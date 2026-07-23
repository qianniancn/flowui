package animation

import (
	"image/color"
	"time"

	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/FlowUI/internal/theme"
)

// FloatTransition stores a component-local float transition.
type FloatTransition struct {
	value float32
	from  float32
	to    float32
	at    time.Time
	ready bool
}

func (t *FloatTransition) Initialize(value float32, now time.Time) {
	if t.ready {
		return
	}
	t.value = value
	t.from = value
	t.to = value
	t.at = now
	t.ready = true
}

func (t *FloatTransition) Set(value, target float32, now time.Time) {
	t.value = value
	t.from = value
	t.to = target
	t.at = now
	t.ready = true
}

func (t *FloatTransition) Value(gtx layout.Context, target float32, duration time.Duration, easing Easing, motions ...theme.MotionTheme) float32 {
	if len(motions) > 0 {
		duration = theme.ResolveMotionDuration(motions[0], duration)
	}
	if !t.ready {
		t.value = target
		t.from = target
		t.to = target
		t.at = gtx.Now
		t.ready = true
		return target
	}
	if target != t.to {
		t.from = t.value
		t.to = target
		t.at = gtx.Now
	}
	if duration <= 0 {
		t.value = target
		t.from = target
		t.to = target
		t.at = gtx.Now
		return target
	}
	if t.from == t.to {
		t.value = t.to
		return t.value
	}
	progress := applyEasing(easing, Progress(gtx.Now.Sub(t.at), duration))
	if progress < 1 {
		gtx.Execute(op.InvalidateCmd{})
	}
	t.value = LerpFloat(t.from, t.to, progress)
	return t.value
}

func (t *FloatTransition) Current() float32 { return t.value }
func (t *FloatTransition) Target() float32  { return t.to }
func (t *FloatTransition) Ready() bool      { return t.ready }
func (t *FloatTransition) Increasing() bool { return t.to >= t.from }

func (t *FloatTransition) Reset() {
	*t = FloatTransition{}
}

// ColorTransition stores a component-local color transition.
type ColorTransition struct {
	value color.NRGBA
	from  color.NRGBA
	to    color.NRGBA
	at    time.Time
	ready bool
}

func (t *ColorTransition) Set(value, target color.NRGBA, now time.Time) {
	t.value = value
	t.from = value
	t.to = target
	t.at = now
	t.ready = true
}

func (t *ColorTransition) Value(gtx layout.Context, target color.NRGBA, duration time.Duration, easing Easing, motions ...theme.MotionTheme) color.NRGBA {
	if len(motions) > 0 {
		duration = theme.ResolveMotionDuration(motions[0], duration)
	}
	if !t.ready {
		t.value = target
		t.from = target
		t.to = target
		t.at = gtx.Now
		t.ready = true
		return target
	}
	if target != t.to {
		t.from = t.value
		t.to = target
		t.at = gtx.Now
	}
	if duration <= 0 {
		t.value = target
		t.from = target
		t.to = target
		t.at = gtx.Now
		return target
	}
	if t.from == t.to {
		t.value = t.to
		return t.value
	}
	progress := applyEasing(easing, Progress(gtx.Now.Sub(t.at), duration))
	if progress < 1 {
		gtx.Execute(op.InvalidateCmd{})
	}
	t.value = LerpColor(t.from, t.to, progress)
	return t.value
}

func (t *ColorTransition) Current() color.NRGBA {
	return t.value
}

func (t *ColorTransition) Reset() {
	*t = ColorTransition{}
}
