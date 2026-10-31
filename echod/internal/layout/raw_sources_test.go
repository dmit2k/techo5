package layout

import (
	"bytes"
	"encoding/binary"
	"io/fs"
	"os"
	"reflect"
	"testing"
)

type testEntry string

func (e testEntry) Name() string               { return string(e) }
func (e testEntry) IsDir() bool                { return false }
func (e testEntry) Type() fs.FileMode          { return 0 }
func (e testEntry) Info() (os.FileInfo, error) { return nil, os.ErrNotExist }

func TestN07RawProcAndIdentityBytes(t *testing.T) {
	src := map[string][]byte{
		"/proc/idme/miccal.0":                {0, 0xff, 0x80, ' ', 0},
		"/proc/idme/unknown_X":               {'\n', 0, 0x13},
		"/proc/idme/empty":                   {},
		"/proc/device-tree/idme/extra/value": {0x99, 0, 0},
		"/proc/device-tree/compatible":       {'a', 0, 'b', 0},
		"/proc/cmdline":                      {'x', '=', 'y', ' ', 0},
	}
	list := func(path string) ([]os.DirEntry, error) {
		switch path {
		case "/proc/idme":
			return []os.DirEntry{testEntry("miccal.0"), testEntry("unknown_X"), testEntry("empty")}, nil
		case "/proc/device-tree/idme":
			return []os.DirEntry{testEntry("extra")}, nil
		}
		return nil, os.ErrNotExist
	}
	read := func(path string) ([]byte, error) {
		b, ok := src[path]
		if !ok {
			return nil, os.ErrNotExist
		}
		return b, nil
	}
	got := rawFactorySources(list, read, false)
	if len(got) != 4 {
		t.Fatalf("raw sources %v", got)
	}
	for _, field := range got {
		if !bytes.Equal(field.Bytes, src[field.Source]) {
			t.Fatalf("%q modified: %x", field.Source, field.Bytes)
		}
	}
	if !reflect.DeepEqual(rawIdentitySources(read), []RawSource{
		{"/proc/device-tree/compatible", src["/proc/device-tree/compatible"]},
		{"/proc/cmdline", src["/proc/cmdline"]},
	}) {
		t.Fatal("identity was normalized")
	}
	// Missing differs from a present zero-length field.
	missing := rawFactorySources(func(string) ([]os.DirEntry, error) { return nil, os.ErrNotExist }, read, false)
	if len(missing) != 0 {
		t.Fatalf("missing sources: %v", missing)
	}
}

func TestN07BootAreaRawIDME(t *testing.T) {
	b := make([]byte, 16)
	copy(b, "beefdeed")
	binary.LittleEndian.PutUint32(b[12:], 2)
	put := func(name string, value []byte) {
		rec := make([]byte, 28)
		copy(rec[:16], name)
		binary.LittleEndian.PutUint32(rec[16:], uint32(len(value)))
		b = append(b, rec...)
		b = append(b, value...)
	}
	put("miccal.0", []byte{0x81, 0, 0xff, ' ', 0})
	put("unrecognized", []byte{})
	got := rawBootIDMESources(b)
	if len(got) != 2 || got[0].Source != "/dev/mmcblk0boot1/miccal.0" ||
		!bytes.Equal(got[0].Bytes, []byte{0x81, 0, 0xff, ' ', 0}) ||
		got[1].Source != "/dev/mmcblk0boot1/unrecognized" || got[1].Bytes == nil {
		t.Fatalf("raw boot IDME: %#v", got)
	}
	// Explicitly verify fallback only when nothing came from /proc and only via fake reader.
	requested := []string{}
	reader := func(path string) ([]byte, error) {
		requested = append(requested, path)
		if path == "/dev/mmcblk0boot1" {
			return b, nil
		}
		return nil, os.ErrNotExist
	}
	from := rawFactorySources(func(string) ([]os.DirEntry, error) { return nil, os.ErrNotExist }, reader, true)
	if len(from) != 2 || requested[len(requested)-1] != "/dev/mmcblk0boot1" {
		t.Fatal("fallback did not use raw image")
	}
	if v := rawBootIDMESources([]byte("bad")); v != nil {
		t.Fatal("malformed IDME accepted")
	}
	// Unknown platform fields are neither mapped into MIC slots nor interpreted.
}
