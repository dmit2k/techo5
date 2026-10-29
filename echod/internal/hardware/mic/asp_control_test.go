package mic

import (
	"bytes"
	"context"
	"errors"
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

func waitASP(t testing.TB, s *Source, predicate func(ASPStatus) bool) ASPStatus {
	t.Helper()
	until := time.Now().Add(3 * time.Second)
	for {
		st := s.ASPStatus()
		if predicate(st) {
			return st
		}
		if time.Now().After(until) {
			t.Fatalf("ASP state did not settle: %+v", st)
		}
		time.Sleep(time.Millisecond)
	}
}

func controlSource(t *testing.T, mode string, saved config.ASPMode, configs ...asp.Config) *Source {
	t.Helper()
	config.Use(filepath.Join(t.TempDir(), "state.json"))
	if err := config.Set().Microphone().ASP(saved); err != nil {
		t.Fatal(err)
	}
	s := testSource()
	// Most control fixtures model an already-open capture stream.
	s.beginASPCapture()
	var children []*externalASP
	var mu sync.Mutex
	s.aspStart = func(ctx context.Context) (*externalASP, error) {
		f, err := startTestASP(ctx, childCommand(mode), configs...)
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
				break
			}
			time.Sleep(time.Millisecond)
		}
		mu.Lock()
		defer mu.Unlock()
		for _, f := range children {
			select {
			case <-f.done:
			case <-time.After(time.Second):
				t.Error("unreaped ASP child")
			}
		}
	})
	return s
}

func waitASPChild(t *testing.T, s *Source) {
	t.Helper()
	until := time.Now().Add(3 * time.Second)
	for {
		s.aspMu.Lock()
		ready := s.asp != nil
		s.aspMu.Unlock()
		if ready {
			return
		}
		if time.Now().After(until) {
			t.Fatal("child not attached", s.ASPStatus())
		}
		time.Sleep(time.Millisecond)
	}
}

func preflightASP(t *testing.T, s *Source, mode string) {
	t.Helper()
	s.SetASP(config.ASPExternal)
	waitASPChild(t, s)
	ch, stop := s.Listen("preflight")
	defer stop()
	stock := testSource()
	for i := 0; i < 4; i++ {
		raw := rawFrame(i, false)
		n := nativePattern(31, i)
		stock.broadcast(bytes.Clone(raw))
		s.broadcastNative(bytes.Clone(raw), n)
		got := <-ch
		if i < 3 {
			if !slices.Equal(got, stock.Recent(20*time.Millisecond)) {
				t.Fatal("preflight PCM escaped")
			}
			st := s.ASPStatus()
			if st.Active != config.ASPBuiltin || st.Confirmed || st.Selected != config.ASPExternal || st.Debug != "H:210:External ASP starting" {
				t.Fatal("false readiness", st)
			}
		} else {
			want := fixturePCM(n, 3)
			if mode == "silence" {
				clear(want)
			}
			if !slices.Equal(got, want) {
				t.Fatal("promotion used wrong current frame")
			}
			st := s.ASPStatus()
			if st.Active != config.ASPExternal || !st.Confirmed || st.Debug != "H:220:External ASP processing OK" {
				t.Fatal("not promoted", st)
			}
		}
	}
	if len(s.Recent(time.Second)) != 4*FrameSamples {
		t.Fatal("history reset or duplicate frame")
	}
	if s.Dropped() != 0 {
		t.Fatal("listener dropped")
	}
}

func TestASPControlFirstSuccessAndSilence(t *testing.T) {
	for _, mode := range []string{"normal", "silence"} {
		t.Run(mode, func(t *testing.T) {
			s := controlSource(t, mode, config.ASPBuiltin)
			changes := s.ASPChanges()
			preflightASP(t, s, mode)
			select {
			case <-changes:
			default:
				t.Fatal("no transition notification")
			}
			// Source owns audio, not disk I/O; the microphone feature commits this confirmed state.
			if config.Get().Microphone.ASP != config.ASPBuiltin {
				t.Fatal("capture wrote configuration")
			}
			s.SetASP(config.ASPBuiltin)
			st := s.ASPStatus()
			if st.Selected != config.ASPBuiltin || st.Active != config.ASPBuiltin || st.Debug != "H:200:Built-in processing OK" {
				t.Fatal(st)
			}
		})
	}
}

func TestASPControlStartupFailuresAndBootKeep(t *testing.T) {
	for _, saved := range []config.ASPMode{config.ASPBuiltin, config.ASPExternal} {
		for _, kind := range []string{"missing", "start", "handshake", "malformed", "none", "reference"} {
			t.Run(string(saved)+"/"+kind, func(t *testing.T) {
				mode := "normal"
				switch kind {
				case "handshake":
					mode = "start-bad"
				case "malformed":
					mode = "length"
				case "none":
					mode = "output-hang"
				}
				s := controlSource(t, mode, saved)
				if kind == "missing" {
					s.aspStart = func(context.Context) (*externalASP, error) { return nil, errASPNotFound }
				}
				if kind == "start" {
					s.aspStart = func(context.Context) (*externalASP, error) { return nil, errASPStart }
				}
				s.SetASP(config.ASPExternal)
				if kind == "malformed" || kind == "none" || kind == "reference" {
					waitASPChild(t, s)
					n := nativePattern(31, 0)
					if kind == "reference" {
						n.ref = nil
					}
					s.broadcastNative(rawFrame(0, false), n)
				}
				st := waitASP(t, s, func(st ASPStatus) bool { return strings.HasPrefix(st.Debug, "H:500:Fallback:") })
				want := map[string]string{"missing": errASPNotFound.Error(), "start": errASPStart.Error(), "handshake": errASPHandshake.Error(), "malformed": errASPInvalid.Error(), "none": errASPTimeout.Error(), "reference": errASPReference.Error()}[kind]
				if st.Debug != "H:500:Fallback: "+want || st.Active != config.ASPBuiltin || st.Confirmed || st.Selected != saved {
					t.Fatalf("wrong failure semantics: %+v", st)
				}
				if config.Get().Microphone.ASP != saved {
					t.Fatal("failure overwrote stored choice")
				}
			})
		}
	}
}

func TestASPControlSavedExternalSuccess(t *testing.T) {
	s := controlSource(t, "normal", config.ASPExternal)
	preflightASP(t, s, "normal")
	if config.Get().Microphone.ASP != config.ASPExternal {
		t.Fatal("saved External erased")
	}
}

func TestASPControlRuntimeFaultsCurrentStockNoRetry(t *testing.T) {
	for _, mode := range []string{"crash", "output-hang", "input-hang", "input-close", "short", "short-open", "length", "sequence", "repeat"} {
		t.Run(mode, func(t *testing.T) {
			s := controlSource(t, "runtime:"+mode, config.ASPExternal)
			preflightASP(t, s, "normal")
			s.aspMu.Lock()
			f := s.asp
			s.aspMu.Unlock()
			ch, stop := s.Listen("same existing consumer")
			defer stop()
			stock := testSource()
			raw := rawFrame(5, false)
			stock.broadcast(bytes.Clone(raw))
			s.broadcastNative(bytes.Clone(raw), nativePattern(31, 5))
			if !slices.Equal(<-ch, stock.Recent(20*time.Millisecond)) {
				t.Fatal("failing tick not current stock")
			}
			st := s.ASPStatus()
			reason := errASPExited.Error()
			switch mode {
			case "output-hang", "input-hang", "short-open":
				reason = errASPTimeout.Error()
			case "short", "length", "sequence", "repeat":
				reason = errASPInvalid.Error()
			}
			if st.Selected != config.ASPExternal || st.Active != config.ASPBuiltin || !st.Confirmed || st.Debug != "H:500:Fallback: "+reason {
				t.Fatal(st)
			}
			for i := 0; i < 5; i++ {
				s.broadcastNative(rawFrame(i, false), nativePattern(31, i))
				<-ch
			}
			if s.ASPStatus() != st {
				t.Fatal("fallback reason erased or automatic retry")
			}
			if config.Get().Microphone.ASP != config.ASPExternal {
				t.Fatal("runtime fallback erased config")
			}
			s.aspMu.Lock()
			attached := s.asp
			s.aspMu.Unlock()
			if attached != nil || !f.broken.Load() {
				t.Fatal("failed instance retained")
			}
			// Re-selecting the same failed External is intentionally not a restart loop.
			s.SetASP(config.ASPExternal)
			if s.ASPStatus() != st {
				t.Fatal("same choice restarted failed helper")
			}
		})
	}
}

func TestASPControlNoCaptureBeforeRunAndNoPCMTimeout(t *testing.T) {
	s := controlSource(t, "normal", config.ASPExternal)
	// Model startup restore before the hardware service opens ALSA.
	s.aspMu.Lock()
	s.aspCapture = false
	s.aspMu.Unlock()
	s.SetASP(config.ASPExternal)
	s.aspMu.Lock()
	started := s.aspStarting
	s.aspMu.Unlock()
	if started {
		t.Fatal("started helper before hardware capture")
	}
	time.Sleep(aspFirstPCMBudget + 20*time.Millisecond)
	if s.ASPStatus().Debug != "H:210:External ASP starting" {
		t.Fatal("timed out before capture goroutine exists")
	}
	s.beginASPCapture()
	waitASPChild(t, s)
	st := waitASP(t, s, func(st ASPStatus) bool { return strings.HasPrefix(st.Debug, "H:500:Fallback:") })
	if st.Selected != config.ASPExternal || st.Debug != "H:504:Fallback: no valid PCM frames received" {
		t.Fatal(st)
	}
}

func TestASPControlLatestCommandAndBoundedStartup(t *testing.T) {
	s := controlSource(t, "normal", config.ASPBuiltin)
	entered := make(chan struct{}, 1)
	release := make(chan struct{})
	var running, maximum, starts atomic.Int32
	s.aspStart = func(ctx context.Context) (*externalASP, error) {
		n := running.Add(1)
		defer running.Add(-1)
		for {
			old := maximum.Load()
			if n <= old || maximum.CompareAndSwap(old, n) {
				break
			}
		}
		starts.Add(1)
		select {
		case entered <- struct{}{}:
		default:
		}
		select {
		case <-release:
			return startTestASP(context.Background(), childCommand("normal"))
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	start := time.Now()
	s.SetASP(config.ASPExternal)
	if time.Since(start) > 50*time.Millisecond {
		t.Fatal("command blocked on startup")
	}
	<-entered
	for i := 0; i < 100; i++ {
		s.SetASP(config.ASPBuiltin)
		s.SetASP(config.ASPExternal)
	}
	s.SetASP(config.ASPBuiltin)
	close(release)
	waitASP(t, s, func(st ASPStatus) bool { return st.Selected == config.ASPBuiltin })
	time.Sleep(10 * time.Millisecond)
	if s.ASPStatus().Active != config.ASPBuiltin || maximum.Load() != 1 {
		t.Fatal("stale result or concurrent startup workers", s.ASPStatus(), maximum.Load())
	}
	t.Log("serialized startup attempts", starts.Load())
}

func TestASPControlCancelPipeWithoutSourceLock(t *testing.T) {
	s := controlSource(t, "normal", config.ASPBuiltin)
	ack, childAck, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer ack.Close()
	defer childAck.Close()
	var child *externalASP
	s.aspStart = func(ctx context.Context) (*externalASP, error) {
		cmd := childCommand("output-hang")
		cmd.Env = append(cmd.Env, "TECHO5_ASP_ACK=1")
		cmd.ExtraFiles = []*os.File{childAck}
		f, err := startTestASP(ctx, cmd)
		child = f
		return f, err
	}
	s.SetASP(config.ASPExternal)
	waitASPChild(t, s)
	done := make(chan struct{})
	go func() { s.broadcastNative(rawFrame(0, false), nativePattern(31, 0)); close(done) }()
	ack.SetReadDeadline(time.Now().Add(time.Second))
	var b [1]byte
	if _, err := ack.Read(b[:]); err != nil {
		t.Fatal("child never received the real native request", err)
	}
	// The child has read the complete current native request and is now blocked without output.
	if !s.mu.TryLock() {
		t.Fatal("Source.mu held during preflight pipe I/O")
	}
	s.mu.Unlock()
	start := time.Now()
	s.SetASP(config.ASPBuiltin)
	if time.Since(start) > 50*time.Millisecond {
		t.Fatal("HA control blocked on hung child")
	}
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("capture stuck after cancellation")
	}
	if s.ASPStatus().Active != config.ASPBuiltin || s.ASPStatus().Debug != "H:200:Built-in processing OK" {
		t.Fatal("stale result won")
	}
	select {
	case <-child.done:
	case <-time.After(time.Second):
		t.Fatal("old child not reaped")
	}
}

func TestASPControlJitterAndHardDeadline(t *testing.T) {
	var stressSink atomic.Uint64
	stop := make(chan struct{})
	burned := make(chan struct{})
	go func() {
		defer close(burned)
		for {
			select {
			case <-stop:
				return
			default:
			}
			sum := uint64(7)
			for i := 0; i < 20000; i++ {
				sum = sum*1664525 + 1013904223
			}
			stressSink.Store(sum)
			runtime.Gosched()
		}
	}()
	defer func() { close(stop); <-burned }()
	f := fixture(t, "jitter")
	n := nativePattern(31, 0)
	var durations []int64
	for i := 0; i < 100; i++ {
		start := time.Now()
		_, err := f.process(n)
		elapsed := time.Since(start)
		if err != nil {
			t.Fatal("routine >10-ms jitter killed ASP", i, elapsed, err)
		}
		durations = append(durations, elapsed.Nanoseconds())
	}
	slices.Sort(durations)
	t.Logf("jitter ns p50=%d p95=%d p99=%d max=%d budget=%s", durations[50], durations[95], durations[99], durations[99], aspFrameBudget)
	hard := fixture(t, "delay:60")
	start := time.Now()
	if _, err := hard.process(n); !errors.Is(err, errASPTimeout) {
		t.Fatal("hard stall not terminal", err)
	}
	t.Logf("hard deadline elapsed=%s", time.Since(start))
	if _, err := hard.process(n); err == nil {
		t.Fatal("late response reused")
	}
}

func TestASPControlLateSuccessfulStartIsDiscarded(t *testing.T) {
	s := controlSource(t, "normal", config.ASPBuiltin)
	entered, release := make(chan struct{}), make(chan struct{})
	returned := make(chan *externalASP, 1)
	s.aspStart = func(context.Context) (*externalASP, error) {
		close(entered)
		<-release
		f, err := startTestASP(context.Background(), childCommand("normal"))
		returned <- f
		return f, err
	}
	s.SetASP(config.ASPExternal)
	<-entered
	s.SetASP(config.ASPBuiltin)
	close(release)
	var old *externalASP
	select {
	case old = <-returned:
	case <-time.After(time.Second):
		t.Fatal("late startup not completed")
	}
	if old == nil {
		t.Fatal("test did not return a successful late child")
	}
	select {
	case <-old.done:
	case <-time.After(time.Second):
		t.Fatal("stale child leaked")
	}
	if s.ASPStatus().Selected != config.ASPBuiltin || s.ASPStatus().Confirmed {
		t.Fatal("old success overruled Built-in")
	}
}

func TestASPStartupCancellationAndNonexecutingFile(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { _, err := startTestASP(ctx, childCommand("start-hang")); done <- err }()
	time.Sleep(10 * time.Millisecond)
	start := time.Now()
	cancel()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("canceled startup succeeded")
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("startup cancellation waited for handshake timeout")
	}
	t.Log("startup cancel duration", time.Since(start))
	path := filepath.Join(t.TempDir(), "external-asp")
	if err := os.WriteFile(path, []byte("not an executable format"), 0700); err != nil {
		t.Fatal(err)
	}
	if _, err := startTestASP(context.Background(), exec.Command(path)); !errors.Is(err, errASPStart) {
		t.Fatal("exec failure misclassified", err)
	}
}
