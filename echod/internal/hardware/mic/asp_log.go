package mic

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"
)

const aspLogCapacity = 64

type aspLogRecord struct {
	level                       slog.Level
	code, kind, origin, message string
	request                     uint64
	current                     *atomic.Uint64
}
type aspLogStats struct {
	Accepted, Logged, Dropped, Stale uint64
	Depth, MaxDepth                  int
}
type aspLogDispatch struct {
	mu          sync.Mutex
	queue       [aspLogCapacity]aspLogRecord
	head, count int
	critical    *aspLogRecord
	stats       aspLogStats
	reported    uint64
	wake        chan struct{}
	stop        chan struct{}
	done        chan struct{}
	write       func(aspLogRecord)
}

func newASPLogDispatch(write func(aspLogRecord)) *aspLogDispatch {
	d := &aspLogDispatch{wake: make(chan struct{}, 1), stop: make(chan struct{}), done: make(chan struct{}), write: write}
	go d.run()
	return d
}
func standardASPLog(r aspLogRecord) {
	slog.LogAttrs(context.Background(), r.level, "ASP",
		slog.String("asp_status_code", r.code), slog.String("asp_kind", r.kind),
		slog.String("asp_origin", r.origin), slog.String("asp_message", r.message),
		slog.Uint64("asp_request", r.request))
}

// One shared consumer for the daemon, not one goroutine per Source, frame or event.
var sharedASPLog = sync.OnceValue(func() *aspLogDispatch { return newASPLogDispatch(standardASPLog) })

func (d *aspLogDispatch) enqueue(r aspLogRecord) {
	d.mu.Lock()
	d.stats.Accepted++
	if r.level >= slog.LevelError {
		// Retain the fatal record independently of saturated informational slots.
		if d.count == aspLogCapacity {
			d.stats.Dropped += uint64(d.count)
			clear(d.queue[:])
			d.head, d.count = 0, 0
		}
		if d.critical != nil {
			d.stats.Dropped++
		}
		d.critical = &r
	} else if d.count == aspLogCapacity {
		d.stats.Dropped++
	} else {
		d.queue[(d.head+d.count)%aspLogCapacity] = r
		d.count++
		d.stats.MaxDepth = max(d.stats.MaxDepth, d.count)
	}
	d.mu.Unlock()
	select {
	case d.wake <- struct{}{}:
	default:
	}
}
func (d *aspLogDispatch) snapshot() aspLogStats {
	d.mu.Lock()
	defer d.mu.Unlock()
	s := d.stats
	s.Depth = d.count
	if d.critical != nil {
		s.Depth++
	}
	return s
}
func (d *aspLogDispatch) next() (aspLogRecord, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.critical != nil {
		r := *d.critical
		d.critical = nil
		return r, true
	}
	if d.stats.Dropped != d.reported {
		d.reported = d.stats.Dropped
		return aspLogRecord{level: slog.LevelWarn, code: "340", kind: "STATE", origin: "echod", message: "ASP diagnostic queue overflow; lower-priority records dropped"}, true
	}
	if d.count == 0 {
		return aspLogRecord{}, false
	}
	r := d.queue[d.head]
	d.queue[d.head] = aspLogRecord{}
	d.head = (d.head + 1) % aspLogCapacity
	d.count--
	return r, true
}
func (d *aspLogDispatch) run() {
	defer close(d.done)
	for {
		select {
		case <-d.stop:
			return
		case <-d.wake:
		}
		for {
			select {
			case <-d.stop:
				return
			default:
			}
			r, ok := d.next()
			if !ok {
				break
			}
			if r.current != nil && r.current.Load() != r.request {
				d.mu.Lock()
				d.stats.Stale++
				d.mu.Unlock()
				continue
			}
			// No capture, ASP or dispatcher lock is held across a possibly blocked slog Handler.
			d.write(r)
			d.mu.Lock()
			d.stats.Logged++
			d.mu.Unlock()
		}
	}
}
