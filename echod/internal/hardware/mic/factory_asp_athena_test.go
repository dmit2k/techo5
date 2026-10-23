//go:build !dot && !spot && !panel

package mic

import (
	"encoding/binary"
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

// The normal FIR path updates athenaMeter, but the native ASP path uses
// factoryAux16. This test prevents silently reporting -150 dBFS when
// diagnostic/raw microphone and AWB samples actually contain signal.
func TestFactoryAux16TracksNonzeroRawMicrophoneAndReference(t *testing.T) {
	const frames = 16
	const step = 6
	c := &pairedCapture{}
	raw := make([]byte, frames*step*4*4)
	ref := make([]byte, frames*step*2*2)
	out := make([]byte, frames*Channels*Bits/8)
	for f := 0; f < frames; f++ {
		binary.LittleEndian.PutUint32(raw[f*step*4*4:], uint32(1<<29))
		binary.LittleEndian.PutUint16(ref[f*step*2*2:], 8192)
	}
	c.factoryAux16(out, raw, ref)
	if c.meter.source != "raw_1_in_6" {
		t.Fatalf("meter source = %q, want explicitly coarse raw", c.meter.source)
	}
	if c.meter.mic[0] <= 0 || c.meter.micPeak[0] <= 0 {
		t.Fatal("active microphone was falsely reported as silent")
	}
	if c.meter.ref[0] <= 0 {
		t.Fatal("active hardware reference was falsely reported as silent")
	}
	if c.meter.mic[1] != 0 || c.meter.ref[1] != 0 {
		t.Fatal("zero side channels were not preserved")
	}
	if out[0] == 0 && out[1] == 0 && out[2] == 0 {
		t.Fatal("diagnostic output did not preserve microphone signal")
	}
}
