package ui

import (
	giotext "gioui.org/text"
	textui "github.com/qianniancn/flowui/internal/components/text"
)

type TextWidget = textui.Widget
type SelectableTextWidget = textui.Widget
type TextAlignment = giotext.Alignment
type TextWrapPolicy = giotext.WrapPolicy

const (
	TextAlignStart  = giotext.Start
	TextAlignEnd    = giotext.End
	TextAlignCenter = giotext.Middle

	TextWrapHeuristically = giotext.WrapHeuristically
	TextWrapWords         = giotext.WrapWords
	TextWrapGraphemes     = giotext.WrapGraphemes
)

func Text(value string) TextWidget {
	return textui.New(value)
}

func SelectableText(key, value string) SelectableTextWidget {
	return textui.Selectable(key, value)
}
