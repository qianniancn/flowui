// Package datepickerstyle contains the visual recipes used by date picker
// providers. It has no date-picker behavior and only consumes the neutral
// theme protocol.
package datepickerstyle

import (
	"image/color"
	"time"

	"gioui.org/font"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/theme/internal/fieldstyle"
	"github.com/qianniancn/flowui/theme/internal/panelstyle"
)

const (
	datePickerPopoverInDuration  = 150 * time.Millisecond
	datePickerPopoverOutDuration = 100 * time.Millisecond
	datePickerCellColorDuration  = 100 * time.Millisecond
	datePickerPressInDuration    = 90 * time.Millisecond
	datePickerPressOutDuration   = 160 * time.Millisecond
	datePickerPressScale         = 0.95
	datePickerPopoverScale       = 0.95
)

// BuildDatePicker returns the complete field and calendar recipe for a date
// picker. The widget owns selection, focus, navigation, and overlay behavior;
// this package owns the visual declaration for each semantic part.
func BuildDatePicker(active *theme.Theme, request theme.DatePickerRequest) theme.DatePickerVisual {
	if active == nil {
		panic(theme.ErrMissingVisualProvider)
	}
	return theme.DatePickerVisual{
		Defaults: fieldstyle.FieldDefaultRules(active, request.Variant, fieldstyle.FieldDeclarationOptions{
			TargetPart:          style.PartContent,
			Radius:              active.Components.DatePicker.Radius,
			FocusRingWidth:      active.Components.Input.FocusRingWidth,
			FocusGlowBlur:       active.Components.Input.FocusGlowBlur,
			FocusHaloBlur:       active.Components.Input.FocusHaloBlur,
			FocusHaloOpacity:    active.Components.Input.FocusHaloOpacity,
			FocusShadowSpread:   active.Components.Input.FocusShadowSpread,
			InvalidOutlineWidth: active.Components.Input.InvalidOutlineWidth,
			BorderWidth:         active.Components.Input.BorderWidth,
			ShadowColor:         active.Components.Input.ShadowColor,
			ShadowOpacity:       active.Components.Input.ShadowOpacity,
			ShadowStrength:      active.Components.Input.ShadowStrength,
		}),
		Override: theme.DatePickerOverrideStyle(request),
		Calendar: calendarDeclaration(active),
		Metrics:  datePickerMetrics(active),
		Motion: theme.DatePickerMotion{
			Policy:       active.Motion,
			PopoverIn:    datePickerPopoverInDuration,
			PopoverOut:   datePickerPopoverOutDuration,
			CellColor:    datePickerCellColorDuration,
			PressIn:      datePickerPressInDuration,
			PressOut:     datePickerPressOutDuration,
			PressScale:   datePickerPressScale,
			PopoverScale: datePickerPopoverScale,
		},
		Cell:        func(value theme.DatePickerCellRequest) style.Style { return cellDeclaration(active, value) },
		Range:       func(value theme.DatePickerRangeRequest) style.Style { return rangeDeclaration(active, value) },
		DateSegment: func(value theme.DatePickerTextRequest) style.Style { return dateSegmentDeclaration(active, value) },
		DateLiteral: func(value theme.DatePickerTextRequest) style.Style {
			return textPartDeclaration(theme.DatePickerPartDateLiteral, value.Base)
		},
		RangeSeparator: func(value theme.DatePickerTextRequest) style.Style {
			return textPartDeclaration(theme.DatePickerPartDateRangeSeparator, value.Base)
		},
		Trigger: func(value theme.DatePickerTextRequest) style.Style {
			return textPartDeclaration(theme.DatePickerPartDateTrigger, value.Base)
		},
		TextColor:     resolveTextColor,
		PaintColor:    resolvePaintColor,
		OutlineColor:  resolveOutlineColor,
		FieldColors:   resolveFieldColors,
		CalendarIcon:  calendarIconRenderer,
		YearIndicator: yearIndicatorRenderer,
		NavButton:     navButtonRenderer,
		CellRenderer:  cellRenderer,
		RangeRenderer: rangeRenderer,
		Strike:        strikeRenderer,
		Focus:         focusRenderer,
	}
}

func resolveTextColor(value style.ResolvedStyle) color.NRGBA {
	if value.Text == nil {
		panic(theme.ErrIncompleteVisualProvider)
	}
	resolved, ok := computed.Color(value.Text.Color)
	if !ok {
		panic(theme.ErrIncompleteVisualProvider)
	}
	return resolved
}

func resolvePaintColor(value style.ResolvedStyle) color.NRGBA {
	if value.Paint == nil {
		panic(theme.ErrIncompleteVisualProvider)
	}
	resolved, ok := computed.Brush(value.Paint.Background)
	if !ok {
		panic(theme.ErrIncompleteVisualProvider)
	}
	return resolved.ColorAt(.5)
}

// resolveOutlineColor returns the accent a focus ring draws with. A field family
// may express its focus ring as an outer outline or as an inset border, so the
// accent is whichever stroke the resolved part actually carries; the calendar's
// own focus part always provides the outline.
func resolveOutlineColor(value style.ResolvedStyle) color.NRGBA {
	if value.Paint == nil {
		panic(theme.ErrIncompleteVisualProvider)
	}
	if value.Paint.Outline != nil {
		if resolved, ok := computed.Color(value.Paint.Outline.Color); ok {
			return resolved
		}
	}
	if value.Paint.Border != nil {
		if resolved, ok := computed.Color(value.Paint.Border.Color); ok {
			return resolved
		}
	}
	panic(theme.ErrIncompleteVisualProvider)
}

func resolveFieldColors(content, placeholder, selection style.ResolvedStyle) theme.DatePickerFieldColors {
	return theme.DatePickerFieldColors{
		Foreground:  resolveTextColor(content),
		Placeholder: resolveTextColor(placeholder),
		Selection:   resolvePaintColor(selection),
	}
}

// datePickerMetrics returns the neutral measurements used by date-picker
// layout. It is embedded in the visual bundle so geometry and declarations
// share one Provider/cache lifecycle.
func datePickerMetrics(active *theme.Theme) theme.DatePickerMetrics {
	if active == nil {
		panic(theme.ErrMissingVisualProvider)
	}
	tokens := active.Components.DatePicker
	return theme.DatePickerMetrics{
		Height:             tokens.Height,
		Radius:             tokens.Radius,
		RangeRadius:        tokens.RangeRadius,
		TextSize:           tokens.TextSize,
		FieldGap:           tokens.FieldGap,
		FieldPaddingX:      active.Components.Input.PaddingX,
		FocusRingWidth:     active.Components.Input.FocusRingWidth,
		SegmentHeight:      tokens.SegmentHeight,
		SegmentRadius:      tokens.SegmentRadius,
		YearSegmentWidth:   tokens.YearSegmentWidth,
		SegmentWidth:       tokens.SegmentWidth,
		SeparatorWidth:     tokens.SeparatorWidth,
		RangeSeparatorSize: tokens.RangeSeparatorSize,
		TriggerWidth:       tokens.TriggerWidth,
		IconSize:           tokens.IconSize,
		IconRadius:         tokens.IconRadius,
		IconStroke:         tokens.IconStroke,
		PopoverGap:         tokens.PopoverGap,
		PopoverPadding:     tokens.PopoverPadding,
		PopoverRadius:      tokens.PopoverRadius,
		PopoverMaxHeight:   tokens.PopoverMaxHeight,
		CalendarWidth:      tokens.CalendarWidth,
		CellSize:           tokens.CellSize,
		MonthCellHeight:    tokens.MonthCellHeight,
		YearCellHeight:     tokens.YearCellHeight,
		YearGridGap:        tokens.YearGridGap,
		NavButtonSize:      tokens.NavButtonSize,
		NavChevronStroke:   tokens.NavChevronStroke,
		HeaderGap:          tokens.HeaderGap,
		HeaderIconSize:     tokens.HeaderIconSize,
		HeaderTextSize:     tokens.HeaderTextSize,
		WeekdayTextSize:    tokens.WeekdayTextSize,
		CellTextSize:       tokens.CellTextSize,
		CellStrikeWidth:    tokens.CellStrikeWidth,
		CellStrikeHalfSize: tokens.CellStrikeHalfSize,
	}
}

// CalendarDeclaration contains all static calendar and date-part defaults.
func calendarDeclaration(active *theme.Theme) style.Style {
	if active == nil {
		panic(theme.ErrMissingVisualProvider)
	}
	sem := active.Semantic()
	tokens := active.Components.DatePicker
	danger := sem.Tone(theme.ToneDanger).Fill
	return style.Style{}.
		Part(theme.DatePickerPartCalendarPanel, panelstyle.Popup(
			active,
			tokens.PopoverRadius,
			tokens.PopoverPadding,
			tokens.PopoverBorderWidth,
		)).
		Part(style.PartDescription, style.Style{}.
			FontSize(active.Components.Description.TextSize).
			TextColor(style.Solid(sem.MutedForeground)).
			On(style.Invalid, style.Style{}.TextColor(style.Solid(danger))).
			On(style.Disabled, style.Style{}.
				TextColor(style.Solid(active.DisabledColor(sem.MutedForeground)))).
			On(style.Invalid|style.Disabled, style.Style{}.
				TextColor(style.Solid(active.DisabledColor(danger))))).
		Part(theme.DatePickerPartCalendarHeader, style.Style{}.
			FontSize(tokens.HeaderTextSize).
			FontWeight(int(font.Medium)).
			TextColor(style.Solid(sem.Foreground)).
			On(style.Selected, style.Style{}.
				TextColor(style.Solid(sem.Tone(theme.TonePrimary).OnSoft)))).
		Part(theme.DatePickerPartCalendarWeekday, style.Style{}.
			FontSize(tokens.WeekdayTextSize).
			FontWeight(int(font.Medium)).
			TextColor(style.Solid(sem.MutedForeground))).
		Part(theme.DatePickerPartCalendarNav, navStyle(active)).
		Part(theme.DatePickerPartCalendarCell, cellRecipe(active)).
		Part(theme.DatePickerPartCalendarRange, style.Style{}.
			Background(style.Solid(sem.Tone(theme.TonePrimary).Soft))).
		Part(theme.DatePickerPartCalendarFocus, style.Style{}.
			TextColor(style.Solid(sem.Ring)).
			Outline(active.Components.Input.FocusRingWidth, 0, style.Solid(sem.Ring))).
		Part(theme.DatePickerPartDateSegment, style.Style{}.
			FontSize(tokens.TextSize).
			FontWeight(int(font.Medium)).
			Background(style.RGBA(0)).
			TextColor(style.Solid(sem.FieldForeground))).
		Part(theme.DatePickerPartDateLiteral, style.Style{}.
			FontSize(tokens.TextSize).
			TextColor(style.Solid(sem.FieldPlaceholder))).
		Part(theme.DatePickerPartDateRangeSeparator, style.Style{}.
			FontSize(tokens.TextSize).
			TextColor(style.Solid(sem.FieldPlaceholder))).
		Part(theme.DatePickerPartDateTrigger, style.Style{}.
			TextColor(style.Solid(sem.FieldPlaceholder)))
}

// CellDeclaration applies calendar-domain flags to a cell declaration. The
// shared interaction state remains in request.State so normal style overrides
// and transitions continue to work.
func cellDeclaration(active *theme.Theme, request theme.DatePickerCellRequest) style.Style {
	if active == nil {
		panic(theme.ErrMissingVisualProvider)
	}
	sem := active.Semantic()
	tokens := active.Components.DatePicker
	background := color.NRGBA{}
	foreground := theme.ColorOr(tokens.CurrentTextColor, sem.Foreground)
	if request.State.Hovered || request.State.Pressed {
		background = sem.SurfaceRaised
	}
	if request.Today {
		background = sem.Tone(theme.TonePrimary).Soft
		foreground = sem.Tone(theme.TonePrimary).OnSoft
	}
	if request.State.Selected {
		background = sem.Tone(theme.TonePrimary).Fill
		foreground = sem.Tone(theme.TonePrimary).OnFill
	}
	if request.Outside {
		foreground = theme.ColorOr(tokens.OutsideTextColor, active.DisabledColor(sem.MutedForeground))
		if request.State.Selected {
			background = sem.SurfaceRaised
		}
	}
	if request.Active && !request.State.Selected && !request.State.Disabled {
		background = sem.SurfaceRaised
		if !request.Today {
			foreground = sem.Tone(theme.TonePrimary).OnSoft
		}
	}
	if request.Range && !request.State.Selected {
		background = color.NRGBA{}
		if !request.Outside {
			foreground = sem.Tone(theme.TonePrimary).OnSoft
		}
	}
	if request.State.Disabled {
		background = active.DisabledColor(background)
		foreground = active.DisabledColor(foreground)
	}
	return style.Style{}.
		Background(style.Solid(background)).
		TextColor(style.Solid(foreground)).
		FontWeight(int(font.Medium))
}

// RangeDeclaration returns the range-track declaration for a calendar cell.
func rangeDeclaration(active *theme.Theme, request theme.DatePickerRangeRequest) style.Style {
	if active == nil {
		panic(theme.ErrMissingVisualProvider)
	}
	value := active.Semantic().Tone(theme.TonePrimary).Soft
	if request.Outside {
		value = active.Semantic().SurfaceRaised
		value.A = byte(uint16(value.A) / 5)
	}
	return style.Style{}.Background(style.Solid(value))
}

// DateSegmentDeclaration returns the field-derived date segment declaration.
func dateSegmentDeclaration(active *theme.Theme, request theme.DatePickerTextRequest) style.Style {
	if active == nil {
		panic(theme.ErrMissingVisualProvider)
	}
	danger := active.Semantic().Tone(theme.ToneDanger)
	return style.Style{}.
		TextColor(style.Solid(request.Base)).
		On(style.Focused, style.Style{}.
			TextColor(style.Solid(active.Semantic().Tone(theme.TonePrimary).OnSoft)).
			Background(style.Solid(active.Semantic().Tone(theme.TonePrimary).Soft))).
		On(style.Invalid, style.Style{}.
			TextColor(style.Solid(danger.Fill))).
		On(style.Invalid|style.Focused, style.Style{}.
			TextColor(style.Solid(danger.OnSoft)).
			Background(style.Solid(danger.Soft)))
}

// TextPartDeclaration returns a simple text-part override for literal and
// separator content. The part is applied by the widget resolver.
func textPartDeclaration(_ style.Part, base color.NRGBA) style.Style {
	return style.Style{}.TextColor(style.Solid(base))
}

func navStyle(active *theme.Theme) style.Style {
	sem := active.Semantic()
	rest := theme.Composite(sem.Overlay, sem.Foreground, .04)
	return style.Style{}.
		Radius(active.Components.DatePicker.NavButtonSize/2).
		TextColor(style.Solid(sem.Tone(theme.TonePrimary).OnSoft)).
		Background(style.Solid(rest)).
		On(style.Hovered, style.Style{}.Background(style.Solid(sem.SurfaceRaised))).
		On(style.Pressed, style.Style{}.Background(style.Solid(sem.SurfacePressed))).
		On(style.Disabled, style.Style{}.
			TextColor(style.Solid(active.DisabledColor(sem.Tone(theme.TonePrimary).OnSoft))).
			Background(style.Solid(active.DisabledColor(rest))))
}

func cellRecipe(active *theme.Theme) style.Style {
	sem := active.Semantic()
	return style.Style{}.
		FontSize(active.Components.DatePicker.CellTextSize).
		FontWeight(int(font.Medium)).
		Radius(active.Components.DatePicker.CellSize/2).
		Background(style.RGBA(0)).
		TextColor(style.Solid(sem.Foreground)).
		On(style.Hovered, style.Style{}.Background(style.Solid(sem.SurfaceRaised))).
		On(style.Pressed, style.Style{}.Background(style.Solid(sem.SurfacePressed))).
		On(style.Disabled, style.Style{}.
			TextColor(style.Solid(active.DisabledColor(sem.Foreground))).
			Background(style.Solid(active.DisabledColor(sem.SurfaceRaised))))
}
