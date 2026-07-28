package input

import (
	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/field"
	flowstyle "github.com/qianniancn/flowui/internal/style"
	"github.com/qianniancn/flowui/internal/theme"
)

type inputStyle = field.Colors

// styleDeclarations follows Button's four-slot protocol:
// defaults (primary chrome) | variant (secondary chrome) | size (unused) | instance.
func (i InputWidget) styleDeclarations(activeTheme *theme.Theme) (defaults, variant, size flowstyle.Style) {
	switch i.variant {
	case InputSecondary:
		return flowstyle.Style{}, inputDefaultDeclaration(activeTheme, InputSecondary, i.fullWidth), flowstyle.Style{}
	default:
		return inputDefaultDeclaration(activeTheme, InputPrimary, i.fullWidth), flowstyle.Style{}, flowstyle.Style{}
	}
}

func (t TextAreaWidget) styleDeclarations(activeTheme *theme.Theme, height unit.Dp) (defaults, variant, size flowstyle.Style) {
	switch t.variant {
	case TextAreaSecondary:
		return flowstyle.Style{}, textAreaDefaultDeclaration(activeTheme, TextAreaSecondary, t.fullWidth, height), flowstyle.Style{}
	default:
		return textAreaDefaultDeclaration(activeTheme, TextAreaPrimary, t.fullWidth, height), flowstyle.Style{}, flowstyle.Style{}
	}
}

func (g InputGroupWidget) styleDeclarations(activeTheme *theme.Theme, minHeight unit.Dp) (defaults, variant, size flowstyle.Style) {
	switch g.variant {
	case InputSecondary:
		return flowstyle.Style{}, inputGroupDefaultDeclaration(activeTheme, InputSecondary, g.fullWidth, minHeight), flowstyle.Style{}
	default:
		return inputGroupDefaultDeclaration(activeTheme, InputPrimary, g.fullWidth, minHeight), flowstyle.Style{}, flowstyle.Style{}
	}
}

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

func textAreaDefaultDeclaration(activeTheme *theme.Theme, variant TextAreaVariant, fullWidth bool, height unit.Dp) flowstyle.Style {
	tokens := activeTheme.Components.TextArea
	return field.DefaultDeclaration(activeTheme, variant, field.DeclarationOptions{
		Height: max(tokens.MinHeight, height), Radius: tokens.Radius,
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
