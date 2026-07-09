package layout

import (
	"image"

	giolayout "gioui.org/layout"
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

func (a Align) Direction() giolayout.Direction {
	switch a {
	case AlignTopStart:
		return giolayout.NW
	case AlignTop:
		return giolayout.N
	case AlignTopEnd:
		return giolayout.NE
	case AlignStart:
		return giolayout.W
	case AlignCenter:
		return giolayout.Center
	case AlignEnd:
		return giolayout.E
	case AlignBottomStart:
		return giolayout.SW
	case AlignBottom:
		return giolayout.S
	case AlignBottomEnd:
		return giolayout.SE
	default:
		return giolayout.NW
	}
}

func (a Align) Position(widget, bounds image.Point) image.Point {
	return a.Direction().Position(widget, bounds)
}
