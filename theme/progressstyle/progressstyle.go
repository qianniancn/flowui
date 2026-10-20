// Package progressstyle owns the visual recipes used by progress providers.
// Progress behavior remains in widget/progress: this package only maps theme
// semantics and component tokens to declarations and neutral layout metrics.
package progressstyle

import (
	"image/color"
	"time"

	"gioui.org/font"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/appearance"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

// BuildProgressBar builds the provider-owned linear progress visual plan.
func BuildProgressBar(active *theme.Theme, request theme.ProgressBarRequest) theme.ProgressBarVisual {
	if active == nil {
		panic(theme.ErrMissingVisualProvider)
	}
	return theme.ProgressBarVisual{
		Defaults: barDefaults(active),
		Variant:  barVariant(active, request.Color),
		Size:     barSize(active, request.Size),
		Override: theme.ProgressOverrideStyle(request.Override),
		Color:    progressColor,
		Metrics: theme.ProgressBarMetrics{
			HeaderGap:             active.Components.ProgressBar.HeaderGap,
			ValueDuration:         progressValueDuration(active.Motion),
			IndeterminatePeriod:   1500 * time.Millisecond,
			IndeterminateFillRate: .4,
		},
		Motion: active.Motion,
	}
}

// BuildProgressCircle builds the provider-owned circular progress visual plan.
func BuildProgressCircle(active *theme.Theme, request theme.ProgressCircleRequest) theme.ProgressCircleVisual {
	if active == nil {
		panic(theme.ErrMissingVisualProvider)
	}
	return theme.ProgressCircleVisual{
		Defaults: circleDefaults(active, request.Color, request.Size),
		Override: theme.ProgressOverrideStyle(request.Override),
		Color:    progressColor,
		Renderer: progressCircleRenderer,
		Metrics: theme.ProgressCircleMetrics{
			StrokeRatio:   active.Components.ProgressCircle.StrokeRatio,
			ValueDuration: progressValueDuration(active.Motion),
			SpinDuration:  active.Components.ProgressCircle.SpinDuration,
		},
		Motion: active.Motion,
	}
}

// progressColor keeps brush and opacity interpretation on the provider side.
// Circle drawing needs a concrete color, while the widget remains agnostic to
// whether a theme uses a solid fill or a gradient declaration.
func progressColor(value style.ResolvedStyle) (color.NRGBA, bool) {
	if value.Paint == nil {
		return color.NRGBA{}, false
	}
	brush, ok := computed.Brush(value.Paint.Background)
	if !ok {
		return color.NRGBA{}, false
	}
	return computed.ApplyOpacity(brush.ColorAt(.5), value.Paint.Opacity), true
}

func progressValueDuration(motion theme.MotionTheme) time.Duration {
	if motion.ControlDuration > 0 {
		return motion.ControlDuration
	}
	return motion.DefaultDuration
}

func barDefaults(active *theme.Theme) style.Style {
	sem := active.Semantic()
	return style.Style{}.
		Part(style.PartTrack, style.Style{}.
			FillWidth().
			Overflow(style.OverflowHidden).
			Background(style.Solid(sem.SurfaceRaised))).
		Part(style.PartLabel, style.Style{}.
			TextColor(style.Solid(sem.Foreground)).
			FontSize(active.Components.ProgressBar.TextSize).
			FontWeight(int(font.Medium))).
		Rules(style.Rule{
			When: style.Disabled,
			Set: style.Style{}.
				Part(style.PartLabel, style.Style{}.
					TextColor(style.Solid(active.DisabledColor(sem.Foreground)))),
		})
}

func barVariant(active *theme.Theme, value theme.ProgressBarColor) style.Style {
	enabled := barColors(active, value, false)
	disabled := barColors(active, value, true)
	parts := func(colors barPalette) style.Style {
		return style.Style{}.
			Part(style.PartTrack, style.Style{}.Background(style.Solid(colors.track))).
			Part(style.PartFill, style.Style{}.Background(style.Solid(colors.fill)))
	}
	return parts(enabled).Rules(style.Rule{When: style.Disabled, Set: parts(disabled)})
}

func barSize(active *theme.Theme, size theme.ProgressBarSize) style.Style {
	height, radius := barDimensions(active, size)
	return style.Style{}.
		Part(style.PartTrack, style.Style{}.Height(height).Radius(radius)).
		Part(style.PartFill, style.Style{}.Radius(radius))
}

func circleDefaults(active *theme.Theme, value theme.ProgressBarColor, size theme.ProgressBarSize) style.Style {
	enabled := circleColors(active, value, false)
	disabled := circleColors(active, value, true)
	diameter := circleDiameter(active, size)
	parts := func(colors circlePalette) style.Style {
		return style.Style{}.
			Part(style.PartTrack, style.Style{}.Background(style.Solid(colors.track))).
			Part(style.PartFill, style.Style{}.Background(style.Solid(colors.fill)))
	}
	return style.Join(style.Style{}.Width(diameter).Height(diameter).AspectRatio(1), parts(enabled)).
		Rules(style.Rule{When: style.Disabled, Set: parts(disabled)})
}

type barPalette struct{ track, fill color.NRGBA }

func barColors(active *theme.Theme, value theme.ProgressBarColor, disabled bool) barPalette {
	sem := active.Semantic()
	result := barPalette{track: sem.SurfaceRaised, fill: tone(active, value).Fill}
	if value == theme.ProgressBarDefault {
		result.fill = sem.Foreground
	}
	if disabled {
		result.track = active.DisabledColor(result.track)
		result.fill = active.DisabledColor(result.fill)
	}
	return result
}

type circlePalette struct{ track, fill color.NRGBA }

func circleColors(active *theme.Theme, value theme.ProgressBarColor, disabled bool) circlePalette {
	neutral := appearance.Tone(active, theme.ToneDefault)
	result := circlePalette{track: neutral.Fill, fill: tone(active, value).Fill}
	if value == theme.ProgressBarDefault {
		result.fill = neutral.OnFill
	}
	if disabled {
		result.track = active.DisabledColor(result.track)
		result.fill = active.DisabledColor(result.fill)
	}
	return result
}

func tone(active *theme.Theme, value theme.ProgressBarColor) appearance.ToneColors {
	switch value {
	case theme.ProgressBarSuccess:
		return appearance.Tone(active, theme.ToneSuccess)
	case theme.ProgressBarWarning:
		return appearance.Tone(active, theme.ToneWarning)
	case theme.ProgressBarDanger:
		return appearance.Tone(active, theme.ToneDanger)
	default:
		return appearance.Tone(active, theme.TonePrimary)
	}
}

func barDimensions(active *theme.Theme, size theme.ProgressBarSize) (unit.Dp, unit.Dp) {
	tokens := active.Components.ProgressBar
	switch size {
	case theme.ProgressBarSmall:
		return tokens.SmallHeight, tokens.SmallRadius
	case theme.ProgressBarLarge:
		return tokens.LargeHeight, tokens.LargeRadius
	default:
		return tokens.MediumHeight, tokens.MediumRadius
	}
}

func circleDiameter(active *theme.Theme, size theme.ProgressBarSize) unit.Dp {
	tokens := active.Components.ProgressCircle
	switch size {
	case theme.ProgressBarSmall:
		return tokens.SmallSize
	case theme.ProgressBarLarge:
		return tokens.LargeSize
	default:
		return tokens.MediumSize
	}
}
