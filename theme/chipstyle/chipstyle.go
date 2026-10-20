// Package chipstyle contains the theme-owned visual recipe for chip behavior.
// It maps semantic color, variant, and size axes to declarations; widget/chip
// remains responsible for label/accessory layout and semantics.
package chipstyle

import (
	"image/color"

	"gioui.org/font"
	"github.com/qianniancn/flowui/appearance"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

// BuildChip builds the complete provider-owned chip visual plan.
func BuildChip(active *theme.Theme, request theme.ChipRequest) theme.ChipVisual {
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
	background, foreground := chipColors(active, request.Color, request.Variant)
	variant := style.Style{}.TextColor(style.Solid(foreground))
	if request.Variant != theme.ChipVariantTertiary {
		variant = variant.Background(style.Solid(background))
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
		Metrics: theme.ChipMetrics{
			LabelPaddingX: tokens.LabelPaddingX,
			ContentGap:    tokens.ContentGap,
		},
	}
}

func chipColors(active *theme.Theme, chipColor theme.ChipColor, variant theme.ChipVariant) (color.NRGBA, color.NRGBA) {
	sem := active.Semantic()
	defaultBackground := sem.SurfaceTertiary
	defaultForeground := sem.SurfaceTertiaryForeground
	tone, tinted := chipTone(active, chipColor)

	solidBackground, solidForeground := defaultBackground, defaultForeground
	softBackground, softForeground := defaultSoftColor(defaultBackground), defaultForeground
	if tinted {
		solidBackground, solidForeground = tone.Fill, tone.OnFill
		softBackground, softForeground = tone.Soft, tone.OnSoft
	}

	switch variant {
	case theme.ChipVariantPrimary:
		return solidBackground, solidForeground
	case theme.ChipVariantTertiary:
		return color.NRGBA{}, softForeground
	case theme.ChipVariantSoft:
		return softBackground, softForeground
	default:
		return defaultBackground, softForeground
	}
}

func chipTone(active *theme.Theme, chipColor theme.ChipColor) (appearance.ToneColors, bool) {
	switch chipColor {
	case theme.ChipColorAccent:
		return appearance.Tone(active, theme.TonePrimary), true
	case theme.ChipColorSuccess:
		return appearance.Tone(active, theme.ToneSuccess), true
	case theme.ChipColorWarning:
		return appearance.Tone(active, theme.ToneWarning), true
	case theme.ChipColorDanger:
		return appearance.Tone(active, theme.ToneDanger), true
	default:
		return appearance.Tone(active, theme.ToneDefault), false
	}
}

func defaultSoftColor(value color.NRGBA) color.NRGBA {
	value.A = byte((uint16(value.A) + 1) / 2)
	return value
}
