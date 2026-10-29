package mic

import (
	"github.com/HuskerMinion/techo5/echod/internal/asp"
	"log/slog"
	"runtime"
	"slices"
	"testing"
	"time"
)

func TestASP2QuantumSchedulingAndHardDeadline(t *testing.T) {
	f := configuredChild(t, "factory", aspShape(4, 2), "TECHO5_ASP_QUANTUM_DELAY_MS=6")
	input := aspInput(aspShape(4, 2))
	stop := make(chan struct{})
	done := make(chan struct{})
	go func() {
		defer close(done)
		var x uint64 = 3
		for {
			select {
			case <-stop:
				return
			default:
			}
			for j := 0; j < 10000; j++ {
				x = x*1664525 + 1013904223
			}
			runtime.KeepAlive(x)
		}
	}()
	defer func() { close(stop); <-done }()
	var ns []int64
	for i := 0; i < 60; i++ {
		start := time.Now()
		r, e := f.exchange(input)
		ns = append(ns, time.Since(start).Nanoseconds())
		if e != nil || i == 0 && r.Type != asp.NotReady || i > 0 && r.Type != asp.Data {
			t.Fatal("2/3 quantum scheduling jitter rejected", i, e)
		}
	}
	slices.Sort(ns)
	t.Logf("factory_2or3_quanta_ns p50=%d p95=%d p99=%d max=%d accepted=60", ns[30], ns[57], ns[59], ns[59])
	hard := configuredChild(t, "factory", aspShape(4, 2), "TECHO5_ASP_QUANTUM_DELAY_MS=10")
	if r, e := hard.exchange(input); e != nil || r.Type != asp.NotReady {
		t.Fatal("two 10ms quanta should fit provisional budget", e)
	}
	start := time.Now()
	r, e := hard.exchange(input)
	elapsed := time.Since(start)
	if e == nil || r.PCM != nil || !hard.broken.Load() || elapsed > 150*time.Millisecond {
		t.Fatal("three 10ms quanta were not terminated", elapsed, e)
	}
	t.Logf("factory_hard_3quanta_ns=%d configured_deadline_ms=25", elapsed.Nanoseconds())
}
func TestASPWarningNeedsExplicitRecovery(t *testing.T) {
	d := testLogDispatch(t, func(aspLogRecord) {})
	s := testSource()
	s.aspLog = d
	a := &aspAttempt{}
	for _, st := range []asp.Status{{Kind: asp.State, Code: 301, Message: "Noise suppressor disabled"}, {Kind: asp.State, Code: 250, Message: "Informational text only"}, {Kind: asp.State, Code: 250, Message: "Informational text only"}} {
		s.aspMu.Lock()
		s.feedbackASPLocked(a, st, "External ASP processing OK")
		s.aspMu.Unlock()
	}
	if s.ASPStatus().Debug != "E:250:Informational text only" {
		t.Fatal("explicit information state was suppressed")
	}
	s.aspMu.Lock()
	s.feedbackASPLocked(a, asp.Status{Kind: asp.State, Code: 200}, "External ASP processing OK")
	s.aspMu.Unlock()
	if s.ASPStatus().Debug != "E:200:External ASP processing OK" {
		t.Fatal("explicit recovery ignored")
	}
	drainLog(t, d, 3)
}
func BenchmarkASPLoggingBoundedEnqueue(b *testing.B) {
	entered := make(chan struct{})
	release := make(chan struct{})
	d := newASPLogDispatch(func(aspLogRecord) {
		select {
		case <-entered:
		default:
			close(entered)
		}
		<-release
	})
	d.enqueue(aspLogRecord{level: slog.LevelInfo})
	<-entered
	b.Cleanup(func() { close(release); close(d.stop); <-d.done })
	r := aspLogRecord{level: slog.LevelInfo, code: "210", kind: "EVENT", origin: "external-asp", message: "Simulated observation"}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d.enqueue(r)
	}
	b.StopTimer()
	if s := d.snapshot(); s.Depth > aspLogCapacity {
		b.Fatal("unbounded queue", s)
	}
}
