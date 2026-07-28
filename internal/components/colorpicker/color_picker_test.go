package colorpicker

import (
	"image"
	"image/color"
	"testing"
	"time"

	"gioui.org/f32"
	"gioui.org/io/input"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/locale"
)

func TestColorPickerOptionsUseValueSemantics(t *testing.T) {
	base := ColorPicker("brand", color.NRGBA{R: 1, A: 255})
	presets := []color.NRGBA{{R: 2, A: 255}}
	configured := base.
		Label("Brand color").
		Disabled(true).
		Alpha(true).
		ShowField().
		Presets(presets).
		OnChange(func(color.NRGBA) {})
	presets[0] = color.NRGBA{}

	// Defaults enable desktop chrome (field/RGB/history); configuration must not mutate base.
	if base.label != "" || base.disabled || base.alpha || !base.showField || !base.showRGB || !base.showHistory || len(base.presets) != 0 || base.onChange != nil {
		t.Fatalf("configuring ColorPicker mutated base: %#v", base)
	}
	if configured.label != "Brand color" || !configured.disabled || !configured.alpha || !configured.showField || configured.presets[0].R != 2 || configured.onChange == nil {
		t.Fatalf("configured ColorPicker = %#v", configured)
	}
}

func TestColorPickerHistoryExcludesPresets(t *testing.T) {
	red := color.NRGBA{R: 0xff, A: 0xff}
	green := color.NRGBA{G: 0xff, A: 0xff}
	blue := color.NRGBA{B: 0xff, A: 0xff}

	got := historyWithoutPresets([]color.NRGBA{red, blue, green}, []color.NRGBA{green, red})
	if len(got) != 1 || got[0] != blue {
		t.Fatalf("filtered history = %#v, want only %#v", got, blue)
	}
}

func TestColorPickerPointerTriggerFocusIsNotKeyboardVisible(t *testing.T) {
	ctx := frame.New(nil, nil, locale.LanguageEnglish)
	router := new(input.Router)
	picker := ColorPicker("brand", color.NRGBA{R: 0x04, G: 0x85, B: 0xf7, A: 255}).Label("Accent color")
	start := time.Unix(1, 0)
	layoutColorPickerFrame(ctx, router, picker, start)
	state := colorPickerTestState(ctx, "brand")

	router.Queue(pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: 1, Buttons: pointer.ButtonPrimary, Position: f32.Pt(60, 18)})
	layoutColorPickerFrame(ctx, router, picker, start.Add(time.Millisecond))
	router.Queue(pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, PointerID: 1, Position: f32.Pt(60, 18)})
	layoutColorPickerFrame(ctx, router, picker, start.Add(2*time.Millisecond))
	layoutColorPickerFrame(ctx, router, picker, start.Add(3*time.Millisecond))

	if !state.open || !router.Source().Focused(&state.trigger) {
		t.Fatal("pointer click did not open and focus the color picker trigger")
	}
	if frame.FocusVisible(ctx, &state.trigger, true) {
		t.Fatal("pointer click exposed the color picker keyboard focus ring")
	}
}

func TestColorPickerKeyboardTriggerFocusRemainsVisible(t *testing.T) {
	ctx := frame.New(nil, nil, locale.LanguageEnglish)
	router := new(input.Router)
	picker := ColorPicker("brand", color.NRGBA{R: 0x04, G: 0x85, B: 0xf7, A: 255}).Label("Accent color")
	start := time.Unix(1, 0)
	layoutColorPickerFrame(ctx, router, picker, start)
	state := colorPickerTestState(ctx, "brand")
	router.Source().Execute(key.FocusCmd{Tag: &state.trigger})
	layoutColorPickerFrame(ctx, router, picker, start.Add(time.Millisecond))

	router.Queue(
		key.Event{Name: key.NameReturn, State: key.Press},
		key.Event{Name: key.NameReturn, State: key.Release},
	)
	layoutColorPickerFrame(ctx, router, picker, start.Add(2*time.Millisecond))
	layoutColorPickerFrame(ctx, router, picker, start.Add(3*time.Millisecond))

	if !state.open || !router.Source().Focused(&state.trigger) {
		t.Fatal("keyboard activation did not open and retain color picker trigger focus")
	}
	if !frame.FocusVisible(ctx, &state.trigger, true) {
		t.Fatal("keyboard activation hid the color picker focus ring")
	}
}

func TestIndependentColorComponentsUseValueSemantics(t *testing.T) {
	value := color.NRGBA{R: 1, G: 2, B: 3, A: 255}
	colors := []color.NRGBA{{R: 4, A: 255}}
	disabled := []color.NRGBA{{R: 5, A: 255}}
	area := ColorArea("area", value)
	configuredArea := area.ShowDots(true).Disabled(true).Label("Area").OnChange(func(color.NRGBA) {})
	picker := ColorSwatchPicker("palette", value, colors).DisabledColors(disabled)
	colors[0] = color.NRGBA{}
	disabled[0] = color.NRGBA{}

	if area.showDots || area.disabled || area.label != "" || area.onChange != nil {
		t.Fatalf("configuring ColorArea mutated base: %#v", area)
	}
	if !configuredArea.showDots || !configuredArea.disabled || configuredArea.label != "Area" || configuredArea.onChange == nil {
		t.Fatalf("configured ColorArea = %#v", configuredArea)
	}
	if picker.colors[0].R != 4 || picker.disabledColors[0].R != 5 {
		t.Fatalf("ColorSwatchPicker retained caller slices: %#v", picker)
	}
}

func TestColorSwatchUsesHeroUISizes(t *testing.T) {
	ctx := frame.New(nil, nil, locale.LanguageEnglish)
	router := new(input.Router)
	now := time.Unix(1, 0)
	for _, test := range []struct {
		size ColorSwatchSize
		want int
	}{
		{ColorSwatchExtraSmall, 16},
		{ColorSwatchSmall, 24},
		{ColorSwatchMedium, 32},
		{ColorSwatchLarge, 36},
		{ColorSwatchExtraLarge, 40},
	} {
		dimensions := layoutColorWidgetFrame(ctx, router, ColorSwatch(color.NRGBA{A: 255}).Size(test.size), now)
		if dimensions.Size != image.Pt(test.want, test.want) {
			t.Fatalf("ColorSwatch size %d = %v, want %dx%d", test.size, dimensions.Size, test.want, test.want)
		}
	}
}

func TestColorSwatchPickerDispatchesSelection(t *testing.T) {
	ctx := frame.New(nil, nil, locale.LanguageEnglish)
	router := new(input.Router)
	red := color.NRGBA{R: 255, A: 255}
	blue := color.NRGBA{B: 255, A: 255}
	var changed color.NRGBA
	picker := ColorSwatchPicker("palette", red, []color.NRGBA{red, blue}).OnChange(func(value color.NRGBA) {
		changed = value
	})
	now := time.Unix(1, 0)
	layoutColorWidgetFrame(ctx, router, picker, now)
	pickerState, ok := frame.PeekState[colorSwatchPickerState](ctx, "palette", stateSlotColorSwatchPicker)
	itemKey := colorSwatchItemKey{index: 1, value: blue}
	if !ok || pickerState.items[itemKey] == nil {
		t.Fatal("missing ColorSwatchPicker item state")
	}
	pickerState.items[itemKey].clickable.Click()
	layoutColorWidgetFrame(ctx, router, picker, now.Add(time.Millisecond))
	if changed != blue {
		t.Fatalf("ColorSwatchPicker change = %#v, want %#v", changed, blue)
	}
}

func TestColorSwatchPickerUsesBorderBoxForSwatchSize(t *testing.T) {
	if got := colorSwatchPickerVisualSize(image.Pt(32, 32), 2, 1); got != image.Pt(28, 28) {
		t.Fatalf("default swatch = %v, want 28x28", got)
	}
	if got := colorSwatchPickerVisualSize(image.Pt(32, 32), 2, .77); got != image.Pt(22, 22) {
		t.Fatalf("selected swatch = %v, want 22x22", got)
	}
	if got := colorSwatchPickerVisualSize(image.Pt(36, 36), 3, 1); got != image.Pt(30, 30) {
		t.Fatalf("large swatch = %v, want 30x30", got)
	}
}

func TestHSVConversionRoundTripsNRGBA(t *testing.T) {
	values := []color.NRGBA{
		{R: 255, A: 255},
		{G: 255, A: 192},
		{B: 255, A: 128},
		{R: 50, G: 85, B: 120, A: 64},
		{R: 127, G: 127, B: 127, A: 255},
		{A: 255},
		{R: 255, G: 255, B: 255, A: 255},
	}
	for _, value := range values {
		if got := hsvToNRGBA(nrgbaToHSV(value)); !nearColor(got, value) {
			t.Fatalf("HSV round trip %#v = %#v", value, got)
		}
	}
}

func TestColorPickerRetainsHueForAchromaticValues(t *testing.T) {
	state := new(colorValueState)
	state.sync(color.NRGBA{B: 255, A: 255})
	state.sync(color.NRGBA{A: 255})
	if got := state.hsv().h; got < .66 || got > .67 {
		t.Fatalf("retained hue = %v, want blue", got)
	}
}

func TestColorPickerAreaChangeKeepsExactHue(t *testing.T) {
	state := new(colorValueState)
	state.sync(color.NRGBA{R: 4, G: 133, B: 247, A: 255})
	before := state.hsv()
	next := before
	next.s = .43
	next.v = .71
	value := hsvToNRGBA(next)
	state.accept(value, next.h)
	state.sync(value)
	after := state.hsv()

	if after.h != before.h {
		t.Fatalf("hue moved after color area change: %v -> %v", before.h, after.h)
	}
	size := image.Pt(240, 24)
	if colorSliderCenter(size, after.h) != colorSliderCenter(size, before.h) {
		t.Fatalf("hue slider moved after color area change")
	}
}

func TestColorSliderRetainsRightHueEndpoint(t *testing.T) {
	state := new(colorValueState)
	value := hsvToNRGBA(hsvColor{h: 1, s: 1, v: 1, a: 1})
	state.accept(value, 1)
	state.sync(value)
	if got := state.hsv().h; got != 1 {
		t.Fatalf("right hue endpoint = %v, want 1", got)
	}
}

func BenchmarkColorPickerChangingColor(b *testing.B) {
	ctx := frame.New(nil, nil, locale.LanguageEnglish)
	router := new(input.Router)
	operations := new(op.Ops)
	start := time.Unix(1, 0)
	value := color.NRGBA{R: 4, G: 133, B: 247, A: 255}
	layoutColorWidgetFrameWithOps(ctx, router, ColorPicker("benchmark", value), start, operations)
	pickerState := colorPickerTestState(ctx, "benchmark")
	pickerState.open = true
	layoutColorWidgetFrameWithOps(ctx, router, ColorPicker("benchmark", value), start.Add(colorPickerEnterDuration), operations)

	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		value = hsvToNRGBA(hsvColor{h: .57, s: float64(index%100) / 100, v: .8, a: 1})
		layoutColorWidgetFrameWithOps(ctx, router, ColorPicker("benchmark", value), start.Add(colorPickerEnterDuration+time.Duration(index+1)*time.Millisecond), operations)
	}
}

func TestHexColorFormattingAndParsing(t *testing.T) {
	value := color.NRGBA{R: 0x32, G: 0x55, B: 0x78, A: 0x80}
	if got := formatHexColor(value, false); got != "#325578" {
		t.Fatalf("hex = %q", got)
	}
	if got := formatHexColor(value, true); got != "#32557880" {
		t.Fatalf("hex alpha = %q", got)
	}
	for text, want := range map[string]color.NRGBA{
		"#0485F7":  {R: 0x04, G: 0x85, B: 0xf7, A: 0x7f},
		"32557880": {R: 0x32, G: 0x55, B: 0x78, A: 0x80},
		"#F43":     {R: 0xff, G: 0x44, B: 0x33, A: 0x7f},
		"#F438":    {R: 0xff, G: 0x44, B: 0x33, A: 0x88},
	} {
		got, ok := parseHexColor(text, 0x7f)
		if !ok || got != want {
			t.Fatalf("parse %q = %#v, %v; want %#v", text, got, ok, want)
		}
	}
	if _, ok := parseHexColor("#12", 255); ok {
		t.Fatal("invalid hex color was accepted")
	}
}

func TestColorPickerAreaPointerChangesControlledValue(t *testing.T) {
	ctx := frame.New(nil, nil, locale.LanguageEnglish)
	router := new(input.Router)
	value := color.NRGBA{R: 0x04, G: 0x85, B: 0xf7, A: 255}
	initial := value
	picker := func() ColorPickerWidget {
		return ColorPicker("brand", value).
			Label("Pick a color").
			OnChange(func(next color.NRGBA) { value = next })
	}
	start := time.Unix(1, 0)

	dimensions := layoutColorPickerFrame(ctx, router, picker(), start)
	if dimensions.Size.Y != 36 {
		t.Fatalf("trigger height = %d, want 36", dimensions.Size.Y)
	}
	state := colorPickerTestState(ctx, "brand")
	state.trigger.Click()
	layoutColorPickerFrame(ctx, router, picker(), start.Add(time.Millisecond))
	layoutColorPickerFrame(ctx, router, picker(), start.Add(time.Millisecond+colorPickerEnterDuration))
	areaState := colorAreaTestState(ctx, "brand/area")

	router.Queue(pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: 1, Buttons: pointer.ButtonPrimary, Position: f32.Pt(124, 166)})
	layoutColorPickerFrame(ctx, router, picker(), start.Add(2*time.Millisecond+colorPickerEnterDuration))
	if !state.open || !areaState.control.dragging {
		t.Fatal("color picker closed when color area pointer was pressed")
	}

	layoutColorPickerFrame(ctx, router, picker(), start.Add(3*time.Millisecond+colorPickerEnterDuration))
	if !state.open {
		t.Fatal("color picker closed while color area pointer remained pressed")
	}

	router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, PointerID: 1, Buttons: pointer.ButtonPrimary, Position: f32.Pt(164, 126)})
	layoutColorPickerFrame(ctx, router, picker(), start.Add(4*time.Millisecond+colorPickerEnterDuration))
	if !state.open {
		t.Fatal("color picker closed while dragging in the color area")
	}

	router.Queue(pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, PointerID: 1, Position: f32.Pt(164, 126)})
	layoutColorPickerFrame(ctx, router, picker(), start.Add(5*time.Millisecond+colorPickerEnterDuration))

	if value == (color.NRGBA{}) || value == initial {
		t.Fatalf("color area change = %#v", value)
	}
	if !state.open {
		t.Fatal("color picker closed when color area pointer was released")
	}

	state.dismiss[0].Click()
	layoutColorPickerFrame(ctx, router, picker(), start.Add(6*time.Millisecond+colorPickerEnterDuration))
	if state.open {
		t.Fatal("outside click did not close color picker")
	}
}

func TestColorPickerHueKeyboardAndEscape(t *testing.T) {
	ctx := frame.New(nil, nil, locale.LanguageEnglish)
	router := new(input.Router)
	initial := color.NRGBA{R: 0x04, G: 0x85, B: 0xf7, A: 255}
	var changed color.NRGBA
	picker := ColorPicker("brand", initial).OnChange(func(value color.NRGBA) { changed = value })
	start := time.Unix(1, 0)

	layoutColorPickerFrame(ctx, router, picker, start)
	state := colorPickerTestState(ctx, "brand")
	state.trigger.Click()
	layoutColorPickerFrame(ctx, router, picker, start.Add(time.Millisecond))
	layoutColorPickerFrame(ctx, router, picker, start.Add(time.Millisecond+colorPickerEnterDuration))
	hueState := colorSliderTestState(ctx, "brand/hue")
	router.Source().Execute(key.FocusCmd{Tag: &hueState.control})
	router.Queue(key.Event{Name: key.NameRightArrow, State: key.Press})
	layoutColorPickerFrame(ctx, router, picker, start.Add(2*time.Millisecond+colorPickerEnterDuration))

	if changed == (color.NRGBA{}) || changed == initial {
		t.Fatalf("keyboard hue change = %#v", changed)
	}
	if !state.open {
		t.Fatal("color picker closed when hue gained focus")
	}

	router.Queue(key.Event{Name: key.NameEscape, State: key.Press})
	layoutColorPickerFrame(ctx, router, picker, start.Add(3*time.Millisecond+colorPickerEnterDuration))
	if state.open {
		t.Fatal("Escape did not close color picker")
	}
}

func nearColor(first, second color.NRGBA) bool {
	near := func(a, b uint8) bool {
		difference := int(a) - int(b)
		if difference < 0 {
			difference = -difference
		}
		return difference <= 1
	}
	return near(first.R, second.R) && near(first.G, second.G) && near(first.B, second.B) && near(first.A, second.A)
}

func colorPickerTestState(ctx *frame.Context, key string) *colorPickerState {
	value, ok := frame.PeekState[colorPickerState](ctx, key, stateSlotColorPicker)
	if !ok {
		panic("missing color picker state")
	}
	return value
}

func colorAreaTestState(ctx *frame.Context, key string) *colorAreaState {
	value, ok := frame.PeekState[colorAreaState](ctx, key, stateSlotColorArea)
	if !ok {
		panic("missing color area state")
	}
	return value
}

func colorSliderTestState(ctx *frame.Context, key string) *colorSliderState {
	value, ok := frame.PeekState[colorSliderState](ctx, key, stateSlotColorSlider)
	if !ok {
		panic("missing color slider state")
	}
	return value
}

func layoutColorPickerFrame(ctx *frame.Context, router *input.Router, picker ColorPickerWidget, now time.Time) layout.Dimensions {
	return layoutColorWidgetFrame(ctx, router, picker, now)
}

func layoutColorWidgetFrame(ctx *frame.Context, router *input.Router, widget frame.Widget, now time.Time) layout.Dimensions {
	return layoutColorWidgetFrameWithOps(ctx, router, widget, now, new(op.Ops))
}

func layoutColorWidgetFrameWithOps(ctx *frame.Context, router *input.Router, widget frame.Widget, now time.Time, operations *op.Ops) layout.Dimensions {
	operations.Reset()
	viewport := image.Pt(400, 500)
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: viewport},
		Source:      router.Source(),
		Ops:         operations,
		Now:         now,
	}
	frame.BeginFrameWithViewport(ctx, viewport)
	dimensions := widget.Layout(ctx, gtx)
	frame.LayoutOverlays(ctx, gtx)
	frame.ApplyFrameCommands(ctx, gtx)
	frame.EndFrame(ctx)
	router.Frame(operations)
	return dimensions
}
