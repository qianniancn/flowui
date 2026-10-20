package theme

import (
	"image/color"
	"math"
	"time"

	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/style"
)

// ProgressBarRequest contains the semantic axes that affect a linear
// progress bar's visual recipe. Value, indeterminate state, and disabled state
// remain behavior inputs resolved by the widget.
type ProgressBarRequest struct {
	Color    ProgressBarColor
	Size     ProgressBarSize
	Override *ProgressOverride
}

// ProgressCircleRequest contains the semantic axes that affect a circular
// progress indicator's visual recipe.
type ProgressCircleRequest struct {
	Color    ProgressBarColor
	Size     ProgressBarSize
	Override *ProgressOverride
}

// ProgressOverride carries an instance declaration across the provider
// boundary. The progress behavior itself does not retain visual state.
type ProgressOverride struct {
	Declaration style.Style
}

// ProgressOverrideStyle extracts the provider-owned instance layer.
func ProgressOverrideStyle(override *ProgressOverride) style.Style {
	if override != nil {
		return override.Declaration
	}
	return style.Style{}
}

// ProgressBarMetrics contains neutral geometry and timing consumed by the
// linear progress behavior layout.
type ProgressBarMetrics struct {
	HeaderGap             unit.Dp
	ValueDuration         time.Duration
	IndeterminatePeriod   time.Duration
	IndeterminateFillRate float32
}

// ProgressColorResolver interprets a provider-owned Track or Fill declaration
// for drawing paths that need a concrete color. The widget supplies the
// resolved Part and never projects a brush or opacity itself.
type ProgressColorResolver func(style.ResolvedStyle) (color.NRGBA, bool)

// ProgressBarVisual is the complete visual plan consumed by BarWidget.
type ProgressBarVisual struct {
	Defaults style.Style
	Variant  style.Style
	Size     style.Style
	Override style.Style
	Color    ProgressColorResolver
	Metrics  ProgressBarMetrics
	Motion   MotionTheme
}

// ProgressCircleMetrics contains neutral geometry and timing consumed by the
// circular progress behavior layout.
type ProgressCircleMetrics struct {
	StrokeRatio   float32
	ValueDuration time.Duration
	SpinDuration  time.Duration
}

// ProgressCircleRenderRequest contains behavior-derived progress inputs and
// provider-owned geometry needed to paint one circle frame. The widget owns
// value state and invalidation timing; the provider owns ring and arc drawing.
type ProgressCircleRenderRequest struct {
	Size          unit.Dp
	StrokeRatio   float32
	Track         color.NRGBA
	Fill          color.NRGBA
	Progress      float32
	Indeterminate bool
	Period        time.Duration
}

// ProgressCircleRenderer is a theme-owned circular progress drawing strategy.
type ProgressCircleRenderer func(layout.Context, ProgressCircleRenderRequest) layout.Dimensions

// ProgressCircleVisual is the complete visual plan consumed by CircleWidget.
type ProgressCircleVisual struct {
	Defaults style.Style
	Override style.Style
	Color    ProgressColorResolver
	Renderer ProgressCircleRenderer
	Metrics  ProgressCircleMetrics
	Motion   MotionTheme
}

// ProgressProvider supplies progress declarations, metrics, and motion policy
// without taking ownership of values, animation state, or semantics.
type ProgressProvider interface {
	ProgressBar(active *Theme, request ProgressBarRequest) ProgressBarVisual
	ProgressCircle(active *Theme, request ProgressCircleRequest) ProgressCircleVisual
}

// ValidateProgressBarVisual checks the linear progress Parts, concrete color
// resolver, metrics, and motion before a widget stores the result in its
// declaration cache. Progress value and indeterminate state remain behavior
// inputs; the provider owns every visual declaration they feed.
func ValidateProgressBarVisual(visual ProgressBarVisual) {
	metrics := visual.Metrics
	if visual.Color == nil {
		invalidVisual("ProgressBar", "Color", "must be provided")
	}
	if metrics.HeaderGap < 0 {
		invalidVisual("ProgressBar", "Metrics.HeaderGap", "must be at least 0")
	}
	if metrics.ValueDuration <= 0 {
		invalidVisual("ProgressBar", "Metrics.ValueDuration", "must be greater than 0")
	}
	if metrics.IndeterminatePeriod <= 0 {
		invalidVisual("ProgressBar", "Metrics.IndeterminatePeriod", "must be greater than 0")
	}
	if metrics.IndeterminateFillRate <= 0 {
		invalidVisual("ProgressBar", "Metrics.IndeterminateFillRate", "must be greater than 0")
	}
	if metrics.IndeterminateFillRate >= 1 {
		invalidVisual("ProgressBar", "Metrics.IndeterminateFillRate", "must be less than 1")
	}

	validateMotion("ProgressBar", "Motion", visual.Motion)
	for _, state := range []style.State{{}, {Disabled: true}} {
		track := style.CascadePart(state, style.PartTrack, visual.Defaults, visual.Variant, visual.Size, visual.Override)
		if track.Box == nil {
			invalidVisual("ProgressBar", "Track.Box", "must be provided")
		}
		if !hasPositiveHeight(track.Box.Height, track.Box.MinHeight) {
			invalidVisual("ProgressBar", "Track.Box.Height", "height or minimum height must be finite and greater than 0")
		}
		if track.Paint == nil {
			invalidVisual("ProgressBar", "Track.Paint", "must be provided")
		}
		if !hasPaint(track.Paint.Background) {
			invalidVisual("ProgressBar", "Track.Paint.Background", "must provide a solid color or nonempty gradient")
		}

		if _, ok := visual.Color(track); !ok {
			invalidVisual("ProgressBar", "Color", "must resolve the declaration color")
		}

		fill := style.CascadePart(state, style.PartFill, visual.Defaults, visual.Variant, visual.Size, visual.Override)
		if fill.Paint == nil {
			invalidVisual("ProgressBar", "Fill.Paint", "must be provided")
		}
		if !hasPaint(fill.Paint.Background) {
			invalidVisual("ProgressBar", "Fill.Paint.Background", "must provide a solid color or nonempty gradient")
		}

		if _, ok := visual.Color(fill); !ok {
			invalidVisual("ProgressBar", "Color", "must resolve the declaration color")
		}

		label := style.CascadePart(state, style.PartLabel, visual.Defaults, visual.Variant, visual.Size, visual.Override)
		if label.Text == nil {
			invalidVisual("ProgressBar", "Label.Text", "must be provided")
		}
		if label.Text.FontSize == nil {
			invalidVisual("ProgressBar", "Label.Text.FontSize", "must be provided")
		}
		if *label.Text.FontSize <= 0 {
			invalidVisual("ProgressBar", "Label.Text.FontSize", "must be greater than 0")
		}
		if math.IsNaN(float64(*label.Text.FontSize)) {
			invalidVisual("ProgressBar", "Label.Text.FontSize", "must be finite")
		}
		if math.IsInf(float64(*label.Text.FontSize), 0) {
			invalidVisual("ProgressBar", "Label.Text.FontSize", "must be finite")
		}
		if !hasColor(label.Text.Color) {
			invalidVisual("ProgressBar", "Label.Text.Color", "must resolve to a solid color")
		}
	}
}

// ValidateProgressCircleVisual checks the circular progress Parts, concrete
// color resolver, metrics, and motion before cache insertion.
func ValidateProgressCircleVisual(visual ProgressCircleVisual) {
	metrics := visual.Metrics
	if visual.Color == nil {
		invalidVisual("ProgressCircle", "Color", "must be provided")
	}
	if visual.Renderer == nil {
		invalidVisual("ProgressCircle", "Renderer", "must be provided")
	}
	if metrics.StrokeRatio <= 0 {
		invalidVisual("ProgressCircle", "Metrics.StrokeRatio", "must be greater than 0")
	}
	if metrics.StrokeRatio >= .5 {
		invalidVisual("ProgressCircle", "Metrics.StrokeRatio", "must be less than .5")
	}
	if metrics.ValueDuration <= 0 {
		invalidVisual("ProgressCircle", "Metrics.ValueDuration", "must be greater than 0")
	}
	if metrics.SpinDuration <= 0 {
		invalidVisual("ProgressCircle", "Metrics.SpinDuration", "must be greater than 0")
	}

	validateMotion("ProgressCircle", "Motion", visual.Motion)
	for _, state := range []style.State{{}, {Disabled: true}} {
		root := style.Cascade(state, visual.Defaults, visual.Override)
		if root.Box == nil {
			invalidVisual("ProgressCircle", "Root.Box", "must be provided")
		}
		if !hasPositiveWidth(root.Box.Width, root.Box.MinWidth) {
			invalidVisual("ProgressCircle", "Root.Box.Width", "width or minimum width must be finite and greater than 0")
		}
		if !hasPositiveHeight(root.Box.Height, root.Box.MinHeight) {
			invalidVisual("ProgressCircle", "Root.Box.Height", "height or minimum height must be finite and greater than 0")
		}

		for _, part := range []style.Part{style.PartTrack, style.PartFill} {
			resolved := style.CascadePart(state, part, visual.Defaults, visual.Override)
			if resolved.Paint == nil {
				invalidVisual("ProgressCircle", "Resolved.Paint", "must be provided")
			}
			if !hasPaint(resolved.Paint.Background) {
				invalidVisual("ProgressCircle", "Resolved.Paint.Background", "must provide a solid color or nonempty gradient")
			}

			if _, ok := visual.Color(resolved); !ok {
				invalidVisual("ProgressCircle", "Color", "must resolve the declaration color")
			}
		}
	}
}
