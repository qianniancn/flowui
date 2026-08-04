package colorpicker

import (
	"image"
	"image/color"
	"math"

	"gioui.org/f32"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/internal/animation"
	"github.com/qianniancn/flowui/internal/components/disclosure"
	"github.com/qianniancn/flowui/internal/frame"
	"github.com/qianniancn/flowui/internal/interact"
	"github.com/qianniancn/flowui/internal/overlay"
	"github.com/qianniancn/flowui/internal/state"
	"github.com/qianniancn/flowui/internal/theme"
)

const stateSlotColorPicker = "color-picker"

func colorPickerStateFor(ctx *frame.Context, key string) (string, *colorPickerState) {
	key = frame.ClaimKey(ctx, state.KindColorPicker, key)
	return key, frame.UseState[colorPickerState](ctx, key, stateSlotColorPicker)
}

type colorPickerState struct {
	disclosure        disclosure.Binding[color.NRGBA]
	trigger           widget.Clickable
	triggerFocus      state.FocusAnimation
	dialog            overlay.ClickArea
	dismiss           [16]overlay.ClickArea
	open              bool
	popoverTransition animation.FloatTransition
	panelList         layout.List
	color             colorValueState
	history           []color.NRGBA
}

func (pickerState *colorPickerState) pushHistory(value color.NRGBA, limit int) {
	if limit <= 0 {
		return
	}
	// Skip duplicate of most recent.
	if len(pickerState.history) > 0 && pickerState.history[0] == value {
		return
	}
	next := make([]color.NRGBA, 0, limit+1)
	next = append(next, value)
	for _, item := range pickerState.history {
		if item == value {
			continue
		}
		next = append(next, item)
		if len(next) >= limit {
			break
		}
	}
	pickerState.history = next
}

type colorControlState struct {
	dragging bool
	pointer  pointer.ID
	focus    state.FocusAnimation
}

func (pickerState *colorPickerState) popoverProgress(gtx layout.Context, open bool, motions ...theme.MotionTheme) float32 {
	target := float32(0)
	duration := colorPickerExitDuration
	if open {
		target = 1
		duration = colorPickerEnterDuration
	}
	return pickerState.popoverTransition.Value(gtx, target, duration, animation.EaseSmoothstep, motions...)
}

func (pickerState *colorPickerState) handleOverlayEvents(ctx *frame.Context, gtx layout.Context, historyLimit int, recordHistory bool) {
	for pickerState.dialog.Clicked(gtx) {
	}
	if pickerState.dialog.TakePressed() {
		frame.PreserveFocus(ctx)
	}
	for i := range pickerState.dismiss {
		closed := false
		for pickerState.dismiss[i].Clicked(gtx) {
			closed = true
		}
		if pickerState.dismiss[i].TakePressed() {
			closed = true
		}
		if closed {
			if recordHistory {
				pickerState.pushHistory(pickerState.color.syncedColor, historyLimit)
			}
			pickerState.open = false
		}
	}
}

func (pickerState *colorPickerState) escapePressed(gtx layout.Context) bool {
	for {
		value, ok := gtx.Event(key.Filter{Name: key.NameEscape})
		if !ok {
			return false
		}
		if eventValue, ok := value.(key.Event); ok && eventValue.State == key.Press {
			return true
		}
	}
}

func (control *colorControlState) updateArea(ctx *frame.Context, gtx layout.Context, size image.Point, current hsvColor, enabled bool) (hsvColor, bool) {
	next := current
	changed := false
	if !enabled {
		control.dragging = false
	}
	for {
		eventValue, ok := interact.NextPointerEvent(gtx, control, pointer.Press|pointer.Drag|pointer.Release|pointer.Cancel)
		if !ok {
			break
		}
		if !enabled {
			continue
		}
		switch eventValue.Kind {
		case pointer.Press:
			if !interact.IsPrimaryPointerPress(eventValue) {
				continue
			}
			frame.PreserveFocus(ctx)
			control.dragging = true
			control.pointer = eventValue.PointerID
			frame.RequestFocusVisible(ctx, control, false)
			interact.GrabPointer(gtx, control, eventValue)
			next.s, next.v = colorAreaPosition(eventValue.Position, size)
			changed = true
		case pointer.Drag:
			if control.dragging && eventValue.PointerID == control.pointer {
				next.s, next.v = colorAreaPosition(eventValue.Position, size)
				changed = true
			}
		case pointer.Release:
			if eventValue.PointerID == control.pointer {
				next.s, next.v = colorAreaPosition(eventValue.Position, size)
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
		)
		if !ok {
			break
		}
		eventValue, ok := value.(key.Event)
		if !ok || eventValue.State != key.Press || !enabled {
			continue
		}
		step := .01
		if eventValue.Modifiers&key.ModShift != 0 {
			step = .1
		}
		switch eventValue.Name {
		case key.NameLeftArrow:
			next.s -= step
		case key.NameRightArrow:
			next.s += step
		case key.NameUpArrow:
			next.v += step
		case key.NameDownArrow:
			next.v -= step
		}
		next.s = clampUnit(next.s)
		next.v = clampUnit(next.v)
		changed = true
	}
	return next, changed
}

func (control *colorControlState) updateAxis(ctx *frame.Context, gtx layout.Context, size image.Point, current, keyboardStep float64, enabled bool) (float64, bool) {
	next := current
	changed := false
	if !enabled {
		control.dragging = false
	}
	for {
		eventValue, ok := interact.NextPointerEvent(gtx, control, pointer.Press|pointer.Drag|pointer.Release|pointer.Cancel)
		if !ok {
			break
		}
		if !enabled {
			continue
		}
		switch eventValue.Kind {
		case pointer.Press:
			if !interact.IsPrimaryPointerPress(eventValue) {
				continue
			}
			frame.PreserveFocus(ctx)
			control.dragging = true
			control.pointer = eventValue.PointerID
			frame.RequestFocusVisible(ctx, control, false)
			interact.GrabPointer(gtx, control, eventValue)
			next = colorAxisPosition(eventValue.Position, size)
			changed = true
		case pointer.Drag:
			if control.dragging && eventValue.PointerID == control.pointer {
				next = colorAxisPosition(eventValue.Position, size)
				changed = true
			}
		case pointer.Release:
			if eventValue.PointerID == control.pointer {
				next = colorAxisPosition(eventValue.Position, size)
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
			key.Filter{Focus: control, Name: key.NameDownArrow},
			key.Filter{Focus: control, Name: key.NameUpArrow},
			key.Filter{Focus: control, Name: key.NameHome},
			key.Filter{Focus: control, Name: key.NameEnd},
			key.Filter{Focus: control, Name: key.NamePageDown},
			key.Filter{Focus: control, Name: key.NamePageUp},
		)
		if !ok {
			break
		}
		eventValue, ok := value.(key.Event)
		if !ok || eventValue.State != key.Press || !enabled {
			continue
		}
		step := keyboardStep
		if eventValue.Modifiers&key.ModShift != 0 {
			step *= 10
		}
		switch eventValue.Name {
		case key.NameLeftArrow, key.NameDownArrow:
			next -= step
		case key.NameRightArrow, key.NameUpArrow:
			next += step
		case key.NameHome:
			next = 0
		case key.NameEnd:
			next = 1
		case key.NamePageDown:
			next -= keyboardStep * 10
		case key.NamePageUp:
			next += keyboardStep * 10
		}
		next = clampUnit(next)
		changed = true
	}
	return next, changed
}

func (control *colorControlState) focusOpacity(ctx *frame.Context, gtx layout.Context) float32 {
	visible := frame.FocusVisible(ctx, control, gtx.Focused(control))
	return control.focus.Opacity(gtx, visible, frame.ActiveTheme(ctx).Motion)
}

func colorAreaPosition(position f32.Point, size image.Point) (float64, float64) {
	width := max(size.X-1, 1)
	height := max(size.Y-1, 1)
	return clampUnit(float64(position.X) / float64(width)), clampUnit(1 - float64(position.Y)/float64(height))
}

func colorAxisPosition(position f32.Point, size image.Point) float64 {
	edge := float64(min(size.X, size.Y)) / 2
	length := max(float64(size.X)-edge*2, 1)
	return clampUnit((float64(position.X) - edge) / length)
}

type hsvColor struct {
	h float64
	s float64
	v float64
	a float64
}

func nrgbaToHSV(value color.NRGBA) hsvColor {
	r := float64(value.R) / 255
	g := float64(value.G) / 255
	b := float64(value.B) / 255
	maximum := max(r, g, b)
	minimum := min(r, g, b)
	delta := maximum - minimum
	hue := float64(0)
	if delta != 0 {
		switch maximum {
		case r:
			hue = math.Mod((g-b)/delta, 6)
		case g:
			hue = (b-r)/delta + 2
		default:
			hue = (r-g)/delta + 4
		}
		hue /= 6
		if hue < 0 {
			hue++
		}
	}
	saturation := float64(0)
	if maximum != 0 {
		saturation = delta / maximum
	}
	return hsvColor{h: hue, s: saturation, v: maximum, a: float64(value.A) / 255}
}

func hsvToNRGBA(value hsvColor) color.NRGBA {
	hue := value.h - math.Floor(value.h)
	saturation := clampUnit(value.s)
	brightness := clampUnit(value.v)
	chroma := brightness * saturation
	section := hue * 6
	x := chroma * (1 - math.Abs(math.Mod(section, 2)-1))
	var r, g, b float64
	switch int(section) % 6 {
	case 0:
		r, g = chroma, x
	case 1:
		r, g = x, chroma
	case 2:
		g, b = chroma, x
	case 3:
		g, b = x, chroma
	case 4:
		r, b = x, chroma
	default:
		r, b = chroma, x
	}
	match := brightness - chroma
	channel := func(value float64) uint8 {
		return uint8(clampUnit(value+match)*255 + .5)
	}
	return color.NRGBA{R: channel(r), G: channel(g), B: channel(b), A: uint8(clampUnit(value.a)*255 + .5)}
}

func clampUnit(value float64) float64 {
	if math.IsNaN(value) {
		return 0
	}
	return min(max(value, 0), 1)
}

// Disclosure helpers
func colorPickerDisclosureCfg(widget ColorPickerWidget) disclosure.Config[color.NRGBA] {
	return disclosure.Config[color.NRGBA]{
		Controlled: widget.hasValue,
		Value:      widget.value,
		HasDefault: widget.hasDefault,
		Default:    widget.defaultValue,
		OnChange:   widget.onChange,
	}
}

func (s *colorPickerState) currentValue(widget ColorPickerWidget) color.NRGBA {
	return s.disclosure.Current(colorPickerDisclosureCfg(widget))
}

func (s *colorPickerState) bind(widget ColorPickerWidget) {
	s.disclosure.Bind(colorPickerDisclosureCfg(widget))
}

func (s *colorPickerState) requestValue(widget ColorPickerWidget, value color.NRGBA) color.NRGBA {
	newValue, _ := s.disclosure.Request(colorPickerDisclosureCfg(widget), value)
	return newValue
}
