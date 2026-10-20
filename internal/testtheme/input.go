package testtheme

import (
	"image/color"
	"time"

	"gioui.org/font"
	"gioui.org/io/pointer"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

func (behaviorProvider) Input(active *theme.Theme, request theme.InputRequest) theme.InputVisual {
	height, radius, paddingX, paddingY, textSize, lineHeight := unit.Dp(36), unit.Dp(4), unit.Dp(12), unit.Dp(0), unit.Sp(14), unit.Sp(20)
	switch request.Mode {
	case theme.InputModeTextArea:
		height, radius, paddingX, paddingY, textSize, lineHeight = max(unit.Dp(38), request.Height), 4, 12, 8, 14, 20
	case theme.InputModeGroup:
		height, radius, paddingX, paddingY, textSize, lineHeight = max(unit.Dp(36), request.Height), 4, 12, 0, 14, 20
	default:
		height = max(height, request.Height)
	}
	content := style.Style{}.
		Background(style.Solid(background)).
		TextColor(style.Solid(foreground)).
		Radius(radius).
		Overflow(style.OverflowHidden).
		Cursor(pointer.CursorText).
		Opacity(1).
		FontSize(textSize).
		LineHeight(lineHeight).
		PaddingX(paddingX).
		PaddingY(paddingY).
		Transition(style.PropBackgroundColor, 100*time.Millisecond).
		On(style.Hovered, style.Style{}.Background(style.Solid(accent))).
		On(style.Focused, style.Style{}.Outline(2, 0, style.Solid(focus))).
		On(style.Invalid, style.Style{}.BorderColor(style.Solid(focus))).
		On(style.Disabled, style.Style{}.Opacity(active.DisabledOpacityValue()))
	if request.Mode == theme.InputModeSingleLine || request.Mode == theme.InputModeTextArea {
		content = content.Height(height)
	} else {
		content = content.MinHeight(height)
	}
	if request.FullWidth {
		content = content.FillWidth()
	}
	defaults := style.Style{}.
		Part(style.PartRoot, content).
		Part(style.PartPlaceholder, style.Style{}.TextColor(style.Solid(muted))).
		Part(style.PartSelection, style.Style{}.Background(style.Solid(focus)))
	if request.Mode == theme.InputModeGroup {
		root := content.PaddingX(0).PaddingY(0)
		groupContent := content
		if request.Height > 0 {
			groupContent = groupContent.PaddingY(8).MinHeight(38)
		} else {
			groupContent = groupContent.MinHeight(36)
		}
		defaults = defaults.
			Part(style.PartRoot, root).
			Cursor(pointer.CursorDefault).
			Part(style.PartContent, groupContent.
				On(style.Disabled, style.Style{}.Cursor(pointer.CursorNotAllowed))).
			Part(style.PartPrefix, style.Style{}.TextColor(style.Solid(muted))).
			Part(style.PartSuffix, style.Style{}.TextColor(style.Solid(muted))).
			Part(style.PartIndicator, style.Style{}.Width(1).Background(style.Solid(muted)))
	}
	return theme.InputVisual{
		Defaults:    defaults,
		Override:    theme.InputOverrideStyle(request),
		EditorStyle: behaviorInputEditorStyle(foreground, muted, focus, textSize, lineHeight),
	}
}

func behaviorInputEditorStyle(foreground, placeholder, selection color.NRGBA, textSize, lineHeight unit.Sp) theme.InputEditorStyleResolver {
	return func(content, hint, selected style.ResolvedStyle) theme.InputEditorStyle {
		result := theme.InputEditorStyle{
			TextSize: textSize, LineHeight: lineHeight,
			Foreground: foreground, Placeholder: placeholder, Selection: selection,
		}
		if content.Text != nil {
			if value, ok := behaviorColor(content.Text.Color); ok {
				result.Foreground = value
			}
			if content.Text.FontSize != nil {
				result.TextSize = *content.Text.FontSize
			}
			if content.Text.LineHeight != nil {
				result.LineHeight = *content.Text.LineHeight
			}
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
		if hint.Text != nil {
			if value, ok := behaviorColor(hint.Text.Color); ok {
				result.Placeholder = value
			}
		}
		if selected.Paint != nil {
			if value, ok := behaviorPaintColor(selected.Paint.Background); ok {
				result.Selection = value
			}
		}
		return result
	}
}
