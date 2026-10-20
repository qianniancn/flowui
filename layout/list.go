package layout

import (
	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
)

type ListWidget struct {
	key           string
	count         int
	item          func(int) Widget
	gap           unit.Dp
	align         layout.Alignment
	disabled      bool
	stickToEnd    bool
	scrollAnyAxis bool
}

// List creates a virtualized list. Its identity is derived from the enclosing
// scope; call Key when the same position may hold different content.
func List(count int, item func(int) Widget) ListWidget {
	return ListWidget{
		key:   core.AutoKey(),
		count: count,
		item:  item,
	}
}

// Key sets an explicit identity for the list state.
func (l ListWidget) Key(key string) ListWidget {
	l.key = key
	return l
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

func (l ListWidget) Layout(ctx *Context, gtx layout.Context) layout.Dimensions {
	key, state := ctx.ClaimListState(l.key)
	visual := visualOutsetStateFor(ctx, core.KindList, key)
	bar := derivedScrollbarState(ctx, key)
	state.Axis = layout.Vertical
	state.Gap = max(gtx.Dp(l.gap), 0)
	state.Alignment = l.align
	state.ScrollToEnd = l.stickToEnd
	state.ScrollAnyAxis = l.scrollAnyAxis
	if l.disabled {
		gtx = gtx.Disabled()
	}
	return layoutScrollbarList(ctx, gtx, "", state, bar, visual, l.count, l.disabled, false, style.Style{}, func(gtx layout.Context, index int) layout.Dimensions {
		item := l.item(index)
		prepareFieldAssociations(ctx, item)
		return item.Layout(ctx, gtx)
	})
}
