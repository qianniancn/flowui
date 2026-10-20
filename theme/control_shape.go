package theme

import "gioui.org/unit"

// ControlShape is the outline dimension of a control.
//
// The name stays clear of Shape, which is the resolved geometry scale — this
// enum is an *input* to that scale, not a second place to store its values.
type ControlShape uint8

const (
	// ShapeDefault keeps whatever radius the component's own theme declares
	// (Components.X.Radius). It is the zero value so an unset shape means "no
	// opinion" rather than "square".
	ShapeDefault ControlShape = iota
	// ShapeCircle is a circle: the box is forced square at the control height
	// and fully rounded. Only meaningful for icon-only controls, because a
	// square box cannot hold a text label.
	ShapeCircle
	// ShapeRound is a pill: the radius follows the control height, so the ends
	// are always half-circles regardless of size.
	ShapeRound
	// ShapeSquare drops the radius entirely.
	ShapeSquare
)

// ControlShapeCount is the number of defined shapes; ShapeSet indexes by it.
const ControlShapeCount = int(ShapeSquare) + 1

// String names the shape for diagnostics and test failures.
func (s ControlShape) String() string {
	switch s {
	case ShapeDefault:
		return "default"
	case ShapeCircle:
		return "circle"
	case ShapeRound:
		return "round"
	case ShapeSquare:
		return "square"
	}
	return "shape(?)"
}

// ShapeSet declares which shapes a component renders. The zero value means
// "only ShapeDefault", because a shape is an opt-in capability: most controls
// have no shape concept at all, while Button and Input implement all four.
//
// It is shaped exactly like VariantSet on purpose — both answer the same
// question ("which of these does this component actually render?") and both
// fall back to their neutral member rather than failing.
type ShapeSet uint8

// ShapeSetAll is the set a component declares when it renders every shape.
//
// It is written as the union of the derived shapes rather than as a sentinel,
// so that adding a shape to the vocabulary forces this constant to be
// revisited instead of silently claiming support for it.
const ShapeSetAll ShapeSet = 1<<ShapeCircle | 1<<ShapeRound | 1<<ShapeSquare

// Has reports whether the component declares support for a shape.
// ShapeDefault is always supported: it is the fallback every component has.
func (s ShapeSet) Has(shape ControlShape) bool {
	if shape == ShapeDefault {
		return true
	}
	if shape > ShapeSquare {
		return false
	}
	return s&(1<<shape) != 0
}

// Resolve returns a shape the component can render, falling back to
// ShapeDefault for one it cannot. It never reports failure: an unsupported
// shape is a no-op, not an error, because the caller is a builder method that
// has no error channel.
func (s ShapeSet) Resolve(shape ControlShape) ControlShape {
	if s.Has(shape) {
		return shape
	}
	return ShapeDefault
}

// ControlRadius is the single place a ControlShape becomes a radius.
//
// fallback is the component's own radius token (Components.X.Radius) and is
// returned unchanged for ShapeDefault and for out-of-range values, so an
// unknown shape degrades to today's geometry instead of a zero radius. The
// derived shapes ignore fallback entirely — that is what makes them shapes
// rather than another radius knob.
//
// Callers must pass the *resolved* height of the control (after size variant
// and any density token), because ShapeRound and ShapeCircle are defined
// relative to it.
func ControlRadius(shape ControlShape, height, fallback unit.Dp) unit.Dp {
	switch shape {
	case ShapeRound, ShapeCircle:
		// Gio clamps a radius to half the smaller dimension, so this is the
		// largest value that still reads as a half-circle rather than a bug.
		return height / 2
	case ShapeSquare:
		return 0
	default:
		return fallback
	}
}

// ControlSquareSide reports the side length for a shape that forces a square
// box, and whether the shape forces one at all.
//
// ShapeCircle is the only such shape, and the two values belong together: a
// circle drawn in a non-square box is an ellipse, so a component that honours
// ShapeCircle must apply both or neither. Keeping them in one function is what
// stops the pair from drifting apart at the call site.
func ControlSquareSide(shape ControlShape, height unit.Dp) (unit.Dp, bool) {
	if shape == ShapeCircle {
		return height, true
	}
	return 0, false
}
