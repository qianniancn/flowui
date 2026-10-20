package testtheme

import (
	"image"
	"image/color"

	"gioui.org/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

func (behaviorProvider) Spinner(active *theme.Theme, request theme.SpinnerRequest) theme.SpinnerVisual {
	tokens := active.Components.Spinner
	diameter := tokens.MediumSize
	switch request.Size {
	case theme.SpinnerSmall:
		diameter = tokens.SmallSize
	case theme.SpinnerLarge:
		diameter = tokens.LargeSize
	case theme.SpinnerExtraLarge:
		diameter = tokens.ExtraLargeSize
	}
	return theme.SpinnerVisual{
		Defaults: style.Style{}.
			Width(diameter).
			Height(diameter).
			AspectRatio(1).
			Part(style.PartIndicator, style.Style{}.
				TextColor(style.Solid(neutralSpinnerColor(request.Color)))),
		Override: theme.SpinnerOverrideStyle(request.Override),
		Color:    neutralSpinnerColorResolver,
		Renderer: neutralSpinnerRenderer,
		Metrics: theme.SpinnerMetrics{
			StrokeRatio:  tokens.StrokeRatio,
			InsetRatio:   tokens.InsetRatio,
			SpinDuration: tokens.SpinDuration,
		},
		Motion: active.Motion,
	}
}

// neutralSpinnerRenderer is intentionally geometry-neutral: behavior tests
// need a valid renderer contract and dimensions, but must not import or reuse
// a shipped theme's spinner recipe.
func neutralSpinnerRenderer(gtx layout.Context, request theme.SpinnerRenderRequest) layout.Dimensions {
	diameter := max(gtx.Dp(request.Size), 0)
	return layout.Dimensions{Size: image.Pt(diameter, diameter)}
}

func neutralSpinnerColor(value theme.SpinnerColor) color.NRGBA {
	switch value {
	case theme.SpinnerCurrent:
		return foreground
	case theme.SpinnerSuccess:
		return color.NRGBA{R: 0x20, G: 0x90, B: 0x50, A: 0xff}
	case theme.SpinnerWarning:
		return color.NRGBA{R: 0xa0, G: 0x70, B: 0x10, A: 0xff}
	case theme.SpinnerDanger:
		return color.NRGBA{R: 0xb0, G: 0x30, B: 0x30, A: 0xff}
	default:
		return accent
	}
}

func neutralSpinnerColorResolver(value style.ResolvedStyle) (color.NRGBA, bool) {
	if value.Text == nil {
		return color.NRGBA{}, false
	}
	var resolved color.NRGBA
	switch source := value.Text.Color.(type) {
	case style.SolidColor:
		resolved = source.Color
	case *style.SolidColor:
		if source == nil {
			return color.NRGBA{}, false
		}
		resolved = source.Color
	default:
		return color.NRGBA{}, false
	}
	if value.Paint != nil && value.Paint.Opacity != nil {
		resolved.A = byte(float32(resolved.A)*min(max(*value.Paint.Opacity, 0), 1) + .5)
	}
	return resolved, true
}
