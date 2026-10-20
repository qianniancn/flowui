package theme

import (
	"image"
	"image/color"

	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/style"
)

// MenubarRequest describes the behavior-owned density and optional instance
// declaration needed by a menubar visual provider.
type MenubarRequest struct {
	Compact  bool
	Override *MenubarOverride
}

// MenubarOverride carries caller customization across the provider boundary.
type MenubarOverride struct {
	Declaration style.Style
}

// MenubarOverrideStyle extracts the provider-owned instance layer.
func MenubarOverrideStyle(request MenubarRequest) style.Style {
	if request.Override == nil {
		return style.Style{}
	}
	return request.Override.Declaration
}

// MenubarMetrics contains neutral geometry used by menubar behavior and
// layout. Interaction and navigation remain owned by the widget.
type MenubarMetrics struct {
	TriggerHeight          unit.Dp
	TriggerPaddingX        unit.Dp
	TriggerRadius          unit.Dp
	TriggerFocusRingWidth  unit.Dp
	TriggerFocusRingOffset unit.Dp
	Gap                    unit.Dp
	PanelGap               unit.Dp
}

// MenubarTriggerRequest contains the interaction state needed to derive one
// trigger's visual state. The provider owns its colors and opacity policy.
type MenubarTriggerRequest struct {
	Open     bool
	Hovered  bool
	Pressed  bool
	Focused  bool
	Disabled bool
}

// MenubarTriggerVisual is the resolved visual state for one trigger.
type MenubarTriggerVisual struct {
	Background color.NRGBA
	Foreground color.NRGBA
	Focus      color.NRGBA
	Opacity    float32
}

// MenubarTriggerRenderer paints the trigger surface and focus indication.
type MenubarTriggerRenderer func(layout.Context, image.Point, MenubarMetrics, MenubarTriggerVisual, float32)

// MenubarLabelRequest contains the semantic trigger text and its resolved
// foreground. The provider owns typography and the Gio text primitive.
type MenubarLabelRequest struct {
	Text       string
	Foreground color.NRGBA
}

// MenubarLabelRenderer measures and paints a trigger label using provider-
// owned typography. Widget code supplies only semantic text and constraints.
type MenubarLabelRenderer func(layout.Context, MenubarLabelRequest) layout.Dimensions

// MenubarVisual is the complete visual plan consumed by menubar behavior.
type MenubarVisual struct {
	Metrics       MenubarMetrics
	Motion        MotionTheme
	Override      style.Style
	Trigger       func(MenubarTriggerRequest) MenubarTriggerVisual
	RenderTrigger MenubarTriggerRenderer
	RenderLabel   MenubarLabelRenderer
}

// MenubarProvider supplies menubar declarations, geometry, motion, colors,
// and drawing strategies without taking ownership of navigation behavior.
type MenubarProvider interface {
	Menubar(active *Theme, request MenubarRequest) MenubarVisual
}

// ValidateMenubarVisual checks the neutral geometry, motion, trigger state
// resolver, and renderer before a widget stores a provider result in its
// visual cache. Transparent backgrounds are valid because a behavior-only
// provider may intentionally leave the trigger surface unpainted.
func ValidateMenubarVisual(visual MenubarVisual) {
	metrics := visual.Metrics
	if !validPositiveDp(metrics.TriggerHeight) {
		invalidVisual("Menubar", "Metrics.TriggerHeight", "must be finite and greater than 0")
	}
	if !validNonNegativeDp(metrics.TriggerPaddingX) {
		invalidVisual("Menubar", "Metrics.TriggerPaddingX", "must be finite and non-negative")
	}
	if !validNonNegativeDp(metrics.TriggerRadius) {
		invalidVisual("Menubar", "Metrics.TriggerRadius", "must be finite and non-negative")
	}
	if !validPositiveDp(metrics.TriggerFocusRingWidth) {
		invalidVisual("Menubar", "Metrics.TriggerFocusRingWidth", "must be finite and greater than 0")
	}
	if !validNonNegativeDp(metrics.TriggerFocusRingOffset) {
		invalidVisual("Menubar", "Metrics.TriggerFocusRingOffset", "must be finite and non-negative")
	}
	if !validNonNegativeDp(metrics.Gap) {
		invalidVisual("Menubar", "Metrics.Gap", "must be finite and non-negative")
	}
	if !validNonNegativeDp(metrics.PanelGap) {
		invalidVisual("Menubar", "Metrics.PanelGap", "must be finite and non-negative")
	}
	if visual.Trigger == nil {
		invalidVisual("Menubar", "Trigger", "must be provided")
	}
	if visual.RenderTrigger == nil {
		invalidVisual("Menubar", "RenderTrigger", "must be provided")
	}
	if visual.RenderLabel == nil {
		invalidVisual("Menubar", "RenderLabel", "must be provided")
	}

	validateMotion("Menubar", "Motion", visual.Motion)
	for _, request := range []MenubarTriggerRequest{
		{},
		{Hovered: true},
		{Pressed: true},
		{Open: true},
		{Focused: true},
		{Disabled: true},
		{Open: true, Hovered: true, Pressed: true, Focused: true, Disabled: true},
	} {
		resolved := visual.Trigger(request)
		if resolved.Foreground.A == 0 {
			invalidVisual("Menubar", "Resolved.Foreground.A", "must be nonzero")
		}
		if resolved.Focus.A == 0 {
			invalidVisual("Menubar", "Resolved.Focus.A", "must be nonzero")
		}
		if !hasValidOpacityValue(resolved.Opacity) {
			invalidVisual("Menubar", "Resolved.Opacity", "must be finite and between 0 and 1")
		}
	}
}
