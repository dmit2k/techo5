package asp

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func cfg2Athena() Config {
	return Config{
		Flags: Interleaved | Synchronized | HardwareRef,
		Mic: Stream{
			Channels: 4, Frames: 1920, Rate: 96000,
			Format: S32, FullBits: 23, Alignment: Low, ValidBits: 24,
		},
		Ref: Stream{
			Channels: 2, Frames: 1920, Rate: 96000,
			Format: S16, FullBits: 15, Alignment: Low, ValidBits: 16,
		},
	}
}

func cfg2Generic() Config {
	return Config{
		Flags: Interleaved,
		Mic: Stream{
			Channels: 1, Frames: 320, Rate: 16000,
			Format: S16, FullBits: 15, Alignment: Low, ValidBits: 16,
		},
	}
}

func cfg2Fixture(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", "cfg2_v2", name))
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func cfg2PutSize(b []byte) {
	binary.LittleEndian.PutUint32(b[8:12], uint32(len(b)))
}

// cfg2InsertTLV intentionally constructs raw on-wire TLVs independently of
// MarshalBinary. The caller is responsible for choosing ordered types.
func cfg2InsertTLV(wire []byte, at int, typ, flags uint16, value []byte) []byte {
	tlv := make([]byte, cfg2TLVHeaderBytes+len(value))
	binary.LittleEndian.PutUint16(tlv[:2], typ)
	binary.LittleEndian.PutUint16(tlv[2:4], flags)
	binary.LittleEndian.PutUint32(tlv[4:8], uint32(len(value)))
	copy(tlv[8:], value)
	next := make([]byte, 0, len(wire)+len(tlv))
	next = append(next, wire[:at]...)
	next = append(next, tlv...)
	next = append(next, wire[at:]...)
	cfg2PutSize(next)
	return next
}

func cfg2MustReject(t *testing.T, wire []byte) {
	t.Helper()
	if cfg, err := ParseCFG2(wire); !errors.Is(err, ErrConfig) {
		t.Fatalf("ParseCFG2 accepted invalid config: cfg=%+v err=%v", cfg, err)
	}
	if cfg, err := ReadCFG2(bytes.NewReader(wire)); err == nil {
		t.Fatalf("ReadCFG2 accepted invalid config: %+v", cfg)
	}
}

func TestCFG2GoldenVectors(t *testing.T) {
	for _, tc := range []struct {
		file string
		cfg  Config
	}{
		{"athena_4mic_2ref.cfg2", cfg2Athena()},
		{"generic_1mic_noref.cfg2", cfg2Generic()},
	} {
		t.Run(tc.file, func(t *testing.T) {
			fixture := cfg2Fixture(t, tc.file)
			if len(fixture) != 64 || string(fixture[:4]) != "CFG2" ||
				binary.LittleEndian.Uint16(fixture[4:6]) != 2 ||
				binary.LittleEndian.Uint16(fixture[6:8]) != 16 ||
				binary.LittleEndian.Uint32(fixture[8:12]) != 64 ||
				binary.LittleEndian.Uint32(fixture[12:16]) != tc.cfg.Flags {
				t.Fatalf("independent golden header fields: %x", fixture)
			}
			for _, off := range []int{16, 40} {
				if binary.LittleEndian.Uint16(fixture[off:off+2]) != uint16(1+(off-16)/24) ||
					binary.LittleEndian.Uint16(fixture[off+2:off+4]) != 1 ||
					binary.LittleEndian.Uint32(fixture[off+4:off+8]) != 16 ||
					binary.LittleEndian.Uint32(fixture[off+20:off+24]) != 0 {
					t.Fatalf("invalid golden TLV fields at offset %d", off)
				}
			}
			if binary.LittleEndian.Uint16(fixture[24:26]) != tc.cfg.Mic.Channels ||
				binary.LittleEndian.Uint16(fixture[26:28]) != tc.cfg.Mic.Frames ||
				binary.LittleEndian.Uint32(fixture[28:32]) != tc.cfg.Mic.Rate ||
				fixture[32] != tc.cfg.Mic.Format || fixture[33] != tc.cfg.Mic.FullBits ||
				fixture[34] != tc.cfg.Mic.Alignment || fixture[35] != tc.cfg.Mic.ValidBits {
				t.Fatal("MIC_STREAM golden byte offsets differ")
			}
			encoded, err := (CFG2{Config: tc.cfg}).MarshalBinary()
			if err != nil || !bytes.Equal(encoded, fixture) {
				t.Fatalf("MarshalBinary differs from independent fixture: %x err=%v", encoded, err)
			}
			cfg, err := ParseCFG2(fixture)
			if err != nil || !reflect.DeepEqual(cfg.Config, tc.cfg) {
				t.Fatalf("ParseCFG2: %+v err=%v", cfg, err)
			}
			roundtrip, err := cfg.MarshalBinary()
			if err != nil || !bytes.Equal(roundtrip, fixture) {
				t.Fatalf("decode-encode not byte exact: %x err=%v", roundtrip, err)
			}
			var dst bytes.Buffer
			if err := WriteCFG2(&dst, cfg); err != nil || !bytes.Equal(dst.Bytes(), fixture) {
				t.Fatalf("WriteCFG2: err=%v", err)
			}
			// A stream reader consumes exactly total_bytes, not the next INP2.
			reader := bytes.NewReader(append(bytes.Clone(fixture), []byte("INP2")...))
			got, err := ReadCFG2(reader)
			if err != nil || !reflect.DeepEqual(got, cfg) || reader.Len() != 4 {
				t.Fatalf("ReadCFG2 message boundary: got=%+v err=%v trailing=%d", got, err, reader.Len())
			}
			if _, err := ParseCFG2(append(bytes.Clone(fixture), 0)); !errors.Is(err, ErrConfig) {
				t.Fatal("ParseCFG2 accepted trailing garbage", err)
			}
		})
	}
}

func TestCFG2OptionalExtensions(t *testing.T) {
	base := cfg2Fixture(t, "athena_4mic_2ref.cfg2")
	// Registered optional types are now typed: malformed values must reject
	// even if REQUIRED is not set. Unknown OPTIONAL remains safely skippable.
	for _, typ := range []uint16{cfg2PlatformInfo, cfg2MicFactoryGain} {
		t.Run(fmt.Sprintf("malformed_known_%04x", typ), func(t *testing.T) {
			cfg2MustReject(t, cfg2InsertTLV(base, len(base), typ, 0, []byte{0xff, 0, 0xa5, 0x7e}))
			cfg2MustReject(t, cfg2InsertTLV(base, len(base), typ, cfg2Required, []byte{7}))
		})
	}
	wire := cfg2InsertTLV(base, len(base), 0x7f01, 0, []byte{0xff, 0, 0xa5, 0x7e})
	got, err := ParseCFG2(wire)
	if err != nil || got.Config != cfg2Athena() || got.PlatformInfo != nil || got.MicFactoryGain != nil {
		t.Fatalf("unknown optional changed typed semantics: %+v err=%v", got, err)
	}
	encoded, err := got.MarshalBinary()
	if err != nil || !bytes.Equal(encoded, base) {
		t.Fatalf("unknown optional retained or modified core: %x err=%v", encoded, err)
	}
	cfg2MustReject(t, cfg2InsertTLV(base, len(base), 0x7f01, cfg2Required, []byte{7}))
}

func TestCFG2MalformedTLVs(t *testing.T) {
	base := cfg2Fixture(t, "athena_4mic_2ref.cfg2")
	mutate := func(f func([]byte)) []byte {
		out := bytes.Clone(base)
		f(out)
		return out
	}
	// All core TLV lengths and flags must be exact, including REQUIRED.
	for _, tc := range []struct {
		name string
		wire []byte
	}{
		{"zero_type", mutate(func(b []byte) { binary.LittleEndian.PutUint16(b[16:18], 0) })},
		{"reserved_type", mutate(func(b []byte) { binary.LittleEndian.PutUint16(b[40:42], 0xffff) })},
		{"duplicate", mutate(func(b []byte) { binary.LittleEndian.PutUint16(b[40:42], 1) })},
		{"reversed_types", mutate(func(b []byte) { binary.LittleEndian.PutUint16(b[16:18], 2) })},
		{"missing_mic", func() []byte {
			b := append(bytes.Clone(base[:16]), base[40:]...)
			cfg2PutSize(b)
			return b
		}()},
		{"missing_ref", func() []byte { b := bytes.Clone(base[:40]); cfg2PutSize(b); return b }()},
		{"core_optional_mic", mutate(func(b []byte) { binary.LittleEndian.PutUint16(b[18:20], 0) })},
		{"core_optional_ref", mutate(func(b []byte) { binary.LittleEndian.PutUint16(b[42:44], 0) })},
		{"core_unknown_flags", mutate(func(b []byte) { binary.LittleEndian.PutUint16(b[18:20], 3) })},
		{"optional_unknown_flags", cfg2InsertTLV(base, len(base), 0x7f01, 2, nil)},
		{"optional_unknown_reserved_type", cfg2InsertTLV(base, len(base), 0xffff, 0, nil)},
		{"core_short_value", mutate(func(b []byte) { binary.LittleEndian.PutUint32(b[20:24], 15) })},
		{"core_long_value", mutate(func(b []byte) { binary.LittleEndian.PutUint32(b[20:24], 17) })},
		{"forged_u32_length", mutate(func(b []byte) { binary.LittleEndian.PutUint32(b[20:24], ^uint32(0)) })},
		{"reserved_mic", mutate(func(b []byte) { b[36] = 1 })},
		{"reserved_ref", mutate(func(b []byte) { b[63] = 1 })},
		{"malformed_ref_shape", mutate(func(b []byte) { b[48] = 0 })},
		{"partial_tlv_header", func() []byte { b := append(bytes.Clone(base), 1, 2, 3); cfg2PutSize(b); return b }()},
		{"overrun_optional", func() []byte {
			b := cfg2InsertTLV(base, len(base), 0x7f01, 0, nil)
			binary.LittleEndian.PutUint32(b[len(base)+4:], ^uint32(0))
			return b
		}()},
		{"nonmonotone_unknown", cfg2InsertTLV(base, 40, 0x7f01, 0, nil)},
		{"duplicate_unknown", func() []byte {
			b := cfg2InsertTLV(base, len(base), 0x7f01, 0, nil)
			return cfg2InsertTLV(b, len(b), 0x7f01, 0, nil)
		}()},
	} {
		t.Run(tc.name, func(t *testing.T) { cfg2MustReject(t, tc.wire) })
	}
}

func TestCFG2HeaderBoundsAndTruncation(t *testing.T) {
	base := cfg2Fixture(t, "generic_1mic_noref.cfg2")
	for n := 0; n < len(base); n++ {
		t.Run(fmt.Sprintf("truncated_%d", n), func(t *testing.T) {
			cfg2MustReject(t, base[:n])
		})
	}
	for _, tc := range []struct {
		name string
		wire []byte
	}{
		{"bad_magic", func() []byte { b := bytes.Clone(base); b[0] = 'A'; return b }()},
		{"bad_version", func() []byte { b := bytes.Clone(base); b[4] = 3; return b }()},
		{"old_header_len_40", func() []byte { b := bytes.Clone(base); b[6] = 40; return b }()},
		{"wrong_header_len_0", func() []byte { b := bytes.Clone(base); b[6] = 0; return b }()},
		{"invalid_flags_0", func() []byte { b := bytes.Clone(base); b[12] = 0; return b }()},
		{"unknown_stream_flags", func() []byte { b := bytes.Clone(base); b[12] = 0x81; return b }()},
		{"unpaired_ref_flags", func() []byte { b := bytes.Clone(base); b[12] = 7; return b }()},
		{"length_15", func() []byte { b := bytes.Clone(base); binary.LittleEndian.PutUint32(b[8:], 15); return b }()},
		{"length_16", func() []byte { b := bytes.Clone(base[:16]); cfg2PutSize(b); return b }()},
		{"length_8193", func() []byte { b := bytes.Clone(base[:16]); binary.LittleEndian.PutUint32(b[8:], 8193); return b }()},
		{"length_max_u32", func() []byte { b := bytes.Clone(base[:16]); binary.LittleEndian.PutUint32(b[8:], ^uint32(0)); return b }()},
		{"body_short_for_8192", func() []byte { b := bytes.Clone(base[:16]); binary.LittleEndian.PutUint32(b[8:], 8192); return b }()},
	} {
		t.Run(tc.name, func(t *testing.T) { cfg2MustReject(t, tc.wire) })
	}

	// Exactly 8192 bytes: 16 header + 48 mandatory + 8 TLV header + 8120
	// opaque optional data. Bounded acceptance does not retain the payload.
	max := cfg2InsertTLV(base, len(base), 0x7f01, 0, make([]byte, 8120))
	if len(max) != CFG2MaxBytes {
		t.Fatal("bad maximal fixture length", len(max))
	}
	if got, err := ParseCFG2(max); err != nil || got.Config != cfg2Generic() {
		t.Fatalf("8192-byte bounded optional TLV must be accepted: %+v err=%v", got, err)
	}
	if _, err := ReadCFG2(bytes.NewReader(max)); err != nil {
		t.Fatal("read 8192-byte packet", err)
	}
	// 65 TLVs are illegal even when the packet fits within 8192 bytes.
	for optional := 62; optional <= 63; optional++ {
		b := bytes.Clone(base)
		for n := 0; n < optional; n++ {
			b = cfg2InsertTLV(b, len(b), uint16(0x4000+n), 0, nil)
		}
		if optional == 62 {
			if _, err := ParseCFG2(b); err != nil {
				t.Fatal("64 TLVs should be allowed", err)
			}
		} else {
			cfg2MustReject(t, b)
		}
	}
}

func TestCFG2ExistingPCMValidation(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(*Config)
	}{
		{"no_mic", func(c *Config) { c.Mic.Channels = 0 }},
		{"mic_17", func(c *Config) { c.Mic.Channels = 17 }},
		{"ref_9", func(c *Config) { c.Ref.Channels = 9 }},
		{"bad_frames", func(c *Config) { c.Mic.Frames = 1 }},
		{"bad_rate", func(c *Config) { c.Ref.Rate = 7999 }},
		{"rate_overflow", func(c *Config) { c.Mic.Rate = ^uint32(0) }},
		{"bad_format", func(c *Config) { c.Mic.Format = 3 }},
		{"bad_full_bits", func(c *Config) { c.Mic.FullBits = 24 }},
		{"bad_valid_bits", func(c *Config) { c.Mic.ValidBits = 33 }},
		{"bad_alignment", func(c *Config) { c.Ref.Alignment = 2 }},
		{"no_ref_but_flags7", func(c *Config) { c.Ref = Stream{} }},
		{"no_sync", func(c *Config) { c.Flags = Interleaved }},
		{"mismatched_ref", func(c *Config) { c.Flags = Interleaved | HardwareRef }},
		{"payload_over_256kiB", func(c *Config) {
			c.Mic.Channels, c.Ref.Channels = 16, 8
			c.Mic.Frames, c.Ref.Frames = 3840, 3840
			c.Mic.Rate, c.Ref.Rate = 192000, 192000
			c.Mic.Format, c.Ref.Format = S32, S32
			c.Mic.ValidBits, c.Ref.ValidBits = 32, 32
			c.Mic.FullBits, c.Ref.FullBits = 31, 31
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := cfg2Athena()
			tc.mutate(&cfg)
			if _, err := (CFG2{Config: cfg}).MarshalBinary(); !errors.Is(err, ErrConfig) {
				t.Fatalf("encoder accepted invalid PCM: %+v err=%v", cfg, err)
			}
			// Inject even invalid stream bytes directly (bypass encoder).
			wire := cfg2Fixture(t, "athena_4mic_2ref.cfg2")
			binary.LittleEndian.PutUint32(wire[12:16], cfg.Flags)
			for _, s := range []struct {
				offset int
				stream Stream
			}{{24, cfg.Mic}, {48, cfg.Ref}} {
				binary.LittleEndian.PutUint16(wire[s.offset:], s.stream.Channels)
				binary.LittleEndian.PutUint16(wire[s.offset+2:], s.stream.Frames)
				binary.LittleEndian.PutUint32(wire[s.offset+4:], s.stream.Rate)
				wire[s.offset+8], wire[s.offset+9], wire[s.offset+10], wire[s.offset+11] =
					s.stream.Format, s.stream.FullBits, s.stream.Alignment, s.stream.ValidBits
			}
			cfg2MustReject(t, wire)
		})
	}
	noRef := cfg2Generic()
	noRef.Ref.Rate = 96000
	if _, err := (CFG2{Config: noRef}).MarshalBinary(); !errors.Is(err, ErrConfig) {
		t.Fatal("absent REF with metadata accepted", err)
	}
}

func TestCFG2LegacyRejectedAndWriterErrors(t *testing.T) {
	legacy, err := os.ReadFile("testdata/rejected_fixed_cfg2.bin")
	if err != nil {
		t.Fatal(err)
	}
	cfg2MustReject(t, legacy)
	if err := WriteCFG2(shortWriter{}, CFG2{Config: cfg2Generic()}); !errors.Is(err, io.ErrShortWrite) {
		t.Fatal("short write not detected", err)
	}
	if err := WriteCFG2(io.Discard, CFG2{Config: Config{}}); !errors.Is(err, ErrConfig) {
		t.Fatal("invalid encoder input did not reject", err)
	}
}

func TestCFG2EncoderBoundsAndOrdering(t *testing.T) {
	newEncoder := func() cfg2Encoder {
		return cfg2Encoder{wire: make([]byte, CFG2HeaderBytes)}
	}
	max := newEncoder()
	if err := max.add(1, cfg2Required, make([]byte, CFG2MaxBytes-CFG2HeaderBytes-cfg2TLVHeaderBytes)); err != nil || len(max.wire) != CFG2MaxBytes {
		t.Fatalf("encoder failed to accept exactly 8192: %v length=%d", err, len(max.wire))
	}
	if err := max.add(2, 0, nil); !errors.Is(err, ErrConfig) {
		t.Fatal("encoder permitted byte length above 8192", err)
	}
	oversized := newEncoder()
	if err := oversized.add(1, 0, make([]byte, CFG2MaxBytes-CFG2HeaderBytes-cfg2TLVHeaderBytes+1)); !errors.Is(err, ErrConfig) {
		t.Fatal("encoder permitted oversized value", err)
	}
	many := newEncoder()
	for typ := uint16(1); typ <= CFG2MaxTLVs; typ++ {
		if err := many.add(typ, 0, nil); err != nil {
			t.Fatal("encoder incorrectly denied TLV 64", typ, err)
		}
	}
	if err := many.add(65, 0, nil); !errors.Is(err, ErrConfig) {
		t.Fatal("encoder accepted 65 TLVs", err)
	}
	for _, tc := range []struct {
		typ, flags uint16
	}{
		{0, 0}, {0xffff, 0}, {1, 2}, {1, 0xffff},
	} {
		e := newEncoder()
		if err := e.add(tc.typ, tc.flags, nil); !errors.Is(err, ErrConfig) {
			t.Fatal("encoder accepted invalid type/flags", tc, err)
		}
	}
	ordered := newEncoder()
	if err := ordered.add(2, cfg2Required, nil); err != nil {
		t.Fatal(err)
	}
	for _, typ := range []uint16{1, 2} {
		if err := ordered.add(typ, 0, nil); !errors.Is(err, ErrConfig) {
			t.Fatal("encoder accepted duplicate or unordered type", typ, err)
		}
	}
}
