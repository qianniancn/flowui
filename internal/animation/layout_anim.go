package animation

import (
	"image"
	"math"
	"time"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"github.com/qianniancn/flowui/internal/frame"
	"github.com/qianniancn/flowui/internal/state"
	"github.com/qianniancn/flowui/internal/theme"
)

const stateSlotLayoutRect = "layout-rect"
const stateSlotLayoutSize = "layout-size"

// RectValue animates an image.Rectangle toward a target (FLIP-style layout helper).
// Edges are interpolated independently with either duration+easing or a spring.
type RectValue struct {
	key         string
	target      image.Rectangle
	duration    time.Duration
	hasDuration bool
	easing      Easing
	spring      SpringConfig
	useSpring   bool
	disabled    bool
}

type rectEdgeState struct {
	minX, minY, maxX, maxY springState
	from, to               image.Rectangle
	at                     time.Time
	lastNow                time.Time
	duration               time.Duration
	easing                 Easing
	useSpring              bool
	ready                  bool
}

// AnimateRect starts a keyed rectangle animation toward target.
func AnimateRect(key string, target image.Rectangle) RectValue {
	return RectValue{key: key, target: target, easing: EaseCubicOut}
}

// Duration sets duration-mode length before Theme.Motion scaling.
func (r RectValue) Duration(duration time.Duration) RectValue {
	if duration < 0 {
		panic("flowui: layout animation duration must not be negative")
	}
	r.duration = duration
	r.hasDuration = true
	return r
}

// Easing sets duration-mode easing. Nil panics.
func (r RectValue) Easing(easing Easing) RectValue {
	if easing == nil {
		panic("flowui: layout animation easing must not be nil")
	}
	r.easing = easing
	return r
}

// Spring switches edge interpolation to spring physics.
func (r RectValue) Spring(config SpringConfig) RectValue {
	r.spring = config.normalized()
	r.useSpring = true
	return r
}

// Disabled snaps to the target rectangle when true.
func (r RectValue) Disabled(disabled bool) RectValue {
	r.disabled = disabled
	return r
}

// Value advances the animation and returns the current rectangle.
func (r RectValue) Value(ctx *frame.Context, gtx layout.Context) image.Rectangle {
	rect, _ := r.Sample(ctx, gtx)
	return rect
}

// Sample advances the animation and reports whether another frame is needed.
func (r RectValue) Sample(ctx *frame.Context, gtx layout.Context) (image.Rectangle, bool) {
	key := frame.ClaimKey(ctx, state.KindLayoutAnim, r.key)
	current := frame.UseState[rectEdgeState](ctx, key, stateSlotLayoutRect)
	enabled := r.motionEnabled(ctx)

	if !current.ready {
		current.ready = true
		current.to = r.target
		current.from = r.target
		current.at = gtx.Now
		current.lastNow = gtx.Now
		current.duration = r.resolveDuration(ctx)
		current.easing = r.easing
		current.useSpring = r.useSpring
		current.snapEdges(r.target)
		return r.target, false
	}

	if !enabled || r.disabled {
		current.snapEdges(r.target)
		current.to = r.target
		current.from = r.target
		return r.target, false
	}

	if r.target != current.to {
		current.from = current.currentRect()
		current.to = r.target
		current.at = gtx.Now
		current.lastNow = gtx.Now
		current.duration = r.resolveDuration(ctx)
		current.easing = r.easing
		current.useSpring = r.useSpring
		if r.useSpring {
			current.minX.target = float32(r.target.Min.X)
			current.minY.target = float32(r.target.Min.Y)
			current.maxX.target = float32(r.target.Max.X)
			current.maxY.target = float32(r.target.Max.Y)
		}
	} else if r.useSpring != current.useSpring {
		current.useSpring = r.useSpring
		current.from = current.currentRect()
		current.to = r.target
		current.at = gtx.Now
		current.lastNow = gtx.Now
		current.duration = r.resolveDuration(ctx)
		current.easing = r.easing
		if r.useSpring {
			current.minX.snap(float32(current.from.Min.X))
			current.minY.snap(float32(current.from.Min.Y))
			current.maxX.snap(float32(current.from.Max.X))
			current.maxY.snap(float32(current.from.Max.Y))
			current.minX.target = float32(r.target.Min.X)
			current.minY.target = float32(r.target.Min.Y)
			current.maxX.target = float32(r.target.Max.X)
			current.maxY.target = float32(r.target.Max.Y)
		}
	}

	var running bool
	if r.useSpring {
		current.useSpring = true
		dt := gtx.Now.Sub(current.lastNow)
		current.lastNow = gtx.Now
		cfg := r.spring
		runMinX := current.minX.step(dt, cfg)
		runMinY := current.minY.step(dt, cfg)
		runMaxX := current.maxX.step(dt, cfg)
		runMaxY := current.maxY.step(dt, cfg)
		running = runMinX || runMinY || runMaxX || runMaxY
	} else {
		current.useSpring = false
		if current.from.Eq(current.to) {
			return current.to, false
		}
		progress := Progress(gtx.Now.Sub(current.at), current.duration)
		if progress >= 1 {
			current.snapEdges(current.to)
			current.from = current.to
			return current.to, false
		}
		eased := applyEasing(current.easing, progress)
		current.minX.value = LerpFloat(float32(current.from.Min.X), float32(current.to.Min.X), eased)
		current.minY.value = LerpFloat(float32(current.from.Min.Y), float32(current.to.Min.Y), eased)
		current.maxX.value = LerpFloat(float32(current.from.Max.X), float32(current.to.Max.X), eased)
		current.maxY.value = LerpFloat(float32(current.from.Max.Y), float32(current.to.Max.Y), eased)
		running = true
	}

	if running {
		gtx.Execute(op.InvalidateCmd{})
	}
	return current.currentRect(), running
}

func (r RectValue) motionEnabled(ctx *frame.Context) bool {
	if r.disabled {
		return false
	}
	motion := frame.ActiveTheme(ctx).Motion
	if !theme.MotionEnabled(motion) {
		return false
	}
	if r.useSpring {
		return true
	}
	return r.resolveDuration(ctx) > 0
}

func (r RectValue) resolveDuration(ctx *frame.Context) time.Duration {
	motion := frame.ActiveTheme(ctx).Motion
	duration := motion.DefaultDuration
	if r.hasDuration {
		duration = r.duration
	}
	if r.disabled {
		return 0
	}
	return theme.ResolveMotionDuration(motion, duration)
}

func (s *rectEdgeState) snapEdges(rect image.Rectangle) {
	s.minX.snap(float32(rect.Min.X))
	s.minY.snap(float32(rect.Min.Y))
	s.maxX.snap(float32(rect.Max.X))
	s.maxY.snap(float32(rect.Max.Y))
}

func (s *rectEdgeState) currentRect() image.Rectangle {
	minX := int(math.Round(float64(s.minX.value)))
	minY := int(math.Round(float64(s.minY.value)))
	maxX := int(math.Round(float64(s.maxX.value)))
	maxY := int(math.Round(float64(s.maxY.value)))
	if maxX < minX {
		maxX = minX
	}
	if maxY < minY {
		maxY = minY
	}
	return image.Rect(minX, minY, maxX, maxY)
}

// AnimateLayoutWidget interpolates a child's reported size when it changes
// between frames (layout auto-animation for size). Requires a stable key.
type AnimateLayoutWidget struct {
	key         string
	child       frame.Widget
	duration    time.Duration
	hasDuration bool
	easing      Easing
	spring      SpringConfig
	useSpring   bool
	disabled    bool
}

type layoutSizeState struct {
	width, height springState
	fromW, fromH  float32
	toW, toH      float32
	at            time.Time
	lastNow       time.Time
	duration      time.Duration
	easing        Easing
	useSpring     bool
	ready         bool
}

// AnimateLayout wraps child and animates reported size changes under key.
func AnimateLayout(key string, child frame.Widget) AnimateLayoutWidget {
	return AnimateLayoutWidget{key: key, child: child, easing: EaseCubicOut}
}

// Duration sets duration-mode length.
func (w AnimateLayoutWidget) Duration(duration time.Duration) AnimateLayoutWidget {
	if duration < 0 {
		panic("flowui: layout animation duration must not be negative")
	}
	w.duration = duration
	w.hasDuration = true
	return w
}

// Easing sets duration-mode easing.
func (w AnimateLayoutWidget) Easing(easing Easing) AnimateLayoutWidget {
	if easing == nil {
		panic("flowui: layout animation easing must not be nil")
	}
	w.easing = easing
	return w
}

// Spring uses spring physics for size changes.
func (w AnimateLayoutWidget) Spring(config SpringConfig) AnimateLayoutWidget {
	w.spring = config.normalized()
	w.useSpring = true
	return w
}

// Disabled snaps sizes immediately when true.
func (w AnimateLayoutWidget) Disabled(disabled bool) AnimateLayoutWidget {
	w.disabled = disabled
	return w
}

// Layout implements frame.Widget.
func (w AnimateLayoutWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	if w.child == nil {
		return layout.Dimensions{}
	}
	// Record child ops first so we can clip them to the animated size.
	macro := op.Record(gtx.Ops)
	dims, placement := frame.TrackOverlayPlacement(ctx, func() layout.Dimensions {
		return w.child.Layout(ctx, gtx)
	})
	placement.PlaceOffset(image.Point{})
	call := macro.Stop()

	targetW := float32(dims.Size.X)
	targetH := float32(dims.Size.Y)

	key := frame.ClaimDerivedKey(ctx, state.KindLayoutAnim, w.key, "size")
	current := frame.UseState[layoutSizeState](ctx, key, stateSlotLayoutSize)
	enabled := w.motionEnabled(ctx)

	if !current.ready {
		current.ready = true
		current.toW, current.toH = targetW, targetH
		current.fromW, current.fromH = targetW, targetH
		current.width.snap(targetW)
		current.height.snap(targetH)
		current.at = gtx.Now
		current.lastNow = gtx.Now
		current.duration = w.resolveDuration(ctx)
		current.easing = w.easing
		current.useSpring = w.useSpring
		call.Add(gtx.Ops)
		return dims
	}

	if !enabled || w.disabled {
		current.width.snap(targetW)
		current.height.snap(targetH)
		current.toW, current.toH = targetW, targetH
		current.fromW, current.fromH = targetW, targetH
		call.Add(gtx.Ops)
		return dims
	}

	if targetW != current.toW || targetH != current.toH {
		current.fromW = current.width.value
		current.fromH = current.height.value
		current.toW, current.toH = targetW, targetH
		current.at = gtx.Now
		current.lastNow = gtx.Now
		current.duration = w.resolveDuration(ctx)
		current.easing = w.easing
		current.useSpring = w.useSpring
		if w.useSpring {
			current.width.target = targetW
			current.height.target = targetH
		}
	} else if w.useSpring != current.useSpring {
		current.useSpring = w.useSpring
		current.fromW = current.width.value
		current.fromH = current.height.value
		current.toW, current.toH = targetW, targetH
		current.at = gtx.Now
		current.lastNow = gtx.Now
		current.duration = w.resolveDuration(ctx)
		current.easing = w.easing
		if w.useSpring {
			current.width.snap(current.fromW)
			current.height.snap(current.fromH)
			current.width.target = targetW
			current.height.target = targetH
		}
	}

	var running bool
	if w.useSpring {
		current.useSpring = true
		dt := gtx.Now.Sub(current.lastNow)
		current.lastNow = gtx.Now
		current.width.target = targetW
		current.height.target = targetH
		runW := current.width.step(dt, w.spring)
		runH := current.height.step(dt, w.spring)
		running = runW || runH
	} else {
		current.useSpring = false
		if current.fromW != current.toW || current.fromH != current.toH {
			progress := Progress(gtx.Now.Sub(current.at), current.duration)
			if progress >= 1 {
				current.width.snap(current.toW)
				current.height.snap(current.toH)
				current.fromW, current.fromH = current.toW, current.toH
			} else {
				eased := applyEasing(current.easing, progress)
				current.width.value = LerpFloat(current.fromW, current.toW, eased)
				current.height.value = LerpFloat(current.fromH, current.toH, eased)
				running = true
			}
		}
	}

	if running {
		gtx.Execute(op.InvalidateCmd{})
	}

	animW := int(current.width.value + 0.5)
	animH := int(current.height.value + 0.5)
	if animW < 0 {
		animW = 0
	}
	if animH < 0 {
		animH = 0
	}
	// Clip visual to the animated size so growing/shrinking does not spill.
	placement.ClipTo(image.Rect(0, 0, animW, animH))
	defer clip.Rect{Max: image.Pt(animW, animH)}.Push(gtx.Ops).Pop()
	call.Add(gtx.Ops)
	return layout.Dimensions{
		Size:     image.Pt(animW, animH),
		Baseline: dims.Baseline,
	}
}

func (w AnimateLayoutWidget) motionEnabled(ctx *frame.Context) bool {
	if w.disabled {
		return false
	}
	motion := frame.ActiveTheme(ctx).Motion
	if !theme.MotionEnabled(motion) {
		return false
	}
	if w.useSpring {
		return true
	}
	return w.resolveDuration(ctx) > 0
}

func (w AnimateLayoutWidget) resolveDuration(ctx *frame.Context) time.Duration {
	motion := frame.ActiveTheme(ctx).Motion
	duration := motion.DefaultDuration
	if w.hasDuration {
		duration = w.duration
	}
	if w.disabled {
		return 0
	}
	return theme.ResolveMotionDuration(motion, duration)
}
