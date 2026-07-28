package animation

import (
	"testing"
	"time"

	"gioui.org/io/input"
	"github.com/qianniancn/FlowUI/internal/theme"
)

func TestTweenSpringApproachesTarget(t *testing.T) {
	ctx := tweenTestContext(nil)
	router := new(input.Router)
	start := time.Unix(20, 0)
	tween := Tween("spring", 100).Initial(0).Spring(SpringSnappy())

	value, wake := tweenFrame(ctx, router, tween, start)
	if value != 0 || !wake {
		t.Fatalf("spring start = %v, wake %v", value, wake)
	}

	// Step several frames; value should move toward 100 without jumping instantly.
	var last float32
	for i := 1; i <= 30; i++ {
		value, wake = tweenFrame(ctx, router, tween, start.Add(time.Duration(i)*16*time.Millisecond))
		if value < last {
			// allow tiny numerical noise
			if last-value > 1 {
				t.Fatalf("spring went backwards: %v -> %v", last, value)
			}
		}
		last = value
	}
	if last <= 10 {
		t.Fatalf("spring did not progress enough: %v", last)
	}

	// Eventually settles at target.
	for i := 31; i <= 200; i++ {
		value, wake = tweenFrame(ctx, router, tween, start.Add(time.Duration(i)*16*time.Millisecond))
		if !wake {
			if value != 100 {
				t.Fatalf("settled value = %v, want 100", value)
			}
			return
		}
	}
	t.Fatalf("spring did not settle; last=%v", value)
}

func TestTweenSpringHonorsMotionDisabled(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	activeTheme.Motion.Enabled = false
	ctx := tweenTestContext(&activeTheme)
	router := new(input.Router)
	value, wake := tweenFrame(ctx, router, Tween("spring-off", 1).Initial(0).Spring(DefaultSpring()), time.Unix(21, 0))
	if value != 1 || wake {
		t.Fatalf("spring with motion off = %v, wake %v", value, wake)
	}
}

func TestTweenModeSwitchFromSpringToDuration(t *testing.T) {
	ctx := tweenTestContext(nil)
	router := new(input.Router)
	start := time.Unix(22, 0)
	spring := Tween("mode", 100).Initial(0).Spring(SpringSnappy())
	// Drive spring until settled at target.
	var value float32
	var wake bool
	for i := range 200 {
		value, wake = tweenFrame(ctx, router, spring, start.Add(time.Duration(i)*16*time.Millisecond))
		if !wake {
			break
		}
	}
	if value != 100 || wake {
		t.Fatalf("spring settle = %v, wake %v", value, wake)
	}

	// Same target, switch to duration mode — must remain at target, not re-run spring.
	duration := Tween("mode", 100).Duration(time.Second).Easing(EaseLinear)
	value, wake = tweenFrame(ctx, router, duration, start.Add(5*time.Second))
	if value != 100 || wake {
		t.Fatalf("duration after mode switch = %v, wake %v", value, wake)
	}
}

func TestTweenSpringRetargetsFromCurrentTime(t *testing.T) {
	start := time.Unix(23, 0)
	config := SpringSnappy()

	referenceCtx := tweenTestContext(nil)
	referenceRouter := new(input.Router)
	forward := Tween("forward", 1).Initial(0).Spring(config)
	tweenFrame(referenceCtx, referenceRouter, forward, start)
	for step := 1; step <= 3; step++ {
		tweenFrame(referenceCtx, referenceRouter, forward, start.Add(time.Duration(step)*100*time.Millisecond))
	}
	want, _ := tweenFrame(referenceCtx, referenceRouter, forward, start.Add(400*time.Millisecond))

	retargetCtx := tweenTestContext(nil)
	retargetRouter := new(input.Router)
	retarget := Tween("retarget", 1).Initial(0).Spring(config)
	tweenFrame(retargetCtx, retargetRouter, retarget, start)
	for step := 1; step <= 3; step++ {
		tweenFrame(retargetCtx, retargetRouter, retarget, start.Add(time.Duration(step)*100*time.Millisecond))
	}
	got, _ := tweenFrame(retargetCtx, retargetRouter, Tween("retarget", 0).Spring(config), start.Add(400*time.Millisecond))
	if got != want {
		t.Fatalf("retarget value = %v, want current-time value %v", got, want)
	}
}
