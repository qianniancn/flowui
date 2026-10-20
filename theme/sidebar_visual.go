package theme

import (
	"image"
	"image/color"
	"math"

	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/style"
)

// Apply applies the behavior-owned per-instance geometry to provider metrics.
func (r SidebarRequest) Apply(metrics SidebarMetrics) SidebarMetrics {
	if r.MetricsMask&SidebarMetricWidth != 0 {
		metrics.Width = r.Metrics.Width
	}
	if r.MetricsMask&SidebarMetricCollapsedWidth != 0 {
		metrics.CollapsedWidth = r.Metrics.CollapsedWidth
	}
	if r.MetricsMask&SidebarMetricPadding != 0 {
		metrics.Padding = r.Metrics.Padding
	}
	if r.MetricsMask&SidebarMetricContentGap != 0 {
		metrics.ContentGap = r.Metrics.ContentGap
	}
	if r.MetricsMask&SidebarMetricItemGap != 0 {
		metrics.ItemGap = r.Metrics.ItemGap
	}
	if r.MetricsMask&SidebarMetricItemHeight != 0 {
		metrics.ItemHeight = r.Metrics.ItemHeight
	}
	if r.MetricsMask&SidebarMetricItemRadius != 0 {
		metrics.ItemRadius = r.Metrics.ItemRadius
	}
	if r.MetricsMask&SidebarMetricItemPaddingX != 0 {
		metrics.ItemPaddingX = r.Metrics.ItemPaddingX
	}
	if r.MetricsMask&SidebarMetricItemContentGap != 0 {
		metrics.ItemContentGap = r.Metrics.ItemContentGap
	}
	if r.MetricsMask&SidebarMetricInlineIndent != 0 {
		metrics.InlineIndent = r.Metrics.InlineIndent
	}
	return metrics
}

// SidebarRequest describes the construction-time visual variant for a
// navigation sidebar. The widget owns navigation state; the provider owns
// every visual declaration and renderer used to present that state.
type SidebarRequest struct {
	Override    *SidebarOverride
	Metrics     SidebarMetrics
	MetricsMask SidebarMetricsMask
}

// SidebarMetricsMask identifies per-instance geometry values supplied by a
// behavior widget. Keeping the values in the neutral request lets the
// provider remain the owner of the resulting visual scale while preserving a
// stable, comparable cache key.
type SidebarMetricsMask uint16

const (
	SidebarPartRoot  style.Part = style.PartRoot
	SidebarPartItem  style.Part = style.PartItem
	SidebarPartLabel style.Part = style.PartLabel
)

const (
	SidebarMetricWidth SidebarMetricsMask = 1 << iota
	SidebarMetricCollapsedWidth
	SidebarMetricPadding
	SidebarMetricContentGap
	SidebarMetricItemGap
	SidebarMetricItemHeight
	SidebarMetricItemRadius
	SidebarMetricItemPaddingX
	SidebarMetricItemContentGap
	SidebarMetricInlineIndent
)

// SidebarOverride carries caller customization across the neutral provider
// boundary.
type SidebarOverride struct {
	Declaration style.Style
}

// SidebarOverrideStyle extracts the provider-owned instance layer.
func SidebarOverrideStyle(request SidebarRequest) style.Style {
	if request.Override == nil {
		return style.Style{}
	}
	return request.Override.Declaration
}

// SidebarMetrics contains all geometry and typography consumed by sidebar
// behavior layout. It intentionally has no colors or drawing policy.
type SidebarMetrics struct {
	Width                 unit.Dp
	CollapsedWidth        unit.Dp
	Padding               unit.Dp
	ContentGap            unit.Dp
	ItemGap               unit.Dp
	ItemHeight            unit.Dp
	ItemRadius            unit.Dp
	ItemPaddingX          unit.Dp
	ItemContentGap        unit.Dp
	InlineIndent          unit.Dp
	SwitcherSize          unit.Dp
	ItemTextSize          unit.Sp
	SectionTextSize       unit.Sp
	SectionHeight         unit.Dp
	SectionSeparatorInset unit.Dp
	BorderWidth           unit.Dp
	FocusRingWidth        unit.Dp
}

// SidebarItemRequest contains behavior-owned interaction state used by the
// provider to choose a visual result.
type SidebarItemRequest struct {
	Selected bool
	Hovered  bool
	Pressed  bool
	Focused  bool
	Disabled bool
}

// SidebarItemVisual is the resolved visual state of one navigation item.
// Widget code may pass it to the provider renderer but does not choose any of
// these values itself.
type SidebarItemVisual struct {
	Background   color.NRGBA
	Foreground   color.NRGBA
	Focus        color.NRGBA
	Opacity      float32
	Radius       unit.Dp
	Border       color.NRGBA
	BorderWidth  unit.Dp
	Outline      color.NRGBA
	OutlineWidth unit.Dp
}

// SidebarTextRequest contains semantic text content and provider-resolved
// color. Initial selects the compact collapsed-sidebar label treatment.
type SidebarTextRequest struct {
	Text    string
	Color   color.NRGBA
	Size    unit.Sp
	Initial bool
}

// SidebarSectionRequest describes a section label or collapsed separator.
type SidebarSectionRequest struct {
	Title     string
	Collapsed bool
}

// SidebarSwitcherRequest describes the provider-owned default group indicator.
// Custom item switcher widgets remain behavior-owned content and bypass this
// renderer.
type SidebarSwitcherRequest struct {
	Expanded   bool
	Foreground color.NRGBA
	Size       unit.Dp
}

// SidebarRootRenderer and SidebarItemRenderer keep painting in the theme
// recipe while the widget retains only layout and interaction behavior.
type SidebarRootRenderer func(layout.Context, image.Point)

type SidebarItemRenderer func(layout.Context, image.Point, SidebarItemVisual, float32)

type SidebarTextRenderer func(layout.Context, SidebarTextRequest) layout.Dimensions

type SidebarSectionRenderer func(layout.Context, SidebarSectionRequest) layout.Dimensions

type SidebarSwitcherRenderer func(layout.Context, SidebarSwitcherRequest) layout.Dimensions

// SidebarVisual is the complete visual plan consumed by sidebar behavior.
type SidebarVisual struct {
	Metrics         SidebarMetrics
	Foreground      color.NRGBA
	Background      color.NRGBA
	MutedForeground color.NRGBA
	Motion          MotionTheme
	ExpansionEasing EaseKind
	CollapseEasing  EaseKind
	Override        style.Style
	Item            func(SidebarItemRequest) SidebarItemVisual
	Root            SidebarRootRenderer
	RenderItem      SidebarItemRenderer
	Text            SidebarTextRenderer
	Section         SidebarSectionRenderer
	Switcher        SidebarSwitcherRenderer
}

// SidebarProvider supplies sidebar metrics, state visuals, and theme-owned
// rendering strategies without taking ownership of navigation behavior.
type SidebarProvider interface {
	Sidebar(active *Theme, request SidebarRequest) SidebarVisual
}

// ValidateSidebarVisual checks the complete sidebar renderer and metric
// contract before a widget stores a provider result in its visual cache.
// Sidebar behavior owns navigation state and semantic focus; the provider
// must supply every visual renderer and all geometry used to present it.
func ValidateSidebarVisual(visual SidebarVisual) {
	metrics := visual.Metrics
	if metrics.Width <= 0 {
		invalidVisual("Sidebar", "Metrics.Width", "must be greater than 0")
	}
	if metrics.CollapsedWidth <= 0 {
		invalidVisual("Sidebar", "Metrics.CollapsedWidth", "must be greater than 0")
	}
	if metrics.Padding < 0 {
		invalidVisual("Sidebar", "Metrics.Padding", "must be at least 0")
	}
	if metrics.ContentGap < 0 {
		invalidVisual("Sidebar", "Metrics.ContentGap", "must be at least 0")
	}
	if metrics.ItemGap < 0 {
		invalidVisual("Sidebar", "Metrics.ItemGap", "must be at least 0")
	}
	if metrics.ItemHeight <= 0 {
		invalidVisual("Sidebar", "Metrics.ItemHeight", "must be greater than 0")
	}
	if metrics.ItemRadius < 0 {
		invalidVisual("Sidebar", "Metrics.ItemRadius", "must be at least 0")
	}
	if metrics.ItemPaddingX < 0 {
		invalidVisual("Sidebar", "Metrics.ItemPaddingX", "must be at least 0")
	}
	if metrics.ItemContentGap < 0 {
		invalidVisual("Sidebar", "Metrics.ItemContentGap", "must be at least 0")
	}
	if metrics.InlineIndent < 0 {
		invalidVisual("Sidebar", "Metrics.InlineIndent", "must be at least 0")
	}
	if metrics.SwitcherSize <= 0 {
		invalidVisual("Sidebar", "Metrics.SwitcherSize", "must be greater than 0")
	}
	if metrics.ItemTextSize <= 0 {
		invalidVisual("Sidebar", "Metrics.ItemTextSize", "must be greater than 0")
	}
	if metrics.SectionTextSize <= 0 {
		invalidVisual("Sidebar", "Metrics.SectionTextSize", "must be greater than 0")
	}
	if metrics.SectionHeight <= 0 {
		invalidVisual("Sidebar", "Metrics.SectionHeight", "must be greater than 0")
	}
	if metrics.SectionSeparatorInset < 0 {
		invalidVisual("Sidebar", "Metrics.SectionSeparatorInset", "must be at least 0")
	}
	if metrics.BorderWidth < 0 {
		invalidVisual("Sidebar", "Metrics.BorderWidth", "must be at least 0")
	}
	if metrics.FocusRingWidth <= 0 {
		invalidVisual("Sidebar", "Metrics.FocusRingWidth", "must be greater than 0")
	}

	validateMotion("Sidebar", "Motion", visual.Motion)
	if visual.Item == nil {
		invalidVisual("Sidebar", "Item", "must be provided")
	}
	if visual.Root == nil {
		invalidVisual("Sidebar", "Root", "must be provided")
	}
	if visual.RenderItem == nil {
		invalidVisual("Sidebar", "RenderItem", "must be provided")
	}
	if visual.Text == nil {
		invalidVisual("Sidebar", "Text", "must be provided")
	}
	if visual.Section == nil {
		invalidVisual("Sidebar", "Section", "must be provided")
	}
	if visual.Switcher == nil {
		invalidVisual("Sidebar", "Switcher", "must be provided")
	}

	if visual.ExpansionEasing > EaseBack {
		invalidVisual("Sidebar", "ExpansionEasing", "must be at most EaseBack")
	}
	if visual.CollapseEasing > EaseBack {
		invalidVisual("Sidebar", "CollapseEasing", "must be at most EaseBack")
	}
}

// ValidateSidebarItemVisual checks a resolved provider-owned item result.
// Instance declarations are applied after this check, so partial overrides do
// not weaken the base visual contract.
func ValidateSidebarItemVisual(visual SidebarItemVisual) {
	if math.IsNaN(float64(visual.Opacity)) {
		invalidVisual("SidebarItem", "Opacity", "must be finite")
	}
	if math.IsInf(float64(visual.Opacity), 0) {
		invalidVisual("SidebarItem", "Opacity", "must be finite")
	}
	if visual.Opacity < 0 {
		invalidVisual("SidebarItem", "Opacity", "must be at least 0")
	}
	if visual.Opacity > 1 {
		invalidVisual("SidebarItem", "Opacity", "must be at most 1")
	}
	if visual.Radius < 0 {
		invalidVisual("SidebarItem", "Radius", "must be at least 0")
	}
	if visual.BorderWidth < 0 {
		invalidVisual("SidebarItem", "BorderWidth", "must be at least 0")
	}
	if visual.OutlineWidth < 0 {
		invalidVisual("SidebarItem", "OutlineWidth", "must be at least 0")
	}
}
