package theme

import (
	"image/color"
	"math"

	"github.com/qianniancn/flowui/style"
)

// ToggleRequest describes the behavior-owned visual inputs for a toggle
// button. Interaction state is supplied by the widget when it resolves the
// returned declaration layers.
type ToggleRequest struct {
	Variant    ToggleButtonVariant
	Size       ToggleButtonSize
	IconOnly   bool
	Foreground color.NRGBA
	Override   *ToggleOverride
}

// ToggleOverride carries caller customization across the neutral provider
// boundary. The provider owns the returned instance layer.
type ToggleOverride struct {
	Declaration style.Style
}

// ToggleOverrideStyle extracts the provider-owned instance layer.
func ToggleOverrideStyle(request ToggleRequest) style.Style {
	if request.Override == nil {
		return style.Style{}
	}
	return request.Override.Declaration
}

// ToggleVisual is the complete visual plan consumed by toggle button
// behavior. Theme packages own all declarations and transition rules.
type ToggleVisual struct {
	Defaults style.Style
	Variant  style.Style
	Size     style.Style
	Override style.Style
}

// ToggleProvider supplies toggle button declarations and neutral layout
// metrics without taking ownership of selection or input behavior.
type ToggleProvider interface {
	ToggleButton(active *Theme, request ToggleRequest) ToggleVisual
}

// ValidateToggleVisual checks the complete toggle declaration before a widget
// stores a provider result in its visual cache. Toggle behavior owns selected
// state, input, and semantics; the provider must still declare the root
// surface, state transitions, typography, and icon geometry it presents.
func ValidateToggleVisual(visual ToggleVisual) {
	states := []style.State{
		{},
		{Hovered: true},
		{Pressed: true},
		{Selected: true},
		{Selected: true, Hovered: true},
		{Selected: true, Pressed: true},
		{Focused: true},
		{FocusVisible: true},
		{Disabled: true},
		{Selected: true, Disabled: true},
	}
	for _, state := range states {
		root := style.Cascade(state, visual.Defaults, visual.Variant, visual.Size, visual.Override)
		if root.Box == nil {
			invalidVisual("Toggle", "Root.Box", "must be provided")
		}
		if !hasPositiveHeight(root.Box.Height, root.Box.MinHeight) {
			invalidVisual("Toggle", "Root.Box.Height", "height or minimum height must be finite and greater than 0")
		}
		if root.Paint == nil {
			invalidVisual("Toggle", "Root.Paint", "must be provided")
		}
		if !hasPaint(root.Paint.Background) {
			invalidVisual("Toggle", "Root.Paint.Background", "must provide a solid color or nonempty gradient")
		}
		if root.Paint.Radius == nil {
			invalidVisual("Toggle", "Root.Paint.Radius", "must be provided")
		}
		if *root.Paint.Radius < 0 {
			invalidVisual("Toggle", "Root.Paint.Radius", "must be at least 0")
		}
		if root.Paint.Opacity == nil {
			invalidVisual("Toggle", "Root.Paint.Opacity", "must be provided")
		}
		if math.IsNaN(float64(*root.Paint.Opacity)) {
			invalidVisual("Toggle", "Root.Paint.Opacity", "must be finite")
		}
		if math.IsInf(float64(*root.Paint.Opacity), 0) {
			invalidVisual("Toggle", "Root.Paint.Opacity", "must be finite")
		}
		if *root.Paint.Opacity < 0 {
			invalidVisual("Toggle", "Root.Paint.Opacity", "must be at least 0")
		}
		if *root.Paint.Opacity > 1 {
			invalidVisual("Toggle", "Root.Paint.Opacity", "must be at most 1")
		}
		if root.Text == nil {
			invalidVisual("Toggle", "Root.Text", "must be provided")
		}
		if !hasColor(root.Text.Color) {
			invalidVisual("Toggle", "Root.Text.Color", "must resolve to a solid color")
		}
		if !hasValidScale(root.Trans) {
			invalidVisual("Toggle", "Root.Trans", "scale axes must be finite and greater than 0")
		}

		label := style.CascadePart(state, style.PartLabel, visual.Defaults, visual.Variant, visual.Size, visual.Override)
		if label.Text == nil {
			invalidVisual("Toggle", "Label.Text", "must be provided")
		}
		if label.Text.FontSize == nil {
			invalidVisual("Toggle", "Label.Text.FontSize", "must be provided")
		}
		if *label.Text.FontSize <= 0 {
			invalidVisual("Toggle", "Label.Text.FontSize", "must be greater than 0")
		}
		if math.IsNaN(float64(*label.Text.FontSize)) {
			invalidVisual("Toggle", "Label.Text.FontSize", "must be finite")
		}
		if math.IsInf(float64(*label.Text.FontSize), 0) {
			invalidVisual("Toggle", "Label.Text.FontSize", "must be finite")
		}
		if !hasColor(label.Text.Color) {
			invalidVisual("Toggle", "Label.Text.Color", "must resolve to a solid color")
		}

		icon := style.CascadePart(state, style.PartIcon, visual.Defaults, visual.Variant, visual.Size, visual.Override)
		if icon.Box == nil {
			invalidVisual("Toggle", "Icon.Box", "must be provided")
		}
		if !hasPositiveHeight(icon.Box.Height, icon.Box.MinHeight) {
			invalidVisual("Toggle", "Icon.Box.Height", "height or minimum height must be finite and greater than 0")
		}
		if !hasPositiveWidth(icon.Box.Width, icon.Box.MinWidth) {
			invalidVisual("Toggle", "Icon.Box.Width", "width or minimum width must be finite and greater than 0")
		}
		if icon.Text == nil {
			invalidVisual("Toggle", "Icon.Text", "must be provided")
		}
		if icon.Text.FontSize == nil {
			invalidVisual("Toggle", "Icon.Text.FontSize", "must be provided")
		}
		if *icon.Text.FontSize <= 0 {
			invalidVisual("Toggle", "Icon.Text.FontSize", "must be greater than 0")
		}
		if !hasColor(icon.Text.Color) {
			invalidVisual("Toggle", "Icon.Text.Color", "must resolve to a solid color")
		}
	}

	focus := style.Cascade(style.State{FocusVisible: true}, visual.Defaults, visual.Variant, visual.Size, visual.Override)
	if focus.Paint == nil {
		invalidVisual("Toggle", "Focus.Paint", "must be provided")
	}
	if focus.Paint.Outline == nil {
		invalidVisual("Toggle", "Focus.Paint.Outline", "must be provided")
	}
	if focus.Paint.Outline.Width <= 0 {
		invalidVisual("Toggle", "Focus.Paint.Outline.Width", "must be greater than 0")
	}
	if !hasColor(focus.Paint.Outline.Color) {
		invalidVisual("Toggle", "Focus.Paint.Outline.Color", "must resolve to a solid color")
	}
}
