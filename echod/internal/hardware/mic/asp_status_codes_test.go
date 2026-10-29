package mic

import (
	"errors"
	"log/slog"
	"testing"

	"github.com/HuskerMinion/techo5/echod/internal/asp"
)

func TestHostASPStatusCodesUnique(t *testing.T) {
	codes := []uint16{
		hostASPBuiltinOK, hostASPStarting, hostASPProcessingOK,
		340, // asynchronous diagnostic queue overflow
		hostASPNotFound, hostASPStartFailed, hostASPHandshakeFailed,
		hostASPReferenceUnavailable, hostASPNoPCM, hostASPTimeout,
		hostASPExited, hostASPInvalidPCM, hostASPTimeline, hostASPInvalidStatus,
	}
	used := make(map[uint16]bool, len(codes))
	for _, code := range codes {
		if code < 200 || code > 599 || used[code] {
			t.Fatalf("invalid/duplicate host ASP code %d", code)
		}
		used[code] = true
	}
	reasons := []struct {
		err  error
		code uint16
	}{
		{errASPNotFound, 500},
		{errASPStart, 501},
		{errASPHandshake, 502},
		{errASPReference, 503},
		{errASPNoPCM, 504},
		{errASPTimeout, 505},
		{errASPExited, 506},
		{errASPInvalid, 507},
		{errASPTimeline, 508},
		{asp.ErrStatus, 509},
		{errors.New("unknown failure"), 506},
	}
	for _, tc := range reasons {
		code, message := aspFailure(tc.err, 0)
		if code != tc.code || message == "" {
			t.Fatalf("%v: %d %q, expected %d", tc.err, code, message, tc.code)
		}
	}
	// External helpers may use either recommended numeric half of the range.
	for _, code := range []uint16{200, 249, 250, 299, 300, 349, 350, 399, 400, 449, 450, 499, 500, 549, 550, 599} {
		if _, err := (asp.Status{Kind: asp.State, Code: code}).Validate(); err != nil {
			t.Fatalf("advisory split must not reject %d: %v", code, err)
		}
	}
}

func TestHostASPStatusPresentation(t *testing.T) {
	var records []aspLogRecord
	d := testLogDispatch(t, func(r aspLogRecord) { records = append(records, r) })
	s := testSource()
	s.aspLog = d
	s.aspStatus.Request = 1
	s.aspLogRequest.Store(1)
	s.aspMu.Lock()
	s.hostASPLocked(hostASPTimeout, "Fallback: external ASP processing timeout", slog.LevelError)
	s.aspMu.Unlock()
	drainLog(t, d, 1)
	if st := s.ASPStatus(); st.Debug != "H:505:Fallback: external ASP processing timeout" {
		t.Fatal(st)
	}
	if len(records) != 1 || records[0].code != "505" || records[0].origin != "echod" ||
		records[0].message != "Fallback: external ASP processing timeout" {
		t.Fatal(records)
	}
}
