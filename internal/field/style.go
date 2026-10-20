package field

import (
	"image/color"

	"github.com/qianniancn/flowui/style"
)

type Variant uint8

const (
	Primary Variant = iota
	Secondary
	// Filled is a resting gray surface with no border.
	Filled
	// Borderless is a transparent surface.
	Borderless
	// Underlined draws only a bottom border.
	Underlined
)

type Colors struct {
	Foreground  color.NRGBA
	Placeholder color.NRGBA
	Selection   color.NRGBA
}

type Resolved struct {
	Content style.ResolvedStyle
	Colors  Colors
}
