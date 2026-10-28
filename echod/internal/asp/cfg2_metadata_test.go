package asp

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"reflect"
	"testing"
)

// All factory coefficients below are invented protocol-test inputs, not
// device IDME values, factory measurements, or a claim about real hardware.
func cfg2SyntheticComplete() MicFactoryGain {
	return MicFactoryGain{
		Encoding: FactoryGainEncodingQ30, Availability: FactoryGainComplete,
		Provenance: FactoryGainProvenanceIDME, Channels: 4, PresentMask: 0xf,
		Coefficients: [16]int64{0x42000000, 0x3c000000, 0x48000000, 0x30000000},
	}
}

func cfg2SyntheticPartial() MicFactoryGain {
	return MicFactoryGain{
		Encoding: FactoryGainEncodingQ30, Availability: FactoryGainPartial,
		Provenance: FactoryGainProvenanceIDME, Channels: 4, PresentMask: 0x5,
		Coefficients: [16]int64{0x38000000, 0, 0x44000000, 0},
	}
}

func cfg2SyntheticNone(ch uint16) MicFactoryGain {
	return MicFactoryGain{Channels: ch}
}

// Deliberately independent of MarshalBinary, this is just the fixed schema
// prefix and raw little-endian fields used by negative wire fixtures.
func cfg2RawGain(ch uint16, encoding, availability, provenance uint8, mask uint32, coeffs ...int64) []byte {
	b := make([]byte, 16+8*len(coeffs))
	binary.LittleEndian.PutUint16(b[0:2], 1)
	b[2], b[3], b[4] = encoding, availability, provenance
	binary.LittleEndian.PutUint16(b[6:8], ch)
	binary.LittleEndian.PutUint32(b[8:12], mask)
	for i, v := range coeffs {
		binary.LittleEndian.PutUint64(b[16+i*8:24+i*8], uint64(v))
	}
	return b
}

func cfg2MetaTLV(t *testing.T, wire []byte, typ uint16) (int, []byte) {
	t.Helper()
	for off := CFG2HeaderBytes; off+8 <= len(wire); {
		id := binary.LittleEndian.Uint16(wire[off:])
		n := binary.LittleEndian.Uint32(wire[off+4:])
		if uint64(n) > uint64(len(wire)-off-8) {
			t.Fatalf("bad independent fixture TLV at %d", off)
		}
		if id == typ {
			return off, wire[off+8 : off+8+int(n)]
		}
		off += 8 + int(n)
	}
	t.Fatalf("missing TLV type 0x%04x", typ)
	return 0, nil
}

func TestCFG2MetadataGoldenVectors(t *testing.T) {
	confirmed := PlatformInfo{Codename: "athena", Detection: PlatformConfirmed}
	inferred := PlatformInfo{Codename: "lab_future_7", Detection: PlatformInferred}
	other := PlatformInfo{Codename: "lab_board-3", Detection: PlatformInferred}
	complete, partial, none := cfg2SyntheticComplete(), cfg2SyntheticPartial(), cfg2SyntheticNone(1)
	for _, tc := range []struct {
		name string
		want CFG2
	}{
		{"athena_confirmed_complete_4mic.cfg2", CFG2{Config: cfg2Athena(), PlatformInfo: &confirmed, MicFactoryGain: &complete}},
		{"unknown_inferred_partial_4mic.cfg2", CFG2{Config: cfg2Athena(), PlatformInfo: &inferred, MicFactoryGain: &partial}},
		{"explicit_none_1mic.cfg2", CFG2{Config: cfg2Generic(), MicFactoryGain: &none}},
		{"platform_only_1mic.cfg2", CFG2{Config: cfg2Generic(), PlatformInfo: &other}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fixture := cfg2Fixture(t, tc.name)
			if string(fixture[:4]) != "CFG2" ||
				binary.LittleEndian.Uint16(fixture[4:6]) != 2 ||
				binary.LittleEndian.Uint16(fixture[6:8]) != 16 ||
				binary.LittleEndian.Uint32(fixture[8:12]) != uint32(len(fixture)) ||
				binary.LittleEndian.Uint32(fixture[12:16]) != tc.want.Config.Flags {
				t.Fatal("golden header or length mismatched")
			}
			for _, entry := range []struct {
				off, typ int
			}{{16, 1}, {40, 2}} {
				if binary.LittleEndian.Uint16(fixture[entry.off:]) != uint16(entry.typ) ||
					binary.LittleEndian.Uint16(fixture[entry.off+2:]) != 1 ||
					binary.LittleEndian.Uint32(fixture[entry.off+4:]) != 16 {
					t.Fatalf("core golden TLV mismatch at offset %d", entry.off)
				}
			}
			if tc.want.PlatformInfo != nil {
				off, raw := cfg2MetaTLV(t, fixture, cfg2PlatformInfo)
				p := tc.want.PlatformInfo
				if off != 64 || binary.LittleEndian.Uint16(fixture[off+2:]) != 0 ||
					binary.LittleEndian.Uint32(fixture[off+4:]) != uint32(4+len(p.Codename)) ||
					len(raw) != 4+len(p.Codename) ||
					raw[0] != 1 || raw[1] != byte(p.Detection) ||
					raw[2] != byte(len(p.Codename)) || raw[3] != 0 ||
					string(raw[4:]) != p.Codename {
					t.Fatalf("PLATFORM_INFO golden offsets or fields: %x", raw)
				}
			}
			if tc.want.MicFactoryGain != nil {
				off, raw := cfg2MetaTLV(t, fixture, cfg2MicFactoryGain)
				wantOff := 64
				if tc.want.PlatformInfo != nil {
					wantOff += 8 + 4 + len(tc.want.PlatformInfo.Codename)
				}
				g := tc.want.MicFactoryGain
				wantBytes := 16
				if g.Availability != FactoryGainNone {
					wantBytes += 8 * int(g.Channels)
				}
				if off != wantOff || binary.LittleEndian.Uint16(fixture[off+2:]) != 0 ||
					binary.LittleEndian.Uint32(fixture[off+4:]) != uint32(wantBytes) || len(raw) != wantBytes ||
					binary.LittleEndian.Uint16(raw[0:2]) != 1 ||
					raw[2] != byte(g.Encoding) || raw[3] != byte(g.Availability) ||
					raw[4] != byte(g.Provenance) || raw[5] != 0 ||
					binary.LittleEndian.Uint16(raw[6:8]) != g.Channels ||
					binary.LittleEndian.Uint32(raw[8:12]) != g.PresentMask ||
					binary.LittleEndian.Uint32(raw[12:16]) != 0 {
					t.Fatalf("MIC_FACTORY_GAIN golden offsets or prefix: %x", raw)
				}
				if g.Availability != FactoryGainNone {
					for i := 0; i < int(g.Channels); i++ {
						if int64(binary.LittleEndian.Uint64(raw[16+i*8:24+i*8])) != g.Coefficients[i] {
							t.Fatalf("captured slot %d mismatched in fixture", i)
						}
					}
				}
			}
			encoded, err := tc.want.MarshalBinary()
			if err != nil || !bytes.Equal(encoded, fixture) {
				t.Fatalf("encoder differed from independent Python struct.pack fixture: %x err=%v", encoded, err)
			}
			got, err := ParseCFG2(fixture)
			if err != nil || !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("typed decode: %+v err=%v want=%+v", got, err, tc.want)
			}
			again, err := got.MarshalBinary()
			if err != nil || !bytes.Equal(again, fixture) {
				t.Fatalf("roundtrip differs: %x err=%v", again, err)
			}
			again2, err := got.MarshalBinary()
			if err != nil || !bytes.Equal(again2, again) {
				t.Fatal("repeated encode nondeterministic", err)
			}
			var output bytes.Buffer
			if err := WriteCFG2(&output, tc.want); err != nil || !bytes.Equal(output.Bytes(), fixture) {
				t.Fatal("WriteCFG2 payload mismatch", err)
			}
			reader := bytes.NewReader(append(bytes.Clone(fixture), []byte("INP2")...))
			read, err := ReadCFG2(reader)
			if err != nil || !reflect.DeepEqual(read, tc.want) || reader.Len() != 4 {
				t.Fatalf("ReadCFG2 typed: %+v err=%v unread=%d", read, err, reader.Len())
			}
			if _, err := ParseCFG2(append(bytes.Clone(fixture), 0)); !errors.Is(err, ErrConfig) {
				t.Fatal("trailing packet bytes accepted", err)
			}
		})
	}
}

func TestCFG2PlatformInvalidWireAndEncoder(t *testing.T) {
	base := cfg2Fixture(t, "generic_1mic_noref.cfg2")
	valid := []byte{1, 1, 6, 0, 'a', 't', 'h', 'e', 'n', 'a'}
	invalid := []struct {
		name string
		raw  []byte
	}{
		{"empty", []byte{1, 1, 0, 0}},
		{"truncated_prefix", []byte{1, 1, 4}},
		{"bad_schema", []byte{2, 1, 6, 0, 'a', 't', 'h', 'e', 'n', 'a'}},
		{"zero_schema", []byte{0, 1, 6, 0, 'a', 't', 'h', 'e', 'n', 'a'}},
		{"bad_detection", []byte{1, 3, 6, 0, 'a', 't', 'h', 'e', 'n', 'a'}},
		{"zero_detection", []byte{1, 0, 6, 0, 'a', 't', 'h', 'e', 'n', 'a'}},
		{"bad_reserved", []byte{1, 1, 6, 1, 'a', 't', 'h', 'e', 'n', 'a'}},
		{"short_claim", []byte{1, 1, 5, 0, 'a', 't', 'h', 'e', 'n', 'a'}},
		{"long_claim", []byte{1, 1, 7, 0, 'a', 't', 'h', 'e', 'n', 'a'}},
		{"over_32_bytes", append([]byte{1, 1, 33, 0}, bytes.Repeat([]byte{'z'}, 33)...)},
		{"over_32_claim", []byte{1, 1, 33, 0, 'a'}},
		{"trailing_zero", append(bytes.Clone(valid), 0)},
		{"uppercase", []byte{1, 1, 1, 0, 'A'}},
		{"slash", []byte{1, 1, 1, 0, '/'}},
		{"space", []byte{1, 1, 1, 0, ' '}},
		{"nul", []byte{1, 1, 1, 0, 0}},
		{"utf8", []byte{1, 1, 2, 0, 0xc3, 0xa9}},
		{"invalid_ascii", []byte{1, 1, 1, 0, 0xff}},
	}
	for _, tc := range invalid {
		t.Run(tc.name, func(t *testing.T) {
			cfg2MustReject(t, cfg2InsertTLV(base, len(base), cfg2PlatformInfo, 0, tc.raw))
		})
	}
	for _, tc := range []struct {
		name string
		p    PlatformInfo
	}{
		{"empty", PlatformInfo{Detection: PlatformConfirmed}},
		{"too_long", PlatformInfo{Codename: string(bytes.Repeat([]byte{'x'}, 33)), Detection: PlatformConfirmed}},
		{"bad_detection", PlatformInfo{Codename: "athena", Detection: 3}},
		{"zero_detection", PlatformInfo{Codename: "athena", Detection: 0}},
		{"upper", PlatformInfo{Codename: "Athena", Detection: PlatformConfirmed}},
		{"nonascii", PlatformInfo{Codename: "roók", Detection: PlatformInferred}},
		{"nul", PlatformInfo{Codename: "foo\x00", Detection: PlatformInferred}},
	} {
		t.Run("marshal_"+tc.name, func(t *testing.T) {
			_, err := (CFG2{Config: cfg2Generic(), PlatformInfo: &tc.p}).MarshalBinary()
			if !errors.Is(err, ErrConfig) {
				t.Fatal("invalid platform serialized", err)
			}
		})
	}
	// Exactly 32 characters and all allowed alphabet classes are accepted.
	p := PlatformInfo{Codename: "abc012_-xyz0123456789_abcdef0123", Detection: PlatformInferred}
	if len(p.Codename) != 32 {
		t.Fatal("bad 32-byte test input", len(p.Codename))
	}
	if _, err := (CFG2{Config: cfg2Generic(), PlatformInfo: &p}).MarshalBinary(); err != nil {
		t.Fatal("rejected legal maximum-length codename", err)
	}
}

func TestCFG2FactoryGainInvalidWire(t *testing.T) {
	base := cfg2Fixture(t, "athena_4mic_2ref.cfg2")
	complete := []int64{0x42000000, 0x3c000000, 0x48000000, 0x30000000}
	good := func() []byte { return cfg2RawGain(4, 1, 1, 1, 0xf, complete...) }
	mutate := func(fn func([]byte)) []byte {
		b := good()
		fn(b)
		return b
	}
	for _, tc := range []struct {
		name string
		raw  []byte
	}{
		{"missing_header", make([]byte, 15)},
		{"extra_byte", append(good(), 0)},
		{"truncated_coeff", good()[:len(good())-1]},
		{"truncated_slot", good()[:len(good())-8]},
		{"schema", mutate(func(b []byte) { binary.LittleEndian.PutUint16(b[:2], 2) })},
		{"encoding", mutate(func(b []byte) { b[2] = 0 })},
		{"encoding_unknown", mutate(func(b []byte) { b[2] = 2 })},
		{"availability_unknown", mutate(func(b []byte) { b[3] = 3 })},
		{"provenance_none", mutate(func(b []byte) { b[4] = 0 })},
		{"provenance_unknown", mutate(func(b []byte) { b[4] = 2 })},
		{"applied_to_pcm", mutate(func(b []byte) { b[5] = 1 })},
		{"channels_3", mutate(func(b []byte) { binary.LittleEndian.PutUint16(b[6:8], 3) })},
		{"channels_0", mutate(func(b []byte) { binary.LittleEndian.PutUint16(b[6:8], 0) })},
		{"channels_17", mutate(func(b []byte) { binary.LittleEndian.PutUint16(b[6:8], 17) })},
		{"mask_above_channels", mutate(func(b []byte) { binary.LittleEndian.PutUint32(b[8:12], 0x1f) })},
		{"complete_missing_one", mutate(func(b []byte) { binary.LittleEndian.PutUint32(b[8:12], 0x7) })},
		{"partial_all", mutate(func(b []byte) { b[3] = 2 })},
		{"partial_empty", cfg2RawGain(4, 1, 2, 1, 0, 0, 0, 0, 0)},
		{"partial_zero_entry", cfg2RawGain(4, 1, 2, 1, 0x5, 0, 0, 0x44000000, 0)},
		{"partial_masked_nonzero", cfg2RawGain(4, 1, 2, 1, 0x5, 0x42000000, 10, 0x48000000, 0)},
		{"zero_present", mutate(func(b []byte) { binary.LittleEndian.PutUint64(b[16:24], 0) })},
		{"negative", mutate(func(b []byte) { binary.LittleEndian.PutUint64(b[16:24], ^uint64(0)) })},
		{"int32_overflow", mutate(func(b []byte) { binary.LittleEndian.PutUint64(b[16:24], uint64(math.MaxInt32)+1) })},
		{"int64_max", mutate(func(b []byte) { binary.LittleEndian.PutUint64(b[16:24], math.MaxInt64) })},
		{"int64_min", mutate(func(b []byte) { binary.LittleEndian.PutUint64(b[16:24], 1<<63) })},
		{"reserved", mutate(func(b []byte) { b[12] = 1 })},
		{"none_wrong_encoding", cfg2RawGain(4, 1, 0, 0, 0)},
		{"none_factory_provenance", cfg2RawGain(4, 0, 0, 1, 0)},
		{"none_present_mask", cfg2RawGain(4, 0, 0, 0, 1)},
		{"none_with_coeffs", cfg2RawGain(4, 0, 0, 0, 0, 1, 0, 0, 0)},
		{"none_wrong_count", cfg2RawGain(1, 0, 0, 0, 0)},
		{"partial_no_coeffs", cfg2RawGain(4, 1, 2, 1, 1)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg2MustReject(t, cfg2InsertTLV(base, len(base), cfg2MicFactoryGain, 0, tc.raw))
		})
	}
	// Forged uint32 length must be rejected before any size-based allocation.
	raw := cfg2InsertTLV(base, len(base), cfg2MicFactoryGain, 0, good())
	binary.LittleEndian.PutUint32(raw[len(base)+4:], ^uint32(0))
	cfg2MustReject(t, raw)
}

func TestCFG2FactoryGainInvalidEncoder(t *testing.T) {
	for _, tc := range []struct {
		name string
		edit func(*MicFactoryGain)
	}{
		{"wrong_channels", func(g *MicFactoryGain) { g.Channels = 3 }},
		{"applied_to_pcm", func(g *MicFactoryGain) { g.AppliedToPCM = true }},
		{"encoding", func(g *MicFactoryGain) { g.Encoding = FactoryGainEncodingNone }},
		{"provenance", func(g *MicFactoryGain) { g.Provenance = FactoryGainProvenanceNone }},
		{"mask_outside", func(g *MicFactoryGain) { g.PresentMask = 0x1f }},
		{"mask_incomplete", func(g *MicFactoryGain) { g.PresentMask = 7 }},
		{"partial_full", func(g *MicFactoryGain) { g.Availability = FactoryGainPartial }},
		{"partial_zero_mask", func(g *MicFactoryGain) {
			g.Availability = FactoryGainPartial
			g.PresentMask = 0
			g.Coefficients = [16]int64{}
		}},
		{"missing_value", func(g *MicFactoryGain) { g.Coefficients[2] = 0 }},
		{"negative", func(g *MicFactoryGain) { g.Coefficients[2] = -5 }},
		{"overflow", func(g *MicFactoryGain) { g.Coefficients[2] = math.MaxInt32 + 1 }},
		{"unused_slot", func(g *MicFactoryGain) { g.Coefficients[5] = 3 }},
		{"unknown_availability", func(g *MicFactoryGain) { g.Availability = 3 }},
		{"none_with_data", func(g *MicFactoryGain) { *g = cfg2SyntheticNone(4); g.Coefficients[0] = 1 }},
		{"none_with_mask", func(g *MicFactoryGain) { *g = cfg2SyntheticNone(4); g.PresentMask = 1 }},
		{"none_with_encoding", func(g *MicFactoryGain) { *g = cfg2SyntheticNone(4); g.Encoding = FactoryGainEncodingQ30 }},
		{"none_with_provenance", func(g *MicFactoryGain) { *g = cfg2SyntheticNone(4); g.Provenance = FactoryGainProvenanceIDME }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := cfg2SyntheticComplete()
			tc.edit(&g)
			if _, err := (CFG2{Config: cfg2Athena(), MicFactoryGain: &g}).MarshalBinary(); !errors.Is(err, ErrConfig) {
				t.Fatal("invalid factory gain serialized", err)
			}
		})
	}
}

func TestCFG2MetadataOrderingAndLimits(t *testing.T) {
	base := cfg2Fixture(t, "athena_4mic_2ref.cfg2")
	complete := cfg2Fixture(t, "athena_confirmed_complete_4mic.cfg2")
	platformTLV, _ := cfg2MetaTLV(t, complete, cfg2PlatformInfo)
	gainTLV, _ := cfg2MetaTLV(t, complete, cfg2MicFactoryGain)
	p := []byte{1, 1, 6, 0, 'a', 't', 'h', 'e', 'n', 'a'}
	g := cfg2RawGain(4, 1, 1, 1, 0xf, 0x42000000, 0x3c000000, 0x48000000, 0x30000000)
	for _, tc := range []struct {
		name string
		wire []byte
	}{
		{"platform_required", cfg2InsertTLV(base, len(base), cfg2PlatformInfo, cfg2Required, p)},
		{"gain_required", cfg2InsertTLV(base, len(base), cfg2MicFactoryGain, cfg2Required, g)},
		{"platform_bad_flags", cfg2InsertTLV(base, len(base), cfg2PlatformInfo, 2, p)},
		{"gain_bad_flags", cfg2InsertTLV(base, len(base), cfg2MicFactoryGain, 2, g)},
		{"duplicate_platform", cfg2InsertTLV(complete, gainTLV, cfg2PlatformInfo, 0, p)},
		{"duplicate_gain", cfg2InsertTLV(complete, len(complete), cfg2MicFactoryGain, 0, g)},
		{"gain_then_platform", cfg2InsertTLV(complete, len(complete), cfg2PlatformInfo, 0, p)},
		{"platform_before_ref", cfg2InsertTLV(base, 40, cfg2PlatformInfo, 0, p)},
		{"gain_before_platform", cfg2InsertTLV(complete, platformTLV, cfg2MicFactoryGain, 0, g)},
		{"unknown_optional_before_gain_wrong_order", cfg2InsertTLV(complete, gainTLV, 0x7f01, 0, nil)},
	} {
		t.Run(tc.name, func(t *testing.T) { cfg2MustReject(t, tc.wire) })
	}
	// The maximum 8192 bytes remains valid with both known optional TLVs.
	max := cfg2InsertTLV(complete, len(complete), 0x7f01, 0, make([]byte, CFG2MaxBytes-len(complete)-8))
	if len(max) != CFG2MaxBytes {
		t.Fatal("bad maximum packet size", len(max))
	}
	if got, err := ParseCFG2(max); err != nil || got.PlatformInfo == nil || got.MicFactoryGain == nil {
		t.Fatal("8192 bytes with metadata should pass", err)
	}
	over := cfg2InsertTLV(complete, len(complete), 0x7f01, 0, make([]byte, CFG2MaxBytes-len(complete)-7))
	cfg2MustReject(t, over)
	// 4 metadata+core tags and 60 unassigned optional tags = 64.
	for extras := 60; extras <= 61; extras++ {
		b := bytes.Clone(complete)
		for i := 0; i < extras; i++ {
			b = cfg2InsertTLV(b, len(b), uint16(0x2000+i), 0, nil)
		}
		if extras == 60 {
			if _, err := ParseCFG2(b); err != nil {
				t.Fatal("64 TLVs must be legal", err)
			}
		} else {
			cfg2MustReject(t, b)
		}
	}
	// Unknown OPTIONAL may also appear *between* the two known metadata
	// tags; the parsed typed metadata must remain byte-identical.
	middle := cfg2InsertTLV(complete, gainTLV, 0x0020, 0, []byte{0xfd, 0x01})
	middleParsed, err := ParseCFG2(middle)
	if err != nil {
		t.Fatal("unknown OPTIONAL between known metadata tags", err)
	}
	middleCanonical, err := middleParsed.MarshalBinary()
	if err != nil || !bytes.Equal(middleCanonical, complete) {
		t.Fatal("middle OPTIONAL affected known metadata", err)
	}
	// A future OPTIONAL with higher type is ignored, not reflected in the
	// canonical roundtrip or in captured channel calibration.
	unknown := cfg2InsertTLV(complete, len(complete), 0x7f01, 0, []byte{0xf0, 0xaa})
	parsed, err := ParseCFG2(unknown)
	if err != nil {
		t.Fatal(err)
	}
	roundtrip, err := parsed.MarshalBinary()
	if err != nil || !bytes.Equal(roundtrip, complete) {
		t.Fatal("future OPTIONAL was retained or modified typed metadata", err)
	}
}

func TestCFG2MetadataChannelBoundaryAndOwnership(t *testing.T) {
	mic16 := cfg2Generic()
	mic16.Mic.Channels = 16
	mic16bytes, _, err := mic16.Lengths()
	if err != nil || mic16bytes != 16*320*2 {
		t.Fatalf("bad raw 16 MIC PCM length: %d %v", mic16bytes, err)
	}
	full := MicFactoryGain{
		Encoding: FactoryGainEncodingQ30, Availability: FactoryGainComplete,
		Provenance: FactoryGainProvenanceIDME, Channels: 16, PresentMask: 0xffff,
	}
	for i := range full.Coefficients {
		full.Coefficients[i] = int64(1+i) * (1 << 26)
	}
	partial := MicFactoryGain{
		Encoding: FactoryGainEncodingQ30, Availability: FactoryGainPartial,
		Provenance: FactoryGainProvenanceIDME, Channels: 16, PresentMask: 0x8001,
	}
	partial.Coefficients[0], partial.Coefficients[15] = 0x38000000, 0x44000000
	for _, gain := range []MicFactoryGain{full, partial} {
		wire, err := (CFG2{Config: mic16, MicFactoryGain: &gain}).MarshalBinary()
		if err != nil {
			t.Fatal("16 MIC encode", err)
		}
		got, err := ParseCFG2(wire)
		if err != nil || !reflect.DeepEqual(*got.MicFactoryGain, gain) {
			t.Fatal("16 MIC parse", err)
		}
	}
	outside := full
	outside.PresentMask |= 0x10000
	if _, err := (CFG2{Config: mic16, MicFactoryGain: &outside}).MarshalBinary(); !errors.Is(err, ErrConfig) {
		t.Fatal("17th MIC bit accepted by encoder", err)
	}
	bad16, err := (CFG2{Config: mic16, MicFactoryGain: &full}).MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	at, _ := cfg2MetaTLV(t, bad16, cfg2MicFactoryGain)
	binary.LittleEndian.PutUint32(bad16[at+16:at+20], 0x1ffff)
	cfg2MustReject(t, bad16)
	// Correct wire values survive mutation of the caller's input bytes.
	wire := cfg2Fixture(t, "unknown_inferred_partial_4mic.cfg2")
	got, err := ParseCFG2(wire)
	if err != nil {
		t.Fatal(err)
	}
	original, err := got.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	for i := range wire {
		wire[i] ^= 0xff
	}
	after, err := got.MarshalBinary()
	if err != nil || !bytes.Equal(original, after) {
		t.Fatal("decoded metadata aliases the source packet", err)
	}
	// Caller changes to the typed struct after a previous encode cannot
	// mutate already-returned on-wire bytes (nor can any byte slice alias).
	gain := cfg2SyntheticPartial()
	platform := PlatformInfo{Codename: "lab_future_7", Detection: PlatformInferred}
	cfg := CFG2{Config: cfg2Athena(), PlatformInfo: &platform, MicFactoryGain: &gain}
	first, err := cfg.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	gain.Coefficients[0] = 1
	platform.Codename = "modified"
	identical := cfg2Fixture(t, "unknown_inferred_partial_4mic.cfg2")
	if !bytes.Equal(first, identical) {
		t.Fatal("earlier encoded packet mutated by external struct")
	}
	// A wrong MIC channel count is never silently reconciled by metadata.
	gain.Channels = 3
	if _, err := cfg.MarshalBinary(); !errors.Is(err, ErrConfig) {
		t.Fatal("metadata count mismatch silently repaired", err)
	}
	// Absence and explicit NONE are different serialized states.
	absent := cfg2Fixture(t, "generic_1mic_noref.cfg2")
	explicit := cfg2Fixture(t, "explicit_none_1mic.cfg2")
	if bytes.Equal(absent, explicit) {
		t.Fatal("NONE indistinguishable from absent")
	}
	none, err := ParseCFG2(explicit)
	if err != nil || none.MicFactoryGain == nil || none.MicFactoryGain.Availability != FactoryGainNone {
		t.Fatal("explicit NONE lost", err)
	}
	notAdvertised, err := ParseCFG2(absent)
	if err != nil || notAdvertised.MicFactoryGain != nil || notAdvertised.PlatformInfo != nil {
		t.Fatal("metadata absence not preserved", err)
	}

}

func TestCFG2MetadataTruncationAndPacketIntegrity(t *testing.T) {
	wire := cfg2Fixture(t, "athena_confirmed_complete_4mic.cfg2")
	for n := 0; n < len(wire); n++ {
		t.Run(fmt.Sprintf("truncated_%03d", n), func(t *testing.T) {
			cfg2MustReject(t, wire[:n])
		})
	}
	if err := WriteCFG2(io.Discard, CFG2{Config: cfg2Athena(), MicFactoryGain: &MicFactoryGain{}}); !errors.Is(err, ErrConfig) {
		t.Fatal("WriteCFG2 accepted invalid gain", err)
	}
}
