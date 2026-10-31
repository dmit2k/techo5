package layout

import (
	"bytes"
	"encoding/binary"
	"os"
	"strings"
)

// RawSource is an unmodified source value and the name of the actual source.
// Values may contain embedded/trailing NUL, whitespace and non-UTF8 bytes.
// Neither IDME parsing nor stock Idme() normalization is used for External.
type RawSource struct {
	Source string
	Bytes  []byte
}

// RawIdentitySources reads the same physical identity sources stock layout already
// consumes, without turning NUL-delimited values or kernel command-line bytes into
// interpreted platform fields. Missing sources are absent, never synthesized.
func RawIdentitySources() []RawSource {
	return rawIdentitySources(os.ReadFile)
}
func rawIdentitySources(read func(string) ([]byte, error)) []RawSource {
	var out []RawSource
	for _, name := range []string{"/proc/device-tree/compatible", "/proc/cmdline"} {
		if b, err := read(name); err == nil {
			out = append(out, RawSource{name, b})
		}
	}
	return out
}

// RawFactorySources enumerates the existing /proc/idme and DT-IDME raw values.
// On Athena-mainline, stock falls back to eMMC independently for every
// absent field. Merge missing original boot-area fields even when /proc
// exports some fields, preserving the /proc values when both are present.
// Call at initialization, not on the real-time capture loop.
func RawFactorySources() []RawSource {
	return rawFactorySources(os.ReadDir, os.ReadFile, AthenaMainline())
}
func rawFactorySources(list func(string) ([]os.DirEntry, error),
	read func(string) ([]byte, error), bootFallback bool) []RawSource {
	var out []RawSource
	for _, dir := range []struct{ root, suffix string }{
		{"/proc/idme", ""}, {"/proc/device-tree/idme", "/value"},
	} {
		entries, err := list(dir.root)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			key := entry.Name()
			if !rawSourceKey(key) {
				continue
			}
			path := dir.root + "/" + key + dir.suffix
			if len(path) > 63 {
				continue
			}
			if b, err := read(path); err == nil {
				out = append(out, RawSource{path, b})
			}
		}
	}
	if bootFallback {
		// The stock Idme path uses eMMC fallback per missing key, not only
		// when the whole /proc tree is empty. Source names stay explicit.
		present := make(map[string]bool, len(out))
		for _, src := range out {
			switch {
			case strings.HasPrefix(src.Source, "/proc/idme/"):
				present[strings.TrimPrefix(src.Source, "/proc/idme/")] = true
			case strings.HasPrefix(src.Source, "/proc/device-tree/idme/") && strings.HasSuffix(src.Source, "/value"):
				name := strings.TrimSuffix(strings.TrimPrefix(src.Source, "/proc/device-tree/idme/"), "/value")
				present[name] = true
			}
		}
		if b, err := read("/dev/mmcblk0boot1"); err == nil {
			for _, src := range rawBootIDMESources(b) {
				key := strings.TrimPrefix(src.Source, "/dev/mmcblk0boot1/")
				if !present[key] {
					out = append(out, src)
				}
			}
		}
	}
	return out
}

// rawSourceKey accepts names representable by the *unchanged* Stage02 wire
// codec. Unknown valid keys pass through; no hardcoded calibration-key list.
func rawSourceKey(s string) bool {
	if len(s) == 0 || len(s) > 63 {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' || c == '-' || c == '.') {
			return false
		}
	}
	return true
}

// rawBootIDMESources extracts the bytes of every IDME v2 entry without any
// string trim, semantic calibration processing or MIC-slot association.
// The envelope is only used to locate the original value byte ranges.
func rawBootIDMESources(b []byte) []RawSource {
	if len(b) < 16 || !bytes.HasPrefix(b, []byte("beefdeed")) {
		return nil
	}
	count := binary.LittleEndian.Uint32(b[12:])
	off := uint64(16)
	var out []RawSource
	for i := uint32(0); i < count; i++ {
		if off+28 > uint64(len(b)) {
			return nil
		} // malformed source, never synthesize
		h := b[int(off) : int(off)+28]
		key := string(bytes.TrimRight(h[:16], "\x00"))
		size := uint64(binary.LittleEndian.Uint32(h[16:]))
		end := off + 28 + size
		if end > uint64(len(b)) {
			return nil
		}
		if rawSourceKey(key) {
			out = append(out, RawSource{"/dev/mmcblk0boot1/" + key, b[int(off)+28 : int(end)]})
		}
		off = end
	}
	return out
}
