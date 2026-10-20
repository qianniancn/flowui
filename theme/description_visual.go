package theme

import (
	"github.com/qianniancn/flowui/style"
)

// DescriptionRequest describes the instance-level visual override for
// supporting text. The widget owns text content and field association; the
// provider owns typography and state colors.
type DescriptionRequest struct {
	Override *DescriptionOverride
}

// DescriptionOverride carries caller customization across the neutral
// provider boundary. The provider owns how this declaration is merged.
type DescriptionOverride struct {
	Declaration style.Style
}

// DescriptionOverrideStyle extracts the provider-owned instance layer.
func DescriptionOverrideStyle(request DescriptionRequest) style.Style {
	if request.Override == nil {
		return style.Style{}
	}
	return request.Override.Declaration
}

// DescriptionVisual is the complete visual plan consumed by description
// behavior. The widget only resolves these declarations and lays out text.
type DescriptionVisual struct {
	Defaults style.Style
	Override style.Style
}

// DescriptionProvider supplies supporting-text declarations without taking
// ownership of text content or field association behavior.
type DescriptionProvider interface {
	Description(active *Theme, request DescriptionRequest) DescriptionVisual
}

// ValidateDescriptionVisual checks the declaration contract before a widget
// stores a provider result in its cache. Description behavior cannot recover a
// missing text declaration without reintroducing theme knowledge.
func ValidateDescriptionVisual(visual DescriptionVisual) {
	requireText := func(state style.State) {
		resolved := style.Cascade(state, visual.Defaults, visual.Override)
		if resolved.Text == nil {
			invalidVisual("Description", "Resolved.Text", "must be provided")
		}
		if resolved.Text.FontSize == nil {
			invalidVisual("Description", "Resolved.Text.FontSize", "must be provided")
		}
		if *resolved.Text.FontSize <= 0 {
			invalidVisual("Description", "Resolved.Text.FontSize", "must be greater than 0")
		}
		if !hasColor(resolved.Text.Color) {
			invalidVisual("Description", "Resolved.Text.Color", "must resolve to a solid color")
		}
	}
	requireText(style.State{})
	requireText(style.State{Disabled: true})
}
