//go:build linux

package mic

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"
	"time"
)

// prepareSecureASPPinned only prepares a command: it never starts or reaps it.
// The caller exclusively owns p until Start returns, closes it on EVERY branch,
// and calls Wait exactly once after successful Start. No pathname re-resolution.
func prepareSecureASPPinned(ctx context.Context, p *secureASPPinned, procFD string, args []string) (*exec.Cmd, error) {
	if p == nil {
		return nil, errors.New("secure ASP exec: nil pinned descriptors")
	}
	if ctx == nil {
		return nil, errors.New("secure ASP exec: nil context")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	// A closed ExtraFiles entry alone is insufficient: exec can succeed with a
	// closed asset FD. Reject BOTH invalid descriptors before attempting Start.
	for i, f := range []*os.File{p.executable, p.externalRoot} {
		if f == nil {
			return nil, errors.New("secure ASP exec: nil descriptor")
		}
		var st syscall.Stat_t
		if err := syscall.Fstat(int(f.Fd()), &st); err != nil {
			return nil, fmt.Errorf("secure ASP exec: descriptor %d: %w", 3+i, err)
		}
		want := uint32(syscall.S_IFREG)
		if i == 1 {
			want = syscall.S_IFDIR
		}
		if st.Mode&syscall.S_IFMT != want {
			return nil, fmt.Errorf("secure ASP exec: wrong descriptor %d type", 3+i)
		}
	}
	// Only admit procfd failure injection, never a fallback pathname.
	if procFD != "/proc/self/fd/3" && procFD != "/proc/self/fd/-1" {
		return nil, errors.New("secure ASP exec: unsupported procfd")
	}
	cmd := exec.CommandContext(ctx, procFD, args...)
	cmd.ExtraFiles = []*os.File{p.executable, p.externalRoot}
	// No inherited loader overrides (LD_PRELOAD, LD_LIBRARY_PATH, etc.).
	// PT_INTERP and system shared libraries remain kernel/loader dependencies.
	cmd.Env = []string{"LANG=C", "LC_ALL=C"}
	cmd.WaitDelay = time.Second
	return cmd, nil
}

// startSecureASPPinned is the standalone probe entry point. The supervised host
// uses preparation directly, attaching its pipes before its sole Start/Wait.
// Both entry points consume p, including on errors and cancellation.
func startSecureASPPinned(ctx context.Context, p *secureASPPinned, args []string, stdout, stderr io.Writer) (*exec.Cmd, error) {
	return startSecureASPPinnedAtProcFD(ctx, p, "/proc/self/fd/3", args, stdout, stderr)
}

func startSecureASPPinnedAtProcFD(ctx context.Context, p *secureASPPinned, procFD string, args []string, stdout, stderr io.Writer) (*exec.Cmd, error) {
	defer p.Close()
	cmd, err := prepareSecureASPPinned(ctx, p, procFD, args)
	if err != nil {
		return nil, err
	}
	cmd.Stdout, cmd.Stderr = stdout, stderr
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("secure ASP exec: Start: %w", err)
	}
	cmd.ExtraFiles = nil
	return cmd, nil
}

// startExternalASPPinned consumes already-verified descriptors. Host startup
// owns the only Start and reaper; release runs immediately after Start, including
// failure, and also on earlier pipe/context errors. FD3/FD4 survive in the child.
func startExternalASPPinned(ctx context.Context, p *secureASPPinned, f *externalASP) (*externalASP, error) {
	return startExternalASPPinnedAtProcFD(ctx, p, f, "/proc/self/fd/3", nil)
}

// procFD and args are private offline fault/probe seams, never user config/ENV.
func startExternalASPPinnedAtProcFD(ctx context.Context, p *secureASPPinned, f *externalASP, procFD string, args []string) (*externalASP, error) {
	defer p.Close()
	if err := validateASPSession(f); err != nil {
		return nil, err
	}
	cmd, err := prepareSecureASPPinned(ctx, p, procFD, args)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errASPStart, err)
	}
	return startExternalASPProcessWithRelease(ctx, cmd, f, func() {
		_ = p.Close()
		cmd.ExtraFiles = nil
	})
}

func pinnedASPIdentity(p *secureASPPinned) (aspPinnedIdentity, error) {
	var id aspPinnedIdentity
	if p == nil {
		return id, errors.New("secure ASP: missing pin")
	}
	for i, f := range []*os.File{p.executable, p.externalRoot} {
		if f == nil {
			return id, errors.New("secure ASP: missing descriptor")
		}
		var st syscall.Stat_t
		if err := syscall.Fstat(int(f.Fd()), &st); err != nil {
			return id, err
		}
		id[i] = aspInode{dev: uint64(st.Dev), ino: uint64(st.Ino)}
	}
	return id, nil
}
