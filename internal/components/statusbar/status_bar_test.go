package statusbar

import (
	"image"
	"image/color"
	"testing"

	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/flowui/internal/frame"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/theme"
)

func TestStatusBarFillsWidthAndScopesDefaultColors(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	ctx := frame.New(nil, &activeTheme, locale.LanguageAuto)
	left := new(statusBarProbe)
	right := new(statusBarProbe)
	var ops op.Ops
	dims := New(left, right).Layout(ctx, layout.Context{
		Constraints: layout.Constraints{Min: image.Pt(300, 0), Max: image.Pt(300, 100)},
		Ops:         &ops,
	})

	if dims.Size != image.Pt(300, 28) {
		t.Fatalf("status bar size = %v, want (300,28)", dims.Size)
	}
	for _, probe := range []*statusBarProbe{left, right} {
		if probe.foreground != activeTheme.Palette.SurfaceSecondaryForeground || probe.background != activeTheme.Palette.SurfaceSecondary {
			t.Fatalf("status bar child colors = %#v/%#v", probe.foreground, probe.background)
		}
		if probe.constraints.Min.Y != 0 {
			t.Fatalf("status bar child minimum height = %d, want 0", probe.constraints.Min.Y)
		}
	}
}

type statusBarProbe struct {
	foreground  color.NRGBA
	background  color.NRGBA
	constraints layout.Constraints
}

func (p *statusBarProbe) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	p.foreground = ctx.ForegroundColor()
	p.background = ctx.BackgroundColor()
	p.constraints = gtx.Constraints
	return layout.Dimensions{Size: gtx.Constraints.Constrain(image.Pt(60, 14))}
}
