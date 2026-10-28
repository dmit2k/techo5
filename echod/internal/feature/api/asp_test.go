package api

import (
	"github.com/HuskerMinion/techo5/echod/internal/feature/microphone"
	"testing"
)

func TestASPEntitiesHiddenOnNoMicBoard(t *testing.T) {
	ids := map[string]bool{"microphone_asp": true, "active_asp": true, "asp_debug": true}
	entities := microphone.Get().Entities()
	found := 0
	for _, e := range entities {
		if ids[e.Object()] {
			found++
		}
	}
	if found != 3 {
		t.Fatal("ASP entity inventory incomplete")
	}
	for _, e := range dropListening(entities) {
		if ids[e.Object()] {
			t.Fatal("ASP entity offered without microphones")
		}
	}
}
