package testtheme

import (
	"image/color"
	"time"

	"gioui.org/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

func (behaviorProvider) DatePicker(active *theme.Theme, request theme.DatePickerRequest) theme.DatePickerVisual {
	tokens := active.Components.DatePicker
	metrics := theme.DatePickerMetrics{
		Height: tokens.Height, Radius: tokens.Radius, RangeRadius: tokens.RangeRadius,
		TextSize: tokens.TextSize, FieldGap: tokens.FieldGap, FieldPaddingX: active.Components.Input.PaddingX,
		FocusRingWidth: active.Components.Input.FocusRingWidth, SegmentHeight: tokens.SegmentHeight,
		SegmentRadius: tokens.SegmentRadius, YearSegmentWidth: tokens.YearSegmentWidth,
		SegmentWidth: tokens.SegmentWidth, SeparatorWidth: tokens.SeparatorWidth,
		RangeSeparatorSize: tokens.RangeSeparatorSize, TriggerWidth: tokens.TriggerWidth,
		IconSize: tokens.IconSize, IconRadius: tokens.IconRadius, IconStroke: tokens.IconStroke,
		PopoverGap: tokens.PopoverGap, PopoverPadding: tokens.PopoverPadding,
		PopoverRadius: tokens.PopoverRadius, PopoverMaxHeight: tokens.PopoverMaxHeight,
		CalendarWidth: tokens.CalendarWidth, CellSize: tokens.CellSize, MonthCellHeight: tokens.MonthCellHeight,
		YearCellHeight: tokens.YearCellHeight, YearGridGap: tokens.YearGridGap,
		NavButtonSize: tokens.NavButtonSize, NavChevronStroke: tokens.NavChevronStroke,
		HeaderGap: tokens.HeaderGap, HeaderIconSize: tokens.HeaderIconSize,
		HeaderTextSize: tokens.HeaderTextSize, WeekdayTextSize: tokens.WeekdayTextSize,
		CellTextSize: tokens.CellTextSize, CellStrikeWidth: tokens.CellStrikeWidth,
		CellStrikeHalfSize: tokens.CellStrikeHalfSize,
	}
	calendar := style.Style{}.
		Part(theme.DatePickerPartCalendarPanel, style.Style{}.
			Background(style.Solid(background)).TextColor(style.Solid(foreground)).
			Padding(tokens.PopoverPadding).Radius(tokens.PopoverRadius)).
		Part(style.PartDescription, style.Style{}.FontSize(12).TextColor(style.Solid(muted))).
		Part(theme.DatePickerPartCalendarHeader, style.Style{}.FontSize(tokens.HeaderTextSize).TextColor(style.Solid(foreground))).
		Part(theme.DatePickerPartCalendarWeekday, style.Style{}.FontSize(tokens.WeekdayTextSize).TextColor(style.Solid(muted))).
		Part(theme.DatePickerPartCalendarNav, style.Style{}.FontSize(tokens.CellTextSize).
			TextColor(style.Solid(foreground)).Background(style.Solid(background)).Radius(tokens.NavButtonSize/2)).
		Part(theme.DatePickerPartCalendarCell, style.Style{}.FontSize(tokens.CellTextSize).
			TextColor(style.Solid(foreground)).Background(style.Solid(background)).Radius(tokens.CellSize/2)).
		Part(theme.DatePickerPartCalendarRange, style.Style{}.Background(style.Solid(accent))).
		Part(theme.DatePickerPartCalendarFocus, style.Style{}.TextColor(style.Solid(focus)).Outline(2, 0, style.Solid(focus))).
		Part(theme.DatePickerPartDateSegment, style.Style{}.FontSize(tokens.TextSize).TextColor(style.Solid(foreground))).
		Part(theme.DatePickerPartDateLiteral, style.Style{}.FontSize(tokens.TextSize).TextColor(style.Solid(muted))).
		Part(theme.DatePickerPartDateRangeSeparator, style.Style{}.FontSize(tokens.TextSize).TextColor(style.Solid(muted))).
		Part(theme.DatePickerPartDateTrigger, style.Style{}.FontSize(tokens.TextSize).TextColor(style.Solid(muted)))
	defaults := style.Style{}.
		Part(style.PartContent, style.Style{}.
			Height(tokens.Height).
			Radius(tokens.Radius).
			FontSize(tokens.TextSize).
			TextColor(style.Solid(foreground)).
			Background(style.Solid(background)).
			Outline(0, 0, style.Solid(focus)).
			On(style.Focused, style.Style{}.Outline(2, 0, style.Solid(focus)))).
		Part(style.PartPlaceholder, style.Style{}.FontSize(tokens.TextSize).TextColor(style.Solid(muted))).
		Part(style.PartSelection, style.Style{}.Background(style.Solid(focus)))
	return theme.DatePickerVisual{
		Defaults: defaults, Override: theme.DatePickerOverrideStyle(request), Calendar: calendar,
		Metrics: metrics,
		Motion:  theme.DatePickerMotion{Policy: active.Motion, PopoverIn: 100 * time.Millisecond, PopoverOut: 100 * time.Millisecond, CellColor: 100 * time.Millisecond, PressIn: 80 * time.Millisecond, PressOut: 100 * time.Millisecond, PressScale: .98, PopoverScale: .98},
		Cell: func(request theme.DatePickerCellRequest) style.Style {
			bg, fg := background, foreground
			if request.State.Selected {
				bg, fg = focus, accentFg
			} else if request.Today || request.Active || request.State.Hovered || request.State.Pressed {
				bg, fg = accent, foreground
			}
			if request.Outside {
				fg = muted
			}
			return style.Style{}.FontSize(tokens.CellTextSize).TextColor(style.Solid(fg)).Background(style.Solid(bg)).Radius(tokens.CellSize / 2)
		},
		Range: func(theme.DatePickerRangeRequest) style.Style { return style.Style{}.Background(style.Solid(accent)) },
		DateSegment: func(request theme.DatePickerTextRequest) style.Style {
			return style.Style{}.
				FontSize(tokens.TextSize).
				Background(style.RGBA(0)).
				TextColor(style.Solid(request.Base))
		},
		DateLiteral: func(request theme.DatePickerTextRequest) style.Style {
			return style.Style{}.FontSize(tokens.TextSize).TextColor(style.Solid(request.Base))
		},
		RangeSeparator: func(request theme.DatePickerTextRequest) style.Style {
			return style.Style{}.FontSize(tokens.TextSize).TextColor(style.Solid(request.Base))
		},
		Trigger: func(request theme.DatePickerTextRequest) style.Style {
			return style.Style{}.FontSize(tokens.TextSize).TextColor(style.Solid(request.Base))
		},
		TextColor: func(value style.ResolvedStyle) color.NRGBA {
			if value.Text == nil {
				panic(theme.ErrIncompleteVisualProvider)
			}
			result, ok := behaviorColor(value.Text.Color)
			if !ok {
				panic(theme.ErrIncompleteVisualProvider)
			}
			return result
		},
		PaintColor: func(value style.ResolvedStyle) color.NRGBA {
			if value.Paint == nil {
				panic(theme.ErrIncompleteVisualProvider)
			}
			result, ok := behaviorPaintColor(value.Paint.Background)
			if !ok {
				panic(theme.ErrIncompleteVisualProvider)
			}
			return result
		},
		OutlineColor: func(value style.ResolvedStyle) color.NRGBA {
			if value.Paint == nil || value.Paint.Outline == nil {
				panic(theme.ErrIncompleteVisualProvider)
			}
			result, ok := behaviorColor(value.Paint.Outline.Color)
			if !ok {
				panic(theme.ErrIncompleteVisualProvider)
			}
			return result
		},
		FieldColors: func(content, placeholder, selection style.ResolvedStyle) theme.DatePickerFieldColors {
			return theme.DatePickerFieldColors{
				Foreground: behaviorResolvedText(content), Placeholder: behaviorResolvedText(placeholder), Selection: behaviorResolvedPaint(selection),
			}
		},
		CalendarIcon:  func(layout.Context, theme.DatePickerIconRenderRequest) {},
		YearIndicator: func(layout.Context, theme.DatePickerYearIndicatorRenderRequest) {},
		NavButton:     func(layout.Context, theme.DatePickerNavButtonRenderRequest) {},
		CellRenderer:  func(layout.Context, theme.DatePickerCellRenderRequest) {},
		RangeRenderer: func(layout.Context, theme.DatePickerRangeRenderRequest) {},
		Strike:        func(layout.Context, theme.DatePickerStrikeRenderRequest) {},
		Focus:         func(layout.Context, theme.DatePickerFocusRenderRequest) {},
	}
}
