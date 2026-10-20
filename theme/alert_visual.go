package theme

import (
	"gioui.org/unit"
	"github.com/qianniancn/flowui/style"
)

// AlertRequest contains the semantic status and optional instance declaration
// needed by an alert visual provider. Title, description, slots, and actions
// remain behavior-owned by the widget.
type AlertRequest struct {
	Status   AlertStatus
	Override *AlertOverride
}

// AlertOverride carries caller customization across the provider boundary.
type AlertOverride struct {
	Declaration style.Style
}

// AlertOverrideStyle extracts the provider-owned instance layer.
func AlertOverrideStyle(request AlertRequest) style.Style {
	if request.Override == nil {
		return style.Style{}
	}
	return request.Override.Declaration
}

// AlertMetrics contains the neutral geometry consumed by alert behavior.
type AlertMetrics struct {
	Gap              unit.Dp
	IndicatorPadding unit.Dp
	IconSize         unit.Dp
}

// AlertVisual is the complete visual plan consumed by alert behavior.
// Icon is a theme-owned glyph selected for the semantic status.
type AlertVisual struct {
	Defaults    style.Style
	Title       style.Style
	Description style.Style
	Content     style.Style
	Indicator   style.Style
	Override    style.Style
	Metrics     AlertMetrics
	Icon        []byte
}

// AlertProvider supplies alert declarations, metrics, and indicator glyphs
// without taking ownership of content, action, or semantic behavior.
type AlertProvider interface {
	Alert(active *Theme, request AlertRequest) AlertVisual
}

// ValidateAlertVisual checks the provider-owned declarations and neutral
// geometry before a widget stores an alert visual in its cache. Alert
// behavior supplies content and slots; it must not repair missing root or
// text declarations locally.
func ValidateAlertVisual(visual AlertVisual) {
	metrics := visual.Metrics
	if !validNonNegativeDp(metrics.Gap) {
		invalidVisual("Alert", "Metrics.Gap", "must be finite and non-negative")
	}
	if !validNonNegativeDp(metrics.IndicatorPadding) {
		invalidVisual("Alert", "Metrics.IndicatorPadding", "must be finite and non-negative")
	}
	if !validPositiveDp(metrics.IconSize) {
		invalidVisual("Alert", "Metrics.IconSize", "must be finite and greater than 0")
	}

	validateAlertRoot("Root", style.Cascade(style.State{}, visual.Defaults))
	validateAlertRoot("Root", style.Cascade(style.State{}, visual.Defaults, visual.Override))
	validateAlertTextPart("Title", style.CascadePart(style.State{}, style.PartLabel, visual.Title))
	validateAlertTextPart("Title", style.CascadePart(style.State{}, style.PartLabel, visual.Title, visual.Override))
	validateAlertTextPart("Description", style.CascadePart(style.State{}, style.PartDescription, visual.Description))
	validateAlertTextPart("Description", style.CascadePart(style.State{}, style.PartDescription, visual.Description, visual.Override))
	validateAlertTextPart("Content", style.CascadePart(style.State{}, style.PartContent, visual.Content))
	validateAlertTextPart("Content", style.CascadePart(style.State{}, style.PartContent, visual.Content, visual.Override))
	validateAlertIndicator("Indicator", style.CascadePart(style.State{}, style.PartIndicator, visual.Indicator))
	validateAlertIndicator("Indicator", style.CascadePart(style.State{}, style.PartIndicator, visual.Indicator, visual.Override))
}

func validateAlertRoot(field string, resolved style.ResolvedStyle) {
	if resolved.Box == nil {
		invalidVisual("Alert", field+".Box", "must be provided")
	}
	if resolved.Box.Padding == nil {
		invalidVisual("Alert", field+".Box.Padding", "must be provided")
	}
	if !validOptionalPadding(resolved.Box.Padding) {
		invalidVisual("Alert", field+".Box.Padding", "padding sides must be finite and non-negative")
	}
	if resolved.Paint == nil {
		invalidVisual("Alert", field+".Paint", "must be provided")
	}
	if !hasPaint(resolved.Paint.Background) {
		invalidVisual("Alert", field+".Paint.Background", "must provide a solid color or nonempty gradient")
	}
	if resolved.Paint.Radius == nil {
		invalidVisual("Alert", field+".Paint.Radius", "must be provided")
	}
	if !validNonNegativeDp(*resolved.Paint.Radius) {
		invalidVisual("Alert", field+".Paint.Radius", "must be finite and non-negative")
	}
	if resolved.Paint.Opacity != nil && !hasValidOpacity(resolved.Paint.Opacity) {
		invalidVisual("Alert", field+".Paint.Opacity", "must be finite and between 0 and 1 when provided")
	}
	if resolved.Text == nil {
		invalidVisual("Alert", field+".Text", "must be provided")
	}
	if !hasColor(resolved.Text.Color) {
		invalidVisual("Alert", field+".Text.Color", "must resolve to a solid color")
	}
}

func validateAlertTextPart(field string, resolved style.ResolvedStyle) {
	if resolved.Text == nil {
		invalidVisual("Alert", field+".Text", "must be provided")
	}
	if resolved.Text.FontSize == nil {
		invalidVisual("Alert", field+".Text.FontSize", "must be provided")
	}
	if !validPositiveSp(*resolved.Text.FontSize) {
		invalidVisual("Alert", field+".Text.FontSize", "must be finite and greater than 0")
	}
	if !hasColor(resolved.Text.Color) {
		invalidVisual("Alert", field+".Text.Color", "must resolve to a solid color")
	}

	if resolved.Text.LineHeight == nil {
		invalidVisual("Alert", field+".Text.LineHeight", "must be provided")
	}
	if !validPositiveSp(*resolved.Text.LineHeight) {
		invalidVisual("Alert", field+".Text.LineHeight", "must be finite and greater than 0")
	}
}

func validateAlertIndicator(field string, resolved style.ResolvedStyle) {
	if resolved.Text == nil {
		invalidVisual("Alert", field+".Text", "must be provided")
	}
	if !hasColor(resolved.Text.Color) {
		invalidVisual("Alert", field+".Text.Color", "must resolve to a solid color")
	}
}
