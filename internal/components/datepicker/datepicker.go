package datepicker

import (
	"image"
	"time"

	"gioui.org/layout"
	layoutui "github.com/qianniancn/FlowUI/internal/components/layout"
	"github.com/qianniancn/FlowUI/internal/field"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/locale"
	flowstyle "github.com/qianniancn/FlowUI/internal/style"
)

type DatePickerWidget struct {
	key          string
	value        time.Time
	hasValue     bool
	defaultValue time.Time
	hasDefault   bool
	hint         string
	hintSet      bool
	label        string
	description  string
	errorMessage string
	locale       DatePickerLocale
	localeSet    bool
	onChange     func(time.Time)
	variant      field.Variant
	disabled     bool
	editable     bool
	invalid      bool
	required     bool
	fullWidth    bool
	minDate      time.Time
	maxDate      time.Time
	rangeMode    bool
	rangeStart   time.Time
	rangeEnd     time.Time
	onDateSelect func(time.Time)
	customStyle  flowstyle.Style
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
		key:      key,
		value:    dateOnly(value),
		hasValue: true,
		editable: true,
	}
}

func (d DatePickerWidget) Value(value time.Time) DatePickerWidget {
	d.value = dateOnly(value)
	d.hasValue = true
	return d
}

func (d DatePickerWidget) DefaultValue(value time.Time) DatePickerWidget {
	d.defaultValue = dateOnly(value)
	d.hasDefault = true
	d.hasValue = false
	return d
}

func (d DatePickerWidget) Hint(hint string) DatePickerWidget {
	d.hint = hint
	d.hintSet = true
	return d
}

func (d DatePickerWidget) Label(value string) DatePickerWidget {
	d.label = value
	return d
}

func (d DatePickerWidget) Description(value string) DatePickerWidget {
	d.description = value
	return d
}

func (d DatePickerWidget) ErrorMessage(value string) DatePickerWidget {
	d.errorMessage = value
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

func (d DatePickerWidget) Editable(editable bool) DatePickerWidget {
	d.editable = editable
	return d
}

func (d DatePickerWidget) Invalid(invalid bool) DatePickerWidget {
	d.invalid = invalid
	return d
}

func (d DatePickerWidget) Required(value bool) DatePickerWidget {
	d.required = value
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

func (d DatePickerWidget) Style(value flowstyle.Style) DatePickerWidget {
	d.customStyle = value
	return d
}

func (d DatePickerWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	d = d.resolveLocale(ctx)
	now := datePickerFrameNow(gtx.Now)
	state := datePickerStateFor(ctx, d.key)

	// Bind disclosure and get current value
	state.bind(d)
	currentValue := state.currentValue(d)

	// Use currentValue for all subsequent operations
	d.value = currentValue

	naturallyDisabled := frame.OverlayNaturallyDisabled(gtx)
	state.beginFrame()
	state.sync(currentValue, d.initialMonth(now))
	state.hover.update(gtx)
	enabled := gtx.Enabled() && !d.disabled
	if d.editable {
		frame.RegisterFieldFocus(ctx, frame.FullKey(ctx, d.key), &state.segments.segments[d.locale.DateOrder[0]].clickable, enabled)
	} else {
		frame.RegisterFieldFocus(ctx, frame.FullKey(ctx, d.key), &state.trigger, enabled)
	}
	if !enabled {
		state.open = false
	}

	inputFocused := d.editable && state.segments.focused(gtx)
	triggerFocused := gtx.Focused(&state.trigger)
	focused := inputFocused || triggerFocused || state.calendarFocused(gtx)
	state.updateFocus(focused, !enabled)
	if state.open && (state.segments.escapePressed(gtx) || state.calendarEscapePressed(gtx)) {
		state.open = false
	}
	if enabled && !state.open {
		state.updateKeys(gtx, &state.trigger)
	}

	invalid := d.invalid || (d.editable && !state.segments.valid) || dateOutsideRange(currentValue, d.minDate, d.maxDate)
	hovered := state.hover.hovered || (d.editable && state.segments.hovered()) || state.trigger.Hovered()
	styleState := flowstyle.StyleState{
		Hovered:      hovered,
		Focused:      focused,
		FocusVisible: state.focusVisible(ctx, gtx),
		Disabled:     !enabled,
		Invalid:      invalid,
		Selected:     !currentValue.IsZero(),
		Open:         state.open,
	}
	fieldState := styleState
	fieldState.Hovered = hovered && !focused
	fieldState.Focused = inputFocused || (!d.editable && triggerFocused)
	fieldState.FocusVisible = state.segments.focusVisible(ctx, gtx)
	if !d.editable {
		fieldState.FocusVisible = frame.FocusVisible(ctx, &state.trigger, triggerFocused)
	}
	tokens := frame.ActiveTheme(ctx).Components
	resolved := field.Resolve(ctx, gtx, frame.FullKey(ctx, d.key), fieldState, d.variant, field.DeclarationOptions{
		Radius:         tokens.DatePicker.Radius,
		FocusRingWidth: tokens.Input.FocusRingWidth, InvalidOutlineWidth: tokens.Input.InvalidOutlineWidth,
		ShadowColor: tokens.Input.ShadowColor, ShadowOpacity: tokens.Input.ShadowOpacity,
		ShadowStrength: tokens.Input.ShadowStrength,
	}, d.customStyle)
	var dims layout.Dimensions
	var anchor image.Rectangle
	dims = layoutui.LayoutStyled(ctx, gtx, frame.FullKey(ctx, d.key), styleState, d.customStyle, frame.WidgetFunc(func(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
		dims, anchor = d.layoutField(ctx, gtx, state, resolved, enabled, invalid)
		return dims
	}))

	progress := state.popoverProgress(gtx, state.open && enabled, frame.ActiveTheme(ctx).Motion)
	if progress == 0 && (!state.open || !enabled) {
		state.endFrame()
		return dims
	}
	d.layoutPopover(ctx, gtx, state, anchor, progress, now, naturallyDisabled)
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
