package ui

import (
	giotext "gioui.org/text"
	textui "github.com/qianniancn/flowui/internal/components/text"
)

type TextWidget = textui.Widget

type SelectableTextWidget = textui.Widget

// TextAlignment controls text alignment within its layout width.
type TextAlignment = giotext.Alignment

// TextWrapPolicy controls where text may wrap.
type TextWrapPolicy = giotext.WrapPolicy

const (
	TextAlignStart  = giotext.Start
	TextAlignEnd    = giotext.End
	TextAlignCenter = giotext.Middle

	TextWrapHeuristically = giotext.WrapHeuristically
	TextWrapWords         = giotext.WrapWords
	TextWrapGraphemes     = giotext.WrapGraphemes
)

// Text creates a non-interactive text widget.
func Text(value string) TextWidget {
	return textui.New(value)
}

// SelectableText creates text that can be selected and copied.
func SelectableText(key, value string) SelectableTextWidget {
	return textui.Selectable(key, value)
}
