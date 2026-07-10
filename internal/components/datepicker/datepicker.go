package datepicker

import (
	"time"

	"gioui.org/layout"
	"github.com/qianniancn/FlowUI/internal/field"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/locale"
)

type DatePickerWidget struct {
	key       string
	value     time.Time
	hint      string
	hintSet   bool
	locale    DatePickerLocale
	localeSet bool
	onChange  func(time.Time)
	variant   field.Variant
	disabled  bool
	invalid   bool
	fullWidth bool
	minDate   time.Time
	maxDate   time.Time
}

const (
	datePickerYearSpan           = 12
	datePickerPopoverInDuration  = 150 * time.Millisecond
	datePickerPopoverOutDuration = 100 * time.Millisecond
	datePickerCellColorDuration  = 100 * time.Millisecond
	datePickerPressInDuration    = 90 * time.Millisecond
	datePickerPressOutDuration   = 160 * time.Millisecond
)

type datePickerViewMode uint8

const (
	datePickerViewDays datePickerViewMode = iota
	datePickerViewMonths
	datePickerViewYears
)

func DatePickerEnglish() DatePickerLocale {
	return datePickerEnglish()
}

func DatePickerChinese() DatePickerLocale {
	return datePickerChinese()
}

func DatePicker(key string, value time.Time) DatePickerWidget {
	return DatePickerWidget{
		key:   key,
		value: dateOnly(value),
	}
}

func (d DatePickerWidget) Hint(hint string) DatePickerWidget {
	d.hint = hint
	d.hintSet = true
	return d
}

func (d DatePickerWidget) Locale(locale DatePickerLocale) DatePickerWidget {
	d.locale = normalizeDatePickerLocale(locale)
	d.localeSet = true
	if !d.hintSet {
		d.hint = d.locale.Hint
	}
	return d
}

func (d DatePickerWidget) OnChange(fn func(time.Time)) DatePickerWidget {
	d.onChange = fn
	return d
}

func (d DatePickerWidget) Disabled(disabled bool) DatePickerWidget {
	d.disabled = disabled
	return d
}

func (d DatePickerWidget) Invalid(invalid bool) DatePickerWidget {
	d.invalid = invalid
	return d
}

func (d DatePickerWidget) Variant(variant field.Variant) DatePickerWidget {
	d.variant = variant
	return d
}

func (d DatePickerWidget) FullWidth() DatePickerWidget {
	d.fullWidth = true
	return d
}

func (d DatePickerWidget) MinDate(date time.Time) DatePickerWidget {
	d.minDate = dateOnly(date)
	return d
}

func (d DatePickerWidget) MaxDate(date time.Time) DatePickerWidget {
	d.maxDate = dateOnly(date)
	return d
}

func (d DatePickerWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	d = d.resolveLocale(ctx)
	now := datePickerFrameNow(gtx.Now)
	state := datePickerStateFor(ctx, d.key)
	naturallyDisabled := frame.OverlayNaturallyDisabled(gtx)
	eventGtx := gtx
	state.beginFrame()
	state.sync(d.value, d.initialMonth(now))
	if d.disabled {
		state.open = false
		eventGtx = eventGtx.Disabled()
	}

	focused := gtx.Focused(&state.trigger)
	state.updateFocus(focused, d.disabled)
	if !d.disabled && !state.open {
		state.updateKeys(gtx, &state.trigger)
	}

	style := field.ResolveStyle(frame.ActiveTheme(ctx), d.variant, state.trigger.Hovered(), focused || state.open, d.disabled, d.invalid)
	style.Background = state.input.Background(gtx, style.Background)
	style.Border = state.input.BorderColor(gtx, style.Border)
	dims := d.layoutInput(ctx, eventGtx, state, style)

	progress := state.popoverProgress(gtx, state.open && !d.disabled)
	if progress == 0 && (!state.open || d.disabled) {
		state.endFrame()
		return dims
	}
	d.layoutPopover(ctx, eventGtx, state, dims, progress, now, naturallyDisabled)
	frame.AfterOverlays(ctx, state.endFrame)
	return dims
}

func (d DatePickerWidget) resolveLocale(ctx *frame.Context) DatePickerWidget {
	if !d.localeSet {
		if frame.ActiveLanguage(ctx) == locale.LanguageChinese {
			d.locale = datePickerChinese()
		} else {
			d.locale = datePickerEnglish()
		}
	}
	d.locale = normalizeDatePickerLocale(d.locale)
	if !d.hintSet {
		d.hint = d.locale.Hint
	}
	return d
}

func (d DatePickerWidget) initialMonth(now time.Time) time.Time {
	if !d.value.IsZero() {
		return d.value
	}
	initial := dateOnly(now)
	if !d.minDate.IsZero() && compareDate(initial, d.minDate) < 0 {
		return d.minDate
	}
	if !d.maxDate.IsZero() && compareDate(initial, d.maxDate) > 0 {
		return d.maxDate
	}
	return initial
}

func datePickerFrameNow(now time.Time) time.Time {
	if now.IsZero() {
		return time.Now()
	}
	return now
}
