package remote_test

import (
	"github.com/HuskerMinion/techo5/echod/internal/feature/microphone"
	"github.com/HuskerMinion/techo5/echod/internal/feature/remote"
	esphome "github.com/ygelfand/go-esphome-device"
	"testing"
)

func TestASPDiagnosticsAreReadOnly(t *testing.T) {
	found := 0
	for _, e := range microphone.Get().Entities() {
		if e.Object() == "active_asp" || e.Object() == "asp_debug" {
			found++
			if _, ok := e.(*esphome.TextSensor); !ok || remote.Writable(e) {
				t.Fatal("ASP diagnostic writable")
			}
		}
	}
	if found != 2 {
		t.Fatal("not exactly two ASP diagnostics")
	}
}
