package update

import (
	"errors"
	"os"

	"github.com/HuskerMinion/techo5/echod/internal/layout"
)

const (
	persistentFrankenFlag   = "/data/techo5-linux/frankenasp-persistent.enabled"
	persistentFrankenActive = "/run/techo5/frankenasp-persistent-active"
)

// Overridable in tests. Not a device-identity proof: update.Install remains
// limited to Athena's rootfs slot install by slotSystem.
var statFrankenMode = os.Stat

var errFrankenBlocksUpdate = errors.New(
	"update: persistent FrankenASP active/requested; run " +
		"sh /data/techo5-linux/frankenasp-trial/persist.sh disable, " +
		"reboot into stock TECHO5, then install the official update; " +
		"the upstream rootfs does not contain FrankenASP")

func guardFrankenFlagsPresent() bool {
	for _, flag := range []string{persistentFrankenFlag, persistentFrankenActive} {
		// Any state except known absence is treated as active (fail closed).
		if _, err := statFrankenMode(flag); !os.IsNotExist(err) {
			return true
		}
	}
	return false
}

// FrankenASPPersistenceActive also lets the optional overnight auto-update
// loop skip the attempt entirely instead of raising a failure every night.
func FrankenASPPersistenceActive() bool {
	return slotSystem() && layout.Athena() && guardFrankenFlagsPresent()
}

func persistentFrankenUpdateGuard() error {
	if FrankenASPPersistenceActive() {
		return errFrankenBlocksUpdate
	}
	return nil
}

func guardFrankenUpdateForSlot(hasSlots, isAthena bool) error {
	if hasSlots && isAthena && guardFrankenFlagsPresent() {
		return errFrankenBlocksUpdate
	}
	return nil
}
