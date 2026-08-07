package datepicker

import (
	"image/color"
	"time"

	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/internal/animation"
	"github.com/qianniancn/flowui/internal/components/disclosure"
	layoutui "github.com/qianniancn/flowui/internal/components/layout"
	"github.com/qianniancn/flowui/internal/components/optionrow"
	"github.com/qianniancn/flowui/internal/frame"
	"github.com/qianniancn/flowui/internal/overlay"
	"github.com/qianniancn/flowui/internal/state"
	"github.com/qianniancn/flowui/internal/theme"
)

const stateSlotDatePicker = "datepicker"

func datePickerStateFor(ctx *frame.Context, key string) *datePickerState {
	key = frame.ClaimKey(ctx, state.KindDatePicker, key)
	return frame.UseState[datePickerState](ctx, key, stateSlotDatePicker)
}

type datePickerState struct {
	disclosure        disclosure.Binding[time.Time]
	segments          dateSegmentsState
	hover             dateInputHoverState
	trigger           widget.Clickable
	dialog            overlay.ClickArea
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
}

func (s *datePickerState) beginFrame() {
	state.BeginFrameMap(&s.frameDays)
	state.BeginFrameMap(&s.frameMonths)
	state.BeginFrameMap(&s.frameYears)
}

func (s *datePickerState) endFrame() {
	state.SweepFrameMap(s.days, s.frameDays)
	state.SweepFrameMap(s.months, s.frameMonths)
	state.SweepFrameMap(s.years, s.frameYears)
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
	return state.UseFrameMap(&s.days, &s.frameDays, key)
}

func (s *datePickerState) month(key string) *datePickerCellState {
	return state.UseFrameMap(&s.months, &s.frameMonths, key)
}

func (s *datePickerState) year(key string) *datePickerCellState {
	return state.UseFrameMap(&s.years, &s.frameYears, key)
}

func (s *datePickerState) popoverProgress(gtx layout.Context, open bool, motions ...theme.MotionTheme) float32 {
	target := float32(0)
	duration := datePickerPopoverOutDuration
	if open {
		target = 1
		duration = datePickerPopoverInDuration
	}
	return s.popoverTransition.Value(gtx, target, duration, animation.EaseSmoothstep, motions...)
}

func (s *datePickerState) navBackground(gtx layout.Context, delta int, target color.NRGBA, motions ...theme.MotionTheme) color.NRGBA {
	if delta < 0 {
		return s.prevBackground.Value(gtx, target, datePickerCellColorDuration, animation.EaseSmoothstep, motions...)
	}
	return s.nextBackground.Value(gtx, target, datePickerCellColorDuration, animation.EaseSmoothstep, motions...)
}

type datePickerCellState struct {
	clickable            widget.Clickable
	date                 time.Time
	backgroundTransition animation.ColorTransition
}

func (s *datePickerCellState) background(gtx layout.Context, target color.NRGBA, motions ...theme.MotionTheme) color.NRGBA {
	return s.backgroundTransition.Value(gtx, target, datePickerCellColorDuration, animation.EaseSmoothstep, motions...)
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

func (s *datePickerState) focusVisible(ctx *frame.Context, gtx layout.Context) bool {
	return s.segments.focusVisible(ctx, gtx) ||
		frame.FocusVisible(ctx, &s.trigger, gtx.Focused(&s.trigger)) ||
		frame.FocusVisible(ctx, &s.header, gtx.Focused(&s.header)) ||
		frame.FocusVisible(ctx, &s.prev, gtx.Focused(&s.prev)) ||
		frame.FocusVisible(ctx, &s.next, gtx.Focused(&s.next)) ||
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

func datePickerCellsFocusVisible(ctx *frame.Context, gtx layout.Context, cells map[string]*datePickerCellState) bool {
	for _, cell := range cells {
		if frame.FocusVisible(ctx, &cell.clickable, gtx.Focused(&cell.clickable)) {
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

func datePickerPressScale(gtx layout.Context, history []widget.Press, disabled bool, motions ...theme.MotionTheme) float32 {
	target := float32(0.95)
	return optionrow.PressScale(gtx, history, disabled, target, datePickerPressInDuration, datePickerPressOutDuration, motions...)
}

// Disclosure helpers
func datePickerDisclosureCfg(widget DatePickerWidget) disclosure.Config[time.Time] {
	return disclosure.Config[time.Time]{
		Controlled: widget.hasValue,
		Value:      widget.value,
		HasDefault: widget.hasDefault,
		Default:    widget.defaultValue,
		OnChange:   widget.onChange,
	}
}

func (s *datePickerState) currentValue(widget DatePickerWidget) time.Time {
	return s.disclosure.Current(datePickerDisclosureCfg(widget))
}

func (s *datePickerState) bind(widget DatePickerWidget) {
	s.disclosure.Bind(datePickerDisclosureCfg(widget))
}

func (s *datePickerState) requestValue(widget DatePickerWidget, value time.Time) time.Time {
	newValue, _ := s.disclosure.Request(datePickerDisclosureCfg(widget), value)
	return newValue
}
