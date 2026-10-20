// Package avatarstyle contains the theme-owned visual recipe for avatar
// behavior. The widget owns image/fallback content and accessibility; this
// package owns geometry, typography and semantic colors.
package avatarstyle

import (
	"image/color"

	"gioui.org/font"
	"gioui.org/layout"
	lucide "github.com/qianniancn/flowui-icons-lucide"
	"github.com/qianniancn/flowui/appearance"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/icon"
)

func avatarIconRenderer(gtx layout.Context, col color.NRGBA) layout.Dimensions {
	return icon.Layout(lucide.UserRound, gtx, col)
}

// BuildAvatar builds the complete provider-owned avatar visual plan.
func BuildAvatar(active *theme.Theme, request theme.AvatarRequest) theme.AvatarVisual {
	if active == nil {
		panic(theme.ErrMissingVisualProvider)
	}
	tokens := active.Components.Avatar
	diameter, radius, textSize, iconSize := tokens.MediumSize, tokens.MediumRadius, tokens.MediumTextSize, tokens.MediumIconSize
	switch request.Size {
	case theme.AvatarSizeSmall:
		diameter, radius, textSize, iconSize = tokens.SmallSize, tokens.SmallRadius, tokens.SmallTextSize, tokens.SmallIconSize
	case theme.AvatarSizeLarge:
		diameter, radius, textSize, iconSize = tokens.LargeSize, tokens.LargeRadius, tokens.LargeTextSize, tokens.LargeIconSize
	}
	background, foreground := avatarColors(active, request.Color, request.Variant)
	return theme.AvatarVisual{
		Defaults: style.Style{}.
			Overflow(style.OverflowHidden).
			Opacity(1).
			Part(style.PartLabel, style.Style{}.FontWeight(int(font.Medium)).MaxLines(1)),
		Variant: style.Style{}.
			Background(style.Solid(background)).
			TextColor(style.Solid(foreground)),
		Size: style.Style{}.
			Width(diameter).
			AspectRatio(1).
			Radius(radius).
			Part(style.PartLabel, style.Style{}.FontSize(textSize)).
			Part(style.PartIcon, style.Style{}.Width(iconSize).Height(iconSize)),
		Override:     theme.AvatarOverrideStyle(request),
		IconRenderer: avatarIconRenderer,
	}
}

func avatarColors(active *theme.Theme, avatarColor theme.AvatarColor, variant theme.AvatarVariant) (color.NRGBA, color.NRGBA) {
	neutral := appearance.Tone(active, theme.ToneDefault)
	tone, tinted := avatarTone(active, avatarColor)
	foreground := neutral.OnFill
	if tinted {
		foreground = tone.OnSoft
	}
	background := neutral.Fill
	if variant != theme.AvatarVariantSoft {
		return background, foreground
	}
	if tinted {
		return tone.Soft, foreground
	}
	background.A = byte((uint16(background.A) + 1) / 2)
	return background, foreground
}

func avatarTone(active *theme.Theme, avatarColor theme.AvatarColor) (appearance.ToneColors, bool) {
	switch avatarColor {
	case theme.AvatarColorAccent:
		return appearance.Tone(active, theme.TonePrimary), true
	case theme.AvatarColorSuccess:
		return appearance.Tone(active, theme.ToneSuccess), true
	case theme.AvatarColorWarning:
		return appearance.Tone(active, theme.ToneWarning), true
	case theme.AvatarColorDanger:
		return appearance.Tone(active, theme.ToneDanger), true
	default:
		return appearance.Tone(active, theme.ToneDefault), false
	}
}
