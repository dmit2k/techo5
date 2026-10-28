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
func (s *Source) hostASPLocked(text string, level slog.Level) {
	s.aspStatus.Debug = (asp.Status{}).Text(text)
	s.logASPLocked(asp.Status{}, "echod", text, level)
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
		s.aspStatus.Debug = st.Text(def)
		s.notifyASPLocked()
		return
	}
	// Every explicitly supplied EVENT is eligible for logging, even consecutive identical events.
	s.logASPLocked(st, "external-asp", def, slog.LevelInfo)
	now := time.Now()
	if s.aspNow != nil {
		now = s.aspNow()
	}
	text := st.Text(def)
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
		s.aspStatus.Debug = terminal.Status.Text(def)
		s.logASPLocked(terminal.Status, "external-asp", def, slog.LevelError)
	} else {
		s.hostASPLocked(aspFailure(err, good), slog.LevelError)
	}
	s.aspWarning = true
}
