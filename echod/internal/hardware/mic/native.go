package mic

import (
	"github.com/HuskerMinion/techo5/echod/internal/asp"
	"github.com/HuskerMinion/techo5/echod/internal/lib/alsa"
)

// NativeBlock borrows the capture's original physical buffers until the NEXT capture Read.
// A synchronous consumer must finish before that read. It must not keep these bytes or
// publish them as output. Geometry describes captured channels, not stock DSP layout.
// Buffers are not concatenated, remapped, resampled, scaled or copied.
type NativeBlock struct {
	Buffers  [][]byte
	Geometry asp.Descriptor
}

// nativeInterleaved describes one actual interleaved capture buffer. The caller supplies
// captured roles, never a DSP-specific channel mapping.
func nativeInterleaved(pcm []byte, rate uint32, format, bits, shift uint8, mics, refs int) NativeBlock {
	width := 3
	switch format {
	case asp.FormatS16:
		width = 2
	case asp.FormatS32:
		width = 4
	}
	channels := mics + refs
	stride := channels * width
	n := NativeBlock{
		Buffers: [][]byte{pcm},
		Geometry: asp.Descriptor{
			Buffers: []asp.Buffer{{Length: uint32(len(pcm)), SampleRate: rate,
				Frames: uint32(len(pcm) / stride), Stride: uint32(stride),
				Format: format, ValidBits: bits, LSBShift: shift, Channels: uint8(channels)}},
		},
	}
	for i := 0; i < channels; i++ {
		role, slot := asp.RoleMIC, i
		if i >= mics {
			role, slot = asp.RoleREF, i-mics
		}
		n.Geometry.Channels = append(n.Geometry.Channels, asp.Channel{
			Role: role, RoleIndex: uint16(slot), FrameOffset: uint32(i * width)})
	}
	return n
}

// zeroNative clears only declared samples, using each channel's frame offset, width and
// stride. Padding and guard bytes are never touched. This is the shared SoftwareCut
// operation on native physical buffers, including both hardware REF L and REF R.
func zeroNative(n NativeBlock) {
	for i, b := range n.Geometry.Buffers {
		width := 2
		switch b.Format {
		case asp.FormatS24_3:
			width = 3
		case asp.FormatS32:
			width = 4
		}
		pcm := n.Buffers[i]
		first := int(b.FirstChannel)
		for _, ch := range n.Geometry.Channels[first : first+int(b.Channels)] {
			for f := uint32(0); f < b.Frames; f++ {
				at := uint64(f)*uint64(b.Stride) + uint64(ch.FrameOffset)
				if at+uint64(width) > uint64(len(pcm)) {
					panic("native PCM geometry exceeds capture buffer")
				}
				clear(pcm[int(at) : int(at)+width])
			}
		}
	}
}

// captureReadWithNative is optional: ordinary capture.Read remains unchanged.
// The callback runs synchronously before stock conversion while its input is borrowed.
type captureReadWithNative interface {
	ReadWithNative([]byte, func() bool, func(NativeBlock)) (int, bool, error)
}

// readCaptured accepts ONE SoftwareCut decision. No second privacy read occurs.
// Plain ALSA already supplies native interleaved bytes; no tap/copy is needed.
// Panel's socket capture is stock-only and is not misrepresented as hardware REF.
func readCaptured(pcm capture, raw []byte, policy func() bool, visit func(NativeBlock)) (int, bool, error) {
	if paired, ok := pcm.(captureReadWithNative); ok {
		return paired.ReadWithNative(raw, policy, visit)
	}
	n, err := pcm.Read(raw)
	if err != nil {
		return n, false, err
	}
	// Decision is made AFTER the hardware returns, never before a blocking Read.
	cut := policy()
	if cut || visit != nil {
		if device, ok := pcm.(interface{ Config() alsa.Config }); ok && n > 0 {
			cfg := device.Config()
			format := asp.FormatS24_3
			switch cfg.Format {
			case alsa.FormatS16_LE:
				format = asp.FormatS16
			case alsa.FormatS32_LE:
				format = asp.FormatS32
			}
			mics := min(Mics, cfg.Channels)
			nb := nativeInterleaved(raw[:n], uint32(cfg.Rate), format, uint8(cfg.Bits), 0, mics, cfg.Channels-mics)
			if cut {
				zeroNative(nb)
			}
			if visit != nil {
				visit(nb)
			}
		} else if cut {
			clear(raw[:n])
		}
	}
	return n, cut, nil
}

// Descriptor returns only the already captured PCM geometry. Input bytes remain
// borrowed in NativeBlock.Buffers; no metadata or PCM conversion is performed.
func (n NativeBlock) Descriptor() asp.Descriptor {
	return n.Geometry
}
