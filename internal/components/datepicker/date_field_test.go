package datepicker

import (
	"image"
	"testing"
	"time"

	"gioui.org/f32"
	"gioui.org/io/input"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/flowui/internal/field"
	"github.com/qianniancn/flowui/internal/frame"
	"github.com/qianniancn/flowui/internal/locale"
)

func TestDateFieldOptionsUseValueSemantics(t *testing.T) {
	base := DateField("date", time.Time{})
	configured := base.
		Label("Date").
		Description("Choose a date").
		ErrorMessage("Invalid date").
		Locale(DatePickerChinese()).
		Variant(field.Secondary).
		Disabled(true).
		Invalid(true).
		Required(true).
		FullWidth().
		MinDate(testDate(2026, 1, 1)).
		MaxDate(testDate(2026, 12, 31)).
		OnChange(func(time.Time) {})

	if base.label != "" || base.description != "" || base.disabled || base.fullWidth || base.onChange != nil {
		t.Fatalf("configuring DateField mutated base: %#v", base)
	}
	if configured.label != "Date" || configured.description == "" || configured.errorMessage == "" || !configured.localeSet || configured.locale.DateOrder[0] != DatePartYear || !configured.disabled || !configured.invalid || !configured.required || !configured.fullWidth || configured.onChange == nil {
		t.Fatalf("configured DateField = %#v", configured)
	}
}

func TestDatePickerTypingUsesSharedDateSegments(t *testing.T) {
	ctx := newContextWithThemeAndLanguage(nil, nil, locale.LanguageEnglish)
	router := new(input.Router)
	value := testDate(2026, 7, 9)
	picker := func() DatePickerWidget {
		return DatePicker("date", value).OnChange(func(next time.Time) { value = next })
	}
	now := testDate(2026, 7, 1)
	layoutDatePickerFrameAt(ctx, router, picker(), now)
	componentState := testComponentState[datePickerState](ctx, "date", stateSlotDatePicker)
	router.Source().Execute(key.FocusCmd{Tag: &componentState.segments.segments[DatePartMonth].clickable})
	layoutDatePickerFrameAt(ctx, router, picker(), now.Add(time.Millisecond))
	router.Queue(
		key.Event{Name: key.Name("1"), State: key.Press},
		key.Event{Name: key.Name("2"), State: key.Press},
	)
	layoutDatePickerFrameAt(ctx, router, picker(), now.Add(2*time.Millisecond))

	if !sameDate(value, testDate(2026, 12, 9)) {
		t.Fatalf("typed DatePicker month = %v, want 2026-12-09", value)
	}
}

func TestDateFieldArrowNavigationUsesLocaleOrder(t *testing.T) {
	ctx := newContextWithThemeAndLanguage(nil, nil, locale.LanguageEnglish)
	router := new(input.Router)
	widget := DateField("date", testDate(2026, 7, 9))
	now := testDate(2026, 7, 1)
	layoutDateComponentFrame(ctx, router, widget, now)
	componentState := testComponentState[dateFieldState](ctx, "date", stateSlotDateField)
	router.Source().Execute(key.FocusCmd{Tag: &componentState.segments.segments[DatePartMonth].clickable})
	layoutDateComponentFrame(ctx, router, widget, now.Add(time.Millisecond))
	router.Queue(key.Event{Name: key.NameRightArrow, State: key.Press})
	layoutDateComponentFrame(ctx, router, widget, now.Add(2*time.Millisecond))
	layoutDateComponentFrame(ctx, router, widget, now.Add(3*time.Millisecond))

	if !router.Source().Focused(&componentState.segments.segments[DatePartDay].clickable) {
		t.Fatal("English right arrow did not move from month to day")
	}
}

func TestDateFieldPointerSegmentFocusIsNotKeyboardVisible(t *testing.T) {
	ctx := newContextWithThemeAndLanguage(nil, nil, locale.LanguageEnglish)
	router := new(input.Router)
	widget := DateField("date", testDate(2026, 7, 9)).FullWidth()
	start := testDate(2026, 7, 1)
	layoutDateComponentFrame(ctx, router, widget, start)
	state := testComponentState[dateFieldState](ctx, "date", stateSlotDateField)
	month := &state.segments.segments[DatePartMonth].clickable

	router.Queue(pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: 1, Buttons: pointer.ButtonPrimary, Position: f32.Pt(20, 18)})
	layoutDateComponentFrame(ctx, router, widget, start.Add(time.Millisecond))
	router.Queue(pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, PointerID: 1, Position: f32.Pt(20, 18)})
	layoutDateComponentFrame(ctx, router, widget, start.Add(2*time.Millisecond))
	layoutDateComponentFrame(ctx, router, widget, start.Add(3*time.Millisecond))

	if !router.Source().Focused(month) {
		t.Fatal("pointer click did not focus the date segment")
	}
	if frame.FocusVisible(ctx, month, true) {
		t.Fatal("pointer click exposed the date segment keyboard focus ring")
	}
}

func TestDateFieldKeyboardChangesMonth(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	value := testDate(2026, 7, 9)
	fieldWidget := func() DateFieldWidget {
		return DateField("date", value).OnChange(func(next time.Time) { value = next })
	}
	now := testDate(2026, 7, 1)
	layoutDateComponentFrame(ctx, router, fieldWidget(), now)
	componentState := testComponentState[dateFieldState](ctx, "date", stateSlotDateField)
	router.Source().Execute(key.FocusCmd{Tag: &componentState.segments.segments[1].clickable})
	layoutDateComponentFrame(ctx, router, fieldWidget(), now.Add(time.Millisecond))
	router.Queue(key.Event{Name: key.NameUpArrow, State: key.Press})
	layoutDateComponentFrame(ctx, router, fieldWidget(), now.Add(2*time.Millisecond))

	if !sameDate(value, testDate(2026, 8, 9)) {
		t.Fatalf("month increment = %v, want 2026-08-09", value)
	}
}

func TestDateFieldTypingReplacesSegment(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	value := testDate(2026, 7, 9)
	fieldWidget := func() DateFieldWidget {
		return DateField("date", value).OnChange(func(next time.Time) { value = next })
	}
	now := testDate(2026, 7, 1)
	layoutDateComponentFrame(ctx, router, fieldWidget(), now)
	componentState := testComponentState[dateFieldState](ctx, "date", stateSlotDateField)
	router.Source().Execute(key.FocusCmd{Tag: &componentState.segments.segments[1].clickable})
	layoutDateComponentFrame(ctx, router, fieldWidget(), now.Add(time.Millisecond))
	router.Queue(
		key.Event{Name: key.Name("1"), State: key.Press},
		key.Event{Name: key.Name("2"), State: key.Press},
	)
	layoutDateComponentFrame(ctx, router, fieldWidget(), now.Add(2*time.Millisecond))

	if !sameDate(value, testDate(2026, 12, 9)) {
		t.Fatalf("typed month = %v, want 2026-12-09", value)
	}
}

func TestDateFieldRejectsInvalidCalendarDate(t *testing.T) {
	segments := new(dateSegmentsState)
	segments.sync(testDate(2026, 2, 28))
	segments.parts.day = 31
	changed := false
	segments.dispatch(time.Time{}, time.Time{}, func(time.Time) { changed = true })

	if segments.valid || changed {
		t.Fatal("invalid February date was dispatched")
	}
}

func TestDateConstraintsRejectValuesOutsideBounds(t *testing.T) {
	minimum := testDate(2026, 7, 10)
	maximum := testDate(2026, 7, 20)
	if !dateOutsideRange(testDate(2026, 7, 9), minimum, maximum) ||
		!dateOutsideRange(testDate(2026, 7, 21), minimum, maximum) ||
		dateOutsideRange(testDate(2026, 7, 15), minimum, maximum) ||
		dateOutsideRange(time.Time{}, minimum, maximum) {
		t.Fatal("date bounds validation is inconsistent")
	}
	if !invalidDateRange(DateRange{Start: minimum, End: testDate(2026, 7, 21)}, minimum, maximum) {
		t.Fatal("date range did not validate its endpoints")
	}
}

func TestDateRangePickerOrdersReverseCalendarSelection(t *testing.T) {
	componentState := new(dateRangePickerState)
	componentState.calendar.open = true
	componentState.sync(DateRange{}, testDate(2026, 7, 1))
	var changed DateRange
	changes := 0
	callback := func(value DateRange) {
		changed = value
		changes++
	}

	componentState.selectDate(testDate(2026, 7, 20), callback)
	if !componentState.selectingEnd || !componentState.calendar.open || changes != 0 {
		t.Fatal("first range date did not start a pending selection")
	}
	componentState.selectDate(testDate(2026, 7, 10), callback)
	if changes != 1 || !sameDate(changed.Start, testDate(2026, 7, 10)) || !sameDate(changed.End, testDate(2026, 7, 20)) {
		t.Fatalf("reverse selection = %#v", changed)
	}
	if componentState.calendar.open || componentState.selectingEnd {
		t.Fatal("completed range selection stayed open")
	}
}

func TestDateRangePickerPreviewsPendingRangeAtHoveredDate(t *testing.T) {
	componentState := new(dateRangePickerState)
	componentState.sync(DateRange{}, testDate(2026, 7, 1))
	componentState.selectDate(testDate(2026, 7, 20), nil)
	start, end := componentState.displayRange(DateRange{}, testDate(2026, 7, 10))

	if !sameDate(start, testDate(2026, 7, 20)) || !sameDate(end, testDate(2026, 7, 10)) {
		t.Fatalf("preview range = %v - %v", start, end)
	}
}

func TestDateRangePickerCancelsPendingSelection(t *testing.T) {
	original := DateRange{Start: testDate(2026, 7, 1), End: testDate(2026, 7, 4)}
	componentState := new(dateRangePickerState)
	componentState.sync(original, original.Start)
	componentState.selectDate(testDate(2026, 8, 20), nil)
	componentState.cancelSelection(original)

	if componentState.selectingEnd || !componentState.pendingStart.IsZero() {
		t.Fatal("pending range selection was not cancelled")
	}
	start, _, _ := componentState.start.date()
	end, _, _ := componentState.end.date()
	if !sameDate(start, original.Start) || !sameDate(end, original.End) {
		t.Fatalf("cancelled range = %v - %v, want %#v", start, end, original)
	}
	if componentState.calendar.viewMonth.Month() != original.Start.Month() {
		t.Fatalf("cancelled calendar month = %v, want %v", componentState.calendar.viewMonth, original.Start)
	}
}

func TestDateRangePickerCalendarSelectsTwoDates(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	value := DateRange{Start: testDate(2026, 7, 1), End: testDate(2026, 7, 4)}
	picker := func() DateRangePickerWidget {
		return DateRangePicker("trip", value).OnChange(func(next DateRange) { value = next })
	}
	now := testDate(2026, 7, 1)
	layoutDateComponentFrame(ctx, router, picker(), now)
	componentState := testComponentState[dateRangePickerState](ctx, "trip", stateSlotDateRangePicker)
	componentState.calendar.trigger.Click()
	layoutDateComponentFrame(ctx, router, picker(), now.Add(time.Millisecond))

	first := componentState.calendar.days["2026-07-10"]
	if first == nil {
		t.Fatal("missing first range calendar day")
	}
	first.clickable.Click()
	layoutDateComponentFrame(ctx, router, picker(), now.Add(2*time.Millisecond))
	if !componentState.calendar.open || !componentState.selectingEnd {
		t.Fatal("range picker closed after selecting the start")
	}

	second := componentState.calendar.days["2026-07-15"]
	if second == nil {
		t.Fatal("missing second range calendar day")
	}
	second.clickable.Click()
	layoutDateComponentFrame(ctx, router, picker(), now.Add(3*time.Millisecond))
	if !sameDate(value.Start, testDate(2026, 7, 10)) || !sameDate(value.End, testDate(2026, 7, 15)) {
		t.Fatalf("selected range = %#v", value)
	}
	if componentState.calendar.open {
		t.Fatal("range picker stayed open after selecting the end")
	}
}

func TestDateRangePickerClosingCalendarRestoresControlledRange(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	original := DateRange{Start: testDate(2026, 7, 1), End: testDate(2026, 7, 4)}
	picker := DateRangePicker("trip", original)
	now := testDate(2026, 7, 1)
	layoutDateComponentFrame(ctx, router, picker, now)
	componentState := testComponentState[dateRangePickerState](ctx, "trip", stateSlotDateRangePicker)
	componentState.calendar.trigger.Click()
	layoutDateComponentFrame(ctx, router, picker, now.Add(time.Millisecond))

	first := componentState.calendar.days["2026-07-10"]
	if first == nil {
		t.Fatal("missing pending range calendar day")
	}
	first.clickable.Click()
	layoutDateComponentFrame(ctx, router, picker, now.Add(2*time.Millisecond))
	componentState.calendar.trigger.Click()
	layoutDateComponentFrame(ctx, router, picker, now.Add(3*time.Millisecond))

	start, _, _ := componentState.start.date()
	end, _, _ := componentState.end.date()
	if componentState.calendar.open || componentState.selectingEnd ||
		!sameDate(start, original.Start) || !sameDate(end, original.End) {
		t.Fatalf("closed range picker retained pending value: %v - %v", start, end)
	}
}

func TestDateRangePickerTriggerFocusDoesNotFocusDateSegments(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	picker := DateRangePicker("trip", DateRange{})
	now := testDate(2026, 7, 1)
	layoutDateComponentFrame(ctx, router, picker, now)
	componentState := testComponentState[dateRangePickerState](ctx, "trip", stateSlotDateRangePicker)
	router.Source().Execute(key.FocusCmd{Tag: &componentState.calendar.trigger})
	layoutDateComponentFrame(ctx, router, picker, now.Add(time.Millisecond))

	gtx := layout.Context{Source: router.Source()}
	if !componentState.focused(gtx) || componentState.inputFocused(gtx) {
		t.Fatal("calendar trigger focus was treated as date segment focus")
	}
}

func TestDateSegmentsScrollWhenWidthIsConstrained(t *testing.T) {
	ctx := newContextWithThemeAndLanguage(nil, nil, locale.LanguageEnglish)
	router := new(input.Router)
	dimensions := layoutDateComponentFrameWithViewport(ctx, router, DateField("date", testDate(2026, 7, 9)).FullWidth(), testDate(2026, 7, 1), image.Pt(80, 100))
	componentState := testComponentState[dateFieldState](ctx, "date", stateSlotDateField)

	if dimensions.Size.X != 80 {
		t.Fatalf("constrained DateField width = %d, want 80", dimensions.Size.X)
	}
	if componentState.segments.list.Position.Length <= dimensions.Size.X-24 {
		t.Fatalf("date segments length = %d, want scrollable content", componentState.segments.list.Position.Length)
	}
}

func TestDateSegmentPlaceholderExpandsPastThemeMinimum(t *testing.T) {
	ctx := newContext(nil)
	gtx := testLayoutContext()
	segments := new(dateSegmentsState)
	segments.sync(time.Time{})
	style := field.Colors{
		Foreground:  frame.ActiveTheme(ctx).Palette.FieldForegroundColor(),
		Placeholder: frame.ActiveTheme(ctx).Palette.FieldPlaceholderColor(),
		Selection:   frame.ActiveTheme(ctx).Palette.Selection,
	}
	dimensions := segments.layoutSegment(ctx, gtx, int(DatePartMonth), style, true, false)
	minimum := gtx.Dp(frame.ActiveTheme(ctx).Components.DatePicker.SegmentWidth)

	if dimensions.Size.X <= minimum {
		t.Fatalf("month placeholder width = %d, want more than theme minimum %d", dimensions.Size.X, minimum)
	}
}

func TestDateBetweenIncludesRangeEndpoints(t *testing.T) {
	start := testDate(2026, 7, 10)
	end := testDate(2026, 7, 20)
	if !dateBetween(start, start, end) || !dateBetween(end, start, end) || !dateBetween(testDate(2026, 7, 15), end, start) {
		t.Fatal("dateBetween did not include or normalize the range")
	}
	if dateBetween(testDate(2026, 7, 21), start, end) {
		t.Fatal("date outside the range was included")
	}
}

func layoutDateComponentFrame(ctx *frame.Context, router *input.Router, widget frame.Widget, now time.Time) layout.Dimensions {
	return layoutDateComponentFrameWithViewport(ctx, router, widget, now, image.Pt(500, 600))
}

func layoutDateComponentFrameWithViewport(ctx *frame.Context, router *input.Router, widget frame.Widget, now time.Time, viewport image.Point) layout.Dimensions {
	var operations op.Ops
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: viewport},
		Source:      router.Source(),
		Ops:         &operations,
		Now:         now,
	}
	frame.BeginFrameWithViewport(ctx, viewport)
	dimensions := widget.Layout(ctx, gtx)
	frame.LayoutOverlays(ctx, gtx)
	frame.ApplyFrameCommands(ctx, gtx)
	frame.EndFrame(ctx)
	router.Frame(&operations)
	return dimensions
}
