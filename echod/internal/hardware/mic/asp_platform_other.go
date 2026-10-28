//go:build dot || spot || panel

package mic

import (
	"github.com/HuskerMinion/techo5/echod/internal/asp"
	"github.com/HuskerMinion/techo5/echod/internal/layout"
)

// No non-IDME hardware evidence is qualified for these device families
// in H1-A1. An absent BoardDoubt (including a fixed build-time board)
// is insufficient to claim CONFIRMED. Do not inspect IDME here.
func platformInfoForSession() *asp.PlatformInfo {
	return platformInfoFromEvidence(layout.Board, "", nil, layout.BoardDoubt)
}
