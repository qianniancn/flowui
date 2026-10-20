package style

import "gioui.org/unit"

func (s Style) editTransform(edit func(*TransformStyle)) Style {
	transform := TransformStyle{}
	if s.trans != nil {
		transform = *s.trans
	}
	s.trans = &transform
	edit(s.trans)
	return s
}

func (s Style) Translate(x, y unit.Dp) Style {
	if !finite(float32(x)) || !finite(float32(y)) {
		return s
	}
	return s.editTransform(func(transform *TransformStyle) {
		transform.TranslateX = new(x)
		transform.TranslateY = new(y)
	})
}

// Scale sets an equal-ratio transform scale: one factor for both axes.
//
// The two-argument form this replaces had no x != y caller anywhere in the
// repo; every site passed the same value twice. If a genuinely two-axis case
// appears, add an explicit two-axis setter rather than widening this one.
func (s Style) Scale(factor float32) Style {
	if !finite(factor) {
		return s
	}
	return s.editTransform(func(transform *TransformStyle) {
		transform.ScaleX = new(factor)
		transform.ScaleY = new(factor)
	})
}

func (s Style) Rotate(radians float32) Style {
	if !finite(radians) {
		return s
	}
	return s.editTransform(func(transform *TransformStyle) { transform.Rotate = new(radians) })
}
