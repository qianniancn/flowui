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
	"gioui.org/widget"
	"github.com/qianniancn/FlowUI/internal/field"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/locale"
	"github.com/qianniancn/FlowUI/internal/state"
	"github.com/qianniancn/FlowUI/internal/theme"
)

func newContext(_ any) *frame.Context {
	return frame.New(nil, nil, locale.LanguageAuto)
}

func newContextWithThemeAndLanguage(_ any, value *theme.Theme, language locale.Language) *frame.Context {
	return frame.New(nil, value, language)
}

func DefaultTheme() theme.Theme {
	return theme.DefaultTheme()
}

func testComponentState[T any](ctx *frame.Context, key, slot string) *T {
	value, _ := frame.PeekState[T](ctx, key, slot)
	return value
}

func testLayoutContext() layout.Context {
	var router input.Router
	var ops op.Ops
	return layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 200)},
		Source:      router.Source(),
		Ops:         &ops,
	}
}

func TestDatePickerOptions(t *testing.T) {
	value := testDate(2026, 7, 9)
	minDate := testDate(2026, 7, 1)
	maxDate := testDate(2026, 7, 31)
	var changed time.Time

	d := DatePicker("date", value).
		Hint("Pick date").
		Label("Date").
		Description("Choose a date").
		ErrorMessage("Invalid date").
		Variant(field.Secondary).
		Invalid(true).
		Required(true).
		Disabled(true).
		FullWidth().
		Locale(DatePickerChinese()).
		MinDate(minDate).
		MaxDate(maxDate).
		OnChange(func(date time.Time) {
			changed = date
		})

	if d.key != "date" {
		t.Fatalf("key = %q, want date", d.key)
	}
	if d.hint != "Pick date" {
		t.Fatal("hint was not set")
	}
	if d.locale.WeekStart != time.Monday {
		t.Fatal("locale was not set")
	}
	if d.variant != field.Secondary {
		t.Fatal("variant was not set")
	}
	if d.label != "Date" || d.description == "" || d.errorMessage == "" || !d.invalid || !d.required || !d.disabled || !d.fullWidth {
		t.Fatal("boolean option was not set")
	}
	if !sameDate(d.value, value) || !sameDate(d.minDate, minDate) || !sameDate(d.maxDate, maxDate) {
		t.Fatal("date option was not set")
	}
	d.onChange(value)
	if !sameDate(changed, value) {
		t.Fatal("on change was not set")
	}
}

func TestDatePickerNormalizesValue(t *testing.T) {
	value := time.Date(2026, 7, 9, 14, 30, 15, 0, time.UTC)
	d := DatePicker("date", value)

	if d.value.Hour() != 0 || d.value.Minute() != 0 || d.value.Second() != 0 {
		t.Fatalf("value = %v, want date only", d.value)
	}
}

func TestDatePickerLocale(t *testing.T) {
	value := testDate(2026, 7, 9)
	chinese := DatePicker("date", value).Locale(DatePickerChinese())

	if chinese.hint != "YYYY年MM月DD日" {
		t.Fatalf("hint = %q, want Chinese hint", chinese.hint)
	}
	if got := chinese.locale.MonthLabel(value); got != "2026年7月" {
		t.Fatalf("month label = %q, want 2026年7月", got)
	}
	if got := orderedDatePickerWeekdays(chinese.locale); got != [7]string{"一", "二", "三", "四", "五", "六", "日"} {
		t.Fatalf("weekdays = %#v, want Monday-first Chinese weekdays", got)
	}
	if chinese.locale.Months[0] != "1月" {
		t.Fatalf("first month = %q, want 1月", chinese.locale.Months[0])
	}
	if got := chinese.locale.DateLabel(value); got != "2026年7月9日" {
		t.Fatalf("date label = %q, want 2026年7月9日", got)
	}
	if got := DatePickerEnglish().DateLabel(value); got != "07 / 09 / 2026" {
		t.Fatalf("English date label = %q, want 07 / 09 / 2026", got)
	}
	if chinese.locale.DateOrder != [3]DatePart{DatePartYear, DatePartMonth, DatePartDay} ||
		DatePickerEnglish().DateOrder != [3]DatePart{DatePartMonth, DatePartDay, DatePartYear} {
		t.Fatal("date segment order does not match the locale")
	}
}

func TestDatePickerCustomHintSurvivesLocale(t *testing.T) {
	d := DatePicker("date", time.Time{}).
		Hint("Pick date").
		Locale(DatePickerChinese())

	if d.hint != "Pick date" {
		t.Fatalf("hint = %q, want custom hint", d.hint)
	}
}

func TestDatePickerUsesContextLocale(t *testing.T) {
	ctx := newContextWithThemeAndLanguage(nil, nil, locale.LanguageChinese)
	d := DatePicker("date", time.Time{}).resolveLocale(ctx)

	if d.hint != "YYYY年MM月DD日" {
		t.Fatalf("hint = %q, want Chinese hint", d.hint)
	}
	if d.locale.WeekStart != time.Monday {
		t.Fatalf("week start = %v, want Monday", d.locale.WeekStart)
	}
}

func TestDatePickerExplicitLocaleOverridesContextLanguage(t *testing.T) {
	ctx := newContextWithThemeAndLanguage(nil, nil, locale.LanguageChinese)
	d := DatePicker("date", time.Time{}).Locale(DatePickerEnglish()).resolveLocale(ctx)

	if d.hint != "MM / DD / YYYY" {
		t.Fatalf("hint = %q, want English hint", d.hint)
	}
	if d.locale.WeekStart != time.Sunday {
		t.Fatalf("week start = %v, want Sunday", d.locale.WeekStart)
	}
}

func TestDatePickerInitialMonthClampsToRange(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	picker := DatePicker("date", time.Time{}).
		MinDate(testDate(2026, 7, 10)).
		MaxDate(testDate(2026, 8, 20))

	layoutDatePickerFrameAt(ctx, router, picker, testDate(2026, 6, 1))
	if got := dateKey(testComponentState[datePickerState](ctx, "date", stateSlotDatePicker).viewMonth); got != "2026-07-01" {
		t.Fatalf("view month before range = %s, want 2026-07-01", got)
	}

	ctx = newContext(nil)
	router = new(input.Router)
	layoutDatePickerFrameAt(ctx, router, picker, testDate(2026, 9, 1))
	if got := dateKey(testComponentState[datePickerState](ctx, "date", stateSlotDatePicker).viewMonth); got != "2026-08-01" {
		t.Fatalf("view month after range = %s, want 2026-08-01", got)
	}
}

func TestDatePickerViewModeNavigation(t *testing.T) {
	state := &datePickerState{viewMonth: testDate(2026, 7, 1)}

	state.move(1)
	if got := dateKey(state.viewMonth); got != "2026-08-01" {
		t.Fatalf("day view move = %s, want 2026-08-01", got)
	}

	state.viewMode = datePickerViewMonths
	state.move(1)
	if got := dateKey(state.viewMonth); got != "2027-08-01" {
		t.Fatalf("month view move = %s, want 2027-08-01", got)
	}

	state.viewMode = datePickerViewYears
	state.move(-1)
	if got := dateKey(state.viewMonth); got != "2015-08-01" {
		t.Fatalf("year view move = %s, want 2015-08-01", got)
	}
}

func TestDatePickerToggleYearPicker(t *testing.T) {
	state := new(datePickerState)

	state.toggleYearPicker(2026)
	if state.viewMode != datePickerViewYears {
		t.Fatalf("view mode = %v, want years", state.viewMode)
	}
	if !state.yearScrollReady || state.yearScrollYear != 2026 {
		t.Fatalf("year scroll target = %d ready=%v, want 2026 ready", state.yearScrollYear, state.yearScrollReady)
	}
	state.toggleYearPicker(2026)
	if state.viewMode != datePickerViewDays {
		t.Fatalf("view mode = %v, want days", state.viewMode)
	}
}

func TestDatePickerOpenResetsToDayView(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	picker := DatePicker("date", testDate(2026, 7, 1))
	layoutDatePickerFrame(ctx, router, picker)
	state := testComponentState[datePickerState](ctx, "date", stateSlotDatePicker)
	state.viewMode = datePickerViewYears
	state.trigger.Click()

	layoutDatePickerFrame(ctx, router, picker)

	if !state.open {
		t.Fatal("date picker did not open")
	}
	if state.viewMode != datePickerViewDays {
		t.Fatalf("view mode = %v, want days", state.viewMode)
	}
}

func TestDatePickerPointerTriggerFocusIsNotKeyboardVisible(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	picker := DatePicker("date", testDate(2026, 7, 1)).FullWidth()
	start := time.Unix(1, 0)
	layoutDatePickerFrameAt(ctx, router, picker, start)
	state := testComponentState[datePickerState](ctx, "date", stateSlotDatePicker)

	router.Queue(pointer.Event{
		Kind:      pointer.Press,
		Source:    pointer.Mouse,
		PointerID: 1,
		Buttons:   pointer.ButtonPrimary,
		Position:  f32.Pt(282, 18),
	})
	layoutDatePickerFrameAt(ctx, router, picker, start.Add(time.Millisecond))
	router.Queue(pointer.Event{
		Kind:      pointer.Release,
		Source:    pointer.Mouse,
		PointerID: 1,
		Position:  f32.Pt(282, 18),
	})
	layoutDatePickerFrameAt(ctx, router, picker, start.Add(2*time.Millisecond))
	layoutDatePickerFrameAt(ctx, router, picker, start.Add(3*time.Millisecond))

	if !router.Source().Focused(&state.trigger) {
		t.Fatal("pointer click did not focus the calendar trigger")
	}
	if frame.FocusVisible(ctx, &state.trigger, true) {
		t.Fatal("pointer click exposed the calendar trigger keyboard focus ring")
	}
}

func TestDatePickerHeaderAndYearClickFlow(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	picker := DatePicker("date", testDate(2026, 7, 1))

	layoutDatePickerFrame(ctx, router, picker)
	state := testComponentState[datePickerState](ctx, "date", stateSlotDatePicker)
	router.Source().Execute(key.FocusCmd{Tag: &state.trigger})
	state.open = true
	layoutDatePickerFrame(ctx, router, picker)

	state.header.Click()
	layoutDatePickerFrame(ctx, router, picker)
	if state.viewMode != datePickerViewYears {
		t.Fatalf("after header click view mode = %v, want years", state.viewMode)
	}

	year := state.years["2026"]
	if year == nil {
		t.Fatal("missing 2026 year state")
	}
	year.clickable.Click()
	layoutDatePickerFrame(ctx, router, picker)
	if state.viewMode != datePickerViewDays {
		t.Fatalf("after year click view mode = %v, want days", state.viewMode)
	}
	if state.viewMonth.Year() != 2026 {
		t.Fatalf("view year = %d, want 2026", state.viewMonth.Year())
	}
}

func TestDatePickerCalendarFocusKeepsPopoverOpenAndEscapeCloses(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	picker := DatePicker("date", testDate(2026, 7, 1))
	layoutDatePickerFrame(ctx, router, picker)
	state := testComponentState[datePickerState](ctx, "date", stateSlotDatePicker)
	router.Source().Execute(key.FocusCmd{Tag: &state.trigger})
	state.openCalendar()
	layoutDatePickerFrame(ctx, router, picker)

	day := state.days["2026-07-10"]
	if day == nil {
		t.Fatal("missing calendar day")
	}
	router.Source().Execute(key.FocusCmd{Tag: &day.clickable})
	layoutDatePickerFrame(ctx, router, picker)
	if !state.open {
		t.Fatal("calendar focus closed the date picker")
	}

	router.Queue(key.Event{Name: key.NameEscape, State: key.Press})
	layoutDatePickerFrame(ctx, router, picker)
	if state.open {
		t.Fatal("Escape from a calendar day did not close the date picker")
	}
}

func TestDatePickerPanelPressKeepsFocusAndBlocksBackground(t *testing.T) {
	tests := []struct {
		name     string
		position f32.Point
		wantMode datePickerViewMode
	}{
		{name: "padding", position: f32.Pt(200, 47), wantMode: datePickerViewDays},
		{name: "header", position: f32.Pt(50, 62), wantMode: datePickerViewYears},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := newContext(nil)
			router := new(input.Router)
			picker := DatePicker("date", testDate(2026, 7, 1)).FullWidth()
			var background widget.Clickable
			backgroundClicked := false
			start := time.Unix(1, 0)

			layoutDatePickerOverBackgroundFrame(ctx, router, picker, &background, &backgroundClicked, start)
			pickerState := testComponentState[datePickerState](ctx, "date", stateSlotDatePicker)
			router.Source().Execute(key.FocusCmd{Tag: &pickerState.trigger})
			pickerState.openCalendar()
			layoutDatePickerOverBackgroundFrame(ctx, router, picker, &background, &backgroundClicked, start.Add(time.Millisecond))
			layoutDatePickerOverBackgroundFrame(ctx, router, picker, &background, &backgroundClicked, start.Add(time.Millisecond+datePickerPopoverInDuration))
			layoutDatePickerOverBackgroundFrame(ctx, router, picker, &background, &backgroundClicked, start.Add(2*time.Millisecond+datePickerPopoverInDuration))

			router.Queue(pointer.Event{
				Kind:      pointer.Press,
				Source:    pointer.Mouse,
				PointerID: 1,
				Buttons:   pointer.ButtonPrimary,
				Position:  test.position,
			})
			layoutDatePickerOverBackgroundFrame(ctx, router, picker, &background, &backgroundClicked, start.Add(3*time.Millisecond+datePickerPopoverInDuration))

			if !pickerState.open {
				t.Fatalf("%s press closed the date picker", test.name)
			}
			if !router.Source().Focused(&pickerState.trigger) {
				t.Fatalf("%s press cleared trigger focus", test.name)
			}
			if pickerState.viewMode != datePickerViewDays {
				t.Fatalf("view mode changed on %s press: got %v, want days", test.name, pickerState.viewMode)
			}
			if state.ActivePresses(background.History()) != 0 || backgroundClicked {
				t.Fatalf("%s press reached the background control", test.name)
			}

			router.Queue(pointer.Event{
				Kind:      pointer.Release,
				Source:    pointer.Mouse,
				PointerID: 1,
				Position:  test.position,
			})
			layoutDatePickerOverBackgroundFrame(ctx, router, picker, &background, &backgroundClicked, start.Add(4*time.Millisecond+datePickerPopoverInDuration))

			if !pickerState.open {
				t.Fatalf("%s release closed the date picker", test.name)
			}
			if !router.Source().Focused(&pickerState.trigger) {
				t.Fatalf("%s release cleared trigger focus", test.name)
			}
			if pickerState.viewMode != test.wantMode {
				t.Fatalf("view mode after %s release = %v, want %v", test.name, pickerState.viewMode, test.wantMode)
			}
			if backgroundClicked {
				t.Fatalf("%s click reached the background control", test.name)
			}
		})
	}
}

func TestDatePickerExitingPanelPaddingBlocksBackground(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	changed := false
	picker := DatePicker("date", testDate(2026, 7, 1)).
		FullWidth().
		OnChange(func(time.Time) {
			changed = true
		})
	var background widget.Clickable
	backgroundClicked := false
	start := time.Unix(1, 0)

	layoutDatePickerOverBackgroundFrame(ctx, router, picker, &background, &backgroundClicked, start)
	pickerState := testComponentState[datePickerState](ctx, "date", stateSlotDatePicker)
	router.Source().Execute(key.FocusCmd{Tag: &pickerState.trigger})
	pickerState.openCalendar()
	layoutDatePickerOverBackgroundFrame(ctx, router, picker, &background, &backgroundClicked, start.Add(time.Millisecond))
	layoutDatePickerOverBackgroundFrame(ctx, router, picker, &background, &backgroundClicked, start.Add(time.Millisecond+datePickerPopoverInDuration))
	layoutDatePickerOverBackgroundFrame(ctx, router, picker, &background, &backgroundClicked, start.Add(2*time.Millisecond+datePickerPopoverInDuration))

	pickerState.open = false
	closeStart := start.Add(3*time.Millisecond + datePickerPopoverInDuration)
	layoutDatePickerOverBackgroundFrame(ctx, router, picker, &background, &backgroundClicked, closeStart)
	midExit := closeStart.Add(datePickerPopoverOutDuration / 2)
	layoutDatePickerOverBackgroundFrame(ctx, router, picker, &background, &backgroundClicked, midExit)
	if progress := pickerState.popoverTransition.Current(); progress <= 0 || progress >= 1 {
		t.Fatalf("exit progress = %v, want between 0 and 1", progress)
	}

	position := f32.Pt(200, 47)
	router.Queue(pointer.Event{
		Kind:      pointer.Press,
		Source:    pointer.Mouse,
		PointerID: 1,
		Buttons:   pointer.ButtonPrimary,
		Position:  position,
	})
	layoutDatePickerOverBackgroundFrame(ctx, router, picker, &background, &backgroundClicked, midExit.Add(time.Millisecond))

	if pickerState.open {
		t.Fatal("exiting panel padding press reopened the date picker")
	}
	if pickerState.popoverTransition.Current() <= 0 {
		t.Fatal("panel finished exiting before the padding press was processed")
	}
	if pickerState.viewMode != datePickerViewDays || changed {
		t.Fatal("exiting panel content handled the padding press")
	}
	if state.ActivePresses(background.History()) != 0 || backgroundClicked {
		t.Fatal("exiting panel padding press reached the background control")
	}

	router.Queue(pointer.Event{
		Kind:      pointer.Release,
		Source:    pointer.Mouse,
		PointerID: 1,
		Position:  position,
	})
	layoutDatePickerOverBackgroundFrame(ctx, router, picker, &background, &backgroundClicked, midExit.Add(2*time.Millisecond))

	if pickerState.open {
		t.Fatal("exiting panel padding release reopened the date picker")
	}
	if pickerState.popoverTransition.Current() <= 0 {
		t.Fatal("panel finished exiting before the padding release was processed")
	}
	if pickerState.viewMode != datePickerViewDays || changed {
		t.Fatal("exiting panel content handled the padding click")
	}
	if backgroundClicked {
		t.Fatal("exiting panel padding click reached the background control")
	}
}

func TestDatePickerIgnoresEscapeWhileAnotherOverlayIsTopmost(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	picker := DatePicker("date", testDate(2026, 7, 1))

	layoutDatePickerFrame(ctx, router, picker)
	state := testComponentState[datePickerState](ctx, "date", stateSlotDatePicker)
	router.Source().Execute(key.FocusCmd{Tag: &state.trigger})
	state.openCalendar()
	layoutDatePickerBelowOverlayFrame(ctx, router, picker)
	layoutDatePickerBelowOverlayFrame(ctx, router, picker)

	router.Queue(key.Event{Name: key.NameEscape, State: key.Press})
	layoutDatePickerBelowOverlayFrame(ctx, router, picker)
	if !state.open {
		t.Fatal("background date picker consumed Escape while another overlay was topmost")
	}
}

func TestDatePickerPopoverProgressAnimatesInAndOut(t *testing.T) {
	start := testDate(2026, 7, 9)
	state := new(datePickerState)
	var ops op.Ops
	gtx := layout.Context{
		Now: start,
		Ops: &ops,
	}

	if got := state.popoverProgress(gtx, false); got != 0 {
		t.Fatalf("initial progress = %v, want 0", got)
	}
	if got := state.popoverProgress(gtx, true); got != 0 {
		t.Fatalf("opening progress = %v, want 0", got)
	}

	gtx.Now = start.Add(datePickerPopoverInDuration / 2)
	if got := state.popoverProgress(gtx, true); got <= 0 || got >= 1 {
		t.Fatalf("mid open progress = %v, want between 0 and 1", got)
	}

	gtx.Now = start.Add(datePickerPopoverInDuration)
	if got := state.popoverProgress(gtx, true); got != 1 {
		t.Fatalf("open progress = %v, want 1", got)
	}

	closeStart := gtx.Now.Add(time.Millisecond)
	gtx.Now = closeStart
	if got := state.popoverProgress(gtx, false); got != 1 {
		t.Fatalf("closing progress = %v, want 1", got)
	}

	gtx.Now = closeStart.Add(datePickerPopoverOutDuration / 2)
	if got := state.popoverProgress(gtx, false); got <= 0 || got >= 1 {
		t.Fatalf("mid close progress = %v, want between 0 and 1", got)
	}

	gtx.Now = closeStart.Add(datePickerPopoverOutDuration)
	if got := state.popoverProgress(gtx, false); got != 0 {
		t.Fatalf("closed progress = %v, want 0", got)
	}
}

func TestDatePickerOpenLayoutDoesNotTakeSpace(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	picker := DatePicker("date", testDate(2026, 7, 1))
	layoutDatePickerFrame(ctx, router, picker)
	state := testComponentState[datePickerState](ctx, "date", stateSlotDatePicker)
	router.Source().Execute(key.FocusCmd{Tag: &state.trigger})
	state.open = true

	dims := layoutDatePickerFrame(ctx, router, picker)

	if dims.Size.Y != 36 {
		t.Fatalf("open date picker height = %d, want 36", dims.Size.Y)
	}
}

func TestDatePickerClickSelectsDate(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	var got time.Time
	picker := DatePicker("date", testDate(2026, 7, 1)).
		OnChange(func(date time.Time) {
			got = date
		})

	layoutDatePickerFrame(ctx, router, picker)
	state := testComponentState[datePickerState](ctx, "date", stateSlotDatePicker)
	router.Source().Execute(key.FocusCmd{Tag: &state.trigger})
	state.open = true
	layoutDatePickerFrame(ctx, router, picker)

	day := state.days["2026-07-10"]
	if day == nil {
		t.Fatal("missing day state")
	}
	day.clickable.Click()
	layoutDatePickerFrame(ctx, router, picker)

	if !sameDate(got, testDate(2026, 7, 10)) {
		t.Fatalf("selected date = %v, want 2026-07-10", got)
	}
	if state.open {
		t.Fatal("date picker stayed open after selection")
	}
}

func TestDatePickerTodayUsesFrameTimeForDays(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	picker := DatePicker("date", time.Time{})
	now := testDate(2026, 7, 9)
	layoutDatePickerFrameAt(ctx, router, picker, now)
	state := testComponentState[datePickerState](ctx, "date", stateSlotDatePicker)
	router.Source().Execute(key.FocusCmd{Tag: &state.trigger})
	state.open = true

	layoutDatePickerFrameAt(ctx, router, picker, now)

	theme := DefaultTheme()
	day := state.days["2026-07-09"]
	if day == nil {
		t.Fatal("missing today state")
	}
	if background := day.backgroundTransition.Current(); background != theme.Palette.AccentSoft {
		t.Fatalf("today background = %#v, want %#v", background, theme.Palette.AccentSoft)
	}
}

func TestDatePickerTodayUsesFrameTimeForMonthsAndYears(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	picker := DatePicker("date", time.Time{})
	now := testDate(2026, 7, 9)
	layoutDatePickerFrameAt(ctx, router, picker, now)
	state := testComponentState[datePickerState](ctx, "date", stateSlotDatePicker)
	router.Source().Execute(key.FocusCmd{Tag: &state.trigger})
	state.open = true
	state.viewMode = datePickerViewMonths
	state.viewMonth = testDate(2026, 6, 1)

	layoutDatePickerFrameAt(ctx, router, picker, now)

	theme := DefaultTheme()
	month := state.months["2026-07"]
	if month == nil {
		t.Fatal("missing today month state")
	}
	if background := month.backgroundTransition.Current(); background != theme.Palette.AccentSoft {
		t.Fatalf("today month background = %#v, want %#v", background, theme.Palette.AccentSoft)
	}

	state.open = true
	state.viewMonth = testDate(2025, 6, 1)
	state.toggleYearPicker(2026)
	layoutDatePickerFrameAt(ctx, router, picker, now)

	year := state.years["2026"]
	if year == nil {
		t.Fatal("missing today year state")
	}
	if background := year.backgroundTransition.Current(); background != theme.Palette.AccentSoft {
		t.Fatalf("today year background = %#v, want %#v", background, theme.Palette.AccentSoft)
	}
}

func TestDatePickerMinMaxDisablesDates(t *testing.T) {
	picker := DatePicker("date", time.Time{}).
		MinDate(testDate(2026, 7, 10)).
		MaxDate(testDate(2026, 7, 20))

	if !picker.isDateDisabled(testDate(2026, 7, 9)) {
		t.Fatal("date before min was enabled")
	}
	if picker.isDateDisabled(testDate(2026, 7, 15)) {
		t.Fatal("date inside range was disabled")
	}
	if !picker.isDateDisabled(testDate(2026, 7, 21)) {
		t.Fatal("date after max was enabled")
	}
}

func TestDatePickerMinMaxDisablesMonthsAndYears(t *testing.T) {
	picker := DatePicker("date", time.Time{}).
		MinDate(testDate(2026, 7, 10)).
		MaxDate(testDate(2026, 8, 20))

	if !picker.isMonthDisabled(testDate(2026, 6, 1)) {
		t.Fatal("month before min was enabled")
	}
	if picker.isMonthDisabled(testDate(2026, 7, 1)) {
		t.Fatal("month overlapping min was disabled")
	}
	if picker.isMonthDisabled(testDate(2026, 8, 1)) {
		t.Fatal("month overlapping max was disabled")
	}
	if !picker.isMonthDisabled(testDate(2026, 9, 1)) {
		t.Fatal("month after max was enabled")
	}
	if !picker.isYearDisabled(2025) {
		t.Fatal("year before min was enabled")
	}
	if picker.isYearDisabled(2026) {
		t.Fatal("year inside range was disabled")
	}
	if !picker.isYearDisabled(2027) {
		t.Fatal("year after max was enabled")
	}
}

func TestDatePickerYearPickerRangeRespectsExplicitMinMax(t *testing.T) {
	picker := DatePicker("date", testDate(2025, 1, 1)).
		MinDate(testDate(2026, 7, 10)).
		MaxDate(testDate(2028, 8, 20))
	state := &datePickerState{viewMonth: testDate(2025, 6, 1)}

	minYear, maxYear := picker.yearPickerRange(state, testDate(2029, 1, 1))
	if minYear != 2026 || maxYear != 2028 {
		t.Fatalf("year range = %d..%d, want 2026..2028", minYear, maxYear)
	}
}

func TestDatePickerComparesDatesByCalendarDay(t *testing.T) {
	shanghai := time.FixedZone("UTC+8", 8*60*60)
	newYork := time.FixedZone("UTC-5", -5*60*60)
	minDate := time.Date(2026, 7, 10, 0, 0, 0, 0, shanghai)
	sameDay := time.Date(2026, 7, 10, 0, 0, 0, 0, newYork)
	picker := DatePicker("date", time.Time{}).MinDate(minDate)

	if picker.isDateDisabled(sameDay) {
		t.Fatal("same calendar day in a different location was disabled")
	}
}

func TestDatePickerYearRangeStart(t *testing.T) {
	if got := datePickerYearRangeStart(2026); got != 2016 {
		t.Fatalf("range start = %d, want 2016", got)
	}
}

func TestDatePickerMonthDays(t *testing.T) {
	days := datePickerMonthDays(testDate(2026, 7, 1), time.Sunday)

	if len(days) != 42 {
		t.Fatalf("days = %d, want 42", len(days))
	}
	if got := dateKey(days[0].date); got != "2026-06-28" {
		t.Fatalf("first day = %s, want 2026-06-28", got)
	}
	if got := dateKey(days[len(days)-1].date); got != "2026-08-08" {
		t.Fatalf("last day = %s, want 2026-08-08", got)
	}
	if !days[0].outside || days[10].outside {
		t.Fatal("outside month flags were not set")
	}
}

func TestDatePickerMonthDaysWithMondayStart(t *testing.T) {
	days := datePickerMonthDays(testDate(2026, 7, 1), time.Monday)

	if got := dateKey(days[0].date); got != "2026-06-29" {
		t.Fatalf("first day = %s, want 2026-06-29", got)
	}
	if got := dateKey(days[len(days)-1].date); got != "2026-08-09" {
		t.Fatalf("last day = %s, want 2026-08-09", got)
	}
}

func layoutDatePickerFrame(ctx *frame.Context, router *input.Router, picker DatePickerWidget) layout.Dimensions {
	return layoutDatePickerFrameAt(ctx, router, picker, time.Time{})
}

func layoutDatePickerFrameAt(ctx *frame.Context, router *input.Router, picker DatePickerWidget, now time.Time) layout.Dimensions {
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 260)},
		Source:      router.Source(),
		Ops:         &ops,
		Now:         now,
	}
	frame.BeginFrame(ctx)
	dims := picker.Layout(ctx, gtx)
	frame.LayoutOverlays(ctx, gtx)
	frame.ApplyFrameCommands(ctx, gtx)
	frame.EndFrame(ctx)
	router.Frame(&ops)
	return dims
}

func layoutDatePickerOverBackgroundFrame(ctx *frame.Context, router *input.Router, picker DatePickerWidget, background *widget.Clickable, backgroundClicked *bool, now time.Time) {
	var ops op.Ops
	viewport := image.Pt(400, 500)
	gtx := layout.Context{
		Constraints: layout.Exact(viewport),
		Source:      router.Source(),
		Ops:         &ops,
		Now:         now,
	}
	frame.BeginFrameWithViewport(ctx, viewport)
	presses := state.ActivePresses(background.History())
	for background.Clicked(gtx) {
		*backgroundClicked = true
	}
	frame.FocusOnPress(ctx, background, background.History(), presses)
	background.Layout(gtx, func(layout.Context) layout.Dimensions {
		return layout.Dimensions{Size: viewport}
	})
	pickerGtx := gtx
	pickerGtx.Constraints = layout.Constraints{Max: viewport}
	picker.Layout(ctx, pickerGtx)
	frame.LayoutOverlays(ctx, gtx)
	frame.ApplyFrameCommands(ctx, gtx)
	frame.EndFrame(ctx)
	router.Frame(&ops)
}

func layoutDatePickerBelowOverlayFrame(ctx *frame.Context, router *input.Router, picker DatePickerWidget) {
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 260)},
		Source:      router.Source(),
		Ops:         &ops,
	}
	frame.BeginFrame(ctx)
	picker.Layout(ctx, gtx)
	frame.RegisterOverlay(ctx, frame.OverlayRequest{
		Key: "test-top-overlay",
		Layout: func(layout.Context, image.Rectangle, bool) layout.Dimensions {
			return layout.Dimensions{}
		},
	})
	frame.LayoutOverlays(ctx, gtx)
	frame.ApplyFrameCommands(ctx, gtx)
	frame.EndFrame(ctx)
	router.Frame(&ops)
}

func testDate(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}
