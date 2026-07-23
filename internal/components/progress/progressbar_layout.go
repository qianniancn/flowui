package progress

import (
	"image"
	"time"

	"gioui.org/layout"
	"gioui.org/op"
	layoutui "github.com/qianniancn/FlowUI/internal/components/layout"
	"github.com/qianniancn/FlowUI/internal/components/text"
	"github.com/qianniancn/FlowUI/internal/frame"
	flowstyle "github.com/qianniancn/FlowUI/internal/style"
	"github.com/qianniancn/FlowUI/internal/theme"
)

func (p ProgressBarWidget) layout(ctx *frame.Context, gtx layout.Context, style progressBarResolvedStyle, progress float32) layout.Dimensions {
	output := p.outputText()
	hasHeader := p.label != "" || output != ""
	if !hasHeader {
		return p.layoutTrack(ctx, gtx, style.track, style.fill, progress)
	}
	return layout.Flex{
		Axis: layout.Vertical,
		Gap:  gtx.Dp(frame.ActiveTheme(ctx).Components.ProgressBar.HeaderGap),
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return p.layoutHeader(ctx, gtx, style.label, output)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return p.layoutTrack(ctx, gtx, style.track, style.fill, progress)
		}),
	)
}

func (p ProgressBarWidget) layoutHeader(ctx *frame.Context, gtx layout.Context, style flowstyle.ResolvedStyle, output string) layout.Dimensions {
	children := make([]layout.FlexChild, 0, 3)
	if p.label != "" {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layoutui.LayoutResolved(ctx, gtx, style, text.New(p.label))
		}))
	}
	if output != "" {
		children = append(children, layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return layout.Dimensions{Size: image.Pt(gtx.Constraints.Max.X, 0)}
		}))
	}
	if output != "" {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layoutui.LayoutResolved(ctx, gtx, style, text.New(output))
		}))
	}
	return layout.Flex{
		Axis:      layout.Horizontal,
		Alignment: layout.Middle,
	}.Layout(gtx, children...)
}

func (p ProgressBarWidget) layoutTrack(ctx *frame.Context, gtx layout.Context, track, fill flowstyle.ResolvedStyle, progress float32) layout.Dimensions {
	period := time.Duration(0)
	if !p.disabled {
		period = theme.ResolveMotionDuration(frame.ActiveTheme(ctx).Motion, progressBarIndeterminatePeriod)
	}
	return layoutui.LayoutResolved(ctx, gtx, track, frame.WidgetFunc(func(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
		size := gtx.Constraints.Constrain(gtx.Constraints.Min)
		p.layoutFill(ctx, gtx, size, fill, progress, period)
		return layout.Dimensions{Size: size}
	}))
}

func (p ProgressBarWidget) layoutFill(ctx *frame.Context, gtx layout.Context, size image.Point, style flowstyle.ResolvedStyle, progress float32, period time.Duration) {
	if size.X <= 0 || size.Y <= 0 {
		return
	}
	width, x := 0, 0
	if p.indeterminate {
		width = max(int(float32(size.X)*progressBarIndeterminateFillRate+0.5), 1)
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
	layoutui.LayoutResolved(ctx, fillGtx, style, nil)
	stack.Pop()
}
