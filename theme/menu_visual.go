package theme

import (
	"image"
	"image/color"
	"time"

	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/style"
)

// MenuRequest describes the behavior-owned menu density and optional instance
// declaration needed by a visual provider. Item state is resolved by the
// widget at render time and is deliberately not part of this request.
type MenuRequest struct {
	Compact  bool
	Override *MenuOverride
}

// MenuOverride carries caller customization across the provider boundary.
type MenuOverride struct {
	Declaration style.Style
}

// MenuOverrideStyle extracts the provider-owned instance layer from a request.
func MenuOverrideStyle(request MenuRequest) style.Style {
	if request.Override == nil {
		return style.Style{}
	}
	return request.Override.Declaration
}

// MenuMetrics contains neutral geometry used by menu behavior and layout.
// Providers may vary these values by density and theme without changing menu
// selection, focus, keyboard, or overlay behavior.
type MenuMetrics struct {
	Width                      unit.Dp
	MaxHeight                  unit.Dp
	MaxWidthFraction           float32
	Padding                    unit.Dp
	Radius                     unit.Dp
	BorderWidth                unit.Dp
	ItemGap                    unit.Dp
	ItemMinHeight              unit.Dp
	ItemRadius                 unit.Dp
	ItemPaddingX               unit.Dp
	ItemPaddingY               unit.Dp
	ItemContentGap             unit.Dp
	ShortcutHeight             unit.Dp
	ShortcutPaddingX           unit.Dp
	IndicatorSize              unit.Dp
	IndicatorContentGap        unit.Dp
	CheckmarkSize              unit.Dp
	RadioDotSize               unit.Dp
	IndicatorOffsetY           unit.Dp
	SubmenuIndicatorSize       unit.Dp
	FocusRingWidth             unit.Dp
	FocusRingOffset            unit.Dp
	PressedScale               float32
	SectionPaddingX            unit.Dp
	SectionPaddingTop          unit.Dp
	SectionPaddingBottom       unit.Dp
	SeparatorMarginX           unit.Dp
	SeparatorMarginY           unit.Dp
	SeparatorWidth             unit.Dp
	DescriptionLeadingHeight   unit.Dp
	DescriptionLeadingInsetTop unit.Dp
	SubmenuGap                 unit.Dp
	ContextMenuOffset          unit.Dp
	EnterScale                 float32
	ExitScale                  float32
	EnterDuration              time.Duration
	ExitDuration               time.Duration
	SubmenuOpenDelay           time.Duration
	AnimationDistance          unit.Dp
	ShadowOpacity              float32
}

// MenuColors contains semantic colors used by the shared menu behavior.
type MenuColors struct {
	PanelBackground color.NRGBA
	PanelForeground color.NRGBA
	PanelBorder     color.NRGBA
	PanelShadow     color.NRGBA
	Separator       color.NRGBA
	Foreground      color.NRGBA
	Description     color.NRGBA
	Shortcut        color.NRGBA
	Indicator       color.NRGBA
	Danger          color.NRGBA
	Focus           color.NRGBA
}

// MenuPanelColors contains the concrete popup surface colors returned by a
// provider after root and instance declarations have been resolved.
type MenuPanelColors struct {
	Background color.NRGBA
	Border     color.NRGBA
	// Valid is false when the resolved panel uses a paint source that the
	// popup chrome API cannot represent as a concrete color.
	Valid bool
}

// MenuPanelColorsResolver interprets the resolved root Part for popup chrome.
// The widget supplies the cascade result but does not inspect Paint sources.
type MenuPanelColorsResolver func(style.ResolvedStyle) MenuPanelColors

// MenuItemRequest identifies the visual variant of one menu item.
type MenuItemRequest struct {
	Variant  uint8
	Disabled bool
}

// MenuTextRole identifies the semantic text slot rendered or measured by a
// menu. Typography and Gio text configuration remain provider-owned.
type MenuTextRole uint8

const (
	MenuTextEmpty MenuTextRole = iota
	MenuTextSection
	MenuTextItem
	MenuTextDescription
	MenuTextShortcut
)

// MenuTextRequest contains semantic text, its provider-resolved fallback
// color, and the resolved item Part when a caller override affects typography.
type MenuTextRequest struct {
	Text  string
	Color color.NRGBA
	Role  MenuTextRole
	Style style.ResolvedStyle
}

// MenuTextRenderer is the theme-owned Gio text primitive used by menu
// content, section labels, empty states and shortcuts.
type MenuTextRenderer func(layout.Context, MenuTextRequest) layout.Dimensions

// MenuTextWidthResolver measures the same provider-owned text primitive used
// by AutoWidth without exposing typography or text implementation to menu.
type MenuTextWidthResolver func(layout.Context, MenuTextRequest) int

// MenuItemColors contains provider-owned colors for one resolved item Part.
// The widget owns state selection and passes the resolved declaration back to
// the provider instead of interpreting text colors itself.
type MenuItemColors struct {
	Foreground  color.NRGBA
	Description color.NRGBA
	Shortcut    color.NRGBA
	Indicator   color.NRGBA
}

// MenuItemColorsResolver interprets a resolved item Part into the colors used
// by text slots and neutral leading/indicator slots.
type MenuItemColorsResolver func(style.ResolvedStyle) MenuItemColors

// MenuItemVisual contains the declaration layers and provider-owned color
// resolver for one item. The widget owns state selection; the returned
// declarations answer that state through the shared style resolver.
type MenuItemVisual struct {
	Defaults style.Style
	Variant  style.Style
	Colors   MenuItemColorsResolver
}

// MenuIndicatorRequest contains neutral facts needed to paint a checkmark or
// radio dot. Type values are intentionally numeric so theme does not import a
// widget package.
type MenuIndicatorRequest struct {
	Type          uint8
	Size          image.Point
	DotSize       int
	CheckmarkSize int
	OffsetY       int
	Color         color.NRGBA
}

// MenuSeparatorRenderer paints a separator inside the supplied local bounds.
type MenuSeparatorRenderer func(layout.Context, image.Point, color.NRGBA)

// MenuIndicatorRenderer paints a selected item indicator inside the supplied
// local constraints.
type MenuIndicatorRenderer func(layout.Context, MenuIndicatorRequest)

// MenuSubmenuIndicatorRenderer paints a submenu affordance and returns its
// dimensions. Providers must supply this renderer for every validated visual.
type MenuSubmenuIndicatorRenderer func(layout.Context, image.Point, color.NRGBA) layout.Dimensions

// MenuVisual is the complete visual plan consumed by menu behavior.
type MenuVisual struct {
	Metrics          MenuMetrics
	Colors           MenuColors
	Motion           MotionTheme
	Root             style.Style
	Override         style.Style
	PanelColors      MenuPanelColorsResolver
	Text             MenuTextRenderer
	TextWidth        MenuTextWidthResolver
	Item             func(MenuItemRequest) MenuItemVisual
	Separator        MenuSeparatorRenderer
	Indicator        MenuIndicatorRenderer
	SubmenuIndicator MenuSubmenuIndicatorRenderer
}

// MenuProvider supplies menu declarations, geometry, colors, motion, and
// drawing strategies without taking ownership of menu behavior.
type MenuProvider interface {
	Menu(active *Theme, request MenuRequest) MenuVisual
}

// ValidateMenuVisual checks the complete menu declaration before a widget
// stores it in a visual cache. The menu owns navigation and overlay state;
// the provider owns panel/item recipes, typography, geometry, motion, and
// drawing strategies.
func ValidateMenuVisual(visual MenuVisual) {
	metrics := visual.Metrics
	if metrics.Width <= 0 {
		invalidVisual("Menu", "Metrics.Width", "must be greater than 0")
	}
	if metrics.MaxHeight < 0 {
		invalidVisual("Menu", "Metrics.MaxHeight", "must be at least 0")
	}
	if metrics.MaxWidthFraction <= 0 {
		invalidVisual("Menu", "Metrics.MaxWidthFraction", "must be greater than 0")
	}
	if metrics.MaxWidthFraction > 1 {
		invalidVisual("Menu", "Metrics.MaxWidthFraction", "must be at most 1")
	}
	if metrics.Padding < 0 {
		invalidVisual("Menu", "Metrics.Padding", "must be at least 0")
	}
	if metrics.Radius < 0 {
		invalidVisual("Menu", "Metrics.Radius", "must be at least 0")
	}
	if metrics.BorderWidth < 0 {
		invalidVisual("Menu", "Metrics.BorderWidth", "must be at least 0")
	}
	if metrics.ItemGap < 0 {
		invalidVisual("Menu", "Metrics.ItemGap", "must be at least 0")
	}
	if metrics.ItemMinHeight <= 0 {
		invalidVisual("Menu", "Metrics.ItemMinHeight", "must be greater than 0")
	}
	if metrics.ItemRadius < 0 {
		invalidVisual("Menu", "Metrics.ItemRadius", "must be at least 0")
	}
	if metrics.ItemPaddingX < 0 {
		invalidVisual("Menu", "Metrics.ItemPaddingX", "must be at least 0")
	}
	if metrics.ItemPaddingY < 0 {
		invalidVisual("Menu", "Metrics.ItemPaddingY", "must be at least 0")
	}
	if metrics.ItemContentGap < 0 {
		invalidVisual("Menu", "Metrics.ItemContentGap", "must be at least 0")
	}
	if metrics.ShortcutHeight <= 0 {
		invalidVisual("Menu", "Metrics.ShortcutHeight", "must be greater than 0")
	}
	if metrics.ShortcutPaddingX < 0 {
		invalidVisual("Menu", "Metrics.ShortcutPaddingX", "must be at least 0")
	}
	if metrics.IndicatorSize <= 0 {
		invalidVisual("Menu", "Metrics.IndicatorSize", "must be greater than 0")
	}
	if metrics.IndicatorContentGap < 0 {
		invalidVisual("Menu", "Metrics.IndicatorContentGap", "must be at least 0")
	}
	if metrics.CheckmarkSize <= 0 {
		invalidVisual("Menu", "Metrics.CheckmarkSize", "must be greater than 0")
	}
	if metrics.RadioDotSize <= 0 {
		invalidVisual("Menu", "Metrics.RadioDotSize", "must be greater than 0")
	}
	if metrics.SubmenuIndicatorSize <= 0 {
		invalidVisual("Menu", "Metrics.SubmenuIndicatorSize", "must be greater than 0")
	}
	// A zero ring is allowed: the item highlight marks the focused row.
	if metrics.FocusRingWidth < 0 {
		invalidVisual("Menu", "Metrics.FocusRingWidth", "must be at least 0")
	}
	if metrics.FocusRingOffset < 0 {
		invalidVisual("Menu", "Metrics.FocusRingOffset", "must be at least 0")
	}
	if metrics.PressedScale <= 0 {
		invalidVisual("Menu", "Metrics.PressedScale", "must be greater than 0")
	}
	if metrics.PressedScale > 1 {
		invalidVisual("Menu", "Metrics.PressedScale", "must be at most 1")
	}
	if metrics.SectionPaddingX < 0 {
		invalidVisual("Menu", "Metrics.SectionPaddingX", "must be at least 0")
	}
	if metrics.SectionPaddingTop < 0 {
		invalidVisual("Menu", "Metrics.SectionPaddingTop", "must be at least 0")
	}
	if metrics.SectionPaddingBottom < 0 {
		invalidVisual("Menu", "Metrics.SectionPaddingBottom", "must be at least 0")
	}
	if metrics.SeparatorMarginX < 0 {
		invalidVisual("Menu", "Metrics.SeparatorMarginX", "must be at least 0")
	}
	if metrics.SeparatorMarginY < 0 {
		invalidVisual("Menu", "Metrics.SeparatorMarginY", "must be at least 0")
	}
	if metrics.SeparatorWidth <= 0 {
		invalidVisual("Menu", "Metrics.SeparatorWidth", "must be greater than 0")
	}
	if metrics.DescriptionLeadingHeight <= 0 {
		invalidVisual("Menu", "Metrics.DescriptionLeadingHeight", "must be greater than 0")
	}
	if metrics.DescriptionLeadingInsetTop < 0 {
		invalidVisual("Menu", "Metrics.DescriptionLeadingInsetTop", "must be at least 0")
	}
	if metrics.SubmenuGap < 0 {
		invalidVisual("Menu", "Metrics.SubmenuGap", "must be at least 0")
	}
	if metrics.ContextMenuOffset < 0 {
		invalidVisual("Menu", "Metrics.ContextMenuOffset", "must be at least 0")
	}
	if metrics.EnterScale <= 0 {
		invalidVisual("Menu", "Metrics.EnterScale", "must be greater than 0")
	}
	if metrics.EnterScale > 1 {
		invalidVisual("Menu", "Metrics.EnterScale", "must be at most 1")
	}
	if metrics.ExitScale <= 0 {
		invalidVisual("Menu", "Metrics.ExitScale", "must be greater than 0")
	}
	if metrics.ExitScale > 1 {
		invalidVisual("Menu", "Metrics.ExitScale", "must be at most 1")
	}
	if metrics.EnterDuration <= 0 {
		invalidVisual("Menu", "Metrics.EnterDuration", "must be greater than 0")
	}
	if metrics.ExitDuration <= 0 {
		invalidVisual("Menu", "Metrics.ExitDuration", "must be greater than 0")
	}
	if metrics.SubmenuOpenDelay <= 0 {
		invalidVisual("Menu", "Metrics.SubmenuOpenDelay", "must be greater than 0")
	}
	if metrics.AnimationDistance < 0 {
		invalidVisual("Menu", "Metrics.AnimationDistance", "must be at least 0")
	}
	if metrics.ShadowOpacity < 0 {
		invalidVisual("Menu", "Metrics.ShadowOpacity", "must be at least 0")
	}
	if metrics.ShadowOpacity > 1 {
		invalidVisual("Menu", "Metrics.ShadowOpacity", "must be at most 1")
	}

	validateMotion("Menu", "Motion", visual.Motion)
	if visual.PanelColors == nil {
		invalidVisual("Menu", "PanelColors", "must be provided")
	}
	if visual.Text == nil {
		invalidVisual("Menu", "Text", "must be provided")
	}
	if visual.TextWidth == nil {
		invalidVisual("Menu", "TextWidth", "must be provided")
	}
	if visual.Item == nil {
		invalidVisual("Menu", "Item", "must be provided")
	}
	if visual.Separator == nil {
		invalidVisual("Menu", "Separator", "must be provided")
	}
	if visual.Indicator == nil {
		invalidVisual("Menu", "Indicator", "must be provided")
	}
	if visual.SubmenuIndicator == nil {
		invalidVisual("Menu", "SubmenuIndicator", "must be provided")
	}

	root := style.Cascade(style.State{}, visual.Root, visual.Override)
	if root.Paint == nil {
		invalidVisual("Menu", "Root.Paint", "must be provided")
	}
	if !hasPaint(root.Paint.Background) {
		invalidVisual("Menu", "Root.Paint.Background", "must provide a solid color or nonempty gradient")
	}
	if root.Paint.Border == nil {
		invalidVisual("Menu", "Root.Paint.Border", "must be provided")
	}
	if !hasColor(root.Paint.Border.Color) {
		invalidVisual("Menu", "Root.Paint.Border.Color", "must resolve to a solid color")
	}
	if root.Paint.Radius == nil {
		invalidVisual("Menu", "Root.Paint.Radius", "must be provided")
	}
	if root.Text == nil {
		invalidVisual("Menu", "Root.Text", "must be provided")
	}
	if !hasColor(root.Text.Color) {
		invalidVisual("Menu", "Root.Text.Color", "must resolve to a solid color")
	}

	for _, request := range []MenuItemRequest{
		{},
		{Disabled: true},
		{Variant: 1},
		{Variant: 1, Disabled: true},
	} {
		ValidateMenuItemVisual(visual.Item(request))
	}
}

// ValidateMenuItemVisual checks a provider-owned item recipe without applying
// an instance override. Menu overrides may intentionally provide only one
// property and are resolved after this base contract is accepted.
func ValidateMenuItemVisual(visual MenuItemVisual) {
	resolved := style.CascadePart(style.State{}, style.PartItem, visual.Defaults, visual.Variant)
	if resolved.Box == nil {
		invalidVisual("MenuItem", "Resolved.Box", "must be provided")
	}
	if !hasPositiveHeight(resolved.Box.Height, resolved.Box.MinHeight) {
		invalidVisual("MenuItem", "Resolved.Box.Height", "height or minimum height must be finite and greater than 0")
	}
	if resolved.Text == nil {
		invalidVisual("MenuItem", "Resolved.Text", "must be provided")
	}
	if !hasColor(resolved.Text.Color) {
		invalidVisual("MenuItem", "Resolved.Text.Color", "must resolve to a solid color")
	}
	if visual.Colors == nil {
		invalidVisual("MenuItem", "Colors", "must be provided")
	}
}
