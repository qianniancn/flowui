package progress

import (
	"image"
	"time"

	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/text"
)

func (p BarWidget) layout(ctx *core.Context, gtx layout.Context, style progressBarResolvedStyle, progress float32) layout.Dimensions {
	output := p.outputText()
	hasHeader := p.label != "" || output != ""
	if !hasHeader {
		return p.layoutTrack(ctx, gtx, style.track, style.fill, progress, style.metrics, style.motion)
	}
	return layout.Flex{
		Axis: layout.Vertical,
		Gap:  gtx.Dp(style.metrics.HeaderGap),
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return p.layoutHeader(ctx, gtx, style.label, output)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return p.layoutTrack(ctx, gtx, style.track, style.fill, progress, style.metrics, style.motion)
		}),
	)
}

func (p BarWidget) layoutHeader(ctx *core.Context, gtx layout.Context, resolved style.ResolvedStyle, output string) layout.Dimensions {
	children := make([]layout.FlexChild, 0, 3)
	if p.label != "" {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layoutui.Resolved(ctx, gtx, resolved, text.New(p.label))
		}))
	}
	if output != "" {
		children = append(children, layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return layout.Dimensions{Size: image.Pt(gtx.Constraints.Max.X, 0)}
		}))
	}
	if output != "" {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layoutui.Resolved(ctx, gtx, resolved, text.New(output))
		}))
	}
	return layout.Flex{
		Axis:      layout.Horizontal,
		Alignment: layout.Middle,
	}.Layout(gtx, children...)
}

func (p BarWidget) layoutTrack(ctx *core.Context, gtx layout.Context, track, fill style.ResolvedStyle, progress float32, metrics theme.ProgressBarMetrics, motion theme.MotionTheme) layout.Dimensions {
	period := time.Duration(0)
	if !p.disabled {
		period = theme.ResolveMotionDuration(motion, metrics.IndeterminatePeriod)
	}
	return layoutui.Resolved(ctx, gtx, track, core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		size := gtx.Constraints.Constrain(gtx.Constraints.Min)
		p.layoutFill(ctx, gtx, size, fill, progress, period, metrics.IndeterminateFillRate)
		return layout.Dimensions{Size: size}
	}))
}

func (p BarWidget) layoutFill(ctx *core.Context, gtx layout.Context, size image.Point, resolved style.ResolvedStyle, progress float32, period time.Duration, fillRate float32) {
	if size.X <= 0 || size.Y <= 0 {
		return
	}
	width, x := 0, 0
	if p.indeterminate {
		if fillRate <= 0 || fillRate >= 1 {
			return
		}
		width = max(int(float32(size.X)*fillRate+0.5), 1)
		x = progressBarIndeterminatePosition(gtx.Now, width, period)
		if period > 0 {
			gtx.Execute(op.InvalidateCmd{})
		}
	} else if progress > 0 {
		width = max(int(float32(size.X)*min(max(progress, 0), 1)+0.5), 1)
		width = min(width, size.X)
	}
	if width <= 0 {
		return
	}
	fillGtx := gtx
	fillGtx.Constraints = layout.Exact(image.Pt(width, size.Y))
	stack := op.Offset(image.Pt(x, 0)).Push(gtx.Ops)
	layoutui.Resolved(ctx, fillGtx, resolved, nil)
	stack.Pop()
}
