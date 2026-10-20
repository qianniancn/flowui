package theme

import (
	"image"
	"image/color"
	"time"

	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/style"
)

// DropdownRequest describes the instance-level declaration supplied by a
// dropdown wrapper. Trigger interaction and popup menu behavior remain owned
// by the widget; the provider owns the trigger visual metrics and renderer.
type DropdownRequest struct {
	Override *DropdownOverride
}

// DropdownOverride carries caller customization across the provider boundary.
type DropdownOverride struct {
	Declaration style.Style
}

// DropdownOverrideStyle extracts the provider-owned instance layer from a
// request without making widgets inspect the override representation.
func DropdownOverrideStyle(request DropdownRequest) style.Style {
	if request.Override == nil {
		return style.Style{}
	}
	return request.Override.Declaration
}

// DropdownTriggerFocusRequest contains the layout facts needed by a theme to
// paint the trigger's focus indication.
type DropdownTriggerFocusRequest struct {
	Size    image.Point
	Opacity float32
}

// DropdownTriggerFocusRenderer is a theme-owned trigger focus renderer.
type DropdownTriggerFocusRenderer func(layout.Context, DropdownTriggerFocusRequest)

// DropdownTriggerGlyphRequest carries the neutral facts needed to paint the
// default split-button trigger glyph.
type DropdownTriggerGlyphRequest struct {
	Color color.NRGBA
}

// DropdownTriggerGlyphRenderer paints the provider-owned default trigger
// glyph and returns its intrinsic dimensions.
type DropdownTriggerGlyphRenderer func(layout.Context, DropdownTriggerGlyphRequest) layout.Dimensions

// DropdownVisual is the complete visual plan consumed by dropdown behavior.
// It intentionally contains no widget state or callbacks.
type DropdownVisual struct {
	TriggerPressedScale float32
	PressDuration       time.Duration
	EnterDuration       time.Duration
	ExitDuration        time.Duration
	LongPressDuration   time.Duration
	HoverOpenDuration   time.Duration
	HoverCloseDuration  time.Duration
	PanelGap            unit.Dp
	ArrowSize           unit.Dp
	Motion              MotionTheme
	TriggerFocus        DropdownTriggerFocusRenderer
	TriggerGlyph        DropdownTriggerGlyphRenderer
	Override            style.Style
}

// DropdownProvider supplies trigger feedback, popup geometry, and motion
// without taking ownership of open state, focus, or menu behavior.
type DropdownProvider interface {
	Dropdown(active *Theme, request DropdownRequest) DropdownVisual
}

// ValidateDropdownVisual checks the trigger renderer and motion declaration
// before a widget stores a provider result in its visual cache.
func ValidateDropdownVisual(visual DropdownVisual) {
	if visual.TriggerFocus == nil {
		invalidVisual("Dropdown", "TriggerFocus", "must be provided")
	}
	if visual.TriggerGlyph == nil {
		invalidVisual("Dropdown", "TriggerGlyph", "must be provided")
	}
	if visual.TriggerPressedScale <= 0 {
		invalidVisual("Dropdown", "TriggerPressedScale", "must be greater than 0")
	}
	if visual.TriggerPressedScale > 1 {
		invalidVisual("Dropdown", "TriggerPressedScale", "must be at most 1")
	}
	if visual.PressDuration <= 0 {
		invalidVisual("Dropdown", "PressDuration", "must be greater than 0")
	}
	if visual.EnterDuration <= 0 {
		invalidVisual("Dropdown", "EnterDuration", "must be greater than 0")
	}
	if visual.ExitDuration <= 0 {
		invalidVisual("Dropdown", "ExitDuration", "must be greater than 0")
	}
	if visual.LongPressDuration <= 0 {
		invalidVisual("Dropdown", "LongPressDuration", "must be greater than 0")
	}
	if visual.HoverOpenDuration <= 0 {
		invalidVisual("Dropdown", "HoverOpenDuration", "must be greater than 0")
	}
	if visual.HoverCloseDuration <= 0 {
		invalidVisual("Dropdown", "HoverCloseDuration", "must be greater than 0")
	}
	if visual.PanelGap < 0 {
		invalidVisual("Dropdown", "PanelGap", "must be at least 0")
	}
	if visual.ArrowSize <= 0 {
		invalidVisual("Dropdown", "ArrowSize", "must be greater than 0")
	}
}
