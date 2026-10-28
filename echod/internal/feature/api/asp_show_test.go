//go:build !dot && !spot && !panel

package api

import (
	"github.com/HuskerMinion/techo5/echod/internal/feature/microphone"
	"github.com/HuskerMinion/techo5/echod/internal/layout"
	"testing"
)

func TestASPEntitiesHiddenOnActualSkylightLayout(t *testing.T) {
	old := layout.Board
	layout.Board = "skylight"
	defer func() { layout.Board = old }()
	if layout.HasMicrophone() {
		t.Fatal("Skylight fixture still has microphone")
	}
	for _, e := range heard(microphone.Get().Entities()) {
		if e.Object() == "microphone_asp" || e.Object() == "active_asp" || e.Object() == "asp_debug" {
			t.Fatal("ASP offered on no-mic board")
		}
	}
}
