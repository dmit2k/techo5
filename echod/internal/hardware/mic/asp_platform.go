package mic

import (
	"strings"

	"github.com/HuskerMinion/techo5/echod/internal/asp"
)

// platformInfoFromEvidence constructs an owned, advisory session snapshot.
// board is the already-selected layout.Board, never a discovery override.
// Only mutually consistent, explicit hardware evidence may confirm that
// selection. BoardDoubt always downgrades it, even when evidence agrees.
//
// This pure function neither reads devices/IDME nor changes capture PCM.
func platformInfoFromEvidence(board, cmdline string, compatible []string, doubt string) *asp.PlatformInfo {
	info := &asp.PlatformInfo{Codename: board, Detection: asp.PlatformInferred}
	// The accepted CFG2 codec is the sole authority for codename validity.
	if info.Validate() != nil {
		return nil
	}

	var identified string
	conflict := false
	note := func(candidate string) {
		if identified != "" && identified != candidate {
			conflict = true
		} else {
			identified = candidate
		}
	}

	// The root DT identifies Athena (including board revisions) and Skylight.
	// Generic SoC compatibles, and near-misses like amzn,athenaeum, do not.
	for _, c := range compatible {
		switch {
		case athenaPlatformCompatible(c):
			note("athena")
		case c == "skylight,cal150":
			note("skylight")
		}
	}

	// Echo Show bootloaders identify checkers/crown/cronos in the panel
	// driver named by lcm=. Other cmdline parameters are NOT identities.
	// Require a whole underscore-delimited board token, not a substring.
	for _, token := range strings.Fields(cmdline) {
		panel, ok := strings.CutPrefix(token, "lcm=")
		if !ok || !strings.HasPrefix(panel, "1-") {
			continue
		}
		for _, field := range strings.Split(panel, "_") {
			switch field {
			case "checkers", "crown", "cronos":
				note(field)
			}
		}
	}

	if doubt == "" && !conflict && identified != "" && identified == board {
		info.Detection = asp.PlatformConfirmed
	}
	return info
}

// Match the root compatibles used by layout.showBoard without accepting
// empty/malformed pseudo-revisions as affirmative hardware evidence.
func athenaPlatformCompatible(c string) bool {
	if c == "amzn,athena" {
		return true
	}
	const prefix = "amzn,athena-"
	if !strings.HasPrefix(c, prefix) {
		return false
	}
	revision := strings.TrimPrefix(c, prefix)
	if len(revision) == 0 || revision[0] == '-' || revision[len(revision)-1] == '-' {
		return false
	}
	for i := 0; i < len(revision); i++ {
		ch := revision[i]
		if !((ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') || ch == '-') {
			return false
		}
	}
	return true
}
