//go:build linux

package mic

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"debug/elf"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"
)

type aspProbe struct {
	path, nonce, interp string
}

func buildASPProbe(t *testing.T, mode string) aspProbe {
	t.Helper()
	cc, err := exec.LookPath("cc") // Test tool only, never executable discovery.
	if err != nil {
		t.Fatal("installed C compiler required:", err)
	}
	var nonce [16]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		t.Fatal(err)
	}
	p := aspProbe{path: filepath.Join(t.TempDir(), "probe"), nonce: fmt.Sprintf("%x", nonce)}
	args := []string{"-std=c11", "-O2", "-Wall", "-Wextra", "-Werror",
		"-DPROBE_NONCE=\"" + p.nonce + "\"", "testdata/asp_pinned_probe.c", "-o", p.path}
	switch mode {
	case "static":
		args = append(args, "-static")
	case "missing-interpreter":
		args = append(args, "-Wl,--dynamic-linker="+filepath.Join(t.TempDir(), "missing-ld.so"))
	}
	out, err := exec.Command(cc, args...).CombinedOutput()
	if err != nil {
		if mode == "static" {
			t.Skipf("static cc unavailable: %v; %s", err, out)
		}
		t.Fatalf("cc: %v: %s", err, out)
	}
	e, err := elf.Open(p.path)
	if err != nil {
		t.Fatal(err)
	}
	for _, ph := range e.Progs {
		if ph.Type == elf.PT_INTERP {
			b, err := io.ReadAll(ph.Open())
			if err != nil {
				t.Fatal(err)
			}
			p.interp = strings.TrimRight(string(b), "\x00")
		}
	}
	e.Close()
	if mode == "static" && p.interp != "" || mode != "static" && p.interp == "" {
		t.Fatalf("unexpected PT_INTERP for %s: %q", mode, p.interp)
	}
	t.Logf("ELF mode=%s nonce=%s PT_INTERP=%q", mode, p.nonce, p.interp)
	if dir := os.Getenv("TECHO5_FD_EVIDENCE_DIR"); dir != "" {
		b, err := os.ReadFile(p.path)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, mode+"-"+p.nonce+".elf"), b, 0700); err != nil {
			t.Fatal(err)
		}
	}
	return p
}

func newASPProbeFixture(t *testing.T, probe aspProbe, asset string) *aspSecureFixture {
	t.Helper()
	f := newASPSecureFixture(t)
	b, err := os.ReadFile(probe.path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(f.part(4), b, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(f.part(2), "etc"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(f.part(2), "etc", "profile.conf"), []byte(asset), 0600); err != nil {
		t.Fatal(err)
	}
	return f
}

func pinASPProbe(t *testing.T, f *aspSecureFixture) *secureASPPinned {
	t.Helper()
	p, err := f.open()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = p.Close() })
	return p
}

func aspProbeArgs(t *testing.T, p *secureASPPinned, asset, mode string) []string {
	t.Helper()
	var bin, root syscall.Stat_t
	if err := syscall.Fstat(int(p.executable.Fd()), &bin); err != nil {
		t.Fatal(err)
	}
	if err := syscall.Fstat(int(p.externalRoot.Fd()), &root); err != nil {
		t.Fatal(err)
	}
	return []string{fmt.Sprint(bin.Dev), fmt.Sprint(bin.Ino), fmt.Sprint(root.Dev), fmt.Sprint(root.Ino), asset, mode}
}

func assertASPConsumed(t *testing.T, p *secureASPPinned) {
	t.Helper()
	for _, f := range []*os.File{p.executable, p.externalRoot} {
		if f != nil {
			if _, err := f.Stat(); !errors.Is(err, os.ErrClosed) {
				t.Fatalf("parent descriptor not closed: %v", err)
			}
		}
	}
}

func runASPProbe(t *testing.T, p *secureASPPinned, probe aspProbe, asset string) string {
	t.Helper()
	args := aspProbeArgs(t, p, "etc/profile.conf", "exit")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var out, stderr bytes.Buffer
	cmd, err := startSecureASPPinned(ctx, p, args, &out, &stderr)
	assertASPConsumed(t, p)
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Wait(); err != nil {
		t.Fatalf("Wait: %v; stdout=%s stderr=%s", err, &out, &stderr)
	}
	if !cmd.ProcessState.Exited() || !cmd.ProcessState.Success() {
		t.Fatal("successful child not reaped")
	}
	s := out.String()
	want := fmt.Sprintf("nonce=%s exe=%s:%s fd3=%s:%s fd4=%s:%s asset=%s\n",
		probe.nonce, args[0], args[1], args[0], args[1], args[2], args[3], asset)
	if s != want {
		t.Fatalf("wrong executable/assets: got %q want %q", s, want)
	}
	t.Log(strings.TrimSpace(s))
	return s
}

func TestASPPinnedExecNativeELF(t *testing.T) {
	for _, mode := range []string{"dynamic", "static"} {
		t.Run(mode, func(t *testing.T) {
			probe := buildASPProbe(t, mode)
			f := newASPProbeFixture(t, probe, "ORIGINAL-ASSET")
			// Unrelated CLOEXEC descriptors must not appear in the child.
			sentinel, err := os.Open(f.base)
			if err != nil {
				t.Fatal(err)
			}
			defer sentinel.Close()
			// Loader overrides must not be passed to the native child.
			t.Setenv("LD_PRELOAD", "/offline-missing-preload.so")
			runASPProbe(t, pinASPProbe(t, f), probe, "ORIGINAL-ASSET")
		})
	}
}

func TestASPPinnedExecAdversarialReplacements(t *testing.T) {
	original, decoy := buildASPProbe(t, "dynamic"), buildASPProbe(t, "dynamic")
	// Prove the decoy itself is executable with its different nonce/assets.
	df := newASPProbeFixture(t, decoy, "DECOY-ASSET")
	runASPProbe(t, pinASPProbe(t, df), decoy, "DECOY-ASSET")
	for _, mode := range []string{"executable", "root", "both", "executable-symlink", "root-symlink", "unlinked-executable"} {
		t.Run(mode, func(t *testing.T) {
			f := newASPProbeFixture(t, original, "ORIGINAL-ASSET")
			p := pinASPProbe(t, f)
			binIno, rootIno := aspSecureInode(t, p.executable), aspSecureInode(t, p.externalRoot)
			if strings.HasPrefix(mode, "executable") || mode == "both" || mode == "unlinked-executable" {
				if err := os.Rename(f.part(4), f.part(4)+".original"); err != nil {
					t.Fatal(err)
				}
				if mode == "executable-symlink" {
					if err := os.Symlink(decoy.path, f.part(4)); err != nil {
						t.Fatal(err)
					}
				} else {
					b, err := os.ReadFile(decoy.path)
					if err != nil {
						t.Fatal(err)
					}
					if err := os.WriteFile(f.part(4), b, 0755); err != nil {
						t.Fatal(err)
					}
				}
				if mode == "unlinked-executable" {
					if err := os.Remove(f.part(4) + ".original"); err != nil {
						t.Fatal(err)
					}
				}
			}
			if mode == "root" || mode == "both" || mode == "root-symlink" {
				if err := os.Rename(f.part(2), f.part(2)+".original"); err != nil {
					t.Fatal(err)
				}
				if mode == "root-symlink" {
					if err := os.Symlink(df.part(2), f.part(2)); err != nil {
						t.Fatal(err)
					}
				} else {
					replacement := newASPProbeFixture(t, decoy, "DECOY-ASSET")
					if err := os.Rename(replacement.part(2), f.part(2)); err != nil {
						t.Fatal(err)
					}
				}
			}
			t.Logf("pinned exe inode=%d root inode=%d visible exe=%d root=%d",
				binIno, rootIno, aspSecurePathInode(t, f.part(4)), aspSecurePathInode(t, f.part(2)))
			if aspSecurePathInode(t, f.part(4)) == binIno {
				t.Fatal("attack did not replace visible executable")
			}
			runASPProbe(t, p, original, "ORIGINAL-ASSET")
		})
	}
}

func TestASPPinnedExecConcurrentRootSwaps(t *testing.T) {
	original, decoy := buildASPProbe(t, "dynamic"), buildASPProbe(t, "dynamic")
	f, d := newASPProbeFixture(t, original, "ORIGINAL-ASSET"), newASPProbeFixture(t, decoy, "DECOY-ASSET")
	const launches = 24
	pins := make([]*secureASPPinned, launches)
	for i := range pins {
		pins[i] = pinASPProbe(t, f)
	}
	// Keep an independent CLOEXEC directory handle for concurrent executable
	// name swaps even while this directory's visible name is being replaced.
	swapRoot, err := os.Open(f.part(2))
	if err != nil {
		t.Fatal(err)
	}
	defer swapRoot.Close()
	b, err := os.ReadFile(decoy.path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(f.part(3), "decoy"), b, 0755); err != nil {
		t.Fatal(err)
	}
	stop, ready := make(chan struct{}), make(chan struct{})
	type swapResult struct {
		count int
		err   error
	}
	result := make(chan swapResult, 1)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 256; i++ {
			for _, pair := range [][2]string{{"bin/external-asp", "bin/swap"}, {"bin/decoy", "bin/external-asp"}, {"bin/swap", "bin/decoy"}} {
				if err := syscall.Renameat(int(swapRoot.Fd()), pair[0], int(swapRoot.Fd()), pair[1]); err != nil {
					if i == 0 {
						close(ready)
					}
					result <- swapResult{i, err}
					return
				}
			}
			for _, pair := range [][2]string{{f.part(2), f.part(2) + ".swap"}, {d.part(2), f.part(2)}, {f.part(2) + ".swap", d.part(2)}} {
				if err := os.Rename(pair[0], pair[1]); err != nil {
					if i == 0 {
						close(ready)
					}
					result <- swapResult{i, err}
					return
				}
			}
			if i == 0 {
				close(ready)
			}
			select {
			case <-stop:
				result <- swapResult{i + 1, nil}
				return
			case <-time.After(time.Millisecond):
			}
		}
		result <- swapResult{256, nil}
	}()
	defer func() {
		close(stop)
		wg.Wait()
		r := <-result
		if r.err != nil || r.count == 0 {
			t.Errorf("concurrent swaps: count=%d err=%v", r.count, r.err)
		}
		t.Logf("completed %d concurrent executable + root swap cycles", r.count)
	}()
	<-ready
	for _, p := range pins {
		runASPProbe(t, p, original, "ORIGINAL-ASSET")
	}
	t.Logf("%d real ELF launches overlapped bounded executable and directory inode swaps", launches)
}

func TestASPPinnedExecRejectInvalidDescriptors(t *testing.T) {
	probe := buildASPProbe(t, "dynamic")
	for _, mode := range []string{"nil-pin", "nil-executable", "nil-root", "closed-pin", "closed-executable", "closed-root", "wrong-executable-type", "wrong-root-type", "nil-context", "cancelled-context", "missing-procfd", "pathname-injection"} {
		t.Run(mode, func(t *testing.T) {
			f := newASPProbeFixture(t, probe, "ORIGINAL-ASSET")
			p := pinASPProbe(t, f)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			procFD := "/proc/self/fd/3"
			switch mode {
			case "nil-pin":
				p.Close()
				p = nil
			case "nil-executable":
				p.executable.Close()
				p.executable = nil
			case "nil-root":
				p.externalRoot.Close()
				p.externalRoot = nil
			case "closed-pin":
				p.Close()
			case "closed-executable":
				p.executable.Close()
			case "closed-root":
				p.externalRoot.Close()
			case "wrong-executable-type":
				p.executable.Close()
				var err error
				p.executable, err = os.Open(f.part(2))
				if err != nil {
					t.Fatal(err)
				}
			case "wrong-root-type":
				p.externalRoot.Close()
				var err error
				p.externalRoot, err = os.Open(f.part(4))
				if err != nil {
					t.Fatal(err)
				}
			case "cancelled-context":
				cancel()
			case "missing-procfd":
				procFD = "/proc/self/fd/-1"
			case "pathname-injection":
				procFD = f.part(4)
			}
			callCtx := ctx
			if mode == "nil-context" {
				callCtx = nil
			}
			var out bytes.Buffer
			cmd, err := startSecureASPPinnedAtProcFD(callCtx, p, procFD, nil, &out, &out)
			if cmd != nil || err == nil {
				if cmd != nil {
					cmd.Process.Kill()
					cmd.Wait()
				}
				t.Fatalf("invalid input launched: %v", err)
			}
			if p != nil {
				assertASPConsumed(t, p)
			}
			if out.Len() != 0 {
				t.Fatal("failed launch produced child output")
			}
			t.Logf("fail closed: %v", err)
		})
	}
}

func TestASPPinnedExecKernelRejectsMalformedAndInterpreter(t *testing.T) {
	probe := buildASPProbe(t, "dynamic")
	b, err := os.ReadFile(probe.path)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name string
		data []byte
	}{
		{"magic-only", []byte{0x7f, 'E', 'L', 'F'}},
		{"truncated-header", b[:32]},
		{"corrupt-ELF", append([]byte{0x7f, 'E', 'L', 'F'}, make([]byte, 128)...)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newASPProbeFixture(t, probe, "ORIGINAL-ASSET")
			if err := os.WriteFile(f.part(4), tc.data, 0755); err != nil {
				t.Fatal(err)
			}
			p := pinASPProbe(t, f) // Real S1-A1 opener deliberately admits ELF magic.
			cmd, err := startSecureASPPinned(context.Background(), p, nil, nil, nil)
			assertASPConsumed(t, p)
			if cmd != nil || !errors.Is(err, syscall.ENOEXEC) {
				if cmd != nil {
					cmd.Process.Kill()
					cmd.Wait()
				}
				t.Fatalf("kernel did not reject ELF with ENOEXEC: %v", err)
			}
			t.Logf("kernel rejection after successful secure open: %v", err)
		})
	}
	t.Run("missing-PT_INTERP", func(t *testing.T) {
		pb := buildASPProbe(t, "missing-interpreter")
		f := newASPProbeFixture(t, pb, "ORIGINAL-ASSET")
		p := pinASPProbe(t, f)
		cmd, err := startSecureASPPinned(context.Background(), p, nil, nil, nil)
		assertASPConsumed(t, p)
		if cmd != nil || !errors.Is(err, syscall.ENOENT) {
			if cmd != nil {
				cmd.Process.Kill()
				cmd.Wait()
			}
			t.Fatalf("missing interpreter did not fail closed: %v", err)
		}
		t.Logf("kernel resolves PT_INTERP separately: %q -> %v", pb.interp, err)
	})
	t.Run("shebang", func(t *testing.T) {
		f := newASPProbeFixture(t, probe, "ORIGINAL-ASSET")
		if err := os.WriteFile(f.part(4), []byte("#!/bin/sh\nexit 0\n"), 0755); err != nil {
			t.Fatal(err)
		}
		p, err := f.open()
		if err == nil || p != nil {
			if p != nil {
				p.Close()
			}
			t.Fatal("shebang admitted")
		}
		t.Logf("secure opener rejects shebang: %v", err)
	})
}

func TestASPPinnedExecAssetsRejectTraversalAndSymlinks(t *testing.T) {
	probe := buildASPProbe(t, "dynamic")
	f := newASPProbeFixture(t, probe, "ORIGINAL-ASSET")
	for _, link := range [][2]string{
		{"etc/profile.conf", "profile-link"}, {"etc", "etc-link"},
	} {
		if err := os.Symlink(link[0], filepath.Join(f.part(2), link[1])); err != nil {
			t.Fatal(err)
		}
	}
	if err := syscall.Mkfifo(filepath.Join(f.part(2), "fifo"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"", "/etc/profile.conf", "../etc/profile.conf", "etc/../etc/profile.conf", "./etc/profile.conf",
		"etc//profile.conf", "etc/profile.conf/", "profile-link", "etc-link/profile.conf", "etc", "fifo", "missing"} {
		t.Run(strings.ReplaceAll(path, "/", "_"), func(t *testing.T) {
			p := pinASPProbe(t, f)
			args := aspProbeArgs(t, p, path, "exit")
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			var out bytes.Buffer
			cmd, err := startSecureASPPinned(ctx, p, args, &out, &out)
			if err != nil {
				t.Fatal(err)
			}
			err = cmd.Wait()
			var exit *exec.ExitError
			if !errors.As(err, &exit) || exit.ExitCode() != 30 || out.Len() != 0 {
				t.Fatalf("unsafe asset accepted or wrong failure: %v output=%q", err, &out)
			}
		})
	}
}

func assertASPReaped(t *testing.T, cmd *exec.Cmd) {
	t.Helper()
	if cmd.ProcessState == nil {
		t.Fatal("Wait did not record child state")
	}
	var status syscall.WaitStatus
	_, err := syscall.Wait4(cmd.Process.Pid, &status, syscall.WNOHANG, nil)
	if !errors.Is(err, syscall.ECHILD) {
		t.Fatalf("child not reaped: %v", err)
	}
}

func TestASPPinnedExecCancellationAndTimeout(t *testing.T) {
	probe := buildASPProbe(t, "dynamic")
	f := newASPProbeFixture(t, probe, "ORIGINAL-ASSET")
	for _, mode := range []string{"cancel", "timeout"} {
		t.Run(mode, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			if mode == "timeout" {
				var done context.CancelFunc
				ctx, done = context.WithTimeout(ctx, 150*time.Millisecond)
				defer done()
			}
			p := pinASPProbe(t, f)
			args := aspProbeArgs(t, p, "etc/profile.conf", "sleep")
			r, w, err := os.Pipe()
			if err != nil {
				t.Fatal(err)
			}
			defer r.Close()
			defer w.Close()
			cmd, err := startSecureASPPinned(ctx, p, args, w, w)
			if err != nil {
				t.Fatal(err)
			}
			waited := false
			defer func() {
				if !waited {
					cmd.Process.Kill()
					cmd.Wait()
				}
			}()
			w.Close()
			assertASPConsumed(t, p)
			if err := r.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
				t.Fatal(err)
			}
			line, err := bufio.NewReader(r).ReadString('\n')
			if err != nil || !strings.Contains(line, "nonce="+probe.nonce) {
				t.Fatalf("child not ready: %q %v", line, err)
			}
			if mode == "cancel" {
				cancel()
			}
			start := time.Now()
			err = cmd.Wait()
			waited = true
			if err == nil || ctx.Err() == nil || time.Since(start) > time.Second {
				t.Fatalf("unbounded cancellation: %v %v", err, ctx.Err())
			}
			assertASPReaped(t, cmd)
			t.Logf("parent FDs closed while child alive; %s -> %v; Wait reaped pid=%d in %s", mode, ctx.Err(), cmd.Process.Pid, time.Since(start))
		})
	}
}

func aspProbeChildren(t *testing.T) string {
	t.Helper()
	names, err := filepath.Glob("/proc/self/task/*/children")
	if err != nil {
		t.Fatal(err)
	}
	var children []string
	for _, name := range names {
		b, err := os.ReadFile(name)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			t.Fatal(err)
		}
		children = append(children, strings.Fields(string(b))...)
	}
	return strings.Join(children, " ")
}

func TestASPPinnedExecLifecycleNoLeaks(t *testing.T) {
	probe := buildASPProbe(t, "dynamic")
	f := newASPProbeFixture(t, probe, "ORIGINAL-ASSET")
	runASPProbe(t, pinASPProbe(t, f), probe, "ORIGINAL-ASSET") // Warm runtime pipes/poller.
	runtime.GC()
	beforeFD, beforeG := aspSecureFDCount(t), runtime.NumGoroutine()
	beforeChildren := aspProbeChildren(t)
	for i := 0; i < 96; i++ {
		p, err := f.open()
		if err != nil {
			t.Fatal(err)
		}
		mode, procFD := "exit", "/proc/self/fd/3"
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		args := aspProbeArgs(t, p, "etc/profile.conf", mode)
		if i%4 == 1 {
			procFD = "/proc/self/fd/-1"
		}
		if i%4 == 2 {
			cancel()
		}
		if i%4 == 3 {
			args[len(args)-1] = "sleep"
		}
		cmd, err := startSecureASPPinnedAtProcFD(ctx, p, procFD, args, io.Discard, io.Discard)
		assertASPConsumed(t, p)
		if i%4 == 1 || i%4 == 2 {
			if err == nil || cmd != nil {
				t.Fatal("failure path launched")
			}
		} else {
			if err != nil {
				t.Fatal(err)
			}
			if i%4 == 3 {
				cancel()
			}
			err = cmd.Wait()
			if i%4 == 0 && err != nil {
				t.Fatal(err)
			}
			if i%4 == 3 && err == nil {
				t.Fatal("cancelled sleeper succeeded")
			}
			assertASPReaped(t, cmd)
		}
		cancel()
	}
	deadline := time.Now().Add(time.Second)
	for runtime.NumGoroutine() > beforeG && time.Now().Before(deadline) {
		runtime.Gosched()
		time.Sleep(time.Millisecond)
	}
	afterFD, afterG := aspSecureFDCount(t), runtime.NumGoroutine()
	afterChildren := aspProbeChildren(t)
	t.Logf("96 mixed launch/fail/cancel iterations: FD %d -> %d; goroutines %d -> %d; children %q -> %q",
		beforeFD, afterFD, beforeG, afterG, beforeChildren, afterChildren)
	if afterFD != beforeFD || afterG > beforeG || afterChildren != beforeChildren {
		t.Fatal("FD/goroutine/process leak")
	}
}
