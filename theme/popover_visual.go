package theme

import (
	"image"
	"image/color"
	"time"

	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/style"
)

// PopoverRequest describes the optional instance declaration supplied by a
// popover wrapper. Placement, dismissal, and open state remain widget-owned.
type PopoverRequest struct {
	Override *PopoverOverride
}

// PopoverOverride carries caller customization across the provider boundary.
type PopoverOverride struct {
	Declaration style.Style
}

// PopoverOverrideStyle extracts the provider-owned instance layer.
func PopoverOverrideStyle(request PopoverRequest) style.Style {
	if request.Override == nil {
		return style.Style{}
	}
	return request.Override.Declaration
}

// PopoverMetrics contains the neutral measurements used by popover behavior.
// The widget may consume these values for layout, but it does not choose them.
type PopoverMetrics struct {
	Offset            unit.Dp
	Padding           unit.Dp
	Radius            unit.Dp
	MaxWidth          unit.Dp
	ArrowWidth        unit.Dp
	ArrowHeight       unit.Dp
	HeadingSize       unit.Sp
	BodyTextSize      unit.Sp
	SectionGap        unit.Dp
	AnimationScale    float32
	EnterDuration     time.Duration
	ExitDuration      time.Duration
	AnimationDistance unit.Dp
}

// PopoverArrowRequest contains the layout facts needed by a theme to paint
// the arrow. Side values use the stable constants below instead of importing
// internal/core into theme packages.
type PopoverArrowRequest struct {
	Side  uint8
	Panel image.Point
	Arrow image.Point
	Color color.NRGBA
}

const (
	PopoverSideTop uint8 = iota
	PopoverSideLeft
	PopoverSideRight
	PopoverSideBottom
)

// PopoverArrowRenderer is a theme-owned arrow drawing strategy.
type PopoverArrowRenderer func(layout.Context, PopoverArrowRequest)

// PopoverPanelColorResolver extracts the solid surface color needed by the
// arrow renderer from the provider-owned panel declaration. The widget passes
// the resolved style through without interpreting paint sources itself.
type PopoverPanelColorResolver func(style.ResolvedStyle) color.NRGBA

// PopoverVisual is the complete visual plan consumed by popover behavior.
type PopoverVisual struct {
	Defaults   style.Style
	Heading    style.Style
	Content    style.Style
	Override   style.Style
	Metrics    PopoverMetrics
	Motion     MotionTheme
	Arrow      PopoverArrowRenderer
	PanelColor PopoverPanelColorResolver
}

// PopoverProvider supplies panel declarations, metrics, motion, and arrow
// drawing without taking ownership of popover behavior.
type PopoverProvider interface {
	Popover(active *Theme, request PopoverRequest) PopoverVisual
}

// ValidatePopoverVisual checks the renderer and panel-color bridge before a
// widget stores a provider result in its declaration cache. Popover behavior
// must not recover a missing arrow or paint interpretation with a local
// fallback.
func ValidatePopoverVisual(visual PopoverVisual) {
	if visual.Arrow == nil {
		invalidVisual("Popover", "Arrow", "must be provided")
	}
	if visual.PanelColor == nil {
		invalidVisual("Popover", "PanelColor", "must be provided")
	}

	metrics := visual.Metrics
	if metrics.MaxWidth <= 0 {
		invalidVisual("Popover", "Metrics.MaxWidth", "must be greater than 0")
	}
	if metrics.Padding < 0 {
		invalidVisual("Popover", "Metrics.Padding", "must be at least 0")
	}
	if metrics.Radius < 0 {
		invalidVisual("Popover", "Metrics.Radius", "must be at least 0")
	}
	if metrics.Offset < 0 {
		invalidVisual("Popover", "Metrics.Offset", "must be at least 0")
	}
	if metrics.ArrowWidth <= 0 {
		invalidVisual("Popover", "Metrics.ArrowWidth", "must be greater than 0")
	}
	if metrics.ArrowHeight <= 0 {
		invalidVisual("Popover", "Metrics.ArrowHeight", "must be greater than 0")
	}
	if metrics.HeadingSize <= 0 {
		invalidVisual("Popover", "Metrics.HeadingSize", "must be greater than 0")
	}
	if metrics.BodyTextSize <= 0 {
		invalidVisual("Popover", "Metrics.BodyTextSize", "must be greater than 0")
	}
	if metrics.SectionGap < 0 {
		invalidVisual("Popover", "Metrics.SectionGap", "must be at least 0")
	}
	if metrics.AnimationScale <= 0 {
		invalidVisual("Popover", "Metrics.AnimationScale", "must be greater than 0")
	}
	if metrics.AnimationScale > 1 {
		invalidVisual("Popover", "Metrics.AnimationScale", "must be at most 1")
	}
	if metrics.EnterDuration <= 0 {
		invalidVisual("Popover", "Metrics.EnterDuration", "must be greater than 0")
	}
	if metrics.ExitDuration <= 0 {
		invalidVisual("Popover", "Metrics.ExitDuration", "must be greater than 0")
	}
	if metrics.AnimationDistance < 0 {
		invalidVisual("Popover", "Metrics.AnimationDistance", "must be at least 0")
	}

	if panel := style.CascadePart(style.State{}, style.PartRoot, visual.Defaults, style.Style{}, style.Style{}, visual.Override); panel.Paint == nil {
		invalidVisual("Popover", "Panel.Paint", "must be provided")
	}
	if panel := style.CascadePart(style.State{}, style.PartRoot, visual.Defaults, style.Style{}, style.Style{}, visual.Override); panel.Paint.Background == nil {
		invalidVisual("Popover", "Panel.Paint.Background", "must be provided")
	}
}
