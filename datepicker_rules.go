package flowui

import "time"

func (d DatePickerWidget) canMoveMonth(month time.Time, delta int) bool {
	next := firstOfMonth(month.AddDate(0, delta, 0))
	if delta < 0 && !d.minDate.IsZero() {
		return compareDate(next, firstOfMonth(d.minDate)) >= 0
	}
	if delta > 0 && !d.maxDate.IsZero() {
		return compareDate(next, firstOfMonth(d.maxDate)) <= 0
	}
	return true
}

func (d DatePickerWidget) canMove(state *datePickerState, delta int) bool {
	switch state.viewMode {
	case datePickerViewMonths:
		year := state.viewMonth.Year() + delta
		return !d.isYearDisabled(year)
	case datePickerViewYears:
		year := state.viewMonth.Year()
		start := datePickerYearRangeStart(year) + delta*datePickerYearSpan
		return !d.isYearRangeDisabled(start, start+datePickerYearSpan-1)
	default:
		return d.canMoveMonth(state.viewMonth, delta)
	}
}

func (d DatePickerWidget) isDateDisabled(date time.Time) bool {
	date = dateOnly(date)
	if !d.minDate.IsZero() && compareDate(date, d.minDate) < 0 {
		return true
	}
	if !d.maxDate.IsZero() && compareDate(date, d.maxDate) > 0 {
		return true
	}
	return false
}

func (d DatePickerWidget) isMonthDisabled(date time.Time) bool {
	start := firstOfMonth(date)
	end := start.AddDate(0, 1, -1)
	if !d.minDate.IsZero() && compareDate(end, d.minDate) < 0 {
		return true
	}
	if !d.maxDate.IsZero() && compareDate(start, d.maxDate) > 0 {
		return true
	}
	return false
}

func (d DatePickerWidget) isYearDisabled(year int) bool {
	if !d.minDate.IsZero() && year < d.minDate.Year() {
		return true
	}
	if !d.maxDate.IsZero() && year > d.maxDate.Year() {
		return true
	}
	return false
}

func (d DatePickerWidget) isYearRangeDisabled(start, end int) bool {
	for year := start; year <= end; year++ {
		if !d.isYearDisabled(year) {
			return false
		}
	}
	return true
}
