package datepicker

import (
	"errors"
	"image/color"
	"strconv"
	"testing"
	"time"

	"gioui.org/io/input"
	"gioui.org/io/key"
	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/field"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/testtheme"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

func TestDatePickerMonthDaysCacheReusesAndInvalidates(t *testing.T) {
	var state datePickerState
	month := testDate(2026, 7, 1)

	first := state.monthDays(month, time.Sunday)
	if len(first) != 42 {
		t.Fatalf("cached month days = %d, want 42", len(first))
	}
	firstAddress := &first[0]
	firstDate := first[0].date
	for _, day := range first {
		if want := strconv.Itoa(day.date.Day()); day.label != want {
			t.Fatalf("day %s label = %q, want %q", day.date.Format("2006-01-02"), day.label, want)
		}
	}

	second := state.monthDays(month, time.Sunday)
	if &second[0] != firstAddress {
		t.Fatal("same month and week start did not reuse the fixed calendar buffer")
	}
	if !sameDate(second[0].date, firstDate) {
		t.Fatalf("cached first date = %v, want %v", second[0].date, firstDate)
	}

	changedMonth := state.monthDays(testDate(2026, 8, 1), time.Sunday)
	if sameDate(changedMonth[0].date, firstDate) {
		t.Fatal("month change reused stale calendar contents")
	}
	changedWeekStart := state.monthDays(month, time.Monday)
	if dateKey(changedWeekStart[0].date) != "2026-06-29" {
		t.Fatalf("week-start change first date = %s, want 2026-06-29", dateKey(changedWeekStart[0].date))
	}

	otherLocation := time.FixedZone("UTC+8", 8*60*60)
	changedLocation := state.monthDays(time.Date(2026, 7, 1, 0, 0, 0, 0, otherLocation), time.Monday)
	if changedLocation[0].date.Location() != otherLocation {
		t.Fatalf("location change retained %v, want %v", changedLocation[0].date.Location(), otherLocation)
	}
}

func TestDatePickerMonthDaysIntoWritesOnlyCalendarSlots(t *testing.T) {
	days := make([]datePickerDay, 43)
	days[42] = datePickerDay{label: "sentinel"}

	datePickerMonthDaysInto(days, testDate(2026, 7, 1), time.Sunday)
	if days[42].label != "sentinel" {
		t.Fatalf("month generator overwrote slot 42 with %q", days[42].label)
	}
}

type countingDatePickerProvider struct {
	theme.ComponentVisualProvider
	calls      int
	incomplete bool
	requests   []theme.DatePickerRequest
}

type datePickerRendererCounts struct {
	calendarIcon  int
	yearIndicator int
	navButton     int
	cell          int
	rangeTrack    int
	strike        int
	focus         int
}

type recordingDatePickerProvider struct {
	theme.ComponentVisualProvider
	counts *datePickerRendererCounts
}

func (p *recordingDatePickerProvider) DatePicker(active *theme.Theme, request theme.DatePickerRequest) theme.DatePickerVisual {
	visual := p.ComponentVisualProvider.DatePicker(active, request)
	calendarIcon := visual.CalendarIcon
	visual.CalendarIcon = func(gtx layout.Context, value theme.DatePickerIconRenderRequest) {
		p.counts.calendarIcon++
		calendarIcon(gtx, value)
	}
	yearIndicator := visual.YearIndicator
	visual.YearIndicator = func(gtx layout.Context, value theme.DatePickerYearIndicatorRenderRequest) {
		p.counts.yearIndicator++
		yearIndicator(gtx, value)
	}
	navButton := visual.NavButton
	visual.NavButton = func(gtx layout.Context, value theme.DatePickerNavButtonRenderRequest) {
		p.counts.navButton++
		navButton(gtx, value)
	}
	cell := visual.CellRenderer
	visual.CellRenderer = func(gtx layout.Context, value theme.DatePickerCellRenderRequest) {
		p.counts.cell++
		cell(gtx, value)
	}
	rangeRenderer := visual.RangeRenderer
	visual.RangeRenderer = func(gtx layout.Context, value theme.DatePickerRangeRenderRequest) {
		p.counts.rangeTrack++
		rangeRenderer(gtx, value)
	}
	strike := visual.Strike
	visual.Strike = func(gtx layout.Context, value theme.DatePickerStrikeRenderRequest) {
		p.counts.strike++
		strike(gtx, value)
	}
	focus := visual.Focus
	visual.Focus = func(gtx layout.Context, value theme.DatePickerFocusRenderRequest) {
		p.counts.focus++
		focus(gtx, value)
	}
	return visual
}

func TestDatePickerInstancesUseProviderRenderers(t *testing.T) {
	active := *testtheme.Behavior()
	counts := &datePickerRendererCounts{}
	baseProvider := testtheme.BehaviorProvider()
	baseVisual := baseProvider.DatePicker(&active, theme.DatePickerRequest{Variant: theme.InputPrimary})
	theme.ValidateDatePickerVisual(baseVisual)
	active.VisualProvider = &recordingDatePickerProvider{
		ComponentVisualProvider: baseProvider,
		counts:                  counts,
	}
	ctx := core.New(nil, &active, locale.LanguageEnglish)
	var router input.Router
	now := testDate(2026, 7, 1)

	picker := New("renderer-picker", testDate(2026, 7, 9))
	layoutDatePickerFrameAt(ctx, &router, picker, now)
	state := testComponentState[datePickerState](ctx, picker.key, stateSlotDatePicker)
	router.Source().Execute(key.FocusCmd{Tag: &state.trigger})
	state.openCalendar()
	layoutDatePickerFrameAt(ctx, &router, picker, now)

	if counts.calendarIcon == 0 || counts.yearIndicator == 0 || counts.navButton == 0 || counts.cell == 0 || counts.focus == 0 {
		t.Fatalf("date picker renderer counts = %#v, want field/calendar renderers", *counts)
	}

	rangePicker := DateRangePicker("renderer-range", DateRange{
		Start: testDate(2026, 7, 5),
		End:   testDate(2026, 7, 15),
	}).MinDate(testDate(2026, 7, 10)).MaxDate(testDate(2026, 7, 20))
	layoutDateComponentFrame(ctx, &router, rangePicker, now)
	rangeState := testComponentState[dateRangePickerState](ctx, "renderer-range", stateSlotDateRangePicker)
	router.Source().Execute(key.FocusCmd{Tag: &rangeState.calendar.trigger})
	rangeState.calendar.openCalendar()
	layoutDateComponentFrame(ctx, &router, rangePicker, now)

	if counts.rangeTrack == 0 || counts.strike == 0 {
		t.Fatalf("date range renderer counts = %#v, want range and strike renderers", *counts)
	}
}

func (p *countingDatePickerProvider) DatePicker(active *theme.Theme, request theme.DatePickerRequest) theme.DatePickerVisual {
	p.calls++
	p.requests = append(p.requests, request)
	if p.incomplete {
		return theme.DatePickerVisual{}
	}
	return p.ComponentVisualProvider.DatePicker(active, request)
}

func TestDatePickerResolvedPartCacheTracksThemeAndScope(t *testing.T) {
	active := *testtheme.Behavior()
	ctx := core.New(nil, &active, locale.LanguageEnglish)
	var cache datePickerResolvedPartCache

	cache.prepare(ctx, style.Style{})
	if !cache.enabled {
		t.Fatal("transition-free theme should enable the resolved part cache")
	}
	value := style.ResolvedStyle{}
	cache.storeCell(7, value)
	if _, ok := cache.cellValue(7); !ok {
		t.Fatal("stored cell style was not returned")
	}
	if cache.hits != 1 || cache.misses != 0 {
		t.Fatalf("cache counters = hits %d misses %d, want one hit", cache.hits, cache.misses)
	}

	restore := core.PushStyle(ctx, style.Style{}.TextColor(style.RGB(0x102030)))
	cache.prepare(ctx, style.Style{})
	restore()
	if cache.scopeInvalidation == 0 {
		t.Fatal("style scope change did not invalidate resolved parts")
	}
	stats := cache.stats()
	if stats.ScopeInvalidations != 1 {
		t.Fatalf("scope invalidations = %d, want 1", stats.ScopeInvalidations)
	}
	if _, ok := cache.cellValue(7); ok {
		t.Fatal("resolved cell style survived a style scope change")
	}

	other := *testtheme.Behavior()
	cache.storeWeekday(0, value)
	core.ReplaceTheme(ctx, other)
	cache.prepare(ctx, style.Style{})
	if cache.themeInvalidation == 0 {
		t.Fatal("theme replacement did not invalidate resolved parts")
	}
	stats = cache.stats()
	if stats.ThemeInvalidations != 1 {
		t.Fatalf("theme invalidations = %d, want 1", stats.ThemeInvalidations)
	}
	if _, ok := cache.weekdayValue(0); ok {
		t.Fatal("weekday style survived a theme replacement")
	}
}

func TestDatePickerResolvedPartCacheTracksGenerationInvalidation(t *testing.T) {
	active := *testtheme.Behavior()
	ctx := core.New(nil, &active, locale.LanguageEnglish)
	var cache datePickerResolvedPartCache
	cache.prepare(ctx, style.Style{}, field.Primary)

	core.ReplaceLanguage(ctx, locale.LanguageChinese)
	cache.prepare(ctx, style.Style{}, field.Primary)

	stats := cache.stats()
	if stats.GenerationInvalidations != 1 {
		t.Fatalf("generation invalidations = %d, want 1", stats.GenerationInvalidations)
	}
	if stats.ThemeInvalidations != 0 {
		t.Fatalf("theme invalidations = %d, want 0 for language-only change", stats.ThemeInvalidations)
	}
}

func TestDatePickerResolvedPartCacheTracksVariant(t *testing.T) {
	active := *testtheme.Behavior()
	ctx := core.New(nil, &active, locale.LanguageEnglish)
	var cache datePickerResolvedPartCache
	cache.prepare(ctx, style.Style{}, field.Primary)
	cache.storeHeader(3, style.ResolvedStyle{})
	if _, ok := cache.headerValue(3); !ok {
		t.Fatal("stored header style was not returned")
	}

	cache.prepare(ctx, style.Style{}, field.Secondary)
	if _, ok := cache.headerValue(3); ok {
		t.Fatal("header style survived a variant change")
	}
	stats := cache.stats()
	if stats.VariantInvalidations != 1 {
		t.Fatalf("variant invalidations = %d, want 1", stats.VariantInvalidations)
	}
}

func TestDatePickerResolvedPartCacheStoresTextPartsByStateAndColor(t *testing.T) {
	active := *testtheme.Behavior()
	ctx := core.New(nil, &active, locale.LanguageEnglish)
	var cache datePickerResolvedPartCache
	cache.prepare(ctx, style.Style{}, field.Primary)
	state := style.State{Focused: true}
	base := color.NRGBA{R: 0x10, G: 0x20, B: 0x30, A: 0xff}
	key := datePickerTextResolvedKey(state, base)
	cache.storeSegment(key, style.ResolvedStyle{})
	if _, ok := cache.segmentValue(key); !ok {
		t.Fatal("stored segment style was not returned")
	}
	if _, ok := cache.segmentValue(datePickerTextResolvedKey(state, color.NRGBA{R: 0x11, G: 0x20, B: 0x30, A: 0xff})); ok {
		t.Fatal("segment style ignored its base color")
	}
	cache.storeLiteral(key, style.ResolvedStyle{})
	if _, ok := cache.literalValue(key); !ok {
		t.Fatal("stored literal style was not returned")
	}
}

func TestDatePickerResolvedPartCacheBypassesTransitions(t *testing.T) {
	active := *testtheme.Behavior()
	ctx := core.New(nil, &active, locale.LanguageEnglish)
	var cache datePickerResolvedPartCache
	cache.prepare(ctx, style.Style{}.Transition(style.PropOpacity, time.Millisecond))
	if !cache.enabled {
		t.Fatal("duration-only transition should remain cache-safe until resolution")
	}
	cache.storeCell(1, style.ResolvedStyle{Transitions: []style.Transition{{Property: style.PropOpacity}}})
	if _, ok := cache.cellValue(1); ok {
		t.Fatal("transition-bearing resolved style was cached")
	}

	cache.prepare(ctx, style.Style{}.Transition(style.PropOpacity, time.Millisecond, style.TransitionEase(func(progress float32) float32 {
		return progress * progress
	})))
	if cache.enabled {
		t.Fatal("custom easing declaration should disable the resolved part cache")
	}
	cache.prepare(ctx, style.Style{}.Transition(style.PropOpacity, time.Millisecond, style.TransitionEase(func(progress float32) float32 {
		return progress * progress
	})))
	if stats := cache.stats(); stats.UnsafeBypasses != 2 {
		t.Fatalf("unsafe bypasses = %d, want 2 across repeated uncached prepares", stats.UnsafeBypasses)
	}
}

func TestDatePickerResolvedPartCacheTracksWeekdayStates(t *testing.T) {
	active := *testtheme.Behavior()
	ctx := core.New(nil, &active, locale.LanguageEnglish)
	var cache datePickerResolvedPartCache
	cache.prepare(ctx, style.Style{})

	value := style.ResolvedStyle{}
	cache.storeWeekday(0, value)
	if _, ok := cache.weekdayValue(0); !ok {
		t.Fatal("enabled weekday style was not returned")
	}
	if _, ok := cache.weekdayValue(1); ok {
		t.Fatal("disabled weekday style unexpectedly hit before storing")
	}
	cache.storeWeekday(1, value)
	if _, ok := cache.weekdayValue(1); !ok {
		t.Fatal("disabled weekday style was not returned")
	}

	restore := core.PushStyle(ctx, style.Style{}.TextColor(style.RGB(0x102030)))
	cache.prepare(ctx, style.Style{})
	restore()
	if _, ok := cache.weekdayValue(0); ok {
		t.Fatal("weekday style survived a style scope change")
	}
}

func TestDatePickerResolvedPartCacheDoesNotStoreTransitionWeekday(t *testing.T) {
	active := *testtheme.Behavior()
	ctx := core.New(nil, &active, locale.LanguageEnglish)
	var cache datePickerResolvedPartCache
	cache.prepare(ctx, style.Style{})
	cache.storeWeekday(0, style.ResolvedStyle{Transitions: []style.Transition{{Property: style.PropOpacity}}})
	if _, ok := cache.weekdayValue(0); ok {
		t.Fatal("transition-bearing weekday style was cached")
	}

	cache.prepare(ctx, style.Style{}.Transition(style.PropOpacity, time.Millisecond, style.TransitionEase(func(progress float32) float32 {
		return progress * progress
	})))
	cache.storeWeekday(0, style.ResolvedStyle{})
	if _, ok := cache.weekdayValue(0); ok {
		t.Fatal("custom easing declaration unexpectedly enabled weekday cache")
	}
}

func TestDatePickerResolvedPartCacheDoesNotStoreTransitionFocus(t *testing.T) {
	active := *testtheme.Behavior()
	ctx := core.New(nil, &active, locale.LanguageEnglish)
	var cache datePickerResolvedPartCache
	cache.prepare(ctx, style.Style{})
	cache.storeFocus(3, style.ResolvedStyle{Transitions: []style.Transition{{Property: style.PropOpacity}}})
	if _, ok := cache.focusValue(3); ok {
		t.Fatal("transition-bearing focus style was cached")
	}

	cache.prepare(ctx, style.Style{}.Transition(style.PropOpacity, time.Millisecond, style.TransitionEase(func(progress float32) float32 {
		return progress * progress
	})))
	cache.storeFocus(3, style.ResolvedStyle{})
	if _, ok := cache.focusValue(3); ok {
		t.Fatal("custom easing declaration unexpectedly enabled focus cache")
	}
}

func TestDatePickerResolvedPartCacheUsedByOpenCalendar(t *testing.T) {
	active := *testtheme.Behavior()
	ctx := core.New(nil, &active, locale.LanguageEnglish)
	picker := New("resolved-cache", testDate(2026, 7, 9))
	state := datePickerStateFor(ctx, picker.key)
	var router input.Router
	router.Source().Execute(key.FocusCmd{Tag: &state.trigger})
	state.openCalendar()

	layoutDatePickerFrameAt(ctx, &router, picker, testDate(2026, 7, 1))
	misses := state.resolvedParts.misses
	if misses == 0 {
		t.Fatal("open calendar did not resolve any parts")
	}
	layoutDatePickerFrameAt(ctx, &router, picker, testDate(2026, 7, 1).AddDate(0, 0, 1))
	if state.resolvedParts.hits == 0 {
		t.Fatal("second open calendar frame did not hit resolved part cache")
	}
	if state.resolvedParts.misses < misses {
		t.Fatalf("cache misses moved backwards: %d -> %d", misses, state.resolvedParts.misses)
	}
}

func TestDatePickerCalendarDeclarationReusesVisualBundle(t *testing.T) {
	active := *testtheme.Behavior()
	base, err := theme.ComponentProvider(&active)
	if err != nil {
		t.Fatal(err)
	}
	provider := &countingDatePickerProvider{ComponentVisualProvider: base}
	active.VisualProvider = provider
	ctx := core.New(nil, &active, locale.LanguageEnglish)
	var router input.Router

	picker := New("visual-bundle-cache", testDate(2026, 7, 9))
	layoutDatePickerFrameAt(ctx, &router, picker, testDate(2026, 7, 1))

	// The field and calendar consume one shared provider bundle. The Calendar
	// layer must not trigger another DatePicker construction through
	// datePickerBaseStyle, and the field reads Defaults from the same bundle.
	if provider.calls != 1 {
		t.Fatalf("DatePicker provider calls = %d, want 1 shared visual bundle", provider.calls)
	}
}

func TestDatePickerVisualCacheUsesSharedProviderProtocol(t *testing.T) {
	active := *testtheme.Behavior()
	base, err := theme.ComponentProvider(&active)
	if err != nil {
		t.Fatal(err)
	}
	provider := &countingDatePickerProvider{ComponentVisualProvider: base}
	active.VisualProvider = provider

	cache := datePickerPartDefaultsCache{}
	cache.visualFor(&active, field.Primary)
	cache.visualFor(&active, field.Primary)
	if provider.calls != 1 {
		t.Fatalf("same DatePicker visual request called Provider %d times, want 1", provider.calls)
	}

	cache.visualFor(&active, field.Secondary)
	if provider.calls != 2 {
		t.Fatalf("variant change called Provider %d times, want 2", provider.calls)
	}

	replacement := active
	replacement.VisualProvider = provider
	cache.visualFor(&replacement, field.Secondary)
	if provider.calls != 3 {
		t.Fatalf("theme replacement called Provider %d times, want 3", provider.calls)
	}
	stats := cache.visualCache.Stats()
	if stats.Hits != 1 || stats.Misses != 3 || stats.KeyInvalidations != 1 || stats.ThemeInvalidations != 1 {
		t.Fatalf("visual cache stats = %#v, want one hit, three misses and one key/theme invalidation", stats)
	}
}

func TestDatePickerVisualCacheRejectsIncompleteProviderWithoutPoisoning(t *testing.T) {
	active := *testtheme.Behavior()
	base, err := theme.ComponentProvider(&active)
	if err != nil {
		t.Fatal(err)
	}
	provider := &countingDatePickerProvider{ComponentVisualProvider: base, incomplete: true}
	active.VisualProvider = provider
	cache := datePickerPartDefaultsCache{}

	func() {
		defer func() {
			recovered := recover()
			err, ok := recovered.(error)
			if !ok || !errors.Is(err, theme.ErrIncompleteVisualProvider) {
				t.Fatalf("incomplete visual panic = %v, want %v", recovered, theme.ErrIncompleteVisualProvider)
			}
		}()
		cache.visualFor(&active, field.Primary)
	}()
	if stats := cache.visualCache.Stats(); stats.Hits != 0 || stats.Misses != 0 {
		t.Fatalf("incomplete visual cache stats = %#v, want no committed request", stats)
	}

	provider.incomplete = false
	cache.visualFor(&active, field.Primary)
	if stats := cache.visualCache.Stats(); stats.Hits != 0 || stats.Misses != 1 {
		t.Fatalf("recovered visual cache stats = %#v, want one committed miss", stats)
	}
}

func TestDatePickerOverrideRejectsIncompleteProvider(t *testing.T) {
	active := *testtheme.Behavior()
	base, err := theme.ComponentProvider(&active)
	if err != nil {
		t.Fatal(err)
	}
	provider := &countingDatePickerProvider{ComponentVisualProvider: base, incomplete: true}
	active.VisualProvider = provider
	cache := datePickerPartDefaultsCache{}

	defer func() {
		recovered := recover()
		err, ok := recovered.(error)
		if !ok || !errors.Is(err, theme.ErrIncompleteVisualProvider) {
			t.Fatalf("incomplete override visual panic = %v, want %v", recovered, theme.ErrIncompleteVisualProvider)
		}
	}()
	cache.visualForRequest(&active, field.Primary, &theme.DatePickerOverride{})
}

func TestDatePickerOverrideVisualBypassesBaseCache(t *testing.T) {
	active := *testtheme.Behavior()
	base, err := theme.ComponentProvider(&active)
	if err != nil {
		t.Fatal(err)
	}
	provider := &countingDatePickerProvider{ComponentVisualProvider: base}
	active.VisualProvider = provider
	cache := datePickerPartDefaultsCache{}

	baseVisual := cache.visualFor(&active, field.Primary)
	declaration := style.Style{}.TextColor(style.RGB(0x102030))
	overrideVisual := cache.visualForRequest(&active, field.Primary, &theme.DatePickerOverride{Declaration: declaration})
	if got := overrideVisual.Override.Hash64(); got != declaration.Hash64() {
		t.Fatalf("DatePicker override visual hash = %x, want %x", got, declaration.Hash64())
	}
	if provider.calls != 2 {
		t.Fatalf("base plus override requests called Provider %d times, want 2", provider.calls)
	}

	cachedBase := cache.visualFor(&active, field.Primary)
	if provider.calls != 2 {
		t.Fatalf("override request replaced the cached base visual; Provider calls = %d, want 2", provider.calls)
	}
	if cachedBase.Override.Hash64() != baseVisual.Override.Hash64() {
		t.Fatal("base visual cache returned instance override data")
	}
}

func TestDateFieldReusesVisualBundleForFieldDefaults(t *testing.T) {
	active := *testtheme.Behavior()
	base, err := theme.ComponentProvider(&active)
	if err != nil {
		t.Fatal(err)
	}
	provider := &countingDatePickerProvider{ComponentVisualProvider: base}
	active.VisualProvider = provider
	ctx := core.New(nil, &active, locale.LanguageEnglish)
	var router input.Router

	layoutDateComponentFrame(ctx, &router, DateField("visual-bundle-field", testDate(2026, 7, 9)), testDate(2026, 7, 1))
	if provider.calls != 1 {
		t.Fatalf("DateField DatePicker provider calls = %d, want 1 shared visual bundle", provider.calls)
	}
}

func TestDateRangePickerReusesVisualBundleForFieldDefaults(t *testing.T) {
	active := *testtheme.Behavior()
	base, err := theme.ComponentProvider(&active)
	if err != nil {
		t.Fatal(err)
	}
	provider := &countingDatePickerProvider{ComponentVisualProvider: base}
	active.VisualProvider = provider
	ctx := core.New(nil, &active, locale.LanguageEnglish)
	var router input.Router

	layoutDateComponentFrame(ctx, &router, DateRangePicker("visual-bundle-range", DateRange{
		Start: testDate(2026, 7, 9),
		End:   testDate(2026, 7, 12),
	}), testDate(2026, 7, 1))
	if provider.calls != 1 {
		t.Fatalf("DateRangePicker DatePicker provider calls = %d, want 1 shared visual bundle", provider.calls)
	}
}

func TestDatePickerVisualBundleUsesRequestedVariant(t *testing.T) {
	active := *testtheme.Behavior()
	base, err := theme.ComponentProvider(&active)
	if err != nil {
		t.Fatal(err)
	}
	provider := &countingDatePickerProvider{ComponentVisualProvider: base}
	active.VisualProvider = provider
	ctx := core.New(nil, &active, locale.LanguageEnglish)
	var router input.Router

	layoutDatePickerFrameAt(ctx, &router, New("visual-bundle-variant", testDate(2026, 7, 9)).Variant(field.Secondary), testDate(2026, 7, 1))
	if len(provider.requests) != 1 || provider.requests[0].Variant != theme.InputSecondary {
		t.Fatalf("DatePicker requests = %#v, want one secondary request", provider.requests)
	}
}

func TestDatePickerProviderResolvesInstanceOverride(t *testing.T) {
	active := *testtheme.Behavior()
	base, err := theme.ComponentProvider(&active)
	if err != nil {
		t.Fatal(err)
	}
	declaration := style.Style{}.TextColor(style.RGB(0x102030))
	visual := base.DatePicker(&active, theme.DatePickerRequest{
		Variant:  theme.InputPrimary,
		Override: &theme.DatePickerOverride{Declaration: declaration},
	})
	if got := visual.Override.Hash64(); got != declaration.Hash64() {
		t.Fatalf("DatePicker visual override hash = %x, want %x", got, declaration.Hash64())
	}
}

func TestDatePickerStyledWrappersPassInstanceOverrideToProvider(t *testing.T) {
	declaration := style.Style{}.TextColor(style.RGB(0x102030))
	cases := []struct {
		name  string
		value func() core.Widget
	}{
		{
			name: "date-picker",
			value: func() core.Widget {
				return (New("styled-picker", testDate(2026, 7, 9))).Style(declaration)
			},
		},
		{
			name: "date-field",
			value: func() core.Widget {
				return (DateField("styled-field", testDate(2026, 7, 9))).Style(declaration)
			},
		},
		{
			name: "date-range-picker",
			value: func() core.Widget {
				return (DateRangePicker("styled-range", DateRange{
					Start: testDate(2026, 7, 9),
					End:   testDate(2026, 7, 12),
				})).Style(declaration)
			},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			active := *testtheme.Behavior()
			base, err := theme.ComponentProvider(&active)
			if err != nil {
				t.Fatal(err)
			}
			provider := &countingDatePickerProvider{ComponentVisualProvider: base}
			active.VisualProvider = provider
			ctx := core.New(nil, &active, locale.LanguageEnglish)
			var router input.Router

			layoutDateComponentFrame(ctx, &router, testCase.value(), testDate(2026, 7, 1))
			if provider.calls != 1 {
				t.Fatalf("DatePicker Provider calls = %d, want 1 complete override visual; requests = %#v", provider.calls, provider.requests)
			}
			for _, request := range provider.requests {
				if request.Override == nil {
					continue
				}
				if got := request.Override.Declaration.Hash64(); got != declaration.Hash64() {
					t.Fatalf("Provider received override hash = %x, want %x", got, declaration.Hash64())
				}
				return
			}
			t.Fatalf("Provider received no DatePicker instance override; requests = %#v", provider.requests)
		})
	}
}
