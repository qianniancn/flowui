package theme

import (
	"time"

	"gioui.org/unit"
	"github.com/qianniancn/flowui/style"
)

// SwitchRequest describes the behavior-owned size tier needed by a switch
// visual provider. Checked state, focus, and interaction state remain in the
// widget and enter the returned declaration at resolve time.
type SwitchRequest struct {
	Size     SwitchSize
	Override *SwitchOverride
}

// SwitchOverride carries caller customization across the neutral provider
// boundary. The provider owns the returned instance layer.
type SwitchOverride struct {
	Declaration style.Style
}

// SwitchOverrideStyle extracts the provider-owned instance layer from a
// request.
func SwitchOverrideStyle(request SwitchRequest) style.Style {
	if request.Override == nil {
		return style.Style{}
	}
	return request.Override.Declaration
}

// SwitchMetrics contains the neutral layout values used around the switch
// control. Track and thumb dimensions stay in the returned declarations so
// provider-owned style overrides can participate in geometry resolution.
type SwitchMetrics struct {
	FocusSpace     unit.Dp
	ContentGap     unit.Dp
	DescriptionGap unit.Dp
}

// SwitchGroupMetrics contains the spacing for a group of switch behaviors.
type SwitchGroupMetrics struct {
	VerticalGap   unit.Dp
	HorizontalGap unit.Dp
}

// SwitchGroupRequest carries caller customization across the provider
// boundary for a switch group.
type SwitchGroupRequest struct {
	Override *SwitchOverride
}

// SwitchGroupVisual contains the provider-owned group declaration and layout
// metrics.
type SwitchGroupVisual struct {
	Metrics  SwitchGroupMetrics
	Override style.Style
}

// SwitchVisual is the complete visual plan consumed by switch behavior.
type SwitchVisual struct {
	Defaults          style.Style
	Override          style.Style
	Metrics           SwitchMetrics
	Motion            MotionTheme
	SelectionDuration time.Duration
	FocusDuration     time.Duration
}

// SwitchProvider supplies switch declarations, layout metrics, and visual
// motion timing without taking ownership of checked state, focus, or
// semantics.
type SwitchProvider interface {
	Switch(active *Theme, request SwitchRequest) SwitchVisual
	SwitchGroup(active *Theme, request SwitchGroupRequest) SwitchGroupVisual
}

// ValidateSwitchVisual checks every switch Part and the layout/motion values
// before a widget stores a provider result in its visual cache. Checked,
// focus, disabled, and invalid state are supplied by switch behavior; the
// provider must still declare their visual resolution for every Part.
func ValidateSwitchVisual(visual SwitchVisual) {
	metrics := visual.Metrics
	if metrics.FocusSpace < 0 {
		invalidVisual("Switch", "Metrics.FocusSpace", "must be at least 0")
	}
	if metrics.ContentGap < 0 {
		invalidVisual("Switch", "Metrics.ContentGap", "must be at least 0")
	}
	if metrics.DescriptionGap < 0 {
		invalidVisual("Switch", "Metrics.DescriptionGap", "must be at least 0")
	}
	if visual.SelectionDuration <= 0 {
		invalidVisual("Switch", "SelectionDuration", "must be greater than 0")
	}
	if visual.FocusDuration <= 0 {
		invalidVisual("Switch", "FocusDuration", "must be greater than 0")
	}

	validateMotion("Switch", "Motion", visual.Motion)
	resolve := func(part style.Part, state style.State) style.ResolvedStyle {
		return style.CascadePart(state, part, visual.Defaults, visual.Override)
	}
	states := []style.State{
		{},
		{Hovered: true},
		{Pressed: true},
		{Focused: true},
		{FocusVisible: true},
		{Checked: true, Selected: true},
		{Invalid: true},
		{Disabled: true},
		{Checked: true, Selected: true, Invalid: true},
		{Checked: true, Selected: true, Disabled: true},
	}
	for _, state := range states {
		for _, part := range []style.Part{style.PartTrack, style.PartThumb} {
			resolved := resolve(part, state)
			if resolved.Box == nil {
				invalidVisual("Switch", "Resolved.Box", "must be provided")
			}
			if resolved.Box.Width == nil {
				invalidVisual("Switch", "Resolved.Box.Width", "must be provided")
			}
			if *resolved.Box.Width <= 0 {
				invalidVisual("Switch", "Resolved.Box.Width", "must be greater than 0")
			}
			if resolved.Box.Height == nil {
				invalidVisual("Switch", "Resolved.Box.Height", "must be provided")
			}
			if *resolved.Box.Height <= 0 {
				invalidVisual("Switch", "Resolved.Box.Height", "must be greater than 0")
			}
			if resolved.Paint == nil {
				invalidVisual("Switch", "Resolved.Paint", "must be provided")
			}
			if resolved.Paint.Background == nil {
				invalidVisual("Switch", "Resolved.Paint.Background", "must be provided")
			}
			if resolved.Paint.Radius == nil {
				invalidVisual("Switch", "Resolved.Paint.Radius", "must be provided")
			}
			if *resolved.Paint.Radius < 0 {
				invalidVisual("Switch", "Resolved.Paint.Radius", "must be at least 0")
			}
		}
		for _, part := range []style.Part{style.PartLabel, style.PartDescription} {
			resolved := resolve(part, state)
			if resolved.Text == nil {
				invalidVisual("Switch", "Resolved.Text", "must be provided")
			}
			if resolved.Text.Color == nil {
				invalidVisual("Switch", "Resolved.Text.Color", "must be provided")
			}
			if resolved.Text.FontSize == nil {
				invalidVisual("Switch", "Resolved.Text.FontSize", "must be provided")
			}
			if *resolved.Text.FontSize <= 0 {
				invalidVisual("Switch", "Resolved.Text.FontSize", "must be greater than 0")
			}
		}
	}
}

// ValidateSwitchGroupVisual checks the provider-owned group spacing before a
// group applies its root declaration to the child layout.
func ValidateSwitchGroupVisual(visual SwitchGroupVisual) {
	if visual.Metrics.VerticalGap < 0 {
		invalidVisual("SwitchGroup", "Metrics.VerticalGap", "must be at least 0")
	}
	if visual.Metrics.HorizontalGap < 0 {
		invalidVisual("SwitchGroup", "Metrics.HorizontalGap", "must be at least 0")
	}
}
