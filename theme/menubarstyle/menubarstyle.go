// Package menubarstyle contains the theme-owned visual recipe for menubar
// behavior. It deliberately does not import widget/menubar.
package menubarstyle

import (
	"image"
	"image/color"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"github.com/qianniancn/flowui/draw"
	"github.com/qianniancn/flowui/theme"
)

// BuildMenubar returns the complete menubar visual protocol for the active
// theme.
func BuildMenubar(active *theme.Theme, request theme.MenubarRequest) theme.MenubarVisual {
	if active == nil {
		panic(theme.ErrMissingVisualProvider)
	}
	tokens := active.Components.Menubar
	if request.Compact {
		tokens = compact(tokens)
	}
	sem := active.Semantic()
	return theme.MenubarVisual{
		Metrics:     metrics(tokens),
		Motion:      active.Motion,
		Override:    theme.MenubarOverrideStyle(request),
		RenderLabel: labelRenderer(active, tokens.TriggerTextSize),
		Trigger: func(state theme.MenubarTriggerRequest) theme.MenubarTriggerVisual {
			background := color.NRGBA{}
			if state.Open || state.Hovered {
				background = sem.Tone(theme.ToneDefault).Fill
			}
			if state.Pressed {
				background = sem.Tone(theme.ToneDefault).Hover
			}
			opacity := float32(1)
			if state.Disabled {
				opacity = active.DisabledOpacityValue()
			}
			return theme.MenubarTriggerVisual{
				Background: background,
				Foreground: sem.Foreground,
				Focus:      sem.Ring,
				Opacity:    opacity,
			}
		},
		RenderTrigger: renderTrigger,
	}
}

func labelRenderer(active *theme.Theme, size unit.Sp) theme.MenubarLabelRenderer {
	return func(gtx layout.Context, request theme.MenubarLabelRequest) layout.Dimensions {
		label := material.Label(theme.MaterialOf(active), size, request.Text)
		label.Color = request.Foreground
		label.SelectionColor = active.Semantic().Selection
		label.Font.Weight = font.Normal
		return label.Layout(gtx)
	}
}

func compact(tokens theme.MenubarTheme) theme.MenubarTheme {
	tokens.TriggerHeight = minDp(tokens.TriggerHeight, 28)
	tokens.TriggerPaddingX = minDp(tokens.TriggerPaddingX, 8)
	tokens.TriggerRadius = minDp(tokens.TriggerRadius, 4)
	tokens.TriggerTextSize = minSp(tokens.TriggerTextSize, 13)
	return tokens
}

func metrics(tokens theme.MenubarTheme) theme.MenubarMetrics {
	return theme.MenubarMetrics{
		TriggerHeight:          tokens.TriggerHeight,
		TriggerPaddingX:        tokens.TriggerPaddingX,
		TriggerRadius:          tokens.TriggerRadius,
		TriggerFocusRingWidth:  tokens.TriggerFocusRingWidth,
		TriggerFocusRingOffset: tokens.TriggerFocusRingOffset,
		Gap:                    tokens.Gap,
		PanelGap:               tokens.PanelGap,
	}
}

func renderTrigger(gtx layout.Context, size image.Point, metrics theme.MenubarMetrics, visual theme.MenubarTriggerVisual, focusOpacity float32) {
	if size.X <= 0 || size.Y <= 0 {
		return
	}
	radius := min(max(gtx.Dp(metrics.TriggerRadius), 0), min(size.X, size.Y)/2)
	rect := image.Rectangle{Max: size}
	if visual.Background.A != 0 {
		paint.FillShape(gtx.Ops, visual.Background, clip.UniformRRect(rect, radius).Op(gtx.Ops))
	}
	if focusOpacity <= 0 || visual.Focus.A == 0 {
		return
	}
	width := max(gtx.Dp(metrics.TriggerFocusRingWidth), 1)
	inset := max(gtx.Dp(metrics.TriggerFocusRingOffset), 0) + (width+1)/2
	focusRect := rect.Inset(inset)
	if focusRect.Empty() {
		return
	}
	focusRadius := max(radius-inset, 0)
	focus := visual.Focus
	focus.A = byte(float32(focus.A)*focusOpacity + 0.5)
	draw.RoundedInsetStroke(gtx, rect, focusRadius, width, inset, focus)
}

func minDp(value, limit unit.Dp) unit.Dp {
	if value < limit {
		return value
	}
	return limit
}

func minSp(value, limit unit.Sp) unit.Sp {
	if value < limit {
		return value
	}
	return limit
}
