package testtheme

import (
	"image"

	"gioui.org/layout"
	"github.com/qianniancn/flowui/theme"
)

func (behaviorProvider) Dropdown(active *theme.Theme, request theme.DropdownRequest) theme.DropdownVisual {
	tokens := active.Components.Dropdown
	return theme.DropdownVisual{
		TriggerPressedScale: tokens.TriggerPressedScale,
		PressDuration:       tokens.PressDuration,
		EnterDuration:       tokens.EnterDuration,
		ExitDuration:        tokens.ExitDuration,
		LongPressDuration:   tokens.LongPressDuration,
		HoverOpenDuration:   tokens.HoverOpenDuration,
		HoverCloseDuration:  tokens.HoverCloseDuration,
		PanelGap:            tokens.PanelGap,
		ArrowSize:           tokens.ArrowSize,
		Motion:              active.Motion,
		TriggerFocus:        func(layout.Context, theme.DropdownTriggerFocusRequest) {},
		TriggerGlyph: func(gtx layout.Context, _ theme.DropdownTriggerGlyphRequest) layout.Dimensions {
			return layout.Dimensions{Size: image.Pt(gtx.Dp(16), gtx.Dp(16))}
		},
		Override: theme.DropdownOverrideStyle(request),
	}
}
