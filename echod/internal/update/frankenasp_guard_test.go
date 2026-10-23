package update

import (
	"errors"
	"os"
	"strings"
	"testing"
)

// Test the guard against an isolated slot flag and inert stat responses.
// No actual device directories or system updates are involved.
func TestFrankenPersistenceGuardsUpdate(t *testing.T) {
	oldStat := statFrankenMode
	defer func() { statFrankenMode = oldStat }()
	if !strings.Contains(persistentFrankenFlag, "/data/") ||
		!strings.Contains(persistentFrankenActive, "/run/") {
		t.Fatal("persistence guards must use expected owner-data and boot-scoped markers")
	}
	statFrankenMode = func(name string) (os.FileInfo, error) {
		if name == persistentFrankenFlag {
			return nil, nil
		}
		return nil, os.ErrNotExist
	}
	// Pure checker when hardware is unavailable on developer workstation:
	// no test should trigger Install() or touch real slot partitions.
	if !guardFrankenFlagsPresent() {
		t.Fatal("persistent mode must not be treated as absent")
	}
	if err := guardFrankenUpdateForSlot(true, true); err == nil ||
		!strings.Contains(err.Error(), "disable") {
		t.Fatal("Athena slot update must be blocked with an actionable message")
	}
	if err := guardFrankenUpdateForSlot(false, true); err != nil {
		t.Fatalf("non-slot update must be unchanged: %v", err)
	}
	if err := guardFrankenUpdateForSlot(true, false); err != nil {
		t.Fatalf("other board update must be unchanged: %v", err)
	}
	statFrankenMode = func(name string) (os.FileInfo, error) { return nil, os.ErrNotExist }
	if guardFrankenFlagsPresent() {
		t.Fatal("ordinary stock mode should not be guarded")
	}
	statFrankenMode = func(name string) (os.FileInfo, error) { return nil, errors.New("EIO") }
	if !guardFrankenFlagsPresent() {
		t.Fatal("I/O error must fail closed")
	}
}

// A custom suffix remains HA-rankable and cannot hide the next upstream
// release. It also must not invent an unattainable future TECHO5 version.
func TestFrankenBuildStampVersionOrdering(t *testing.T) {
	stamped := "v1.6.0_frankenasp.abcdef1234"
	if err := ValidVersion(stamped); err != nil {
		t.Fatalf("stamped version rejected: %v", err)
	}
	if !Newer("v1.6.1", stamped) {
		t.Fatal("a newer official release must remain visible")
	}
	if Newer("v1.6.0", stamped) {
		t.Fatal("an identical official base version is not newer")
	}
}
