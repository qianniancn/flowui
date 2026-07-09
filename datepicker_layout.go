package flowui

import (
	"fmt"
	"image"
	"image/color"
	"time"

	"gioui.org/f32"
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
)

func (d DatePickerWidget) layoutInput(ctx *Context, gtx layout.Context, state *datePickerState, style inputStyle) layout.Dimensions {
	presses := activePresses(state.trigger.History())
	if !d.disabled {
		for state.trigger.Clicked(gtx) {
			if state.open {
				state.open = false
			} else {
				state.openCalendar()
			}
			ctx.requestFocus(&state.trigger)
		}
		ctx.focusOnPress(&state.trigger, state.trigger.History(), presses)
	}

	frameConstraints := gtx.Constraints
	if d.fullWidth {
		frameConstraints.Min.X = frameConstraints.Max.X
	}
	theme := ctx.Theme.Components.DatePicker
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
	childDims := d.layoutSegments(ctx, childGtx, style)
	call := macro.Stop()

	size := image.Pt(childDims.Size.X+horizontalPadding, childDims.Size.Y)
	size = frameConstraints.Constrain(size)
	rect := image.Rectangle{Max: size}
	radius := min(max(gtx.Dp(theme.Radius), 1), min(size.X, size.Y)/2)
	drawInputFrame(gtx, rect, radius, style)

	stack := op.Offset(image.Pt(left, max((size.Y-childDims.Size.Y)/2, 0))).Push(gtx.Ops)
	call.Add(gtx.Ops)
	stack.Pop()

	iconSize := image.Pt(gtx.Dp(theme.IconSize), gtx.Dp(theme.IconSize))
	iconOffset := image.Pt(size.X-right+(right-iconSize.X)/2, (size.Y-iconSize.Y)/2)
	stack = op.Offset(iconOffset).Push(gtx.Ops)
	drawDatePickerCalendarIcon(gtx, ctx.Theme, iconSize, style.placeholder)
	stack.Pop()

	return state.trigger.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Dimensions{Size: size}
	})
}

func (d DatePickerWidget) layoutSegments(ctx *Context, gtx layout.Context, style inputStyle) layout.Dimensions {
	if !d.value.IsZero() {
		return datePickerSegment(ctx, d.locale.DateLabel(d.value), style.fg, font.Medium)(gtx)
	}
	return Text(d.hint).
		Size(float32(ctx.Theme.Components.DatePicker.TextSize)).
		Color(style.placeholder).
		Layout(ctx, gtx)
}

func datePickerSegment(ctx *Context, text string, col color.NRGBA, weight font.Weight) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		widget := Text(text).Size(float32(ctx.Theme.Components.DatePicker.TextSize)).Color(col)
		if weight != 0 {
			widget = widget.Weight(weight)
		}
		return widget.Layout(ctx, gtx)
	}
}

func (d DatePickerWidget) layoutPopover(ctx *Context, gtx layout.Context, state *datePickerState, inputDims layout.Dimensions, progress float32, now time.Time) {
	theme := ctx.Theme.Components.DatePicker
	gap := gtx.Dp(theme.PopoverGap)
	maxY := gtx.Constraints.Max.Y - inputDims.Size.Y - gap
	if maxY <= 0 {
		return
	}
	panelWidth := gtx.Dp(theme.CalendarWidth) + gtx.Dp(theme.PopoverPadding)*2
	panelConstraints := layout.Constraints{
		Min: image.Pt(panelWidth, 0),
		Max: image.Pt(panelWidth, min(maxY, gtx.Dp(theme.PopoverMaxHeight))),
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
		opacity := paint.PushOpacity(gtx.Ops, progress)
		transform := op.Affine(f32.AffineId().Scale(origin, f32.Pt(scale, scale))).Push(gtx.Ops)
		dims := d.layoutCalendarPanel(ctx, gtx, state, now)
		transform.Pop()
		opacity.Pop()
		stack.Pop()
		return dims
	})
}

func (d DatePickerWidget) layoutCalendarPanel(ctx *Context, gtx layout.Context, state *datePickerState, now time.Time) layout.Dimensions {
	macro := op.Record(gtx.Ops)
	theme := ctx.Theme.Components.DatePicker
	dims := layout.UniformInset(theme.PopoverPadding).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return d.layoutCalendar(ctx, gtx, state, now)
	})
	call := macro.Stop()

	radius := min(max(gtx.Dp(theme.PopoverRadius), 1), min(dims.Size.X, dims.Size.Y)/2)
	rect := image.Rectangle{Max: dims.Size}
	drawDatePickerPopover(gtx, ctx.Theme, rect, radius)
	clipStack := clip.UniformRRect(rect, radius).Push(gtx.Ops)
	call.Add(gtx.Ops)
	clipStack.Pop()
	return dims
}

func (d DatePickerWidget) layoutCalendar(ctx *Context, gtx layout.Context, state *datePickerState, now time.Time) layout.Dimensions {
	gtx.Constraints.Min.X = min(gtx.Dp(ctx.Theme.Components.DatePicker.CalendarWidth), gtx.Constraints.Max.X)
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

func (d DatePickerWidget) layoutCalendarHeader(ctx *Context, gtx layout.Context, state *datePickerState) layout.Dimensions {
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
				for state.header.Clicked(gtx) {
					state.toggleYearPicker(state.viewMonth.Year())
					ctx.requestFocus(&state.trigger)
				}
				return state.header.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					gtx.Constraints.Min.X = gtx.Constraints.Max.X
					return layout.W.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return d.layoutHeaderTrigger(ctx, gtx, state)
					})
				})
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return d.layoutNavButton(ctx, gtx, state, -1)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return d.layoutNavButton(ctx, gtx, state, 1)
			}),
		)
	})
}

func (d DatePickerWidget) layoutHeaderTrigger(ctx *Context, gtx layout.Context, state *datePickerState) layout.Dimensions {
	label := d.headerLabel(state)
	col := ctx.Theme.Palette.Foreground
	if state.viewMode == datePickerViewYears {
		col = ctx.Theme.Palette.AccentSoftForeground
	}
	gap := gtx.Dp(unit.Dp(4))
	iconSize := gtx.Dp(unit.Dp(14))
	return layout.Flex{
		Axis:      layout.Horizontal,
		Alignment: layout.Middle,
		Gap:       gap,
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return Text(label).
				Size(float32(ctx.Theme.Components.DatePicker.HeaderTextSize)).
				Weight(font.Medium).
				Color(col).
				Layout(ctx, gtx)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			size := image.Pt(iconSize, iconSize)
			gtx.Constraints = layout.Exact(size)
			drawDatePickerYearPickerIndicator(gtx, ctx.Theme, size, state.viewMode == datePickerViewYears, col)
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

func (d DatePickerWidget) layoutNavButton(ctx *Context, gtx layout.Context, state *datePickerState, delta int) layout.Dimensions {
	size := image.Pt(gtx.Dp(ctx.Theme.Components.DatePicker.NavButtonSize), gtx.Dp(ctx.Theme.Components.DatePicker.NavButtonSize))
	if state.viewMode == datePickerViewYears {
		return layout.Dimensions{Size: size}
	}

	clickable := &state.prev
	if delta > 0 {
		clickable = &state.next
	}
	disabled := !d.canMove(state, delta)
	presses := activePresses(clickable.History())
	if disabled {
		gtx = gtx.Disabled()
	} else {
		for clickable.Clicked(gtx) {
			state.move(delta)
			ctx.requestFocus(&state.trigger)
		}
		ctx.focusOnPress(&state.trigger, clickable.History(), presses)
	}
	gtx.Constraints = layout.Exact(size)
	return clickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		style := datePickerNavStyle(ctx.Theme, clickable.Hovered(), clickable.Pressed(), disabled)
		style.bg = state.navBackground(gtx, delta, style.bg)
		scale := datePickerPressScale(gtx, clickable.History(), disabled)
		stack := buttonScale(size, scale).Push(gtx.Ops)
		drawDatePickerNavButton(gtx, size, delta, style)
		stack.Pop()
		return layout.Dimensions{Size: size}
	})
}

func (d DatePickerWidget) layoutDayView(ctx *Context, gtx layout.Context, state *datePickerState, now time.Time) layout.Dimensions {
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

func (d DatePickerWidget) layoutWeekdays(ctx *Context, gtx layout.Context) layout.Dimensions {
	cell := gtx.Dp(ctx.Theme.Components.DatePicker.CellSize)
	return layout.Flex{Axis: layout.Horizontal}.Layout(gtx, datePickerWeekdayChildren(ctx, cell, orderedDatePickerWeekdays(d.locale))...)
}

func datePickerWeekdayChildren(ctx *Context, cell int, weekdays [7]string) []layout.FlexChild {
	children := make([]layout.FlexChild, 0, len(weekdays))
	for _, weekday := range weekdays {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints = layout.Exact(image.Pt(cell, cell))
			return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return Text(weekday).
					Size(float32(ctx.Theme.Components.DatePicker.WeekdayTextSize)).
					Weight(font.Medium).
					Color(ctx.Theme.Palette.MutedForeground).
					Layout(ctx, gtx)
			})
		}))
	}
	return children
}

func (d DatePickerWidget) layoutDays(ctx *Context, gtx layout.Context, state *datePickerState, now time.Time) layout.Dimensions {
	days := datePickerMonthDays(state.viewMonth, d.locale.WeekStart)
	cell := gtx.Dp(ctx.Theme.Components.DatePicker.CellSize)
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx, datePickerWeekRows(ctx, d, state, days, cell, now)...)
}

func (d DatePickerWidget) layoutMonths(ctx *Context, gtx layout.Context, state *datePickerState, now time.Time) layout.Dimensions {
	year, _, _ := state.viewMonth.Date()
	theme := ctx.Theme.Components.DatePicker
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

func (d DatePickerWidget) layoutMonth(ctx *Context, gtx layout.Context, state *datePickerState, year int, month time.Month, now time.Time) layout.Dimensions {
	date := time.Date(year, month, 1, 0, 0, 0, 0, state.viewMonth.Location())
	key := fmt.Sprintf("%04d-%02d", year, int(month))
	monthState := state.month(key)
	disabled := d.isMonthDisabled(date)
	presses := activePresses(monthState.clickable.History())
	if disabled {
		gtx = gtx.Disabled()
	} else {
		for monthState.clickable.Clicked(gtx) {
			state.viewMonth = date
			state.viewMode = datePickerViewDays
			ctx.requestFocus(&state.trigger)
		}
		ctx.focusOnPress(&state.trigger, monthState.clickable.History(), presses)
	}

	valueYear, valueMonth, _ := d.value.Date()
	selected := !d.value.IsZero() && valueYear == year && valueMonth == month
	today := now.Year() == year && now.Month() == month
	active := state.viewMonth.Year() == year && state.viewMonth.Month() == month
	return d.layoutPickerCell(ctx, gtx, monthState, d.locale.Months[int(month)-1], selected, active, today, false, disabled)
}

func (d DatePickerWidget) layoutYears(ctx *Context, gtx layout.Context, state *datePickerState, now time.Time) layout.Dimensions {
	minYear, maxYear := d.yearPickerRange(state, now)
	count := maxYear - minYear + 1
	if count <= 0 {
		return layout.Dimensions{}
	}

	theme := ctx.Theme.Components.DatePicker
	rows := (count + 2) / 3
	cellWidth := gtx.Dp(theme.CalendarWidth) / 3
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
	return state.yearList.Layout(gtx, rows, func(gtx layout.Context, row int) layout.Dimensions {
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

func (d DatePickerWidget) layoutYearCell(ctx *Context, gtx layout.Context, state *datePickerState, year, maxYear int, now time.Time) layout.Dimensions {
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

func (d DatePickerWidget) layoutYear(ctx *Context, gtx layout.Context, state *datePickerState, year int, now time.Time) layout.Dimensions {
	yearState := state.year(fmt.Sprintf("%04d", year))
	disabled := d.isYearDisabled(year)
	presses := activePresses(yearState.clickable.History())
	if disabled {
		gtx = gtx.Disabled()
	} else {
		for yearState.clickable.Clicked(gtx) {
			_, month, _ := state.viewMonth.Date()
			state.viewMonth = time.Date(year, month, 1, 0, 0, 0, 0, state.viewMonth.Location())
			state.viewMode = datePickerViewDays
			ctx.requestFocus(&state.trigger)
		}
		ctx.focusOnPress(&state.trigger, yearState.clickable.History(), presses)
	}

	selected := !d.value.IsZero() && d.value.Year() == year
	today := now.Year() == year
	active := state.viewMonth.Year() == year
	return d.layoutPickerCell(ctx, gtx, yearState, d.locale.YearLabel(year), selected, active, today, false, disabled)
}

func (d DatePickerWidget) layoutPickerCell(ctx *Context, gtx layout.Context, state *datePickerCellState, label string, selected, active, today, outside, disabled bool) layout.Dimensions {
	return state.clickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		size := gtx.Constraints.Max
		style := datePickerCellStyleFor(ctx.Theme, state.clickable.Hovered(), state.clickable.Pressed(), selected, today, outside, disabled)
		if active && !selected && !disabled {
			style.bg = ctx.Theme.Palette.SurfaceRaised
			if !today {
				style.fg = ctx.Theme.Palette.AccentSoftForeground
			}
		}
		style.bg = state.background(gtx, style.bg)
		scale := datePickerPressScale(gtx, state.clickable.History(), disabled)
		stack := buttonScale(size, scale).Push(gtx.Ops)
		drawDatePickerCell(gtx, size, style)
		layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return Text(label).
				Size(float32(ctx.Theme.Components.DatePicker.CellTextSize)).
				Weight(font.Medium).
				Color(style.fg).
				Layout(ctx, gtx)
		})
		stack.Pop()
		return layout.Dimensions{Size: size}
	})
}

func datePickerWeekRows(ctx *Context, picker DatePickerWidget, state *datePickerState, days []datePickerDay, cell int, now time.Time) []layout.FlexChild {
	rows := make([]layout.FlexChild, 0, 6)
	for week := range 6 {
		weekDays := days[week*7 : week*7+7]
		rows = append(rows, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Horizontal}.Layout(gtx, datePickerDayChildren(ctx, picker, state, weekDays, cell, now)...)
		}))
	}
	return rows
}

func datePickerDayChildren(ctx *Context, picker DatePickerWidget, state *datePickerState, days []datePickerDay, cell int, now time.Time) []layout.FlexChild {
	children := make([]layout.FlexChild, 0, len(days))
	for _, day := range days {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints = layout.Exact(image.Pt(cell, cell))
			return picker.layoutDay(ctx, gtx, state, day, now)
		}))
	}
	return children
}

func (d DatePickerWidget) layoutDay(ctx *Context, gtx layout.Context, state *datePickerState, day datePickerDay, now time.Time) layout.Dimensions {
	key := dateKey(day.date)
	dayState := state.day(key)
	disabled := d.isDateDisabled(day.date)
	presses := activePresses(dayState.clickable.History())
	if disabled {
		gtx = gtx.Disabled()
	} else {
		for dayState.clickable.Clicked(gtx) {
			date := dateOnly(day.date)
			state.open = false
			state.sync(date, date)
			if d.onChange != nil {
				d.onChange(date)
			}
			ctx.requestFocus(&state.trigger)
		}
		ctx.focusOnPress(&state.trigger, dayState.clickable.History(), presses)
	}

	selected := sameDate(d.value, day.date)
	today := sameDate(now, day.date)
	return dayState.clickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		size := gtx.Constraints.Max
		style := datePickerCellStyleFor(ctx.Theme, dayState.clickable.Hovered(), dayState.clickable.Pressed(), selected, today, day.outside, disabled)
		style.bg = dayState.background(gtx, style.bg)
		scale := datePickerPressScale(gtx, dayState.clickable.History(), disabled)
		stack := buttonScale(size, scale).Push(gtx.Ops)
		drawDatePickerCell(gtx, size, style)
		layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return Text(fmt.Sprintf("%d", day.date.Day())).
				Size(float32(ctx.Theme.Components.DatePicker.CellTextSize)).
				Weight(font.Medium).
				Color(style.fg).
				Layout(ctx, gtx)
		})
		if disabled && !day.outside {
			drawDatePickerStrike(gtx, ctx.Theme, size, style.fg)
		}
		stack.Pop()
		return layout.Dimensions{Size: size}
	})
}
