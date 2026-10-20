package theme

import (
	"image"
	"time"

	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/style"
)

// SelectRequest describes the semantic field treatment needed by a select
// trigger. The provider owns the trigger's parts and panel declarations.
type SelectRequest struct {
	Variant  InputVariant
	Override *SelectOverride
}

// SelectOverride carries caller customization across the neutral provider
// boundary. The provider owns the returned instance layer.
type SelectOverride struct {
	Declaration style.Style
}

// SelectOverrideStyle extracts the provider-owned instance layer from a
// request.
func SelectOverrideStyle(request SelectRequest) style.Style {
	if request.Override == nil {
		return style.Style{}
	}
	return request.Override.Declaration
}

// SelectVisual contains the provider-owned declaration layers consumed by
// select behavior. The request variant is resolved before this bundle is
// returned; only the default and instance layers cross the behavior boundary.
type SelectVisual struct {
	Defaults       style.Style
	Override       style.Style
	Metrics        SelectMetrics
	Motion         MotionTheme
	Indicator      SelectIndicatorRenderer
	Text           SelectTextRenderer
	IndicatorWidth SelectPartWidthResolver
	IconSize       SelectPartSizeResolver
}

// SelectMetrics contains the geometry and motion timing used by select
// behavior. The widget consumes these values without reading the theme
// component token table directly.
type SelectMetrics struct {
	ContentGap        unit.Dp
	PanelGap          unit.Dp
	PanelMaxHeight    unit.Dp
	AnimationScale    float32
	AnimationDistance unit.Dp
	OpenDuration      time.Duration
	CloseDuration     time.Duration
	IndicatorDuration time.Duration
}

// Select part names are stable semantic customization points shared by the
// behavior widget and visual providers. Overlay is the panel surface; its
// lifecycle remains owned by the widget.
const (
	SelectPartTrigger     style.Part = style.PartContent
	SelectPartPlaceholder style.Part = style.PartPlaceholder
	SelectPartSelection   style.Part = style.PartSelection
	SelectPartIndicator   style.Part = style.PartIndicator
	SelectPartIcon        style.Part = style.PartIcon
	SelectPartPanel       style.Part = style.PartPanel
	SelectPartOverlay     style.Part = SelectPartPanel
	SelectPartDescription style.Part = style.PartDescription
)

// SelectTextRequest contains the semantic text and the already-resolved part
// declaration. The provider owns how typography and color are interpreted.
type SelectTextRequest struct {
	Text  string
	Style style.ResolvedStyle
}

// SelectTextRenderer is the theme-owned text primitive used by the trigger
// value and supporting message. Widget code supplies content and constraints
// but does not translate a declaration into a Gio label.
type SelectTextRenderer func(layout.Context, SelectTextRequest) layout.Dimensions

// SelectIndicatorRenderer is the theme-owned default disclosure mark. The
// widget owns when the indicator is shown and the transition progress; the
// provider owns its geometry, color, and drawing strategy.
type SelectIndicatorRenderer func(layout.Context, image.Point, float32, style.ResolvedStyle) layout.Dimensions

// SelectPartWidthResolver and SelectPartSizeResolver keep provider-owned part
// geometry out of the behavior widget while preserving instance overrides.
type SelectPartWidthResolver func(style.ResolvedStyle) unit.Dp

type SelectPartSizeResolver func(layout.Context, style.ResolvedStyle) image.Point

// SelectProvider supplies select trigger and panel visuals.
type SelectProvider interface {
	Select(active *Theme, request SelectRequest) SelectVisual
}

// ValidateSelectVisual checks the complete declaration and renderer contract
// before a widget stores a provider result in its cache. The check lives in
// theme so behavior packages never need to understand concrete color or brush
// representations.
func ValidateSelectVisual(visual SelectVisual) {
	if visual.Indicator == nil {
		invalidVisual("Select", "Indicator", "must be provided")
	}
	if visual.Text == nil {
		invalidVisual("Select", "Text", "must be provided")
	}
	if visual.IndicatorWidth == nil {
		invalidVisual("Select", "IndicatorWidth", "must be provided")
	}
	if visual.IconSize == nil {
		invalidVisual("Select", "IconSize", "must be provided")
	}

	metrics := visual.Metrics
	if metrics.ContentGap < 0 {
		invalidVisual("Select", "Metrics.ContentGap", "must be at least 0")
	}
	if metrics.PanelGap < 0 {
		invalidVisual("Select", "Metrics.PanelGap", "must be at least 0")
	}
	if metrics.PanelMaxHeight <= 0 {
		invalidVisual("Select", "Metrics.PanelMaxHeight", "must be greater than 0")
	}
	if metrics.AnimationScale <= 0 {
		invalidVisual("Select", "Metrics.AnimationScale", "must be greater than 0")
	}
	if metrics.AnimationScale > 1 {
		invalidVisual("Select", "Metrics.AnimationScale", "must be at most 1")
	}
	if metrics.AnimationDistance < 0 {
		invalidVisual("Select", "Metrics.AnimationDistance", "must be at least 0")
	}
	if metrics.OpenDuration <= 0 {
		invalidVisual("Select", "Metrics.OpenDuration", "must be greater than 0")
	}
	if metrics.CloseDuration <= 0 {
		invalidVisual("Select", "Metrics.CloseDuration", "must be greater than 0")
	}
	if metrics.IndicatorDuration <= 0 {
		invalidVisual("Select", "Metrics.IndicatorDuration", "must be greater than 0")
	}

	resolve := func(part style.Part, state style.State) style.ResolvedStyle {
		return style.CascadePart(state, part, visual.Defaults, visual.Override)
	}
	requireText := func(field string, value style.ResolvedStyle, requireSize bool) {
		if value.Text == nil {
			invalidVisual("Select", field, "must be provided")
		}
		if !hasColor(value.Text.Color) {
			invalidVisual("Select", field, "must resolve to a solid color")
		}

		if requireSize && (value.Text.FontSize == nil || *value.Text.FontSize <= 0) {
			invalidVisual("Select", field+".Text.FontSize", "must be provided and greater than 0 when size is required")
		}
	}

	trigger := resolve(SelectPartTrigger, style.State{})
	if trigger.Box == nil {
		invalidVisual("Select", "Trigger.Box", "must be provided")
	}
	if trigger.Box.Height == nil && trigger.Box.MinHeight == nil {
		invalidVisual("Select", "Trigger.Box.Height", "height or minimum height must be provided")
	}

	requireText("Trigger", trigger, true)
	requireText("Placeholder", resolve(SelectPartPlaceholder, style.State{}), false)
	selection := resolve(SelectPartSelection, style.State{})
	if selection.Paint == nil {
		invalidVisual("Select", "Selection.Paint", "must be provided")
	}
	if !hasPaint(selection.Paint.Background) {
		invalidVisual("Select", "Selection.Paint.Background", "must provide a solid color or nonempty gradient")
	}

	indicator := resolve(SelectPartIndicator, style.State{})
	if indicator.Box == nil {
		invalidVisual("Select", "Indicator.Box", "must be provided")
	}
	if indicator.Box.Width == nil {
		invalidVisual("Select", "Indicator.Box.Width", "must be provided")
	}
	if *indicator.Box.Width <= 0 {
		invalidVisual("Select", "Indicator.Box.Width", "must be greater than 0")
	}

	requireText("Indicator", indicator, false)
	icon := resolve(SelectPartIcon, style.State{})
	if icon.Box == nil {
		invalidVisual("Select", "Icon.Box", "must be provided")
	}
	if icon.Box.Width == nil {
		invalidVisual("Select", "Icon.Box.Width", "must be provided")
	}
	if icon.Box.Height == nil {
		invalidVisual("Select", "Icon.Box.Height", "must be provided")
	}
	if *icon.Box.Width <= 0 {
		invalidVisual("Select", "Icon.Box.Width", "must be greater than 0")
	}
	if *icon.Box.Height <= 0 {
		invalidVisual("Select", "Icon.Box.Height", "must be greater than 0")
	}

	requireText("Icon", icon, false)
	panel := resolve(SelectPartPanel, style.State{})
	if panel.Box == nil {
		invalidVisual("Select", "Panel.Box", "must be provided")
	}
	if panel.Box.Padding == nil {
		invalidVisual("Select", "Panel.Box.Padding", "must be provided")
	}
	if panel.Paint == nil {
		invalidVisual("Select", "Panel.Paint", "must be provided")
	}
	if !hasPaint(panel.Paint.Background) {
		invalidVisual("Select", "Panel.Paint.Background", "must provide a solid color or nonempty gradient")
	}
	if panel.Paint.Radius == nil {
		invalidVisual("Select", "Panel.Paint.Radius", "must be provided")
	}

	requireText("Panel", panel, false)
	for _, state := range []style.State{
		{},
		{Disabled: true},
		{Invalid: true},
		{Disabled: true, Invalid: true},
	} {
		requireText("Description", resolve(SelectPartDescription, state), true)
	}
}
