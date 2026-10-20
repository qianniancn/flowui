package layout

import (
	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/style"
)

const (
	pageMaxWidth = unit.Dp(760)
	pagePadding  = unit.Dp(24)
)

// PageWidget is the standard page skeleton: its children stack in a column,
// centered horizontally, with the page scrolling vertically.
type PageWidget struct {
	children []Widget
	gap      unit.Dp
	style    style.Style
}

// Page wraps page content in the standard skeleton. The default max width and
// padding keep long forms and galleries readable; Style overrides them.
func Page(children ...Widget) PageWidget {
	return PageWidget{children: children, gap: 16}
}

// Gap sets the spacing between children, in dp.
func (p PageWidget) Gap(dp int) PageWidget {
	p.gap = unit.Dp(dp)
	return p
}

// Style merges instance declarations into the page box.
func (p PageWidget) Style(values ...style.Style) PageWidget {
	p.style = style.Join(append([]style.Style{p.style}, values...)...)
	return p
}

func (p PageWidget) Layout(ctx *Context, gtx layout.Context) layout.Dimensions {
	column := Column(p.children...).Gap(int(p.gap))
	page := Box(Scroll(column).Vertical()).Style(
		style.New().FillWidth().MaxWidth(pageMaxWidth).Padding(pagePadding),
		p.style,
	)
	return Center(page).Layout(ctx, gtx)
}
