package datepicker

import (
	"image"
	"time"

	"gioui.org/io/pointer"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"github.com/qianniancn/FlowUI/internal/components/description"
	"github.com/qianniancn/FlowUI/internal/components/label"
	layoutui "github.com/qianniancn/FlowUI/internal/components/layout"
	"github.com/qianniancn/FlowUI/internal/components/text"
	"github.com/qianniancn/FlowUI/internal/field"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/locale"
	"github.com/qianniancn/FlowUI/internal/state"
	flowstyle "github.com/qianniancn/FlowUI/internal/style"
)

const stateSlotDateRangePicker = "date-range-picker"

type DateRange struct {
	Start time.Time
	End   time.Time
}

type DateRangePickerWidget struct {
	key          string
	value        DateRange
	label        string
	description  string
	errorMessage string
	locale       DatePickerLocale
	localeSet    bool
	onChange     func(DateRange)
	customStyle  flowstyle.Style
	variant      field.Variant
	disabled     bool
	invalid      bool
	required     bool
	fullWidth    bool
	minDate      time.Time
	maxDate      time.Time
}

type dateRangePickerState struct {
	start        dateSegmentsState
	end          dateSegmentsState
	calendar     datePickerState
	syncedValue  DateRange
	pendingStart time.Time
	selectingEnd bool
	ready        bool
	hover        dateInputHoverState
}

func DateRangePicker(key string, value DateRange) DateRangePickerWidget {
	return DateRangePickerWidget{key: key, value: normalizeDateRange(value)}
}

func (d DateRangePickerWidget) Label(value string) DateRangePickerWidget {
	d.label = value
	return d
}

func (d DateRangePickerWidget) Description(value string) DateRangePickerWidget {
	d.description = value
	return d
}

func (d DateRangePickerWidget) ErrorMessage(value string) DateRangePickerWidget {
	d.errorMessage = value
	return d
}

func (d DateRangePickerWidget) Locale(value DatePickerLocale) DateRangePickerWidget {
	d.locale = normalizeDatePickerLocale(value)
	d.localeSet = true
	return d
}

func (d DateRangePickerWidget) OnChange(fn func(DateRange)) DateRangePickerWidget {
	d.onChange = fn
	return d
}

func (d DateRangePickerWidget) Variant(value field.Variant) DateRangePickerWidget {
	d.variant = value
	return d
}

func (d DateRangePickerWidget) Disabled(value bool) DateRangePickerWidget {
	d.disabled = value
	return d
}

func (d DateRangePickerWidget) Invalid(value bool) DateRangePickerWidget {
	d.invalid = value
	return d
}

func (d DateRangePickerWidget) Required(value bool) DateRangePickerWidget {
	d.required = value
	return d
}

func (d DateRangePickerWidget) FullWidth() DateRangePickerWidget {
	d.fullWidth = true
	return d
}

func (d DateRangePickerWidget) MinDate(value time.Time) DateRangePickerWidget {
	d.minDate = dateOnly(value)
	return d
}

func (d DateRangePickerWidget) MaxDate(value time.Time) DateRangePickerWidget {
	d.maxDate = dateOnly(value)
	return d
}

func (d DateRangePickerWidget) Style(value flowstyle.Style) DateRangePickerWidget {
	d.customStyle = value
	return d
}

func (d DateRangePickerWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	d = d.resolveLocale(ctx)
	key := frame.ClaimKey(ctx, state.KindDateRangePicker, d.key)
	componentState := frame.UseState[dateRangePickerState](ctx, key, stateSlotDateRangePicker)
	now := datePickerFrameNow(gtx.Now)
	componentState.calendar.beginFrame()
	componentState.sync(d.value, d.initialMonth(now))
	componentState.hover.update(gtx)

	enabled := gtx.Enabled() && !d.disabled
	frame.RegisterFieldFocus(ctx, key, &componentState.start.segments[d.locale.DateOrder[0]].clickable, enabled)
	focused := componentState.focused(gtx)
	inputFocused := componentState.inputFocused(gtx)
	if !enabled {
		componentState.calendar.open = false
		componentState.cancelSelection(d.value)
	} else if componentState.calendar.open && !focused {
		componentState.calendar.open = false
		componentState.cancelSelection(d.value)
	}
	componentState.updateEscape(gtx, d.value)
	if enabled && !componentState.calendar.open {
		componentState.calendar.updateKeys(gtx, &componentState.calendar.trigger)
	}

	invalid := d.invalid ||
		!componentState.start.valid ||
		!componentState.end.valid ||
		invalidDateRange(d.value, d.minDate, d.maxDate)
	styleState := flowstyle.StyleState{
		Hovered:      componentState.hovered(),
		Focused:      focused,
		FocusVisible: componentState.start.focusVisible(ctx, gtx) || componentState.end.focusVisible(ctx, gtx) || componentState.calendar.focusVisible(ctx, gtx),
		Disabled:     !enabled,
		Invalid:      invalid,
		Selected:     !d.value.Start.IsZero() || !d.value.End.IsZero(),
		Open:         componentState.calendar.open,
	}
	fieldState := styleState
	fieldState.Hovered = componentState.hovered() && !focused
	fieldState.Focused = inputFocused
	fieldState.FocusVisible = componentState.start.focusVisible(ctx, gtx) || componentState.end.focusVisible(ctx, gtx)
	tokens := frame.ActiveTheme(ctx).Components
	resolved := field.Resolve(ctx, gtx, key, fieldState, d.variant, field.DeclarationOptions{
		Radius:         tokens.DatePicker.Radius,
		FocusRingWidth: tokens.Input.FocusRingWidth, InvalidOutlineWidth: tokens.Input.InvalidOutlineWidth,
		ShadowColor: tokens.Input.ShadowColor, ShadowOpacity: tokens.Input.ShadowOpacity,
		ShadowStrength: tokens.Input.ShadowStrength,
	}, d.customStyle)
	var dims layout.Dimensions
	var anchor image.Rectangle
	dims = layoutui.LayoutStyled(ctx, gtx, key, styleState, d.customStyle, frame.WidgetFunc(func(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
		dims, anchor = d.layoutField(ctx, gtx, componentState, resolved, enabled, invalid)
		return dims
	}))

	progress := componentState.calendar.popoverProgress(gtx, componentState.calendar.open && enabled, frame.ActiveTheme(ctx).Motion)
	if progress == 0 && (!componentState.calendar.open || !enabled) {
		componentState.calendar.endFrame()
		return dims
	}
	calendar := d.calendar(componentState)
	calendar.layoutPopover(ctx, gtx, &componentState.calendar, anchor, progress, now, frame.OverlayNaturallyDisabled(gtx))
	frame.AfterOverlays(ctx, componentState.calendar.endFrame)
	return dims
}

func (d DateRangePickerWidget) resolveLocale(ctx *frame.Context) DateRangePickerWidget {
	if !d.localeSet {
		if frame.ActiveLanguage(ctx) == locale.LanguageChinese {
			d.locale = datePickerChinese()
		} else {
			d.locale = datePickerEnglish()
		}
	}
	d.locale = normalizeDatePickerLocale(d.locale)
	return d
}

func (d DateRangePickerWidget) initialMonth(now time.Time) time.Time {
	value := d.value.Start
	if value.IsZero() {
		value = d.value.End
	}
	return DatePickerWidget{value: value, minDate: d.minDate, maxDate: d.maxDate}.initialMonth(now)
}

func (d DateRangePickerWidget) layoutField(ctx *frame.Context, gtx layout.Context, componentState *dateRangePickerState, style field.Resolved, enabled, invalid bool) (layout.Dimensions, image.Rectangle) {
	var children [3]layout.FlexChild
	count := 0
	labelHeight := 0
	inputAnchor := image.Rectangle{}
	gap := gtx.Dp(frame.ActiveTheme(ctx).Components.DatePicker.FieldGap)
	if d.label != "" {
		children[count] = layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			dimensions := label.Label(d.label).
				For(d.key).
				Required(d.required).
				Disabled(!enabled).
				Invalid(invalid).
				Layout(ctx, gtx)
			labelHeight = dimensions.Size.Y
			return dimensions
		})
		count++
	}
	children[count] = layout.Rigid(func(gtx layout.Context) layout.Dimensions {
		dimensions := d.layoutInput(ctx, gtx, componentState, style, enabled, invalid)
		y := labelHeight
		if d.label != "" {
			y += gap
		}
		inputAnchor = image.Rectangle{Min: image.Pt(0, y), Max: image.Pt(dimensions.Size.X, y+dimensions.Size.Y)}
		addDateInputHover(gtx, &componentState.hover, dimensions.Size, enabled, false)
		return dimensions
	})
	count++
	if invalid && d.errorMessage != "" {
		children[count] = layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return text.New(d.errorMessage).
				Size(float32(frame.ActiveTheme(ctx).Components.Description.TextSize)).
				Color(frame.ActiveTheme(ctx).Palette.Danger).
				Layout(ctx, gtx)
		})
		count++
	} else if d.description != "" {
		children[count] = layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return description.Description(d.description).
				For(d.key).
				Disabled(!enabled).
				Layout(ctx, gtx)
		})
		count++
	}
	dimensions := layout.Flex{
		Axis: layout.Vertical,
		Gap:  gap,
	}.Layout(gtx, children[:count]...)
	return dimensions, inputAnchor
}

func (d DateRangePickerWidget) layoutInput(ctx *frame.Context, gtx layout.Context, componentState *dateRangePickerState, style field.Resolved, enabled, invalid bool) layout.Dimensions {
	return layoutui.LayoutResolved(ctx, gtx, style.Content, frame.WidgetFunc(func(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
		return d.layoutInputContent(ctx, gtx, componentState, style.Colors, enabled, invalid)
	}))
}

func (d DateRangePickerWidget) layoutInputContent(ctx *frame.Context, gtx layout.Context, componentState *dateRangePickerState, style field.Colors, enabled, invalid bool) layout.Dimensions {
	presses := state.ActivePresses(componentState.calendar.trigger.History())
	focusVisible := frame.FocusVisible(ctx, &componentState.calendar.trigger, gtx.Focused(&componentState.calendar.trigger))
	if enabled {
		for componentState.calendar.trigger.Clicked(gtx) {
			wasOpen := componentState.calendar.open
			componentState.calendar.open = !wasOpen
			if componentState.calendar.open {
				componentState.calendar.viewMode = datePickerViewDays
			} else if wasOpen {
				componentState.cancelSelection(d.value)
			}
			frame.RequestFocusVisible(ctx, &componentState.calendar.trigger, focusVisible)
		}
		frame.FocusOnPress(ctx, &componentState.calendar.trigger, componentState.calendar.trigger.History(), presses)
	}

	frameConstraints := gtx.Constraints
	if d.fullWidth {
		frameConstraints.Min.X = frameConstraints.Max.X
	}
	tokens := frame.ActiveTheme(ctx).Components.DatePicker
	height := min(gtx.Dp(tokens.Height), frameConstraints.Max.Y)
	frameConstraints.Min.Y = min(max(frameConstraints.Min.Y, height), frameConstraints.Max.Y)
	left := gtx.Dp(frame.ActiveTheme(ctx).Components.Input.PaddingX)
	right := gtx.Dp(tokens.TriggerWidth)
	maxX := max(frameConstraints.Max.X-left-right, 0)
	minX := min(max(frameConstraints.Min.X-left-right, 0), maxX)

	macro := op.Record(gtx.Ops)
	contentGtx := gtx
	contentGtx.Constraints = layout.Constraints{Min: image.Pt(minX, 0), Max: image.Pt(maxX, height)}
	contentDims := layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(contentGtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return componentState.start.layout(ctx, gtx, d.locale, style, enabled, invalid, d.minDate, d.maxDate, func(value time.Time) {
				componentState.changeStart(value, d.value, d.onChange)
			})
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return d.layoutRangeSeparator(ctx, gtx, style)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return componentState.end.layout(ctx, gtx, d.locale, style, enabled, invalid, d.minDate, d.maxDate, func(value time.Time) {
				componentState.changeEnd(value, d.value, d.onChange)
			})
		}),
	)
	call := macro.Stop()

	size := frameConstraints.Constrain(image.Pt(contentDims.Size.X+left+right, height))
	offset := op.Offset(image.Pt(left, max((size.Y-contentDims.Size.Y)/2, 0))).Push(gtx.Ops)
	contentClip := clip.Rect{Max: image.Pt(max(size.X-left-right, 0), contentDims.Size.Y)}.Push(gtx.Ops)
	call.Add(gtx.Ops)
	contentClip.Pop()
	offset.Pop()

	triggerSize := image.Pt(right, size.Y)
	triggerGtx := gtx
	triggerGtx.Constraints = layout.Exact(triggerSize)
	offset = op.Offset(image.Pt(size.X-right, 0)).Push(gtx.Ops)
	if !enabled {
		triggerGtx = triggerGtx.Disabled()
	}
	componentState.calendar.trigger.Layout(triggerGtx, func(gtx layout.Context) layout.Dimensions {
		clipped := clip.Rect{Max: triggerSize}.Push(gtx.Ops)
		if enabled {
			pointer.CursorPointer.Add(gtx.Ops)
		}
		semantic.Button.Add(gtx.Ops)
		semantic.LabelOp(d.calendarLabel(ctx)).Add(gtx.Ops)
		semantic.EnabledOp(enabled).Add(gtx.Ops)
		iconSize := image.Pt(gtx.Dp(tokens.IconSize), gtx.Dp(tokens.IconSize))
		iconOffset := op.Offset(image.Pt((right-iconSize.X)/2, (size.Y-iconSize.Y)/2)).Push(gtx.Ops)
		drawDatePickerCalendarIcon(gtx, iconSize, style.Placeholder)
		iconOffset.Pop()
		drawDatePickerTriggerFocus(
			gtx,
			frame.ActiveTheme(ctx),
			triggerSize,
			enabled && frame.FocusVisible(ctx, &componentState.calendar.trigger, gtx.Focused(&componentState.calendar.trigger)),
		)
		clipped.Pop()
		return layout.Dimensions{Size: triggerSize}
	})
	offset.Pop()
	return layout.Dimensions{Size: size}
}

func (d DateRangePickerWidget) layoutRangeSeparator(ctx *frame.Context, gtx layout.Context, style field.Colors) layout.Dimensions {
	tokens := frame.ActiveTheme(ctx).Components.DatePicker
	size := image.Pt(gtx.Dp(tokens.RangeSeparatorSize), min(gtx.Dp(tokens.SegmentHeight), gtx.Constraints.Max.Y))
	gtx.Constraints = layout.Exact(size)
	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return text.New("–").Size(float32(tokens.TextSize)).Color(style.Placeholder).Layout(ctx, gtx)
	})
}

func (d DateRangePickerWidget) calendar(componentState *dateRangePickerState) DatePickerWidget {
	start, end := componentState.displayRange(d.value, componentState.calendar.hoveredDay())
	return DatePickerWidget{
		key:        d.key,
		value:      start,
		locale:     d.locale,
		localeSet:  true,
		minDate:    d.minDate,
		maxDate:    d.maxDate,
		rangeMode:  true,
		rangeStart: start,
		rangeEnd:   end,
		onDateSelect: func(value time.Time) {
			componentState.selectDate(value, d.onChange)
		},
	}
}

func (d DateRangePickerWidget) calendarLabel(ctx *frame.Context) string {
	if frame.ActiveLanguage(ctx) == locale.LanguageChinese {
		return "选择日期范围"
	}
	return "Choose date range"
}

func (s *dateRangePickerState) sync(value DateRange, initialMonth time.Time) {
	value = normalizeDateRange(value)
	if s.ready && sameDateRange(value, s.syncedValue) {
		if !s.selectingEnd {
			s.start.sync(value.Start)
			s.end.sync(value.End)
		}
		return
	}
	s.ready = true
	s.syncedValue = value
	s.start.sync(value.Start)
	s.end.sync(value.End)
	s.pendingStart = time.Time{}
	s.selectingEnd = !value.Start.IsZero() && value.End.IsZero()
	if s.selectingEnd {
		s.pendingStart = value.Start
	}
	anchor := value.Start
	if anchor.IsZero() {
		anchor = value.End
	}
	s.calendar.sync(anchor, initialMonth)
}

func (s *dateRangePickerState) focused(gtx layout.Context) bool {
	return s.inputFocused(gtx) || gtx.Focused(&s.calendar.trigger) || s.calendar.calendarFocused(gtx)
}

func (s *dateRangePickerState) inputFocused(gtx layout.Context) bool {
	return s.start.focused(gtx) || s.end.focused(gtx)
}

func (s *dateRangePickerState) hovered() bool {
	return s.hover.hovered || s.start.hovered() || s.end.hovered() || s.calendar.trigger.Hovered()
}

func (s *dateRangePickerState) updateEscape(gtx layout.Context, current DateRange) {
	if !s.calendar.open {
		return
	}
	if s.start.escapePressed(gtx) || s.end.escapePressed(gtx) || s.calendar.calendarEscapePressed(gtx) {
		s.calendar.open = false
		s.cancelSelection(current)
	}
}

func (s *dateRangePickerState) changeStart(value time.Time, current DateRange, onChange func(DateRange)) {
	next := normalizeDateRange(current)
	next.Start = dateOnly(value)
	s.syncedValue = next
	s.pendingStart = next.Start
	s.selectingEnd = !next.Start.IsZero() && next.End.IsZero()
	if onChange != nil {
		onChange(next)
	}
}

func (s *dateRangePickerState) changeEnd(value time.Time, current DateRange, onChange func(DateRange)) {
	next := normalizeDateRange(current)
	next.End = dateOnly(value)
	s.syncedValue = next
	s.selectingEnd = false
	s.pendingStart = time.Time{}
	if onChange != nil {
		onChange(next)
	}
}

func (s *dateRangePickerState) selectDate(value time.Time, onChange func(DateRange)) {
	value = dateOnly(value)
	if !s.selectingEnd || s.pendingStart.IsZero() {
		s.pendingStart = value
		s.selectingEnd = true
		s.start.sync(value)
		s.end.sync(time.Time{})
		s.calendar.sync(value, value)
		return
	}

	next := DateRange{Start: s.pendingStart, End: value}
	if compareDate(next.Start, next.End) > 0 {
		next.Start, next.End = next.End, next.Start
	}
	s.syncedValue = next
	s.pendingStart = time.Time{}
	s.selectingEnd = false
	s.start.sync(next.Start)
	s.end.sync(next.End)
	s.calendar.open = false
	s.calendar.sync(next.Start, next.Start)
	if onChange != nil {
		onChange(next)
	}
}

func (s *dateRangePickerState) cancelSelection(current DateRange) {
	current = normalizeDateRange(current)
	s.pendingStart = time.Time{}
	s.selectingEnd = !current.Start.IsZero() && current.End.IsZero()
	if s.selectingEnd {
		s.pendingStart = current.Start
	}
	s.start.sync(current.Start)
	s.end.sync(current.End)
	anchor := current.Start
	if anchor.IsZero() {
		anchor = current.End
	}
	s.calendar.sync(anchor, anchor)
}

func (s *dateRangePickerState) displayRange(current DateRange, hovered time.Time) (time.Time, time.Time) {
	if s.selectingEnd && !s.pendingStart.IsZero() {
		return s.pendingStart, dateOnly(hovered)
	}
	current = normalizeDateRange(current)
	return current.Start, current.End
}

func normalizeDateRange(value DateRange) DateRange {
	value.Start = dateOnly(value.Start)
	value.End = dateOnly(value.End)
	return value
}

func sameDateRange(first, second DateRange) bool {
	return sameDate(first.Start, second.Start) && sameDate(first.End, second.End)
}

func invalidDateRange(value DateRange, minDate, maxDate time.Time) bool {
	value = normalizeDateRange(value)
	return !value.Start.IsZero() && !value.End.IsZero() && compareDate(value.Start, value.End) > 0 ||
		dateOutsideRange(value.Start, minDate, maxDate) ||
		dateOutsideRange(value.End, minDate, maxDate)
}
