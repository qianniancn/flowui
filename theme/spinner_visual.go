package theme

import (
	"image/color"
	"math"
	"time"

	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/style"
)

// SpinnerRequest contains the semantic axes that affect a standalone
// spinner's visual recipe. Loading state and accessible label remain behavior
// inputs owned by the widget.
type SpinnerRequest struct {
	Color    SpinnerColor
	Size     SpinnerSize
	Override *SpinnerOverride
}

// SpinnerOverride carries an instance declaration across the neutral provider
// boundary. Spinner behavior itself does not retain visual state.
type SpinnerOverride struct {
	Declaration style.Style
}

// SpinnerOverrideStyle extracts the provider-owned instance layer.
func SpinnerOverrideStyle(override *SpinnerOverride) style.Style {
	if override != nil {
		return override.Declaration
	}
	return style.Style{}
}

// SpinnerMetrics contains neutral geometry and timing consumed by spinner
// behavior. The provider owns the visual ratios and base animation period.
type SpinnerMetrics struct {
	StrokeRatio  float32
	InsetRatio   float32
	SpinDuration time.Duration
}

// SpinnerRenderRequest contains the provider-owned geometry and motion values
// needed to paint one standalone spinner frame. The widget decides when the
// indicator is shown and when to invalidate; the provider decides how it is
// drawn.
type SpinnerRenderRequest struct {
	Size        unit.Dp
	StrokeRatio float32
	InsetRatio  float32
	Color       color.NRGBA
	Period      time.Duration
}

// SpinnerRenderer is a theme-owned spinner drawing strategy. Keeping the
// renderer in the neutral protocol lets a theme replace the indicator shape
// without making widget/spinner import or implement a visual recipe.
type SpinnerRenderer func(layout.Context, SpinnerRenderRequest) layout.Dimensions

// SpinnerColorResolver interprets the provider-owned Indicator declaration
// for the renderer. The widget supplies the resolved Part and does not project
// a theme color itself.
type SpinnerColorResolver func(style.ResolvedStyle) (color.NRGBA, bool)

// SpinnerVisual is the complete visual plan consumed by standalone spinner
// behavior.
type SpinnerVisual struct {
	Defaults style.Style
	Override style.Style
	Color    SpinnerColorResolver
	Renderer SpinnerRenderer
	Metrics  SpinnerMetrics
	Motion   MotionTheme
}

// SpinnerProvider supplies standalone spinner declarations, metrics, and
// motion policy without taking ownership of loading state or semantics.
type SpinnerProvider interface {
	Spinner(active *Theme, request SpinnerRequest) SpinnerVisual
}

// ValidateSpinnerVisual checks the standalone spinner declaration, indicator
// color resolver, geometry ratios, and motion before cache insertion.
func ValidateSpinnerVisual(visual SpinnerVisual) {
	metrics := visual.Metrics
	invalidStroke := math.IsNaN(float64(metrics.StrokeRatio)) || math.IsInf(float64(metrics.StrokeRatio), 0)
	invalidInset := math.IsNaN(float64(metrics.InsetRatio)) || math.IsInf(float64(metrics.InsetRatio), 0)
	if visual.Color == nil {
		invalidVisual("Spinner", "Color", "must be provided")
	}
	if visual.Renderer == nil {
		invalidVisual("Spinner", "Renderer", "must be provided")
	}
	if invalidStroke {
		invalidVisual("Spinner", "Metrics.StrokeRatio", "must be finite")
	}
	if metrics.StrokeRatio <= 0 {
		invalidVisual("Spinner", "Metrics.StrokeRatio", "must be greater than 0")
	}
	if metrics.StrokeRatio >= .5 {
		invalidVisual("Spinner", "Metrics.StrokeRatio", "must be less than .5")
	}
	if invalidInset {
		invalidVisual("Spinner", "Metrics.InsetRatio", "must be finite")
	}
	if metrics.InsetRatio < 0 {
		invalidVisual("Spinner", "Metrics.InsetRatio", "must be at least 0")
	}
	if metrics.InsetRatio >= .5 {
		invalidVisual("Spinner", "Metrics.InsetRatio", "must be less than .5")
	}
	if metrics.SpinDuration <= 0 {
		invalidVisual("Spinner", "Metrics.SpinDuration", "must be greater than 0")
	}

	validateMotion("Spinner", "Motion", visual.Motion)
	root := style.Cascade(style.State{}, visual.Defaults, visual.Override)
	if root.Box == nil {
		invalidVisual("Spinner", "Root.Box", "must be provided")
	}
	if !hasPositiveWidth(root.Box.Width, root.Box.MinWidth) {
		invalidVisual("Spinner", "Root.Box.Width", "width or minimum width must be finite and greater than 0")
	}
	if !hasPositiveHeight(root.Box.Height, root.Box.MinHeight) {
		invalidVisual("Spinner", "Root.Box.Height", "height or minimum height must be finite and greater than 0")
	}

	indicator := style.CascadePart(style.State{}, style.PartIndicator, visual.Defaults, visual.Override)
	if indicator.Text == nil {
		invalidVisual("Spinner", "Indicator.Text", "must be provided")
	}
	if !hasColor(indicator.Text.Color) {
		invalidVisual("Spinner", "Indicator.Text.Color", "must resolve to a solid color")
	}

	if _, ok := visual.Color(indicator); !ok {
		invalidVisual("Spinner", "Color", "must resolve the declaration color")
	}
}
