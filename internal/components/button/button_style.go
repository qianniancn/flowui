package button

import (
	"image/color"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/FlowUI/internal/frame"
	flowstyle "github.com/qianniancn/FlowUI/internal/style"
	styleruntime "github.com/qianniancn/FlowUI/internal/style/runtime"
	"github.com/qianniancn/FlowUI/internal/theme"
)

type buttonStyle struct {
	root          flowstyle.ResolvedStyle
	content       flowstyle.ResolvedStyle
	indicatorPart flowstyle.ResolvedStyle
	height        unit.Dp
	inset         layout.Inset
	textSize      float32
	text          *flowstyle.TextStyle
	bg            color.NRGBA
	fg            color.NRGBA
	indicator     color.NRGBA
	border        color.NRGBA
	hasBorder     bool
	radius        unit.Dp
	borderWidth   unit.Dp
	opacity       float32
	scaleX        float32
	scaleY        float32
}

type buttonPalette struct {
	bg        color.NRGBA
	hover     color.NRGBA
	pressed   color.NRGBA
	fg        color.NRGBA
	border    color.NRGBA
	hasBorder bool
}

// resolveStyle is the reference composite-widget style assembly:
// defaults + variant + size + StyleScope(runtime) + instance, with StyleState
// from interact. Other controls should follow this four-slot pattern.
func (b ButtonWidget) resolveStyle(ctx *frame.Context, gtx layout.Context, key string, state flowstyle.StyleState) buttonStyle {
	activeTheme := frame.ActiveTheme(ctx)
	defaults, variant, size := b.styleDeclarations(activeTheme)
	resolved := styleruntime.Resolve(
		ctx,
		gtx,
		key,
		state,
		defaults,
		variant,
		size,
		b.customStyle,
	)
	contentPart := flowstyle.PartLabel
	if b.iconOnly {
		contentPart = flowstyle.PartIcon
	}
	content := styleruntime.ResolvePart(ctx, gtx, key, contentPart, state, defaults, variant, size, b.customStyle)
	indicator := styleruntime.ResolvePart(ctx, gtx, key, flowstyle.PartIndicator, state, defaults, variant, size, b.customStyle)
	return buttonStyleFrom(resolved, content, indicator, activeTheme)
}

func (b ButtonWidget) staticStyle(ctx *frame.Context, state flowstyle.StyleState) buttonStyle {
	activeTheme := frame.ActiveTheme(ctx)
	defaults, variant, size := b.styleDeclarations(activeTheme)
	resolved := styleruntime.ResolveStatic(
		ctx,
		state,
		defaults,
		variant,
		size,
		b.customStyle,
	)
	contentPart := flowstyle.PartLabel
	if b.iconOnly {
		contentPart = flowstyle.PartIcon
	}
	content := styleruntime.ResolvePartStatic(ctx, contentPart, state, defaults, variant, size, b.customStyle)
	indicator := styleruntime.ResolvePartStatic(ctx, flowstyle.PartIndicator, state, defaults, variant, size, b.customStyle)
	return buttonStyleFrom(resolved, content, indicator, activeTheme)
}

// styleDeclarations returns the three component-owned cascade layers
// (defaults, variant, size). Instance style is b.customStyle; StyleScope is
// injected by the style runtime.
func (b ButtonWidget) styleDeclarations(activeTheme *theme.Theme) (defaults, variant, size flowstyle.Style) {
	return buttonDefaultDeclaration(activeTheme),
		buttonVariantDeclaration(activeTheme, b.variant),
		flowstyle.Join(
			buttonSizeDeclaration(activeTheme, b.size, b.iconOnly, b.group.grouped),
			buttonGroupDeclaration(activeTheme, b.group),
		)
}

func buttonDefaultDeclaration(activeTheme *theme.Theme) flowstyle.Style {
	return flowstyle.Style{}.
		Radius(activeTheme.Components.Button.Radius).
		FontWeight(int(font.Medium)).
		Outline(
			activeTheme.Components.Button.FocusRingWidth,
			1,
			flowstyle.WithAlpha(flowstyle.TokenFocus, 0),
		).
		Opacity(1).
		Scale(1, 1).
		Transition(flowstyle.PropBackgroundColor, buttonColorDuration).
		Transition(flowstyle.PropOpacity, buttonColorDuration).
		Transition(flowstyle.PropOutlineColor, buttonFocusDuration).
		Transition(flowstyle.PropTransform, buttonPressOutDuration).
		When(flowstyle.Disabled,
			flowstyle.Style{}.Opacity(activeTheme.DisabledOpacityValue()),
		).
		When(flowstyle.All(flowstyle.FocusVisible, flowstyle.Not(flowstyle.Disabled)),
			flowstyle.Style{}.Outline(
				activeTheme.Components.Button.FocusRingWidth,
				1,
				flowstyle.TokenFocus,
			),
		)

}

func buttonVariantDeclaration(activeTheme *theme.Theme, variant ButtonVariant) flowstyle.Style {
	colors := buttonColors(activeTheme, variant)
	declaration := flowstyle.Style{}.
		Background(solid(colors.bg)).
		TextColor(solid(colors.fg))
	if colors.hasBorder {
		declaration = declaration.
			BorderColor(solid(colors.border)).
			BorderWidth(activeTheme.Components.Button.BorderWidth)
	}
	return declaration.
		When(flowstyle.Hovered,
			flowstyle.Style{}.Background(solid(colors.hover)),
		).
		When(flowstyle.Pressed,
			flowstyle.Style{}.Background(solid(colors.pressed)),
		)

}

func buttonSizeDeclaration(activeTheme *theme.Theme, size ButtonSize, iconOnly, grouped bool) flowstyle.Style {
	height := activeTheme.Spacing.ControlHeight
	padding := activeTheme.Spacing.ControlPaddingX
	textSize := activeTheme.Typography.ControlSize
	switch size {
	case ButtonSmall:
		height = activeTheme.Spacing.SmallControlHeight
		padding = activeTheme.Spacing.SmallControlPaddingX
	case ButtonLarge:
		height = activeTheme.Spacing.LargeControlHeight
		padding = activeTheme.Spacing.LargeControlPaddingX
		textSize += 2
	}
	if iconOnly {
		padding = 0
	}
	declaration := flowstyle.Style{}.
		Height(height).
		PaddingX(padding).
		FontSize(textSize)
	if iconOnly {
		declaration = declaration.Width(height)
	}
	if !grouped {
		declaration = declaration.When(flowstyle.Pressed,
			flowstyle.Style{}.
				Scale(buttonPressedScale(activeTheme, size), buttonPressedScale(activeTheme, size)).
				Transition(flowstyle.PropTransform, buttonPressInDuration),
		)
	}
	return declaration
}

func buttonGroupDeclaration(activeTheme *theme.Theme, group buttonGroupItemStyle) flowstyle.Style {
	if !group.grouped || group.position == buttonGroupSingle {
		return flowstyle.Style{}
	}
	corners := buttonGroupCorners(group)
	builder := flowstyle.Style{}.Radius(0)
	radius := activeTheme.Components.Button.Radius
	if corners.nw {
		builder = builder.RadiusTopLeft(radius)
	}
	if corners.ne {
		builder = builder.RadiusTopRight(radius)
	}
	if corners.se {
		builder = builder.RadiusBottomRight(radius)
	}
	if corners.sw {
		builder = builder.RadiusBottomLeft(radius)
	}
	return builder
}

func buttonStyleFrom(resolved, content, indicator flowstyle.ResolvedStyle, activeTheme *theme.Theme) buttonStyle {
	result := buttonStyle{
		root:          resolved,
		content:       content,
		indicatorPart: indicator,
		opacity:       1,
		scaleX:        1,
		scaleY:        1,
		borderWidth:   activeTheme.Components.Button.BorderWidth,
	}
	if resolved.Box != nil {
		if resolved.Box.Height != nil {
			result.height = *resolved.Box.Height
		}
		if resolved.Box.Padding != nil {
			result.inset = layout.Inset{
				Top:    resolved.Box.Padding.Top,
				Right:  resolved.Box.Padding.Right,
				Bottom: resolved.Box.Padding.Bottom,
				Left:   resolved.Box.Padding.Left,
			}
		}
	}
	if resolved.Paint != nil {
		if background, ok := solidColor(resolved.Paint.Background); ok {
			result.bg = background
		}
		if resolved.Paint.Border != nil {
			result.border, _ = solidColor(resolved.Paint.Border.Color)
			if resolved.Paint.Border.Width != nil {
				result.borderWidth = *resolved.Paint.Border.Width
				result.hasBorder = result.borderWidth > 0
			}
		}
		if resolved.Paint.Radius != nil {
			result.radius = *resolved.Paint.Radius
		}
		if resolved.Paint.Opacity != nil {
			result.opacity = *resolved.Paint.Opacity
		}
	}
	if content.Text != nil {
		result.text = content.Text
		result.fg, _ = solidColor(content.Text.Color)
		if content.Text.FontSize != nil {
			result.textSize = float32(*content.Text.FontSize)
		}
	}
	result.indicator = result.fg
	if indicator.Text != nil {
		result.indicator, _ = solidColor(indicator.Text.Color)
	}
	if resolved.Trans != nil {
		if resolved.Trans.ScaleX != nil {
			result.scaleX = *resolved.Trans.ScaleX
		}
		if resolved.Trans.ScaleY != nil {
			result.scaleY = *resolved.Trans.ScaleY
		}
	}
	return result
}

func solid(value color.NRGBA) flowstyle.SolidColor {
	return flowstyle.SolidColor{Color: value}
}

func solidColor(source any) (color.NRGBA, bool) {
	switch value := source.(type) {
	case flowstyle.SolidColor:
		return value.Color, true
	case *flowstyle.SolidColor:
		return value.Color, true
	default:
		return color.NRGBA{}, false
	}
}

type buttonCorners struct {
	nw bool
	ne bool
	se bool
	sw bool
}

func buttonGroupCorners(group buttonGroupItemStyle) buttonCorners {
	all := buttonCorners{nw: true, ne: true, se: true, sw: true}
	if !group.grouped || group.position == buttonGroupSingle {
		return all
	}
	if group.orientation == ButtonGroupVertical {
		switch group.position {
		case buttonGroupStart:
			return buttonCorners{nw: true, ne: true}
		case buttonGroupEnd:
			return buttonCorners{se: true, sw: true}
		default:
			return buttonCorners{}
		}
	}
	switch group.position {
	case buttonGroupStart:
		return buttonCorners{nw: true, sw: true}
	case buttonGroupEnd:
		return buttonCorners{ne: true, se: true}
	default:
		return buttonCorners{}
	}
}

func buttonSpinnerSize(activeTheme *theme.Theme, size ButtonSize) unit.Dp {
	switch size {
	case ButtonSmall:
		return activeTheme.Components.Button.SpinnerSmall
	case ButtonLarge:
		return activeTheme.Components.Button.SpinnerLarge
	default:
		return activeTheme.Components.Button.SpinnerMedium
	}
}

func buttonLoadingInset(gtx layout.Context, activeTheme *theme.Theme, size ButtonSize, inset layout.Inset) layout.Inset {
	padding := gtx.Dp(inset.Left) + gtx.Dp(inset.Right)
	extra := gtx.Dp(buttonSpinnerSize(activeTheme, size)) + gtx.Dp(activeTheme.Components.Button.ContentGap)
	remaining := max(padding-extra, 0)
	left := remaining / 2
	inset.Left = gtx.Metric.PxToDp(left)
	inset.Right = gtx.Metric.PxToDp(remaining - left)
	return inset
}

func buttonLoadingStyle(gtx layout.Context, activeTheme *theme.Theme, size ButtonSize, style buttonStyle) buttonStyle {
	style.inset = buttonLoadingInset(gtx, activeTheme, size, style.inset)
	if style.root.Box == nil {
		style.root.Box = &flowstyle.BoxStyle{}
	}
	padding := flowstyle.Insets{}
	if style.root.Box.Padding != nil {
		padding = *style.root.Box.Padding
	}
	padding.Left = style.inset.Left
	padding.Right = style.inset.Right
	style.root.Box.Padding = &padding
	return style
}

func buttonColors(activeTheme *theme.Theme, variant ButtonVariant) buttonPalette {
	transparent := color.NRGBA{}
	foreground := activeTheme.Palette.DefaultForegroundColor()
	defaultBg := activeTheme.Palette.DefaultColor()
	defaultHover := activeTheme.Palette.DefaultHoverColor()
	accent := activeTheme.Palette.Accent
	accentHover := activeTheme.Palette.AccentHover
	accentFg := activeTheme.Palette.AccentForeground
	accentSoftFg := activeTheme.Palette.AccentSoftForeground
	danger := activeTheme.Palette.Danger
	dangerHover := activeTheme.Palette.DangerHover
	dangerPressed := activeTheme.Palette.DangerPressed
	dangerSoft := activeTheme.Palette.DangerSoft
	dangerSoftHover := activeTheme.Palette.DangerSoftHover
	dangerSoftFg := activeTheme.Palette.DangerSoftForeground
	border := activeTheme.Palette.Border
	outlineHover := defaultBg
	outlineHover.A = byte(uint16(outlineHover.A) * 0x99 / 0xff)

	switch variant {
	case ButtonSecondary:
		return buttonPalette{bg: defaultBg, hover: defaultHover, pressed: defaultHover, fg: accentSoftFg}
	case ButtonTertiary:
		return buttonPalette{bg: defaultBg, hover: defaultHover, pressed: defaultHover, fg: foreground}
	case ButtonGhost:
		return buttonPalette{bg: transparent, hover: defaultBg, pressed: defaultBg, fg: foreground}
	case ButtonOutline:
		return buttonPalette{
			bg:        transparent,
			hover:     outlineHover,
			pressed:   defaultBg,
			fg:        foreground,
			border:    border,
			hasBorder: true,
		}
	case ButtonDanger:
		return buttonPalette{bg: danger, hover: dangerHover, pressed: dangerPressed, fg: accentFg}
	case ButtonDangerSoft:
		return buttonPalette{bg: dangerSoft, hover: dangerSoftHover, pressed: dangerSoftHover, fg: dangerSoftFg}
	default:
		return buttonPalette{bg: accent, hover: accentHover, pressed: activeTheme.Palette.AccentPressed, fg: accentFg}
	}
}
