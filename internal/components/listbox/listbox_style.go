package listbox

import (
	"image/color"

	"github.com/qianniancn/flowui/internal/style"
	"github.com/qianniancn/flowui/internal/theme"
)

type listBoxItemStyle struct {
	fg          color.NRGBA
	description color.NRGBA
	indicator   color.NRGBA
	selected    float32
}

func listBoxItemDefaultDeclaration(activeTheme *theme.Theme) style.Style {
	tokens := activeTheme.Components.ListBox
	pressedScale := tokens.PressedScale
	if pressedScale <= 0 || pressedScale > 1 {
		pressedScale = 0.98
	}
	item := style.Style{}.
		MinHeight(tokens.ItemMinHeight).
		PaddingX(tokens.ItemPaddingX).
		PaddingY(tokens.ItemPaddingY).
		Radius(tokens.ItemRadius).
		Background(style.RGBA(0x00000000)).
		TextColor(style.SolidColor{Color: activeTheme.Palette.Foreground}).
		BorderWidth(tokens.FocusRingWidth).
		BorderColor(style.WithAlpha(style.TokenFocus, 0)).
		Opacity(1).
		Scale(1, 1).
		Transition(style.PropBackgroundColor, listBoxItemColorDuration).
		Transition(style.PropBorderColor, listBoxItemFocusDuration).
		Transition(style.PropTransform, listBoxItemPressOutDuration).
		When(style.Hovered, style.Style{}.Background(style.TokenSurfaceRaised)).
		When(style.Pressed, style.Style{}.
			Background(style.TokenSurfacePressed).
			Scale(pressedScale, pressedScale).
			Transition(style.PropTransform, listBoxItemPressInDuration)).
		When(style.FocusVisible, style.Style{}.BorderColor(style.TokenFocus)).
		When(style.Disabled, style.Style{}.Opacity(activeTheme.DisabledOpacityValue()))
	return style.Style{}.Part(style.PartItem, item)
}

func listBoxItemVariantDeclaration(activeTheme *theme.Theme, variant ListBoxItemVariant) style.Style {
	if variant != ListBoxItemDanger {
		return style.Style{}
	}
	item := style.Style{}.
		TextColor(style.TokenDanger).
		When(style.FocusVisible, style.Style{}.BorderColor(
			style.WithAlpha(style.TokenDanger, float32(activeTheme.Palette.Focus.A)/255),
		))
	return style.Style{}.Part(style.PartItem, item)
}

func listBoxItemStyleFor(theme *theme.Theme, variant ListBoxItemVariant, disabled bool) listBoxItemStyle {
	style := listBoxItemStyle{
		fg:          theme.Palette.Foreground,
		description: theme.Palette.MutedForeground,
		indicator:   theme.Palette.Foreground,
	}
	if variant == ListBoxItemDanger {
		style.fg = theme.Palette.Danger
		style.indicator = theme.Palette.Danger
	}
	if disabled {
		style.fg = theme.DisabledColor(style.fg)
		style.description = theme.DisabledColor(style.description)
		style.indicator = theme.DisabledColor(style.indicator)
	}
	return style
}
