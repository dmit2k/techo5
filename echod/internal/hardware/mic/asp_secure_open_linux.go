//go:build linux

package mic

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"syscall"
)

var aspSecureComponents = [...]string{"data", "techo5-linux", "external-asp", "bin", "external-asp"}

// secureASPPinned owns the exact verified launcher and external-asp directory
// inodes. Neither file is opened again by pathname. Close is safe to call more
// than once, including concurrently, and closes both even if one close fails.
type secureASPPinned struct {
	executable   *os.File
	externalRoot *os.File
	closeOnce    sync.Once
	closeErr     error
}

func (p *secureASPPinned) Close() error {
	if p == nil {
		return nil
	}
	p.closeOnce.Do(func() {
		p.closeErr = errors.Join(p.executable.Close(), p.externalRoot.Close())
	})
	return p.closeErr
}

// openSecureExternalASP is the production entry point. The authority root and
// required UID cannot be supplied by configuration or environment variables.
// Discovery and every secure activation use this same inode authority.
func openSecureExternalASP() (*secureASPPinned, error) {
	root, err := os.Open("/")
	if err != nil {
		return nil, fmt.Errorf("secure ASP root: %w", err)
	}
	defer root.Close()
	return secureOpenASPAtRoot(root, aspSecurePreferredPath, 0, nil)
}

// secureOpenASPAtRoot only admits the literal production path. An injected
// trusted root FD and owner UID allow unprivileged tests to reproduce the
// device's ownership policy without chown or touching the real /data.
// afterDir is a private *test seam* for deterministic rename/symlink races;
// production passes nil. It cannot skip or override inode verification.
func secureOpenASPAtRoot(root *os.File, absolutePath string, ownerUID uint32, afterDir func(string) error) (_ *secureASPPinned, retErr error) {
	if absolutePath != aspSecurePreferredPath {
		return nil, fmt.Errorf("secure ASP: unsupported absolute path %q", absolutePath)
	}
	if root == nil {
		return nil, errors.New("secure ASP: nil authority root")
	}
	if err := secureASPStat(root, ownerUID, -1); err != nil {
		return nil, fmt.Errorf("secure ASP: untrusted authority root: %w", err)
	}

	current := root // caller owns root, even on failure
	var openedDir, externalRoot *os.File
	defer func() {
		if openedDir != nil && openedDir != externalRoot {
			_ = openedDir.Close()
		}
		if retErr != nil && externalRoot != nil {
			_ = externalRoot.Close()
		}
	}()

	for i, component := range aspSecureComponents[:len(aspSecureComponents)-1] {
		fd, err := syscall.Openat(int(current.Fd()), component,
			syscall.O_RDONLY|syscall.O_DIRECTORY|syscall.O_NOFOLLOW|syscall.O_CLOEXEC, 0)
		if err != nil {
			return nil, fmt.Errorf("secure ASP: open directory %s: %w", component, err)
		}
		dir := os.NewFile(uintptr(fd), component)
		if err := secureASPStat(dir, ownerUID, i); err != nil {
			_ = dir.Close()
			return nil, fmt.Errorf("secure ASP: directory %s: %w", component, err)
		}
		if openedDir != nil && openedDir != externalRoot {
			_ = openedDir.Close()
		}
		openedDir, current = dir, dir
		if i == 2 {
			externalRoot = dir
		}
		if afterDir != nil {
			// Only used by offline tests; the validated current FD remains
			// authoritative even if its name is renamed during this hook.
			if err := afterDir(component); err != nil {
				return nil, fmt.Errorf("secure ASP: test step %s: %w", component, err)
			}
		}
	}

	fd, err := syscall.Openat(int(current.Fd()), aspSecureComponents[len(aspSecureComponents)-1],
		syscall.O_RDONLY|syscall.O_NOFOLLOW|syscall.O_CLOEXEC|syscall.O_NONBLOCK, 0)
	if err != nil {
		return nil, fmt.Errorf("secure ASP: open executable: %w", err)
	}
	executable := os.NewFile(uintptr(fd), "external-asp")
	if err := secureASPStat(executable, ownerUID, len(aspSecureComponents)-1); err != nil {
		_ = executable.Close()
		return nil, fmt.Errorf("secure ASP: executable: %w", err)
	}
	// A cheap ELF-candidate check, not a complete ELF/ABI validation. S1-A2
	// must still prove that this pinned descriptor can actually be executed.
	var magic [4]byte
	if _, err := executable.ReadAt(magic[:], 0); err != nil {
		_ = executable.Close()
		return nil, fmt.Errorf("secure ASP: read ELF header: %w", err)
	}
	if magic != [4]byte{0x7f, 'E', 'L', 'F'} {
		_ = executable.Close()
		return nil, errors.New("secure ASP: not an ELF executable")
	}
	// The bin FD is closed by the deferred cleanup; externalRoot survives.
	return &secureASPPinned{executable: executable, externalRoot: externalRoot}, nil
}

// position = -1: pinned authority /; 0/1: ONLY the two factory 0777
// ancestors; 2/3: our private 0700 tree; 4: regular ELF launcher.
// fstat always operates on the already-opened inode, not a pathname.
func secureASPStat(file *os.File, ownerUID uint32, position int) error {
	var st syscall.Stat_t
	if err := syscall.Fstat(int(file.Fd()), &st); err != nil {
		return err
	}
	if st.Uid != ownerUID {
		return fmt.Errorf("unexpected owner UID %d (want %d)", st.Uid, ownerUID)
	}
	mode := st.Mode
	perm := mode & 07777
	if position < 4 {
		if mode&syscall.S_IFMT != syscall.S_IFDIR {
			return fmt.Errorf("not a directory (mode %#o)", mode)
		}
		if position == -1 {
			if perm&0022 != 0 || perm&07000 != 0 {
				return fmt.Errorf("unsafe authority root mode %#o", perm)
			}
		} else if position >= 2 {
			if perm != 0700 {
				return fmt.Errorf("private directory mode %#o (want 0700)", perm)
			}
		} else if perm&0022 != 0 && perm != 0777 {
			// Writable ancestors are trusted ONLY at exact /data and
			// /data/techo5-linux positions and ONLY in known 0777 form.
			return fmt.Errorf("unsafe ancestor directory mode %#o", perm)
		} else if perm&07000 != 0 {
			return fmt.Errorf("unsafe ancestor special bits %#o", perm)
		}
		return nil
	}
	if mode&syscall.S_IFMT != syscall.S_IFREG {
		return fmt.Errorf("not a regular file (mode %#o)", mode)
	}
	if perm&07000 != 0 || perm&0022 != 0 || perm&0500 != 0500 {
		return fmt.Errorf("unsafe executable mode %#o", perm)
	}
	if st.Size < 4 {
		return fmt.Errorf("short executable (%d bytes)", st.Size)
	}
	return nil
}
