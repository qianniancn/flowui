package theme

import (
	"image/color"

	"gioui.org/unit"
)

// IconRequest describes the inherited foreground supplied by icon behavior.
// The provider owns the default icon size and resolves the effective color.
type IconRequest struct {
	Foreground color.NRGBA
}

// IconVisual is the complete provider-owned visual input consumed by icon
// behavior. Explicit widget properties are resolved by the widget afterward.
type IconVisual struct {
	Size  unit.Dp
	Color color.NRGBA
}

// IconProvider supplies default icon geometry and paint without taking
// ownership of IconVG decoding or rendering.
type IconProvider interface {
	Icon(active *Theme, request IconRequest) IconVisual
}

// ValidateIconVisual checks the provider-owned default geometry and paint
// before icon behavior stores a result in its cache.
func ValidateIconVisual(visual IconVisual) {
	if !validPositiveDp(visual.Size) {
		invalidVisual("Icon", "Size", "must be finite and greater than 0")
	}
	if visual.Color.A == 0 {
		invalidVisual("Icon", "Color.A", "must be nonzero")
	}
}
