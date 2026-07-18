package datepicker

import (
	"fmt"
	"image"
	"image/color"
	"time"

	"gioui.org/f32"
	"gioui.org/font"
	"gioui.org/io/pointer"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"github.com/qianniancn/FlowUI/internal/components/description"
	"github.com/qianniancn/FlowUI/internal/components/label"
	layoutui "github.com/qianniancn/FlowUI/internal/components/layout"
	"github.com/qianniancn/FlowUI/internal/components/text"
	"github.com/qianniancn/FlowUI/internal/field"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/locale"
	"github.com/qianniancn/FlowUI/internal/overlay"
	"github.com/qianniancn/FlowUI/internal/render"
	"github.com/qianniancn/FlowUI/internal/state"
)

func (d DatePickerWidget) layoutField(ctx *frame.Context, gtx layout.Context, pickerState *datePickerState, style field.Style, enabled, invalid bool) (layout.Dimensions, image.Rectangle) {
	var children [3]layout.FlexChild
	count := 0
	labelHeight := 0
	inputAnchor := image.Rectangle{}
	gap := gtx.Dp(frame.ActiveTheme(ctx).Components.DatePicker.FieldGap)
	if d.label != "" {
		children[count] = layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			dimensions := label.Label(d.label).
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
		dimensions := d.layoutInput(ctx, gtx, pickerState, style, enabled, invalid)
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
				Size(float32(frame.ActiveTheme(ctx).Components.Description.TextSize)).
				Color(frame.ActiveTheme(ctx).Palette.Danger).
				Layout(ctx, gtx)
		})
		count++
	} else if d.description != "" {
		children[count] = layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return description.Description(d.description).
				For(d.key).
				Disabled(!enabled).
				Layout(ctx, gtx)
		})
		count++
	}
	dimensions := layout.Flex{Axis: layout.Vertical, Gap: gap}.Layout(gtx, children[:count]...)
	return dimensions, inputAnchor
}

func (d DatePickerWidget) layoutInput(ctx *frame.Context, gtx layout.Context, pickerState *datePickerState, style field.Style, enabled, invalid bool) layout.Dimensions {
	presses := state.ActivePresses(pickerState.trigger.History())
	focusVisible := frame.FocusVisible(ctx, &pickerState.trigger, gtx.Focused(&pickerState.trigger))
	if enabled {
		for pickerState.trigger.Clicked(gtx) {
			if pickerState.open {
				pickerState.open = false
			} else {
				pickerState.openCalendar()
			}
			frame.RequestFocusVisible(ctx, &pickerState.trigger, focusVisible)
		}
		frame.FocusOnPress(ctx, &pickerState.trigger, pickerState.trigger.History(), presses)
	}

	frameConstraints := gtx.Constraints
	if d.fullWidth {
		frameConstraints.Min.X = frameConstraints.Max.X
	}
	theme := frame.ActiveTheme(ctx).Components.DatePicker
	height := min(gtx.Dp(theme.Height), frameConstraints.Max.Y)
	frameConstraints.Min.Y = min(max(frameConstraints.Min.Y, height), frameConstraints.Max.Y)

	left := gtx.Dp(frame.ActiveTheme(ctx).Components.Input.PaddingX)
	right := gtx.Dp(theme.TriggerWidth)
	horizontalPadding := left + right
	maxX := max(frameConstraints.Max.X-horizontalPadding, 0)
	minX := min(max(frameConstraints.Min.X-horizontalPadding, 0), maxX)

	macro := op.Record(gtx.Ops)
	contentGtx := gtx
	contentGtx.Constraints = layout.Constraints{
		Min: image.Pt(minX, 0),
		Max: image.Pt(maxX, height),
	}
	contentDims := pickerState.segments.layout(ctx, contentGtx, d.locale, style, enabled, invalid, d.minDate, d.maxDate, d.onChange)
	call := macro.Stop()
	showHint := d.hintSet && pickerState.segments.empty() && !pickerState.segments.focused(gtx)
	var hintCall op.CallOp
	var hintDims layout.Dimensions
	if showHint {
		hintMacro := op.Record(gtx.Ops)
		hintColor := style.Placeholder
		if invalid {
			hintColor = frame.ActiveTheme(ctx).Palette.Danger
		}
		hintGtx := contentGtx
		hintGtx.Constraints.Min = image.Point{}
		hintDims = text.New(d.hint).
			Size(float32(theme.TextSize)).
			Color(hintColor).
			MaxLines(1).
			Truncator("...").
			Layout(ctx, hintGtx)
		hintCall = hintMacro.Stop()
		contentDims.Size.X = max(contentDims.Size.X, hintDims.Size.X)
		contentDims.Size.Y = max(contentDims.Size.Y, hintDims.Size.Y)
	}

	size := frameConstraints.Constrain(image.Pt(contentDims.Size.X+horizontalPadding, height))
	rect := image.Rectangle{Max: size}
	radius := min(max(gtx.Dp(theme.Radius), 1), min(size.X, size.Y)/2)
	field.DrawFrame(gtx, rect, radius, style)
	clipped := clip.UniformRRect(rect, radius).Push(gtx.Ops)
	stack := op.Offset(image.Pt(left, max((size.Y-contentDims.Size.Y)/2, 0))).Push(gtx.Ops)
	contentClip := clip.Rect{Max: image.Pt(max(size.X-horizontalPadding, 0), contentDims.Size.Y)}.Push(gtx.Ops)
	call.Add(gtx.Ops)
	if showHint {
		paint.FillShape(gtx.Ops, style.Background, clip.Rect{Max: image.Pt(max(size.X-horizontalPadding, 0), contentDims.Size.Y)}.Op())
		hintOffset := op.Offset(image.Pt(0, max((contentDims.Size.Y-hintDims.Size.Y)/2, 0))).Push(gtx.Ops)
		hintCall.Add(gtx.Ops)
		hintOffset.Pop()
	}
	contentClip.Pop()
	stack.Pop()

	triggerSize := image.Pt(right, size.Y)
	triggerGtx := gtx
	triggerGtx.Constraints = layout.Exact(triggerSize)
	stack = op.Offset(image.Pt(size.X-right, 0)).Push(gtx.Ops)
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
		iconSize := image.Pt(gtx.Dp(theme.IconSize), gtx.Dp(theme.IconSize))
		iconOffset := op.Offset(image.Pt((right-iconSize.X)/2, (size.Y-iconSize.Y)/2)).Push(gtx.Ops)
		drawDatePickerCalendarIcon(gtx, iconSize, style.Placeholder)
		iconOffset.Pop()
		drawDatePickerTriggerFocus(gtx, frame.ActiveTheme(ctx), triggerSize, enabled && frame.FocusVisible(ctx, &pickerState.trigger, gtx.Focused(&pickerState.trigger)))
		return layout.Dimensions{Size: triggerSize}
	})
	stack.Pop()
	clipped.Pop()
	return layout.Dimensions{Size: size}
}

func (d DatePickerWidget) calendarLabel(ctx *frame.Context) string {
	if frame.ActiveLanguage(ctx) == locale.LanguageChinese {
		return "选择日期"
	}
	return "Choose date"
}

func (d DatePickerWidget) layoutPopover(ctx *frame.Context, gtx layout.Context, state *datePickerState, inputAnchor image.Rectangle, progress float32, now time.Time, naturallyDisabled bool) {
	frame.RegisterOverlay(ctx, frame.OverlayRequest{
		Key:       frame.FullKey(ctx, d.key),
		Layer:     frame.OverlayLayerPopup,
		Anchor:    inputAnchor,
		HasAnchor: true,
		Disabled:  naturallyDisabled,
		Layout: func(gtx layout.Context, anchor image.Rectangle, interactive bool) layout.Dimensions {
			contentInteractive := interactive && state.open && gtx.Enabled()
			if contentInteractive {
				state.updateKeys(gtx, &state.trigger)
			}
			panelGtx := gtx
			if !contentInteractive {
				panelGtx = panelGtx.Disabled()
			}
			return d.layoutCalendarOverlay(ctx, gtx, panelGtx, state, anchor, progress, now, contentInteractive)
		},
	})
}

func (d DatePickerWidget) layoutCalendarOverlay(ctx *frame.Context, gtx, panelGtx layout.Context, pickerState *datePickerState, anchor image.Rectangle, progress float32, now time.Time, interactive bool) layout.Dimensions {
	if interactive {
		for pickerState.dialog.Clicked(gtx) {
			frame.RequestFocusVisible(ctx, &pickerState.trigger, false)
		}
		if pickerState.dialog.TakePressed() {
			frame.RequestFocusVisible(ctx, &pickerState.trigger, false)
		}
	}
	theme := frame.ActiveTheme(ctx).Components.DatePicker
	viewport := gtx.Constraints.Max
	gap := gtx.Dp(theme.PopoverGap)
	panelWidth := min(gtx.Dp(theme.CalendarWidth)+gtx.Dp(theme.PopoverPadding)*2, max(viewport.X, 0))
	panelMaxY := min(gtx.Dp(theme.PopoverMaxHeight), max(viewport.Y-gap, 0))
	if panelWidth <= 0 || panelMaxY <= 0 {
		return layout.Dimensions{}
	}
	panelGtx.Constraints = layout.Constraints{
		Min: image.Pt(panelWidth, 0),
		Max: image.Pt(panelWidth, panelMaxY),
	}

	macro := op.Record(gtx.Ops)
	dims, tracked := frame.TrackOverlayPlacement(ctx, func() layout.Dimensions {
		return d.layoutCalendarPanel(ctx, panelGtx, pickerState, now)
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

func (d DatePickerWidget) layoutCalendarPanel(ctx *frame.Context, gtx layout.Context, state *datePickerState, now time.Time) layout.Dimensions {
	macro := op.Record(gtx.Ops)
	theme := frame.ActiveTheme(ctx).Components.DatePicker
	dims := layout.UniformInset(theme.PopoverPadding).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return d.layoutCalendar(ctx, gtx, state, now)
	})
	call := macro.Stop()

	radius := min(max(gtx.Dp(theme.PopoverRadius), 1), min(dims.Size.X, dims.Size.Y)/2)
	rect := image.Rectangle{Max: dims.Size}
	drawDatePickerPopover(gtx, frame.ActiveTheme(ctx), rect, radius)
	clipStack := clip.UniformRRect(rect, radius).Push(gtx.Ops)
	call.Add(gtx.Ops)
	clipStack.Pop()
	return dims
}

func (d DatePickerWidget) layoutCalendar(ctx *frame.Context, gtx layout.Context, state *datePickerState, now time.Time) layout.Dimensions {
	gtx.Constraints.Min.X = min(gtx.Dp(frame.ActiveTheme(ctx).Components.DatePicker.CalendarWidth), gtx.Constraints.Max.X)
	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return d.layoutCalendarHeader(ctx, gtx, state)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			switch state.viewMode {
			case datePickerViewMonths:
				return d.layoutMonths(ctx, gtx, state, now)
			case datePickerViewYears:
				return d.layoutYears(ctx, gtx, state, now)
			default:
				return d.layoutDayView(ctx, gtx, state, now)
			}
		}),
	)
}

func (d DatePickerWidget) layoutCalendarHeader(ctx *frame.Context, gtx layout.Context, pickerState *datePickerState) layout.Dimensions {
	return layout.Inset{
		Left:   2,
		Right:  2,
		Bottom: 16,
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{
			Axis:      layout.Horizontal,
			Alignment: layout.Middle,
		}.Layout(gtx,
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				presses := state.ActivePresses(pickerState.header.History())
				focusVisible := frame.FocusVisible(ctx, &pickerState.header, gtx.Focused(&pickerState.header))
				for pickerState.header.Clicked(gtx) {
					pickerState.toggleYearPicker(pickerState.viewMonth.Year())
					frame.RequestFocusVisible(ctx, &pickerState.trigger, focusVisible)
				}
				frame.FocusOnPress(ctx, &pickerState.trigger, pickerState.header.History(), presses)
				return pickerState.header.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					gtx.Constraints.Min.X = gtx.Constraints.Max.X
					return layout.W.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return d.layoutHeaderTrigger(ctx, gtx, pickerState)
					})
				})
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return d.layoutNavButton(ctx, gtx, pickerState, -1)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return d.layoutNavButton(ctx, gtx, pickerState, 1)
			}),
		)
	})
}

func (d DatePickerWidget) layoutHeaderTrigger(ctx *frame.Context, gtx layout.Context, state *datePickerState) layout.Dimensions {
	label := d.headerLabel(state)
	col := frame.ActiveTheme(ctx).Palette.Foreground
	if state.viewMode == datePickerViewYears {
		col = frame.ActiveTheme(ctx).Palette.AccentSoftForeground
	}
	gap := gtx.Dp(unit.Dp(4))
	iconSize := gtx.Dp(unit.Dp(14))
	return layout.Flex{
		Axis:      layout.Horizontal,
		Alignment: layout.Middle,
		Gap:       gap,
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return text.New(label).
				Size(float32(frame.ActiveTheme(ctx).Components.DatePicker.HeaderTextSize)).
				Weight(font.Medium).
				Color(col).
				Layout(ctx, gtx)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			size := image.Pt(iconSize, iconSize)
			gtx.Constraints = layout.Exact(size)
			drawDatePickerYearPickerIndicator(gtx, size, state.viewMode == datePickerViewYears, col)
			return layout.Dimensions{Size: size}
		}),
	)
}

func (d DatePickerWidget) headerLabel(state *datePickerState) string {
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

func (d DatePickerWidget) layoutNavButton(ctx *frame.Context, gtx layout.Context, pickerState *datePickerState, delta int) layout.Dimensions {
	size := image.Pt(gtx.Dp(frame.ActiveTheme(ctx).Components.DatePicker.NavButtonSize), gtx.Dp(frame.ActiveTheme(ctx).Components.DatePicker.NavButtonSize))
	if pickerState.viewMode == datePickerViewYears {
		return layout.Dimensions{Size: size}
	}

	clickable := &pickerState.prev
	if delta > 0 {
		clickable = &pickerState.next
	}
	disabled := !d.canMove(pickerState, delta)
	focusVisible := frame.FocusVisible(ctx, clickable, gtx.Focused(clickable))
	presses := state.ActivePresses(clickable.History())
	if disabled {
		gtx = gtx.Disabled()
	} else {
		for clickable.Clicked(gtx) {
			pickerState.move(delta)
			frame.RequestFocusVisible(ctx, &pickerState.trigger, focusVisible)
		}
		frame.FocusOnPress(ctx, &pickerState.trigger, clickable.History(), presses)
	}
	gtx.Constraints = layout.Exact(size)
	return clickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		style := datePickerNavStyle(frame.ActiveTheme(ctx), clickable.Hovered(), clickable.Pressed(), disabled)
		motion := frame.ActiveTheme(ctx).Motion
		style.bg = pickerState.navBackground(gtx, delta, style.bg, motion)
		scale := datePickerPressScale(gtx, clickable.History(), disabled, motion)
		stack := render.Scale(size, scale).Push(gtx.Ops)
		drawDatePickerNavButton(gtx, size, delta, style)
		stack.Pop()
		drawDatePickerControlFocus(gtx, frame.ActiveTheme(ctx), size, min(size.X, size.Y)/2, focusVisible && !disabled)
		return layout.Dimensions{Size: size}
	})
}

func (d DatePickerWidget) layoutDayView(ctx *frame.Context, gtx layout.Context, state *datePickerState, now time.Time) layout.Dimensions {
	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return d.layoutWeekdays(ctx, gtx)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return d.layoutDays(ctx, gtx, state, now)
		}),
	)
}

func (d DatePickerWidget) layoutWeekdays(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	cell := gtx.Dp(frame.ActiveTheme(ctx).Components.DatePicker.CellSize)
	return layout.Flex{Axis: layout.Horizontal}.Layout(gtx, datePickerWeekdayChildren(ctx, cell, orderedDatePickerWeekdays(d.locale))...)
}

func datePickerWeekdayChildren(ctx *frame.Context, cell int, weekdays [7]string) []layout.FlexChild {
	children := make([]layout.FlexChild, 0, len(weekdays))
	for _, weekday := range weekdays {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints = layout.Exact(image.Pt(cell, cell))
			return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return text.New(weekday).
					Size(float32(frame.ActiveTheme(ctx).Components.DatePicker.WeekdayTextSize)).
					Weight(font.Medium).
					Color(frame.ActiveTheme(ctx).Palette.MutedForeground).
					Layout(ctx, gtx)
			})
		}))
	}
	return children
}

func (d DatePickerWidget) layoutDays(ctx *frame.Context, gtx layout.Context, state *datePickerState, now time.Time) layout.Dimensions {
	days := datePickerMonthDays(state.viewMonth, d.locale.WeekStart)
	cell := gtx.Dp(frame.ActiveTheme(ctx).Components.DatePicker.CellSize)
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx, datePickerWeekRows(ctx, d, state, days, cell, now)...)
}

func (d DatePickerWidget) layoutMonths(ctx *frame.Context, gtx layout.Context, state *datePickerState, now time.Time) layout.Dimensions {
	year, _, _ := state.viewMonth.Date()
	theme := frame.ActiveTheme(ctx).Components.DatePicker
	cellWidth := gtx.Dp(theme.CalendarWidth) / 3
	cellHeight := gtx.Dp(theme.MonthCellHeight)
	children := make([]layout.FlexChild, 0, 4)
	for row := range 4 {
		monthStart := row*3 + 1
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					gtx.Constraints = layout.Exact(image.Pt(cellWidth, cellHeight))
					return d.layoutMonth(ctx, gtx, state, year, time.Month(monthStart), now)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					gtx.Constraints = layout.Exact(image.Pt(cellWidth, cellHeight))
					return d.layoutMonth(ctx, gtx, state, year, time.Month(monthStart+1), now)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					gtx.Constraints = layout.Exact(image.Pt(cellWidth, cellHeight))
					return d.layoutMonth(ctx, gtx, state, year, time.Month(monthStart+2), now)
				}),
			)
		}))
	}
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx, children...)
}

func (d DatePickerWidget) layoutMonth(ctx *frame.Context, gtx layout.Context, pickerState *datePickerState, year int, month time.Month, now time.Time) layout.Dimensions {
	date := time.Date(year, month, 1, 0, 0, 0, 0, pickerState.viewMonth.Location())
	key := fmt.Sprintf("%04d-%02d", year, int(month))
	monthState := pickerState.month(key)
	focusVisible := frame.FocusVisible(ctx, &monthState.clickable, gtx.Focused(&monthState.clickable))
	disabled := d.isMonthDisabled(date)
	presses := state.ActivePresses(monthState.clickable.History())
	if disabled {
		gtx = gtx.Disabled()
	} else {
		for monthState.clickable.Clicked(gtx) {
			pickerState.viewMonth = date
			pickerState.viewMode = datePickerViewDays
			frame.RequestFocusVisible(ctx, &pickerState.trigger, focusVisible)
		}
		frame.FocusOnPress(ctx, &pickerState.trigger, monthState.clickable.History(), presses)
	}

	valueYear, valueMonth, _ := d.value.Date()
	selected := !d.value.IsZero() && valueYear == year && valueMonth == month
	today := now.Year() == year && now.Month() == month
	active := pickerState.viewMonth.Year() == year && pickerState.viewMonth.Month() == month
	return d.layoutPickerCell(ctx, gtx, monthState, d.locale.Months[int(month)-1], selected, active, today, false, disabled)
}

func (d DatePickerWidget) layoutYears(ctx *frame.Context, gtx layout.Context, state *datePickerState, now time.Time) layout.Dimensions {
	minYear, maxYear := d.yearPickerRange(state, now)
	count := maxYear - minYear + 1
	if count <= 0 {
		return layout.Dimensions{}
	}

	theme := frame.ActiveTheme(ctx).Components.DatePicker
	rows := (count + 2) / 3
	cellHeight := gtx.Dp(theme.YearCellHeight)
	if cellHeight == 0 {
		cellHeight = gtx.Dp(theme.MonthCellHeight)
	}
	state.yearList.Axis = layout.Vertical
	state.yearList.Gap = gtx.Dp(theme.YearGridGap)
	if state.yearScrollReady {
		target := max(state.yearScrollYear, minYear)
		if target > maxYear {
			target = maxYear
		}
		state.yearList.ScrollTo((target - minYear) / 3)
		state.yearScrollReady = false
	}
	return layoutui.LayoutTrackedScrollbar(ctx, gtx, &state.yearList, &state.yearBar, rows, !gtx.Enabled(), false, func(gtx layout.Context, row int) layout.Dimensions {
		cellWidth := gtx.Constraints.Max.X / 3
		yearStart := minYear + row*3
		return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints = layout.Exact(image.Pt(cellWidth, cellHeight))
				return d.layoutYearCell(ctx, gtx, state, yearStart, maxYear, now)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints = layout.Exact(image.Pt(cellWidth, cellHeight))
				return d.layoutYearCell(ctx, gtx, state, yearStart+1, maxYear, now)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints = layout.Exact(image.Pt(cellWidth, cellHeight))
				return d.layoutYearCell(ctx, gtx, state, yearStart+2, maxYear, now)
			}),
		)
	})
}

func (d DatePickerWidget) layoutYearCell(ctx *frame.Context, gtx layout.Context, state *datePickerState, year, maxYear int, now time.Time) layout.Dimensions {
	if year > maxYear {
		return layout.Dimensions{Size: gtx.Constraints.Max}
	}
	return d.layoutYear(ctx, gtx, state, year, now)
}

func (d DatePickerWidget) yearPickerRange(state *datePickerState, now time.Time) (int, int) {
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

func (d DatePickerWidget) layoutYear(ctx *frame.Context, gtx layout.Context, pickerState *datePickerState, year int, now time.Time) layout.Dimensions {
	yearState := pickerState.year(fmt.Sprintf("%04d", year))
	focusVisible := frame.FocusVisible(ctx, &yearState.clickable, gtx.Focused(&yearState.clickable))
	disabled := d.isYearDisabled(year)
	presses := state.ActivePresses(yearState.clickable.History())
	if disabled {
		gtx = gtx.Disabled()
	} else {
		for yearState.clickable.Clicked(gtx) {
			_, month, _ := pickerState.viewMonth.Date()
			pickerState.viewMonth = time.Date(year, month, 1, 0, 0, 0, 0, pickerState.viewMonth.Location())
			pickerState.viewMode = datePickerViewDays
			frame.RequestFocusVisible(ctx, &pickerState.trigger, focusVisible)
		}
		frame.FocusOnPress(ctx, &pickerState.trigger, yearState.clickable.History(), presses)
	}

	selected := !d.value.IsZero() && d.value.Year() == year
	today := now.Year() == year
	active := pickerState.viewMonth.Year() == year
	return d.layoutPickerCell(ctx, gtx, yearState, d.locale.YearLabel(year), selected, active, today, false, disabled)
}

func (d DatePickerWidget) layoutPickerCell(ctx *frame.Context, gtx layout.Context, state *datePickerCellState, label string, selected, active, today, outside, disabled bool) layout.Dimensions {
	focusVisible := frame.FocusVisible(ctx, &state.clickable, gtx.Focused(&state.clickable))
	return state.clickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		size := gtx.Constraints.Max
		style := datePickerCellStyleFor(frame.ActiveTheme(ctx), state.clickable.Hovered(), state.clickable.Pressed(), selected, today, outside, disabled)
		if active && !selected && !disabled {
			style.bg = frame.ActiveTheme(ctx).Palette.SurfaceRaised
			if !today {
				style.fg = frame.ActiveTheme(ctx).Palette.AccentSoftForeground
			}
		}
		motion := frame.ActiveTheme(ctx).Motion
		style.bg = state.background(gtx, style.bg, motion)
		scale := datePickerPressScale(gtx, state.clickable.History(), disabled, motion)
		stack := render.Scale(size, scale).Push(gtx.Ops)
		drawDatePickerCell(gtx, size, style)
		layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return text.New(label).
				Size(float32(frame.ActiveTheme(ctx).Components.DatePicker.CellTextSize)).
				Weight(font.Medium).
				Color(style.fg).
				Layout(ctx, gtx)
		})
		stack.Pop()
		drawDatePickerControlFocus(gtx, frame.ActiveTheme(ctx), size, min(size.X, size.Y)/2, focusVisible && !disabled)
		return layout.Dimensions{Size: size}
	})
}

func datePickerWeekRows(ctx *frame.Context, picker DatePickerWidget, state *datePickerState, days []datePickerDay, cell int, now time.Time) []layout.FlexChild {
	rows := make([]layout.FlexChild, 0, 6)
	for week := range 6 {
		weekDays := days[week*7 : week*7+7]
		rows = append(rows, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal}.Layout(gtx, datePickerDayChildren(ctx, picker, state, weekDays, cell, now)...)
		}))
	}
	return rows
}

func datePickerDayChildren(ctx *frame.Context, picker DatePickerWidget, state *datePickerState, days []datePickerDay, cell int, now time.Time) []layout.FlexChild {
	children := make([]layout.FlexChild, 0, len(days))
	for column, day := range days {
		previousOutside := column > 0 && days[column-1].outside
		nextOutside := column < len(days)-1 && days[column+1].outside
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints = layout.Exact(image.Pt(cell, cell))
			return picker.layoutDay(ctx, gtx, state, day, column, previousOutside, nextOutside, now)
		}))
	}
	return children
}

func (d DatePickerWidget) layoutDay(ctx *frame.Context, gtx layout.Context, pickerState *datePickerState, day datePickerDay, column int, previousOutside, nextOutside bool, now time.Time) layout.Dimensions {
	key := dateKey(day.date)
	dayState := pickerState.day(key)
	dayState.date = day.date
	focusVisible := frame.FocusVisible(ctx, &dayState.clickable, gtx.Focused(&dayState.clickable))
	disabled := d.isDateDisabled(day.date)
	presses := state.ActivePresses(dayState.clickable.History())
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
				if d.onChange != nil {
					d.onChange(date)
				}
			}
			frame.RequestFocusVisible(ctx, &pickerState.trigger, focusVisible)
		}
		frame.FocusOnPress(ctx, &pickerState.trigger, dayState.clickable.History(), presses)
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
		style := datePickerCellStyleFor(frame.ActiveTheme(ctx), dayState.clickable.Hovered(), dayState.clickable.Pressed(), selected, today, day.outside, disabled)
		if rangeSelected && !selected {
			style.bg = color.NRGBA{}
			if !day.outside {
				style.fg = frame.ActiveTheme(ctx).Palette.AccentSoftForeground
			}
		}
		motion := frame.ActiveTheme(ctx).Motion
		style.bg = dayState.background(gtx, style.bg, motion)
		if rangeSelected {
			startRadius, endRadius := 0, 0
			if column == 0 {
				startRadius = gtx.Dp(frame.ActiveTheme(ctx).Components.DatePicker.RangeRadius)
			}
			if column == 6 {
				endRadius = gtx.Dp(frame.ActiveTheme(ctx).Components.DatePicker.RangeRadius)
			}
			if !day.outside && previousOutside {
				startRadius = gtx.Dp(frame.ActiveTheme(ctx).Components.DatePicker.RangeRadius)
			}
			if !day.outside && nextOutside {
				endRadius = gtx.Dp(frame.ActiveTheme(ctx).Components.DatePicker.RangeRadius)
			}
			if selectionStart {
				startRadius = size.Y / 2
			}
			if selectionEnd {
				endRadius = size.Y / 2
			}
			trackColor := frame.ActiveTheme(ctx).Palette.AccentSoft
			if day.outside {
				trackColor = frame.ActiveTheme(ctx).Palette.SurfaceRaised
				trackColor.A = byte(uint16(trackColor.A) / 5)
			}
			drawDatePickerRangeTrack(gtx, trackColor, size, startRadius, endRadius)
		}
		scale := datePickerPressScale(gtx, dayState.clickable.History(), disabled, motion)
		stack := render.Scale(size, scale).Push(gtx.Ops)
		drawDatePickerCell(gtx, size, style)
		layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return text.New(fmt.Sprintf("%d", day.date.Day())).
				Size(float32(frame.ActiveTheme(ctx).Components.DatePicker.CellTextSize)).
				Weight(font.Medium).
				Color(style.fg).
				Layout(ctx, gtx)
		})
		if disabled && !day.outside {
			drawDatePickerStrike(gtx, frame.ActiveTheme(ctx), size, style.fg)
		}
		stack.Pop()
		drawDatePickerControlFocus(gtx, frame.ActiveTheme(ctx), size, min(size.X, size.Y)/2, focusVisible && !disabled)
		return layout.Dimensions{Size: size}
	})
}
