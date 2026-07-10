package layoutui

import (
	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/FlowUI/internal/frame"
)

type ListWidget struct {
	key           string
	count         int
	item          func(int) frame.Widget
	gap           unit.Dp
	align         layout.Alignment
	disabled      bool
	stickToEnd    bool
	scrollAnyAxis bool
}

func List(key string, count int, item func(int) frame.Widget) ListWidget {
	return ListWidget{
		key:   key,
		count: count,
		item:  item,
	}
}

func (l ListWidget) Gap(dp int) ListWidget {
	l.gap = unit.Dp(dp)
	return l
}

func (l ListWidget) Disabled(disabled bool) ListWidget {
	l.disabled = disabled
	return l
}

func (l ListWidget) AlignStart() ListWidget {
	l.align = layout.Start
	return l
}

func (l ListWidget) AlignMiddle() ListWidget {
	l.align = layout.Middle
	return l
}

func (l ListWidget) AlignEnd() ListWidget {
	l.align = layout.End
	return l
}

func (l ListWidget) StickToEnd() ListWidget {
	l.stickToEnd = true
	return l
}

func (l ListWidget) ScrollAnyAxis() ListWidget {
	l.scrollAnyAxis = true
	return l
}

func (l ListWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	state := ctx.ListState(l.key)
	state.Axis = layout.Vertical
	state.Gap = gtx.Dp(l.gap)
	state.Alignment = l.align
	state.ScrollToEnd = l.stickToEnd
	state.ScrollAnyAxis = l.scrollAnyAxis
	if l.disabled {
		gtx = gtx.Disabled()
	}
	return state.Layout(gtx, l.count, func(gtx layout.Context, index int) layout.Dimensions {
		item := l.item(index)
		prepareFieldAssociations(ctx, item)
		return item.Layout(ctx, gtx)
	})
}
