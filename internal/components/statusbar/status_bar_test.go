package statusbar

import (
	"image"
	"image/color"
	"testing"

	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/locale"
	"github.com/qianniancn/FlowUI/internal/theme"
)

func TestStatusBarOptions(t *testing.T) {
	base := New(nil, nil)
	styled := base.Variant(Accent).Height(32).Border(false)
	if base.variant != Default || base.height != 0 || !base.border {
		t.Fatal("status bar options mutated the original value")
	}
	if styled.variant != Accent || styled.height != 32 || styled.border {
		t.Fatalf("status bar options = %#v", styled)
	}
}

func TestStatusBarRejectsNonPositiveHeight(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("non-positive height did not panic")
		}
	}()
	New(nil, nil).Height(0)
}

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

func TestStatusBarAccentColors(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	ctx := frame.New(nil, &activeTheme, locale.LanguageAuto)
	probe := new(statusBarProbe)
	var ops op.Ops
	New(probe, nil).Variant(Accent).Layout(ctx, layout.Context{
		Constraints: layout.Constraints{Min: image.Pt(240, 0), Max: image.Pt(240, 80)},
		Ops:         &ops,
	})

	if probe.foreground != activeTheme.Palette.AccentForeground || probe.background != activeTheme.Palette.Accent {
		t.Fatalf("accent child colors = %#v/%#v", probe.foreground, probe.background)
	}
}

func TestStatusBarUnknownVariantFallsBackToDefault(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	got := statusBarStyleFor(&activeTheme, Variant(255))
	want := statusBarStyleFor(&activeTheme, Default)
	if got != want {
		t.Fatalf("unknown variant style = %#v, want %#v", got, want)
	}
}

func TestStatusBarTheme(t *testing.T) {
	tokens := theme.DefaultTheme().Components.StatusBar
	if tokens.Height != 28 || tokens.PaddingX != 10 || tokens.Gap != 8 || tokens.BorderWidth != 1 {
		t.Fatalf("status bar theme = %#v", tokens)
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
