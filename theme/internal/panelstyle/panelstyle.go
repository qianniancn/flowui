// Package panelstyle builds the shared chrome for popup panels.
//
// Select, ComboBox, and DatePicker render different triggers and content, but
// their popup surfaces share the same chrome: an overlay fill, a hairline
// border, the menu shadow profile, and clipped rounded corners. Keeping that
// recipe here prevents the components from drifting apart when one panel
// changes.
package panelstyle

import (
	"gioui.org/unit"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

// Popup returns the panel declaration for an option list.
func Popup(active *theme.Theme, radius, padding, borderWidth unit.Dp) style.Style {
	sem := active.Semantic()
	return style.Style{}.
		Background(style.Solid(sem.Overlay)).
		TextColor(style.Solid(sem.OverlayForeground)).
		BorderColor(style.Solid(sem.Border)).
		BorderWidth(borderWidth).
		Radius(radius).
		Padding(padding).
		Shadow(style.ShadowMenu).
		Overflow(style.OverflowHidden)
}
