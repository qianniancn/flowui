package theme

import (
	"image/color"

	"gioui.org/unit"
	"github.com/qianniancn/flowui/style"
)

// LabelRequest describes the inherited text color and instance-level visual
// override needed by a form label. The widget owns field interaction and
// required-state behavior; the provider owns typography and state colors.
type LabelRequest struct {
	Foreground color.NRGBA
	Override   *LabelOverride
}

// LabelOverride carries caller customization across the neutral provider
// boundary. The provider owns how this declaration is merged.
type LabelOverride struct {
	Declaration style.Style
}

// LabelOverrideStyle extracts the provider-owned instance layer.
func LabelOverrideStyle(request LabelRequest) style.Style {
	if request.Override == nil {
		return style.Style{}
	}
	return request.Override.Declaration
}

// LabelVisual is the complete visual plan consumed by label behavior.
type LabelVisual struct {
	Defaults        style.Style
	Required        style.Style
	Override        style.Style
	RequiredMarkGap unit.Dp
}

// LabelProvider supplies form-label declarations without taking ownership of
// field association, focus, click, or required-mark behavior.
type LabelProvider interface {
	Label(active *Theme, request LabelRequest) LabelVisual
}

// ValidateLabelVisual checks the root and required-mark declarations before a
// widget stores a provider result in its cache. Label behavior cannot recover
// typography or state colors without reintroducing theme knowledge.
func ValidateLabelVisual(visual LabelVisual) {
	if visual.RequiredMarkGap < 0 {
		invalidVisual("Label", "RequiredMarkGap", "must be at least 0")
	}

	for _, state := range []style.State{
		{},
		{Invalid: true},
		{Disabled: true},
		{Disabled: true, Invalid: true},
	} {
		resolved := style.Cascade(state, visual.Defaults, visual.Override)
		if resolved.Text == nil {
			invalidVisual("Label", "Resolved.Text", "must be provided")
		}
		if resolved.Text.FontSize == nil {
			invalidVisual("Label", "Resolved.Text.FontSize", "must be provided")
		}
		if *resolved.Text.FontSize <= 0 {
			invalidVisual("Label", "Resolved.Text.FontSize", "must be greater than 0")
		}
		if resolved.Text.FontWeight == nil {
			invalidVisual("Label", "Resolved.Text.FontWeight", "must be provided")
		}
		if !hasColor(resolved.Text.Color) {
			invalidVisual("Label", "Resolved.Text.Color", "must resolve to a solid color")
		}

		required := style.Cascade(state, visual.Required)
		if required.Text == nil {
			invalidVisual("Label", "Required.Text", "must be provided")
		}
		if !hasColor(required.Text.Color) {
			invalidVisual("Label", "Required.Text.Color", "must resolve to a solid color")
		}
	}
}
