// Package host implements the single-box layout host: ResolvedStyle + optional
// interaction + one child. Public surface: ui.Box, ui.LayoutBox, ui.LayoutInteractiveBox.
package host

import (
	"gioui.org/layout"
	"github.com/qianniancn/FlowUI/internal/frame"
	flowstyle "github.com/qianniancn/FlowUI/internal/style"
)

// LayoutBox lays out a pre-resolved non-interactive box.
func LayoutBox(ctx *frame.Context, gtx layout.Context, resolved flowstyle.ResolvedStyle, child frame.Widget) layout.Dimensions {
	return LayoutResolved(ctx, gtx, resolved, child)
}

// LayoutInteractiveBox keeps margin outside the interaction host.
func LayoutInteractiveBox(
	ctx *frame.Context,
	gtx layout.Context,
	resolved flowstyle.ResolvedStyle,
	child frame.Widget,
	hostFn func(layout.Context, layout.Widget) layout.Dimensions,
) layout.Dimensions {
	return LayoutInteractiveResolved(ctx, gtx, resolved, child, hostFn)
}
