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
	"github.com/qianniancn/FlowUI/internal/locale"
	"github.com/qianniancn/FlowUI/internal/theme"
)

func TestTweenSnapsOnFirstFrameByDefault(t *testing.T) {
	ctx := tweenTestContext(nil)
	router := new(input.Router)
	value, wake := tweenFrame(ctx, router, Tween("position", 1), time.Unix(1, 0))
	if value != 1 || wake {
		t.Fatalf("default first frame = %v, wake %v", value, wake)
	}
}

func TestTweenAnimatesExplicitInitialValue(t *testing.T) {
	ctx := tweenTestContext(nil)
	router := new(input.Router)
	start := time.Unix(2, 0)
	tween := Tween("position", 1).Initial(0).Duration(time.Second).Easing(EaseCubicInOut)
	if value, wake := tweenFrame(ctx, router, tween, start); value != 0 || !wake {
		t.Fatalf("entry start = %v, wake %v", value, wake)
	}
	if value, wake := tweenFrame(ctx, router, tween, start.Add(500*time.Millisecond)); value != 0.5 || !wake {
		t.Fatalf("entry midpoint = %v, wake %v", value, wake)
	}
	if value, wake := tweenFrame(ctx, router, tween, start.Add(time.Second)); value != 1 || wake {
		t.Fatalf("entry end = %v, wake %v", value, wake)
	}
}

func TestTweenTargetChangesRemainContinuous(t *testing.T) {
	ctx := tweenTestContext(nil)
	router := new(input.Router)
	start := time.Unix(3, 0)
	forward := Tween("position", 1).Initial(0).Duration(time.Second).Easing(EaseLinear)
	tweenFrame(ctx, router, forward, start)
	if value, _ := tweenFrame(ctx, router, forward, start.Add(500*time.Millisecond)); value != 0.5 {
		t.Fatalf("forward midpoint = %v", value)
	}
	reverse := Tween("position", 0).Duration(time.Second).Easing(EaseLinear)
	if value, _ := tweenFrame(ctx, router, reverse, start.Add(500*time.Millisecond)); value != 0.5 {
		t.Fatalf("reversal jumped to %v", value)
	}
	if value, _ := tweenFrame(ctx, router, reverse, start.Add(750*time.Millisecond)); value != 0.375 {
		t.Fatalf("reverse midpoint = %v, want 0.375", value)
	}
}

func TestTweenHonorsGlobalMotionControls(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	activeTheme.Motion.Enabled = false
	ctx := tweenTestContext(&activeTheme)
	router := new(input.Router)
	value, wake := tweenFrame(ctx, router, Tween("disabled", 1).Initial(0), time.Unix(4, 0))
	if value != 1 || wake {
		t.Fatalf("disabled motion = %v, wake %v", value, wake)
	}

	activeTheme = theme.DefaultTheme()
	activeTheme.Motion.DurationScale = 0.5
	ctx = tweenTestContext(&activeTheme)
	router = new(input.Router)
	start := time.Unix(5, 0)
	tween := Tween("scaled", 1).Initial(0).Duration(time.Second).Easing(EaseLinear)
	tweenFrame(ctx, router, tween, start)
	if value, wake := tweenFrame(ctx, router, tween, start.Add(500*time.Millisecond)); value != 1 || wake {
		t.Fatalf("scaled duration end = %v, wake %v", value, wake)
	}
}

func TestTweenInvalidatesUntilTimeCompletes(t *testing.T) {
	ctx := tweenTestContext(nil)
	router := new(input.Router)
	start := time.Unix(6, 0)
	tween := Tween("constant-output", 1).
		Initial(0).
		Duration(time.Second).
		Easing(func(float32) float32 { return 1 })
	tweenFrame(ctx, router, tween, start)
	if value, wake := tweenFrame(ctx, router, tween, start.Add(250*time.Millisecond)); value != 1 || !wake {
		t.Fatalf("early target value = %v, wake %v", value, wake)
	}
}

func TestTweenRevisionRestartsFromInitialValue(t *testing.T) {
	ctx := tweenTestContext(nil)
	router := new(input.Router)
	start := time.Unix(9, 0)
	first := Tween("revision", 1).Initial(0).Revision(1).Duration(time.Second).Easing(EaseLinear)
	tweenFrame(ctx, router, first, start)
	if value, _ := tweenFrame(ctx, router, first, start.Add(time.Second)); value != 1 {
		t.Fatalf("first revision end = %v", value)
	}
	second := Tween("revision", 1).Initial(0).Revision(2).Duration(time.Second).Easing(EaseLinear)
	if value, wake := tweenFrame(ctx, router, second, start.Add(2*time.Second)); value != 0 || !wake {
		t.Fatalf("second revision start = %v, wake %v", value, wake)
	}
}

func TestTweenRejectsInvalidConfigurationAndEasing(t *testing.T) {
	tests := []func(){
		func() { Tween("bad", float32(math.NaN())) },
		func() { Tween("bad", 1).Initial(float32(math.Inf(1))) },
		func() { Tween("bad", 1).Duration(-time.Second) },
		func() { Tween("bad", 1).Easing(nil) },
	}
	for _, run := range tests {
		func() {
			defer func() {
				if recover() == nil {
					t.Fatal("invalid tween configuration did not panic")
				}
			}()
			run()
		}()
	}

	ctx := tweenTestContext(nil)
	router := new(input.Router)
	start := time.Unix(7, 0)
	tween := Tween("bad-easing", 1).Initial(0).Duration(time.Second).Easing(func(float32) float32 {
		return float32(math.NaN())
	})
	tweenFrame(ctx, router, tween, start)
	defer func() {
		if recover() == nil {
			t.Fatal("non-finite easing output did not panic")
		}
	}()
	tweenFrame(ctx, router, tween, start.Add(time.Millisecond))
}

func TestTweenStateIsReleasedWhenUnused(t *testing.T) {
	ctx := tweenTestContext(nil)
	router := new(input.Router)
	tweenFrame(ctx, router, Tween("temporary", 1), time.Unix(8, 0))
	if frame.StateLen(ctx) != 1 {
		t.Fatalf("Tween retained states = %d, want 1", frame.StateLen(ctx))
	}
	frame.BeginFrame(ctx)
	frame.EndFrame(ctx)
	if frame.StateLen(ctx) != 0 {
		t.Fatalf("unused Tween states = %d, want 0", frame.StateLen(ctx))
	}
}

func tweenTestContext(activeTheme *theme.Theme) *frame.Context {
	return frame.New(nil, activeTheme, locale.LanguageEnglish)
}

func tweenFrame(ctx *frame.Context, router *input.Router, tween TweenValue, now time.Time) (float32, bool) {
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Exact(image.Pt(100, 100)),
		Source:      router.Source(),
		Ops:         &ops,
		Now:         now,
	}
	frame.BeginFrame(ctx)
	value := tween.Value(ctx, gtx)
	frame.EndFrame(ctx)
	router.Frame(&ops)
	_, wake := router.WakeupTime()
	return value, wake
}
