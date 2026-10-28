package mic

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	"github.com/HuskerMinion/techo5/echod/internal/asp"
)

func TestPlatformInfoEvidenceConfidence(t *testing.T) {
	const (
		checkersLCM = "lcm=1-st7701s_wsvga_dsi_vdo_checkers_st_kd_hsd"
		crownLCM    = "lcm=1-jd936x_wxga_dsi_vdo_crown_st_kd_hsd"
		cronosLCM   = "lcm=1-st7701s_wsvga_dsi_vdo_cronos_st_truly"
	)
	tests := []struct {
		name, board, cmdline, doubt string
		compatible                  []string
		want                        asp.PlatformDetection
	}{
		{"athena bare DT", "athena", "", "", []string{"amzn,athena"}, asp.PlatformConfirmed},
		{"athena stock revisions", "athena", "console=tty0", "", []string{"amzn,athena-pvt", "amzn,athena-dvt", "amzn,athena"}, asp.PlatformConfirmed},
		{"athena single revision", "athena", "", "", []string{"amzn,athena-evt"}, asp.PlatformConfirmed},
		{"athena hyphenated revision", "athena", "", "", []string{"amzn,athena-pvt-1"}, asp.PlatformConfirmed},
		{"athena unrelated DT", "athena", "", "", []string{"google,kukui", "mediatek,mt8183"}, asp.PlatformInferred},
		{"athena similar DT", "athena", "", "", []string{"amzn,athenaeum"}, asp.PlatformInferred},
		{"athena empty revision", "athena", "", "", []string{"amzn,athena-"}, asp.PlatformInferred},
		{"athena malformed revision", "athena", "", "", []string{"amzn,athena-pvt!"}, asp.PlatformInferred},
		{"athena mismatching DT", "cronos", "", "", []string{"amzn,athena-pvt"}, asp.PlatformInferred},
		{"athena conflicted lcm", "athena", crownLCM, "", []string{"amzn,athena-pvt"}, asp.PlatformInferred},
		{"cronos conflicted DT", "cronos", cronosLCM, "", []string{"amzn,athena"}, asp.PlatformInferred},
		{"checkers lcm", "checkers", checkersLCM, "", nil, asp.PlatformConfirmed},
		{"crown lcm", "crown", crownLCM, "", nil, asp.PlatformConfirmed},
		{"cronos lcm", "cronos", cronosLCM, "", nil, asp.PlatformConfirmed},
		{"kernel extra tokens", "crown", "serialno=TESTCASE " + crownLCM + " console=tty0", "", nil, asp.PlatformConfirmed},
		{"unrelated DT plus lcm", "checkers", checkersLCM, "", []string{"mediatek,mt8163"}, asp.PlatformConfirmed},
		{"lcm different board", "checkers", cronosLCM, "", nil, asp.PlatformInferred},
		{"multiple conflicting lcm", "checkers", checkersLCM + " " + cronosLCM, "", nil, asp.PlatformInferred},
		{"two boards in one lcm", "checkers", "lcm=1-panel_checkers_cronos_st", "", nil, asp.PlatformInferred},
		{"non-lcm param", "checkers", "androidboot.product=checkers", "", nil, asp.PlatformInferred},
		{"not an lcm token", "checkers", "notlcm=foo_checkers_bar", "", nil, asp.PlatformInferred},
		{"malformed lcm prefix", "checkers", "lcm=panel_checkers_st", "", nil, asp.PlatformInferred},
		{"lcm substring near miss", "checkers", "lcm=1-panel_checkersfake_st", "", nil, asp.PlatformInferred},
		{"lcm prefixed near miss", "crown", "lcm=1-panel_xcrown_st", "", nil, asp.PlatformInferred},
		{"gating fallback no lcm", "checkers", "console=tty0", "", nil, asp.PlatformInferred},
		{"no gating fallback no lcm", "cronos", "console=tty0", "", nil, asp.PlatformInferred},
		{"empty cmdline", "cronos", "", "", nil, asp.PlatformInferred},
		{"empty DT", "athena", "", "", nil, asp.PlatformInferred},
		{"doubt overrides lcm", "crown", crownLCM, "ambiguous", nil, asp.PlatformInferred},
		{"doubt overrides DT", "athena", "", "fallback reason", []string{"amzn,athena"}, asp.PlatformInferred},
		{"doubt with no evidence", "cronos", "", "unidentified", nil, asp.PlatformInferred},
		{"skylight DT", "skylight", "", "", []string{"skylight,cal150", "rockchip,rk3566"}, asp.PlatformConfirmed},
		{"skylight lookalike", "skylight", "", "", []string{"skylight,cal150x"}, asp.PlatformInferred},
		{"skylight conflicting DT", "skylight", "", "", []string{"skylight,cal150", "amzn,athena"}, asp.PlatformInferred},
		{"fixed spot no proof", "rook", "", "", nil, asp.PlatformInferred},
		{"non-show dot no proof", "biscuit", "", "", nil, asp.PlatformInferred},
		{"unknown but valid board", "future-board_1", "", "", nil, asp.PlatformInferred},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := platformInfoFromEvidence(tt.board, tt.cmdline, tt.compatible, tt.doubt)
			if got == nil || got.Codename != tt.board || got.Detection != tt.want {
				t.Fatalf("platformInfoFromEvidence() = %+v, want board %q detection %d", got, tt.board, tt.want)
			}
		})
	}
}

func TestPlatformInfoInvalidCandidatesOmitted(t *testing.T) {
	for _, board := range []string{
		"", "Athena", "echo show", "board/device", "UPPER",
		"écho", "board\x00other", strings.Repeat("x", 33),
	} {
		t.Run(strings.ReplaceAll(board, "\x00", "NUL"), func(t *testing.T) {
			if got := platformInfoFromEvidence(board, "lcm=1-panel_cronos_test", []string{"amzn,athena"}, ""); got != nil {
				t.Fatalf("invalid board %q advertised as %+v", board, got)
			}
		})
	}
	// Do not over-constrain beyond the accepted encoder's validity rules.
	for _, board := range []string{"a", "x_y-z", strings.Repeat("x", 32)} {
		got := platformInfoFromEvidence(board, "", nil, "")
		if got == nil || got.Detection != asp.PlatformInferred {
			t.Fatalf("encoder-valid board %q was rejected or confirmed: %+v", board, got)
		}
	}
}

func TestPlatformInfoSnapshotWirePrivacyAndPCMIndependence(t *testing.T) {
	compatible := []string{"amzn,athena-pvt", "mediatek,mt8183"}
	const cmdline = "serialno=SYNTHETIC_SERIAL_98765 androidboot.product=SECRET_PRODUCT lcm=1-panel_unrelated_test"
	const doubt = ""
	info := platformInfoFromEvidence("athena", cmdline, compatible, doubt)
	if info == nil || info.Detection != asp.PlatformConfirmed {
		t.Fatalf("expected confirmed Athena from DT, got %+v", info)
	}
	compatible[0] = "skylight,cal150"
	if info.Codename != "athena" || info.Detection != asp.PlatformConfirmed {
		t.Fatalf("returned metadata aliases mutable evidence: %+v", info)
	}

	micPCM := bytes.Repeat([]byte{0x01, 0xff, 0x55, 0x80}, nativeMicBytes/4)
	refPCM := bytes.Repeat([]byte{0x12, 0x34}, nativeRefBytes/2)
	originalMic, originalRef := bytes.Clone(micPCM), bytes.Clone(refPCM)
	native := nativeFrame{mics: micPCM, ref: refPCM, fullBits: 23}
	before := native.input()
	if err := before.Validate(); err != nil {
		t.Fatalf("native fixture invalid: %v", err)
	}
	plain, err := (asp.CFG2{Config: before.Config}).MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	packet, err := (asp.CFG2{Config: before.Config, PlatformInfo: info}).MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(plain, packet) || len(packet) != len(plain)+8+4+len("athena") {
		t.Fatalf("unexpected PLATFORM_INFO TLV size: core=%d metadata=%d", len(plain), len(packet))
	}
	for _, secret := range []string{"SYNTHETIC_SERIAL_98765", "SECRET_PRODUCT", "amzn,athena-pvt", "mediatek,mt8183", "BoardDoubt"} {
		if bytes.Contains(packet, []byte(secret)) {
			t.Fatalf("private evidence %q appeared in CFG2", secret)
		}
	}
	decoded, err := asp.ParseCFG2(packet)
	if err != nil {
		t.Fatal(err)
	}
	if decoded.PlatformInfo == nil || *decoded.PlatformInfo != *info || decoded.MicFactoryGain != nil {
		t.Fatalf("PLATFORM_INFO round trip mismatch: %+v", decoded)
	}
	if decoded.Config != before.Config {
		t.Fatalf("metadata altered MIC/REF descriptor: %+v != %+v", decoded.Config, before.Config)
	}
	after := native.input()
	if before.Config != after.Config || !reflect.DeepEqual(before.Mic, after.Mic) ||
		!reflect.DeepEqual(before.Ref, after.Ref) || !bytes.Equal(native.mics, originalMic) ||
		!bytes.Equal(native.ref, originalRef) {
		t.Fatal("platform metadata builder changed native input or PCM")
	}
}
