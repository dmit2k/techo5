//go:build !dot && !spot && !panel

package mic

// An explicitly opt-in PRIVATE experiment on athena mainline: keep the original
// factory ARM32 library in /vendor, with the original Android 9 Bionic runtime
// from this device. Only a small standalone DSP worker travels with the test
// build, never the factory library or configuration files.

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/HuskerMinion/techo5/echod/internal/layout"
)

const (
	factoryModeFile     = "/data/techo5-linux/fireos-asp.enabled"
	factoryRoot         = "/data/techo5-linux/amazon-arm32"
	factoryInputFrames  = 768
	factoryInputBytes   = factoryInputFrames * 6 * 4
	factoryOutputFrames = 128
)

type factoryASP struct {
	cmd         *exec.Cmd
	stdin       io.WriteCloser
	stdout      io.ReadCloser
	cal         [4]int64
	pendingIn   []byte
	pendingOut  []int16
	outputFrame []int16
	blocks      uint64
}

// factoryASPSelected gates the hardware AWB capture and native DSP together.
func factoryASPSelected() (bool, error) {
	if !layout.AthenaMainline() {
		return false, nil
	}
	_, err := os.Stat(factoryModeFile)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return err == nil, err
}

// Start only on athena mainline when the operator has explicitly selected
// experimental stereo + factory ASP. Failure does not quietly switch engines.
func openFactoryASP() (*factoryASP, error) {
	enabled, err := factoryASPSelected()
	if err != nil {
		return nil, err
	}
	if !enabled {
		return nil, nil
	}
	f := &factoryASP{}
	for ch := 0; ch < 4; ch++ {
		f.cal[ch] = 1 << 30
		v := layout.Idme(fmt.Sprintf("miccal.%d", ch))
		if v == "" {
			slog.Warn("factory ASP: IDME calibration absent, using unity", "channel", ch)
			continue
		}
		n, e := strconv.ParseInt(v, 10, 32)
		if e != nil || n <= 0 {
			slog.Warn("factory ASP: invalid calibration, using unity", "channel", ch)
			continue
		}
		if n&0x7fff0000 == 0 {
			n <<= 16
		}
		f.cal[ch] = n
	}
	shell := factoryRoot + "/system/bin/sh-native"
	preload := factoryRoot + "/libasp_stream_worker.so"
	cmd := exec.Command(shell, "-c", "exit 0")
	cmd.Env = append(os.Environ(),
		"TECHO5_ASP_STREAM=1",
		"LD_PRELOAD="+preload,
		"LD_LIBRARY_PATH=/vendor/lib:"+factoryRoot+"/system/lib:"+factoryRoot+"/system/lib/vndk-28:"+factoryRoot+"/system/lib/vndk-sp-28")
	in, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	out, err := cmd.StdoutPipe()
	if err != nil {
		_ = in.Close()
		return nil, err
	}
	cmd.Stderr = os.Stderr
	if err = cmd.Start(); err != nil {
		_ = in.Close()
		_ = out.Close()
		return nil, fmt.Errorf("factory ASP exec: %w", err)
	}
	// A stalled or broken Bionic loader must never silently fall back to TECHO5.
	ready := make(chan error, 1)
	go func() {
		var hdr [8]byte
		e := func() error {
			if _, e := io.ReadFull(out, hdr[:]); e != nil {
				return e
			}
			if string(hdr[:]) != "ASPREADY" {
				return fmt.Errorf("ASP startup handshake: %q", hdr)
			}
			return nil
		}()
		ready <- e
	}()
	select {
	case err = <-ready:
	case <-time.After(8 * time.Second):
		err = errors.New("factory ASP startup timeout")
	}
	if err != nil {
		_ = in.Close()
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		return nil, fmt.Errorf("factory ASP startup: %w", err)
	}
	f.cmd, f.stdin, f.stdout = cmd, in, out
	slog.Info("factory Amazon ASP enabled", "worker", cmd.Process.Pid, "ref", "hardware hw:0,1 DL1_AWB_Record stereo S16LE 96k")
	return f, nil
}

func (f *factoryASP) Close() {
	if f == nil {
		return
	}
	_ = f.stdin.Close()
	_ = f.stdout.Close()
	if f.cmd != nil {
		_ = f.cmd.Process.Kill()
		_ = f.cmd.Wait()
	}
}

func clamp24(v int64) int32 {
	if v > (1<<23)-1 {
		return (1 << 23) - 1
	}
	if v < -(1 << 23) {
		return -(1 << 23)
	}
	return int32(v)
}

// Process20ms accepts one native S32LE 96k four-microphone read and the
// *actual* stereo 96k S16LE hardware DL1_AWB_Record data. Both streams use
// the same AFE clock. The remaining start/latency offset must be measured.
// The original ASP needs 8ms chunks, while TECHO5 listeners use 20ms.
func (f *factoryASP) Process20ms(raw, reference []byte, outputFrames int) ([]int16, error) {
	if f == nil {
		return nil, errors.New("ASP is not enabled")
	}
	if len(raw)%16 != 0 {
		return nil, errors.New("ASP microphones must be four S32 channels")
	}
	n := len(raw) / 16
	if n != outputFrames*6 {
		return nil, fmt.Errorf("ASP mic frames %d, output frames %d", n, outputFrames)
	}
	if len(reference) != n*4 {
		return nil, fmt.Errorf("ASP needs real AWB stereo ref: got %d bytes, need %d", len(reference), n*4)
	}
	var packet [24]byte
	for i := 0; i < n; i++ {
		at := i * 16
		for ch := 0; ch < 4; ch++ {
			mic := int32(binary.LittleEndian.Uint32(raw[at+4*ch:])) >> 8
			scaled := clamp24(int64(mic) * f.cal[ch] >> 30)
			binary.LittleEndian.PutUint32(packet[ch*4:], uint32(scaled))
		}
		// The factory ASP uses 32-bit reference slots; observed input
		// representation is signed S16 shifted to bits 31:16.
		for ch := 0; ch < 2; ch++ {
			x := int16(binary.LittleEndian.Uint16(reference[i*4+ch*2:]))
			binary.LittleEndian.PutUint32(packet[16+ch*4:], uint32(int32(x)<<16))
		}
		f.pendingIn = append(f.pendingIn, packet[:]...)
	}
	for len(f.pendingIn) >= factoryInputBytes {
		if _, err := f.stdin.Write(f.pendingIn[:factoryInputBytes]); err != nil {
			return nil, fmt.Errorf("ASP input write: %w", err)
		}
		var block [factoryOutputFrames * 2]byte
		if _, err := io.ReadFull(f.stdout, block[:]); err != nil {
			return nil, fmt.Errorf("ASP output read: %w", err)
		}
		for i := 0; i < factoryOutputFrames; i++ {
			f.pendingOut = append(f.pendingOut, int16(binary.LittleEndian.Uint16(block[i*2:])))
		}
		f.pendingIn = f.pendingIn[factoryInputBytes:]
		f.blocks++
	}
	// Fixed 4ms startup FIFO: first 20ms input contains 2*8ms outputs;
	// subsequent reads alternate 3 and 2 blocks. Do not reorder audio.
	if f.blocks == 2 && len(f.pendingOut) == 256 {
		aligned := make([]int16, 64, 64+len(f.pendingOut))
		f.pendingOut = append(aligned, f.pendingOut...)
	}
	if len(f.pendingOut) < outputFrames {
		return nil, fmt.Errorf("ASP output underrun: %d samples available, need %d", len(f.pendingOut), outputFrames)
	}
	f.outputFrame = append(f.outputFrame[:0], f.pendingOut[:outputFrames]...)
	f.pendingOut = f.pendingOut[outputFrames:]
	return f.outputFrame, nil
}
