package slider

import (
	"image"
	"image/color"
	"math"
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
	"github.com/qianniancn/FlowUI/internal/theme"
)

func TestSliderDefaults(t *testing.T) {
	value := Slider("volume", 30)
	if value.key != "volume" || value.value != 30 || value.minValue != 0 || value.maxValue != 100 || value.step != 1 {
		t.Fatalf("slider defaults = %#v", value)
	}
	if value.rangeMode || value.orientation != SliderHorizontal {
		t.Fatal("single slider should default to horizontal single-value mode")
	}
}

func TestRangeSliderSortsAndSnapsValues(t *testing.T) {
	values := RangeSlider("price", 83, 17).Range(0, 100).Step(5).resolvedValues()
	if values.lower != 15 || values.upper != 85 {
		t.Fatalf("range values = %v to %v, want 15 to 85", values.lower, values.upper)
	}
}

func TestSliderClampsInvalidNumericValues(t *testing.T) {
	values := Slider("value", 150).Range(0, 100).resolvedValues()
	if values.lower != 100 {
		t.Fatalf("clamped value = %v, want 100", values.lower)
	}
	values = Slider("value", math.NaN()).resolvedValues()
	if values.lower != 0 {
		t.Fatalf("NaN value = %v, want 0", values.lower)
	}
}

func TestSliderRejectsInvalidRangeAndStep(t *testing.T) {
	assertSliderPanic(t, func() { Slider("value", 0).Range(10, 10) })
	assertSliderPanic(t, func() { Slider("value", 0).Range(math.NaN(), 10) })
	assertSliderPanic(t, func() { Slider("value", 0).Step(0) })
	assertSliderPanic(t, func() { Slider("value", 0).Step(math.Inf(1)) })
}

func assertSliderPanic(t *testing.T, fn func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Fatal("expected slider configuration panic")
		}
	}()
	fn()
}

func TestSliderOptionsKeepValueSemantics(t *testing.T) {
	base := Slider("volume", 20)
	styled := base.
		Label("Volume").
		ShowValue().
		ValueText("20 percent").
		FormatValue(func(float64) string { return "formatted" }).
		Orientation(SliderVertical).
		Disabled(true).
		OnChange(func(float64) {}).
		OnRangeChange(func(float64, float64) {})
	if base.label != "" || base.showValue || base.orientation != SliderHorizontal || base.disabled || base.onChange != nil {
		t.Fatal("slider options mutated the original value")
	}
	if styled.label != "Volume" || !styled.showValue || styled.valueText != "20 percent" || !styled.hasValueText {
		t.Fatal("slider text options were not retained")
	}
	if styled.orientation != SliderVertical || !styled.disabled || styled.onChange == nil || styled.onRangeChange == nil || styled.formatValue == nil {
		t.Fatal("slider behavior options were not retained")
	}
	if got := base.Vertical().orientation; got != SliderVertical {
		t.Fatalf("vertical orientation = %v", got)
	}
}

func TestSliderOutputFormatting(t *testing.T) {
	single := Slider("volume", 30).ShowValue()
	if got := single.outputText(single.resolvedValues()); got != "30" {
		t.Fatalf("single output = %q, want 30", got)
	}
	ranged := RangeSlider("price", 100, 500).
		Range(0, 1000).
		ShowValue().
		FormatValue(func(value float64) string { return "$" + single.format(value) })
	if got := ranged.outputText(ranged.resolvedValues()); got != "$100 - $500" {
		t.Fatalf("range output = %q, want $100 - $500", got)
	}
	custom := single.ValueText("Quiet")
	if got := custom.outputText(custom.resolvedValues()); got != "Quiet" {
		t.Fatalf("custom output = %q, want Quiet", got)
	}
}

func TestSliderDecimalStepDoesNotLeakFloatingPointNoise(t *testing.T) {
	value := Slider("decimal", .3).Range(0, 1).Step(.1).ShowValue()
	if got := value.outputText(value.resolvedValues()); got != "0.3" {
		t.Fatalf("decimal output = %q, want 0.3", got)
	}
}

func TestSliderHeroUIGeometry(t *testing.T) {
	horizontal := newSliderGeometry(image.Pt(300, 20), layout.Horizontal, 12, .25, .75, true, 24, 4)
	if horizontal.centers[0] != image.Pt(81, 10) || horizontal.centers[1] != image.Pt(219, 10) {
		t.Fatalf("horizontal centers = %v", horizontal.centers)
	}
	if horizontal.thumbRects[0].Size() != image.Pt(28, 20) {
		t.Fatalf("horizontal thumb = %v, want (28,20)", horizontal.thumbRects[0].Size())
	}
	vertical := newSliderGeometry(image.Pt(20, 300), layout.Vertical, 12, .25, .75, true, 24, 4)
	if vertical.centers[0] != image.Pt(10, 219) || vertical.centers[1] != image.Pt(10, 81) {
		t.Fatalf("vertical centers = %v", vertical.centers)
	}
	if vertical.thumbRects[0].Size() != image.Pt(20, 28) {
		t.Fatalf("vertical thumb = %v, want (20,28)", vertical.thumbRects[0].Size())
	}
}

func TestSliderGeometryKeepsCentersInsideTinyConstraints(t *testing.T) {
	geometry := newSliderGeometry(image.Pt(10, 20), layout.Horizontal, 12, 0, 1, true, 24, 4)
	if geometry.centers[0].X < 0 || geometry.centers[0].X > 10 || geometry.centers[1].X < 0 || geometry.centers[1].X > 10 {
		t.Fatalf("tiny slider centers = %v, want centers within width", geometry.centers)
	}
	if geometry.edge != 5 || geometry.inner != 0 {
		t.Fatalf("tiny slider geometry edge=%d inner=%d, want edge=5 inner=0", geometry.edge, geometry.inner)
	}
}

func TestSliderFillGeometry(t *testing.T) {
	geometry := newSliderGeometry(image.Pt(300, 20), layout.Horizontal, 12, .25, .75, true, 24, 4)
	if got := sliderFillRect(geometry, true); got != image.Rect(81, 0, 219, 20) {
		t.Fatalf("range fill = %v", got)
	}
	if got := sliderFillRect(geometry, false); got != image.Rect(0, 0, 81, 20) {
		t.Fatalf("single fill = %v", got)
	}
}

func TestDisabledSliderStyleUsesThemeOpacity(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	enabled := sliderStyleFor(&activeTheme, false)
	disabled := sliderStyleFor(&activeTheme, true)
	if enabled.fill != activeTheme.Palette.Accent || enabled.track != activeTheme.Palette.SurfaceRaised {
		t.Fatalf("enabled style = %#v", enabled)
	}
	if disabled.fill.A >= enabled.fill.A || disabled.track.A >= enabled.track.A {
		t.Fatal("disabled style should reduce control opacity")
	}
	if disabled.focus != (color.NRGBA{}) {
		t.Fatal("disabled slider should not draw a focus ring")
	}
}

func TestSliderDefaultAndVerticalLayout(t *testing.T) {
	ctx := sliderTestContext()
	var ops op.Ops
	dims := Slider("horizontal", 30).Layout(ctx, layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 200)},
		Ops:         &ops,
	})
	if dims.Size != image.Pt(300, 20) {
		t.Fatalf("horizontal size = %v, want (300,20)", dims.Size)
	}

	ctx = sliderTestContext()
	ops.Reset()
	dims = Slider("vertical", 30).Vertical().Layout(ctx, layout.Context{
		Constraints: layout.Exact(image.Pt(80, 300)),
		Ops:         &ops,
	})
	if dims.Size != image.Pt(80, 300) {
		t.Fatalf("vertical size = %v, want (80,300)", dims.Size)
	}
}

func TestSliderRegistersFieldFocus(t *testing.T) {
	ctx := sliderTestContext()
	frame.BeginFrame(ctx)
	var ops op.Ops
	Slider("volume", 30).Layout(ctx, layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 100)},
		Ops:         &ops,
	})
	frame.EndFrame(ctx)
	if !frame.HasFieldFocus(ctx, "volume") {
		t.Fatal("slider field focus was not retained at frame end")
	}
}

func TestSliderArrowKeyUsesConfiguredStep(t *testing.T) {
	ctx := sliderTestContext()
	router := new(input.Router)
	value := 30.0
	widget := func() SliderWidget {
		return Slider("volume", value).Step(5).OnChange(func(next float64) { value = next })
	}
	layoutSliderFrame(ctx, router, widget(), time.Unix(1, 0), image.Pt(300, 100))
	state := sliderTestState(ctx, "volume")
	router.Source().Execute(key.FocusCmd{Tag: &state.lowerThumb.clickable})
	layoutSliderFrame(ctx, router, widget(), time.Unix(1, int64(time.Millisecond)), image.Pt(300, 100))
	router.Queue(key.Event{Name: key.NameRightArrow, State: key.Press})
	layoutSliderFrame(ctx, router, widget(), time.Unix(1, int64(2*time.Millisecond)), image.Pt(300, 100))
	if value != 35 {
		t.Fatalf("keyboard value = %v, want 35", value)
	}
}

func TestSliderPointerChangesSingleValue(t *testing.T) {
	ctx := sliderTestContext()
	router := new(input.Router)
	value := 0.0
	widget := func() SliderWidget {
		return Slider("volume", value).OnChange(func(next float64) { value = next })
	}
	layoutSliderFrame(ctx, router, widget(), time.Unix(1, 0), image.Pt(300, 100))
	pressSliderAt(router, 1, f32.Pt(150, 10))
	layoutSliderFrame(ctx, router, widget(), time.Unix(1, int64(time.Millisecond)), image.Pt(300, 100))
	if value != 50 {
		t.Fatalf("pointer value = %v, want 50", value)
	}
}

func TestSliderPointerFocusStaysHiddenUntilKeyboardInput(t *testing.T) {
	ctx := sliderTestContext()
	router := new(input.Router)
	value := 0.0
	widget := func() SliderWidget {
		return Slider("volume", value).OnChange(func(next float64) { value = next })
	}
	layoutSliderFrame(ctx, router, widget(), time.Unix(1, 0), image.Pt(300, 100))
	pressSliderAt(router, 1, f32.Pt(150, 10))
	layoutSliderFrame(ctx, router, widget(), time.Unix(1, int64(time.Millisecond)), image.Pt(300, 100))

	state := sliderTestState(ctx, "volume")
	if !state.lowerThumb.pointerFocus || !state.lowerThumb.pointerFocusPending {
		t.Fatal("pointer focus origin was not retained while the focus command was pending")
	}
	layoutSliderFrame(ctx, router, widget(), time.Unix(1, int64(2*time.Millisecond)), image.Pt(300, 100))
	if !state.lowerThumb.pointerFocus || state.lowerThumb.pointerFocusPending {
		t.Fatal("pointer focus origin was not retained after the deferred focus command")
	}

	router.Queue(key.Event{Name: key.NameRightArrow, State: key.Press})
	layoutSliderFrame(ctx, router, widget(), time.Unix(1, int64(3*time.Millisecond)), image.Pt(300, 100))
	if state.lowerThumb.pointerFocus || state.lowerThumb.pointerFocusPending {
		t.Fatal("keyboard input did not switch the thumb to keyboard-visible focus")
	}
	if value != 51 {
		t.Fatalf("keyboard value = %v, want 51", value)
	}
}

func TestRangeSliderPointerTargetsNearestThumb(t *testing.T) {
	ctx := sliderTestContext()
	router := new(input.Router)
	lower, upper := 20.0, 80.0
	widget := func() SliderWidget {
		return RangeSlider("price", lower, upper).OnRangeChange(func(nextLower, nextUpper float64) {
			lower, upper = nextLower, nextUpper
		})
	}
	layoutSliderFrame(ctx, router, widget(), time.Unix(1, 0), image.Pt(300, 100))
	pressSliderAt(router, 1, f32.Pt(100, 10))
	layoutSliderFrame(ctx, router, widget(), time.Unix(1, int64(time.Millisecond)), image.Pt(300, 100))
	if lower != 32 || upper != 80 {
		t.Fatalf("left press values = %v to %v, want 32 to 80", lower, upper)
	}

	pressSliderAt(router, 2, f32.Pt(200, 10))
	layoutSliderFrame(ctx, router, widget(), time.Unix(1, int64(2*time.Millisecond)), image.Pt(300, 100))
	if lower != 32 || upper != 68 {
		t.Fatalf("right press values = %v to %v, want 32 to 68", lower, upper)
	}
}

func TestSliderDragStartingOnThumbChangesValue(t *testing.T) {
	ctx := sliderTestContext()
	router := new(input.Router)
	value := 20.0
	widget := func() SliderWidget {
		return Slider("volume", value).OnChange(func(next float64) { value = next })
	}
	layoutSliderFrame(ctx, router, widget(), time.Unix(1, 0), image.Pt(300, 100))
	router.Queue(pointer.Event{
		Kind:      pointer.Press,
		Source:    pointer.Mouse,
		PointerID: 1,
		Buttons:   pointer.ButtonPrimary,
		Position:  f32.Pt(67, 10),
	})
	layoutSliderFrame(ctx, router, widget(), time.Unix(1, int64(time.Millisecond)), image.Pt(300, 100))
	router.Queue(pointer.Event{
		Kind:      pointer.Move,
		Source:    pointer.Mouse,
		PointerID: 1,
		Buttons:   pointer.ButtonPrimary,
		Position:  f32.Pt(150, 10),
	})
	layoutSliderFrame(ctx, router, widget(), time.Unix(1, int64(2*time.Millisecond)), image.Pt(300, 100))
	if value != 50 {
		t.Fatalf("dragged value = %v, want 50", value)
	}
}

func TestVerticalSliderMapsTopToMaximum(t *testing.T) {
	ctx := sliderTestContext()
	router := new(input.Router)
	value := 0.0
	widget := func() SliderWidget {
		return Slider("intensity", value).Vertical().OnChange(func(next float64) { value = next })
	}
	layoutSliderFrame(ctx, router, widget(), time.Unix(1, 0), image.Pt(20, 300))
	pressSliderAt(router, 1, f32.Pt(10, 12))
	layoutSliderFrame(ctx, router, widget(), time.Unix(1, int64(time.Millisecond)), image.Pt(20, 300))
	if value != 100 {
		t.Fatalf("vertical top value = %v, want 100", value)
	}
}

func TestDisabledSliderIgnoresPointerInput(t *testing.T) {
	ctx := sliderTestContext()
	router := new(input.Router)
	changed := false
	widget := Slider("disabled", 30).Disabled(true).OnChange(func(float64) { changed = true })
	layoutSliderFrame(ctx, router, widget, time.Unix(1, 0), image.Pt(300, 100))
	pressSliderAt(router, 1, f32.Pt(200, 10))
	layoutSliderFrame(ctx, router, widget, time.Unix(1, int64(time.Millisecond)), image.Pt(300, 100))
	if changed {
		t.Fatal("disabled slider dispatched a pointer change")
	}
}

func sliderTestContext() *frame.Context {
	return frame.New(nil, nil, locale.LanguageAuto)
}

func sliderTestState(ctx *frame.Context, key string) *sliderState {
	value, _ := frame.PeekState[sliderState](ctx, key, stateSlotSlider)
	return value
}

func layoutSliderFrame(ctx *frame.Context, router *input.Router, widget SliderWidget, now time.Time, size image.Point) layout.Dimensions {
	frame.BeginFrameWithViewport(ctx, size)
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: size},
		Source:      router.Source(),
		Ops:         &ops,
		Now:         now,
	}
	dims := widget.Layout(ctx, gtx)
	frame.ApplyFrameCommands(ctx, gtx)
	frame.EndFrame(ctx)
	router.Frame(gtx.Ops)
	return dims
}

func pressSliderAt(router *input.Router, id pointer.ID, position f32.Point) {
	router.Queue(
		pointer.Event{
			Kind:      pointer.Press,
			Source:    pointer.Mouse,
			PointerID: id,
			Buttons:   pointer.ButtonPrimary,
			Position:  position,
		},
		pointer.Event{
			Kind:      pointer.Release,
			Source:    pointer.Mouse,
			PointerID: id,
			Position:  position,
		},
	)
}
