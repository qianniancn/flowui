package layout

import (
	"image"

	"gioui.org/layout"
)

type Overflow int

const (
	OverflowVisible Overflow = iota
	OverflowHidden
)

type Align int

const (
	AlignTopStart Align = iota
	AlignTop
	AlignTopEnd
	AlignStart
	AlignCenter
	AlignEnd
	AlignBottomStart
	AlignBottom
	AlignBottomEnd
)

func (a Align) Direction() layout.Direction {
	switch a {
	case AlignTopStart:
		return layout.NW
	case AlignTop:
		return layout.N
	case AlignTopEnd:
		return layout.NE
	case AlignStart:
		return layout.W
	case AlignCenter:
		return layout.Center
	case AlignEnd:
		return layout.E
	case AlignBottomStart:
		return layout.SW
	case AlignBottom:
		return layout.S
	case AlignBottomEnd:
		return layout.SE
	default:
		return layout.NW
	}
}

func (a Align) Position(widget, bounds image.Point) image.Point {
	return a.Direction().Position(widget, bounds)
}
