// Package asp defines the bounded ASP2 binary contract shared by capture and helpers.
// It describes captured slots, never DSP geometry, device discovery or calibration.
package asp

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	MaxPayload   = 256 << 10
	MaxMessage   = 160
	Samples      = 320
	NoSample     = math.MaxUint64
	S16          = 1
	S32          = 2
	Low          = 0
	High         = 1
	Interleaved  = 1
	Synchronized = 2
	HardwareRef  = 4
	NoReference  = 1
	Accept       = 1
	Reject       = 2
	Data         = 1
	NotReady     = 2
	Fatal        = 3
	State        = 1
	Event        = 2
)

var (
	ErrConfig = errors.New("invalid external ASP configuration")
	ErrStatus = errors.New("invalid external ASP status")
	ErrPCM    = errors.New("invalid external ASP PCM")
)

type Stream struct {
	Channels, Frames                       uint16
	Rate                                   uint32
	Format, FullBits, Alignment, ValidBits uint8
}
type Config struct {
	Flags    uint32
	Mic, Ref Stream
}
type Input struct {
	Config   Config
	Mic, Ref []byte
}
type Status struct {
	Kind    uint8
	Code    uint16
	Message string
}

func (s Status) Present() bool  { return s.Kind != 0 }
func (s Status) Terminal() bool { return s.Code >= 500 && s.Code <= 599 }
func (s Status) Warning() bool  { return s.Code >= 300 && s.Code < 500 }
func (s Status) Text(def string) string {
	if s.Message != "" {
		def = s.Message
	}
	return fmt.Sprintf("%03d: %s", s.Code, def)
}

// Terminal is an already validated complete helper status, preserved through Kill/Wait.
type Terminal struct {
	Status        Status
	Configuration bool
}

func (e *Terminal) Error() string { return "external ASP rejected processing" }

func (s Stream) bytes(maxChannels uint16) (uint32, error) {
	if s.Channels < 1 || s.Channels > maxChannels || s.Rate < 8000 || s.Rate > 192000 || s.Rate%50 != 0 || uint32(s.Frames) != s.Rate/50 {
		return 0, ErrConfig
	}
	width := uint8(0)
	switch s.Format {
	case S16:
		width = 16
	case S32:
		width = 32
	default:
		return 0, ErrConfig
	}
	if s.ValidBits < 2 || s.ValidBits > width || s.Alignment > High {
		return 0, ErrConfig
	}
	scale := s.ValidBits - 1
	if s.Alignment == High {
		scale = width - 1
	}
	if s.FullBits != scale {
		return 0, ErrConfig
	}
	n := uint64(s.Frames) * uint64(s.Channels) * uint64(width/8)
	if n > MaxPayload {
		return 0, ErrConfig
	}
	return uint32(n), nil
}
func (c Config) Lengths() (uint32, uint32, error) {
	m, e := c.Mic.bytes(16)
	if e != nil {
		return 0, 0, e
	}
	if c.Ref.Channels == 0 {
		if c.Ref != (Stream{}) || c.Flags != Interleaved {
			return 0, 0, ErrConfig
		}
		return m, 0, nil
	}
	if c.Flags != Interleaved|Synchronized|HardwareRef {
		return 0, 0, ErrConfig
	}
	r, e := c.Ref.bytes(8)
	if e != nil || uint64(m)+uint64(r) > MaxPayload {
		return 0, 0, ErrConfig
	}
	return m, r, nil
}
func (i Input) Validate() error {
	m, r, e := i.Config.Lengths()
	if e != nil {
		return e
	}
	if uint64(len(i.Mic)) != uint64(m) || uint64(len(i.Ref)) != uint64(r) {
		return ErrConfig
	}
	return nil
}

// Unsafe but framed text is ignored as a whole, never partially echoed. Invalid UTF-8 is an error.
// This conservative plain-text policy rejects controls/markup/paths/IP literals, not just ANSI.
func cleanText(s string) (string, error) {
	if len(s) > MaxMessage || !utf8.ValidString(s) {
		return "", ErrStatus
	}
	safe := true
	for _, r := range s {
		if !(unicode.IsLetter(r) || unicode.IsNumber(r) || r == ' ' || strings.ContainsRune(".,:;!?()-+%", r)) {
			safe = false
			break
		}
	}
	for _, word := range strings.Fields(s) {
		if net.ParseIP(strings.Trim(word, "(),;[]")) != nil {
			safe = false
		}
	}
	if !safe {
		return "", nil
	}
	return strings.TrimSpace(s), nil
}
func (s Status) Validate() (Status, error) {
	if s.Kind == 0 {
		if s.Code != 0 || s.Message != "" {
			return Status{}, ErrStatus
		}
		return s, nil
	}
	if s.Kind != State && s.Kind != Event {
		return Status{}, ErrStatus
	}
	// Zero is reserved exclusively for a wholly absent status (Kind=0).
	// The host/external halves of the nonzero range are advisory, not enforced.
	if s.Code < 200 || s.Code > 599 {
		return Status{}, ErrStatus
	}
	if s.Kind == Event && s.Code >= 300 && !s.Terminal() {
		return Status{}, ErrStatus
	}
	t, e := cleanText(s.Message)
	s.Message = t
	return s, e
}
func readStatus(r io.Reader, kind uint8, code, length uint16) (Status, error) {
	if length > MaxMessage {
		return Status{}, ErrStatus
	}
	var b [MaxMessage]byte
	if _, e := io.ReadFull(r, b[:length]); e != nil {
		return Status{}, ErrStatus
	}
	return (Status{kind, code, string(b[:length])}).Validate()
}
func WriteAll(w io.Writer, b []byte) error {
	for len(b) > 0 {
		n, e := w.Write(b)
		if e != nil {
			return e
		}
		if n <= 0 || n > len(b) {
			return io.ErrShortWrite
		}
		b = b[n:]
	}
	return nil
}

type ConfigResult struct {
	Type   uint8
	Flags  uint16
	Status Status
}

func ReadResult(r io.Reader) (ConfigResult, error) {
	var h [12]byte
	var a ConfigResult
	if n, e := io.ReadFull(r, h[:]); e != nil {
		if n > 0 {
			return a, ErrStatus
		}
		return a, e
	}
	if string(h[:4]) != "ACK2" {
		return a, ErrStatus
	}
	a.Type, a.Flags = h[4], binary.LittleEndian.Uint16(h[10:])
	s, e := readStatus(r, h[5], binary.LittleEndian.Uint16(h[6:]), binary.LittleEndian.Uint16(h[8:]))
	a.Status = s
	if e != nil {
		return a, e
	}
	if s.Terminal() {
		return a, &Terminal{s, true}
	}
	if a.Type != Accept && a.Type != Reject || a.Flags & ^uint16(NoReference) != 0 {
		return a, ErrStatus
	}
	if a.Type == Reject {
		return a, ErrStatus // A reject must carry a typed terminal status.
	}
	return a, nil
}
func WriteResult(w io.Writer, a ConfigResult) error {
	if len(a.Status.Message) > MaxMessage {
		return ErrStatus
	}
	var h [12]byte
	copy(h[:], "ACK2")
	h[4], h[5] = a.Type, a.Status.Kind
	binary.LittleEndian.PutUint16(h[6:], a.Status.Code)
	binary.LittleEndian.PutUint16(h[8:], uint16(len(a.Status.Message)))
	binary.LittleEndian.PutUint16(h[10:], a.Flags)
	if e := WriteAll(w, h[:]); e != nil {
		return e
	}
	return WriteAll(w, []byte(a.Status.Message))
}

type Reply struct {
	Type            uint8
	Sequence, Start uint64
	PCM             []int16
	Status          Status
	ConfigStatus    Status // host-local first-exchange metadata, never serialized
}

func ReadReply(r io.Reader) (Reply, error) {
	var h [32]byte
	var a Reply
	if n, e := io.ReadFull(r, h[:]); e != nil {
		if n > 0 {
			return a, ErrStatus
		}
		return a, e
	}
	if string(h[:4]) != "RSP2" || binary.LittleEndian.Uint16(h[10:]) != 0 {
		return a, ErrStatus
	}
	a.Type, a.Sequence, a.Start = h[4], binary.LittleEndian.Uint64(h[12:]), binary.LittleEndian.Uint64(h[20:])
	s, e := readStatus(r, h[5], binary.LittleEndian.Uint16(h[6:]), binary.LittleEndian.Uint16(h[8:]))
	a.Status = s
	if e != nil {
		return a, e
	}
	// A complete fatal envelope preempts attached PCM, even if the worker never sends its bytes.
	if s.Terminal() {
		return a, &Terminal{s, false}
	}
	if a.Type == Fatal {
		return a, ErrStatus // A fatal response must carry a typed terminal status.
	}
	size := binary.LittleEndian.Uint32(h[28:])
	if a.Type == NotReady {
		if size != 0 || a.Start != NoSample {
			return a, ErrPCM
		}
		return a, nil
	}
	if a.Type != Data || size != Samples*2 || a.Start == NoSample {
		return a, ErrPCM
	}
	var b [Samples * 2]byte
	if _, e := io.ReadFull(r, b[:]); e != nil {
		return a, e
	}
	a.PCM = make([]int16, Samples)
	for i := range a.PCM {
		a.PCM[i] = int16(binary.LittleEndian.Uint16(b[2*i:]))
	}
	return a, nil
}
func WriteReply(w io.Writer, a Reply) error {
	if len(a.Status.Message) > MaxMessage {
		return ErrStatus
	}
	if len(a.PCM) > Samples {
		return ErrPCM
	}
	var h [32]byte
	copy(h[:], "RSP2")
	h[4], h[5] = a.Type, a.Status.Kind
	binary.LittleEndian.PutUint16(h[6:], a.Status.Code)
	binary.LittleEndian.PutUint16(h[8:], uint16(len(a.Status.Message)))
	binary.LittleEndian.PutUint64(h[12:], a.Sequence)
	binary.LittleEndian.PutUint64(h[20:], a.Start)
	binary.LittleEndian.PutUint32(h[28:], uint32(len(a.PCM)*2))
	if e := WriteAll(w, h[:]); e != nil {
		return e
	}
	if e := WriteAll(w, []byte(a.Status.Message)); e != nil {
		return e
	}
	var b [Samples * 2]byte
	if len(a.PCM) > Samples {
		return ErrPCM
	}
	for i, v := range a.PCM {
		binary.LittleEndian.PutUint16(b[2*i:], uint16(v))
	}
	return WriteAll(w, b[:len(a.PCM)*2])
}
func WriteRequest(w io.Writer, seq uint64, i Input) error {
	if e := i.Validate(); e != nil {
		return e
	}
	var h [20]byte
	copy(h[:], "INP2")
	binary.LittleEndian.PutUint64(h[4:], seq)
	binary.LittleEndian.PutUint32(h[12:], uint32(len(i.Mic)))
	binary.LittleEndian.PutUint32(h[16:], uint32(len(i.Ref)))
	for _, p := range [][]byte{h[:], i.Mic, i.Ref} {
		if e := WriteAll(w, p); e != nil {
			return e
		}
	}
	return nil
}
func ReadRequest(r io.Reader, c Config, mic, ref []byte) (uint64, error) {
	var h [20]byte
	if _, e := io.ReadFull(r, h[:]); e != nil {
		return 0, e
	}
	m, n, e := c.Lengths()
	if e != nil {
		return 0, e
	}
	if string(h[:4]) != "INP2" || binary.LittleEndian.Uint32(h[12:]) != m || binary.LittleEndian.Uint32(h[16:]) != n || len(mic) != int(m) || len(ref) != int(n) {
		return 0, ErrConfig
	}
	if _, e = io.ReadFull(r, mic); e != nil {
		return 0, e
	}
	if _, e = io.ReadFull(r, ref); e != nil {
		return 0, e
	}
	return binary.LittleEndian.Uint64(h[4:]), nil
}
