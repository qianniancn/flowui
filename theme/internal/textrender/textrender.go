// Package textrender contains the small Gio adapters shared by theme recipes.
// It is intentionally below theme/internal so behavior widgets cannot import
// it and start interpreting provider declarations again.
package textrender

import (
	"image"
	"image/color"

	"gioui.org/font"
	"gioui.org/layout"
	giotext "gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

// Label renders a provider-resolved text part using the active theme's Gio
// material bridge. All declaration interpretation stays in this theme-side
// package; widgets only forward the resolved part and semantic text.
func Label(active *theme.Theme, gtx layout.Context, value string, resolved style.ResolvedStyle, size unit.Sp) layout.Dimensions {
	if active == nil {
		panic(theme.ErrMissingVisualProvider)
	}
	if resolved.Text == nil {
		panic(theme.ErrIncompleteVisualProvider)
	}
	if size <= 0 {
		size = active.Typography.BodySize
		if resolved.Text.FontSize != nil && *resolved.Text.FontSize > 0 {
			size = *resolved.Text.FontSize
		}
	}
	label := material.Label(theme.MaterialOf(active), size, value)
	label.SelectionColor = active.Semantic().Selection
	if value, ok := computed.Color(resolved.Text.Color); ok {
		label.Color = value
	} else {
		panic(theme.ErrIncompleteVisualProvider)
	}
	if resolved.Text.FontWeight != nil {
		label.Font.Weight = font.Weight(*resolved.Text.FontWeight)
	}
	if resolved.Text.Typeface != nil {
		label.Font.Typeface = *resolved.Text.Typeface
	}
	if resolved.Text.FontStyle != nil {
		label.Font.Style = *resolved.Text.FontStyle
	}
	if resolved.Text.Align != nil {
		label.Alignment = alignment(*resolved.Text.Align)
	}
	if resolved.Text.MaxLines != nil {
		label.MaxLines = *resolved.Text.MaxLines
	}
	if resolved.Text.LineHeight != nil {
		label.LineHeight = *resolved.Text.LineHeight
	}
	if resolved.Text.LineHeightScale != nil {
		label.LineHeightScale = *resolved.Text.LineHeightScale
	}
	if resolved.Text.Wrap != nil {
		label.WrapPolicy = *resolved.Text.Wrap
	} else {
		label.WrapPolicy = giotext.WrapWords
	}
	if resolved.Text.Truncator != nil {
		label.Truncator = *resolved.Text.Truncator
	}
	return label.Layout(gtx)
}

func alignment(value style.TextAlign) giotext.Alignment {
	switch value {
	case style.TextAlignCenter:
		return giotext.Middle
	case style.TextAlignEnd:
		return giotext.End
	default:
		return giotext.Start
	}
}

// Color resolves a semantic text part for a theme-owned icon renderer.
func Color(resolved style.ResolvedStyle) color.NRGBA {
	value, ok := ColorValue(resolved)
	if !ok {
		panic(theme.ErrIncompleteVisualProvider)
	}
	return value
}

// ColorValue reports whether a resolved text Part carries a concrete color.
func ColorValue(resolved style.ResolvedStyle) (color.NRGBA, bool) {
	if resolved.Text == nil {
		return color.NRGBA{}, false
	}
	return computed.Color(resolved.Text.Color)
}

// ColorOr resolves a text Part while retaining a provider-owned fallback for
// declarations that intentionally inherit their parent typography.
func ColorOr(resolved style.ResolvedStyle, fallback color.NRGBA) color.NRGBA {
	if value, ok := ColorValue(resolved); ok {
		return value
	}
	return fallback
}

// PaintColor resolves the center color of a selection brush for a theme-owned
// editor adapter.
func PaintColor(resolved style.ResolvedStyle) color.NRGBA {
	if resolved.Paint == nil {
		panic(theme.ErrIncompleteVisualProvider)
	}
	brush, ok := computed.Brush(resolved.Paint.Background)
	if !ok {
		panic(theme.ErrIncompleteVisualProvider)
	}
	return brush.ColorAt(.5)
}

// PartWidth resolves provider-owned horizontal geometry.
func PartWidth(resolved style.ResolvedStyle) unit.Dp {
	if resolved.Box == nil || resolved.Box.Width == nil {
		panic(theme.ErrIncompleteVisualProvider)
	}
	return *resolved.Box.Width
}

// PartSize resolves provider-owned icon geometry in device pixels.
func PartSize(gtx layout.Context, resolved style.ResolvedStyle) image.Point {
	if resolved.Box == nil {
		panic(theme.ErrIncompleteVisualProvider)
	}
	width, height := 0, 0
	if resolved.Box.Width != nil {
		width = max(gtx.Dp(*resolved.Box.Width), 0)
	}
	if resolved.Box.Height != nil {
		height = max(gtx.Dp(*resolved.Box.Height), 0)
	}
	if width == 0 {
		width = height
	}
	if height == 0 {
		height = width
	}
	if width <= 0 || height <= 0 {
		panic(theme.ErrIncompleteVisualProvider)
	}
	return image.Pt(width, height)
}

// TypographySize returns the resolved font size or the active body size when
// the part intentionally inherits the provider's body typography.
func TypographySize(active *theme.Theme, resolved style.ResolvedStyle) unit.Sp {
	if active == nil || resolved.Text == nil {
		panic(theme.ErrIncompleteVisualProvider)
	}
	if resolved.Text.FontSize != nil && *resolved.Text.FontSize > 0 {
		return *resolved.Text.FontSize
	}
	return active.Typography.BodySize
}

// InputEditorStyle resolves the concrete bridge needed by Gio's editor. The
// fallback values come from the active theme, while explicit Part declarations
// take precedence; this keeps color and typography interpretation in the theme
// side of the provider boundary.
func InputEditorStyle(active *theme.Theme, content, placeholder, selection style.ResolvedStyle) theme.InputEditorStyle {
	if active == nil {
		panic(theme.ErrMissingVisualProvider)
	}
	semantic := active.Semantic()
	result := theme.InputEditorStyle{
		Foreground:  semantic.FieldForeground,
		Placeholder: semantic.FieldPlaceholder,
		Selection:   semantic.Selection,
	}
	if content.Text != nil {
		if value, ok := computed.Color(content.Text.Color); ok {
			result.Foreground = value
		}
		result.TextSize, result.LineHeight = textMetrics(content)
		if content.Text.LineHeightScale != nil {
			result.LineHeightScale = *content.Text.LineHeightScale
			result.HasLineHeightScale = true
		}
		if content.Text.Typeface != nil {
			result.Typeface = *content.Text.Typeface
			result.HasTypeface = true
		}
		if content.Text.FontStyle != nil {
			result.FontStyle = *content.Text.FontStyle
			result.HasFontStyle = true
		}
		if content.Text.FontWeight != nil {
			result.FontWeight = font.Weight(*content.Text.FontWeight)
			result.HasFontWeight = true
		}
	}
	if placeholder.Text != nil {
		if value, ok := computed.Color(placeholder.Text.Color); ok {
			result.Placeholder = value
		}
	}
	if selection.Paint != nil {
		if brush, ok := computed.Brush(selection.Paint.Background); ok {
			result.Selection = brush.ColorAt(.5)
		}
	}
	return result
}

func textMetrics(resolved style.ResolvedStyle) (size, lineHeight unit.Sp) {
	if resolved.Text == nil {
		return 0, 0
	}
	if resolved.Text.FontSize != nil {
		size = *resolved.Text.FontSize
	}
	if resolved.Text.LineHeight != nil {
		lineHeight = *resolved.Text.LineHeight
	}
	return size, lineHeight
}
