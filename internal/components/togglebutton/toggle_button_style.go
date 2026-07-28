package togglebutton

import (
	"image/color"

	"gioui.org/font"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/frame"
	flowstyle "github.com/qianniancn/flowui/internal/style"
	styleruntime "github.com/qianniancn/flowui/internal/style/runtime"
	"github.com/qianniancn/flowui/internal/theme"
)

type toggleButtonWidget struct {
	background color.NRGBA
	foreground color.NRGBA
	radius     unit.Dp
	opacity    float32
}

type toggleButtonResolvedStyle struct {
	root    flowstyle.ResolvedStyle
	content flowstyle.ResolvedStyle
}

func (b ToggleButtonWidget) resolveStyle(ctx *frame.Context, gtx layout.Context, key string, state flowstyle.StyleState) toggleButtonResolvedStyle {
	activeTheme := frame.ActiveTheme(ctx)
	defaults := toggleButtonDefaultDeclaration(activeTheme)
	variant := toggleButtonVariantDeclaration(activeTheme, ctx.ForegroundColor(), b.variant)
	size := toggleButtonSizeDeclaration(activeTheme, b.size, b.iconOnly)
	part := flowstyle.PartLabel
	if b.iconOnly {
		part = flowstyle.PartIcon
	}
	return toggleButtonResolvedStyle{
		root:    styleruntime.Resolve(ctx, gtx, key, state, defaults, variant, size, b.customStyle),
		content: styleruntime.ResolvePart(ctx, gtx, key, part, state, defaults, variant, size, b.customStyle),
	}
}

func toggleButtonDefaultDeclaration(activeTheme *theme.Theme) flowstyle.Style {
	tokens := activeTheme.Components.ToggleButton
	return flowstyle.Style{}.
		Radius(tokens.Radius).
		Outline(tokens.FocusRingWidth, tokens.FocusRingOffset, flowstyle.WithAlpha(flowstyle.TokenFocus, 0)).
		Opacity(1).
		Scale(1, 1).
		Cursor(pointer.CursorPointer).
		Part(flowstyle.PartLabel, flowstyle.Style{}.FontWeight(int(font.Medium))).
		Transition(flowstyle.PropBackgroundColor, toggleButtonColorDuration).
		Transition(flowstyle.PropOutlineColor, toggleButtonColorDuration).
		Transition(flowstyle.PropTransform, toggleButtonScaleDuration).
		When(flowstyle.All(flowstyle.FocusVisible, flowstyle.Not(flowstyle.Disabled)),
			flowstyle.Style{}.Outline(tokens.FocusRingWidth, tokens.FocusRingOffset, flowstyle.TokenFocus),
		).
		When(flowstyle.Disabled,
			flowstyle.Style{}.Opacity(activeTheme.DisabledOpacityValue()).Cursor(pointer.CursorDefault),
		)

}

func toggleButtonVariantDeclaration(activeTheme *theme.Theme, foreground color.NRGBA, variant ToggleButtonVariant) flowstyle.Style {
	idle := toggleButtonWidgetFor(activeTheme, foreground, variant, false, false, false, false)
	hovered := toggleButtonWidgetFor(activeTheme, foreground, variant, false, true, false, false)
	selected := toggleButtonWidgetFor(activeTheme, foreground, variant, true, false, false, false)
	selectedHovered := toggleButtonWidgetFor(activeTheme, foreground, variant, true, true, false, false)
	hoveredOrPressed := flowstyle.Any(flowstyle.Hovered, flowstyle.Pressed)
	return flowstyle.Style{}.
		Background(flowstyle.SolidColor{Color: idle.background}).
		TextColor(flowstyle.SolidColor{Color: idle.foreground}).
		When(hoveredOrPressed,
			flowstyle.Style{}.Background(flowstyle.SolidColor{Color: hovered.background}),
		).
		When(flowstyle.Selected,
			flowstyle.Style{}.
				Background(flowstyle.SolidColor{Color: selected.background}).
				TextColor(flowstyle.SolidColor{Color: selected.foreground}),
		).
		When(flowstyle.All(flowstyle.Selected, hoveredOrPressed),
			flowstyle.Style{}.Background(flowstyle.SolidColor{Color: selectedHovered.background}),
		)

}

func toggleButtonSizeDeclaration(activeTheme *theme.Theme, size ToggleButtonSize, iconOnly bool) flowstyle.Style {
	resolved := toggleButtonSizeStyleFor(activeTheme, size)
	pressedScale := resolvedToggleButtonPressedScale(resolved.pressedScale, size)
	builder := flowstyle.Style{}.
		Height(resolved.height).
		PaddingX(resolved.paddingX).
		Part(flowstyle.PartLabel, flowstyle.Style{}.FontSize(unit.Sp(resolved.textSize))).
		When(flowstyle.Pressed,
			flowstyle.Style{}.Scale(pressedScale, pressedScale),
		)
	if iconOnly {
		builder = builder.AspectRatio(1).PaddingX(0)
	}
	return builder
}

type toggleButtonSizeStyle struct {
	height       unit.Dp
	paddingX     unit.Dp
	textSize     float32
	pressedScale float32
}

func toggleButtonWidgetFor(activeTheme *theme.Theme, currentForeground color.NRGBA, variant ToggleButtonVariant, selected, hovered, pressed, disabled bool) toggleButtonWidget {
	background := activeTheme.Palette.SurfaceRaised
	foreground := currentForeground
	if variant == ToggleButtonGhost {
		background = color.NRGBA{}
		foreground = activeTheme.Palette.Foreground
	}
	if hovered || pressed {
		background = activeTheme.Palette.SurfacePressed
	}
	if selected {
		background = activeTheme.Palette.AccentSoft
		foreground = activeTheme.Palette.AccentSoftForeground
		if hovered || pressed {
			background = activeTheme.Palette.AccentSoftHover
		}
	}
	opacity := float32(1)
	if disabled {
		opacity = activeTheme.DisabledOpacityValue()
	}
	return toggleButtonWidget{
		background: background,
		foreground: foreground,
		radius:     activeTheme.Components.ToggleButton.Radius,
		opacity:    opacity,
	}
}

func toggleButtonSizeStyleFor(activeTheme *theme.Theme, size ToggleButtonSize) toggleButtonSizeStyle {
	tokens := activeTheme.Components.ToggleButton
	style := toggleButtonSizeStyle{
		height:       tokens.MediumHeight,
		paddingX:     tokens.MediumPaddingX,
		textSize:     float32(tokens.MediumTextSize),
		pressedScale: tokens.PressedScaleMedium,
	}
	switch size {
	case ToggleButtonSmall:
		style.height = tokens.SmallHeight
		style.paddingX = tokens.SmallPaddingX
		style.textSize = float32(tokens.SmallTextSize)
		style.pressedScale = tokens.PressedScaleSmall
	case ToggleButtonLarge:
		style.height = tokens.LargeHeight
		style.paddingX = tokens.LargePaddingX
		style.textSize = float32(tokens.LargeTextSize)
		style.pressedScale = tokens.PressedScaleLarge
	}
	return style
}

func resolvedToggleButtonPressedScale(scale float32, size ToggleButtonSize) float32 {
	if scale > 0 && scale <= 1 {
		return scale
	}
	switch size {
	case ToggleButtonSmall:
		return 0.98
	case ToggleButtonLarge:
		return 0.96
	default:
		return 0.97
	}
}
