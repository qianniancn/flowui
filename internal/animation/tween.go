package animation

import (
	"math"
	"time"

	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/state"
	"github.com/qianniancn/FlowUI/internal/theme"
)

const stateSlotTween = "tween"

// TweenValue describes a frame-persistent float transition.
// By default it uses duration + easing. Call Spring to switch to a damped spring.
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
	spring      SpringConfig
	useSpring   bool
}

type tweenState struct {
	value       float32
	from        float32
	to          float32
	at          time.Time
	lastNow     time.Time
	duration    time.Duration
	easing      Easing
	revision    uint64
	hasRevision bool
	ready       bool
	useSpring   bool
	spring      springState
}

// Tween creates a keyed frame-persistent float transition.
func Tween(key string, target float32) TweenValue {
	validateFinite("target", target)
	return TweenValue{key: key, target: target, easing: EaseCubicOut}
}

// Duration sets the transition length for duration-mode tweens.
func (t TweenValue) Duration(duration time.Duration) TweenValue {
	if duration < 0 {
		panic("flowui: tween duration must not be negative")
	}
	t.duration = duration
	t.hasDuration = true
	return t
}

// Easing sets the easing function for duration-mode tweens.
func (t TweenValue) Easing(easing Easing) TweenValue {
	if easing == nil {
		panic("flowui: tween easing must not be nil")
	}
	t.easing = easing
	return t
}

// Spring switches this tween to spring physics toward the target.
// Duration and Easing are ignored while spring mode is active.
func (t TweenValue) Spring(config SpringConfig) TweenValue {
	t.spring = config.normalized()
	t.useSpring = true
	return t
}

// Initial enables an entry transition from value on the first frame.
func (t TweenValue) Initial(value float32) TweenValue {
	validateFinite("initial value", value)
	t.initial = value
	t.hasInitial = true
	return t
}

// Disabled freezes the tween at its target when true.
func (t TweenValue) Disabled(disabled bool) TweenValue {
	t.disabled = disabled
	return t
}

// Revision restarts the transition when revision changes.
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

// Sample advances the transition and reports whether another timed frame is required.
func (t TweenValue) Sample(ctx *frame.Context, gtx layout.Context) (float32, bool) {
	key := frame.ClaimKey(ctx, state.KindTween, t.key)
	current := frame.UseState[tweenState](ctx, key, stateSlotTween)
	enabled := t.motionEnabled(ctx)

	if !current.ready {
		current.ready = true
		current.to = t.target
		current.easing = t.easing
		current.duration = t.resolveDuration(ctx)
		current.at = gtx.Now
		current.lastNow = gtx.Now
		current.revision = t.revision
		current.hasRevision = t.hasRevision
		current.useSpring = t.useSpring
		if enabled && t.hasInitial && t.initial != t.target {
			current.value = t.initial
			current.from = t.initial
			if t.useSpring {
				current.spring.value = t.initial
				current.spring.target = t.target
				current.spring.velocity = 0
				current.spring.ready = true
				current.spring.config = t.spring
			}
			gtx.Execute(op.InvalidateCmd{})
			return current.value, true
		}
		current.value = t.target
		current.from = t.target
		if t.useSpring {
			current.spring.snap(t.target)
			current.spring.config = t.spring
		}
		return current.value, false
	}

	if !enabled || t.disabled {
		current.revision = t.revision
		current.hasRevision = t.hasRevision
		current.snap(t.target, gtx.Now)
		return current.value, false
	}

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
		current.lastNow = gtx.Now
		current.duration = t.resolveDuration(ctx)
		current.easing = t.easing
		current.useSpring = t.useSpring
		if t.useSpring {
			current.spring.value = current.value
			current.spring.target = t.target
			current.spring.velocity = 0
			current.spring.ready = true
			current.spring.config = t.spring
		}
	} else if t.target != current.to {
		current.advanceTo(gtx.Now)
		current.from = current.value
		current.to = t.target
		current.at = gtx.Now
		current.lastNow = gtx.Now
		current.duration = t.resolveDuration(ctx)
		current.easing = t.easing
		current.useSpring = t.useSpring
		if t.useSpring {
			current.spring.value = current.value
			current.spring.target = t.target
			// keep velocity for continuous retargeting
			current.spring.ready = true
			current.spring.config = t.spring
		}
	} else if t.useSpring != current.useSpring {
		// Mode switch under the same target must not stick on the previous path.
		current.advanceTo(gtx.Now)
		current.useSpring = t.useSpring
		current.from = current.value
		current.to = t.target
		current.at = gtx.Now
		current.lastNow = gtx.Now
		current.duration = t.resolveDuration(ctx)
		current.easing = t.easing
		if t.useSpring {
			current.spring.value = current.value
			current.spring.target = t.target
			current.spring.velocity = 0
			current.spring.ready = true
			current.spring.config = t.spring
		}
	}

	var running bool
	if t.useSpring {
		current.useSpring = true
		if !current.spring.ready {
			current.spring.snap(current.value)
			current.spring.target = t.target
		}
		current.spring.target = t.target
		dt := gtx.Now.Sub(current.lastNow)
		current.lastNow = gtx.Now
		running = current.spring.step(dt, t.spring)
		current.value = current.spring.value
		current.to = t.target
		current.from = current.value
	} else {
		current.useSpring = false
		if current.from == current.to {
			current.value = current.to
			return current.value, false
		}
		running = current.advance(gtx.Now)
	}
	if running {
		gtx.Execute(op.InvalidateCmd{})
	}
	return current.value, running
}

func (t TweenValue) motionEnabled(ctx *frame.Context) bool {
	if t.disabled {
		return false
	}
	motion := frame.ActiveTheme(ctx).Motion
	if !theme.MotionEnabled(motion) {
		return false
	}
	if t.useSpring {
		return true
	}
	return t.resolveDuration(ctx) > 0
}

func (t TweenValue) resolveDuration(ctx *frame.Context) time.Duration {
	motion := frame.ActiveTheme(ctx).Motion
	duration := motion.DefaultDuration
	if t.hasDuration {
		duration = t.duration
	}
	if t.disabled {
		return 0
	}
	return theme.ResolveMotionDuration(motion, duration)
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

func (s *tweenState) advanceTo(now time.Time) {
	if s.useSpring {
		s.spring.step(now.Sub(s.lastNow), s.spring.config)
		s.value = s.spring.value
		s.lastNow = now
		return
	}
	s.advance(now)
}

func (s *tweenState) snap(target float32, now time.Time) {
	s.value = target
	s.from = target
	s.to = target
	s.at = now
	s.lastNow = now
	s.spring.snap(target)
}

func validateFinite(name string, value float32) {
	if math.IsNaN(float64(value)) || math.IsInf(float64(value), 0) {
		panic("flowui: tween " + name + " must be finite")
	}
}
