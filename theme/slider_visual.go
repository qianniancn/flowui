package theme

import (
	"time"

	"gioui.org/unit"
	"github.com/qianniancn/flowui/style"
)

// SliderRequest contains the axis and optional instance declaration needed by
// a slider visual provider. Values, focus, and interaction state remain owned
// by slider behavior and are resolved at layout time.
type SliderRequest struct {
	Orientation SliderOrientation
	Override    *SliderOverride
}

// SliderOverride carries caller customization across the neutral provider
// boundary without storing visual state on the behavior widget.
type SliderOverride struct {
	Declaration style.Style
}

// SliderOverrideStyle extracts the provider-owned instance layer.
func SliderOverrideStyle(request SliderRequest) style.Style {
	if request.Override == nil {
		return style.Style{}
	}
	return request.Override.Declaration
}

// SliderMetrics contains neutral geometry consumed by slider layout.
type SliderMetrics struct {
	TrackThickness  unit.Dp
	TrackRadius     unit.Dp
	EdgeInset       unit.Dp
	ThumbLength     unit.Dp
	ThumbExtra      unit.Dp
	HeaderGap       unit.Dp
	VerticalGap     unit.Dp
	DraggingScale   float32
	FocusRingWidth  unit.Dp
	FocusRingOffset unit.Dp
}

// SliderVisual is the complete visual plan consumed by slider behavior.
type SliderVisual struct {
	Defaults         style.Style
	Override         style.Style
	Metrics          SliderMetrics
	Motion           MotionTheme
	FocusDuration    time.Duration
	DraggingDuration time.Duration
}

// SliderProvider supplies part declarations, layout metrics, and motion
// without taking ownership of slider values, input, focus, or semantics.
type SliderProvider interface {
	Slider(active *Theme, request SliderRequest) SliderVisual
}

// ValidateSliderVisual checks the complete slider declaration before a widget
// stores it in its provider cache. Slider layout owns values, hit testing, and
// semantics; every part geometry, paint, typography, metric, and motion input
// must come from the provider bundle.
func ValidateSliderVisual(visual SliderVisual) {
	metrics := visual.Metrics
	if !validPositiveDp(metrics.TrackThickness) {
		invalidVisual("Slider", "Metrics.TrackThickness", "must be finite and greater than 0")
	}
	if !validNonNegativeDp(metrics.TrackRadius) {
		invalidVisual("Slider", "Metrics.TrackRadius", "must be finite and non-negative")
	}
	if !validNonNegativeDp(metrics.EdgeInset) {
		invalidVisual("Slider", "Metrics.EdgeInset", "must be finite and non-negative")
	}
	if !validPositiveDp(metrics.ThumbLength) {
		invalidVisual("Slider", "Metrics.ThumbLength", "must be finite and greater than 0")
	}
	if !validNonNegativeDp(metrics.ThumbExtra) {
		invalidVisual("Slider", "Metrics.ThumbExtra", "must be finite and non-negative")
	}
	if !validNonNegativeDp(metrics.HeaderGap) {
		invalidVisual("Slider", "Metrics.HeaderGap", "must be finite and non-negative")
	}
	if !validNonNegativeDp(metrics.VerticalGap) {
		invalidVisual("Slider", "Metrics.VerticalGap", "must be finite and non-negative")
	}
	if !validFiniteFloat32(metrics.DraggingScale) {
		invalidVisual("Slider", "Metrics.DraggingScale", "must be finite")
	}
	if metrics.DraggingScale <= 0 {
		invalidVisual("Slider", "Metrics.DraggingScale", "must be greater than 0")
	}
	if !validNonNegativeDp(metrics.FocusRingWidth) {
		invalidVisual("Slider", "Metrics.FocusRingWidth", "must be finite and non-negative")
	}
	if !validNonNegativeDp(metrics.FocusRingOffset) {
		invalidVisual("Slider", "Metrics.FocusRingOffset", "must be finite and non-negative")
	}
	if visual.FocusDuration <= 0 {
		invalidVisual("Slider", "FocusDuration", "must be greater than 0")
	}
	if visual.DraggingDuration <= 0 {
		invalidVisual("Slider", "DraggingDuration", "must be greater than 0")
	}

	validateMotion("Slider", "Motion", visual.Motion)

	states := []style.State{
		{},
		{Hovered: true},
		{Pressed: true},
		{Focused: true},
		{FocusVisible: true},
		{Disabled: true},
		{Dragging: true},
		{Disabled: true, FocusVisible: true},
		{Hovered: true, Pressed: true},
	}
	for _, state := range states {
		track := style.CascadePart(state, style.PartTrack, visual.Defaults, visual.Override)
		if track.Box == nil {
			invalidVisual("Slider", "Track.Box", "must be provided")
		}
		if track.Paint == nil {
			invalidVisual("Slider", "Track.Paint", "must be provided")
		}
		if !hasPaint(track.Paint.Background) {
			invalidVisual("Slider", "Track.Paint.Background", "must provide a solid color or nonempty gradient")
		}
		if !validSliderTrackBox(track.Box) {
			invalidVisual("Slider", "Track.Box", "must have positive width and height, minimum dimensions, or fill on each axis")
		}

		fill := style.CascadePart(state, style.PartFill, visual.Defaults, visual.Override)
		if fill.Paint == nil {
			invalidVisual("Slider", "Fill.Paint", "must be provided")
		}
		if !hasPaint(fill.Paint.Background) {
			invalidVisual("Slider", "Fill.Paint.Background", "must provide a solid color or nonempty gradient")
		}

		thumb := style.CascadePart(state, style.PartThumb, visual.Defaults, visual.Override)
		if thumb.Box == nil {
			invalidVisual("Slider", "Thumb.Box", "must be provided")
		}
		if !hasPositiveWidth(thumb.Box.Width, thumb.Box.MinWidth) {
			invalidVisual("Slider", "Thumb.Box.Width", "width or minimum width must be finite and greater than 0")
		}
		if !hasPositiveHeight(thumb.Box.Height, thumb.Box.MinHeight) {
			invalidVisual("Slider", "Thumb.Box.Height", "height or minimum height must be finite and greater than 0")
		}
		if thumb.Paint == nil {
			invalidVisual("Slider", "Thumb.Paint", "must be provided")
		}
		if !hasPaint(thumb.Paint.Background) {
			invalidVisual("Slider", "Thumb.Paint.Background", "must provide a solid color or nonempty gradient")
		}

		label := style.CascadePart(state, style.PartLabel, visual.Defaults, visual.Override)
		if label.Text == nil {
			invalidVisual("Slider", "Label.Text", "must be provided")
		}
		if label.Text.FontSize == nil {
			invalidVisual("Slider", "Label.Text.FontSize", "must be provided")
		}
		if !validPositiveSp(*label.Text.FontSize) {
			invalidVisual("Slider", "Label.Text.FontSize", "must be finite and greater than 0")
		}
		if !hasColor(label.Text.Color) {
			invalidVisual("Slider", "Label.Text.Color", "must resolve to a solid color")
		}
	}
}

func validSliderTrackBox(box *style.BoxStyle) bool {
	horizontal := hasPositiveHeight(box.Height, box.MinHeight) &&
		(hasPositiveWidth(box.Width, box.MinWidth) || hasSliderFill(box.FillWidth))
	vertical := hasPositiveWidth(box.Width, box.MinWidth) &&
		(hasPositiveHeight(box.Height, box.MinHeight) || hasSliderFill(box.FillHeight))
	return horizontal || vertical
}

func hasSliderFill(value *bool) bool {
	return value != nil && *value
}
