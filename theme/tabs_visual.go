package theme

import (
	"image"
	"image/color"
	"time"

	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/style"
)

// TabsRequest describes the semantic axes and optional instance declaration
// needed by a tabs visual provider. Selection, focus, keyboard navigation and
// panel lifetime remain owned by the widget.
type TabsRequest struct {
	Variant     uint8
	Orientation uint8
	Size        uint8
	Color       uint8
	Override    *TabsOverride
}

// TabsOverride carries caller customization across the provider boundary.
type TabsOverride struct {
	Declaration style.Style
}

// TabsOverrideStyle extracts the provider-owned instance layer.
func TabsOverrideStyle(request TabsRequest) style.Style {
	if request.Override == nil {
		return style.Style{}
	}
	return request.Override.Declaration
}

// TabsMetrics contains geometry and motion values used by tabs layout. It is
// intentionally a neutral copy of the public measurements rather than a
// reference to the theme's component token bag.
type TabsMetrics struct {
	RootGap            unit.Dp
	ListPadding        unit.Dp
	TabHeight          unit.Dp
	SmallTabHeight     unit.Dp
	LargeTabHeight     unit.Dp
	TabMinWidth        unit.Dp
	TabPaddingX        unit.Dp
	SmallTabPaddingX   unit.Dp
	LargeTabPaddingX   unit.Dp
	TabGap             unit.Dp
	IconSize           unit.Dp
	IconGap            unit.Dp
	CloseButtonSize    unit.Dp
	CloseButtonGap     unit.Dp
	ExtraContentGap    unit.Dp
	IndicatorLineWidth unit.Dp
	IndicatorWidth     unit.Dp
	IndicatorMinWidth  unit.Dp
	IndicatorInset     unit.Dp
	PanelGap           unit.Dp
	ColorDuration      time.Duration
	IndicatorDuration  time.Duration
	PanelDuration      time.Duration
	ScrollButtonSize   unit.Dp
	ScrollButtonInset  unit.Dp
}

// TabsColors contains provider-owned semantic colors for tabs chrome. The
// item resolver supplies state-sensitive colors while the widget owns only
// interpolation progress and interaction state.
type TabsColors struct {
	ListBackground     color.NRGBA
	ScrollBackground   color.NRGBA
	ListBorder         color.NRGBA
	Foreground         color.NRGBA
	SelectedForeground color.NRGBA
	Indicator          color.NRGBA
	Separator          color.NRGBA
	Focus              color.NRGBA
	PanelBackground    color.NRGBA
	PanelBorder        color.NRGBA
	ScrollForeground   color.NRGBA
	DisabledForeground color.NRGBA
}

// TabsItemRequest identifies one item's visual state and the interpolated
// foreground selected by tabs behavior.
type TabsItemRequest struct {
	State      style.State
	Foreground color.NRGBA
}

// TabsItemVisual contains the provider-owned item declaration and state
// colors used by indicator/separator/focus drawing.
type TabsItemVisual struct {
	Declaration        style.Style
	Foreground         color.NRGBA
	SelectedForeground color.NRGBA
	Indicator          color.NRGBA
	Separator          color.NRGBA
	Focus              color.NRGBA
}

// TabsItemVisualResolver returns a complete item visual for the supplied
// interaction state. Themes may use it for disabled or hover treatment.
type TabsItemVisualResolver func(TabsItemRequest) TabsItemVisual

// TabsIndicatorRenderRequest contains the behavior-computed indicator bounds
// and neutral axis/variant facts needed by a theme renderer.
type TabsIndicatorRenderRequest struct {
	Bounds      image.Rectangle
	Orientation uint8
	Variant     uint8
	Color       color.NRGBA
}

// TabsIndicatorRenderer paints the selected-tab indicator inside the supplied
// local operation stream. Selection animation and target geometry stay in the
// widget; indicator shape and stroke ownership stay in the theme.
type TabsIndicatorRenderer func(layout.Context, TabsIndicatorRenderRequest)

// TabsFocusRenderRequest contains the resolved focus opacity and item bounds
// needed by a theme-owned focus ring renderer.
type TabsFocusRenderRequest struct {
	Size    image.Point
	Variant uint8
	Opacity float32
	Color   color.NRGBA
}

// TabsFocusRenderer paints focus indication without owning focus state.
type TabsFocusRenderer func(layout.Context, TabsFocusRenderRequest)

// TabsSeparatorRenderRequest contains the neutral slot facts for an optional
// separator between neighboring tab items.
type TabsSeparatorRenderRequest struct {
	Size        image.Point
	Orientation uint8
	Color       color.NRGBA
}

// TabsSeparatorRenderer paints a visual separator while widget owns whether
// the separator should exist for the current item state.
type TabsSeparatorRenderer func(layout.Context, TabsSeparatorRenderRequest)

// TabsListRenderRequest contains the neutral facts needed to paint the tab
// list surface. The widget owns the list bounds and scrolling state; the
// theme owns the surface shape and border treatment.
type TabsListRenderRequest struct {
	Size        image.Point
	Orientation uint8
	Variant     uint8
	Background  color.NRGBA
	Border      color.NRGBA
}

// TabsListRenderer paints the tab list chrome without owning list behavior.
type TabsListRenderer func(layout.Context, TabsListRenderRequest)

// TabsScrollRenderRequest contains the visible overflow edges and list chrome
// colors used by a theme-owned shadow/border renderer.
type TabsScrollRenderRequest struct {
	Size        image.Point
	Orientation uint8
	Variant     uint8
	Previous    bool
	Next        bool
	Background  color.NRGBA
	Border      color.NRGBA
}

// TabsScrollRenderer paints overflow shadows and their optional edge border.
// The widget decides whether an edge is visible; shadow geometry stays in the
// theme so it cannot become an implicit default visual in widget/tabs.
type TabsScrollRenderer func(layout.Context, TabsScrollRenderRequest)

// TabsChevronRenderRequest contains the neutral facts for an overflow button
// glyph. Direction and orientation are behavior facts; glyph choice and icon
// geometry belong to the theme.
type TabsChevronRenderRequest struct {
	Size        image.Point
	Orientation uint8
	Direction   int
	Color       color.NRGBA
}

// TabsChevronRenderer paints an overflow button glyph.
type TabsChevronRenderer func(layout.Context, TabsChevronRenderRequest)

// TabsCloseRenderRequest contains the neutral facts for a closable-tab glyph.
type TabsCloseRenderRequest struct {
	Size  image.Point
	Color color.NRGBA
}

// TabsCloseRenderer paints the close glyph for a closable tab.
type TabsCloseRenderer func(layout.Context, TabsCloseRenderRequest)

// TabsAddRenderRequest contains the neutral facts for the add-tab glyph.
type TabsAddRenderRequest struct {
	Size  image.Point
	Color color.NRGBA
}

// TabsAddRenderer paints the add-tab affordance.
type TabsAddRenderer func(layout.Context, TabsAddRenderRequest)

// TabsOverflowRenderRequest contains the neutral facts for the overflow-menu
// trigger glyph.
type TabsOverflowRenderRequest struct {
	Size  image.Point
	Color color.NRGBA
}

// TabsOverflowRenderer paints the overflow-menu trigger affordance.
type TabsOverflowRenderer func(layout.Context, TabsOverflowRenderRequest)

// TabsVisual is the complete visual plan consumed by tabs behavior.
type TabsVisual struct {
	Root      style.Style
	Panel     style.Style
	Override  style.Style
	Metrics   TabsMetrics
	Colors    TabsColors
	Motion    MotionTheme
	Item      TabsItemVisualResolver
	Indicator TabsIndicatorRenderer
	Focus     TabsFocusRenderer
	Separator TabsSeparatorRenderer
	List      TabsListRenderer
	Scroll    TabsScrollRenderer
	Chevron   TabsChevronRenderer
	Close     TabsCloseRenderer
	Add       TabsAddRenderer
	Overflow  TabsOverflowRenderer
}

// TabsProvider supplies tabs declarations, geometry, colors, and motion
// without taking ownership of selection, focus, keyboard, or panel behavior.
type TabsProvider interface {
	Tabs(active *Theme, request TabsRequest) TabsVisual
}

// ValidateTabsVisual checks the provider-owned declarations and geometry
// before a widget stores a tabs visual in its cache. Tabs behavior supplies
// selection and interaction state; it cannot recover missing chrome locally.
func ValidateTabsVisual(visual TabsVisual) {
	m := visual.Metrics
	if m.RootGap < 0 {
		invalidVisual("Tabs", "Metrics.RootGap", "must be at least 0")
	}
	if m.ListPadding < 0 {
		invalidVisual("Tabs", "Metrics.ListPadding", "must be at least 0")
	}
	if m.TabHeight <= 0 {
		invalidVisual("Tabs", "Metrics.TabHeight", "must be greater than 0")
	}
	if m.SmallTabHeight <= 0 {
		invalidVisual("Tabs", "Metrics.SmallTabHeight", "must be greater than 0")
	}
	if m.LargeTabHeight <= 0 {
		invalidVisual("Tabs", "Metrics.LargeTabHeight", "must be greater than 0")
	}
	if m.TabMinWidth < 0 {
		invalidVisual("Tabs", "Metrics.TabMinWidth", "must be at least 0")
	}
	if m.TabPaddingX < 0 {
		invalidVisual("Tabs", "Metrics.TabPaddingX", "must be at least 0")
	}
	if m.SmallTabPaddingX < 0 {
		invalidVisual("Tabs", "Metrics.SmallTabPaddingX", "must be at least 0")
	}
	if m.LargeTabPaddingX < 0 {
		invalidVisual("Tabs", "Metrics.LargeTabPaddingX", "must be at least 0")
	}
	if m.TabGap < 0 {
		invalidVisual("Tabs", "Metrics.TabGap", "must be at least 0")
	}
	if m.IconSize <= 0 {
		invalidVisual("Tabs", "Metrics.IconSize", "must be greater than 0")
	}
	if m.IconGap < 0 {
		invalidVisual("Tabs", "Metrics.IconGap", "must be at least 0")
	}
	if m.CloseButtonSize <= 0 {
		invalidVisual("Tabs", "Metrics.CloseButtonSize", "must be greater than 0")
	}
	if m.CloseButtonGap < 0 {
		invalidVisual("Tabs", "Metrics.CloseButtonGap", "must be at least 0")
	}
	if m.ExtraContentGap < 0 {
		invalidVisual("Tabs", "Metrics.ExtraContentGap", "must be at least 0")
	}
	if m.IndicatorLineWidth <= 0 {
		invalidVisual("Tabs", "Metrics.IndicatorLineWidth", "must be greater than 0")
	}
	if m.IndicatorWidth < 0 {
		invalidVisual("Tabs", "Metrics.IndicatorWidth", "must be at least 0")
	}
	if m.IndicatorMinWidth < 0 {
		invalidVisual("Tabs", "Metrics.IndicatorMinWidth", "must be at least 0")
	}
	if m.IndicatorInset < 0 {
		invalidVisual("Tabs", "Metrics.IndicatorInset", "must be at least 0")
	}
	if m.PanelGap < 0 {
		invalidVisual("Tabs", "Metrics.PanelGap", "must be at least 0")
	}
	if m.ColorDuration <= 0 {
		invalidVisual("Tabs", "Metrics.ColorDuration", "must be greater than 0")
	}
	if m.IndicatorDuration <= 0 {
		invalidVisual("Tabs", "Metrics.IndicatorDuration", "must be greater than 0")
	}
	if m.PanelDuration <= 0 {
		invalidVisual("Tabs", "Metrics.PanelDuration", "must be greater than 0")
	}
	if m.ScrollButtonSize <= 0 {
		invalidVisual("Tabs", "Metrics.ScrollButtonSize", "must be greater than 0")
	}
	if m.ScrollButtonInset < 0 {
		invalidVisual("Tabs", "Metrics.ScrollButtonInset", "must be at least 0")
	}
	if visual.Item == nil {
		invalidVisual("Tabs", "Item", "must be provided")
	}
	if visual.Indicator == nil {
		invalidVisual("Tabs", "Indicator", "must be provided")
	}
	if visual.Focus == nil {
		invalidVisual("Tabs", "Focus", "must be provided")
	}
	if visual.Separator == nil {
		invalidVisual("Tabs", "Separator", "must be provided")
	}
	if visual.List == nil {
		invalidVisual("Tabs", "List", "must be provided")
	}
	if visual.Scroll == nil {
		invalidVisual("Tabs", "Scroll", "must be provided")
	}
	if visual.Chevron == nil {
		invalidVisual("Tabs", "Chevron", "must be provided")
	}
	if visual.Close == nil {
		invalidVisual("Tabs", "Close", "must be provided")
	}
	if visual.Add == nil {
		invalidVisual("Tabs", "Add", "must be provided")
	}
	if visual.Overflow == nil {
		invalidVisual("Tabs", "Overflow", "must be provided")
	}

	root := style.Cascade(style.State{}, visual.Root, visual.Override)
	if root.Text == nil {
		invalidVisual("Tabs", "Root.Text", "must be provided")
	}
	if !hasColor(root.Text.Color) {
		invalidVisual("Tabs", "Root.Text.Color", "must resolve to a solid color")
	}

	if !hasColor(style.Solid(visual.Colors.ScrollBackground)) {
		invalidVisual("Tabs", "Colors.ScrollBackground", "must resolve to a solid color")
	}
	if !hasColor(style.Solid(visual.Colors.Foreground)) {
		invalidVisual("Tabs", "Colors.Foreground", "must resolve to a solid color")
	}
	if !hasColor(style.Solid(visual.Colors.ScrollForeground)) {
		invalidVisual("Tabs", "Colors.ScrollForeground", "must resolve to a solid color")
	}
	if !hasColor(style.Solid(visual.Colors.DisabledForeground)) {
		invalidVisual("Tabs", "Colors.DisabledForeground", "must resolve to a solid color")
	}

	panel := style.CascadePart(style.State{Selected: true}, style.PartPanel, visual.Panel, visual.Override)
	if panel.Box == nil {
		invalidVisual("Tabs", "Panel.Box", "must be provided")
	}
	if panel.Box.Padding == nil {
		invalidVisual("Tabs", "Panel.Box.Padding", "must be provided")
	}
	if panel.Paint == nil {
		invalidVisual("Tabs", "Panel.Paint", "must be provided")
	}

	for _, request := range []TabsItemRequest{
		{Foreground: visual.Colors.Foreground},
		{State: style.State{Hovered: true}, Foreground: visual.Colors.Foreground},
		{State: style.State{Selected: true}, Foreground: visual.Colors.SelectedForeground},
		{State: style.State{Disabled: true}, Foreground: visual.Colors.Foreground},
	} {
		item := visual.Item(request)
		resolved := style.CascadePart(request.State, style.PartItem, item.Declaration, visual.Override)
		if resolved.Text == nil {
			invalidVisual("Tabs", "Resolved.Text", "must be provided")
		}
		if !hasColor(resolved.Text.Color) {
			invalidVisual("Tabs", "Resolved.Text.Color", "must resolve to a solid color")
		}
		if resolved.Text.FontSize == nil {
			invalidVisual("Tabs", "Resolved.Text.FontSize", "must be provided")
		}
		if !validPositiveSp(*resolved.Text.FontSize) {
			invalidVisual("Tabs", "Resolved.Text.FontSize", "must be finite and greater than 0")
		}
		if resolved.Text.FontWeight == nil {
			invalidVisual("Tabs", "Resolved.Text.FontWeight", "must be provided")
		}
	}
	validateMotion("Tabs", "Motion", visual.Motion)
}
