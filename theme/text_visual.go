package theme

import (
	"image"
	"image/color"

	"gioui.org/layout"
	"github.com/qianniancn/flowui/style"
)

// TextRequest describes the inherited foreground supplied by a text behavior.
// The provider owns the remaining default typography and selection treatment.
type TextRequest struct {
	Foreground color.NRGBA
}

// TextUnderlineRequest contains the resolved text bounds and foreground
// needed by a provider-owned underline renderer.
type TextUnderlineRequest struct {
	Size     image.Point
	Baseline int
	Color    color.NRGBA
}

// TextUnderlineRenderer paints the underline requested by a resolved text
// declaration. The provider owns thickness and baseline treatment.
type TextUnderlineRenderer func(layout.Context, TextUnderlineRequest)

// TextVisual is the complete default visual plan consumed by text behavior.
// Explicit text instance properties are resolved by the widget after Defaults.
type TextVisual struct {
	Defaults  style.Style
	Selection color.NRGBA
	Underline TextUnderlineRenderer
}

// TextProvider supplies default typography and selection treatment without
// taking ownership of text content, semantics, or layout behavior.
type TextProvider interface {
	Text(active *Theme, request TextRequest) TextVisual
}

// ValidateTextVisual checks the default text declaration before a widget
// stores a provider result in its cache. Text behavior cannot invent a
// fallback typography without violating the provider boundary.
func ValidateTextVisual(visual TextVisual) {
	if visual.Underline == nil {
		invalidVisual("Text", "Underline", "must be provided")
	}

	resolved := style.Cascade(style.State{}, visual.Defaults)
	if resolved.Text == nil {
		invalidVisual("Text", "Resolved.Text", "must be provided")
	}
	if resolved.Text.FontSize == nil {
		invalidVisual("Text", "Resolved.Text.FontSize", "must be provided")
	}
	if *resolved.Text.FontSize <= 0 {
		invalidVisual("Text", "Resolved.Text.FontSize", "must be greater than 0")
	}
	if !hasColor(resolved.Text.Color) {
		invalidVisual("Text", "Resolved.Text.Color", "must resolve to a solid color")
	}
	if resolved.Text.Typeface == nil {
		invalidVisual("Text", "Resolved.Text.Typeface", "must be provided")
	}
}
