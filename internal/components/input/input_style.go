package input

import (
	"image/color"

	"gioui.org/unit"
	"github.com/qianniancn/FlowUI/internal/render"
	"github.com/qianniancn/FlowUI/internal/theme"
)

const primaryHoverStrength float32 = 0.65

type inputStyle struct {
	Background    color.NRGBA
	Foreground    color.NRGBA
	Placeholder   color.NRGBA
	Selection     color.NRGBA
	Ring          color.NRGBA
	RingWidth     unit.Dp
	ShadowOpacity float32
	Opacity       float32
}

func inputStyleFor(activeTheme *theme.Theme, variant InputVariant, hovered, focused, disabled, invalid bool) inputStyle {
	tokens := activeTheme.Components.Input
	return fieldStyleFor(activeTheme, variant, hovered, focused, disabled, invalid, tokens.FocusRingWidth, tokens.InvalidOutlineWidth, tokens.ShadowOpacity)
}

func textAreaStyleFor(activeTheme *theme.Theme, variant TextAreaVariant, hovered, focused, disabled, invalid bool) inputStyle {
	tokens := activeTheme.Components.TextArea
	return fieldStyleFor(activeTheme, variant, hovered, focused, disabled, invalid, tokens.FocusRingWidth, tokens.InvalidOutlineWidth, tokens.ShadowOpacity)
}

func fieldStyleFor(activeTheme *theme.Theme, variant InputVariant, hovered, focused, disabled, invalid bool, focusRingWidth, invalidOutlineWidth unit.Dp, shadowOpacity float32) inputStyle {
	background := activeTheme.Palette.Surface
	hoverBackground := render.LerpColor(
		activeTheme.Palette.Surface,
		activeTheme.Palette.SurfaceSecondary,
		primaryHoverStrength,
	)
	focusBackground := activeTheme.Palette.Surface
	if variant == InputSecondary {
		background = activeTheme.Palette.SurfacePressed
		hoverBackground = activeTheme.Palette.Border
		focusBackground = activeTheme.Palette.SurfacePressed
		shadowOpacity = 0
	}
	if hovered && !focused {
		background = hoverBackground
	}
	if focused || invalid {
		background = focusBackground
	}

	ring := color.NRGBA{}
	ringWidth := unit.Dp(0)
	if focused {
		ring = activeTheme.Palette.Focus
		ringWidth = focusRingWidth
	}
	if invalid {
		ring = activeTheme.Palette.Danger
		ringWidth = invalidOutlineWidth
		if focused {
			ringWidth = focusRingWidth
		}
	}

	opacity := float32(1)
	if disabled {
		opacity = activeTheme.DisabledOpacityValue()
	}
	return inputStyle{
		Background:    background,
		Foreground:    activeTheme.Palette.Foreground,
		Placeholder:   activeTheme.Palette.MutedForeground,
		Selection:     activeTheme.Palette.Selection,
		Ring:          ring,
		RingWidth:     ringWidth,
		ShadowOpacity: shadowOpacity,
		Opacity:       opacity,
	}
}
