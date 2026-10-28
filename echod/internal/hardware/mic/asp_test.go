package mic

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"os/exec"
	"reflect"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/HuskerMinion/techo5/echod/internal/asp"
	"github.com/HuskerMinion/techo5/echod/internal/asp/synthetic"
	"github.com/HuskerMinion/techo5/echod/internal/config"
	"github.com/HuskerMinion/techo5/echod/internal/lib/alsa"
	"github.com/HuskerMinion/techo5/echod/internal/lib/spectrum"
)

// TestASPChild is a standalone synthetic executable when launched with this environment.
// It opens no hardware, files, sockets or services. stdout contains ONLY the binary protocol.
func TestASPChild(t *testing.T) {
	mode := os.Getenv("TECHO5_ASP_CHILD")
	if mode == "" {
		return
	}
	if mode == "start-hang" {
		time.Sleep(time.Hour)
		os.Exit(0)
	}
	if mode == "start-bad" {
		os.Stdout.Write([]byte("NOPE"))
		os.Exit(0)
	}
	os.Stdout.Write([]byte("ASP2"))
	cfg2, err := asp.ReadCFG2(os.Stdin)
	c := cfg2.Config
	if err != nil {
		os.Exit(10)
	}
	envStatus := func(prefix string) asp.Status {
		code, _ := strconv.Atoi(os.Getenv(prefix + "_CODE"))
		kind, _ := strconv.Atoi(os.Getenv(prefix + "_KIND"))
		msg := os.Getenv(prefix + "_MSG")
		if kind == 0 && (code != 0 || msg != "") {
			kind = asp.State
		}
		return asp.Status{Kind: uint8(kind), Code: uint16(code), Message: msg}
	}
	if mode == "config-hang" {
		time.Sleep(time.Hour)
		os.Exit(0)
	}
	if mode == "config-crash" {
		os.Exit(17)
	}
	ack := asp.ConfigResult{Type: asp.Accept, Status: envStatus("TECHO5_ASP_ACK")}
	if c.Ref.Channels == 0 {
		ack.Flags = asp.NoReference
	}
	if mode == "no-ref-reject" {
		ack.Flags = 0
	}
	var emu *synthetic.Emulator
	if strings.HasPrefix(mode, "factory") {
		mapping := synthetic.Mapping{}
		if os.Getenv("TECHO5_ASP_CHILD_BAD_MAP") == "1" {
			mapping.Mic = []int{0, 0, 2, 3} // intentional duplicate for negative IPC test
		}
		a, e := synthetic.New(c, synthetic.Profile{Mic: 4, Ref: 2, FactoryShape: true, GeometryValidated: true, IdentityValidated: true}, mapping)
		// The factory-shaped TEST profile is stricter than generic synthetic:
		// a codename never substitutes for actual PCM shape or confirmed board.
		if cfg2.PlatformInfo == nil || cfg2.PlatformInfo.Detection != asp.PlatformConfirmed || cfg2.PlatformInfo.Codename != "athena" {
			e = synthetic.ErrMapping
		}
		if e != nil {
			ack.Type = asp.Reject
			ack.Status = asp.Status{Kind: asp.State, Code: 522, Message: "Required microphone mapping unavailable"}
		} else {
			emu = &synthetic.Emulator{Adapter: a, Markers: false}
		}
	}
	if mode == "reject" {
		ack.Type = asp.Reject
	}
	// H1-B2b test-only ACK corruption and split-delay fixtures.
	if d, _ := strconv.Atoi(os.Getenv("TECHO5_ASP_CONFIG_DELAY_MS")); d > 0 {
		time.Sleep(time.Duration(d) * time.Millisecond)
	}
	var ackBytes bytes.Buffer
	asp.WriteResult(&ackBytes, ack)
	ackWire := ackBytes.Bytes()
	switch mode {
	case "ack-magic":
		ackWire[0] = 'X'
	case "ack-type":
		ackWire[4] = 99
	case "ack-ref":
		ackWire[10] ^= asp.NoReference
	case "ack-flags":
		ackWire[10] = 2
	case "ack-short":
		ackWire = ackWire[:7]
	case "ack-large":
		binary.LittleEndian.PutUint16(ackWire[8:], 161)
	}
	os.Stdout.Write(ackWire)
	if mode == "ack-short" {
		os.Exit(0)
	}
	if ack.Type == asp.Reject || ack.Status.Terminal() {
		time.Sleep(time.Hour)
		os.Exit(0)
	}
	if mode == "input-close" {
		os.Stdin.Close()
		time.Sleep(time.Hour)
		os.Exit(0)
	}
	if mode == "input-hang" {
		time.Sleep(time.Hour)
		os.Exit(0)
	}
	mb, rb, _ := c.Lengths()
	i := asp.Input{Config: c, Mic: make([]byte, mb), Ref: make([]byte, rb)}
	frames := 0
	for {
		fault := mode
		if strings.HasPrefix(mode, "runtime:") {
			if frames >= 4 {
				fault = strings.TrimPrefix(mode, "runtime:")
			} else {
				fault = "normal"
			}
		}
		if strings.HasPrefix(mode, "factory:") {
			parts := strings.Split(mode, ":")
			at, _ := strconv.Atoi(parts[2])
			fault = "factory"
			if frames == at {
				fault = parts[1]
			}
		}
		if fault == "input-hang" {
			time.Sleep(time.Hour)
			os.Exit(0)
		}
		if fault == "input-close" {
			os.Stdin.Close()
			time.Sleep(time.Hour)
			os.Exit(0)
		}
		seq, e := asp.ReadRequest(os.Stdin, c, i.Mic, i.Ref)
		if e != nil {
			os.Exit(0)
		}
		frames++
		if emu != nil {
			d, _ := strconv.Atoi(os.Getenv("TECHO5_ASP_QUANTUM_DELAY_MS"))
			if d > 0 {
				quanta := 2
				if seq%2 == 1 {
					quanta = 3
				}
				time.Sleep(time.Duration(d*quanta) * time.Millisecond)
			}
		}
		if strings.HasPrefix(fault, "delay:") {
			d, _ := strconv.Atoi(strings.TrimPrefix(fault, "delay:"))
			time.Sleep(time.Duration(d) * time.Millisecond)
		}
		if fault == "jitter" {
			time.Sleep([]time.Duration{0, 12 * time.Millisecond, 18 * time.Millisecond, 0, 16 * time.Millisecond}[seq%5])
		}
		if os.Getenv("TECHO5_ASP_ACK") == "1" {
			syscall.Write(3, []byte{1})
		}
		if fault == "crash" {
			os.Exit(17)
		}
		if fault == "output-hang" {
			time.Sleep(time.Hour)
			os.Exit(0)
		}
		r := asp.Reply{Type: asp.Data, Sequence: seq, Start: seq * FrameSamples, PCM: fixturePCM(nativeFrame{i.Mic, i.Ref, uint32(c.Mic.FullBits)}, seq)}
		if emu != nil {
			r, e = emu.Step(i, seq)
			if e != nil {
				os.Exit(19)
			}
		}
		if fault == "silence" {
			clear(r.PCM)
		}
		r.Status = envStatus("TECHO5_ASP_REPLY")
		if at := os.Getenv("TECHO5_ASP_REPLY_ON"); at != "" {
			n, _ := strconv.Atoi(at)
			if seq != uint64(n) {
				r.Status = asp.Status{}
			}
		}
		if typ := os.Getenv("TECHO5_ASP_REPLY_TYPE"); typ != "" && r.Status.Present() {
			n, _ := strconv.Atoi(typ)
			r.Type = uint8(n)
		}
		if fault == "sequence" || fault == "repeat" && seq > 0 {
			r.Sequence = seq - 1
		}
		if fault == "origin" {
			r.Start += FrameSamples
		}
		if fault == "origin-repeat" {
			r.Start -= FrameSamples
		}
		if fault == "fatal" {
			r.Status = asp.Status{Kind: asp.Event, Code: 503, Message: "ASP processing failed"}
		}

		if fault == "late-not-ready" {
			r.Type = asp.NotReady
			r.Start = asp.NoSample
			r.PCM = nil
		}
		var b bytes.Buffer
		asp.WriteReply(&b, r)
		out := b.Bytes()
		if fault == "bad-status" {
			out[5] = asp.State
			binary.LittleEndian.PutUint16(out[8:], 1)
			out = append(out[:32], 0xff)
		}
		if fault == "large-status" {
			binary.LittleEndian.PutUint16(out[8:], 161)
		}
		if fault == "bad-kind" {
			out[5] = 99
		}
		if fault == "bad-code" {
			out[5] = asp.State
			binary.LittleEndian.PutUint16(out[6:], 199)
		}
		if fault == "truncated-status" {
			out = out[:15]
		}
		if fault == "length" {
			binary.LittleEndian.PutUint32(out[28:], ^uint32(0))
		}
		if fault == "short" || fault == "short-open" {
			out = out[:min(len(out), 32+101)]
		}
		if fault == "fragments" {
			for at := 0; at < len(out); {
				end := min(at+7, len(out))
				os.Stdout.Write(out[at:end])
				at = end
			}
		} else {
			os.Stdout.Write(out)
		}
		if fault == "short-open" {
			time.Sleep(time.Hour)
			os.Exit(0)
		}
		if fault == "truncated-status" || fault == "bad-status" || fault == "large-status" || fault == "short" || fault == "once" || fault == "length" || fault == "sequence" {
			os.Exit(0)
		}
	}
}

func fixturePCM(n nativeFrame, seq uint64) []int16 {
	c := crc32.NewIEEE()
	c.Write(n.mics)
	c.Write(n.ref)
	seed := c.Sum32() + n.fullBits + uint32(seq)*997
	out := make([]int16, FrameSamples)
	for i := range out {
		out[i] = int16(seed + uint32(i)*37)
	}
	return out
}

func nativePattern(bits uint32, seed int) nativeFrame {
	n := nativeFrame{mics: make([]byte, nativeMicBytes), ref: make([]byte, nativeRefBytes), fullBits: bits}
	values := []int32{-1, 0, 1, -1 << (bits), int32((uint32(1) << bits) - 1)}
	for i := 0; i < 1920; i++ {
		for ch := 0; ch < 4; ch++ {
			binary.LittleEndian.PutUint32(n.mics[i*16+ch*4:], uint32(values[(i+ch+seed)%len(values)]))
		}
		binary.LittleEndian.PutUint16(n.ref[i*4:], uint16(int16(i*17+seed)))
		binary.LittleEndian.PutUint16(n.ref[i*4+2:], uint16(int16(-i*31-seed-123)))
	}
	return n
}

func childCommand(mode string) *exec.Cmd {
	exe, err := os.Executable()
	if err != nil {
		panic(err)
	}
	cmd := exec.Command(exe, "-test.run=^TestASPChild$")
	cmd.Env = append(os.Environ(), "TECHO5_ASP_CHILD="+mode, "GORACE=atexit_sleep_ms=0")
	return cmd
}

func fixture(t testing.TB, mode string, configs ...asp.Config) *externalASP {
	t.Helper()
	f, err := startTestASP(context.Background(), childCommand(mode), configs...)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		f.Close()
		select {
		case <-f.done:
		case <-time.After(3 * time.Second):
			t.Error("child not reaped")
		}
	})
	return f
}

func TestASPNativePipes(t *testing.T) {
	for _, bits := range []uint32{23, 31} {
		for _, mode := range []string{"normal", "fragments"} {
			t.Run(fmt.Sprint(bits, "/", mode), func(t *testing.T) {
				f := fixture(t, mode, nativeASPConfig(bits))
				var held [][]int16
				for tick := 0; tick < 8; tick++ {
					n := nativePattern(bits, tick)
					beforeMic, beforeRef := bytes.Clone(n.mics), bytes.Clone(n.ref)
					out, err := f.process(n)
					if err != nil {
						t.Fatal(err)
					}
					if !slices.Equal(out, fixturePCM(n, uint64(tick))) {
						t.Fatal("channel/scale/sequence corruption")
					}
					if !bytes.Equal(n.mics, beforeMic) || !bytes.Equal(n.ref, beforeRef) {
						t.Fatal("native mutated")
					}
					for _, old := range held {
						if &old[0] == &out[0] {
							t.Fatal("reused output storage")
						}
					}
					held = append(held, out)
				}
			})
		}
	}
}

func TestASPFaultBounded(t *testing.T) {
	for _, mode := range []string{"input-close", "input-hang", "output-hang", "crash", "short", "short-open", "length", "sequence"} {
		t.Run(mode, func(t *testing.T) {
			f := fixture(t, mode)
			start := time.Now()
			out, err := f.process(nativePattern(31, 0))
			elapsed := time.Since(start)
			if err == nil || out != nil || !f.broken.Load() {
				t.Fatal("bad output accepted")
			}
			// This tolerance is a host scheduling guard, NOT a device realtime claim.
			if elapsed > 250*time.Millisecond {
				t.Fatalf("stall %s", elapsed)
			}
			start = time.Now()
			out, err = f.process(nativePattern(31, 1))
			if out != nil || err == nil || time.Since(start) > 50*time.Millisecond {
				t.Fatal("failed child reused or blocked")
			}
			t.Logf("failure duration: %s (configured budget %s)", elapsed, aspFrameBudget)
		})
	}
}

func TestASPStartupFailure(t *testing.T) {
	for _, mode := range []string{"start-bad", "start-hang"} {
		start := time.Now()
		f, err := startTestASP(context.Background(), childCommand(mode))
		if err == nil || f != nil {
			t.Fatal("invalid startup accepted")
		}
		if time.Since(start) > 2*time.Second {
			t.Fatal("startup unbounded")
		}
	}
	if f, err := startTestASP(context.Background(), exec.Command("/nonexistent-external-asp")); err == nil || f != nil {
		t.Fatal("missing helper accepted")
	}
}

func TestASPRepeatedReplyRejected(t *testing.T) {
	f := fixture(t, "repeat", nativeASPConfig(23))
	n := nativePattern(23, 0)
	if _, err := f.process(n); err != nil {
		t.Fatal(err)
	}
	if out, err := f.process(n); err == nil || out != nil {
		t.Fatal("stale response accepted")
	}
}

func TestASPMissingNative(t *testing.T) {
	for _, n := range []nativeFrame{{}, {mics: make([]byte, nativeMicBytes), fullBits: 31}, {mics: make([]byte, nativeMicBytes), ref: make([]byte, nativeRefBytes), fullBits: 24}} {
		f := fixture(t, "normal")
		if out, err := f.process(n); err == nil || out != nil {
			t.Fatal("unsupported native accepted")
		}
	}
}

// Selection after all stock processors must leave every analytical result identical.
func TestASPFinalSelectionAndSensors(t *testing.T) {
	stock, ext := testSource(), testSource()
	for _, s := range []*Source{stock, ext} {
		s.leveling.Store(true)
		s.denoising.Store(true)
		s.canceling = true
		s.SetNoise(true)
		s.SetStandout(true)
	}
	attachReadyASP(t, ext, fixture(t, "normal"))
	a, sa := ext.Listen("MWW")
	defer sa()
	b, sb := ext.Listen("STT")
	defer sb()
	c, sc := ext.Listen("call/talkback/announce")
	defer sc()
	rawCh, sr := ext.ListenRaw()
	defer sr()
	for tick := 0; tick < 8; tick++ {
		stock.Facing()
		ext.Facing()
		raw := rawFrame(tick, true)
		n := nativePattern(31, tick)
		stock.broadcast(bytes.Clone(raw))
		ext.broadcastNative(bytes.Clone(raw), n)
		want := fixturePCM(n, uint64(tick))
		x, y, z := <-a, <-b, <-c
		if !slices.Equal(x, want) || !slices.Equal(y, want) || !slices.Equal(z, want) || &x[0] != &y[0] || &x[0] != &z[0] {
			t.Fatal("selected stream differs between listeners or processed twice")
		}
		if got := ext.Recent(20 * time.Millisecond); !slices.Equal(got, want) {
			t.Fatal("history differs")
		}
		if got := <-rawCh; !bytes.Equal(got, raw) {
			t.Fatal("ListenRaw changed")
		}
		if ext.Gain() != stock.Gain() || ext.Floor() != stock.Floor() || ext.Level() != stock.Level() || ext.Clipped() != stock.Clipped() || ext.Standout(time.Second) != stock.Standout(time.Second) || ext.facing.Load() != stock.facing.Load() || ext.ERLE() != stock.ERLE() || ext.Canceling() != stock.Canceling() {
			t.Fatal("stock analytics changed")
		}
		na, ma := stock.Noise()
		nb, mb := ext.Noise()
		if na != nb || ma != mb {
			t.Fatal("noise changed")
		}
	}
	// The stock processor's evolving internals have run on exactly the same frames.
	if !reflect.DeepEqual(ext.leveler.lp, stock.leveler.lp) || ext.standout.floor != stock.standout.floor {
		t.Fatal("stock state diverged")
	}
	frames, pre, stop := ext.ListenHistory("history join", time.Second)
	defer stop()
	if len(pre) != 8*FrameSamples {
		t.Fatal("history purged")
	}
	ext.closeASP()
	n := nativePattern(31, 9)
	raw := rawFrame(9, false)
	stock.broadcast(bytes.Clone(raw))
	ext.broadcastNative(bytes.Clone(raw), n)
	if !slices.Equal(<-a, stock.Recent(20*time.Millisecond)) {
		t.Fatal("disable fallback changed stock")
	}
	<-b
	<-c
	<-rawCh
	if !slices.Equal(<-frames, ext.Recent(20*time.Millisecond)) {
		t.Fatal("ListenHistory join changed")
	}
	if len(ext.listeners) != 4 {
		t.Fatal("subscriptions replaced")
	}
}

func TestASPFaultFallsBackSameTickAndReconnect(t *testing.T) {
	for _, mode := range []string{"crash", "output-hang", "short", "length", "sequence"} {
		t.Run(mode, func(t *testing.T) {
			ext, stock := testSource(), testSource()
			attachReadyASP(t, ext, fixture(t, mode))
			frames, stop := ext.Listen("existing")
			defer stop()
			for i := 0; i < 2; i++ {
				raw := rawFrame(i, false)
				stock.broadcast(bytes.Clone(raw))
				ext.broadcastNative(bytes.Clone(raw), nativePattern(31, i))
				if !slices.Equal(<-frames, stock.Recent(20*time.Millisecond)) {
					t.Fatal("fallback not current stock")
				}
			}
			if ext.asp != nil {
				t.Fatal("failed child retained")
			}
			attachReadyASP(t, ext, fixture(t, "normal"))
			n := nativePattern(31, 3)
			ext.broadcastNative(rawFrame(3, false), n)
			if !slices.Equal(<-frames, fixturePCM(n, 0)) {
				t.Fatal("reconnect not fresh")
			}
			if len(ext.Recent(time.Second)) != 3*FrameSamples || len(ext.listeners) != 1 {
				t.Fatal("history/subscriber reset")
			}
		})
	}
}

// Spectrum must see the selected frame as well, at the same single broadcast point.
func TestASPSpectrumSelected(t *testing.T) {
	old := spectrum.Mic
	defer func() { spectrum.Mic = old }()
	got, want := spectrum.New(Rate, 512), spectrum.New(Rate, 512)
	now := time.Now()
	got.Bands(8, now)
	want.Bands(8, now)
	spectrum.Mic = got
	s := testSource()
	attachReadyASP(t, s, fixture(t, "normal"))
	n := nativePattern(31, 2)
	s.broadcastNative(rawFrame(0, false), n)
	want.Push(fixturePCM(n, 0))
	g, _ := got.Bands(8, now)
	w, _ := want.Bands(8, now)
	if !slices.Equal(g, w) {
		t.Fatal("spectrum did not receive selected PCM")
	}
}

type scriptedNativeCapture struct {
	tick    int
	cancel  context.CancelFunc
	overrun bool
}

func (c *scriptedNativeCapture) Read(p []byte) (int, error) {
	if c.tick == 4 {
		c.cancel()
		return 0, io.EOF
	}
	c.tick++
	if c.overrun && c.tick == 2 {
		return 0, alsa.ErrOverrun
	}
	return copy(p, rawFrame(c.tick, false)), nil
}
func (c *scriptedNativeCapture) nativeInput() nativeFrame { return nativePattern(31, c.tick) }
func (*scriptedNativeCapture) Close() error               { return nil }

func TestASPRunUsesCurrentNativeOnce(t *testing.T) {
	for _, overrun := range []bool{false, true} {
		s := testSource()
		f := fixture(t, "normal")
		attachReadyASP(t, s, f)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		s.pcm = &scriptedNativeCapture{cancel: cancel, overrun: overrun}
		ch, stop := s.Listen("one")
		defer stop()
		if err := s.Run(ctx); !errors.Is(err, io.EOF) {
			t.Fatal(err)
		}
		seq := uint64(0)
		for tick := 1; tick <= 4; tick++ {
			if overrun && tick == 2 {
				continue
			}
			if !slices.Equal(<-ch, fixturePCM(nativePattern(31, tick), seq)) {
				t.Fatal("wrong native capture tick")
			}
			seq++
		}
		select {
		case <-ch:
			t.Fatal("duplicate frame")
		default:
		}
		if s.asp != nil || !f.broken.Load() {
			t.Fatal("Run did not close asp")
		}
	}
}

type chunkWriter struct {
	bytes.Buffer
	chunk int
	zero  bool
}

func (w *chunkWriter) Write(p []byte) (int, error) {
	if w.zero {
		return 0, nil
	}
	return w.Buffer.Write(p[:min(w.chunk, len(p))])
}
func TestASPExactWrites(t *testing.T) {
	p := nativePattern(31, 2).mics
	for _, chunk := range []int{1, 7, 4096} {
		w := &chunkWriter{chunk: chunk}
		if err := writeExact(w, p); err != nil || !bytes.Equal(w.Bytes(), p) {
			t.Fatal("partial write lost bytes")
		}
	}
	if !errors.Is(writeExact(&chunkWriter{zero: true}, p), io.ErrShortWrite) {
		t.Fatal("zero progress accepted")
	}
}

func TestASPCloseInterruptsAndNoLeaks(t *testing.T) {
	fdBefore, _ := os.ReadDir("/proc/self/fd")
	gBefore := runtime.NumGoroutine()
	for i := 0; i < 8; i++ {
		f := fixture(t, "output-hang")
		done := make(chan error, 1)
		go func() { _, err := f.process(nativePattern(31, i)); done <- err }()
		f.Close()
		select {
		case err := <-done:
			if err == nil {
				t.Fatal("close delivered PCM")
			}
		case <-time.After(time.Second):
			t.Fatal("Close did not interrupt")
		}
		select {
		case <-f.done:
		case <-time.After(3 * time.Second):
			t.Fatal("child leaked")
		}
	}
	runtime.GC()
	fdAfter, _ := os.ReadDir("/proc/self/fd")
	if len(fdAfter) > len(fdBefore)+2 {
		t.Fatalf("fd leak %d -> %d", len(fdBefore), len(fdAfter))
	}
	if runtime.NumGoroutine() > gBefore+2 {
		t.Fatalf("goroutine leak %d -> %d", gBefore, runtime.NumGoroutine())
	}
}

func BenchmarkASPPipes(b *testing.B) {
	f := fixture(b, "normal")
	n := nativePattern(31, 0)
	b.ReportAllocs()
	b.SetBytes(nativeMicBytes + nativeRefBytes + FrameSamples*2)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := f.process(n); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkASPStockBroadcast(b *testing.B) {
	s := testSource()
	raw := rawFrame(0, false)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.broadcast(raw)
	}
}

// Deterministic proof of the 8-to-20 causal boundary: no zeros, dropped samples or duplicate output.
// This is a test framer, NOT a factory adapter. One capture-tick delay is necessary for this model.
func TestASPEightToTwentyFIFO(t *testing.T) {
	var nativePending int
	var output []int
	next := 0
	for tick := 0; tick < 100; tick++ {
		nativePending += 1920
		for nativePending >= 768 {
			nativePending -= 768
			for i := 0; i < 128; i++ {
				output = append(output, next)
				next++
			}
		}
		if tick == 0 {
			if len(output) != 256 {
				t.Fatal("first tick cannot produce 320")
			}
			continue
		}
		for i, v := range output[:320] {
			if v != (tick-1)*320+i {
				t.Fatal("FIFO reordered/invented samples")
			}
		}
		output = output[320:]
		if len(output) > 320 || nativePending >= 768 {
			t.Fatal("unbounded FIFO")
		}
	}
	if next != 32000 || len(output) != 320 || nativePending != 0 {
		t.Fatal("sample conservation failed")
	}
}

func TestASPHostTiming(t *testing.T) {
	if os.Getenv("TECHO5_ASP_METRICS") == "" {
		t.Skip("explicit measurement only")
	}
	f := fixture(t, "normal")
	n := nativePattern(31, 0)
	var ns []int64
	var before, after syscall.Rusage
	syscall.Getrusage(syscall.RUSAGE_SELF, &before)
	childStat, _ := os.ReadFile(fmt.Sprintf("/proc/%d/stat", f.cmd.Process.Pid))
	for i := 0; i < 1000; i++ {
		start := time.Now()
		if _, err := f.process(n); err != nil {
			t.Fatal(err)
		}
		ns = append(ns, time.Since(start).Nanoseconds())
	}
	syscall.Getrusage(syscall.RUSAGE_SELF, &after)
	childEnd, _ := os.ReadFile(fmt.Sprintf("/proc/%d/stat", f.cmd.Process.Pid))
	rss, _ := os.ReadFile(fmt.Sprintf("/proc/%d/status", f.cmd.Process.Pid))
	selfRss, _ := os.ReadFile("/proc/self/status")
	ticks := func(p []byte) int64 {
		x := strings.Fields(string(p))
		a, _ := strconv.ParseInt(x[13], 10, 64)
		b, _ := strconv.ParseInt(x[14], 10, 64)
		return a + b
	}
	micros := func(x syscall.Timeval) int64 { return int64(x.Sec)*1000000 + int64(x.Usec) }
	t.Logf("self cpu microseconds=%d; child cpu clock ticks=%d over 1000 frames", micros(after.Utime)+micros(after.Stime)-micros(before.Utime)-micros(before.Stime), ticks(childEnd)-ticks(childStat))
	for _, entry := range []struct {
		name string
		data []byte
	}{{"child", rss}, {"self", selfRss}} {
		for _, line := range strings.Split(string(entry.data), "\n") {
			if strings.HasPrefix(line, "VmRSS:") || strings.HasPrefix(line, "VmHWM:") {
				t.Log(entry.name, line)
			}
		}
	}
	slices.Sort(ns)
	t.Logf("pipe ns p50=%d p95=%d p99=%d max=%d count=%s", ns[500], ns[950], ns[990], ns[999], strconv.Itoa(len(ns)))
}

func attachReadyASP(t testing.TB, s *Source, f *externalASP) {
	ctx, cancel := context.WithCancel(context.Background())
	s.aspMu.Lock()
	s.asp = f
	s.aspAttempt = &aspAttempt{ctx: ctx, cancel: cancel, keep: true, good: aspPreflightFrames}
	s.aspStatus = ASPStatus{Selected: config.ASPExternal, Active: config.ASPBuiltin, Debug: "External ASP starting"}
	s.aspMu.Unlock()
	t.Cleanup(s.closeASP)
}

// Test setup always freezes configuration before starting the supervised child.
func startTestASP(ctx context.Context, cmd *exec.Cmd, configs ...asp.Config) (*externalASP, error) {
	c := nativeASPConfig(31)
	if len(configs) > 0 {
		c = configs[0]
	}
	session, err := buildSessionCFG2(c, confirmedFactoryAthena(), nil)
	if err != nil {
		return nil, err
	}
	return startExternalASP(ctx, cmd, c, session)
}
func rejectedCFG2(t testing.TB) []byte {
	t.Helper()
	b, err := os.ReadFile("../../asp/testdata/rejected_fixed_cfg2.bin")
	if err != nil {
		t.Fatal(err)
	}
	return b
}
