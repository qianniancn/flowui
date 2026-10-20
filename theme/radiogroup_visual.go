package theme

import (
	"image"
	"image/color"
	"math"
	"time"

	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/style"
)

// RadioIndicatorRenderRequest contains the behavior-derived selection and
// press progress plus the provider-resolved inputs needed to paint the dot.
// The widget owns the state transition; the provider owns the indicator shape.
type RadioIndicatorRenderRequest struct {
	Size            image.Point
	Foreground      color.NRGBA
	Selection       float32
	Pressed         bool
	DotScale        float32
	DotPressedScale float32
}

// RadioIndicatorRenderer is a theme-owned selected-indicator drawing strategy.
type RadioIndicatorRenderer func(layout.Context, RadioIndicatorRenderRequest) layout.Dimensions

// RadioGroupRequest describes the semantic variant and optional instance
// declaration needed by a radio-group visual provider. Selection, focus and
// disabled state are supplied later when the widget resolves Parts.
type RadioGroupRequest struct {
	Variant  RadioGroupVariant
	Override *RadioGroupOverride
}

// RadioGroupOverride carries caller customization across the neutral
// provider boundary without storing visual state on the behavior widget.
type RadioGroupOverride struct {
	Declaration style.Style
}

// RadioGroupOverrideStyle extracts the provider-owned instance layer.
func RadioGroupOverrideStyle(request RadioGroupRequest) style.Style {
	if request.Override == nil {
		return style.Style{}
	}
	return request.Override.Declaration
}

// RadioGroupMetrics contains geometry and spacing used by radio behavior.
type RadioGroupMetrics struct {
	FocusSpace      unit.Dp
	ContentGap      unit.Dp
	VerticalGap     unit.Dp
	HorizontalGap   unit.Dp
	DescriptionGap  unit.Dp
	DotScale        float32
	DotPressedScale float32
	PressedScale    float32
}

// RadioGroupIndicatorColorResolver interprets the provider-owned selected
// indicator declaration. The widget supplies the resolved Part and never
// derives a color or falls back to a theme token itself.
type RadioGroupIndicatorColorResolver func(style.ResolvedStyle) (color.NRGBA, bool)

// RadioGroupVisual is the complete visual plan consumed by radio behavior.
type RadioGroupVisual struct {
	Defaults          style.Style
	Override          style.Style
	Metrics           RadioGroupMetrics
	IndicatorColor    RadioGroupIndicatorColorResolver
	IndicatorRenderer RadioIndicatorRenderer
	Motion            MotionTheme
	SelectionDuration time.Duration
	FocusDuration     time.Duration
	PressInDuration   time.Duration
	PressOutDuration  time.Duration
}

// RadioGroupProvider supplies radio declarations, geometry, colors, and
// motion timing without taking ownership of selection or keyboard behavior.
type RadioGroupProvider interface {
	RadioGroup(active *Theme, request RadioGroupRequest) RadioGroupVisual
}

// ValidateRadioGroupVisual checks every RadioGroup Part and the interaction
// metrics before a widget stores a provider result in its visual cache.
// Selection, focus, keyboard navigation, and semantics remain owned by the
// widget; the provider must supply the complete indicator/text contract.
func ValidateRadioGroupVisual(visual RadioGroupVisual) {
	metrics := visual.Metrics
	if metrics.FocusSpace < 0 {
		invalidVisual("RadioGroup", "Metrics.FocusSpace", "must be at least 0")
	}
	if metrics.ContentGap < 0 {
		invalidVisual("RadioGroup", "Metrics.ContentGap", "must be at least 0")
	}
	if metrics.VerticalGap < 0 {
		invalidVisual("RadioGroup", "Metrics.VerticalGap", "must be at least 0")
	}
	if metrics.HorizontalGap < 0 {
		invalidVisual("RadioGroup", "Metrics.HorizontalGap", "must be at least 0")
	}
	if metrics.DescriptionGap < 0 {
		invalidVisual("RadioGroup", "Metrics.DescriptionGap", "must be at least 0")
	}
	if metrics.DotScale <= 0 {
		invalidVisual("RadioGroup", "Metrics.DotScale", "must be greater than 0")
	}
	if metrics.DotScale > 1 {
		invalidVisual("RadioGroup", "Metrics.DotScale", "must be at most 1")
	}
	if metrics.DotPressedScale <= 0 {
		invalidVisual("RadioGroup", "Metrics.DotPressedScale", "must be greater than 0")
	}
	if metrics.DotPressedScale > 1 {
		invalidVisual("RadioGroup", "Metrics.DotPressedScale", "must be at most 1")
	}
	if metrics.PressedScale <= 0 {
		invalidVisual("RadioGroup", "Metrics.PressedScale", "must be greater than 0")
	}
	if metrics.PressedScale > 1 {
		invalidVisual("RadioGroup", "Metrics.PressedScale", "must be at most 1")
	}
	if visual.SelectionDuration <= 0 {
		invalidVisual("RadioGroup", "SelectionDuration", "must be greater than 0")
	}
	if visual.FocusDuration <= 0 {
		invalidVisual("RadioGroup", "FocusDuration", "must be greater than 0")
	}
	if visual.PressInDuration <= 0 {
		invalidVisual("RadioGroup", "PressInDuration", "must be greater than 0")
	}
	if visual.PressOutDuration <= 0 {
		invalidVisual("RadioGroup", "PressOutDuration", "must be greater than 0")
	}
	if visual.IndicatorColor == nil {
		invalidVisual("RadioGroup", "IndicatorColor", "must be provided")
	}
	if visual.IndicatorRenderer == nil {
		invalidVisual("RadioGroup", "IndicatorRenderer", "must be provided")
	}

	validateMotion("RadioGroup", "Motion", visual.Motion)

	resolve := func(part style.Part, state style.State) style.ResolvedStyle {
		return style.CascadePart(state, part, visual.Defaults, visual.Override)
	}
	states := []style.State{
		{},
		{Hovered: true},
		{Pressed: true},
		{Focused: true},
		{FocusVisible: true},
		{Selected: true, Checked: true},
		{Invalid: true},
		{Disabled: true},
		{Disabled: true, Invalid: true},
		{Selected: true, Checked: true, Hovered: true},
		{Selected: true, Checked: true, Pressed: true},
		{Selected: true, Checked: true, Disabled: true},
	}
	for _, state := range states {
		indicator := resolve(style.PartIndicator, state)
		if indicator.Box == nil {
			invalidVisual("RadioGroup", "Indicator.Box", "must be provided")
		}
		if !hasPositiveWidth(indicator.Box.Width, indicator.Box.MinWidth) {
			invalidVisual("RadioGroup", "Indicator.Box.Width", "width or minimum width must be finite and greater than 0")
		}
		if !hasPositiveHeight(indicator.Box.Height, indicator.Box.MinHeight) {
			invalidVisual("RadioGroup", "Indicator.Box.Height", "height or minimum height must be finite and greater than 0")
		}
		if indicator.Paint == nil {
			invalidVisual("RadioGroup", "Indicator.Paint", "must be provided")
		}
		if !hasPaint(indicator.Paint.Background) {
			invalidVisual("RadioGroup", "Indicator.Paint.Background", "must provide a solid color or nonempty gradient")
		}
		if indicator.Paint.Border == nil {
			invalidVisual("RadioGroup", "Indicator.Paint.Border", "must be provided")
		}
		if indicator.Paint.Border.Width == nil {
			invalidVisual("RadioGroup", "Indicator.Paint.Border.Width", "must be provided")
		}
		if *indicator.Paint.Border.Width < 0 {
			invalidVisual("RadioGroup", "Indicator.Paint.Border.Width", "must be at least 0")
		}
		if indicator.Paint.Radius == nil {
			invalidVisual("RadioGroup", "Indicator.Paint.Radius", "must be provided")
		}
		if *indicator.Paint.Radius < 0 {
			invalidVisual("RadioGroup", "Indicator.Paint.Radius", "must be at least 0")
		}
		if indicator.Text == nil {
			invalidVisual("RadioGroup", "Indicator.Text", "must be provided")
		}
		if !hasColor(indicator.Text.Color) {
			invalidVisual("RadioGroup", "Indicator.Text.Color", "must resolve to a solid color")
		}

		if _, ok := visual.IndicatorColor(indicator); !ok {
			invalidVisual("RadioGroup", "IndicatorColor", "must resolve the declaration color")
		}

		label := resolve(style.PartLabel, state)
		if label.Text == nil {
			invalidVisual("RadioGroup", "Label.Text", "must be provided")
		}
		if label.Text.FontSize == nil {
			invalidVisual("RadioGroup", "Label.Text.FontSize", "must be provided")
		}
		if *label.Text.FontSize <= 0 {
			invalidVisual("RadioGroup", "Label.Text.FontSize", "must be greater than 0")
		}
		if math.IsNaN(float64(*label.Text.FontSize)) {
			invalidVisual("RadioGroup", "Label.Text.FontSize", "must be finite")
		}
		if math.IsInf(float64(*label.Text.FontSize), 0) {
			invalidVisual("RadioGroup", "Label.Text.FontSize", "must be finite")
		}
		if !hasColor(label.Text.Color) {
			invalidVisual("RadioGroup", "Label.Text.Color", "must resolve to a solid color")
		}

		description := resolve(style.PartDescription, state)
		if description.Text == nil {
			invalidVisual("RadioGroup", "Description.Text", "must be provided")
		}
		if description.Text.FontSize == nil {
			invalidVisual("RadioGroup", "Description.Text.FontSize", "must be provided")
		}
		if *description.Text.FontSize <= 0 {
			invalidVisual("RadioGroup", "Description.Text.FontSize", "must be greater than 0")
		}
		if math.IsNaN(float64(*description.Text.FontSize)) {
			invalidVisual("RadioGroup", "Description.Text.FontSize", "must be finite")
		}
		if math.IsInf(float64(*description.Text.FontSize), 0) {
			invalidVisual("RadioGroup", "Description.Text.FontSize", "must be finite")
		}
		if !hasColor(description.Text.Color) {
			invalidVisual("RadioGroup", "Description.Text.Color", "must resolve to a solid color")
		}
	}

	focus := resolve(style.PartIndicator, style.State{FocusVisible: true})
	if focus.Paint == nil {
		invalidVisual("RadioGroup", "Focus.Paint", "must be provided")
	}
	if focus.Paint.Outline == nil {
		invalidVisual("RadioGroup", "Focus.Paint.Outline", "must be provided")
	}
	if focus.Paint.Outline.Width <= 0 {
		invalidVisual("RadioGroup", "Focus.Paint.Outline.Width", "must be greater than 0")
	}
	if !hasColor(focus.Paint.Outline.Color) {
		invalidVisual("RadioGroup", "Focus.Paint.Outline.Color", "must resolve to a solid color")
	}
}
