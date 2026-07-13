package animation

import (
	"math"
	"time"

	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/state"
)

const stateSlotTween = "tween"

// TweenValue describes a frame-persistent float transition.
type TweenValue struct {
	key         string
	target      float32
	duration    time.Duration
	hasDuration bool
	easing      Easing
	initial     float32
	hasInitial  bool
	revision    uint64
	hasRevision bool
	disabled    bool
}

type tweenState struct {
	value       float32
	from        float32
	to          float32
	at          time.Time
	duration    time.Duration
	easing      Easing
	revision    uint64
	hasRevision bool
	ready       bool
}

func Tween(key string, target float32) TweenValue {
	validateFinite("target", target)
	return TweenValue{key: key, target: target, easing: EaseCubicOut}
}

func (t TweenValue) Duration(duration time.Duration) TweenValue {
	if duration < 0 {
		panic("flowui: tween duration must not be negative")
	}
	t.duration = duration
	t.hasDuration = true
	return t
}

func (t TweenValue) Easing(easing Easing) TweenValue {
	if easing == nil {
		panic("flowui: tween easing must not be nil")
	}
	t.easing = easing
	return t
}

// Initial enables an entry transition from value on the first frame.
func (t TweenValue) Initial(value float32) TweenValue {
	validateFinite("initial value", value)
	t.initial = value
	t.hasInitial = true
	return t
}

func (t TweenValue) Disabled(disabled bool) TweenValue {
	t.disabled = disabled
	return t
}

// Revision restarts the transition when revision changes. If Initial is set,
// the restarted transition begins from that value.
func (t TweenValue) Revision(revision uint64) TweenValue {
	t.revision = revision
	t.hasRevision = true
	return t
}

// Value advances the transition and requests another frame while it is active.
func (t TweenValue) Value(ctx *frame.Context, gtx layout.Context) float32 {
	value, _ := t.Sample(ctx, gtx)
	return value
}

// Sample advances the transition and reports whether another timed frame is
// required. It is useful when rendering the animated value is expensive.
func (t TweenValue) Sample(ctx *frame.Context, gtx layout.Context) (float32, bool) {
	key := frame.ClaimKey(ctx, state.KindTween, t.key)
	current := frame.UseState[tweenState](ctx, key, stateSlotTween)
	duration, enabled := t.resolveMotion(ctx)

	if !current.ready {
		current.ready = true
		current.to = t.target
		current.easing = t.easing
		current.duration = duration
		current.at = gtx.Now
		current.revision = t.revision
		current.hasRevision = t.hasRevision
		if enabled && t.hasInitial && t.initial != t.target {
			current.value = t.initial
			current.from = t.initial
			gtx.Execute(op.InvalidateCmd{})
			return current.value, true
		}
		current.value = t.target
		current.from = t.target
		return current.value, false
	}

	if !enabled {
		current.revision = t.revision
		current.hasRevision = t.hasRevision
		current.snap(t.target, gtx.Now)
		return current.value, false
	}

	current.advance(gtx.Now)
	restarted := t.hasRevision && (!current.hasRevision || t.revision != current.revision)
	if restarted {
		current.revision = t.revision
		current.hasRevision = true
		if t.hasInitial {
			current.value = t.initial
		}
		current.from = current.value
		current.to = t.target
		current.at = gtx.Now
		current.duration = duration
		current.easing = t.easing
	} else if t.target != current.to {
		current.from = current.value
		current.to = t.target
		current.at = gtx.Now
		current.duration = duration
		current.easing = t.easing
	}
	if current.from == current.to {
		current.value = current.to
		return current.value, false
	}
	running := current.advance(gtx.Now)
	if running {
		gtx.Execute(op.InvalidateCmd{})
	}
	return current.value, running
}

func (t TweenValue) resolveMotion(ctx *frame.Context) (time.Duration, bool) {
	motion := frame.ActiveTheme(ctx).Motion
	duration := motion.DefaultDuration
	if t.hasDuration {
		duration = t.duration
	}
	scale := motion.DurationScale
	if t.disabled || !motion.Enabled || duration <= 0 || scale <= 0 || math.IsNaN(float64(scale)) || math.IsInf(float64(scale), 0) {
		return 0, false
	}
	scaled := time.Duration(float64(duration) * float64(scale))
	if scaled <= 0 {
		return 0, false
	}
	return scaled, true
}

func (s *tweenState) advance(now time.Time) bool {
	if s.from == s.to {
		s.value = s.to
		return false
	}
	progress := Progress(now.Sub(s.at), s.duration)
	if progress >= 1 {
		s.value = s.to
		s.from = s.to
		return false
	}
	s.value = LerpFloat(s.from, s.to, applyEasing(s.easing, progress))
	return true
}

func (s *tweenState) snap(target float32, now time.Time) {
	s.value = target
	s.from = target
	s.to = target
	s.at = now
}

func validateFinite(name string, value float32) {
	if math.IsNaN(float64(value)) || math.IsInf(float64(value), 0) {
		panic("flowui: tween " + name + " must be finite")
	}
}
