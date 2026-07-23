package style

import (
	"maps"
	"time"

	"gioui.org/font"
	"gioui.org/io/pointer"
	giotext "gioui.org/text"
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

func (s Style) editPaint(edit func(*PaintStyle)) Style {
	paint := PaintStyle{}
	if s.paint != nil {
		paint = *s.paint
		if paint.Border != nil {
			border := *paint.Border
			paint.Border = &border
		}
		paint.Shadows = append([]Shadow(nil), paint.Shadows...)
	}
	s.paint = &paint
	edit(s.paint)
	return s
}

func (s Style) editText(edit func(*TextStyle)) Style {
	text := TextStyle{}
	if s.text != nil {
		text = *s.text
	}
	s.text = &text
	edit(s.text)
	return s
}

func (s Style) editTransform(edit func(*TransformStyle)) Style {
	transform := TransformStyle{}
	if s.trans != nil {
		transform = *s.trans
	}
	s.trans = &transform
	edit(s.trans)
	return s
}

// Use applies theme metrics as defaults within this declaration. Explicit
// properties in the same Style take precedence over tokens.
func (s Style) Use(tokens ...StyleToken) Style {
	s.tokens = append(append([]StyleToken(nil), s.tokens...), tokens...)
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
	if !finite(float32(offsetX)) || !finite(float32(offsetY)) ||
		!finite(float32(blur)) || !finite(float32(spread)) {
		return s
	}
	return s.editPaint(func(paint *PaintStyle) {
		paint.Shadows = append(paint.Shadows, Shadow{
			OffsetX: offsetX,
			OffsetY: offsetY,
			Blur:    max(blur, 0),
			Spread:  spread,
			Color:   cloneColorSource(col),
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

func (s Style) TextColor(value ColorSource) Style {
	return s.editText(func(text *TextStyle) { text.Color = cloneColorSource(value) })
}

func (s Style) FontSize(value unit.Sp) Style {
	if !finite(float32(value)) {
		return s
	}
	return s.editText(func(text *TextStyle) { text.FontSize = new(value) })
}

func (s Style) FontWeight(value int) Style {
	return s.editText(func(text *TextStyle) { text.FontWeight = new(value) })
}

func (s Style) Typeface(value font.Typeface) Style {
	return s.editText(func(text *TextStyle) { text.Typeface = new(value) })
}

func (s Style) FontStyle(value font.Style) Style {
	return s.editText(func(text *TextStyle) { text.FontStyle = new(value) })
}

func (s Style) LineHeight(value unit.Sp) Style {
	if !finite(float32(value)) {
		return s
	}
	return s.editText(func(text *TextStyle) { text.LineHeight = new(value) })
}

func (s Style) LineHeightScale(value float32) Style {
	if !finite(value) {
		return s
	}
	return s.editText(func(text *TextStyle) { text.LineHeightScale = new(max(value, 0)) })
}

func (s Style) MaxLines(value int) Style {
	return s.editText(func(text *TextStyle) { text.MaxLines = new(max(value, 0)) })
}

func (s Style) TextAlign(value TextAlign) Style {
	return s.editText(func(text *TextStyle) { text.Align = new(value) })
}

func (s Style) Wrap(value giotext.WrapPolicy) Style {
	return s.editText(func(text *TextStyle) { text.Wrap = new(value) })
}

func (s Style) Truncator(value string) Style {
	return s.editText(func(text *TextStyle) { text.Truncator = new(value) })
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

func (s Style) Scale(x, y float32) Style {
	if !finite(x) || !finite(y) {
		return s
	}
	return s.editTransform(func(transform *TransformStyle) {
		transform.ScaleX = new(x)
		transform.ScaleY = new(y)
	})
}

func (s Style) Rotate(radians float32) Style {
	if !finite(radians) {
		return s
	}
	return s.editTransform(func(transform *TransformStyle) { transform.Rotate = new(radians) })
}

func (s Style) Transition(property PropertyID, duration time.Duration, options ...TransitionOption) Style {
	s.transitions = append([]Transition(nil), s.transitions...)
	transition := transition(&s.transitions, property)
	transition.Duration = max(duration, 0)
	for _, option := range options {
		if option != nil {
			option(transition)
		}
	}
	return s
}

func (s Style) When(predicate Condition, override Style) Style {
	if predicate == nil {
		return s
	}
	s.conditions = append(append([]condition(nil), s.conditions...), condition{predicate: predicate, override: override})
	return s
}

func (s Style) Part(part Part, override Style) Style {
	if part == PartRoot {
		return merge(s, override)
	}
	result := s
	result.parts = make(map[Part]Style, len(s.parts)+1)
	maps.Copy(result.parts, s.parts)
	result.parts[part] = merge(result.parts[part], override)
	return result
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

func transition(transitions *[]Transition, property PropertyID) *Transition {
	for index := range *transitions {
		if (*transitions)[index].Property == property {
			return &(*transitions)[index]
		}
	}
	*transitions = append(*transitions, Transition{Property: property})
	return &(*transitions)[len(*transitions)-1]
}
