package datepicker

import (
	"image"
	"strconv"
	"time"

	"gioui.org/font"
	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"github.com/qianniancn/FlowUI/internal/components/description"
	"github.com/qianniancn/FlowUI/internal/components/label"
	"github.com/qianniancn/FlowUI/internal/components/text"
	"github.com/qianniancn/FlowUI/internal/field"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/locale"
	"github.com/qianniancn/FlowUI/internal/state"
	"github.com/qianniancn/FlowUI/internal/theme"
)

const (
	stateSlotDateField = "date-field"
	dateSegmentTimeout = time.Second
)

type DateFieldWidget struct {
	theme        func(*theme.Theme)
	key          string
	value        time.Time
	label        string
	description  string
	errorMessage string
	locale       DatePickerLocale
	localeSet    bool
	onChange     func(time.Time)
	variant      field.Variant
	disabled     bool
	invalid      bool
	required     bool
	fullWidth    bool
	minDate      time.Time
	maxDate      time.Time
}

type dateFieldState struct {
	input    field.State
	segments dateSegmentsState
	hover    dateInputHoverState
}

type dateInputHoverState struct {
	hovered bool
}

const (
	dateSegmentYear  = DatePartYear
	dateSegmentMonth = DatePartMonth
	dateSegmentDay   = DatePartDay
)

type dateParts struct {
	year  int
	month int
	day   int
}

type dateSegmentsState struct {
	segments    [3]dateSegmentState
	list        layout.List
	parts       dateParts
	syncedValue time.Time
	location    *time.Location
	ready       bool
	valid       bool
}

type dateSegmentState struct {
	clickable widget.Clickable
	typed     string
	typedAt   time.Time
}

func DateField(key string, value time.Time) DateFieldWidget {
	return DateFieldWidget{key: key, value: dateOnly(value)}
}

func (d DateFieldWidget) Label(value string) DateFieldWidget {
	d.label = value
	return d
}

func (d DateFieldWidget) Description(value string) DateFieldWidget {
	d.description = value
	return d
}

func (d DateFieldWidget) ErrorMessage(value string) DateFieldWidget {
	d.errorMessage = value
	return d
}

func (d DateFieldWidget) Locale(value DatePickerLocale) DateFieldWidget {
	d.locale = normalizeDatePickerLocale(value)
	d.localeSet = true
	return d
}

func (d DateFieldWidget) OnChange(fn func(time.Time)) DateFieldWidget {
	d.onChange = fn
	return d
}

func (d DateFieldWidget) Variant(value field.Variant) DateFieldWidget {
	d.variant = value
	return d
}

func (d DateFieldWidget) Disabled(value bool) DateFieldWidget {
	d.disabled = value
	return d
}

func (d DateFieldWidget) Invalid(value bool) DateFieldWidget {
	d.invalid = value
	return d
}

func (d DateFieldWidget) Required(value bool) DateFieldWidget {
	d.required = value
	return d
}

func (d DateFieldWidget) FullWidth() DateFieldWidget {
	d.fullWidth = true
	return d
}

func (d DateFieldWidget) MinDate(value time.Time) DateFieldWidget {
	d.minDate = dateOnly(value)
	return d
}

func (d DateFieldWidget) MaxDate(value time.Time) DateFieldWidget {
	d.maxDate = dateOnly(value)
	return d
}

func (d DateFieldWidget) Theme(fn func(*theme.Theme)) DateFieldWidget {
	d.theme = fn
	return d
}

func (d DateFieldWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	if restore := frame.PushInstanceTheme(ctx, d.theme); restore != nil {
		defer restore()
	}
	d = d.resolveLocale(ctx)
	key := frame.ClaimKey(ctx, state.KindDateField, d.key)
	componentState := frame.UseState[dateFieldState](ctx, key, stateSlotDateField)
	componentState.segments.sync(d.value)
	componentState.hover.update(gtx)
	enabled := gtx.Enabled() && !d.disabled
	frame.RegisterFieldFocus(ctx, key, &componentState.segments.segments[d.locale.DateOrder[0]].clickable, enabled)

	focused := componentState.segments.focused(gtx)
	hovered := componentState.hover.hovered || componentState.segments.hovered()
	invalid := d.invalid || !componentState.segments.valid || dateOutsideRange(d.value, d.minDate, d.maxDate)
	style := field.ResolveStyle(frame.ActiveTheme(ctx), d.variant, hovered, focused, !enabled, invalid)
	style.Background = componentState.input.Background(gtx, style.Background, frame.ActiveTheme(ctx).Motion)
	style.Border = componentState.input.BorderColor(gtx, style.Border, frame.ActiveTheme(ctx).Motion)

	var children [3]layout.FlexChild
	count := 0
	if d.label != "" {
		children[count] = layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return label.Label(d.label).
				For(d.key).
				Required(d.required).
				Disabled(!enabled).
				Invalid(invalid).
				Layout(ctx, gtx)
		})
		count++
	}
	children[count] = layout.Rigid(func(gtx layout.Context) layout.Dimensions {
		dimensions := layoutDateInputFrame(ctx, gtx, style, d.fullWidth, func(gtx layout.Context) layout.Dimensions {
			return componentState.segments.layout(ctx, gtx, d.locale, style, enabled, invalid, d.minDate, d.maxDate, d.onChange)
		})
		addDateInputHover(gtx, &componentState.hover, dimensions.Size, enabled, true)
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

	return layout.Flex{
		Axis: layout.Vertical,
		Gap:  gtx.Dp(frame.ActiveTheme(ctx).Components.DatePicker.FieldGap),
	}.Layout(gtx, children[:count]...)
}

func (d DateFieldWidget) resolveLocale(ctx *frame.Context) DateFieldWidget {
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

func layoutDateInputFrame(ctx *frame.Context, gtx layout.Context, style field.Style, fullWidth bool, content layout.Widget) layout.Dimensions {
	frameConstraints := gtx.Constraints
	if fullWidth {
		frameConstraints.Min.X = frameConstraints.Max.X
	}
	tokens := frame.ActiveTheme(ctx).Components.DatePicker
	height := min(gtx.Dp(tokens.Height), frameConstraints.Max.Y)
	frameConstraints.Min.Y = min(max(frameConstraints.Min.Y, height), frameConstraints.Max.Y)
	padding := gtx.Dp(frame.ActiveTheme(ctx).Components.Input.PaddingX)
	maxX := max(frameConstraints.Max.X-padding*2, 0)
	minX := min(max(frameConstraints.Min.X-padding*2, 0), maxX)

	macro := op.Record(gtx.Ops)
	contentGtx := gtx
	contentGtx.Constraints = layout.Constraints{Min: image.Pt(minX, 0), Max: image.Pt(maxX, height)}
	contentDims := content(contentGtx)
	call := macro.Stop()
	size := frameConstraints.Constrain(image.Pt(contentDims.Size.X+padding*2, height))
	radius := min(max(gtx.Dp(tokens.Radius), 1), min(size.X, size.Y)/2)
	field.DrawFrame(gtx, image.Rectangle{Max: size}, radius, style)
	clipped := clip.UniformRRect(image.Rectangle{Max: size}, radius).Push(gtx.Ops)
	offset := op.Offset(image.Pt(padding, max((size.Y-contentDims.Size.Y)/2, 0))).Push(gtx.Ops)
	call.Add(gtx.Ops)
	offset.Pop()
	clipped.Pop()
	return layout.Dimensions{Size: size}
}

func (s *dateSegmentsState) sync(value time.Time) {
	value = dateOnly(value)
	if s.ready && sameDate(value, s.syncedValue) {
		return
	}
	s.ready = true
	s.valid = true
	s.syncedValue = value
	for index := range s.segments {
		s.segments[index].typed = ""
	}
	if value.IsZero() {
		s.parts = dateParts{}
		if s.location == nil {
			s.location = time.Local
		}
		return
	}
	s.location = value.Location()
	year, month, day := value.Date()
	s.parts = dateParts{year: year, month: int(month), day: day}
}

func (s *dateSegmentsState) focused(gtx layout.Context) bool {
	for index := range s.segments {
		if gtx.Focused(&s.segments[index].clickable) {
			return true
		}
	}
	return false
}

func (s *dateSegmentsState) hovered() bool {
	for index := range s.segments {
		if s.segments[index].clickable.Hovered() {
			return true
		}
	}
	return false
}

func (s *dateSegmentsState) empty() bool {
	return s.parts == (dateParts{})
}

func (s *dateSegmentsState) escapePressed(gtx layout.Context) bool {
	pressed := false
	for index := range s.segments {
		for {
			value, ok := gtx.Event(key.Filter{Focus: &s.segments[index].clickable, Name: key.NameEscape})
			if !ok {
				break
			}
			if eventValue, ok := value.(key.Event); ok && eventValue.State == key.Press {
				pressed = true
			}
		}
	}
	return pressed
}

func (s *dateSegmentsState) layout(ctx *frame.Context, gtx layout.Context, locale DatePickerLocale, style field.Style, enabled, invalid bool, minDate, maxDate time.Time, onChange func(time.Time)) layout.Dimensions {
	for _, part := range locale.DateOrder {
		s.updateSegment(ctx, gtx, int(part), locale.DateOrder, enabled, minDate, maxDate, onChange)
	}
	var children [7]layout.Widget
	partItems := [3]int{-1, -1, -1}
	count := 0
	addLiteral := func(value string) {
		if value == "" {
			return
		}
		children[count] = func(gtx layout.Context) layout.Dimensions {
			return layoutDateSegmentLiteral(ctx, gtx, value, style)
		}
		count++
	}
	addLiteral(locale.DateLiterals[0])
	for position, part := range locale.DateOrder {
		part := int(part)
		partItems[part] = count
		children[count] = func(gtx layout.Context) layout.Dimensions {
			return s.layoutSegment(ctx, gtx, part, style, enabled, invalid)
		}
		count++
		addLiteral(locale.DateLiterals[position+1])
	}
	if s.list.Position.Count > 0 {
		for part, item := range partItems {
			if item >= 0 && gtx.Focused(&s.segments[part].clickable) &&
				(item < s.list.Position.First || item >= s.list.Position.First+s.list.Position.Count) {
				s.list.ScrollTo(item)
			}
		}
	}
	s.list.Axis = layout.Horizontal
	s.list.Alignment = layout.Middle
	return s.list.Layout(gtx, count, func(gtx layout.Context, index int) layout.Dimensions {
		return children[index](gtx)
	})
}

func (s *dateSegmentsState) updateSegment(ctx *frame.Context, gtx layout.Context, index int, order [3]DatePart, enabled bool, minDate, maxDate time.Time, onChange func(time.Time)) {
	segment := &s.segments[index]
	presses := state.ActivePresses(segment.clickable.History())
	if enabled {
		for segment.clickable.Clicked(gtx) {
			segment.typed = ""
			frame.RequestFocus(ctx, &segment.clickable)
		}
		frame.FocusOnPress(ctx, &segment.clickable, segment.clickable.History(), presses)
	}
	if !enabled || !gtx.Focused(&segment.clickable) {
		return
	}

	var filters [17]event.Filter
	filterCount := 0
	for digit := 0; digit <= 9; digit++ {
		filters[filterCount] = key.Filter{Focus: &segment.clickable, Name: key.Name(strconv.Itoa(digit))}
		filterCount++
	}
	for _, name := range [...]key.Name{
		key.NameLeftArrow,
		key.NameRightArrow,
		key.NameUpArrow,
		key.NameDownArrow,
		key.NameDeleteBackward,
		key.NameDeleteForward,
		key.NameHome,
	} {
		filters[filterCount] = key.Filter{Focus: &segment.clickable, Name: name}
		filterCount++
	}
	for {
		value, ok := gtx.Event(filters[:filterCount]...)
		if !ok {
			return
		}
		eventValue, ok := value.(key.Event)
		if !ok || eventValue.State != key.Press {
			continue
		}
		switch eventValue.Name {
		case key.NameLeftArrow:
			position := datePartPosition(order, DatePart(index))
			s.focusSegment(ctx, int(order[max(position-1, 0)]))
		case key.NameRightArrow:
			position := datePartPosition(order, DatePart(index))
			s.focusSegment(ctx, int(order[min(position+1, len(order)-1)]))
		case key.NameUpArrow:
			s.adjust(index, 1, minDate, maxDate, onChange)
		case key.NameDownArrow:
			s.adjust(index, -1, minDate, maxDate, onChange)
		case key.NameDeleteBackward, key.NameDeleteForward:
			s.clear(index, onChange)
		case key.NameHome:
			s.setPart(index, s.minimum(index))
			s.dispatch(minDate, maxDate, onChange)
		default:
			if eventValue.Name >= "0" && eventValue.Name <= "9" {
				s.typeDigit(ctx, gtx.Now, index, order, byte(eventValue.Name[0]), minDate, maxDate, onChange)
			}
		}
	}
}

func (s *dateSegmentsState) focusSegment(ctx *frame.Context, index int) {
	s.segments[index].typed = ""
	frame.RequestFocus(ctx, &s.segments[index].clickable)
}

func (s *dateSegmentsState) typeDigit(ctx *frame.Context, now time.Time, index int, order [3]DatePart, digit byte, minDate, maxDate time.Time, onChange func(time.Time)) {
	segment := &s.segments[index]
	maxDigits := 2
	if index == int(dateSegmentYear) {
		maxDigits = 4
	}
	if segment.typed == "" || now.Sub(segment.typedAt) > dateSegmentTimeout || len(segment.typed) >= maxDigits {
		segment.typed = ""
	}
	segment.typed += string(digit)
	segment.typedAt = now
	value, _ := strconv.Atoi(segment.typed)
	s.setPart(index, value)
	s.dispatch(minDate, maxDate, onChange)

	advance := len(segment.typed) >= maxDigits
	if len(segment.typed) == 1 {
		advance = index == int(dateSegmentMonth) && value > 1 || index == int(dateSegmentDay) && value > 3
	}
	position := datePartPosition(order, DatePart(index))
	if advance && position < len(order)-1 {
		s.focusSegment(ctx, int(order[position+1]))
	}
}

func datePartPosition(order [3]DatePart, part DatePart) int {
	for position, candidate := range order {
		if candidate == part {
			return position
		}
	}
	return 0
}

func (s *dateSegmentsState) adjust(index, delta int, minDate, maxDate time.Time, onChange func(time.Time)) {
	s.segments[index].typed = ""
	value := s.part(index)
	minimum := s.minimum(index)
	maximum := s.maximum(index)
	if value == 0 {
		value = minimum
	} else {
		value += delta
		if value < minimum {
			value = maximum
		} else if value > maximum {
			value = minimum
		}
	}
	s.setPart(index, value)
	s.dispatch(minDate, maxDate, onChange)
}

func (s *dateSegmentsState) clear(index int, onChange func(time.Time)) {
	s.segments[index].typed = ""
	s.setPart(index, 0)
	s.valid = true
	if !s.syncedValue.IsZero() {
		s.syncedValue = time.Time{}
		if onChange != nil {
			onChange(time.Time{})
		}
	}
}

func (s *dateSegmentsState) dispatch(minDate, maxDate time.Time, onChange func(time.Time)) {
	value, complete, valid := s.date()
	if !complete {
		s.valid = true
		return
	}
	if valid && !minDate.IsZero() && compareDate(value, minDate) < 0 {
		valid = false
	}
	if valid && !maxDate.IsZero() && compareDate(value, maxDate) > 0 {
		valid = false
	}
	s.valid = valid
	if !valid || sameDate(value, s.syncedValue) {
		return
	}
	s.syncedValue = value
	if onChange != nil {
		onChange(value)
	}
}

func (s *dateSegmentsState) date() (time.Time, bool, bool) {
	if s.parts.year == 0 || s.parts.month == 0 || s.parts.day == 0 {
		return time.Time{}, false, true
	}
	if s.parts.year < 1 || s.parts.year > 9999 || s.parts.month < 1 || s.parts.month > 12 {
		return time.Time{}, true, false
	}
	location := s.location
	if location == nil {
		location = time.Local
	}
	value := time.Date(s.parts.year, time.Month(s.parts.month), s.parts.day, 0, 0, 0, 0, location)
	year, month, day := value.Date()
	valid := year == s.parts.year && int(month) == s.parts.month && day == s.parts.day
	return value, true, valid
}

func (s *dateSegmentsState) part(index int) int {
	switch DatePart(index) {
	case dateSegmentMonth:
		return s.parts.month
	case dateSegmentDay:
		return s.parts.day
	default:
		return s.parts.year
	}
}

func (s *dateSegmentsState) setPart(index, value int) {
	switch DatePart(index) {
	case dateSegmentMonth:
		s.parts.month = value
	case dateSegmentDay:
		s.parts.day = value
	default:
		s.parts.year = value
	}
}

func (s *dateSegmentsState) minimum(index int) int {
	return 1
}

func (s *dateSegmentsState) maximum(index int) int {
	switch DatePart(index) {
	case dateSegmentMonth:
		return 12
	case dateSegmentDay:
		year := s.parts.year
		if year == 0 {
			year = 2000
		}
		month := s.parts.month
		if month < 1 || month > 12 {
			return 31
		}
		return time.Date(year, time.Month(month)+1, 0, 0, 0, 0, 0, time.UTC).Day()
	default:
		return 9999
	}
}

func (s *dateSegmentsState) layoutSegment(ctx *frame.Context, gtx layout.Context, index int, style field.Style, enabled, invalid bool) layout.Dimensions {
	tokens := frame.ActiveTheme(ctx).Components.DatePicker
	width := gtx.Dp(tokens.SegmentWidth)
	if index == int(dateSegmentYear) {
		width = gtx.Dp(tokens.YearSegmentWidth)
	}
	segment := &s.segments[index]
	focused := gtx.Focused(&segment.clickable)
	value, placeholder := s.segmentText(index, focused)
	col := style.Foreground
	if placeholder {
		col = style.Placeholder
	}
	if invalid {
		col = frame.ActiveTheme(ctx).Palette.Danger
		if focused {
			col = frame.ActiveTheme(ctx).Palette.DangerSoftForeground
		}
	} else if focused {
		col = frame.ActiveTheme(ctx).Palette.AccentSoftForeground
	}
	textMacro := op.Record(gtx.Ops)
	textGtx := gtx
	textGtx.Constraints.Min = image.Point{}
	textDims := text.New(value).
		Size(float32(tokens.TextSize)).
		Weight(font.Medium).
		Color(col).
		Layout(ctx, textGtx)
	textCall := textMacro.Stop()
	width = max(width, textDims.Size.X+gtx.Dp(unit.Dp(4)))
	size := image.Pt(width, min(gtx.Dp(tokens.SegmentHeight), gtx.Constraints.Max.Y))
	segmentGtx := gtx
	segmentGtx.Constraints = layout.Exact(size)
	if !enabled {
		segmentGtx = segmentGtx.Disabled()
	}
	return segment.clickable.Layout(segmentGtx, func(gtx layout.Context) layout.Dimensions {
		clipped := clip.UniformRRect(image.Rectangle{Max: size}, min(gtx.Dp(tokens.SegmentRadius), min(size.X, size.Y)/2)).Push(gtx.Ops)
		if enabled {
			pointer.CursorText.Add(gtx.Ops)
		}
		if focused {
			background := frame.ActiveTheme(ctx).Palette.AccentSoft
			if invalid {
				background = frame.ActiveTheme(ctx).Palette.DangerSoft
			}
			paint.Fill(gtx.Ops, background)
		}
		semantic.Editor.Add(gtx.Ops)
		semantic.LabelOp(dateSegmentLabel(ctx, index)).Add(gtx.Ops)
		semantic.EnabledOp(enabled).Add(gtx.Ops)
		key.InputHintOp{Tag: &segment.clickable, Hint: key.HintNumeric}.Add(gtx.Ops)
		offset := op.Offset(image.Pt((size.X-textDims.Size.X)/2, max((size.Y-textDims.Size.Y)/2, 0))).Push(gtx.Ops)
		textCall.Add(gtx.Ops)
		offset.Pop()
		clipped.Pop()
		return layout.Dimensions{Size: size}
	})
}

func (s *dateSegmentsState) segmentText(index int, focused bool) (string, bool) {
	segment := &s.segments[index]
	if focused && segment.typed != "" {
		return segment.typed, false
	}
	value := s.part(index)
	if value == 0 {
		switch DatePart(index) {
		case dateSegmentMonth:
			return "MM", true
		case dateSegmentDay:
			return "DD", true
		default:
			return "YYYY", true
		}
	}
	if index == int(dateSegmentYear) {
		return leftPadNumber(value, 4), false
	}
	return leftPadNumber(value, 2), false
}

func layoutDateSegmentLiteral(ctx *frame.Context, gtx layout.Context, value string, style field.Style) layout.Dimensions {
	tokens := frame.ActiveTheme(ctx).Components.DatePicker
	height := min(gtx.Dp(tokens.SegmentHeight), gtx.Constraints.Max.Y)
	minimumWidth := gtx.Dp(tokens.SeparatorWidth)
	gtx.Constraints.Min = image.Point{}
	gtx.Constraints.Max.Y = height
	macro := op.Record(gtx.Ops)
	dimensions := text.New(value).Size(float32(tokens.TextSize)).Color(style.Placeholder).Layout(ctx, gtx)
	call := macro.Stop()
	size := image.Pt(max(dimensions.Size.X, minimumWidth), height)
	offset := op.Offset(image.Pt((size.X-dimensions.Size.X)/2, max((size.Y-dimensions.Size.Y)/2, 0))).Push(gtx.Ops)
	call.Add(gtx.Ops)
	offset.Pop()
	return layout.Dimensions{Size: size}
}

func leftPadNumber(value, width int) string {
	text := strconv.Itoa(value)
	for len(text) < width {
		text = "0" + text
	}
	return text
}

func dateSegmentLabel(ctx *frame.Context, index int) string {
	if frame.ActiveLanguage(ctx) == locale.LanguageChinese {
		return [...]string{"年", "月", "日"}[index]
	}
	return [...]string{"Year", "Month", "Day"}[index]
}

func (s *dateInputHoverState) update(gtx layout.Context) {
	for {
		value, ok := gtx.Event(pointer.Filter{Target: s, Kinds: pointer.Enter | pointer.Leave | pointer.Cancel})
		if !ok {
			return
		}
		eventValue, ok := value.(pointer.Event)
		if !ok {
			continue
		}
		s.hovered = eventValue.Kind == pointer.Enter
	}
}

func addDateInputHover(gtx layout.Context, target event.Tag, size image.Point, enabled, textCursor bool) {
	if !enabled || size.X <= 0 || size.Y <= 0 {
		return
	}
	clipped := clip.Rect{Max: size}.Push(gtx.Ops)
	pass := pointer.PassOp{}.Push(gtx.Ops)
	if textCursor {
		pointer.CursorText.Add(gtx.Ops)
	}
	event.Op(gtx.Ops, target)
	pass.Pop()
	clipped.Pop()
}
