package button

import (
	"image/color"

	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

type buttonPreparedContent struct {
	call         op.CallOp
	dims         layout.Dimensions
	width        int
	declarations buttonDeclarationBundle
}

func (b Widget) prepareContent(ctx *core.Context, gtx layout.Context, overrides ...*theme.ButtonOverride) buttonPreparedContent {
	activeTheme := core.ActiveTheme(ctx)
	key := core.FullKey(ctx, b.key)
	declarations := b.declarationsForOverride(activeTheme, core.UseState[buttonDeclarationCache](ctx, key, stateSlotStyle), b.buttonOverride(overrides))
	style := b.staticContentStyleWithDeclarations(ctx, style.State{Disabled: b.disabled, Loading: b.loading}, declarations)
	if b.loading && !b.iconOnly && !b.fullWidth {
		style = buttonLoadingStyle(gtx, style)
	}
	measureGtx := gtx
	if b.disabled || b.loading {
		measureGtx = measureGtx.Disabled()
	}
	measureGtx.Constraints.Min.X = 0
	measureGtx.Constraints.Min.Y = 0
	macro := op.Record(gtx.Ops)
	dims := b.layoutContent(ctx, measureGtx, style)
	call := macro.Stop()
	visualMacro := op.Record(gtx.Ops)
	visualDims := layoutui.Resolved(ctx, measureGtx, style.root, core.WidgetFunc(func(_ *core.Context, gtx layout.Context) layout.Dimensions {
		return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			call.Add(gtx.Ops)
			return dims
		})
	}))
	_ = visualMacro.Stop()
	return buttonPreparedContent{call: call, dims: dims, width: visualDims.Size.X, declarations: declarations}
}

func (b Widget) layoutContent(ctx *core.Context, gtx layout.Context, resolved buttonStyle) layout.Dimensions {
	content := core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		return layoutui.Resolved(ctx, gtx, resolved.content, b.child)
	})
	if !b.loading {
		return content.Layout(ctx, gtx)
	}
	spinner := core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		return layoutui.Resolved(ctx, gtx, resolved.indicatorPart, core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
			// The indicator part has already gone through the component cascade.
			// Keep the spinner on that resolved color so instance and scoped
			// PartIndicator overrides apply to the loading visual as well.
			indicator := buttonForeground(resolved.visual, resolved.indicatorPart)
			return loadingWithVisual(gtx, resolved.visual, indicator)
		}))
	})
	if b.iconOnly {
		return spinner.Layout(ctx, gtx)
	}
	if resolved.visual.ContentGap == nil {
		panic(theme.ErrIncompleteVisualProvider)
	}
	gap := resolved.visual.ContentGap(theme.ButtonContentGapRequest{IconOnly: b.iconOnly, Loading: b.loading})
	if gap < 0 {
		panic(theme.ErrIncompleteVisualProvider)
	}
	return layout.Flex{
		Axis:      layout.Horizontal,
		Alignment: layout.Middle,
		Gap:       gtx.Dp(gap),
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return spinner.Layout(ctx, gtx)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return content.Layout(ctx, gtx)
		}),
	)
}

func loadingWithVisual(gtx layout.Context, visual theme.ButtonVisual, col color.NRGBA) layout.Dimensions {
	if visual.Spinner == nil || visual.SpinnerSize == nil || visual.SpinnerStrokeWidth == nil || visual.SpinnerPeriod == nil {
		panic(theme.ErrIncompleteVisualProvider)
	}
	spinner := visual.SpinnerSize()
	strokeWidth := visual.SpinnerStrokeWidth()
	period := theme.ResolveMotionDuration(visual.Motion, visual.SpinnerPeriod())
	if spinner <= 0 || strokeWidth <= 0 || period < 0 {
		panic(theme.ErrIncompleteVisualProvider)
	}
	return visual.Spinner(gtx, theme.ButtonSpinnerRequest{
		Size:        spinner,
		StrokeWidth: strokeWidth,
		Color:       col,
		Period:      period,
	})
}
