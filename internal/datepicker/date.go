package datepicker

import "time"

func DateOnly(date time.Time) time.Time {
	if date.IsZero() {
		return time.Time{}
	}
	year, month, day := date.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, date.Location())
}

func SameDate(a, b time.Time) bool {
	if a.IsZero() || b.IsZero() {
		return a.IsZero() && b.IsZero()
	}
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	return ay == by && am == bm && ad == bd
}

func CompareDate(a, b time.Time) int {
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
