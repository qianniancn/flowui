package colorpicker

import (
	"image"
	"image/color"
	"math"

	"gioui.org/f32"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/i18n"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

const stateSlotColorWheel = "color-wheel"

type ColorWheelWidget struct {
	override *theme.ColorPickerOverride
	key      string
	value    color.NRGBA
	label    string
	size     int
	hasSize  bool
	onChange func(color.NRGBA)
	disabled bool
}

type colorWheelState struct {
	control     colorControlState
	color       colorValueState
	visualCache colorPickerVisualCache
}

func ColorWheel(key string, value color.NRGBA) ColorWheelWidget {
	return ColorWheelWidget{key: key, value: value}
}

func (wheel ColorWheelWidget) Size(size int) ColorWheelWidget {
	wheel.size = max(size, 0)
	wheel.hasSize = true
	return wheel
}

func (wheel ColorWheelWidget) Label(label string) ColorWheelWidget {
	wheel.label = label
	return wheel
}

func (wheel ColorWheelWidget) OnChange(fn func(color.NRGBA)) ColorWheelWidget {
	wheel.onChange = fn
	return wheel
}

func (wheel ColorWheelWidget) Disabled(disabled bool) ColorWheelWidget {
	wheel.disabled = disabled
	return wheel
}

func (wheel ColorWheelWidget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	return wheel.layoutWithOverride(ctx, gtx, wheel.override)
}

func (wheel ColorWheelWidget) layoutWithOverride(ctx *core.Context, gtx layout.Context, override *theme.ColorPickerOverride) layout.Dimensions {
	key := core.ClaimKey(ctx, core.KindColorWheel, wheel.key)
	wheelState := core.UseState[colorWheelState](ctx, key, stateSlotColorWheel)
	visual := wheelState.visualCache.resolve(core.ActiveTheme(ctx), colorPickerRequest(override))
	wheelState.color.sync(wheel.value)

	enabled := gtx.Enabled() && !wheel.disabled
	diameter := visual.Wheel.Size
	if wheel.hasSize {
		diameter = unit.Dp(wheel.size)
	}
	side := min(gtx.Dp(diameter), min(gtx.Constraints.Max.X, gtx.Constraints.Max.Y))
	size := image.Pt(max(side, 0), max(side, 0))
	wheelGtx := gtx
	wheelGtx.Constraints = layout.Exact(size)
	current := wheelState.color.hsv()
	if next, changed := wheelState.control.updateWheel(ctx, wheelGtx, size, current, enabled); changed {
		nextColor := hsvToNRGBA(next)
		wheelState.color.accept(nextColor, next.h)
		if wheel.onChange != nil && nextColor != wheel.value {
			wheel.onChange(nextColor)
		}
		current = next
	}

	focused := gtx.Focused(&wheelState.control)
	return layoutui.Styled(ctx, gtx, key, style.State{
		Pressed:      wheelState.control.dragging,
		Focused:      focused,
		FocusVisible: core.FocusVisible(ctx, &wheelState.control, focused),
		Disabled:     !enabled,
		Dragging:     wheelState.control.dragging,
	}, visual.Override, core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		opacity := paint.PushOpacity(gtx.Ops, func() float32 {
			if enabled {
				return 1
			}
			return visual.DisabledOpacity
		}())
		visual.Wheel.Render(gtx, theme.ColorWheelRenderRequest{
			Size:  size,
			Value: colorPickerHSV(current),
			Focus: wheelState.control.focusOpacity(ctx, gtx, visual.Motion),
		})
		opacity.Pop()
		addColorControlInput(gtx, &wheelState.control, size, enabled, true, wheel.semanticLabel(ctx), formatHexColor(wheel.value, wheel.value.A != 255))
		return layout.Dimensions{Size: size}
	}))
}

func (wheel ColorWheelWidget) semanticLabel(ctx *core.Context) string {
	if wheel.label != "" {
		return wheel.label
	}
	return ctx.Text(i18n.ColorPickerWheel)
}

func (control *colorControlState) updateWheel(ctx *core.Context, gtx layout.Context, size image.Point, current hsvColor, enabled bool) (hsvColor, bool) {
	next := current
	changed := false
	if !enabled {
		control.dragging = false
	}
	for {
		eventValue, ok := core.NextPointerEvent(gtx, control, pointer.Press|pointer.Drag|pointer.Release|pointer.Cancel)
		if !ok {
			break
		}
		if !enabled {
			continue
		}
		switch eventValue.Kind {
		case pointer.Press:
			if !core.IsPrimaryPointerPress(eventValue) {
				continue
			}
			core.PreserveFocus(ctx)
			control.dragging = true
			control.pointer = eventValue.PointerID
			core.RequestFocusVisible(ctx, control, false)
			core.GrabPointer(gtx, control, eventValue)
			next = colorWheelPosition(eventValue.Position, size, next)
			changed = true
		case pointer.Drag:
			if control.dragging && eventValue.PointerID == control.pointer {
				next = colorWheelPosition(eventValue.Position, size, next)
				changed = true
			}
		case pointer.Release:
			if eventValue.PointerID == control.pointer {
				next = colorWheelPosition(eventValue.Position, size, next)
				control.dragging = false
				changed = true
			}
		case pointer.Cancel:
			if eventValue.PointerID == control.pointer {
				control.dragging = false
			}
		}
	}

	for {
		value, ok := gtx.Event(
			key.Filter{Focus: control, Name: key.NameLeftArrow},
			key.Filter{Focus: control, Name: key.NameRightArrow},
			key.Filter{Focus: control, Name: key.NameUpArrow},
			key.Filter{Focus: control, Name: key.NameDownArrow},
			key.Filter{Focus: control, Name: key.NameHome},
			key.Filter{Focus: control, Name: key.NameEnd},
		)
		if !ok {
			break
		}
		eventValue, ok := value.(key.Event)
		if !ok || eventValue.State != key.Press || !enabled {
			continue
		}
		hueStep := 1.0 / 360
		saturationStep := .01
		if eventValue.Modifiers&key.ModShift != 0 {
			hueStep *= 10
			saturationStep *= 10
		}
		switch eventValue.Name {
		case key.NameLeftArrow:
			next.h = wrapUnit(next.h - hueStep)
		case key.NameRightArrow:
			next.h = wrapUnit(next.h + hueStep)
		case key.NameUpArrow:
			next.s = clampUnit(next.s + saturationStep)
		case key.NameDownArrow:
			next.s = clampUnit(next.s - saturationStep)
		case key.NameHome:
			next.s = 0
		case key.NameEnd:
			next.s = 1
		}
		changed = true
	}
	return next, changed
}

func colorWheelPosition(position f32.Point, size image.Point, current hsvColor) hsvColor {
	radius := float64(min(size.X, size.Y)) / 2
	if radius <= 0 {
		return current
	}
	center := f32.Pt(float32(size.X)/2, float32(size.Y)/2)
	dx := float64(position.X - center.X)
	dy := float64(position.Y - center.Y)
	distance := math.Hypot(dx, dy)
	current.s = clampUnit(distance / radius)
	if distance > 0 {
		current.h = wrapUnit(math.Atan2(dy, dx) / (2 * math.Pi))
	}
	return current
}

func wrapUnit(value float64) float64 {
	value -= math.Floor(value)
	if value < 0 {
		value++
	}
	return value
}

// Style merges instance declarations into the widget. The declaration
// reaches the provider at layout time, so Style can sit anywhere in the chain.
func (w ColorWheelWidget) Style(values ...style.Style) ColorWheelWidget {
	declaration := style.Join(values...)
	if w.override == nil {
		w.override = &theme.ColorPickerOverride{}
	}
	w.override.Declaration = style.Join(w.override.Declaration, declaration)
	return w
}
