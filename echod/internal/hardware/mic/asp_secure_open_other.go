//go:build !linux

package mic

import "errors"

// The pinned openat resolver is Linux-only. Never silently fall back to a
// pathname check or regular exec on unsupported systems.
type secureASPPinned struct{}

func openSecureExternalASP() (*secureASPPinned, error) {
	return nil, errors.New("secure ASP: pinned open requires Linux")
}

func (*secureASPPinned) Close() error { return nil }

func pinnedASPIdentity(*secureASPPinned) (aspPinnedIdentity, error) {
	return aspPinnedIdentity{}, errors.New("secure ASP: pinned identity requires Linux")
}
