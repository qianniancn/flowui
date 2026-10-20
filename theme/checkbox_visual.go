package theme

import (
	"image/color"
	"math"
	"time"

	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/style"
)

// CheckboxRequest describes the behavior-owned semantic variant needed by a
// checkbox visual provider. Interaction state is supplied when declarations
// are resolved by the widget.
type CheckboxRequest struct {
	Variant  CheckboxVariant
	Override *CheckboxOverride
}

// CheckboxOverride carries caller customization across the neutral provider
// boundary. The provider owns the returned instance layer.
type CheckboxOverride struct {
	Declaration style.Style
}

// CheckboxOverrideStyle extracts the provider-owned instance layer.
func CheckboxOverrideStyle(request CheckboxRequest) style.Style {
	if request.Override == nil {
		return style.Style{}
	}
	return request.Override.Declaration
}

// CheckboxMetrics contains geometry and spacing used by checkbox behavior.
type CheckboxMetrics struct {
	FocusSpace          unit.Dp
	IndicatorSize       unit.Dp
	IndeterminateStroke unit.Dp
	CheckStroke         unit.Dp
	DescriptionGap      unit.Dp
	DescriptionIndent   unit.Dp
	LabelGap            unit.Dp
	RequiredMarkOffset  unit.Dp
}

// CheckboxColors contains provider-owned colors needed when a semantic part
// is rendered without an explicit declaration.
type CheckboxColors struct {
	RequiredMark         color.NRGBA
	DisabledRequiredMark color.NRGBA
}

// CheckboxIndicatorRenderRequest contains only the resolved state and
// geometry needed to draw the default checkmark or indeterminate bar.
// Indicator surface, border, text color, and focus treatment remain resolved
// through the Checkbox Parts before this renderer is called.
type CheckboxIndicatorRenderRequest struct {
	Foreground          color.NRGBA
	Selection           float32
	Indeterminate       bool
	IndeterminateStroke unit.Dp
	CheckStroke         unit.Dp
}

// CheckboxIndicatorRenderer is the theme-owned strategy for drawing the
// default selected indicator inside an already-resolved Checkbox Part.
type CheckboxIndicatorRenderer func(layout.Context, CheckboxIndicatorRenderRequest) layout.Dimensions

// CheckboxVisual is the complete visual plan consumed by checkbox behavior.
type CheckboxVisual struct {
	Defaults style.Style
	// Variant follows inherited text and precedes scope and instance styles.
	Variant           style.Style
	Override          style.Style
	Metrics           CheckboxMetrics
	Colors            CheckboxColors
	IndicatorRenderer CheckboxIndicatorRenderer
	Motion            MotionTheme
	SelectionDuration time.Duration
	FocusDuration     time.Duration
}

// CheckboxProvider supplies checkbox declarations, geometry, colors, and
// motion timing without taking ownership of checked state or semantics.
type CheckboxProvider interface {
	Checkbox(active *Theme, request CheckboxRequest) CheckboxVisual
}

// ValidateCheckboxVisual checks every checkbox Part and the geometry/motion
// values before a widget stores a provider result in its visual cache.
// Checkbox behavior owns selection, focus, and semantics; the provider must
// still declare every state used by the indicator and supporting text.
func ValidateCheckboxVisual(visual CheckboxVisual) {
	metrics := visual.Metrics
	if metrics.FocusSpace < 0 {
		invalidVisual("Checkbox", "Metrics.FocusSpace", "must be at least 0")
	}
	if metrics.IndicatorSize <= 0 {
		invalidVisual("Checkbox", "Metrics.IndicatorSize", "must be greater than 0")
	}
	if metrics.IndeterminateStroke <= 0 {
		invalidVisual("Checkbox", "Metrics.IndeterminateStroke", "must be greater than 0")
	}
	if metrics.CheckStroke <= 0 {
		invalidVisual("Checkbox", "Metrics.CheckStroke", "must be greater than 0")
	}
	if metrics.DescriptionGap < 0 {
		invalidVisual("Checkbox", "Metrics.DescriptionGap", "must be at least 0")
	}
	if metrics.DescriptionIndent < 0 {
		invalidVisual("Checkbox", "Metrics.DescriptionIndent", "must be at least 0")
	}
	if metrics.LabelGap < 0 {
		invalidVisual("Checkbox", "Metrics.LabelGap", "must be at least 0")
	}
	if metrics.RequiredMarkOffset < 0 {
		invalidVisual("Checkbox", "Metrics.RequiredMarkOffset", "must be at least 0")
	}
	if visual.SelectionDuration <= 0 {
		invalidVisual("Checkbox", "SelectionDuration", "must be greater than 0")
	}
	if visual.FocusDuration <= 0 {
		invalidVisual("Checkbox", "FocusDuration", "must be greater than 0")
	}
	if visual.IndicatorRenderer == nil {
		invalidVisual("Checkbox", "IndicatorRenderer", "must be provided")
	}

	validateMotion("Checkbox", "Motion", visual.Motion)

	resolve := func(part style.Part, state style.State) style.ResolvedStyle {
		return style.CascadePart(state, part, visual.Defaults, visual.Variant, visual.Override)
	}
	states := []style.State{
		{},
		{Hovered: true},
		{Pressed: true},
		{FocusVisible: true},
		{Checked: true, Selected: true},
		{Indeterminate: true, Selected: true},
		{Invalid: true},
		{Disabled: true},
		{Disabled: true, Invalid: true},
		{Checked: true, Selected: true, Hovered: true},
		{Checked: true, Selected: true, Disabled: true},
	}
	for _, state := range states {
		indicator := resolve(style.PartIndicator, state)
		if indicator.Box == nil {
			invalidVisual("Checkbox", "Indicator.Box", "must be provided")
		}
		if !hasPositiveWidth(indicator.Box.Width, indicator.Box.MinWidth) {
			invalidVisual("Checkbox", "Indicator.Box.Width", "width or minimum width must be finite and greater than 0")
		}
		if !hasPositiveHeight(indicator.Box.Height, indicator.Box.MinHeight) {
			invalidVisual("Checkbox", "Indicator.Box.Height", "height or minimum height must be finite and greater than 0")
		}
		if indicator.Paint == nil {
			invalidVisual("Checkbox", "Indicator.Paint", "must be provided")
		}
		if !hasPaint(indicator.Paint.Background) {
			invalidVisual("Checkbox", "Indicator.Paint.Background", "must provide a solid color or nonempty gradient")
		}
		if indicator.Paint.Border == nil {
			invalidVisual("Checkbox", "Indicator.Paint.Border", "must be provided")
		}
		if indicator.Paint.Border.Width == nil {
			invalidVisual("Checkbox", "Indicator.Paint.Border.Width", "must be provided")
		}
		if *indicator.Paint.Border.Width < 0 {
			invalidVisual("Checkbox", "Indicator.Paint.Border.Width", "must be at least 0")
		}
		if indicator.Paint.Radius == nil {
			invalidVisual("Checkbox", "Indicator.Paint.Radius", "must be provided")
		}
		if *indicator.Paint.Radius < 0 {
			invalidVisual("Checkbox", "Indicator.Paint.Radius", "must be at least 0")
		}
		if indicator.Text == nil {
			invalidVisual("Checkbox", "Indicator.Text", "must be provided")
		}
		if !hasColor(indicator.Text.Color) {
			invalidVisual("Checkbox", "Indicator.Text.Color", "must resolve to a solid color")
		}

		label := resolve(style.PartLabel, state)
		if label.Text == nil {
			invalidVisual("Checkbox", "Label.Text", "must be provided")
		}
		if label.Text.FontSize == nil {
			invalidVisual("Checkbox", "Label.Text.FontSize", "must be provided")
		}
		if *label.Text.FontSize <= 0 {
			invalidVisual("Checkbox", "Label.Text.FontSize", "must be greater than 0")
		}
		if math.IsNaN(float64(*label.Text.FontSize)) {
			invalidVisual("Checkbox", "Label.Text.FontSize", "must be finite")
		}
		if math.IsInf(float64(*label.Text.FontSize), 0) {
			invalidVisual("Checkbox", "Label.Text.FontSize", "must be finite")
		}
		if !hasColor(label.Text.Color) {
			invalidVisual("Checkbox", "Label.Text.Color", "must resolve to a solid color")
		}

		description := resolve(style.PartDescription, state)
		if description.Text == nil {
			invalidVisual("Checkbox", "Description.Text", "must be provided")
		}
		if !hasColor(description.Text.Color) {
			invalidVisual("Checkbox", "Description.Text.Color", "must resolve to a solid color")
		}
	}

	focus := resolve(style.PartIndicator, style.State{FocusVisible: true})
	if focus.Paint == nil {
		invalidVisual("Checkbox", "Focus.Paint", "must be provided")
	}
	if focus.Paint.Outline == nil {
		invalidVisual("Checkbox", "Focus.Paint.Outline", "must be provided")
	}
	if focus.Paint.Outline.Width <= 0 {
		invalidVisual("Checkbox", "Focus.Paint.Outline.Width", "must be greater than 0")
	}
	if !hasColor(focus.Paint.Outline.Color) {
		invalidVisual("Checkbox", "Focus.Paint.Outline.Color", "must resolve to a solid color")
	}
}
