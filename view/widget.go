package view

import "gioui.org/layout"

type Widget[C any] interface {
	Layout(ctx *C, gtx layout.Context) layout.Dimensions
}
