// Package colorpickerstyle owns the visual recipe for the color picker panel.
package colorpickerstyle

import (
	"image"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/widget/material"
	"github.com/qianniancn/flowui/draw"
	"github.com/qianniancn/flowui/theme"
)

func BuildColorPicker(active *theme.Theme, request theme.ColorPickerRequest) theme.ColorPickerVisual {
	if active == nil {
		panic(theme.ErrMissingVisualProvider)
	}
	tokens := active.Components.ColorPicker
	wheel := active.Components.ColorWheel
	area := active.Components.ColorArea
	slider := active.Components.ColorSlider
	field := active.Components.ColorField
	swatch := active.Components.ColorSwatch
	picker := active.Components.ColorSwatchPicker
	sem := active.Semantic()
	return theme.ColorPickerVisual{
		Override: theme.ColorPickerOverrideStyle(request),
		Metrics: theme.ColorPickerMetrics{
			TriggerGap: tokens.TriggerGap, PanelGap: tokens.PanelGap,
			PanelWidth: tokens.PanelWidth, PanelPadding: tokens.PanelPadding,
			PanelRadius: tokens.PanelRadius, PanelMaxHeight: tokens.PanelMaxHeight,
			ContentGap:              tokens.ContentGap,
			CompactContentGap:       tokens.CompactContentGap,
			ChannelGap:              8,
			EnterDuration:           tokens.EnterDuration,
			ExitDuration:            tokens.ExitDuration,
			EnterScale:              tokens.EnterScale,
			SwatchSelectionDuration: picker.SelectionDuration,
			SwatchCheckDuration:     picker.CheckDuration,
			SwatchHoverDuration:     picker.HoverDuration,
		},
		Wheel:           theme.ColorWheelVisual{Size: wheel.Size, ThumbSize: wheel.ThumbSize, ThumbBorderWidth: wheel.ThumbBorderWidth, FocusRingWidth: wheel.FocusRingWidth, Render: wheelRenderer(wheel, sem.Ring)},
		Area:            theme.ColorAreaVisual{Size: area.Size, Radius: area.Radius, ThumbSize: area.ThumbSize, DraggingThumbSize: area.DraggingThumbSize, ThumbBorderWidth: area.ThumbBorderWidth, FocusRingWidth: area.FocusRingWidth, DotSize: area.DotSize, DotGap: area.DotGap, Render: areaRenderer(area, sem.Ring)},
		Slider:          theme.ColorSliderVisual{TextSize: slider.TextSize, HeaderGap: slider.HeaderGap, TrackHeight: slider.TrackHeight, ThumbSize: slider.ThumbSize, ThumbBorderWidth: slider.ThumbBorderWidth, FocusRingWidth: slider.FocusRingWidth, Render: sliderRenderer(slider, sem.Ring)},
		Field:           theme.ColorFieldVisual{Gap: field.Gap, ErrorTextSize: active.Components.Description.TextSize, ErrorForeground: sem.Tone(theme.ToneDanger).Fill},
		Swatch:          theme.ColorSwatchVisual{ExtraSmallSize: swatch.ExtraSmallSize, SmallSize: swatch.SmallSize, MediumSize: swatch.MediumSize, LargeSize: swatch.LargeSize, ExtraLargeSize: swatch.ExtraLargeSize, SquareRadius: swatch.SquareRadius, Render: swatchRenderer(swatch)},
		SwatchPicker:    theme.ColorSwatchPickerVisual{Gap: picker.Gap, ExtraSmallBorderWidth: picker.ExtraSmallBorderWidth, BorderWidth: picker.BorderWidth, LargeBorderWidth: picker.LargeBorderWidth, FocusRingWidth: picker.FocusRingWidth, FocusRingGap: picker.FocusRingGap, CheckStroke: picker.CheckStroke, SquareItemRadiusExtraSmall: picker.SquareItemRadiusExtraSmall, SquareItemRadiusSmall: picker.SquareItemRadiusSmall, SquareItemRadius: picker.SquareItemRadius, SquareSwatchRadiusExtraSmall: picker.SquareSwatchRadiusExtraSmall, SquareSwatchRadiusSmall: picker.SquareSwatchRadiusSmall, SquareSwatchRadius: picker.SquareSwatchRadius, ShadowOpacity: picker.ShadowOpacity, SelectedScale: picker.SelectedScale, Render: swatchPickerRenderer(active, picker)},
		PanelForeground: sem.OverlayForeground, PanelBackground: sem.Overlay,
		Foreground: sem.Foreground, MutedForeground: sem.MutedForeground, FocusColor: sem.Ring,
		DisabledOpacity: active.DisabledOpacityValue(), Motion: active.Motion,
		TriggerLabel: func(gtx layout.Context, value string) layout.Dimensions {
			label := material.Label(theme.MaterialOf(active), tokens.TriggerTextSize, value)
			label.Color = sem.Foreground
			return label.Layout(gtx)
		},
		TriggerFocus: func(gtx layout.Context, size image.Point, opacity float32) {
			if opacity <= 0 {
				return
			}
			width := max(gtx.Dp(tokens.FocusRingWidth), 1)
			rect := image.Rectangle{Max: size}.Inset(width)
			if rect.Empty() {
				return
			}
			col := sem.Ring
			col.A = byte(float32(col.A)*opacity + .5)
			radius := max(gtx.Dp(tokens.TriggerRadius)-width, 0)
			stroke := clip.Stroke{Path: clip.UniformRRect(rect, radius).Path(gtx.Ops), Width: float32(width)}.Op().Push(gtx.Ops)
			paint.Fill(gtx.Ops, col)
			stroke.Pop()
		},
		Panel: func(gtx layout.Context, size image.Point) {
			radius := min(max(gtx.Dp(tokens.PanelRadius), 1), min(size.X, size.Y)/2)
			draw.Surface(gtx, image.Rectangle{Max: size}, radius, sem.Overlay,
				draw.ThemeShadow(active.Shadows.Overlay, sem.OverlayShadow, 1))
		},
	}
}
