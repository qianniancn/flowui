package frame

import (
	"gioui.org/layout"
	"gioui.org/op"
)

func DeferOverlay(ctx *Context, gtx layout.Context, overlay layout.Widget) {
	macro := op.Record(gtx.Ops)
	overlay(gtx)
	op.Defer(gtx.Ops, macro.Stop())
}
