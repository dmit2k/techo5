package mic

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/token"
	"math"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/HuskerMinion/techo5/echod/internal/asp"
)

func testAthenaFactoryConfig() asp.Config {
	return nativeFrame{
		mics:     make([]byte, nativeMicBytes),
		ref:      make([]byte, nativeRefBytes),
		fullBits: 31,
	}.input().Config
}

func confirmedFactoryAthena() *asp.PlatformInfo {
	return &asp.PlatformInfo{Codename: "athena", Detection: asp.PlatformConfirmed}
}

func TestFactoryIDMEQ30StrictDecimal(t *testing.T) {
	tests := []struct {
		name, source string
		want         int64
		ok           bool
	}{
		{"low one shift", "1", 65536, true},
		{"low boundary", "32767", 2147418112, true}, // 0x7fff << 16
		{"shift overflow", "32768", 0, false},       // 0x8000 still selects shift
		{"high shift overflow", "65535", 0, false},
		{"first bypass", "65536", 65536, true}, // 0x10000 unchanged
		{"raw max signed32", "2147483647", math.MaxInt32, true},
		{"upper valid Q30", "1073741824", 1 << 30, true},
		{"zero", "0", 0, false},
		{"negative", "-1", 0, false},
		{"signed min", "-2147483648", 0, false},
		{"signed overflow", "2147483648", 0, false},
		{"long decimal overflow", "99999999999999999999999999999", 0, false},
		{"empty", "", 0, false},
		{"plus", "+1", 0, false},
		{"spaces prefix", " 1", 0, false},
		{"spaces suffix", "1 ", 0, false},
		{"newline", "1\n", 0, false},
		{"tab", "\t65536", 0, false},
		{"hex", "0x10000", 0, false},
		{"decimal point", "1.0", 0, false},
		{"exponent", "1e3", 0, false},
		{"underscore", "1_000", 0, false},
		{"ASCII letters", "abc", 0, false},
		{"Unicode Arabic digit", "١", 0, false},
		{"Unicode fullwidth digit", "１", 0, false},
		{"NUL", "1\x00", 0, false},
		{"zeros", "00000", 0, false},
		{"zero-padded decimal", "00001", 65536, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := factoryIDMEQ30(tt.source)
			if got != tt.want || ok != tt.ok {
				t.Fatalf("factoryIDMEQ30(%q) = (%d,%t), want (%d,%t)", tt.source, got, ok, tt.want, tt.ok)
			}
		})
	}
}

func TestFactoryGainMasksAndAvailability(t *testing.T) {
	tests := []struct {
		name   string
		fields []string
		mask   uint32
		state  asp.FactoryGainAvailability
		gains  [4]int64
	}{
		{"all four", []string{"1", "65536", "32767", "2147483647"}, 0xf, asp.FactoryGainComplete, [4]int64{65536, 65536, 2147418112, math.MaxInt32}},
		{"one malformed", []string{"1", "65536", "invalid", "32767"}, 0xb, asp.FactoryGainPartial, [4]int64{65536, 65536, 0, 2147418112}},
		{"sparse slots one and three", []string{"", "1", "bad", "65536"}, 0xa, asp.FactoryGainPartial, [4]int64{0, 65536, 0, 65536}},
		{"sparse first only", []string{"1", "", "", ""}, 0x1, asp.FactoryGainPartial, [4]int64{65536, 0, 0, 0}},
		{"sparse final only", []string{"", "", "", "1073741824"}, 0x8, asp.FactoryGainPartial, [4]int64{0, 0, 0, 1073741824}},
		{"all absent", []string{"", "", "", ""}, 0, asp.FactoryGainNone, [4]int64{}},
		{"all malformed", []string{"0", "-1", "32768", "xx"}, 0, asp.FactoryGainNone, [4]int64{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := micFactoryGainFromValues(tt.fields, testAthenaFactoryConfig(), confirmedFactoryAthena())
			if g == nil {
				t.Fatal("qualified four-field snapshot omitted")
			}
			if g.Channels != 4 || g.PresentMask != tt.mask || g.Availability != tt.state || g.AppliedToPCM {
				t.Fatalf("metadata mismatch: %+v", g)
			}
			if !reflect.DeepEqual(g.Coefficients[:4], tt.gains[:]) {
				t.Fatal("Q30 values lost captured-slot alignment")
			}
			for i := 4; i < len(g.Coefficients); i++ {
				if g.Coefficients[i] != 0 {
					t.Fatal("coefficient outside captured channels", i)
				}
			}
			if tt.mask == 0 {
				if g.Encoding != asp.FactoryGainEncodingNone || g.Provenance != asp.FactoryGainProvenanceNone {
					t.Fatal("NONE claims real factory values", g)
				}
			} else if g.Encoding != asp.FactoryGainEncodingQ30 || g.Provenance != asp.FactoryGainProvenanceIDME {
				t.Fatal("valid gain not labeled factory Q30", g)
			}
			// No invalid/absent field may be advertised as implicit 1<<30.
			for i := 0; i < 4; i++ {
				if tt.mask&(1<<uint(i)) == 0 && g.Coefficients[i] != 0 {
					t.Fatal("fabricated missing-channel unity")
				}
			}
		})
	}
}

func TestFactoryGainQualificationAndNotAdvertised(t *testing.T) {
	ready := testAthenaFactoryConfig()
	confirmed := confirmedFactoryAthena()
	fields := []string{"1", "1", "1", "1"}
	mutations := []struct {
		name     string
		config   asp.Config
		platform *asp.PlatformInfo
	}{
		{"not confirmed", ready, &asp.PlatformInfo{Codename: "athena", Detection: asp.PlatformInferred}},
		{"other confirmed", ready, &asp.PlatformInfo{Codename: "cronos", Detection: asp.PlatformConfirmed}},
		{"invalid platform", ready, &asp.PlatformInfo{Codename: "Athena", Detection: asp.PlatformConfirmed}},
		{"no platform", ready, nil},
	}
	for _, tt := range mutations {
		t.Run(tt.name, func(t *testing.T) {
			if g := micFactoryGainFromValues(fields, tt.config, tt.platform); g != nil {
				t.Fatalf("unqualified factory gain emitted: %+v", g)
			}
		})
	}
	alterations := []struct {
		name string
		edit func(*asp.Config)
	}{
		{"three mics", func(c *asp.Config) { c.Mic.Channels = 3 }},
		{"five mics", func(c *asp.Config) { c.Mic.Channels = 5 }},
		{"wrong frames", func(c *asp.Config) { c.Mic.Frames = 960 }},
		{"wrong rate", func(c *asp.Config) { c.Mic.Rate = 48000; c.Mic.Frames = 960 }},
		{"mic not S32", func(c *asp.Config) { c.Mic.Format = asp.S16 }},
		{"mic not 24 bits", func(c *asp.Config) { c.Mic.ValidBits = 32 }},
		{"bad low align", func(c *asp.Config) { c.Mic.Alignment = asp.Low }},
		{"valid low exponent only", func(c *asp.Config) { c.Mic.FullBits = 23 }},
		{"no hardware REF", func(c *asp.Config) { c.Ref = asp.Stream{}; c.Flags = asp.Interleaved }},
		{"one REF channel", func(c *asp.Config) { c.Ref.Channels = 1 }},
		{"wrong REF rate", func(c *asp.Config) { c.Ref.Rate = 48000; c.Ref.Frames = 960 }},
		{"wrong REF format", func(c *asp.Config) { c.Ref.Format = asp.S32 }},
		{"no sync", func(c *asp.Config) { c.Flags = asp.Interleaved | asp.HardwareRef }},
	}
	for _, tt := range alterations {
		t.Run(tt.name, func(t *testing.T) {
			c := ready
			tt.edit(&c)
			if g := micFactoryGainFromValues(fields, c, confirmed); g != nil {
				t.Fatalf("mismatched native MIC/REF source emits factory metadata: %+v", g)
			}
		})
	}
	for _, tc := range []struct {
		name   string
		fields []string
	}{
		{"nil source", nil},
		{"zero-field snapshot", []string{}},
		{"short snapshot", []string{"1", "1", "1"}},
		{"overlong snapshot", []string{"1", "1", "1", "1", "1"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if g := micFactoryGainFromValues(tc.fields, ready, confirmed); g != nil {
				t.Fatal("invalid source shape must omit TLV, not claim NONE")
			}
		})
	}
	// The native capture supports both low/high 24-bit alignment.
	low := ready
	low.Mic.FullBits, low.Mic.Alignment = 23, asp.Low
	if micFactoryGainFromValues(fields, low, confirmed) == nil {
		t.Fatal("valid low-aligned Athena MIC stream rejected")
	}
}

func TestFactoryGainOneShotReaderAndSnapshot(t *testing.T) {
	fields := []string{"1", "65536", "32767", "2147483647"}
	var calls []string
	read := func(name string) string {
		calls = append(calls, name)
		return fields[len(calls)-1]
	}
	config, platform := testAthenaFactoryConfig(), confirmedFactoryAthena()
	g := micFactoryGainFromReader(config, platform, "athena", read)
	if g == nil || g.PresentMask != 0xf || g.Availability != asp.FactoryGainComplete {
		t.Fatalf("expected complete one-shot factory snapshot: %+v", g)
	}
	wantNames := []string{"miccal.0", "miccal.1", "miccal.2", "miccal.3"}
	if !reflect.DeepEqual(calls, wantNames) {
		t.Fatalf("IDME reads out of order or repeated: %v", calls)
	}
	first := *g
	fields[0] = "0"
	fields[1] = "0"
	if *g != first {
		t.Fatal("returned factory descriptor retained caller's mutable snapshot")
	}
	calls = nil
	g2 := micFactoryGainFromReader(config, platform, "athena", read)
	if g2 == nil || g2.Availability != asp.FactoryGainPartial || g2.PresentMask != 0xc ||
		!reflect.DeepEqual(calls, wantNames) || *g != first {
		t.Fatal("new session should get a fresh snapshot without changing first")
	}
	// No I/O on unqualified input, unknown board, nil reader or inferred identity.
	for _, tc := range []struct {
		name     string
		board    string
		cfg      asp.Config
		platform *asp.PlatformInfo
		reader   func(string) string
	}{
		{"mismatched board", "cronos", config, platform, read},
		{"inferred board", "athena", config, &asp.PlatformInfo{Codename: "athena", Detection: asp.PlatformInferred}, read},
		{"no reader", "athena", config, platform, nil},
		{"wrong channels", "athena", func() asp.Config { c := config; c.Mic.Channels = 3; return c }(), platform, read},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls = nil
			if v := micFactoryGainFromReader(tc.cfg, tc.platform, tc.board, tc.reader); v != nil {
				t.Fatal("reader should omit metadata when unsupported")
			}
			if len(calls) != 0 {
				t.Fatal("unqualified hardware read attempted:", calls)
			}
		})
	}
}

func TestFactoryGainCFG2WireAndRawPCMInvariants(t *testing.T) {
	micPCM := bytes.Repeat([]byte{0x10, 0xff, 0x80, 0x00}, nativeMicBytes/4)
	refPCM := bytes.Repeat([]byte{0xff, 0x00, 0x42, 0x7f}, nativeRefBytes/4)
	origMIC, origREF := bytes.Clone(micPCM), bytes.Clone(refPCM)
	for _, exponent := range []uint32{23, 31} {
		t.Run(strings.Repeat("x", int(exponent-22)), func(t *testing.T) {
			native := nativeFrame{mics: micPCM, ref: refPCM, fullBits: exponent}
			in := native.input()
			if err := in.Validate(); err != nil {
				t.Fatal(err)
			}
			shapeBefore, err := (asp.CFG2{Config: in.Config}).MarshalBinary()
			if err != nil {
				t.Fatal(err)
			}
			core, err := (asp.CFG2{Config: in.Config}).MarshalBinary()
			if err != nil {
				t.Fatal(err)
			}
			for _, tc := range []struct {
				name   string
				fields []string
				want   asp.FactoryGainAvailability
				tlv    bool
			}{
				{"complete", []string{"1", "65536", "32767", "2147483647"}, asp.FactoryGainComplete, true},
				{"partial", []string{"1", "", "0", "65536"}, asp.FactoryGainPartial, true},
				{"none", []string{"", "", "", ""}, asp.FactoryGainNone, true},
				{"not advertised", nil, asp.FactoryGainNone, false},
			} {
				t.Run(tc.name, func(t *testing.T) {
					g := micFactoryGainFromValues(tc.fields, in.Config, confirmedFactoryAthena())
					if (g != nil) != tc.tlv {
						t.Fatalf("TLV presence mismatch: %+v", g)
					}
					w, err := (asp.CFG2{Config: in.Config, MicFactoryGain: g}).MarshalBinary()
					if err != nil {
						t.Fatal(err)
					}
					parsed, err := asp.ParseCFG2(w)
					if err != nil {
						t.Fatal(err)
					}
					if parsed.Config != in.Config || !reflect.DeepEqual(parsed.MicFactoryGain, g) {
						t.Fatalf("CFG2 roundtrip mismatch: original=%+v decoded=%+v", g, parsed.MicFactoryGain)
					}
					if parsed.PlatformInfo != nil {
						t.Fatal("unrequested platform metadata synthesized")
					}
					if !bytes.Equal(core[16:], w[16:len(core)]) {
						t.Fatal("factory metadata altered mandatory MIC/REF bytes")
					}
					if tc.tlv {
						if g.Availability != tc.want || len(w) == len(core) {
							t.Fatal("factory descriptor was not serialized as distinct TLV")
						}
					} else if !bytes.Equal(w, core) {
						t.Fatal("nil must omit optional TLV")
					}
					shapeAfter, err := (asp.CFG2{Config: native.input().Config}).MarshalBinary()
					if err != nil || !bytes.Equal(shapeBefore, shapeAfter) ||
						!bytes.Equal(native.mics, origMIC) || !bytes.Equal(native.ref, origREF) ||
						!bytes.Equal(in.Mic, origMIC) || !bytes.Equal(in.Ref, origREF) {
						t.Fatal("metadata mutated native input, PCM or CFG2 shape")
					}
				})
			}
		})
	}
}

func TestFactoryGainNoCapturePathCallers(t *testing.T) {
	// AST regression guard against accidental insertion into the 20-ms tick
	// and the legacy ASP2 host. Only the opt-in session adapter may use
	// layout.Idme; synthetic tests invoke an injected reader instead.
	for _, path := range []string{"asp.go", "athena.go", "capture_alsa.go", "asp_path.go"} {
		t.Run(path, func(t *testing.T) {
			if _, err := os.Stat(path); err != nil {
				if os.IsNotExist(err) {
					t.Skip("source file excluded in this build tree")
				}
				t.Fatal(err)
			}
			tree, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
			if err != nil {
				t.Fatal(err)
			}
			ast.Inspect(tree, func(node ast.Node) bool {
				call, ok := node.(*ast.CallExpr)
				if !ok {
					return true
				}
				switch f := call.Fun.(type) {
				case *ast.Ident:
					switch f.Name {
					case "micFactoryGainForSession", "micFactoryGainFromReader", "micFactoryGainFromValues":
						t.Errorf("factory metadata builder called in capture/legacy host: %s", f.Name)
					}
				case *ast.SelectorExpr:
					if f.Sel.Name == "Idme" {
						t.Error("IDME read in capture/legacy host")
					}
				}
				return true
			})
		})
	}
}
