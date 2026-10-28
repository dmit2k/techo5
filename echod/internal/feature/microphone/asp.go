package microphone

import (
	"context"
	"log/slog"

	esphome "github.com/ygelfand/go-esphome-device"

	"github.com/HuskerMinion/techo5/echod/internal/component"
	"github.com/HuskerMinion/techo5/echod/internal/config"
	"github.com/HuskerMinion/techo5/echod/internal/hardware/mic"
)

// Only the existing Source control surface is needed here; tests substitute deterministic state.
type aspSource interface {
	SetASP(config.ASPMode) config.ASPMode
	ASPStatus() mic.ASPStatus
	ASPChanges() <-chan struct{}
}

func (m *Microphone) bindASP(source aspSource, have bool) {
	m.aspSource, m.aspHave = source, have
	m.asp = &esphome.Select{Base: esphome.Base{
		ObjectID: "microphone_asp", Name: "ASP frontend", Icon: "mdi:microphone-settings",
		Category: esphome.CategoryConfig, DeviceID: component.DeviceMicrophone,
	}}
	m.activeASP = &esphome.TextSensor{Base: esphome.Base{
		ObjectID: "active_asp", Name: "Active ASP", Icon: "mdi:microphone",
		Category: esphome.CategoryDiagnostic, DeviceID: component.DeviceMicrophone,
	}}
	m.aspDebug = &esphome.TextSensor{Base: esphome.Base{
		ObjectID: "asp_debug", Name: "ASP debug", Icon: "mdi:information-outline",
		Category: esphome.CategoryDiagnostic, DeviceID: component.DeviceMicrophone,
	}}
	m.asp.OnCommand = m.setASP
	m.syncASP()
}

// ASP first activation is asynchronous: component.Bind's unconditional immediate Save cannot
// be used. This is the only custom binding; persistence still uses the ordinary writer/store.
func (m *Microphone) setASP(label string) {
	m.aspMu.Lock()
	defer m.aspMu.Unlock()
	want, ok := config.ByLabel(config.ASPModes(), label)
	if !ok {
		return
	}
	if want == config.ASPExternal && !m.aspHave && m.aspSource.ASPStatus().Selected != config.ASPExternal {
		return
	}
	m.asp.Set(want.Label())
	m.aspSource.SetASP(want)
	if want == config.ASPBuiltin {
		if err := config.Set().Microphone().ASP(want); err != nil {
			slog.Error("saving ASP selection failed", "err", err)
		}
	}
	m.publishASPLocked(m.aspSource.ASPStatus())
}

func (m *Microphone) restoreASP(saved config.ASPMode) {
	m.aspMu.Lock()
	defer m.aspMu.Unlock()
	component.Restore(m.asp, saved, m.aspSource.SetASP)
	m.publishASPLocked(m.aspSource.ASPStatus())
}

func (m *Microphone) syncASP() {
	m.aspMu.Lock()
	defer m.aspMu.Unlock()
	m.publishASPLocked(m.aspSource.ASPStatus())
}

func (m *Microphone) publishASPLocked(st mic.ASPStatus) {
	// HA commands and this save share aspMu. An older completion cannot save after a newer
	// Built-in command, including when Store.Update waits on its own lock or disk.
	if st.Confirmed && st.Selected == config.ASPExternal && config.Get().Microphone.ASP != config.ASPExternal {
		if err := config.Set().Microphone().ASP(config.ASPExternal); err != nil {
			slog.Error("saving confirmed ASP selection failed", "err", err)
		}
	}
	modes := []config.ASPMode{config.ASPBuiltin}
	if m.aspHave || st.Selected == config.ASPExternal {
		modes = config.ASPModes()
	}
	m.asp.SetOptions(config.Labels(modes))
	if m.asp.Get() != st.Selected.Label() {
		m.asp.Set(st.Selected.Label())
	}
	if m.activeASP.Get() != st.Active.Label() {
		m.activeASP.Set(st.Active.Label())
	}
	if m.aspDebug.Get() != st.Debug {
		m.aspDebug.Set(st.Debug)
	}
}

// Run adds transition delivery to the existing microphone component, not a diagnostics poller.
// Initial values are already present at build/Restore, before this loop or HA subscription.
func (m *Microphone) Run(ctx context.Context) error {
	changes := m.aspSource.ASPChanges()
	m.syncASP()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-changes:
			m.syncASP()
		}
	}
}
