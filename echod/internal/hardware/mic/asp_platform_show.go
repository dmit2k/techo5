//go:build !dot && !spot && !panel

package mic

import (
	"os"
	"strings"

	"github.com/HuskerMinion/techo5/echod/internal/asp"
	"github.com/HuskerMinion/techo5/echod/internal/layout"
)

// platformInfoForSession is an opt-in, one-shot adapter for a future CFG2
// session builder. It is deliberately NOT invoked by the current ASP2 host
// or from a 20-ms capture loop. layout.Board remains the chosen identity.
// These are the same read-only kernel sources used by layout.showBoard.
// Reading them anew allows the separate confidence assessment to detect
// an ambiguous or changed identity without mutating the layout selection.
func platformInfoForSession() *asp.PlatformInfo {
	board, doubt := layout.Board, layout.BoardDoubt
	cmdline, _ := os.ReadFile("/proc/cmdline")
	rawCompatible, _ := os.ReadFile("/proc/device-tree/compatible")
	var compatible []string
	for _, c := range strings.Split(string(rawCompatible), "\x00") {
		if c = strings.TrimSpace(c); c != "" {
			compatible = append(compatible, c)
		}
	}
	return platformInfoFromEvidence(board, string(cmdline), compatible, doubt)
}
