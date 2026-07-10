package layoutui

import (
	"image"
	"image/color"
	"testing"

	"gioui.org/layout"
	"gioui.org/op"
)

func TestDividerLayout(t *testing.T) {
	var ops op.Ops

	dims := Divider().Thickness(2).Layout(newContext(nil), layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 100)},
		Ops:         &ops,
	})

	if dims.Size != image.Pt(300, 2) {
		t.Fatalf("divider size = %v, want (300,2)", dims.Size)
	}
}

func TestSeparatorLayout(t *testing.T) {
	var ops op.Ops

	dims := Separator().Thickness(3).Layout(newContext(nil), layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 100)},
		Ops:         &ops,
	})

	if dims.Size != image.Pt(3, 100) {
		t.Fatalf("separator size = %v, want (3,100)", dims.Size)
	}
}

func TestDividerColor(t *testing.T) {
	d := Divider().Color(color.NRGBA{A: 0xff})

	if !d.hasColor {
		t.Fatal("divider color was not set")
	}
}
