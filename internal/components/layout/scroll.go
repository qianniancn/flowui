package layoutui

import (
	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/frame"
	stateutil "github.com/qianniancn/flowui/internal/state"
	flowstyle "github.com/qianniancn/flowui/internal/style"
)

type ScrollWidget struct {
	key           string
	child         frame.Widget
	axis          layout.Axis
	align         layout.Alignment
	disabled      bool
	stickToEnd    bool
	scrollAnyAxis bool
}

func Scroll(key string, child frame.Widget) ScrollWidget {
	return ScrollWidget{
		key:   key,
		child: child,
		axis:  layout.Vertical,
	}
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

func (s ScrollWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	prepareFieldAssociations(ctx, s.child)
	state := ctx.ScrollState(s.key)
	visual := visualOutsetStateFor(ctx, stateutil.KindScroll, s.key)
	bar := derivedScrollbarState(ctx, s.key)
	state.Axis = s.axis
	state.Gap = 0
	state.Alignment = s.align
	state.ScrollToEnd = s.stickToEnd
	state.ScrollAnyAxis = s.scrollAnyAxis
	if s.disabled {
		gtx = gtx.Disabled()
	}
	return layoutScrollbarList(ctx, gtx, "", state, bar, visual, 1, s.disabled, false, flowstyle.Style{}, func(gtx layout.Context, _ int) layout.Dimensions {
		return s.child.Layout(ctx, gtx)
	})
}
