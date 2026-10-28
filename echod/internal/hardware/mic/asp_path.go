package mic

import (
	"context"
	"fmt"
	"sync"

	"github.com/HuskerMinion/techo5/echod/internal/asp"
	"github.com/HuskerMinion/techo5/echod/internal/layout"
)

const ASPHelper = "external-asp"
const aspSecurePreferredPath = "/data/techo5-linux/external-asp/bin/external-asp"

type aspInode struct{ dev, ino uint64 }
type aspPinnedIdentity [2]aspInode // executable and external asset root

// Cached discovery retains identity, never descriptors. Replacement requires
// daemon restart. No filename, ENV, HELLO or board name grants qualification.
type aspCandidate struct {
	identity aspPinnedIdentity
}

func (c aspCandidate) operational() bool {
	return c.identity != (aspPinnedIdentity{})
}

// Availability depends only on secure discovery of the helper and its asset root.
var discoveredASP = sync.OnceValue(func() aspCandidate {
	return discoverASPCandidate(openSecureExternalASP)
})

func discoverASPCandidate(opener func() (*secureASPPinned, error)) aspCandidate {
	p, err := opener()
	if p != nil {
		defer p.Close()
	}
	if err != nil {
		return aspCandidate{}
	}
	id, err := pinnedASPIdentity(p)
	if err != nil {
		return aspCandidate{}
	}
	return aspCandidate{identity: id}
}
func haveExternalASPCandidate(board string, c aspCandidate) bool {
	return board == "athena" && c.operational()
}
func HaveExternalASP() bool {
	return layout.Board == "athena" && haveExternalASPCandidate(layout.Board, discoveredASP())
}

// The startup worker snapshots the held capture's shape and metadata. Missing
// native hardware REF or an unavailable snapshot fails closed before launch.
func (s *Source) openExternalASP(ctx context.Context) (*externalASP, error) {
	if ctx == nil {
		return nil, errASPStart
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if layout.Board != "athena" {
		return nil, errASPReference
	}
	c := discoveredASP()
	if !haveExternalASPCandidate(layout.Board, c) {
		return nil, errASPNotFound
	}
	capture, ok := s.device().(interface{ aspConfig() asp.Config })
	if !ok {
		return nil, errASPReference
	}
	config := capture.aspConfig()
	session, err := prepareASP2SessionCFG2(config)
	if err != nil {
		return nil, errASPReference
	}
	return startASPCandidate(ctx, c, &externalASP{config: config, session: session}, openSecureExternalASP)
}

// Reopen through the sole secure authority and compare both cached inodes.
// There is one FD-pinned launch and no pathname retry.
func startASPCandidate(ctx context.Context, c aspCandidate, f *externalASP, opener func() (*secureASPPinned, error)) (*externalASP, error) {
	if !c.operational() {
		return nil, errASPNotFound
	}
	if ctx == nil {
		return nil, errASPStart
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := validateASPSession(f); err != nil {
		return nil, err
	}
	p, err := opener()
	if p != nil {
		defer p.Close()
	}
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errASPStart, err)
	}
	id, err := pinnedASPIdentity(p)
	if err != nil || id != c.identity {
		return nil, fmt.Errorf("%w: secure candidate changed or unavailable", errASPStart)
	}
	return startExternalASPPinned(ctx, p, f)
}
