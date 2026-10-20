package theme

import (
	"time"

	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/style"
)

// ToastRequest describes the semantic variant and optional instance
// declaration needed by a toast visual provider. Stacking, timing, focus,
// dismissal, and action behavior remain owned by the widget.
type ToastRequest struct {
	Variant  ToastVariant
	Override *ToastOverride
}

// ToastOverride carries caller customization across the provider boundary.
type ToastOverride struct {
	Declaration style.Style
}

// ToastOverrideStyle extracts the provider-owned instance layer.
func ToastOverrideStyle(request ToastRequest) style.Style {
	if request.Override == nil {
		return style.Style{}
	}
	return request.Override.Declaration
}

// ToastMetrics contains neutral geometry and timing values consumed by toast
// behavior. The widget never reads Components.Toast directly.
type ToastMetrics struct {
	Width             unit.Dp
	Inset             unit.Dp
	Gap               unit.Dp
	ContentGap        unit.Dp
	IndicatorSize     unit.Dp
	CloseInset        unit.Dp
	CloseButtonSize   unit.Dp
	ActionInsetTop    unit.Dp
	MaxVisible        int
	ScaleFactor       float32
	AnimationDuration time.Duration
	DefaultTimeout    time.Duration
}

// ToastTextRole identifies the semantic text slot being rendered.
type ToastTextRole uint8

const (
	ToastTextTitle ToastTextRole = iota
	ToastTextDescription
)

// ToastTextRenderRequest contains semantic text only. Typography, color, and
// line treatment remain owned by the visual provider.
type ToastTextRenderRequest struct {
	Text string
	Role ToastTextRole
}

// ToastTextRenderer paints one provider-owned toast text slot.
type ToastTextRenderer func(layout.Context, ToastTextRenderRequest) layout.Dimensions

// ToastIndicatorRenderRequest contains neutral indicator facts. Custom is a
// caller-owned indicator slot; the provider still owns its surrounding box,
// loading treatment, and default glyph.
type ToastIndicatorRenderRequest struct {
	Size    unit.Dp
	Loading bool
	Custom  func(layout.Context) layout.Dimensions
}

// ToastIndicatorRenderer paints the default, loading, or custom indicator.
type ToastIndicatorRenderer func(layout.Context, ToastIndicatorRenderRequest) layout.Dimensions

// ToastActionRenderRequest contains behavior state for an action slot. The
// provider owns the button visual while the widget retains the Clickable and
// invokes the action callback.
type ToastActionRenderRequest struct {
	Label     string
	Variant   ButtonVariant
	Clickable *widget.Clickable
	State     style.State
}

// ToastActionRenderer paints an action while preserving the widget-owned
// Clickable lifecycle.
type ToastActionRenderer func(layout.Context, ToastActionRenderRequest) layout.Dimensions

// ToastCloseRenderRequest contains behavior state for the dismiss slot.
type ToastCloseRenderRequest struct {
	Label     string
	Clickable *widget.Clickable
	Size      unit.Dp
	State     style.State
}

// ToastCloseRenderer paints the provider-owned dismiss control.
type ToastCloseRenderer func(layout.Context, ToastCloseRenderRequest) layout.Dimensions

// ToastVisual is the complete visual plan consumed by toast behavior. The
// provider owns colors, typography, glyph choice, and the root declaration;
// the widget only executes the returned values.
type ToastVisual struct {
	Defaults  style.Style
	Override  style.Style
	Metrics   ToastMetrics
	Motion    MotionTheme
	Text      ToastTextRenderer
	Indicator ToastIndicatorRenderer
	Action    ToastActionRenderer
	Close     ToastCloseRenderer
}

// ToastProvider supplies declarations, metrics, colors, and glyphs without
// taking ownership of toast stacking, timers, focus, or actions.
type ToastProvider interface {
	Toast(active *Theme, request ToastRequest) ToastVisual
}

// ValidateToastVisual checks the declaration and neutral metrics before a
// widget stores a provider result in its visual cache. Toast behavior must not
// recover missing geometry or colors with a local theme fallback.
func ValidateToastVisual(visual ToastVisual) {
	if visual.Text == nil {
		invalidVisual("Toast", "Text", "must be provided")
	}
	if visual.Indicator == nil {
		invalidVisual("Toast", "Indicator", "must be provided")
	}
	if visual.Action == nil {
		invalidVisual("Toast", "Action", "must be provided")
	}
	if visual.Close == nil {
		invalidVisual("Toast", "Close", "must be provided")
	}

	metrics := visual.Metrics
	if metrics.Width <= 0 {
		invalidVisual("Toast", "Metrics.Width", "must be greater than 0")
	}
	if metrics.Inset < 0 {
		invalidVisual("Toast", "Metrics.Inset", "must be at least 0")
	}
	if metrics.Gap < 0 {
		invalidVisual("Toast", "Metrics.Gap", "must be at least 0")
	}
	if metrics.ContentGap < 0 {
		invalidVisual("Toast", "Metrics.ContentGap", "must be at least 0")
	}
	if metrics.ActionInsetTop < 0 {
		invalidVisual("Toast", "Metrics.ActionInsetTop", "must be at least 0")
	}
	if metrics.IndicatorSize <= 0 {
		invalidVisual("Toast", "Metrics.IndicatorSize", "must be greater than 0")
	}
	if metrics.CloseButtonSize <= 0 {
		invalidVisual("Toast", "Metrics.CloseButtonSize", "must be greater than 0")
	}
	if metrics.MaxVisible <= 0 {
		invalidVisual("Toast", "Metrics.MaxVisible", "must be greater than 0")
	}
	if metrics.ScaleFactor < 0 {
		invalidVisual("Toast", "Metrics.ScaleFactor", "must be at least 0")
	}
	if metrics.ScaleFactor > 1 {
		invalidVisual("Toast", "Metrics.ScaleFactor", "must be at most 1")
	}
	if metrics.AnimationDuration <= 0 {
		invalidVisual("Toast", "Metrics.AnimationDuration", "must be greater than 0")
	}
	if metrics.DefaultTimeout < 0 {
		invalidVisual("Toast", "Metrics.DefaultTimeout", "must be at least 0")
	}

	root := style.Cascade(style.State{}, visual.Defaults, visual.Override)
	if root.Box == nil {
		invalidVisual("Toast", "Root.Box", "must be provided")
	}
	if root.Box.Padding == nil {
		invalidVisual("Toast", "Root.Box.Padding", "must be provided")
	}
	if root.Box.Overflow == nil {
		invalidVisual("Toast", "Root.Box.Overflow", "must be provided")
	}
	if *root.Box.Overflow != style.OverflowHidden {
		invalidVisual("Toast", "Root.Box.Overflow", "must be style.OverflowHidden")
	}
	if root.Paint == nil {
		invalidVisual("Toast", "Root.Paint", "must be provided")
	}
	if root.Paint.Background == nil {
		invalidVisual("Toast", "Root.Paint.Background", "must be provided")
	}

	baseRoot := style.Cascade(style.State{}, visual.Defaults)
	if baseRoot.Box == nil {
		invalidVisual("Toast", "BaseRoot.Box", "must be provided")
	}
	if baseRoot.Box.Overflow == nil {
		invalidVisual("Toast", "BaseRoot.Box.Overflow", "must be provided")
	}
	if *baseRoot.Box.Overflow != style.OverflowHidden {
		invalidVisual("Toast", "BaseRoot.Box.Overflow", "must be style.OverflowHidden")
	}
}
