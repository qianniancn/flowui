package flowui

import (
	"image"

	"gioui.org/f32"
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/widget"
)

func (c ComboBoxWidget) layoutInput(ctx *Context, gtx layout.Context, state *comboBoxState, editor *widget.Editor, style inputStyle, child layout.Widget) layout.Dimensions {
	frameConstraints := gtx.Constraints
	if c.fullWidth {
		frameConstraints.Min.X = frameConstraints.Max.X
	}
	theme := ctx.Theme.Components.ComboBox
	height := min(gtx.Dp(theme.Height), frameConstraints.Max.Y)
	frameConstraints.Min.Y = min(max(frameConstraints.Min.Y, height), frameConstraints.Max.Y)

	left := gtx.Dp(ctx.Theme.Components.Input.PaddingX)
	right := gtx.Dp(theme.TriggerWidth)
	horizontalPadding := left + right
	maxX := max(frameConstraints.Max.X-horizontalPadding, 0)
	minX := min(max(frameConstraints.Min.X-horizontalPadding, 0), maxX)

	macro := op.Record(gtx.Ops)
	childGtx := gtx
	childGtx.Constraints = layout.Constraints{
		Min: image.Pt(minX, 0),
		Max: image.Pt(maxX, frameConstraints.Max.Y),
	}
	childDims := child(childGtx)
	call := macro.Stop()

	size := image.Pt(childDims.Size.X+horizontalPadding, childDims.Size.Y)
	size = frameConstraints.Constrain(size)
	rect := image.Rectangle{Max: size}
	radius := min(max(gtx.Dp(theme.Radius), 1), min(size.X, size.Y)/2)

	drawInputFrame(gtx, rect, radius, style)
	stack := op.Offset(image.Pt(left, max((size.Y-childDims.Size.Y)/2, 0))).Push(gtx.Ops)
	call.Add(gtx.Ops)
	stack.Pop()

	c.layoutTrigger(ctx, gtx, state, editor, size, style)
	state.input.addPointer(gtx, size, c.disabled)
	return layout.Dimensions{Size: size}
}

func (c ComboBoxWidget) layoutTrigger(ctx *Context, gtx layout.Context, state *comboBoxState, editor *widget.Editor, size image.Point, style inputStyle) {
	triggerSize := image.Pt(min(gtx.Dp(ctx.Theme.Components.ComboBox.TriggerWidth), size.X), size.Y)
	presses := activePresses(state.trigger.History())
	if !c.disabled {
		for state.trigger.Clicked(gtx) {
			state.open = !state.open
			ctx.requestFocus(editor)
		}
		ctx.focusOnPress(editor, state.trigger.History(), presses)
	}

	triggerGtx := gtx
	triggerGtx.Constraints = layout.Exact(triggerSize)
	if c.disabled {
		triggerGtx = triggerGtx.Disabled()
	}
	stack := op.Offset(image.Pt(size.X-triggerSize.X, 0)).Push(gtx.Ops)
	state.trigger.Layout(triggerGtx, func(gtx layout.Context) layout.Dimensions {
		drawComboBoxChevron(gtx, ctx.Theme, triggerSize, state.iconProgress(gtx, state.open), style.placeholder)
		return layout.Dimensions{Size: triggerSize}
	})
	stack.Pop()
}

func (c ComboBoxWidget) layoutOpen(ctx *Context, gtx layout.Context, state *comboBoxState, editor *widget.Editor, inputDims layout.Dimensions, visible []int, progress float32) layout.Dimensions {
	theme := ctx.Theme.Components.ComboBox
	gap := gtx.Dp(theme.PanelGap)
	panelMaxY := gtx.Constraints.Max.Y - inputDims.Size.Y - gap
	if panelMaxY <= 0 {
		state.endFrame()
		return inputDims
	}
	panelMaxY = min(panelMaxY, gtx.Dp(theme.PanelMaxHeight))
	panelConstraints := layout.Constraints{
		Min: image.Pt(inputDims.Size.X, 0),
		Max: image.Pt(inputDims.Size.X, panelMaxY),
	}
	overlayBounds := gtx.Constraints.Max

	ctx.deferOverlay(gtx, func(gtx layout.Context) layout.Dimensions {
		gtx.Constraints = panelConstraints
		placement := overlayPlacement{side: overlaySideBottom, align: overlayAlignStart}
		result := overlayResolvePosition(overlayPositionConfig{
			Trigger:       inputDims.Size,
			Panel:         panelConstraints.Max,
			Bounds:        overlayBounds,
			Offset:        gap,
			Placement:     placement,
			AvoidOverflow: true,
		})
		origin := overlayPanelTransformOrigin(inputDims.Size, result.Position, panelConstraints.Max, result.Placement)
		scale := 0.95 + 0.05*progress
		stack := op.Offset(result.Position).Push(gtx.Ops)
		transform := op.Affine(f32.AffineId().Scale(origin, f32.Pt(scale, scale))).Push(gtx.Ops)
		dims := c.layoutPanel(ctx, gtx, state, editor, visible)
		transform.Pop()
		stack.Pop()
		return dims
	})

	state.endFrame()
	return inputDims
}

func (c ComboBoxWidget) layoutPanel(ctx *Context, gtx layout.Context, state *comboBoxState, editor *widget.Editor, visible []int) layout.Dimensions {
	macro := op.Record(gtx.Ops)
	theme := ctx.Theme.Components.ComboBox
	inset := layout.UniformInset(theme.PanelPadding)
	dims := inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		if len(visible) == 0 {
			return c.layoutEmpty(ctx, gtx)
		}
		state.list.Axis = layout.Vertical
		return state.list.Layout(gtx, len(visible), func(gtx layout.Context, index int) layout.Dimensions {
			item := c.items[visible[index]]
			return c.layoutItem(ctx, gtx, state, editor, item, item.Key == c.selectedKey, index == state.highlight)
		})
	})
	call := macro.Stop()

	radius := min(max(gtx.Dp(theme.PanelRadius), 1), min(dims.Size.X, dims.Size.Y)/2)
	rect := image.Rectangle{Max: dims.Size}
	drawComboBoxPanel(gtx, ctx.Theme, rect, radius)
	clipStack := clip.UniformRRect(rect, radius).Push(gtx.Ops)
	call.Add(gtx.Ops)
	clipStack.Pop()
	return dims
}

func (c ComboBoxWidget) layoutEmpty(ctx *Context, gtx layout.Context) layout.Dimensions {
	theme := ctx.Theme.Components.ComboBox
	gtx.Constraints.Min.Y = min(max(gtx.Constraints.Min.Y, gtx.Dp(theme.ItemHeight)), gtx.Constraints.Max.Y)
	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return Text(c.emptyText).
			Size(float32(theme.ItemTextSize)).
			Color(ctx.Theme.Palette.MutedForeground).
			Layout(ctx, gtx)
	})
}

func (c ComboBoxWidget) layoutItem(ctx *Context, gtx layout.Context, state *comboBoxState, editor *widget.Editor, item ComboBoxItem, selected, highlighted bool) layout.Dimensions {
	itemState := state.item(item.Key)
	presses := activePresses(itemState.clickable.History())
	if !item.Disabled {
		for itemState.clickable.Clicked(gtx) {
			c.selectItem(editor, state, item)
			ctx.requestFocus(editor)
		}
		ctx.focusOnPress(editor, itemState.clickable.History(), presses)
	}
	if item.Disabled {
		gtx = gtx.Disabled()
	}

	return itemState.clickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		height := min(gtx.Dp(ctx.Theme.Components.ComboBox.ItemHeight), gtx.Constraints.Max.Y)
		gtx.Constraints.Min.Y = min(max(gtx.Constraints.Min.Y, height), gtx.Constraints.Max.Y)
		macro := op.Record(gtx.Ops)
		dims := c.layoutItemContent(ctx, gtx, item, selected)
		call := macro.Stop()
		dims.Size = gtx.Constraints.Constrain(dims.Size)
		style := comboBoxItemStyleFor(ctx.Theme, itemState.clickable.Hovered() || highlighted, itemState.clickable.Pressed(), selected, item.Disabled)
		style.bg = itemState.background(gtx, style.bg)
		scale := comboBoxItemScale(gtx, itemState.clickable.History(), item.Disabled)
		stack := comboBoxItemTransform(dims.Size, scale).Push(gtx.Ops)
		drawComboBoxItem(gtx, ctx.Theme, dims.Size, style)
		call.Add(gtx.Ops)
		stack.Pop()
		return dims
	})
}

func (c ComboBoxWidget) layoutItemContent(ctx *Context, gtx layout.Context, item ComboBoxItem, selected bool) layout.Dimensions {
	theme := ctx.Theme.Components.ComboBox
	return layout.Inset{
		Top:    theme.ItemPaddingY,
		Right:  theme.ItemPaddingX,
		Bottom: theme.ItemPaddingY,
		Left:   theme.ItemPaddingX,
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{
			Axis:      layout.Horizontal,
			Alignment: layout.Middle,
		}.Layout(gtx,
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				if item.Description == "" {
					return Text(item.Label).
						Size(float32(theme.ItemTextSize)).
						Weight(font.Medium).
						Color(comboBoxItemTextColor(ctx.Theme, item.Disabled)).
						Layout(ctx, gtx)
				}
				return layout.Flex{
					Axis: layout.Vertical,
				}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return Text(item.Label).
							Size(float32(theme.ItemTextSize)).
							Weight(font.Medium).
							Color(comboBoxItemTextColor(ctx.Theme, item.Disabled)).
							Layout(ctx, gtx)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return Text(item.Description).
							Size(float32(theme.ItemDescriptionSize)).
							Color(comboBoxItemDescriptionColor(ctx.Theme, item.Disabled)).
							Layout(ctx, gtx)
					}),
				)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				size := image.Pt(gtx.Dp(theme.ItemCheckSize), gtx.Dp(theme.ItemCheckSize))
				if selected {
					drawComboBoxCheck(gtx, ctx.Theme, size)
				}
				return layout.Dimensions{Size: size}
			}),
		)
	})
}
