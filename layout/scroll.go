package layout

import (
	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
)

type ScrollWidget struct {
	key           string
	child         Widget
	axis          layout.Axis
	align         layout.Alignment
	disabled      bool
	stickToEnd    bool
	scrollAnyAxis bool
	overlay       bool
}

// Scroll creates a scrollable viewport. Its identity is derived from the
// enclosing scope; call Key when the same position may hold different content.
func Scroll(child Widget) ScrollWidget {
	return ScrollWidget{
		key:   core.AutoKey(),
		child: child,
		axis:  layout.Vertical,
	}
}

// Key sets an explicit identity for the scroll state.
func (s ScrollWidget) Key(key string) ScrollWidget {
	s.key = key
	return s
}

func (s ScrollWidget) Vertical() ScrollWidget {
	s.axis = layout.Vertical
	return s
}

func (s ScrollWidget) Horizontal() ScrollWidget {
	s.axis = layout.Horizontal
	return s
}

func (s ScrollWidget) Disabled(disabled bool) ScrollWidget {
	s.disabled = disabled
	return s
}

func (s ScrollWidget) AlignStart() ScrollWidget {
	s.align = layout.Start
	return s
}

func (s ScrollWidget) AlignMiddle() ScrollWidget {
	s.align = layout.Middle
	return s
}

func (s ScrollWidget) AlignEnd() ScrollWidget {
	s.align = layout.End
	return s
}

func (s ScrollWidget) StickToEnd() ScrollWidget {
	s.stickToEnd = true
	return s
}

func (s ScrollWidget) ScrollAnyAxis() ScrollWidget {
	s.scrollAnyAxis = true
	return s
}

// Overlay draws the scrollbar over the content instead of reserving a gutter.
// Use it for pages whose height changes, such as tabbed settings pages, so the
// content width does not jump when a short page follows a long one.
func (s ScrollWidget) Overlay() ScrollWidget {
	s.overlay = true
	return s
}

func (s ScrollWidget) Layout(ctx *Context, gtx layout.Context) layout.Dimensions {
	prepareFieldAssociations(ctx, s.child)
	key, state := ctx.ClaimScrollState(s.key)
	visual := visualOutsetStateFor(ctx, core.KindScroll, key)
	bar := derivedScrollbarState(ctx, key)
	state.Axis = s.axis
	state.Gap = 0
	state.Alignment = s.align
	state.ScrollToEnd = s.stickToEnd
	state.ScrollAnyAxis = s.scrollAnyAxis
	if s.disabled {
		gtx = gtx.Disabled()
	}
	return layoutScrollbarList(
		ctx,
		gtx,
		"",
		state,
		bar,
		visual,
		1,
		s.disabled,
		s.overlay,
		style.Style{},
		func(gtx layout.Context, _ int) layout.Dimensions {
			return s.child.Layout(ctx, gtx)
		},
	)
}
