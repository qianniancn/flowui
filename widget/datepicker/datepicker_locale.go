package datepicker

import (
	"fmt"
	"time"

	"github.com/qianniancn/flowui/i18n"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
)

type DatePart uint8

const (
	DatePartYear DatePart = iota
	DatePartMonth
	DatePartDay
)

type Locale struct {
	Hint           string
	DateOrder      [3]DatePart
	DateLiterals   [4]string
	Weekdays       [7]string
	Months         [12]string
	WeekStart      time.Weekday
	DateLabel      func(time.Time) string
	MonthLabel     func(time.Time) string
	YearLabel      func(int) string
	YearRangeLabel func(int, int) string
}

type datePickerLocaleCache struct {
	context    *core.Context
	generation uint64
	value      Locale
	valid      bool
}

func (c *datePickerLocaleCache) resolve(ctx *core.Context) Locale {
	generation := core.ThemeGeneration(ctx)
	if c.valid && c.context == ctx && c.generation == generation {
		return c.value
	}
	c.context = ctx
	c.generation = generation
	c.value = datePickerLocaleForContext(ctx, core.ActiveLanguage(ctx))
	c.valid = true
	return c.value
}

var (
	englishDatePickerLocale = newDatePickerEnglish()
	chineseDatePickerLocale = newDatePickerChinese()
)

func datePickerEnglish() Locale {
	return englishDatePickerLocale
}

func datePickerChinese() Locale {
	return chineseDatePickerLocale
}

func newDatePickerEnglish() Locale {
	return Locale{
		Hint:         "MM / DD / YYYY",
		DateOrder:    [3]DatePart{DatePartMonth, DatePartDay, DatePartYear},
		DateLiterals: [4]string{"", "/", "/", ""},
		Weekdays:     calendarWeekdays(i18n.LanguageEnglish),
		Months:       calendarMonths(i18n.LanguageEnglish),
		WeekStart:    time.Sunday,
		DateLabel: func(date time.Time) string {
			year, month, day := date.Date()
			return fmt.Sprintf("%02d / %02d / %04d", int(month), day, year)
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

func newDatePickerChinese() Locale {
	return Locale{
		Hint:         "YYYY年MM月DD日",
		DateOrder:    [3]DatePart{DatePartYear, DatePartMonth, DatePartDay},
		DateLiterals: [4]string{"", "年", "月", "日"},
		Weekdays:     calendarWeekdays(i18n.LanguageChinese),
		Months:       calendarMonths(i18n.LanguageChinese),
		WeekStart:    time.Monday,
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

func calendarMonths(language i18n.Language) [12]string {
	keys := [...]i18n.Key{
		i18n.CalendarMonthJanuary,
		i18n.CalendarMonthFebruary,
		i18n.CalendarMonthMarch,
		i18n.CalendarMonthApril,
		i18n.CalendarMonthMay,
		i18n.CalendarMonthJune,
		i18n.CalendarMonthJuly,
		i18n.CalendarMonthAugust,
		i18n.CalendarMonthSeptember,
		i18n.CalendarMonthOctober,
		i18n.CalendarMonthNovember,
		i18n.CalendarMonthDecember,
	}
	var months [12]string
	for index, key := range keys {
		months[index] = (i18n.Catalog{}).Text(language, key)
	}
	return months
}

func calendarWeekdays(language i18n.Language) [7]string {
	keys := [...]i18n.Key{
		i18n.CalendarWeekdaySunday,
		i18n.CalendarWeekdayMonday,
		i18n.CalendarWeekdayTuesday,
		i18n.CalendarWeekdayWednesday,
		i18n.CalendarWeekdayThursday,
		i18n.CalendarWeekdayFriday,
		i18n.CalendarWeekdaySaturday,
	}
	var weekdays [7]string
	for index, key := range keys {
		weekdays[index] = (i18n.Catalog{}).Text(language, key)
	}
	return weekdays
}

func datePickerLocaleForContext(ctx *core.Context, language locale.Language) Locale {
	value := datePickerEnglish()
	if locale.IsChinese(language) {
		value = datePickerChinese()
	}
	if ctx == nil {
		return value
	}
	value = withCatalogCalendarLabels(ctx, value)
	value.MonthLabel = func(date time.Time) string {
		_, month, _ := date.Date()
		monthName := ctx.Text(calendarFullMonthKey(month))
		return ctx.Textf(i18n.CalendarMonthYear, monthName, date.Year())
	}
	return value
}

func withCatalogCalendarLabels(ctx *core.Context, value Locale) Locale {
	monthKeys := [...]i18n.Key{
		i18n.CalendarMonthJanuary,
		i18n.CalendarMonthFebruary,
		i18n.CalendarMonthMarch,
		i18n.CalendarMonthApril,
		i18n.CalendarMonthMay,
		i18n.CalendarMonthJune,
		i18n.CalendarMonthJuly,
		i18n.CalendarMonthAugust,
		i18n.CalendarMonthSeptember,
		i18n.CalendarMonthOctober,
		i18n.CalendarMonthNovember,
		i18n.CalendarMonthDecember,
	}
	weekdayKeys := [...]i18n.Key{
		i18n.CalendarWeekdaySunday,
		i18n.CalendarWeekdayMonday,
		i18n.CalendarWeekdayTuesday,
		i18n.CalendarWeekdayWednesday,
		i18n.CalendarWeekdayThursday,
		i18n.CalendarWeekdayFriday,
		i18n.CalendarWeekdaySaturday,
	}
	for index, key := range monthKeys {
		value.Months[index] = ctx.Text(key)
	}
	for index, key := range weekdayKeys {
		value.Weekdays[index] = ctx.Text(key)
	}
	return value
}

func calendarFullMonthKey(month time.Month) i18n.Key {
	keys := [...]i18n.Key{
		i18n.CalendarMonthFullJanuary,
		i18n.CalendarMonthFullFebruary,
		i18n.CalendarMonthFullMarch,
		i18n.CalendarMonthFullApril,
		i18n.CalendarMonthFullMay,
		i18n.CalendarMonthFullJune,
		i18n.CalendarMonthFullJuly,
		i18n.CalendarMonthFullAugust,
		i18n.CalendarMonthFullSeptember,
		i18n.CalendarMonthFullOctober,
		i18n.CalendarMonthFullNovember,
		i18n.CalendarMonthFullDecember,
	}
	if month < time.January || month > time.December {
		return ""
	}
	return keys[month-time.January]
}

func normalizeDatePickerLocale(locale Locale) Locale {
	fallback := datePickerEnglish()
	if locale.Hint == "" {
		locale.Hint = fallback.Hint
	}
	if !validDateOrder(locale.DateOrder) {
		locale.DateOrder = fallback.DateOrder
		if locale.DateLiterals == [4]string{} {
			locale.DateLiterals = fallback.DateLiterals
		}
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

func validDateOrder(order [3]DatePart) bool {
	var seen [3]bool
	for _, part := range order {
		if part > DatePartDay || seen[part] {
			return false
		}
		seen[part] = true
	}
	return true
}

func orderedDatePickerWeekdays(locale Locale) [7]string {
	var weekdays [7]string
	start := int(locale.WeekStart)
	for i := range weekdays {
		weekdays[i] = locale.Weekdays[(start+i)%7]
	}
	return weekdays
}
