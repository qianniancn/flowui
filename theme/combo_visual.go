package theme

import (
	"image"
	"image/color"
	"time"

	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/style"
)

// ComboRequest describes the semantic field treatment needed by a combo box.
type ComboRequest struct {
	Variant  InputVariant
	Override *ComboOverride
}

// ComboOverride carries caller customization across the neutral provider
// boundary. The provider owns the returned instance layer.
type ComboOverride struct {
	Declaration style.Style
}

// ComboOverrideStyle extracts the provider-owned instance layer from a
// request.
func ComboOverrideStyle(request ComboRequest) style.Style {
	if request.Override == nil {
		return style.Style{}
	}
	return request.Override.Declaration
}

// ComboBox part names are stable semantic customization points shared by the
// behavior widget and visual providers. ComboPartPanel is also the popup
// overlay surface: the overlay lifecycle belongs to the widget, while its
// visual declaration belongs to the provider.
const (
	ComboPartTrigger     style.Part = style.PartContent
	ComboPartPlaceholder style.Part = style.PartPlaceholder
	ComboPartSelection   style.Part = style.PartSelection
	ComboPartIndicator   style.Part = style.PartIndicator
	ComboPartIcon        style.Part = style.PartIcon
	ComboPartPanel       style.Part = style.PartPanel
	ComboPartOverlay     style.Part = ComboPartPanel
	ComboPartItem        style.Part = style.PartItem
	ComboPartDescription style.Part = style.PartDescription
)

// ComboVisual contains the provider-owned declaration layers consumed by
// combo behavior. The request variant is resolved before this bundle is
// returned; only the default and instance layers cross the behavior boundary.
type ComboVisual struct {
	Defaults       style.Style
	Override       style.Style
	Metrics        ComboMetrics
	Motion         MotionTheme
	Chevron        ComboChevronRenderer
	Check          ComboCheckRenderer
	Text           ComboTextRenderer
	EditorStyle    ComboEditorStyleResolver
	IndicatorWidth ComboPartWidthResolver
	IconSize       ComboPartSizeResolver
}

// ComboMetrics contains visual geometry and motion values needed by the
// shared ComboBox behavior. It is returned by the active provider so widgets
// do not reach into a concrete theme's component token table.
type ComboMetrics struct {
	PanelGap           unit.Dp
	PanelMaxHeight     unit.Dp
	ItemHeight         unit.Dp
	ItemTextSize       unit.Sp
	AnimationDuration  time.Duration
	ItemSelectDuration time.Duration
	IconDuration       time.Duration
	AnimationScale     float32
}

// ComboTextRequest contains the semantic text and its resolved declaration.
// The provider owns typography and color interpretation.
type ComboTextRequest struct {
	Text  string
	Style style.ResolvedStyle
	Size  unit.Sp
}

// ComboTextRenderer is the theme-owned text primitive used by trigger
// support text, empty states, and option rows.
type ComboTextRenderer func(layout.Context, ComboTextRequest) layout.Dimensions

// ComboEditorStyle is the provider-owned bridge to Gio's stateful editor.
// Combo behavior applies these values to the editor adapter without deriving
// them from ResolvedStyle itself.
type ComboEditorStyle struct {
	TextSize    unit.Sp
	LineHeight  unit.Sp
	Foreground  color.NRGBA
	Placeholder color.NRGBA
	Selection   color.NRGBA
}

// ComboEditorStyleResolver resolves the stateful editor colors and typography
// from provider-owned declarations.
type ComboEditorStyleResolver func(content, placeholder, selection style.ResolvedStyle) ComboEditorStyle

// ComboPartWidthResolver and ComboPartSizeResolver keep provider-owned part
// geometry out of the behavior widget while preserving instance overrides.
type ComboPartWidthResolver func(style.ResolvedStyle) unit.Dp

type ComboPartSizeResolver func(layout.Context, style.ResolvedStyle) image.Point

// ComboChevronRenderer and ComboCheckRenderer are theme-owned drawing
// strategies for the ComboBox trigger and selection mark. The widget owns
// when these parts are shown; the provider owns how they are drawn.
type ComboChevronRenderer func(layout.Context, image.Point, float32, style.ResolvedStyle)

type ComboCheckRenderer func(layout.Context, image.Point, float32, style.ResolvedStyle)

// ComboProvider supplies combo trigger, panel, item, and description visuals.
type ComboProvider interface {
	Combo(active *Theme, request ComboRequest) ComboVisual
}

// ValidateComboVisual checks the complete declaration and renderer contract
// before a widget stores a provider result in its cache.
func ValidateComboVisual(visual ComboVisual) {
	if visual.Chevron == nil {
		invalidVisual("Combo", "Chevron", "must be provided")
	}
	if visual.Check == nil {
		invalidVisual("Combo", "Check", "must be provided")
	}
	if visual.Text == nil {
		invalidVisual("Combo", "Text", "must be provided")
	}
	if visual.EditorStyle == nil {
		invalidVisual("Combo", "EditorStyle", "must be provided")
	}
	if visual.IndicatorWidth == nil {
		invalidVisual("Combo", "IndicatorWidth", "must be provided")
	}
	if visual.IconSize == nil {
		invalidVisual("Combo", "IconSize", "must be provided")
	}

	metrics := visual.Metrics
	if metrics.PanelGap < 0 {
		invalidVisual("Combo", "Metrics.PanelGap", "must be at least 0")
	}
	if metrics.PanelMaxHeight <= 0 {
		invalidVisual("Combo", "Metrics.PanelMaxHeight", "must be greater than 0")
	}
	if metrics.ItemHeight <= 0 {
		invalidVisual("Combo", "Metrics.ItemHeight", "must be greater than 0")
	}
	if metrics.ItemTextSize <= 0 {
		invalidVisual("Combo", "Metrics.ItemTextSize", "must be greater than 0")
	}
	if metrics.AnimationDuration <= 0 {
		invalidVisual("Combo", "Metrics.AnimationDuration", "must be greater than 0")
	}
	if metrics.ItemSelectDuration <= 0 {
		invalidVisual("Combo", "Metrics.ItemSelectDuration", "must be greater than 0")
	}
	if metrics.IconDuration <= 0 {
		invalidVisual("Combo", "Metrics.IconDuration", "must be greater than 0")
	}
	if metrics.AnimationScale <= 0 {
		invalidVisual("Combo", "Metrics.AnimationScale", "must be greater than 0")
	}
	if metrics.AnimationScale > 1 {
		invalidVisual("Combo", "Metrics.AnimationScale", "must be at most 1")
	}

	resolve := func(part style.Part, state style.State) style.ResolvedStyle {
		return style.CascadePart(state, part, visual.Defaults, visual.Override)
	}
	requireText := func(field string, value style.ResolvedStyle, requireSize bool) {
		if value.Text == nil {
			invalidVisual("Combo", field, "must be provided")
		}
		if !hasColor(value.Text.Color) {
			invalidVisual("Combo", field, "must resolve to a solid color")
		}

		if requireSize && (value.Text.FontSize == nil || *value.Text.FontSize <= 0) {
			invalidVisual("Combo", field+".Text.FontSize", "must be provided and greater than 0 when size is required")
		}
	}

	trigger := resolve(ComboPartTrigger, style.State{})
	if trigger.Box == nil {
		invalidVisual("Combo", "Trigger.Box", "must be provided")
	}
	if trigger.Box.Height == nil && trigger.Box.MinHeight == nil {
		invalidVisual("Combo", "Trigger.Box.Height", "height or minimum height must be provided")
	}

	requireText("Trigger", trigger, true)
	requireText("Placeholder", resolve(ComboPartPlaceholder, style.State{}), false)
	selection := resolve(ComboPartSelection, style.State{})
	if selection.Paint == nil {
		invalidVisual("Combo", "Selection.Paint", "must be provided")
	}
	if !hasPaint(selection.Paint.Background) {
		invalidVisual("Combo", "Selection.Paint.Background", "must provide a solid color or nonempty gradient")
	}

	indicator := resolve(ComboPartIndicator, style.State{})
	if indicator.Box == nil {
		invalidVisual("Combo", "Indicator.Box", "must be provided")
	}
	if indicator.Box.Width == nil {
		invalidVisual("Combo", "Indicator.Box.Width", "must be provided")
	}
	if *indicator.Box.Width <= 0 {
		invalidVisual("Combo", "Indicator.Box.Width", "must be greater than 0")
	}

	requireText("Indicator", indicator, false)
	icon := resolve(ComboPartIcon, style.State{})
	if icon.Box == nil {
		invalidVisual("Combo", "Icon.Box", "must be provided")
	}
	if icon.Box.Width == nil {
		invalidVisual("Combo", "Icon.Box.Width", "must be provided")
	}
	if icon.Box.Height == nil {
		invalidVisual("Combo", "Icon.Box.Height", "must be provided")
	}
	if *icon.Box.Width <= 0 {
		invalidVisual("Combo", "Icon.Box.Width", "must be greater than 0")
	}
	if *icon.Box.Height <= 0 {
		invalidVisual("Combo", "Icon.Box.Height", "must be greater than 0")
	}

	requireText("Icon", icon, false)
	panel := resolve(ComboPartPanel, style.State{})
	if panel.Box == nil {
		invalidVisual("Combo", "Panel.Box", "must be provided")
	}
	if panel.Box.Padding == nil {
		invalidVisual("Combo", "Panel.Box.Padding", "must be provided")
	}
	if panel.Paint == nil {
		invalidVisual("Combo", "Panel.Paint", "must be provided")
	}
	if !hasPaint(panel.Paint.Background) {
		invalidVisual("Combo", "Panel.Paint.Background", "must provide a solid color or nonempty gradient")
	}
	if panel.Paint.Radius == nil {
		invalidVisual("Combo", "Panel.Paint.Radius", "must be provided")
	}

	requireText("Panel", panel, false)

	states := []style.State{
		{},
		{Hovered: true},
		{Pressed: true},
		{Focused: true, FocusVisible: true},
		{Disabled: true},
		{Selected: true},
		{Open: true},
	}
	for _, state := range states {
		requireText("Placeholder", resolve(ComboPartPlaceholder, state), false)
		requireText("Description", resolve(ComboPartDescription, state), true)
		requireText("Item", resolve(ComboPartItem, state), true)
	}
	item := resolve(ComboPartItem, style.State{})
	if item.Box == nil {
		invalidVisual("Combo", "Item.Box", "must be provided")
	}
	if item.Box.MinHeight == nil {
		invalidVisual("Combo", "Item.Box.MinHeight", "must be provided")
	}
	if *item.Box.MinHeight <= 0 {
		invalidVisual("Combo", "Item.Box.MinHeight", "must be greater than 0")
	}
}
