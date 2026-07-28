package colorpicker

import (
	"image"
	"image/color"
	"math"

	"gioui.org/f32"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	layoutui "github.com/qianniancn/flowui/internal/components/layout"
	"github.com/qianniancn/flowui/internal/frame"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/state"
	flowstyle "github.com/qianniancn/flowui/internal/style"
)

const stateSlotColorWheel = "color-wheel"

type ColorWheelWidget struct {
	key         string
	value       color.NRGBA
	label       string
	size        int
	hasSize     bool
	onChange    func(color.NRGBA)
	disabled    bool
	customStyle flowstyle.Style
}

type colorWheelState struct {
	control colorControlState
	color   colorValueState
	cache   colorWheelImageCache
}

type colorWheelImageCache struct {
	size image.Point
	op   paint.ImageOp
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

func (wheel ColorWheelWidget) Style(value flowstyle.Style) ColorWheelWidget {
	wheel.customStyle = value
	return wheel
}

func (wheel ColorWheelWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	key := frame.ClaimKey(ctx, state.KindColorWheel, wheel.key)
	wheelState := frame.UseState[colorWheelState](ctx, key, stateSlotColorWheel)
	wheelState.color.sync(wheel.value)

	enabled := gtx.Enabled() && !wheel.disabled
	tokens := frame.ActiveTheme(ctx).Components.ColorWheel
	diameter := tokens.Size
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
	return layoutui.LayoutStyled(ctx, gtx, key, flowstyle.StyleState{
		Pressed:      wheelState.control.dragging,
		Focused:      focused,
		FocusVisible: frame.FocusVisible(ctx, &wheelState.control, focused),
		Disabled:     !enabled,
		Dragging:     wheelState.control.dragging,
	}, wheel.customStyle, frame.WidgetFunc(func(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
		opacity := paint.PushOpacity(gtx.Ops, func() float32 {
			if enabled {
				return 1
			}
			return frame.ActiveTheme(ctx).DisabledOpacityValue()
		}())
		drawColorWheel(
			gtx,
			size,
			current,
			&wheelState.cache,
			max(gtx.Dp(tokens.ThumbSize), 1),
			max(gtx.Dp(tokens.ThumbBorderWidth), 1),
			max(gtx.Dp(tokens.FocusRingWidth), 1),
			wheelState.control.focusOpacity(ctx, gtx),
			frame.ActiveTheme(ctx).Palette.Focus,
		)
		opacity.Pop()
		addColorControlInput(gtx, &wheelState.control, size, enabled, true, wheel.semanticLabel(ctx), formatHexColor(wheel.value, wheel.value.A != 255))
		return layout.Dimensions{Size: size}
	}))
}

func (wheel ColorWheelWidget) semanticLabel(ctx *frame.Context) string {
	if wheel.label != "" {
		return wheel.label
	}
	if frame.ActiveLanguage(ctx) == locale.LanguageChinese {
		return "色轮"
	}
	return "Color wheel"
}

func (control *colorControlState) updateWheel(ctx *frame.Context, gtx layout.Context, size image.Point, current hsvColor, enabled bool) (hsvColor, bool) {
	next := current
	changed := false
	if !enabled {
		control.dragging = false
	}
	for {
		value, ok := gtx.Event(pointer.Filter{Target: control, Kinds: pointer.Press | pointer.Drag | pointer.Release | pointer.Cancel})
		if !ok {
			break
		}
		eventValue, ok := value.(pointer.Event)
		if !ok || !enabled {
			continue
		}
		switch eventValue.Kind {
		case pointer.Press:
			if eventValue.Source != pointer.Touch && !eventValue.Buttons.Contain(pointer.ButtonPrimary) {
				continue
			}
			frame.PreserveFocus(ctx)
			control.dragging = true
			control.pointer = eventValue.PointerID
			frame.RequestFocusVisible(ctx, control, false)
			gtx.Execute(pointer.GrabCmd{Tag: control, ID: eventValue.PointerID})
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

func colorWheelSelectorCenter(size image.Point, value hsvColor, inset int) image.Point {
	radius := max(float64(min(size.X, size.Y))/2-float64(max(inset, 0)), 0)
	angle := wrapUnit(value.h) * 2 * math.Pi
	centerX := float64(size.X) / 2
	centerY := float64(size.Y) / 2
	distance := radius * clampUnit(value.s)
	return image.Pt(
		int(math.Round(centerX+math.Cos(angle)*distance)),
		int(math.Round(centerY+math.Sin(angle)*distance)),
	)
}

func wrapUnit(value float64) float64 {
	value -= math.Floor(value)
	if value < 0 {
		value++
	}
	return value
}

func (cache *colorWheelImageCache) update(size image.Point) bool {
	if size.X <= 0 || size.Y <= 0 || cache.size == size && cache.op.Size() == size {
		return false
	}
	cache.size = size
	cache.op = paint.NewImageOp(buildColorWheelImage(size))
	return true
}

func drawColorWheel(gtx layout.Context, size image.Point, value hsvColor, cache *colorWheelImageCache, thumbSize, thumbBorder, focusWidth int, focus float32, focusColor color.NRGBA) {
	if size.X <= 0 || size.Y <= 0 {
		return
	}
	cache.update(size)
	clipped := clip.Ellipse(image.Rectangle{Max: size}).Push(gtx.Ops)
	cache.op.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)
	clipped.Pop()
	drawRoundedStroke(gtx, image.Rectangle{Max: size}, min(size.X, size.Y)/2, 1, color.NRGBA{A: 26})
	thumbColor := hsvToNRGBA(hsvColor{h: value.h, s: value.s, v: value.v, a: 1})
	selectorInset := (thumbSize + focusWidth*4 + 1) / 2
	drawColorThumb(gtx, colorWheelSelectorCenter(size, value, selectorInset), thumbSize, thumbBorder, focusWidth, thumbColor, focus, focusColor)
}

func buildColorWheelImage(size image.Point) *image.NRGBA {
	value := image.NewNRGBA(image.Rectangle{Max: size})
	centerX := float64(size.X-1) / 2
	centerY := float64(size.Y-1) / 2
	radius := float64(min(size.X, size.Y)) / 2
	if radius <= 0 {
		return value
	}
	for y := range size.Y {
		for x := range size.X {
			dx := float64(x) - centerX
			dy := float64(y) - centerY
			distance := math.Hypot(dx, dy)
			coverage := min(max(radius+.5-distance, 0), 1)
			if coverage == 0 {
				continue
			}
			pixel := hsvToNRGBA(hsvColor{
				h: wrapUnit(math.Atan2(dy, dx) / (2 * math.Pi)),
				s: clampUnit(distance / radius),
				v: 1,
				a: coverage,
			})
			value.SetNRGBA(x, y, pixel)
		}
	}
	return value
}
