package input

import (
	"image/color"

	"gioui.org/unit"
	"github.com/qianniancn/FlowUI/internal/render"
	"github.com/qianniancn/FlowUI/internal/theme"
)

type inputGroupStyle struct {
	Background    color.NRGBA
	Foreground    color.NRGBA
	Placeholder   color.NRGBA
	Selection     color.NRGBA
	Ring          color.NRGBA
	Divider       color.NRGBA
	RingWidth     unit.Dp
	ShadowOpacity float32
	Opacity       float32
}

func inputGroupStyleFor(activeTheme *theme.Theme, variant InputVariant, hovered, focused, disabled, invalid bool) inputGroupStyle {
	background := activeTheme.Palette.Surface
	hoverBackground := render.LerpColor(
		activeTheme.Palette.Surface,
		activeTheme.Palette.SurfaceSecondary,
		primaryHoverStrength,
	)
	focusBackground := activeTheme.Palette.Surface
	shadowOpacity := activeTheme.Components.InputGroup.ShadowOpacity
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
		ringWidth = activeTheme.Components.InputGroup.FocusRingWidth
	}
	if invalid {
		ring = activeTheme.Palette.Danger
		ringWidth = activeTheme.Components.InputGroup.InvalidOutlineWidth
		if focused {
			ringWidth = activeTheme.Components.InputGroup.FocusRingWidth
		}
	}

	opacity := float32(1)
	if disabled {
		opacity = activeTheme.DisabledOpacityValue()
	}
	return inputGroupStyle{
		Background:    background,
		Foreground:    activeTheme.Palette.Foreground,
		Placeholder:   activeTheme.Palette.MutedForeground,
		Selection:     activeTheme.Palette.Selection,
		Ring:          ring,
		Divider:       activeTheme.Palette.Border,
		RingWidth:     ringWidth,
		ShadowOpacity: shadowOpacity,
		Opacity:       opacity,
	}
}
