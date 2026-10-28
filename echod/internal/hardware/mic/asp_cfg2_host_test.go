package mic

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/HuskerMinion/techo5/echod/internal/asp"
	"github.com/HuskerMinion/techo5/echod/internal/config"
)

func hostCFG2Session(t testing.TB, c asp.Config, p *asp.PlatformInfo, g *asp.MicFactoryGain) sessionCFG2 {
	t.Helper()
	s, err := buildSessionCFG2(c, p, g)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func hostCFG2Command(mode string, env ...string) *exec.Cmd {
	cmd := childCommand(mode)
	cmd.Env = append(cmd.Env, env...)
	return cmd
}

func hostCFG2Reaped(t testing.TB, f *externalASP) {
	t.Helper()
	f.Close()
	select {
	case <-f.done:
	case <-time.After(3 * time.Second):
		t.Error("child/reaper leaked")
	}
}

func hostCFG2Fixture(t testing.TB, mode string, c asp.Config, session sessionCFG2, env ...string) *externalASP {
	t.Helper()
	f, err := startExternalASP(context.Background(), hostCFG2Command(mode, env...), c, session)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { hostCFG2Reaped(t, f) })
	return f
}

// Real HELLO/CFG2/ACK2/INP2/RSP2 pipes with the production 25ms deadline.
func TestHostCFG2MetadataPCM(t *testing.T) {
	for _, bits := range []uint32{23, 31} {
		c := nativePattern(bits, 0).input().Config
		for _, avail := range []asp.FactoryGainAvailability{asp.FactoryGainComplete, asp.FactoryGainPartial, asp.FactoryGainNone} {
			t.Run(fmt.Sprint(bits, "/", avail), func(t *testing.T) {
				p := confirmedFactoryAthena()
				mask := uint32(15)
				if avail == asp.FactoryGainPartial {
					mask = 5
				}
				if avail == asp.FactoryGainNone {
					mask = 0
				}
				g := cfg2Gain(4, mask, avail)
				session := hostCFG2Session(t, c, p, g)
				before := session.Bytes()
				f := hostCFG2Fixture(t, "fragments", c, session,
					"TECHO5_ASP_ACK_KIND=1", "TECHO5_ASP_ACK_CODE=200", "TECHO5_ASP_ACK_MSG=Ready",
					"TECHO5_ASP_REPLY_KIND=1", "TECHO5_ASP_REPLY_CODE=301")
				// Mutating caller metadata and an owned copy after startup cannot
				// change what the host sends on its first exchange.
				p.Codename = "other"
				g.Coefficients[0] = 0
				copyBytes := session.Bytes()
				copyBytes[0] = 0
				var held [][]int16
				for seq := uint64(0); seq < 8; seq++ {
					n := nativePattern(bits, int(seq))
					m, r := bytes.Clone(n.mics), bytes.Clone(n.ref)
					got, err := f.exchange(n.input())
					if err != nil || got.Type != asp.Data || got.Sequence != seq || got.Start != seq*FrameSamples ||
						!slices.Equal(got.PCM, fixturePCM(n, seq)) {
						t.Fatalf("PCM/sequence/origin: seq=%d err=%v", seq, err)
					}
					if seq == 0 && (got.ConfigStatus.Code != 200 || got.ConfigStatus.Message != "Ready") ||
						seq > 0 && got.ConfigStatus.Present() || got.Status.Code != 301 {
						t.Fatal("feedback lost", got)
					}
					if !bytes.Equal(m, n.mics) || !bytes.Equal(r, n.ref) {
						t.Fatal("raw MIC/REF changed")
					}
					for _, old := range held {
						if &old[0] == &got.PCM[0] {
							t.Fatal("borrowed output storage")
						}
					}
					held = append(held, got.PCM)
				}
				if !bytes.Equal(before, session.Bytes()) || !bytes.Equal(before, f.session.Bytes()) {
					t.Fatal("snapshot mutated")
				}
			})
		}
	}
	for _, p := range []*asp.PlatformInfo{nil, {Codename: "athena", Detection: asp.PlatformInferred}, {Codename: "future", Detection: asp.PlatformInferred}} {
		c := genericSessionConfig()
		session := hostCFG2Session(t, c, p, nil)
		f := hostCFG2Fixture(t, "normal", c, session)
		for seq := uint64(0); seq < 3; seq++ {
			i := aspInput(c)
			got, err := f.exchange(i)
			if err != nil || got.Start != seq*FrameSamples || !slices.Equal(got.PCM, fixturePCM(nativeFrame{i.Mic, i.Ref, uint32(c.Mic.FullBits)}, seq)) {
				t.Fatal("generic absent/inferred failed", err)
			}
		}
	}
}

func TestHostCFG2SnapshotRejectedBeforeLaunch(t *testing.T) {
	c := nativePattern(31, 0).input().Config
	good := hostCFG2Session(t, c, confirmedFactoryAthena(), nil)
	change := func(fn func([]byte)) sessionCFG2 { b := good.Bytes(); fn(b); return sessionCFG2{packet: string(b)} }
	stale := nativePattern(23, 0).input().Config
	cases := []struct {
		name string
		c    asp.Config
		s    sessionCFG2
	}{
		{"zero", c, sessionCFG2{}},
		{"stale", stale, good},
		{"truncated", c, sessionCFG2{packet: good.packet[:len(good.packet)-1]}},
		{"oversized", c, sessionCFG2{packet: strings.Repeat("x", 8193)}},
		{"TLV-overflow", c, change(func(b []byte) { binary.LittleEndian.PutUint32(b[20:], ^uint32(0)) })},
		{"bad-platform", c, change(func(b []byte) { b[cfg2TLVOffset(b, 3)+12] = 'A' })},
	}
	// Codec-valid gain on an inferred platform violates host qualification.
	wire := cfg2Wire(t, asp.CFG2{Config: c, PlatformInfo: &asp.PlatformInfo{Codename: "athena", Detection: asp.PlatformInferred}, MicFactoryGain: cfg2Gain(4, 15, asp.FactoryGainComplete)})
	cases = append(cases, struct {
		name string
		c    asp.Config
		s    sessionCFG2
	}{"unqualified-gain", c, sessionCFG2{packet: string(wire)}})
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := hostCFG2Command("normal")
			f, err := startExternalASP(context.Background(), cmd, tc.c, tc.s)
			if !errors.Is(err, errASPInvalid) || f != nil || cmd.Process != nil {
				t.Fatal("invalid snapshot launched child", err)
			}
		})
	}
	for _, after := range []bool{false, true} {
		t.Run(fmt.Sprint("native-change/", after), func(t *testing.T) {
			f := hostCFG2Fixture(t, "normal", c, good)
			if after {
				if _, err := f.exchange(nativePattern(31, 0).input()); err != nil {
					t.Fatal(err)
				}
			}
			if r, err := f.exchange(nativePattern(23, 1).input()); !errors.Is(err, errASPReference) || r.PCM != nil || !f.broken.Load() {
				t.Fatal("changed native accepted", err)
			}
		})
	}
}

func hostCFG2Source(t *testing.T, mode string, env ...string) *Source {
	t.Helper()
	return hostCFG2SourceConfig(t, mode, nativePattern(31, 0).input().Config, env...)
}

func hostCFG2SourceConfig(t *testing.T, mode string, c asp.Config, env ...string) *Source {
	t.Helper()
	s := controlSource(t, "normal", config.ASPExternal)
	var mu sync.Mutex
	var children []*externalASP
	s.aspStart = func(ctx context.Context) (*externalASP, error) {
		session, err := prepareASP2SessionCFG2WithReaders(c, confirmedFactoryAthena,
			func(c asp.Config, p *asp.PlatformInfo) *asp.MicFactoryGain {
				return cfg2Gain(4, 15, asp.FactoryGainComplete)
			})
		if err != nil {
			return nil, err
		}
		f, err := startExternalASP(ctx, hostCFG2Command(mode, env...), c, session)
		if f != nil {
			mu.Lock()
			children = append(children, f)
			mu.Unlock()
		}
		return f, err
	}
	t.Cleanup(func() {
		s.SetASP(config.ASPBuiltin)
		until := time.Now().Add(3 * time.Second)
		for {
			s.aspMu.Lock()
			running := s.aspStarting
			s.aspMu.Unlock()
			if !running {
				break
			}
			if time.Now().After(until) {
				t.Error("startup worker leaked")
				return
			}
			runtime.Gosched()
		}
		mu.Lock()
		defer mu.Unlock()
		for _, f := range children {
			hostCFG2Reaped(t, f)
		}
	})
	return s
}

func TestHostCFG2PreflightAndFactoryOrigin(t *testing.T) {
	for _, mode := range []string{"normal", "factory"} {
		t.Run(mode, func(t *testing.T) {
			s := hostCFG2SourceConfig(t, mode, nativePattern(23, 0).input().Config)
			s.mixer = Center{}
			s.SetASP(config.ASPExternal)
			waitASPChild(t, s)
			ch, stop := s.Listen("CFG2 preflight")
			defer stop()
			activate := 3
			if mode == "factory" {
				activate = 4
			}
			for tick := 0; tick < 12; tick++ {
				raw, n := markerInput(tick)
				s.broadcastNative(raw, n)
				got := <-ch
				if tick < activate {
					checkMarker(t, got, uint64(tick*320))
				} else if mode == "factory" {
					checkMarker(t, got, uint64((tick-1)*320))
				} else if !slices.Equal(got, fixturePCM(n, uint64(tick))) {
					t.Fatal("activation sample")
				}
				st := s.ASPStatus()
				if (st.Active == config.ASPExternal) != (tick >= activate) || st.Confirmed != (tick >= activate) {
					t.Fatal("preflight false activation", tick, st)
				}
			}
			if s.aspAttempt.good != 3 || len(s.Recent(time.Second)) != 12*320 || s.Dropped() != 0 {
				t.Fatal("preflight/history")
			}
			s.SetASP(config.ASPBuiltin)
			raw, n := markerInput(12)
			s.broadcastNative(raw, n)
			checkMarker(t, <-ch, 12*320)
		})
	}
}

func TestHostCFG2FaultFallback(t *testing.T) {
	for _, mode := range []string{
		"reject", "ack-magic", "ack-type", "ack-ref", "ack-flags", "ack-short", "ack-large", "config-hang", "config-crash",
		"input-close", "input-hang", "output-hang", "crash", "short", "short-open", "length", "sequence", "origin",
		"truncated-status", "bad-status", "large-status", "bad-kind", "bad-code", "fatal",
		"runtime:sequence", "runtime:origin-repeat", "runtime:late-not-ready", "runtime:crash", "runtime:output-hang",
		"runtime:short", "runtime:bad-status",
	} {
		t.Run(mode, func(t *testing.T) {
			s := hostCFG2Source(t, mode)
			s.SetASP(config.ASPExternal)
			waitASPChild(t, s)
			f := s.asp
			ch, stop := s.Listen("fallback")
			defer stop()
			stock := testSource()
			for tick := 0; tick < 7; tick++ {
				raw := rawFrame(tick, false)
				stock.broadcast(bytes.Clone(raw))
				s.broadcastNative(bytes.Clone(raw), nativePattern(31, tick))
				got := <-ch
				failing := !strings.HasPrefix(mode, "runtime:") || tick >= 4
				if failing || tick < 3 {
					if !slices.Equal(got, stock.Recent(20*time.Millisecond)) {
						t.Fatal("fallback not current stock", tick)
					}
				}
				if failing && s.ASPStatus().Active != config.ASPBuiltin {
					t.Fatal("false active")
				}
			}
			st := s.ASPStatus()
			if !strings.Contains(st.Debug, "Fallback:") && !strings.Contains(st.Debug, "503:") {
				t.Fatal("missing failure", st)
			}
			if s.asp != nil || !f.broken.Load() || st.Confirmed != strings.HasPrefix(mode, "runtime:") {
				t.Fatal("supervision", st)
			}
			hostCFG2Reaped(t, f)
			s.SetASP(config.ASPExternal)
			if s.ASPStatus() != st {
				t.Fatal("implicit restart")
			}
		})
	}
	t.Run("terminal-ACK-status", func(t *testing.T) {
		s := hostCFG2Source(t, "normal", "TECHO5_ASP_ACK_KIND=1", "TECHO5_ASP_ACK_CODE=522", "TECHO5_ASP_ACK_MSG=Required mapping unavailable")
		s.SetASP(config.ASPExternal)
		waitASPChild(t, s)
		f := s.asp
		s.broadcastNative(rawFrame(0, false), nativePattern(31, 0))
		st := s.ASPStatus()
		if st.Active != config.ASPBuiltin || st.Confirmed || !strings.HasPrefix(st.Debug, "522:") {
			t.Fatal(st)
		}
		hostCFG2Reaped(t, f)
	})
	t.Run("source-timeline", func(t *testing.T) {
		s := hostCFG2Source(t, "normal")
		s.SetASP(config.ASPExternal)
		waitASPChild(t, s)
		f := s.asp
		s.broadcastNative(rawFrame(0, false), nativePattern(31, 0))
		s.aspMu.Lock()
		s.aspTicks += 3
		s.aspMu.Unlock()
		stock := testSource()
		raw := rawFrame(1, false)
		stock.broadcast(bytes.Clone(raw))
		s.broadcastNative(raw, nativePattern(31, 1))
		if !slices.Equal(s.Recent(20*time.Millisecond), stock.Recent(20*time.Millisecond)) ||
			s.ASPStatus().Active != config.ASPBuiltin || !strings.Contains(s.ASPStatus().Debug, "timeline") {
			t.Fatal("source timeline accepted", s.ASPStatus())
		}
		hostCFG2Reaped(t, f)
	})
}

func TestHostCFG2StartupReadersAndCancellation(t *testing.T) {
	s := hostCFG2Source(t, "normal")
	c := nativePattern(31, 0).input().Config
	entered, release := make(chan struct{}), make(chan struct{})
	var platformCalls, gainCalls atomic.Int32
	s.aspStart = func(ctx context.Context) (*externalASP, error) {
		session, err := prepareASP2SessionCFG2WithReaders(c, func() *asp.PlatformInfo {
			platformCalls.Add(1)
			close(entered)
			<-release
			return confirmedFactoryAthena()
		}, func(asp.Config, *asp.PlatformInfo) *asp.MicFactoryGain { gainCalls.Add(1); return nil })
		if err != nil {
			return nil, err
		}
		return startExternalASP(ctx, hostCFG2Command("normal"), c, session)
	}
	s.SetASP(config.ASPExternal)
	<-entered
	ch, stop := s.Listen("stock during metadata")
	defer stop()
	s.mixer = Center{}
	for tick := 0; tick < 5; tick++ {
		raw, n := markerInput(tick)
		s.broadcastNative(raw, n)
		checkMarker(t, <-ch, uint64(tick*320))
	}
	s.SetASP(config.ASPBuiltin)
	close(release)
	until := time.Now().Add(time.Second)
	for {
		s.aspMu.Lock()
		running := s.aspStarting
		s.aspMu.Unlock()
		if !running {
			break
		}
		if time.Now().After(until) {
			t.Fatal("startup cancellation stuck")
		}
		runtime.Gosched()
	}
	if platformCalls.Load() != 1 || gainCalls.Load() != 1 || s.ASPStatus().Active != config.ASPBuiltin || s.asp != nil {
		t.Fatal("canceled snapshot attached")
	}
	// Invalid metadata fails in startup before any child or capture exchange.
	s.aspStart = func(ctx context.Context) (*externalASP, error) {
		session, err := prepareASP2SessionCFG2WithReaders(c, func() *asp.PlatformInfo { return &asp.PlatformInfo{Codename: "Bad", Detection: asp.PlatformConfirmed} },
			func(asp.Config, *asp.PlatformInfo) *asp.MicFactoryGain {
				t.Error("invalid platform consulted gains")
				return nil
			})
		if err != nil {
			return nil, fmt.Errorf("%w: invalid metadata", errASPInvalid)
		}
		return startExternalASP(ctx, hostCFG2Command("normal"), c, session)
	}
	s.SetASP(config.ASPExternal)
	st := waitASP(t, s, func(st ASPStatus) bool { return strings.Contains(st.Debug, "Fallback:") })
	if st.Active != config.ASPBuiltin || st.Confirmed {
		t.Fatal(st)
	}
}

func TestHostCFG2OneShotReadersNoHotPath(t *testing.T) {
	s := hostCFG2Source(t, "normal")
	c := nativePattern(31, 0).input().Config
	var pc, gc atomic.Int32
	var child *externalASP
	s.aspStart = func(ctx context.Context) (*externalASP, error) {
		session, err := prepareASP2SessionCFG2WithReaders(c, func() *asp.PlatformInfo { pc.Add(1); return confirmedFactoryAthena() },
			func(asp.Config, *asp.PlatformInfo) *asp.MicFactoryGain {
				gc.Add(1)
				return cfg2Gain(4, 5, asp.FactoryGainPartial)
			})
		if err != nil {
			return nil, err
		}
		child, err = startExternalASP(ctx, hostCFG2Command("normal"), c, session)
		return child, err
	}
	s.SetASP(config.ASPExternal)
	waitASPChild(t, s)
	for tick := 0; tick < 10; tick++ {
		s.broadcastNative(rawFrame(tick, false), nativePattern(31, tick))
	}
	if pc.Load() != 1 || gc.Load() != 1 || s.ASPStatus().Active != config.ASPExternal {
		t.Fatal("collectors repeated in capture")
	}
	s.SetASP(config.ASPBuiltin)
	hostCFG2Reaped(t, child)
}

func TestHostCFG2CancelAndFreshEpoch(t *testing.T) {
	s := hostCFG2Source(t, "normal")
	s.SetASP(config.ASPExternal)
	waitASPChild(t, s)
	f := s.asp
	n := nativePattern(31, 0)
	old := s.processASP(n)
	if old.err != nil {
		t.Fatal(old.err)
	}
	s.SetASP(config.ASPBuiltin)
	hostCFG2Reaped(t, f)
	s.SetASP(config.ASPExternal)
	waitASPChild(t, s)
	before := s.ASPStatus()
	fresh := s.aspAttempt
	stock := make([]int16, 320)
	s.aspMu.Lock()
	got := s.chooseASP(stock, old)
	s.aspMu.Unlock()
	s.retireASP(old)
	if !slices.Equal(got, stock) || s.ASPStatus() != before || fresh.next != 0 || fresh.good != 0 || fresh.based {
		t.Fatal("stale epoch escaped")
	}
	for tick := 0; tick < 5; tick++ {
		s.broadcastNative(rawFrame(tick, false), nativePattern(31, tick))
	}
	if s.ASPStatus().Active != config.ASPExternal {
		t.Fatal("fresh epoch never activated")
	}
}

func TestHostCFG2CloseCancelNoLeaks(t *testing.T) {
	fdBefore, _ := os.ReadDir("/proc/self/fd")
	gBefore := runtime.NumGoroutine()
	for tick := 0; tick < 8; tick++ {
		c := nativePattern(31, 0).input().Config
		f := hostCFG2Fixture(t, "output-hang", c, hostCFG2Session(t, c, nil, nil))
		done := make(chan error, 1)
		go func() { _, err := f.exchange(nativePattern(31, tick).input()); done <- err }()
		f.Close()
		select {
		case err := <-done:
			if err == nil {
				t.Fatal("canceled output")
			}
		case <-time.After(time.Second):
			t.Fatal("capture stuck")
		}
		hostCFG2Reaped(t, f)
	}
	runtime.GC()
	fdAfter, _ := os.ReadDir("/proc/self/fd")
	if len(fdAfter) > len(fdBefore)+2 || runtime.NumGoroutine() > gBefore+2 {
		t.Fatal("pipe/goroutine leak", len(fdBefore), len(fdAfter))
	}
}

// Virtual cumulative pipe costs make each phase independently deterministic.
// A partial Write/Read never resets the deadline; both pipes must receive the
// exact same absolute timestamp once per exchange.
type hostCFG2Clock struct {
	spent                         time.Duration
	cost                          [4]time.Duration // config write, ACK read, input write, reply read
	deadline                      time.Time
	writeDeadlines, readDeadlines int
	closedIn, closedOut           bool
}

func (c *hostCFG2Clock) step(phase int) error {
	c.spent += c.cost[phase]
	if c.spent >= aspFrameBudget {
		return os.ErrDeadlineExceeded
	}
	return nil
}

type hostCFG2Writer struct {
	c                       *hostCFG2Clock
	wire                    bytes.Buffer
	configBytes, inputBytes int
}

func (w *hostCFG2Writer) SetWriteDeadline(d time.Time) error {
	w.c.writeDeadlines++
	w.c.deadline = d
	w.c.spent = 0
	return nil
}
func (w *hostCFG2Writer) Write(p []byte) (int, error) {
	off := w.wire.Len()
	if off == 0 {
		if err := w.c.step(0); err != nil {
			return 0, err
		}
	}
	if off >= w.configBytes && (off-w.configBytes)%w.inputBytes == 0 {
		if err := w.c.step(2); err != nil {
			return 0, err
		}
	}
	// Fragment both the CFG2 packet and raw PCM writes deliberately.
	n := min(7, len(p))
	if off < w.configBytes {
		n = min(n, w.configBytes-off)
	}
	return w.wire.Write(p[:n])
}
func (w *hostCFG2Writer) Close() error { w.c.closedIn = true; return nil }

type hostCFG2Reader struct {
	c    *hostCFG2Clock
	wire []byte
	at   int
}

func (r *hostCFG2Reader) SetReadDeadline(d time.Time) error {
	r.c.readDeadlines++
	if d != r.c.deadline {
		return errors.New("different read/write absolute deadline")
	}
	return nil
}
func (r *hostCFG2Reader) Read(p []byte) (int, error) {
	if r.at == 0 {
		if err := r.c.step(1); err != nil {
			return 0, err
		}
	}
	if r.at >= 12 && (r.at-12)%(32+640) == 0 {
		if err := r.c.step(3); err != nil {
			return 0, err
		}
	}
	if r.at == len(r.wire) {
		return 0, io.EOF
	}
	n := min(3, len(p), len(r.wire)-r.at)
	copy(p, r.wire[r.at:r.at+n])
	r.at += n
	return n, nil
}
func (r *hostCFG2Reader) Close() error { r.c.closedOut = true; return nil }

func TestHostCFG2AbsoluteDeadlineDeterministic(t *testing.T) {
	for _, tc := range []struct {
		name string
		cost [4]time.Duration
		fail bool
	}{
		{"config-write-hang", [4]time.Duration{26, 0, 0, 0}, true},
		{"ACK-hang", [4]time.Duration{0, 26, 0, 0}, true},
		{"input-write-hang", [4]time.Duration{0, 0, 26, 0}, true},
		{"output-read-hang", [4]time.Duration{0, 0, 0, 26}, true},
		{"split-cost-no-reset", [4]time.Duration{7, 7, 7, 7}, true},
		{"partial-IO-within-budget", [4]time.Duration{6, 6, 6, 6}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := nativePattern(31, 0).input().Config
			session := hostCFG2Session(t, c, confirmedFactoryAthena(), nil)
			clock := &hostCFG2Clock{}
			for j, v := range tc.cost {
				clock.cost[j] = v * time.Millisecond
			}
			w := &hostCFG2Writer{c: clock, configBytes: len(session.packet), inputBytes: 20 + nativeMicBytes + nativeRefBytes}
			var output bytes.Buffer
			asp.WriteResult(&output, asp.ConfigResult{Type: asp.Accept})
			for seq := uint64(0); seq < 2; seq++ {
				asp.WriteReply(&output, asp.Reply{Type: asp.Data, Sequence: seq, Start: seq * 320, PCM: make([]int16, 320)})
			}
			r := &hostCFG2Reader{c: clock, wire: output.Bytes()}
			f := &externalASP{in: w, out: r, session: session, config: c}
			got, err := f.exchange(nativePattern(31, 0).input())
			if clock.readDeadlines != 1 || clock.writeDeadlines != 1 {
				t.Fatal("deadline reset per syscall/phase", clock)
			}
			if tc.fail {
				if !errors.Is(err, errASPTimeout) || got.PCM != nil || !clock.closedIn || !clock.closedOut || !f.broken.Load() {
					t.Fatal("unbounded or nonterminal deadline", err)
				}
			} else {
				if err != nil || !bytes.Equal(w.wire.Bytes()[:len(session.packet)], session.Bytes()) {
					t.Fatal("partial CFG2 writes", err)
				}
				clock.cost = [4]time.Duration{0, 0, 10 * time.Millisecond, 16 * time.Millisecond}
				got, err = f.exchange(nativePattern(31, 1).input())
				if !errors.Is(err, errASPTimeout) || got.PCM != nil || clock.readDeadlines != 2 || clock.writeDeadlines != 2 {
					t.Fatal("later exchange not fresh 25ms", err)
				}
			}
			t.Logf("deterministic total=%s deadline=%s write/read deadline calls=%d/%d terminal=%t", clock.spent, aspFrameBudget, clock.writeDeadlines, clock.readDeadlines, f.broken.Load())
		})
	}
}

func TestHostCFG2AbsoluteDeadlineRealPipes(t *testing.T) {
	c := nativePattern(31, 0).input().Config
	session := hostCFG2Session(t, c, nil, nil)
	// Both individual delays fit 25ms; their sum cannot fit the first exchange.
	f := hostCFG2Fixture(t, "delay:18", c, session, "TECHO5_ASP_CONFIG_DELAY_MS=18")
	start := time.Now()
	got, err := f.exchange(nativePattern(31, 0).input())
	elapsed := time.Since(start)
	if !errors.Is(err, errASPTimeout) || got.PCM != nil || !f.broken.Load() {
		t.Fatal("first deadline reset after ACK", err)
	}
	t.Logf("first real config18ms+reply18ms elapsed=%s, enforced budget=25ms; elapsed is scheduler dependent", elapsed)
}

func TestHostCFG2ControlCancelPipe(t *testing.T) {
	s := hostCFG2Source(t, "normal")
	ack, childAck, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer ack.Close()
	defer childAck.Close()
	c := nativePattern(31, 0).input().Config
	session := hostCFG2Session(t, c, nil, nil)
	var child *externalASP
	s.aspStart = func(ctx context.Context) (*externalASP, error) {
		cmd := hostCFG2Command("output-hang", "TECHO5_ASP_ACK=1")
		cmd.ExtraFiles = []*os.File{childAck}
		var err error
		child, err = startExternalASP(ctx, cmd, c, session)
		return child, err
	}
	s.SetASP(config.ASPExternal)
	waitASPChild(t, s)
	done := make(chan struct{})
	go func() { s.broadcastNative(rawFrame(0, false), nativePattern(31, 0)); close(done) }()
	ack.SetReadDeadline(time.Now().Add(time.Second))
	var b [1]byte
	if _, err := io.ReadFull(ack, b[:]); err != nil {
		t.Fatal("no native input acknowledgement", err)
	}
	if !s.mu.TryLock() {
		t.Fatal("Source lock held during exchange")
	}
	s.mu.Unlock()
	s.SetASP(config.ASPBuiltin)
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("cancel did not interrupt PCM exchange")
	}
	if st := s.ASPStatus(); st.Active != config.ASPBuiltin || st.Debug != "000: Built-in processing OK" {
		t.Fatal("stale failure replaced manual selection", st)
	}
	hostCFG2Reaped(t, child)
}

func TestHostCFG2StartupContext(t *testing.T) {
	c := nativePattern(31, 0).input().Config
	session := hostCFG2Session(t, c, nil, nil)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	cmd := hostCFG2Command("normal")
	if f, err := startExternalASP(ctx, cmd, c, session); f != nil || !errors.Is(err, context.Canceled) || cmd.Process != nil {
		t.Fatal("canceled launch", err)
	}
	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() {
		_, err := startExternalASP(ctx, hostCFG2Command("start-hang"), c, session)
		done <- err
	}()
	// The cancellation must interrupt either scheduling/startup or HELLO.
	cancel()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("canceled HELLO accepted")
		}
	case <-time.After(time.Second):
		t.Fatal("HELLO cancellation unbounded")
	}
}

func TestHostCFG2RepeatedSourceCancellation(t *testing.T) {
	s := hostCFG2Source(t, "normal")
	for cycle := 0; cycle < 10; cycle++ {
		s.SetASP(config.ASPExternal)
		waitASPChild(t, s)
		f := s.asp
		for tick := 0; tick < 5; tick++ {
			s.broadcastNative(rawFrame(tick, false), nativePattern(31, tick))
		}
		if s.ASPStatus().Active != config.ASPExternal {
			t.Fatal("fresh epoch never active", cycle)
		}
		s.SetASP(config.ASPBuiltin)
		hostCFG2Reaped(t, f)
	}
	// Concurrent HA commands/status snapshots exercise the existing single
	// startup worker and attempt cancellation with CFG2 children under -race.
	var wg sync.WaitGroup
	for worker := 0; worker < 3; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				s.SetASP(config.ASPBuiltin)
				s.SetASP(config.ASPExternal)
				_ = s.ASPStatus()
			}
		}()
	}
	wg.Wait()
	s.SetASP(config.ASPBuiltin)
}
