package pagination

import (
	"image"
	"image/color"
	"strconv"

	"gioui.org/font"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/qianniancn/FlowUI/internal/components/icon"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/render"
	"github.com/qianniancn/FlowUI/internal/state"
	"github.com/qianniancn/FlowUI/internal/theme"
	"github.com/qianniancn/flowui-icons-lucide"
)

const stateSlotPagination = "pagination"

type paginationState struct {
	items      map[string]*paginationItemState
	frameItems map[string]struct{}
}

type paginationItemState struct {
	clickable widget.Clickable
	focus     state.FocusAnimation
}

type paginationSizeStyle struct {
	size     unit.Dp
	textSize unit.Sp
	paddingX unit.Dp
	iconSize unit.Dp
}

func (p Widget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	key := frame.ClaimKey(ctx, state.KindPagination, p.key)
	value := frame.UseState[paginationState](ctx, key, stateSlotPagination)
	value.beginFrame()
	defer value.endFrame()

	semantic.LabelOp("Pagination").Add(gtx.Ops)
	content := func(gtx layout.Context) layout.Dimensions {
		return p.layoutContent(ctx, gtx, value)
	}
	if p.summary == nil {
		return content(gtx)
	}
	tokens := frame.ActiveTheme(ctx).Components.Pagination
	if gtx.Constraints.Max.X < gtx.Dp(tokens.CompactWidth) {
		return layout.Flex{Axis: layout.Vertical, Alignment: layout.Start, Gap: gtx.Dp(tokens.ContentGap)}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions { return p.layoutSummary(ctx, gtx) }),
			layout.Rigid(content),
		)
	}
	return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions { return p.layoutSummary(ctx, gtx) }),
		layout.Rigid(content),
	)
}

func (p Widget) layoutSummary(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	restore := frame.PushColors(ctx, frame.ActiveTheme(ctx).Palette.MutedForeground, ctx.BackgroundColor())
	defer restore()
	return p.summary.Layout(ctx, gtx)
}

func (p Widget) layoutContent(ctx *frame.Context, gtx layout.Context, value *paginationState) layout.Dimensions {
	items := pageItems(p.page, p.total, p.boundaries, p.siblings)
	children := make([]layout.FlexChild, 0, len(items)+2)
	if p.showControls {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return p.layoutNav(ctx, gtx, value.item("previous"), p.page-1, p.previousLabel, false, p.page <= 1)
		}))
	}
	for _, page := range items {
		page := page
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if page == 0 {
				return p.layoutEllipsis(ctx, gtx)
			}
			return p.layoutPage(ctx, gtx, value.item("page:"+strconv.Itoa(page)), page, page == p.page)
		}))
	}
	if p.showControls {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return p.layoutNav(ctx, gtx, value.item("next"), p.page+1, p.nextLabel, true, p.page >= p.total)
		}))
	}
	return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle, Gap: gtx.Dp(frame.ActiveTheme(ctx).Components.Pagination.ItemGap)}.Layout(gtx, children...)
}

func (p Widget) layoutPage(ctx *frame.Context, gtx layout.Context, stateValue *paginationItemState, page int, active bool) layout.Dimensions {
	sizeStyle := paginationSize(frame.ActiveTheme(ctx), p.size)
	size := gtx.Dp(sizeStyle.size)
	gtx.Constraints = layout.Exact(image.Pt(size, size))
	label := strconv.Itoa(page)
	return p.layoutButton(ctx, gtx, stateValue, page, active, false, label, func(gtx layout.Context, foreground color.NRGBA) layout.Dimensions {
		return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return paginationLabel(ctx, gtx, label, sizeStyle.textSize, foreground)
		})
	})
}

func (p Widget) layoutNav(ctx *frame.Context, gtx layout.Context, stateValue *paginationItemState, target int, label string, next, disabled bool) layout.Dimensions {
	sizeStyle := paginationSize(frame.ActiveTheme(ctx), p.size)
	size := min(gtx.Dp(sizeStyle.size), gtx.Constraints.Max.Y)
	gtx.Constraints.Min.Y = size
	gtx.Constraints.Max.Y = size
	return p.layoutButton(ctx, gtx, stateValue, target, false, disabled, label, func(gtx layout.Context, foreground color.NRGBA) layout.Dimensions {
		return layout.Inset{Left: sizeStyle.paddingX, Right: sizeStyle.paddingX}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			iconChild := func(gtx layout.Context) layout.Dimensions {
				diameter := min(gtx.Dp(sizeStyle.iconSize), min(gtx.Constraints.Max.X, gtx.Constraints.Max.Y))
				iconGtx := gtx
				iconGtx.Constraints = layout.Exact(image.Pt(diameter, diameter))
				data := lucide.ChevronLeft
				if next {
					data = lucide.ChevronRight
				}
				return icon.Layout(data, iconGtx, foreground)
			}
			labelChild := func(gtx layout.Context) layout.Dimensions {
				return paginationLabel(ctx, gtx, label, sizeStyle.textSize, foreground)
			}
			children := []layout.FlexChild{layout.Rigid(iconChild), layout.Rigid(labelChild)}
			if next {
				children[0], children[1] = children[1], children[0]
			}
			return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle, Gap: gtx.Dp(6)}.Layout(gtx, children...)
		})
	})
}

func (p Widget) layoutEllipsis(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	sizeStyle := paginationSize(frame.ActiveTheme(ctx), p.size)
	size := min(gtx.Dp(sizeStyle.size), min(gtx.Constraints.Max.X, gtx.Constraints.Max.Y))
	gtx.Constraints = layout.Exact(image.Pt(size, size))
	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return paginationLabel(ctx, gtx, "...", sizeStyle.textSize, frame.ActiveTheme(ctx).Palette.MutedForeground)
	})
}

func (p Widget) layoutButton(ctx *frame.Context, gtx layout.Context, item *paginationItemState, target int, active, disabled bool, label string, child func(layout.Context, color.NRGBA) layout.Dimensions) layout.Dimensions {
	disabled = disabled || p.disabled || !gtx.Enabled()
	if !disabled {
		for item.clickable.Clicked(gtx) {
			if target != p.page && p.onChange != nil {
				p.onChange(target)
			}
		}
		frame.FocusOnPress(ctx, &item.clickable, item.clickable.History(), state.ActivePresses(item.clickable.History()))
	} else {
		gtx = gtx.Disabled()
	}
	return item.clickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		semantic.Button.Add(gtx.Ops)
		semantic.LabelOp(label).Add(gtx.Ops)
		semantic.SelectedOp(active).Add(gtx.Ops)
		semantic.EnabledOp(!disabled).Add(gtx.Ops)

		activeTheme := frame.ActiveTheme(ctx)
		background := color.NRGBA{}
		if active {
			background = activeTheme.Palette.DefaultColor()
		}
		if item.clickable.Hovered() || item.clickable.Pressed() {
			background = activeTheme.Palette.DefaultHoverColor()
		}
		foreground := activeTheme.Palette.DefaultForegroundColor()
		if disabled {
			background = activeTheme.DisabledColor(background)
			foreground = activeTheme.DisabledColor(foreground)
		}
		focusVisible := frame.FocusVisible(ctx, &item.clickable, gtx.Focused(&item.clickable))
		focus := item.focus.Opacity(gtx, focusVisible && !disabled)

		macro := op.Record(gtx.Ops)
		dims := layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return child(gtx, foreground)
		})
		call := macro.Stop()
		scale := float32(1)
		if item.clickable.Pressed() && !disabled {
			scale = 0.97
		}
		transform := render.Scale(dims.Size, scale).Push(gtx.Ops)
		drawPaginationButton(gtx, dims.Size, background, activeTheme.Palette.Focus, focus, activeTheme.Components.Pagination.FocusRingWidth)
		call.Add(gtx.Ops)
		transform.Pop()
		return dims
	})
}

func paginationLabel(ctx *frame.Context, gtx layout.Context, value string, size unit.Sp, foreground color.NRGBA) layout.Dimensions {
	label := material.Label(frame.ActiveTheme(ctx).Material, size, value)
	label.Color = foreground
	label.Font.Weight = font.Medium
	return label.Layout(gtx)
}

func paginationSize(activeTheme *theme.Theme, size Size) paginationSizeStyle {
	tokens := activeTheme.Components.Pagination
	style := paginationSizeStyle{size: tokens.MediumSize, textSize: tokens.MediumTextSize, paddingX: tokens.MediumPaddingX, iconSize: tokens.IconSize}
	switch size {
	case SizeSmall:
		style.size, style.textSize, style.paddingX = tokens.SmallSize, tokens.SmallTextSize, tokens.SmallPaddingX
	case SizeLarge:
		style.size, style.textSize, style.paddingX = tokens.LargeSize, tokens.LargeTextSize, tokens.LargePaddingX
	}
	return style
}

func drawPaginationButton(gtx layout.Context, size image.Point, background, focusColor color.NRGBA, focus float32, focusWidth unit.Dp) {
	rect := image.Rectangle{Max: size}
	radius := min(size.X, size.Y) / 2
	if background.A != 0 {
		paint.FillShape(gtx.Ops, background, clip.UniformRRect(rect, radius).Op(gtx.Ops))
	}
	if focus <= 0 {
		return
	}
	width := max(gtx.Dp(focusWidth), 1)
	focusRect := rect.Inset(width + 1)
	if focusRect.Empty() {
		return
	}
	focusColor.A = byte(float32(focusColor.A)*focus + 0.5)
	stroke := clip.Stroke{Path: clip.UniformRRect(focusRect, max(radius-width-1, 0)).Path(gtx.Ops), Width: float32(width)}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, focusColor)
	stroke.Pop()
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
