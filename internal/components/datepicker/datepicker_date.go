package datepicker

import "time"

type datePickerDay struct {
	date    time.Time
	outside bool
}

func datePickerMonthDays(month time.Time, weekStart time.Weekday) []datePickerDay {
	month = firstOfMonth(month)
	offset := (int(month.Weekday()) - int(weekStart) + 7) % 7
	start := month.AddDate(0, 0, -offset)
	days := make([]datePickerDay, 42)
	for i := range days {
		date := start.AddDate(0, 0, i)
		days[i] = datePickerDay{
			date:    date,
			outside: date.Month() != month.Month(),
		}
	}
	return days
}

func datePickerYearRangeStart(year int) int {
	return year - year%datePickerYearSpan
}

func firstOfMonth(date time.Time) time.Time {
	date = dateOnly(date)
	if date.IsZero() {
		date = dateOnly(time.Now())
	}
	year, month, _ := date.Date()
	return time.Date(year, month, 1, 0, 0, 0, 0, date.Location())
}

func dateOnly(date time.Time) time.Time {
	if date.IsZero() {
		return time.Time{}
	}
	year, month, day := date.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, date.Location())
}

func sameDate(a, b time.Time) bool {
	if a.IsZero() || b.IsZero() {
		return a.IsZero() && b.IsZero()
	}
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	return ay == by && am == bm && ad == bd
}

func compareDate(a, b time.Time) int {
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	if ay != by {
		if ay < by {
			return -1
		}
		return 1
	}
	if am != bm {
		if am < bm {
			return -1
		}
		return 1
	}
	if ad != bd {
		if ad < bd {
			return -1
		}
		return 1
	}
	return 0
}

func dateKey(date time.Time) string {
	return dateOnly(date).Format("2006-01-02")
}
