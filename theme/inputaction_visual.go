package theme

import (
	"math"

	"gioui.org/unit"
	"github.com/qianniancn/flowui/style"
)

// InputActionRequest describes the visual inputs for a compact prefix/suffix
// action used by an input group.
type InputActionRequest struct {
	// Override is an optional instance declaration supplied by the caller.
	// The provider owns how it is merged into the action declaration.
	Override *InputActionOverride
}

// InputActionOverride carries caller customization across the neutral
// provider boundary. It is deliberately not stored by the action behavior.
type InputActionOverride struct {
	Declaration style.Style
}

// InputActionOverrideStyle extracts the provider-owned instance layer.
func InputActionOverrideStyle(request InputActionRequest) style.Style {
	if request.Override == nil {
		return style.Style{}
	}
	return request.Override.Declaration
}

// InputActionMetrics contains neutral geometry used by the input group when
// placing a prefix or suffix action.
type InputActionMetrics struct {
	SlotPadding unit.Dp
}

// InputActionVisual is the complete visual plan for a compact input action.
// The widget executes the returned declaration and uses the metrics for slot
// placement; it does not construct a fallback declaration or spacing value.
// Supply a positive MinWidth for content-sized actions, or Width for fixed
// actions, and a positive Height. Text actions should be allowed to grow.
type InputActionVisual struct {
	Declaration style.Style
	Metrics     InputActionMetrics
}

// InputActionProvider supplies the compact prefix/suffix action used by input
// groups.
type InputActionProvider interface {
	InputAction(active *Theme, request InputActionRequest) InputActionVisual
}

// ValidateInputActionVisual checks the compact action declaration before it is
// cached. Group behavior supplies interaction and focus semantics; it must not
// invent action geometry or text color when a provider omits them.
func ValidateInputActionVisual(visual InputActionVisual) {
	if visual.Metrics.SlotPadding < 0 {
		invalidVisual("InputAction", "Metrics.SlotPadding", "must be at least 0")
	}
	if math.IsNaN(float64(visual.Metrics.SlotPadding)) {
		invalidVisual("InputAction", "Metrics.SlotPadding", "must be finite")
	}
	if math.IsInf(float64(visual.Metrics.SlotPadding), 0) {
		invalidVisual("InputAction", "Metrics.SlotPadding", "must be finite")
	}

	for _, state := range []style.State{
		{},
		{Focused: true},
		{FocusVisible: true},
		{Disabled: true},
		{Focused: true, Disabled: true},
	} {
		resolved := visual.Declaration.Resolve(state)
		if resolved.Box == nil {
			invalidVisual("InputAction", "Resolved.Box", "must be provided")
		}
		if resolved.Box.Width != nil {
			if *resolved.Box.Width <= 0 {
				invalidVisual("InputAction", "Resolved.Box.Width", "must be greater than 0")
			}
		} else if resolved.Box.MinWidth == nil || *resolved.Box.MinWidth <= 0 {
			invalidVisual("InputAction", "Resolved.Box.MinWidth", "must be greater than 0 when Width is omitted")
		}
		if resolved.Box.Height == nil {
			invalidVisual("InputAction", "Resolved.Box.Height", "must be provided")
		}
		if *resolved.Box.Height <= 0 {
			invalidVisual("InputAction", "Resolved.Box.Height", "must be greater than 0")
		}
		if resolved.Text == nil {
			invalidVisual("InputAction", "Resolved.Text", "must be provided")
		}
		if !hasColor(resolved.Text.Color) {
			invalidVisual("InputAction", "Resolved.Text.Color", "must resolve to a solid color")
		}
	}
}
