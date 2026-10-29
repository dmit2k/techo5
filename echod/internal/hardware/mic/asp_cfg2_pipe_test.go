package mic

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/HuskerMinion/techo5/echod/internal/asp"
)

// cfg2Pipe is deliberately NOT externalASP: it is an independent parent of
// the existing test executable; host supervision is tested separately.
type cfg2Pipe struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
	done   chan struct{}
	exit   error
}

func newCFG2Pipe(t *testing.T, mode string, env ...string) *cfg2Pipe {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	base := childCommand(mode)
	cmd := exec.CommandContext(ctx, base.Path, base.Args[1:]...)
	cmd.Env = append(os.Environ(), "TECHO5_ASP_CHILD="+mode,
		"GORACE=atexit_sleep_ms=0")
	cmd.Env = append(cmd.Env, env...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		cancel()
		t.Fatal(err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		cancel()
		t.Fatal(err)
	}
	p := &cfg2Pipe{cmd: cmd, stdin: stdin, stdout: stdout, done: make(chan struct{})}
	go func() { p.exit = cmd.Wait(); close(p.done) }()
	t.Cleanup(func() {
		_ = p.stdin.Close()
		cancel()
		_ = p.stdout.Close()
		select {
		case <-p.done:
		case <-time.After(2 * time.Second):
			t.Errorf("synthetic child was not reaped (pid=%d)", p.cmd.Process.Pid)
		}
	})
	if f, ok := stdout.(*os.File); ok {
		_ = f.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	}
	var hello [4]byte
	start := time.Now()
	if _, err := io.ReadFull(stdout, hello[:]); err != nil || string(hello[:]) != "ASP2" {
		t.Fatalf("HELLO: %q, error=%v", hello, err)
	}
	t.Logf("HELLO within 500ms bound: %s", time.Since(start))
	return p
}

func (p *cfg2Pipe) readDeadline(d time.Duration) {
	if f, ok := p.stdout.(*os.File); ok {
		_ = f.SetReadDeadline(time.Now().Add(d))
	}
}

func (p *cfg2Pipe) write(t *testing.T, wire []byte, chunk int) {
	t.Helper()
	if chunk <= 0 {
		t.Fatal("invalid chunk")
	}
	for at := 0; at < len(wire); {
		end := min(len(wire), at+chunk)
		if err := asp.WriteAll(p.stdin, wire[at:end]); err != nil {
			t.Fatalf("fragmented stdin write [%d:%d]: %v", at, end, err)
		}
		at = end
	}
}

func (p *cfg2Pipe) ack(t *testing.T, cfg asp.Config) {
	t.Helper()
	p.readDeadline(250 * time.Millisecond) // scheduling guard, not native latency proof
	got, err := asp.ReadResult(p.stdout)
	if err != nil || got.Type != asp.Accept || got.Status.Terminal() {
		t.Fatalf("expected ACCEPT ACK2, got %+v: %v", got, err)
	}
	if cfg.Ref.Channels == 0 && got.Flags != asp.NoReference ||
		cfg.Ref.Channels != 0 && got.Flags != 0 {
		t.Fatalf("incorrect no-REF ACK2 flags: %d", got.Flags)
	}
}

func (p *cfg2Pipe) exchange(t *testing.T, seq uint64, input asp.Input, inChunk int) asp.Reply {
	t.Helper()
	var packet bytes.Buffer
	if err := asp.WriteRequest(&packet, seq, input); err != nil {
		t.Fatal(err)
	}
	p.write(t, packet.Bytes(), inChunk)
	p.readDeadline(250 * time.Millisecond) // generous Go scheduler allowance
	start := time.Now()
	r, err := asp.ReadReply(p.stdout)
	if err != nil {
		t.Fatalf("RSP2: %v", err)
	}
	if r.Sequence != seq {
		t.Fatalf("incorrect sequence: %d != %d", r.Sequence, seq)
	}
	t.Logf("RSP2 read=%s; independent protocol smoke test, not the host exchange deadline", time.Since(start))
	return r
}

func (p *cfg2Pipe) rejectedExit10(t *testing.T) {
	t.Helper()
	_ = p.stdin.Close() // release incomplete input; avoid leaking a blocked child
	p.readDeadline(500 * time.Millisecond)
	_, err := asp.ReadResult(p.stdout)
	if err == nil {
		t.Fatal("invalid packet produced a valid ACK2")
	}
	select {
	case <-p.done:
		var status *exec.ExitError
		if !errors.As(p.exit, &status) || status.ExitCode() != 10 {
			t.Fatalf("expected protocol-reject exit 10, got %v", p.exit)
		}
	case <-time.After(time.Second):
		t.Fatal("invalid config child failed to terminate")
	}
}

func cfg2Wire(t *testing.T, cfg asp.CFG2) []byte {
	t.Helper()
	b, err := cfg.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func cfg2Gain(ch uint16, mask uint32, availability asp.FactoryGainAvailability) *asp.MicFactoryGain {
	g := &asp.MicFactoryGain{Channels: ch, Availability: availability}
	if availability != asp.FactoryGainNone {
		g.Encoding = asp.FactoryGainEncodingQ30
		g.Provenance = asp.FactoryGainProvenanceIDME
		g.PresentMask = mask
		for i := uint16(0); i < ch; i++ {
			if mask&(1<<i) != 0 {
				g.Coefficients[i] = 1000000000 + int64(i)*50000
			}
		}
	}
	return g
}

func TestSyntheticCFG2PipeGoldenAndPCMIdentity(t *testing.T) {
	// Six accepted C0-B fixture bytes also verified by the independent C0-C decoder.
	names := []string{
		"athena_4mic_2ref.cfg2",
		"athena_confirmed_complete_4mic.cfg2",
		"explicit_none_1mic.cfg2",
		"generic_1mic_noref.cfg2",
		"platform_only_1mic.cfg2",
		"unknown_inferred_partial_4mic.cfg2",
	}
	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			wire, err := os.ReadFile(filepath.Join("..", "..", "asp", "testdata", "cfg2_v2", name))
			if err != nil {
				t.Fatal(err)
			}
			cfg, err := asp.ParseCFG2(wire)
			if err != nil {
				t.Fatal(err)
			}
			p := newCFG2Pipe(t, "fragments")
			p.write(t, wire, 1) // deliberate one-byte CFG2 header/TLV fragmentation
			p.ack(t, cfg.Config)
			i := aspInput(cfg.Config)
			micBefore, refBefore := bytes.Clone(i.Mic), bytes.Clone(i.Ref)
			for seq := uint64(0); seq < 2; seq++ {
				r := p.exchange(t, seq, i, 4096)
				want := fixturePCM(nativeFrame{i.Mic, i.Ref, uint32(cfg.Config.Mic.FullBits)}, seq)
				if r.Type != asp.Data || r.Start != seq*asp.Samples || !slices.Equal(r.PCM, want) {
					t.Fatalf("PCM identity or origin mismatch: type=%d start=%d", r.Type, r.Start)
				}
			}
			if !bytes.Equal(i.Mic, micBefore) || !bytes.Equal(i.Ref, refBefore) {
				t.Fatal("metadata or IPC modified captured PCM")
			}
			t.Logf("C0-B/C0-C golden=%x metadata platform=%t gain=%t",
				sha256.Sum256(wire), cfg.PlatformInfo != nil, cfg.MicFactoryGain != nil)
		})
	}
}

func TestSyntheticCFG2PipeMetadataAdvisoryAndShapes(t *testing.T) {
	configs := []struct {
		name string
		cfg  asp.CFG2
	}{
		{"generic-absent", asp.CFG2{Config: aspShape(1, 0)}},
		{"generic-inferred", asp.CFG2{Config: aspShape(1, 0), PlatformInfo: &asp.PlatformInfo{Codename: "mystery", Detection: asp.PlatformInferred}}},
		{"generic-explicit-none", asp.CFG2{Config: aspShape(1, 0), MicFactoryGain: cfg2Gain(1, 0, asp.FactoryGainNone)}},
		{"factory-confirmed-absent-gain", asp.CFG2{Config: aspShape(4, 2), PlatformInfo: &asp.PlatformInfo{Codename: "athena", Detection: asp.PlatformConfirmed}}},
		{"factory-complete", asp.CFG2{Config: aspShape(4, 2), PlatformInfo: &asp.PlatformInfo{Codename: "athena", Detection: asp.PlatformConfirmed}, MicFactoryGain: cfg2Gain(4, 15, asp.FactoryGainComplete)}},
		{"factory-partial", asp.CFG2{Config: aspShape(4, 2), PlatformInfo: &asp.PlatformInfo{Codename: "athena", Detection: asp.PlatformConfirmed}, MicFactoryGain: cfg2Gain(4, 5, asp.FactoryGainPartial)}},
		{"factory-none", asp.CFG2{Config: aspShape(4, 2), PlatformInfo: &asp.PlatformInfo{Codename: "athena", Detection: asp.PlatformConfirmed}, MicFactoryGain: cfg2Gain(4, 0, asp.FactoryGainNone)}},
	}
	var canonical []int16
	for _, tc := range configs {
		t.Run(tc.name, func(t *testing.T) {
			p := newCFG2Pipe(t, "normal")
			p.write(t, cfg2Wire(t, tc.cfg), 3)
			p.ack(t, tc.cfg.Config)
			i := aspInput(tc.cfg.Config)
			r := p.exchange(t, 0, i, 1024)
			want := fixturePCM(nativeFrame{i.Mic, i.Ref, uint32(i.Config.Mic.FullBits)}, 0)
			if r.Type != asp.Data || !slices.Equal(r.PCM, want) || r.Start != 0 {
				t.Fatal("metadata affected generic DSP PCM or origin")
			}
			if tc.cfg.Config.Mic.Channels == 4 {
				if canonical == nil {
					canonical = slices.Clone(r.PCM)
				} else if !slices.Equal(canonical, r.PCM) {
					t.Fatal("factory metadata altered identical MIC/REF raw samples")
				}
			}
		})
	}
	// Descriptors, not codename, determine actual PCM bytes and rates.
	for _, c := range []asp.Config{
		aspShape(2, 0),
		aspShape(9, 8),
		{Flags: asp.Interleaved, Mic: asp.Stream{Channels: 1, Frames: 320, Rate: 16000, Format: asp.S16, FullBits: 15, ValidBits: 16}},
		{Flags: asp.Interleaved | asp.Synchronized | asp.HardwareRef,
			Mic: asp.Stream{Channels: 4, Frames: 960, Rate: 48000, Format: asp.S32, FullBits: 31, Alignment: asp.High, ValidBits: 24},
			Ref: asp.Stream{Channels: 1, Frames: 320, Rate: 16000, Format: asp.S16, FullBits: 15, ValidBits: 16}},
	} {
		t.Run(fmt.Sprint(c.Mic.Channels, "/", c.Mic.Rate, "/", c.Ref.Channels), func(t *testing.T) {
			p := newCFG2Pipe(t, "normal")
			p.write(t, cfg2Wire(t, asp.CFG2{Config: c}), 13)
			p.ack(t, c)
			i := aspInput(c)
			r := p.exchange(t, 13, i, 4096)
			if r.Type != asp.Data || r.Start != 13*asp.Samples ||
				!slices.Equal(r.PCM, fixturePCM(nativeFrame{i.Mic, i.Ref, uint32(c.Mic.FullBits)}, 13)) {
				t.Fatal("non-Athena native shape ignored")
			}
		})
	}
}

func appendCFG2TLV(wire []byte, typ, flags uint16, value []byte) []byte {
	result := bytes.Clone(wire)
	var h [8]byte
	binary.LittleEndian.PutUint16(h[0:2], typ)
	binary.LittleEndian.PutUint16(h[2:4], flags)
	binary.LittleEndian.PutUint32(h[4:8], uint32(len(value)))
	result = append(result, h[:]...)
	result = append(result, value...)
	binary.LittleEndian.PutUint32(result[8:12], uint32(len(result)))
	return result
}

func cfg2TLVOffset(wire []byte, typ uint16) int {
	for at := asp.CFG2HeaderBytes; at+8 <= len(wire); {
		kind := binary.LittleEndian.Uint16(wire[at:])
		if kind == typ {
			return at
		}
		at += 8 + int(binary.LittleEndian.Uint32(wire[at+4:]))
	}
	return -1
}

func TestSyntheticCFG2PipeStrictRejection(t *testing.T) {
	base := cfg2Wire(t, asp.CFG2{Config: aspShape(4, 2),
		PlatformInfo:   &asp.PlatformInfo{Codename: "athena", Detection: asp.PlatformConfirmed},
		MicFactoryGain: cfg2Gain(4, 15, asp.FactoryGainComplete)})
	mutate := func(f func([]byte)) []byte { b := bytes.Clone(base); f(b); return b }
	cases := []struct {
		name string
		wire []byte
	}{
		{"broken-magic", mutate(func(b []byte) { b[0] = 'X' })},
		{"wrong-version", mutate(func(b []byte) { b[4]++ })},
		{"legacy-header40", mutate(func(b []byte) { binary.LittleEndian.PutUint16(b[6:], 40) })},
		{"length15", mutate(func(b []byte) { binary.LittleEndian.PutUint32(b[8:], 15) })},
		{"length8193", mutate(func(b []byte) { binary.LittleEndian.PutUint32(b[8:], 8193) })},
		{"length-overflow", mutate(func(b []byte) { binary.LittleEndian.PutUint32(b[8:], ^uint32(0)) })},
		{"bad-stream-flags", mutate(func(b []byte) { b[12] = 255 })},
		{"bad-PCM-shape", mutate(func(b []byte) { b[asp.CFG2HeaderBytes+8+2]++ })},
		{"bad-stream-reserved", mutate(func(b []byte) { b[asp.CFG2HeaderBytes+8+12] = 1 })},
		{"bad-platform-schema", mutate(func(b []byte) { b[cfg2TLVOffset(b, 3)+8] = 2 })},
		{"bad-platform-codename", mutate(func(b []byte) { b[cfg2TLVOffset(b, 3)+8+4] = 'A' })},
		{"bad-gain-applied", mutate(func(b []byte) { b[cfg2TLVOffset(b, 0x101)+8+5] = 1 })},
		{"bad-gain-mask", mutate(func(b []byte) { b[cfg2TLVOffset(b, 0x101)+8+8] = 0 })},
		{"bad-known-TLV-length", mutate(func(b []byte) { binary.LittleEndian.PutUint32(b[cfg2TLVOffset(b, 3)+4:], 1000) })},
		{"required-unknown", appendCFG2TLV(base, 0x200, 1, []byte{1, 2, 3})},
		{"duplicate-typed", appendCFG2TLV(base, 0x101, 0, nil)},
		{"out-of-order", appendCFG2TLV(base, 3, 0, []byte{1, 1, 1, 0, 'x'})},
		{"unknown-flags", appendCFG2TLV(base, 0x200, 2, nil)},
		{"truncated-body", base[:len(base)-1]},
		{"truncated-header", base[:7]},
	}
	// MAX_TLVS=64: existing 4 + 61 unknown OPTIONAL TLVs -> 65.
	tooMany := bytes.Clone(base)
	for n := 0; n < 61; n++ {
		tooMany = appendCFG2TLV(tooMany, uint16(0x200+n), 0, nil)
	}
	cases = append(cases, struct {
		name string
		wire []byte
	}{"TLV-count65", tooMany})
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := newCFG2Pipe(t, "normal")
			p.write(t, tc.wire, 97)
			p.rejectedExit10(t)
		})
	}
	// Unknown OPTIONAL is accepted and skipped; original raw MIC still wins.
	t.Run("unknown-optional-accepted", func(t *testing.T) {
		cfg, err := asp.ParseCFG2(base)
		if err != nil {
			t.Fatal(err)
		}
		p := newCFG2Pipe(t, "normal")
		p.write(t, appendCFG2TLV(base, 0x200, 0, []byte{1, 2, 3}), 3)
		p.ack(t, cfg.Config)
		i := aspInput(cfg.Config)
		r := p.exchange(t, 0, i, 4096)
		if !slices.Equal(r.PCM, fixturePCM(nativeFrame{i.Mic, i.Ref, uint32(i.Config.Mic.FullBits)}, 0)) {
			t.Fatal("unknown optional affected PCM")
		}
	})
}

func TestSyntheticCFG2PipeFactoryCapabilities(t *testing.T) {
	for _, tc := range []struct {
		name   string
		cfg    asp.CFG2
		badMap bool
	}{
		{"missing-board", asp.CFG2{Config: aspShape(4, 2)}, false},
		{"inferred-athena", asp.CFG2{Config: aspShape(4, 2), PlatformInfo: &asp.PlatformInfo{Codename: "athena", Detection: asp.PlatformInferred}}, false},
		{"confirmed-other-board", asp.CFG2{Config: aspShape(4, 2), PlatformInfo: &asp.PlatformInfo{Codename: "cronos", Detection: asp.PlatformConfirmed}}, false},
		{"wrong-mic-shape", asp.CFG2{Config: aspShape(3, 2), PlatformInfo: &asp.PlatformInfo{Codename: "athena", Detection: asp.PlatformConfirmed}}, false},
		{"wrong-ref-shape", asp.CFG2{Config: aspShape(4, 0), PlatformInfo: &asp.PlatformInfo{Codename: "athena", Detection: asp.PlatformConfirmed}}, false},
		{"duplicate-mic-map", asp.CFG2{Config: aspShape(4, 2), PlatformInfo: &asp.PlatformInfo{Codename: "athena", Detection: asp.PlatformConfirmed}}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			env := []string{}
			if tc.badMap {
				env = append(env, "TECHO5_ASP_CHILD_BAD_MAP=1")
			}
			p := newCFG2Pipe(t, "factory", env...)
			p.write(t, cfg2Wire(t, tc.cfg), 7)
			p.readDeadline(250 * time.Millisecond)
			ack, err := asp.ReadResult(p.stdout)
			var terminal *asp.Terminal
			if !errors.As(err, &terminal) || ack.Type != asp.Reject || terminal.Status.Code != 522 {
				t.Fatalf("expected terminal factory mapping reject 522; ack=%+v err=%v", ack, err)
			}
		})
	}
	t.Run("confirmed-factory-not-ready-then-data", func(t *testing.T) {
		cfg := asp.CFG2{Config: aspShape(4, 2),
			PlatformInfo: &asp.PlatformInfo{Codename: "athena", Detection: asp.PlatformConfirmed}}
		p := newCFG2Pipe(t, "factory")
		p.write(t, cfg2Wire(t, cfg), 5)
		p.ack(t, cfg.Config)
		i := aspInput(cfg.Config)
		for seq := uint64(0); seq < 6; seq++ {
			r := p.exchange(t, seq, i, 8192)
			if seq == 0 {
				if r.Type != asp.NotReady || r.Start != asp.NoSample || r.PCM != nil {
					t.Fatal("factory 8ms quanta must preserve first-tick NotReady")
				}
			} else if r.Type != asp.Data || r.Start != (seq-1)*asp.Samples || len(r.PCM) != asp.Samples {
				t.Fatalf("factory stream origin/shape seq=%d reply=%+v", seq, r)
			}
		}
	})
}

func TestSyntheticCFG2PipeLegacyMismatchAndFaults(t *testing.T) {
	c := aspShape(4, 2)
	newWire := cfg2Wire(t, asp.CFG2{Config: c})
	p := newCFG2Pipe(t, "normal")
	p.write(t, rejectedCFG2(t), 40)
	p.rejectedExit10(t)

	for _, mode := range []string{"fragments", "sequence", "origin", "late-not-ready", "fatal", "crash", "output-hang"} {
		t.Run(mode, func(t *testing.T) {
			p := newCFG2Pipe(t, mode)
			p.write(t, newWire, 3)
			p.ack(t, c)
			i := aspInput(c)
			var request bytes.Buffer
			if err := asp.WriteRequest(&request, 0, i); err != nil {
				t.Fatal(err)
			}
			p.write(t, request.Bytes(), 8192)
			if mode == "output-hang" {
				// Deliberately inject a short local read deadline to exercise
				// io deadline propagation; production host uses 5 seconds.
				p.readDeadline(25 * time.Millisecond)
				_, err := asp.ReadReply(p.stdout)
				if !errors.Is(err, os.ErrDeadlineExceeded) {
					t.Fatalf("expected injected 25ms read timeout, got %v", err)
				}
				return
			}
			p.readDeadline(250 * time.Millisecond)
			r, err := asp.ReadReply(p.stdout)
			switch mode {
			case "crash":
				if err == nil {
					t.Fatal("crashed child produced valid RSP2")
				}
			case "fatal":
				var term *asp.Terminal
				if !errors.As(err, &term) || term.Status.Code != 503 {
					t.Fatalf("terminal status lost: %+v %v", r, err)
				}
			case "sequence":
				if err != nil || r.Sequence == 0 {
					t.Fatalf("injected stale sequence not observable: %+v %v", r, err)
				}
			case "origin":
				if err != nil || r.Start != asp.Samples {
					t.Fatalf("injected invalid origin not observable: %+v %v", r, err)
				}
			case "late-not-ready":
				if err != nil || r.Type != asp.NotReady || r.Start != asp.NoSample || r.PCM != nil {
					t.Fatalf("NotReady envelope corrupted: %+v %v", r, err)
				}
			default:
				if err != nil || r.Type != asp.Data || r.Sequence != 0 || r.Start != 0 {
					t.Fatalf("fragmented RSP2 corrupted: %+v %v", r, err)
				}
			}
		})
	}
	for _, invalid := range []struct {
		name           string
		micLen, refLen uint32
	}{
		{"wrong-mic-length", 1, uint32(aspShape(4, 2).Ref.Frames) * 4},
		{"wrong-ref-length", 30720, 1},
	} {
		t.Run(invalid.name, func(t *testing.T) {
			p := newCFG2Pipe(t, "normal")
			p.write(t, newWire, 11)
			p.ack(t, c)
			var hdr [20]byte
			copy(hdr[:], "INP2")
			binary.LittleEndian.PutUint32(hdr[12:16], invalid.micLen)
			binary.LittleEndian.PutUint32(hdr[16:20], invalid.refLen)
			p.write(t, hdr[:], 2)
			_ = p.stdin.Close()
			p.readDeadline(250 * time.Millisecond)
			if _, err := asp.ReadReply(p.stdout); err == nil {
				t.Fatal("wrong request PCM size generated RSP2")
			}
		})
	}
}

func TestSyntheticCFG2PipeMaxEnvelopeAndStatus(t *testing.T) {
	base := cfg2Wire(t, asp.CFG2{Config: aspShape(1, 0)})
	maxValue := make([]byte, asp.CFG2MaxBytes-len(base)-8)
	maxPacket := appendCFG2TLV(base, 0x200, 0, maxValue)
	if len(maxPacket) != asp.CFG2MaxBytes {
		t.Fatal("incorrect 8192-byte boundary fixture")
	}
	t.Run("8192-byte-optional-accepted", func(t *testing.T) {
		p := newCFG2Pipe(t, "normal")
		p.write(t, maxPacket, 127)
		p.ack(t, aspShape(1, 0))
		i := aspInput(aspShape(1, 0))
		r := p.exchange(t, 0, i, 256)
		if r.Type != asp.Data || r.Start != 0 ||
			!slices.Equal(r.PCM, fixturePCM(nativeFrame{i.Mic, i.Ref, uint32(i.Config.Mic.FullBits)}, 0)) {
			t.Fatal("maximum-sized optional TLV changed PCM")
		}
	})
	t.Run("8193-byte-rejected", func(t *testing.T) {
		p := newCFG2Pipe(t, "normal")
		packet := appendCFG2TLV(maxPacket, 0x201, 0, []byte{0})
		p.write(t, packet[:16], 4) // invalid total is rejected before allocating or waiting for body
		p.rejectedExit10(t)
	})
	t.Run("ACK2-and-RSP2-status", func(t *testing.T) {
		p := newCFG2Pipe(t, "normal",
			"TECHO5_ASP_ACK_KIND=1", "TECHO5_ASP_ACK_CODE=200",
			"TECHO5_ASP_ACK_MSG=Ready",
			"TECHO5_ASP_REPLY_KIND=1", "TECHO5_ASP_REPLY_CODE=301",
			"TECHO5_ASP_REPLY_MSG=Noise suppressor disabled")
		p.write(t, base, 1)
		p.readDeadline(250 * time.Millisecond)
		ack, err := asp.ReadResult(p.stdout)
		if err != nil || ack.Type != asp.Accept || ack.Flags != asp.NoReference ||
			ack.Status.Kind != asp.State || ack.Status.Code != 200 || ack.Status.Message != "Ready" {
			t.Fatalf("ACK2 status not retained: %+v %v", ack, err)
		}
		i := aspInput(aspShape(1, 0))
		r := p.exchange(t, 0, i, 2048)
		if r.Type != asp.Data || r.Status.Kind != asp.State || r.Status.Code != 301 ||
			r.Status.Message != "Noise suppressor disabled" {
			t.Fatalf("RSP2 status not retained: %+v", r)
		}
	})
}
