package mic

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/HuskerMinion/techo5/echod/internal/asp"
)

func (s *Source) logASPLocked(st asp.Status, origin, def string, level slog.Level) {
	if st.Warning() {
		level = slog.LevelWarn
	}
	if st.Terminal() {
		level = slog.LevelError
	}
	kind := "STATE"
	if st.Kind == asp.Event {
		kind = "EVENT"
	}
	msg := def
	if st.Message != "" {
		msg = st.Message
	}
	d := s.aspLog
	if d == nil {
		d = sharedASPLog()
	}
	d.enqueue(aspLogRecord{level: level, code: fmt.Sprintf("%03d", st.Code), kind: kind, origin: origin, message: msg, request: s.aspStatus.Request, current: &s.aspLogRequest})
}

// Host codes are local echod assignments; only the ASP2 numeric status travels on wire.
const (
	hostASPBuiltinOK            uint16 = 200
	hostASPStarting             uint16 = 210
	hostASPProcessingOK         uint16 = 220
	hostASPNotFound             uint16 = 500
	hostASPStartFailed          uint16 = 501
	hostASPHandshakeFailed      uint16 = 502
	hostASPReferenceUnavailable uint16 = 503
	hostASPNoPCM                uint16 = 504
	hostASPTimeout              uint16 = 505
	hostASPExited               uint16 = 506
	hostASPInvalidPCM           uint16 = 507
	hostASPTimeline             uint16 = 508
	hostASPInvalidStatus        uint16 = 509
)

func hostASPText(code uint16, text string) string {
	return fmt.Sprintf("H:%03d:%s", code, text)
}

func externalASPText(st asp.Status, def string) string {
	if st.Message != "" {
		def = st.Message
	}
	return fmt.Sprintf("E:%03d:%s", st.Code, def)
}

func (s *Source) hostASPLocked(code uint16, text string, level slog.Level) {
	s.aspStatus.Debug = hostASPText(code, text)
	s.logASPLocked(asp.Status{Kind: asp.State, Code: code}, "echod", text, level)
}
func statusDefault(st asp.Status, def string) string {
	if st.Kind == asp.Event {
		return "External ASP event"
	}
	if st.Code >= 400 && st.Code < 500 {
		return "External ASP configuration degraded"
	}
	if st.Code >= 300 && st.Code < 400 {
		return "External ASP processing degraded"
	}
	return def
}
func (s *Source) feedbackASPLocked(a *aspAttempt, st asp.Status, def string) {
	if !st.Present() {
		return
	}
	def = statusDefault(st, def)
	if st.Kind == asp.State {
		if a.observed == st {
			return
		}
		a.observed = st
		s.logASPLocked(st, "external-asp", def, slog.LevelInfo)
		if s.aspWarning && st.Code == 0 {
			return
		}
		a.feedback = st
		s.aspWarning = st.Warning()
		s.aspStatus.Debug = externalASPText(st, def)
		s.notifyASPLocked()
		return
	}
	// Every explicitly supplied EVENT is eligible for logging, even consecutive identical events.
	s.logASPLocked(st, "external-asp", def, slog.LevelInfo)
	now := time.Now()
	if s.aspNow != nil {
		now = s.aspNow()
	}
	text := externalASPText(st, def)
	if s.aspWarning || text == s.aspEventText || !s.aspEventAt.IsZero() && now.Sub(s.aspEventAt) < 5*time.Second {
		s.aspEventsSuppressed++
		return
	}
	s.aspEventText, s.aspEventAt = text, now
	s.aspStatus.Debug = text
	s.notifyASPLocked()
}
func (s *Source) failureASPLocked(err error, good int) {
	var terminal *asp.Terminal
	if errors.As(err, &terminal) {
		def := "Fallback: external ASP fatal error"
		if terminal.Configuration {
			def = "Fallback: external ASP configuration rejected"
		}
		s.aspStatus.Debug = externalASPText(terminal.Status, def)
		s.logASPLocked(terminal.Status, "external-asp", def, slog.LevelError)
	} else {
		code, reason := aspFailure(err, good)
		s.hostASPLocked(code, reason, slog.LevelError)
	}
	s.aspWarning = true
}
