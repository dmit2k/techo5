package mic

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/HuskerMinion/techo5/echod/internal/lib/spectrum"
	"math/rand"
	"os"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/HuskerMinion/techo5/echod/internal/asp"
	"github.com/HuskerMinion/techo5/echod/internal/asp/synthetic"
	"github.com/HuskerMinion/techo5/echod/internal/config"
)

func aspShape(m, r int) asp.Config {
	c := asp.Config{Flags: asp.Interleaved, Mic: asp.Stream{Channels: uint16(m), Frames: 1920, Rate: 96000, Format: asp.S32, FullBits: 23, ValidBits: 24}}
	if r > 0 {
		c.Flags |= asp.Synchronized | asp.HardwareRef
		c.Ref = asp.Stream{Channels: uint16(r), Frames: 1920, Rate: 96000, Format: asp.S16, FullBits: 15, ValidBits: 16}
	}
	return c
}
func aspInput(c asp.Config) asp.Input {
	m, r, _ := c.Lengths()
	i := asp.Input{Config: c, Mic: make([]byte, m), Ref: make([]byte, r)}
	for at := range i.Mic {
		i.Mic[at] = byte(at*37 + 11)
	}
	for at := range i.Ref {
		i.Ref[at] = byte(at*19 + 7)
	}
	return i
}
func configuredChild(t testing.TB, mode string, c asp.Config, env ...string) *externalASP {
	t.Helper()
	cmd := childCommand(mode)
	cmd.Env = append(cmd.Env, env...)
	f, e := startTestASP(context.Background(), cmd, c)
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() {
		f.Close()
		select {
		case <-f.done:
		case <-time.After(time.Second):
			t.Error("unreaped ASP2 child")
		}
	})
	return f
}
func TestASP2PipeMatrixAndImmutability(t *testing.T) {
	for m := 1; m <= 8; m++ {
		for r := 0; r <= 2; r++ {
			t.Run(fmt.Sprintf("%d+%d", m, r), func(t *testing.T) {
				f := configuredChild(t, "normal", aspShape(m, r))
				i := aspInput(aspShape(m, r))
				beforeM, beforeR := bytes.Clone(i.Mic), bytes.Clone(i.Ref)
				for tick := 0; tick < 2; tick++ {
					got, e := f.exchange(i)
					want := fixturePCM(nativeFrame{i.Mic, i.Ref, uint32(i.Config.Mic.FullBits)}, uint64(tick))
					if e != nil || got.Start != uint64(tick*320) || !slices.Equal(got.PCM, want) {
						t.Fatal(got.Start, e)
					}
				}
				if !bytes.Equal(beforeM, i.Mic) || !bytes.Equal(beforeR, i.Ref) {
					t.Fatal("mutated capture slots")
				}
				changed := i
				changed.Config.Mic.FullBits = 31
				changed.Config.Mic.ValidBits = 32
				if _, e := f.exchange(changed); e == nil {
					t.Fatal("renegotiated in place")
				}
			})
		}
	}
	for _, m := range []int{9, 12, 16} {
		f := configuredChild(t, "normal", aspShape(m, 8))
		if _, e := f.exchange(aspInput(aspShape(m, 8))); e != nil {
			t.Fatal(m, e)
		}
	}
	for _, shape := range []asp.Config{
		{Flags: asp.Interleaved, Mic: asp.Stream{Channels: 8, Frames: 320, Rate: 16000, Format: asp.S16, FullBits: 15, ValidBits: 16}},
		{Flags: asp.Interleaved | asp.Synchronized | asp.HardwareRef, Mic: asp.Stream{Channels: 7, Frames: 960, Rate: 48000, Format: asp.S32, FullBits: 31, ValidBits: 24, Alignment: asp.High}, Ref: asp.Stream{Channels: 2, Frames: 320, Rate: 16000, Format: asp.S16, FullBits: 15, ValidBits: 16}},
	} {
		f := configuredChild(t, "normal", shape)
		if _, e := f.exchange(aspInput(shape)); e != nil {
			t.Fatal(e)
		}
	}
	f := configuredChild(t, "no-ref-reject", aspShape(8, 0))
	if _, e := f.exchange(aspInput(aspShape(8, 0))); !errors.Is(e, asp.ErrStatus) {
		t.Fatal("no explicit no-ref ACK", e)
	}
	f = configuredChild(t, "factory", aspShape(7, 2))
	_, e := f.exchange(aspInput(aspShape(7, 2)))
	var terminal *asp.Terminal
	if !errors.As(e, &terminal) || terminal.Status.Code != 522 {
		t.Fatal("factory topology not rejected", e)
	}
}
func markerInput(tick int) ([]byte, nativeFrame) {
	raw := make([]byte, FrameSamples*Channels*3)
	for f := 0; f < FrameSamples; f++ {
		for ch := 0; ch < Mics; ch++ {
			putS24LE3(raw[(f*Channels+ch)*3:], int32(synthetic.Marker(uint64(tick*FrameSamples+f)))<<8)
		}
	}
	n := nativeFrame{mics: make([]byte, nativeMicBytes), ref: make([]byte, nativeRefBytes), fullBits: 23}
	for f := 0; f < 1920; f++ {
		for ch := 0; ch < 4; ch++ {
			binary.LittleEndian.PutUint32(n.mics[(f*4+ch)*4:], uint32(int32(synthetic.Marker(uint64(tick*320+f/6)))<<8))
		}
	}
	return raw, n
}
func checkMarker(t testing.TB, frame []int16, start uint64) {
	t.Helper()
	if len(frame) != 320 {
		t.Fatal("frame size", len(frame))
	}
	for j, v := range frame {
		if v != synthetic.Marker(start+uint64(j)) {
			t.Fatalf("sample discontinuity at %d: got %d want %d", start+uint64(j), v, synthetic.Marker(start+uint64(j)))
		}
	}
}

// Policy A's exact cross-processor timeline remains in inputs/POLICY_A_asp2_test.go.
// Handoffs now permit one source overlap/skip, never a missing publication or delayed Built-in.
func TestASP2PublishedTimeline(t *testing.T) {
	cases := []string{"factory", "manual-preflight", "manual-active"}
	for at := 0; at <= 8; at++ {
		for _, fault := range []string{"crash", "output-hang", "short", "short-open", "sequence", "origin", "origin-repeat", "late-not-ready", "reference", "fatal", "bad-status", "large-status", "bad-kind", "bad-code", "truncated-status"} {
			cases = append(cases, fmt.Sprintf("%s:%d", fault, at))
		}
	}
	for _, which := range cases {
		t.Run(which, func(t *testing.T) {
			mode := "factory"
			faultAt := -1
			if parts := strings.Split(which, ":"); len(parts) == 2 {
				faultAt, _ = strconv.Atoi(parts[1])
				if parts[0] != "reference" {
					mode = "factory:" + parts[0] + ":" + parts[1]
				}
			}
			s := controlSource(t, mode, config.ASPExternal, nativeASPConfig(23))
			s.mixer = Center{}
			live, stop := s.Listen("MWW")
			defer stop()
			stt, stopSTT := s.Listen("STT")
			defer stopSTT()
			oldSpectrum := spectrum.Mic
			defer func() { spectrum.Mic = oldSpectrum }()
			actual, expected := spectrum.New(Rate, 512), spectrum.New(Rate, 512)
			spectrum.Mic = actual
			s.SetASP(config.ASPExternal)
			waitASPChild(t, s)
			child := s.asp
			var history []int16
			var previousEnd uint64
			overlap, skip := 0, 0
			for tick := 0; tick < 65; tick++ {
				if which == "manual-preflight" && tick == 2 || which == "manual-active" && tick == 7 {
					s.SetASP(config.ASPBuiltin)
				}
				raw, n := markerInput(tick)
				if strings.HasPrefix(which, "reference:") && tick == faultAt {
					n.ref = nil
				}
				now := time.Now()
				actual.Bands(24, now)
				expected.Bands(24, now)
				s.broadcastNative(raw, n)
				wantExternal := tick >= 4 && (faultAt < 0 || which == "late-not-ready:0" || tick < faultAt)
				if which == "manual-preflight" && tick >= 2 || which == "manual-active" && tick >= 7 {
					wantExternal = false
				}
				if (s.ASPStatus().Active == config.ASPExternal) != wantExternal {
					t.Fatal("unexpected early fallback or activation", tick, s.ASPStatus())
				}
				var got []int16
				select {
				case got = <-live:
				default:
					t.Fatal("missing capture-tick publication", tick)
				}
				start := uint64(tick * 320)
				if s.ASPStatus().Active == config.ASPExternal {
					start -= 320
				}
				checkMarker(t, got, start)
				if tick > 0 {
					if start < previousEnd {
						if previousEnd-start != 320 {
							t.Fatal("unbounded overlap")
						}
						overlap++
					}
					if start > previousEnd {
						if start-previousEnd != 320 {
							t.Fatal("unbounded skip")
						}
						skip++
					}
				}
				previousEnd = start + 320
				other := <-stt
				if !slices.Equal(got, other) || !slices.Equal(s.Recent(20*time.Millisecond), got) {
					t.Fatal("split history/MWW/STT publication")
				}
				for _, ch := range []<-chan []int16{live, stt} {
					select {
					case <-ch:
						t.Fatal("two publications per tick")
					default:
					}
				}
				expected.Push(got)
				a, _ := actual.Bands(24, now.Add(time.Millisecond))
				b, _ := expected.Bands(24, now.Add(time.Millisecond))
				if !slices.Equal(a, b) {
					t.Fatal("spectrum differs from selected PCM", tick)
				}
				history = append(history, got...)
				if tick == 6 || tick == 64 {
					tailCh, h, tailStop := s.ListenHistory("mid-turn", 2*time.Second)
					if !slices.Equal(history, h) {
						t.Fatal("history differs from actual publication sequence")
					}
					select {
					case <-tailCh:
						t.Fatal("duplicate history/live tail")
					default:
					}
					tailStop()
				}
			}
			if overlap > 1 || skip > 1 || s.Dropped() != 0 {
				t.Fatal("unbounded boundary/drop", overlap, skip, s.Dropped())
			}
			if which == "factory" && (overlap != 1 || skip != 0 || s.ASPStatus().Active != config.ASPExternal || child.sequence != 65) {
				t.Fatal("factory never remained active", s.ASPStatus(), child.sequence)
			}
			if strings.Contains(which, ":") && s.ASPStatus().Selected != config.ASPExternal {
				t.Fatal("saved selection erased")
			}
			if faultAt >= 0 && which != "late-not-ready:0" {
				if s.ASPStatus().Active != config.ASPBuiltin || !child.broken.Load() {
					t.Fatal("fault retained External")
				}
				select {
				case <-child.done:
				case <-time.After(time.Second):
					t.Fatal("fault child not reaped")
				}
			}
			if which == "factory" {
				t.Logf("ticks=65 publications=65 samples=%d overlap_samples=320 skip_samples=0 built_in_offset_samples=0 listener_drops=%d", len(history), s.Dropped())
			}
		})
	}
}

func TestASP2LongPublishedAndSeededSwitches(t *testing.T) {
	ticks := 300
	if os.Getenv("TECHO5_ASP_LONG") == "1" {
		ticks = 100000
	}
	s := controlSource(t, "factory", config.ASPExternal, nativeASPConfig(23))
	s.mixer = Center{}
	s.SetASP(config.ASPExternal)
	waitASPChild(t, s)
	child := s.asp
	ch, stop := s.Listen("long MWW")
	defer stop()
	stt, stopSTT := s.Listen("long STT")
	defer stopSTT()
	before := time.Now()
	samples := 0
	for tick := 0; tick < ticks; tick++ {
		raw, n := markerInput(tick)
		s.broadcastNative(raw, n)
		var frame []int16
		select {
		case frame = <-ch:
		default:
			t.Fatal("missing publication", tick)
		}
		start := uint64(tick * 320)
		if tick >= 4 {
			start -= 320
		}
		checkMarker(t, frame, start)
		if !slices.Equal(frame, <-stt) {
			t.Fatal("split consumers")
		}
		samples += len(frame)
		if tick >= 4 && s.ASPStatus().Active != config.ASPExternal {
			t.Fatal("long run silently fell back", tick, s.ASPStatus())
		}
	}
	st := s.ASPStatus()
	if child.sequence != uint64(ticks) || !st.Confirmed || st.Active != config.ASPExternal || samples != ticks*320 || s.Dropped() != 0 || s.aspAttempt.next != uint64(ticks-1)*320 {
		t.Fatal("External continuity not proved", child.sequence, st, samples, s.aspAttempt.next)
	}
	t.Logf("actual_external_exchanges=%d active=External published_samples=%d gaps=0 steady_external_duplicates=0 steady_external_skips=0 activation_overlap_samples=320 host_retained_pcm_samples=0 listener_drops=%d elapsed_seconds=%.3f", child.sequence, samples, s.Dropped(), time.Since(before).Seconds())
	// Return to current stock now and fifty ticks later; no DSP tail is drained.
	s.SetASP(config.ASPBuiltin)
	for tick := ticks; tick < ticks+55; tick++ {
		raw, n := markerInput(tick)
		s.broadcastNative(raw, n)
		frame := <-ch
		checkMarker(t, frame, uint64(tick*320))
		if !slices.Equal(frame, <-stt) {
			t.Fatal("fallback consumers")
		}
	}
	select {
	case <-ch:
		t.Fatal("EOF/tail extra publication")
	default:
	}
}

func TestASP2SeededReactivationNoLagAccumulation(t *testing.T) {
	s := controlSource(t, "factory:fatal:6", config.ASPExternal, nativeASPConfig(23))
	s.mixer = Center{}
	ch, stop := s.Listen("seed MWW")
	defer stop()
	stt, stopSTT := s.Listen("seed STT")
	defer stopSTT()
	rng := rand.New(rand.NewSource(0xA532))
	tick := 37
	activeRuns, fatalRuns, early := 0, 0, 0
	feed := func(wantExternal bool) {
		raw, n := markerInput(tick)
		s.broadcastNative(raw, n)
		start := uint64(tick * 320)
		if wantExternal {
			start -= 320
		}
		var frame []int16
		select {
		case frame = <-ch:
		default:
			t.Fatal("switch pause", tick)
		}
		checkMarker(t, frame, start)
		if !slices.Equal(frame, <-stt) || !slices.Equal(frame, s.Recent(20*time.Millisecond)) {
			t.Fatal("switch split publication")
		}
		if (s.ASPStatus().Active == config.ASPExternal) != wantExternal {
			t.Fatal("wrong active source", tick, s.ASPStatus())
		}
		for _, c := range []<-chan []int16{ch, stt} {
			select {
			case <-c:
				t.Fatal("extra publication")
			default:
			}
		}
		tick++
	}
	for cycle := 0; cycle < 120; cycle++ {
		s.SetASP(config.ASPBuiltin)
		feed(false)
		s.SetASP(config.ASPExternal)
		waitASPChild(t, s)
		child := s.asp
		count := rng.Intn(12)
		for local := 0; local < count; local++ {
			feed(local >= 4 && local < 6)
		}
		if count >= 5 {
			activeRuns++
		}
		if count >= 7 {
			fatalRuns++
		} else if count < 4 {
			early++
		}
		if child.sequence != uint64(min(count, 6)) {
			t.Fatal("session origin reused or early fallback", cycle, count, child.sequence, s.ASPStatus())
		}
		s.SetASP(config.ASPBuiltin)
		feed(false)
		feed(false)
		select {
		case <-child.done:
		case <-time.After(time.Second):
			t.Fatal("switch child not reaped")
		}
	}
	for i := 0; i < 55; i++ {
		feed(false)
	}
	if activeRuns < 20 || fatalRuns < 20 || early < 20 || s.Dropped() != 0 {
		t.Fatal(activeRuns, fatalRuns, early, s.Dropped())
	}
	t.Logf("seed=0xA532 cycles=120 active_runs=%d fatal_runs=%d early_preflight_runs=%d final_tick=%d lag_accumulation_samples=0 missing_publications=0 listener_drops=0", activeRuns, fatalRuns, early, tick)
}

func TestASP2FeedbackStateFatalAndPersistence(t *testing.T) {
	for _, saved := range []config.ASPMode{config.ASPBuiltin, config.ASPExternal} {
		for _, code := range []int{0, 200, 201, 301, 422, 503, 522} {
			t.Run(fmt.Sprintf("%s/%d", saved, code), func(t *testing.T) {
				s := controlSource(t, "normal", saved, nativeASPConfig(23))
				s.aspStart = func(ctx context.Context) (*externalASP, error) {
					cmd := childCommand("normal")
					cmd.Env = append(cmd.Env, "TECHO5_ASP_ACK_KIND=1", "TECHO5_ASP_ACK_CODE="+strconv.Itoa(code), "TECHO5_ASP_ACK_MSG=Bounded helper state")
					return startTestASP(ctx, cmd, nativeASPConfig(23))
				}
				s.SetASP(config.ASPExternal)
				waitASPChild(t, s)
				for tick := 0; tick < 4; tick++ {
					s.broadcastNative(rawFrame(tick, false), nativePattern(23, tick))
				}
				st := s.ASPStatus()
				if st.Debug != fmt.Sprintf("%03d: Bounded helper state", code) {
					t.Fatal(st)
				}
				if code >= 500 {
					if st.Active != config.ASPBuiltin || st.Selected != saved || config.Get().Microphone.ASP != saved {
						t.Fatal(st)
					}
				} else if st.Active != config.ASPExternal || !st.Confirmed {
					t.Fatal(st)
				}
			})
		}
	}
	for _, typ := range []int{asp.Data, asp.NotReady, asp.Fatal} {
		s := controlSource(t, "normal", config.ASPExternal, nativeASPConfig(23))
		s.aspStart = func(ctx context.Context) (*externalASP, error) {
			cmd := childCommand("normal")
			cmd.Env = append(cmd.Env, "TECHO5_ASP_REPLY_KIND=2", "TECHO5_ASP_REPLY_CODE=503", "TECHO5_ASP_REPLY_MSG=ASP processing failed", "TECHO5_ASP_REPLY_ON=4", "TECHO5_ASP_REPLY_TYPE="+strconv.Itoa(typ))
			return startTestASP(ctx, cmd)
		}
		preflightASP(t, s, "normal")
		s.broadcastNative(rawFrame(4, false), nativePattern(31, 4))
		st := s.ASPStatus()
		if st.Debug != "503: ASP processing failed" || st.Active != config.ASPBuiltin || st.Selected != config.ASPExternal {
			t.Fatal(st)
		}
		s.broadcastNative(rawFrame(5, false), nativePattern(31, 5))
		if s.ASPStatus().Debug != st.Debug {
			t.Fatal("exit overwrote complete fatal")
		}
	}
}
