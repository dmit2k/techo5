//go:build !dot && !spot && !panel

package mic

import (
	"os"
	"strings"
	"testing"
	"time"
)

// A Bionic DSP that stops reading or writing must never block the 20-ms
// microphone reader forever; the trial watchdog can then roll back safely.
func TestFactoryASPStalledPipeFailsClosed(t *testing.T) {
	inputReader, inputWriter, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer inputReader.Close()
	defer inputWriter.Close()
	outputReader, outputWriter, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer outputReader.Close()
	defer outputWriter.Close()

	f := &factoryASP{stdin: inputWriter, stdout: outputReader}
	for i := range f.cal {
		f.cal[i] = 1 << 30
	}

	// 320 output frames = 20ms at 16k; four raw mics + stereo ref at 96k.
	const outputFrames = 320
	raw := make([]byte, outputFrames*6*4*4)
	ref := make([]byte, outputFrames*6*2*2)
	start := time.Now()
	_, err = f.Process20ms(raw, ref, outputFrames)
	elapsed := time.Since(start)
	if err == nil {
		t.Fatal("stalled worker unexpectedly returned valid PCM")
	}
	if elapsed > time.Second {
		t.Fatalf("stalled native IPC exceeded fail-closed deadline: %s: %v", elapsed, err)
	}
	if !strings.Contains(err.Error(), "timeout") && !strings.Contains(err.Error(), "deadline") {
		t.Fatalf("unexpected stalled-worker error: %v", err)
	}
}
