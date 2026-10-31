package layout

import (
	"bytes"
	"encoding/binary"
	"io/fs"
	"os"
	"testing"
)

type fakeIDMEEntry string

func (e fakeIDMEEntry) Name() string               { return string(e) }
func (e fakeIDMEEntry) IsDir() bool                { return false }
func (e fakeIDMEEntry) Type() fs.FileMode          { return 0 }
func (e fakeIDMEEntry) Info() (os.FileInfo, error) { return nil, os.ErrNotExist }

// A partially exported /proc/idme must not cause us to lose absent raw eMMC
// factory calibration. Stock Idme() tries the boot-area fallback per field.
func TestN07PartialProcBootFallback(t *testing.T) {
	b := make([]byte, 16)
	copy(b, "beefdeed")
	binary.LittleEndian.PutUint32(b[12:], 2)
	add := func(name string, val []byte) {
		h := make([]byte, 28)
		copy(h, name)
		binary.LittleEndian.PutUint32(h[16:], uint32(len(val)))
		b = append(b, h...)
		b = append(b, val...)
	}
	add("serial", []byte{'B', 'O', 'O', 'T', 0})
	calibration := []byte{0x81, 0, 0xff, ' ', 0}
	add("miccal.0", calibration)
	list := func(path string) ([]os.DirEntry, error) {
		if path == "/proc/idme" {
			return []os.DirEntry{fakeIDMEEntry("serial")}, nil
		}
		return nil, os.ErrNotExist
	}
	read := func(path string) ([]byte, error) {
		switch path {
		case "/proc/idme/serial":
			return []byte{'P', 'R', 'O', 'C', 0}, nil
		case "/dev/mmcblk0boot1":
			return b, nil
		default:
			return nil, os.ErrNotExist
		}
	}
	got := rawFactorySources(list, read, true)
	if len(got) != 2 ||
		got[0].Source != "/proc/idme/serial" ||
		!bytes.Equal(got[0].Bytes, []byte{'P', 'R', 'O', 'C', 0}) ||
		got[1].Source != "/dev/mmcblk0boot1/miccal.0" ||
		!bytes.Equal(got[1].Bytes, calibration) {
		t.Fatalf("partial /proc loses raw calibration, or duplicates fallback values: %+v", got)
	}
}
