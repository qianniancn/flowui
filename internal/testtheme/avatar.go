package testtheme

import (
	"image/color"
	"sync"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/widget"
	lucide "github.com/qianniancn/flowui-icons-lucide"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

var neutralAvatarIcon struct {
	sync.Once
	value *widget.Icon
}

func neutralAvatarIconRenderer(gtx layout.Context, col color.NRGBA) layout.Dimensions {
	neutralAvatarIcon.Do(func() {
		value, err := widget.NewIcon(lucide.UserRound)
		if err == nil {
			neutralAvatarIcon.value = value
		}
	})
	if neutralAvatarIcon.value == nil {
		return layout.Dimensions{}
	}
	return neutralAvatarIcon.value.Layout(gtx, col)
}

func (behaviorProvider) Avatar(active *theme.Theme, request theme.AvatarRequest) theme.AvatarVisual {
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
	sem := active.Semantic()
	background, foreground := sem.Tone(theme.ToneDefault).Fill, sem.Tone(theme.ToneDefault).OnFill
	tone, tinted := sem.Tone(theme.ToneDefault), false
	switch request.Color {
	case theme.AvatarColorAccent:
		tone, tinted = sem.Tone(theme.TonePrimary), true
	case theme.AvatarColorSuccess:
		tone, tinted = sem.Tone(theme.ToneSuccess), true
	case theme.AvatarColorWarning:
		tone, tinted = sem.Tone(theme.ToneWarning), true
	case theme.AvatarColorDanger:
		tone, tinted = sem.Tone(theme.ToneDanger), true
	}
	if tinted {
		foreground = tone.OnSoft
	}
	if request.Variant == theme.AvatarVariantSoft {
		if tinted {
			background = tone.Soft
		} else {
			background.A = byte((uint16(background.A) + 1) / 2)
		}
	}
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
		IconRenderer: neutralAvatarIconRenderer,
	}
}
