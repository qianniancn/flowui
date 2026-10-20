package colorpicker

import (
	"image"
	"image/color"

	"gioui.org/layout"
	"gioui.org/op/paint"
	"github.com/qianniancn/flowui/i18n"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

const stateSlotColorArea = "color-area"

type ColorAreaWidget struct {
	override *theme.ColorPickerOverride
	key      string
	value    color.NRGBA
	label    string
	onChange func(color.NRGBA)
	disabled bool
	showDots bool
	color    *colorValueState
}

type colorAreaState struct {
	control     colorControlState
	color       colorValueState
	visualCache colorPickerVisualCache
}

func ColorArea(key string, value color.NRGBA) ColorAreaWidget {
	return ColorAreaWidget{key: key, value: value}
}

func (area ColorAreaWidget) Label(label string) ColorAreaWidget {
	area.label = label
	return area
}

func (area ColorAreaWidget) OnChange(fn func(color.NRGBA)) ColorAreaWidget {
	area.onChange = fn
	return area
}

func (area ColorAreaWidget) Disabled(disabled bool) ColorAreaWidget {
	area.disabled = disabled
	return area
}

func (area ColorAreaWidget) ShowDots(show bool) ColorAreaWidget {
	area.showDots = show
	return area
}

func (area ColorAreaWidget) withColorState(state *colorValueState) ColorAreaWidget {
	area.color = state
	return area
}

func (area ColorAreaWidget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	return area.layoutWithOverride(ctx, gtx, area.override)
}

func (area ColorAreaWidget) layoutWithOverride(ctx *core.Context, gtx layout.Context, override *theme.ColorPickerOverride) layout.Dimensions {
	key := core.ClaimKey(ctx, core.KindColorArea, area.key)
	areaState := core.UseState[colorAreaState](ctx, key, stateSlotColorArea)
	visual := areaState.visualCache.resolve(core.ActiveTheme(ctx), colorPickerRequest(override))
	valueState := &areaState.color
	if area.color != nil {
		valueState = area.color
	} else {
		valueState.sync(area.value)
	}

	enabled := gtx.Enabled() && !area.disabled
	side := min(gtx.Dp(visual.Area.Size), min(gtx.Constraints.Max.X, gtx.Constraints.Max.Y))
	size := image.Pt(max(side, 0), max(side, 0))
	areaGtx := gtx
	areaGtx.Constraints = layout.Exact(size)
	current := valueState.hsv()
	if next, changed := areaState.control.updateArea(ctx, areaGtx, size, current, enabled); changed {
		nextColor := hsvToNRGBA(next)
		valueState.accept(nextColor, next.h)
		if area.onChange != nil && nextColor != area.value {
			area.onChange(nextColor)
		}
		current = next
	}

	focused := gtx.Focused(&areaState.control)
	return layoutui.Styled(ctx, gtx, key, style.State{
		Pressed:      areaState.control.dragging,
		Focused:      focused,
		FocusVisible: core.FocusVisible(ctx, &areaState.control, focused),
		Disabled:     !enabled,
		Dragging:     areaState.control.dragging,
	}, visual.Override, core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		opacity := paint.PushOpacity(gtx.Ops, func() float32 {
			if enabled {
				return 1
			}
			return visual.DisabledOpacity
		}())
		visual.Area.Render(gtx, theme.ColorAreaRenderRequest{
			Size:     size,
			Value:    colorPickerHSV(current),
			Dragging: areaState.control.dragging,
			Focus:    areaState.control.focusOpacity(ctx, gtx, visual.Motion),
			ShowDots: area.showDots,
		})
		opacity.Pop()
		addColorControlInput(gtx, &areaState.control, size, enabled, true, area.semanticLabel(ctx), formatHexColor(area.value, area.value.A != 255))
		return layout.Dimensions{Size: size}
	}))
}

func (area ColorAreaWidget) semanticLabel(ctx *core.Context) string {
	if area.label != "" {
		return area.label
	}
	return ctx.Text(i18n.ColorPickerArea)
}

// Style merges instance declarations into the widget. The declaration
// reaches the provider at layout time, so Style can sit anywhere in the chain.
func (w ColorAreaWidget) Style(values ...style.Style) ColorAreaWidget {
	declaration := style.Join(values...)
	if w.override == nil {
		w.override = &theme.ColorPickerOverride{}
	}
	w.override.Declaration = style.Join(w.override.Declaration, declaration)
	return w
}
