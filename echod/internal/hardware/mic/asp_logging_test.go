package mic

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/HuskerMinion/techo5/echod/internal/asp"
	"github.com/HuskerMinion/techo5/echod/internal/config"
)

func testLogDispatch(t testing.TB, write func(aspLogRecord)) *aspLogDispatch {
	t.Helper()
	d := newASPLogDispatch(write)
	t.Cleanup(func() {
		close(d.stop)
		select {
		case <-d.done:
		case <-time.After(time.Second):
			t.Error("logging consumer did not finish")
		}
	})
	return d
}
func drainLog(t testing.TB, d *aspLogDispatch, want uint64) {
	t.Helper()
	until := time.Now().Add(3 * time.Second)
	for d.snapshot().Logged < want {
		if time.Now().After(until) {
			t.Fatal("logs not drained", d.snapshot(), want)
		}
		runtime.Gosched()
	}
}
func TestASPLoggingAndHAPriority(t *testing.T) {
	var mu sync.Mutex
	var records []aspLogRecord
	d := testLogDispatch(t, func(r aspLogRecord) { mu.Lock(); records = append(records, r); mu.Unlock() })
	s := testSource()
	s.aspLog = d
	s.aspStatus.Request = 1
	s.aspLogRequest.Store(1)
	a := &aspAttempt{}
	now := time.Unix(100, 0)
	s.aspNow = func() time.Time { return now }
	emit := func(st asp.Status) {
		s.aspMu.Lock()
		s.feedbackASPLocked(a, st, "External ASP processing OK")
		s.aspMu.Unlock()
	}
	emit(asp.Status{Kind: asp.State, Code: 200, Message: "Started"})
	emit(asp.Status{Kind: asp.State, Code: 200, Message: "Started"})
	if s.ASPStatus().Debug != "200: Started" {
		t.Fatal(s.ASPStatus())
	}
	emit(asp.Status{Kind: asp.Event, Code: 210, Message: "Simulated beam changed"})
	emit(asp.Status{Kind: asp.Event, Code: 210, Message: "Simulated beam changed"})
	if s.ASPStatus().Debug != "210: Simulated beam changed" {
		t.Fatal(s.ASPStatus())
	}
	emit(asp.Status{Kind: asp.State, Code: 301, Message: "Noise suppressor disabled"})
	emit(asp.Status{Kind: asp.Event, Code: 210, Message: "Simulated event during warning"})
	if s.ASPStatus().Debug != "301: Noise suppressor disabled" {
		t.Fatal("event hid warning")
	}
	emit(asp.Status{Kind: asp.State, Code: 422, Message: "Reduced array mode active"})
	emit(asp.Status{Kind: asp.Event, Code: 0, Message: "Informational observation"})
	if s.ASPStatus().Debug != "422: Reduced array mode active" {
		t.Fatal("event hid configuration warning")
	}
	emit(asp.Status{Kind: asp.State, Code: 201})
	if s.ASPStatus().Debug != "201: External ASP processing OK" {
		t.Fatal("no explicit recovery")
	}
	now = now.Add(6 * time.Second)
	emit(asp.Status{Kind: asp.Event, Code: 210, Message: "Simulated new event"})
	drainLog(t, d, 9)
	mu.Lock()
	defer mu.Unlock()
	if len(records) != 9 {
		t.Fatal(len(records))
	}
	if records[0].level != slog.LevelInfo || records[3].level != slog.LevelWarn || records[5].level != slog.LevelWarn {
		t.Fatal(records)
	}
	for _, r := range records {
		if len(r.code) != 3 || r.kind == "" || r.origin != "external-asp" || r.message == "" {
			t.Fatal("missing structured attrs", r)
		}
	}
	if s.aspEventsSuppressed != 3 {
		t.Fatal("unexpected HA suppression", s.aspEventsSuppressed)
	}
	t.Logf("meaningful_helper_records=9 logged=%d HA_events=2 HA_suppressed=3 repeated_STATE_logs=0 duplicate_EVENT_logs=2", d.snapshot().Logged)
}

type blockedASPHandler struct {
	entered, release chan struct{}
	mu               sync.Mutex
	records          []slog.Record
	once             sync.Once
}

func (h *blockedASPHandler) Enabled(context.Context, slog.Level) bool { return true }
func (h *blockedASPHandler) Handle(_ context.Context, r slog.Record) error {
	block := false
	r.Attrs(func(a slog.Attr) bool {
		if a.Key == "asp_message" && a.Value.String() == "Block logging" {
			block = true
		}
		return true
	})
	if block {
		h.once.Do(func() { close(h.entered) })
		<-h.release
	}
	h.mu.Lock()
	h.records = append(h.records, r.Clone())
	h.mu.Unlock()
	return nil
}
func (h *blockedASPHandler) WithAttrs([]slog.Attr) slog.Handler { return h }
func (h *blockedASPHandler) WithGroup(string) slog.Handler      { return h }
func TestASPLoggingBlockedHandlerFatalPriority(t *testing.T) {
	h := &blockedASPHandler{entered: make(chan struct{}), release: make(chan struct{})}
	previous := slog.Default()
	slog.SetDefault(slog.New(h))
	defer slog.SetDefault(previous)
	released := false
	defer func() {
		if !released {
			close(h.release)
		}
	}()
	d := testLogDispatch(t, standardASPLog)
	s := controlSource(t, "normal", config.ASPExternal)
	s.aspLog = d
	s.aspStart = func(ctx context.Context) (*externalASP, error) {
		cmd := childCommand("normal")
		cmd.Env = append(cmd.Env, "TECHO5_ASP_REPLY_KIND=2", "TECHO5_ASP_REPLY_CODE=503", "TECHO5_ASP_REPLY_MSG=ASP processing failed", "TECHO5_ASP_REPLY_ON=6")
		return startTestASP(ctx, cmd, nativeASPConfig(23))
	}
	s.SetASP(config.ASPExternal)
	waitASPChild(t, s)
	s.aspMu.Lock()
	child := s.asp
	s.aspMu.Unlock()
	d.enqueue(aspLogRecord{level: slog.LevelInfo, code: "210", kind: "EVENT", origin: "external-asp", message: "Block logging"})
	select {
	case <-h.entered:
	case <-time.After(time.Second):
		t.Fatal("handler did not block")
	}
	for i := 0; i < 200; i++ {
		d.enqueue(aspLogRecord{level: slog.LevelInfo, code: "210", kind: "EVENT", origin: "external-asp", message: "Simulated burst"})
	}
	before := runtime.NumGoroutine()
	var max time.Duration
	for tick := 0; tick < 8; tick++ {
		start := time.Now()
		s.broadcastNative(rawFrame(tick, false), nativePattern(23, tick))
		max = maxDuration(max, time.Since(start))
	}
	st := s.ASPStatus()
	if st.Debug != "503: ASP processing failed" || st.Active != config.ASPBuiltin || st.Selected != config.ASPExternal || config.Get().Microphone.ASP != config.ASPExternal {
		t.Fatal(st)
	}
	select {
	case <-child.done:
	case <-time.After(time.Second):
		t.Fatal("blocked logger prevented reap")
	}
	stats := d.snapshot()
	if stats.Dropped < 136 || stats.Depth > aspLogCapacity+1 || runtime.NumGoroutine() > before+2 {
		t.Fatal("unbounded dispatch", stats, before, runtime.NumGoroutine())
	}
	if max > aspFrameBudget+30*time.Millisecond {
		t.Fatal("logger stalled capture", max)
	}
	close(h.release)
	released = true
	drainLog(t, d, 4)
	h.mu.Lock()
	defer h.mu.Unlock()
	foundError := false
	blockAt, errorAt := -1, -1
	for i, r := range h.records {
		r.Attrs(func(a slog.Attr) bool {
			if a.Key == "asp_message" && a.Value.String() == "Block logging" {
				blockAt = i
			}
			if a.Key == "asp_status_code" && a.Value.String() == "503" {
				if r.Level != slog.LevelError {
					t.Fatal("fatal level")
				}
				foundError = true
				errorAt = i
			}
			return true
		})
	}
	if !foundError || errorAt != blockAt+1 {
		t.Fatal("fatal not first after blocked I/O", blockAt, errorAt)
	}
	t.Logf("blocked_slog_capture_max_ns=%d capacity=%d accepted=%d overflow_or_preempt_drops=%d SIGKILL_reaped=true fatal_log_next_after_unblock=true", max.Nanoseconds(), aspLogCapacity, stats.Accepted, stats.Dropped)
}
func maxDuration(a, b time.Duration) time.Duration {
	if a > b {
		return a
	}
	return b
}
func TestASPLoggingVirtual125PerSecondAndStale(t *testing.T) {
	var mu sync.Mutex
	count := 0
	d := testLogDispatch(t, func(aspLogRecord) { mu.Lock(); count++; mu.Unlock() })
	s := testSource()
	s.aspLog = d
	s.aspStatus.Request = 1
	s.aspLogRequest.Store(1)
	a := &aspAttempt{}
	clock := time.Unix(1000, 0)
	s.aspNow = func() time.Time { return clock }
	publications := 0
	previous := ""
	for n := 0; n < 15000; n++ {
		s.aspMu.Lock()
		s.feedbackASPLocked(a, asp.Status{Kind: asp.Event, Code: 210, Message: fmt.Sprintf("Simulated observation %d", n)}, "External ASP event")
		s.aspMu.Unlock()
		current := s.ASPStatus().Debug
		if current != previous {
			publications++
			previous = current
		}
		clock = clock.Add(8 * time.Millisecond)
		// Separate simulated pacing from saturation; drain each accepted record.
		drainLog(t, d, uint64(n+1))
	}
	stats := d.snapshot()
	if publications != 24 || s.aspEventsSuppressed != 14976 || stats.Logged != 15000 || stats.Dropped != 0 {
		t.Fatal(publications, s.aspEventsSuppressed, stats)
	}
	old := s.aspStatus.Request
	s.aspLogRequest.Store(old + 1)
	d.enqueue(aspLogRecord{level: slog.LevelInfo, code: "210", kind: "EVENT", origin: "external-asp", message: "Old child observation", request: old, current: &s.aspLogRequest})
	until := time.Now().Add(time.Second)
	for d.snapshot().Stale < 1 {
		if time.Now().After(until) {
			t.Fatal("stale log not rejected")
		}
		runtime.Gosched()
	}
	t.Logf("virtual_seconds=120 dispatcher_rate=125 accepted=15000 logged=15000 HA_events=%d HA_suppressed=%d queue_dropped=0 stale_dropped=1", publications, s.aspEventsSuppressed)
}
func TestASPLoggingReal125PerSecond(t *testing.T) {
	if os.Getenv("TECHO5_ASP_STRESS") != "1" {
		t.Skip("120-second paced dispatcher test is an explicit separate stress gate")
	}
	d := testLogDispatch(t, func(aspLogRecord) {})
	s := testSource()
	s.aspLog = d
	s.aspStatus.Request = 1
	s.aspLogRequest.Store(1)
	a := &aspAttempt{}
	ticker := time.NewTicker(8 * time.Millisecond)
	defer ticker.Stop()
	started := time.Now()
	publications := 0
	previous := ""
	for n := 0; n < 15000; n++ {
		<-ticker.C
		s.aspMu.Lock()
		s.feedbackASPLocked(a, asp.Status{Kind: asp.Event, Code: 210, Message: fmt.Sprintf("Simulated paced observation %d", n)}, "External ASP event")
		s.aspMu.Unlock()
		current := s.ASPStatus().Debug
		if current != previous {
			publications++
			previous = current
		}
	}
	drainLog(t, d, 15000)
	stats := d.snapshot()
	if stats.Dropped != 0 || publications > 25 {
		t.Fatal(stats, publications)
	}
	t.Logf("real_seconds=%.3f accepted=%d logged=%d HA_events=%d HA_suppressed=%d queue_max=%d dropped=%d", time.Since(started).Seconds(), stats.Accepted, stats.Logged, publications, s.aspEventsSuppressed, stats.MaxDepth, stats.Dropped)
}
func TestASPLogSanitizationAndNoFeedback(t *testing.T) {
	for _, message := range []string{"/private/recording", "<b>markup</b>", "line\nspoof", "192.0.2.1"} {
		st, e := (asp.Status{Kind: asp.Event, Code: 210, Message: message}).Validate()
		if e != nil || st.Message != "" {
			t.Fatal(st, e)
		}
	}
	var n int
	var mu sync.Mutex
	d := testLogDispatch(t, func(r aspLogRecord) {
		mu.Lock()
		n++
		mu.Unlock()
		if strings.ContainsAny(r.message, "<>/\n") {
			t.Error("unsafe log")
		}
	})
	s := testSource()
	s.aspLog = d
	a := &aspAttempt{}
	for i := 0; i < 1000; i++ {
		s.aspMu.Lock()
		s.feedbackASPLocked(a, asp.Status{}, "unused")
		s.aspMu.Unlock()
	}
	time.Sleep(time.Millisecond)
	mu.Lock()
	defer mu.Unlock()
	if n != 0 || d.snapshot().Accepted != 0 {
		t.Fatal("logged empty feedback")
	}
}

// Reaping must finish even when every enabled backend write, including DEBUG, blocks.
type p3AllBlockedASPHandler struct {
	entered, release chan struct{}
	calls            atomic.Uint64
}

func (h *p3AllBlockedASPHandler) Enabled(context.Context, slog.Level) bool { return true }
func (h *p3AllBlockedASPHandler) Handle(context.Context, slog.Record) error {
	if h.calls.Add(1) == 1 {
		close(h.entered)
	}
	<-h.release
	return nil
}
func (h *p3AllBlockedASPHandler) WithAttrs([]slog.Attr) slog.Handler { return h }
func (h *p3AllBlockedASPHandler) WithGroup(string) slog.Handler      { return h }
func TestASPLoggingReaperHasNoBackendIO(t *testing.T) {
	h := &p3AllBlockedASPHandler{entered: make(chan struct{}), release: make(chan struct{})}
	previous := slog.Default()
	slog.SetDefault(slog.New(h))
	defer slog.SetDefault(previous)
	defer close(h.release)
	d := testLogDispatch(t, standardASPLog)
	d.enqueue(aspLogRecord{code: "210", kind: "EVENT", origin: "external-asp", message: "Simulated blocked backend"})
	select {
	case <-h.entered:
	case <-time.After(time.Second):
		t.Fatal("handler not blocked")
	}
	runtime.GC()
	before := runtime.NumGoroutine()
	for i := 0; i < 16; i++ {
		f := configuredChild(t, "normal", nativeASPConfig(31))
		f.Close()
		select {
		case <-f.done:
		case <-time.After(time.Second):
			t.Fatal("child not reaped")
		}
	}
	time.Sleep(20 * time.Millisecond)
	growth := runtime.NumGoroutine() - before
	t.Logf("reaped_children=16 blocked_backend_calls=%d goroutine_growth=%d", h.calls.Load(), growth)
	if h.calls.Load() != 1 || growth > 2 {
		t.Fatal("reaper bypassed bounded logging dispatcher")
	}
}
