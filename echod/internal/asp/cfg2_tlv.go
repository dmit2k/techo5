package asp

import (
	"encoding/binary"
	"fmt"
	"io"
)

const (
	CFG2HeaderBytes = 16
	CFG2MaxBytes    = 8192
	CFG2MaxTLVs     = 64

	cfg2Version        = 2
	cfg2TLVHeaderBytes = 8
	cfg2StreamBytes    = 16
	cfg2Required       = 1
	cfg2MicStream      = 0x0001
	cfg2RefStream      = 0x0002
)

// CFG2 is the extensible, pre-release CFG2/TLV session descriptor.
//
// Optional metadata fields distinguish unadvertised (nil) from explicit values,
// including factory gain NONE. Unknown optional wire values are never retained.
type CFG2 struct {
	Config         Config
	PlatformInfo   *PlatformInfo
	MicFactoryGain *MicFactoryGain
}

// MarshalBinary serializes canonical ordered CFG2 descriptors. It never
// fabricates platform identity, factory calibration or playback reference.
func (c CFG2) MarshalBinary() ([]byte, error) {
	if _, _, err := c.Config.Lengths(); err != nil {
		return nil, fmt.Errorf("%w: invalid PCM stream shape", ErrConfig)
	}

	// All typed metadata passes through the same bounded sorted TLV writer.
	e := cfg2Encoder{wire: make([]byte, CFG2HeaderBytes, 64)}
	copy(e.wire[:4], "CFG2")
	binary.LittleEndian.PutUint16(e.wire[4:6], cfg2Version)
	binary.LittleEndian.PutUint16(e.wire[6:8], CFG2HeaderBytes)
	binary.LittleEndian.PutUint32(e.wire[12:16], c.Config.Flags)
	if err := e.addStream(cfg2MicStream, c.Config.Mic); err != nil {
		return nil, err
	}
	if err := e.addStream(cfg2RefStream, c.Config.Ref); err != nil {
		return nil, err
	}
	if c.PlatformInfo != nil {
		value, err := c.PlatformInfo.marshalCFG2()
		if err != nil {
			return nil, err
		}
		if err := e.add(cfg2PlatformInfo, 0, value); err != nil {
			return nil, err
		}
	}
	if c.MicFactoryGain != nil {
		value, err := c.MicFactoryGain.marshalCFG2(c.Config.Mic.Channels)
		if err != nil {
			return nil, err
		}
		if err := e.add(cfg2MicFactoryGain, 0, value); err != nil {
			return nil, err
		}
	}
	binary.LittleEndian.PutUint32(e.wire[8:12], uint32(len(e.wire)))
	return e.wire, nil
}

// cfg2Encoder is private so unvalidated arbitrary optional vendor bytes
// cannot be accidentally advertised through the public configuration API.
type cfg2Encoder struct {
	wire     []byte
	previous uint16
	count    int
}

func (e *cfg2Encoder) add(typ, flags uint16, value []byte) error {
	if typ == 0 || typ == 0xffff || typ <= e.previous ||
		flags & ^uint16(cfg2Required) != 0 || e.count >= CFG2MaxTLVs ||
		len(e.wire) > CFG2MaxBytes-cfg2TLVHeaderBytes ||
		len(value) > CFG2MaxBytes-len(e.wire)-cfg2TLVHeaderBytes {
		return fmt.Errorf("%w: invalid or oversized CFG2 TLV", ErrConfig)
	}
	var hdr [cfg2TLVHeaderBytes]byte
	binary.LittleEndian.PutUint16(hdr[:2], typ)
	binary.LittleEndian.PutUint16(hdr[2:4], flags)
	binary.LittleEndian.PutUint32(hdr[4:8], uint32(len(value)))
	e.wire = append(e.wire, hdr[:]...)
	e.wire = append(e.wire, value...)
	e.previous, e.count = typ, e.count+1
	return nil
}

func (e *cfg2Encoder) addStream(typ uint16, stream Stream) error {
	var b [cfg2StreamBytes]byte
	encodeCFG2Stream(b[:], stream)
	return e.add(typ, cfg2Required, b[:])
}

// WriteCFG2 writes one packet, including handling partial io.Writer writes.
func WriteCFG2(w io.Writer, c CFG2) error {
	wire, err := c.MarshalBinary()
	if err != nil {
		return err
	}
	return WriteAll(w, wire)
}

// ReadCFG2 reads exactly one CFG2 message, leaving subsequent protocol bytes
// unread. The fixed header is checked before any variable-sized allocation.
func ReadCFG2(r io.Reader) (CFG2, error) {
	var header [CFG2HeaderBytes]byte
	if _, err := io.ReadFull(r, header[:]); err != nil {
		return CFG2{}, fmt.Errorf("%w: incomplete CFG2 header: %v", ErrConfig, err)
	}
	total, _, err := validateCFG2Header(header[:])
	if err != nil {
		return CFG2{}, err
	}
	wire := make([]byte, total)
	copy(wire, header[:])
	if _, err := io.ReadFull(r, wire[CFG2HeaderBytes:]); err != nil {
		return CFG2{}, fmt.Errorf("%w: incomplete CFG2 body: %v", ErrConfig, err)
	}
	return ParseCFG2(wire)
}

// ParseCFG2 validates a complete standalone packet. Unlike ReadCFG2, it
// rejects extra bytes beyond total_bytes: it never silently discards garbage.
func ParseCFG2(wire []byte) (CFG2, error) {
	if len(wire) < CFG2HeaderBytes {
		return CFG2{}, fmt.Errorf("%w: short CFG2 header", ErrConfig)
	}
	total, flags, err := validateCFG2Header(wire[:CFG2HeaderBytes])
	if err != nil {
		return CFG2{}, err
	}
	if len(wire) != total {
		return CFG2{}, fmt.Errorf("%w: CFG2 packet size mismatch", ErrConfig)
	}

	cfg := CFG2{Config: Config{Flags: flags}}
	var haveMic, haveRef bool
	offset, count := CFG2HeaderBytes, 0
	var previous uint16
	for offset < total {
		if total-offset < cfg2TLVHeaderBytes {
			return CFG2{}, fmt.Errorf("%w: incomplete TLV header", ErrConfig)
		}
		tlvType := binary.LittleEndian.Uint16(wire[offset : offset+2])
		tlvFlags := binary.LittleEndian.Uint16(wire[offset+2 : offset+4])
		size := binary.LittleEndian.Uint32(wire[offset+4 : offset+8])
		offset += cfg2TLVHeaderBytes
		count++
		if count > CFG2MaxTLVs || tlvType == 0 || tlvType == 0xffff ||
			tlvType <= previous || tlvFlags & ^uint16(cfg2Required) != 0 {
			return CFG2{}, fmt.Errorf("%w: invalid TLV type, order or flags", ErrConfig)
		}
		previous = tlvType
		// Compare as uint64 before converting or slicing: a forged u32
		// value_bytes must never overflow, allocate or go out of bounds.
		if uint64(size) > uint64(total-offset) {
			return CFG2{}, fmt.Errorf("%w: TLV exceeds CFG2 boundary", ErrConfig)
		}
		value := wire[offset : offset+int(size)]
		offset += int(size)

		switch tlvType {
		case cfg2MicStream:
			if tlvFlags != cfg2Required || len(value) != cfg2StreamBytes {
				return CFG2{}, fmt.Errorf("%w: invalid MIC_STREAM", ErrConfig)
			}
			cfg.Config.Mic, err = decodeCFG2Stream(value)
			haveMic = true
		case cfg2RefStream:
			if tlvFlags != cfg2Required || len(value) != cfg2StreamBytes {
				return CFG2{}, fmt.Errorf("%w: invalid REF_STREAM", ErrConfig)
			}
			cfg.Config.Ref, err = decodeCFG2Stream(value)
			haveRef = true
		case cfg2PlatformInfo:
			if tlvFlags != 0 {
				return CFG2{}, fmt.Errorf("%w: PLATFORM_INFO must be OPTIONAL", ErrConfig)
			}
			var platform PlatformInfo
			platform, err = parseCFG2PlatformInfo(value)
			if err == nil {
				cfg.PlatformInfo = &platform
			}
		case cfg2MicFactoryGain:
			if tlvFlags != 0 {
				return CFG2{}, fmt.Errorf("%w: MIC_FACTORY_GAIN must be OPTIONAL", ErrConfig)
			}
			var gain MicFactoryGain
			gain, err = parseCFG2MicFactoryGain(value, cfg.Config.Mic.Channels)
			if err == nil {
				cfg.MicFactoryGain = &gain
			}
		default:
			// Unknown optional TLVs are skipped without retaining wire bytes.
			if tlvFlags == cfg2Required {
				return CFG2{}, fmt.Errorf("%w: unsupported REQUIRED TLV", ErrConfig)
			}
		}
		if err != nil {
			return CFG2{}, err
		}
	}
	if !haveMic || !haveRef {
		return CFG2{}, fmt.Errorf("%w: missing required stream TLV", ErrConfig)
	}
	if _, _, err := cfg.Config.Lengths(); err != nil {
		return CFG2{}, fmt.Errorf("%w: invalid MIC/REF PCM shape", ErrConfig)
	}
	return cfg, nil
}

func validateCFG2Header(header []byte) (int, uint32, error) {
	if string(header[:4]) != "CFG2" ||
		binary.LittleEndian.Uint16(header[4:6]) != cfg2Version ||
		binary.LittleEndian.Uint16(header[6:8]) != CFG2HeaderBytes {
		return 0, 0, fmt.Errorf("%w: unsupported CFG2 header", ErrConfig)
	}
	size := binary.LittleEndian.Uint32(header[8:12])
	flags := binary.LittleEndian.Uint32(header[12:16])
	if size < CFG2HeaderBytes || size > CFG2MaxBytes ||
		(flags != Interleaved && flags != Interleaved|Synchronized|HardwareRef) {
		return 0, 0, fmt.Errorf("%w: CFG2 length or stream flags", ErrConfig)
	}
	return int(size), flags, nil
}

func encodeCFG2Stream(dst []byte, stream Stream) {
	// Caller owns a zero-initialized 16-byte descriptor, including reserved.
	binary.LittleEndian.PutUint16(dst[0:2], stream.Channels)
	binary.LittleEndian.PutUint16(dst[2:4], stream.Frames)
	binary.LittleEndian.PutUint32(dst[4:8], stream.Rate)
	dst[8], dst[9], dst[10], dst[11] =
		stream.Format, stream.FullBits, stream.Alignment, stream.ValidBits
}

func decodeCFG2Stream(value []byte) (Stream, error) {
	if binary.LittleEndian.Uint32(value[12:16]) != 0 {
		return Stream{}, fmt.Errorf("%w: nonzero stream reserved bytes", ErrConfig)
	}
	return Stream{
		Channels:  binary.LittleEndian.Uint16(value[0:2]),
		Frames:    binary.LittleEndian.Uint16(value[2:4]),
		Rate:      binary.LittleEndian.Uint32(value[4:8]),
		Format:    value[8],
		FullBits:  value[9],
		Alignment: value[10],
		ValidBits: value[11],
	}, nil
}
