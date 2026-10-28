package mic

import (
	"context"
	"errors"
	"fmt"
	"github.com/HuskerMinion/techo5/echod/internal/asp"
	"io"
	"os"
	"os/exec"
	"sync"
	"sync/atomic"
	"time"
)

const (
	nativeMicBytes = 1920 * 4 * 4
	nativeRefBytes = 1920 * 2 * 2
	aspFrameBudget = 25 * time.Millisecond
	aspStartBudget = 500 * time.Millisecond
)

// nativeFrame is borrowed from the single capture owner until its next Read. Microphones are
// interleaved S32_LE, references interleaved stereo S16_LE, all at 96 kHz for exactly 20 ms.
// fullBits describes the original signed microphone scale; no normalization or boost is applied.
type nativeFrame struct {
	mics, ref []byte
	fullBits  uint32
}

func (n nativeFrame) valid() bool {
	return len(n.mics) == nativeMicBytes && len(n.ref) == nativeRefBytes && (n.fullBits == 23 || n.fullBits == 31)
}

// externalASP is one source-owned child with bounded synchronous exchanges, never an ALSA owner.
// The capture owner serializes process; Close may
// interrupt it. Failure permanently closes this instance: no queued audio can cross a restart.
// Startup: ASP2 HELLO; the first actual input negotiates immutable CFG2 under one deadline.
// Typed replies carry explicit PCM origin and optional bounded helper feedback.
type externalASP struct {
	cmd        *exec.Cmd
	in         aspInputPipe
	out        aspOutputPipe
	done       chan struct{}
	once       sync.Once
	broken     atomic.Bool
	sequence   uint64
	configured bool
	config     asp.Config
	session    sessionCFG2
	havePCM    bool
	nextSample uint64
}

// Private pipe interfaces allow deterministic deadline/partial-I/O tests.
// Real children always use the original os.Pipe descriptors.
type aspInputPipe interface {
	io.WriteCloser
	SetWriteDeadline(time.Time) error
}
type aspOutputPipe interface {
	io.ReadCloser
	SetReadDeadline(time.Time) error
}

// startExternalASP requires a frozen CFG2 snapshot before any child is started.
func startExternalASP(ctx context.Context, cmd *exec.Cmd, config asp.Config, session sessionCFG2) (*externalASP, error) {
	return startExternalASPProcessWithRelease(ctx, cmd, &externalASP{config: config, session: session}, nil)
}

func validateASPSession(f *externalASP) error {
	if f == nil {
		return fmt.Errorf("%w: missing session", errASPInvalid)
	}
	config, session := f.config, f.session
	if len(session.packet) < asp.CFG2HeaderBytes || len(session.packet) > asp.CFG2MaxBytes {
		return fmt.Errorf("%w: invalid session size", errASPInvalid)
	}
	cfg, err := asp.ParseCFG2(session.Bytes())
	if err != nil || cfg.Config != config {
		return fmt.Errorf("%w: invalid or stale session", errASPInvalid)
	}
	// Recheck the session builder's host metadata qualification, including
	// factory gain tied to confirmed Athena captured slots. Never infer gain.
	if _, err := buildSessionCFG2(config, cfg.PlatformInfo, cfg.MicFactoryGain); err != nil {
		return fmt.Errorf("%w: invalid session metadata", errASPInvalid)
	}
	return nil
}

// release owns launch-only resources, not the live child. Pipes are attached
// before the single Start; the sole Wait belongs to the reaper below.
func startExternalASPProcessWithRelease(ctx context.Context, cmd *exec.Cmd, f *externalASP, release func()) (*externalASP, error) {
	closeLaunch := sync.OnceFunc(func() {
		if release != nil {
			release()
		}
	})
	defer closeLaunch()
	if ctx == nil || cmd == nil || f == nil {
		return nil, fmt.Errorf("%w: missing startup state", errASPStart)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := validateASPSession(f); err != nil {
		return nil, err
	}

	childIn, in, err := os.Pipe()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errASPStart, err)
	}
	out, childOut, err := os.Pipe()
	if err != nil {
		in.Close()
		childIn.Close()
		return nil, fmt.Errorf("%w: %w", errASPStart, err)
	}
	cmd.Stdin, cmd.Stdout, cmd.Stderr = childIn, childOut, os.Stderr
	err = cmd.Start()
	closeLaunch()
	if err != nil {
		in.Close()
		childIn.Close()
		out.Close()
		childOut.Close()
		return nil, fmt.Errorf("%w: %w", errASPStart, err)
	}
	childIn.Close()
	childOut.Close()
	f.cmd, f.in, f.out, f.done = cmd, in, out, make(chan struct{})
	go func() {
		_ = cmd.Wait()
		f.broken.Store(true)
		close(f.done)
	}()
	stopCancel := context.AfterFunc(ctx, f.Close)
	defer stopCancel()
	var ready [4]byte
	if err = out.SetReadDeadline(time.Now().Add(aspStartBudget)); err == nil {
		_, err = io.ReadFull(out, ready[:])
	}
	if err == nil && string(ready[:]) != "ASP2" {
		err = errASPHandshake
	}
	if err == nil {
		err = ctx.Err()
	}
	if err != nil {
		f.Close()
		return nil, fmt.Errorf("%w: %w", errASPHandshake, err)
	}
	return f, nil
}

// writeExact also supports writers that make partial progress without an error.
func writeExact(w io.Writer, p []byte) error {
	for len(p) > 0 {
		n, err := w.Write(p)
		if err != nil {
			return err
		}
		if n <= 0 || n > len(p) {
			return io.ErrShortWrite
		}
		p = p[n:]
	}
	return nil
}

func (n nativeFrame) input() asp.Input {
	if !n.valid() {
		return asp.Input{}
	}
	return asp.Input{Config: nativeASPConfig(n.fullBits), Mic: n.mics, Ref: n.ref}
}

// nativeASPConfig describes the captured representation without touching PCM.
func nativeASPConfig(fullBits uint32) asp.Config {
	if fullBits != 23 && fullBits != 31 {
		return asp.Config{}
	}
	bits, valid, align := uint8(fullBits), uint8(24), uint8(asp.Low)
	if bits == 31 {
		align = asp.High
	}
	return asp.Config{Flags: asp.Interleaved | asp.Synchronized | asp.HardwareRef,
		Mic: asp.Stream{Channels: 4, Frames: 1920, Rate: 96000, Format: asp.S32, FullBits: bits, ValidBits: valid, Alignment: align},
		Ref: asp.Stream{Channels: 2, Frames: 1920, Rate: 96000, Format: asp.S16, FullBits: 15, ValidBits: 16}}
}

// process is retained for the native 20-ms comparison fixtures.
func (f *externalASP) process(n nativeFrame) ([]int16, error) {
	r, e := f.exchange(n.input())
	return r.PCM, e
}

func (f *externalASP) exchange(i asp.Input) (reply asp.Reply, err error) {
	if f.broken.Load() {
		f.Close()
		return reply, errASPExited
	}
	defer func() {
		if err == nil {
			return
		}
		reply.PCM = nil
		var terminal *asp.Terminal
		switch {
		case errors.As(err, &terminal), errors.Is(err, asp.ErrStatus), errors.Is(err, errASPReference), errors.Is(err, errASPInvalid):
		case errors.Is(err, asp.ErrPCM), errors.Is(err, asp.ErrConfig):
			err = errASPInvalid
		case errors.Is(err, os.ErrDeadlineExceeded):
			err = errASPTimeout
		case errors.Is(err, io.ErrUnexpectedEOF):
			err = errASPInvalid
		default:
			err = errASPExited
		}
		f.Close()
	}()
	if i.Validate() != nil || i.Config != f.config {
		return reply, errASPReference
	}
	deadline := time.Now().Add(aspFrameBudget)
	if err = f.in.SetWriteDeadline(deadline); err != nil {
		return reply, err
	}
	if err = f.out.SetReadDeadline(deadline); err != nil {
		return reply, err
	}
	if !f.configured {
		wire := f.session.Bytes() // validated before startup; no metadata I/O
		if err = writeExact(f.in, wire); err != nil {
			return reply, err
		}
		var a asp.ConfigResult
		a, err = asp.ReadResult(f.out)
		if err != nil {
			return reply, err
		}
		if i.Config.Ref.Channels == 0 && (a.Flags&asp.NoReference == 0) || i.Config.Ref.Channels > 0 && a.Flags != 0 {
			return reply, asp.ErrStatus
		}
		f.configured = true
		f.config = i.Config
		reply.ConfigStatus = a.Status
	}
	if err = asp.WriteRequest(f.in, f.sequence, i); err != nil {
		return reply, err
	}
	c := reply.ConfigStatus
	reply, err = asp.ReadReply(f.out)
	reply.ConfigStatus = c
	if err != nil {
		return reply, err
	}
	if reply.Sequence != f.sequence {
		return reply, errASPInvalid
	}
	if reply.Type == asp.NotReady {
		if f.havePCM || f.sequence != 0 {
			return reply, errASPInvalid
		}
	} else {
		// Only current or exactly one capture-tick delayed origins are supported by this seam.
		if reply.Start != f.nextSample || reply.Start > f.sequence*FrameSamples || f.sequence*FrameSamples-reply.Start > FrameSamples {
			return reply, errASPInvalid
		}
		f.nextSample += FrameSamples
		f.havePCM = true
	}
	f.sequence++
	return reply, nil
}

// Close closes our actual descriptors as well as killing the child. A descendant inheriting a
// pipe cannot prolong a blocked transfer. Wait has a single reaper; capture never waits on reaping.
func (f *externalASP) Close() {
	f.once.Do(func() {
		f.broken.Store(true)
		f.in.Close()
		f.out.Close()
		if f.cmd != nil && f.cmd.Process != nil {
			_ = f.cmd.Process.Kill()
		}
	})
}

// sessionCFG2 is a validated, independently owned, immutable wire snapshot.
// A string intentionally prevents mutation of the stored CFG2 bytes. Bytes
// returns a fresh copy for each read (including when sent to an io.Writer).
// The zero value is not a valid session; only the builder constructs one.
type sessionCFG2 struct {
	packet string
}

// Bytes returns an owned copy. Mutating it never changes the session.
func (s sessionCFG2) Bytes() []byte { return []byte(s.packet) }

// buildSessionCFG2 is pure: no layout, IDME, filesystem or PCM access.
// It describes the supplied native capture format without modifying it.
// nil platform/gain means NOT ADVERTISED; a non-nil gain with NONE means
// explicitly available metadata reporting no valid factory coefficients.
//
// Factory gain is accepted only for the H1-A2-qualified confirmed Athena
// four-MIC/two-REF native capture shape. Reject incompatible caller-supplied
// metadata instead of silently advertising fabricated factory calibration.
func buildSessionCFG2(config asp.Config, platform *asp.PlatformInfo, gain *asp.MicFactoryGain) (sessionCFG2, error) {
	if _, _, err := config.Lengths(); err != nil {
		return sessionCFG2{}, fmt.Errorf("%w: invalid session MIC/REF config: %v", asp.ErrConfig, err)
	}
	if gain != nil && !qualifiedAthenaFactoryMIC(config, platform) {
		return sessionCFG2{}, fmt.Errorf("%w: unqualified session factory calibration", asp.ErrConfig)
	}

	// Copy typed values before serialization, with no aliases to caller data.
	// The accepted codec validates all typed fields and wire bounds.
	cfg := asp.CFG2{Config: config}
	if platform != nil {
		p := *platform
		cfg.PlatformInfo = &p
	}
	if gain != nil {
		g := *gain // coefficients are a fixed array, not a borrowed slice
		cfg.MicFactoryGain = &g
	}
	packet, err := cfg.MarshalBinary()
	if err != nil {
		return sessionCFG2{}, err
	}
	return sessionCFG2{packet: string(packet)}, nil
}

// prepareASP2SessionCFG2 collects metadata once in the startup worker, before
// capture exchange deadlines. It must never run in the capture PCM loop.
func prepareASP2SessionCFG2(config asp.Config) (sessionCFG2, error) {
	return prepareASP2SessionCFG2WithReaders(config, platformInfoForSession, micFactoryGainForSession)
}

// The reader seam enables offline tests without inspecting a real device.
// Both collectors are called at most once and only after config validation.
// Invalid platform metadata is rejected before attempting any gain read.
func prepareASP2SessionCFG2WithReaders(
	config asp.Config,
	platformReader func() *asp.PlatformInfo,
	gainReader func(asp.Config, *asp.PlatformInfo) *asp.MicFactoryGain,
) (sessionCFG2, error) {
	if _, _, err := config.Lengths(); err != nil {
		return sessionCFG2{}, fmt.Errorf("%w: invalid session MIC/REF config: %v", asp.ErrConfig, err)
	}
	if platformReader == nil || gainReader == nil {
		return sessionCFG2{}, fmt.Errorf("%w: missing session metadata collector", asp.ErrConfig)
	}
	platform := platformReader()
	if platform != nil {
		if err := platform.Validate(); err != nil {
			return sessionCFG2{}, err
		}
	}
	gain := gainReader(config, platform)
	return buildSessionCFG2(config, platform, gain)
}
