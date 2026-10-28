package mic

import (
	"context"
	"slices"
	"testing"
	"time"

	"github.com/HuskerMinion/techo5/echod/internal/config"
)

func TestASP2AlternatingSessionLag(t *testing.T) {
	s := controlSource(t, "factory", config.ASPExternal, nativeASPConfig(23))
	s.mixer = Center{}
	ch, stop := s.Listen("mixed-session MWW")
	defer stop()
	tick := 0
	for cycle := 0; cycle < 12; cycle++ {
		s.SetASP(config.ASPBuiltin)
		mode := "factory"
		if cycle%2 == 1 {
			mode = "normal"
		}
		s.aspStart = func(ctx context.Context) (*externalASP, error) {
			return startTestASP(ctx, childCommand(mode), nativeASPConfig(23))
		}
		s.SetASP(config.ASPExternal)
		waitASPChild(t, s)
		child := s.asp
		for local := 0; local < 8; local++ {
			raw, n := markerInput(tick)
			s.broadcastNative(raw, n)
			got := <-ch
			activeAt := 4
			if mode == "normal" {
				activeAt = 3
			}
			if local < activeAt {
				checkMarker(t, got, uint64(tick*320))
			} else if mode == "factory" {
				checkMarker(t, got, uint64((tick-1)*320))
			} else if !slices.Equal(got, fixturePCM(n, uint64(local))) {
				t.Fatal("same-tick session inherited old origin", cycle, local)
			}
			if (s.ASPStatus().Active == config.ASPExternal) != (local >= activeAt) {
				t.Fatal("session failed", cycle, local, s.ASPStatus())
			}
			tick++
		}
		s.SetASP(config.ASPBuiltin)
		raw, n := markerInput(tick)
		s.broadcastNative(raw, n)
		checkMarker(t, <-ch, uint64(tick*320))
		tick++
		select {
		case <-child.done:
		case <-time.After(time.Second):
			t.Fatal("retired child not reaped")
		}
	}
	if s.Dropped() != 0 {
		t.Fatal("subscriber drops", s.Dropped())
	}
	t.Log("12 alternating lag320/lag0 sessions: preflight, active, current stock and fresh origin PASS")
}

func TestASP2StartupAlwaysCurrent(t *testing.T) {
	s := controlSource(t, "start-hang", config.ASPBuiltin, nativeASPConfig(23))
	s.mixer = Center{}
	ch, stop := s.Listen("startup current stock")
	defer stop()
	s.SetASP(config.ASPExternal)
	for tick := 0; tick < 8; tick++ {
		raw, n := markerInput(tick)
		s.broadcastNative(raw, n)
		checkMarker(t, <-ch, uint64(tick*320))
		if s.ASPStatus().Active != config.ASPBuiltin {
			t.Fatal("HELLO faked PCM readiness")
		}
	}
	s.SetASP(config.ASPBuiltin)
	raw, n := markerInput(8)
	s.broadcastNative(raw, n)
	checkMarker(t, <-ch, 8*320)
}

func TestASP2CanceledResultCannotChangeNewSession(t *testing.T) {
	s := controlSource(t, "factory", config.ASPExternal, nativeASPConfig(23))
	s.mixer = Center{}
	s.SetASP(config.ASPExternal)
	waitASPChild(t, s)
	// A full validated response was read, but a newer manual command wins before publication.
	_, n := markerInput(0)
	old := s.processASP(n)
	s.SetASP(config.ASPBuiltin)
	s.SetASP(config.ASPExternal)
	waitASPChild(t, s)
	fresh := s.aspAttempt
	before := s.ASPStatus()
	stock := make([]int16, FrameSamples)
	for j := range stock {
		stock[j] = int16(j*19 + 23)
	}
	s.aspMu.Lock()
	got := s.chooseASP(stock, old)
	s.aspMu.Unlock()
	s.retireASP(old)
	if !slices.Equal(got, stock) || s.ASPStatus() != before || fresh.next != 0 || fresh.good != 0 || fresh.based {
		t.Fatal("canceled child altered new origin/status", s.ASPStatus())
	}
	ch, stop := s.Listen("fresh after stale result")
	defer stop()
	for tick := 1; tick < 8; tick++ {
		raw, n := markerInput(tick)
		s.broadcastNative(raw, n)
		start := tick * 320
		if tick >= 5 {
			start -= 320
		}
		checkMarker(t, <-ch, uint64(start))
	}
	if s.ASPStatus().Active != config.ASPExternal {
		t.Fatal("new child not active", s.ASPStatus())
	}
}
