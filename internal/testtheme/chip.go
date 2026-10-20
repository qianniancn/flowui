package testtheme

import (
	"gioui.org/font"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

func (behaviorProvider) Chip(active *theme.Theme, request theme.ChipRequest) theme.ChipVisual {
	if active == nil {
		panic(theme.ErrMissingVisualProvider)
	}
	tokens := active.Components.Chip
	height, paddingX, paddingY, textSize := tokens.MediumHeight, tokens.MediumPaddingX, tokens.MediumPaddingY, tokens.MediumTextSize
	switch request.Size {
	case theme.ChipSizeSmall:
		height, paddingX, paddingY, textSize = tokens.SmallHeight, tokens.SmallPaddingX, tokens.SmallPaddingY, tokens.SmallTextSize
	case theme.ChipSizeLarge:
		height, paddingX, paddingY, textSize = tokens.LargeHeight, tokens.LargePaddingX, tokens.LargePaddingY, tokens.LargeTextSize
	}
	sem := active.Semantic()
	defaultBackground := sem.SurfaceTertiary
	defaultForeground := sem.SurfaceTertiaryForeground
	tone, tinted := sem.Tone(theme.ToneDefault), false
	switch request.Color {
	case theme.ChipColorAccent:
		tone, tinted = sem.Tone(theme.TonePrimary), true
	case theme.ChipColorSuccess:
		tone, tinted = sem.Tone(theme.ToneSuccess), true
	case theme.ChipColorWarning:
		tone, tinted = sem.Tone(theme.ToneWarning), true
	case theme.ChipColorDanger:
		tone, tinted = sem.Tone(theme.ToneDanger), true
	}
	solidBackground, solidForeground := defaultBackground, defaultForeground
	softBackground, softForeground := defaultBackground, defaultForeground
	softBackground.A = byte((uint16(softBackground.A) + 1) / 2)
	if tinted {
		solidBackground, solidForeground = tone.Fill, tone.OnFill
		softBackground, softForeground = tone.Soft, tone.OnSoft
	}
	variant := style.Style{}.TextColor(style.Solid(softForeground))
	switch request.Variant {
	case theme.ChipVariantPrimary:
		variant = variant.Background(style.Solid(solidBackground)).TextColor(style.Solid(solidForeground))
	case theme.ChipVariantSoft:
		variant = variant.Background(style.Solid(softBackground))
	case theme.ChipVariantSecondary:
		variant = variant.Background(style.Solid(defaultBackground))
	}
	return theme.ChipVisual{
		Defaults: style.Style{}.
			Radius(tokens.Radius).
			Opacity(1).
			Part(style.PartLabel, style.Style{}.FontWeight(int(font.Medium)).MaxLines(1)),
		Variant: variant,
		Size: style.Style{}.
			Height(height).
			PaddingX(paddingX).
			PaddingY(paddingY).
			Part(style.PartLabel, style.Style{}.FontSize(textSize).LineHeight(tokens.LineHeight)),
		Override: theme.ChipOverrideStyle(request),
		Metrics:  theme.ChipMetrics{LabelPaddingX: tokens.LabelPaddingX, ContentGap: tokens.ContentGap},
	}
}
