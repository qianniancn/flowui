package datepicker

import (
	"image/color"

	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

// DatePicker parts are the stable visual customization points owned by the
// date picker. They change the appearance of existing calendar elements; they
// do not replace the component's interaction or layout structure.
const (
	PartCalendarPanel      = theme.DatePickerPartCalendarPanel
	PartCalendarNav        = theme.DatePickerPartCalendarNav
	PartCalendarHeader     = theme.DatePickerPartCalendarHeader
	PartCalendarWeekday    = theme.DatePickerPartCalendarWeekday
	PartCalendarCell       = theme.DatePickerPartCalendarCell
	PartCalendarRange      = theme.DatePickerPartCalendarRange
	PartCalendarFocus      = theme.DatePickerPartCalendarFocus
	PartDateSegment        = theme.DatePickerPartDateSegment
	PartDateLiteral        = theme.DatePickerPartDateLiteral
	PartDateRangeSeparator = theme.DatePickerPartDateRangeSeparator
	PartDateTrigger        = theme.DatePickerPartDateTrigger
)

// datePickerDateSegmentDeclaration supplies the field-derived base color while
// keeping focused and invalid visuals in the DatePicker Part protocol.
func datePickerDateSegmentDeclaration(visual theme.DatePickerVisual, base color.NRGBA, states ...style.State) style.Style {
	var state style.State
	if len(states) > 0 {
		state = states[0]
	}
	return style.Style{}.Part(PartDateSegment, visual.DateSegment(theme.DatePickerTextRequest{Base: base, State: state}))
}

func datePickerTextPartDeclaration(visual theme.DatePickerVisual, part style.Part, request theme.DatePickerTextRequest) style.Style {
	switch part {
	case PartDateLiteral:
		return style.Style{}.Part(part, visual.DateLiteral(request))
	case PartDateRangeSeparator:
		return style.Style{}.Part(part, visual.RangeSeparator(request))
	case PartDateTrigger:
		return style.Style{}.Part(part, visual.Trigger(request))
	default:
		panic(theme.ErrIncompleteVisualProvider)
	}
}

func datePickerCellDeclarationFromVisual(visual theme.DatePickerVisual, cell CellStyleContext) style.Style {
	state := style.State{
		Hovered:  cell.Hovered,
		Pressed:  cell.Pressed,
		Selected: cell.Selected,
		Disabled: cell.Disabled,
	}
	return visual.Cell(theme.DatePickerCellRequest{
		State:   state,
		Today:   cell.Today,
		Outside: cell.Outside,
		Active:  cell.Active,
		Range:   cell.Range,
	})
}

func datePickerRangeDeclarationFromVisual(visual theme.DatePickerVisual, outside bool) style.Style {
	return visual.Range(theme.DatePickerRangeRequest{Outside: outside})
}

func datePickerTextColor(visual theme.DatePickerVisual, value style.ResolvedStyle) color.NRGBA {
	return visual.TextColor(value)
}

func datePickerPaintColor(visual theme.DatePickerVisual, value style.ResolvedStyle) color.NRGBA {
	return visual.PaintColor(value)
}

func datePickerOutlineColor(visual theme.DatePickerVisual, value style.ResolvedStyle) color.NRGBA {
	return visual.OutlineColor(value)
}

// CellStyleContext carries the per-cell interaction state a day cell
// resolves against. It is a parameter packet rather than a replaceable-visual
// seam: the flags are all same-typed booleans, so a named struct keeps the call
// site readable and order-proof.
type CellStyleContext struct {
	Hovered  bool
	Pressed  bool
	Selected bool
	Today    bool
	Outside  bool
	Disabled bool
	Active   bool
	Range    bool
}

func datePickerCellContextKey(cell CellStyleContext) uint8 {
	var key uint8
	if cell.Hovered {
		key |= 1 << 0
	}
	if cell.Pressed {
		key |= 1 << 1
	}
	if cell.Selected {
		key |= 1 << 2
	}
	if cell.Today {
		key |= 1 << 3
	}
	if cell.Outside {
		key |= 1 << 4
	}
	if cell.Disabled {
		key |= 1 << 5
	}
	if cell.Active {
		key |= 1 << 6
	}
	if cell.Range {
		key |= 1 << 7
	}
	return key
}

func datePickerStyleStateKey(state style.State) uint16 {
	var key uint16
	flags := [...]bool{
		state.Hovered, state.Pressed, state.Focused, state.FocusVisible,
		state.Disabled, state.Selected, state.Checked, state.Indeterminate,
		state.ReadOnly, state.Invalid, state.Loading, state.Open,
		state.Expanded, state.Dragging, state.DropTarget, state.Warning,
	}
	for index, value := range flags {
		if value {
			key |= 1 << index
		}
	}
	return key
}

func datePickerResolvedCellKey(context CellStyleContext, state style.State) uint32 {
	return uint32(datePickerCellContextKey(context))<<16 | uint32(datePickerStyleStateKey(state))
}

func datePickerTextResolvedKey(state style.State, base color.NRGBA) uint64 {
	return uint64(datePickerStyleStateKey(state))<<32 |
		uint64(base.R)<<24 | uint64(base.G)<<16 | uint64(base.B)<<8 | uint64(base.A)
}
