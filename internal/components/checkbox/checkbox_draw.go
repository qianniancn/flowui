package checkbox

import (
	"image"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	layoutui "github.com/qianniancn/FlowUI/internal/components/layout"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/render"
	flowstyle "github.com/qianniancn/FlowUI/internal/style"
	styleruntime "github.com/qianniancn/FlowUI/internal/style/runtime"
	"github.com/qianniancn/FlowUI/internal/theme"
)

// ControlOptions describes the shared Checkbox control used by Checkbox and Table.
type ControlOptions struct {
	Variant         CheckboxVariant
	Selection       float32
	Indeterminate   bool
	Hovered         bool
	Pressed         bool
	Focused         float32
	Disabled        bool
	Invalid         bool
	CustomIndicator bool
	Indicator       frame.Widget
	resolvedStyle   *checkboxIndicatorStyle
}

// DrawControl renders a HeroUI-aligned checkbox control without owning interaction state.
func DrawControl(ctx *frame.Context, gtx layout.Context, options ControlOptions) layout.Dimensions {
	activeTheme := frame.ActiveTheme(ctx)
	style := resolveControlStyle(ctx, activeTheme, options)
	if options.resolvedStyle != nil {
		style = *options.resolvedStyle
	}
	return drawCheckbox(ctx, gtx, activeTheme, style, min(max(options.Selection, 0), 1), min(max(options.Focused, 0), 1), options.Indeterminate, options.CustomIndicator, options.Indicator)
}

func resolveControlStyle(ctx *frame.Context, activeTheme *theme.Theme, options ControlOptions) checkboxIndicatorStyle {
	state := flowstyle.StyleState{
		Hovered: options.Hovered, Pressed: options.Pressed,
		FocusVisible: options.Focused > 0, Disabled: options.Disabled,
		Selected: options.Selection > 0, Checked: options.Selection > 0,
		Indeterminate: options.Indeterminate, Invalid: options.Invalid,
	}
	visual := checkboxStyleFor(activeTheme, options.Variant, state.Hovered, state.Pressed, state.Disabled, state.Invalid)
	defaults := checkboxStyleDeclaration(activeTheme, visual, state)
	offState := state
	offState.Checked, offState.Selected, offState.Indeterminate = false, false, false
	onState := state
	onState.Checked, onState.Selected = true, true
	return checkboxIndicatorStyle{
		off: styleruntime.ResolvePartStatic(ctx, flowstyle.PartIndicator, offState, defaults, flowstyle.Style{}, flowstyle.Style{}, flowstyle.Style{}),
		on:  styleruntime.ResolvePartStatic(ctx, flowstyle.PartIndicator, onState, defaults, flowstyle.Style{}, flowstyle.Style{}, flowstyle.Style{}),
	}
}

func drawCheckbox(ctx *frame.Context, gtx layout.Context, activeTheme *theme.Theme, style checkboxIndicatorStyle, selected, focused float32, indeterminate, customIndicator bool, indicator frame.Widget) layout.Dimensions {
	controlSize := checkboxIndicatorSize(gtx, style, activeTheme.Components.Checkbox.Size)
	focusSpace := max(gtx.Dp(activeTheme.Components.Checkbox.FocusSpace), 1)
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
	content := frame.WidgetFunc(func(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
		local := image.Rectangle{Max: gtx.Constraints.Min}
		if customIndicator {
			layoutCheckboxIndicator(ctx, gtx, activeTheme, local, indicator)
		} else if indeterminate {
			drawCheckboxIndeterminate(ctx, gtx, activeTheme, local, selected)
		} else {
			drawCheckboxCheck(ctx, gtx, activeTheme, local, selected)
		}
		return layout.Dimensions{Size: gtx.Constraints.Min}
	})
	layoutCheckboxLayer(ctx, gtx, rect, style.off, 1-selected, focused, nil)
	layoutCheckboxLayer(ctx, gtx, rect, style.on, selected, focused, content)

	return layout.Dimensions{Size: dims}
}

func checkboxIndicatorSize(gtx layout.Context, style checkboxIndicatorStyle, fallbackDp unit.Dp) image.Point {
	var size image.Point
	var hasWidth, hasHeight bool
	for _, endpoint := range [...]flowstyle.ResolvedStyle{style.off, style.on} {
		if endpoint.Box == nil {
			continue
		}
		if endpoint.Box.Width != nil {
			size.X = max(size.X, gtx.Dp(*endpoint.Box.Width))
			hasWidth = true
		}
		if endpoint.Box.Height != nil {
			size.Y = max(size.Y, gtx.Dp(*endpoint.Box.Height))
			hasHeight = true
		}
	}
	if !hasWidth {
		size.X = gtx.Dp(fallbackDp)
	}
	if !hasHeight {
		size.Y = gtx.Dp(fallbackDp)
	}
	return size
}

func layoutCheckboxLayer(ctx *frame.Context, gtx layout.Context, rect image.Rectangle, style flowstyle.ResolvedStyle, opacity, focused float32, child frame.Widget) {
	layerGtx := gtx
	layerGtx.Constraints = layout.Exact(rect.Size())
	stack := op.Offset(rect.Min).Push(gtx.Ops)
	fade := paint.PushOpacity(gtx.Ops, opacity)
	layoutui.LayoutResolved(ctx, layerGtx, styleruntime.ApplyOutlineOpacity(style, focused), child)
	fade.Pop()
	stack.Pop()
}

func drawCheckboxCheck(ctx *frame.Context, gtx layout.Context, activeTheme *theme.Theme, rect image.Rectangle, selected float32) {
	if selected == 0 {
		return
	}
	width := float32(rect.Dx())
	height := float32(rect.Dy())
	x := float32(rect.Min.X)
	y := float32(rect.Min.Y)
	points := [3]f32.Point{
		f32.Pt(x+width*0.27, y+height*0.52),
		f32.Pt(x+width*0.43, y+height*0.68),
		f32.Pt(x+width*0.75, y+height*0.31),
	}
	path := render.CheckPath(gtx.Ops, points, selected)
	col := ctx.ForegroundColor()
	col.A = byte(float32(col.A)*selected + 0.5)
	stroke := clip.Stroke{
		Path:  path,
		Width: max(render.DpFloat(gtx, activeTheme.Components.Checkbox.CheckStroke), 1),
	}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, col)
	stroke.Pop()
}

func drawCheckboxIndeterminate(ctx *frame.Context, gtx layout.Context, activeTheme *theme.Theme, rect image.Rectangle, selected float32) {
	if selected == 0 {
		return
	}
	length := max(int(float32(rect.Dx())*0.58*selected+0.5), 1)
	thickness := max(gtx.Dp(activeTheme.Components.Checkbox.IndeterminateStroke), 1)
	center := rect.Min.Add(image.Pt(rect.Dx()/2, rect.Dy()/2))
	line := image.Rect(center.X-length/2, center.Y-thickness/2, center.X+(length+1)/2, center.Y+(thickness+1)/2)
	col := ctx.ForegroundColor()
	col.A = byte(float32(col.A)*selected + 0.5)
	paint.FillShape(gtx.Ops, col, clip.UniformRRect(line, max(thickness/2, 1)).Op(gtx.Ops))
}

func layoutCheckboxIndicator(ctx *frame.Context, gtx layout.Context, activeTheme *theme.Theme, rect image.Rectangle, indicator frame.Widget) {
	if indicator == nil {
		return
	}
	diameter := min(gtx.Dp(activeTheme.Components.Checkbox.IndicatorSize), min(rect.Dx(), rect.Dy()))
	if diameter <= 0 {
		return
	}
	indicatorRect := image.Rectangle{Min: image.Pt((rect.Min.X+rect.Max.X-diameter)/2, (rect.Min.Y+rect.Max.Y-diameter)/2)}
	indicatorRect.Max = indicatorRect.Min.Add(image.Pt(diameter, diameter))
	stack := op.Offset(indicatorRect.Min).Push(gtx.Ops)
	indicatorGtx := gtx
	indicatorGtx.Constraints = layout.Exact(indicatorRect.Size())
	clipStack := clip.Rect{Max: indicatorRect.Size()}.Push(gtx.Ops)
	layout.Center.Layout(indicatorGtx, func(gtx layout.Context) layout.Dimensions {
		return indicator.Layout(ctx, gtx)
	})
	clipStack.Pop()
	stack.Pop()
}
