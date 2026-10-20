// Package iconstyle contains the theme-owned visual recipe for icons.
// Icon behavior remains in widget/icon; this package only resolves the
// provider-owned default size and inherited foreground.
package iconstyle

import "github.com/qianniancn/flowui/theme"

// BuildIcon builds the provider-owned icon visual.
func BuildIcon(active *theme.Theme, request theme.IconRequest) theme.IconVisual {
	if active == nil {
		panic(theme.ErrMissingVisualProvider)
	}
	return theme.IconVisual{
		Size:  active.Components.Icon.Size,
		Color: request.Foreground,
	}
}
