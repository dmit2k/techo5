package microphone

import (
	"context"
	"path/filepath"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/HuskerMinion/techo5/echod/internal/component"
	"github.com/HuskerMinion/techo5/echod/internal/config"
	"github.com/HuskerMinion/techo5/echod/internal/hardware/mic"
	esphome "github.com/ygelfand/go-esphome-device"
)

// Source's real pipe/PCM transitions are tested in hardware/mic. This double controls notification
// ordering for the HA/store boundary, including coalescing and a completed old request.
type aspFake struct {
	mu      sync.Mutex
	state   mic.ASPStatus
	changed chan struct{}
}

func (f *aspFake) SetASP(v config.ASPMode) config.ASPMode {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.state = mic.ASPStatus{Selected: v, Active: config.ASPBuiltin, Debug: "000: Built-in processing OK", Request: f.state.Request + 1}
	if v == config.ASPExternal {
		f.state.Debug = "000: External ASP starting"
	}
	return v
}
func (f *aspFake) ASPStatus() mic.ASPStatus    { f.mu.Lock(); defer f.mu.Unlock(); return f.state }
func (f *aspFake) ASPChanges() <-chan struct{} { return f.changed }
func (f *aspFake) transition(selected, active config.ASPMode, debug string, confirmed bool) {
	f.mu.Lock()
	f.state.Selected = selected
	f.state.Active = active
	f.state.Debug = debug
	f.state.Confirmed = confirmed
	f.mu.Unlock()
	select {
	case f.changed <- struct{}{}:
	default:
	}
}

func aspFeature(t *testing.T, have bool, saved config.ASPMode) (*Microphone, *aspFake, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "state.json")
	config.Use(path)
	if err := config.Set().Microphone().ASP(saved); err != nil {
		t.Fatal(err)
	}
	f := &aspFake{state: mic.ASPStatus{Selected: config.ASPBuiltin, Active: config.ASPBuiltin, Debug: "000: Built-in processing OK"}, changed: make(chan struct{}, 1)}
	m := &Microphone{}
	m.bindASP(f, have)
	return m, f, path
}

func assertASPFeature(t *testing.T, m *Microphone, selection, active, debug string, saved config.ASPMode) {
	t.Helper()
	if m.asp.Get() != selection || m.activeASP.Get() != active || m.aspDebug.Get() != debug || config.Get().Microphone.ASP != saved {
		t.Fatalf("got select=%s active=%s debug=%s saved=%s", m.asp.Get(), m.activeASP.Get(), m.aspDebug.Get(), config.Get().Microphone.ASP)
	}
}

func TestASPFeatureFirstSaveAndExplicitBuiltin(t *testing.T) {
	m, f, path := aspFeature(t, true, config.ASPBuiltin)
	assertASPFeature(t, m, "Built-in", "Built-in", "000: Built-in processing OK", config.ASPBuiltin)
	m.asp.OnCommand("External")
	assertASPFeature(t, m, "External", "Built-in", "000: External ASP starting", config.ASPBuiltin)
	reload, err := config.Load(path)
	if err != nil || reload.Get().Microphone.ASP != config.ASPBuiltin {
		t.Fatal("saved on click")
	}
	f.transition(config.ASPExternal, config.ASPExternal, "000: External ASP processing OK", true)
	m.syncASP()
	assertASPFeature(t, m, "External", "External", "000: External ASP processing OK", config.ASPExternal)
	reload, err = config.Load(path)
	if err != nil || reload.Get().Microphone.ASP != config.ASPExternal {
		t.Fatal("confirmed ASP not saved/reloaded")
	}
	f.transition(config.ASPExternal, config.ASPBuiltin, "000: Fallback: external ASP exited", true)
	m.syncASP()
	assertASPFeature(t, m, "External", "Built-in", "000: Fallback: external ASP exited", config.ASPExternal)
	m.asp.OnCommand("Built-in")
	assertASPFeature(t, m, "Built-in", "Built-in", "000: Built-in processing OK", config.ASPBuiltin)
	reload, err = config.Load(path)
	if err != nil || reload.Get().Microphone.ASP != config.ASPBuiltin {
		t.Fatal("explicit Built-in not saved")
	}
}

func TestASPFeatureFirstFailVersusSavedBootFail(t *testing.T) {
	for _, saved := range []config.ASPMode{config.ASPBuiltin, config.ASPExternal} {
		m, f, _ := aspFeature(t, true, saved)
		m.restoreASP(config.ASPExternal)
		selected := saved
		f.transition(selected, config.ASPBuiltin, "000: Fallback: external ASP binary not found", false)
		m.syncASP()
		assertASPFeature(t, m, selected.Label(), "Built-in", "000: Fallback: external ASP binary not found", saved)
	}
}

func TestASPFeatureUnavailableOptionsHonestRestore(t *testing.T) {
	m, f, _ := aspFeature(t, false, config.ASPBuiltin)
	if !slices.Equal(m.asp.Options, []string{"Built-in"}) {
		t.Fatal("offered unavailable External")
	}
	m.asp.OnCommand("External")
	assertASPFeature(t, m, "Built-in", "Built-in", "000: Built-in processing OK", config.ASPBuiltin)
	if err := config.Set().Microphone().ASP(config.ASPExternal); err != nil {
		t.Fatal(err)
	}
	m.restoreASP(config.ASPExternal)
	f.transition(config.ASPExternal, config.ASPBuiltin, "000: Fallback: external ASP binary not found", false)
	m.syncASP()
	assertASPFeature(t, m, "External", "Built-in", "000: Fallback: external ASP binary not found", config.ASPExternal)
	if !slices.Equal(m.asp.Options, []string{"Built-in", "External"}) {
		t.Fatal("cannot display saved unavailable choice")
	}
	m.asp.OnCommand("Built-in")
	if !slices.Equal(m.asp.Options, []string{"Built-in"}) {
		t.Fatal("unavailable External still offered after explicit Built-in")
	}
}

func TestASPFeatureCoalescedPromotionFailureStillSaves(t *testing.T) {
	m, f, _ := aspFeature(t, true, config.ASPBuiltin)
	m.asp.OnCommand("External")
	f.transition(config.ASPExternal, config.ASPExternal, "000: External ASP processing OK", true)
	f.transition(config.ASPExternal, config.ASPBuiltin, "000: Fallback: external ASP exited", true)
	m.syncASP()
	assertASPFeature(t, m, "External", "Built-in", "000: Fallback: external ASP exited", config.ASPExternal)
}

func TestASPFeatureLatestCommandDoesNotSaveStaleSuccess(t *testing.T) {
	m, f, _ := aspFeature(t, true, config.ASPBuiltin)
	m.asp.OnCommand("External")
	f.transition(config.ASPExternal, config.ASPExternal, "000: External ASP processing OK", true)
	// The queued notification survives, but the callback reads current state after the newer command.
	m.asp.OnCommand("Built-in")
	m.syncASP()
	assertASPFeature(t, m, "Built-in", "Built-in", "000: Built-in processing OK", config.ASPBuiltin)
}

func TestASPFeaturePushWithoutPolling(t *testing.T) {
	m, f, _ := aspFeature(t, true, config.ASPBuiltin)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- m.Run(ctx) }()
	m.asp.OnCommand("External")
	f.transition(config.ASPExternal, config.ASPExternal, "000: External ASP processing OK", true)
	until := time.Now().Add(time.Second)
	for m.activeASP.Get() != "External" {
		if time.Now().After(until) {
			t.Fatal("diagnostics waited for poll")
		}
		time.Sleep(time.Millisecond)
	}
	cancel()
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	assertASPFeature(t, m, "External", "External", "000: External ASP processing OK", config.ASPExternal)
}

func TestASPFeatureEntityInventory(t *testing.T) {
	m := build()
	expected := map[string]string{"microphone_asp": "ASP frontend", "active_asp": "Active ASP", "asp_debug": "ASP debug"}
	seen := map[string]bool{}
	for _, e := range m.Entities() {
		name, ok := expected[e.Object()]
		if !ok {
			continue
		}
		if seen[e.Object()] {
			t.Fatal("duplicate ASP entity")
		}
		seen[e.Object()] = true
		switch x := e.(type) {
		case *esphome.Select:
			if x.Name != name || x.Category != esphome.CategoryConfig || x.DeviceID != component.DeviceMicrophone || x.OnCommand == nil {
				t.Fatal("wrong ASP selector metadata")
			}
		case *esphome.TextSensor:
			if x.Name != name || x.Category != esphome.CategoryDiagnostic || x.DeviceID != component.DeviceMicrophone {
				t.Fatal("wrong ASP diagnostic metadata")
			}
		default:
			t.Fatal("ASP diagnostic has writable entity type")
		}
	}
	if len(seen) != 3 {
		t.Fatal("ASP inventory incomplete", seen)
	}
}

func TestASPFeatureRealSourceDefersUntilCapture(t *testing.T) {
	m, _, _ := aspFeature(t, false, config.ASPExternal)
	source := mic.New()
	m.bindASP(source, false)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- m.Run(ctx) }()
	m.restoreASP(config.ASPExternal)
	// The host must not attempt a factory launch before ALSA capture is open.
	assertASPFeature(t, m, "External", "Built-in", "000: External ASP starting", config.ASPExternal)
	select {
	case <-source.ASPChanges():
		// The initial requested-state notification is allowed.
	default:
	}
	time.Sleep(30 * time.Millisecond)
	assertASPFeature(t, m, "External", "Built-in", "000: External ASP starting", config.ASPExternal)
	m.asp.OnCommand("Built-in")
	assertASPFeature(t, m, "Built-in", "Built-in", "000: Built-in processing OK", config.ASPBuiltin)
	cancel()
	<-done
}
