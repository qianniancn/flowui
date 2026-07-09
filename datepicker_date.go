package flowui

import (
	"time"

	flowdatepicker "github.com/qianniancn/FlowUI/internal/datepicker"
)

type datePickerDay struct {
	date    time.Time
	outside bool
}

func datePickerMonthDays(month time.Time, weekStart time.Weekday) []datePickerDay {
	componentDays := flowdatepicker.MonthDays(month, weekStart)
	days := make([]datePickerDay, len(componentDays))
	for i, day := range componentDays {
		days[i] = datePickerDay{
			date:    day.Date,
			outside: day.Outside,
		}
	}
	return days
}

func datePickerYearRangeStart(year int) int {
	return flowdatepicker.YearRangeStart(year, datePickerYearSpan)
}

func firstOfMonth(date time.Time) time.Time {
	return flowdatepicker.FirstOfMonth(date)
}

func dateOnly(date time.Time) time.Time {
	return flowdatepicker.DateOnly(date)
}

func sameDate(a, b time.Time) bool {
	return flowdatepicker.SameDate(a, b)
}

func compareDate(a, b time.Time) int {
	return flowdatepicker.CompareDate(a, b)
}

func dateKey(date time.Time) string {
	return dateOnly(date).Format("2006-01-02")
}

func normalizeDatePickerLocale(locale DatePickerLocale) DatePickerLocale {
	return flowdatepicker.Normalize(locale)
}

func orderedDatePickerWeekdays(locale DatePickerLocale) [7]string {
	return flowdatepicker.OrderedWeekdays(locale)
}
