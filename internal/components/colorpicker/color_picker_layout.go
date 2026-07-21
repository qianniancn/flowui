package colorpicker

import (
	"image"

	"gioui.org/f32"
	"gioui.org/io/event"
	"gioui.org/io/pointer"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"github.com/qianniancn/FlowUI/internal/components/input"
	"github.com/qianniancn/FlowUI/internal/components/text"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/locale"
	"github.com/qianniancn/FlowUI/internal/overlay"
	"github.com/qianniancn/FlowUI/internal/state"
)

func (picker ColorPickerWidget) layoutTrigger(ctx *frame.Context, gtx layout.Context, pickerState *colorPickerState, enabled bool) layout.Dimensions {
	presses := state.SnapshotPresses(pickerState.trigger.History())
	if enabled {
		for pickerState.trigger.Clicked(gtx) {
			pickerState.open = !pickerState.open
			frame.RequestFocusVisible(ctx, &pickerState.trigger, presses.ClickFocusVisible(pickerState.trigger.History()))
		}
		frame.FocusOnPress(ctx, &pickerState.trigger, pickerState.trigger.History(), presses.Active())
	}

	return pickerState.trigger.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		semantic.Button.Add(gtx.Ops)
		semantic.LabelOp(picker.semanticLabel(ctx)).Add(gtx.Ops)
		semantic.EnabledOp(enabled).Add(gtx.Ops)
		macro := op.Record(gtx.Ops)
		dimensions := picker.layoutTriggerContent(ctx, gtx)
		call := macro.Stop()
		if !enabled {
			opacity := paint.PushOpacity(gtx.Ops, frame.ActiveTheme(ctx).DisabledOpacityValue())
			call.Add(gtx.Ops)
			opacity.Pop()
		} else {
			call.Add(gtx.Ops)
		}
		focusVisible := frame.FocusVisible(ctx, &pickerState.trigger, gtx.Focused(&pickerState.trigger))
		focus := pickerState.triggerFocus.Opacity(gtx, focusVisible && enabled, frame.ActiveTheme(ctx).Motion)
		tokens := frame.ActiveTheme(ctx).Components.ColorPicker
		drawTriggerFocus(
			gtx,
			dimensions.Size,
			gtx.Dp(tokens.TriggerRadius),
			max(gtx.Dp(tokens.FocusRingWidth), 1),
			focus,
			frame.ActiveTheme(ctx).Palette.Focus,
		)
		return dimensions
	})
}

func (picker ColorPickerWidget) layoutTriggerContent(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	tokens := frame.ActiveTheme(ctx).Components.ColorPicker
	children := []layout.FlexChild{
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ColorSwatch(picker.value).Size(ColorSwatchLarge).Layout(ctx, gtx)
		}),
	}
	if picker.label != "" {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return text.New(picker.label).
				Size(float32(tokens.TriggerTextSize)).
				Color(frame.ActiveTheme(ctx).Palette.Foreground).
				Layout(ctx, gtx)
		}))
	}
	return layout.Flex{
		Axis:      layout.Horizontal,
		Alignment: layout.Middle,
		Gap:       gtx.Dp(tokens.TriggerGap),
	}.Layout(gtx, children...)
}

func (picker ColorPickerWidget) layoutPopover(ctx *frame.Context, gtx layout.Context, pickerState *colorPickerState, key string, trigger layout.Dimensions, progress float32, naturallyDisabled bool) {
	frame.RegisterOverlay(ctx, frame.OverlayRequest{
		Key:       key,
		Layer:     frame.OverlayLayerPopup,
		Anchor:    image.Rectangle{Max: trigger.Size},
		HasAnchor: true,
		Disabled:  naturallyDisabled,
		Layout: func(gtx layout.Context, anchor image.Rectangle, interactive bool) layout.Dimensions {
			contentInteractive := interactive && pickerState.open && gtx.Enabled()
			if contentInteractive {
				pickerState.handleOverlayEvents(ctx, gtx)
				contentInteractive = pickerState.open
			}
			if contentInteractive && pickerState.escapePressed(gtx) {
				pickerState.open = false
				frame.RequestFocus(ctx, &pickerState.trigger)
				contentInteractive = false
			}
			panelGtx := gtx
			if !contentInteractive {
				panelGtx = panelGtx.Disabled()
			}
			return picker.layoutOverlay(ctx, gtx, panelGtx, pickerState, anchor, progress)
		},
	})
}

func (picker ColorPickerWidget) layoutOverlay(ctx *frame.Context, gtx, panelGtx layout.Context, pickerState *colorPickerState, anchor image.Rectangle, progress float32) layout.Dimensions {
	tokens := frame.ActiveTheme(ctx).Components.ColorPicker
	viewport := gtx.Constraints.Max
	gap := gtx.Dp(tokens.PanelGap)
	panelWidth := min(gtx.Dp(tokens.PanelWidth), max(viewport.X, 0))
	panelHeight := min(gtx.Dp(tokens.PanelMaxHeight), max(viewport.Y-gap, 0))
	if panelWidth <= 0 || panelHeight <= 0 {
		return layout.Dimensions{}
	}
	panelGtx.Constraints = layout.Constraints{
		Min: image.Pt(panelWidth, 0),
		Max: image.Pt(panelWidth, panelHeight),
	}

	macro := op.Record(gtx.Ops)
	dimensions, tracked := frame.TrackOverlayPlacement(ctx, func() layout.Dimensions {
		return picker.layoutPanel(ctx, panelGtx, pickerState)
	})
	call := macro.Stop()
	placement := overlay.Placement{Side: overlay.SideBottom, Align: overlay.AlignStart}
	result := overlay.ResolvePosition(overlay.PositionConfig{
		Trigger:          anchor.Size(),
		TriggerOrigin:    anchor.Min,
		HasTriggerOrigin: true,
		Panel:            dimensions.Size,
		Bounds:           viewport,
		Offset:           gap,
		Placement:        placement,
		Flip:             true,
		AvoidOverflow:    true,
	})
	origin := overlay.PanelTransformOriginAt(anchor, result.Position, dimensions.Size, result.Placement)
	scale := float32(.95) + .05*progress
	scaleTransform := f32.AffineId().Scale(origin, f32.Pt(scale, scale))
	panelTransform := f32.AffineId().Offset(f32.Pt(float32(result.Position.X), float32(result.Position.Y))).Mul(scaleTransform)
	tracked.PlaceTransform(panelTransform)
	tracked.SetOpacity(progress)
	picker.layoutDismissAreas(gtx, pickerState, viewport, anchor, overlay.AffineRectBounds(image.Rectangle{Max: dimensions.Size}, panelTransform))

	offset := op.Offset(result.Position).Push(gtx.Ops)
	opacity := paint.PushOpacity(gtx.Ops, progress)
	transform := op.Affine(scaleTransform).Push(gtx.Ops)
	layoutColorPickerPanelBlocker(gtx, pickerState, dimensions.Size)
	call.Add(gtx.Ops)
	transform.Pop()
	opacity.Pop()
	offset.Pop()
	return dimensions
}

func (picker ColorPickerWidget) layoutDismissAreas(gtx layout.Context, pickerState *colorPickerState, viewport image.Point, anchor, panel image.Rectangle) {
	index := 0
	for _, outsideAnchor := range overlay.DismissRects(image.Rectangle{Max: viewport}, anchor) {
		if outsideAnchor.Empty() {
			continue
		}
		for _, area := range overlay.DismissRects(outsideAnchor, panel) {
			if area.Empty() || index >= len(pickerState.dismiss) {
				continue
			}
			areaGtx := gtx
			areaGtx.Constraints = layout.Exact(area.Size())
			offset := op.Offset(area.Min).Push(gtx.Ops)
			pass := pointer.PassOp{}.Push(gtx.Ops)
			pickerState.dismiss[index].Layout(areaGtx, func(layout.Context) layout.Dimensions {
				return layout.Dimensions{Size: area.Size()}
			})
			pass.Pop()
			offset.Pop()
			index++
		}
	}
}

func (picker ColorPickerWidget) layoutPanel(ctx *frame.Context, gtx layout.Context, pickerState *colorPickerState) layout.Dimensions {
	activeTheme := frame.ActiveTheme(ctx)
	restoreColors := frame.PushColors(ctx, activeTheme.Palette.OverlayForegroundColor(), activeTheme.Palette.OverlayColor())
	defer restoreColors()
	pickerState.panelList.Axis = layout.Vertical
	macro := op.Record(gtx.Ops)
	dimensions := pickerState.panelList.Layout(gtx, 1, func(gtx layout.Context, _ int) layout.Dimensions {
		padding := frame.ActiveTheme(ctx).Components.ColorPicker.PanelPadding
		return layout.UniformInset(padding).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Min.X = gtx.Constraints.Max.X
			return picker.layoutPanelContent(ctx, gtx, pickerState)
		})
	})
	call := macro.Stop()
	radius := min(max(gtx.Dp(frame.ActiveTheme(ctx).Components.ColorPicker.PanelRadius), 1), min(dimensions.Size.X, dimensions.Size.Y)/2)
	drawColorPickerPanel(gtx, frame.ActiveTheme(ctx), dimensions.Size, radius)
	clipped := clip.UniformRRect(image.Rectangle{Max: dimensions.Size}, radius).Push(gtx.Ops)
	call.Add(gtx.Ops)
	clipped.Pop()
	return dimensions
}

func (picker ColorPickerWidget) layoutPanelContent(ctx *frame.Context, gtx layout.Context, pickerState *colorPickerState) layout.Dimensions {
	tokens := frame.ActiveTheme(ctx).Components.ColorPicker
	restore := frame.PushKey(ctx, picker.key)
	defer restore()
	var children [5]layout.FlexChild
	count := 0
	children[count] = layout.Rigid(func(gtx layout.Context) layout.Dimensions {
		return ColorArea("area", picker.value).
			withColorState(&pickerState.color).
			OnChange(picker.onChange).
			Disabled(picker.disabled).
			Layout(ctx, gtx)
	})
	count++
	children[count] = layout.Rigid(func(gtx layout.Context) layout.Dimensions {
		return ColorSlider("hue", picker.value, ColorChannelHue).
			withColorState(&pickerState.color).
			OnChange(picker.onChange).
			Disabled(picker.disabled).
			Layout(ctx, gtx)
	})
	count++
	if picker.alpha {
		children[count] = layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ColorSlider("alpha", picker.value, ColorChannelAlpha).
				withColorState(&pickerState.color).
				OnChange(picker.onChange).
				Disabled(picker.disabled).
				Layout(ctx, gtx)
		})
		count++
	}
	if len(picker.presets) > 0 {
		children[count] = layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ColorSwatchPicker("presets", picker.value, picker.presets).
				Size(ColorSwatchExtraSmall).
				OnChange(picker.onChange).
				Disabled(picker.disabled).
				Layout(ctx, gtx)
		})
		count++
	}
	if picker.showField {
		children[count] = layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ColorField("field", picker.value).
				Alpha(picker.alpha).
				Swatch(true).
				Variant(input.InputSecondary).
				OnChange(picker.onChange).
				Disabled(picker.disabled).
				Layout(ctx, gtx)
		})
		count++
	}
	gap := tokens.ContentGap
	if picker.alpha || picker.showField || len(picker.presets) > 0 {
		gap = tokens.CompactContentGap
	}
	return layout.Flex{Axis: layout.Vertical, Gap: gtx.Dp(gap)}.Layout(gtx, children[:count]...)
}

func addColorControlInput(gtx layout.Context, tag event.Tag, size image.Point, enabled, area bool, label, description string) {
	clipped := clip.Rect{Max: size}.Push(gtx.Ops)
	semantic.LabelOp(label).Add(gtx.Ops)
	semantic.DescriptionOp(description).Add(gtx.Ops)
	semantic.EnabledOp(enabled).Add(gtx.Ops)
	if enabled {
		if area {
			pointer.CursorCrosshair.Add(gtx.Ops)
		} else {
			pointer.CursorPointer.Add(gtx.Ops)
		}
		event.Op(gtx.Ops, tag)
	}
	clipped.Pop()
}

func layoutColorPickerPanelBlocker(gtx layout.Context, pickerState *colorPickerState, size image.Point) {
	if size.X <= 0 || size.Y <= 0 {
		return
	}
	blockerGtx := gtx
	blockerGtx.Constraints = layout.Exact(size)
	pickerState.dialog.Layout(blockerGtx, func(layout.Context) layout.Dimensions {
		return layout.Dimensions{Size: size}
	})
}

func (picker ColorPickerWidget) semanticLabel(ctx *frame.Context) string {
	if picker.label != "" {
		return picker.label
	}
	if frame.ActiveLanguage(ctx) == locale.LanguageChinese {
		return "选择颜色"
	}
	return "Pick a color"
}
