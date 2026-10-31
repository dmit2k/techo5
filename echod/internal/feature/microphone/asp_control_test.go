package microphone

import (
	"path/filepath"
	"slices"
	"testing"

	"github.com/HuskerMinion/techo5/echod/internal/config"
	"github.com/HuskerMinion/techo5/echod/internal/hardware/mic"
)

func TestL09SelectedPersistenceAndAlwaysAvailableHA(t *testing.T) {
	config.Use(filepath.Join(t.TempDir(), "state.json"))
	m := build()
	if !slices.Contains(m.aspSelected.Options, "External") || !slices.Contains(m.aspSelected.Options, "Built-in") {
		t.Fatalf("option unavailable without helper: %v", m.aspSelected.Options)
	}
	m.aspSelected.OnCommand("External")
	if saved := config.Get().Microphone.ASP; saved != config.ASPExternal {
		t.Fatalf("selection not persisted: %q", saved)
	}
	if got := mic.Get().SelectedASP(); got != config.ASPExternal {
		t.Fatalf("selected %q", got)
	}
	if got := mic.Get().ActiveASP(); got != config.ASPBuiltin {
		t.Fatalf("active before first PCM %q", got)
	}
	st, err := config.Load(filepath.Join(t.TempDir(), "missing.json"))
	if err != nil || st.Get().Microphone.ASP != config.ASPBuiltin {
		t.Fatal("default mode is not Built-in", err)
	}
	m.aspSelected.OnCommand("Built-in")
	if got := config.Get().Microphone.ASP; got != config.ASPBuiltin {
		t.Fatalf("built-in not persisted: %q", got)
	}
	c := config.Get()
	c.Microphone.ASP = config.ASPExternal
	m.Restore(c)
	if got := mic.Get().SelectedASP(); got != config.ASPExternal {
		t.Fatal("restore did not select External")
	}
	mic.Get().SetASP(config.ASPBuiltin)
}
