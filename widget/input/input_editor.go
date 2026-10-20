package input

import (
	"gioui.org/layout"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/theme"
)

func applyInputEditorStyle(target *material.EditorStyle, resolved theme.InputEditorStyle) {
	if resolved.TextSize > 0 {
		target.TextSize = resolved.TextSize
	}
	if resolved.LineHeight > 0 {
		target.LineHeight = resolved.LineHeight
	}
	if resolved.HasLineHeightScale {
		target.LineHeightScale = resolved.LineHeightScale
	}
	if resolved.HasTypeface {
		target.Font.Typeface = resolved.Typeface
	}
	if resolved.HasFontStyle {
		target.Font.Style = resolved.FontStyle
	}
	if resolved.HasFontWeight {
		target.Font.Weight = resolved.FontWeight
	}
	target.Color = resolved.Foreground
	target.HintColor = resolved.Placeholder
	target.SelectionColor = resolved.Selection
}

func editorLayoutFor(ctx *core.Context, editor *widget.Editor, hint string, resolved theme.InputEditorStyle) layout.Widget {
	editorStyle := material.Editor(core.ActiveMaterial(ctx), editor, hint)
	applyInputEditorStyle(&editorStyle, resolved)
	return editorStyle.Layout
}
