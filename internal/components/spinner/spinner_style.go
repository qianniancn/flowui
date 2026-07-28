package spinner

import (
	"image/color"

	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/frame"
	stateutil "github.com/qianniancn/flowui/internal/state"
	flowstyle "github.com/qianniancn/flowui/internal/style"
	styleruntime "github.com/qianniancn/flowui/internal/style/runtime"
	"github.com/qianniancn/flowui/internal/theme"
)

type spinnerStyle struct {
	color color.NRGBA
}

type spinnerResolvedStyle struct {
	root   flowstyle.ResolvedStyle
	visual spinnerStyle
	size   spinnerSizeStyle
}

func (s SpinnerWidget) resolveStyle(ctx *frame.Context, gtx layout.Context) spinnerResolvedStyle {
	activeTheme := frame.ActiveTheme(ctx)
	state := flowstyle.StyleState{}
	defaults := flowstyle.Style{}
	variant := spinnerColorDeclaration(activeTheme, s.color)
	size := spinnerSizeDeclaration(activeTheme, s.size)
	root := styleruntime.ResolveStatic(
		ctx,
		state,
		defaults,
		variant,
		size,
		s.customStyle,
	)
	indicator := styleruntime.ResolvePartStatic(ctx, flowstyle.PartIndicator, state, defaults, variant, size, s.customStyle)
	if styleruntime.HasTransitions(root, indicator) {
		key := frame.ClaimKey(ctx, stateutil.KindStyle, "spinner")
		root = styleruntime.ApplyTransitions(ctx, gtx, key, root)
		indicator = styleruntime.ApplyPartTransitions(ctx, gtx, key, flowstyle.PartIndicator, indicator)
	}
	visual := spinnerStyleFor(activeTheme, s.color)
	geometry := spinnerSizeStyleFor(activeTheme, s.size)
	if indicator.Text != nil {
		visual.color, _ = styleruntime.Color(indicator.Text.Color)
	}
	if indicator.Paint != nil {
		visual.color = styleruntime.ApplyOpacity(visual.color, indicator.Paint.Opacity)
	}
	return spinnerResolvedStyle{root: root, visual: visual, size: geometry}
}

func spinnerColorDeclaration(activeTheme *theme.Theme, spinnerColor SpinnerColor) flowstyle.Style {
	return flowstyle.Style{}.
		Part(flowstyle.PartIndicator, flowstyle.Style{}.TextColor(flowstyle.SolidColor{Color: spinnerStyleFor(activeTheme, spinnerColor).color}))

}

func spinnerSizeDeclaration(activeTheme *theme.Theme, size SpinnerSize) flowstyle.Style {
	diameter := spinnerSizeStyleFor(activeTheme, size).diameter
	return flowstyle.Style{}.Width(diameter).Height(diameter).AspectRatio(1)
}

type spinnerSizeStyle struct {
	diameter    unit.Dp
	strokeRatio float32
	insetRatio  float32
}

func spinnerStyleFor(activeTheme *theme.Theme, spinnerColor SpinnerColor) spinnerStyle {
	style := spinnerStyle{color: activeTheme.Palette.Accent}
	switch spinnerColor {
	case SpinnerCurrent:
		style.color = activeTheme.Palette.Foreground
	case SpinnerSuccess:
		style.color = activeTheme.Palette.Success
	case SpinnerWarning:
		style.color = activeTheme.Palette.Warning
	case SpinnerDanger:
		style.color = activeTheme.Palette.Danger
	}
	return style
}

func spinnerSizeStyleFor(activeTheme *theme.Theme, size SpinnerSize) spinnerSizeStyle {
	tokens := activeTheme.Components.Spinner
	diameter := tokens.MediumSize
	switch size {
	case SpinnerSmall:
		diameter = tokens.SmallSize
	case SpinnerLarge:
		diameter = tokens.LargeSize
	case SpinnerExtraLarge:
		diameter = tokens.ExtraLargeSize
	}
	return spinnerSizeStyle{
		diameter:    diameter,
		strokeRatio: tokens.StrokeRatio,
		insetRatio:  tokens.InsetRatio,
	}
}
