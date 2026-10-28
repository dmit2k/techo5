package mic

import (
	"fmt"
	"math"
	"strconv"

	"github.com/HuskerMinion/techo5/echod/internal/asp"
)

const athenaFactoryMICCount = 4

// qualifiedAthenaFactoryMIC requires an independently confirmed Athena
// identity and the actual native four-MIC/two-REF capture description.
// The identity alone does not authenticate capsule geometry or vendor
// logical-channel mapping: that qualification belongs to the factory helper.
func qualifiedAthenaFactoryMIC(config asp.Config, platform *asp.PlatformInfo) bool {
	if platform == nil || platform.Validate() != nil ||
		platform.Codename != "athena" || platform.Detection != asp.PlatformConfirmed {
		return false
	}
	if config.Flags != asp.Interleaved|asp.Synchronized|asp.HardwareRef {
		return false
	}
	m := config.Mic
	if m.Channels != athenaFactoryMICCount || m.Frames != 1920 ||
		m.Rate != 96000 || m.Format != asp.S32 || m.ValidBits != 24 ||
		!((m.FullBits == 23 && m.Alignment == asp.Low) ||
			(m.FullBits == 31 && m.Alignment == asp.High)) {
		return false
	}
	if config.Ref != (asp.Stream{
		Channels: 2, Frames: 1920, Rate: 96000, Format: asp.S16,
		FullBits: 15, Alignment: asp.Low, ValidBits: 16,
	}) {
		return false
	}
	_, _, err := config.Lengths()
	return err == nil
}

// factoryIDMEQ30 parses the factory's decimal signed-int32 representation.
// Only ASCII decimal digits are accepted: no whitespace, signs, separators,
// or nondecimal Unicode. The legacy left shift is checked BEFORE shifting;
// the signed-i64 wire container does not relax the signed-int32 gain limit.
// This function does not read IDME and never substitutes a unity coefficient.
func factoryIDMEQ30(raw string) (int64, bool) {
	if raw == "" {
		return 0, false
	}
	for i := 0; i < len(raw); i++ {
		if raw[i] < '0' || raw[i] > '9' {
			return 0, false
		}
	}
	n, err := strconv.ParseInt(raw, 10, 32)
	if err != nil || n <= 0 {
		return 0, false
	}
	if n&0x7fff0000 == 0 {
		if n > math.MaxInt32>>16 {
			return 0, false
		}
		n <<= 16
	}
	if n <= 0 || n > math.MaxInt32 {
		return 0, false
	}
	return n, true
}

// micFactoryGainFromValues is pure: values are explicitly supplied in the
// captured MIC PCM slot order. No vendor logical map or PCM scaling occurs.
// A nil or incorrectly sized source, unconfirmed identity, or mismatched
// native shape is NOT ADVERTISED (nil). A qualified four-field snapshot
// with zero valid coefficients advertises explicit NONE. Missing/malformed
// individual fields contribute no bit and their coefficient remains zero.
// The result owns a fixed array; it never retains the input slice.
func micFactoryGainFromValues(values []string, config asp.Config, platform *asp.PlatformInfo) *asp.MicFactoryGain {
	if len(values) != athenaFactoryMICCount || !qualifiedAthenaFactoryMIC(config, platform) {
		return nil
	}
	g := &asp.MicFactoryGain{Channels: config.Mic.Channels}
	for slot, raw := range values {
		if q30, valid := factoryIDMEQ30(raw); valid {
			g.PresentMask |= uint32(1) << uint(slot)
			g.Coefficients[slot] = q30
		}
	}
	switch g.PresentMask {
	case 0:
		g.Availability = asp.FactoryGainNone
		g.Encoding = asp.FactoryGainEncodingNone
		g.Provenance = asp.FactoryGainProvenanceNone
	case (1 << athenaFactoryMICCount) - 1:
		g.Availability = asp.FactoryGainComplete
		g.Encoding = asp.FactoryGainEncodingQ30
		g.Provenance = asp.FactoryGainProvenanceIDME
	default:
		g.Availability = asp.FactoryGainPartial
		g.Encoding = asp.FactoryGainEncodingQ30
		g.Provenance = asp.FactoryGainProvenanceIDME
	}
	// The CFG2 codec remains the normative validation authority; no
	// pointer is emitted if a future schema rejects this representation.
	if _, err := (asp.CFG2{Config: config, MicFactoryGain: g}).MarshalBinary(); err != nil {
		return nil
	}
	return g
}

// micFactoryGainFromReader is a one-shot session boundary, never a tick
// callback. It reads each qualified IDME name exactly once. board is the
// independently selected layout.Board; read is injected for offline tests.
// Missing IDME is represented by an empty string, not fabricated factory gain.
func micFactoryGainFromReader(config asp.Config, platform *asp.PlatformInfo, board string, read func(string) string) *asp.MicFactoryGain {
	if board != "athena" || read == nil || !qualifiedAthenaFactoryMIC(config, platform) {
		return nil
	}
	values := make([]string, athenaFactoryMICCount)
	for slot := range values {
		values[slot] = read(fmt.Sprintf("miccal.%d", slot))
	}
	return micFactoryGainFromValues(values, config, platform)
}
