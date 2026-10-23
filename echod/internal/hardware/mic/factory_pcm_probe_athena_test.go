//go:build !dot && !spot && !panel

package mic

import (
	"bufio"
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
)

func TestFactoryPCMProbeCapturesMatchedInputOutput(t *testing.T) {
	root := t.TempDir()
	p := &factoryPCMProbe{}
	names := []string{"mic.raw", "awb.raw", "factory.pcm"}
	for i, name := range names {
		file, err := os.OpenFile(filepath.Join(root, name), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if err != nil {
			t.Fatal(err)
		}
		p.files[i] = file
		p.writers[i] = bufio.NewWriter(file)
	}
	frames := 320
	mic := make([]byte, frames*6*4*4)
	ref := make([]byte, frames*6*2*2)
	out := make([]int16, frames)
	for i := 0; i < frames*6; i++ {
		binary.LittleEndian.PutUint32(mic[i*16:], 123456)
		binary.LittleEndian.PutUint16(ref[i*4:], 42)
	}
	for i := range out {
		out[i] = 1500
	}
	p.passed = factoryProbeSkipFrames
	for i := 0; i < factoryProbeMaxFrames/frames; i++ {
		p.push(mic, ref, out)
	}
	if !p.closed {
		t.Fatal("probe did not stop after 8s")
	}
	if p.saved != factoryProbeMaxFrames {
		t.Fatalf("captured %d frames", p.saved)
	}
	if p.count != uint64(factoryProbeMaxFrames) || p.nonzero != p.count || p.fullscale != 0 {
		t.Fatalf("invalid stats count=%d nonzero=%d fullscale=%d", p.count, p.nonzero, p.fullscale)
	}
	expected := []int{factoryProbeMaxFrames * 6 * 4 * 4, factoryProbeMaxFrames * 6 * 2 * 2, factoryProbeMaxFrames * 2}
	for i, name := range names {
		got, err := os.Stat(filepath.Join(root, name))
		if err != nil {
			t.Fatal(err)
		}
		if got.Size() != int64(expected[i]) {
			t.Fatalf("file %s: %d want %d", name, got.Size(), expected[i])
		}
	}
	file, err := os.ReadFile(filepath.Join(root, "factory.pcm"))
	if err != nil {
		t.Fatal(err)
	}
	if binary.LittleEndian.Uint16(file[:2]) != 1500 {
		t.Fatal("wrong output PCM content")
	}
}

func TestFactoryPCMProbeSkipAndNoExtra(t *testing.T) {
	p := &factoryPCMProbe{}
	p.passed = 0
	p.push(make([]byte, 320*6*4*4), make([]byte, 320*6*2*2), make([]int16, 320))
	if p.saved != 0 || p.passed != 320 {
		t.Fatalf("warmup not skipped: %+v", p)
	}
}
