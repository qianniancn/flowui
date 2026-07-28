package statusbar

import (
	"image"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	layoutui "github.com/qianniancn/flowui/internal/components/layout"
	"github.com/qianniancn/flowui/internal/components/surface"
	"github.com/qianniancn/flowui/internal/frame"
)

const (
	barHeight  unit.Dp = 28
	barPadding unit.Dp = 10
	barGap     unit.Dp = 8
	barBorder  unit.Dp = 1
)

// Widget presents compact application state at the bottom of a window.
type Widget struct {
	left  frame.Widget
	right frame.Widget
}

func New(left, right frame.Widget) Widget {
	return Widget{left: left, right: right}
}

func (w Widget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	size := gtx.Constraints.Constrain(image.Pt(gtx.Constraints.Max.X, gtx.Dp(barHeight)))
	rootGtx := gtx
	rootGtx.Constraints = layout.Exact(size)
	children := make([]frame.Widget, 0, 3)
	if w.left != nil {
		children = append(children, w.left)
	}
	children = append(children, layoutui.Expanded(layoutui.Spacer(0, 0)))
	if w.right != nil {
		children = append(children, w.right)
	}
	content := frame.WidgetFunc(func(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
		return layoutui.LayoutTrackedDirection(ctx, gtx, layout.Center, func(gtx layout.Context) layout.Dimensions {
			return layoutui.LayoutTrackedInset(ctx, gtx, layout.Inset{Left: barPadding, Right: barPadding}, func(gtx layout.Context) layout.Dimensions {
				return layoutui.LayoutTrackedFlex(ctx, gtx, layout.Horizontal, barGap, layout.Middle, children...)
			})
		})
	})
	dims := surface.Surface(content).Variant(surface.SurfaceSecondary).Layout(ctx, rootGtx)
	if dims.Size.X > 0 && dims.Size.Y > 0 {
		width := min(max(gtx.Dp(barBorder), 1), dims.Size.Y)
		paint.FillShape(gtx.Ops, frame.ActiveTheme(ctx).Palette.SeparatorColor(), clip.Rect{Max: image.Pt(dims.Size.X, width)}.Op())
	}
	return dims
}
