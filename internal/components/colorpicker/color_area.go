package colorpicker

import (
	"image"
	"image/color"

	"gioui.org/layout"
	"gioui.org/op/paint"
	layoutui "github.com/qianniancn/flowui/internal/components/layout"
	"github.com/qianniancn/flowui/internal/frame"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/state"
	flowstyle "github.com/qianniancn/flowui/internal/style"
)

const stateSlotColorArea = "color-area"

type ColorAreaWidget struct {
	key         string
	value       color.NRGBA
	label       string
	onChange    func(color.NRGBA)
	disabled    bool
	showDots    bool
	color       *colorValueState
	customStyle flowstyle.Style
}

type colorAreaState struct {
	control colorControlState
	color   colorValueState
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

func (area ColorAreaWidget) Style(value flowstyle.Style) ColorAreaWidget {
	area.customStyle = value
	return area
}

func (area ColorAreaWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	key := frame.ClaimKey(ctx, state.KindColorArea, area.key)
	areaState := frame.UseState[colorAreaState](ctx, key, stateSlotColorArea)
	valueState := &areaState.color
	if area.color != nil {
		valueState = area.color
	} else {
		valueState.sync(area.value)
	}

	enabled := gtx.Enabled() && !area.disabled
	tokens := frame.ActiveTheme(ctx).Components.ColorArea
	side := min(gtx.Dp(tokens.Size), min(gtx.Constraints.Max.X, gtx.Constraints.Max.Y))
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
	return layoutui.LayoutStyled(ctx, gtx, key, flowstyle.StyleState{
		Pressed:      areaState.control.dragging,
		Focused:      focused,
		FocusVisible: frame.FocusVisible(ctx, &areaState.control, focused),
		Disabled:     !enabled,
		Dragging:     areaState.control.dragging,
	}, area.customStyle, frame.WidgetFunc(func(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
		opacity := paint.PushOpacity(gtx.Ops, func() float32 {
			if enabled {
				return 1
			}
			return frame.ActiveTheme(ctx).DisabledOpacityValue()
		}())
		drawColorArea(
			gtx,
			size,
			current,
			gtx.Dp(tokens.Radius),
			func() int {
				if areaState.control.dragging {
					return max(gtx.Dp(tokens.DraggingThumbSize), 1)
				}
				return max(gtx.Dp(tokens.ThumbSize), 1)
			}(),
			max(gtx.Dp(tokens.ThumbBorderWidth), 1),
			max(gtx.Dp(tokens.FocusRingWidth), 1),
			areaState.control.focusOpacity(ctx, gtx),
			frame.ActiveTheme(ctx).Palette.Focus,
			area.showDots,
			max(gtx.Dp(tokens.DotSize), 1),
			max(gtx.Dp(tokens.DotGap), 1),
		)
		opacity.Pop()
		addColorControlInput(gtx, &areaState.control, size, enabled, true, area.semanticLabel(ctx), formatHexColor(area.value, area.value.A != 255))
		return layout.Dimensions{Size: size}
	}))
}

func (area ColorAreaWidget) semanticLabel(ctx *frame.Context) string {
	if area.label != "" {
		return area.label
	}
	if frame.ActiveLanguage(ctx) == locale.LanguageChinese {
		return "颜色区域"
	}
	return "Color area"
}
