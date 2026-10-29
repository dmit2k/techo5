package asp

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
	"testing"
)

func shape(m, r int) Config {
	c := Config{Flags: Interleaved, Mic: Stream{Channels: uint16(m), Frames: 1920, Rate: 96000, Format: S32, FullBits: 23, ValidBits: 24}}
	if r > 0 {
		c.Flags |= Synchronized | HardwareRef
		c.Ref = Stream{Channels: uint16(r), Frames: 1920, Rate: 96000, Format: S16, FullBits: 15, ValidBits: 16}
	}
	return c
}
func TestConfigMatrixAndBytes(t *testing.T) {
	for m := 1; m <= 8; m++ {
		for r := 0; r <= 2; r++ {
			t.Run(fmt.Sprintf("%d+%d", m, r), func(t *testing.T) {
				c := shape(m, r)
				b, e := (CFG2{Config: c}).MarshalBinary()
				if e != nil {
					t.Fatal(e)
				}
				got, e := ReadCFG2(bytes.NewReader(b[:]))
				if e != nil || got.Config != c {
					t.Fatal(got, e)
				}
				mb, rb, e := got.Config.Lengths()
				if e != nil || mb != uint32(1920*m*4) || rb != uint32(1920*r*2) {
					t.Fatal(mb, rb, e)
				}
				if b[24] != byte(m) || b[48] != byte(r) || b[32] != S32 || b[33] != 23 || b[35] != 24 {
					t.Fatal("wire offsets")
				}
			})
		}
	}
	for _, m := range []int{9, 12, 16} {
		c := shape(m, 8)
		if _, _, e := c.Lengths(); e != nil {
			t.Fatal(m, e)
		}
	}
	c := shape(16, 8)
	c.Mic.Rate, c.Ref.Rate = 192000, 192000
	c.Mic.Frames, c.Ref.Frames = 3840, 3840
	c.Ref.Format = S32
	c.Ref.FullBits = 31
	c.Ref.ValidBits = 32
	if _, _, e := c.Lengths(); e == nil {
		t.Fatal("combined 368640 bytes not capped")
	}
	c = shape(7, 2)
	c.Mic.Rate, c.Mic.Frames = 48000, 960
	c.Ref.Rate, c.Ref.Frames = 16000, 320
	c.Mic.Alignment, c.Mic.FullBits = High, 31
	if _, _, e := c.Lengths(); e != nil {
		t.Fatal("independent formats/rates", e)
	}
}
func TestConfigRejectsAndTruncation(t *testing.T) {
	for _, mutate := range []func(*Config){
		func(c *Config) { c.Mic.Channels = 0 }, func(c *Config) { c.Mic.Channels = 17 }, func(c *Config) { c.Ref.Channels = 9 },
		func(c *Config) { c.Mic.Rate = 7999 }, func(c *Config) { c.Mic.Rate = 192001 }, func(c *Config) { c.Mic.Frames = 1 },
		func(c *Config) { c.Mic.Format = 99 }, func(c *Config) { c.Mic.FullBits = 31 }, func(c *Config) { c.Mic.ValidBits = 33 },
		func(c *Config) { c.Mic.Alignment = 2 }, func(c *Config) { c.Flags = Interleaved }, func(c *Config) { c.Flags |= 128 },
		func(c *Config) { c.Ref = Stream{} }, func(c *Config) { c.Mic.Rate = ^uint32(0); c.Mic.Frames = ^uint16(0) },
	} {
		c := shape(4, 2)
		mutate(&c)
		if _, _, e := c.Lengths(); e == nil {
			t.Fatal("accepted invalid", c)
		}
	}
	c := shape(4, 0)
	c.Ref.Rate = 96000
	if _, _, e := c.Lengths(); e == nil {
		t.Fatal("invented absent ref")
	}
	b, _ := (CFG2{Config: shape(4, 2)}).MarshalBinary()
	for n := 0; n < len(b); n++ {
		if _, e := ReadCFG2(bytes.NewReader(b[:n])); e == nil {
			t.Fatal("truncated", n)
		}
	}
	for _, off := range []int{0, 4, 6, 36, 39} {
		x := bytes.Clone(b)
		x[off] ^= 0xff
		if _, e := ReadCFG2(bytes.NewReader(x[:])); e == nil {
			t.Fatal("corrupt", off)
		}
	}
}
func TestStatusOptionalExactFormatAndClasses(t *testing.T) {
	cases := []struct {
		s         Status
		def, want string
	}{
		{Status{}, "", "000: "},
		{Status{State, 200, "External ASP started (8 mic, 2 ref)"}, "External ASP starting", "200: External ASP started (8 mic, 2 ref)"},
		{Status{State, 201, ""}, "External ASP processing OK", "201: External ASP processing OK"},
		{Status{State, 301, "Noise suppressor disabled"}, "unused", "301: Noise suppressor disabled"},
		{Status{State, 422, "Reduced array mode active"}, "unused", "422: Reduced array mode active"},
		{Status{State, 522, "Required hardware reference unavailable"}, "unused", "522: Required hardware reference unavailable"},
		{Status{State, 503, "ASP processing failed"}, "unused", "503: ASP processing failed"},
	}
	for _, x := range cases {
		v, e := x.s.Validate()
		if e != nil || v.Text(x.def) != x.want {
			t.Fatal(x, v, e)
		}
	}
	for _, code := range []uint16{1, 99, 100, 199, 600, 65535} {
		if _, e := (Status{State, code, ""}).Validate(); e == nil {
			t.Fatal(code)
		}
	}
	for _, code := range []uint16{200, 249, 250, 299, 300, 349, 350, 399, 400, 449, 450, 499, 500, 549, 550, 599} {
		s := Status{State, code, ""}
		if _, e := s.Validate(); e != nil {
			t.Fatal(code, e)
		}
	}
	for _, s := range []Status{{9, 200, "x"}, {0, 200, ""}, {State, 0, "x"}, {Event, 0, "x"}, {Event, 301, "x"}, {Event, 422, "x"}} {
		if _, e := s.Validate(); e == nil {
			t.Fatal(s)
		}
	}
	for _, msg := range []string{"hello\nworld", "<b>test</b>", "/private/file", "key=secret", "192.0.2.1", "a\u202eb", "x\x1b[31m"} {
		s, e := (Status{Event, 210, msg}).Validate()
		if e != nil || s.Message != "" {
			t.Fatal("unsafe text leaked", s, e)
		}
	}
	for _, msg := range []string{string([]byte{0xff}), strings.Repeat("a", 161)} {
		if _, e := (Status{State, 200, msg}).Validate(); e == nil {
			t.Fatal("malformed accepted")
		}
	}
}
func TestTypedFatalPreemptsPCMAndAccept(t *testing.T) {
	for _, typ := range []uint8{Data, NotReady, Fatal, 99} {
		for _, kind := range []uint8{State, Event} {
			var b bytes.Buffer
			r := Reply{Type: typ, Sequence: 0, Start: 0, PCM: make([]int16, Samples), Status: Status{kind, 503, "ASP processing failed"}}
			WriteReply(&b, r)
			// Header + complete status only. Never wait for attached PCM after a valid fatal.
			_, e := ReadReply(bytes.NewReader(b.Bytes()[:32+len(r.Status.Message)]))
			var terminal *Terminal
			if !errors.As(e, &terminal) || terminal.Status != r.Status {
				t.Fatal(typ, kind, e)
			}
		}
	}
	for _, typ := range []uint8{Accept, Reject, 99} {
		var b bytes.Buffer
		WriteResult(&b, ConfigResult{Type: typ, Status: Status{State, 522, "Required hardware reference unavailable"}})
		_, e := ReadResult(&b)
		var fatal *Terminal
		if !errors.As(e, &fatal) {
			t.Fatal(typ, e)
		}
	}
	for _, typ := range []uint8{Reject, Fatal} {
		for _, code := range []uint16{0, 201, 301, 422} {
			s := Status{State, code, ""}
			var b bytes.Buffer
			if typ == Reject {
				WriteResult(&b, ConfigResult{Type: Reject, Status: s})
				_, err := ReadResult(&b)
				if !errors.Is(err, ErrStatus) {
					t.Fatal(code, err)
				}
			} else {
				WriteReply(&b, Reply{Type: Fatal, Status: s})
				_, err := ReadReply(&b)
				if !errors.Is(err, ErrStatus) {
					t.Fatal(code, err)
				}
			}
		}
	}
}
func TestReplyByteBoundsAndRequest(t *testing.T) {
	pcm := make([]int16, Samples)
	for i := range pcm {
		pcm[i] = int16(i * 37)
	}
	r := Reply{Type: Data, Sequence: 7, Start: 320, PCM: pcm, Status: Status{Event, 210, "Simulated beam changed"}}
	var b bytes.Buffer
	WriteReply(&b, r)
	data := bytes.Clone(b.Bytes())
	got, e := ReadReply(&b)
	if e != nil || !reflect.DeepEqual(got, r) {
		t.Fatal(got, e)
	}
	for n := 0; n < len(data); n++ {
		if _, e := ReadReply(bytes.NewReader(data[:n])); e == nil {
			t.Fatal(n)
		}
	}
	for _, off := range []int{0, 5, 8, 10, 28} {
		x := bytes.Clone(data)
		x[off] = 255
		if _, e := ReadReply(bytes.NewReader(x)); e == nil {
			t.Fatal("invalid", off)
		}
	}
	nr := Reply{Type: NotReady, Sequence: 0, Start: NoSample}
	b.Reset()
	WriteReply(&b, nr)
	if _, e := ReadReply(&b); e != nil {
		t.Fatal(e)
	}
	c := shape(8, 0)
	m, n, _ := c.Lengths()
	input := Input{c, make([]byte, m), make([]byte, n)}
	b.Reset()
	if e := WriteRequest(&b, 9, input); e != nil {
		t.Fatal(e)
	}
	wire := bytes.Clone(b.Bytes())
	seq, e := ReadRequest(&b, c, input.Mic, input.Ref)
	if e != nil || seq != 9 {
		t.Fatal(seq, e)
	}
	binary.LittleEndian.PutUint32(wire[12:], ^uint32(0))
	if _, e := ReadRequest(bytes.NewReader(wire), c, input.Mic, input.Ref); e == nil {
		t.Fatal("oversize")
	}
	if e := WriteAll(shortWriter{}, []byte{1}); !errors.Is(e, io.ErrShortWrite) {
		t.Fatal(e)
	}
}

type shortWriter struct{}

func (shortWriter) Write([]byte) (int, error) { return 0, nil }
