package testtheme

import (
	"image"
	"image/color"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"github.com/qianniancn/flowui/theme"
)

func (behaviorProvider) ColorPicker(active *theme.Theme, request theme.ColorPickerRequest) theme.ColorPickerVisual {
	visual := colorPickerBehaviorVisual(active)
	visual.Override = theme.ColorPickerOverrideStyle(request)
	return visual
}

func colorPickerBehaviorVisual(active *theme.Theme) theme.ColorPickerVisual {
	base := active.Components
	return theme.ColorPickerVisual{
		Metrics: theme.ColorPickerMetrics{
			TriggerGap: 8, PanelGap: 4, PanelWidth: 248,
			PanelPadding: 8, PanelRadius: 4, PanelMaxHeight: 520,
			ContentGap:              8,
			CompactContentGap:       6,
			ChannelGap:              8,
			EnterDuration:           base.ColorPicker.EnterDuration,
			ExitDuration:            base.ColorPicker.ExitDuration,
			EnterScale:              base.ColorPicker.EnterScale,
			SwatchSelectionDuration: base.ColorSwatchPicker.SelectionDuration,
			SwatchCheckDuration:     base.ColorSwatchPicker.CheckDuration,
			SwatchHoverDuration:     base.ColorSwatchPicker.HoverDuration,
		},
		Wheel:           theme.ColorWheelVisual{Size: base.ColorWheel.Size, ThumbSize: base.ColorWheel.ThumbSize, ThumbBorderWidth: base.ColorWheel.ThumbBorderWidth, FocusRingWidth: base.ColorWheel.FocusRingWidth, Render: func(layout.Context, theme.ColorWheelRenderRequest) {}},
		Area:            theme.ColorAreaVisual{Size: base.ColorArea.Size, Radius: base.ColorArea.Radius, ThumbSize: base.ColorArea.ThumbSize, DraggingThumbSize: base.ColorArea.DraggingThumbSize, ThumbBorderWidth: base.ColorArea.ThumbBorderWidth, FocusRingWidth: base.ColorArea.FocusRingWidth, DotSize: base.ColorArea.DotSize, DotGap: base.ColorArea.DotGap, Render: func(layout.Context, theme.ColorAreaRenderRequest) {}},
		Slider:          theme.ColorSliderVisual{TextSize: base.ColorSlider.TextSize, HeaderGap: base.ColorSlider.HeaderGap, TrackHeight: base.ColorSlider.TrackHeight, ThumbSize: base.ColorSlider.ThumbSize, ThumbBorderWidth: base.ColorSlider.ThumbBorderWidth, FocusRingWidth: base.ColorSlider.FocusRingWidth, Render: func(layout.Context, theme.ColorSliderRenderRequest) {}},
		Field:           theme.ColorFieldVisual{Gap: base.ColorField.Gap, ErrorTextSize: base.Description.TextSize, ErrorForeground: color.NRGBA{R: 200, A: 255}},
		Swatch:          theme.ColorSwatchVisual{ExtraSmallSize: base.ColorSwatch.ExtraSmallSize, SmallSize: base.ColorSwatch.SmallSize, MediumSize: base.ColorSwatch.MediumSize, LargeSize: base.ColorSwatch.LargeSize, ExtraLargeSize: base.ColorSwatch.ExtraLargeSize, SquareRadius: base.ColorSwatch.SquareRadius, Render: func(layout.Context, theme.ColorSwatchRenderRequest) {}},
		SwatchPicker:    theme.ColorSwatchPickerVisual{Gap: base.ColorSwatchPicker.Gap, ExtraSmallBorderWidth: base.ColorSwatchPicker.ExtraSmallBorderWidth, BorderWidth: base.ColorSwatchPicker.BorderWidth, LargeBorderWidth: base.ColorSwatchPicker.LargeBorderWidth, FocusRingWidth: base.ColorSwatchPicker.FocusRingWidth, FocusRingGap: base.ColorSwatchPicker.FocusRingGap, CheckStroke: base.ColorSwatchPicker.CheckStroke, SquareItemRadiusExtraSmall: base.ColorSwatchPicker.SquareItemRadiusExtraSmall, SquareItemRadiusSmall: base.ColorSwatchPicker.SquareItemRadiusSmall, SquareItemRadius: base.ColorSwatchPicker.SquareItemRadius, SquareSwatchRadiusExtraSmall: base.ColorSwatchPicker.SquareSwatchRadiusExtraSmall, SquareSwatchRadius: base.ColorSwatchPicker.SquareSwatchRadius, SquareSelectedSmallRadius: base.ColorSwatchPicker.SquareSelectedSmallRadius, ShadowOpacity: base.ColorSwatchPicker.ShadowOpacity, SelectedScale: base.ColorSwatchPicker.SelectedScale, Render: func(layout.Context, theme.ColorSwatchPickerItemRequest) {}},
		PanelForeground: color.NRGBA{A: 255},
		PanelBackground: color.NRGBA{R: 255, G: 255, B: 255, A: 255},
		Foreground:      color.NRGBA{A: 255}, MutedForeground: color.NRGBA{A: 255}, FocusColor: color.NRGBA{B: 255, A: 255},
		DisabledOpacity: 0.5, Motion: active.Motion,
		TriggerLabel: func(gtx layout.Context, value string) layout.Dimensions {
			label := material.Label(theme.MaterialOf(active), unit.Sp(14), value)
			return label.Layout(gtx)
		},
		TriggerFocus: func(gtx layout.Context, size image.Point, opacity float32) {
			if opacity == 0 {
				return
			}
			paint.FillShape(gtx.Ops, color.NRGBA{A: byte(255 * opacity)}, clip.Rect{Max: size}.Op())
		},
		Panel: func(gtx layout.Context, size image.Point) {
			paint.FillShape(gtx.Ops, color.NRGBA{R: 255, G: 255, B: 255, A: 255}, clip.Rect{Max: size}.Op())
		},
	}
}
