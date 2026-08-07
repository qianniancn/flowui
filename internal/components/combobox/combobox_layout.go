package combobox

import (
	"image"

	"gioui.org/f32"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/widget"
	layoutui "github.com/qianniancn/flowui/internal/components/layout"
	"github.com/qianniancn/flowui/internal/components/optionrow"
	"github.com/qianniancn/flowui/internal/components/text"
	"github.com/qianniancn/flowui/internal/field"
	"github.com/qianniancn/flowui/internal/frame"
	"github.com/qianniancn/flowui/internal/overlay"
	"github.com/qianniancn/flowui/internal/render"
	"github.com/qianniancn/flowui/internal/state"
)

func (c ComboBoxWidget) layoutInput(ctx *frame.Context, gtx layout.Context, state *comboBoxState, editor *widget.Editor, style field.Resolved, child layout.Widget) layout.Dimensions {
	return layoutui.LayoutResolved(ctx, gtx, style.Content, frame.WidgetFunc(func(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
		return c.layoutInputContent(ctx, gtx, state, editor, style.Colors, child)
	}))
}

func (c ComboBoxWidget) layoutInputContent(ctx *frame.Context, gtx layout.Context, state *comboBoxState, editor *widget.Editor, colors field.Colors, child layout.Widget) layout.Dimensions {
	frameConstraints := gtx.Constraints
	if c.fullWidth {
		frameConstraints.Min.X = frameConstraints.Max.X
	}
	theme := frame.ActiveTheme(ctx).Components.ComboBox
	height := min(gtx.Dp(theme.Height), frameConstraints.Max.Y)
	frameConstraints.Min.Y = min(max(frameConstraints.Min.Y, height), frameConstraints.Max.Y)

	left := gtx.Dp(frame.ActiveTheme(ctx).Components.Input.PaddingX)
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
	stack := op.Offset(image.Pt(left, max((size.Y-childDims.Size.Y)/2, 0))).Push(gtx.Ops)
	call.Add(gtx.Ops)
	stack.Pop()

	state.input.AddPointer(gtx, size, c.disabled)
	c.layoutTrigger(ctx, gtx, state, editor, size, colors)
	return layout.Dimensions{Size: size}
}

func (c ComboBoxWidget) layoutTrigger(ctx *frame.Context, gtx layout.Context, comboState *comboBoxState, editor *widget.Editor, size image.Point, colors field.Colors) {
	triggerSize := image.Pt(min(gtx.Dp(frame.ActiveTheme(ctx).Components.ComboBox.TriggerWidth), size.X), size.Y)
	presses := state.SnapshotPresses(comboState.trigger.History())
	if !c.disabled {
		for comboState.trigger.Clicked(gtx) {
			comboState.open = !comboState.open
			frame.RequestFocusVisible(ctx, editor, presses.ClickFocusVisible(comboState.trigger.History()))
		}
	}

	triggerGtx := gtx
	triggerGtx.Constraints = layout.Exact(triggerSize)
	if c.disabled {
		triggerGtx = triggerGtx.Disabled()
	}
	stack := op.Offset(image.Pt(size.X-triggerSize.X, 0)).Push(gtx.Ops)
	comboState.trigger.Layout(triggerGtx, func(gtx layout.Context) layout.Dimensions {
		if !c.disabled {
			pointer.CursorPointer.Add(gtx.Ops)
		}
		drawComboBoxChevron(gtx, frame.ActiveTheme(ctx), triggerSize, comboState.iconProgress(gtx, comboState.open, frame.ActiveTheme(ctx).Motion), colors.Placeholder)
		return layout.Dimensions{Size: triggerSize}
	})
	stack.Pop()
}

func (c ComboBoxWidget) layoutOpen(ctx *frame.Context, gtx layout.Context, state *comboBoxState, editor *widget.Editor, inputDims layout.Dimensions, visible []int, progress float32, naturallyDisabled bool) layout.Dimensions {
	frame.RegisterOverlay(ctx, frame.OverlayRequest{
		Key:       frame.FullKey(ctx, c.key),
		Layer:     frame.OverlayLayerPopup,
		Anchor:    image.Rectangle{Max: inputDims.Size},
		HasAnchor: true,
		Disabled:  naturallyDisabled,
		Layout: func(gtx layout.Context, anchor image.Rectangle, interactive bool) layout.Dimensions {
			panelVisible := visible
			contentInteractive := interactive && state.open && gtx.Enabled()
			if contentInteractive {
				state.clampHighlight(c.items, panelVisible)
				if index, ok := state.updateKeys(gtx, editor, c.items, panelVisible); ok {
					c.selectItem(editor, state, c.items[panelVisible[index]])
				}
				c.updateEditor(editor, state, gtx)
				query := editor.Text()
				selectedLabel, _ := state.selectedLabel(c)
				panelVisible = state.visibleItems(c, query, selectedLabel)
				state.clampHighlight(c.items, panelVisible)
			}
			panelGtx := gtx
			if !contentInteractive {
				panelGtx = panelGtx.Disabled()
			}
			return c.layoutPanelOverlay(ctx, gtx, panelGtx, state, editor, panelVisible, anchor, progress, contentInteractive)
		},
	})
	frame.AfterOverlays(ctx, state.endFrame)
	return inputDims
}

func (c ComboBoxWidget) layoutPanelOverlay(ctx *frame.Context, gtx, panelGtx layout.Context, comboState *comboBoxState, editor *widget.Editor, visible []int, anchor image.Rectangle, progress float32, interactive bool) layout.Dimensions {
	if interactive {
		for comboState.dialog.Clicked(gtx) {
			frame.RequestFocusVisible(ctx, editor, false)
		}
		if comboState.dialog.TakePressed() {
			frame.RequestFocusVisible(ctx, editor, false)
		}
	}
	theme := frame.ActiveTheme(ctx).Components.ComboBox
	viewport := gtx.Constraints.Max
	gap := gtx.Dp(theme.PanelGap)
	panelWidth := min(max(anchor.Dx(), 0), max(viewport.X, 0))
	panelMaxY := min(gtx.Dp(theme.PanelMaxHeight), max(viewport.Y-gap, 0))
	if panelWidth <= 0 || panelMaxY <= 0 {
		return layout.Dimensions{}
	}
	panelGtx.Constraints = layout.Constraints{
		Min: image.Pt(panelWidth, 0),
		Max: image.Pt(panelWidth, panelMaxY),
	}

	macro := op.Record(gtx.Ops)
	dims, tracked := frame.TrackOverlayPlacement(ctx, func() layout.Dimensions {
		return c.layoutPanel(ctx, panelGtx, comboState, editor, visible)
	})
	call := macro.Stop()
	placement := overlay.Placement{Side: overlay.SideBottom, Align: overlay.AlignStart}
	result := overlay.ResolvePosition(overlay.PositionConfig{
		Trigger:          anchor.Size(),
		TriggerOrigin:    anchor.Min,
		HasTriggerOrigin: true,
		Panel:            dims.Size,
		Bounds:           viewport,
		Offset:           gap,
		Placement:        placement,
		Flip:             true,
		AvoidOverflow:    true,
	})
	origin := overlay.PanelTransformOriginAt(anchor, result.Position, dims.Size, result.Placement)
	scale := 0.95 + 0.05*progress
	scaleTransform := f32.AffineId().Scale(origin, f32.Pt(scale, scale))
	tracked.PlaceTransform(f32.AffineId().Offset(f32.Pt(float32(result.Position.X), float32(result.Position.Y))).Mul(scaleTransform))

	stack := op.Offset(result.Position).Push(gtx.Ops)
	transform := op.Affine(scaleTransform).Push(gtx.Ops)
	layoutComboBoxPanelBlocker(gtx, comboState, dims.Size)
	call.Add(gtx.Ops)
	transform.Pop()
	stack.Pop()
	return dims
}

func layoutComboBoxPanelBlocker(gtx layout.Context, state *comboBoxState, size image.Point) {
	if size.X <= 0 || size.Y <= 0 {
		return
	}
	blockerGtx := gtx
	blockerGtx.Constraints = layout.Exact(size)
	state.dialog.Layout(blockerGtx, func(layout.Context) layout.Dimensions {
		return layout.Dimensions{Size: size}
	})
}

func (c ComboBoxWidget) layoutPanel(ctx *frame.Context, gtx layout.Context, state *comboBoxState, editor *widget.Editor, visible []int) layout.Dimensions {
	macro := op.Record(gtx.Ops)
	theme := frame.ActiveTheme(ctx).Components.ComboBox
	inset := layout.UniformInset(theme.PanelPadding)
	dims := inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		if len(visible) == 0 {
			return c.layoutEmpty(ctx, gtx)
		}
		state.list.Axis = layout.Vertical
		return layoutui.LayoutTrackedScrollbarWithVisualOutset(ctx, gtx, &state.list, &state.bar, &state.visualOutset, len(visible), !gtx.Enabled(), false, func(gtx layout.Context, index int) layout.Dimensions {
			item := c.items[visible[index]]
			return c.layoutItem(ctx, gtx, state, editor, item, item.Key == c.selectedKey, index == state.highlight)
		})
	})
	call := macro.Stop()

	radius := min(max(gtx.Dp(theme.PanelRadius), 1), min(dims.Size.X, dims.Size.Y)/2)
	rect := image.Rectangle{Max: dims.Size}
	drawComboBoxPanel(gtx, frame.ActiveTheme(ctx), rect, radius)
	clipStack := clip.UniformRRect(rect, radius).Push(gtx.Ops)
	call.Add(gtx.Ops)
	clipStack.Pop()
	return dims
}

func (c ComboBoxWidget) layoutEmpty(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	theme := frame.ActiveTheme(ctx).Components.ComboBox
	gtx.Constraints.Min.Y = min(max(gtx.Constraints.Min.Y, gtx.Dp(theme.ItemHeight)), gtx.Constraints.Max.Y)
	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return text.New(c.emptyText).
			Size(float32(theme.ItemTextSize)).
			Color(frame.ActiveTheme(ctx).Palette.MutedForeground).
			Layout(ctx, gtx)
	})
}

func (c ComboBoxWidget) layoutItem(ctx *frame.Context, gtx layout.Context, comboState *comboBoxState, editor *widget.Editor, item ComboBoxItem, selected, highlighted bool) layout.Dimensions {
	itemState := comboState.item(item.Key)
	presses := state.SnapshotPresses(itemState.Clickable.History())
	if !item.Disabled {
		for itemState.Clickable.Clicked(gtx) {
			c.selectItem(editor, comboState, item)
			frame.RequestFocusVisible(ctx, editor, presses.ClickFocusVisible(itemState.Clickable.History()))
		}
		frame.FocusOnPress(ctx, editor, itemState.Clickable.History(), presses.Active())
	}
	if item.Disabled {
		gtx = gtx.Disabled()
	}

	return itemState.Clickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		minHeight := min(gtx.Dp(frame.ActiveTheme(ctx).Components.ComboBox.ItemHeight), gtx.Constraints.Max.Y)
		macro := op.Record(gtx.Ops)
		contentGtx := gtx
		contentGtx.Constraints.Min.Y = 0
		motion := frame.ActiveTheme(ctx).Motion
		contentDims := c.layoutItemContent(ctx, contentGtx, item, itemState.Selection(gtx, selected, comboBoxItemSelectDuration, motion))
		call := macro.Stop()
		size, contentOffset := comboBoxItemFrame(gtx.Constraints, minHeight, contentDims.Size)
		dims := contentDims
		dims.Size = size
		dims.Baseline += max(size.Y-contentOffset.Y-contentDims.Size.Y, 0)
		style := comboBoxItemStyleFor(frame.ActiveTheme(ctx), itemState.Clickable.Hovered() || highlighted, itemState.Clickable.Pressed(), item.Disabled)
		style.bg = itemState.Background(gtx, style.bg, comboBoxItemColorDuration, motion)
		scale := optionrow.PressScale(gtx, itemState.Clickable.History(), item.Disabled, 0.98, comboBoxItemPressInDuration, comboBoxItemPressDuration, motion)
		stack := render.Scale(dims.Size, scale).Push(gtx.Ops)
		drawComboBoxItem(gtx, frame.ActiveTheme(ctx), dims.Size, style)
		offset := op.Offset(contentOffset).Push(gtx.Ops)
		call.Add(gtx.Ops)
		offset.Pop()
		stack.Pop()
		return dims
	})
}

func comboBoxItemFrame(constraints layout.Constraints, minHeight int, content image.Point) (size, offset image.Point) {
	return optionrow.Frame(constraints, minHeight, content)
}

func (c ComboBoxWidget) layoutItemContent(ctx *frame.Context, gtx layout.Context, item ComboBoxItem, selection float32) layout.Dimensions {
	theme := frame.ActiveTheme(ctx).Components.ComboBox
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
				return optionrow.LayoutText(
					ctx, gtx, item.Label, item.Description,
					float32(theme.ItemTextSize), float32(theme.ItemDescriptionSize),
					comboBoxItemTextColor(frame.ActiveTheme(ctx), item.Disabled),
					comboBoxItemDescriptionColor(frame.ActiveTheme(ctx), item.Disabled),
				)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				size := image.Pt(gtx.Dp(theme.ItemCheckSize), gtx.Dp(theme.ItemCheckSize))
				drawComboBoxCheck(gtx, frame.ActiveTheme(ctx), size, selection)
				return layout.Dimensions{Size: size}
			}),
		)
	})
}
