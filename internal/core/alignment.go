package core

import (
	"image"

	"gioui.org/layout"
)

// LayoutAlign positions a child within its layout bounds.
//
// It is kept in the core package so internal layout hosts share one type
// without depending back on the public layout package.
type LayoutAlign int

const (
	LayoutAlignTopStart LayoutAlign = iota
	LayoutAlignTop
	LayoutAlignTopEnd
	LayoutAlignStart
	LayoutAlignCenter
	LayoutAlignEnd
	LayoutAlignBottomStart
	LayoutAlignBottom
	LayoutAlignBottomEnd
)

func (a LayoutAlign) Direction() layout.Direction {
	switch a {
	case LayoutAlignTopStart:
		return layout.NW
	case LayoutAlignTop:
		return layout.N
	case LayoutAlignTopEnd:
		return layout.NE
	case LayoutAlignStart:
		return layout.W
	case LayoutAlignCenter:
		return layout.Center
	case LayoutAlignEnd:
		return layout.E
	case LayoutAlignBottomStart:
		return layout.SW
	case LayoutAlignBottom:
		return layout.S
	case LayoutAlignBottomEnd:
		return layout.SE
	default:
		return layout.NW
	}
}

func (a LayoutAlign) Position(widget, bounds image.Point) image.Point {
	return a.Direction().Position(widget, bounds)
}
