package style

import (
	"gioui.org/font"
	giotext "gioui.org/text"
	"gioui.org/unit"
)

func (s Style) editText(edit func(*TextStyle)) Style {
	text := TextStyle{}
	if s.text != nil {
		text = *s.text
	}
	s.text = &text
	edit(s.text)
	return s
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

// Underline toggles a rule under every text line.
func (s Style) Underline(value bool) Style {
	return s.editText(func(text *TextStyle) { text.Underline = new(value) })
}
