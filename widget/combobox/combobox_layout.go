package combobox

import (
	"image"

	"gioui.org/f32"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/i18n"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/optionrow"
)

func (c Widget) layoutInput(ctx *core.Context, gtx layout.Context, state *comboBoxState, editor *widget.Editor, resolved comboResolvedStyle, child layout.Widget) layout.Dimensions {
	return layoutui.InteractiveResolved(ctx, gtx, resolved.content, core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		return c.layoutInputContent(ctx, gtx, state, editor, resolved, child)
	}), func(gtx layout.Context, visual layout.Widget) layout.Dimensions {
		dims := visual(gtx)
		state.input.AddPointer(gtx, dims.Size, c.disabled)
		return dims
	})
}

func (c Widget) layoutInputContent(ctx *core.Context, gtx layout.Context, state *comboBoxState, editor *widget.Editor, resolved comboResolvedStyle, child layout.Widget) layout.Dimensions {
	frameConstraints := gtx.Constraints
	if c.fullWidth {
		frameConstraints.Min.X = frameConstraints.Max.X
	}
	indicatorWidth := gtx.Dp(resolved.visual.IndicatorWidth(resolved.indicator))
	maxX := max(frameConstraints.Max.X-indicatorWidth, 0)
	minX := min(max(frameConstraints.Min.X-indicatorWidth, 0), maxX)

	macro := op.Record(gtx.Ops)
	childGtx := gtx
	childGtx.Constraints = layout.Constraints{
		Min: image.Pt(minX, 0),
		Max: image.Pt(maxX, frameConstraints.Max.Y),
	}
	childDims := child(childGtx)
	call := macro.Stop()

	size := image.Pt(childDims.Size.X+indicatorWidth, childDims.Size.Y)
	size = frameConstraints.Constrain(size)
	stack := op.Offset(image.Pt(0, max((size.Y-childDims.Size.Y)/2, 0))).Push(gtx.Ops)
	call.Add(gtx.Ops)
	stack.Pop()

	c.layoutTrigger(ctx, gtx, state, editor, size, resolved)
	return layout.Dimensions{Size: size}
}

func (c Widget) layoutTrigger(ctx *core.Context, gtx layout.Context, comboState *comboBoxState, editor *widget.Editor, size image.Point, resolved comboResolvedStyle) {
	triggerSize := image.Pt(min(gtx.Dp(resolved.visual.IndicatorWidth(resolved.indicator)), size.X), size.Y)
	presses := core.SnapshotPresses(comboState.trigger.History())
	if !c.disabled {
		for comboState.trigger.Clicked(gtx) {
			comboState.open = !comboState.open
			core.RequestFocusVisible(ctx, editor, presses.ClickFocusVisible(comboState.trigger.History()))
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
		if resolved.visual.Chevron != nil {
			resolved.visual.Chevron(gtx, triggerSize,
				comboState.iconProgress(gtx, comboState.open, resolved.visual.Metrics.IconDuration, resolved.visual.Motion),
				resolved.icon)
		}
		return layout.Dimensions{Size: triggerSize}
	})
	stack.Pop()
}

func (c Widget) layoutOpen(ctx *core.Context, gtx layout.Context, state *comboBoxState, editor *widget.Editor, inputDims layout.Dimensions, visible []int, progress float32, naturallyDisabled bool, overrides ...style.Style) layout.Dimensions {
	bundle := comboStyleDeclarationsCached(core.ActiveTheme(ctx), c.variant, &state.styleCache)
	return c.layoutOpenWithBundle(ctx, gtx, state, editor, inputDims, visible, progress, naturallyDisabled, firstComboStyle(overrides), bundle)
}

func (c Widget) layoutOpenWithBundle(ctx *core.Context, gtx layout.Context, state *comboBoxState, editor *widget.Editor, inputDims layout.Dimensions, visible []int, progress float32, naturallyDisabled bool, instanceStyle style.Style, bundle comboDeclarationBundle) layout.Dimensions {
	core.RegisterOverlay(ctx, core.OverlayRequest{
		Key:       core.FullKey(ctx, c.key),
		Layer:     core.OverlayLayerPopup,
		Anchor:    image.Rectangle{Max: inputDims.Size},
		HasAnchor: true,
		Disabled:  naturallyDisabled,
		Layout: func(gtx layout.Context, anchor image.Rectangle, interactive bool) layout.Dimensions {
			panelVisible := visible
			contentInteractive := interactive && state.open && gtx.Enabled()
			if contentInteractive {
				state.clampHighlight(c.items, panelVisible)
				if index, ok := state.updateKeys(gtx, editor, c.items, panelVisible, true); ok {
					c.selectItem(editor, state, c.items[panelVisible[index]])
				}
				c.updateEditor(editor, state, gtx)
				query := editor.Text()
				selectedLabel, _ := state.selectedLabel(c)
				panelVisible = state.visibleItems(c, query, selectedLabel)
				state.clampHighlight(c.items, panelVisible)
			}
			panelGtx := gtx
			// Keep an open panel enabled even when a nested overlay owns
			// policy input; disabling it would unregister that overlay.
			if !state.open || !gtx.Enabled() {
				panelGtx = panelGtx.Disabled()
			}
			return c.layoutPanelOverlayWithBundle(ctx, gtx, panelGtx, state, editor, panelVisible, anchor, progress, contentInteractive, instanceStyle, bundle)
		},
	})
	core.AfterOverlays(ctx, state.endFrame)
	return inputDims
}

func (c Widget) layoutPanelOverlayWithBundle(ctx *core.Context, gtx, panelGtx layout.Context, comboState *comboBoxState, editor *widget.Editor, visible []int, anchor image.Rectangle, progress float32, interactive bool, instanceStyle style.Style, declarations comboDeclarationBundle) layout.Dimensions {
	if interactive {
		for comboState.dialog.Clicked(gtx) {
			core.RequestFocusVisible(ctx, editor, false)
		}
		if comboState.dialog.TakePressed() {
			core.RequestFocusVisible(ctx, editor, false)
		}
	}
	metrics := declarations.visual.Metrics
	viewport := gtx.Constraints.Max
	gap := gtx.Dp(metrics.PanelGap)
	panelWidth := min(max(anchor.Dx(), 0), max(viewport.X, 0))
	panelMaxY := min(gtx.Dp(metrics.PanelMaxHeight), max(viewport.Y-gap, 0))
	if panelWidth <= 0 || panelMaxY <= 0 {
		return layout.Dimensions{}
	}
	panelGtx.Constraints = layout.Constraints{
		Min: image.Pt(panelWidth, 0),
		Max: image.Pt(panelWidth, panelMaxY),
	}

	macro := op.Record(gtx.Ops)
	dims, tracked := core.TrackOverlayPlacement(ctx, func() layout.Dimensions {
		return c.layoutPanelWithBundle(ctx, panelGtx, comboState, editor, visible, instanceStyle, declarations)
	})
	call := macro.Stop()
	placement := core.Placement{Side: core.SideBottom, Align: core.AlignStart}
	result := core.ResolvePosition(core.PositionConfig{
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
	origin := core.PanelTransformOriginAt(anchor, result.Position, dims.Size, result.Placement)
	animationScale := metrics.AnimationScale
	if animationScale <= 0 || animationScale > 1 {
		animationScale = 0.95
	}
	scale := animationScale + (1-animationScale)*progress
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

func (c Widget) layoutPanelWithBundle(ctx *core.Context, gtx layout.Context, state *comboBoxState, editor *widget.Editor, visible []int, instanceStyle style.Style, declarations comboDeclarationBundle) layout.Dimensions {
	key := core.FullKey(ctx, c.key)
	defaults := declarations.visual.Defaults
	panel := computed.ResolvePart(ctx, gtx, key, theme.ComboPartOverlay,
		style.State{Disabled: !gtx.Enabled(), Open: true},
		defaults, style.Style{}, style.Style{}, instanceStyle)
	return layoutui.Resolved(ctx, gtx, panel, core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		if len(visible) == 0 {
			return c.layoutEmptyWithBundle(ctx, gtx, instanceStyle, declarations)
		}
		state.list.Axis = layout.Vertical
		return layoutui.TrackedScrollbarWithVisualOutset(ctx, gtx, &state.list, &state.bar, &state.visualOutset, len(visible), !gtx.Enabled(), false, func(gtx layout.Context, index int) layout.Dimensions {
			item := c.items[visible[index]]
			return c.layoutItemWithBundle(ctx, gtx, state, editor, item, item.Key == c.selectedKey, index == state.highlight, instanceStyle, declarations)
		})
	}))
}

func (c Widget) layoutEmptyWithBundle(ctx *core.Context, gtx layout.Context, instanceStyle style.Style, declarations comboDeclarationBundle) layout.Dimensions {
	metrics := declarations.visual.Metrics
	emptyText := c.emptyText
	if !c.hasEmptyText {
		emptyText = ctx.Text(i18n.ComboBoxNoResults)
	}
	gtx.Constraints.Min.Y = min(max(gtx.Constraints.Min.Y, gtx.Dp(metrics.ItemHeight)), gtx.Constraints.Max.Y)
	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		description := computed.ResolvePart(ctx, gtx, core.FullKey(ctx, c.key), theme.ComboPartDescription,
			style.State{Disabled: !gtx.Enabled()}, declarations.visual.Defaults, style.Style{}, style.Style{}, instanceStyle)
		return declarations.visual.Text(gtx, theme.ComboTextRequest{
			Text: emptyText, Style: description, Size: metrics.ItemTextSize,
		})
	})
}

func (c Widget) layoutItem(ctx *core.Context, gtx layout.Context, comboState *comboBoxState, editor *widget.Editor, item Item, selected, highlighted bool, overrides ...style.Style) layout.Dimensions {
	bundle := comboStyleDeclarationsCached(core.ActiveTheme(ctx), c.variant, &comboState.styleCache)
	return c.layoutItemWithBundle(ctx, gtx, comboState, editor, item, selected, highlighted, firstComboStyle(overrides), bundle)
}

func (c Widget) layoutItemWithBundle(ctx *core.Context, gtx layout.Context, comboState *comboBoxState, editor *widget.Editor, item Item, selected, highlighted bool, instanceStyle style.Style, declarations comboDeclarationBundle) layout.Dimensions {
	itemState := comboState.item(item.Key)
	presses := core.SnapshotPresses(itemState.Clickable.History())
	if !item.Disabled {
		for itemState.Clickable.Clicked(gtx) {
			c.selectItem(editor, comboState, item)
			core.RequestFocusVisible(ctx, editor, presses.ClickFocusVisible(itemState.Clickable.History()))
		}
		core.FocusOnPress(ctx, editor, itemState.Clickable.History(), presses.Active())
	}
	animGtx := gtx
	disabled := item.Disabled || !gtx.Enabled()
	focused := animGtx.Focused(&itemState.Clickable)
	focusVisible := core.FocusVisible(ctx, &itemState.Clickable, focused)
	styleState := style.State{
		Hovered:      !disabled && (itemState.Clickable.Hovered() || highlighted),
		Pressed:      !disabled && itemState.Clickable.Pressed(),
		Focused:      focused,
		FocusVisible: !disabled && focusVisible,
		Disabled:     disabled,
		Selected:     selected,
	}
	defaults := declarations.visual.Defaults
	itemKey := comboState.derivedItemKey(ctx, core.FullKey(ctx, c.key), item.Key)
	itemResolved := computed.ResolvePart(ctx, animGtx, itemKey, theme.ComboPartItem,
		styleState, defaults, style.Style{}, style.Style{}, instanceStyle)
	descriptionResolved := computed.ResolvePart(ctx, animGtx, itemKey, theme.ComboPartDescription,
		styleState, defaults, style.Style{}, style.Style{}, instanceStyle)
	iconResolved := computed.ResolvePart(ctx, animGtx, itemKey, theme.ComboPartIcon,
		styleState, defaults, style.Style{}, style.Style{}, instanceStyle)
	if item.Disabled {
		gtx = gtx.Disabled()
	}

	selection := itemState.Selection(animGtx, selected, declarations.visual.Metrics.ItemSelectDuration, declarations.visual.Motion)
	check := declarations.visual.Check
	content := core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		return c.layoutItemContent(ctx, gtx, item, itemResolved, descriptionResolved, iconResolved, selection, check, declarations.visual)
	})
	return layoutui.InteractiveResolved(ctx, gtx, itemResolved, content, func(gtx layout.Context, visual layout.Widget) layout.Dimensions {
		return itemState.Clickable.Layout(gtx, visual)
	})
}

func firstComboStyle(overrides []style.Style) style.Style {
	if len(overrides) == 0 {
		return style.Style{}
	}
	return overrides[0]
}

func comboBoxItemFrame(constraints layout.Constraints, minHeight int, content image.Point) (size, offset image.Point) {
	return optionrow.Frame(constraints, minHeight, content)
}

func (c Widget) layoutItemContent(
	ctx *core.Context,
	gtx layout.Context,
	item Item,
	itemResolved,
	descriptionResolved,
	iconResolved style.ResolvedStyle,
	selection float32,
	check theme.ComboCheckRenderer,
	visual theme.ComboVisual,
) layout.Dimensions {
	return layout.Flex{
		Axis:      layout.Horizontal,
		Alignment: layout.Middle,
	}.Layout(gtx,
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return visual.Text(gtx, theme.ComboTextRequest{Text: item.Label, Style: itemResolved})
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if item.Description == "" {
						return layout.Dimensions{}
					}
					return visual.Text(gtx, theme.ComboTextRequest{Text: item.Description, Style: descriptionResolved})
				}),
			)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			size := visual.IconSize(gtx, iconResolved)
			if check != nil {
				check(gtx, size, selection, iconResolved)
			}
			return layout.Dimensions{Size: size}
		}),
	)
}
