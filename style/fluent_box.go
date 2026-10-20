package style

import (
	"gioui.org/io/pointer"
	"gioui.org/unit"
)

const (
	sideTop uint8 = 1 << iota
	sideRight
	sideBottom
	sideLeft
	sidesAll = sideTop | sideRight | sideBottom | sideLeft
)

func (s Style) editBox(edit func(*BoxStyle)) Style {
	box := BoxStyle{}
	if s.box != nil {
		box = *s.box
	}
	s.box = &box
	edit(s.box)
	return s
}

func (s Style) Width(value unit.Dp) Style {
	if !finite(float32(value)) {
		return s
	}
	return s.editBox(func(box *BoxStyle) { box.Width = new(value) })
}

func (s Style) Height(value unit.Dp) Style {
	if !finite(float32(value)) {
		return s
	}
	return s.editBox(func(box *BoxStyle) { box.Height = new(value) })
}

func (s Style) MinWidth(value unit.Dp) Style {
	if !finite(float32(value)) {
		return s
	}
	return s.editBox(func(box *BoxStyle) { box.MinWidth = new(value) })
}

func (s Style) MaxWidth(value unit.Dp) Style {
	if !finite(float32(value)) {
		return s
	}
	return s.editBox(func(box *BoxStyle) { box.MaxWidth = new(value) })
}

func (s Style) MinHeight(value unit.Dp) Style {
	if !finite(float32(value)) {
		return s
	}
	return s.editBox(func(box *BoxStyle) { box.MinHeight = new(value) })
}

func (s Style) MaxHeight(value unit.Dp) Style {
	if !finite(float32(value)) {
		return s
	}
	return s.editBox(func(box *BoxStyle) { box.MaxHeight = new(value) })
}

func (s Style) FillWidth() Style {
	return s.editBox(func(box *BoxStyle) { box.FillWidth = new(true) })
}

func (s Style) FillHeight() Style {
	return s.editBox(func(box *BoxStyle) { box.FillHeight = new(true) })
}

// AspectRatio sets the preferred width divided by height. It resolves the
// unconstrained axis when Width, Height, FillWidth, or FillHeight fixes one.
func (s Style) AspectRatio(value float32) Style {
	if !(value > 0) || !finite(value) {
		return s
	}
	return s.editBox(func(box *BoxStyle) { box.AspectRatio = new(value) })
}

func (s Style) Padding(value unit.Dp) Style {
	if !finite(float32(value)) {
		return s
	}
	return s.editBox(func(box *BoxStyle) {
		box.Padding = &Insets{Top: value, Right: value, Bottom: value, Left: value}
		box.paddingMask = sidesAll
	})
}

func (s Style) PaddingX(value unit.Dp) Style {
	if !finite(float32(value)) {
		return s
	}
	return s.editBox(func(box *BoxStyle) {
		padding := insets(box.Padding)
		padding.Left, padding.Right = value, value
		box.Padding = &padding
		box.paddingMask |= sideLeft | sideRight
	})
}

func (s Style) PaddingY(value unit.Dp) Style {
	if !finite(float32(value)) {
		return s
	}
	return s.editBox(func(box *BoxStyle) {
		padding := insets(box.Padding)
		padding.Top, padding.Bottom = value, value
		box.Padding = &padding
		box.paddingMask |= sideTop | sideBottom
	})
}

func (s Style) PaddingTop(value unit.Dp) Style {
	return s.paddingSide(sideTop, value)
}

func (s Style) PaddingRight(value unit.Dp) Style {
	return s.paddingSide(sideRight, value)
}

func (s Style) PaddingBottom(value unit.Dp) Style {
	return s.paddingSide(sideBottom, value)
}

func (s Style) PaddingLeft(value unit.Dp) Style {
	return s.paddingSide(sideLeft, value)
}

func (s Style) paddingSide(side uint8, value unit.Dp) Style {
	if !finite(float32(value)) {
		return s
	}
	return s.editBox(func(box *BoxStyle) {
		padding := insets(box.Padding)
		setInsetSide(&padding, side, value)
		box.Padding = &padding
		box.paddingMask |= side
	})
}

func (s Style) Margin(value unit.Dp) Style {
	if !finite(float32(value)) {
		return s
	}
	return s.editBox(func(box *BoxStyle) {
		box.Margin = &Insets{Top: value, Right: value, Bottom: value, Left: value}
		box.marginMask = sidesAll
	})
}

func (s Style) MarginX(value unit.Dp) Style {
	if !finite(float32(value)) {
		return s
	}
	return s.editBox(func(box *BoxStyle) {
		margin := insets(box.Margin)
		margin.Left, margin.Right = value, value
		box.Margin = &margin
		box.marginMask |= sideLeft | sideRight
	})
}

func (s Style) MarginY(value unit.Dp) Style {
	if !finite(float32(value)) {
		return s
	}
	return s.editBox(func(box *BoxStyle) {
		margin := insets(box.Margin)
		margin.Top, margin.Bottom = value, value
		box.Margin = &margin
		box.marginMask |= sideTop | sideBottom
	})
}

func (s Style) MarginTop(value unit.Dp) Style {
	return s.marginSide(sideTop, value)
}

func (s Style) MarginRight(value unit.Dp) Style {
	return s.marginSide(sideRight, value)
}

func (s Style) MarginBottom(value unit.Dp) Style {
	return s.marginSide(sideBottom, value)
}

func (s Style) MarginLeft(value unit.Dp) Style {
	return s.marginSide(sideLeft, value)
}

func (s Style) marginSide(side uint8, value unit.Dp) Style {
	if !finite(float32(value)) {
		return s
	}
	return s.editBox(func(box *BoxStyle) {
		margin := insets(box.Margin)
		setInsetSide(&margin, side, value)
		box.Margin = &margin
		box.marginMask |= side
	})
}

func (s Style) Overflow(value Overflow) Style {
	return s.editBox(func(box *BoxStyle) { box.Overflow = new(value) })
}

// Cursor sets the pointer cursor over the rendered box.
func (s Style) Cursor(value pointer.Cursor) Style {
	return s.editBox(func(box *BoxStyle) { box.Cursor = new(value) })
}

func insets(value *Insets) Insets {
	if value == nil {
		return Insets{}
	}
	return *value
}

func setInsetSide(insets *Insets, side uint8, value unit.Dp) {
	switch side {
	case sideTop:
		insets.Top = value
	case sideRight:
		insets.Right = value
	case sideBottom:
		insets.Bottom = value
	case sideLeft:
		insets.Left = value
	}
}
