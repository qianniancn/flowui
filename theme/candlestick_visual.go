package theme

import (
	"image/color"
	"time"

	"gioui.org/font"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/style"
)

// CandlestickRequest contains the instance-level visual declaration for a
// candlestick chart. OHLC data, scales, and pointer behavior remain widget-owned.
type CandlestickRequest struct {
	Override *CandlestickOverride
}

// CandlestickOverride carries caller customization across the neutral provider
// boundary. The provider owns the returned instance layer.
type CandlestickOverride struct {
	Declaration style.Style
}

// CandlestickOverrideStyle extracts the provider-owned instance layer.
func CandlestickOverrideStyle(request CandlestickRequest) style.Style {
	if request.Override == nil {
		return style.Style{}
	}
	return request.Override.Declaration
}

// CandlestickMetrics contains geometry and typography consumed by candlestick
// layout and drawing. It contains no theme or semantic objects.
type CandlestickMetrics struct {
	AnimationDuration     time.Duration
	UpdateDuration        time.Duration
	Height                unit.Dp
	PlotPaddingTop        unit.Dp
	PlotPaddingRight      unit.Dp
	PlotPaddingBottom     unit.Dp
	PlotPaddingLeft       unit.Dp
	AxisNameGap           unit.Dp
	TickLabelGap          unit.Dp
	AxisTextSize          unit.Sp
	GridWidth             unit.Dp
	AxisWidth             unit.Dp
	AnnotationLineWidth   unit.Dp
	AnnotationPointSize   unit.Dp
	CrosshairWidth        unit.Dp
	CrosshairLabelPadding unit.Dp
	CrosshairLabelRadius  unit.Dp
	WickWidth             unit.Dp
	BorderWidth           unit.Dp
	EmphasisBorderWidth   unit.Dp
	TooltipGap            unit.Dp
	TooltipRowGap         unit.Dp
	TooltipTextSize       unit.Sp
	TextWeight            font.Weight
	EmphasisWeight        font.Weight
}

// CandlestickColors contains all palette decisions used by candlestick
// behavior. Explicit OHLC colors remain widget inputs; the provider owns defaults.
type CandlestickColors struct {
	Axis                     color.NRGBA
	AxisLabel                color.NRGBA
	Grid                     color.NRGBA
	Crosshair                color.NRGBA
	CrosshairLabelBackground color.NRGBA
	CrosshairLabelForeground color.NRGBA
	OverlayForeground        color.NRGBA
	AnnotationArea           color.NRGBA
	AnnotationLine           color.NRGBA
	AnnotationPoint          color.NRGBA
	Up                       color.NRGBA
	Down                     color.NRGBA
	Doji                     color.NRGBA
}

// CandlestickVisual is the complete visual plan consumed by candlestick
// behavior. The root declaration owns disabled opacity and instance styling.
type CandlestickVisual struct {
	Defaults        style.Style
	Override        style.Style
	Metrics         CandlestickMetrics
	Colors          CandlestickColors
	DisabledOpacity float32
	Motion          MotionTheme
}

// CandlestickProvider supplies candlestick metrics and palette without taking
// ownership of OHLC data, scales, selection, or pointer behavior.
type CandlestickProvider interface {
	Candlestick(active *Theme, request CandlestickRequest) CandlestickVisual
}

// ValidateCandlestickVisual checks the geometry, palette, and root declaration
// contract before a candlestick visual enters its provider cache.
func ValidateCandlestickVisual(visual CandlestickVisual) {
	m := visual.Metrics
	if !validPositiveDp(m.Height) {
		invalidVisual("Candlestick", "Metrics.Height", "must be finite and greater than 0")
	}
	if !validNonNegativeDp(m.PlotPaddingTop) {
		invalidVisual("Candlestick", "Metrics.PlotPaddingTop", "must be finite and non-negative")
	}
	if !validNonNegativeDp(m.PlotPaddingRight) {
		invalidVisual("Candlestick", "Metrics.PlotPaddingRight", "must be finite and non-negative")
	}
	if !validNonNegativeDp(m.PlotPaddingBottom) {
		invalidVisual("Candlestick", "Metrics.PlotPaddingBottom", "must be finite and non-negative")
	}
	if !validNonNegativeDp(m.PlotPaddingLeft) {
		invalidVisual("Candlestick", "Metrics.PlotPaddingLeft", "must be finite and non-negative")
	}
	if !validNonNegativeDp(m.AxisNameGap) {
		invalidVisual("Candlestick", "Metrics.AxisNameGap", "must be finite and non-negative")
	}
	if !validNonNegativeDp(m.TickLabelGap) {
		invalidVisual("Candlestick", "Metrics.TickLabelGap", "must be finite and non-negative")
	}
	if !validPositiveSp(m.AxisTextSize) {
		invalidVisual("Candlestick", "Metrics.AxisTextSize", "must be finite and greater than 0")
	}
	if !validPositiveDp(m.GridWidth) {
		invalidVisual("Candlestick", "Metrics.GridWidth", "must be finite and greater than 0")
	}
	if !validPositiveDp(m.AxisWidth) {
		invalidVisual("Candlestick", "Metrics.AxisWidth", "must be finite and greater than 0")
	}
	if !validPositiveDp(m.AnnotationLineWidth) {
		invalidVisual("Candlestick", "Metrics.AnnotationLineWidth", "must be finite and greater than 0")
	}
	if !validPositiveDp(m.AnnotationPointSize) {
		invalidVisual("Candlestick", "Metrics.AnnotationPointSize", "must be finite and greater than 0")
	}
	if !validPositiveDp(m.CrosshairWidth) {
		invalidVisual("Candlestick", "Metrics.CrosshairWidth", "must be finite and greater than 0")
	}
	if !validNonNegativeDp(m.CrosshairLabelPadding) {
		invalidVisual("Candlestick", "Metrics.CrosshairLabelPadding", "must be finite and non-negative")
	}
	if !validNonNegativeDp(m.CrosshairLabelRadius) {
		invalidVisual("Candlestick", "Metrics.CrosshairLabelRadius", "must be finite and non-negative")
	}
	if !validPositiveDp(m.WickWidth) {
		invalidVisual("Candlestick", "Metrics.WickWidth", "must be finite and greater than 0")
	}
	if !validPositiveDp(m.BorderWidth) {
		invalidVisual("Candlestick", "Metrics.BorderWidth", "must be finite and greater than 0")
	}
	if !validPositiveDp(m.EmphasisBorderWidth) {
		invalidVisual("Candlestick", "Metrics.EmphasisBorderWidth", "must be finite and greater than 0")
	}
	if !validNonNegativeDp(m.TooltipGap) {
		invalidVisual("Candlestick", "Metrics.TooltipGap", "must be finite and non-negative")
	}
	if !validNonNegativeDp(m.TooltipRowGap) {
		invalidVisual("Candlestick", "Metrics.TooltipRowGap", "must be finite and non-negative")
	}
	if !validPositiveSp(m.TooltipTextSize) {
		invalidVisual("Candlestick", "Metrics.TooltipTextSize", "must be finite and greater than 0")
	}
	if !hasValidOpacityValue(visual.DisabledOpacity) {
		invalidVisual("Candlestick", "DisabledOpacity", "must be finite and between 0 and 1")
	}

	if visual.Colors.Axis.A == 0 {
		invalidVisual("Candlestick", "Colors.Axis", "must have nonzero alpha")
	}
	if visual.Colors.AxisLabel.A == 0 {
		invalidVisual("Candlestick", "Colors.AxisLabel", "must have nonzero alpha")
	}
	if visual.Colors.Grid.A == 0 {
		invalidVisual("Candlestick", "Colors.Grid", "must have nonzero alpha")
	}
	if visual.Colors.Crosshair.A == 0 {
		invalidVisual("Candlestick", "Colors.Crosshair", "must have nonzero alpha")
	}
	if visual.Colors.CrosshairLabelBackground.A == 0 {
		invalidVisual("Candlestick", "Colors.CrosshairLabelBackground", "must have nonzero alpha")
	}
	if visual.Colors.CrosshairLabelForeground.A == 0 {
		invalidVisual("Candlestick", "Colors.CrosshairLabelForeground", "must have nonzero alpha")
	}
	if visual.Colors.OverlayForeground.A == 0 {
		invalidVisual("Candlestick", "Colors.OverlayForeground", "must have nonzero alpha")
	}
	if visual.Colors.AnnotationArea.A == 0 {
		invalidVisual("Candlestick", "Colors.AnnotationArea", "must have nonzero alpha")
	}
	if visual.Colors.AnnotationLine.A == 0 {
		invalidVisual("Candlestick", "Colors.AnnotationLine", "must have nonzero alpha")
	}
	if visual.Colors.AnnotationPoint.A == 0 {
		invalidVisual("Candlestick", "Colors.AnnotationPoint", "must have nonzero alpha")
	}
	if visual.Colors.Up.A == 0 {
		invalidVisual("Candlestick", "Colors.Up", "must have nonzero alpha")
	}
	if visual.Colors.Down.A == 0 {
		invalidVisual("Candlestick", "Colors.Down", "must have nonzero alpha")
	}
	root := style.Cascade(style.State{}, visual.Defaults, visual.Override)
	if root.Box == nil {
		invalidVisual("Candlestick", "Root.Box", "must be provided")
	}
	if root.Paint == nil {
		invalidVisual("Candlestick", "Root.Paint", "must be provided")
	}
	if !hasValidOpacity(root.Paint.Opacity) {
		invalidVisual("Candlestick", "Root.Paint.Opacity", "must be finite and between 0 and 1")
	}
}
