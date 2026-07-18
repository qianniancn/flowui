package input

import (
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/qianniancn/FlowUI/internal/frame"
)

func editorLayoutFor(ctx *frame.Context, editor *widget.Editor, hint string, style inputStyle, textSize, lineHeight unit.Sp) layout.Widget {
	editorStyle := material.Editor(frame.ActiveTheme(ctx).Material, editor, hint)
	editorStyle.TextSize = textSize
	editorStyle.LineHeight = lineHeight
	editorStyle.Color = style.Foreground
	editorStyle.HintColor = style.Placeholder
	editorStyle.SelectionColor = style.Selection
	return editorStyle.Layout
}
