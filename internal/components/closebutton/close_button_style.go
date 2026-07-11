package closebutton

import (
	"image/color"

	"gioui.org/unit"
	"github.com/qianniancn/FlowUI/internal/theme"
)

type closeButtonStyle struct {
	background   color.NRGBA
	foreground   color.NRGBA
	focusColor   color.NRGBA
	radius       unit.Dp
	padding      unit.Dp
	iconSize     unit.Dp
	focusWidth   unit.Dp
	pressedScale float32
	focusOpacity float32
}

func closeButtonStyleFor(activeTheme *theme.Theme, hovered, disabled bool) closeButtonStyle {
	component := activeTheme.Components.CloseButton
	background := activeTheme.Palette.SurfaceRaised
	if hovered && !disabled {
		background = activeTheme.Palette.SurfacePressed
	}
	foreground := activeTheme.Palette.MutedForeground
	if disabled {
		background = activeTheme.DisabledColor(background)
		foreground = activeTheme.DisabledColor(foreground)
	}
	return closeButtonStyle{
		background:   background,
		foreground:   foreground,
		focusColor:   activeTheme.Palette.Focus,
		radius:       component.Radius,
		padding:      component.Padding,
		iconSize:     component.IconSize,
		focusWidth:   component.FocusRingWidth,
		pressedScale: component.PressedScale,
	}
}
