package progress

import (
	"image/color"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/frame"
	flowstyle "github.com/qianniancn/flowui/internal/style"
	styleruntime "github.com/qianniancn/flowui/internal/style/runtime"
	"github.com/qianniancn/flowui/internal/theme"
)

type progressBarStyle struct {
	track color.NRGBA
	fill  color.NRGBA
}

type progressBarResolvedStyle struct {
	root  flowstyle.ResolvedStyle
	track flowstyle.ResolvedStyle
	fill  flowstyle.ResolvedStyle
	label flowstyle.ResolvedStyle
}

func (p ProgressBarWidget) resolveStyle(ctx *frame.Context, gtx layout.Context, key string) progressBarResolvedStyle {
	activeTheme := frame.ActiveTheme(ctx)
	state := flowstyle.StyleState{Disabled: p.disabled}
	defaults := progressBarDefaultDeclaration(activeTheme, p.disabled)
	variant := progressBarVariantDeclaration(activeTheme, p.color, p.disabled)
	size := progressBarSizeDeclaration(activeTheme, p.size)
	root := styleruntime.Resolve(
		ctx,
		gtx,
		key,
		state,
		defaults,
		variant,
		size,
		p.customStyle,
	)
	track := styleruntime.ResolvePart(ctx, gtx, key, flowstyle.PartTrack, state, defaults, variant, size, p.customStyle)
	fill := styleruntime.ResolvePart(ctx, gtx, key, flowstyle.PartFill, state, defaults, variant, size, p.customStyle)
	label := styleruntime.ResolvePart(ctx, gtx, key, flowstyle.PartLabel, state, defaults, variant, size, p.customStyle)
	return progressBarResolvedStyle{root: root, track: track, fill: fill, label: label}
}

func progressBarDefaultDeclaration(activeTheme *theme.Theme, disabled bool) flowstyle.Style {
	label := progressBarLabelDeclaration(activeTheme, disabled)
	return flowstyle.Style{}.
		Part(flowstyle.PartTrack, flowstyle.Style{}.
			FillWidth().
			Overflow(flowstyle.OverflowHidden).
			Background(flowstyle.SolidColor{Color: activeTheme.Palette.SurfaceRaised})).
		Part(flowstyle.PartLabel, label)

}

func progressBarVariantDeclaration(activeTheme *theme.Theme, barColor ProgressBarColor, disabled bool) flowstyle.Style {
	resolved := progressBarStyleFor(activeTheme, barColor, disabled)
	return flowstyle.Style{}.
		Part(flowstyle.PartTrack, flowstyle.Style{}.Background(flowstyle.SolidColor{Color: resolved.track})).
		Part(flowstyle.PartFill, flowstyle.Style{}.Background(flowstyle.SolidColor{Color: resolved.fill}))

}

func progressBarSizeDeclaration(activeTheme *theme.Theme, size ProgressBarSize) flowstyle.Style {
	resolved := progressBarSizeStyleFor(activeTheme, size)
	return flowstyle.Style{}.
		Part(flowstyle.PartTrack, flowstyle.Style{}.Height(resolved.height).Radius(resolved.radius)).
		Part(flowstyle.PartFill, flowstyle.Style{}.Radius(resolved.radius))

}

func progressBarLabelDeclaration(activeTheme *theme.Theme, disabled bool) flowstyle.Style {
	foreground := activeTheme.Palette.Foreground
	if disabled {
		foreground = activeTheme.DisabledColor(foreground)
	}
	return flowstyle.Style{}.
		TextColor(flowstyle.SolidColor{Color: foreground}).
		FontSize(activeTheme.Components.ProgressBar.TextSize).
		FontWeight(int(font.Medium))

}

type progressBarSizeStyle struct {
	height unit.Dp
	radius unit.Dp
}

func progressBarStyleFor(theme *theme.Theme, barColor ProgressBarColor, disabled bool) progressBarStyle {
	style := progressBarStyle{
		track: theme.Palette.SurfaceRaised,
		fill:  theme.Palette.Accent,
	}
	switch barColor {
	case ProgressBarDefault:
		style.fill = theme.Palette.Foreground
	case ProgressBarSuccess:
		style.fill = theme.Palette.Success
	case ProgressBarWarning:
		style.fill = theme.Palette.Warning
	case ProgressBarDanger:
		style.fill = theme.Palette.Danger
	default:
		style.fill = theme.Palette.Accent
	}
	if disabled {
		style.track = theme.DisabledColor(style.track)
		style.fill = theme.DisabledColor(style.fill)
	}
	return style
}

func progressBarSizeStyleFor(theme *theme.Theme, size ProgressBarSize) progressBarSizeStyle {
	progressTheme := theme.Components.ProgressBar
	switch size {
	case ProgressBarSmall:
		return progressBarSizeStyle{
			height: progressTheme.SmallHeight,
			radius: progressTheme.SmallRadius,
		}
	case ProgressBarLarge:
		return progressBarSizeStyle{
			height: progressTheme.LargeHeight,
			radius: progressTheme.LargeRadius,
		}
	default:
		return progressBarSizeStyle{
			height: progressTheme.MediumHeight,
			radius: progressTheme.MediumRadius,
		}
	}
}
