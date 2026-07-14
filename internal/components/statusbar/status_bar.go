package statusbar

import (
	"image"
	"image/color"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	layoutui "github.com/qianniancn/FlowUI/internal/components/layout"
	"github.com/qianniancn/FlowUI/internal/components/surface"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/render"
	"github.com/qianniancn/FlowUI/internal/theme"
)

type Variant uint8

const (
	Default Variant = iota
	Accent
	Transparent
)

// Widget presents compact application state at the bottom of a window.
type Widget struct {
	left    frame.Widget
	right   frame.Widget
	variant Variant
	height  unit.Dp
	border  bool
}

func New(left, right frame.Widget) Widget {
	return Widget{left: left, right: right, border: true}
}

func (w Widget) Variant(variant Variant) Widget {
	w.variant = variant
	return w
}

func (w Widget) Height(dp int) Widget {
	if dp <= 0 {
		panic("flowui: status bar height must be positive")
	}
	w.height = unit.Dp(dp)
	return w
}

func (w Widget) Border(visible bool) Widget {
	w.border = visible
	return w
}

func (w Widget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	activeTheme := frame.ActiveTheme(ctx)
	tokens := activeTheme.Components.StatusBar
	height := tokens.Height
	if w.height > 0 {
		height = w.height
	}
	size := gtx.Constraints.Constrain(image.Pt(gtx.Constraints.Max.X, max(gtx.Dp(height), 1)))
	rootGtx := gtx
	rootGtx.Constraints = layout.Exact(size)

	content := frame.WidgetFunc(func(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
		padding := max(tokens.PaddingX, 0)
		return layoutui.LayoutTrackedDirection(ctx, gtx, layout.Center, func(gtx layout.Context) layout.Dimensions {
			return layoutui.LayoutTrackedInset(ctx, gtx, layout.Inset{Left: padding, Right: padding}, func(gtx layout.Context) layout.Dimensions {
				return layoutui.LayoutTrackedFlex(ctx, gtx, layout.Horizontal, tokens.Gap, layout.Middle, statusBarChildren(w.left, w.right)...)
			})
		})
	})
	style := statusBarStyleFor(activeTheme, w.variant)
	bar := surface.Surface(content).Variant(style.surface)
	if style.custom {
		bar = bar.Background(render.SolidBrush(style.background)).Foreground(style.foreground)
	}
	dims := bar.Layout(ctx, rootGtx)
	if w.border && dims.Size.X > 0 && dims.Size.Y > 0 {
		width := min(max(gtx.Dp(tokens.BorderWidth), 1), dims.Size.Y)
		paint.FillShape(gtx.Ops, style.border, clip.Rect{Max: image.Pt(dims.Size.X, width)}.Op())
	}
	return dims
}

func statusBarChildren(left, right frame.Widget) []frame.Widget {
	children := make([]frame.Widget, 0, 3)
	if left != nil {
		children = append(children, left)
	}
	children = append(children, layoutui.Expanded(layoutui.Spacer(0, 0)))
	if right != nil {
		children = append(children, right)
	}
	return children
}

type statusBarStyle struct {
	surface    surface.SurfaceVariant
	background color.NRGBA
	foreground color.NRGBA
	border     color.NRGBA
	custom     bool
}

func statusBarStyleFor(activeTheme *theme.Theme, variant Variant) statusBarStyle {
	switch variant {
	case Accent:
		return statusBarStyle{
			background: activeTheme.Palette.Accent,
			foreground: activeTheme.Palette.AccentForeground,
			border:     theme.ColorOr(activeTheme.Palette.AccentHover, activeTheme.Palette.Accent),
			custom:     true,
		}
	case Transparent:
		return statusBarStyle{surface: surface.SurfaceTransparent, border: activeTheme.Palette.SeparatorColor()}
	default:
		return statusBarStyle{surface: surface.SurfaceSecondary, border: activeTheme.Palette.SeparatorColor()}
	}
}
