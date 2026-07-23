package input

import (
	"gioui.org/unit"
	"github.com/qianniancn/FlowUI/internal/field"
	flowstyle "github.com/qianniancn/FlowUI/internal/style"
	"github.com/qianniancn/FlowUI/internal/theme"
)

type inputStyle = field.Colors

func inputDefaultDeclaration(activeTheme *theme.Theme, variant InputVariant, fullWidth bool) flowstyle.Style {
	tokens := activeTheme.Components.Input
	return field.DefaultDeclaration(activeTheme, variant, field.DeclarationOptions{
		Height: tokens.Height, Radius: tokens.Radius, PaddingX: tokens.PaddingX,
		TextSize: tokens.TextSize, LineHeight: tokens.LineHeight,
		FocusRingWidth: tokens.FocusRingWidth, InvalidOutlineWidth: tokens.InvalidOutlineWidth,
		ShadowColor: tokens.ShadowColor, ShadowOpacity: tokens.ShadowOpacity,
		ShadowStrength: tokens.ShadowStrength, FillWidth: fullWidth,
	})
}

func textAreaDefaultDeclaration(activeTheme *theme.Theme, variant TextAreaVariant, fullWidth bool, minHeight unit.Dp) flowstyle.Style {
	tokens := activeTheme.Components.TextArea
	return field.DefaultDeclaration(activeTheme, variant, field.DeclarationOptions{
		MinHeight: max(tokens.MinHeight, minHeight), Radius: tokens.Radius,
		PaddingX: tokens.PaddingX, PaddingY: tokens.PaddingY,
		TextSize: tokens.TextSize, LineHeight: tokens.LineHeight,
		FocusRingWidth: tokens.FocusRingWidth, InvalidOutlineWidth: tokens.InvalidOutlineWidth,
		ShadowColor: tokens.ShadowColor, ShadowOpacity: tokens.ShadowOpacity,
		ShadowStrength: tokens.ShadowStrength, FillWidth: fullWidth,
	})
}

func resolvedInputStyle(root, placeholder, selection flowstyle.ResolvedStyle, activeTheme *theme.Theme) inputStyle {
	return field.ResolvedColors(root, placeholder, selection, activeTheme)
}

func resolvedTypography(root flowstyle.ResolvedStyle, textSize unit.Sp, lineHeight unit.Sp) (unit.Sp, unit.Sp) {
	if root.Text == nil {
		return textSize, lineHeight
	}
	if root.Text.FontSize != nil {
		textSize = *root.Text.FontSize
	}
	if root.Text.LineHeight != nil {
		lineHeight = *root.Text.LineHeight
	}
	return textSize, lineHeight
}
