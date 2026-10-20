package testtheme

import (
	"image"
	"time"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/widget/material"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

func (behaviorProvider) Combo(active *theme.Theme, request theme.ComboRequest) theme.ComboVisual {
	content := style.Style{}.
		Background(style.Solid(background)).
		TextColor(style.Solid(foreground)).
		Radius(4).
		MinHeight(40).
		PaddingX(8).
		PaddingY(8).
		FontSize(14)
	defaults := style.Style{}.
		Part(theme.ComboPartTrigger, content).
		Part(theme.ComboPartPlaceholder, style.Style{}.
			TextColor(style.Solid(muted)).
			FontSize(14)).
		Part(theme.ComboPartSelection, style.Style{}.
			Background(style.Solid(accent))).
		Part(theme.ComboPartIndicator, style.Style{}.
			Width(16).
			Height(16).
			TextColor(style.Solid(muted))).
		Part(theme.ComboPartIcon, style.Style{}.
			Width(16).
			Height(16).
			TextColor(style.Solid(muted))).
		Part(theme.ComboPartPanel, style.Style{}.
			Background(style.Solid(background)).
			TextColor(style.Solid(foreground)).
			Radius(4).
			Padding(8)).
		Part(theme.ComboPartDescription, style.Style{}.
			TextColor(style.Solid(muted)).
			FontSize(12)).
		Part(theme.ComboPartItem, style.Style{}.
			MinHeight(40).
			PaddingX(8).
			PaddingY(8).
			Radius(4).
			Background(style.RGBA(0x00000000)).
			TextColor(style.Solid(foreground)).
			FontSize(14).
			FontWeight(int(font.Medium)).
			On(style.Hovered, style.Style{}.Background(style.Solid(background))).
			On(style.Disabled, style.Style{}.
				TextColor(style.Solid(muted))))
	return theme.ComboVisual{
		Defaults: defaults,
		Override: theme.ComboOverrideStyle(request),
		Metrics: theme.ComboMetrics{
			PanelGap:           4,
			PanelMaxHeight:     240,
			ItemHeight:         40,
			ItemTextSize:       14,
			AnimationDuration:  150 * time.Millisecond,
			ItemSelectDuration: 100 * time.Millisecond,
			IconDuration:       150 * time.Millisecond,
			AnimationScale:     .96,
		},
		Motion:  active.Motion,
		Chevron: func(layout.Context, image.Point, float32, style.ResolvedStyle) {},
		Check:   func(layout.Context, image.Point, float32, style.ResolvedStyle) {},
		Text:    behaviorComboText(active),
		EditorStyle: func(style.ResolvedStyle, style.ResolvedStyle, style.ResolvedStyle) theme.ComboEditorStyle {
			return theme.ComboEditorStyle{TextSize: 14, Foreground: foreground, Placeholder: muted, Selection: accent}
		},
		IndicatorWidth: behaviorPartWidth,
		IconSize:       behaviorPartSize,
	}
}

func behaviorComboText(active *theme.Theme) theme.ComboTextRenderer {
	return func(gtx layout.Context, request theme.ComboTextRequest) layout.Dimensions {
		size := request.Size
		if size <= 0 {
			size = 14
		}
		label := material.Label(theme.MaterialOf(active), size, request.Text)
		label.Color = foreground
		return label.Layout(gtx)
	}
}
