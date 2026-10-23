//go:build !dot && !spot && !panel

package mic

// Private, time-bounded research probe of the *actual* Amazon ASP output.
// It is disabled unless the owner explicitly creates pcm-probe.enable.
// Only owner-local data; no HTTP, HA, or upstream publication.

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"log/slog"
	"math"
	"os"
	"path/filepath"
)

const (
	factoryProbeRoot       = "/data/techo5-linux/frankenasp-trial"
	factoryProbeFlag       = factoryProbeRoot + "/pcm-probe.enable"
	factoryProbeSkipFrames = 2 * Rate // skip initial DSP warmup
	factoryProbeMaxFrames  = 8 * Rate // capture eight seconds exactly
)

type factoryPCMProbe struct {
	files                     [3]*os.File
	writers                   [3]*bufio.Writer
	passed, saved             int
	count, nonzero, fullscale uint64
	squared                   float64
	peak                      int
	closed                    bool
}

func openFactoryPCMProbe() (*factoryPCMProbe, error) {
	if _, err := os.Stat(factoryProbeFlag); err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("factory PCM probe flag: %w", err)
	}
	p := &factoryPCMProbe{}
	names := [...]string{"pcm-probe-mics-s32le-4ch-96k.raw", "pcm-probe-awb-s16le-2ch-96k.raw", "pcm-probe-factory-s16le-1ch-16k.raw"}
	for i, name := range names {
		filename := filepath.Join(factoryProbeRoot, name)
		file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if err != nil {
			p.Close()
			return nil, fmt.Errorf("factory PCM probe open %s: %w", name, err)
		}
		p.files[i] = file
		p.writers[i] = bufio.NewWriterSize(file, 128*1024)
	}
	slog.Info("factory PCM probe armed", "skip_seconds", 2, "capture_seconds", 8, "output_private", true)
	return p, nil
}

func (p *factoryPCMProbe) push(raw, ref []byte, out []int16) {
	if p == nil || p.closed {
		return
	}
	frames := len(out)
	if frames == 0 {
		return
	}
	if p.passed < factoryProbeSkipFrames {
		p.passed += frames
		return
	}
	if p.saved >= factoryProbeMaxFrames {
		return
	}
	if p.saved+frames > factoryProbeMaxFrames {
		frames = factoryProbeMaxFrames - p.saved
	}
	micBytes := frames * 6 * 4 * 4 // four 96k S32LE microphone channels
	refBytes := frames * 6 * 2 * 2 // two 96k S16LE hardware AWB channels
	if len(raw) < micBytes || len(ref) < refBytes {
		slog.Error("factory PCM probe unexpected input length", "raw", len(raw), "ref", len(ref), "out", len(out))
		p.Close()
		return
	}
	var pcm [320 * 2]byte
	for i := 0; i < frames; i++ {
		v := out[i]
		binary.LittleEndian.PutUint16(pcm[2*i:], uint16(v))
		a := int(v)
		if a < 0 {
			a = -a
		}
		if a > p.peak {
			p.peak = a
		}
		if v != 0 {
			p.nonzero++
		}
		if v == -32768 || v == 32767 {
			p.fullscale++
		}
		p.squared += float64(v) * float64(v)
		p.count++
	}
	for i, value := range [][]byte{raw[:micBytes], ref[:refBytes], pcm[:frames*2]} {
		if _, err := p.writers[i].Write(value); err != nil {
			slog.Error("factory PCM probe write failed", "stream", i, "err", err)
			p.Close()
			return
		}
	}
	p.saved += frames
	if p.saved == factoryProbeMaxFrames {
		p.Close()
	}
}

func (p *factoryPCMProbe) Close() {
	if p == nil || p.closed {
		return
	}
	p.closed = true
	for i := range p.files {
		if p.writers[i] != nil {
			if err := p.writers[i].Flush(); err != nil {
				slog.Error("factory PCM probe flush failed", "stream", i, "err", err)
			}
		}
		if p.files[i] != nil {
			if err := p.files[i].Close(); err != nil {
				slog.Error("factory PCM probe close failed", "stream", i, "err", err)
			}
		}
	}
	var rms float64
	if p.count > 0 {
		rms = math.Sqrt(p.squared/float64(p.count)) / 32768
	}
	dbfs := -150.0
	if rms > 0 {
		dbfs = 20 * math.Log10(rms)
	}
	slog.Info("factory PCM probe completed", "frames", p.saved, "duration_seconds", float64(p.saved)/Rate, "rms_dbfs", dbfs, "peak_pcm16", p.peak,
		"nonzero_samples", p.nonzero, "clipped_samples", p.fullscale, "private_files", true)
}
