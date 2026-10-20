package style

import "gioui.org/unit"

func (s Style) editPaint(edit func(*PaintStyle)) Style {
	paint := PaintStyle{}
	if s.paint != nil {
		paint = *s.paint
		if paint.Border != nil {
			border := *paint.Border
			paint.Border = &border
		}
		if paint.BorderBottom != nil {
			border := *paint.BorderBottom
			paint.BorderBottom = &border
		}
		paint.Shadows = append([]Shadow(nil), paint.Shadows...)
	}
	s.paint = &paint
	edit(s.paint)
	return s
}

func (s Style) Background(source PaintSource) Style {
	return s.editPaint(func(paint *PaintStyle) {
		paint.Background = clonePaintSource(source)
		paint.backgroundSet = true
	})
}

func (s Style) BackgroundNone() Style {
	return s.editPaint(func(paint *PaintStyle) {
		paint.Background = nil
		paint.backgroundSet = true
	})
}

func (s Style) BorderColor(value ColorSource) Style {
	return s.editPaint(func(paint *PaintStyle) {
		if paint.Border == nil {
			paint.Border = &BorderStyle{}
		}
		paint.Border.Color = cloneColorSource(value)
	})
}

func (s Style) BorderWidth(value unit.Dp) Style {
	if !finite(float32(value)) {
		return s
	}
	return s.editPaint(func(paint *PaintStyle) {
		if paint.Border == nil {
			paint.Border = &BorderStyle{}
		}
		paint.Border.Width = new(value)
	})
}

// BorderLine selects the border stroke pattern. It only applies to Border,
// not to BorderBottom.
func (s Style) BorderLine(value LineStyle) Style {
	return s.editPaint(func(paint *PaintStyle) {
		if paint.Border == nil {
			paint.Border = &BorderStyle{}
		}
		paint.Border.LineStyle = new(value)
	})
}

// BorderBottom sets the bottom border width and color without changing the
// other sides. The pair travels together: the split setters this replaced
// had no caller that set only one half (every site chained both), and the
// invalid-width degradation matches the old chained behavior exactly -- the
// color still lands, only the width write is skipped.
func (s Style) BorderBottom(width unit.Dp, color ColorSource) Style {
	return s.editPaint(func(paint *PaintStyle) {
		if paint.BorderBottom == nil {
			paint.BorderBottom = &BorderStyle{}
		}
		paint.BorderBottom.Color = cloneColorSource(color)
		if finite(float32(width)) {
			paint.BorderBottom.Width = new(width)
		}
	})
}

func (s Style) Radius(value unit.Dp) Style {
	if !finite(float32(value)) {
		return s
	}
	return s.editPaint(func(paint *PaintStyle) {
		paint.Radius = new(value)
		paint.Radii = nil
		paint.radiusMask = 0
	})
}

func (s Style) RadiusTopLeft(value unit.Dp) Style {
	return s.radiusCorner(sideTop|sideLeft, value)
}

func (s Style) RadiusTopRight(value unit.Dp) Style {
	return s.radiusCorner(sideTop|sideRight, value)
}

func (s Style) RadiusBottomRight(value unit.Dp) Style {
	return s.radiusCorner(sideBottom|sideRight, value)
}

func (s Style) RadiusBottomLeft(value unit.Dp) Style {
	return s.radiusCorner(sideBottom|sideLeft, value)
}

func (s Style) radiusCorner(corner uint8, value unit.Dp) Style {
	if !finite(float32(value)) {
		return s
	}
	return s.editPaint(func(paint *PaintStyle) {
		radii := CornerRadii{}
		if paint.Radii != nil {
			radii = *paint.Radii
		} else if paint.Radius != nil {
			radii = uniformRadii(*paint.Radius)
		}
		switch corner {
		case sideTop | sideLeft:
			radii.TopLeft = value
		case sideTop | sideRight:
			radii.TopRight = value
		case sideBottom | sideRight:
			radii.BottomRight = value
		case sideBottom | sideLeft:
			radii.BottomLeft = value
		}
		paint.Radii = &radii
		paint.radiusMask |= cornerMask(corner)
	})
}

func (s Style) BoxShadow(offsetX, offsetY, blur, spread unit.Dp, col ColorSource) Style {
	return s.boxShadow(offsetX, offsetY, blur, spread, col, false)
}

// BoxShadowTransient is BoxShadow for decoration that must not reserve visual
// overflow in clipping containers. Focus and validation halos use it so a
// state change cannot reflow a scroll viewport.
func (s Style) BoxShadowTransient(offsetX, offsetY, blur, spread unit.Dp, col ColorSource) Style {
	return s.boxShadow(offsetX, offsetY, blur, spread, col, true)
}

func (s Style) boxShadow(offsetX, offsetY, blur, spread unit.Dp, col ColorSource, transient bool) Style {
	if !finite(float32(offsetX)) || !finite(float32(offsetY)) ||
		!finite(float32(blur)) || !finite(float32(spread)) {
		return s
	}
	return s.editPaint(func(paint *PaintStyle) {
		paint.Shadows = append(paint.Shadows, Shadow{
			OffsetX:   offsetX,
			OffsetY:   offsetY,
			Blur:      max(blur, 0),
			Spread:    spread,
			Color:     cloneColorSource(col),
			Transient: transient,
		})
		paint.shadowsSet = true
	})
}

// Shadow applies a theme-controlled multi-layer shadow profile.
func (s Style) Shadow(profile ShadowProfile) Style {
	return s.editPaint(func(paint *PaintStyle) {
		paint.Shadows = append(paint.Shadows, Shadow{Profile: new(profile)})
		paint.shadowsSet = true
	})
}

func (s Style) BoxShadowNone() Style {
	return s.editPaint(func(paint *PaintStyle) {
		paint.Shadows = nil
		paint.shadowsSet = true
	})
}

// VisualOutset declares the minimum visual space this box may paint beyond
// its layout bounds. Clipping containers reserve the declaration together
// with the box's shadow and outline extents.
func (s Style) VisualOutset(top, right, bottom, left unit.Dp) Style {
	if !finite(float32(top)) || !finite(float32(right)) ||
		!finite(float32(bottom)) || !finite(float32(left)) {
		return s
	}
	return s.editPaint(func(paint *PaintStyle) {
		paint.VisualOutset = &Insets{
			Top:    max(top, 0),
			Right:  max(right, 0),
			Bottom: max(bottom, 0),
			Left:   max(left, 0),
		}
		paint.visualOutsetSet = true
	})
}

// VisualOutsetNone clears a visual overflow declaration inherited from a
// lower-precedence style.
func (s Style) VisualOutsetNone() Style {
	return s.editPaint(func(paint *PaintStyle) {
		paint.VisualOutset = nil
		paint.visualOutsetSet = true
	})
}

func (s Style) Outline(width, offset unit.Dp, col ColorSource) Style {
	if !finite(float32(width)) || !finite(float32(offset)) {
		return s
	}
	return s.editPaint(func(paint *PaintStyle) {
		paint.Outline = &OutlineStyle{Width: max(width, 0), Offset: max(offset, 0), Color: cloneColorSource(col)}
	})
}

func (s Style) Opacity(value float32) Style {
	if !finite(value) {
		return s
	}
	return s.editPaint(func(paint *PaintStyle) { paint.Opacity = new(clamp01(value)) })
}

// Dim scales the alpha of every color this declaration resolves to, leaving
// shadow layers alone. It is a declaration-time transform, not a draw-time one:
// Dim(0.5) on a solid color produces a different solid color, whereas
// Opacity(0.5) produces the same color composited as a faded group. The two are
// not interchangeable wherever colors overlap -- a thumb's border over its own
// fill, or a shadow behind it -- so pick by what the design asks for, not by
// which one looks equivalent on an opaque background.
//
// It exists because a rule table cannot express "scale whatever the base said
// by half". Rule.Set is an absolute declaration, so a component whose disabled
// state is a transform over its own enabled colors would otherwise have to
// restate the whole cross product of its other states.
//
// Shadows are deliberately out of scope: a shadow whose alpha is derived from a
// color has to be restated by whoever wrote it, because the two built-in cases
// disagree -- the switch's shadow is independent of its palette and must not
// dim, the slider's is derived from its thumb and must.
func (s Style) Dim(factor float32) Style {
	if !finite(factor) {
		return s
	}
	result := s
	result.dim = new(clamp01(factor))
	return result
}

func uniformRadii(value unit.Dp) CornerRadii {
	return CornerRadii{TopLeft: value, TopRight: value, BottomRight: value, BottomLeft: value}
}

func cornerMask(corner uint8) uint8 {
	switch corner {
	case sideTop | sideLeft:
		return 1 << 0
	case sideTop | sideRight:
		return 1 << 1
	case sideBottom | sideRight:
		return 1 << 2
	case sideBottom | sideLeft:
		return 1 << 3
	default:
		return 0
	}
}
