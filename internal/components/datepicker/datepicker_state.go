package datepicker

import (
	"image/color"
	"time"

	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
	"github.com/qianniancn/FlowUI/internal/field"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/render"
	"github.com/qianniancn/FlowUI/internal/state"
)

const stateSlotDatePicker = "datepicker"

func datePickerStateFor(ctx *frame.Context, key string) *datePickerState {
	key = frame.ClaimKey(ctx, state.KindDatePicker, key)
	return frame.UseState[datePickerState](ctx, key, stateSlotDatePicker)
}

type datePickerState struct {
	input           field.State
	trigger         widget.Clickable
	header          widget.Clickable
	prev            widget.Clickable
	next            widget.Clickable
	open            bool
	viewMode        datePickerViewMode
	viewMonth       time.Time
	syncedValue     time.Time
	monthReady      bool
	popover         float32
	popoverFrom     float32
	popoverTo       float32
	popoverAt       time.Time
	popoverDuration time.Duration
	popoverReady    bool
	prevBg          color.NRGBA
	prevBgFrom      color.NRGBA
	prevBgTo        color.NRGBA
	prevBgAt        time.Time
	prevBgReady     bool
	nextBg          color.NRGBA
	nextBgFrom      color.NRGBA
	nextBgTo        color.NRGBA
	nextBgAt        time.Time
	nextBgReady     bool
	yearList        layout.List
	yearScrollYear  int
	yearScrollReady bool
	days            map[string]*datePickerCellState
	months          map[string]*datePickerCellState
	years           map[string]*datePickerCellState
	frameDays       map[string]struct{}
	frameMonths     map[string]struct{}
	frameYears      map[string]struct{}
}

func (s *datePickerState) beginFrame() {
	if s.frameDays == nil {
		s.frameDays = make(map[string]struct{})
	} else {
		clear(s.frameDays)
	}
	if s.frameMonths == nil {
		s.frameMonths = make(map[string]struct{})
	} else {
		clear(s.frameMonths)
	}
	if s.frameYears == nil {
		s.frameYears = make(map[string]struct{})
	} else {
		clear(s.frameYears)
	}
}

func (s *datePickerState) endFrame() {
	for key := range s.days {
		if _, ok := s.frameDays[key]; !ok {
			delete(s.days, key)
		}
	}
	for key := range s.months {
		if _, ok := s.frameMonths[key]; !ok {
			delete(s.months, key)
		}
	}
	for key := range s.years {
		if _, ok := s.frameYears[key]; !ok {
			delete(s.years, key)
		}
	}
}

func (s *datePickerState) sync(value, initialMonth time.Time) {
	value = dateOnly(value)
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
	if s.days == nil {
		s.days = make(map[string]*datePickerCellState)
	}
	s.frameDays[key] = struct{}{}
	if day := s.days[key]; day != nil {
		return day
	}
	day := new(datePickerCellState)
	s.days[key] = day
	return day
}

func (s *datePickerState) month(key string) *datePickerCellState {
	if s.months == nil {
		s.months = make(map[string]*datePickerCellState)
	}
	s.frameMonths[key] = struct{}{}
	if month := s.months[key]; month != nil {
		return month
	}
	month := new(datePickerCellState)
	s.months[key] = month
	return month
}

func (s *datePickerState) year(key string) *datePickerCellState {
	if s.years == nil {
		s.years = make(map[string]*datePickerCellState)
	}
	s.frameYears[key] = struct{}{}
	if year := s.years[key]; year != nil {
		return year
	}
	year := new(datePickerCellState)
	s.years[key] = year
	return year
}

func (s *datePickerState) popoverProgress(gtx layout.Context, open bool) float32 {
	target := float32(0)
	duration := datePickerPopoverOutDuration
	if open {
		target = 1
		duration = datePickerPopoverInDuration
	}
	if !s.popoverReady {
		s.popover = target
		s.popoverFrom = target
		s.popoverTo = target
		s.popoverAt = gtx.Now
		s.popoverDuration = duration
		s.popoverReady = true
		return target
	}
	if target != s.popoverTo {
		s.popoverFrom = s.popover
		s.popoverTo = target
		s.popoverAt = gtx.Now
		s.popoverDuration = duration
	}
	if s.popoverFrom == s.popoverTo {
		s.popover = s.popoverTo
		return s.popover
	}
	progress := render.Ease(render.Progress(gtx.Now.Sub(s.popoverAt), s.popoverDuration))
	if progress < 1 {
		gtx.Execute(op.InvalidateCmd{})
	}
	s.popover = render.Lerp(s.popoverFrom, s.popoverTo, progress)
	return s.popover
}

func (s *datePickerState) navBackground(gtx layout.Context, delta int, target color.NRGBA) color.NRGBA {
	if delta < 0 {
		return datePickerColor(gtx, target, &s.prevBg, &s.prevBgFrom, &s.prevBgTo, &s.prevBgAt, &s.prevBgReady)
	}
	return datePickerColor(gtx, target, &s.nextBg, &s.nextBgFrom, &s.nextBgTo, &s.nextBgAt, &s.nextBgReady)
}

type datePickerCellState struct {
	clickable widget.Clickable
	bg        color.NRGBA
	bgFrom    color.NRGBA
	bgTo      color.NRGBA
	bgAt      time.Time
	bgReady   bool
}

func (s *datePickerCellState) background(gtx layout.Context, target color.NRGBA) color.NRGBA {
	return datePickerColor(gtx, target, &s.bg, &s.bgFrom, &s.bgTo, &s.bgAt, &s.bgReady)
}

func datePickerColor(gtx layout.Context, target color.NRGBA, value, from, to *color.NRGBA, at *time.Time, ready *bool) color.NRGBA {
	if !*ready {
		*value = target
		*from = target
		*to = target
		*at = gtx.Now
		*ready = true
		return target
	}
	if target != *to {
		*from = *value
		*to = target
		*at = gtx.Now
	}
	if *from == *to {
		*value = *to
		return *value
	}
	progress := render.Ease(render.Progress(gtx.Now.Sub(*at), datePickerCellColorDuration))
	if progress < 1 {
		gtx.Execute(op.InvalidateCmd{})
	}
	*value = render.LerpColor(*from, *to, progress)
	return *value
}

func datePickerPressScale(gtx layout.Context, history []widget.Press, disabled bool) float32 {
	if disabled || len(history) == 0 {
		return 1
	}
	press := history[len(history)-1]
	target := float32(0.95)
	if press.End.IsZero() {
		progress := render.Ease(render.Progress(gtx.Now.Sub(press.Start), datePickerPressInDuration))
		if progress < 1 {
			gtx.Execute(op.InvalidateCmd{})
		}
		return render.Lerp(1, target, progress)
	}
	progress := render.Ease(render.Progress(gtx.Now.Sub(press.End), datePickerPressOutDuration))
	if progress < 1 {
		gtx.Execute(op.InvalidateCmd{})
	}
	return render.Lerp(target, 1, progress)
}
