package animation

import (
	"image"
	"testing"
	"time"

	"gioui.org/io/input"
	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/theme"
)

func TestTimelineSamplesKeyframes(t *testing.T) {
	ctx := tweenTestContext(nil)
	router := new(input.Router)
	start := time.Unix(30, 0)
	tl := Timeline("pulse").
		Keyframe(0, 0).
		Keyframe(0.5, 1).
		Keyframe(1, 0).
		Duration(time.Second).
		Easing(EaseLinear)

	if value, wake := sampleTimeline(ctx, router, tl, start); value != 0 || !wake {
		t.Fatalf("timeline start = %v, wake %v", value, wake)
	}
	if value, wake := sampleTimeline(ctx, router, tl, start.Add(500*time.Millisecond)); value != 1 || !wake {
		t.Fatalf("timeline midpoint = %v, wake %v", value, wake)
	}
	if value, wake := sampleTimeline(ctx, router, tl, start.Add(time.Second)); value != 0 || wake {
		t.Fatalf("timeline end = %v, wake %v", value, wake)
	}
}

func TestTimelineDisabledSnapsToEnd(t *testing.T) {
	ctx := tweenTestContext(nil)
	router := new(input.Router)
	tl := Timeline("done").
		Keyframe(0, 0).
		Keyframe(1, 10).
		Duration(time.Second).
		Disabled(true)
	if value, wake := sampleTimeline(ctx, router, tl, time.Unix(31, 0)); value != 10 || wake {
		t.Fatalf("disabled timeline = %v, wake %v", value, wake)
	}
}

func TestTimelineHonorsMotionOff(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	activeTheme.Motion.Enabled = false
	ctx := tweenTestContext(&activeTheme)
	router := new(input.Router)
	tl := Timeline("off").
		Keyframe(0, 0).
		Keyframe(1, 5).
		Duration(time.Second)
	if value, wake := sampleTimeline(ctx, router, tl, time.Unix(32, 0)); value != 5 || wake {
		t.Fatalf("motion-off timeline = %v, wake %v", value, wake)
	}
}

func TestTimelineRebindsLoopAfterCompletion(t *testing.T) {
	ctx := tweenTestContext(nil)
	router := new(input.Router)
	start := time.Unix(33, 0)
	base := Timeline("loop").Keyframe(0, 0).Keyframe(1, 10).Duration(time.Second).Easing(EaseLinear)
	sampleTimeline(ctx, router, base, start)
	if value, wake := sampleTimeline(ctx, router, base, start.Add(time.Second)); value != 10 || wake {
		t.Fatalf("finished timeline = %v, wake %v", value, wake)
	}
	looping := base.Loop(true)
	if value, wake := sampleTimeline(ctx, router, looping, start.Add(1100*time.Millisecond)); value != 0 || !wake {
		t.Fatalf("rebound loop = %v, wake %v", value, wake)
	}
	if value, wake := sampleTimeline(ctx, router, looping, start.Add(1600*time.Millisecond)); value != 5 || !wake {
		t.Fatalf("rebound loop midpoint = %v, wake %v", value, wake)
	}
}

func TestTimelineRebindsDurationAndEasing(t *testing.T) {
	ctx := tweenTestContext(nil)
	router := new(input.Router)
	start := time.Unix(34, 0)
	base := Timeline("config").Keyframe(0, 0).Keyframe(1, 1).Duration(time.Second).Easing(EaseLinear)
	sampleTimeline(ctx, router, base, start)
	if value, _ := sampleTimeline(ctx, router, base, start.Add(250*time.Millisecond)); value != .25 {
		t.Fatalf("base progress = %v", value)
	}
	longer := base.Duration(2 * time.Second)
	if value, _ := sampleTimeline(ctx, router, longer, start.Add(250*time.Millisecond)); value != .25 {
		t.Fatalf("duration rebound progress = %v", value)
	}
	quadratic := longer.Easing(func(progress float32) float32 { return progress * progress })
	if value, _ := sampleTimeline(ctx, router, quadratic, start.Add(250*time.Millisecond)); value != .0625 {
		t.Fatalf("easing rebound progress = %v", value)
	}
}

func TestTimelinePausesAtCurrentFrameAndResumes(t *testing.T) {
	ctx := tweenTestContext(nil)
	router := new(input.Router)
	start := time.Unix(35, 0)
	base := Timeline("pause").Keyframe(0, 0).Keyframe(1, 10).Duration(time.Second).Easing(EaseLinear)
	sampleTimeline(ctx, router, base, start)

	paused := base.Playing(false)
	if value, wake := sampleTimeline(ctx, router, paused, start.Add(500*time.Millisecond)); value != 5 || wake {
		t.Fatalf("paused value = %v, wake %v; want 5, false", value, wake)
	}

	playing := base.Playing(true)
	if value, wake := sampleTimeline(ctx, router, playing, start.Add(1500*time.Millisecond)); value != 5 || !wake {
		t.Fatalf("resumed value = %v, wake %v; want 5, true", value, wake)
	}
}

func sampleTimeline(ctx *frame.Context, router *input.Router, timeline TimelineValue, now time.Time) (float32, bool) {
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Exact(image.Pt(100, 100)),
		Source:      router.Source(),
		Ops:         &ops,
		Now:         now,
	}
	frame.BeginFrame(ctx)
	value, running := timeline.Sample(ctx, gtx)
	frame.EndFrame(ctx)
	router.Frame(&ops)
	_, wake := router.WakeupTime()
	if running {
		wake = true
	}
	return value, wake
}
