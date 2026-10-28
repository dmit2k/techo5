//go:build dot || spot || panel

package mic

import "github.com/HuskerMinion/techo5/echod/internal/asp"

// No factory four-MIC capture source is qualified on these build targets.
// Do not read IDME or advertise factory calibration.
func micFactoryGainForSession(config asp.Config, platform *asp.PlatformInfo) *asp.MicFactoryGain {
	return nil
}
