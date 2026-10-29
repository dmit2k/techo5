package mic

import (
	"context"
	"errors"
	"github.com/HuskerMinion/techo5/echod/internal/asp"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/HuskerMinion/techo5/echod/internal/config"
)

var (
	errASPNotFound  = errors.New("external ASP binary not found")
	errASPStart     = errors.New("external ASP start failed")
	errASPHandshake = errors.New("external ASP handshake failed")
	errASPReference = errors.New("hardware reference unavailable")
	errASPNoPCM     = errors.New("no valid PCM frames received")
	errASPTimeout   = errors.New("external ASP processing timeout")
	errASPExited    = errors.New("external ASP exited")
	errASPInvalid   = errors.New("invalid external ASP PCM")
	errASPTimeline  = errors.New("external ASP timeline discontinuity")
)

const aspPreflightFrames = 3
const aspFirstPCMBudget = time.Second

// ASPStatus separates the requested choice from the last committed PCM source.
// Request orders user commands only; it is not an audio epoch and never resets consumers.
// Confirmed remains true across runtime fallback, so a coalesced notification cannot lose a save.
type ASPStatus struct {
	Selected, Active config.ASPMode
	Debug            string
	Request          uint64
	Confirmed        bool
}

type aspAttempt struct {
	ctx                     context.Context
	cancel                  context.CancelFunc
	keep, confirmed, failed bool
	good                    int
	feedback                asp.Status
	observed                asp.Status
	base, next              uint64
	based                   bool
	timer                   *time.Timer
}

type aspControl struct {
	aspMu                   sync.Mutex
	asp                     *externalASP
	aspAttempt              *aspAttempt
	aspStatus               ASPStatus
	aspChanged              chan struct{}
	aspStarting, aspCapture bool
	// Private injection point for deterministic host tests, not an executable registry.
	aspStart            func(context.Context) (*externalASP, error)
	aspTicks            uint64
	aspWarning          bool
	aspEventText        string
	aspEventAt          time.Time
	aspEventsSuppressed uint64
	aspNow              func() time.Time
	aspLog              *aspLogDispatch
	aspLogRequest       atomic.Uint64
}

func (s *Source) statusASPLocked() ASPStatus {
	st := s.aspStatus
	if st.Selected == "" {
		st.Selected = config.ASPBuiltin
	}
	if st.Active == "" {
		st.Active = config.ASPBuiltin
	}
	if st.Debug == "" {
		st.Debug = hostASPText(hostASPBuiltinOK, "Built-in processing OK")
	}
	return st
}

func (s *Source) ASP() config.ASPMode { return s.ASPStatus().Selected }

func (s *Source) ASPStatus() ASPStatus {
	s.aspMu.Lock()
	defer s.aspMu.Unlock()
	return s.statusASPLocked()
}

// ASPChanges coalesces state changes. The consumer must read the current snapshot, not replay
// stale events. No callbacks, disk I/O or entity publication run under source locks.
func (s *Source) ASPChanges() <-chan struct{} {
	s.aspMu.Lock()
	defer s.aspMu.Unlock()
	if s.aspChanged == nil {
		s.aspChanged = make(chan struct{}, 1)
	}
	return s.aspChanged
}

func (s *Source) notifyASPLocked() {
	if s.aspChanged != nil {
		select {
		case s.aspChanged <- struct{}{}:
		default:
		}
	}
}

// SetASP requests a path without waiting for a child or for capture. Only explicit commands
// create attempts; a failed external instance is never automatically restarted.
func (s *Source) SetASP(want config.ASPMode) config.ASPMode {
	if want != config.ASPExternal {
		want = config.ASPBuiltin
	}
	keep := config.Get().Microphone.ASP == config.ASPExternal
	s.aspMu.Lock()
	if want == config.ASPExternal && s.statusASPLocked().Selected == want {
		s.aspMu.Unlock()
		return want
	}
	old, a := s.asp, s.aspAttempt
	if a != nil && a.timer != nil {
		a.timer.Stop()
	}
	s.asp = nil
	s.aspAttempt = nil
	s.aspStatus = ASPStatus{Selected: want, Active: config.ASPBuiltin, Request: s.aspStatus.Request + 1}
	s.aspLogRequest.Store(s.aspStatus.Request)
	s.aspWarning = false
	s.aspEventText = ""
	s.aspEventAt = time.Time{}
	if want == config.ASPBuiltin {
		s.hostASPLocked(hostASPBuiltinOK, "Built-in processing OK", slog.LevelInfo)
	}
	launch := false
	if want == config.ASPExternal {
		ctx, cancel := context.WithCancel(context.Background())
		next := &aspAttempt{ctx: ctx, cancel: cancel, keep: keep}
		s.aspAttempt = next
		s.hostASPLocked(hostASPStarting, "External ASP starting", slog.LevelInfo)
		s.armASPLocked(next)
		// Startup configuration can be restored before ALSA is acquired.
		// Launch only once capture is ready, not against a nil hardware REF.
		if s.aspCapture && !s.aspStarting {
			s.aspStarting = true
			launch = true
		}
	}
	s.notifyASPLocked()
	s.aspMu.Unlock()
	if a != nil {
		a.cancel()
	}
	if old != nil {
		old.Close()
	}
	if launch {
		go s.startASPLoop()
	}
	return want
}

// At most one startup worker exists, even under arbitrarily rapid command toggles.
func (s *Source) startASPLoop() {
	for {
		s.aspMu.Lock()
		a := s.aspAttempt
		if a == nil || a.failed || s.asp != nil {
			s.aspStarting = false
			s.aspMu.Unlock()
			return
		}
		start := s.aspStart
		s.aspMu.Unlock()
		if start == nil {
			start = s.openExternalASP
		}
		f, err := start(a.ctx)
		s.aspMu.Lock()
		current := s.aspAttempt == a && a.ctx.Err() == nil && !a.failed
		if current {
			if err != nil {
				s.failASPLocked(a, err)
			} else {
				s.asp = f
			}
			s.aspStarting = false
		}
		s.aspMu.Unlock()
		if !current || err != nil {
			if f != nil {
				f.Close()
			}
		}
		if current {
			return
		}
	}
}

func (s *Source) armASPLocked(a *aspAttempt) {
	if !s.aspCapture || a.timer != nil || a.failed {
		return
	}
	a.timer = time.AfterFunc(aspFirstPCMBudget, func() {
		s.aspMu.Lock()
		var f *externalASP
		expired := false
		if s.aspAttempt == a && !a.confirmed && !a.failed {
			expired = true
			f = s.asp
			reason := errASPNoPCM
			if a.good > 0 {
				reason = errASPTimeout
			}
			s.failASPLocked(a, reason)
		}
		s.aspMu.Unlock()
		if expired {
			a.cancel()
		}
		if f != nil {
			f.Close()
		}
	})
}

func (s *Source) beginASPCapture() {
	s.aspMu.Lock()
	s.aspCapture = true
	launch := false
	if a := s.aspAttempt; a != nil && !a.failed {
		s.armASPLocked(a)
		if !s.aspStarting && s.asp == nil {
			s.aspStarting = true
			launch = true
		}
	}
	s.aspMu.Unlock()
	if launch {
		go s.startASPLoop()
	}
}

func aspFailure(err error, _ int) (uint16, string) {
	// Keep the wire status untouched for errors that originate in a helper.
	code := hostASPExited
	reason := errASPExited
	switch {
	case errors.Is(err, errASPNotFound):
		code, reason = hostASPNotFound, errASPNotFound
	case errors.Is(err, errASPStart):
		code, reason = hostASPStartFailed, errASPStart
	case errors.Is(err, errASPHandshake):
		code, reason = hostASPHandshakeFailed, errASPHandshake
	case errors.Is(err, errASPReference):
		code, reason = hostASPReferenceUnavailable, errASPReference
	case errors.Is(err, asp.ErrStatus):
		return hostASPInvalidStatus, "Fallback: invalid external ASP status"
	case errors.Is(err, errASPTimeline):
		return hostASPTimeline, "Fallback: external ASP timeline discontinuity"
	case errors.Is(err, errASPInvalid):
		code, reason = hostASPInvalidPCM, errASPInvalid
	case errors.Is(err, errASPNoPCM):
		code, reason = hostASPNoPCM, errASPNoPCM
	case errors.Is(err, errASPTimeout):
		code, reason = hostASPTimeout, errASPTimeout
	}
	return code, "Fallback: " + reason.Error()
}

func (s *Source) failASPLocked(a *aspAttempt, err error) {
	if a.failed {
		return
	}
	a.failed = true
	if a.timer != nil {
		a.timer.Stop()
	}
	s.asp = nil
	s.aspStatus.Active = config.ASPBuiltin
	s.failureASPLocked(err, a.good)
	s.aspStatus.Confirmed = a.confirmed
	if !a.keep && !a.confirmed {
		s.aspStatus.Selected = config.ASPBuiltin
	}
	s.notifyASPLocked()
}

type aspResult struct {
	attempt *aspAttempt
	child   *externalASP
	reply   asp.Reply
	tick    uint64
	err     error
}

// The single capture goroutine owns borrowed native bytes until this returns.
func (s *Source) processASP(n nativeFrame) aspResult {
	s.aspMu.Lock()
	a, f := s.aspAttempt, s.asp
	tick := s.aspTicks
	s.aspTicks++
	if f != nil && !a.based {
		a.base = tick * FrameSamples
		a.based = true
	}
	s.aspMu.Unlock()
	if f == nil {
		return aspResult{tick: tick}
	}
	reply, err := f.exchange(n.input())
	return aspResult{attempt: a, child: f, reply: reply, tick: tick, err: err}
}

// Called under Source.mu -> aspMu until history/spectrum/listener commit.
// Built-in always publishes this capture tick. A delayed External may overlap one source
// interval on activation; returning to Built-in skips that interval without retaining latency.
func (s *Source) chooseASP(stock []int16, r aspResult) []int16 {
	current := r.tick * FrameSamples
	if r.child == nil || r.child != s.asp || r.attempt != s.aspAttempt {
		return stock
	}
	a := r.attempt
	s.feedbackASPLocked(a, r.reply.ConfigStatus, "External ASP starting")
	if r.err != nil {
		s.failASPLocked(a, r.err)
		return stock
	}
	phase := "External ASP starting"
	if a.confirmed {
		phase = "External ASP processing OK"
	}
	s.feedbackASPLocked(a, r.reply.Status, phase)
	if r.reply.Type == asp.NotReady {
		return stock
	}
	// Origins belong to this attempt, including preflight, never to the preceding stock lane.
	// exchange validates wire order; this check also binds the run to real capture ticks.
	if current < a.base || r.reply.Start != a.next || r.reply.Start > current-a.base || current-a.base-r.reply.Start > FrameSamples {
		s.failASPLocked(a, errASPTimeline)
		return stock
	}
	a.next += FrameSamples
	if a.good < aspPreflightFrames {
		a.good++
		return stock
	}
	if !a.confirmed {
		a.confirmed = true
		if a.timer != nil {
			a.timer.Stop()
		}
		s.aspStatus.Active = config.ASPExternal
		s.aspStatus.Confirmed = true
		if a.feedback.Present() && a.feedback.Message == "" {
			def := statusDefault(a.feedback, "External ASP processing OK")
			text := externalASPText(a.feedback, def)
			if text != s.aspStatus.Debug {
				s.aspStatus.Debug = text
				s.logASPLocked(a.feedback, "external-asp", def, slog.LevelInfo)
			}
		}
		if !a.feedback.Present() && !s.aspWarning {
			s.hostASPLocked(hostASPProcessingOK, "External ASP processing OK", slog.LevelInfo)
		}
		s.notifyASPLocked()
	}
	return r.reply.PCM
}

func (s *Source) closeASP() {
	s.aspMu.Lock()
	f, a := s.asp, s.aspAttempt
	s.aspCapture = false
	if a != nil && !a.failed {
		s.failASPLocked(a, errASPExited)
	}
	s.aspMu.Unlock()
	if a != nil {
		a.cancel()
	}
	if f != nil {
		f.Close()
	}
}

func (s *Source) retireASP(r aspResult) {
	s.aspMu.Lock()
	retire := r.child != s.asp
	s.aspMu.Unlock()
	if retire {
		r.child.Close()
	}
}
