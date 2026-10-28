package asp

import (
	"encoding/binary"
	"fmt"
	"math"
)

const (
	cfg2PlatformInfo   = 0x0003
	cfg2MicFactoryGain = 0x0101

	cfg2PlatformPrefixBytes = 4
	cfg2MaxCodenameBytes    = 32
	cfg2GainPrefixBytes     = 16
	cfg2MaxMICChannels      = 16
)

// PlatformDetection states the confidence in board identification, not
// permission to load factory assets or an assertion about PCM stream shape.
type PlatformDetection uint8

const (
	PlatformConfirmed PlatformDetection = 1
	PlatformInferred  PlatformDetection = 2
)

// PlatformInfo is advisory metadata. A nil *PlatformInfo in CFG2 means the
// TLV was not advertised; an invalid or empty instance is never serialized.
type PlatformInfo struct {
	Codename  string
	Detection PlatformDetection
}

// Validate uses the same rules as the CFG2 encoder, permitting callers to
// omit invalid optional platform metadata without duplicating wire checks.
func (p PlatformInfo) Validate() error { return p.validate() }

func (p PlatformInfo) validate() error {
	if p.Detection != PlatformConfirmed && p.Detection != PlatformInferred {
		return fmt.Errorf("%w: invalid platform detection", ErrConfig)
	}
	if len(p.Codename) < 1 || len(p.Codename) > cfg2MaxCodenameBytes {
		return fmt.Errorf("%w: platform codename length", ErrConfig)
	}
	for i := 0; i < len(p.Codename); i++ {
		c := p.Codename[i]
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '_' || c == '-') {
			return fmt.Errorf("%w: invalid platform codename ASCII", ErrConfig)
		}
	}
	return nil
}

func (p PlatformInfo) marshalCFG2() ([]byte, error) {
	if err := p.validate(); err != nil {
		return nil, err
	}
	value := make([]byte, cfg2PlatformPrefixBytes+len(p.Codename))
	value[0], value[1], value[2] = 1, byte(p.Detection), byte(len(p.Codename))
	copy(value[cfg2PlatformPrefixBytes:], p.Codename)
	return value, nil
}

func parseCFG2PlatformInfo(value []byte) (PlatformInfo, error) {
	if len(value) < cfg2PlatformPrefixBytes || value[0] != 1 ||
		value[3] != 0 || int(value[2]) != len(value)-cfg2PlatformPrefixBytes {
		return PlatformInfo{}, fmt.Errorf("%w: invalid PLATFORM_INFO envelope", ErrConfig)
	}
	// This conversion creates an owned string. No caller-owned packet backing
	// is retained in the returned typed descriptor.
	p := PlatformInfo{Codename: string(value[cfg2PlatformPrefixBytes:]), Detection: PlatformDetection(value[1])}
	if err := p.validate(); err != nil {
		return PlatformInfo{}, err
	}
	return p, nil
}

// FactoryGainEncoding describes the only two schema-v1 representations.
type FactoryGainEncoding uint8

const (
	FactoryGainEncodingNone FactoryGainEncoding = 0
	FactoryGainEncodingQ30  FactoryGainEncoding = 1
)

// FactoryGainAvailability distinguishes explicit NONE, COMPLETE, and PARTIAL.
// A nil *MicFactoryGain in CFG2 means NOT ADVERTISED (different from NONE).
type FactoryGainAvailability uint8

const (
	FactoryGainNone     FactoryGainAvailability = 0
	FactoryGainComplete FactoryGainAvailability = 1
	FactoryGainPartial  FactoryGainAvailability = 2
)

// FactoryGainProvenance explicitly distinguishes factory-IDME values from
// missing metadata. Fabricated unity must never claim factory provenance.
type FactoryGainProvenance uint8

const (
	FactoryGainProvenanceNone FactoryGainProvenance = 0
	FactoryGainProvenanceIDME FactoryGainProvenance = 1
)

// MicFactoryGain associates values with *captured* MIC slots, without
// applying gain to PCM or permuting channels for any DSP logical mapping.
// Coefficients is a fixed-size owned array, not a slice into packet bytes.
// Slots outside Channels and masked-out slots MUST be zero.
type MicFactoryGain struct {
	Encoding     FactoryGainEncoding
	Availability FactoryGainAvailability
	Provenance   FactoryGainProvenance
	AppliedToPCM bool
	Channels     uint16
	PresentMask  uint32
	Coefficients [cfg2MaxMICChannels]int64
}

func (g MicFactoryGain) validate(micChannels uint16) error {
	if g.Channels != micChannels || micChannels < 1 || micChannels > cfg2MaxMICChannels || g.AppliedToPCM {
		return fmt.Errorf("%w: factory gain channel association or PCM applied flag", ErrConfig)
	}
	all := uint32(1)<<micChannels - 1
	if g.PresentMask & ^all != 0 {
		return fmt.Errorf("%w: factory gain mask outside captured MIC slots", ErrConfig)
	}
	switch g.Availability {
	case FactoryGainNone:
		if g.Encoding != FactoryGainEncodingNone || g.Provenance != FactoryGainProvenanceNone || g.PresentMask != 0 {
			return fmt.Errorf("%w: invalid factory gain NONE representation", ErrConfig)
		}
	case FactoryGainComplete, FactoryGainPartial:
		if g.Encoding != FactoryGainEncodingQ30 || g.Provenance != FactoryGainProvenanceIDME {
			return fmt.Errorf("%w: invalid factory Q30 representation or provenance", ErrConfig)
		}
		if (g.Availability == FactoryGainComplete && g.PresentMask != all) ||
			(g.Availability == FactoryGainPartial && (g.PresentMask == 0 || g.PresentMask == all)) {
			return fmt.Errorf("%w: inconsistent factory gain availability/mask", ErrConfig)
		}
	default:
		return fmt.Errorf("%w: unknown factory gain availability", ErrConfig)
	}
	for i, v := range g.Coefficients {
		present := i < int(micChannels) && g.PresentMask&(uint32(1)<<uint(i)) != 0
		if present {
			if v <= 0 || v > math.MaxInt32 {
				return fmt.Errorf("%w: factory Q30 outside positive signed int32 range", ErrConfig)
			}
		} else if v != 0 {
			return fmt.Errorf("%w: absent factory gain slot has nonzero value", ErrConfig)
		}
	}
	return nil
}

func (g MicFactoryGain) marshalCFG2(micChannels uint16) ([]byte, error) {
	if err := g.validate(micChannels); err != nil {
		return nil, err
	}
	size := cfg2GainPrefixBytes
	if g.Availability != FactoryGainNone {
		size += 8 * int(micChannels)
	}
	value := make([]byte, size)
	binary.LittleEndian.PutUint16(value[0:2], 1)
	value[2], value[3], value[4] = byte(g.Encoding), byte(g.Availability), byte(g.Provenance)
	// value[5] is the mandatory zero applied_to_pcm field.
	binary.LittleEndian.PutUint16(value[6:8], g.Channels)
	binary.LittleEndian.PutUint32(value[8:12], g.PresentMask)
	// value[12:16] remains reserved zero.
	if g.Availability != FactoryGainNone {
		for i := 0; i < int(micChannels); i++ {
			binary.LittleEndian.PutUint64(value[16+8*i:24+8*i], uint64(g.Coefficients[i]))
		}
	}
	return value, nil
}

func parseCFG2MicFactoryGain(value []byte, micChannels uint16) (MicFactoryGain, error) {
	if len(value) < cfg2GainPrefixBytes || binary.LittleEndian.Uint16(value[0:2]) != 1 ||
		value[5] != 0 || binary.LittleEndian.Uint32(value[12:16]) != 0 {
		return MicFactoryGain{}, fmt.Errorf("%w: invalid MIC_FACTORY_GAIN envelope", ErrConfig)
	}
	g := MicFactoryGain{
		Encoding:     FactoryGainEncoding(value[2]),
		Availability: FactoryGainAvailability(value[3]),
		Provenance:   FactoryGainProvenance(value[4]),
		Channels:     binary.LittleEndian.Uint16(value[6:8]),
		PresentMask:  binary.LittleEndian.Uint32(value[8:12]),
	}
	// Check captured MIC count, availability, and exact length before
	// touching any coefficient slot. Nothing is allocated from wire sizes.
	if g.Channels != micChannels || micChannels < 1 || micChannels > cfg2MaxMICChannels {
		return MicFactoryGain{}, fmt.Errorf("%w: MIC_FACTORY_GAIN channel mismatch", ErrConfig)
	}
	want := cfg2GainPrefixBytes
	if g.Availability == FactoryGainComplete || g.Availability == FactoryGainPartial {
		want += 8 * int(micChannels)
	}
	if len(value) != want {
		return MicFactoryGain{}, fmt.Errorf("%w: MIC_FACTORY_GAIN payload size", ErrConfig)
	}
	if g.Availability != FactoryGainNone {
		for i := 0; i < int(micChannels); i++ {
			g.Coefficients[i] = int64(binary.LittleEndian.Uint64(value[16+8*i : 24+8*i]))
		}
	}
	if err := g.validate(micChannels); err != nil {
		return MicFactoryGain{}, err
	}
	return g, nil
}
