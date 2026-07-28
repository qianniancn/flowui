package chart

import (
	"image"
	"math"

	"gioui.org/io/pointer"
)

const minimumDataWindowSpan = 0.02

// DataWindow is a normalized visible range where 0 is the data start and 1 is
// the data end. Float64 is intentional: time-based charts map this range back
// to wall-clock values and should not accumulate float32 rounding drift.
type DataWindow struct {
	Start float64
	End   float64
}

// VisibleCategoryRange converts a normalized data window to a non-empty
// category index range.
func VisibleCategoryRange(count int, window DataWindow) (int, int) {
	if count <= 0 {
		return 0, 0
	}
	start := int(math.Floor(float64(window.Start) * float64(count)))
	end := int(math.Ceil(float64(window.End) * float64(count)))
	start = min(max(start, 0), count-1)
	end = min(max(end, start+1), count)
	return start, end
}

// FullDataWindow returns the complete normalized data range.
func FullDataWindow() DataWindow {
	return DataWindow{End: 1}
}

// NewDataWindow validates and returns a normalized data range.
func NewDataWindow[T ~float32 | ~float64](start, end T) DataWindow {
	startValue, endValue := float64(start), float64(end)
	if !finite64(startValue) || !finite64(endValue) || startValue < 0 || endValue > 1 || endValue <= startValue {
		panic("flowui: chart data window must satisfy 0 <= start < end <= 1")
	}
	return DataWindow{Start: startValue, End: endValue}
}

func (w DataWindow) IsFull() bool {
	return w.Start == 0 && w.End == 1
}

func (w DataWindow) zoom(anchor, factor float64) DataWindow {
	anchor = min(max(anchor, 0), 1)
	span := w.End - w.Start
	nextSpan := min(max(span*factor, minimumDataWindowSpan), 1)
	anchorValue := w.Start + anchor*span
	start := anchorValue - anchor*nextSpan
	start = min(max(start, 0), 1-nextSpan)
	return DataWindow{Start: start, End: start + nextSpan}
}

func (w DataWindow) pan(delta float64) DataWindow {
	span := w.End - w.Start
	start := w.Start + delta
	start = min(max(start, 0), 1-span)
	return DataWindow{Start: start, End: start + span}
}

func finite64(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

// DataWindowGesture tracks pointer-driven zooming and panning.
type DataWindowGesture struct {
	active        bool
	dragging      bool
	pointerID     pointer.ID
	startPosition float32
	startWindow   DataWindow
}

// Dragging reports whether the gesture is actively panning the data window.
func (g *DataWindowGesture) Dragging() bool {
	return g.dragging
}

// Update applies one pointer event to the controlled data window.
func (g *DataWindowGesture) Update(event pointer.Event, plot image.Rectangle, current DataWindow, vertical bool) (DataWindow, bool) {
	if plot.Empty() {
		g.Cancel()
		return current, false
	}
	switch event.Kind {
	case pointer.Scroll:
		if event.Scroll.Y != 0 {
			factor := 0.8
			if event.Scroll.Y > 0 {
				factor = 1.25
			}
			position := event.Position.X
			minimum := float64(plot.Min.X)
			length := float64(plot.Dx())
			if vertical {
				position = event.Position.Y
				minimum = float64(plot.Min.Y)
				length = float64(plot.Dy())
			}
			anchor := (float64(position) - minimum) / length
			next := current.zoom(anchor, factor)
			return next, next != current
		}
		if event.Scroll.X != 0 {
			direction := 0.1
			if event.Scroll.X < 0 {
				direction = -direction
			}
			next := current.pan(direction * (current.End - current.Start))
			return next, next != current
		}
	case pointer.Press:
		if event.Buttons.Contain(pointer.ButtonPrimary) {
			g.active = true
			g.dragging = false
			g.pointerID = event.PointerID
			g.startPosition = event.Position.X
			if vertical {
				g.startPosition = event.Position.Y
			}
			g.startWindow = current
		}
	case pointer.Drag:
		if g.active && event.PointerID == g.pointerID {
			g.dragging = true
			position := event.Position.X
			length := float32(plot.Dx())
			if vertical {
				position = event.Position.Y
				length = float32(plot.Dy())
			}
			delta := (float64(g.startPosition) - float64(position)) / float64(length) * (g.startWindow.End - g.startWindow.Start)
			next := g.startWindow.pan(delta)
			return next, next != current
		}
	case pointer.Release, pointer.Cancel:
		if event.PointerID == g.pointerID {
			g.Cancel()
		}
	}
	return current, false
}

func (g *DataWindowGesture) Cancel() {
	g.active = false
	g.dragging = false
	g.pointerID = 0
}
