package slider

import (
	"image"
	"image/color"
	"math"
	"testing"
	"time"

	"gioui.org/f32"
	"gioui.org/gpu/headless"
	"gioui.org/io/input"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/testtheme"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

func TestSliderDefaults(t *testing.T) {
	value := New("volume", 30)
	if value.key != "volume" || value.value != 30 || value.minValue != 0 || value.maxValue != 100 || value.step != 1 {
		t.Fatalf("slider defaults = %#v", value)
	}
	if value.rangeMode || value.orientation != Horizontal {
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
	values := New("value", 150).Range(0, 100).resolvedValues()
	if values.lower != 100 {
		t.Fatalf("clamped value = %v, want 100", values.lower)
	}
	values = New("value", math.NaN()).resolvedValues()
	if values.lower != 0 {
		t.Fatalf("NaN value = %v, want 0", values.lower)
	}
}

func TestSliderRejectsInvalidRangeAndStep(t *testing.T) {
	assertSliderPanic(t, func() { New("value", 0).Range(10, 10) })
	assertSliderPanic(t, func() { New("value", 0).Range(math.NaN(), 10) })
	assertSliderPanic(t, func() { New("value", 0).Step(0) })
	assertSliderPanic(t, func() { New("value", 0).Step(math.Inf(1)) })
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

func TestSliderOutputFormatting(t *testing.T) {
	single := New("volume", 30).ShowValue()
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
	value := New("decimal", .3).Range(0, 1).Step(.1).ShowValue()
	if got := value.outputText(value.resolvedValues()); got != "0.3" {
		t.Fatalf("decimal output = %q, want 0.3", got)
	}
}

func TestSliderGeometry(t *testing.T) {
	horizontal := newSliderGeometry(image.Pt(300, 20), layout.Horizontal, 12, .25, .75, true, image.Pt(28, 20))
	if horizontal.centers[0] != image.Pt(81, 10) || horizontal.centers[1] != image.Pt(219, 10) {
		t.Fatalf("horizontal centers = %v", horizontal.centers)
	}
	if horizontal.thumbRects[0].Size() != image.Pt(28, 20) {
		t.Fatalf("horizontal thumb = %v, want (28,20)", horizontal.thumbRects[0].Size())
	}
	vertical := newSliderGeometry(image.Pt(20, 300), layout.Vertical, 12, .25, .75, true, image.Pt(20, 28))
	if vertical.centers[0] != image.Pt(10, 219) || vertical.centers[1] != image.Pt(10, 81) {
		t.Fatalf("vertical centers = %v", vertical.centers)
	}
	if vertical.thumbRects[0].Size() != image.Pt(20, 28) {
		t.Fatalf("vertical thumb = %v, want (20,28)", vertical.thumbRects[0].Size())
	}
}

func TestSliderEdgeInsetFollowsSmallerCustomThumb(t *testing.T) {
	geometry := newSliderGeometry(image.Pt(300, 4), layout.Horizontal, 12, 1, 1, false, image.Pt(16, 16))
	if geometry.edge != 8 || geometry.centers[0].X != 292 || geometry.thumbRects[0].Max.X != 300 {
		t.Fatalf("compact geometry edge=%d center=%v thumb=%v, want thumb flush with track end", geometry.edge, geometry.centers[0], geometry.thumbRects[0])
	}
}

func TestSliderGeometryKeepsCentersInsideTinyConstraints(t *testing.T) {
	geometry := newSliderGeometry(image.Pt(10, 20), layout.Horizontal, 12, 0, 1, true, image.Pt(28, 20))
	if geometry.centers[0].X < 0 || geometry.centers[0].X > 10 || geometry.centers[1].X < 0 || geometry.centers[1].X > 10 {
		t.Fatalf("tiny slider centers = %v, want centers within width", geometry.centers)
	}
	if geometry.edge != 5 || geometry.inner != 0 {
		t.Fatalf("tiny slider geometry edge=%d inner=%d, want edge=5 inner=0", geometry.edge, geometry.inner)
	}
}

func TestSliderFillGeometry(t *testing.T) {
	geometry := newSliderGeometry(image.Pt(300, 20), layout.Horizontal, 12, .25, .75, true, image.Pt(28, 20))
	if got := sliderFillRect(geometry, true); got != image.Rect(81, 0, 219, 20) {
		t.Fatalf("range fill = %v", got)
	}
	if got := sliderFillRect(geometry, false); got != image.Rect(0, 0, 81, 20) {
		t.Fatalf("single fill = %v", got)
	}
	if got := sliderFillRect(newSliderGeometry(image.Pt(300, 20), layout.Horizontal, 12, 0, 0, false, image.Pt(28, 20)), false); !got.Empty() {
		t.Fatalf("horizontal minimum fill = %v, want empty", got)
	}
	if got := sliderFillRect(newSliderGeometry(image.Pt(300, 20), layout.Horizontal, 12, 1, 1, false, image.Pt(28, 20)), false); got != image.Rect(0, 0, 300, 20) {
		t.Fatalf("horizontal maximum fill = %v, want full track", got)
	}
	vertical := newSliderGeometry(image.Pt(20, 300), layout.Vertical, 12, 0, 0, false, image.Pt(20, 28))
	if got := sliderFillRect(vertical, false); !got.Empty() {
		t.Fatalf("vertical minimum fill = %v, want empty", got)
	}
	vertical = newSliderGeometry(image.Pt(20, 300), layout.Vertical, 12, 1, 1, false, image.Pt(20, 28))
	if got := sliderFillRect(vertical, false); got != image.Rect(0, 0, 20, 300) {
		t.Fatalf("vertical maximum fill = %v, want full track", got)
	}
}

func TestSliderResolvesSemanticParts(t *testing.T) {
	track := color.NRGBA{R: 1, A: 0xff}
	fill := color.NRGBA{G: 2, A: 0xff}
	thumb := color.NRGBA{B: 3, A: 0xff}
	custom := style.Style{}.
		Part(style.PartTrack, style.Style{}.
			Background(style.SolidColor{Color: track}).
			BorderWidth(2)).
		Part(style.PartFill, style.Style{}.Background(style.SolidColor{Color: fill})).
		Part(style.PartThumb, style.Style{}.
			Width(32).
			Height(24).
			Background(style.SolidColor{Color: thumb}).
			Radius(7))

	ctx := core.New(nil, testtheme.Behavior(), locale.LanguageAuto)
	var ops op.Ops
	styled := (New("volume", 50)).Style(custom)
	visual := resolveSliderProvider(core.ActiveTheme(ctx), Horizontal, styled.override)
	resolved := styled.resolveStyle(ctx, layout.Context{Constraints: layout.Constraints{Max: image.Pt(200, 40)}, Ops: &ops}, "volume", style.State{}, visual)
	if got := resolved.track.Paint.Background.(style.SolidColor).Color; got != track {
		t.Fatalf("track background = %#v", got)
	}
	if resolved.track.Paint.Border == nil || resolved.track.Paint.Border.Width == nil || *resolved.track.Paint.Border.Width != 2 {
		t.Fatalf("track border = %#v", resolved.track.Paint.Border)
	}
	if got := resolved.fill.Paint.Background.(style.SolidColor).Color; got != fill {
		t.Fatalf("fill background = %#v", got)
	}
	if resolved.fill.Paint.Radii == nil || resolved.fill.Paint.Radii.TopLeft != 12 || resolved.fill.Paint.Radii.BottomLeft != 12 || resolved.fill.Paint.Radii.TopRight != 0 || resolved.fill.Paint.Radii.BottomRight != 0 {
		t.Fatalf("horizontal fill radii = %#v", resolved.fill.Paint)
	}
	if got := resolved.thumb.Paint.Background.(style.SolidColor).Color; got != thumb {
		t.Fatalf("thumb background = %#v", got)
	}
	if resolved.thumb.Box == nil || resolved.thumb.Box.Width == nil || *resolved.thumb.Box.Width != 32 || resolved.thumb.Box.Height == nil || *resolved.thumb.Box.Height != 24 {
		t.Fatalf("thumb box = %#v", resolved.thumb.Box)
	}
	if resolved.thumb.Paint.Radius == nil || *resolved.thumb.Paint.Radius != 7 {
		t.Fatalf("thumb radius = %#v", resolved.thumb.Paint)
	}
	outer, inner, _, layered := sliderThumbLayers(resolved.thumb)
	if !layered || len(outer.Paint.Shadows) != 0 || len(inner.Paint.Shadows) != 2 {
		t.Fatalf("thumb shadow layers = outer %d, inner %d; want inner shadow", len(outer.Paint.Shadows), len(inner.Paint.Shadows))
	}
}

func TestSliderDefaultAndVerticalLayout(t *testing.T) {
	ctx := sliderTestContext()
	var ops op.Ops
	dims := New("horizontal", 30).Layout(ctx, layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 200)},
		Ops:         &ops,
	})
	if dims.Size != image.Pt(300, 20) {
		t.Fatalf("horizontal size = %v, want (300,20)", dims.Size)
	}

	ctx = sliderTestContext()
	ops.Reset()
	dims = New("vertical", 30).Vertical().Layout(ctx, layout.Context{
		Constraints: layout.Exact(image.Pt(80, 300)),
		Ops:         &ops,
	})
	if dims.Size != image.Pt(80, 300) {
		t.Fatalf("vertical size = %v, want (80,300)", dims.Size)
	}
}

func TestSliderThumbOuterEdgeDoesNotLightenFilledTrack(t *testing.T) {
	window, err := headless.NewWindow(300, 20)
	if err != nil {
		t.Skipf("headless renderer unavailable: %v", err)
	}
	defer window.Release()

	ctx := sliderTestContext()
	var router input.Router
	var ops op.Ops
	New("volume", 50).Layout(ctx, layout.Context{
		Constraints: layout.Exact(image.Pt(300, 20)),
		Source:      router.Source(),
		Ops:         &ops,
	})
	if err := window.Frame(&ops); err != nil {
		t.Fatal(err)
	}
	pixels := image.NewRGBA(image.Rect(0, 0, 300, 20))
	if err := window.Screenshot(pixels); err != nil {
		t.Fatal(err)
	}
	accent := theme.NeutralTheme().Palette.Accent
	lightLimit := int(accent.R) + 16
	for y := 2; y <= 8; y++ {
		lightSeen := false
		for x := 120; x <= 150; x++ {
			got := color.NRGBAModel.Convert(pixels.At(x, y)).(color.NRGBA)
			light := int(got.R) > lightLimit
			if lightSeen && !light {
				t.Fatalf("thumb row %d returns to blue at x=%d after a light edge pixel; white fringe is visible", y, x)
			}
			lightSeen = lightSeen || light
		}
	}
}

func TestSliderLabelPartUsesCommonBoxRenderer(t *testing.T) {
	gtx := layout.Context{Constraints: layout.Constraints{Max: image.Pt(300, 200)}, Ops: new(op.Ops)}
	base := New("volume", 30).
		Label("Volume").
		Layout(sliderTestContext(), gtx)
	gtx.Ops = new(op.Ops)
	styled := (New("volume", 30).Label("Volume")).Style(style.Style{}.Part(style.PartLabel, style.Style{}.PaddingY(7))).Layout(sliderTestContext(), gtx)
	if styled.Size.Y != base.Size.Y+14 {
		t.Fatalf("styled label height = %d, want %d", styled.Size.Y, base.Size.Y+14)
	}
}

func TestSliderTrackPartUsesCommonBoxRenderer(t *testing.T) {
	gtx := layout.Context{Constraints: layout.Constraints{Max: image.Pt(300, 200)}, Ops: new(op.Ops)}
	base := New("volume", 30).Layout(sliderTestContext(), gtx)
	gtx.Ops = new(op.Ops)
	styled := (New("volume", 30)).Style(style.Style{}.Part(style.PartTrack, style.Style{}.MarginY(7))).Layout(sliderTestContext(), gtx)
	if styled.Size.Y != base.Size.Y+14 {
		t.Fatalf("styled track height = %d, want %d", styled.Size.Y, base.Size.Y+14)
	}
}

func TestSliderRegistersFieldFocus(t *testing.T) {
	ctx := sliderTestContext()
	core.BeginFrame(ctx)
	var ops op.Ops
	New("volume", 30).Layout(ctx, layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 100)},
		Ops:         &ops,
	})
	core.EndFrame(ctx)
	if !core.HasFieldFocus(ctx, "volume") {
		t.Fatal("slider field focus was not retained at frame end")
	}
}

func TestSliderArrowKeyUsesConfiguredStep(t *testing.T) {
	ctx := sliderTestContext()
	router := new(input.Router)
	value := 30.0
	widget := func() Widget {
		return New("volume", value).Step(5).OnChange(func(next float64) { value = next })
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
	widget := func() Widget {
		return New("volume", value).OnChange(func(next float64) { value = next })
	}
	layoutSliderFrame(ctx, router, widget(), time.Unix(1, 0), image.Pt(300, 100))
	pressSliderAt(router, 1, f32.Pt(150, 10))
	layoutSliderFrame(ctx, router, widget(), time.Unix(1, int64(time.Millisecond)), image.Pt(300, 100))
	if value != 50 {
		t.Fatalf("pointer value = %v, want 50", value)
	}
}

func TestCompactSliderKeepsMinimumHitTarget(t *testing.T) {
	ctx := sliderTestContext()
	router := new(input.Router)
	value := 0.0
	custom := style.Style{}.
		Part(style.PartTrack, style.Style{}.Height(4).MarginY(6)).
		Part(style.PartThumb, style.Style{}.Width(16).Height(16).Radius(8))
	widget := func() sliderLayoutWidget {
		return (New("compact-hit", value).OnChange(func(next float64) { value = next })).Style(custom)
	}
	layoutSliderFrame(ctx, router, widget(), time.Unix(1, 0), image.Pt(300, 40))
	pressSliderAt(router, 1, f32.Pt(150, 18))
	layoutSliderFrame(ctx, router, widget(), time.Unix(1, int64(time.Millisecond)), image.Pt(300, 40))
	if value != 50 {
		t.Fatalf("compact slider value = %v, want 50 from click outside its 4px visual track", value)
	}
}

func TestSliderCursorStyleCoversExpandedHitTarget(t *testing.T) {
	ctx := sliderTestContext()
	router := new(input.Router)
	cursorStyle := style.Style{}.
		Cursor(pointer.CursorDefault).
		OnAny(style.Pressed|style.Dragging, style.Style{}.Cursor(pointer.CursorPointer))
	custom := style.Style{}.
		Part(style.PartTrack, style.Style{}.
			Height(4).
			MarginY(6).
			Cursor(pointer.CursorDefault).
			OnAny(style.Pressed|style.Dragging, style.Style{}.Cursor(pointer.CursorPointer))).
		Part(style.PartThumb, cursorStyle)
	widget := (New("compact-cursor", 50)).Style(custom)
	position := f32.Pt(150, 16)
	router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, PointerID: 1, Position: position})
	layoutSliderFrame(ctx, router, widget, time.Unix(1, 0), image.Pt(300, 40))
	if got := router.Cursor(); got != pointer.CursorDefault {
		t.Fatalf("released compact cursor = %v, want default", got)
	}
	router.Queue(pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: 1, Buttons: pointer.ButtonPrimary, Position: position})
	layoutSliderFrame(ctx, router, widget, time.Unix(1, int64(time.Millisecond)), image.Pt(300, 40))
	if got := router.Cursor(); got != pointer.CursorPointer {
		t.Fatalf("pressed compact cursor = %v, want pointer", got)
	}
	router.Queue(pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, PointerID: 1, Position: position})
	layoutSliderFrame(ctx, router, widget, time.Unix(1, int64(2*time.Millisecond)), image.Pt(300, 40))
	if got := router.Cursor(); got != pointer.CursorDefault {
		t.Fatalf("released compact cursor = %v, want default", got)
	}
}

func TestSliderPointerFocusStaysHiddenUntilKeyboardInput(t *testing.T) {
	ctx := sliderTestContext()
	router := new(input.Router)
	value := 0.0
	widget := func() Widget {
		return New("volume", value).OnChange(func(next float64) { value = next })
	}
	layoutSliderFrame(ctx, router, widget(), time.Unix(1, 0), image.Pt(300, 100))
	pressSliderAt(router, 1, f32.Pt(150, 10))
	layoutSliderFrame(ctx, router, widget(), time.Unix(1, int64(time.Millisecond)), image.Pt(300, 100))

	state := sliderTestState(ctx, "volume")
	if core.FocusVisible(ctx, &state.lowerThumb.clickable, true) {
		t.Fatal("pointer focus was visible while the focus command was pending")
	}
	layoutSliderFrame(ctx, router, widget(), time.Unix(1, int64(2*time.Millisecond)), image.Pt(300, 100))
	if core.FocusVisible(ctx, &state.lowerThumb.clickable, true) {
		t.Fatal("pointer focus was visible after the deferred focus command")
	}

	router.Queue(key.Event{Name: key.NameRightArrow, State: key.Press})
	layoutSliderFrame(ctx, router, widget(), time.Unix(1, int64(3*time.Millisecond)), image.Pt(300, 100))
	if !core.FocusVisible(ctx, &state.lowerThumb.clickable, true) {
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
	widget := func() Widget {
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
	widget := func() Widget {
		return New("volume", value).OnChange(func(next float64) { value = next })
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
	widget := func() Widget {
		return New("intensity", value).Vertical().OnChange(func(next float64) { value = next })
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
	widget := New("disabled", 30).Disabled(true).OnChange(func(float64) { changed = true })
	layoutSliderFrame(ctx, router, widget, time.Unix(1, 0), image.Pt(300, 100))
	pressSliderAt(router, 1, f32.Pt(200, 10))
	layoutSliderFrame(ctx, router, widget, time.Unix(1, int64(time.Millisecond)), image.Pt(300, 100))
	if changed {
		t.Fatal("disabled slider dispatched a pointer change")
	}
}

func sliderTestContext() *core.Context {
	return core.New(nil, testtheme.Behavior(), locale.LanguageAuto)
}

func sliderTestState(ctx *core.Context, key string) *sliderState {
	value, _ := core.PeekState[sliderState](ctx, key, stateSlotSlider)
	return value
}

type sliderLayoutWidget interface {
	Layout(*core.Context, layout.Context) layout.Dimensions
}

func layoutSliderFrame(ctx *core.Context, router *input.Router, widget sliderLayoutWidget, now time.Time, size image.Point) layout.Dimensions {
	core.BeginFrameWithViewport(ctx, size)
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: size},
		Source:      router.Source(),
		Ops:         &ops,
		Now:         now,
	}
	dims := widget.Layout(ctx, gtx)
	core.ApplyFrameCommands(ctx, gtx)
	core.EndFrame(ctx)
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
