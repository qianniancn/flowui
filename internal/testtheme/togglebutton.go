package testtheme

import (
	"time"

	"gioui.org/io/pointer"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

func (behaviorProvider) ToggleButton(active *theme.Theme, request theme.ToggleRequest) theme.ToggleVisual {
	tokens := active.Components.ToggleButton
	height, padding, textSize, pressedScale := tokens.MediumHeight, tokens.MediumPaddingX, tokens.MediumTextSize, tokens.PressedScaleMedium
	switch request.Size {
	case theme.ToggleButtonSmall:
		height, padding, textSize, pressedScale = tokens.SmallHeight, tokens.SmallPaddingX, tokens.SmallTextSize, tokens.PressedScaleSmall
	case theme.ToggleButtonLarge:
		height, padding, textSize, pressedScale = tokens.LargeHeight, tokens.LargePaddingX, tokens.LargeTextSize, tokens.PressedScaleLarge
	}
	if request.IconOnly {
		padding = 0
	}
	defaults := style.Style{}.
		Radius(tokens.Radius).
		Outline(tokens.FocusRingWidth, tokens.FocusRingOffset, style.RGBA(0)).
		Opacity(1).
		Scale(1).
		Cursor(pointer.CursorPointer).
		Part(style.PartLabel, style.Style{}.FontWeight(500)).
		Transition(style.PropBackgroundColor, 100*time.Millisecond).
		Transition(style.PropOutlineColor, 100*time.Millisecond).
		Transition(style.PropTransform, 160*time.Millisecond).
		On(style.Disabled, style.Style{}.Opacity(active.DisabledOpacityValue()).Cursor(pointer.CursorDefault)).
		OnNot(style.FocusVisible, style.Disabled, style.Style{}.Outline(tokens.FocusRingWidth, tokens.FocusRingOffset, style.Solid(focus)))
	variant := style.Style{}.
		Background(style.Solid(background)).
		TextColor(style.Solid(request.Foreground)).
		On(style.Hovered, style.Style{}.Background(style.Solid(accent))).
		On(style.Pressed, style.Style{}.Background(style.Solid(muted))).
		On(style.Selected, style.Style{}.
			Background(style.Solid(active.Semantic().Tone(theme.TonePrimary).Soft)).
			TextColor(style.Solid(active.Semantic().Tone(theme.TonePrimary).OnSoft))).
		On(style.Selected|style.Hovered, style.Style{}.Background(style.Solid(accent))).
		On(style.Selected|style.Pressed, style.Style{}.Background(style.Solid(accent)))
	size := style.Style{}.
		Height(height).
		PaddingX(padding).
		Part(style.PartLabel, style.Style{}.FontSize(textSize)).
		Part(style.PartIcon, style.Style{}.
			FontSize(textSize).
			MinWidth(unit.Dp(textSize)).
			MinHeight(unit.Dp(textSize))).
		Rules(style.Rule{When: style.Pressed, Set: style.Style{}.Scale(pressedScale)})
	if request.IconOnly {
		size = size.AspectRatio(1).PaddingX(0)
	}
	return theme.ToggleVisual{
		Defaults: defaults, Variant: variant, Size: size,
		Override: theme.ToggleOverrideStyle(request),
	}
}
