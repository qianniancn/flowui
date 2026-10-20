package draw_test

import (
	"image"
	"image/color"
	"testing"

	"gioui.org/f32"
	giolayout "gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/draw"
	"github.com/qianniancn/flowui/theme"
)

func TestDrawingPrimitivesAreComposable(t *testing.T) {
	red := color.NRGBA{R: 0xff, A: 0xff}
	blue := color.NRGBA{B: 0xff, A: 0xff}
	brush := draw.LinearGradient(
		draw.GradientStop{Offset: 0, Color: red},
		draw.GradientStop{Offset: 1, Color: blue},
	).Angle(90)
	if got := brush.ColorAt(0); got != red {
		t.Fatalf("gradient start = %#v, want %#v", got, red)
	}
	if got := brush.ColorAt(1); got != blue {
		t.Fatalf("gradient end = %#v, want %#v", got, blue)
	}

	var ops op.Ops
	gtx := giolayout.Context{Ops: &ops}
	rect := image.Rect(0, 0, 80, 40)
	draw.Fill(gtx, rect, 6, brush)
	draw.RoundedBorder(gtx, rect, 6, 1, red)
	draw.BottomBorder(gtx, rect, 1, blue)
	draw.PaintLinearGradient(gtx, f32.Pt(0, 0), red, f32.Pt(80, 0), blue)

	if got := draw.SolidBrush(red).ColorAt(0.5); got != red {
		t.Fatalf("solid brush sample = %#v, want %#v", got, red)
	}
}

func TestShadowShapesAndLayersExposeStableValues(t *testing.T) {
	shape := draw.RoundedShadowCorners(unit.Dp(4), unit.Dp(8), unit.Dp(12), unit.Dp(16))
	if shape.Kind != draw.ShadowRoundedRect || shape.Radii.NE != unit.Dp(8) {
		t.Fatalf("rounded shape = %#v", shape)
	}
	if draw.EllipseShadow().Kind != draw.ShadowEllipse {
		t.Fatal("ellipse shape kind changed")
	}

	shadow := draw.BoxShadow{Layers: []draw.ShadowLayer{{Blur: 4, Color: color.NRGBA{A: 0xff}}}}
	if got := shadow.EffectiveLayers(); len(got) != 1 || got[0].Blur != 4 {
		t.Fatalf("effective shadow layers = %#v", got)
	}
	if got := draw.ShadowRasterPadding(4, 2, -3); got <= 0 {
		t.Fatalf("shadow raster padding = %d, want positive", got)
	}

	var ops op.Ops
	gtx := giolayout.Context{Ops: &ops}
	_ = shape.ClipOp(gtx, image.Rect(0, 0, 40, 20))
	_ = clip.Rect{Max: image.Pt(40, 20)}.Op()
}

func TestThemeShadowAcceptsPublicStyleTheme(t *testing.T) {
	shadow := draw.ThemeShadow(
		theme.DefaultShadows().Surface,
		color.NRGBA{A: 0xff},
		1,
	)
	if got := len(shadow.EffectiveLayers()); got == 0 {
		t.Fatal("public style shadow theme produced no drawable layers")
	}
}
