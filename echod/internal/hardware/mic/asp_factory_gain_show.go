//go:build !dot && !spot && !panel

package mic

import (
	"github.com/HuskerMinion/techo5/echod/internal/asp"
	"github.com/HuskerMinion/techo5/echod/internal/layout"
)

// micFactoryGainForSession is opt-in and currently has NO production caller.
// Call once before starting an ASP2 session, only after obtaining an
// independently verified platform descriptor and native capture config.
// Never call from the 20-ms capture path or an active ASP2 exchange.
func micFactoryGainForSession(config asp.Config, platform *asp.PlatformInfo) *asp.MicFactoryGain {
	return micFactoryGainFromReader(config, platform, layout.Board, layout.Idme)
}
