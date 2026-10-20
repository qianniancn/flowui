package theme

import (
	"image"
	"image/color"
	"time"

	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/style"
)

// TooltipRequest describes the optional instance declaration supplied by a
// tooltip wrapper. Trigger state, placement, and content remain widget-owned.
type TooltipRequest struct {
	Override *TooltipOverride
}

// TooltipOverride carries caller customization across the provider boundary.
type TooltipOverride struct {
	Declaration style.Style
}

// TooltipOverrideStyle extracts the provider-owned instance layer.
func TooltipOverrideStyle(request TooltipRequest) style.Style {
	if request.Override == nil {
		return style.Style{}
	}
	return request.Override.Declaration
}

// TooltipMetrics contains neutral geometry, timing, and motion values used by
// tooltip behavior. The widget consumes these values without interpreting
// theme tokens or selecting visual fallbacks.
type TooltipMetrics struct {
	Offset            unit.Dp
	ArrowOffset       unit.Dp
	Padding           unit.Dp
	Radius            unit.Dp
	BorderWidth       unit.Dp
	MaxWidth          unit.Dp
	ArrowSize         unit.Dp
	TextSize          unit.Sp
	AnimationScale    float32
	ExitScale         float32
	AnimationDistance unit.Dp
	Delay             time.Duration
	CloseDelay        time.Duration
	EnterDuration     time.Duration
	ExitDuration      time.Duration
	MoveDuration      time.Duration
}

// TooltipPanelColors contains the resolved colors needed by an arrow
// renderer. The provider interprets the resolved declaration; the widget
// passes it through without inspecting paint sources.
type TooltipPanelColors struct {
	Surface color.NRGBA
	Border  color.NRGBA
}

// TooltipPanelColorResolver extracts arrow colors from the resolved panel.
type TooltipPanelColorResolver func(style.ResolvedStyle) TooltipPanelColors

// TooltipArrowRequest contains the layout facts needed to paint one tooltip
// arrow. Anchor is measured along the panel cross-axis in local coordinates.
type TooltipArrowRequest struct {
	Side        uint8
	Panel       image.Point
	Anchor      float32
	ArrowSize   int
	BorderWidth int
	Colors      TooltipPanelColors
}

const (
	TooltipSideTop uint8 = iota
	TooltipSideLeft
	TooltipSideRight
	TooltipSideBottom
)

// TooltipArrowRenderer is a theme-owned arrow drawing strategy.
type TooltipArrowRenderer func(layout.Context, TooltipArrowRequest)

// TooltipVisual is the complete visual plan consumed by tooltip behavior.
type TooltipVisual struct {
	Defaults   style.Style
	Content    style.Style
	Override   style.Style
	Metrics    TooltipMetrics
	Motion     MotionTheme
	Arrow      TooltipArrowRenderer
	PanelColor TooltipPanelColorResolver
}

// TooltipProvider supplies panel declarations, content typography, geometry,
// motion, and arrow drawing without taking ownership of trigger behavior.
type TooltipProvider interface {
	Tooltip(active *Theme, request TooltipRequest) TooltipVisual
}

// ValidateTooltipVisual checks the declaration, motion, and arrow bridge
// before a widget stores a provider result in its visual cache.
func ValidateTooltipVisual(visual TooltipVisual) {
	if visual.Arrow == nil {
		invalidVisual("Tooltip", "Arrow", "must be provided")
	}
	if visual.PanelColor == nil {
		invalidVisual("Tooltip", "PanelColor", "must be provided")
	}

	metrics := visual.Metrics
	if metrics.MaxWidth <= 0 {
		invalidVisual("Tooltip", "Metrics.MaxWidth", "must be greater than 0")
	}
	if metrics.Padding < 0 {
		invalidVisual("Tooltip", "Metrics.Padding", "must be at least 0")
	}
	if metrics.Radius < 0 {
		invalidVisual("Tooltip", "Metrics.Radius", "must be at least 0")
	}
	if metrics.BorderWidth < 0 {
		invalidVisual("Tooltip", "Metrics.BorderWidth", "must be at least 0")
	}
	if metrics.Offset < 0 {
		invalidVisual("Tooltip", "Metrics.Offset", "must be at least 0")
	}
	if metrics.ArrowOffset < 0 {
		invalidVisual("Tooltip", "Metrics.ArrowOffset", "must be at least 0")
	}
	if metrics.ArrowSize <= 0 {
		invalidVisual("Tooltip", "Metrics.ArrowSize", "must be greater than 0")
	}
	if metrics.TextSize <= 0 {
		invalidVisual("Tooltip", "Metrics.TextSize", "must be greater than 0")
	}
	if metrics.AnimationScale <= 0 {
		invalidVisual("Tooltip", "Metrics.AnimationScale", "must be greater than 0")
	}
	if metrics.AnimationScale > 1 {
		invalidVisual("Tooltip", "Metrics.AnimationScale", "must be at most 1")
	}
	if metrics.ExitScale <= 0 {
		invalidVisual("Tooltip", "Metrics.ExitScale", "must be greater than 0")
	}
	if metrics.ExitScale > 1 {
		invalidVisual("Tooltip", "Metrics.ExitScale", "must be at most 1")
	}
	if metrics.AnimationDistance < 0 {
		invalidVisual("Tooltip", "Metrics.AnimationDistance", "must be at least 0")
	}
	if metrics.Delay < 0 {
		invalidVisual("Tooltip", "Metrics.Delay", "must be at least 0")
	}
	if metrics.CloseDelay < 0 {
		invalidVisual("Tooltip", "Metrics.CloseDelay", "must be at least 0")
	}
	if metrics.EnterDuration <= 0 {
		invalidVisual("Tooltip", "Metrics.EnterDuration", "must be greater than 0")
	}
	if metrics.ExitDuration <= 0 {
		invalidVisual("Tooltip", "Metrics.ExitDuration", "must be greater than 0")
	}
	if metrics.MoveDuration <= 0 {
		invalidVisual("Tooltip", "Metrics.MoveDuration", "must be greater than 0")
	}

	root := style.Cascade(style.State{}, visual.Defaults, visual.Override)
	if root.Box == nil {
		invalidVisual("Tooltip", "Root.Box", "must be provided")
	}
	if root.Box.Padding == nil {
		invalidVisual("Tooltip", "Root.Box.Padding", "must be provided")
	}
	if root.Paint == nil {
		invalidVisual("Tooltip", "Root.Paint", "must be provided")
	}
	if root.Paint.Background == nil {
		invalidVisual("Tooltip", "Root.Paint.Background", "must be provided")
	}

	content := style.Cascade(style.State{}, visual.Content)
	if content.Text == nil {
		invalidVisual("Tooltip", "Content.Text", "must be provided")
	}
	if content.Text.FontSize == nil {
		invalidVisual("Tooltip", "Content.Text.FontSize", "must be provided")
	}
	if !validPositiveSp(*content.Text.FontSize) {
		invalidVisual("Tooltip", "Content.Text.FontSize", "must be finite and greater than 0")
	}
	if !hasColor(content.Text.Color) {
		invalidVisual("Tooltip", "Content.Text.Color", "must resolve to a solid color")
	}
}
