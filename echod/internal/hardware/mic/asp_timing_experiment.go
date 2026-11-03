package mic

// Temporary hardware timing instrumentation. Kept only on the experimental branch.
// Records timing/return status, NEVER microphone samples or PCM payloads.

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

type aspTimingEvent struct {
	at      time.Time
	kind    string
	elapsed time.Duration
	reason  string
}

var aspTimingEvents = make(chan aspTimingEvent, 4096)
var aspTimingDropped atomic.Uint64

func init() { go aspTimingWriter() }

func aspTimingRecord(kind string, elapsed time.Duration, err error) {
	reason := "ok"
	if err != nil {
		reason = strings.ReplaceAll(err.Error(), ",", ";")
		reason = strings.ReplaceAll(reason, "\n", " ")
	}
	select {
	case aspTimingEvents <- aspTimingEvent{time.Now(), kind, elapsed, reason}:
	default: // Measurement must never block microphone capture.
		aspTimingDropped.Add(1)
	}
}

func aspTimingWriter() {
	// /data survives rootfs changes and reboot on both tested devices.
	const dir = "/data/techo5-linux/asp-pcm-timing"
	if err := os.MkdirAll(dir, 0700); err != nil {
		return
	}
	bootID, err := os.ReadFile("/proc/sys/kernel/random/boot_id")
	if err != nil {
		return
	}
	id := strings.TrimSpace(string(bootID))
	path := filepath.Join(dir, fmt.Sprintf("timing-%s-pid%d.csv", id, os.Getpid()))
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return
	}
	defer f.Close()
	stat, _ := f.Stat()
	w := bufio.NewWriterSize(f, 64*1024)
	if stat == nil || stat.Size() == 0 {
		fmt.Fprintln(w, "wall_utc,boot_uptime_ms,kind,exchange_us,result,total,over20,over40,over80,over200,over1000,max_us")
	}

	var atBoot int64
	bootStamp := time.Now()
	if b, e := os.ReadFile("/proc/uptime"); e == nil {
		parts := strings.Fields(string(b))
		if len(parts) != 0 {
			if up, e := strconv.ParseFloat(parts[0], 64); e == nil {
				atBoot = int64(up * 1000)
			}
		}
	}
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	var fullUntil time.Time
	var minute time.Time
	var total, over20, over40, over80, over200, over1000 int
	var maxUS int64
	var slowLogged int

	row := func(e aspTimingEvent, count, a, b, c, d, g int, max int64) {
		up := atBoot + e.at.Sub(bootStamp).Milliseconds()
		fmt.Fprintf(w, "%s,%d,%s,%d,%s,%d,%d,%d,%d,%d,%d,%d\n",
			e.at.UTC().Format(time.RFC3339Nano), up, e.kind,
			e.elapsed.Microseconds(), e.reason, count, a, b, c, d, g, max)
	}
	summary := func(now time.Time) {
		queuedDrops := aspTimingDropped.Swap(0)
		if total == 0 && queuedDrops == 0 {
			return
		}
		e := aspTimingEvent{at: now, kind: "summary", reason: fmt.Sprintf("logger_drops=%d", queuedDrops)}
		row(e, total, over20, over40, over80, over200, over1000, maxUS)
		total, over20, over40, over80, over200, over1000, maxUS = 0, 0, 0, 0, 0, 0, 0
		slowLogged = 0
	}
	for {
		select {
		case e := <-aspTimingEvents:
			if minute.IsZero() {
				minute = e.at
			}
			if e.at.Sub(minute) >= time.Minute {
				summary(e.at)
				minute = e.at
			}
			if e.kind == "launch_begin" {
				fullUntil = e.at.Add(5 * time.Minute)
			}
			if e.kind == "exchange" {
				total++
				us := e.elapsed.Microseconds()
				if us > maxUS {
					maxUS = us
				}
				if us > 20000 {
					over20++
				}
				if us > 40000 {
					over40++
				}
				if us > 80000 {
					over80++
				}
				if us > 200000 {
					over200++
				}
				if us > 1000000 {
					over1000++
				}
				// All exchanges for the first five minutes of each session, then
				// at most 20 late samples per minute plus one aggregate summary.
				if e.at.Before(fullUntil) || (us > 20000 && slowLogged < 20) || e.reason != "ok" {
					row(e, 0, 0, 0, 0, 0, 0, 0)
					if e.at.After(fullUntil) && us > 20000 {
						slowLogged++
					}
				}
			} else {
				// Launch, READY, activation, fallback, native geometry and XRUN.
				row(e, 0, 0, 0, 0, 0, 0, 0)
				_ = w.Flush()
				if e.kind == "fallback" || e.kind == "native_fallback" {
					_ = f.Sync()
				}
			}
		case <-ticker.C:
			_ = w.Flush()
			if time.Since(bootStamp) >= time.Minute {
				// Periodic durability without fsync on the capture goroutine.
				_ = f.Sync()
			}
		}
	}
}
