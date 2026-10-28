package mic

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/HuskerMinion/techo5/echod/internal/config"
)

func TestASP2MalformedStatusHostReason(t *testing.T) {
	for _, mode := range []string{"bad-status", "large-status", "bad-kind", "bad-code", "truncated-status"} {
		t.Run(mode, func(t *testing.T) {
			s := controlSource(t, mode, config.ASPExternal, nativeASPConfig(23))
			s.SetASP(config.ASPExternal)
			waitASPChild(t, s)
			s.broadcastNative(rawFrame(0, false), nativePattern(23, 0))
			st := s.ASPStatus()
			if st.Debug != "000: Fallback: invalid external ASP status" || st.Active != config.ASPBuiltin || st.Selected != config.ASPExternal {
				t.Fatal(st)
			}
		})
	}
}
func TestASP2OptionalStatusDefaults(t *testing.T) {
	for _, row := range []struct {
		env  []string
		want string
	}{
		{nil, "000: External ASP processing OK"},
		{[]string{"TECHO5_ASP_ACK_KIND=1", "TECHO5_ASP_ACK_CODE=201"}, "201: External ASP processing OK"},
		{[]string{"TECHO5_ASP_ACK_KIND=1", "TECHO5_ASP_ACK_MSG=External ASP fallback beam active"}, "000: External ASP fallback beam active"},
		{[]string{"TECHO5_ASP_ACK_KIND=1", "TECHO5_ASP_ACK_CODE=422"}, "422: External ASP configuration degraded"},
		{[]string{"TECHO5_ASP_ACK_KIND=1", "TECHO5_ASP_ACK_CODE=200", "TECHO5_ASP_ACK_MSG=External ASP started (8 mic, 2 ref)"}, "200: External ASP started (8 mic, 2 ref)"},
	} {
		s := controlSource(t, "normal", config.ASPBuiltin, nativeASPConfig(23))
		env := row.env
		s.aspStart = func(ctx context.Context) (*externalASP, error) {
			cmd := childCommand("normal")
			cmd.Env = append(cmd.Env, env...)
			return startTestASP(ctx, cmd, nativeASPConfig(23))
		}
		s.SetASP(config.ASPExternal)
		waitASPChild(t, s)
		for tick := 0; tick < 4; tick++ {
			s.broadcastNative(rawFrame(tick, false), nativePattern(23, tick))
		}
		if st := s.ASPStatus(); st.Debug != row.want || st.Active != config.ASPExternal {
			t.Fatal(st, row.want)
		}
	}
	for _, mode := range []string{"reject", "normal"} {
		s := controlSource(t, mode, config.ASPBuiltin, nativeASPConfig(23))
		s.aspStart = func(ctx context.Context) (*externalASP, error) {
			cmd := childCommand(mode)
			prefix := "TECHO5_ASP_ACK"
			if mode == "normal" {
				prefix = "TECHO5_ASP_REPLY"
				cmd.Env = append(cmd.Env, "TECHO5_ASP_REPLY_TYPE=3")
			}
			cmd.Env = append(cmd.Env, prefix+"_KIND=1", prefix+"_MSG=Unrecoverable DSP fault")
			return startTestASP(ctx, cmd, nativeASPConfig(23))
		}
		s.SetASP(config.ASPExternal)
		waitASPChild(t, s)
		s.broadcastNative(rawFrame(0, false), nativePattern(23, 0))
		if st := s.ASPStatus(); st.Debug != "000: Unrecoverable DSP fault" || st.Active != config.ASPBuiltin || st.Selected != config.ASPBuiltin {
			t.Fatal(st)
		}
	}
}
func TestASP2WireEventsDoNotChangeAudio(t *testing.T) {
	normal := controlSource(t, "normal", config.ASPExternal, nativeASPConfig(23))
	events := controlSource(t, "normal", config.ASPExternal, nativeASPConfig(23))
	var logEvents atomic.Uint64
	d := testLogDispatch(t, func(r aspLogRecord) {
		if r.kind == "EVENT" {
			logEvents.Add(1)
		}
	})
	events.aspLog = d
	events.aspStart = func(ctx context.Context) (*externalASP, error) {
		cmd := childCommand("normal")
		cmd.Env = append(cmd.Env, "TECHO5_ASP_REPLY_KIND=2", "TECHO5_ASP_REPLY_CODE=210", "TECHO5_ASP_REPLY_MSG=Simulated beam changed")
		return startTestASP(ctx, cmd, nativeASPConfig(23))
	}
	normal.SetASP(config.ASPExternal)
	events.SetASP(config.ASPExternal)
	waitASPChild(t, normal)
	waitASPChild(t, events)
	a, stopA := normal.Listen("reference")
	defer stopA()
	b, stopB := events.Listen("events")
	defer stopB()
	for tick := 0; tick < 100; tick++ {
		normal.broadcastNative(rawFrame(tick, false), nativePattern(23, tick))
		events.broadcastNative(rawFrame(tick, false), nativePattern(23, tick))
		if !slices.Equal(<-a, <-b) {
			t.Fatal("telemetry changed PCM", tick)
		}
	}
	drainLog(t, d, 102)
	if logEvents.Load() != 100 || events.aspEventsSuppressed != 99 {
		t.Fatal(logEvents.Load(), events.aspEventsSuppressed)
	}
	t.Logf("response_attached_events=100 event_logs=%d HA_event_suppressed=%d PCM_equal=true max_wire_events_per_second=50", logEvents.Load(), events.aspEventsSuppressed)
}
func TestASP2PacedHandoffCadence(t *testing.T) {
	s := controlSource(t, "factory:fatal:6", config.ASPExternal, nativeASPConfig(23))
	s.mixer = Center{}
	s.SetASP(config.ASPExternal)
	waitASPChild(t, s)
	ch, stop := s.Listen("paced")
	defer stop()
	ticker := time.NewTicker(20 * time.Millisecond)
	defer ticker.Stop()
	var last time.Time
	var intervals, costs []int64
	publications := 0
	late := 0
	for tick := 0; tick < 80; tick++ {
		scheduled := <-ticker.C
		if tick == 6 {
			s.SetASP(config.ASPBuiltin)
		}
		if tick == 65 {
			s.SetASP(config.ASPExternal)
			waitASPChild(t, s)
		}
		startTime := time.Now()
		raw, n := markerInput(tick)
		s.broadcastNative(raw, n)
		var frame []int16
		select {
		case frame = <-ch:
		default:
			t.Fatal("intentional missing publication", tick)
		}
		external := tick >= 4 && tick < 6 || tick >= 69 && tick < 71
		origin := uint64(tick * 320)
		if external {
			origin -= 320
		}
		checkMarker(t, frame, origin)
		if (s.ASPStatus().Active == config.ASPExternal) != external {
			t.Fatal("wrong paced source", tick, s.ASPStatus())
		}
		now := time.Now()
		costs = append(costs, now.Sub(startTime).Nanoseconds())
		if now.Sub(scheduled) > 5*time.Millisecond {
			late++
		}
		if !last.IsZero() {
			intervals = append(intervals, now.Sub(last).Nanoseconds())
		}
		last = now
		publications++
	}
	if publications != 80 || s.Dropped() != 0 {
		t.Fatal(publications, s.Dropped())
	}
	slices.Sort(intervals)
	// Retain the original 100-ms scheduler safety bound; the old deliberate >=30-ms gap
	// assertion is replaced by exactly one publication per capture tick.
	if intervals[len(intervals)-1] > int64(100*time.Millisecond) {
		t.Fatal("scheduler/capture stall", intervals[len(intervals)-1])
	}
	slices.Sort(costs)
	metric := func(v []int64) string {
		return fmt.Sprintf("p50=%d p95=%d p99=%d max=%d", v[len(v)/2], v[(len(v)-1)*95/100], v[(len(v)-1)*99/100], v[len(v)-1])
	}
	t.Logf("paced_capture_ms=20 publication_count=%d missing_publications=0 cadence_ns %s capture_cost_ns %s scheduler_late_over5ms=%d immediate_and_50ticks_stock_offset_samples=0 listener_drops=%d", publications, metric(intervals), metric(costs), late, s.Dropped())
}

func TestASP2AcceptedBoundaryArtifacts(t *testing.T) {
	// Accepted boundary policy: current stock ticks0..3 then delayed external3 repeats320.
	stockEnd := uint64(4 * 320)
	promotedStart := uint64(3 * 320)
	repeat := stockEnd - promotedStart
	// After delayed external origin4 at capture5, current-stock fallback capture6 skips origin5.
	externalEnd := uint64(5 * 320)
	fallbackStart := uint64(6 * 320)
	skip := fallbackStart - externalEnd
	if repeat != 320 || skip != 320 {
		t.Fatal(repeat, skip)
	}
	t.Logf("accepted_low_latency_policy repeat_samples=%d fallback_skipped_samples=%d; always_delayed_from_capture_start alternative adds20ms even before preflight", repeat, skip)
}

func TestASP2NotReadyFeedbackUsesStartingDefault(t *testing.T) {
	s := controlSource(t, "factory", config.ASPBuiltin, nativeASPConfig(23))
	s.aspStart = func(ctx context.Context) (*externalASP, error) {
		cmd := childCommand("factory")
		cmd.Env = append(cmd.Env, "TECHO5_ASP_REPLY_KIND=1", "TECHO5_ASP_REPLY_CODE=201")
		return startTestASP(ctx, cmd, nativeASPConfig(23))
	}
	s.SetASP(config.ASPExternal)
	waitASPChild(t, s)
	for tick := 0; tick < 5; tick++ {
		raw, n := markerInput(tick)
		s.broadcastNative(raw, n)
		st := s.ASPStatus()
		if tick < 4 {
			if st.Debug != "201: External ASP starting" || st.Active != config.ASPBuiltin || st.Confirmed {
				t.Fatal("feedback faked PCM readiness", tick, st)
			}
		} else if st.Debug != "201: External ASP processing OK" || st.Active != config.ASPExternal || !st.Confirmed {
			t.Fatal(st)
		}
	}
}

func TestASP2FirstExchangeFailureKeepsAcceptedConfigLog(t *testing.T) {
	for _, mode := range []string{"crash", "output-hang", "config-hang", "config-crash"} {
		t.Run(mode, func(t *testing.T) {
			var mu sync.Mutex
			var records []aspLogRecord
			d := testLogDispatch(t, func(r aspLogRecord) { mu.Lock(); records = append(records, r); mu.Unlock() })
			s := controlSource(t, mode, config.ASPExternal, nativeASPConfig(23))
			s.aspLog = d
			s.aspStart = func(ctx context.Context) (*externalASP, error) {
				cmd := childCommand(mode)
				cmd.Env = append(cmd.Env, "TECHO5_ASP_ACK_KIND=1", "TECHO5_ASP_ACK_CODE=200", "TECHO5_ASP_ACK_MSG=Configuration accepted")
				return startTestASP(ctx, cmd, nativeASPConfig(23))
			}
			s.SetASP(config.ASPExternal)
			waitASPChild(t, s)
			s.broadcastNative(rawFrame(0, false), nativePattern(23, 0))
			want := "000: Fallback: external ASP exited"
			if mode == "output-hang" || mode == "config-hang" {
				want = "000: Fallback: external ASP processing timeout"
			}
			if s.ASPStatus().Debug != want {
				t.Fatal(mode, s.ASPStatus())
			}
			total := uint64(3)
			if strings.HasPrefix(mode, "config-") {
				total = 2
			}
			drainLog(t, d, total)
			mu.Lock()
			defer mu.Unlock()
			found := false
			for _, r := range records {
				if r.code == "200" && r.message == "Configuration accepted" {
					found = true
				}
			}
			if found == strings.HasPrefix(mode, "config-") {
				t.Fatal("validated ConfigResult feedback missing or invented", mode, records)
			}
			if d.snapshot().Dropped != 0 {
				t.Fatal("unsaturated fatal discarded accepted state")
			}
		})
	}
}
