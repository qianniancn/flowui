package datepicker

import (
	"image"
	"time"

	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/field"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

type Widget struct {
	override     *theme.DatePickerOverride
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
	locale       Locale
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
}

const (
	datePickerYearSpan = 12
)

type datePickerViewMode uint8

const (
	datePickerViewDays datePickerViewMode = iota
	datePickerViewMonths
	datePickerViewYears
)

func English() Locale {
	return datePickerEnglish()
}

func Chinese() Locale {
	return datePickerChinese()
}

func New(key string, value time.Time) Widget {
	return Widget{
		key:      key,
		value:    dateOnly(value),
		hasValue: true,
		editable: true,
	}
}

// Uncontrolled uses the current date only as the initial internal value.
func (d Widget) Uncontrolled() Widget {
	d.defaultValue = d.value
	d.hasDefault = true
	d.hasValue = false
	return d
}

func (d Widget) Value(value time.Time) Widget {
	d.value = dateOnly(value)
	d.hasValue = true
	return d
}

func (d Widget) DefaultValue(value time.Time) Widget {
	d.defaultValue = dateOnly(value)
	d.hasDefault = true
	d.hasValue = false
	return d
}

func (d Widget) Hint(hint string) Widget {
	d.hint = hint
	d.hintSet = true
	return d
}

func (d Widget) Label(value string) Widget {
	d.label = value
	return d
}

func (d Widget) Description(value string) Widget {
	d.description = value
	return d
}

func (d Widget) ErrorMessage(value string) Widget {
	d.errorMessage = value
	return d
}

func (d Widget) Locale(locale Locale) Widget {
	d.locale = normalizeDatePickerLocale(locale)
	d.localeSet = true
	if !d.hintSet {
		d.hint = d.locale.Hint
	}
	return d
}

func (d Widget) OnChange(fn func(time.Time)) Widget {
	d.onChange = fn
	return d
}

func (d Widget) Disabled(disabled bool) Widget {
	d.disabled = disabled
	return d
}

func (d Widget) Editable(editable bool) Widget {
	d.editable = editable
	return d
}

func (d Widget) Invalid(invalid bool) Widget {
	d.invalid = invalid
	return d
}

func (d Widget) Required(value bool) Widget {
	d.required = value
	return d
}

func (d Widget) Variant(variant field.Variant) Widget {
	d.variant = variant
	return d
}

func (d Widget) FullWidth() Widget {
	d.fullWidth = true
	return d
}

func (d Widget) MinDate(date time.Time) Widget {
	d.minDate = dateOnly(date)
	return d
}

func (d Widget) MaxDate(date time.Time) Widget {
	d.maxDate = dateOnly(date)
	return d
}

func (d Widget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	return d.layoutWithOverride(ctx, gtx, d.override)
}

func (d Widget) layoutWithOverride(ctx *core.Context, gtx layout.Context, override *theme.DatePickerOverride) layout.Dimensions {
	core.RequireVisualProvider(ctx)
	state := datePickerStateFor(ctx, d.key)
	activeTheme := core.ActiveTheme(ctx)
	visual := state.partDefaults.visualForRequest(activeTheme, d.variant, override)
	restoreOverride := core.PushInstanceStyle(ctx, visual.Override)
	defer restoreOverride()
	d = d.resolveLocale(ctx, &state.localeCache)
	now := datePickerFrameNow(gtx.Now)

	// Bind disclosure and get current value
	state.bind(d)
	currentValue := state.currentValue(d)

	// Use currentValue for all subsequent operations
	d.value = currentValue

	naturallyDisabled := core.OverlayNaturallyDisabled(gtx)
	state.beginFrame()
	state.resolvedParts.prepare(ctx, core.InstanceStyle(ctx), d.variant)
	state.sync(currentValue, d.initialMonth(now))
	metrics := visual.Metrics
	state.calendarStyle = visual.Calendar
	state.hover.update(gtx)
	enabled := gtx.Enabled() && !d.disabled
	if d.editable {
		core.RegisterFieldFocus(ctx, core.FullKey(ctx, d.key), &state.segments.segments[d.locale.DateOrder[0]].clickable, enabled)
	} else {
		core.RegisterFieldFocus(ctx, core.FullKey(ctx, d.key), &state.trigger, enabled)
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
	styleState := style.State{
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
		fieldState.FocusVisible = core.FocusVisible(ctx, &state.trigger, triggerFocused)
	}
	resolved := resolveDatePickerField(ctx, gtx, core.FullKey(ctx, d.key), fieldState, style.Style{}, visual)
	var dims layout.Dimensions
	var anchor image.Rectangle
	dims = layoutui.Styled(ctx, gtx, core.FullKey(ctx, d.key), styleState, style.Style{}, core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		dims, anchor = d.layoutField(ctx, gtx, state, resolved, metrics, enabled, invalid)
		return dims
	}))

	motion := state.partDefaults.motion(core.ActiveTheme(ctx))
	progress := state.popoverProgress(gtx, state.open && enabled, motion)
	if progress == 0 && (!state.open || !enabled) {
		state.endFrame()
		return dims
	}
	d.layoutPopover(ctx, gtx, state, anchor, progress, now, naturallyDisabled)
	core.AfterOverlays(ctx, state.endFrame)
	return dims
}

func (d Widget) resolveLocale(ctx *core.Context, caches ...*datePickerLocaleCache) Widget {
	if !d.localeSet {
		if len(caches) > 0 && caches[0] != nil {
			d.locale = caches[0].resolve(ctx)
		} else {
			d.locale = datePickerLocaleForContext(ctx, core.ActiveLanguage(ctx))
		}
	}
	if !d.hintSet {
		d.hint = d.locale.Hint
	}
	return d
}

func (d Widget) initialMonth(now time.Time) time.Time {
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

// Style merges instance declarations into the widget. The declaration
// reaches the provider at layout time, so Style can sit anywhere in the chain.
func (w Widget) Style(values ...style.Style) Widget {
	declaration := style.Join(values...)
	if w.override == nil {
		w.override = &theme.DatePickerOverride{}
	}
	w.override.Declaration = style.Join(w.override.Declaration, declaration)
	return w
}
