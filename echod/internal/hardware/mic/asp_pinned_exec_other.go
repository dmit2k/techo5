//go:build !linux

package mic

import "context"

func startExternalASPPinned(context.Context, *secureASPPinned, *externalASP) (*externalASP, error) {
	return nil, errASPStart
}
