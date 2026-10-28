package mic

import (
	"bytes"
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/HuskerMinion/techo5/echod/internal/asp"
)

func genericSessionConfig() asp.Config {
	return asp.Config{
		Flags: asp.Interleaved,
		Mic: asp.Stream{
			Channels: 1, Frames: 320, Rate: 16000,
			Format: asp.S16, FullBits: 15, Alignment: asp.Low, ValidBits: 16,
		},
	}
}

func TestSessionCFG2AthenaLowHighAndFactoryStates(t *testing.T) {
	for _, bits := range []uint32{23, 31} {
		t.Run(strings.Repeat("x", int(bits-22)), func(t *testing.T) {
			micPCM := bytes.Repeat([]byte{0x00, 0xff, 0x80, 0x7f}, nativeMicBytes/4)
			refPCM := bytes.Repeat([]byte{0x80, 0xff, 0x01, 0x00}, nativeRefBytes/4)
			beforeMIC, beforeREF := bytes.Clone(micPCM), bytes.Clone(refPCM)
			n := nativeFrame{mics: micPCM, ref: refPCM, fullBits: bits}
			input := n.input()
			if err := input.Validate(); err != nil {
				t.Fatal(err)
			}
			originalConfig := input.Config
			shapeBefore, err := (asp.CFG2{Config: input.Config}).MarshalBinary()
			if err != nil {
				t.Fatal(err)
			}
			core, err := (asp.CFG2{Config: input.Config}).MarshalBinary()
			if err != nil {
				t.Fatal(err)
			}

			for _, tc := range []struct {
				name   string
				values []string
				avail  asp.FactoryGainAvailability
				have   bool
			}{
				{"complete", []string{"1", "65536", "32767", "2147483647"}, asp.FactoryGainComplete, true},
				{"partial", []string{"1", "", "0", "65536"}, asp.FactoryGainPartial, true},
				{"none", []string{"", "", "", ""}, asp.FactoryGainNone, true},
				{"absent", nil, asp.FactoryGainNone, false},
			} {
				t.Run(tc.name, func(t *testing.T) {
					sourceConfig := originalConfig
					p := confirmedFactoryAthena()
					g := micFactoryGainFromValues(tc.values, sourceConfig, p)
					if (g != nil) != tc.have {
						t.Fatalf("gain qualification mismatch: %+v", g)
					}
					expected := asp.CFG2{Config: originalConfig, PlatformInfo: p, MicFactoryGain: g}
					expectedWire, err := expected.MarshalBinary()
					if err != nil {
						t.Fatal(err)
					}
					s, err := buildSessionCFG2(sourceConfig, p, g)
					if err != nil {
						t.Fatal(err)
					}
					if s.packet == "" || !bytes.Equal(s.Bytes(), expectedWire) {
						t.Fatal("session wire differs from canonical ordered CFG2")
					}
					if len(s.Bytes()) > asp.CFG2MaxBytes || !bytes.Equal(s.Bytes()[16:len(core)], core[16:]) {
						t.Fatal("session changes the native MIC/REF descriptor or exceeds wire limit")
					}
					decoded, err := asp.ParseCFG2(s.Bytes())
					if err != nil {
						t.Fatal(err)
					}
					if decoded.Config != originalConfig || !reflect.DeepEqual(decoded.PlatformInfo, p) ||
						!reflect.DeepEqual(decoded.MicFactoryGain, g) {
						t.Fatalf("roundtrip mismatch: %+v", decoded)
					}
					if tc.have && decoded.MicFactoryGain.Availability != tc.avail {
						t.Fatal("gain availability lost")
					}
					reencoded, err := decoded.MarshalBinary()
					if err != nil || !bytes.Equal(reencoded, expectedWire) {
						t.Fatal("re-encoding not byte-identical", err)
					}

					// Caller mutation, returned-byte mutation and repeated reads
					// must not alter the live session snapshot.
					first := s.Bytes()
					p.Codename, p.Detection = "other", asp.PlatformInferred
					if g != nil {
						g.PresentMask = 0
						g.Coefficients[0] = 0
					}
					sourceConfig.Mic.Rate = 8000
					first[0] ^= 0xff
					if !bytes.Equal(s.Bytes(), expectedWire) || !bytes.Equal(s.Bytes(), reencoded) {
						t.Fatal("session retained caller alias or exposed mutable packet")
					}
					if n.input().Config != originalConfig || !bytes.Equal(micPCM, beforeMIC) ||
						!bytes.Equal(refPCM, beforeREF) ||
						!bytes.Equal(input.Mic, beforeMIC) || !bytes.Equal(input.Ref, beforeREF) {
						t.Fatal("builder altered native streams or audio PCM")
					}
					shapeAfter, err := (asp.CFG2{Config: n.input().Config}).MarshalBinary()
					if err != nil || !bytes.Equal(shapeBefore, shapeAfter) {
						t.Fatal("builder altered the CFG2 stream shape")
					}
				})
			}
		})
	}
}

func TestSessionCFG2GenericAndUnknownPlatform(t *testing.T) {
	cfg := genericSessionConfig()
	core, err := (asp.CFG2{Config: cfg}).MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name     string
		platform *asp.PlatformInfo
	}{
		{"absent", nil},
		{"inferred Athena", &asp.PlatformInfo{Codename: "athena", Detection: asp.PlatformInferred}},
		{"inferred unknown", &asp.PlatformInfo{Codename: "future-board_1", Detection: asp.PlatformInferred}},
		{"confirmed other", &asp.PlatformInfo{Codename: "rook", Detection: asp.PlatformConfirmed}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s, err := buildSessionCFG2(cfg, tc.platform, nil)
			if err != nil {
				t.Fatal(err)
			}
			decoded, err := asp.ParseCFG2(s.Bytes())
			if err != nil {
				t.Fatal(err)
			}
			if decoded.Config != cfg || decoded.Config.Flags != asp.Interleaved ||
				decoded.Config.Ref != (asp.Stream{}) || decoded.MicFactoryGain != nil ||
				!reflect.DeepEqual(decoded.PlatformInfo, tc.platform) {
				t.Fatalf("unexpected inferred metadata/REF/gain: %+v", decoded)
			}
			if tc.platform == nil && !bytes.Equal(s.Bytes(), core) {
				t.Fatal("nil optional metadata must not synthesize optional TLV")
			}
			reencoded, err := decoded.MarshalBinary()
			if err != nil || !bytes.Equal(s.Bytes(), reencoded) {
				t.Fatal("generic CFG2 roundtrip failed", err)
			}
		})
	}
}

func TestSessionCFG2RejectInvalidConfigAndTypedMetadata(t *testing.T) {
	cfg := testAthenaFactoryConfig()
	p := confirmedFactoryAthena()
	g := micFactoryGainFromValues([]string{"1", "1", "1", "1"}, cfg, p)
	if g == nil {
		t.Fatal("invalid test fixture")
	}
	for _, tc := range []struct {
		name string
		cfg  asp.Config
	}{
		{"zero config", asp.Config{}},
		{"missing mic", func() asp.Config { c := cfg; c.Mic.Channels = 0; return c }()},
		{"mic frames", func() asp.Config { c := cfg; c.Mic.Frames--; return c }()},
		{"mic rate", func() asp.Config { c := cfg; c.Mic.Rate = 44100; return c }()},
		{"mic format", func() asp.Config { c := cfg; c.Mic.Format = 99; return c }()},
		{"mic bits", func() asp.Config { c := cfg; c.Mic.ValidBits = 33; return c }()},
		{"mic alignment", func() asp.Config { c := cfg; c.Mic.Alignment = 7; return c }()},
		{"unsupported flags", func() asp.Config { c := cfg; c.Flags |= 8; return c }()},
		{"ref mismatch", func() asp.Config { c := cfg; c.Ref.Frames--; return c }()},
		{"no REF with hardware flags", func() asp.Config { c := cfg; c.Ref = asp.Stream{}; return c }()},
		{"ref without hardware flags", func() asp.Config { c := cfg; c.Flags = asp.Interleaved; return c }()},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := buildSessionCFG2(tc.cfg, nil, nil); !errors.Is(err, asp.ErrConfig) {
				t.Fatalf("invalid config accepted: %v", err)
			}
		})
	}
	for _, tc := range []struct {
		name string
		p    *asp.PlatformInfo
		g    *asp.MicFactoryGain
		cfg  asp.Config
	}{
		{"bad uppercase codename", &asp.PlatformInfo{Codename: "Athena", Detection: asp.PlatformConfirmed}, nil, cfg},
		{"empty codename", &asp.PlatformInfo{Detection: asp.PlatformInferred}, nil, cfg},
		{"long codename", &asp.PlatformInfo{Codename: strings.Repeat("a", 33), Detection: asp.PlatformInferred}, nil, cfg},
		{"embedded NUL", &asp.PlatformInfo{Codename: "a\x00b", Detection: asp.PlatformConfirmed}, nil, cfg},
		{"bad detection", &asp.PlatformInfo{Codename: "athena", Detection: 99}, nil, cfg},
		{"absent platform with gain", nil, g, cfg},
		{"inferred with gain", &asp.PlatformInfo{Codename: "athena", Detection: asp.PlatformInferred}, g, cfg},
		{"other confirmed with gain", &asp.PlatformInfo{Codename: "rook", Detection: asp.PlatformConfirmed}, g, cfg},
		{"generic with gain", p, g, genericSessionConfig()},
		{"bad gain channels", p, func() *asp.MicFactoryGain { v := *g; v.Channels = 3; return &v }(), cfg},
		{"bad mask", p, func() *asp.MicFactoryGain { v := *g; v.PresentMask = 0x1f; return &v }(), cfg},
		{"missing coefficient", p, func() *asp.MicFactoryGain { v := *g; v.Coefficients[0] = 0; return &v }(), cfg},
		{"oversized Q30", p, func() *asp.MicFactoryGain { v := *g; v.Coefficients[0] = 1 << 32; return &v }(), cfg},
		{"PCM applied", p, func() *asp.MicFactoryGain { v := *g; v.AppliedToPCM = true; return &v }(), cfg},
		{"incomplete complete", p, func() *asp.MicFactoryGain { v := *g; v.PresentMask = 0x7; v.Coefficients[3] = 0; return &v }(), cfg},
		{"NONE carrying gain", p, func() *asp.MicFactoryGain { v := *g; v.Availability = asp.FactoryGainNone; return &v }(), cfg},
		{"unqualified ref", p, g, func() asp.Config { c := cfg; c.Ref.Channels = 1; return c }()},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := buildSessionCFG2(tc.cfg, tc.p, tc.g); !errors.Is(err, asp.ErrConfig) {
				t.Fatalf("invalid metadata accepted: %v", err)
			}
		})
	}
}

func TestSessionCFG2OneShotInjectedReaders(t *testing.T) {
	config := testAthenaFactoryConfig()
	var platformCalls, gainCalls int
	readNames := []string{}
	fields := []string{"1", "65536", "", "32767"}
	platformReader := func() *asp.PlatformInfo {
		platformCalls++
		return confirmedFactoryAthena()
	}
	gainReader := func(c asp.Config, p *asp.PlatformInfo) *asp.MicFactoryGain {
		gainCalls++
		if c != config || p == nil || p.Detection != asp.PlatformConfirmed {
			t.Fatal("incorrect one-shot collection arguments")
		}
		return micFactoryGainFromReader(c, p, "athena", func(name string) string {
			readNames = append(readNames, name)
			return fields[len(readNames)-1]
		})
	}
	s, err := prepareASP2SessionCFG2WithReaders(config, platformReader, gainReader)
	if err != nil {
		t.Fatal(err)
	}
	if platformCalls != 1 || gainCalls != 1 ||
		!reflect.DeepEqual(readNames, []string{"miccal.0", "miccal.1", "miccal.2", "miccal.3"}) {
		t.Fatal("one-shot collectors not bounded to startup", platformCalls, gainCalls, readNames)
	}
	before := s.Bytes()
	fields[0] = "0"
	fields[1] = "0"
	if !bytes.Equal(before, s.Bytes()) {
		t.Fatal("IDME source changed established session")
	}
	decoded, err := asp.ParseCFG2(s.Bytes())
	if err != nil || decoded.MicFactoryGain == nil ||
		decoded.MicFactoryGain.Availability != asp.FactoryGainPartial || decoded.MicFactoryGain.PresentMask != 0xb {
		t.Fatalf("wrong synthetic IDME snapshot: %+v, %v", decoded, err)
	}

	// Invalid native config or platform must fail before gain reads.
	platformCalls, gainCalls = 0, 0
	if _, err := prepareASP2SessionCFG2WithReaders(asp.Config{}, platformReader, gainReader); err == nil ||
		platformCalls != 0 || gainCalls != 0 {
		t.Fatal("invalid config must not consult metadata readers")
	}
	if _, err := prepareASP2SessionCFG2WithReaders(config, func() *asp.PlatformInfo {
		platformCalls++
		return &asp.PlatformInfo{Codename: "Bad", Detection: asp.PlatformConfirmed}
	}, gainReader); err == nil || platformCalls != 1 || gainCalls != 0 {
		t.Fatal("invalid platform must stop before gain reads")
	}
	if _, err := prepareASP2SessionCFG2WithReaders(config, nil, gainReader); !errors.Is(err, asp.ErrConfig) {
		t.Fatal("missing reader must fail closed")
	}
	// Unknown/inferred board: valid snapshot but no factory gain and no IDME.
	noGainReads := 0
	inferred, err := prepareASP2SessionCFG2WithReaders(genericSessionConfig(),
		func() *asp.PlatformInfo {
			return &asp.PlatformInfo{Codename: "unknown", Detection: asp.PlatformInferred}
		}, func(c asp.Config, p *asp.PlatformInfo) *asp.MicFactoryGain {
			return micFactoryGainFromReader(c, p, "unknown", func(string) string {
				noGainReads++
				return "1"
			})
		})
	if err != nil || noGainReads != 0 {
		t.Fatal("inferred platform must not read IDME", err)
	}
	v, err := asp.ParseCFG2(inferred.Bytes())
	if err != nil || v.PlatformInfo == nil || v.PlatformInfo.Detection != asp.PlatformInferred || v.MicFactoryGain != nil {
		t.Fatal("inferred platform metadata fabricated", err)
	}
}

func TestSessionCFG2ConcurrentIndependentReads(t *testing.T) {
	cfg := testAthenaFactoryConfig()
	p := confirmedFactoryAthena()
	g := micFactoryGainFromValues([]string{"1", "65536", "", ""}, cfg, p)
	s, err := buildSessionCFG2(cfg, p, g)
	if err != nil {
		t.Fatal(err)
	}
	before := s.Bytes()
	var wg sync.WaitGroup
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				v := s.Bytes()
				if !bytes.Equal(v, before) {
					t.Error("independent reader observed mutation")
					return
				}
				v[0], v[len(v)-1] = 0, 0
			}
		}()
	}
	wg.Wait()
	if !bytes.Equal(s.Bytes(), before) {
		t.Fatal("concurrent reader mutated immutable packet")
	}
}

func TestSessionCFG2NoHotPathMetadataCallers(t *testing.T) {
	// Only startup functions may build or collect metadata; inspect capture
	// function bodies even when cold and hot code share a source file.
	for _, path := range []string{
		"asp.go", "asp_control.go", "asp_path.go", "mic.go", "athena.go", "capture_alsa.go",
	} {
		t.Run(path, func(t *testing.T) {
			tree, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
			if err != nil {
				t.Fatal(err)
			}
			for _, decl := range tree.Decls {
				fn, ok := decl.(*ast.FuncDecl)
				if !ok || fn.Body == nil {
					continue
				}
				switch fn.Name.Name {
				case "input", "exchange", "process", "processASP", "broadcastNative", "Run", "Read", "retainNative", "convert":
				default:
					continue
				}
				ast.Inspect(fn.Body, func(n ast.Node) bool {
					call, ok := n.(*ast.CallExpr)
					if !ok {
						return true
					}
					switch f := call.Fun.(type) {
					case *ast.Ident:
						switch f.Name {
						case "buildSessionCFG2", "prepareASP2SessionCFG2",
							"prepareASP2SessionCFG2WithReaders", "platformInfoForSession", "micFactoryGainForSession":
							t.Errorf("session builder/collector called in active hot path: %s", f.Name)
						}
					case *ast.SelectorExpr:
						if f.Sel.Name == "Idme" {
							t.Error("IDME accessed in active hot path")
						}
					}
					return true
				})
			}
		})
	}
}
