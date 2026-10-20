package datepicker

import (
	"fmt"
	"image"
	"image/color"
	"time"

	"gioui.org/f32"
	"gioui.org/io/pointer"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"github.com/qianniancn/flowui/draw"
	"github.com/qianniancn/flowui/i18n"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/field"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/description"
	"github.com/qianniancn/flowui/widget/label"
	"github.com/qianniancn/flowui/widget/text"
)

func (d Widget) resolveDatePickerPart(ctx *core.Context, gtx layout.Context, part style.Part, state style.State, calendar *style.Style) style.ResolvedStyle {
	return resolveDatePickerPartFor(ctx, gtx, core.FullKey(ctx, d.key), part, state, core.InstanceStyle(ctx), style.Style{}, calendar)
}

func (d Widget) resolveDatePickerPartWithDefaults(ctx *core.Context, gtx layout.Context, part style.Part, state style.State, defaults style.Style) style.ResolvedStyle {
	return computed.ResolvePart(ctx, gtx, core.FullKey(ctx, d.key), part, state, defaults, style.Style{}, style.Style{}, core.InstanceStyle(ctx))
}

func datePickerBaseStyle(calendar *style.Style) style.Style {
	if calendar == nil {
		panic(theme.ErrIncompleteVisualProvider)
	}
	return *calendar
}

func resolveDatePickerVisual(activeTheme *theme.Theme, variant field.Variant, override *theme.DatePickerOverride) theme.DatePickerVisual {
	return theme.ResolveVisual(activeTheme, theme.DatePickerRequest{
		Variant:  theme.InputVariant(variant),
		Override: override,
	}, theme.ComponentVisualProvider.DatePicker, theme.ValidateDatePickerVisual)
}

func datePickerVisualDeclarationForVariant(activeTheme *theme.Theme, variant field.Variant) theme.DatePickerVisual {
	return resolveDatePickerVisual(activeTheme, variant, nil)
}

func resolveDatePickerPartFor(ctx *core.Context, gtx layout.Context, key string, part style.Part, state style.State, custom, declaration style.Style, calendar *style.Style) style.ResolvedStyle {
	base := datePickerBaseStyle(calendar)
	// Merge only the requested part. Joining the whole declaration deep-clones
	// every calendar part for each cell before the resolver sees this declaration.
	defaults := style.JoinPart(base, declaration, part)
	return computed.ResolvePart(ctx, gtx, key, part, state, defaults, style.Style{}, style.Style{}, custom)
}

func (d Widget) datePickerDescriptionColor(ctx *core.Context, gtx layout.Context, invalid, enabled bool, visual theme.DatePickerVisual, calendar *style.Style) color.NRGBA {
	state := style.State{Disabled: !enabled, Invalid: invalid}
	resolved := d.resolveDatePickerPart(ctx, gtx, style.PartDescription, state, calendar)
	return datePickerTextColor(visual, resolved)
}

func (d Widget) layoutField(ctx *core.Context, gtx layout.Context, pickerState *datePickerState, fieldStyle field.Resolved, metrics theme.DatePickerMetrics, enabled, invalid bool) (layout.Dimensions, image.Rectangle) {
	var children [3]layout.FlexChild
	count := 0
	labelHeight := 0
	inputAnchor := image.Rectangle{}
	gap := gtx.Dp(metrics.FieldGap)
	descriptionStyle := d.resolveDatePickerPart(ctx, gtx, style.PartDescription, style.State{
		Disabled: !enabled,
		Invalid:  invalid,
	}, &pickerState.calendarStyle)
	if d.label != "" {
		children[count] = layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			dimensions := label.New(d.label).
				For(d.key).
				Required(d.required).
				Disabled(!enabled).
				Invalid(invalid).
				Layout(ctx, gtx)
			labelHeight = dimensions.Size.Y
			return dimensions
		})
		count++
	}
	children[count] = layout.Rigid(func(gtx layout.Context) layout.Dimensions {
		dimensions := d.layoutInput(ctx, gtx, pickerState, fieldStyle, metrics, enabled, invalid)
		y := labelHeight
		if d.label != "" {
			y += gap
		}
		inputAnchor = image.Rectangle{Min: image.Pt(0, y), Max: image.Pt(dimensions.Size.X, y+dimensions.Size.Y)}
		addDateInputHover(gtx, &pickerState.hover, dimensions.Size, enabled, false)
		return dimensions
	})
	count++
	if invalid && d.errorMessage != "" {
		children[count] = layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return text.New(d.errorMessage).
				Style(style.TextDeclaration(descriptionStyle.Text)).
				Layout(ctx, gtx)
		})
		count++
	} else if d.description != "" {
		children[count] = layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return (description.New(d.description).
				For(d.key).
				Disabled(!enabled)).Style(style.TextDeclaration(descriptionStyle.Text)).Layout(ctx, gtx)
		})
		count++
	}
	dimensions := layout.Flex{Axis: layout.Vertical, Gap: gap}.Layout(gtx, children[:count]...)
	return dimensions, inputAnchor
}

func (d Widget) layoutInput(ctx *core.Context, gtx layout.Context, pickerState *datePickerState, fieldStyle field.Resolved, metrics theme.DatePickerMetrics, enabled, invalid bool) layout.Dimensions {
	return layoutui.Resolved(ctx, gtx, fieldStyle.Content, core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		return d.layoutInputContent(ctx, gtx, pickerState, fieldStyle, metrics, enabled, invalid)
	}))
}

func (d Widget) layoutInputContent(ctx *core.Context, gtx layout.Context, pickerState *datePickerState, fieldStyle field.Resolved, metrics theme.DatePickerMetrics, enabled, invalid bool) layout.Dimensions {
	colors := fieldStyle.Colors
	visual := pickerState.partDefaults.currentVisual(core.ActiveTheme(ctx))
	presses := core.ActivePresses(pickerState.trigger.History())
	focusVisible := core.FocusVisible(ctx, &pickerState.trigger, gtx.Focused(&pickerState.trigger))
	if enabled {
		for pickerState.trigger.Clicked(gtx) {
			if pickerState.open {
				pickerState.open = false
			} else {
				pickerState.openCalendar()
			}
			core.RequestFocusVisible(ctx, &pickerState.trigger, focusVisible)
		}
		core.FocusOnPress(ctx, &pickerState.trigger, pickerState.trigger.History(), presses)
	}

	frameConstraints := gtx.Constraints
	if d.fullWidth {
		frameConstraints.Min.X = frameConstraints.Max.X
	}
	height := min(gtx.Dp(metrics.Height), frameConstraints.Max.Y)
	frameConstraints.Min.Y = min(max(frameConstraints.Min.Y, height), frameConstraints.Max.Y)

	left := gtx.Dp(metrics.FieldPaddingX)
	right := gtx.Dp(metrics.TriggerWidth)
	horizontalPadding := left + right
	maxX := max(frameConstraints.Max.X-horizontalPadding, 0)
	minX := min(max(frameConstraints.Min.X-horizontalPadding, 0), maxX)

	macro := op.Record(gtx.Ops)
	contentGtx := gtx
	contentGtx.Constraints = layout.Constraints{
		Min: image.Pt(minX, 0),
		Max: image.Pt(maxX, height),
	}
	var contentDims layout.Dimensions
	showHint := false
	if d.editable {
		contentDims = pickerState.segments.layout(ctx, contentGtx, d.locale, dateSegmentsLayoutOptions{
			key:           core.FullKey(ctx, d.key),
			instanceStyle: core.InstanceStyle(ctx),
			calendarStyle: &pickerState.calendarStyle,
			resolvedParts: &pickerState.resolvedParts,
			colors:        colors,
			metrics:       metrics,
			visual:        visual,
			enabled:       enabled,
			invalid:       invalid,
			minDate:       d.minDate,
			maxDate:       d.maxDate,
		}, func(value time.Time) {
			pickerState.requestValue(d, value)
		})
		showHint = d.hintSet && pickerState.segments.empty() && !pickerState.segments.focused(gtx)
	} else {
		value := d.hint
		color := colors.Placeholder
		if !d.value.IsZero() {
			value = d.locale.DateLabel(d.value)
			color = colors.Foreground
		}
		contentGtx.Constraints.Min = image.Point{}
		contentDims = text.New(value).
			Size(float32(metrics.TextSize)).
			Color(color).
			MaxLines(1).
			Truncator("...").
			Layout(ctx, contentGtx)
	}
	call := macro.Stop()
	var hintCall op.CallOp
	var hintDims layout.Dimensions
	if showHint {
		hintMacro := op.Record(gtx.Ops)
		hintColor := colors.Placeholder
		if invalid {
			hintColor = d.datePickerDescriptionColor(ctx, gtx, true, enabled, visual, &pickerState.calendarStyle)
		}
		hintGtx := contentGtx
		hintGtx.Constraints.Min = image.Point{}
		hintDims = text.New(d.hint).
			Size(float32(metrics.TextSize)).
			Color(hintColor).
			MaxLines(1).
			Truncator("...").
			Layout(ctx, hintGtx)
		hintCall = hintMacro.Stop()
		contentDims.Size.X = max(contentDims.Size.X, hintDims.Size.X)
		contentDims.Size.Y = max(contentDims.Size.Y, hintDims.Size.Y)
	}

	size := frameConstraints.Constrain(image.Pt(contentDims.Size.X+horizontalPadding, height))
	stack := op.Offset(image.Pt(left, max((size.Y-contentDims.Size.Y)/2, 0))).Push(gtx.Ops)
	contentClip := clip.Rect{Max: image.Pt(max(size.X-horizontalPadding, 0), contentDims.Size.Y)}.Push(gtx.Ops)
	if !showHint {
		call.Add(gtx.Ops)
	}
	if showHint {
		hintOffset := op.Offset(image.Pt(0, max((contentDims.Size.Y-hintDims.Size.Y)/2, 0))).Push(gtx.Ops)
		hintCall.Add(gtx.Ops)
		hintOffset.Pop()
	}
	contentClip.Pop()
	stack.Pop()

	triggerX := size.X - right
	triggerSize := image.Pt(right, size.Y)
	if !d.editable {
		triggerX = 0
		triggerSize.X = size.X
	}
	triggerGtx := gtx
	triggerGtx.Constraints = layout.Exact(triggerSize)
	stack = op.Offset(image.Pt(triggerX, 0)).Push(gtx.Ops)
	if !enabled {
		triggerGtx = triggerGtx.Disabled()
	}
	pickerState.trigger.Layout(triggerGtx, func(gtx layout.Context) layout.Dimensions {
		if enabled {
			pointer.CursorPointer.Add(gtx.Ops)
		}
		semantic.Button.Add(gtx.Ops)
		semantic.LabelOp(d.calendarLabel(ctx)).Add(gtx.Ops)
		semantic.EnabledOp(enabled).Add(gtx.Ops)
		iconSize := image.Pt(gtx.Dp(metrics.IconSize), gtx.Dp(metrics.IconSize))
		iconOffset := op.Offset(image.Pt(triggerSize.X-right+(right-iconSize.X)/2, (size.Y-iconSize.Y)/2)).Push(gtx.Ops)
		visual.CalendarIcon(gtx, theme.DatePickerIconRenderRequest{
			Size: iconSize, Color: colors.Placeholder, Metrics: metrics,
		})
		iconOffset.Pop()
		if d.editable {
			focusColor := datePickerOutlineColor(visual, fieldStyle.Content)
			visual.Focus(gtx, theme.DatePickerFocusRenderRequest{
				Size: triggerSize, Color: focusColor,
				Visible: enabled && core.FocusVisible(ctx, &pickerState.trigger, gtx.Focused(&pickerState.trigger)),
				Trigger: true, FocusWidth: metrics.FocusRingWidth,
				IconSize: metrics.IconSize, SegmentRadius: metrics.SegmentRadius, Metrics: metrics,
			})
		}
		return layout.Dimensions{Size: triggerSize}
	})
	stack.Pop()
	return layout.Dimensions{Size: size}
}

func (d Widget) calendarLabel(ctx *core.Context) string {
	return ctx.Text(i18n.DatePickerChooseDate)
}

func (d Widget) layoutPopover(ctx *core.Context, gtx layout.Context, state *datePickerState, inputAnchor image.Rectangle, progress float32, now time.Time, naturallyDisabled bool) {
	metrics := state.partDefaults.currentVisual(core.ActiveTheme(ctx)).Metrics
	core.RegisterOverlay(ctx, core.OverlayRequest{
		Key:       core.FullKey(ctx, d.key),
		Layer:     core.OverlayLayerPopup,
		Anchor:    inputAnchor,
		HasAnchor: true,
		Disabled:  naturallyDisabled,
		Layout: func(gtx layout.Context, anchor image.Rectangle, interactive bool) layout.Dimensions {
			contentInteractive := interactive && state.open && gtx.Enabled()
			if contentInteractive {
				state.updateKeys(gtx, &state.trigger)
			}
			panelGtx := gtx
			// Keep an open panel enabled even when a nested overlay owns
			// policy input; disabling it would unregister that overlay.
			if !state.open || !gtx.Enabled() {
				panelGtx = panelGtx.Disabled()
			}
			return d.layoutCalendarOverlay(ctx, gtx, panelGtx, state, metrics, anchor, progress, now, contentInteractive)
		},
	})
}

func (d Widget) layoutCalendarOverlay(ctx *core.Context, gtx, panelGtx layout.Context, pickerState *datePickerState, metrics theme.DatePickerMetrics, anchor image.Rectangle, progress float32, now time.Time, interactive bool) layout.Dimensions {
	if interactive {
		for pickerState.dialog.Clicked(gtx) {
			core.RequestFocusVisible(ctx, &pickerState.trigger, false)
		}
		if pickerState.dialog.TakePressed() {
			core.RequestFocusVisible(ctx, &pickerState.trigger, false)
		}
	}
	viewport := gtx.Constraints.Max
	gap := gtx.Dp(metrics.PopoverGap)
	panelResolved := d.resolveDatePickerPart(ctx, gtx, PartCalendarPanel, style.State{
		Disabled: !gtx.Enabled(),
		Open:     pickerState.open,
	}, &pickerState.calendarStyle)
	panelWidth := min(gtx.Dp(metrics.CalendarWidth)+datePickerHorizontalPadding(gtx, panelResolved), max(viewport.X, 0))
	panelMaxY := min(gtx.Dp(metrics.PopoverMaxHeight), max(viewport.Y-gap, 0))
	if panelWidth <= 0 || panelMaxY <= 0 {
		return layout.Dimensions{}
	}
	panelGtx.Constraints = layout.Constraints{
		Min: image.Pt(panelWidth, 0),
		Max: image.Pt(panelWidth, panelMaxY),
	}

	macro := op.Record(gtx.Ops)
	dims, tracked := core.TrackOverlayPlacement(ctx, func() layout.Dimensions {
		resolved := panelResolved
		// A disabled overlay has a different style state. Preserve the old
		// disabled resolution only for that uncommon path; the interactive path
		// reuses the panel declaration already needed for width measurement.
		if panelGtx.Enabled() != gtx.Enabled() {
			resolved = d.resolveDatePickerPart(ctx, panelGtx, PartCalendarPanel, style.State{
				Disabled: !panelGtx.Enabled(),
				Open:     pickerState.open,
			}, &pickerState.calendarStyle)
		}
		return d.layoutCalendarPanel(ctx, panelGtx, pickerState, metrics, now, resolved)
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
	motion := pickerState.partDefaults.motion(core.ActiveTheme(ctx))
	popoverScale := motion.PopoverScale
	if popoverScale <= 0 || popoverScale > 1 {
		popoverScale = 1
	}
	scale := popoverScale + (1-popoverScale)*progress
	scaleTransform := f32.AffineId().Scale(origin, f32.Pt(scale, scale))
	tracked.PlaceTransform(f32.AffineId().Offset(f32.Pt(float32(result.Position.X), float32(result.Position.Y))).Mul(scaleTransform))
	tracked.SetOpacity(progress)

	stack := op.Offset(result.Position).Push(gtx.Ops)
	opacity := paint.PushOpacity(gtx.Ops, progress)
	transform := op.Affine(scaleTransform).Push(gtx.Ops)
	layoutDatePickerPanelBlocker(gtx, pickerState, dims.Size)
	call.Add(gtx.Ops)
	transform.Pop()
	opacity.Pop()
	stack.Pop()
	return dims
}

func layoutDatePickerPanelBlocker(gtx layout.Context, state *datePickerState, size image.Point) {
	if size.X <= 0 || size.Y <= 0 {
		return
	}
	blockerGtx := gtx
	blockerGtx.Constraints = layout.Exact(size)
	state.dialog.Layout(blockerGtx, func(layout.Context) layout.Dimensions {
		return layout.Dimensions{Size: size}
	})
}

func (d Widget) layoutCalendarPanel(ctx *core.Context, gtx layout.Context, state *datePickerState, metrics theme.DatePickerMetrics, now time.Time, resolved style.ResolvedStyle) layout.Dimensions {
	return layoutui.Resolved(ctx, gtx, resolved, core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		return d.layoutCalendar(ctx, gtx, state, metrics, now)
	}))
}

func datePickerHorizontalPadding(gtx layout.Context, resolved style.ResolvedStyle) int {
	if resolved.Box == nil || resolved.Box.Padding == nil {
		panic(theme.ErrIncompleteVisualProvider)
	}
	return gtx.Dp(resolved.Box.Padding.Left) + gtx.Dp(resolved.Box.Padding.Right)
}

func datePickerPaintRadius(gtx layout.Context, resolved style.ResolvedStyle) int {
	if resolved.Paint == nil || resolved.Paint.Radius == nil {
		panic(theme.ErrIncompleteVisualProvider)
	}
	return max(gtx.Dp(*resolved.Paint.Radius), 0)
}

func (d Widget) layoutCalendar(ctx *core.Context, gtx layout.Context, state *datePickerState, metrics theme.DatePickerMetrics, now time.Time) layout.Dimensions {
	gtx.Constraints.Min.X = min(gtx.Dp(metrics.CalendarWidth), gtx.Constraints.Max.X)
	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return d.layoutCalendarHeader(ctx, gtx, state, metrics)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layoutui.Divider().Layout(ctx, gtx)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layoutui.Spacer(0, 8).Layout(ctx, gtx)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			switch state.viewMode {
			case datePickerViewMonths:
				return d.layoutMonths(ctx, gtx, state, metrics, now)
			case datePickerViewYears:
				return d.layoutYears(ctx, gtx, state, metrics, now)
			default:
				return d.layoutDayView(ctx, gtx, state, metrics, now)
			}
		}),
	)
}

func (d Widget) layoutCalendarHeader(ctx *core.Context, gtx layout.Context, pickerState *datePickerState, metrics theme.DatePickerMetrics) layout.Dimensions {
	return layout.Inset{
		Left:   2,
		Right:  2,
		Bottom: 10,
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{
			Axis:      layout.Horizontal,
			Alignment: layout.Middle,
		}.Layout(gtx,
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				presses := core.ActivePresses(pickerState.header.History())
				focusVisible := core.FocusVisible(ctx, &pickerState.header, gtx.Focused(&pickerState.header))
				for pickerState.header.Clicked(gtx) {
					pickerState.toggleYearPicker(pickerState.viewMonth.Year())
					core.RequestFocusVisible(ctx, &pickerState.trigger, focusVisible)
				}
				core.FocusOnPress(ctx, &pickerState.trigger, pickerState.header.History(), presses)
				return pickerState.header.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					gtx.Constraints.Min.X = gtx.Constraints.Max.X
					return layout.W.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return d.layoutHeaderTrigger(ctx, gtx, pickerState, metrics)
					})
				})
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return d.layoutNavButton(ctx, gtx, pickerState, metrics, -1)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return d.layoutNavButton(ctx, gtx, pickerState, metrics, 1)
			}),
		)
	})
}

func (d Widget) layoutHeaderTrigger(ctx *core.Context, gtx layout.Context, state *datePickerState, metrics theme.DatePickerMetrics) layout.Dimensions {
	visual := state.partDefaults.currentVisual(core.ActiveTheme(ctx))
	label := d.headerLabel(state)
	focused := gtx.Focused(&state.header)
	headerState := style.State{
		Focused:      focused,
		FocusVisible: core.FocusVisible(ctx, &state.header, focused),
		Selected:     state.viewMode == datePickerViewYears,
	}
	headerKey := datePickerStyleStateKey(headerState)
	resolved, cached := state.resolvedParts.headerValue(headerKey)
	if !cached {
		resolved = d.resolveDatePickerPart(ctx, gtx, PartCalendarHeader, headerState, &state.calendarStyle)
		state.resolvedParts.storeHeader(headerKey, resolved)
	}
	gap := gtx.Dp(metrics.HeaderGap)
	iconSize := gtx.Dp(metrics.HeaderIconSize)
	return layout.Flex{
		Axis:      layout.Horizontal,
		Alignment: layout.Middle,
		Gap:       gap,
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			labelWidget := text.WithResolvedText(text.New(label), resolved.Text)
			return labelWidget.Layout(ctx, gtx)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			size := image.Pt(iconSize, iconSize)
			gtx.Constraints = layout.Exact(size)
			color := datePickerTextColor(visual, resolved)
			visual.YearIndicator(gtx, theme.DatePickerYearIndicatorRenderRequest{
				Size: size, Open: state.viewMode == datePickerViewYears, Color: color, Metrics: metrics,
			})
			return layout.Dimensions{Size: size}
		}),
	)
}

func (d Widget) headerLabel(state *datePickerState) string {
	year, _, _ := state.viewMonth.Date()
	switch state.viewMode {
	case datePickerViewMonths:
		return d.locale.YearLabel(year)
	case datePickerViewYears:
		return d.locale.MonthLabel(state.viewMonth)
	default:
		return d.locale.MonthLabel(state.viewMonth)
	}
}

func (d Widget) layoutNavButton(ctx *core.Context, gtx layout.Context, pickerState *datePickerState, metrics theme.DatePickerMetrics, delta int) layout.Dimensions {
	visual := pickerState.partDefaults.currentVisual(core.ActiveTheme(ctx))
	size := image.Pt(gtx.Dp(metrics.NavButtonSize), gtx.Dp(metrics.NavButtonSize))
	if pickerState.viewMode == datePickerViewYears {
		return layout.Dimensions{Size: size}
	}

	clickable := &pickerState.prev
	if delta > 0 {
		clickable = &pickerState.next
	}
	disabled := !d.canMove(pickerState, delta)
	focusVisible := core.FocusVisible(ctx, clickable, gtx.Focused(clickable))
	presses := core.ActivePresses(clickable.History())
	if disabled {
		gtx = gtx.Disabled()
	} else {
		for clickable.Clicked(gtx) {
			pickerState.move(delta)
			core.RequestFocusVisible(ctx, &pickerState.trigger, focusVisible)
		}
		core.FocusOnPress(ctx, &pickerState.trigger, clickable.History(), presses)
	}
	gtx.Constraints = layout.Exact(size)
	return clickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		focused := gtx.Focused(clickable)
		state := style.State{
			Hovered:      clickable.Hovered() && !disabled,
			Pressed:      clickable.Pressed() && !disabled,
			Focused:      focused,
			FocusVisible: focusVisible && !disabled,
			Disabled:     disabled,
		}
		navKey := datePickerStyleStateKey(state)
		resolved, cached := pickerState.resolvedParts.navValue(navKey)
		if !cached {
			resolved = d.resolveDatePickerPart(ctx, gtx, PartCalendarNav, state, &pickerState.calendarStyle)
			pickerState.resolvedParts.storeNav(navKey, resolved)
		}
		foreground := datePickerTextColor(visual, resolved)
		background := datePickerPaintColor(visual, resolved)
		radius := datePickerPaintRadius(gtx, resolved)
		focusKey := datePickerStyleStateKey(state)
		controlFocus, cached := pickerState.resolvedParts.focusValue(focusKey)
		if !cached {
			controlFocus = d.resolveDatePickerPart(ctx, gtx, PartCalendarFocus, state, &pickerState.calendarStyle)
			pickerState.resolvedParts.storeFocus(focusKey, controlFocus)
		}
		focusColor := datePickerOutlineColor(visual, controlFocus)
		navVisual := theme.DatePickerNavButtonRenderRequest{
			Size: size, Direction: delta, Background: background, Foreground: foreground,
			Radius: radius, Metrics: metrics,
		}
		motion := pickerState.partDefaults.motion(core.ActiveTheme(ctx))
		navVisual.Background = pickerState.navBackground(gtx, delta, navVisual.Background, motion)
		scale := datePickerPressScale(gtx, clickable.History(), disabled, motion)
		stack := draw.Scale(size, scale).Push(gtx.Ops)
		visual.NavButton(gtx, navVisual)
		stack.Pop()
		visual.Focus(gtx, theme.DatePickerFocusRenderRequest{
			Size: size, Color: focusColor, Radius: radius,
			Visible: focusVisible && !disabled, FocusWidth: metrics.FocusRingWidth, Metrics: metrics,
		})
		return layout.Dimensions{Size: size}
	})
}

func (d Widget) layoutDayView(ctx *core.Context, gtx layout.Context, state *datePickerState, metrics theme.DatePickerMetrics, now time.Time) layout.Dimensions {
	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return d.layoutWeekdays(ctx, gtx, state, metrics)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layoutui.Spacer(0, 4).Layout(ctx, gtx)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return d.layoutDays(ctx, gtx, state, metrics, now)
		}),
	)
}

func (d Widget) layoutWeekdays(ctx *core.Context, gtx layout.Context, state *datePickerState, metrics theme.DatePickerMetrics) layout.Dimensions {
	cell := gtx.Dp(metrics.CellSize)
	weekdayState := style.State{Disabled: !gtx.Enabled()}
	weekdayKey := uint8(0)
	if weekdayState.Disabled {
		weekdayKey = 1
	}
	resolved, ok := state.resolvedParts.weekdayValue(weekdayKey)
	if !ok {
		resolved = d.resolveDatePickerPart(ctx, gtx, PartCalendarWeekday, weekdayState, &state.calendarStyle)
		state.resolvedParts.storeWeekday(weekdayKey, resolved)
	}
	return layout.Flex{Axis: layout.Horizontal}.Layout(gtx, d.datePickerWeekdayChildren(ctx, cell, orderedDatePickerWeekdays(d.locale), resolved)...)
}

func (d Widget) datePickerWeekdayChildren(ctx *core.Context, cell int, weekdays [7]string, resolved style.ResolvedStyle) []layout.FlexChild {
	var children [len(weekdays)]layout.FlexChild
	for index, weekday := range weekdays {
		children[index] = layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints = layout.Exact(image.Pt(cell, cell))
			return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				weekdayWidget := text.WithResolvedText(text.New(weekday), resolved.Text)
				return weekdayWidget.Layout(ctx, gtx)
			})
		})
	}
	return children[:]
}

func (d Widget) layoutDays(ctx *core.Context, gtx layout.Context, state *datePickerState, metrics theme.DatePickerMetrics, now time.Time) layout.Dimensions {
	days := state.monthDays(state.viewMonth, d.locale.WeekStart)
	cell := gtx.Dp(metrics.CellSize)
	state.calendarLayout.prepare(ctx, d, state, days, cell, metrics, now)
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx, state.calendarLayout.rows[:]...)
}

func (d Widget) layoutMonths(ctx *core.Context, gtx layout.Context, state *datePickerState, metrics theme.DatePickerMetrics, now time.Time) layout.Dimensions {
	year, _, _ := state.viewMonth.Date()
	cellWidth := gtx.Dp(metrics.CalendarWidth) / 3
	cellHeight := gtx.Dp(metrics.MonthCellHeight)
	children := make([]layout.FlexChild, 0, 4)
	for row := range 4 {
		monthStart := row*3 + 1
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					gtx.Constraints = layout.Exact(image.Pt(cellWidth, cellHeight))
					return d.layoutMonth(ctx, gtx, state, year, time.Month(monthStart), metrics, now)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					gtx.Constraints = layout.Exact(image.Pt(cellWidth, cellHeight))
					return d.layoutMonth(ctx, gtx, state, year, time.Month(monthStart+1), metrics, now)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					gtx.Constraints = layout.Exact(image.Pt(cellWidth, cellHeight))
					return d.layoutMonth(ctx, gtx, state, year, time.Month(monthStart+2), metrics, now)
				}),
			)
		}))
	}
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx, children...)
}

func (d Widget) layoutMonth(ctx *core.Context, gtx layout.Context, pickerState *datePickerState, year int, month time.Month, metrics theme.DatePickerMetrics, now time.Time) layout.Dimensions {
	date := time.Date(year, month, 1, 0, 0, 0, 0, pickerState.viewMonth.Location())
	key := fmt.Sprintf("%04d-%02d", year, int(month))
	monthState := pickerState.month(key)
	focusVisible := core.FocusVisible(ctx, &monthState.clickable, gtx.Focused(&monthState.clickable))
	disabled := d.isMonthDisabled(date)
	presses := core.ActivePresses(monthState.clickable.History())
	if disabled {
		gtx = gtx.Disabled()
	} else {
		for monthState.clickable.Clicked(gtx) {
			pickerState.viewMonth = date
			pickerState.viewMode = datePickerViewDays
			core.RequestFocusVisible(ctx, &pickerState.trigger, focusVisible)
		}
		core.FocusOnPress(ctx, &pickerState.trigger, monthState.clickable.History(), presses)
	}

	valueYear, valueMonth, _ := d.value.Date()
	selected := !d.value.IsZero() && valueYear == year && valueMonth == month
	today := now.Year() == year && now.Month() == month
	active := pickerState.viewMonth.Year() == year && pickerState.viewMonth.Month() == month
	return d.layoutPickerCell(ctx, gtx, pickerState, monthState, d.locale.Months[int(month)-1], metrics, selected, active, today, false, disabled)
}

func (d Widget) layoutYears(ctx *core.Context, gtx layout.Context, state *datePickerState, metrics theme.DatePickerMetrics, now time.Time) layout.Dimensions {
	minYear, maxYear := d.yearPickerRange(state, now)
	count := maxYear - minYear + 1
	if count <= 0 {
		return layout.Dimensions{}
	}

	rows := (count + 2) / 3
	cellHeight := gtx.Dp(metrics.YearCellHeight)
	if cellHeight == 0 {
		cellHeight = gtx.Dp(metrics.MonthCellHeight)
	}
	state.yearList.Axis = layout.Vertical
	state.yearList.Gap = gtx.Dp(metrics.YearGridGap)
	if state.yearScrollReady {
		target := min(max(state.yearScrollYear, minYear), maxYear)
		state.yearList.ScrollTo((target - minYear) / 3)
		state.yearScrollReady = false
	}
	return layoutui.TrackedScrollbarWithVisualOutset(ctx, gtx, &state.yearList, &state.yearBar, &state.yearVisualOutset, rows, !gtx.Enabled(), false, func(gtx layout.Context, row int) layout.Dimensions {
		cellWidth := gtx.Constraints.Max.X / 3
		yearStart := minYear + row*3
		return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints = layout.Exact(image.Pt(cellWidth, cellHeight))
				return d.layoutYearCell(ctx, gtx, state, yearStart, maxYear, metrics, now)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints = layout.Exact(image.Pt(cellWidth, cellHeight))
				return d.layoutYearCell(ctx, gtx, state, yearStart+1, maxYear, metrics, now)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints = layout.Exact(image.Pt(cellWidth, cellHeight))
				return d.layoutYearCell(ctx, gtx, state, yearStart+2, maxYear, metrics, now)
			}),
		)
	})
}

func (d Widget) layoutYearCell(ctx *core.Context, gtx layout.Context, state *datePickerState, year, maxYear int, metrics theme.DatePickerMetrics, now time.Time) layout.Dimensions {
	if year > maxYear {
		return layout.Dimensions{Size: gtx.Constraints.Max}
	}
	return d.layoutYear(ctx, gtx, state, year, metrics, now)
}

func (d Widget) yearPickerRange(state *datePickerState, now time.Time) (int, int) {
	hasMin := !d.minDate.IsZero()
	hasMax := !d.maxDate.IsZero()
	minYear := 1900
	maxYear := 2099
	if hasMin {
		minYear = d.minDate.Year()
	}
	if hasMax {
		maxYear = d.maxDate.Year()
	}
	for _, year := range []int{state.viewMonth.Year(), now.Year()} {
		if !hasMin && year < minYear {
			minYear = year
		}
		if !hasMax && year > maxYear {
			maxYear = year
		}
	}
	if !d.value.IsZero() {
		year := d.value.Year()
		if !hasMin && year < minYear {
			minYear = year
		}
		if !hasMax && year > maxYear {
			maxYear = year
		}
	}
	if minYear > maxYear {
		minYear, maxYear = maxYear, minYear
	}
	return minYear, maxYear
}

func (d Widget) layoutYear(ctx *core.Context, gtx layout.Context, pickerState *datePickerState, year int, metrics theme.DatePickerMetrics, now time.Time) layout.Dimensions {
	yearState := pickerState.year(fmt.Sprintf("%04d", year))
	focusVisible := core.FocusVisible(ctx, &yearState.clickable, gtx.Focused(&yearState.clickable))
	disabled := d.isYearDisabled(year)
	presses := core.ActivePresses(yearState.clickable.History())
	if disabled {
		gtx = gtx.Disabled()
	} else {
		for yearState.clickable.Clicked(gtx) {
			_, month, _ := pickerState.viewMonth.Date()
			pickerState.viewMonth = time.Date(year, month, 1, 0, 0, 0, 0, pickerState.viewMonth.Location())
			pickerState.viewMode = datePickerViewDays
			core.RequestFocusVisible(ctx, &pickerState.trigger, focusVisible)
		}
		core.FocusOnPress(ctx, &pickerState.trigger, yearState.clickable.History(), presses)
	}

	selected := !d.value.IsZero() && d.value.Year() == year
	today := now.Year() == year
	active := pickerState.viewMonth.Year() == year
	return d.layoutPickerCell(ctx, gtx, pickerState, yearState, d.locale.YearLabel(year), metrics, selected, active, today, false, disabled)
}

func (d Widget) layoutPickerCell(ctx *core.Context, gtx layout.Context, pickerState *datePickerState, state *datePickerCellState, label string, metrics theme.DatePickerMetrics, selected, active, today, outside, disabled bool) layout.Dimensions {
	visual := pickerState.partDefaults.currentVisual(core.ActiveTheme(ctx))
	focusVisible := core.FocusVisible(ctx, &state.clickable, gtx.Focused(&state.clickable))
	return state.clickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		size := gtx.Constraints.Max
		cellState := style.State{
			Hovered:      state.clickable.Hovered() && !disabled,
			Pressed:      state.clickable.Pressed() && !disabled,
			Focused:      gtx.Focused(&state.clickable),
			FocusVisible: focusVisible && !disabled,
			Disabled:     disabled,
			Selected:     selected,
		}
		cellContext := CellStyleContext{
			Hovered:  cellState.Hovered,
			Pressed:  cellState.Pressed,
			Selected: selected,
			Today:    today,
			Outside:  outside,
			Disabled: disabled,
			Active:   active,
		}
		cellKey := datePickerResolvedCellKey(cellContext, cellState)
		resolved, cached := pickerState.resolvedParts.cellValue(cellKey)
		if !cached {
			base := datePickerBaseStyle(&pickerState.calendarStyle)
			defaults := pickerState.partDefaults.cellDefaults(core.ActiveTheme(ctx), base, cellContext)
			resolved = d.resolveDatePickerPartWithDefaults(ctx, gtx, PartCalendarCell, cellState, defaults)
			pickerState.resolvedParts.storeCell(cellKey, resolved)
		}
		focusKey := datePickerStyleStateKey(cellState)
		focusResolved, cached := pickerState.resolvedParts.focusValue(focusKey)
		if !cached {
			focusResolved = d.resolveDatePickerPart(ctx, gtx, PartCalendarFocus, cellState, &pickerState.calendarStyle)
			pickerState.resolvedParts.storeFocus(focusKey, focusResolved)
		}
		focusColor := datePickerOutlineColor(visual, focusResolved)
		motion := pickerState.partDefaults.motion(core.ActiveTheme(ctx))
		background := state.background(gtx, datePickerPaintColor(visual, resolved), motion)
		radius := datePickerPaintRadius(gtx, resolved)
		scale := datePickerPressScale(gtx, state.clickable.History(), disabled, motion)
		stack := draw.Scale(size, scale).Push(gtx.Ops)
		visual.CellRenderer(gtx, theme.DatePickerCellRenderRequest{
			Size: size, Background: background, Radius: radius, Metrics: metrics,
		})
		layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			labelWidget := text.WithResolvedText(text.New(label), resolved.Text)
			return labelWidget.Layout(ctx, gtx)
		})
		stack.Pop()
		visual.Focus(gtx, theme.DatePickerFocusRenderRequest{
			Size: size, Color: focusColor, Radius: radius,
			Visible: focusVisible && !disabled, FocusWidth: metrics.FocusRingWidth, Metrics: metrics,
		})
		return layout.Dimensions{Size: size}
	})
}

func (d Widget) layoutDay(ctx *core.Context, gtx layout.Context, pickerState *datePickerState, day datePickerDay, column int, previousOutside, nextOutside bool, metrics theme.DatePickerMetrics, now time.Time) layout.Dimensions {
	visual := pickerState.partDefaults.currentVisual(core.ActiveTheme(ctx))
	key := dateKey(day.date)
	dayState := pickerState.day(key)
	dayState.date = day.date
	focusVisible := core.FocusVisible(ctx, &dayState.clickable, gtx.Focused(&dayState.clickable))
	disabled := d.isDateDisabled(day.date)
	presses := core.ActivePresses(dayState.clickable.History())
	if disabled {
		gtx = gtx.Disabled()
	} else {
		for dayState.clickable.Clicked(gtx) {
			date := dateOnly(day.date)
			if d.onDateSelect != nil {
				d.onDateSelect(date)
			} else {
				pickerState.open = false
				pickerState.sync(date, date)
				pickerState.requestValue(d, date)
			}
			core.RequestFocusVisible(ctx, &pickerState.trigger, focusVisible)
		}
		core.FocusOnPress(ctx, &pickerState.trigger, dayState.clickable.History(), presses)
	}

	selectionStart := d.rangeMode && !day.outside && sameDate(d.rangeStart, day.date)
	selectionEnd := d.rangeMode && !day.outside && sameDate(d.rangeEnd, day.date)
	rangeSelected := d.rangeMode && dateBetween(day.date, d.rangeStart, d.rangeEnd)
	selected := sameDate(d.value, day.date)
	if d.rangeMode {
		selected = selectionStart || selectionEnd
	}
	today := sameDate(now, day.date)
	return dayState.clickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		size := gtx.Constraints.Max
		cellState := style.State{
			Hovered:      dayState.clickable.Hovered() && !disabled,
			Pressed:      dayState.clickable.Pressed() && !disabled,
			Focused:      gtx.Focused(&dayState.clickable),
			FocusVisible: focusVisible && !disabled,
			Disabled:     disabled,
			Selected:     selected,
		}
		cellContext := CellStyleContext{
			Hovered:  cellState.Hovered,
			Pressed:  cellState.Pressed,
			Selected: selected,
			Today:    today,
			Outside:  day.outside,
			Disabled: disabled,
			Range:    rangeSelected && !selected,
		}
		cellKey := datePickerResolvedCellKey(cellContext, cellState)
		resolved, cached := pickerState.resolvedParts.cellValue(cellKey)
		if !cached {
			base := datePickerBaseStyle(&pickerState.calendarStyle)
			defaults := pickerState.partDefaults.cellDefaults(core.ActiveTheme(ctx), base, cellContext)
			resolved = d.resolveDatePickerPartWithDefaults(ctx, gtx, PartCalendarCell, cellState, defaults)
			pickerState.resolvedParts.storeCell(cellKey, resolved)
		}
		foreground := datePickerTextColor(visual, resolved)
		focusKey := datePickerStyleStateKey(cellState)
		focusResolved, cached := pickerState.resolvedParts.focusValue(focusKey)
		if !cached {
			focusResolved = d.resolveDatePickerPart(ctx, gtx, PartCalendarFocus, cellState, &pickerState.calendarStyle)
			pickerState.resolvedParts.storeFocus(focusKey, focusResolved)
		}
		focusColor := datePickerOutlineColor(visual, focusResolved)
		motion := pickerState.partDefaults.motion(core.ActiveTheme(ctx))
		background := dayState.background(gtx, datePickerPaintColor(visual, resolved), motion)
		radius := datePickerPaintRadius(gtx, resolved)
		if rangeSelected {
			startRadius, endRadius := 0, 0
			if column == 0 {
				startRadius = gtx.Dp(metrics.RangeRadius)
			}
			if column == 6 {
				endRadius = gtx.Dp(metrics.RangeRadius)
			}
			if !day.outside && previousOutside {
				startRadius = gtx.Dp(metrics.RangeRadius)
			}
			if !day.outside && nextOutside {
				endRadius = gtx.Dp(metrics.RangeRadius)
			}
			if selectionStart {
				startRadius = size.Y / 2
			}
			if selectionEnd {
				endRadius = size.Y / 2
			}
			rangeKey := focusKey << 1
			if day.outside {
				rangeKey |= 1
			}
			rangeResolved, cached := pickerState.resolvedParts.rangeValue(rangeKey)
			if !cached {
				base := datePickerBaseStyle(&pickerState.calendarStyle)
				rangeDefaults := pickerState.partDefaults.rangeDefaults(core.ActiveTheme(ctx), base, day.outside)
				rangeResolved = d.resolveDatePickerPartWithDefaults(ctx, gtx, PartCalendarRange, cellState, rangeDefaults)
				pickerState.resolvedParts.storeRange(rangeKey, rangeResolved)
			}
			trackColor := datePickerPaintColor(visual, rangeResolved)
			visual.RangeRenderer(gtx, theme.DatePickerRangeRenderRequest{
				Size: size, Color: trackColor, StartRadius: startRadius, EndRadius: endRadius, Metrics: metrics,
			})
		}
		scale := datePickerPressScale(gtx, dayState.clickable.History(), disabled, motion)
		stack := draw.Scale(size, scale).Push(gtx.Ops)
		visual.CellRenderer(gtx, theme.DatePickerCellRenderRequest{
			Size: size, Background: background, Radius: radius, Metrics: metrics,
		})
		layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			labelWidget := text.WithResolvedText(text.New(day.label), resolved.Text)
			return labelWidget.Layout(ctx, gtx)
		})
		if disabled && !day.outside {
			visual.Strike(gtx, theme.DatePickerStrikeRenderRequest{
				Size: size, Color: foreground, HalfSize: metrics.CellStrikeHalfSize,
				Width: metrics.CellStrikeWidth, Metrics: metrics,
			})
		}
		stack.Pop()
		visual.Focus(gtx, theme.DatePickerFocusRenderRequest{
			Size: size, Color: focusColor, Radius: radius,
			Visible: focusVisible && !disabled, FocusWidth: metrics.FocusRingWidth, Metrics: metrics,
		})
		return layout.Dimensions{Size: size}
	})
}
