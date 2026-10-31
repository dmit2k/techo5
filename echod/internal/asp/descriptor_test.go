package asp

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func b(o, l, r, f, s uint32, format, valid, shift, channels uint8, first uint16) Buffer {
	return Buffer{o, l, r, f, s, format, valid, shift, channels, first}
}
func c(bi, role uint8, idx uint16, off uint32) Channel { return Channel{bi, role, idx, off} }

// vectors are hand-specified PCM geometries, independent of the Python wire oracle.
func vectors() map[string]Descriptor {
	maxBuffers := make([]Buffer, 255)
	maxChannels := make([]Channel, 255)
	for i := 0; i < 255; i++ {
		maxBuffers[i] = b(uint32(i*2), 2, 16000, 1, 2, FormatS16, 16, 0, 1, uint16(i))
		maxChannels[i] = c(uint8(i), RoleMIC, uint16(i), 0)
	}
	return map[string]Descriptor{
		"stereo_s24": {
			Buffers:  []Buffer{b(0, 36, 16000, 2, 18, FormatS24_3, 24, 0, 6, 0)},
			Channels: []Channel{c(0, RoleMIC, 0, 0), c(0, RoleMIC, 1, 3), c(0, RoleMIC, 2, 6), c(0, RoleMIC, 3, 9), c(0, RoleREF, 0, 12), c(0, RoleREF, 1, 15)},
		},
		"separate_three": {
			Buffers:  []Buffer{b(0, 8, 16000, 2, 4, FormatS16, 16, 0, 2, 0), b(8, 4, 16000, 2, 2, FormatS16, 16, 0, 1, 2), b(12, 4, 16000, 2, 2, FormatS16, 16, 0, 1, 3)},
			Channels: []Channel{c(0, RoleMIC, 0, 0), c(0, RoleMIC, 1, 2), c(1, RoleREF, 0, 0), c(2, RoleREF, 1, 0)},
		},
		"s32_low24": {
			Buffers:  []Buffer{b(0, 16, 48000, 2, 8, FormatS32, 24, 0, 2, 0)},
			Channels: []Channel{c(0, RoleMIC, 0, 0), c(0, RoleREF, 0, 4)},
		},
		"s32_high24": {
			Buffers:  []Buffer{b(0, 16, 48000, 2, 8, FormatS32, 24, 8, 2, 0)},
			Channels: []Channel{c(0, RoleMIC, 0, 0), c(0, RoleREF, 0, 4)},
		},
		"ref_absent": {
			Buffers:  []Buffer{b(0, 8, 16000, 2, 4, FormatS16, 16, 0, 2, 0)},
			Channels: []Channel{c(0, RoleMIC, 0, 0), c(0, RoleMIC, 1, 2)},
		},
		"packed_sparse": {
			Buffers:  []Buffer{b(0, 24, 16000, 2, 12, FormatS24_3, 20, 4, 2, 0)},
			Channels: []Channel{c(0, RoleREF, 0, 7), c(0, RoleMIC, 0, 0)},
		},
		"max_255": {Buffers: maxBuffers, Channels: maxChannels},
	}
}
func TestIndependentExpectedBytes(t *testing.T) {
	decoder := os.Getenv("ASP_C_DECODER")
	for name, desc := range vectors() {
		t.Run(name, func(t *testing.T) {
			got, e := Encode(desc)
			if e != nil {
				t.Fatal(e)
			}
			expected, e := os.ReadFile(filepath.Join("testdata", name+".bin"))
			if e != nil {
				t.Fatal(e)
			}
			if !bytes.Equal(got, expected) {
				t.Fatalf("writer differs from independent PCM-only vector at %s", name)
			}
			if len(got) != HeaderSize+len(desc.Buffers)*BufferSize+len(desc.Channels)*ChannelSize ||
				len(got) > MaxDescriptorSize || got[4] != 1 || got[5] != 0 {
				t.Fatal("metadata/trailing fields or incorrect PCM-only header")
			}
			if HeaderSize != 16 || MaxDescriptorSize != 9196 {
				t.Fatal("PCM-only contract constants")
			}
			if name == "max_255" && len(got) != 9196 {
				t.Fatal("maximum descriptor not 9196")
			}

			if decoder != "" {
				cmd := exec.Command(decoder)
				cmd.Stdin = bytes.NewReader(got)
				out, e := cmd.Output()
				if e != nil {
					t.Fatalf("compiled C reader: %v", e)
				}
				expect, e := os.ReadFile(filepath.Join("testdata", name+".expect"))
				if e != nil {
					t.Fatal(e)
				}
				if !bytes.Equal(out, expect) {
					t.Fatal("compiled C decoded transcript differs from oracle")
				}
			}
		})
	}
}
func TestReadyAndStereoREF(t *testing.T) {
	ready, e := os.ReadFile("testdata/readiness.bin")
	if e != nil {
		t.Fatal(e)
	}
	if !bytes.Equal(ready, []byte{ReadyByte}) || OutputBytes != 640 {
		t.Fatal("startup/output contract")
	}
	pcm, e := os.ReadFile("testdata/stereo_s24.pcm")
	if e != nil {
		t.Fatal(e)
	}
	if !bytes.Equal(pcm[12:15], []byte{0xa1, 0xb2, 0xc3}) ||
		!bytes.Equal(pcm[15:18], []byte{0xd4, 0xe5, 0xf6}) {
		t.Fatal("independent REF source bytes lost")
	}
	if decoder := os.Getenv("ASP_C_DECODER"); decoder != "" {
		descriptor, e := Encode(vectors()["stereo_s24"])
		if e != nil {
			t.Fatal(e)
		}
		cmd := exec.Command(decoder, "--sample")
		cmd.Stdin = bytes.NewReader(append(descriptor, pcm...))
		out, e := cmd.Output()
		if e != nil {
			t.Fatal(e)
		}
		if !bytes.Contains(out, []byte("S 2 0 a1b2c3\n")) || !bytes.Contains(out, []byte("S 2 1 d4e5f6\n")) {
			t.Fatal("stereo REF L/R mismatch")
		}
	}
}
func TestPCMBlocksAreByteExactAcrossCompiledC(t *testing.T) {
 decoder:=os.Getenv("ASP_C_DECODER")
 if decoder=="" {t.Skip("compiled C decoder supplied by C07 matrix")}
 for _,name:=range []string{"stereo_s24","separate_three"} {
  desc,err:=Encode(vectors()[name]);if err!=nil{t.Fatal(err)}
  pcm,err:=os.ReadFile(filepath.Join("testdata",name+".pcm"));if err!=nil{t.Fatal(err)}
  cmd:=exec.Command(decoder,"--roundtrip")
  cmd.Stdin=bytes.NewReader(append(desc,pcm...))
  got,err:=cmd.Output();if err!=nil{t.Fatalf("%s: %v",name,err)}
  if !bytes.Equal(got,pcm){t.Fatalf("%s: changed physical PCM bytes",name)}
 }
}
func TestRejectBadWriterInputs(t *testing.T) {
	base := vectors()["ref_absent"]
	bad := map[string]func(*Descriptor){
		"no buffers":      func(d *Descriptor) { d.Buffers = nil },
		"no mic":          func(d *Descriptor) { d.Channels[0].Role = RoleREF; d.Channels[1].Role = RoleREF },
		"invalid format":  func(d *Descriptor) { d.Buffers[0].Format = 4 },
		"bad shift":       func(d *Descriptor) { d.Buffers[0].LSBShift = 1 },
		"zero rate":       func(d *Descriptor) { d.Buffers[0].SampleRate = 0 },
		"bad offset":      func(d *Descriptor) { d.Buffers[0].Offset = 1 },
		"bad stride":      func(d *Descriptor) { d.Buffers[0].Stride = 3 },
		"last sample":     func(d *Descriptor) { d.Buffers[0].Length = 7 },
		"duplicate mic":   func(d *Descriptor) { d.Channels[1].RoleIndex = 0 },
		"bad role":        func(d *Descriptor) { d.Channels[1].Role = 3 },
		"buffer index":    func(d *Descriptor) { d.Channels[0].BufferIndex = 1 },
		"channel overlap": func(d *Descriptor) { d.Channels[1].FrameOffset = 1 },
	}
	for name, mutate := range bad {
		t.Run(name, func(t *testing.T) {
			d := Descriptor{Buffers: append([]Buffer(nil), base.Buffers...), Channels: append([]Channel(nil), base.Channels...)}
			mutate(&d)
			if _, e := Encode(d); e == nil {
				t.Fatal("invalid descriptor accepted")
			}
		})
	}
}
func TestSharedSHA256Manifest(t *testing.T) {
	raw, e := os.ReadFile("testdata/SHA256SUMS")
	if e != nil {
		t.Fatal(e)
	}
	for _, line := range strings.Split(strings.TrimSpace(string(raw)), "\n") {
		parts := strings.SplitN(line, "  ", 2)
		if len(parts) != 2 {
			t.Fatal("manifest format")
		}
		data, e := os.ReadFile(filepath.Join("testdata", parts[1]))
		if e != nil {
			t.Fatal(e)
		}
		digest := sha256.Sum256(data)
		if hex.EncodeToString(digest[:]) != parts[0] {
			t.Fatal(fmt.Sprintf("digest mismatch: %s", parts[1]))
		}
	}
}

// The encoder has no raw metadata fields and cannot append unknown sections.
func TestPCMOnlyNoExtraSection(t *testing.T) {
	d := vectors()["ref_absent"]
	got, err := Encode(d)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 60 {
		t.Fatalf("unexpected descriptor length %d", len(got))
	}
	expect, err := os.ReadFile("testdata/ref_absent.bin")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, expect) {
		t.Fatal("exact PCM-only wire oracle mismatch")
	}
	// At 16 bytes the first buffer starts immediately: no extra 8-byte header.
	if got[16] != 0 || got[20] != 8 || got[24] != 0x80 || got[25] != 0x3e {
		t.Fatalf("16-byte boundary not PCM buffer record: %x", got[16:28])
	}
}
