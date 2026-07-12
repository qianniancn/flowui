package menu

import (
	"image"

	"gioui.org/font"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"github.com/qianniancn/FlowUI/internal/components/icon"
	"github.com/qianniancn/FlowUI/internal/components/text"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/render"
	"github.com/qianniancn/FlowUI/internal/state"
	"github.com/qianniancn/flowui-icons-lucide"
)

func (m Widget) layout(ctx *frame.Context, gtx layout.Context, menuState *menuState, interactive bool) layout.Dimensions {
	menuState.beginFrame()
	defer menuState.endFrame()
	items := m.actionableItems()
	menuState.checkItems(items)
	if interactive && !m.disabled {
		result := menuState.updateKeys(gtx, items, m.nested)
		if result.focusKey != "" {
			frame.RequestFocus(ctx, &menuState.item(result.focusKey).clickable)
			menuState.reveal(m.entries(), result.focusKey)
		}
		if result.actionKey != "" {
			if item, ok := itemByKey(items, result.actionKey); ok {
				if item.Kind == ItemSubmenu {
					menuState.openSubmenu = item.Key
					menuState.submenuFocusVisible = true
				} else {
					m.activate(item)
				}
			}
		}
		if result.openKey != "" {
			menuState.openSubmenu = result.openKey
			menuState.submenuFocusVisible = true
		}
		if result.close {
			m.closeToParent(ctx)
		}
	}

	m.applyConstraints(ctx, &gtx)
	style := menuPanelStyle(frame.ActiveTheme(ctx))
	macro := op.Record(gtx.Ops)
	var contentDims layout.Dimensions
	func() {
		restore := frame.PushColors(ctx, style.foreground, style.background)
		defer restore()
		contentDims = m.layoutContent(ctx, gtx, menuState, interactive)
	}()
	call := macro.Stop()
	m.registerSubmenus(ctx, gtx, menuState, interactive)
	contentDims.Size = gtx.Constraints.Constrain(contentDims.Size)
	rect := image.Rectangle{Max: contentDims.Size}
	radius := min(max(gtx.Dp(frame.ActiveTheme(ctx).Components.Menu.Radius), 1), min(rect.Dx(), rect.Dy())/2)
	drawMenuPanel(gtx, frame.ActiveTheme(ctx), rect, radius, style)
	clipStack := clip.UniformRRect(rect, radius).Push(gtx.Ops)
	call.Add(gtx.Ops)
	clipStack.Pop()
	return contentDims
}

func (m Widget) applyConstraints(ctx *frame.Context, gtx *layout.Context) {
	tokens := frame.ActiveTheme(ctx).Components.Menu
	width := tokens.Width
	if m.width > 0 {
		width = m.width
	}
	widthPx := min(max(gtx.Dp(width), 0), gtx.Constraints.Max.X)
	gtx.Constraints.Min.X = widthPx
	gtx.Constraints.Max.X = widthPx
	maxHeight := gtx.Dp(tokens.MaxHeight)
	if maxHeight > 0 {
		gtx.Constraints.Max.Y = min(gtx.Constraints.Max.Y, maxHeight)
		gtx.Constraints.Min.Y = min(gtx.Constraints.Min.Y, gtx.Constraints.Max.Y)
	}
}

func (m Widget) layoutContent(ctx *frame.Context, gtx layout.Context, menuState *menuState, interactive bool) layout.Dimensions {
	tokens := frame.ActiveTheme(ctx).Components.Menu
	padding := gtx.Dp(tokens.Padding)
	innerGtx := gtx
	innerGtx.Constraints.Min = image.Pt(max(gtx.Constraints.Min.X-padding*2, 0), 0)
	innerGtx.Constraints.Max = image.Pt(max(gtx.Constraints.Max.X-padding*2, 0), max(gtx.Constraints.Max.Y-padding*2, 0))
	entries := m.entries()
	if len(entries) == 0 {
		return layout.UniformInset(tokens.Padding).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return m.layoutEmpty(ctx, gtx)
		})
	}

	menuState.list.Axis = layout.Vertical
	menuState.list.Alignment = layout.Start
	menuState.list.Gap = gtx.Dp(tokens.ItemGap)
	menuState.list.ScrollToEnd = false
	menuState.list.ScrollAnyAxis = false
	entrySizes := make([]image.Point, len(entries))
	content := menuState.list.Layout(innerGtx, len(entries), func(gtx layout.Context, index int) layout.Dimensions {
		entry := entries[index]
		gtx.Constraints.Min.X = innerGtx.Constraints.Min.X
		gtx.Constraints.Max.X = innerGtx.Constraints.Max.X
		var dims layout.Dimensions
		switch entry.kind {
		case ItemSeparator:
			dims = m.layoutSeparator(ctx, gtx)
		case ItemGroupLabel:
			dims = m.layoutGroupLabel(ctx, gtx, entry.label)
		default:
			dims = m.layoutItem(ctx, gtx, menuState, entry.item, interactive)
		}
		entrySizes[index] = dims.Size
		return dims
	})
	y := padding - menuState.list.Position.Offset
	last := min(menuState.list.Position.First+menuState.list.Position.Count, len(entries))
	for index := menuState.list.Position.First; index < last; index++ {
		entry := entries[index]
		size := entrySizes[index]
		if entry.kind != ItemSeparator && entry.kind != ItemGroupLabel && size.Y > 0 {
			menuState.anchors[entry.item.Key] = image.Rect(padding, y, padding+size.X, y+size.Y)
		}
		y += size.Y + menuState.list.Gap
	}
	content.Size = content.Size.Add(image.Pt(padding*2, padding*2))
	content.Size = gtx.Constraints.Constrain(content.Size)
	return content
}

func (m Widget) layoutEmpty(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	tokens := frame.ActiveTheme(ctx).Components.Menu
	height := min(gtx.Dp(tokens.ItemMinHeight), gtx.Constraints.Max.Y)
	gtx.Constraints.Min.Y = min(max(gtx.Constraints.Min.Y, height), gtx.Constraints.Max.Y)
	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return text.New(m.emptyText).Size(float32(tokens.ItemTextSize)).Color(frame.ActiveTheme(ctx).Palette.MutedForeground).Layout(ctx, gtx)
	})
}

func (m Widget) layoutGroupLabel(ctx *frame.Context, gtx layout.Context, label string) layout.Dimensions {
	tokens := frame.ActiveTheme(ctx).Components.Menu
	return layout.Inset{Top: tokens.SectionPaddingY, Right: tokens.SectionPaddingX, Bottom: tokens.SectionPaddingY, Left: tokens.SectionPaddingX}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return text.New(label).Size(float32(tokens.SectionTextSize)).Weight(font.Medium).Color(frame.ActiveTheme(ctx).Palette.MutedForeground).Layout(ctx, gtx)
	})
}

func (m Widget) layoutSeparator(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	tokens := frame.ActiveTheme(ctx).Components.Menu
	marginX := gtx.Dp(tokens.SeparatorMarginX)
	marginY := gtx.Dp(tokens.SeparatorMarginY)
	width := max(gtx.Constraints.Min.X-marginX*2, 0)
	height := max(gtx.Dp(tokens.SeparatorWidth), 1)
	size := gtx.Constraints.Constrain(image.Pt(width+marginX*2, height+marginY*2))
	offset := op.Offset(image.Pt(marginX, marginY)).Push(gtx.Ops)
	drawMenuSeparator(gtx, image.Pt(max(size.X-marginX*2, 0), min(height, size.Y)), frame.ActiveTheme(ctx).Palette.Border)
	offset.Pop()
	return layout.Dimensions{Size: size}
}

func (m Widget) layoutItem(ctx *frame.Context, gtx layout.Context, menuState *menuState, item Item, interactive bool) layout.Dimensions {
	itemState := menuState.item(item.Key)
	disabled := m.disabled || item.Disabled
	animGtx := gtx
	eventGtx := gtx
	if disabled || !interactive {
		eventGtx = eventGtx.Disabled()
	} else {
		presses := state.ActivePresses(itemState.clickable.History())
		for itemState.clickable.Clicked(eventGtx) {
			frame.RequestFocus(ctx, &itemState.clickable)
			if item.Kind == ItemSubmenu {
				menuState.openSubmenu = item.Key
				menuState.submenuFocusVisible = false
			} else {
				m.activate(item)
			}
		}
		frame.FocusOnPress(ctx, &itemState.clickable, itemState.clickable.History(), presses)
	}

	dims := itemState.clickable.Layout(eventGtx, func(gtx layout.Context) layout.Dimensions {
		class := semantic.Button
		if item.Kind == ItemCheckbox {
			class = semantic.CheckBox
		} else if item.Kind == ItemRadio {
			class = semantic.RadioButton
		}
		class.Add(gtx.Ops)
		semantic.LabelOp(item.Label).Add(gtx.Ops)
		if item.Description != "" {
			semantic.DescriptionOp(item.Description).Add(gtx.Ops)
		}
		if item.Kind == ItemCheckbox || item.Kind == ItemRadio {
			semantic.SelectedOp(item.Checked).Add(gtx.Ops)
		}
		semantic.EnabledOp(!disabled).Add(gtx.Ops)

		tokens := frame.ActiveTheme(ctx).Components.Menu
		minHeight := min(gtx.Dp(tokens.ItemMinHeight), gtx.Constraints.Max.Y)
		focusVisible := itemState.focus.Visible(gtx.Focused(&itemState.clickable), itemState.clickable.History())
		style := menuItemStyle(frame.ActiveTheme(ctx), item.Variant, itemState.clickable.Hovered(), disabled)
		style.focus = itemState.focus.Opacity(animGtx, focusVisible && !disabled)
		scale := menuItemScale(animGtx, itemState.clickable.History(), frame.ActiveTheme(ctx), disabled)
		macro := op.Record(gtx.Ops)
		contentGtx := gtx
		contentGtx.Constraints.Min.Y = 0
		contentDims := m.layoutItemContent(ctx, contentGtx, item, style)
		call := macro.Stop()
		size := gtx.Constraints.Constrain(image.Pt(max(contentDims.Size.X, gtx.Constraints.Min.X), max(minHeight, contentDims.Size.Y)))
		radius := min(max(gtx.Dp(tokens.ItemRadius), 1), min(size.X, size.Y)/2)
		opacity := paint.PushOpacity(gtx.Ops, style.opacity)
		transform := render.Scale(size, scale).Push(gtx.Ops)
		drawMenuItem(gtx, frame.ActiveTheme(ctx), size, radius, style)
		offset := op.Offset(image.Pt(0, max((size.Y-contentDims.Size.Y)/2, 0))).Push(gtx.Ops)
		call.Add(gtx.Ops)
		offset.Pop()
		transform.Pop()
		opacity.Pop()
		return layout.Dimensions{Size: size}
	})
	if interactive && !disabled {
		m.updateSubmenuHover(gtx, menuState, item, itemState.clickable.Hovered())
	}
	return dims
}

func (m Widget) layoutItemContent(ctx *frame.Context, gtx layout.Context, item Item, style itemStyle) layout.Dimensions {
	tokens := frame.ActiveTheme(ctx).Components.Menu
	hasIndicators := m.hasIndicators()
	return layout.Inset{Top: tokens.ItemPaddingY, Right: tokens.ItemPaddingX, Bottom: tokens.ItemPaddingY, Left: tokens.ItemPaddingX}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		children := make([]layout.FlexChild, 0, 11)
		gap := func(dp unit.Dp) layout.FlexChild {
			return layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Spacer{Width: dp}.Layout(gtx)
			})
		}
		if hasIndicators {
			children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return m.layoutSelectionIndicator(ctx, gtx, item, style)
			}))
			children = append(children, gap(tokens.IndicatorContentGap))
		}
		if item.Leading != nil {
			children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return item.Leading.Layout(ctx, gtx)
			}))
			children = append(children, gap(tokens.ItemContentGap))
		}
		children = append(children, layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return m.layoutItemText(ctx, gtx, item, style)
		}))
		if item.Shortcut != "" {
			children = append(children, gap(tokens.ItemContentGap))
			children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return text.New(item.Shortcut).Size(float32(tokens.ShortcutTextSize)).Color(style.shortcut).Layout(ctx, gtx)
			}))
		}
		if item.Trailing != nil {
			children = append(children, gap(tokens.ItemContentGap))
			children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return item.Trailing.Layout(ctx, gtx)
			}))
		}
		if item.Kind == ItemSubmenu {
			children = append(children, gap(tokens.ItemContentGap))
			children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return icon.New(lucide.ChevronRight).Size(float32(tokens.SubmenuIndicatorSize)).Color(style.shortcut).Layout(ctx, gtx)
			}))
		}
		return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx, children...)
	})
}

func (m Widget) layoutItemText(ctx *frame.Context, gtx layout.Context, item Item, style itemStyle) layout.Dimensions {
	tokens := frame.ActiveTheme(ctx).Components.Menu
	if item.Description == "" {
		return text.New(item.Label).Size(float32(tokens.ItemTextSize)).Color(style.foreground).Layout(ctx, gtx)
	}
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return text.New(item.Label).Size(float32(tokens.ItemTextSize)).Color(style.foreground).Layout(ctx, gtx)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return text.New(item.Description).Size(float32(tokens.ItemDescriptionSize)).Color(style.description).Layout(ctx, gtx)
		}),
	)
}

func (m Widget) layoutSelectionIndicator(ctx *frame.Context, gtx layout.Context, item Item, style itemStyle) layout.Dimensions {
	sizeDp := frame.ActiveTheme(ctx).Components.Menu.IndicatorSize
	sizePx := gtx.Dp(sizeDp)
	size := image.Pt(sizePx, sizePx)
	gtx.Constraints = layout.Exact(size)
	if !item.Checked || (item.Kind != ItemCheckbox && item.Kind != ItemRadio) {
		return layout.Dimensions{Size: size}
	}
	if item.Kind == ItemRadio {
		drawRadioIndicator(gtx, size, gtx.Dp(frame.ActiveTheme(ctx).Components.Menu.RadioDotSize), style.indicator)
		return layout.Dimensions{Size: size}
	}
	iconSize := frame.ActiveTheme(ctx).Components.Menu.CheckmarkSize
	iconGtx := gtx
	iconGtx.Constraints = layout.Exact(image.Pt(gtx.Dp(iconSize), gtx.Dp(iconSize)))
	return layout.Center.Layout(gtx, func(layout.Context) layout.Dimensions {
		return icon.New(lucide.Check).Size(float32(iconSize)).Color(style.indicator).Layout(ctx, iconGtx)
	})
}

func (m Widget) hasIndicators() bool {
	for _, item := range m.actionableItems() {
		if item.Kind == ItemCheckbox || item.Kind == ItemRadio {
			return true
		}
	}
	return false
}
