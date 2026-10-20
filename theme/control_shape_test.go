package theme

import (
	"testing"

	"gioui.org/unit"
)

// TestControlShapeVocabularyIsComplete pins the shape vocabulary: every value
// below ControlShapeCount has a distinct name, and an out-of-range shape
// reports itself as unknown rather than borrowing a real name.
//
// It exists so that adding a shape cannot be finished by editing the enum
// alone — the new member shows up here as a missing name.
func TestControlShapeVocabularyIsComplete(t *testing.T) {
	seen := make(map[string]ControlShape, ControlShapeCount)
	for value := range ControlShapeCount {
		shape := ControlShape(value)
		name := shape.String()
		if name == "" || name == "shape(?)" {
			t.Errorf("ControlShape(%d).String() = %q, want a real name", value, name)
			continue
		}
		if previous, duplicate := seen[name]; duplicate {
			t.Errorf("ControlShape(%d) and ControlShape(%d) share the name %q", previous, value, name)
			continue
		}
		seen[name] = shape
	}
	if got := ControlShape(ControlShapeCount).String(); got != "shape(?)" {
		t.Errorf("out-of-range shape String() = %q, want %q", got, "shape(?)")
	}
}

// TestShapeSetAllHasEveryDerivedShape guards the constant against a new shape
// being added to the enum while ShapeSetAll quietly keeps claiming support for
// only the old ones.
func TestShapeSetAllHasEveryDerivedShape(t *testing.T) {
	for value := range ControlShapeCount {
		shape := ControlShape(value)
		if !ShapeSetAll.Has(shape) {
			t.Errorf("ShapeSetAll.Has(%s) = false, want true for every defined shape", shape)
		}
	}
}

// TestZeroShapeSetIsDefaultOnly is the contract that makes ShapeSet usable as a
// zero value: a component that says nothing has no shape concept, so everything
// but ShapeDefault is refused rather than silently honoured.
func TestZeroShapeSetIsDefaultOnly(t *testing.T) {
	var set ShapeSet
	if !set.Has(ShapeDefault) {
		t.Error("zero ShapeSet refuses ShapeDefault; the fallback must always be available")
	}
	for _, shape := range []ControlShape{ShapeCircle, ShapeRound, ShapeSquare} {
		if set.Has(shape) {
			t.Errorf("zero ShapeSet claims to render %s", shape)
		}
		if got := set.Resolve(shape); got != ShapeDefault {
			t.Errorf("zero ShapeSet.Resolve(%s) = %s, want default", shape, got)
		}
	}
}

// TestShapeSetResolveKeepsDeclaredShapes is the positive half: a declared shape
// survives resolution unchanged, so Resolve is a filter and not a transform.
func TestShapeSetResolveKeepsDeclaredShapes(t *testing.T) {
	set := ShapeSetAll | 1<<ShapeCircle
	for _, shape := range []ControlShape{ShapeDefault, ShapeCircle, ShapeRound, ShapeSquare} {
		if got := set.Resolve(shape); got != shape {
			t.Errorf("ShapeSetAll.Resolve(%s) = %s, want %s", shape, got, shape)
		}
	}
}

// TestShapeSetAssertsIndependence checks that a set built from individual bits
// only claims the shapes it was given.
func TestShapeSetAssertsIndependence(t *testing.T) {
	set := ShapeSet(1<<ShapeRound | 1<<ShapeSquare)
	if !set.Has(ShapeRound) || !set.Has(ShapeSquare) {
		t.Fatal("explicitly declared shapes are missing")
	}
	if set.Has(ShapeCircle) {
		t.Error("set claims ShapeCircle without declaring it")
	}
	if got := set.Resolve(ShapeCircle); got != ShapeDefault {
		t.Errorf("Resolve(ShapeCircle) = %s, want default for an undeclared shape", got)
	}
}

// TestControlRadiusFollowsTheHeightForDerivedShapes is the geometric core of
// the shape dimension: circled and pill shapes are defined relative to the
// resolved control height, and they ignore the component's radius token.
func TestControlRadiusFollowsTheHeightForDerivedShapes(t *testing.T) {
	const fallback = unit.Dp(6)
	cases := []struct {
		name   string
		shape  ControlShape
		height unit.Dp
		want   unit.Dp
	}{
		{"pill at medium", ShapeRound, 40, 20},
		{"pill at small", ShapeRound, 24, 12},
		// unit.Dp is a float, so an odd height keeps its half rather than
		// rounding: the radius is a geometric value, not a pixel count.
		{"pill at odd height", ShapeRound, 33, 16.5},
		{"circle at medium", ShapeCircle, 40, 20},
		{"square drops the radius", ShapeSquare, 40, 0},
		{"default keeps the token", ShapeDefault, 40, fallback},
		// An unknown shape must degrade to today's geometry, never to zero:
		// a zero radius on a shape the caller mistyped looks like a layout
		// bug rather than a rejected value.
		{"unknown keeps the token", ControlShape(200), 40, fallback},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			if got := ControlRadius(testCase.shape, testCase.height, fallback); got != testCase.want {
				t.Errorf("ControlRadius(%s, %v, %v) = %v, want %v",
					testCase.shape, testCase.height, fallback, got, testCase.want)
			}
		})
	}
}

// TestControlRadiusIgnoresTheFallbackForDerivedShapes states the precedence
// rule directly: a derived shape wins over the component token, so a component
// cannot half-implement a shape by also setting its radius.
func TestControlRadiusIgnoresTheFallbackForDerivedShapes(t *testing.T) {
	for _, shape := range []ControlShape{ShapeRound, ShapeCircle, ShapeSquare} {
		low := ControlRadius(shape, 40, 2)
		high := ControlRadius(shape, 40, 99)
		if low != high {
			t.Errorf("%s radius depends on the fallback token (%v vs %v); a shape must own the radius outright",
				shape, low, high)
		}
	}
}

// TestControlSquareSidePairsWithCircle is the invariant that keeps a circle
// from being drawn as an ellipse: the shape that forces a square box and the
// radius that completes it must agree on which shape that is.
func TestControlSquareSidePairsWithCircle(t *testing.T) {
	if side, ok := ControlSquareSide(ShapeCircle, 40); !ok || side != 40 {
		t.Errorf("ControlSquareSide(ShapeCircle, 40) = (%v, %v), want (40, true)", side, ok)
	}
	for _, shape := range []ControlShape{ShapeDefault, ShapeRound, ShapeSquare, ControlShape(200)} {
		side, ok := ControlSquareSide(shape, 40)
		if ok {
			t.Errorf("ControlSquareSide(%s, 40) forces a square side %v; only ShapeCircle may", shape, side)
		}
		if side != 0 {
			t.Errorf("ControlSquareSide(%s, 40) returned side %v with ok=false", shape, side)
		}
	}
}
