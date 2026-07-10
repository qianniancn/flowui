package frame

import "gioui.org/layout"

// Widget is a FlowUI node that lays itself out with the current frame context.
type Widget interface {
	Layout(ctx *Context, gtx layout.Context) layout.Dimensions
}
