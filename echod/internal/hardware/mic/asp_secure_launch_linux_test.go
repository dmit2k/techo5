//go:build linux

package mic

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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

func buildASPHostProbe(t *testing.T, marker int) aspProbe {
	t.Helper()
	path := filepath.Join(t.TempDir(), "host-probe")
	out, err := exec.Command("cc", "-std=c11", "-O2", "-Wall", "-Wextra", "-Werror", fmt.Sprintf("-DCHILD_MARKER=%d", marker), "testdata/asp_host_probe.c", "-o", path).CombinedOutput()
	if err != nil {
		t.Fatalf("native independent IPC ELF: %v: %s", err, out)
	}
	if dir := os.Getenv("TECHO5_FD_EVIDENCE_DIR"); dir != "" {
		b, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, fmt.Sprintf("host-probe-%d.elf", marker)), b, 0700); err != nil {
			t.Fatal(err)
		}
	}
	return aspProbe{path: path}
}

func hostProbeFixture(t *testing.T, probe aspProbe, mode, proto string, asset int) *aspSecureFixture {
	t.Helper()
	return newASPProbeFixture(t, probe, fmt.Sprintf("%s %s %d", mode, proto, asset))
}

func probeHostState(t *testing.T, bits uint32) *externalASP {
	t.Helper()
	f := &externalASP{config: nativeASPConfig(bits)}
	f.session = hostCFG2Session(t, f.config, confirmedFactoryAthena(), nil)
	return f
}

func probeHostPCM(n nativeFrame, seq uint64, marker, asset int) []int16 {
	p := fixturePCM(n, seq)
	for i := range p {
		p[i] += int16(marker + asset)
	}
	return p
}

func waitPinnedHost(t *testing.T, f *externalASP) {
	t.Helper()
	f.Close()
	select {
	case <-f.done:
	case <-time.After(time.Second):
		t.Fatal("single host reaper did not finish")
	}
	assertASPReaped(t, f.cmd)
}

func startProbeCandidate(t *testing.T, fixture *aspSecureFixture, bits uint32) *externalASP {
	t.Helper()
	c := discoverASPCandidate(fixture.open)
	if c.identity == (aspPinnedIdentity{}) || !haveExternalASPCandidate("athena", c) {
		t.Fatal("qualified synthetic candidate missing")
	}
	f, err := startASPCandidate(context.Background(), c, probeHostState(t, bits), fixture.open)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { waitPinnedHost(t, f) })
	return f
}

func TestASPSecureLaunchNativeHost(t *testing.T) {
	probe := buildASPHostProbe(t, 17)
	// Transport-looking ENV and loader overrides must not reach the child or
	// select a host mode; the child independently demands exactly LANG/LC_ALL.
	t.Setenv("LD_PRELOAD", "/offline-missing-preload.so")
	t.Setenv("LD_LIBRARY_PATH", "/offline-missing-libs")
	t.Setenv("TECHO5_ASP_CHILD_PROTO", "cfg2-tlv")
	for _, bits := range []uint32{23, 31} {
		for _, proto := range []string{"cfg2"} {
			t.Run(fmt.Sprint(bits, "/", proto), func(t *testing.T) {
				fixture := hostProbeFixture(t, probe, "normal", proto, 71)
				c := discoverASPCandidate(fixture.open)
				var pin *secureASPPinned
				f, err := startASPCandidate(context.Background(), c, probeHostState(t, bits), func() (*secureASPPinned, error) { var e error; pin, e = fixture.open(); return pin, e })
				if err != nil {
					t.Fatal(err)
				}
				defer waitPinnedHost(t, f)
				assertASPConsumed(t, pin)
				if f.cmd.Path != "/proc/self/fd/3" || len(f.cmd.ExtraFiles) != 0 {
					t.Fatal("lost pinned launch or explicit transport")
				}
				for seq := uint64(0); seq < 8; seq++ {
					n := nativePattern(bits, int(seq))
					m, r := bytes.Clone(n.mics), bytes.Clone(n.ref)
					got, e := f.process(n)
					if e != nil || !slices.Equal(got, probeHostPCM(n, seq, 17, 71)) {
						t.Fatalf("IPC/inode/assets/raw PCM seq=%d: %v", seq, e)
					}
					if !bytes.Equal(n.mics, m) || !bytes.Equal(n.ref, r) {
						t.Fatal("raw MIC/REF changed")
					}
				}
				t.Log("real FD3 ELF + FD4 profile: HELLO, exact CFG, ACK, 8 INP/RSP, raw MIC/REF identity PASS")
			})
		}
	}
}

func TestASPSecureLaunchDiscoveryPolicy(t *testing.T) {
	probe := buildASPHostProbe(t, 17)
	fixture := hostProbeFixture(t, probe, "normal", "cfg2", 71)
	c := discoverASPCandidate(fixture.open)
	if !c.operational() || !haveExternalASPCandidate("athena", c) || haveExternalASPCandidate("rook", c) {
		t.Fatal("secure Athena helper not made available")
	}
	if e := os.Chmod(fixture.part(4), 0777); e != nil {
		t.Fatal(e)
	}
	if discoverASPCandidate(fixture.open).operational() {
		t.Fatal("unsafe helper advertised")
	}
	if discoverASPCandidate(func() (*secureASPPinned, error) { return nil, os.ErrNotExist }).operational() {
		t.Fatal("unavailable helper advertised")
	}
}

func TestASPSecureLaunchDiscoveryReplacement(t *testing.T) {
	original, decoy := buildASPHostProbe(t, 17), buildASPHostProbe(t, 333)
	// The decoy is a valid native child, distinguishable by both code and assets.
	d := hostProbeFixture(t, decoy, "normal", "cfg2", 900)
	f := startProbeCandidate(t, d, 31)
	n := nativePattern(31, 0)
	if got, e := f.process(n); e != nil || !slices.Equal(got, probeHostPCM(n, 0, 333, 900)) {
		t.Fatal("decoy control invalid", e)
	}
	for _, part := range []int{0, 1, 2, 4} {
		t.Run(fmt.Sprint(part), func(t *testing.T) {
			fixture := hostProbeFixture(t, original, "normal", "cfg2", 71)
			c := discoverASPCandidate(fixture.open)
			if e := os.Rename(fixture.part(part), fixture.part(part)+".original"); e != nil {
				t.Fatal(e)
			}
			replacement := hostProbeFixture(t, decoy, "normal", "cfg2", 900)
			if e := os.Rename(replacement.part(part), fixture.part(part)); e != nil {
				t.Fatal(e)
			}
			var pin *secureASPPinned
			child, e := startASPCandidate(context.Background(), c, probeHostState(t, 31), func() (*secureASPPinned, error) { var err error; pin, err = fixture.open(); return pin, err })
			if child != nil || !errors.Is(e, errASPStart) {
				t.Fatal("discovery name replacement executed", e)
			}
			assertASPConsumed(t, pin)
		})
	}
}

func TestASPSecureLaunchSwapAfterPin(t *testing.T) {
	original, decoy := buildASPHostProbe(t, 17), buildASPHostProbe(t, 333)
	for _, part := range []int{0, 1, 2, 4} {
		t.Run(fmt.Sprint(part), func(t *testing.T) {
			fixture := hostProbeFixture(t, original, "normal", "cfg2", 71)
			c := discoverASPCandidate(fixture.open)
			var pin *secureASPPinned
			child, e := startASPCandidate(context.Background(), c, probeHostState(t, 31), func() (*secureASPPinned, error) {
				p, e := fixture.open()
				if e != nil {
					return nil, e
				}
				pin = p
				if e = os.Rename(fixture.part(part), fixture.part(part)+".original"); e != nil {
					return p, e
				}
				replacement := hostProbeFixture(t, decoy, "normal", "cfg2", 900)
				return p, os.Rename(replacement.part(part), fixture.part(part))
			})
			if e != nil {
				t.Fatal(e)
			}
			defer waitPinnedHost(t, child)
			assertASPConsumed(t, pin)
			for seq := uint64(0); seq < 4; seq++ {
				n := nativePattern(31, int(seq))
				if got, e := child.process(n); e != nil || !slices.Equal(got, probeHostPCM(n, seq, 17, 71)) {
					t.Fatal("post-pin decoy executed or asset changed", e)
				}
			}
		})
	}
}

func TestASPSecureLaunchFormatsAndFaults(t *testing.T) {
	probe := buildASPHostProbe(t, 17)
	for _, tc := range []struct {
		mode, proto string
		want        error
	}{
		{"normal", "cfg2", nil},
		{"ack-bad", "cfg2", asp.ErrStatus}, {"reject", "cfg2", nil},
		{"config-hang", "cfg2", errASPTimeout}, {"response-hang", "cfg2", errASPTimeout},
		{"closed-reader", "cfg2", errASPExited}, {"sequence", "cfg2", errASPInvalid},
		{"origin", "cfg2", errASPInvalid}, {"fatal", "cfg2", nil},
	} {
		t.Run(fmt.Sprint(tc.mode, "/", tc.proto), func(t *testing.T) {
			fixture := hostProbeFixture(t, probe, tc.mode, tc.proto, 71)
			child := startProbeCandidate(t, fixture, 31)
			start := time.Now()
			got, e := child.process(nativePattern(31, 0))
			elapsed := time.Since(start)
			if tc.mode == "normal" {
				if e != nil || len(got) != 320 {
					t.Fatal(e)
				}
				return
			}
			if tc.mode == "fatal" || tc.mode == "reject" {
				var terminal *asp.Terminal
				if !errors.As(e, &terminal) || terminal.Status.Code < 500 {
					t.Fatal("terminal lost", e)
				}
			} else if !errors.Is(e, tc.want) {
				t.Fatalf("wrong failure: %v want %v", e, tc.want)
			}
			limit := 250 * time.Millisecond
			if tc.mode == "config-hang" || tc.mode == "response-hang" {
				limit = aspFrameBudget + 2*time.Second
				if elapsed < aspFrameBudget-250*time.Millisecond {
					t.Fatal("hang timed out before configured budget", e, elapsed)
				}
			}
			if got != nil || !child.broken.Load() || elapsed > limit {
				t.Fatal("fault not bounded", e, elapsed, limit)
			}
			firstError := e
			if got, e = child.process(nativePattern(31, 1)); got != nil || e == nil {
				t.Fatal("failed session reused/downgraded")
			}
			t.Logf("terminal failure %v in %s; budget=%s, no downgrade/retry", firstError, elapsed, aspFrameBudget)
		})
	}
	for _, field := range []string{"mic", "ref"} {
		t.Run("changed-"+field, func(t *testing.T) {
			child := startProbeCandidate(t, hostProbeFixture(t, probe, "normal", "cfg2", 71), 31)
			i := nativePattern(31, 0).input()
			if field == "mic" {
				i.Config.Mic.ValidBits = 23
			} else {
				i.Config.Ref.ValidBits = 15
				i.Config.Ref.FullBits = 14
			}
			if _, e := child.exchange(i); !errors.Is(e, errASPReference) {
				t.Fatal("stale MIC/REF accepted", e)
			}
		})
	}
}

func TestASPSecureLaunchInvalidSnapshotBeforeOpen(t *testing.T) {
	c := aspCandidate{identity: aspPinnedIdentity{{dev: 1, ino: 1}, {dev: 1, ino: 2}}}
	for _, f := range []*externalASP{nil, {}, {session: sessionCFG2{packet: "invalid"}}} {
		called := false
		if child, e := startASPCandidate(context.Background(), c, f, func() (*secureASPPinned, error) { called = true; return nil, nil }); child != nil || e == nil || called {
			t.Fatal("invalid snapshot opened/launched helper", e)
		}
	}
	good := probeHostState(t, 31)
	good.config.Mic.FullBits = 23
	good.config.Mic.Alignment = asp.Low
	called := false
	if child, e := startASPCandidate(context.Background(), c, good, func() (*secureASPPinned, error) { called = true; return nil, nil }); child != nil || !errors.Is(e, errASPInvalid) || called {
		t.Fatal("stale snapshot launched", e)
	}
}

func TestASPSecureLaunchStartErrors(t *testing.T) {
	probe := buildASPHostProbe(t, 17)
	for _, mode := range []string{"open", "nil", "closed-exe", "closed-root", "missing-proc", "bad-elf", "cancel", "nil-context"} {
		t.Run(mode, func(t *testing.T) {
			fixture := hostProbeFixture(t, probe, "normal", "cfg2", 71)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			p := pinASPProbe(t, fixture)
			proc := "/proc/self/fd/3"
			switch mode {
			case "open":
				p.Close()
				c := aspCandidate{identity: aspPinnedIdentity{{dev: 1, ino: 1}, {dev: 1, ino: 2}}}
				child, e := startASPCandidate(ctx, c, probeHostState(t, 31), func() (*secureASPPinned, error) { return nil, os.ErrPermission })
				if child != nil || !errors.Is(e, errASPStart) {
					t.Fatal(e)
				}
				return
			case "nil":
				p.Close()
				p = nil
			case "closed-exe":
				p.executable.Close()
			case "closed-root":
				p.externalRoot.Close()
			case "missing-proc":
				proc = "/proc/self/fd/-1"
			case "bad-elf":
				p.Close()
				if e := os.WriteFile(fixture.part(4), aspSecureELF("truncated"), 0755); e != nil {
					t.Fatal(e)
				}
				p = pinASPProbe(t, fixture)
			case "cancel":
				cancel()
			case "nil-context":
				ctx = nil

			}
			child, e := startExternalASPPinnedAtProcFD(ctx, p, probeHostState(t, 31), proc, nil)
			if child != nil || e == nil {
				t.Fatal("invalid launch accepted", e)
			}
			if p != nil {
				assertASPConsumed(t, p)
			}
		})
	}
}

func TestASPSecureLaunchStartupBounded(t *testing.T) {
	probe := buildASPHostProbe(t, 17)
	for _, mode := range []string{"start-bad", "start-hang", "hold-start", "cancel"} {
		t.Run(mode, func(t *testing.T) {
			profile := mode
			if mode == "cancel" {
				profile = "start-hang"
			}
			fixture := hostProbeFixture(t, probe, profile, "cfg2", 71)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			p := pinASPProbe(t, fixture)
			// Capture the failed instance only to verify the host's single reaper.
			state := probeHostState(t, 31)
			done := make(chan error, 1)
			start := time.Now()
			go func() { _, e := startExternalASPPinned(ctx, p, state); done <- e }()
			if mode == "cancel" {
				time.Sleep(20 * time.Millisecond)
				cancel()
			}
			select {
			case e := <-done:
				if !errors.Is(e, errASPHandshake) {
					t.Fatal(e)
				}
			case <-time.After(time.Second):
				t.Fatal("startup/descendant pipe blocked")
			}
			elapsed := time.Since(start)
			assertASPConsumed(t, p)
			if mode == "start-hang" || mode == "hold-start" {
				if elapsed < 450*time.Millisecond {
					t.Fatal("500-ms HELLO bypassed")
				}
			}
			if mode == "cancel" && elapsed > 250*time.Millisecond {
				t.Fatal("cancellation waited for HELLO timeout")
			}
			waitPinnedHost(t, state)
			t.Logf("%s returned in %s, sole Wait reaped child", mode, elapsed)
		})
	}
}

func TestASPSecureLaunchSourceFallbackAndPreflight(t *testing.T) {
	probe := buildASPHostProbe(t, 17)
	for _, proto := range []string{"cfg2"} {
		t.Run(proto, func(t *testing.T) {
			fixture := hostProbeFixture(t, probe, "runtime-fatal", proto, 71)
			c := discoverASPCandidate(fixture.open)
			s := controlSource(t, "normal", config.ASPExternal)
			var starts atomic.Int32
			var child *externalASP
			s.aspStart = func(ctx context.Context) (*externalASP, error) {
				starts.Add(1)
				var e error
				child, e = startASPCandidate(ctx, c, probeHostState(t, 31), fixture.open)
				return child, e
			}
			s.SetASP(config.ASPExternal)
			waitASPChild(t, s)
			ch, stop := s.Listen("pinned current stock")
			defer stop()
			stock := testSource()
			for tick := 0; tick < 8; tick++ {
				raw, n := rawFrame(tick, false), nativePattern(31, tick)
				stock.broadcast(bytes.Clone(raw))
				s.broadcastNative(bytes.Clone(raw), n)
				got := <-ch
				want := stock.Recent(20 * time.Millisecond)
				if tick == 3 {
					want = probeHostPCM(n, 3, 17, 71)
				}
				if !slices.Equal(got, want) {
					t.Fatal("preflight/activation/fallback current tick", tick)
				}
				if (s.ASPStatus().Active == config.ASPExternal) != (tick == 3) {
					t.Fatal("wrong readiness", tick, s.ASPStatus())
				}
			}
			s.SetASP(config.ASPExternal)
			if starts.Load() != 1 || !s.ASPStatus().Confirmed || !strings.Contains(s.ASPStatus().Debug, "503:") {
				t.Fatal("terminal fallback retried or lost status", s.ASPStatus(), starts.Load())
			}
			waitPinnedHost(t, child)
		})
	}
}

func TestASPSecureLaunchSourceConcurrentCancel(t *testing.T) {
	probe := buildASPHostProbe(t, 17)
	fixture := hostProbeFixture(t, probe, "start-hang", "cfg2", 71)
	c := discoverASPCandidate(fixture.open)
	s := controlSource(t, "normal", config.ASPExternal)
	var running, maximum atomic.Int32
	var statesMu sync.Mutex
	var states []*externalASP
	entered := make(chan struct{})
	signal := sync.OnceFunc(func() { close(entered) })
	s.aspStart = func(ctx context.Context) (*externalASP, error) {
		n := running.Add(1)
		signal()
		defer running.Add(-1)
		for old := maximum.Load(); n > old && !maximum.CompareAndSwap(old, n); old = maximum.Load() {
		}
		f := probeHostState(t, 31)
		_, e := startASPCandidate(ctx, c, f, fixture.open)
		statesMu.Lock()
		if f.cmd != nil {
			states = append(states, f)
		}
		statesMu.Unlock()
		return nil, e
	}
	s.SetASP(config.ASPExternal)
	select {
	case <-entered:
	case <-time.After(time.Second):
		t.Fatal("startup worker never entered")
	}
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				s.SetASP(config.ASPBuiltin)
				s.SetASP(config.ASPExternal)
			}
		}()
	}
	wg.Wait()
	s.SetASP(config.ASPBuiltin)
	until := time.Now().Add(time.Second)
	for {
		s.aspMu.Lock()
		busy := s.aspStarting
		s.aspMu.Unlock()
		if !busy {
			break
		}
		if time.Now().After(until) {
			t.Fatal("serialized startup stuck")
		}
		time.Sleep(time.Millisecond)
	}
	if maximum.Load() != 1 || s.ASPStatus().Active != config.ASPBuiltin {
		t.Fatal("concurrent launch ownership", maximum.Load())
	}
	statesMu.Lock()
	defer statesMu.Unlock()
	for _, f := range states {
		waitPinnedHost(t, f)
	}
	// Explicit fresh epoch after all cancellations, using the same cached identity.
	if e := os.WriteFile(filepath.Join(fixture.part(2), "etc/profile.conf"), []byte("normal cfg2 71"), 0600); e != nil {
		t.Fatal(e)
	}
	f, e := startASPCandidate(context.Background(), c, probeHostState(t, 31), fixture.open)
	if e != nil {
		t.Fatal(e)
	}
	defer waitPinnedHost(t, f)
	if _, e = f.process(nativePattern(31, 0)); e != nil {
		t.Fatal("cancellation poisoned next startup", e)
	}
}

func TestASPSecureLaunchLifecycleNoLeaks(t *testing.T) {
	probe := buildASPHostProbe(t, 17)
	fixture := hostProbeFixture(t, probe, "normal", "cfg2", 71)
	c := discoverASPCandidate(fixture.open)
	warm, e := startASPCandidate(context.Background(), c, probeHostState(t, 31), fixture.open)
	if e != nil {
		t.Fatal(e)
	}
	waitPinnedHost(t, warm)
	runtime.GC()
	beforeFD, beforeG := aspSecureFDCount(t), runtime.NumGoroutine()
	beforeChildren := aspProbeChildren(t)
	for i := 0; i < 64; i++ {
		ctx, cancel := context.WithCancel(context.Background())
		state := probeHostState(t, 31)
		var child *externalASP
		var e error
		switch i % 4 {
		case 0:
			child, e = startASPCandidate(ctx, c, state, fixture.open)
		case 1:
			p, err := fixture.open()
			if err != nil {
				t.Fatal(err)
			}
			child, e = startExternalASPPinnedAtProcFD(ctx, p, state, "/proc/self/fd/-1", nil)
			assertASPConsumed(t, p)
		case 2:
			cancel()
			child, e = startASPCandidate(ctx, c, state, fixture.open)
		case 3:
			child, e = startASPCandidate(ctx, c, state, fixture.open)
			cancel()
		}
		if i%4 == 1 || i%4 == 2 {
			if e == nil || child != nil {
				t.Fatal("failure launched")
			}
		} else {
			if e == nil && child != nil {
				waitPinnedHost(t, child)
			} else if state.cmd != nil {
				waitPinnedHost(t, state)
			} else {
				t.Fatal("unexpected startup failure", e)
			}
		}
		cancel()
	}
	until := time.Now().Add(time.Second)
	for runtime.NumGoroutine() > beforeG && time.Now().Before(until) {
		runtime.Gosched()
		time.Sleep(time.Millisecond)
	}
	afterFD, afterG := aspSecureFDCount(t), runtime.NumGoroutine()
	afterChildren := aspProbeChildren(t)
	t.Logf("64 host starts/errors/cancels: FD %d -> %d; goroutines %d -> %d; children %q -> %q", beforeFD, afterFD, beforeG, afterG, beforeChildren, afterChildren)
	if beforeFD != afterFD || afterG > beforeG || beforeChildren != afterChildren {
		t.Fatal("host descriptor/goroutine/child leak")
	}
}

func TestASPSecureLaunchReleaseBeforeHELLO(t *testing.T) {
	probe := buildASPHostProbe(t, 17)
	fixture := hostProbeFixture(t, probe, "start-hang", "cfg2", 71)
	p := pinASPProbe(t, fixture)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cmd, e := prepareSecureASPPinned(ctx, p, "/proc/self/fd/3", nil)
	if e != nil {
		t.Fatal(e)
	}
	// Preparation must not spawn. Only the host is allowed to Start.
	if cmd.Process != nil {
		t.Fatal("preparation launched child")
	}
	released := make(chan struct{})
	done := make(chan error, 1)
	state := probeHostState(t, 31)
	go func() {
		_, e := startExternalASPProcessWithRelease(ctx, cmd, state, func() { p.Close(); cmd.ExtraFiles = nil; close(released) })
		done <- e
	}()
	select {
	case <-released:
	case <-time.After(time.Second):
		t.Fatal("parent pins held during HELLO")
	}
	assertASPConsumed(t, p)
	if cmd.Process == nil {
		t.Fatal("release not tied to actual Start")
	}
	cancel()
	select {
	case e := <-done:
		if e == nil {
			t.Fatal("cancelled startup succeeded")
		}
	case <-time.After(250 * time.Millisecond):
		t.Fatal("cancel blocked")
	}
	waitPinnedHost(t, state)
}

func TestASPSecureLaunchConcurrentPinnedHostSwaps(t *testing.T) {
	original, decoy := buildASPHostProbe(t, 17), buildASPHostProbe(t, 333)
	// Pin at the activation seam, then make the visible name actively hostile.
	// The cached expected identity is checked by startASPCandidate itself.
	for i := 0; i < 16; i++ {
		fixture := hostProbeFixture(t, original, "normal", "cfg2", 71)
		replacement := hostProbeFixture(t, decoy, "normal", "cfg2", 900)
		c := discoverASPCandidate(fixture.open)
		p, e := fixture.open()
		if e != nil {
			t.Fatal(e)
		}
		stop, ready := make(chan struct{}), make(chan struct{})
		result := make(chan error, 1)
		var cycles atomic.Int32
		go func() {
			for j := 0; ; j++ {
				for _, pair := range [][2]string{{fixture.part(2), fixture.part(2) + ".swap"}, {replacement.part(2), fixture.part(2)}, {fixture.part(2) + ".swap", replacement.part(2)}} {
					if e := os.Rename(pair[0], pair[1]); e != nil {
						if j == 0 {
							close(ready)
						}
						result <- e
						return
					}
				}
				cycles.Add(1)
				if j == 0 {
					close(ready)
				}
				select {
				case <-stop:
					result <- nil
					return
				case <-time.After(time.Millisecond):
				}
			}
		}()
		<-ready
		child, e := startASPCandidate(context.Background(), c, probeHostState(t, 31), func() (*secureASPPinned, error) { return p, nil })
		if e == nil {
			n := nativePattern(31, i)
			got, err := child.process(n)
			if err != nil || !slices.Equal(got, probeHostPCM(n, 0, 17, 71)) {
				t.Errorf("concurrent host swap selected decoy: %v", err)
			}
			waitPinnedHost(t, child)
		}
		close(stop)
		swapErr := <-result
		assertASPConsumed(t, p)
		if e != nil || swapErr != nil || cycles.Load() < 1 {
			t.Fatal("concurrent swap/host failed", e, swapErr)
		}
	}
	t.Log("16 real supervised CFG2 host sessions overlapped directory rename cycles: original ELF/assets only")
}

func TestASPSecureLaunchUnavailableSourceStock(t *testing.T) {
	probe := buildASPHostProbe(t, 17)
	for _, mode := range []string{"open-error", "changed-inode", "invalid-snapshot"} {
		t.Run(mode, func(t *testing.T) {
			fixture := hostProbeFixture(t, probe, "normal", "cfg2", 71)
			c := discoverASPCandidate(fixture.open)
			opener := fixture.open
			state := probeHostState(t, 31)
			switch mode {
			case "open-error":
				opener = func() (*secureASPPinned, error) { return nil, os.ErrPermission }
			case "changed-inode":
				if e := os.Rename(fixture.part(4), fixture.part(4)+".old"); e != nil {
					t.Fatal(e)
				}
				b, e := os.ReadFile(probe.path)
				if e != nil {
					t.Fatal(e)
				}
				if e = os.WriteFile(fixture.part(4), b, 0755); e != nil {
					t.Fatal(e)
				}
			case "invalid-snapshot":
				state.session = sessionCFG2{}
			}
			s := controlSource(t, "normal", config.ASPExternal)
			var starts atomic.Int32
			s.aspStart = func(ctx context.Context) (*externalASP, error) {
				starts.Add(1)
				return startASPCandidate(ctx, c, state, opener)
			}
			s.SetASP(config.ASPExternal)
			waitASP(t, s, func(st ASPStatus) bool { return strings.Contains(st.Debug, "Fallback:") })
			ch, stop := s.Listen("unavailable candidate")
			defer stop()
			stock := testSource()
			for tick := 0; tick < 4; tick++ {
				raw := rawFrame(tick, false)
				stock.broadcast(bytes.Clone(raw))
				s.broadcastNative(bytes.Clone(raw), nativePattern(31, tick))
				if !slices.Equal(<-ch, stock.Recent(20*time.Millisecond)) {
					t.Fatal("unavailable candidate changed stock")
				}
			}
			s.SetASP(config.ASPExternal)
			if starts.Load() != 1 || s.ASPStatus().Active != config.ASPBuiltin || state.cmd != nil {
				t.Fatal("unavailable candidate spawned/retried")
			}
		})
	}
}

func TestASPSecureLaunchClosedHostReader(t *testing.T) {
	probe := buildASPHostProbe(t, 17)
	child := startProbeCandidate(t, hostProbeFixture(t, probe, "normal", "cfg2", 71), 31)
	child.out.Close()
	if got, e := child.process(nativePattern(31, 0)); got != nil || !errors.Is(e, errASPExited) {
		t.Fatal("closed host reader accepted", e)
	}
	waitPinnedHost(t, child)
}
