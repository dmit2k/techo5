package synthetic

import (
	"encoding/binary"
	"fmt"
	"slices"
	"testing"

	"github.com/HuskerMinion/techo5/echod/internal/asp"
)

func config(m, r int) asp.Config {
	c := asp.Config{Flags: asp.Interleaved, Mic: asp.Stream{Channels: uint16(m), Frames: 1920, Rate: 96000, Format: asp.S32, FullBits: 23, ValidBits: 24}}
	if r > 0 {
		c.Flags |= asp.Synchronized | asp.HardwareRef
		c.Ref = asp.Stream{Channels: uint16(r), Frames: 1920, Rate: 96000, Format: asp.S16, FullBits: 15, ValidBits: 16}
	}
	return c
}
func input(c asp.Config) asp.Input {
	m, r, _ := c.Lengths()
	return asp.Input{Config: c, Mic: make([]byte, m), Ref: make([]byte, r)}
}
func TestMappingMatrixAndIsolation(t *testing.T) {
	for m := 1; m <= 8; m++ {
		for r := 0; r <= 2; r++ {
			t.Run(fmt.Sprintf("%d+%d", m, r), func(t *testing.T) {
				c := config(m, r)
				p := Profile{Mic: m, Ref: r, GeometryValidated: true, IdentityValidated: true, NoReference: true}
				a, e := New(c, p, Mapping{})
				if e != nil {
					t.Fatal(e)
				}
				in := input(c)
				for f := 0; f < 1920; f++ {
					for ch := 0; ch < m; ch++ {
						binary.LittleEndian.PutUint32(in.Mic[(f*m+ch)*4:], uint32(int32(ch*100-f-17)))
					}
					for ch := 0; ch < r; ch++ {
						binary.LittleEndian.PutUint16(in.Ref[(f*r+ch)*2:], uint16(int16(ch*333-f-19)))
					}
				}
				for _, f := range []int{0, 11, 1919} {
					out := make([]int32, m+r)
					if e := a.PrepareFrame(in, f, out); e != nil {
						t.Fatal(e)
					}
					for ch := 0; ch < m; ch++ {
						if out[ch] != int32(ch*100-f-17) {
							t.Fatal("mic")
						}
					}
					for ch := 0; ch < r; ch++ {
						if out[m+ch] != int32(int16(ch*333-f-19))<<16 {
							t.Fatal("ref")
						}
					}
				}
			})
		}
	}
	for _, m := range []int{4, 7, 8} {
		c := config(m, 2)
		mm := make([]int, m)
		for ch := range mm {
			mm[ch] = (ch + 2) % m
		}
		a, e := New(c, Profile{Mic: m, Ref: 2, GeometryValidated: true}, Mapping{mm, []int{1, 0}})
		if e != nil {
			t.Fatal(e)
		}
		// Mutating caller's map must not change the accepted session.
		want := slices.Clone(mm)
		mm[0] = 0
		in := input(c)
		for ch := 0; ch < m; ch++ {
			binary.LittleEndian.PutUint32(in.Mic[ch*4:], uint32(ch+11))
		}
		binary.LittleEndian.PutUint16(in.Ref, 0xfffe)
		binary.LittleEndian.PutUint16(in.Ref[2:], 77)
		out := make([]int32, m+2)
		a.PrepareFrame(in, 0, out)
		for ch, v := range want {
			if out[ch] != int32(v+11) {
				t.Fatal(m, ch, out)
			}
		}
		if out[m] != 77<<16 || out[m+1] != -2<<16 || !slices.Equal(a.CalibrationIndices(), want) {
			t.Fatal("L/R or capture-indexed gain association", out)
		}
	}
	c := config(8, 0)
	p := Profile{Mic: 4, GeometryValidated: true, AllowSubset: true, NoReference: true}
	a, e := New(c, p, Mapping{Mic: []int{7, 3, 0, 6}})
	if e != nil {
		t.Fatal(e)
	}
	in := input(c)
	for ch := 0; ch < 8; ch++ {
		binary.LittleEndian.PutUint32(in.Mic[ch*4:], uint32(ch+1))
	}
	dst := make([]int32, 4)
	a.PrepareFrame(in, 0, dst)
	if !slices.Equal(dst, []int32{8, 4, 1, 7}) {
		t.Fatal(dst)
	}
	if a.Validate(c, Mapping{Mic: []int{7, 3, 0, 5}}) == nil {
		t.Fatal("dynamic map change")
	}
}
func TestMappingRejectsAndFactoryProfile(t *testing.T) {
	c := config(4, 2)
	p := Profile{Mic: 4, Ref: 2, GeometryValidated: true}
	for _, m := range []Mapping{
		{}, {Mic: []int{0, 1, 2}, Ref: []int{0, 1}}, {Mic: []int{0, 1, 1, 3}, Ref: []int{0, 1}},
		{Mic: []int{0, 1, 2, 4}, Ref: []int{0, 1}}, {Mic: []int{0, 1, 2, -1}, Ref: []int{0, 1}},
		{Mic: []int{0, 1, 2, 3}, Ref: []int{0, 2}}, {Mic: []int{0, 1, 2, 3}, Ref: []int{0, 0}},
	} {
		if _, e := New(c, p, m); e == nil {
			t.Fatal("bad map accepted", m)
		}
	}
	if _, e := New(c, Profile{Mic: 4, Ref: 2, IdentityValidated: true}, Mapping{}); e == nil {
		t.Fatal("unvalidated geometry")
	}
	for m := 1; m <= 8; m++ {
		for r := 0; r <= 2; r++ {
			_, e := New(config(m, r), Profile{Mic: 4, Ref: 2, GeometryValidated: true, IdentityValidated: true, FactoryShape: true}, Mapping{})
			if (e == nil) != (m == 4 && r == 2) {
				t.Fatal(m, r, e)
			}
		}
	}
	if _, e := New(config(8, 0), Profile{Mic: 8, GeometryValidated: true, IdentityValidated: true}, Mapping{}); e == nil {
		t.Fatal("implicit no-ref")
	}
}
func TestSignedScaleClippingAndSilence(t *testing.T) {
	for _, s := range []asp.Stream{
		{Format: asp.S32, FullBits: 23, ValidBits: 24}, {Format: asp.S32, FullBits: 31, ValidBits: 24, Alignment: asp.High},
		{Format: asp.S32, FullBits: 31, ValidBits: 32}, {Format: asp.S16, FullBits: 15, ValidBits: 16},
	} {
		var values []int32
		if s.FullBits == 23 {
			values = []int32{-8388608, -257, -1, 0, 1, 257, 8388607, 2147483647, -2147483648}
		} else if s.FullBits == 31 {
			values = []int32{-2147483648, -257, -256, 0, 256, 257, 2147483647}
		} else {
			values = []int32{-32768, -1, 0, 1, 32767}
		}
		for _, v := range values {
			want := int64(v)
			if s.FullBits > 23 {
				want >>= s.FullBits - 23
			} else {
				want <<= 23 - s.FullBits
			}
			want = max(int64(-8388608), min(int64(8388607), want))
			if got := scaleMic(v, s); got != int32(want) {
				t.Fatal(s, v, got, want)
			}
		}
	}
	for _, v := range []int32{-32768, -1, 0, 1, 32767} {
		if scaleRef(v, asp.Stream{FullBits: 15}) != v<<16 {
			t.Fatal("reference sign/scale")
		}
	}
	// Identity low24 must not lose the lowest eight valid bits.
	if scaleMic(257, asp.Stream{FullBits: 23}) != 257 {
		t.Fatal("unconditional historical >>8")
	}
}
func TestCausalHundredThousandTicks(t *testing.T) {
	c := config(4, 2)
	a, e := New(c, Profile{Mic: 4, Ref: 2, GeometryValidated: true, IdentityValidated: true, FactoryShape: true}, Mapping{})
	if e != nil {
		t.Fatal(e)
	}
	emu := Emulator{Adapter: a, Markers: true}
	in := input(c)
	var next uint64
	for tick := 0; tick < 100000; tick++ {
		q := emu.Quanta
		r, e := emu.Step(in, uint64(tick))
		if e != nil {
			t.Fatal(tick, e)
		}
		wantQ := uint64(2)
		if tick%2 == 1 {
			wantQ = 3
		}
		if emu.Quanta-q != wantQ {
			t.Fatal("2/3 quantum order")
		}
		if tick == 0 {
			if r.Type != asp.NotReady || r.Start != asp.NoSample || len(r.PCM) != 0 {
				t.Fatal("fabricated initial output")
			}
		} else {
			if r.Type != asp.Data || r.Start != next || len(r.PCM) != 320 {
				t.Fatal(tick, r.Start, next)
			}
			for j, v := range r.PCM {
				if v != Marker(next+uint64(j)) {
					t.Fatal("duplicate/drop/padding", tick, j)
				}
			}
			next += 320
		}
		n, m := emu.Pending()
		if n >= 768 || m > 320 {
			t.Fatal("unbounded carry", n, m)
		}
	}
	n, m := emu.Pending()
	if next != 31999680 || n != 0 || m != 320 || emu.InputFrames != 192000000 || emu.Quanta != 250000 {
		t.Fatal(next, n, m, emu.InputFrames, emu.Quanta)
	}
	t.Logf("ticks=100000 input_frames=%d quanta=%d emitted=%d pending_native=%d pending_mono=%d max_native_carry=%d max_mono_carry=%d", emu.InputFrames, emu.Quanta, next, n, m, emu.MaxNativeCarry, emu.MaxOutputCarry)
}
