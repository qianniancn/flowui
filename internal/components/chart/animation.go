package chart

import (
	"time"

	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/animation"
	"github.com/qianniancn/flowui/internal/frame"
)

// DataAnimation owns the common revision and motion state for chart data.
type DataAnimation[T any] struct {
	ready     bool
	revision  uint64
	from      T
	target    T
	displayed T
	duration  time.Duration
	easing    animation.Easing
}

func (a *DataAnimation[T]) Update(
	ctx *frame.Context,
	gtx layout.Context,
	target T,
	enabled bool,
	initialDuration, updateDuration time.Duration,
	initialEasing, updateEasing animation.Easing,
	same func(T, T) bool,
	baseline func(T) T,
	transitionFrom func(T, T) T,
	interpolate func(T, T, float32) T,
) T {
	if !a.ready {
		a.ready = true
		a.revision = 1
		a.from = baseline(target)
		a.target = target
		a.duration = initialDuration
		a.easing = initialEasing
	} else if !same(a.target, target) {
		a.revision++
		a.from = transitionFrom(a.displayed, target)
		a.target = target
		a.duration = updateDuration
		a.easing = updateEasing
	} else {
		a.target = target
	}

	progress, running := animation.Tween("data", 1).
		Initial(0).
		Revision(a.revision).
		Duration(a.duration).
		Easing(a.easing).
		Disabled(!enabled).
		Sample(ctx, gtx)
	if !running && progress == 1 {
		a.from = *new(T)
		a.displayed = target
		return target
	}
	a.displayed = interpolate(a.from, target, progress)
	return a.displayed
}

func ValidateAnimationDuration(duration time.Duration, name string) {
	if duration < 0 {
		panic("flowui: " + name + " animation duration must not be negative")
	}
}

func ValidateAnimationEasing(easing animation.Easing, name string) {
	if easing == nil {
		panic("flowui: " + name + " animation easing must not be nil")
	}
}
