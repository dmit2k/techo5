//go:build linux

package mic

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"testing"
)

// A private unprivileged fixture is the sole test authority. Tests neither
// touch /data nor need chown/root, and only the literal production path is
// interpreted relative to this already-open test directory FD.
type aspSecureFixture struct {
	t    *testing.T
	base string
	uid  uint32
	root *os.File
}

func newASPSecureFixture(t *testing.T) *aspSecureFixture {
	t.Helper()
	base := t.TempDir()
	names := []string{"data", "techo5-linux", "external-asp", "bin"}
	at := base
	for i, name := range names {
		at = filepath.Join(at, name)
		if err := os.Mkdir(at, 0700); err != nil {
			t.Fatal(err)
		}
		mode := os.FileMode(0700)
		if i < 2 {
			mode = 0777
		}
		if err := os.Chmod(at, mode); err != nil {
			t.Fatal(err)
		}
	}
	binary := filepath.Join(at, "external-asp")
	if err := os.WriteFile(binary, aspSecureELF("ORIGINAL"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(binary, 0755); err != nil {
		t.Fatal(err)
	}
	root, err := os.Open(base)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = root.Close() })
	return &aspSecureFixture{t: t, base: base, uid: uint32(os.Geteuid()), root: root}
}

func aspSecureELF(payload string) []byte {
	return append([]byte{0x7f, 'E', 'L', 'F', 2, 1, 1, 0}, []byte(payload)...)
}

func (f *aspSecureFixture) part(i int) string {
	p := f.base
	for n := 0; n <= i; n++ {
		p = filepath.Join(p, aspSecureComponents[n])
	}
	return p
}

func (f *aspSecureFixture) open() (*secureASPPinned, error) {
	return secureOpenASPAtRoot(f.root, aspSecurePreferredPath, f.uid, nil)
}

func aspSecureInode(t *testing.T, file *os.File) uint64 {
	t.Helper()
	var st syscall.Stat_t
	if err := syscall.Fstat(int(file.Fd()), &st); err != nil {
		t.Fatal(err)
	}
	return st.Ino
}

func aspSecurePathInode(t *testing.T, path string) uint64 {
	t.Helper()
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	return aspSecureInode(t, file)
}

func aspSecureFDCount(t *testing.T) int {
	t.Helper()
	entries, err := os.ReadDir("/proc/self/fd")
	if err != nil {
		t.Skipf("Linux procfs required for FD leak check: %v", err)
	}
	return len(entries)
}

func aspSecureAssertCloexec(t *testing.T, file *os.File) {
	t.Helper()
	flag, _, e := syscall.Syscall(syscall.SYS_FCNTL, file.Fd(), uintptr(syscall.F_GETFD), 0)
	if e != 0 {
		t.Fatalf("F_GETFD: %v", e)
	}
	if flag&syscall.FD_CLOEXEC == 0 {
		t.Fatalf("descriptor %d lacks CLOEXEC", file.Fd())
	}
}

func TestASPSecureOpenGoodPinnedInodesAndClose(t *testing.T) {
	f := newASPSecureFixture(t)
	binIno := aspSecurePathInode(t, f.part(4))
	rootIno := aspSecurePathInode(t, f.part(2))
	p, err := f.open()
	if err != nil {
		t.Fatal(err)
	}
	if aspSecureInode(t, p.executable) != binIno || aspSecureInode(t, p.externalRoot) != rootIno {
		t.Fatal("returned FDs do not pin the tested inodes")
	}
	aspSecureAssertCloexec(t, p.executable)
	aspSecureAssertCloexec(t, p.externalRoot)
	var header [8]byte
	if _, err := p.executable.ReadAt(header[:], 0); err != nil || !bytes.Equal(header[:], aspSecureELF("")[:8]) {
		t.Fatalf("pinned ELF contents: %x, %v", header, err)
	}
	binFD, rootFD := int(p.executable.Fd()), int(p.externalRoot.Fd())
	var wg sync.WaitGroup
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := p.Close(); err != nil {
				t.Errorf("close: %v", err)
			}
		}()
	}
	wg.Wait()
	for _, fd := range []int{binFD, rootFD} {
		var st syscall.Stat_t
		if err := syscall.Fstat(fd, &st); !errors.Is(err, syscall.EBADF) {
			t.Fatalf("FD %d not closed: %v", fd, err)
		}
	}
	if err := p.Close(); err != nil {
		t.Fatal(err)
	}
	var empty *secureASPPinned
	if err := empty.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestASPSecureOpenExactPathOnly(t *testing.T) {
	f := newASPSecureFixture(t)
	paths := []string{
		"", "data/techo5-linux/external-asp/bin/external-asp",
		"/data/../data/techo5-linux/external-asp/bin/external-asp",
		"/data/techo5-linux/external-asp/bin/../bin/external-asp",
		"/data/techo5-linux/external-asp/bin/external-asp/",
		"/data/techo5-linux/external-asp/bin/external-asp/extra",
		"/data/techo5-linux/bin/external-asp",
		"/data/techo5-linux/external-asp2/bin/external-asp",
		"/data//techo5-linux/external-asp/bin/external-asp",
	}
	for _, path := range paths {
		t.Run(strings.ReplaceAll(path, "/", "_"), func(t *testing.T) {
			p, err := secureOpenASPAtRoot(f.root, path, f.uid, nil)
			if err == nil || p != nil {
				if p != nil {
					_ = p.Close()
				}
				t.Fatalf("accepted unauthorized path %q: %v", path, err)
			}
		})
	}
	if p, err := secureOpenASPAtRoot(nil, aspSecurePreferredPath, f.uid, nil); err == nil || p != nil {
		t.Fatal("accepted nil root authority")
	}
}

func TestASPSecureOpenModesAndOwnership(t *testing.T) {
	cases := []struct {
		name string
		pos  int
		mode os.FileMode
		ok   bool
	}{
		{"data_writable_0777", 0, 0777, true},
		{"parent_writable_0777", 1, 0777, true},
		{"data_nonwritable_0755", 0, 0755, true},
		{"parent_nonwritable_0700", 1, 0700, true},
		{"data_group_writable", 0, 0775, false},
		{"parent_group_writable", 1, 0770, false},
		{"parent_sticky_1777", 1, 0777 | os.ModeSticky, false},
		{"external_world_writable", 2, 0777, false},
		{"external_group_writable", 2, 0770, false},
		{"external_not_private_0755", 2, 0755, false},
		{"bin_world_writable", 3, 0777, false},
		{"bin_group_writable", 3, 0775, false},
		{"bin_readable_group", 3, 0750, false},
		{"launcher_world_writable", 4, 0777, false},
		{"launcher_group_writable", 4, 0775, false},
		{"launcher_no_exec", 4, 0644, false},
		{"launcher_nonexec_owner", 4, 0055, false},
		{"launcher_setuid", 4, 0755 | os.ModeSetuid, false},
		{"launcher_private_exec", 4, 0700, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := newASPSecureFixture(t)
			if err := os.Chmod(f.part(tc.pos), tc.mode); err != nil {
				t.Fatal(err)
			}
			p, err := f.open()
			if (err == nil) != tc.ok {
				if p != nil {
					_ = p.Close()
				}
				t.Fatalf("mode %#o accepted=%v expected=%v (%v)", tc.mode, err == nil, tc.ok, err)
			}
			if p != nil {
				if err := p.Close(); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
	t.Run("wrong_owner_authority", func(t *testing.T) {
		f := newASPSecureFixture(t)
		if p, err := secureOpenASPAtRoot(f.root, aspSecurePreferredPath, f.uid+1, nil); err == nil || p != nil {
			if p != nil {
				_ = p.Close()
			}
			t.Fatal("incorrect required UID accepted")
		}
	})
	t.Run("wrong_owner_opened_inode", func(t *testing.T) {
		f := newASPSecureFixture(t)
		// This is a REAL fstat of our inode, with a deliberately wrong
		// policy UID. No fake stat data or production UID bypass.
		fd, err := os.Open(f.part(2))
		if err != nil {
			t.Fatal(err)
		}
		defer fd.Close()
		if err := secureASPStat(fd, f.uid+1, 2); err == nil {
			t.Fatal("untrusted owner accepted")
		}
	})
	t.Run("root_writable_rejected", func(t *testing.T) {
		f := newASPSecureFixture(t)
		if err := os.Chmod(f.base, 0777); err != nil {
			t.Fatal(err)
		}
		p, err := f.open()
		if err == nil {
			_ = p.Close()
			t.Fatal("unsafe root FD accepted")
		}
	})
}

func TestASPSecureOpenSymlinkAtEveryLevel(t *testing.T) {
	for pos := 0; pos < len(aspSecureComponents); pos++ {
		t.Run(fmt.Sprint(pos), func(t *testing.T) {
			f := newASPSecureFixture(t)
			path := f.part(pos)
			old := path + ".real"
			if err := os.Rename(path, old); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(old, path); err != nil {
				t.Fatal(err)
			}
			p, err := f.open()
			if p != nil {
				_ = p.Close()
			}
			if err == nil {
				t.Fatalf("symlink at component %d accepted", pos)
			}
		})
	}
}

func TestASPSecureOpenMissingTypesAndFormat(t *testing.T) {
	for pos := 0; pos < len(aspSecureComponents); pos++ {
		t.Run(fmt.Sprintf("missing_%d", pos), func(t *testing.T) {
			f := newASPSecureFixture(t)
			if err := os.Rename(f.part(pos), f.part(pos)+".missing"); err != nil {
				t.Fatal(err)
			}
			p, err := f.open()
			if p != nil {
				_ = p.Close()
			}
			if err == nil {
				t.Fatal("missing path accepted")
			}
		})
		t.Run(fmt.Sprintf("wrong_type_%d", pos), func(t *testing.T) {
			f := newASPSecureFixture(t)
			path := f.part(pos)
			if err := os.Rename(path, path+".old"); err != nil {
				t.Fatal(err)
			}
			var err error
			if pos == 4 {
				err = os.Mkdir(path, 0700)
			} else {
				err = os.WriteFile(path, aspSecureELF("fake"), 0755)
			}
			if err != nil {
				t.Fatal(err)
			}
			p, err := f.open()
			if p != nil {
				_ = p.Close()
			}
			if err == nil {
				t.Fatal("incorrect inode type accepted")
			}
		})
	}
	for _, bytes := range [][]byte{nil, {0x7f, 'E', 'L'}, []byte("#!/bin/sh\nexit 0\n")} {
		t.Run(fmt.Sprintf("not_elf_%d", len(bytes)), func(t *testing.T) {
			f := newASPSecureFixture(t)
			if err := os.WriteFile(f.part(4), bytes, 0755); err != nil {
				t.Fatal(err)
			}
			p, err := f.open()
			if p != nil {
				_ = p.Close()
			}
			if err == nil {
				t.Fatal("invalid ELF candidate accepted")
			}
		})
	}
	t.Run("fifo_nonblocking", func(t *testing.T) {
		f := newASPSecureFixture(t)
		if err := os.Remove(f.part(4)); err != nil {
			t.Fatal(err)
		}
		if err := syscall.Mkfifo(f.part(4), 0755); err != nil {
			t.Fatal(err)
		}
		p, err := f.open() // must NOT wait for a FIFO writer
		if p != nil {
			_ = p.Close()
		}
		if err == nil {
			t.Fatal("FIFO accepted")
		}
	})
	t.Run("closed_authority", func(t *testing.T) {
		f := newASPSecureFixture(t)
		if err := f.root.Close(); err != nil {
			t.Fatal(err)
		}
		p, err := f.open()
		if p != nil {
			_ = p.Close()
		}
		if err == nil {
			t.Fatal("closed root FD accepted")
		}
	})
}

func TestASPSecureOpenSwapBeforeFinalOpenRejects(t *testing.T) {
	f := newASPSecureFixture(t)
	var swapped bool
	p, err := secureOpenASPAtRoot(f.root, aspSecurePreferredPath, f.uid, func(component string) error {
		if component != "bin" {
			return nil
		}
		swapped = true
		target := f.part(4)
		if err := os.Rename(target, target+".original"); err != nil {
			return err
		}
		return os.Symlink(target+".original", target)
	})
	if p != nil {
		_ = p.Close()
	}
	if !swapped || err == nil {
		t.Fatalf("unexpected open after final-component attack: swapped=%v err=%v", swapped, err)
	}
}

func TestASPSecureOpenPinnedDirectorySurvivesRename(t *testing.T) {
	f := newASPSecureFixture(t)
	originalBin := aspSecurePathInode(t, f.part(4))
	originalRoot := aspSecurePathInode(t, f.part(2))
	var swapped bool
	p, err := secureOpenASPAtRoot(f.root, aspSecurePreferredPath, f.uid, func(component string) error {
		if component != "external-asp" {
			return nil
		}
		swapped = true
		oldRoot := f.part(2) + ".renamed"
		if err := os.Rename(f.part(2), oldRoot); err != nil {
			return err
		}
		if err := os.Mkdir(f.part(2), 0700); err != nil {
			return err
		}
		if err := os.Mkdir(f.part(3), 0700); err != nil {
			return err
		}
		if err := os.WriteFile(f.part(4), aspSecureELF("DECOY"), 0755); err != nil {
			return err
		}
		return os.Chmod(f.part(4), 0755)
	})
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	if !swapped || aspSecureInode(t, p.executable) != originalBin ||
		aspSecureInode(t, p.externalRoot) != originalRoot {
		t.Fatal("renamed external root caused a new pathname to replace pinned inodes")
	}
	if aspSecurePathInode(t, f.part(4)) == originalBin {
		t.Fatal("test failed to replace visible path")
	}
	data := make([]byte, 64)
	n, err := p.executable.ReadAt(data, 0)
	if err != nil && !errors.Is(err, io.EOF) {
		t.Fatal(err)
	}
	if !bytes.Contains(data[:n], []byte("ORIGINAL")) || bytes.Contains(data[:n], []byte("DECOY")) {
		t.Fatalf("wrong pinned contents: %q", data[:n])
	}
}

func TestASPSecureOpenPinnedExecutableSurvivesReplacement(t *testing.T) {
	f := newASPSecureFixture(t)
	p, err := f.open()
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	original := aspSecureInode(t, p.executable)
	path := f.part(4)
	if err := os.Rename(path, path+".old"); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, aspSecureELF("REPLACED"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, 0755); err != nil {
		t.Fatal(err)
	}
	if aspSecurePathInode(t, path) == original {
		t.Fatal("visible path was not replaced")
	}
	if aspSecureInode(t, p.executable) != original {
		t.Fatal("pinned executable inode changed after rename")
	}
	payload := make([]byte, 64)
	n, _ := p.executable.ReadAt(payload, 0)
	if !bytes.Contains(payload[:n], []byte("ORIGINAL")) || bytes.Contains(payload[:n], []byte("REPLACED")) {
		t.Fatalf("pinned executable was re-resolved: %q", payload[:n])
	}
}

func TestASPSecureOpenFDErrorPathsNoLeaks(t *testing.T) {
	f := newASPSecureFixture(t)
	// First touch the FD enumerator and test opener before establishing the
	// baseline to avoid attributing one-time Go runtime allocation to leaks.
	p, err := f.open()
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Close(); err != nil {
		t.Fatal(err)
	}
	runtime.GC()
	start := aspSecureFDCount(t)
	for i := 0; i < 240; i++ {
		var hook func(string) error
		switch i % 6 {
		case 0:
			hook = func(component string) error {
				if component == "data" {
					return errors.New("injected early failure")
				}
				return nil
			}
		case 1:
			hook = func(component string) error {
				if component == "external-asp" {
					return errors.New("injected retained-root failure")
				}
				return nil
			}
		case 2:
			hook = func(component string) error {
				if component == "bin" {
					return errors.New("injected terminal-preopen failure")
				}
				return nil
			}
		case 3:
			hook = func(string) error { return nil }
		case 4:
			// Force a real openat(2) ENOENT after several parent FDs.
			if err := os.Rename(f.part(4), f.part(4)+".hidden"); err != nil {
				t.Fatal(err)
			}
		case 5:
			// Force a real fstat-policy rejection after final openat.
			if err := os.Chmod(f.part(4), 0777); err != nil {
				t.Fatal(err)
			}
		}
		p, err := secureOpenASPAtRoot(f.root, aspSecurePreferredPath, f.uid, hook)
		switch i % 6 {
		case 3:
			if err != nil {
				t.Fatal(err)
			}
			if err := p.Close(); err != nil {
				t.Fatal(err)
			}
		default:
			if err == nil || p != nil {
				t.Fatalf("failure path accepted iteration %d", i)
			}
		}
		switch i % 6 {
		case 4:
			if err := os.Rename(f.part(4)+".hidden", f.part(4)); err != nil {
				t.Fatal(err)
			}
		case 5:
			if err := os.Chmod(f.part(4), 0755); err != nil {
				t.Fatal(err)
			}
		}
	}
	end := aspSecureFDCount(t)
	if end > start+1 {
		t.Fatalf("FD leak after repeated failure/success paths: before=%d after=%d", start, end)
	}
	t.Logf("FD bounded: before=%d after=%d, 240 attempts (including ENOENT and fstat rejection)", start, end)
}
