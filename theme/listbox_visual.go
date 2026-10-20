package theme

import (
	"image"
	"image/color"
	"time"

	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/style"
)

// ListBoxRequest contains the construction-time visual variant for an item
// collection. The numeric variant is deliberately neutral so the protocol
// does not import widget/listbox or copy its behavior types into theme.
type ListBoxRequest struct {
	Variant  uint8
	Override *ListBoxOverride
}

// ListBoxOverride carries caller customization across the neutral provider
// boundary. The provider owns the returned instance layer.
type ListBoxOverride struct {
	Declaration style.Style
}

// ListBoxOverrideStyle extracts the provider-owned instance layer from a
// request.
func ListBoxOverrideStyle(request ListBoxRequest) style.Style {
	if request.Override == nil {
		return style.Style{}
	}
	return request.Override.Declaration
}

// ListBoxItemRequest contains behavior-owned item state needed to choose the
// visual declaration and semantic colors. The provider never owns selection or
// focus transitions; it only describes their visual result.
type ListBoxItemRequest struct {
	Variant uint8
	State   style.State
}

// ListBoxMetrics contains the neutral geometry and motion values read by the
// listbox layout. Typography and text drawing stay inside ListBoxTextRenderer.
type ListBoxMetrics struct {
	Padding               unit.Dp
	Gap                   unit.Dp
	MaxHeight             unit.Dp
	SectionHeaderPaddingX unit.Dp
	SectionHeaderPaddingY unit.Dp
	ItemMinHeight         unit.Dp
	ItemContentGap        unit.Dp
	ItemIndicatorSize     unit.Dp
	SelectionDuration     time.Duration
}

// ListBoxColors contains the provider-owned color used by listbox text slots
// that do not belong to a concrete item.
type ListBoxColors struct {
	MutedForeground color.NRGBA
}

// ListBoxItemColors is the concrete text and indicator palette for one item.
type ListBoxItemColors struct {
	Foreground  color.NRGBA
	Description color.NRGBA
	Indicator   color.NRGBA
}

// ListBoxItemColorsResolver lets the theme interpret a resolved Item Part.
// The widget owns the state and cascade; the provider owns color meaning.
type ListBoxItemColorsResolver func(style.ResolvedStyle) ListBoxItemColors

// ListBoxTextRole identifies the semantic text slot rendered by a listbox.
// The provider owns the typography and Gio text configuration for each role;
// the widget only supplies the semantic value and its resolved color.
type ListBoxTextRole uint8

const (
	ListBoxTextSectionHeader ListBoxTextRole = iota
	ListBoxTextEmpty
	ListBoxTextItem
	ListBoxTextDescription
)

// ListBoxTextRequest contains the semantic text and provider-resolved color
// for one listbox text slot.
type ListBoxTextRequest struct {
	Text  string
	Color color.NRGBA
	Role  ListBoxTextRole
}

// ListBoxTextRenderer is the theme-owned text primitive used by listbox
// section headers, empty states, labels and descriptions.
type ListBoxTextRenderer func(layout.Context, ListBoxTextRequest) layout.Dimensions

// ListBoxItemVisual is an immutable declaration bundle for one item variant.
type ListBoxItemVisual struct {
	Defaults style.Style
	Variant  style.Style
	Colors   ListBoxItemColorsResolver
	// CacheUnsafe is populated by the widget cache loader from Defaults and
	// Variant. It keeps the hot item path from rescanning immutable declarations
	// every frame; providers should leave it at its zero value.
	CacheUnsafe bool
}

// ListBoxVisual contains the complete visual plan consumed by listbox.
type ListBoxVisual struct {
	Metrics   ListBoxMetrics
	Colors    ListBoxColors
	Motion    MotionTheme
	Override  style.Style
	Text      ListBoxTextRenderer
	Item      func(ListBoxItemRequest) ListBoxItemVisual
	Indicator ListBoxIndicatorRenderer
}

// ListBoxIndicatorRenderer is the theme-owned default selection mark.
type ListBoxIndicatorRenderer func(layout.Context, image.Point, float32, color.NRGBA)

// ListBoxProvider supplies listbox layout metrics, item recipes and drawing
// strategies without taking ownership of listbox behavior.
type ListBoxProvider interface {
	ListBox(active *Theme, request ListBoxRequest) ListBoxVisual
}

// ValidateListBoxVisual checks the complete listbox declaration before a
// widget stores it in a visual cache. Listbox behavior supplies selection,
// focus, and semantic state; the provider must supply every visual primitive
// and the geometry used to present those states.
func ValidateListBoxVisual(visual ListBoxVisual) {
	metrics := visual.Metrics
	if metrics.Padding < 0 {
		invalidVisual("ListBox", "Metrics.Padding", "must be at least 0")
	}
	if metrics.Gap < 0 {
		invalidVisual("ListBox", "Metrics.Gap", "must be at least 0")
	}
	if metrics.MaxHeight < 0 {
		invalidVisual("ListBox", "Metrics.MaxHeight", "must be at least 0")
	}
	if metrics.SectionHeaderPaddingX < 0 {
		invalidVisual("ListBox", "Metrics.SectionHeaderPaddingX", "must be at least 0")
	}
	if metrics.SectionHeaderPaddingY < 0 {
		invalidVisual("ListBox", "Metrics.SectionHeaderPaddingY", "must be at least 0")
	}
	if metrics.ItemMinHeight <= 0 {
		invalidVisual("ListBox", "Metrics.ItemMinHeight", "must be greater than 0")
	}
	if metrics.ItemContentGap < 0 {
		invalidVisual("ListBox", "Metrics.ItemContentGap", "must be at least 0")
	}
	if metrics.ItemIndicatorSize <= 0 {
		invalidVisual("ListBox", "Metrics.ItemIndicatorSize", "must be greater than 0")
	}
	if metrics.SelectionDuration <= 0 {
		invalidVisual("ListBox", "Metrics.SelectionDuration", "must be greater than 0")
	}

	validateMotion("ListBox", "Motion", visual.Motion)
	if visual.Text == nil {
		invalidVisual("ListBox", "Text", "must be provided")
	}
	if visual.Item == nil {
		invalidVisual("ListBox", "Item", "must be provided")
	}
	if visual.Indicator == nil {
		invalidVisual("ListBox", "Indicator", "must be provided")
	}

	// Item declarations are stateful recipes. Validate every state the widget
	// can submit before the root visual enters the shared cache, so a provider
	// cannot defer an incomplete focus or disabled declaration until first use.
	for _, state := range []style.State{
		{},
		{Hovered: true},
		{Pressed: true},
		{Focused: true},
		{FocusVisible: true},
		{Selected: true},
		{Disabled: true},
		{Selected: true, Hovered: true},
		{Selected: true, Pressed: true},
		{Selected: true, FocusVisible: true},
		{Selected: true, Disabled: true},
	} {
		ValidateListBoxItemVisual(visual.Item(ListBoxItemRequest{
			State: state,
		}), state)
	}
}

// ValidateListBoxItemVisual checks a provider-owned item recipe without
// applying an instance override. Overrides are allowed to be partial and are
// resolved by the widget only after this base declaration is accepted.
func ValidateListBoxItemVisual(visual ListBoxItemVisual, state style.State) {
	resolved := style.CascadePart(state, style.PartItem, visual.Defaults, visual.Variant)
	if resolved.Box == nil {
		invalidVisual("ListBoxItem", "Resolved.Box", "must be provided")
	}
	if !hasPositiveHeight(resolved.Box.Height, resolved.Box.MinHeight) {
		invalidVisual("ListBoxItem", "Resolved.Box.Height", "height or minimum height must be finite and greater than 0")
	}
	if resolved.Box.Padding == nil {
		invalidVisual("ListBoxItem", "Resolved.Box.Padding", "must be provided")
	}
	if resolved.Box.Padding.Top < 0 {
		invalidVisual("ListBoxItem", "Resolved.Box.Padding.Top", "must be at least 0")
	}
	if resolved.Box.Padding.Right < 0 {
		invalidVisual("ListBoxItem", "Resolved.Box.Padding.Right", "must be at least 0")
	}
	if resolved.Box.Padding.Bottom < 0 {
		invalidVisual("ListBoxItem", "Resolved.Box.Padding.Bottom", "must be at least 0")
	}
	if resolved.Box.Padding.Left < 0 {
		invalidVisual("ListBoxItem", "Resolved.Box.Padding.Left", "must be at least 0")
	}
	if resolved.Text == nil {
		invalidVisual("ListBoxItem", "Resolved.Text", "must be provided")
	}
	if !hasColor(resolved.Text.Color) {
		invalidVisual("ListBoxItem", "Resolved.Text.Color", "must resolve to a solid color")
	}
	if visual.Colors == nil {
		invalidVisual("ListBoxItem", "Colors", "must be provided")
	}
	if resolved.Paint == nil {
		invalidVisual("ListBoxItem", "Resolved.Paint", "must be provided")
	}
	if !hasPaint(resolved.Paint.Background) {
		invalidVisual("ListBoxItem", "Resolved.Paint.Background", "must provide a solid color or nonempty gradient")
	}
	if resolved.Paint.Border == nil {
		invalidVisual("ListBoxItem", "Resolved.Paint.Border", "must be provided")
	}
	if resolved.Paint.Border.Width == nil {
		invalidVisual("ListBoxItem", "Resolved.Paint.Border.Width", "must be provided")
	}
	if *resolved.Paint.Border.Width < 0 {
		invalidVisual("ListBoxItem", "Resolved.Paint.Border.Width", "must be at least 0")
	}
	if !hasColor(resolved.Paint.Border.Color) {
		invalidVisual("ListBoxItem", "Resolved.Paint.Border.Color", "must resolve to a solid color")
	}
	if resolved.Paint.Radius == nil {
		invalidVisual("ListBoxItem", "Resolved.Paint.Radius", "must be provided")
	}
	if *resolved.Paint.Radius < 0 {
		invalidVisual("ListBoxItem", "Resolved.Paint.Radius", "must be at least 0")
	}
	if !hasValidOpacity(resolved.Paint.Opacity) {
		invalidVisual("ListBoxItem", "Resolved.Paint.Opacity", "must be finite and between 0 and 1")
	}
	if !hasOptionalScale(resolved.Trans) {
		invalidVisual("ListBoxItem", "Resolved.Trans", "scale axes must be finite and greater than 0")
	}

	if state.FocusVisible {
		if resolved.Paint.Outline != nil {
			if resolved.Paint.Outline.Width <= 0 {
				invalidVisual("ListBoxItem", "Resolved.Paint.Outline.Width", "must be greater than 0")
			}
			if !hasColor(resolved.Paint.Outline.Color) {
				invalidVisual("ListBoxItem", "Resolved.Paint.Outline.Color", "must resolve to a solid color")
			}
		} else if resolved.Paint.Border == nil {
			invalidVisual("ListBoxItem", "Resolved.Paint.Border", "must be provided")
		}
		if resolved.Paint.Border.Width == nil {
			invalidVisual("ListBoxItem", "Resolved.Paint.Border.Width", "must be provided")
		}
		if *resolved.Paint.Border.Width <= 0 {
			invalidVisual("ListBoxItem", "Resolved.Paint.Border.Width", "must be greater than 0")
		}
		if !hasColor(resolved.Paint.Border.Color) {
			invalidVisual("ListBoxItem", "Resolved.Paint.Border.Color", "must resolve to a solid color")
		}
	}
}
