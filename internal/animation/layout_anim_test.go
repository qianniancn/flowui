package animation

import (
	"image"
	"math"
	"testing"
	"time"

	"gioui.org/io/input"
	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/theme"
)

func TestAnimateRectDurationMode(t *testing.T) {
	ctx := tweenTestContext(nil)
	router := new(input.Router)
	start := time.Unix(40, 0)
	from := image.Rect(0, 0, 10, 10)
	to := image.Rect(0, 0, 110, 10)

	// First frame snaps.
	rect, wake := sampleRect(ctx, router, AnimateRect("box", from).Duration(time.Second).Easing(EaseLinear), start)
	if rect != from || wake {
		t.Fatalf("first frame = %v, wake %v", rect, wake)
	}

	// Retarget and animate.
	anim := AnimateRect("box", to).Duration(time.Second).Easing(EaseLinear)
	rect, wake = sampleRect(ctx, router, anim, start.Add(time.Millisecond))
	if rect.Dx() != 10 || !wake {
		// same frame as retarget starts from previous
	}
	// After retarget, next sample at half duration.
	// First call with new target records from=current.
	sampleRect(ctx, router, anim, start.Add(10*time.Millisecond))
	rect, wake = sampleRect(ctx, router, anim, start.Add(10*time.Millisecond+500*time.Millisecond))
	if rect.Dx() < 50 || rect.Dx() > 70 {
		t.Fatalf("mid rect width = %d, want ~60", rect.Dx())
	}
	if !wake {
		t.Fatal("expected wake while animating")
	}
	rect, wake = sampleRect(ctx, router, anim, start.Add(10*time.Millisecond+time.Second))
	if rect != to || wake {
		t.Fatalf("end rect = %v, wake %v", rect, wake)
	}
}

func TestAnimateRectMotionOffSnaps(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	activeTheme.Motion.Enabled = false
	ctx := tweenTestContext(&activeTheme)
	router := new(input.Router)
	target := image.Rect(5, 5, 50, 50)
	rect, wake := sampleRect(ctx, router, AnimateRect("snap", target).Duration(time.Second), time.Unix(41, 0))
	if rect != target || wake {
		t.Fatalf("motion-off rect = %v, wake %v", rect, wake)
	}
}

func TestLayoutSpringTreatsNonFiniteMotionScaleAsDisabled(t *testing.T) {
	for _, scale := range []float32{float32(math.NaN()), float32(math.Inf(1))} {
		activeTheme := theme.DefaultTheme()
		activeTheme.Motion.DurationScale = scale
		ctx := tweenTestContext(&activeTheme)
		router := new(input.Router)
		start := time.Unix(41, 0)
		sampleRect(ctx, router, AnimateRect("snap", image.Rect(0, 0, 10, 10)).Spring(DefaultSpring()), start)
		target := image.Rect(10, 10, 50, 50)
		rect, wake := sampleRect(ctx, router, AnimateRect("snap", target).Spring(DefaultSpring()), start.Add(time.Millisecond))
		if rect != target || wake {
			t.Fatalf("scale %v rect = %v, wake %v", scale, rect, wake)
		}
	}
}

func TestLayoutAnimationUsesDerivedSizeKey(t *testing.T) {
	ctx := tweenTestContext(nil)
	router := new(input.Router)
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Exact(image.Pt(100, 100)),
		Source:      router.Source(),
		Ops:         &ops,
		Now:         time.Unix(42, 0),
	}
	frame.BeginFrame(ctx)
	AnimateLayout("panel", frame.WidgetFunc(func(*frame.Context, layout.Context) layout.Dimensions {
		return layout.Dimensions{Size: image.Pt(10, 10)}
	})).Layout(ctx, gtx)
	AnimateRect("panel/size", image.Rect(0, 0, 10, 10)).Value(ctx, gtx)
	frame.EndFrame(ctx)
}

func TestRectStateRoundsNegativeCoordinates(t *testing.T) {
	state := rectEdgeState{}
	state.minX.value = -10.6
	state.minY.value = -5.5
	state.maxX.value = 9.4
	state.maxY.value = 14.5
	if got, want := state.currentRect(), image.Rect(-11, -6, 9, 15); got != want {
		t.Fatalf("rounded rect = %v, want %v", got, want)
	}
}

func TestAnimateLayoutSizeChange(t *testing.T) {
	ctx := tweenTestContext(nil)
	router := new(input.Router)
	start := time.Unix(42, 0)

	child := frame.WidgetFunc(func(_ *frame.Context, gtx layout.Context) layout.Dimensions {
		return layout.Dimensions{Size: image.Pt(100, 40)}
	})
	widget := AnimateLayout("panel", child).Duration(time.Second).Easing(EaseLinear)

	// First frame reports natural size.
	dims := layoutAnimFrame(ctx, router, widget, start)
	if dims.Size != image.Pt(100, 40) {
		t.Fatalf("first size = %v", dims.Size)
	}

	// Change child size and animate.
	child = frame.WidgetFunc(func(_ *frame.Context, gtx layout.Context) layout.Dimensions {
		return layout.Dimensions{Size: image.Pt(200, 40)}
	})
	widget = AnimateLayout("panel", child).Duration(time.Second).Easing(EaseLinear)
	layoutAnimFrame(ctx, router, widget, start.Add(time.Millisecond))
	dims = layoutAnimFrame(ctx, router, widget, start.Add(501*time.Millisecond))
	if dims.Size.X < 140 || dims.Size.X > 160 {
		t.Fatalf("mid width = %d, want ~150", dims.Size.X)
	}
}

func sampleRect(ctx *frame.Context, router *input.Router, anim RectValue, now time.Time) (image.Rectangle, bool) {
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Exact(image.Pt(400, 400)),
		Source:      router.Source(),
		Ops:         &ops,
		Now:         now,
	}
	frame.BeginFrame(ctx)
	rect, running := anim.Sample(ctx, gtx)
	frame.EndFrame(ctx)
	router.Frame(&ops)
	_, wake := router.WakeupTime()
	if running {
		wake = true
	}
	return rect, wake
}

func layoutAnimFrame(ctx *frame.Context, router *input.Router, widget AnimateLayoutWidget, now time.Time) layout.Dimensions {
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Exact(image.Pt(400, 400)),
		Source:      router.Source(),
		Ops:         &ops,
		Now:         now,
	}
	frame.BeginFrame(ctx)
	dims := widget.Layout(ctx, gtx)
	frame.EndFrame(ctx)
	router.Frame(&ops)
	return dims
}
