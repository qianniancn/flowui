package pagination

import (
	"image"
	"strconv"

	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/animation"
	"github.com/qianniancn/flowui/i18n"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

const stateSlotPagination = "pagination"

type paginationState struct {
	items      map[string]*paginationItemState
	frameItems map[string]struct{}
}

type paginationItemState struct {
	clickable widget.Clickable
	focus     animation.FocusAnimation
}

func (p Widget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	return p.layoutWithOverride(ctx, gtx, p.override)
}

func (p Widget) layoutWithOverride(ctx *core.Context, gtx layout.Context, override *theme.PaginationOverride) layout.Dimensions {
	key := core.ClaimKey(ctx, core.KindPagination, p.key)
	value := core.UseState[paginationState](ctx, key, stateSlotPagination)
	value.beginFrame()
	defer value.endFrame()
	visual := p.visualWithCache(ctx, override)

	semantic.LabelOp(ctx.Text(i18n.PaginationLabel)).Add(gtx.Ops)
	content := func(gtx layout.Context) layout.Dimensions {
		return p.layoutContent(ctx, gtx, value, visual)
	}
	layoutContent := func(gtx layout.Context) layout.Dimensions {
		if p.summary == nil {
			return content(gtx)
		}
		if gtx.Constraints.Max.X < gtx.Dp(visual.Metrics.CompactWidth) {
			return layout.Flex{Axis: layout.Vertical, Alignment: layout.Start, Gap: gtx.Dp(visual.Metrics.ContentGap)}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions { return p.layoutSummary(ctx, gtx, visual) }),
				layout.Rigid(content),
			)
		}
		return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions { return p.layoutSummary(ctx, gtx, visual) }),
			layout.Rigid(content),
		)
	}
	hovered, pressed, focused := false, false, false
	for _, item := range value.items {
		hovered = hovered || item.clickable.Hovered()
		pressed = pressed || item.clickable.Pressed()
		focused = focused || gtx.Focused(&item.clickable)
	}
	state := style.State{Hovered: hovered, Pressed: pressed, Focused: focused, Disabled: p.disabled || !gtx.Enabled(), Selected: true}
	resolved := computed.ResolveStatic(ctx, state, visual.Defaults, style.Style{}, style.Style{}, visual.Override)
	return layoutui.Resolved(ctx, gtx, resolved, core.WidgetFunc(func(_ *core.Context, gtx layout.Context) layout.Dimensions { return layoutContent(gtx) }))
}

func (p Widget) layoutSummary(ctx *core.Context, gtx layout.Context, visual theme.PaginationVisual) layout.Dimensions {
	resolved := computed.ResolveStatic(ctx, style.State{}, visual.Summary, style.Style{}, style.Style{}, visual.Override)
	return layoutui.Resolved(ctx, gtx, resolved, p.summary)
}

func (p Widget) layoutContent(ctx *core.Context, gtx layout.Context, value *paginationState, visual theme.PaginationVisual) layout.Dimensions {
	items := pageItems(p.page, p.total, p.boundaries, p.siblings)
	children := make([]layout.FlexChild, 0, len(items)+2)
	if p.showControls {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return p.layoutNav(ctx, gtx, value.item("previous"), p.page-1, p.previousText(ctx), false, p.page <= 1, visual)
		}))
	}
	for _, page := range items {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if page == 0 {
				return p.layoutEllipsis(ctx, gtx, visual)
			}
			return p.layoutPage(ctx, gtx, value.item("page:"+strconv.Itoa(page)), page, page == p.page, visual)
		}))
	}
	if p.showControls {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return p.layoutNav(ctx, gtx, value.item("next"), p.page+1, p.nextText(ctx), true, p.page >= p.total, visual)
		}))
	}
	return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle, Gap: gtx.Dp(visual.Metrics.ItemGap)}.Layout(gtx, children...)
}

func (p Widget) previousText(ctx *core.Context) string {
	if p.hasPreviousLabel {
		return p.previousLabel
	}
	return ctx.Text(i18n.PaginationPrevious)
}

func (p Widget) nextText(ctx *core.Context) string {
	if p.hasNextLabel {
		return p.nextLabel
	}
	return ctx.Text(i18n.PaginationNext)
}

func (p Widget) layoutPage(ctx *core.Context, gtx layout.Context, stateValue *paginationItemState, page int, active bool, visual theme.PaginationVisual) layout.Dimensions {
	size := gtx.Dp(visual.Metrics.Size)
	gtx.Constraints = layout.Exact(image.Pt(size, size))
	label := strconv.Itoa(page)
	return p.layoutButton(ctx, gtx, stateValue, page, active, false, label, theme.PaginationPage, visual)
}

func (p Widget) layoutNav(ctx *core.Context, gtx layout.Context, stateValue *paginationItemState, target int, label string, next, disabled bool, visual theme.PaginationVisual) layout.Dimensions {
	size := min(gtx.Dp(visual.Metrics.Size), gtx.Constraints.Max.Y)
	gtx.Constraints.Min.Y = size
	gtx.Constraints.Max.Y = size
	kind := theme.PaginationPrevious
	if next {
		kind = theme.PaginationNext
	}
	return p.layoutButton(ctx, gtx, stateValue, target, false, disabled, label, kind, visual)
}

func (p Widget) layoutEllipsis(ctx *core.Context, gtx layout.Context, visual theme.PaginationVisual) layout.Dimensions {
	size := min(gtx.Dp(visual.Metrics.Size), min(gtx.Constraints.Max.X, gtx.Constraints.Max.Y))
	gtx.Constraints = layout.Exact(image.Pt(size, size))
	return p.layoutButton(ctx, gtx, nil, 0, false, true, "...", theme.PaginationEllipsis, visual)
}

func (p Widget) layoutButton(ctx *core.Context, gtx layout.Context, item *paginationItemState, target int, active, disabled bool, label string, kind theme.PaginationButtonKind, visual theme.PaginationVisual) layout.Dimensions {
	disabled = disabled || p.disabled || !gtx.Enabled()
	if item != nil && !disabled {
		for item.clickable.Clicked(gtx) {
			if target != p.page && p.onChange != nil {
				p.onChange(target)
			}
		}
		core.FocusOnPress(ctx, &item.clickable, item.clickable.History(), core.ActivePresses(item.clickable.History()))
	} else if disabled {
		gtx = gtx.Disabled()
	}
	if item == nil {
		return visual.Button(gtx, theme.PaginationButtonRequest{Kind: kind, Label: label, Disabled: true})
	}
	return item.clickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		semantic.Button.Add(gtx.Ops)
		semantic.LabelOp(label).Add(gtx.Ops)
		semantic.SelectedOp(active).Add(gtx.Ops)
		semantic.EnabledOp(!disabled).Add(gtx.Ops)

		focusVisible := core.FocusVisible(ctx, &item.clickable, gtx.Focused(&item.clickable))
		focus := item.focus.Opacity(gtx, focusVisible && !disabled, visual.Motion)
		return visual.Button(gtx, theme.PaginationButtonRequest{Kind: kind, Label: label, Active: active, Disabled: disabled, Hovered: item.clickable.Hovered(), Pressed: item.clickable.Pressed(), FocusVisible: focusVisible, Focus: focus})
	})
}

func (s *paginationState) beginFrame() {
	if s.frameItems == nil {
		s.frameItems = make(map[string]struct{})
	} else {
		clear(s.frameItems)
	}
}

func (s *paginationState) endFrame() {
	for key := range s.items {
		if _, ok := s.frameItems[key]; !ok {
			delete(s.items, key)
		}
	}
}

func (s *paginationState) item(key string) *paginationItemState {
	if s.items == nil {
		s.items = make(map[string]*paginationItemState)
	}
	s.frameItems[key] = struct{}{}
	if item := s.items[key]; item != nil {
		return item
	}
	item := new(paginationItemState)
	s.items[key] = item
	return item
}
