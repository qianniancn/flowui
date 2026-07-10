package progress

import (
	"image"

	"gioui.org/font"
	"gioui.org/layout"
	"github.com/qianniancn/FlowUI/internal/components/text"
	"github.com/qianniancn/FlowUI/internal/frame"
)

func (p ProgressBarWidget) layout(ctx *frame.Context, gtx layout.Context, style progressBarStyle, sizeStyle progressBarSizeStyle, progress float32) layout.Dimensions {
	output := p.outputText()
	hasHeader := p.label != "" || output != ""
	if !hasHeader {
		return p.layoutTrack(gtx, style, sizeStyle, progress)
	}
	return layout.Flex{
		Axis: layout.Vertical,
		Gap:  gtx.Dp(frame.ActiveTheme(ctx).Components.ProgressBar.HeaderGap),
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return p.layoutHeader(ctx, gtx, style, output)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return p.layoutTrack(gtx, style, sizeStyle, progress)
		}),
	)
}

func (p ProgressBarWidget) layoutHeader(ctx *frame.Context, gtx layout.Context, style progressBarStyle, output string) layout.Dimensions {
	children := make([]layout.FlexChild, 0, 3)
	if p.label != "" {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return text.New(p.label).
				Size(float32(frame.ActiveTheme(ctx).Components.ProgressBar.TextSize)).
				Weight(font.Medium).
				Color(style.label).
				Layout(ctx, gtx)
		}))
	}
	if output != "" {
		children = append(children, layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return layout.Dimensions{Size: image.Pt(gtx.Constraints.Max.X, 0)}
		}))
	}
	if output != "" {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return text.New(output).
				Size(float32(frame.ActiveTheme(ctx).Components.ProgressBar.TextSize)).
				Weight(font.Medium).
				Color(style.output).
				Layout(ctx, gtx)
		}))
	}
	return layout.Flex{
		Axis:      layout.Horizontal,
		Alignment: layout.Middle,
	}.Layout(gtx, children...)
}

func (p ProgressBarWidget) layoutTrack(gtx layout.Context, style progressBarStyle, sizeStyle progressBarSizeStyle, progress float32) layout.Dimensions {
	height := min(gtx.Dp(sizeStyle.height), gtx.Constraints.Max.Y)
	if height <= 0 {
		height = gtx.Dp(sizeStyle.height)
	}
	size := gtx.Constraints.Constrain(image.Pt(gtx.Constraints.Max.X, height))
	drawProgressBar(gtx, size, sizeStyle.radius, style, progress, p.indeterminate, !p.disabled)
	return layout.Dimensions{Size: size}
}
