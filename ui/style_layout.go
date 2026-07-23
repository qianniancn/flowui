package ui

import (
	"gioui.org/layout"
	layoutui "github.com/qianniancn/FlowUI/internal/components/layout"
)

// LayoutResolvedStyle renders a Style returned by ResolveStyle or
// ResolveStylePart. It lets custom components reuse the common box renderer.
func LayoutResolvedStyle(ctx *Context, gtx layout.Context, resolved ResolvedStyle, child Widget) layout.Dimensions {
	return layoutui.LayoutResolved(ctx, gtx, resolved, child)
}

// LayoutInteractiveResolvedStyle keeps margin outside an interaction host and
// applies the remaining resolved Style to its visual and hit area.
func LayoutInteractiveResolvedStyle(
	ctx *Context,
	gtx layout.Context,
	resolved ResolvedStyle,
	child Widget,
	host func(layout.Context, layout.Widget) layout.Dimensions,
) layout.Dimensions {
	return layoutui.LayoutInteractiveResolved(ctx, gtx, resolved, child, host)
}
