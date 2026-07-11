package togglebutton

import (
	"image/color"

	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/FlowUI/internal/theme"
)

type toggleButtonStyle struct {
	background  color.NRGBA
	foreground  color.NRGBA
	radius      unit.Dp
	focusColor  color.NRGBA
	focusWidth  unit.Dp
	focusOffset unit.Dp
	focus       float32
	opacity     float32
}

type toggleButtonSizeStyle struct {
	height       unit.Dp
	paddingX     unit.Dp
	textSize     float32
	pressedScale float32
}

func toggleButtonStyleFor(activeTheme *theme.Theme, currentForeground color.NRGBA, variant ToggleButtonVariant, selected, hovered, pressed, disabled bool) toggleButtonStyle {
	background := activeTheme.Palette.SurfaceRaised
	foreground := currentForeground
	if variant == ToggleButtonGhost {
		background = color.NRGBA{}
		foreground = activeTheme.Palette.Foreground
	}
	if hovered || pressed {
		background = activeTheme.Palette.SurfacePressed
	}
	if selected {
		background = activeTheme.Palette.AccentSoft
		foreground = activeTheme.Palette.AccentSoftForeground
		if hovered || pressed {
			background = activeTheme.Palette.AccentSoftHover
		}
	}
	opacity := float32(1)
	if disabled {
		opacity = activeTheme.DisabledOpacityValue()
	}
	return toggleButtonStyle{
		background:  background,
		foreground:  foreground,
		radius:      activeTheme.Components.ToggleButton.Radius,
		focusColor:  activeTheme.Palette.Focus,
		focusWidth:  activeTheme.Components.ToggleButton.FocusRingWidth,
		focusOffset: activeTheme.Components.ToggleButton.FocusRingOffset,
		opacity:     opacity,
	}
}

func toggleButtonSizeStyleFor(activeTheme *theme.Theme, size ToggleButtonSize) toggleButtonSizeStyle {
	tokens := activeTheme.Components.ToggleButton
	style := toggleButtonSizeStyle{
		height:       tokens.MediumHeight,
		paddingX:     tokens.MediumPaddingX,
		textSize:     float32(tokens.MediumTextSize),
		pressedScale: tokens.PressedScaleMedium,
	}
	switch size {
	case ToggleButtonSmall:
		style.height = tokens.SmallHeight
		style.paddingX = tokens.SmallPaddingX
		style.textSize = float32(tokens.SmallTextSize)
		style.pressedScale = tokens.PressedScaleSmall
	case ToggleButtonLarge:
		style.height = tokens.LargeHeight
		style.paddingX = tokens.LargePaddingX
		style.textSize = float32(tokens.LargeTextSize)
		style.pressedScale = tokens.PressedScaleLarge
	}
	return style
}

func (s toggleButtonSizeStyle) inset(iconOnly bool) layout.Inset {
	if iconOnly {
		return layout.Inset{}
	}
	return layout.Inset{Left: s.paddingX, Right: s.paddingX}
}

func resolvedToggleButtonPressedScale(scale float32, size ToggleButtonSize) float32 {
	if scale > 0 && scale <= 1 {
		return scale
	}
	switch size {
	case ToggleButtonSmall:
		return 0.98
	case ToggleButtonLarge:
		return 0.96
	default:
		return 0.97
	}
}
