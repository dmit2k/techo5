package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestASPDefaultsSaveReload(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	st, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if st.Get().Microphone.ASP != ASPBuiltin {
		t.Fatal("not Built-in by default")
	}
	if err := st.Set().Microphone().ASP(ASPExternal); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]map[string]json.RawMessage
	if err := json.Unmarshal(b, &raw); err != nil {
		t.Fatal(err)
	}
	var value string
	json.Unmarshal(raw["microphone"]["asp"], &value)
	if value != "external" {
		t.Fatal("wrong ordinary JSON setting")
	}
	reload, err := Load(path)
	if err != nil || reload.Get().Microphone.ASP != ASPExternal {
		t.Fatal("lost saved ASP", err)
	}
	if err := reload.Set().Microphone().ASP(ASPBuiltin); err != nil {
		t.Fatal(err)
	}
	again, err := Load(path)
	if err != nil || again.Get().Microphone.ASP != ASPBuiltin {
		t.Fatal("Builtin not saved")
	}
}

func TestASPExportRestoreMatchesCancelEngine(t *testing.T) {
	for _, external := range []bool{false, true} {
		c := Defaults()
		if external {
			c.Microphone.ASP = ASPExternal
			c.Microphone.CancelEngine = CancelWebRTC
		}
		b, err := Export(c, "model-a", "test")
		if err != nil {
			t.Fatal(err)
		}
		for _, model := range []string{"model-a", "model-b"} {
			cur := Defaults()
			cur.Microphone.ASP = ASPExternal
			cur.Microphone.CancelEngine = CancelWebRTC
			r, err := Restored(cur, b, model)
			if err != nil {
				t.Fatal(err)
			}
			if r.Config.Microphone.ASP != c.Microphone.ASP || r.Config.Microphone.CancelEngine != c.Microphone.CancelEngine {
				t.Fatal("ASP export/restore differs from WebRTC policy")
			}
		}
	}
}
