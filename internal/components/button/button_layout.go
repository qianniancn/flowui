package button

import (
	"gioui.org/layout"
	"gioui.org/op"
	layoutui "github.com/qianniancn/flowui/internal/components/layout"
	"github.com/qianniancn/flowui/internal/frame"
	flowstyle "github.com/qianniancn/flowui/internal/style"
	"github.com/qianniancn/flowui/internal/theme"
)

type buttonPreparedContent struct {
	call  op.CallOp
	dims  layout.Dimensions
	width int
}

func (b ButtonWidget) prepareContent(ctx *frame.Context, gtx layout.Context) buttonPreparedContent {
	activeTheme := frame.ActiveTheme(ctx)
	style := b.staticStyle(ctx, flowstyle.StyleState{Disabled: b.disabled, Loading: b.loading})
	if b.loading && !b.iconOnly && !b.fullWidth {
		style = buttonLoadingStyle(gtx, activeTheme, b.size, style)
	}
	measureGtx := gtx
	if b.disabled || b.loading {
		measureGtx = measureGtx.Disabled()
	}
	measureGtx.Constraints.Min.X = 0
	measureGtx.Constraints.Min.Y = 0
	macro := op.Record(gtx.Ops)
	dims := b.layoutContent(ctx, measureGtx, style, activeTheme)
	call := macro.Stop()
	visualMacro := op.Record(gtx.Ops)
	visualDims := layoutui.LayoutResolved(ctx, measureGtx, style.root, frame.WidgetFunc(func(_ *frame.Context, gtx layout.Context) layout.Dimensions {
		return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			call.Add(gtx.Ops)
			return dims
		})
	}))
	_ = visualMacro.Stop()
	return buttonPreparedContent{call: call, dims: dims, width: visualDims.Size.X}
}

func (b ButtonWidget) layoutContent(ctx *frame.Context, gtx layout.Context, style buttonStyle, activeTheme *theme.Theme) layout.Dimensions {
	content := frame.WidgetFunc(func(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
		return layoutui.LayoutResolved(ctx, gtx, style.content, b.child)
	})
	if !b.loading {
		return content.Layout(ctx, gtx)
	}
	spinner := frame.WidgetFunc(func(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
		return layoutui.LayoutResolved(ctx, gtx, style.indicatorPart, frame.WidgetFunc(func(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
			period := theme.ResolveMotionDuration(activeTheme.Motion, buttonSpinnerPeriod)
			return drawButtonSpinner(gtx, buttonSpinnerSize(activeTheme, b.size), activeTheme.Components.Button.SpinnerStrokeWidth, ctx.ForegroundColor(), period)
		}))
	})
	if b.iconOnly {
		return spinner.Layout(ctx, gtx)
	}
	return layout.Flex{
		Axis:      layout.Horizontal,
		Alignment: layout.Middle,
		Gap:       gtx.Dp(activeTheme.Components.Button.ContentGap),
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return spinner.Layout(ctx, gtx)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return content.Layout(ctx, gtx)
		}),
	)
}
