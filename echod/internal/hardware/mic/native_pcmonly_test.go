package mic

import (
	"bytes"
	"strings"
	"testing"

	"github.com/HuskerMinion/techo5/echod/internal/asp"
	"github.com/HuskerMinion/techo5/echod/internal/lib/alsa"
)

// Stage03-R: metadata is NOT an argument to any native capture API. The same
// synthetic PCM descriptor must encode identically regardless of the existence,
// completeness or volume of unrelated factory bytes.
func TestN07PCMOnlyDescriptorIgnoresAllSyntheticMetadata(t *testing.T) {
	mic := []byte{1, 2, 3, 4, 5, 6}
	block := nativeInterleaved(mic, 16000, asp.FormatS24_3, 24, 0, 1, 0)
	expected, err := asp.Encode(block.Descriptor())
	if err != nil {
		t.Fatal(err)
	}
	if len(expected) != 16+28+8 {
		t.Fatalf("non-PCM descriptor: %d", len(expected))
	}
	variants := [][]byte{nil, {}, {0, 0xff, 0x81}, bytes.Repeat([]byte{0xae}, 1<<20)}
	for _, foreignBytes := range variants {
		_ = foreignBytes // deliberately NO metadata input to production descriptor
		encoded, err := asp.Encode(block.Descriptor())
		if err != nil || !bytes.Equal(expected, encoded) {
			t.Fatalf("synthetic metadata length %d changed PCM descriptor: %v", len(foreignBytes), err)
		}
	}
	if &block.Buffers[0][0] != &mic[0] {
		t.Fatal("borrowed input copied")
	}
	for _, ch := range block.Geometry.Channels {
		if ch.Role == asp.RoleREF {
			t.Fatal("absent REF was invented")
		}
	}
}

// Plain device channel layout is known only for its MIC slots, or the complete
// stock MIC/REF layout. Unknown ALSA formats and extra channels are errors to a
// requested native consumer, never guessed as S24 or REF.
func TestN07PlainRejectsAmbiguousFormatAndRoles(t *testing.T) {
	original := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}
	cases := []struct {
		name string
		cfg  alsa.Config
	}{
		{"unknown-format", alsa.Config{Channels: Mics, Rate: 16000, Format: 199, Bits: 24}},
		{"unknown-extra-channel", alsa.Config{Channels: Channels + 1, Rate: 16000, Format: alsa.FormatS24_3LE, Bits: 24}},
		{"inconsistent-width", alsa.Config{Channels: Mics, Rate: 16000, Format: alsa.FormatS16_LE, Bits: 24}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			frames := bytes.Repeat(original, max(1, tc.cfg.Channels*3))
			capture := &fakePlainCapture{cfg: tc.cfg, bytes: frames}
			out := make([]byte, len(frames))
			calls, visits := 0, 0
			n, cut, err := readCaptured(capture, out, func() bool { calls++; return false }, func(NativeBlock) { visits++ })
			if err == nil || !strings.Contains(err.Error(), "mic:") || n != len(frames) || cut || visits != 0 || calls != 1 {
				t.Fatalf("silent guessed geometry: n=%d cut=%t err=%v visits=%d policy=%d", n, cut, err, visits, calls)
			}
			if !bytes.Equal(frames, out) {
				t.Fatal("stock non-muted capture modified")
			}
			// With no native consumer, the previous stock Read path remains usable.
			n, _, err = readCaptured(capture, out, fixedCut(false), nil)
			if err != nil || n != len(frames) || !bytes.Equal(frames, out) {
				t.Fatal("stock capture broken by invalid optional native geometry")
			}
			n, cut, err = readCaptured(capture, out, fixedCut(true), func(NativeBlock) { t.Fatal("muted unknown geometry published") })
			if err == nil || !cut || n != len(frames) || !bytes.Equal(out, make([]byte, n)) {
				t.Fatalf("unknown PCM must fail closed under cut: cut=%t err=%v", cut, err)
			}
		})
	}
}

// Both hardware REF slots remain separate and byte-exact in the original single
// interleaved buffer. This is independent of stock AEC's own output mix.
func TestN02PlainStereoHardwareRefByteExact(t *testing.T) {
	if Refs != 2 {
		t.Skip("stock Panel has a mono software reference, not hardware stereo REF")
	}
	cfg := alsa.Config{Channels: Channels, Rate: 48000, Format: alsa.FormatS24_3LE, Bits: 24}
	buf := make([]byte, 2*Channels*3)
	left, right := []byte{0xa1, 0xb2, 0xc3}, []byte{0xd4, 0xe5, 0xf6}
	for f := 0; f < 2; f++ {
		off := f * Channels * 3
		for c := 0; c < Mics; c++ {
			copy(buf[off+c*3:], []byte{byte(c + 1), 0x55, 0x33})
		}
		copy(buf[off+RefFirst*3:], left)
		copy(buf[off+(RefFirst+1)*3:], right)
	}
	fake := &fakePlainCapture{cfg: cfg, bytes: buf}
	out := make([]byte, len(buf))
	n, cut, err := readCaptured(fake, out, fixedCut(false), func(b NativeBlock) {
		if &out[0] != &b.Buffers[0][0] || !bytes.Equal(buf, b.Buffers[0]) {
			t.Fatal("native PCM copied or mutated")
		}
		if len(b.Geometry.Channels) != Channels {
			t.Fatal("missing channel")
		}
		refs := b.Geometry.Channels[Mics:]
		if len(refs) != Refs || refs[0].Role != asp.RoleREF || refs[0].RoleIndex != 0 ||
			refs[1].Role != asp.RoleREF || refs[1].RoleIndex != 1 ||
			refs[0].FrameOffset != uint32(RefFirst*3) || refs[1].FrameOffset != uint32((RefFirst+1)*3) {
			t.Fatalf("stereo reference mapping incorrect: %+v", refs)
		}
		if !bytes.Equal(b.Buffers[0][RefFirst*3:RefFirst*3+3], left) ||
			!bytes.Equal(b.Buffers[0][(RefFirst+1)*3:(RefFirst+2)*3], right) {
			t.Fatal("REF L/R not byte exact")
		}
		if _, err := asp.Encode(b.Descriptor()); err != nil {
			t.Fatal(err)
		}
	})
	if err != nil || cut || n != len(buf) || !bytes.Equal(out, buf) {
		t.Fatalf("stock changed %v", err)
	}
}

func TestN07PCMOnlyValidPlainS16AndS32(t *testing.T) {
	cases := []struct {
		format int
		bits   int
		codec  uint8
	}{
		{alsa.FormatS16_LE, 16, asp.FormatS16},
		{alsa.FormatS32_LE, 32, asp.FormatS32},
	}
	for _, tc := range cases {
		cfg := alsa.Config{Channels: Mics, Rate: 32000, Format: tc.format, Bits: tc.bits}
		src := bytes.Repeat([]byte{0xb7}, 2*Mics*(tc.bits/8))
		n, err := plainNative(src, cfg)
		if err != nil || n.Geometry.Buffers[0].Format != tc.codec || len(n.Geometry.Channels) != Mics {
			t.Fatalf("format %d: %v %+v", tc.format, err, n.Geometry)
		}
		if _, err := asp.Encode(n.Descriptor()); err != nil {
			t.Fatal(err)
		}
	}
}
