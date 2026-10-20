package checkbox

import (
	"image"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

// ControlOptions describes the shared Checkbox control used by Checkbox and Table.
type ControlOptions struct {
	Variant         Variant
	Selection       float32
	Indeterminate   bool
	Hovered         bool
	Pressed         bool
	Focused         float32
	Disabled        bool
	Invalid         bool
	CustomIndicator bool
	Indicator       core.Widget
	// VisualCache optionally reuses provider resolution for compound widgets.
	VisualCache    *VisualCache
	resolvedStyle  *checkboxIndicatorStyle
	resolvedVisual *theme.CheckboxVisual
}

// DrawControl renders the framework checkbox control without owning interaction
// state. Its appearance comes from the active theme.
func DrawControl(ctx *core.Context, gtx layout.Context, options ControlOptions) layout.Dimensions {
	core.RequireVisualProvider(ctx)
	activeTheme := core.ActiveTheme(ctx)
	var visual theme.CheckboxVisual
	if options.resolvedVisual != nil {
		visual = *options.resolvedVisual
	} else if options.VisualCache != nil {
		visual = options.VisualCache.Resolve(activeTheme, options.Variant)
	} else {
		visual = checkboxVisual(activeTheme, options.Variant)
	}
	style := resolveControlStyle(ctx, options, visual)
	if options.resolvedStyle != nil {
		style = *options.resolvedStyle
	}
	return drawCheckbox(ctx, gtx, visual, style, min(max(options.Selection, 0), 1), min(max(options.Focused, 0), 1), options.Indeterminate, options.CustomIndicator, options.Indicator)
}

func checkboxVisual(activeTheme *theme.Theme, variant Variant) theme.CheckboxVisual {
	return resolveCheckboxVisual(activeTheme, variant, nil)
}

func resolveControlStyle(ctx *core.Context, options ControlOptions, visual theme.CheckboxVisual) checkboxIndicatorStyle {
	state := style.State{
		Hovered: options.Hovered, Pressed: options.Pressed,
		FocusVisible: options.Focused > 0, Disabled: options.Disabled,
		Selected: options.Selection > 0, Checked: options.Selection > 0,
		Indeterminate: options.Indeterminate, Invalid: options.Invalid,
	}
	defaults := visual.Defaults
	offState := state
	offState.Checked, offState.Selected, offState.Indeterminate = false, false, false
	onState := state
	onState.Checked, onState.Selected = true, true
	return checkboxIndicatorStyle{
		off: computed.ResolvePartStatic(ctx, style.PartIndicator, offState, defaults, visual.Variant, style.Style{}, visual.Override),
		on:  computed.ResolvePartStatic(ctx, style.PartIndicator, onState, defaults, visual.Variant, style.Style{}, visual.Override),
	}
}

func drawCheckbox(ctx *core.Context, gtx layout.Context, visual theme.CheckboxVisual, resolved checkboxIndicatorStyle, selected, focused float32, indeterminate, customIndicator bool, indicator core.Widget) layout.Dimensions {
	metrics := visual.Metrics
	controlSize := checkboxIndicatorSize(gtx, resolved)
	focusSpace := max(gtx.Dp(metrics.FocusSpace), 1)
	controlSize.X = min(controlSize.X, max(gtx.Constraints.Max.X-focusSpace*2, 0))
	controlSize.Y = min(controlSize.Y, max(gtx.Constraints.Max.Y-focusSpace*2, 0))
	bounds := controlSize.Add(image.Pt(focusSpace*2, focusSpace*2))
	dims := gtx.Constraints.Constrain(bounds)
	if controlSize.X <= 0 || controlSize.Y <= 0 {
		return layout.Dimensions{Size: dims}
	}

	origin := image.Pt((dims.X-controlSize.X)/2, (dims.Y-controlSize.Y)/2)
	rect := image.Rectangle{
		Min: origin,
		Max: origin.Add(controlSize),
	}
	content := core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		local := image.Rectangle{Max: gtx.Constraints.Min}
		var custom layout.Widget
		if customIndicator && indicator != nil {
			custom = func(indicatorGtx layout.Context) layout.Dimensions {
				return layoutCheckboxIndicator(ctx, indicatorGtx, metrics, local, indicator)
			}
		}
		if customIndicator {
			if custom != nil {
				return custom(gtx)
			}
			return layout.Dimensions{Size: gtx.Constraints.Min}
		}
		if visual.IndicatorRenderer == nil {
			panic(theme.ErrIncompleteVisualProvider)
		}
		return visual.IndicatorRenderer(gtx, theme.CheckboxIndicatorRenderRequest{
			Foreground:          ctx.ForegroundColor(),
			Selection:           selected,
			Indeterminate:       indeterminate,
			IndeterminateStroke: metrics.IndeterminateStroke,
			CheckStroke:         metrics.CheckStroke,
		})
	})
	layoutCheckboxLayer(ctx, gtx, rect, resolved.off, 1-selected, focused, nil)
	layoutCheckboxLayer(ctx, gtx, rect, resolved.on, selected, focused, content)

	return layout.Dimensions{Size: dims}
}

func checkboxIndicatorSize(gtx layout.Context, resolved checkboxIndicatorStyle) image.Point {
	var size image.Point
	for _, endpoint := range [...]style.ResolvedStyle{resolved.off, resolved.on} {
		if endpoint.Box == nil || endpoint.Box.Width == nil || endpoint.Box.Height == nil {
			panic(theme.ErrIncompleteVisualProvider)
		}
		size.X = max(size.X, gtx.Dp(*endpoint.Box.Width))
		size.Y = max(size.Y, gtx.Dp(*endpoint.Box.Height))
	}
	if size.X <= 0 || size.Y <= 0 {
		panic(theme.ErrIncompleteVisualProvider)
	}
	return size
}

func layoutCheckboxLayer(ctx *core.Context, gtx layout.Context, rect image.Rectangle, resolved style.ResolvedStyle, opacity, focused float32, child core.Widget) {
	layerGtx := gtx
	layerGtx.Constraints = layout.Exact(rect.Size())
	stack := op.Offset(rect.Min).Push(gtx.Ops)
	fade := paint.PushOpacity(gtx.Ops, opacity)
	layoutui.Resolved(ctx, layerGtx, computed.ApplyOutlineOpacity(resolved, focused), child)
	fade.Pop()
	stack.Pop()
}

func layoutCheckboxIndicator(ctx *core.Context, gtx layout.Context, metrics theme.CheckboxMetrics, rect image.Rectangle, indicator core.Widget) layout.Dimensions {
	if indicator == nil {
		return layout.Dimensions{Size: gtx.Constraints.Min}
	}
	diameter := min(gtx.Dp(metrics.IndicatorSize), min(rect.Dx(), rect.Dy()))
	if diameter <= 0 {
		return layout.Dimensions{Size: gtx.Constraints.Min}
	}
	indicatorRect := image.Rectangle{Min: image.Pt((rect.Min.X+rect.Max.X-diameter)/2, (rect.Min.Y+rect.Max.Y-diameter)/2)}
	indicatorRect.Max = indicatorRect.Min.Add(image.Pt(diameter, diameter))
	stack := op.Offset(indicatorRect.Min).Push(gtx.Ops)
	indicatorGtx := gtx
	indicatorGtx.Constraints = layout.Exact(indicatorRect.Size())
	clipStack := clip.Rect{Max: indicatorRect.Size()}.Push(gtx.Ops)
	dims := layout.Center.Layout(indicatorGtx, func(gtx layout.Context) layout.Dimensions {
		return indicator.Layout(ctx, gtx)
	})
	clipStack.Pop()
	stack.Pop()
	return dims
}
