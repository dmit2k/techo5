package mic

import (
	"context"
	"encoding/binary"
	"errors"
	"slices"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/HuskerMinion/techo5/echod/internal/asp"
	"github.com/HuskerMinion/techo5/echod/internal/config"
	"github.com/HuskerMinion/techo5/echod/internal/lib/alsa"
)

type stage06Process struct {
	once        sync.Once
	canceled    chan struct{}
	done        chan struct{}
	wait        <-chan struct{}
	value       int16
	entered     chan struct{}
	fail        error
	honorCancel bool
}

func stage06Fake(v int16) *stage06Process {
	return &stage06Process{canceled: make(chan struct{}), done: make(chan struct{}), value: v}
}
func (p *stage06Process) Exchange(ctx context.Context, bufs [][]byte) ([]byte, error) {
	if p.entered != nil {
		select {
		case p.entered <- struct{}{}:
		default:
		}
	}
	if p.fail != nil {
		return nil, p.fail
	}

	if p.wait != nil {
		if p.honorCancel {
			select {
			case <-p.wait:
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		} else {
			<-p.wait
		} // deliberately hostile, simulates stale reply
	}
	b := make([]byte, asp.OutputBytes)
	for i := 0; i < len(b); i += 2 {
		binary.LittleEndian.PutUint16(b[i:], uint16(p.value))
	}
	return b, nil
}
func (p *stage06Process) Cancel() { p.once.Do(func() { close(p.canceled) }) }
func (p *stage06Process) Close() error {
	p.Cancel()
	select {
	case <-p.done:
	default:
		close(p.done)
	}
	return nil
}

func stage06Block() NativeBlock {
	b := make([]byte, FrameSamples*Channels*3)
	return nativeInterleaved(b, 16000, asp.FormatS24_3, 24, 0, Mics, Refs)
}
func stage06Stock() []byte {
	return make([]byte, FrameSamples*Channels*3)
}
func stage06Await(t *testing.T, pred func() bool) {
	t.Helper()
	end := time.After(2 * time.Second)
	ticker := time.NewTicker(time.Millisecond)
	defer ticker.Stop()
	for !pred() {
		select {
		case <-end:
			t.Fatal("ASP startup/cleanup did not finish")
		case <-ticker.C:
		}
	}
}
func stage06Ready(s *Source) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.aspCurrent != nil && s.aspCurrent.process != nil
}
func stage06Source(t *testing.T, opener func(context.Context, asp.Descriptor) (aspProcess, error)) *Source {
	t.Helper()
	s := New()
	s.aspOpen = opener
	s.BindASP(context.Background())
	if got := s.SetASP(config.ASPExternal); got != config.ASPExternal {
		t.Fatal(got)
	}
	t.Cleanup(func() { s.StopASP(); s.WaitASP() })
	return s
}

func TestL01SlowREADYAndFirstPCMActive(t *testing.T) {
	ready := make(chan struct{})
	p := stage06Fake(1642)
	s := stage06Source(t, func(ctx context.Context, d asp.Descriptor) (aspProcess, error) {
		select {
		case <-ready:
			return p, nil
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	})
	ch, unsub := s.Listen("stage06")
	defer unsub()
	if got := s.captureNative(stage06Block()); got != nil {
		t.Fatal("launch blocked capture")
	}
	s.broadcastSelected(stage06Stock(), false, nil)
	if len(<-ch) != FrameSamples || s.ActiveASP() != config.ASPBuiltin {
		t.Fatal("Built-in not running during READY")
	}
	close(ready)
	stage06Await(t, func() bool { return stage06Ready(s) })
	r := s.captureNative(stage06Block())
	s.broadcastSelected(stage06Stock(), false, r)
	f := <-ch
	if r == nil || len(f) != FrameSamples || f[0] != 1642 || s.ActiveASP() != config.ASPExternal || s.SelectedASP() != config.ASPExternal {
		t.Fatal("first External output failed")
	}
	h := s.Recent(20 * time.Millisecond)
	if !slices.Equal(h, f) {
		t.Fatal("history differs from listeners")
	}
}

func TestL02FallbackNoRetryAndManualReactivation(t *testing.T) {
	var launches atomic.Int32
	s := stage06Source(t, func(context.Context, asp.Descriptor) (aspProcess, error) {
		launches.Add(1)
		if launches.Load() == 1 {
			return nil, errors.New("missing executable")
		}
		return stage06Fake(803), nil
	})
	if s.captureNative(stage06Block()) != nil {
		t.Fatal("launch should be async")
	}
	stage06Await(t, func() bool {
		s.mu.Lock()
		defer s.mu.Unlock()
		return s.aspCurrent == nil && !s.aspLaunchWanted
	})
	for i := 0; i < 5; i++ {
		s.captureNative(stage06Block())
	}
	if launches.Load() != 1 || s.ActiveASP() != config.ASPBuiltin || s.SelectedASP() != config.ASPExternal {
		t.Fatal("fallback retried or overwrote Selected")
	}
	s.SetASP(config.ASPBuiltin)
	s.SetASP(config.ASPExternal)
	s.captureNative(stage06Block())
	stage06Await(t, func() bool { return stage06Ready(s) })
	if launches.Load() != 2 {
		t.Fatal("manual retry did not re-evaluate executable")
	}
}

func TestL03OldLaunchCannotReplaceNewAndStop(t *testing.T) {
	oldReady := make(chan struct{})
	first := stage06Fake(11)
	second := stage06Fake(22)
	var calls atomic.Int32
	s := stage06Source(t, func(ctx context.Context, d asp.Descriptor) (aspProcess, error) {
		calls.Add(1)
		if calls.Load() == 1 {
			// Deliberately ignore cancellation: late old READY.
			<-oldReady
			return first, nil
		}
		return second, nil
	})
	s.captureNative(stage06Block())
	stage06Await(t, func() bool { return calls.Load() == 1 })
	s.SetASP(config.ASPBuiltin)
	s.SetASP(config.ASPExternal)
	s.captureNative(stage06Block())
	stage06Await(t, func() bool { return stage06Ready(s) })
	close(oldReady)
	stage06Await(t, func() bool {
		select {
		case <-first.done:
			return true
		default:
			return false
		}
	})
	r := s.captureNative(stage06Block())
	if r == nil || r.process != second {
		t.Fatal("stale launch replaced current session")
	}
	s.StopASP()
	select {
	case <-second.canceled:
	default:
		t.Fatal("shutdown failed to cancel current process")
	}
	s.WaitASP()
}

func TestL04StalePublicationAndExternalOwnership(t *testing.T) {
	p := stage06Fake(1593)
	s := stage06Source(t, func(context.Context, asp.Descriptor) (aspProcess, error) { return p, nil })
	s.captureNative(stage06Block())
	stage06Await(t, func() bool { return stage06Ready(s) })
	out, unsub := s.Listen("owned")
	defer unsub()
	r := s.captureNative(stage06Block())
	if r == nil {
		t.Fatal("no External response")
	}
	s.broadcastSelected(stage06Stock(), false, r)
	a := <-out
	saved := slices.Clone(a)
	for i := 0; i < 3; i++ {
		s.broadcastSelected(stage06Stock(), false, s.captureNative(stage06Block()))
		<-out
	}
	if !slices.Equal(a, saved) || a[0] != 1593 {
		t.Fatal("old External output mutated")
	}
	late := s.captureNative(stage06Block())
	s.SetASP(config.ASPBuiltin)
	s.broadcastSelected(stage06Stock(), false, late)
	stale := <-out
	if stale[0] == 1593 || s.ActiveASP() != config.ASPBuiltin || !slices.Equal(a, saved) {
		t.Fatal("stale External result published after selection changed")
	}
	if h := s.Recent(20 * time.Millisecond); h[0] == 1593 {
		t.Fatal("stale External leaked into history")
	}
}

func TestL05CloseDoesNotCancelSession(t *testing.T) {
	p := stage06Fake(37)
	s := stage06Source(t, func(context.Context, asp.Descriptor) (aspProcess, error) { return p, nil })
	s.captureNative(stage06Block())
	stage06Await(t, func() bool { return stage06Ready(s) })
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	select {
	case <-p.canceled:
		t.Fatal("capture Close killed resident helper")
	default:
	}
	if s.captureNative(stage06Block()) == nil {
		t.Fatal("reopened capture cannot use same helper")
	}
	s.StopASP()
	select {
	case <-p.canceled:
	default:
		t.Fatal("final stop did not cancel")
	}
	if err := s.Start(context.Background()); !errors.Is(err, context.Canceled) {
		t.Fatalf("new capture accepted after stop: %v", err)
	}
}

func TestL06FinalStopUnblocksExchangeBeforeWait(t *testing.T) {
	block := make(chan struct{})
	p := stage06Fake(4)
	p.wait = block
	p.honorCancel = true
	s := stage06Source(t, func(context.Context, asp.Descriptor) (aspProcess, error) { return p, nil })
	s.captureNative(stage06Block())
	stage06Await(t, func() bool { return stage06Ready(s) })
	done := make(chan struct{})
	go func() { defer close(done); s.captureNative(stage06Block()) }()
	// The Exchange may race with shutdown, but must not prevent reaping.
	s.StopASP()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("blocked exchange survived cancellation")
	}
	s.WaitASP()
	select {
	case <-p.done:
	default:
		t.Fatal("sole process cleanup not joined")
	}
}

func TestL07NativeSoftwareCutAndStockRaw(t *testing.T) {
	raw := stage06Stock()
	for i := range raw {
		raw[i] = 0x6b
	}
	fake := &fakePlainCapture{cfg: alsaConfig(), bytes: raw}
	dst := make([]byte, len(raw))
	visited := false
	_, cut, err := readCaptured(fake, dst, func() bool { return true }, func(b NativeBlock) {
		visited = true
		for _, native := range b.Buffers {
			for _, v := range native {
				if v != 0 {
					t.Fatal("native MIC not muted")
				}
			}
		}
	})
	if err != nil || !cut || !visited {
		t.Fatal("cut native capture failed", err)
	}
	for _, v := range dst {
		if v != 0 {
			t.Fatal("stock mute changed")
		}
	}
}

// The stock capture reader can report an ALSA overrun and then exit. Neither
// situation may terminate/relaunch the resident helper (Rewire uses Close).
type stage06OverrunCapture struct{ reads int }

func (c *stage06OverrunCapture) Read(buf []byte) (int, error) {
	c.reads++
	if c.reads == 1 {
		return 0, alsa.ErrOverrun
	}
	return 0, errors.New("simulated capture failure")
}
func (*stage06OverrunCapture) Close() error        { return nil }
func (*stage06OverrunCapture) Config() alsa.Config { return alsaConfig() }
func TestL05RunExitAndXRUNPreserveResidentProcess(t *testing.T) {
	p := stage06Fake(55)
	s := stage06Source(t, func(context.Context, asp.Descriptor) (aspProcess, error) { return p, nil })
	s.captureNative(stage06Block())
	stage06Await(t, func() bool { return stage06Ready(s) })
	capture := &stage06OverrunCapture{}
	s.devMu.Lock()
	s.pcm = capture
	s.devMu.Unlock()
	if err := s.Run(context.Background()); err == nil {
		t.Fatal("capture failure suppressed")
	}
	if capture.reads != 2 {
		t.Fatalf("xrun should continue stock capture, got %d reads", capture.reads)
	}
	select {
	case <-p.canceled:
		t.Fatal("capture Run/overrun terminated helper")
	default:
	}
	if got := s.captureNative(stage06Block()); got == nil {
		t.Fatal("resident session was not preserved")
	}
}

// A malicious/late PCM reply from a canceled old process must lose against a
// fresh manual session, even when its Exchange ignores cancellation entirely.
func TestL03LatePCMReadAfterManualReplacement(t *testing.T) {
	release := make(chan struct{})
	old := stage06Fake(111)
	old.wait = release
	old.entered = make(chan struct{}, 1)
	fresh := stage06Fake(222)
	var launched atomic.Int32
	s := stage06Source(t, func(context.Context, asp.Descriptor) (aspProcess, error) {
		if launched.Add(1) == 1 {
			return old, nil
		}
		return fresh, nil
	})
	s.captureNative(stage06Block())
	stage06Await(t, func() bool { return stage06Ready(s) })
	ch, unsub := s.Listen("late-source")
	defer unsub()
	staleDone := make(chan *aspResponse, 1)
	go func() { staleDone <- s.captureNative(stage06Block()) }()
	select {
	case <-old.entered:
	case <-time.After(time.Second):
		t.Fatal("old exchange never started")
	}
	s.SetASP(config.ASPBuiltin)
	s.SetASP(config.ASPExternal)
	s.captureNative(stage06Block())
	stage06Await(t, func() bool { return stage06Ready(s) })
	close(release)
	var stale *aspResponse
	select {
	case stale = <-staleDone:
	case <-time.After(time.Second):
		t.Fatal("stale exchange blocked")
	}
	s.broadcastSelected(stage06Stock(), false, stale)
	if got := <-ch; got[0] == 111 || s.ActiveASP() != config.ASPBuiltin {
		t.Fatal("stale PCM published")
	}
	r := s.captureNative(stage06Block())
	s.broadcastSelected(stage06Stock(), false, r)
	if got := <-ch; got[0] != 222 || s.ActiveASP() != config.ASPExternal {
		t.Fatal("old cleanup replaced current helper")
	}
}

func TestL03SwitchCancelsBlockedPCM(t *testing.T) {
	blocked := make(chan struct{})
	p := stage06Fake(23)
	p.wait = blocked
	p.honorCancel = true
	p.entered = make(chan struct{}, 1)
	s := stage06Source(t, func(context.Context, asp.Descriptor) (aspProcess, error) { return p, nil })
	s.captureNative(stage06Block())
	stage06Await(t, func() bool { return stage06Ready(s) })
	finished := make(chan *aspResponse, 1)
	go func() { finished <- s.captureNative(stage06Block()) }()
	select {
	case <-p.entered:
	case <-time.After(time.Second):
		t.Fatal("exchange not entered")
	}
	s.SetASP(config.ASPBuiltin)
	select {
	case response := <-finished:
		if response != nil {
			t.Fatal("canceled PCM accepted")
		}
	case <-time.After(time.Second):
		t.Fatal("mode change did not unblock PCM")
	}
	if s.ActiveASP() != config.ASPBuiltin {
		t.Fatal("late activation")
	}
}

func TestL02PCMFailureFallbackAndManualOnlyRetry(t *testing.T) {
	var starts atomic.Int32
	s := stage06Source(t, func(context.Context, asp.Descriptor) (aspProcess, error) {
		starts.Add(1)
		p := stage06Fake(31)
		p.fail = errors.New("simulated helper EOF")
		return p, nil
	})
	s.captureNative(stage06Block())
	stage06Await(t, func() bool { return stage06Ready(s) })
	if r := s.captureNative(stage06Block()); r != nil {
		t.Fatal("failed exchange published")
	}
	if s.ActiveASP() != config.ASPBuiltin || s.SelectedASP() != config.ASPExternal {
		t.Fatal("fallback altered Selected/Active")
	}
	for i := 0; i < 4; i++ {
		s.captureNative(stage06Block())
	}
	if starts.Load() != 1 {
		t.Fatal("automatic retry after PCM failure")
	}
	s.SetASP(config.ASPBuiltin)
	s.SetASP(config.ASPExternal)
	s.captureNative(stage06Block())
	stage06Await(t, func() bool { return stage06Ready(s) })
	if starts.Load() != 2 {
		t.Fatal("manual retry did not launch fresh session")
	}
}

type stage06PartialCapture struct{ cause error }

func (c *stage06PartialCapture) Read(dst []byte) (int, error) {
	copy(dst, []byte{1, 2, 3, 4, 5, 6})
	return 6, c.cause
}
func (*stage06PartialCapture) Close() error        { return nil }
func (*stage06PartialCapture) Config() alsa.Config { return alsaConfig() }

func TestL08PartialHardwareReadIsNotExternalGeometryFallback(t *testing.T) {
	p := stage06Fake(18)
	s := stage06Source(t, func(context.Context, asp.Descriptor) (aspProcess, error) { return p, nil })
	s.captureNative(stage06Block())
	stage06Await(t, func() bool { return stage06Ready(s) })
	failure := errors.New("actual capture driver failed")
	s.devMu.Lock()
	s.pcm = &stage06PartialCapture{cause: failure}
	s.devMu.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := s.Run(ctx); !errors.Is(err, failure) {
		t.Fatalf("capture read error masked as External geometry: %v", err)
	}
	select {
	case <-p.canceled:
		t.Fatal("capture failure incorrectly killed resident ASP")
	default:
	}
}
