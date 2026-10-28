// Package synthetic is an offline DSP-shape emulator, not a factory runtime.
// Channel maps and validated geometry belong to this helper side, never to the capture host.
package synthetic

import (
	"encoding/binary"
	"errors"

	"github.com/HuskerMinion/techo5/echod/internal/asp"
)

var ErrMapping = errors.New("unsupported synthetic DSP mapping")

type Profile struct {
	Mic, Ref                                                                     int
	GeometryValidated, IdentityValidated, AllowSubset, NoReference, FactoryShape bool
}
type Mapping struct{ Mic, Ref []int }
type Adapter struct {
	config   asp.Config
	mic, ref []int
	profile  Profile
}

func channelMap(given []int, captured, logical int, identity, subset bool) ([]int, error) {
	if logical > captured || logical < 0 || !subset && logical != captured {
		return nil, ErrMapping
	}
	if logical == 0 && captured == 0 && len(given) == 0 {
		return []int{}, nil
	}
	if given == nil {
		if !identity {
			return nil, ErrMapping
		}
		given = make([]int, logical)
		for i := range given {
			given[i] = i
		}
	}
	if len(given) != logical {
		return nil, ErrMapping
	}
	seen := uint32(0)
	for _, v := range given {
		if v < 0 || v >= captured || seen&(1<<v) != 0 {
			return nil, ErrMapping
		}
		seen |= 1 << v
	}
	return append([]int{}, given...), nil
}
func New(c asp.Config, p Profile, m Mapping) (*Adapter, error) {
	if _, _, e := c.Lengths(); e != nil {
		return nil, e
	}
	if !p.GeometryValidated {
		return nil, ErrMapping
	}
	if c.Ref.Channels == 0 && !p.NoReference {
		return nil, ErrMapping
	}
	if p.FactoryShape && (c.Mic.Channels != 4 || c.Ref.Channels != 2 || c.Mic.Rate != 96000 || c.Ref.Rate != 96000 || c.Mic.Format != asp.S32 || c.Ref.Format != asp.S16 || p.Mic != 4 || p.Ref != 2) {
		return nil, ErrMapping
	}
	mic, e := channelMap(m.Mic, int(c.Mic.Channels), p.Mic, p.IdentityValidated, p.AllowSubset)
	if e != nil {
		return nil, e
	}
	ref, e := channelMap(m.Ref, int(c.Ref.Channels), p.Ref, p.IdentityValidated, p.AllowSubset)
	if e != nil {
		return nil, e
	}
	return &Adapter{c, mic, ref, p}, nil
}
func (a *Adapter) Validate(c asp.Config, m Mapping) error {
	b, e := New(c, a.profile, m)
	if e != nil {
		return e
	}
	if c != a.config || !equal(a.mic, b.mic) || !equal(a.ref, b.ref) {
		return ErrMapping
	}
	return nil
}
func equal(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// CalibrationIndices keeps synthetic capture-indexed gains attached to their physical signals.
// It deliberately applies no gain: real libasp internal /proc miccal visibility is unresolved.
func (a *Adapter) CalibrationIndices() []int { return append([]int{}, a.mic...) }
func sample(b []byte, s asp.Stream, frame, ch int) int32 {
	width := 2
	if s.Format == asp.S32 {
		width = 4
	}
	at := (frame*int(s.Channels) + ch) * width
	if width == 2 {
		return int32(int16(binary.LittleEndian.Uint16(b[at:])))
	}
	return int32(binary.LittleEndian.Uint32(b[at:]))
}
func clamp24(v int64) int32 {
	if v > 8388607 {
		return 8388607
	}
	if v < -8388608 {
		return -8388608
	}
	return int32(v)
}
func scaleMic(v int32, s asp.Stream) int32 {
	// fullBits23 low24 is already the factory low24 scale. FullBits31 maps explicitly by >>8.
	n := int(s.FullBits) - 23
	if n >= 0 {
		return clamp24(int64(v) >> n)
	}
	return clamp24(int64(v) << (-n))
}
func scaleRef(v int32, s asp.Stream) int32 {
	n := 31 - int(s.FullBits)
	if n >= 0 {
		return int32(int64(v) << n)
	}
	return v >> (-n)
}

// PrepareFrame leaves source order/time intact and routes DSP destinations before interleave.
// Different stream rates are transport-valid but this fixed-quantum emulator requires equal rates.
func (a *Adapter) PrepareFrame(i asp.Input, frame int, dst []int32) error {
	if i.Config != a.config || i.Validate() != nil || a.config.Mic.Rate != a.config.Ref.Rate && a.config.Ref.Channels != 0 || len(dst) != len(a.mic)+len(a.ref) || frame < 0 || frame >= int(a.config.Mic.Frames) {
		return ErrMapping
	}
	return a.prepareFrame(i, frame, dst)
}
func (a *Adapter) prepareFrame(i asp.Input, frame int, dst []int32) error {
	for ch, src := range a.mic {
		dst[ch] = scaleMic(sample(i.Mic, i.Config.Mic, frame, src), i.Config.Mic)
	}
	for ch, src := range a.ref {
		dst[len(a.mic)+ch] = scaleRef(sample(i.Ref, i.Config.Ref, frame, src), i.Config.Ref)
	}
	return nil
}
func Marker(n uint64) int16 { n ^= n >> 17; n *= 0x9e3779b185ebca87; n ^= n >> 29; return int16(n) }

type Emulator struct {
	Adapter                        *Adapter
	Markers                        bool
	quantum                        [768 * 6]int32
	inputCount                     int
	output                         [640]int16
	outputCount                    int
	generated, emitted             uint64
	Quanta, InputFrames            uint64
	MaxNativeCarry, MaxOutputCarry int
}

func (e *Emulator) Step(i asp.Input, seq uint64) (asp.Reply, error) {
	if e.Adapter == nil || !e.Adapter.profile.FactoryShape || i.Config != e.Adapter.config || i.Validate() != nil {
		return asp.Reply{}, ErrMapping
	}
	// Exactly 1920 native samples arrive; each complete quantum consumes 768 and produces 128.
	for f := 0; f < int(i.Config.Mic.Frames); f++ {
		if err := e.Adapter.prepareFrame(i, f, e.quantum[e.inputCount*6:(e.inputCount+1)*6]); err != nil {
			return asp.Reply{}, err
		}
		e.inputCount++
		e.InputFrames++
		if e.inputCount == 768 {
			if e.outputCount+128 > len(e.output) {
				return asp.Reply{}, ErrMapping
			}
			for j := 0; j < 128; j++ {
				v := int16(e.quantum[j*6*6] >> 8)
				if e.Markers {
					v = Marker(e.generated + uint64(j))
				}
				e.output[e.outputCount+j] = v
			}
			e.outputCount += 128
			e.generated += 128
			e.inputCount = 0
			e.Quanta++
		}
	}
	e.MaxNativeCarry = max(e.MaxNativeCarry, e.inputCount)
	r := asp.Reply{Type: asp.NotReady, Sequence: seq, Start: asp.NoSample}
	if e.outputCount >= asp.Samples {
		r.Type, r.Start = asp.Data, e.emitted
		r.PCM = append([]int16(nil), e.output[:asp.Samples]...)
		e.outputCount = copy(e.output[:], e.output[asp.Samples:e.outputCount])
		e.emitted += asp.Samples
	}
	e.MaxOutputCarry = max(e.MaxOutputCarry, e.outputCount)
	return r, nil
}
func (e *Emulator) Pending() (native, mono int) { return e.inputCount, e.outputCount }
