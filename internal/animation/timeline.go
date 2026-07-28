package animation

import (
	"math"
	"sort"
	"time"

	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/flowui/internal/frame"
	"github.com/qianniancn/flowui/internal/state"
	"github.com/qianniancn/flowui/internal/theme"
)

const stateSlotTimeline = "timeline"

// TimelineValue is a keyed multi-keyframe float animation (track B extension).
// Progress advances over Duration; values are linearly interpolated between
// sorted keyframes. Optional Easing remaps overall progress before sampling.
type TimelineValue struct {
	key         string
	frames      []timelineFrame
	duration    time.Duration
	hasDuration bool
	easing      Easing
	loop        bool
	revision    uint64
	hasRevision bool
	disabled    bool
	playing     bool
	hasPlaying  bool
}

type timelineFrame struct {
	at    float32 // 0..1
	value float32
}

type timelineState struct {
	value       float32
	at          time.Time
	pausedAt    time.Time
	paused      bool
	duration    time.Duration
	easing      Easing
	frames      []timelineFrame
	loop        bool
	revision    uint64
	hasRevision bool
	ready       bool
	finished    bool
}

// Timeline starts a keyed keyframe animation. Call Keyframe at least twice
// (typically at 0 and 1) before Value/Sample.
func Timeline(key string) TimelineValue {
	return TimelineValue{key: key, easing: EaseLinear, playing: true, hasPlaying: true}
}

// Keyframe adds or replaces a keyframe at normalized progress in [0, 1].
func (t TimelineValue) Keyframe(at, value float32) TimelineValue {
	if at < 0 || at > 1 || !isFinite(at) {
		panic("flowui: timeline keyframe progress must be finite in [0, 1]")
	}
	validateFinite("keyframe value", value)
	frames := append([]timelineFrame(nil), t.frames...)
	replaced := false
	for i := range frames {
		if frames[i].at == at {
			frames[i].value = value
			replaced = true
			break
		}
	}
	if !replaced {
		frames = append(frames, timelineFrame{at: at, value: value})
	}
	sort.Slice(frames, func(i, j int) bool { return frames[i].at < frames[j].at })
	t.frames = frames
	return t
}

// Duration sets the full timeline length before Theme.Motion scaling.
func (t TimelineValue) Duration(duration time.Duration) TimelineValue {
	if duration < 0 {
		panic("flowui: timeline duration must not be negative")
	}
	t.duration = duration
	t.hasDuration = true
	return t
}

// Easing remaps overall timeline progress. Nil panics.
func (t TimelineValue) Easing(easing Easing) TimelineValue {
	if easing == nil {
		panic("flowui: timeline easing must not be nil")
	}
	t.easing = easing
	return t
}

// Loop restarts from the beginning when the timeline completes.
func (t TimelineValue) Loop(loop bool) TimelineValue {
	t.loop = loop
	return t
}

// Playing freezes progress when false (holds current value).
func (t TimelineValue) Playing(playing bool) TimelineValue {
	t.playing = playing
	t.hasPlaying = true
	return t
}

// Revision restarts the timeline when revision changes.
func (t TimelineValue) Revision(revision uint64) TimelineValue {
	t.revision = revision
	t.hasRevision = true
	return t
}

// Disabled snaps to the last keyframe when true.
func (t TimelineValue) Disabled(disabled bool) TimelineValue {
	t.disabled = disabled
	return t
}

// Value advances the timeline and returns the current float.
func (t TimelineValue) Value(ctx *frame.Context, gtx layout.Context) float32 {
	value, _ := t.Sample(ctx, gtx)
	return value
}

// Sample advances the timeline and reports whether another timed frame is needed.
func (t TimelineValue) Sample(ctx *frame.Context, gtx layout.Context) (float32, bool) {
	if len(t.frames) == 0 {
		panic("flowui: timeline requires at least one keyframe")
	}
	key := frame.ClaimKey(ctx, state.KindTimeline, t.key)
	current := frame.UseState[timelineState](ctx, key, stateSlotTimeline)
	duration, enabled := t.resolveMotion(ctx)
	endValue := t.frames[len(t.frames)-1].value

	if !current.ready {
		current.ready = true
		current.frames = append([]timelineFrame(nil), t.frames...)
		current.duration = duration
		current.easing = t.easing
		current.loop = t.loop
		current.at = gtx.Now
		current.revision = t.revision
		current.hasRevision = t.hasRevision
		current.value = t.frames[0].value
		if !enabled || t.disabled || duration <= 0 {
			current.value = endValue
			current.finished = true
			return current.value, false
		}
		if len(t.frames) == 1 {
			current.finished = true
			return current.value, false
		}
		gtx.Execute(op.InvalidateCmd{})
		return current.value, true
	}

	restarted := t.hasRevision && (!current.hasRevision || t.revision != current.revision)
	if restarted {
		current.revision = t.revision
		current.hasRevision = true
		current.frames = append([]timelineFrame(nil), t.frames...)
		current.duration = duration
		current.easing = t.easing
		current.loop = t.loop
		current.at = gtx.Now
		current.paused = false
		current.finished = false
		current.value = t.frames[0].value
	} else if !framesEqual(current.frames, t.frames) {
		// Declarative View rebuilds may change keyframes without Revision;
		// rebind values while preserving overall progress when possible.
		elapsed := gtx.Now.Sub(current.at)
		if current.loop && current.duration > 0 {
			elapsed %= current.duration
		}
		progress := Progress(elapsed, current.duration)
		current.frames = append([]timelineFrame(nil), t.frames...)
		current.duration = duration
		current.easing = t.easing
		current.loop = t.loop
		if progress > 0 && progress < 1 && duration > 0 {
			current.at = gtx.Now.Add(-time.Duration(float64(duration) * float64(progress)))
		}
		current.value = sampleKeyframes(current.frames, applyEasing(current.easing, progress))
	} else {
		oldLoop := current.loop
		if current.duration != duration {
			elapsed := gtx.Now.Sub(current.at)
			if oldLoop && current.duration > 0 {
				elapsed %= current.duration
			}
			progress := Progress(elapsed, current.duration)
			current.duration = duration
			if progress > 0 && progress < 1 && duration > 0 {
				current.at = gtx.Now.Add(-time.Duration(float64(duration) * float64(progress)))
			}
		}
		if oldLoop && !t.loop && current.duration > 0 {
			elapsed := gtx.Now.Sub(current.at) % current.duration
			current.at = gtx.Now.Add(-elapsed)
		}
		if !oldLoop && t.loop && current.finished {
			current.at = gtx.Now
			current.value = current.frames[0].value
			current.finished = false
		}
		current.easing = t.easing
		current.loop = t.loop
	}

	if !enabled || t.disabled {
		current.value = endValue
		current.finished = true
		current.paused = false
		return current.value, false
	}
	if t.hasPlaying && !t.playing {
		// Freeze the clock so resume continues from the held progress.
		if !current.paused {
			elapsed := gtx.Now.Sub(current.at)
			if current.loop && current.duration > 0 {
				elapsed %= current.duration
			}
			progress := Progress(elapsed, current.duration)
			if progress >= 1 && !current.loop {
				current.value = endValue
				current.finished = true
			} else {
				current.value = sampleKeyframes(current.frames, applyEasing(current.easing, progress))
			}
			current.pausedAt = gtx.Now
			current.paused = true
		}
		return current.value, false
	}
	if current.paused {
		current.at = current.at.Add(gtx.Now.Sub(current.pausedAt))
		current.paused = false
	}
	if current.finished && !t.loop {
		return current.value, false
	}

	elapsed := gtx.Now.Sub(current.at)
	if current.loop && duration > 0 {
		elapsed = elapsed % duration
	}
	progress := Progress(elapsed, duration)
	if progress >= 1 && !current.loop {
		current.value = endValue
		current.finished = true
		return current.value, false
	}
	eased := applyEasing(current.easing, progress)
	current.value = sampleKeyframes(current.frames, eased)
	if !current.finished || current.loop {
		gtx.Execute(op.InvalidateCmd{})
		return current.value, true
	}
	return current.value, false
}

func framesEqual(a, b []timelineFrame) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func (t TimelineValue) resolveMotion(ctx *frame.Context) (time.Duration, bool) {
	motion := frame.ActiveTheme(ctx).Motion
	duration := motion.DefaultDuration
	if t.hasDuration {
		duration = t.duration
	}
	if t.disabled {
		return 0, false
	}
	duration = theme.ResolveMotionDuration(motion, duration)
	return duration, duration > 0
}

func sampleKeyframes(frames []timelineFrame, progress float32) float32 {
	if len(frames) == 0 {
		return 0
	}
	if progress <= frames[0].at {
		return frames[0].value
	}
	last := frames[len(frames)-1]
	if progress >= last.at {
		return last.value
	}
	for i := 1; i < len(frames); i++ {
		right := frames[i]
		left := frames[i-1]
		if progress > right.at {
			continue
		}
		span := right.at - left.at
		if span <= 0 {
			return right.value
		}
		local := (progress - left.at) / span
		return LerpFloat(left.value, right.value, local)
	}
	return last.value
}

func isFinite(value float32) bool {
	return !math.IsNaN(float64(value)) && !math.IsInf(float64(value), 0)
}
