package datepicker

import (
	"image"
	"image/color"
	"time"

	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/animation"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/field"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/disclosure"
	"github.com/qianniancn/flowui/widget/optionrow"
)

const stateSlotDatePicker = "datepicker"

func datePickerStateFor(ctx *core.Context, key string) *datePickerState {
	key = core.ClaimKey(ctx, core.KindDatePicker, key)
	return core.UseState[datePickerState](ctx, key, stateSlotDatePicker)
}

type datePickerState struct {
	disclosure        disclosure.Binding[time.Time]
	segments          dateSegmentsState
	hover             dateInputHoverState
	trigger           widget.Clickable
	dialog            core.ClickArea
	header            widget.Clickable
	prev              widget.Clickable
	next              widget.Clickable
	open              bool
	viewMode          datePickerViewMode
	viewMonth         time.Time
	syncedValue       time.Time
	monthReady        bool
	popoverTransition animation.FloatTransition
	prevBackground    animation.ColorTransition
	nextBackground    animation.ColorTransition
	yearList          layout.List
	yearBar           widget.Scrollbar
	yearVisualOutset  layoutui.VisualOutsetState
	yearScrollYear    int
	yearScrollReady   bool
	days              map[string]*datePickerCellState
	months            map[string]*datePickerCellState
	years             map[string]*datePickerCellState
	frameDays         map[string]struct{}
	frameMonths       map[string]struct{}
	frameYears        map[string]struct{}
	calendarDays      [42]datePickerDay
	calendarDaysYear  int
	calendarDaysMonth time.Month
	calendarDaysWeek  time.Weekday
	calendarDaysLoc   *time.Location
	calendarDaysValid bool
	calendarStyle     style.Style
	partDefaults      datePickerPartDefaultsCache
	resolvedParts     datePickerResolvedPartCache
	localeCache       datePickerLocaleCache
	calendarLayout    datePickerCalendarLayoutCache
}

// datePickerCalendarLayoutCache owns the fixed Gio child descriptors used by
// the six-week calendar grid. The descriptors only need their slot index;
// frame-specific date, theme-independent geometry, and picker values are
// refreshed before the grid is laid out. Keeping the descriptors on state
// avoids rebuilding 42 escaping closures on every open calendar frame.
type datePickerCalendarLayoutCache struct {
	ctx         *core.Context
	picker      Widget
	state       *datePickerState
	days        []datePickerDay
	cell        int
	metrics     theme.DatePickerMetrics
	now         time.Time
	rows        [6]layout.FlexChild
	dayChildren [6][7]layout.FlexChild
	initialized bool
}

func (c *datePickerCalendarLayoutCache) prepare(ctx *core.Context, picker Widget, state *datePickerState, days []datePickerDay, cell int, metrics theme.DatePickerMetrics, now time.Time) {
	if !c.initialized {
		c.initialize()
	}
	c.ctx = ctx
	c.picker = picker
	c.state = state
	c.days = days
	c.cell = cell
	c.metrics = metrics
	c.now = now
}

func (c *datePickerCalendarLayoutCache) initialize() {
	for row := range c.rows {
		c.rows[row] = layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return c.layoutWeek(gtx, row)
		})
		for column := range c.dayChildren[row] {
			c.dayChildren[row][column] = layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return c.layoutDay(gtx, row, column)
			})
		}
	}
	c.initialized = true
}

func (c *datePickerCalendarLayoutCache) layoutWeek(gtx layout.Context, row int) layout.Dimensions {
	return layout.Flex{Axis: layout.Horizontal}.Layout(gtx, c.dayChildren[row][:]...)
}

func (c *datePickerCalendarLayoutCache) layoutDay(gtx layout.Context, row, column int) layout.Dimensions {
	gtx.Constraints = layout.Exact(image.Pt(c.cell, c.cell))
	index := row*7 + column
	day := c.days[index]
	previousOutside := column > 0 && c.days[index-1].outside
	nextOutside := column < 6 && c.days[index+1].outside
	return c.picker.layoutDay(c.ctx, gtx, c.state, day, column, previousOutside, nextOutside, c.metrics, c.now)
}

func (s *datePickerState) monthDays(month time.Time, weekStart time.Weekday) []datePickerDay {
	month = firstOfMonth(month)
	year, valueMonth, _ := month.Date()
	if s.calendarDaysValid && s.calendarDaysYear == year && s.calendarDaysMonth == valueMonth && s.calendarDaysWeek == weekStart && s.calendarDaysLoc == month.Location() {
		return s.calendarDays[:]
	}
	datePickerMonthDaysInto(s.calendarDays[:], month, weekStart)
	s.calendarDaysYear = year
	s.calendarDaysMonth = valueMonth
	s.calendarDaysWeek = weekStart
	s.calendarDaysLoc = month.Location()
	s.calendarDaysValid = true
	return s.calendarDays[:]
}

const datePickerResolvedPartCacheLimit = 256

// DatePickerResolvedCacheStats is a snapshot of the transition-free resolved
// part cache. It is intentionally limited to evidence used by tests and
// benchmarks; applications do not need to depend on cache internals.
type DatePickerResolvedCacheStats struct {
	Hits                    uint64
	Misses                  uint64
	Evictions               uint64
	ThemeInvalidations      uint64
	GenerationInvalidations uint64
	ScopeInvalidations      uint64
	VariantInvalidations    uint64
	UnsafeBypasses          uint64
}

// datePickerResolvedPartCache keeps transition-free calendar parts after the
// declaration cascade. Calendar frames revisit the same small set of state
// combinations, so retaining the resolved value avoids re-hashing the full
// inherited/custom style stack for every cell. Unsafe declarations and styles
// with transitions deliberately bypass this cache because they need per-use
// ownership or animation state.
type datePickerResolvedPartCache struct {
	theme                  *theme.Theme
	themeGeneration        uint64
	scopeHash              uint64
	variant                field.Variant
	enabled                bool
	cell                   map[uint32]style.ResolvedStyle
	focus                  map[uint16]style.ResolvedStyle
	rangeStyles            map[uint16]style.ResolvedStyle
	header                 map[uint16]style.ResolvedStyle
	nav                    map[uint16]style.ResolvedStyle
	segments               map[uint64]style.ResolvedStyle
	literals               map[uint64]style.ResolvedStyle
	weekday                [2]style.ResolvedStyle
	weekdayValid           [2]bool
	hits                   uint64
	misses                 uint64
	evictions              uint64
	themeInvalidation      uint64
	generationInvalidation uint64
	scopeInvalidation      uint64
	variantInvalidation    uint64
	unsafeBypass           uint64
}

func (c *datePickerResolvedPartCache) stats() DatePickerResolvedCacheStats {
	if c == nil {
		return DatePickerResolvedCacheStats{}
	}
	return DatePickerResolvedCacheStats{
		Hits:                    c.hits,
		Misses:                  c.misses,
		Evictions:               c.evictions,
		ThemeInvalidations:      c.themeInvalidation,
		GenerationInvalidations: c.generationInvalidation,
		ScopeInvalidations:      c.scopeInvalidation,
		VariantInvalidations:    c.variantInvalidation,
		UnsafeBypasses:          c.unsafeBypass,
	}
}

func (c *datePickerResolvedPartCache) prepare(ctx *core.Context, custom style.Style, variants ...field.Variant) {
	activeTheme := core.ActiveTheme(ctx)
	generation := core.ThemeGeneration(ctx)
	enabled := datePickerStyleLayersCacheSafe(ctx, custom)
	variant := field.Primary
	if len(variants) > 0 {
		variant = variants[0]
	}
	scopeHash := uint64(0)
	if enabled {
		scopeHash = datePickerStyleScopeHash(ctx, custom)
	}
	if c.theme == activeTheme && c.themeGeneration == generation && c.scopeHash == scopeHash && c.enabled == enabled && c.variant == variant {
		if !enabled {
			c.unsafeBypass++
		}
		return
	}
	if c.theme != nil {
		switch {
		case c.theme != activeTheme:
			c.themeInvalidation++
		case c.themeGeneration != generation:
			c.generationInvalidation++
		case c.scopeHash != scopeHash:
			c.scopeInvalidation++
		case c.variant != variant:
			c.variantInvalidation++
		}
	}
	if !enabled {
		c.unsafeBypass++
	}
	c.theme = activeTheme
	c.themeGeneration = generation
	c.scopeHash = scopeHash
	c.variant = variant
	c.enabled = enabled
	clear(c.cell)
	clear(c.focus)
	clear(c.rangeStyles)
	clear(c.header)
	clear(c.nav)
	clear(c.segments)
	clear(c.literals)
	c.weekdayValid = [2]bool{}
}

func (c *datePickerResolvedPartCache) cellValue(key uint32) (style.ResolvedStyle, bool) {
	if !c.enabled {
		return style.ResolvedStyle{}, false
	}
	value, ok := c.cell[key]
	if ok {
		c.hits++
	} else {
		c.misses++
	}
	return value, ok
}

func (c *datePickerResolvedPartCache) focusValue(key uint16) (style.ResolvedStyle, bool) {
	if !c.enabled {
		return style.ResolvedStyle{}, false
	}
	value, ok := c.focus[key]
	if ok {
		c.hits++
	} else {
		c.misses++
	}
	return value, ok
}

func (c *datePickerResolvedPartCache) rangeValue(key uint16) (style.ResolvedStyle, bool) {
	if !c.enabled {
		return style.ResolvedStyle{}, false
	}
	value, ok := c.rangeStyles[key]
	if ok {
		c.hits++
	} else {
		c.misses++
	}
	return value, ok
}

func (c *datePickerResolvedPartCache) headerValue(key uint16) (style.ResolvedStyle, bool) {
	if !c.enabled {
		return style.ResolvedStyle{}, false
	}
	value, ok := c.header[key]
	if ok {
		c.hits++
	} else {
		c.misses++
	}
	return value, ok
}

func (c *datePickerResolvedPartCache) navValue(key uint16) (style.ResolvedStyle, bool) {
	if !c.enabled {
		return style.ResolvedStyle{}, false
	}
	value, ok := c.nav[key]
	if ok {
		c.hits++
	} else {
		c.misses++
	}
	return value, ok
}

func (c *datePickerResolvedPartCache) segmentValue(key uint64) (style.ResolvedStyle, bool) {
	if !c.enabled {
		return style.ResolvedStyle{}, false
	}
	value, ok := c.segments[key]
	if ok {
		c.hits++
	} else {
		c.misses++
	}
	return value, ok
}

func (c *datePickerResolvedPartCache) literalValue(key uint64) (style.ResolvedStyle, bool) {
	if !c.enabled {
		return style.ResolvedStyle{}, false
	}
	value, ok := c.literals[key]
	if ok {
		c.hits++
	} else {
		c.misses++
	}
	return value, ok
}

func (c *datePickerResolvedPartCache) weekdayValue(key uint8) (style.ResolvedStyle, bool) {
	if !c.enabled || key >= uint8(len(c.weekday)) || !c.weekdayValid[key] {
		if c.enabled {
			c.misses++
		}
		return style.ResolvedStyle{}, false
	}
	c.hits++
	return c.weekday[key], true
}

func (c *datePickerResolvedPartCache) storeCell(key uint32, value style.ResolvedStyle) {
	if !c.enabled || len(value.Transitions) != 0 {
		return
	}
	if len(c.cell) >= datePickerResolvedPartCacheLimit {
		clear(c.cell)
		c.evictions++
	}
	if c.cell == nil {
		c.cell = make(map[uint32]style.ResolvedStyle)
	}
	c.cell[key] = value
}

func (c *datePickerResolvedPartCache) storeFocus(key uint16, value style.ResolvedStyle) {
	if !c.enabled || len(value.Transitions) != 0 {
		return
	}
	if len(c.focus) >= datePickerResolvedPartCacheLimit {
		clear(c.focus)
		c.evictions++
	}
	if c.focus == nil {
		c.focus = make(map[uint16]style.ResolvedStyle)
	}
	c.focus[key] = value
}

func (c *datePickerResolvedPartCache) storeRange(key uint16, value style.ResolvedStyle) {
	if !c.enabled || len(value.Transitions) != 0 {
		return
	}
	if len(c.rangeStyles) >= datePickerResolvedPartCacheLimit {
		clear(c.rangeStyles)
		c.evictions++
	}
	if c.rangeStyles == nil {
		c.rangeStyles = make(map[uint16]style.ResolvedStyle)
	}
	c.rangeStyles[key] = value
}

func (c *datePickerResolvedPartCache) storeHeader(key uint16, value style.ResolvedStyle) {
	if !c.enabled || len(value.Transitions) != 0 {
		return
	}
	if len(c.header) >= datePickerResolvedPartCacheLimit {
		clear(c.header)
		c.evictions++
	}
	if c.header == nil {
		c.header = make(map[uint16]style.ResolvedStyle)
	}
	c.header[key] = value
}

func (c *datePickerResolvedPartCache) storeNav(key uint16, value style.ResolvedStyle) {
	if !c.enabled || len(value.Transitions) != 0 {
		return
	}
	if len(c.nav) >= datePickerResolvedPartCacheLimit {
		clear(c.nav)
		c.evictions++
	}
	if c.nav == nil {
		c.nav = make(map[uint16]style.ResolvedStyle)
	}
	c.nav[key] = value
}

func (c *datePickerResolvedPartCache) storeSegment(key uint64, value style.ResolvedStyle) {
	if !c.enabled || len(value.Transitions) != 0 {
		return
	}
	if len(c.segments) >= datePickerResolvedPartCacheLimit {
		clear(c.segments)
		c.evictions++
	}
	if c.segments == nil {
		c.segments = make(map[uint64]style.ResolvedStyle)
	}
	c.segments[key] = value
}

func (c *datePickerResolvedPartCache) storeLiteral(key uint64, value style.ResolvedStyle) {
	if !c.enabled || len(value.Transitions) != 0 {
		return
	}
	if len(c.literals) >= datePickerResolvedPartCacheLimit {
		clear(c.literals)
		c.evictions++
	}
	if c.literals == nil {
		c.literals = make(map[uint64]style.ResolvedStyle)
	}
	c.literals[key] = value
}

func (c *datePickerResolvedPartCache) storeWeekday(key uint8, value style.ResolvedStyle) {
	if !c.enabled || key >= uint8(len(c.weekday)) || len(value.Transitions) != 0 {
		return
	}
	c.weekday[key] = value
	c.weekdayValid[key] = true
}

func datePickerStyleLayersCacheSafe(ctx *core.Context, custom style.Style) bool {
	if custom.CacheUnsafe() {
		return false
	}
	for _, value := range core.ActiveInheritedStylesReadOnly(ctx) {
		if value.CacheUnsafe() {
			return false
		}
	}
	for _, value := range core.ActiveStylesReadOnly(ctx) {
		if value.CacheUnsafe() {
			return false
		}
	}
	return true
}

func datePickerStyleScopeHash(ctx *core.Context, custom style.Style) uint64 {
	const (
		offset64 = uint64(14695981039346656037)
		prime64  = uint64(1099511628211)
	)
	hash := offset64
	mix := func(value uint64) {
		for range 8 {
			hash ^= value & 0xff
			hash *= prime64
			value >>= 8
		}
	}
	mix(uint64(len(core.ActiveInheritedStylesReadOnly(ctx))))
	for _, value := range core.ActiveInheritedStylesReadOnly(ctx) {
		mix(value.Hash64())
	}
	mix(uint64(len(core.ActiveStylesReadOnly(ctx))))
	for _, value := range core.ActiveStylesReadOnly(ctx) {
		mix(value.Hash64())
	}
	mix(custom.Hash64())
	return hash
}

// datePickerPartDefaultsCache stores the two calendar part declarations after they have
// been joined with the theme declaration. A calendar frame revisits the same
// small set of boolean combinations for every cell; rebuilding the part map
// for each one otherwise dominates the style resolver's allocation path.
type datePickerPartDefaultsCache struct {
	theme           *theme.Theme
	activeVariant   field.Variant
	visualCache     computed.ProviderCache[field.Variant, theme.DatePickerVisual]
	activeVisual    theme.DatePickerVisual
	activeBundleSet bool
	cell            [256]style.Style
	cellValid       [256]bool
	rangeStyle      [2]style.Style
	rangeValid      [2]bool
}

func (c *datePickerPartDefaultsCache) resetFor(activeTheme *theme.Theme) {
	if c.theme == activeTheme {
		return
	}
	c.theme = activeTheme
	c.activeVariant = field.Primary
	c.activeBundleSet = false
	for i := range c.cellValid {
		c.cellValid[i] = false
	}
	for i := range c.rangeValid {
		c.rangeValid[i] = false
	}
}

func (c *datePickerPartDefaultsCache) prepareVariant(activeTheme *theme.Theme, variant field.Variant) field.Variant {
	c.resetFor(activeTheme)
	key := variant
	if key > field.Underlined {
		key = field.Primary
	}
	if c.activeVariant != key {
		c.activeVariant = key
		c.activeBundleSet = false
		for i := range c.cellValid {
			c.cellValid[i] = false
		}
		for i := range c.rangeValid {
			c.rangeValid[i] = false
		}
	}
	return key
}

func (c *datePickerPartDefaultsCache) visualFor(activeTheme *theme.Theme, variant field.Variant) theme.DatePickerVisual {
	key := c.prepareVariant(activeTheme, variant)
	visual := c.visualCache.Resolve(activeTheme, key, func() theme.DatePickerVisual {
		return datePickerVisualDeclarationForVariant(activeTheme, key)
	})
	c.activeVisual = visual
	c.activeBundleSet = true
	return visual
}

// visualForRequest resolves the complete bundle needed by one layout frame.
// Instance overrides intentionally bypass visualCache, but the returned
// bundle is retained for the rest of this frame so calendar parts do not
// issue a second Provider request for the same visual.
func (c *datePickerPartDefaultsCache) visualForRequest(activeTheme *theme.Theme, variant field.Variant, override *theme.DatePickerOverride) theme.DatePickerVisual {
	key := c.prepareVariant(activeTheme, variant)
	if override == nil {
		return c.visualFor(activeTheme, key)
	}
	visual := resolveDatePickerVisual(activeTheme, key, override)
	c.activeVisual = visual
	c.activeBundleSet = true
	return visual
}

func (c *datePickerPartDefaultsCache) currentVisual(activeTheme *theme.Theme) theme.DatePickerVisual {
	c.resetFor(activeTheme)
	if c.activeBundleSet {
		return c.activeVisual
	}
	return c.visualFor(activeTheme, c.activeVariant)
}

func (c *datePickerPartDefaultsCache) motion(activeTheme *theme.Theme) theme.DatePickerMotion {
	return c.currentVisual(activeTheme).Motion
}

func (c *datePickerPartDefaultsCache) cellDefaults(activeTheme *theme.Theme, base style.Style, context CellStyleContext) style.Style {
	c.resetFor(activeTheme)
	key := datePickerCellContextKey(context)
	if c.cellValid[key] {
		return c.cell[key]
	}
	declaration := datePickerCellDeclarationFromVisual(c.currentVisual(activeTheme), context)
	value := style.JoinPart(base, style.Style{}.
		Part(PartCalendarCell, declaration), PartCalendarCell)
	c.cell[key] = value
	c.cellValid[key] = true
	return value
}

func (c *datePickerPartDefaultsCache) rangeDefaults(activeTheme *theme.Theme, base style.Style, outside bool) style.Style {
	c.resetFor(activeTheme)
	key := uint8(0)
	if outside {
		key = 1
	}
	if c.rangeValid[key] {
		return c.rangeStyle[key]
	}
	declaration := datePickerRangeDeclarationFromVisual(c.currentVisual(activeTheme), outside)
	value := style.JoinPart(base, style.Style{}.
		Part(PartCalendarRange, declaration), PartCalendarRange)
	c.rangeStyle[key] = value
	c.rangeValid[key] = true
	return value
}

func (s *datePickerState) beginFrame() {
	core.BeginFrameMap(&s.frameDays)
	core.BeginFrameMap(&s.frameMonths)
	core.BeginFrameMap(&s.frameYears)
}

func (s *datePickerState) endFrame() {
	core.SweepFrameMap(s.days, s.frameDays)
	core.SweepFrameMap(s.months, s.frameMonths)
	core.SweepFrameMap(s.years, s.frameYears)
}

func (s *datePickerState) sync(value, initialMonth time.Time) {
	value = dateOnly(value)
	s.segments.sync(value)
	if !s.monthReady {
		s.viewMonth = firstOfMonth(initialMonth)
		s.syncedValue = value
		s.monthReady = true
		return
	}
	if sameDate(value, s.syncedValue) {
		return
	}
	s.syncedValue = value
	if !value.IsZero() {
		s.viewMonth = firstOfMonth(value)
	}
}

func (s *datePickerState) toggleYearPicker(year int) {
	if s.viewMode == datePickerViewYears {
		s.viewMode = datePickerViewDays
		return
	}
	s.viewMode = datePickerViewYears
	s.yearScrollYear = year
	s.yearScrollReady = true
}

func (s *datePickerState) openCalendar() {
	if !s.open {
		s.viewMode = datePickerViewDays
	}
	s.open = true
}

func (s *datePickerState) move(delta int) {
	switch s.viewMode {
	case datePickerViewMonths:
		s.viewMonth = firstOfMonth(s.viewMonth.AddDate(delta, 0, 0))
	case datePickerViewYears:
		s.viewMonth = firstOfMonth(s.viewMonth.AddDate(delta*datePickerYearSpan, 0, 0))
	default:
		s.viewMonth = firstOfMonth(s.viewMonth.AddDate(0, delta, 0))
	}
}

func (s *datePickerState) updateFocus(focused, disabled bool) {
	if disabled {
		s.open = false
		return
	}
	if !focused {
		s.open = false
	}
}

func (s *datePickerState) updateKeys(gtx layout.Context, focus event.Tag) {
	filters := []event.Filter{
		key.Filter{Focus: focus, Name: key.NameReturn},
		key.Filter{Focus: focus, Name: key.NameEnter},
		key.Filter{Focus: focus, Name: key.NameSpace},
	}
	if s.open {
		filters = append(filters, key.Filter{Focus: focus, Name: key.NameEscape})
	}
	for {
		e, ok := gtx.Event(filters...)
		if !ok {
			return
		}
		event, ok := e.(key.Event)
		if !ok || event.State != key.Press {
			continue
		}
		switch event.Name {
		case key.NameReturn, key.NameEnter, key.NameSpace:
			s.openCalendar()
		case key.NameEscape:
			s.open = false
		}
	}
}

func (s *datePickerState) day(key string) *datePickerCellState {
	return core.UseFrameMap(&s.days, &s.frameDays, key)
}

func (s *datePickerState) month(key string) *datePickerCellState {
	return core.UseFrameMap(&s.months, &s.frameMonths, key)
}

func (s *datePickerState) year(key string) *datePickerCellState {
	return core.UseFrameMap(&s.years, &s.frameYears, key)
}

func (s *datePickerState) popoverProgress(gtx layout.Context, open bool, motion theme.DatePickerMotion) float32 {
	target := float32(0)
	duration := motion.PopoverOut
	if open {
		target = 1
		duration = motion.PopoverIn
	}
	return s.popoverTransition.Value(gtx, target, duration, animation.EasingFor(motion.Policy.Ease), motion.Policy)
}

func (s *datePickerState) navBackground(gtx layout.Context, delta int, target color.NRGBA, motion theme.DatePickerMotion) color.NRGBA {
	if delta < 0 {
		return s.prevBackground.Value(gtx, target, motion.CellColor, animation.EasingFor(motion.Policy.Ease), motion.Policy)
	}
	return s.nextBackground.Value(gtx, target, motion.CellColor, animation.EasingFor(motion.Policy.Ease), motion.Policy)
}

type datePickerCellState struct {
	clickable            widget.Clickable
	date                 time.Time
	backgroundTransition animation.ColorTransition
}

func (s *datePickerCellState) background(gtx layout.Context, target color.NRGBA, motion theme.DatePickerMotion) color.NRGBA {
	return s.backgroundTransition.Value(gtx, target, motion.CellColor, animation.EasingFor(motion.Policy.Ease), motion.Policy)
}

func (s *datePickerState) hoveredDay() time.Time {
	for _, day := range s.days {
		if day.clickable.Hovered() {
			return day.date
		}
	}
	return time.Time{}
}

func (s *datePickerState) calendarFocused(gtx layout.Context) bool {
	return gtx.Focused(&s.header) ||
		gtx.Focused(&s.prev) ||
		gtx.Focused(&s.next) ||
		datePickerCellsFocused(gtx, s.days) ||
		datePickerCellsFocused(gtx, s.months) ||
		datePickerCellsFocused(gtx, s.years)
}

func (s *datePickerState) focusVisible(ctx *core.Context, gtx layout.Context) bool {
	return s.segments.focusVisible(ctx, gtx) ||
		core.FocusVisible(ctx, &s.trigger, gtx.Focused(&s.trigger)) ||
		core.FocusVisible(ctx, &s.header, gtx.Focused(&s.header)) ||
		core.FocusVisible(ctx, &s.prev, gtx.Focused(&s.prev)) ||
		core.FocusVisible(ctx, &s.next, gtx.Focused(&s.next)) ||
		datePickerCellsFocusVisible(ctx, gtx, s.days) ||
		datePickerCellsFocusVisible(ctx, gtx, s.months) ||
		datePickerCellsFocusVisible(ctx, gtx, s.years)
}

func (s *datePickerState) calendarEscapePressed(gtx layout.Context) bool {
	return datePickerEscapePressed(gtx, &s.header) ||
		datePickerEscapePressed(gtx, &s.prev) ||
		datePickerEscapePressed(gtx, &s.next) ||
		datePickerCellsEscapePressed(gtx, s.days) ||
		datePickerCellsEscapePressed(gtx, s.months) ||
		datePickerCellsEscapePressed(gtx, s.years)
}

func datePickerCellsFocused(gtx layout.Context, cells map[string]*datePickerCellState) bool {
	for _, cell := range cells {
		if gtx.Focused(&cell.clickable) {
			return true
		}
	}
	return false
}

func datePickerCellsFocusVisible(ctx *core.Context, gtx layout.Context, cells map[string]*datePickerCellState) bool {
	for _, cell := range cells {
		if core.FocusVisible(ctx, &cell.clickable, gtx.Focused(&cell.clickable)) {
			return true
		}
	}
	return false
}

func datePickerCellsEscapePressed(gtx layout.Context, cells map[string]*datePickerCellState) bool {
	for _, cell := range cells {
		if datePickerEscapePressed(gtx, &cell.clickable) {
			return true
		}
	}
	return false
}

func datePickerEscapePressed(gtx layout.Context, target event.Tag) bool {
	for {
		value, ok := gtx.Event(key.Filter{Focus: target, Name: key.NameEscape})
		if !ok {
			return false
		}
		if eventValue, ok := value.(key.Event); ok && eventValue.State == key.Press {
			return true
		}
	}
}

func datePickerPressScale(gtx layout.Context, history []widget.Press, disabled bool, motion theme.DatePickerMotion) float32 {
	target := motion.PressScale
	if target <= 0 || target > 1 {
		target = 1
	}
	return optionrow.PressScaleWithEasing(
		gtx,
		history,
		disabled,
		target,
		motion.PressIn,
		motion.PressOut,
		animation.EasingFor(motion.Policy.Ease),
		motion.Policy,
	)
}

// Disclosure helpers
func datePickerDisclosureCfg(widget Widget) disclosure.Config[time.Time] {
	return disclosure.Config[time.Time]{
		Controlled: widget.hasValue,
		Value:      widget.value,
		HasDefault: widget.hasDefault,
		Default:    widget.defaultValue,
		OnChange:   widget.onChange,
	}
}

func (s *datePickerState) currentValue(widget Widget) time.Time {
	return s.disclosure.Current(datePickerDisclosureCfg(widget))
}

func (s *datePickerState) bind(widget Widget) {
	s.disclosure.Bind(datePickerDisclosureCfg(widget))
}

func (s *datePickerState) requestValue(widget Widget, value time.Time) time.Time {
	newValue, _ := s.disclosure.Request(datePickerDisclosureCfg(widget), value)
	return newValue
}
