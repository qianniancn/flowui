package theme

import (
	"image"
	"image/color"
	"time"

	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/style"
)

// DatePickerRequest describes the field face needed by date picker behavior.
// Calendar geometry and date selection remain owned by widget/datepicker.
type DatePickerRequest struct {
	Variant InputVariant
	// Override is the optional instance declaration supplied by the caller.
	// The provider owns the returned instance layer.
	Override *DatePickerOverride
}

// DatePickerOverride carries caller customization across the neutral provider
// boundary without storing visual state on the behavior widget.
type DatePickerOverride struct {
	Declaration style.Style
}

// DatePickerOverrideStyle extracts the provider-owned instance layer.
func DatePickerOverrideStyle(request DatePickerRequest) style.Style {
	if request.Override == nil {
		return style.Style{}
	}
	return request.Override.Declaration
}

// DatePickerMetrics contains the geometry and typography measurements needed
// to lay out date-picker parts. It deliberately contains no colors, brushes,
// or behavior flags: the widget may use these values to measure and hit-test,
// while the provider remains the owner of the visual scale.
type DatePickerMetrics struct {
	Height             unit.Dp
	Radius             unit.Dp
	RangeRadius        unit.Dp
	TextSize           unit.Sp
	FieldGap           unit.Dp
	FieldPaddingX      unit.Dp
	FocusRingWidth     unit.Dp
	SegmentHeight      unit.Dp
	SegmentRadius      unit.Dp
	YearSegmentWidth   unit.Dp
	SegmentWidth       unit.Dp
	SeparatorWidth     unit.Dp
	RangeSeparatorSize unit.Dp
	TriggerWidth       unit.Dp
	IconSize           unit.Dp
	IconRadius         unit.Dp
	IconStroke         unit.Dp
	PopoverGap         unit.Dp
	PopoverPadding     unit.Dp
	PopoverRadius      unit.Dp
	PopoverMaxHeight   unit.Dp
	CalendarWidth      unit.Dp
	CellSize           unit.Dp
	MonthCellHeight    unit.Dp
	YearCellHeight     unit.Dp
	YearGridGap        unit.Dp
	NavButtonSize      unit.Dp
	NavChevronStroke   unit.Dp
	HeaderGap          unit.Dp
	HeaderIconSize     unit.Dp
	HeaderTextSize     unit.Sp
	WeekdayTextSize    unit.Sp
	CellTextSize       unit.Sp
	CellStrikeWidth    unit.Dp
	CellStrikeHalfSize unit.Dp
}

// DatePickerCellRequest describes calendar-domain flags in addition to the
// interaction state shared by all style providers.
type DatePickerCellRequest struct {
	State   style.State
	Today   bool
	Outside bool
	Active  bool
	Range   bool
}

// DatePickerRangeRequest describes the range track's calendar-domain state.
type DatePickerRangeRequest struct {
	Outside bool
}

// DatePickerTextRequest supplies the field-derived base color for a date text
// part while keeping focus and validation state in the neutral protocol.
type DatePickerTextRequest struct {
	Base  color.NRGBA
	State style.State
}

// DatePickerFieldColors are the concrete colors needed by the field bridge
// after a provider resolves content, placeholder, and selection parts.
type DatePickerFieldColors struct {
	Foreground  color.NRGBA
	Placeholder color.NRGBA
	Selection   color.NRGBA
}

// DatePickerTextColorResolver interprets a provider-owned text declaration.
// Widgets consume the result but never inspect ColorSource themselves.
type DatePickerTextColorResolver func(style.ResolvedStyle) color.NRGBA

// DatePickerPaintColorResolver interprets a provider-owned background
// declaration for date-picker drawing.
type DatePickerPaintColorResolver func(style.ResolvedStyle) color.NRGBA

// DatePickerOutlineColorResolver interprets a provider-owned focus outline
// declaration for date-picker drawing.
type DatePickerOutlineColorResolver func(style.ResolvedStyle) color.NRGBA

// DatePickerFieldColorsResolver resolves the field bridge colors from the
// provider-owned part declarations without exposing style representation to
// widget behavior.
type DatePickerFieldColorsResolver func(
	content style.ResolvedStyle,
	placeholder style.ResolvedStyle,
	selection style.ResolvedStyle,
) DatePickerFieldColors

// DatePickerIconRenderRequest contains the neutral slot facts for the field
// calendar affordance. The provider owns the glyph and its geometry.
type DatePickerIconRenderRequest struct {
	Size    image.Point
	Color   color.NRGBA
	Metrics DatePickerMetrics
}

type DatePickerIconRenderer func(layout.Context, DatePickerIconRenderRequest)

// DatePickerYearIndicatorRenderRequest contains the neutral state for the
// calendar header's year/month disclosure mark.
type DatePickerYearIndicatorRenderRequest struct {
	Size    image.Point
	Open    bool
	Color   color.NRGBA
	Metrics DatePickerMetrics
}

type DatePickerYearIndicatorRenderer func(layout.Context, DatePickerYearIndicatorRenderRequest)

// DatePickerNavButtonRenderRequest contains the resolved button chrome and
// direction. Navigation state and enabled/focus decisions remain in widget.
type DatePickerNavButtonRenderRequest struct {
	Size       image.Point
	Direction  int
	Background color.NRGBA
	Foreground color.NRGBA
	Radius     int
	Metrics    DatePickerMetrics
}

type DatePickerNavButtonRenderer func(layout.Context, DatePickerNavButtonRenderRequest)

// DatePickerCellRenderRequest contains the resolved calendar-cell surface.
type DatePickerCellRenderRequest struct {
	Size       image.Point
	Background color.NRGBA
	Radius     int
	Metrics    DatePickerMetrics
}

type DatePickerCellRenderer func(layout.Context, DatePickerCellRenderRequest)

// DatePickerRangeRenderRequest contains the range track geometry computed by
// calendar behavior. The provider owns the track shape and insets.
type DatePickerRangeRenderRequest struct {
	Size        image.Point
	Color       color.NRGBA
	StartRadius int
	EndRadius   int
	Metrics     DatePickerMetrics
}

type DatePickerRangeRenderer func(layout.Context, DatePickerRangeRenderRequest)

// DatePickerStrikeRenderRequest contains the disabled-day strike facts.
type DatePickerStrikeRenderRequest struct {
	Size     image.Point
	Color    color.NRGBA
	HalfSize unit.Dp
	Width    unit.Dp
	Metrics  DatePickerMetrics
}

type DatePickerStrikeRenderer func(layout.Context, DatePickerStrikeRenderRequest)

// DatePickerFocusRenderRequest contains focus visibility and local bounds.
// The provider owns ring shape, inset, and stroke treatment.
type DatePickerFocusRenderRequest struct {
	Size          image.Point
	Color         color.NRGBA
	Radius        int
	Visible       bool
	Trigger       bool
	FocusWidth    unit.Dp
	IconSize      unit.Dp
	SegmentRadius unit.Dp
	Metrics       DatePickerMetrics
}

type DatePickerFocusRenderer func(layout.Context, DatePickerFocusRenderRequest)

// DatePickerMotion contains the visual transition policy for date-picker
// parts. The widget owns when a transition starts and what state it targets;
// the active provider owns the timing, easing policy, and visual scale.
//
// Durations are raw theme values. Consumers must apply Policy through
// ResolveMotionDuration so a provider can be disabled or globally scaled
// without duplicating that policy in widget code.
type DatePickerMotion struct {
	Policy       MotionTheme
	PopoverIn    time.Duration
	PopoverOut   time.Duration
	CellColor    time.Duration
	PressIn      time.Duration
	PressOut     time.Duration
	PressScale   float32
	PopoverScale float32
}

// DatePicker part names are stable semantic customization points shared by
// widgets and visual providers. They describe existing elements, not behavior.
const (
	DatePickerPartCalendarPanel      style.Part = style.PartPanel
	DatePickerPartCalendarNav        style.Part = "datepicker-nav"
	DatePickerPartCalendarHeader     style.Part = "datepicker-header"
	DatePickerPartCalendarWeekday    style.Part = "datepicker-weekday"
	DatePickerPartCalendarCell       style.Part = "datepicker-cell"
	DatePickerPartCalendarRange      style.Part = "datepicker-range"
	DatePickerPartCalendarFocus      style.Part = "datepicker-focus"
	DatePickerPartDateSegment        style.Part = "datepicker-date-segment"
	DatePickerPartDateLiteral        style.Part = "datepicker-date-literal"
	DatePickerPartDateRangeSeparator style.Part = "datepicker-range-separator"
	DatePickerPartDateTrigger        style.Part = "datepicker-date-trigger"
)

// DatePickerVisual contains the declaration layers consumed by date picker
// behavior. Calendar and dynamic part recipes are owned by the active theme;
// widget code only resolves and draws the returned declarations.
type DatePickerVisual struct {
	Defaults       style.Style
	Override       style.Style
	Calendar       style.Style
	Metrics        DatePickerMetrics
	Motion         DatePickerMotion
	Cell           func(DatePickerCellRequest) style.Style
	Range          func(DatePickerRangeRequest) style.Style
	DateSegment    func(DatePickerTextRequest) style.Style
	DateLiteral    func(DatePickerTextRequest) style.Style
	RangeSeparator func(DatePickerTextRequest) style.Style
	Trigger        func(DatePickerTextRequest) style.Style
	TextColor      DatePickerTextColorResolver
	PaintColor     DatePickerPaintColorResolver
	OutlineColor   DatePickerOutlineColorResolver
	FieldColors    DatePickerFieldColorsResolver
	CalendarIcon   DatePickerIconRenderer
	YearIndicator  DatePickerYearIndicatorRenderer
	NavButton      DatePickerNavButtonRenderer
	CellRenderer   DatePickerCellRenderer
	RangeRenderer  DatePickerRangeRenderer
	Strike         DatePickerStrikeRenderer
	Focus          DatePickerFocusRenderer
}

// DatePickerProvider supplies the complete date picker visual bundle,
// including the neutral geometry consumed by widget layout.
type DatePickerProvider interface {
	DatePicker(active *Theme, request DatePickerRequest) DatePickerVisual
}

// ValidateDatePickerVisual checks the complete field and calendar declaration
// before a widget stores a provider result in its cache. Dynamic calendar and
// text parts are sampled across the states used by date-picker behavior so a
// provider cannot silently fall back to a second visual implementation.
func ValidateDatePickerVisual(visual DatePickerVisual) {
	validateMotion("DatePicker", "Motion.Policy", visual.Motion.Policy)
	if visual.Motion.PopoverIn <= 0 {
		invalidVisual("DatePicker", "Motion.PopoverIn", "must be greater than 0")
	}
	if visual.Motion.PopoverOut <= 0 {
		invalidVisual("DatePicker", "Motion.PopoverOut", "must be greater than 0")
	}
	if visual.Motion.CellColor <= 0 {
		invalidVisual("DatePicker", "Motion.CellColor", "must be greater than 0")
	}
	if visual.Motion.PressIn <= 0 {
		invalidVisual("DatePicker", "Motion.PressIn", "must be greater than 0")
	}
	if visual.Motion.PressOut <= 0 {
		invalidVisual("DatePicker", "Motion.PressOut", "must be greater than 0")
	}
	if !validFiniteFloat32(visual.Motion.PressScale) {
		invalidVisual("DatePicker", "Motion.PressScale", "must be finite")
	}
	if visual.Motion.PressScale <= 0 {
		invalidVisual("DatePicker", "Motion.PressScale", "must be greater than 0")
	}
	if visual.Motion.PressScale > 1 {
		invalidVisual("DatePicker", "Motion.PressScale", "must be at most 1")
	}
	if !validFiniteFloat32(visual.Motion.PopoverScale) {
		invalidVisual("DatePicker", "Motion.PopoverScale", "must be finite")
	}
	if visual.Motion.PopoverScale <= 0 {
		invalidVisual("DatePicker", "Motion.PopoverScale", "must be greater than 0")
	}
	if visual.Motion.PopoverScale > 1 {
		invalidVisual("DatePicker", "Motion.PopoverScale", "must be at most 1")
	}

	metrics := visual.Metrics
	if !validPositiveDp(metrics.Height) {
		invalidVisual("DatePicker", "Metrics.Height", "must be finite and greater than 0")
	}
	if !validNonNegativeDp(metrics.Radius) {
		invalidVisual("DatePicker", "Metrics.Radius", "must be finite and non-negative")
	}
	if !validNonNegativeDp(metrics.RangeRadius) {
		invalidVisual("DatePicker", "Metrics.RangeRadius", "must be finite and non-negative")
	}
	if !validPositiveSp(metrics.TextSize) {
		invalidVisual("DatePicker", "Metrics.TextSize", "must be finite and greater than 0")
	}
	if !validNonNegativeDp(metrics.FieldGap) {
		invalidVisual("DatePicker", "Metrics.FieldGap", "must be finite and non-negative")
	}
	if !validNonNegativeDp(metrics.FieldPaddingX) {
		invalidVisual("DatePicker", "Metrics.FieldPaddingX", "must be finite and non-negative")
	}
	if !validNonNegativeDp(metrics.FocusRingWidth) {
		invalidVisual("DatePicker", "Metrics.FocusRingWidth", "must be finite and non-negative")
	}
	if !validPositiveDp(metrics.SegmentHeight) {
		invalidVisual("DatePicker", "Metrics.SegmentHeight", "must be finite and greater than 0")
	}
	if !validNonNegativeDp(metrics.SegmentRadius) {
		invalidVisual("DatePicker", "Metrics.SegmentRadius", "must be finite and non-negative")
	}
	if !validPositiveDp(metrics.YearSegmentWidth) {
		invalidVisual("DatePicker", "Metrics.YearSegmentWidth", "must be finite and greater than 0")
	}
	if !validPositiveDp(metrics.SegmentWidth) {
		invalidVisual("DatePicker", "Metrics.SegmentWidth", "must be finite and greater than 0")
	}
	if !validNonNegativeDp(metrics.SeparatorWidth) {
		invalidVisual("DatePicker", "Metrics.SeparatorWidth", "must be finite and non-negative")
	}
	if !validPositiveDp(metrics.RangeSeparatorSize) {
		invalidVisual("DatePicker", "Metrics.RangeSeparatorSize", "must be finite and greater than 0")
	}
	if !validPositiveDp(metrics.TriggerWidth) {
		invalidVisual("DatePicker", "Metrics.TriggerWidth", "must be finite and greater than 0")
	}
	if !validPositiveDp(metrics.IconSize) {
		invalidVisual("DatePicker", "Metrics.IconSize", "must be finite and greater than 0")
	}
	if !validNonNegativeDp(metrics.IconRadius) {
		invalidVisual("DatePicker", "Metrics.IconRadius", "must be finite and non-negative")
	}
	if !validPositiveDp(metrics.IconStroke) {
		invalidVisual("DatePicker", "Metrics.IconStroke", "must be finite and greater than 0")
	}
	if !validNonNegativeDp(metrics.PopoverGap) {
		invalidVisual("DatePicker", "Metrics.PopoverGap", "must be finite and non-negative")
	}
	if !validNonNegativeDp(metrics.PopoverPadding) {
		invalidVisual("DatePicker", "Metrics.PopoverPadding", "must be finite and non-negative")
	}
	if !validNonNegativeDp(metrics.PopoverRadius) {
		invalidVisual("DatePicker", "Metrics.PopoverRadius", "must be finite and non-negative")
	}
	if !validPositiveDp(metrics.PopoverMaxHeight) {
		invalidVisual("DatePicker", "Metrics.PopoverMaxHeight", "must be finite and greater than 0")
	}
	if !validPositiveDp(metrics.CalendarWidth) {
		invalidVisual("DatePicker", "Metrics.CalendarWidth", "must be finite and greater than 0")
	}
	if !validPositiveDp(metrics.CellSize) {
		invalidVisual("DatePicker", "Metrics.CellSize", "must be finite and greater than 0")
	}
	if !validPositiveDp(metrics.MonthCellHeight) {
		invalidVisual("DatePicker", "Metrics.MonthCellHeight", "must be finite and greater than 0")
	}
	if !validPositiveDp(metrics.YearCellHeight) {
		invalidVisual("DatePicker", "Metrics.YearCellHeight", "must be finite and greater than 0")
	}
	if !validNonNegativeDp(metrics.YearGridGap) {
		invalidVisual("DatePicker", "Metrics.YearGridGap", "must be finite and non-negative")
	}
	if !validPositiveDp(metrics.NavButtonSize) {
		invalidVisual("DatePicker", "Metrics.NavButtonSize", "must be finite and greater than 0")
	}
	if !validPositiveDp(metrics.NavChevronStroke) {
		invalidVisual("DatePicker", "Metrics.NavChevronStroke", "must be finite and greater than 0")
	}
	if !validNonNegativeDp(metrics.HeaderGap) {
		invalidVisual("DatePicker", "Metrics.HeaderGap", "must be finite and non-negative")
	}
	if !validPositiveDp(metrics.HeaderIconSize) {
		invalidVisual("DatePicker", "Metrics.HeaderIconSize", "must be finite and greater than 0")
	}
	if !validPositiveSp(metrics.HeaderTextSize) {
		invalidVisual("DatePicker", "Metrics.HeaderTextSize", "must be finite and greater than 0")
	}
	if !validPositiveSp(metrics.WeekdayTextSize) {
		invalidVisual("DatePicker", "Metrics.WeekdayTextSize", "must be finite and greater than 0")
	}
	if !validPositiveSp(metrics.CellTextSize) {
		invalidVisual("DatePicker", "Metrics.CellTextSize", "must be finite and greater than 0")
	}
	if !validPositiveDp(metrics.CellStrikeWidth) {
		invalidVisual("DatePicker", "Metrics.CellStrikeWidth", "must be finite and greater than 0")
	}
	if !validPositiveDp(metrics.CellStrikeHalfSize) {
		invalidVisual("DatePicker", "Metrics.CellStrikeHalfSize", "must be finite and greater than 0")
	}

	if visual.Cell == nil {
		invalidVisual("DatePicker", "Cell", "must be provided")
	}
	if visual.Range == nil {
		invalidVisual("DatePicker", "Range", "must be provided")
	}
	if visual.DateSegment == nil {
		invalidVisual("DatePicker", "DateSegment", "must be provided")
	}
	if visual.DateLiteral == nil {
		invalidVisual("DatePicker", "DateLiteral", "must be provided")
	}
	if visual.RangeSeparator == nil {
		invalidVisual("DatePicker", "RangeSeparator", "must be provided")
	}
	if visual.Trigger == nil {
		invalidVisual("DatePicker", "Trigger", "must be provided")
	}
	if visual.TextColor == nil {
		invalidVisual("DatePicker", "TextColor", "must be provided")
	}
	if visual.PaintColor == nil {
		invalidVisual("DatePicker", "PaintColor", "must be provided")
	}
	if visual.OutlineColor == nil {
		invalidVisual("DatePicker", "OutlineColor", "must be provided")
	}
	if visual.FieldColors == nil {
		invalidVisual("DatePicker", "FieldColors", "must be provided")
	}
	if visual.CalendarIcon == nil {
		invalidVisual("DatePicker", "CalendarIcon", "must be provided")
	}
	if visual.YearIndicator == nil {
		invalidVisual("DatePicker", "YearIndicator", "must be provided")
	}
	if visual.NavButton == nil {
		invalidVisual("DatePicker", "NavButton", "must be provided")
	}
	if visual.CellRenderer == nil {
		invalidVisual("DatePicker", "CellRenderer", "must be provided")
	}
	if visual.RangeRenderer == nil {
		invalidVisual("DatePicker", "RangeRenderer", "must be provided")
	}
	if visual.Strike == nil {
		invalidVisual("DatePicker", "Strike", "must be provided")
	}
	if visual.Focus == nil {
		invalidVisual("DatePicker", "Focus", "must be provided")
	}

	panel := style.CascadePart(style.State{}, DatePickerPartCalendarPanel, visual.Calendar)
	if panel.Box == nil {
		invalidVisual("DatePicker", "Panel.Box", "must be provided")
	}
	if panel.Box.Padding == nil {
		invalidVisual("DatePicker", "Panel.Box.Padding", "must be provided")
	}

	requirePaint := func(field string, value style.ResolvedStyle) {
		if value.Paint == nil {
			invalidVisual("DatePicker", field, "must be provided")
		}
		if !hasPaint(value.Paint.Background) {
			invalidVisual("DatePicker", field, "must provide a solid color or nonempty gradient")
		}
	}
	requireText := func(field string, value style.ResolvedStyle) {
		if value.Text == nil {
			invalidVisual("DatePicker", field, "must be provided")
		}
		if !hasColor(value.Text.Color) {
			invalidVisual("DatePicker", field, "must resolve to a solid color")
		}
	}
	requireRadius := func(field string, value style.ResolvedStyle) {
		if value.Paint == nil {
			invalidVisual("DatePicker", field, "must be provided")
		}
		if value.Paint.Radius == nil {
			invalidVisual("DatePicker", field, "must be provided")
		}
	}
	requireOutline := func(field string, value style.ResolvedStyle) {
		if value.Paint == nil {
			invalidVisual("DatePicker", field, "must be provided")
		}
		if value.Paint.Outline == nil {
			invalidVisual("DatePicker", field, "must be provided")
		}
		if !hasColor(value.Paint.Outline.Color) {
			invalidVisual("DatePicker", field, "must resolve to a solid color")
		}
	}

	requirePaint("CalendarPanel", panel)
	nav := style.CascadePart(style.State{}, DatePickerPartCalendarNav, visual.Calendar)
	requireText("CalendarNav", nav)
	requirePaint("CalendarNav", nav)
	requireRadius("CalendarNav", nav)
	requireText("CalendarHeader", style.CascadePart(style.State{}, DatePickerPartCalendarHeader, visual.Calendar))
	requireText("CalendarWeekday", style.CascadePart(style.State{}, DatePickerPartCalendarWeekday, visual.Calendar))
	for _, state := range []style.State{
		{},
		{Disabled: true},
		{Invalid: true},
		{Disabled: true, Invalid: true},
	} {
		requireText("Description", style.CascadePart(state, style.PartDescription, visual.Calendar))
	}
	requireOutline("CalendarFocus", style.CascadePart(style.State{}, DatePickerPartCalendarFocus, visual.Calendar))
	requireText("DateSegment", style.CascadePart(style.State{}, DatePickerPartDateSegment, visual.Calendar))
	requireText("DateLiteral", style.CascadePart(style.State{}, DatePickerPartDateLiteral, visual.Calendar))
	requireText("DateRangeSeparator", style.CascadePart(style.State{}, DatePickerPartDateRangeSeparator, visual.Calendar))
	requireText("DateTrigger", style.CascadePart(style.State{}, DatePickerPartDateTrigger, visual.Calendar))

	for _, request := range []DatePickerCellRequest{
		{},
		{State: style.State{Hovered: true}},
		{State: style.State{Selected: true}},
		{Today: true},
		{Range: true},
		{State: style.State{Disabled: true}, Outside: true},
	} {
		declaration := visual.Cell(request)
		resolved := style.CascadePart(request.State, DatePickerPartCalendarCell,
			visual.Calendar, style.Style{}.Part(DatePickerPartCalendarCell, declaration))
		requireText("CalendarCell", resolved)
		requirePaint("CalendarCell", resolved)
		requireRadius("CalendarCell", resolved)
	}
	for _, outside := range []bool{false, true} {
		declaration := visual.Range(DatePickerRangeRequest{Outside: outside})
		resolved := style.CascadePart(style.State{}, DatePickerPartCalendarRange,
			visual.Calendar, style.Style{}.Part(DatePickerPartCalendarRange, declaration))
		requirePaint("CalendarRange", resolved)
	}

	textParts := []struct {
		name        string
		part        style.Part
		declaration func(DatePickerTextRequest) style.Style
		base        color.NRGBA
		paint       bool
	}{
		{"DateSegment", DatePickerPartDateSegment, visual.DateSegment, color.NRGBA{}, true},
		{"DateLiteral", DatePickerPartDateLiteral, visual.DateLiteral, color.NRGBA{}, false},
		{"RangeSeparator", DatePickerPartDateRangeSeparator, visual.RangeSeparator, color.NRGBA{}, false},
		{"Trigger", DatePickerPartDateTrigger, visual.Trigger, color.NRGBA{}, false},
	}
	for _, part := range textParts {
		declaration := part.declaration(DatePickerTextRequest{Base: part.base})
		resolved := style.CascadePart(style.State{}, part.part,
			visual.Calendar, style.Style{}.Part(part.part, declaration))
		requireText(part.name, resolved)
		visual.TextColor(resolved)
		if part.paint {
			requirePaint(part.name, resolved)
			visual.PaintColor(resolved)
		}
	}
	visual.PaintColor(nav)
	visual.OutlineColor(style.CascadePart(style.State{}, DatePickerPartCalendarFocus, visual.Calendar))
	content := style.CascadePart(style.State{}, style.PartContent,
		visual.Defaults, visual.Override)
	placeholder := style.CascadePart(style.State{}, style.PartPlaceholder,
		visual.Defaults, visual.Override)
	selection := style.CascadePart(style.State{}, style.PartSelection,
		visual.Defaults, visual.Override)
	visual.FieldColors(content, placeholder, selection)
}
