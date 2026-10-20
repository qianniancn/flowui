// Package badgestyle contains the theme-owned visual recipe for badge
// behavior. It maps semantic color, variant, and size axes to declarations;
// widget/badge remains responsible for anchor and placement behavior.
package badgestyle

import (
	"image/color"

	"gioui.org/font"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/appearance"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

// BuildBadge builds the complete provider-owned badge visual plan.
func BuildBadge(active *theme.Theme, request theme.BadgeRequest) theme.BadgeVisual {
	if active == nil {
		panic(theme.ErrMissingVisualProvider)
	}
	tokens := active.Components.Badge
	background, foreground := badgeColors(active, request.Color, request.Variant)
	minSize, radius, textSize, lineHeight := badgeSize(tokens, request.Size)
	border := active.Semantic().Surface
	return theme.BadgeVisual{
		Defaults: style.Style{}.
			BorderColor(style.Solid(border)).
			BorderWidth(tokens.BorderWidth).
			Opacity(1).
			Part(style.PartLabel, style.Style{}.
				FontWeight(int(font.Medium)).
				MaxLines(1)),
		Variant: style.Style{}.
			Background(style.Solid(background)).
			TextColor(style.Solid(foreground)),
		Size: style.Style{}.
			MinWidth(minSize).
			MinHeight(minSize).
			PaddingX(tokens.LabelPaddingX).
			Radius(radius).
			Part(style.PartLabel, style.Style{}.
				FontSize(textSize).
				LineHeight(lineHeight)),
		Override: theme.BadgeOverrideStyle(request),
		Metrics:  theme.BadgeMetrics{PlacementOffsetRatio: tokens.PlacementOffsetRatio},
	}
}

func badgeSize(tokens theme.BadgeTheme, size theme.BadgeSize) (minSize, radius unit.Dp, textSize, lineHeight unit.Sp) {
	minSize, radius, textSize, lineHeight = tokens.MediumMinSize, tokens.MediumRadius, tokens.MediumTextSize, tokens.MediumLineHeight
	switch size {
	case theme.BadgeSizeSmall:
		minSize, radius, textSize, lineHeight = tokens.SmallMinSize, tokens.SmallRadius, tokens.SmallTextSize, tokens.SmallLineHeight
	case theme.BadgeSizeLarge:
		minSize, radius, textSize, lineHeight = tokens.LargeMinSize, tokens.LargeRadius, tokens.LargeTextSize, tokens.LargeLineHeight
	}
	return
}

func badgeColors(active *theme.Theme, badgeColor theme.BadgeColor, variant theme.BadgeVariant) (color.NRGBA, color.NRGBA) {
	neutral := appearance.Tone(active, theme.ToneDefault)
	tone, tinted := badgeTone(active, badgeColor)
	switch variant {
	case theme.BadgeVariantSecondary:
		return neutral.Fill, tone.OnSoft
	case theme.BadgeVariantSoft:
		if tinted {
			return tone.Soft, tone.OnSoft
		}
		background := neutral.Fill
		background.A = byte((uint16(background.A) + 1) / 2)
		return background, neutral.OnFill
	default:
		return tone.Fill, tone.OnFill
	}
}

func badgeTone(active *theme.Theme, badgeColor theme.BadgeColor) (appearance.ToneColors, bool) {
	switch badgeColor {
	case theme.BadgeColorAccent:
		return appearance.Tone(active, theme.TonePrimary), true
	case theme.BadgeColorSuccess:
		return appearance.Tone(active, theme.ToneSuccess), true
	case theme.BadgeColorWarning:
		return appearance.Tone(active, theme.ToneWarning), true
	case theme.BadgeColorDanger:
		return appearance.Tone(active, theme.ToneDanger), true
	default:
		return appearance.Tone(active, theme.ToneDefault), false
	}
}
