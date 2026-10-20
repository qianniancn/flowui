package testtheme

import (
	"image"
	"image/color"
	"time"

	"gioui.org/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

func (behaviorProvider) ProgressBar(active *theme.Theme, request theme.ProgressBarRequest) theme.ProgressBarVisual {
	height, radius := active.Components.ProgressBar.MediumHeight, active.Components.ProgressBar.MediumRadius
	switch request.Size {
	case theme.ProgressBarSmall:
		height, radius = active.Components.ProgressBar.SmallHeight, active.Components.ProgressBar.SmallRadius
	case theme.ProgressBarLarge:
		height, radius = active.Components.ProgressBar.LargeHeight, active.Components.ProgressBar.LargeRadius
	}
	defaults := style.Style{}.
		Part(style.PartTrack, style.Style{}.FillWidth().Overflow(style.OverflowHidden).Background(style.Solid(background))).
		Part(style.PartLabel, style.Style{}.TextColor(style.Solid(foreground)).FontSize(active.Components.ProgressBar.TextSize)).
		Rules(style.Rule{When: style.Disabled, Set: style.Style{}.Part(style.PartLabel, style.Style{}.TextColor(style.Solid(muted)))})
	variant := style.Style{}.
		Part(style.PartTrack, style.Style{}.Background(style.Solid(accent))).
		Part(style.PartFill, style.Style{}.Background(style.Solid(focus))).
		Rules(style.Rule{When: style.Disabled, Set: style.Style{}.Part(style.PartTrack, style.Style{}.Background(style.Solid(muted))).Part(style.PartFill, style.Style{}.Background(style.Solid(muted)))})
	size := style.Style{}.
		Part(style.PartTrack, style.Style{}.Height(height).Radius(radius)).
		Part(style.PartFill, style.Style{}.Radius(radius))
	return theme.ProgressBarVisual{
		Defaults: defaults, Variant: variant, Size: size,
		Override: theme.ProgressOverrideStyle(request.Override),
		Color:    neutralProgressColor,
		Metrics:  theme.ProgressBarMetrics{HeaderGap: active.Components.ProgressBar.HeaderGap, ValueDuration: active.Motion.ControlDuration, IndeterminatePeriod: 1500 * time.Millisecond, IndeterminateFillRate: .4},
		Motion:   active.Motion,
	}
}

func (behaviorProvider) ProgressCircle(active *theme.Theme, request theme.ProgressCircleRequest) theme.ProgressCircleVisual {
	diameter := active.Components.ProgressCircle.MediumSize
	switch request.Size {
	case theme.ProgressBarSmall:
		diameter = active.Components.ProgressCircle.SmallSize
	case theme.ProgressBarLarge:
		diameter = active.Components.ProgressCircle.LargeSize
	}
	defaults := style.Style{}.
		Width(diameter).Height(diameter).AspectRatio(1).
		Part(style.PartTrack, style.Style{}.Background(style.Solid(background))).
		Part(style.PartFill, style.Style{}.Background(style.Solid(focus))).
		Rules(style.Rule{When: style.Disabled, Set: style.Style{}.
			Part(style.PartTrack, style.Style{}.Background(style.Solid(muted))).
			Part(style.PartFill, style.Style{}.Background(style.Solid(muted)))})
	return theme.ProgressCircleVisual{
		Defaults: defaults,
		Override: theme.ProgressOverrideStyle(request.Override),
		Color:    neutralProgressColor,
		Renderer: neutralProgressCircleRenderer,
		Metrics:  theme.ProgressCircleMetrics{StrokeRatio: active.Components.ProgressCircle.StrokeRatio, ValueDuration: active.Motion.ControlDuration, SpinDuration: active.Components.ProgressCircle.SpinDuration},
		Motion:   active.Motion,
	}
}

// neutralProgressCircleRenderer keeps behavior fixtures independent of the
// shipped progress recipe while satisfying the renderer contract.
func neutralProgressCircleRenderer(gtx layout.Context, request theme.ProgressCircleRenderRequest) layout.Dimensions {
	diameter := max(gtx.Dp(request.Size), 0)
	return layout.Dimensions{Size: image.Pt(diameter, diameter)}
}

func neutralProgressColor(value style.ResolvedStyle) (color.NRGBA, bool) {
	if value.Paint == nil {
		return color.NRGBA{}, false
	}
	switch source := value.Paint.Background.(type) {
	case style.SolidColor:
		return applyProgressOpacity(source.Color, value.Paint.Opacity), true
	case *style.SolidColor:
		if source != nil {
			return applyProgressOpacity(source.Color, value.Paint.Opacity), true
		}
	}
	return color.NRGBA{}, false
}

func applyProgressOpacity(value color.NRGBA, opacity *float32) color.NRGBA {
	if opacity == nil {
		return value
	}
	value.A = byte(float32(value.A)*min(max(*opacity, 0), 1) + .5)
	return value
}
