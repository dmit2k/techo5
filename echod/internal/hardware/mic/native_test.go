package mic

import (
	"bytes"
	"testing"
	"time"

	"github.com/HuskerMinion/techo5/echod/internal/asp"
	"github.com/HuskerMinion/techo5/echod/internal/lib/alsa"
)

type fakePlainCapture struct {
	cfg    alsa.Config
	bytes  []byte
	onRead func()
}

func (c *fakePlainCapture) Read(buf []byte) (int, error) {
	if c.onRead != nil {
		c.onRead()
	}
	return copy(buf, c.bytes), nil
}
func (c *fakePlainCapture) Close() error        { return nil }
func (c *fakePlainCapture) Config() alsa.Config { return c.cfg }

func TestN02PlainBorrowedAndNoREF(t *testing.T) {
	// Fake hardware with no REF, not the stock software-generated REF.
	raw := make([]byte, 2*Mics*3)
	for i := range raw {
		raw[i] = byte(i + 1)
	}
	fake := &fakePlainCapture{cfg: alsaConfig(), bytes: raw}
	dst := make([]byte, len(raw))
	visited := 0
	n, _, err := readCaptured(fake, dst, fixedCut(false), func(b NativeBlock) {
		visited++
		if &dst[0] != &b.Buffers[0][0] {
			t.Fatal("plain capture introduced input copy")
		}
		if len(b.Geometry.Buffers) != 1 || b.Geometry.Buffers[0].Frames != 2 ||
			b.Geometry.Buffers[0].SampleRate != 48000 ||
			b.Geometry.Buffers[0].Format != asp.FormatS24_3 ||
			len(b.Geometry.Channels) != Mics || b.Geometry.Channels[len(b.Geometry.Channels)-1].Role != asp.RoleMIC {
			t.Fatalf("wrong plain geometry: %+v", b.Geometry)
		}
		if _, err := asp.Encode(b.Geometry); err != nil {
			t.Fatalf("Stage02 descriptor: %v", err)
		}
	})
	if err != nil || n != len(raw) || visited != 1 || !bytes.Equal(dst, raw) {
		t.Fatalf("plain read %d/%v/%d %x", n, err, visited, dst)
	}
	_, _, err = readCaptured(fake, dst, fixedCut(true), func(b NativeBlock) {
		if !bytes.Equal(b.Buffers[0], make([]byte, len(raw))) {
			t.Fatal("plain native not cut")
		}
	})
	if err != nil || !bytes.Equal(dst, make([]byte, len(raw))) {
		t.Fatal("plain stock not cut")
	}
}
func alsaConfig() alsa.Config {
	return alsa.Config{Channels: Mics, Rate: 48000, Format: alsa.FormatS24_3LE, Bits: 24}
}

func TestN03SparseNativeZeroBounds(t *testing.T) {
	// S32 high-aligned MIC and 2 REF channels in a sparse 20-byte stride.
	// Padding bytes and outer canaries must survive the cut unchanged.
	raw := bytes.Repeat([]byte{0x7a}, 43)
	pcm := raw[1:41]
	n := NativeBlock{
		Buffers: [][]byte{pcm},
		Geometry: asp.Descriptor{
			Buffers: []asp.Buffer{{Length: uint32(len(pcm)), Frames: 2, SampleRate: 96000, Stride: 20,
				Format: asp.FormatS32, ValidBits: 24, LSBShift: 8, Channels: 3}},
			Channels: []asp.Channel{
				{Role: asp.RoleMIC, RoleIndex: 0, FrameOffset: 2},
				{Role: asp.RoleREF, RoleIndex: 0, FrameOffset: 8},
				{Role: asp.RoleREF, RoleIndex: 1, FrameOffset: 14},
			},
		},
	}
	if _, err := asp.Encode(n.Geometry); err != nil {
		t.Fatal(err)
	}
	zeroNative(n)
	for i, v := range raw {
		want := byte(0x7a)
		for _, base := range []int{1, 21} {
			for _, off := range []int{2, 8, 14} {
				if i >= base+off && i < base+off+4 {
					want = 0
				}
			}
		}
		if v != want {
			t.Fatalf("byte %d=%02x wanted %02x", i, v, want)
		}
	}
}

func TestN05OwnedStockRetentionHistoryAndRaw(t *testing.T) {
	s := testSource()
	s.mixer = &reusedBuffer{}
	frames, stop := s.Listen("native-retain")
	defer stop()
	raws, stopRaw := s.ListenRaw()
	defer stopRaw()
	raw := rawFrame(0, false)
	s.broadcastWithCut(raw, false)
	first := <-frames
	original := append([]int16(nil), first...)
	rawFirst := <-raws
	originalRaw := bytes.Clone(rawFirst)
	for i := 1; i <= 5; i++ {
		// Simulate the capture buffer being reused and overwritten.
		copy(raw, rawFrame(i, false))
		s.broadcastWithCut(raw, false)
		<-frames
		<-raws
	}
	for i, v := range first {
		if v != original[i] {
			t.Fatalf("held frame %d mutated", i)
		}
	}
	if !bytes.Equal(rawFirst, originalRaw) {
		t.Fatal("ListenRaw frame mutated")
	}
	recent := s.Recent(20 * time.Millisecond)
	if len(recent) != FrameSamples || recent[0] != 6 {
		t.Fatalf("history incorrect: %v", recent[:min(2, len(recent))])
	}
}

func TestN07NativePCMOnlyDescriptor(t *testing.T) {
	mic := []byte{1, 2, 3, 4, 5, 6}
	block := nativeInterleaved(mic, 16000, asp.FormatS24_3, 24, 0, 1, 0)
	desc := block.Descriptor()
	wire, err := asp.Encode(desc)
	if err != nil {
		t.Fatal(err)
	}
	if len(wire) != 16+28+8 {
		t.Fatalf("unexpected descriptor size %d", len(wire))
	}
	if &block.Buffers[0][0] != &mic[0] {
		t.Fatal("native PCM copied")
	}
	if !bytes.Equal(block.Buffers[0], mic) {
		t.Fatal("input mutated")
	}
}

// fixedCut supplies a synthetic software policy without touching real privacy controls.
func fixedCut(v bool) func() bool { return func() bool { return v } }

func TestN03PolicyExactlyOnceAfterPlainRead(t *testing.T) {
	order, calls := 0, 0
	c := &fakePlainCapture{cfg: alsaConfig(), bytes: make([]byte, Mics*3),
		onRead: func() { order = 1 }}
	out := make([]byte, len(c.bytes))
	_, cut, err := readCaptured(c, out, func() bool {
		if order != 1 {
			t.Fatal("SoftwareCut sampled before capture read")
		}
		calls++
		return true
	}, nil)
	if err != nil || !cut || calls != 1 || !bytes.Equal(out, make([]byte, len(out))) {
		t.Fatalf("single capture policy result: cut=%v calls=%d err=%v", cut, calls, err)
	}
}
