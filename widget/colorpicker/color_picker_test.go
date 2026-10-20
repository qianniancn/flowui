package colorpicker

import (
	"image"
	"image/color"
	"slices"
	"testing"
	"time"

	"gioui.org/f32"
	"gioui.org/io/input"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/testtheme"
)

const colorPickerEnterDuration = 150 * time.Millisecond

func TestColorPickerOptionsUseValueSemantics(t *testing.T) {
	base := New("brand", color.NRGBA{R: 1, A: 255})
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
	ctx := core.New(nil, testtheme.Behavior(), locale.LanguageEnglish)
	router := new(input.Router)
	picker := New("brand", color.NRGBA{R: 0x04, G: 0x85, B: 0xf7, A: 255}).Label("Accent color")
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
	if core.FocusVisible(ctx, &state.trigger, true) {
		t.Fatal("pointer click exposed the color picker keyboard focus ring")
	}
}

func TestColorPickerKeyboardTriggerFocusRemainsVisible(t *testing.T) {
	ctx := core.New(nil, testtheme.Behavior(), locale.LanguageEnglish)
	router := new(input.Router)
	picker := New("brand", color.NRGBA{R: 0x04, G: 0x85, B: 0xf7, A: 255}).Label("Accent color")
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
	if !core.FocusVisible(ctx, &state.trigger, true) {
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

func TestColorSwatchUsesStandardSizes(t *testing.T) {
	ctx := core.New(nil, testtheme.Behavior(), locale.LanguageEnglish)
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
	ctx := core.New(nil, testtheme.Behavior(), locale.LanguageEnglish)
	router := new(input.Router)
	red := color.NRGBA{R: 255, A: 255}
	blue := color.NRGBA{B: 255, A: 255}
	var changed color.NRGBA
	picker := ColorSwatchPicker("palette", red, []color.NRGBA{red, blue}).OnChange(func(value color.NRGBA) {
		changed = value
	})
	now := time.Unix(1, 0)
	layoutColorWidgetFrame(ctx, router, picker, now)
	pickerState, ok := core.PeekState[colorSwatchPickerState](ctx, "palette", stateSlotColorSwatchPicker)
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

func TestColorSwatchReusesStandaloneVisualCache(t *testing.T) {
	ctx := core.New(nil, testtheme.Behavior(), locale.LanguageEnglish)
	router := new(input.Router)
	swatch := ColorSwatch(color.NRGBA{R: 42, G: 128, B: 220, A: 255})
	now := time.Unix(1, 0)
	layoutColorWidgetFrame(ctx, router, swatch, now)
	layoutColorWidgetFrame(ctx, router, swatch, now.Add(time.Millisecond))
	cache, ok := core.PeekState[colorPickerVisualCache](ctx, standaloneColorPickerVisualKey, standaloneColorPickerVisualSlot)
	if !ok {
		t.Fatal("missing standalone ColorSwatch visual cache")
	}
	stats := cache.cache.Stats()
	if stats.Misses != 1 || stats.Hits < 1 {
		t.Fatalf("standalone visual cache stats = %+v, want one miss and a hit", stats)
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
	ctx := core.New(nil, testtheme.Behavior(), locale.LanguageEnglish)
	router := new(input.Router)
	operations := new(op.Ops)
	start := time.Unix(1, 0)
	value := color.NRGBA{R: 4, G: 133, B: 247, A: 255}
	layoutColorWidgetFrameWithOps(ctx, router, New("benchmark", value), start, operations)
	pickerState := colorPickerTestState(ctx, "benchmark")
	pickerState.open = true
	layoutColorWidgetFrameWithOps(ctx, router, New("benchmark", value), start.Add(colorPickerEnterDuration), operations)

	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		value = hsvToNRGBA(hsvColor{h: .57, s: float64(index%100) / 100, v: .8, a: 1})
		layoutColorWidgetFrameWithOps(ctx, router, New("benchmark", value), start.Add(colorPickerEnterDuration+time.Duration(index+1)*time.Millisecond), operations)
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
	ctx := core.New(nil, testtheme.Behavior(), locale.LanguageEnglish)
	router := new(input.Router)
	value := color.NRGBA{R: 0x04, G: 0x85, B: 0xf7, A: 255}
	initial := value
	picker := func() Widget {
		return New("brand", value).
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
	ctx := core.New(nil, testtheme.Behavior(), locale.LanguageEnglish)
	router := new(input.Router)
	initial := color.NRGBA{R: 0x04, G: 0x85, B: 0xf7, A: 255}
	var changed color.NRGBA
	picker := New("brand", initial).OnChange(func(value color.NRGBA) { changed = value })
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
	if !router.Source().Focused(&state.trigger) {
		t.Fatal("Escape did not restore focus to the color picker trigger")
	}
}

func TestColorPickerOutsideDismissCommitsHistoryOnce(t *testing.T) {
	ctx := core.New(nil, testtheme.Behavior(), locale.LanguageEnglish)
	router := new(input.Router)
	initial := color.NRGBA{R: 0x04, G: 0x85, B: 0xf7, A: 255}
	picker := New("brand", initial).HistorySize(4)
	start := time.Unix(1, 0)

	layoutColorPickerFrame(ctx, router, picker, start)
	state := colorPickerTestState(ctx, "brand")
	state.trigger.Click()
	layoutColorPickerFrame(ctx, router, picker, start.Add(time.Millisecond))
	layoutColorPickerFrame(ctx, router, picker, start.Add(time.Millisecond+colorPickerEnterDuration))
	state.dismiss[0].Click()
	layoutColorPickerFrame(ctx, router, picker, start.Add(2*time.Millisecond+colorPickerEnterDuration))

	if state.open {
		t.Fatal("outside dismiss left the color picker open")
	}
	if len(state.history) != 1 || state.history[0] != initial {
		t.Fatalf("outside dismiss history = %#v, want one committed value %#v", state.history, initial)
	}

	// The closing transition may be laid out again; it must not add a duplicate.
	layoutColorPickerFrame(ctx, router, picker, start.Add(3*time.Millisecond+colorPickerEnterDuration))
	if len(state.history) != 1 {
		t.Fatalf("closing transition duplicated history: %#v", state.history)
	}
}

func TestColorPickerAlphaSliderKeyboardAndDisabledPicker(t *testing.T) {
	initial := color.NRGBA{R: 0x04, G: 0x85, B: 0xf7, A: 93}
	ctx := core.New(nil, testtheme.Behavior(), locale.LanguageEnglish)
	router := new(input.Router)
	value := initial
	var changed color.NRGBA
	picker := func() Widget {
		return New("brand", value).
			Alpha(true).
			ShowRGB(false).
			HideField().
			ShowHistory(false).
			OnChange(func(next color.NRGBA) {
				changed = next
				value = next
			})
	}
	start := time.Unix(1, 0)

	layoutColorPickerFrame(ctx, router, picker(), start)
	state := colorPickerTestState(ctx, "brand")
	state.trigger.Click()
	layoutColorPickerFrame(ctx, router, picker(), start.Add(time.Millisecond))
	layoutColorPickerFrame(ctx, router, picker(), start.Add(time.Millisecond+colorPickerEnterDuration))
	alpha := colorSliderTestState(ctx, "brand/alpha")
	router.Source().Execute(key.FocusCmd{Tag: &alpha.control})
	router.Queue(key.Event{Name: key.NameEnd, State: key.Press})
	layoutColorPickerFrame(ctx, router, picker(), start.Add(2*time.Millisecond+colorPickerEnterDuration))
	if changed.A != 255 {
		t.Fatalf("alpha End value = %#v, want alpha 255", changed)
	}
	router.Source().Execute(key.FocusCmd{Tag: &alpha.control})
	router.Queue(key.Event{Name: key.NameHome, State: key.Press})
	layoutColorPickerFrame(ctx, router, picker(), start.Add(3*time.Millisecond+colorPickerEnterDuration))
	if changed.A != 0 {
		t.Fatalf("alpha Home value = %#v, want alpha 0", changed)
	}

	disabledCtx := core.New(nil, testtheme.Behavior(), locale.LanguageEnglish)
	disabledRouter := new(input.Router)
	disabled := New("disabled", initial).Disabled(true)
	layoutColorPickerFrame(disabledCtx, disabledRouter, disabled, start)
	disabledState := colorPickerTestState(disabledCtx, "disabled")
	disabledRouter.Queue(pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: 1, Buttons: pointer.ButtonPrimary, Position: f32.Pt(24, 18)})
	layoutColorPickerFrame(disabledCtx, disabledRouter, disabled, start.Add(time.Millisecond))
	if disabledState.open {
		t.Fatal("disabled color picker opened from a pointer press")
	}
}

func TestColorPickerHexFieldCommitsValidValuesAndRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name      string
		alpha     bool
		text      string
		want      color.NRGBA
		wantMax   int
		wantInput string
	}{
		{
			name:      "rgb value",
			text:      "#112233",
			want:      color.NRGBA{R: 0x11, G: 0x22, B: 0x33, A: 0x7a},
			wantMax:   7,
			wantInput: "#112233",
		},
		{
			name:      "rgba value",
			alpha:     true,
			text:      "#11223344",
			want:      color.NRGBA{R: 0x11, G: 0x22, B: 0x33, A: 0x44},
			wantMax:   9,
			wantInput: "#11223344",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			initial := color.NRGBA{R: 4, G: 133, B: 247, A: 0x7a}
			value := initial
			changes := 0
			picker := func() Widget {
				return New("brand", value).
					Alpha(test.alpha).
					OnChange(func(next color.NRGBA) {
						changes++
						value = next
					})
			}
			ctx := core.New(nil, testtheme.Behavior(), locale.LanguageEnglish)
			router := new(input.Router)
			start := time.Unix(1, 0)
			layoutColorPickerFrame(ctx, router, picker(), start)
			state := colorPickerTestState(ctx, "brand")
			state.trigger.Click()
			layoutColorPickerFrame(ctx, router, picker(), start.Add(time.Millisecond))
			layoutColorPickerFrame(ctx, router, picker(), start.Add(time.Millisecond+colorPickerEnterDuration))

			editor := colorEditorTestState(ctx, "brand/field/input")
			if editor.MaxLen != test.wantMax || editor.Text() != formatHexColor(initial, test.alpha) {
				t.Fatalf("hex editor before edit = text %q max %d, want %q max %d", editor.Text(), editor.MaxLen, formatHexColor(initial, test.alpha), test.wantMax)
			}
			editor.SetText(test.text)
			layoutColorPickerFrame(ctx, router, picker(), start.Add(2*time.Millisecond+colorPickerEnterDuration))
			if changes != 1 || value != test.want {
				t.Fatalf("valid hex change = %#v with %d callbacks, want %#v and one callback", value, changes, test.want)
			}

			for _, invalid := range []string{"#12", "#gggggg", ""} {
				editor.SetText(invalid)
				layoutColorPickerFrame(ctx, router, picker(), start.Add(3*time.Millisecond+colorPickerEnterDuration))
				if changes != 1 || value != test.want {
					t.Fatalf("invalid hex %q changed value to %#v with %d callbacks", invalid, value, changes)
				}
			}
		})
	}
}

func TestColorPickerRGBFieldsCommitChannelsAndRejectInvalidInput(t *testing.T) {
	tests := []struct {
		name  string
		alpha bool
		key   string
		text  string
		want  color.NRGBA
	}{
		{name: "red", key: "brand/r/value", text: "12", want: color.NRGBA{R: 12, G: 20, B: 30, A: 200}},
		{name: "green", key: "brand/g/value", text: "34", want: color.NRGBA{R: 10, G: 34, B: 30, A: 200}},
		{name: "blue", key: "brand/b/value", text: "56", want: color.NRGBA{R: 10, G: 20, B: 56, A: 200}},
		{name: "alpha", alpha: true, key: "brand/a/value", text: "78", want: color.NRGBA{R: 10, G: 20, B: 30, A: 78}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			initial := color.NRGBA{R: 10, G: 20, B: 30, A: 200}
			value := initial
			changes := 0
			picker := func() Widget {
				return New("brand", value).
					Alpha(test.alpha).
					OnChange(func(next color.NRGBA) {
						changes++
						value = next
					})
			}
			ctx := core.New(nil, testtheme.Behavior(), locale.LanguageEnglish)
			router := new(input.Router)
			start := time.Unix(1, 0)
			layoutColorPickerFrame(ctx, router, picker(), start)
			state := colorPickerTestState(ctx, "brand")
			state.trigger.Click()
			layoutColorPickerFrame(ctx, router, picker(), start.Add(time.Millisecond))
			layoutColorPickerFrame(ctx, router, picker(), start.Add(time.Millisecond+colorPickerEnterDuration))

			editor, ok := core.PeekState[widget.Editor](ctx, test.key, "editor")
			if !ok {
				t.Fatalf("missing RGB editor %q", test.key)
			}
			editor.SetText(test.text)
			layoutColorPickerFrame(ctx, router, picker(), start.Add(2*time.Millisecond+colorPickerEnterDuration))
			if changes != 1 || value != test.want {
				t.Fatalf("channel change = %#v with %d callbacks, want %#v and one callback", value, changes, test.want)
			}

			for _, invalid := range []string{"256", "-1", "letters", ""} {
				editor.SetText(invalid)
				layoutColorPickerFrame(ctx, router, picker(), start.Add(3*time.Millisecond+colorPickerEnterDuration))
				if changes != 1 || value != test.want {
					t.Fatalf("invalid channel %q changed value to %#v with %d callbacks", invalid, value, changes)
				}
			}
		})
	}

	ctx := core.New(nil, testtheme.Behavior(), locale.LanguageEnglish)
	router := new(input.Router)
	start := time.Unix(1, 0)
	picker := New("rgb-only", color.NRGBA{R: 1, G: 2, B: 3, A: 4}).ShowRGB(true)
	layoutColorPickerFrame(ctx, router, picker, start)
	state := colorPickerTestState(ctx, "rgb-only")
	state.trigger.Click()
	layoutColorPickerFrame(ctx, router, picker, start.Add(time.Millisecond))
	layoutColorPickerFrame(ctx, router, picker, start.Add(time.Millisecond+colorPickerEnterDuration))
	if _, ok := core.PeekState[widget.Editor](ctx, "rgb-only/a/value", "editor"); ok {
		t.Fatal("alpha editor exists while Alpha is disabled")
	}
}

func TestColorPickerPresetSelectionHonorsDisabledAndCurrentItems(t *testing.T) {
	red := color.NRGBA{R: 255, A: 255}
	blue := color.NRGBA{B: 255, A: 255}
	green := color.NRGBA{G: 255, A: 255}
	tests := []struct {
		name          string
		initial       color.NRGBA
		item          color.NRGBA
		wantCallbacks int
		wantValue     color.NRGBA
	}{
		{name: "select enabled preset", initial: red, item: blue, wantCallbacks: 1, wantValue: blue},
		{name: "current preset", initial: red, item: red, wantCallbacks: 0, wantValue: red},
		{name: "select another preset", initial: red, item: green, wantCallbacks: 1, wantValue: green},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value := test.initial
			callbacks := 0
			picker := func() Widget {
				return New("brand", value).
					Presets([]color.NRGBA{red, blue, green}).
					OnChange(func(next color.NRGBA) {
						callbacks++
						value = next
					})
			}
			ctx := core.New(nil, testtheme.Behavior(), locale.LanguageEnglish)
			router := new(input.Router)
			start := time.Unix(1, 0)
			layoutColorPickerFrame(ctx, router, picker(), start)
			state := colorPickerTestState(ctx, "brand")
			state.trigger.Click()
			layoutColorPickerFrame(ctx, router, picker(), start.Add(time.Millisecond))
			layoutColorPickerFrame(ctx, router, picker(), start.Add(time.Millisecond+colorPickerEnterDuration))
			presetState, ok := core.PeekState[colorSwatchPickerState](ctx, "brand/presets", stateSlotColorSwatchPicker)
			if !ok {
				t.Fatal("missing embedded preset state")
			}
			var itemState *colorSwatchItemState
			for key, candidate := range presetState.items {
				if key.value == test.item {
					itemState = candidate
					break
				}
			}
			if itemState == nil {
				t.Fatalf("missing preset item %#v", test.item)
			}
			itemState.clickable.Click()
			layoutColorPickerFrame(ctx, router, picker(), start.Add(2*time.Millisecond+colorPickerEnterDuration))
			if callbacks != test.wantCallbacks || value != test.wantValue {
				t.Fatalf("preset result = %#v with %d callbacks, want %#v with %d", value, callbacks, test.wantValue, test.wantCallbacks)
			}
		})
	}
}

func TestColorSwatchPickerDisabledAndCurrentItemsDoNotDispatch(t *testing.T) {
	red := color.NRGBA{R: 255, A: 255}
	blue := color.NRGBA{B: 255, A: 255}
	tests := []struct {
		name          string
		initial       color.NRGBA
		disabled      []color.NRGBA
		item          color.NRGBA
		wantCallbacks int
		wantValue     color.NRGBA
	}{
		{name: "disabled item", initial: red, disabled: []color.NRGBA{blue}, item: blue, wantValue: red},
		{name: "current item", initial: red, item: red, wantValue: red},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value := test.initial
			callbacks := 0
			picker := ColorSwatchPicker("palette", value, []color.NRGBA{red, blue}).
				DisabledColors(test.disabled).
				OnChange(func(next color.NRGBA) {
					callbacks++
					value = next
				})
			ctx := core.New(nil, testtheme.Behavior(), locale.LanguageEnglish)
			router := new(input.Router)
			layoutColorWidgetFrame(ctx, router, picker, time.Unix(1, 0))
			state, ok := core.PeekState[colorSwatchPickerState](ctx, "palette", stateSlotColorSwatchPicker)
			if !ok {
				t.Fatal("missing swatch picker state")
			}
			var itemState *colorSwatchItemState
			for key, candidate := range state.items {
				if key.value == test.item {
					itemState = candidate
					break
				}
			}
			if itemState == nil {
				t.Fatalf("missing item %#v", test.item)
			}
			itemState.clickable.Click()
			layoutColorWidgetFrame(ctx, router, picker, time.Unix(1, 1))
			if callbacks != test.wantCallbacks || value != test.wantValue {
				t.Fatalf("item result = %#v with %d callbacks, want %#v with %d", value, callbacks, test.wantValue, test.wantCallbacks)
			}
		})
	}
}

func TestColorPickerControlledAndUncontrolledValuesPersistAcrossFrames(t *testing.T) {
	tests := []struct {
		name         string
		uncontrolled bool
	}{
		{name: "controlled", uncontrolled: false},
		{name: "uncontrolled", uncontrolled: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			initial := color.NRGBA{R: 1, G: 2, B: 3, A: 255}
			selected := color.NRGBA{R: 10, G: 20, B: 30, A: 255}
			value := initial
			callbacks := 0
			picker := New("brand", initial)
			if test.uncontrolled {
				picker = picker.DefaultValue(initial)
			}
			picker = picker.OnChange(func(next color.NRGBA) {
				callbacks++
				if !test.uncontrolled {
					value = next
				}
			})
			frame := func() Widget {
				current := initial
				if !test.uncontrolled {
					current = value
				}
				result := New("brand", current).OnChange(picker.onChange)
				if test.uncontrolled {
					result = result.DefaultValue(initial)
				}
				return result
			}
			ctx := core.New(nil, testtheme.Behavior(), locale.LanguageEnglish)
			router := new(input.Router)
			start := time.Unix(1, 0)
			layoutColorPickerFrame(ctx, router, frame(), start)
			state := colorPickerTestState(ctx, "brand")
			state.trigger.Click()
			layoutColorPickerFrame(ctx, router, frame(), start.Add(time.Millisecond))
			layoutColorPickerFrame(ctx, router, frame(), start.Add(time.Millisecond+colorPickerEnterDuration))
			presets, ok := core.PeekState[colorSwatchPickerState](ctx, "brand/presets", stateSlotColorSwatchPicker)
			if !ok {
				// Add the preset only for this interaction while preserving the
				// controlled/uncontrolled declaration under test.
				current := initial
				if !test.uncontrolled {
					current = value
				}
				presetWidget := New("brand", current).Presets([]color.NRGBA{selected}).OnChange(picker.onChange)
				if test.uncontrolled {
					presetWidget = presetWidget.DefaultValue(initial)
				}
				layoutColorPickerFrame(ctx, router, presetWidget, start.Add(2*time.Millisecond+colorPickerEnterDuration))
				presets, ok = core.PeekState[colorSwatchPickerState](ctx, "brand/presets", stateSlotColorSwatchPicker)
			}
			if !ok {
				t.Fatal("missing preset state")
			}
			for key, item := range presets.items {
				if key.value == selected {
					item.clickable.Click()
				}
			}
			current := initial
			if !test.uncontrolled {
				current = value
			}
			selectedWidget := New("brand", current).Presets([]color.NRGBA{selected}).OnChange(picker.onChange)
			if test.uncontrolled {
				selectedWidget = selectedWidget.DefaultValue(initial)
			}
			layoutColorPickerFrame(ctx, router, selectedWidget, start.Add(3*time.Millisecond+colorPickerEnterDuration))
			if callbacks != 1 {
				t.Fatalf("callbacks = %d, want one", callbacks)
			}
			if !test.uncontrolled && value != selected {
				t.Fatalf("controlled value = %#v, want %#v", value, selected)
			}
			if test.uncontrolled && state.currentValue(picker) != selected {
				t.Fatalf("uncontrolled value = %#v, want %#v", state.currentValue(picker), selected)
			}
		})
	}
}

func TestColorPickerHistoryDeduplicatesAndFiltersPresets(t *testing.T) {
	red := color.NRGBA{R: 255, A: 255}
	green := color.NRGBA{G: 255, A: 255}
	blue := color.NRGBA{B: 255, A: 255}
	state := new(colorPickerState)
	state.pushHistory(red, 4)
	state.pushHistory(green, 4)
	state.pushHistory(red, 4)
	state.pushHistory(blue, 2)
	if want := []color.NRGBA{blue, red}; !slices.Equal(state.history, want) {
		t.Fatalf("history = %#v, want %#v", state.history, want)
	}
	if got := historyWithoutPresets(state.history, []color.NRGBA{red}); !slices.Equal(got, []color.NRGBA{blue}) {
		t.Fatalf("history without presets = %#v, want %#v", got, []color.NRGBA{blue})
	}
}

func TestColorPickerHistoryTracksDiscreteCommitsOnly(t *testing.T) {
	initial := color.NRGBA{R: 4, G: 133, B: 247, A: 255}
	preset := color.NRGBA{R: 200, G: 100, B: 50, A: 255}
	value := initial
	picker := func() Widget {
		return New("brand", value).
			Presets([]color.NRGBA{preset}).
			OnChange(func(next color.NRGBA) { value = next })
	}
	ctx := core.New(nil, testtheme.Behavior(), locale.LanguageEnglish)
	router := new(input.Router)
	start := time.Unix(1, 0)
	layoutColorPickerFrame(ctx, router, picker(), start)
	state := colorPickerTestState(ctx, "brand")
	state.trigger.Click()
	layoutColorPickerFrame(ctx, router, picker(), start.Add(time.Millisecond))
	layoutColorPickerFrame(ctx, router, picker(), start.Add(time.Millisecond+colorPickerEnterDuration))

	// Area drag is continuous feedback and must not create a recent-color item.
	areaState := colorAreaTestState(ctx, "brand/area")
	router.Queue(pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: 1, Buttons: pointer.ButtonPrimary, Position: f32.Pt(124, 166)})
	layoutColorPickerFrame(ctx, router, picker(), start.Add(2*time.Millisecond+colorPickerEnterDuration))
	router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, PointerID: 1, Buttons: pointer.ButtonPrimary, Position: f32.Pt(164, 126)})
	layoutColorPickerFrame(ctx, router, picker(), start.Add(3*time.Millisecond+colorPickerEnterDuration))
	router.Queue(pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, PointerID: 1, Position: f32.Pt(164, 126)})
	layoutColorPickerFrame(ctx, router, picker(), start.Add(4*time.Millisecond+colorPickerEnterDuration))
	if !areaState.control.dragging && len(state.history) != 0 {
		t.Fatalf("area drag history = %#v, want no continuous commit", state.history)
	}

	// Hex is a discrete commit and should be the first history entry.
	editor := colorEditorTestState(ctx, "brand/field/input")
	editor.SetText("#102030")
	layoutColorPickerFrame(ctx, router, picker(), start.Add(5*time.Millisecond+colorPickerEnterDuration))
	hex := color.NRGBA{R: 0x10, G: 0x20, B: 0x30, A: 255}
	if len(state.history) != 1 || state.history[0] != hex {
		t.Fatalf("hex history = %#v, want %#v", state.history, []color.NRGBA{hex})
	}

	// A preset is a discrete selection, but presets are filtered from the
	// history swatch row so they do not appear twice in the panel.
	presetState, ok := core.PeekState[colorSwatchPickerState](ctx, "brand/presets", stateSlotColorSwatchPicker)
	if !ok {
		t.Fatal("missing preset state")
	}
	for key, item := range presetState.items {
		if key.value == preset {
			item.clickable.Click()
		}
	}
	layoutColorPickerFrame(ctx, router, picker(), start.Add(6*time.Millisecond+colorPickerEnterDuration))
	if len(state.history) != 2 || state.history[0] != preset || state.history[1] != hex {
		t.Fatalf("preset history = %#v, want preset then hex", state.history)
	}
	if got := historyWithoutPresets(state.history, []color.NRGBA{preset}); !slices.Equal(got, []color.NRGBA{hex}) {
		t.Fatalf("filtered history = %#v, want %#v", got, []color.NRGBA{hex})
	}

	// Selecting the existing history value moves it to the front without a
	// duplicate entry.
	historyState, ok := core.PeekState[colorSwatchPickerState](ctx, "brand/history", stateSlotColorSwatchPicker)
	if !ok {
		t.Fatal("missing history state")
	}
	for key, item := range historyState.items {
		if key.value == hex {
			item.clickable.Click()
		}
	}
	layoutColorPickerFrame(ctx, router, picker(), start.Add(7*time.Millisecond+colorPickerEnterDuration))
	if len(state.history) != 2 || state.history[0] != hex || state.history[1] != preset {
		t.Fatalf("history reselect = %#v, want hex then preset", state.history)
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

func colorPickerTestState(ctx *core.Context, key string) *colorPickerState {
	value, ok := core.PeekState[colorPickerState](ctx, key, stateSlotColorPicker)
	if !ok {
		panic("missing color picker state")
	}
	return value
}

func colorAreaTestState(ctx *core.Context, key string) *colorAreaState {
	value, ok := core.PeekState[colorAreaState](ctx, key, stateSlotColorArea)
	if !ok {
		panic("missing color area state")
	}
	return value
}

func colorSliderTestState(ctx *core.Context, key string) *colorSliderState {
	value, ok := core.PeekState[colorSliderState](ctx, key, stateSlotColorSlider)
	if !ok {
		panic("missing color slider state")
	}
	return value
}

func colorEditorTestState(ctx *core.Context, key string) *widget.Editor {
	value, ok := core.PeekState[widget.Editor](ctx, key, "editor")
	if !ok {
		panic("missing color picker editor state")
	}
	return value
}

func layoutColorPickerFrame(ctx *core.Context, router *input.Router, picker Widget, now time.Time) layout.Dimensions {
	return layoutColorWidgetFrame(ctx, router, picker, now)
}

func layoutColorWidgetFrame(ctx *core.Context, router *input.Router, widget core.Widget, now time.Time) layout.Dimensions {
	return layoutColorWidgetFrameWithOps(ctx, router, widget, now, new(op.Ops))
}

func layoutColorWidgetFrameWithOps(ctx *core.Context, router *input.Router, widget core.Widget, now time.Time, operations *op.Ops) layout.Dimensions {
	operations.Reset()
	viewport := image.Pt(400, 500)
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: viewport},
		Source:      router.Source(),
		Ops:         operations,
		Now:         now,
	}
	core.BeginFrameWithViewport(ctx, viewport)
	dimensions := widget.Layout(ctx, gtx)
	core.LayoutOverlays(ctx, gtx)
	core.ApplyFrameCommands(ctx, gtx)
	core.EndFrame(ctx)
	router.Frame(operations)
	return dimensions
}
