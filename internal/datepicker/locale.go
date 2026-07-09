package datepicker

import (
	"fmt"
	"time"
)

type Locale struct {
	Hint           string
	Weekdays       [7]string
	Months         [12]string
	WeekStart      time.Weekday
	DateLabel      func(time.Time) string
	MonthLabel     func(time.Time) string
	YearLabel      func(int) string
	YearRangeLabel func(int, int) string
}

func English() Locale {
	return Locale{
		Hint:      "YYYY / MM / DD",
		Weekdays:  [7]string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"},
		Months:    [12]string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"},
		WeekStart: time.Sunday,
		DateLabel: func(date time.Time) string {
			year, month, day := date.Date()
			return fmt.Sprintf("%04d / %02d / %02d", year, int(month), day)
		},
		MonthLabel: func(date time.Time) string {
			return date.Format("January 2006")
		},
		YearLabel: func(year int) string {
			return fmt.Sprintf("%04d", year)
		},
		YearRangeLabel: func(start, end int) string {
			return fmt.Sprintf("%04d - %04d", start, end)
		},
	}
}

func Chinese() Locale {
	return Locale{
		Hint:      "请选择日期",
		Weekdays:  [7]string{"日", "一", "二", "三", "四", "五", "六"},
		Months:    [12]string{"1月", "2月", "3月", "4月", "5月", "6月", "7月", "8月", "9月", "10月", "11月", "12月"},
		WeekStart: time.Monday,
		DateLabel: func(date time.Time) string {
			year, month, day := date.Date()
			return fmt.Sprintf("%04d年%d月%d日", year, int(month), day)
		},
		MonthLabel: func(date time.Time) string {
			year, month, _ := date.Date()
			return fmt.Sprintf("%04d年%d月", year, int(month))
		},
		YearLabel: func(year int) string {
			return fmt.Sprintf("%04d年", year)
		},
		YearRangeLabel: func(start, end int) string {
			return fmt.Sprintf("%04d年 - %04d年", start, end)
		},
	}
}

func Normalize(locale Locale) Locale {
	fallback := English()
	if locale.Hint == "" {
		locale.Hint = fallback.Hint
	}
	for i, weekday := range locale.Weekdays {
		if weekday == "" {
			locale.Weekdays[i] = fallback.Weekdays[i]
		}
	}
	for i, month := range locale.Months {
		if month == "" {
			locale.Months[i] = fallback.Months[i]
		}
	}
	if locale.MonthLabel == nil {
		locale.MonthLabel = fallback.MonthLabel
	}
	if locale.YearLabel == nil {
		locale.YearLabel = fallback.YearLabel
	}
	if locale.YearRangeLabel == nil {
		locale.YearRangeLabel = fallback.YearRangeLabel
	}
	if locale.WeekStart < time.Sunday || locale.WeekStart > time.Saturday {
		locale.WeekStart = fallback.WeekStart
	}
	if locale.DateLabel == nil {
		locale.DateLabel = fallback.DateLabel
	}
	return locale
}

func OrderedWeekdays(locale Locale) [7]string {
	var weekdays [7]string
	start := int(locale.WeekStart)
	for i := range weekdays {
		weekdays[i] = locale.Weekdays[(start+i)%7]
	}
	return weekdays
}
