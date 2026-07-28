package colorpicker

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
	"github.com/qianniancn/flowui/internal/frame"
	"github.com/qianniancn/flowui/internal/locale"
)

func TestColorWheelOptionsUseValueSemantics(t *testing.T) {
	base := ColorWheel("brand", color.NRGBA{R: 1, A: 255})
	configured := base.
		Size(180).
		Label("Brand color").
		Disabled(true).
		OnChange(func(color.NRGBA) {})

	if base.hasSize || base.size != 0 || base.label != "" || base.disabled || base.onChange != nil {
		t.Fatalf("configuring ColorWheel mutated base: %#v", base)
	}
	if !configured.hasSize || configured.size != 180 || configured.label != "Brand color" || !configured.disabled || configured.onChange == nil {
		t.Fatalf("configured ColorWheel = %#v", configured)
	}
}

func TestColorWheelUsesConfiguredAndConstrainedSize(t *testing.T) {
	ctx := frame.New(nil, nil, locale.LanguageEnglish)
	router := new(input.Router)
	now := time.Unix(1, 0)

	dimensions := layoutColorWidgetFrame(ctx, router, ColorWheel("configured", color.NRGBA{A: 255}).Size(164), now)
	if dimensions.Size != image.Pt(164, 164) {
		t.Fatalf("configured ColorWheel size = %v, want 164x164", dimensions.Size)
	}
}

func TestColorWheelPositionMapsCenterEdgeAndOutside(t *testing.T) {
	size := image.Pt(200, 200)
	current := hsvColor{h: .61, s: .4, v: .37, a: .42}

	center := colorWheelPosition(f32.Pt(100, 100), size, current)
	if center.h != current.h || center.s != 0 || center.v != current.v || center.a != current.a {
		t.Fatalf("center position = %#v, want retained hue with zero saturation", center)
	}

	right := colorWheelPosition(f32.Pt(200, 100), size, current)
	if right.h != 0 || right.s != 1 {
		t.Fatalf("right edge = %#v, want hue 0 and saturation 1", right)
	}

	top := colorWheelPosition(f32.Pt(100, 0), size, current)
	if math.Abs(top.h-.75) > 1e-9 || top.s != 1 {
		t.Fatalf("top edge = %#v, want hue .75 and saturation 1", top)
	}

	outside := colorWheelPosition(f32.Pt(-100, 100), size, current)
	if math.Abs(outside.h-.5) > 1e-9 || outside.s != 1 {
		t.Fatalf("outside position = %#v, want hue .5 and clamped saturation", outside)
	}
}

func TestColorWheelChangePreservesBrightnessAndAlpha(t *testing.T) {
	value := color.NRGBA{R: 40, G: 80, B: 120, A: 93}
	current := nrgbaToHSV(value)
	next := colorWheelPosition(f32.Pt(200, 100), image.Pt(200, 200), current)
	got := hsvToNRGBA(next)

	if got.A != value.A {
		t.Fatalf("ColorWheel alpha = %d, want %d", got.A, value.A)
	}
	if maximum := max(got.R, got.G, got.B); maximum != max(value.R, value.G, value.B) {
		t.Fatalf("ColorWheel brightness max channel = %d, want %d", maximum, max(value.R, value.G, value.B))
	}
}

func TestColorWheelImageCacheOnlyRebuildsForSizeChanges(t *testing.T) {
	var cache colorWheelImageCache
	if !cache.update(image.Pt(16, 16)) {
		t.Fatal("initial ColorWheel cache update did not build an image")
	}
	if cache.update(image.Pt(16, 16)) {
		t.Fatal("ColorWheel cache rebuilt for an unchanged size")
	}
	if !cache.update(image.Pt(24, 24)) {
		t.Fatal("ColorWheel cache did not rebuild after a size change")
	}
	if cache.op.Size() != image.Pt(24, 24) {
		t.Fatalf("ColorWheel cached image size = %v, want 24x24", cache.op.Size())
	}
}

func TestDisabledColorWheelDoesNotDispatchPointerChanges(t *testing.T) {
	ctx := frame.New(nil, nil, locale.LanguageEnglish)
	router := new(input.Router)
	initial := color.NRGBA{R: 4, G: 133, B: 247, A: 93}
	changes := 0
	wheel := ColorWheel("disabled", initial).Size(100).Disabled(true).OnChange(func(color.NRGBA) {
		changes++
	})
	start := time.Unix(1, 0)
	layoutColorWidgetFrame(ctx, router, wheel, start)

	router.Queue(pointer.Event{
		Kind:      pointer.Press,
		Source:    pointer.Mouse,
		PointerID: 1,
		Buttons:   pointer.ButtonPrimary,
		Position:  f32.Pt(100, 50),
	})
	layoutColorWidgetFrame(ctx, router, wheel, start.Add(time.Millisecond))
	if changes != 0 {
		t.Fatalf("disabled ColorWheel dispatched %d changes", changes)
	}
}

func TestColorWheelPointerDispatchesChange(t *testing.T) {
	ctx := frame.New(nil, nil, locale.LanguageEnglish)
	router := new(input.Router)
	initial := color.NRGBA{R: 4, G: 133, B: 247, A: 93}
	var changed color.NRGBA
	wheel := ColorWheel("pointer", initial).Size(100).OnChange(func(value color.NRGBA) {
		changed = value
	})
	start := time.Unix(1, 0)
	layoutColorWidgetFrame(ctx, router, wheel, start)
	router.Queue(pointer.Event{
		Kind:      pointer.Press,
		Source:    pointer.Mouse,
		PointerID: 1,
		Buttons:   pointer.ButtonPrimary,
		Position:  f32.Pt(99, 50),
	})
	layoutColorWidgetFrame(ctx, router, wheel, start.Add(time.Millisecond))

	if changed == (color.NRGBA{}) || changed == initial {
		t.Fatalf("pointer ColorWheel change = %#v", changed)
	}
	if changed.A != initial.A {
		t.Fatalf("pointer ColorWheel alpha = %d, want %d", changed.A, initial.A)
	}
}

func TestColorWheelKeyboardChangePreservesAlpha(t *testing.T) {
	ctx := frame.New(nil, nil, locale.LanguageEnglish)
	router := new(input.Router)
	initial := color.NRGBA{R: 4, G: 133, B: 247, A: 93}
	var changed color.NRGBA
	wheel := ColorWheel("keyboard", initial).Size(100).OnChange(func(value color.NRGBA) {
		changed = value
	})
	start := time.Unix(1, 0)
	layoutColorWidgetFrame(ctx, router, wheel, start)
	layoutColorWidgetFrame(ctx, router, wheel, start.Add(time.Millisecond))
	wheelState, ok := frame.PeekState[colorWheelState](ctx, "keyboard", stateSlotColorWheel)
	if !ok {
		t.Fatal("missing ColorWheel state")
	}
	router.Source().Execute(key.FocusCmd{Tag: &wheelState.control})
	if !router.Source().Focused(&wheelState.control) {
		t.Fatal("ColorWheel did not accept keyboard focus")
	}
	router.Queue(key.Event{Name: key.NameRightArrow, State: key.Press})
	if !router.Source().Focused(&wheelState.control) {
		t.Fatal("ColorWheel lost keyboard focus while queueing the key event")
	}
	layoutColorWidgetFrame(ctx, router, wheel, start.Add(2*time.Millisecond))

	if changed == (color.NRGBA{}) || changed == initial {
		t.Fatalf("keyboard ColorWheel change = %#v, focused=%v, retained=%#v", changed, router.Source().Focused(&wheelState.control), wheelState.color.hsv())
	}
	if changed.A != initial.A {
		t.Fatalf("keyboard ColorWheel alpha = %d, want %d", changed.A, initial.A)
	}
}
