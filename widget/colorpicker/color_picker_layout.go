package colorpicker

import (
	"image"
	"image/color"
	"slices"

	"gioui.org/f32"
	"gioui.org/io/event"
	"gioui.org/io/pointer"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"github.com/qianniancn/flowui/i18n"
	"github.com/qianniancn/flowui/internal/core"
	flowlayout "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/input"
)

func (picker Widget) layoutTrigger(ctx *core.Context, gtx layout.Context, pickerState *colorPickerState, visual theme.ColorPickerVisual, enabled bool) layout.Dimensions {
	presses := core.SnapshotPresses(pickerState.trigger.History())
	if enabled {
		for pickerState.trigger.Clicked(gtx) {
			pickerState.open = !pickerState.open
			core.RequestFocusVisible(ctx, &pickerState.trigger, presses.ClickFocusVisible(pickerState.trigger.History()))
		}
		core.FocusOnPress(ctx, &pickerState.trigger, pickerState.trigger.History(), presses.Active())
	}

	return pickerState.trigger.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		semantic.Button.Add(gtx.Ops)
		semantic.LabelOp(picker.semanticLabel(ctx)).Add(gtx.Ops)
		semantic.EnabledOp(enabled).Add(gtx.Ops)
		macro := op.Record(gtx.Ops)
		dimensions := picker.layoutTriggerContent(ctx, gtx, visual)
		call := macro.Stop()
		if !enabled {
			opacity := paint.PushOpacity(gtx.Ops, visual.DisabledOpacity)
			call.Add(gtx.Ops)
			opacity.Pop()
		} else {
			call.Add(gtx.Ops)
		}
		focusVisible := core.FocusVisible(ctx, &pickerState.trigger, gtx.Focused(&pickerState.trigger))
		focus := pickerState.triggerFocus.Opacity(gtx, focusVisible && enabled, visual.Motion)
		visual.TriggerFocus(gtx, dimensions.Size, focus)
		return dimensions
	})
}

func (picker Widget) layoutTriggerContent(ctx *core.Context, gtx layout.Context, visual theme.ColorPickerVisual) layout.Dimensions {
	children := []layout.FlexChild{
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ColorSwatch(picker.value).Size(ColorSwatchLarge).Layout(ctx, gtx)
		}),
	}
	if picker.label != "" {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return visual.TriggerLabel(gtx, picker.label)
		}))
	}
	return layout.Flex{
		Axis:      layout.Horizontal,
		Alignment: layout.Middle,
		Gap:       gtx.Dp(visual.Metrics.TriggerGap),
	}.Layout(gtx, children...)
}

func (picker Widget) layoutPopover(ctx *core.Context, gtx layout.Context, pickerState *colorPickerState, visual theme.ColorPickerVisual, key string, trigger layout.Dimensions, progress float32, naturallyDisabled bool) {
	core.RegisterOverlay(ctx, core.OverlayRequest{
		Key:       key,
		Layer:     core.OverlayLayerPopup,
		Anchor:    image.Rectangle{Max: trigger.Size},
		HasAnchor: true,
		Disabled:  naturallyDisabled,
		Layout: func(gtx layout.Context, anchor image.Rectangle, interactive bool) layout.Dimensions {
			contentInteractive := interactive && pickerState.open && gtx.Enabled()
			if contentInteractive {
				pickerState.handleOverlayEvents(ctx, gtx, picker.historySize, picker.showHistory)
				contentInteractive = pickerState.open
			}
			if contentInteractive && pickerState.escapePressed(gtx) {
				if picker.showHistory {
					pickerState.pushHistory(pickerState.color.syncedColor, picker.historySize)
				}
				pickerState.open = false
				core.RequestFocus(ctx, &pickerState.trigger)
				contentInteractive = false
			}
			panelGtx := gtx
			// Keep the open panel enabled even when a nested overlay (such as
			// an input's context menu) owns policy input. Disabling it would
			// stop that overlay from being registered here, so it would
			// disappear on every frame it is not the event owner.
			if !pickerState.open || !gtx.Enabled() {
				panelGtx = panelGtx.Disabled()
			}
			return picker.layoutOverlay(ctx, gtx, panelGtx, pickerState, visual, anchor, progress)
		},
	})
}

func (picker Widget) layoutOverlay(ctx *core.Context, gtx, panelGtx layout.Context, pickerState *colorPickerState, visual theme.ColorPickerVisual, anchor image.Rectangle, progress float32) layout.Dimensions {
	viewport := gtx.Constraints.Max
	gap := gtx.Dp(visual.Metrics.PanelGap)
	panelWidth := min(gtx.Dp(visual.Metrics.PanelWidth), max(viewport.X, 0))
	panelHeight := min(gtx.Dp(visual.Metrics.PanelMaxHeight), max(viewport.Y-gap, 0))
	if panelWidth <= 0 || panelHeight <= 0 {
		return layout.Dimensions{}
	}
	panelGtx.Constraints = layout.Constraints{
		Min: image.Pt(panelWidth, 0),
		Max: image.Pt(panelWidth, panelHeight),
	}

	macro := op.Record(gtx.Ops)
	dimensions, tracked := core.TrackOverlayPlacement(ctx, func() layout.Dimensions {
		return picker.layoutPanel(ctx, panelGtx, pickerState, visual)
	})
	call := macro.Stop()
	placement := core.Placement{Side: core.SideBottom, Align: core.AlignStart}
	result := core.ResolvePosition(core.PositionConfig{
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
	origin := core.PanelTransformOriginAt(anchor, result.Position, dimensions.Size, result.Placement)
	scale := visual.Metrics.EnterScale + (1-visual.Metrics.EnterScale)*progress
	scaleTransform := f32.AffineId().Scale(origin, f32.Pt(scale, scale))
	panelTransform := f32.AffineId().Offset(f32.Pt(float32(result.Position.X), float32(result.Position.Y))).Mul(scaleTransform)
	tracked.PlaceTransform(panelTransform)
	tracked.SetOpacity(progress)
	picker.layoutDismissAreas(gtx, pickerState, viewport, anchor, core.AffineRectBounds(image.Rectangle{Max: dimensions.Size}, panelTransform))

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

func (picker Widget) layoutDismissAreas(gtx layout.Context, pickerState *colorPickerState, viewport image.Point, anchor, panel image.Rectangle) {
	index := 0
	for _, outsideAnchor := range core.DismissRects(image.Rectangle{Max: viewport}, anchor) {
		if outsideAnchor.Empty() {
			continue
		}
		for _, area := range core.DismissRects(outsideAnchor, panel) {
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

func (picker Widget) layoutPanel(ctx *core.Context, gtx layout.Context, pickerState *colorPickerState, visual theme.ColorPickerVisual) layout.Dimensions {
	restoreColors := core.PushColors(ctx, visual.PanelForeground, visual.PanelBackground)
	defer restoreColors()
	pickerState.panelList.Axis = layout.Vertical
	macro := op.Record(gtx.Ops)
	dimensions := pickerState.panelList.Layout(gtx, 1, func(gtx layout.Context, _ int) layout.Dimensions {
		return flowlayout.TrackedInset(ctx, gtx, layout.UniformInset(visual.Metrics.PanelPadding), func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Min.X = gtx.Constraints.Max.X
			return picker.layoutPanelContent(ctx, gtx, pickerState, visual)
		})
	})
	call := macro.Stop()
	visual.Panel(gtx, dimensions.Size)
	radius := min(max(gtx.Dp(visual.Metrics.PanelRadius), 1), min(dimensions.Size.X, dimensions.Size.Y)/2)
	clipped := clip.UniformRRect(image.Rectangle{Max: dimensions.Size}, radius).Push(gtx.Ops)
	call.Add(gtx.Ops)
	clipped.Pop()
	return dimensions
}

func (picker Widget) layoutPanelContent(ctx *core.Context, gtx layout.Context, pickerState *colorPickerState, visual theme.ColorPickerVisual) layout.Dimensions {
	restore := core.PushKey(ctx, picker.key)
	defer restore()
	// Continuous drag (area/slider) must not flood recent-history; discrete
	// commits (hex, RGB, presets, history swatches) record history.
	dragChange := picker.changeHandler(pickerState, false)
	commitChange := picker.changeHandler(pickerState, true)
	var children [8]layout.Widget
	count := 0
	children[count] = func(gtx layout.Context) layout.Dimensions {
		return ColorArea("area", picker.value).
			withColorState(&pickerState.color).
			OnChange(dragChange).
			Disabled(picker.disabled).
			Layout(ctx, gtx)
	}
	count++
	children[count] = func(gtx layout.Context) layout.Dimensions {
		return ColorSlider("hue", picker.value, ColorChannelHue).
			withColorState(&pickerState.color).
			OnChange(dragChange).
			Disabled(picker.disabled).
			Layout(ctx, gtx)
	}
	count++
	if picker.alpha {
		children[count] = func(gtx layout.Context) layout.Dimensions {
			return ColorSlider("alpha", picker.value, ColorChannelAlpha).
				withColorState(&pickerState.color).
				OnChange(dragChange).
				Disabled(picker.disabled).
				Layout(ctx, gtx)
		}
		count++
	}
	if picker.showRGB {
		children[count] = func(gtx layout.Context) layout.Dimensions {
			return picker.layoutRGBChannels(ctx, gtx, pickerState, visual, commitChange)
		}
		count++
	}
	if picker.showField {
		children[count] = func(gtx layout.Context) layout.Dimensions {
			return ColorField("field", picker.value).
				Alpha(picker.alpha).
				Swatch(true).
				Variant(input.Secondary).
				OnChange(commitChange).
				Disabled(picker.disabled).
				Layout(ctx, gtx)
		}
		count++
	}
	if len(picker.presets) > 0 {
		children[count] = func(gtx layout.Context) layout.Dimensions {
			return ColorSwatchPicker("presets", picker.value, picker.presets).
				Size(ColorSwatchExtraSmall).
				OnChange(commitChange).
				Disabled(picker.disabled).
				Layout(ctx, gtx)
		}
		count++
	}
	gap := visual.Metrics.ContentGap
	if picker.alpha || picker.showField || picker.showRGB || len(picker.presets) > 0 || picker.showHistory {
		gap = visual.Metrics.CompactContentGap
	}
	history := historyWithoutPresets(pickerState.history, picker.presets)
	if picker.showHistory && len(history) > 0 {
		children[count] = func(gtx layout.Context) layout.Dimensions {
			return ColorSwatchPicker("history", picker.value, history).
				Size(ColorSwatchExtraSmall).
				OnChange(commitChange).
				Disabled(picker.disabled).
				Layout(ctx, gtx)
		}
		count++
	}
	return flowlayout.Items(ctx, gtx, false, 0, gtx.Dp(gap), children[:count])
}

func historyWithoutPresets(history, presets []color.NRGBA) []color.NRGBA {
	if len(history) == 0 || len(presets) == 0 {
		return history
	}
	filtered := make([]color.NRGBA, 0, len(history))
	for _, recent := range history {
		preset := slices.Contains(presets, recent)
		if !preset {
			filtered = append(filtered, recent)
		}
	}
	return filtered
}

func (picker Widget) layoutRGBChannels(ctx *core.Context, gtx layout.Context, pickerState *colorPickerState, visual theme.ColorPickerVisual, onChange func(color.NRGBA)) layout.Dimensions {
	value := pickerState.color.syncedColor
	if !pickerState.color.ready {
		value = picker.value
	}
	type channelSpec struct {
		key   string
		label string
		get   func(color.NRGBA) byte
		set   func(color.NRGBA, byte) color.NRGBA
	}
	channels := []channelSpec{
		{"r", "R", func(c color.NRGBA) byte { return c.R }, func(c color.NRGBA, v byte) color.NRGBA { c.R = v; return c }},
		{"g", "G", func(c color.NRGBA) byte { return c.G }, func(c color.NRGBA, v byte) color.NRGBA { c.G = v; return c }},
		{"b", "B", func(c color.NRGBA) byte { return c.B }, func(c color.NRGBA, v byte) color.NRGBA { c.B = v; return c }},
	}
	if picker.alpha {
		channels = append(channels, channelSpec{
			"a", "A",
			func(c color.NRGBA) byte { return c.A },
			func(c color.NRGBA, v byte) color.NRGBA { c.A = v; return c },
		})
	}
	children := make([]layout.Widget, 0, len(channels))
	for _, channel := range channels {
		children = append(children, func(gtx layout.Context) layout.Dimensions {
			return picker.layoutChannelField(ctx, gtx, channel.key, channel.label, channel.get(value), func(next byte) {
				onChange(channel.set(value, next))
			})
		})
	}
	return layoutTrackedRow(ctx, gtx, gtx.Dp(visual.Metrics.ChannelGap), children)
}

// layoutTrackedRow lays out equal-width children in a row while recording each
// child's horizontal offset, so overlays anchored inside them (such as an
// input's context menu) resolve to the right viewport position.
func layoutTrackedRow(ctx *core.Context, gtx layout.Context, gap int, children []layout.Widget) layout.Dimensions {
	if len(children) == 0 {
		return layout.Dimensions{}
	}
	width := (gtx.Constraints.Max.X - gap*(len(children)-1)) / len(children)
	width = max(width, 0)
	type placed struct {
		call      op.CallOp
		dims      layout.Dimensions
		placement core.OverlayPlacement
		x         int
	}
	calls := make([]placed, 0, len(children))
	height := 0
	for index, child := range children {
		childGtx := gtx
		childGtx.Constraints.Min.X = width
		childGtx.Constraints.Max.X = width
		macro := op.Record(gtx.Ops)
		dims, placement := core.TrackOverlayPlacement(ctx, func() layout.Dimensions {
			return child(childGtx)
		})
		call := macro.Stop()
		calls = append(calls, placed{call: call, dims: dims, placement: placement, x: index * (width + gap)})
		height = max(height, dims.Size.Y)
	}
	for _, child := range calls {
		offset := image.Pt(child.x, height-child.dims.Size.Y)
		child.placement.PlaceOffset(offset)
		push := op.Offset(offset).Push(gtx.Ops)
		child.call.Add(gtx.Ops)
		push.Pop()
	}
	return layout.Dimensions{Size: image.Pt(width*len(children)+gap*(len(children)-1), height)}
}

func (picker Widget) layoutChannelField(ctx *core.Context, gtx layout.Context, key, label string, value byte, onChange func(byte)) layout.Dimensions {
	restore := core.PushKey(ctx, key)
	defer restore()
	control := input.New(itoaChannel(value)).Key("value").
		Placeholder("0").
		Label(label).
		Variant(input.Secondary).
		Disabled(picker.disabled).
		OnChange(func(text string) {
			if parsed, ok := parseChannel(text); ok {
				onChange(parsed)
			}
		})
	return control.Layout(ctx, gtx)
}

func itoaChannel(value byte) string {
	if value == 0 {
		return "0"
	}
	var buf [3]byte
	i := len(buf)
	for value > 0 {
		i--
		buf[i] = byte('0' + value%10)
		value /= 10
	}
	return string(buf[i:])
}

func parseChannel(text string) (byte, bool) {
	text = trimSpaceASCII(text)
	// Incomplete input (empty or still typing) must not commit a value, or
	// clearing the field while typing "128" would snap to 0 mid-edit.
	if text == "" {
		return 0, false
	}
	var n int
	for _, ch := range text {
		if ch < '0' || ch > '9' {
			return 0, false
		}
		n = n*10 + int(ch-'0')
		if n > 255 {
			return 0, false
		}
	}
	return byte(n), true
}

func trimSpaceASCII(value string) string {
	start, end := 0, len(value)
	for start < end && (value[start] == ' ' || value[start] == '\t') {
		start++
	}
	for end > start && (value[end-1] == ' ' || value[end-1] == '\t') {
		end--
	}
	return value[start:end]
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

func (picker Widget) semanticLabel(ctx *core.Context) string {
	if picker.label != "" {
		return picker.label
	}
	return ctx.Text(i18n.ColorPickerPickColor)
}
