package datepicker

import "time"

type Day struct {
	Date    time.Time
	Outside bool
}

func MonthDays(month time.Time, weekStart time.Weekday) []Day {
	month = FirstOfMonth(month)
	offset := (int(month.Weekday()) - int(weekStart) + 7) % 7
	start := month.AddDate(0, 0, -offset)
	days := make([]Day, 42)
	for i := range days {
		date := start.AddDate(0, 0, i)
		days[i] = Day{
			Date:    date,
			Outside: date.Month() != month.Month(),
		}
	}
	return days
}

func YearRangeStart(year, span int) int {
	return year - year%span
}

func FirstOfMonth(date time.Time) time.Time {
	date = DateOnly(date)
	if date.IsZero() {
		date = DateOnly(time.Now())
	}
	year, month, _ := date.Date()
	return time.Date(year, month, 1, 0, 0, 0, 0, date.Location())
}
