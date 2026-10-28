//go:build !dot && !spot && !panel

package mic

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/HuskerMinion/techo5/echod/internal/lib/alsa"
)

func TestASPAthenaBorrowedBeforeConvert(t *testing.T) {
	asShow8(t)
	for _, k := range []*athenaKernel{&athenaAmazonKernel, &athenaMainlineKernel} {
		c := newPairedCapture(nil, &alsa.Capture{}, 96000)
		c.k = k
		bits := uint32(23)
		if k.micFull == 1<<31 {
			bits = 31
		}
		n := nativePattern(bits, 7)
		m, r := bytes.Clone(n.mics), bytes.Clone(n.ref)
		c.retainNative(n.mics, n.ref)
		view := c.nativeInput()
		if c.aspConfig() != n.input().Config {
			t.Fatal("cold capture shape differs from native PCM")
		}
		if !view.valid() || view.fullBits != bits || &view.mics[0] != &n.mics[0] || &view.ref[0] != &n.ref[0] {
			t.Fatal("not the original borrowed streams")
		}
		c.convert(make([]byte, FrameSamples*Channels*Bits/8), n.mics, n.ref)
		if !bytes.Equal(view.mics, m) || !bytes.Equal(view.ref, r) {
			t.Fatal("convert changed native buffers")
		}
		c.ref = nil
		c.retainNative(n.mics, n.ref)
		if _, _, err := c.aspConfig().Lengths(); err == nil {
			t.Fatal("unavailable native capture has a valid cold snapshot")
		}
		if c.nativeInput().valid() {
			t.Fatal("missing hardware reference accepted")
		}
		c.ref = &alsa.Capture{}
		c.micRate = 16000
		c.retainNative(n.mics, n.ref)
		if _, _, err := c.aspConfig().Lengths(); err == nil {
			t.Fatal("unavailable native capture has a valid cold snapshot")
		}
		if c.nativeInput().valid() {
			t.Fatal("downsampled paired capture accepted")
		}
		c.micRate = 96000
		c.retainNative(n.mics[:len(n.mics)-1], n.ref)
		if c.nativeInput().valid() {
			t.Fatal("short input accepted")
		}
	}
}

// The exact Read cannot run without ALSA ioctls. Assert its audited call-order directly as well
// as exercising retainNative/convert above and Source.Run using a synthetic capture separately.
func TestASPAthenaReadOrder(t *testing.T) {
	f, err := parser.ParseFile(token.NewFileSet(), "athena.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var calls []string
	var resetFirst bool
	for _, d := range f.Decls {
		fn, ok := d.(*ast.FuncDecl)
		if !ok || fn.Name.Name != "Read" || fn.Recv == nil {
			continue
		}
		a, ok := fn.Body.List[0].(*ast.AssignStmt)
		resetFirst = ok && len(a.Lhs) == 1
		if resetFirst {
			s, ok := a.Lhs[0].(*ast.SelectorExpr)
			resetFirst = ok && s.Sel.Name == "native"
		}
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			if c, ok := n.(*ast.CallExpr); ok {
				switch x := c.Fun.(type) {
				case *ast.Ident:
					if x.Name == "readFull" {
						calls = append(calls, x.Name)
					}
				case *ast.SelectorExpr:
					if x.Sel.Name == "retainNative" || x.Sel.Name == "convert" {
						calls = append(calls, x.Sel.Name)
					}
				}
			}
			return true
		})
	}
	want := []string{"readFull", "readFull", "retainNative", "convert"}
	if !resetFirst || len(calls) != len(want) {
		t.Fatal("Read reset/capture order changed", calls)
	}
	for i := range want {
		if calls[i] != want[i] {
			t.Fatal("tap not after paired reads and before convert", calls)
		}
	}
}
