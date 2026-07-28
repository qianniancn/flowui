package colorpicker

import (
	"image/color"
	"time"

	"gioui.org/layout"
	layoutui "github.com/qianniancn/flowui/internal/components/layout"
	"github.com/qianniancn/flowui/internal/frame"
	flowstyle "github.com/qianniancn/flowui/internal/style"
)

type ColorPickerWidget struct {
	key          string
	value        color.NRGBA
	hasValue     bool
	defaultValue color.NRGBA
	hasDefault   bool
	label        string
	onChange     func(color.NRGBA)
	disabled     bool
	alpha        bool
	showField    bool
	showRGB      bool
	showHistory  bool
	historySize  int
	presets      []color.NRGBA
	customStyle  flowstyle.Style
}

const defaultColorHistorySize = 8

const (
	colorPickerEnterDuration = 150 * time.Millisecond
	colorPickerExitDuration  = 100 * time.Millisecond
)

func ColorPicker(key string, value color.NRGBA) ColorPickerWidget {
	return ColorPickerWidget{
		key:         key,
		value:       value,
		hasValue:    true,
		showField:   true,
		showRGB:     true,
		showHistory: true,
		historySize: defaultColorHistorySize,
	}
}

func (picker ColorPickerWidget) Value(value color.NRGBA) ColorPickerWidget {
	picker.value = value
	picker.hasValue = true
	return picker
}

func (picker ColorPickerWidget) DefaultValue(value color.NRGBA) ColorPickerWidget {
	picker.defaultValue = value
	picker.hasDefault = true
	picker.hasValue = false
	return picker
}

func (picker ColorPickerWidget) Label(label string) ColorPickerWidget {
	picker.label = label
	return picker
}

func (picker ColorPickerWidget) OnChange(fn func(color.NRGBA)) ColorPickerWidget {
	picker.onChange = fn
	return picker
}

func (picker ColorPickerWidget) changeHandler(state *colorPickerState, recordHistory bool) func(color.NRGBA) {
	return func(value color.NRGBA) {
		picker.reportChange(state, value, recordHistory)
	}
}

func (picker ColorPickerWidget) Disabled(disabled bool) ColorPickerWidget {
	picker.disabled = disabled
	return picker
}

func (picker ColorPickerWidget) Alpha(enabled bool) ColorPickerWidget {
	picker.alpha = enabled
	return picker
}

// ShowField shows a hex text field. Defaults to true for desktop pickers.
func (picker ColorPickerWidget) ShowField() ColorPickerWidget {
	picker.showField = true
	return picker
}

// HideField hides the hex text field.
func (picker ColorPickerWidget) HideField() ColorPickerWidget {
	picker.showField = false
	return picker
}

// ShowRGB shows R/G/B (and A when Alpha is enabled) channel fields.
// Defaults to true for desktop pickers.
func (picker ColorPickerWidget) ShowRGB(show bool) ColorPickerWidget {
	picker.showRGB = show
	return picker
}

// ShowHistory shows recent colors chosen in this picker instance.
// Defaults to true for desktop pickers.
func (picker ColorPickerWidget) ShowHistory(show bool) ColorPickerWidget {
	picker.showHistory = show
	return picker
}

// HistorySize sets how many recent colors to keep (default 8, max 16).
func (picker ColorPickerWidget) HistorySize(size int) ColorPickerWidget {
	if size < 0 {
		size = 0
	}
	if size > 16 {
		size = 16
	}
	picker.historySize = size
	return picker
}

func (picker ColorPickerWidget) Presets(values []color.NRGBA) ColorPickerWidget {
	picker.presets = append([]color.NRGBA(nil), values...)
	return picker
}

func (picker ColorPickerWidget) reportChange(state *colorPickerState, value color.NRGBA, recordHistory bool) {
	// Keep prior hue only for achromatic colors so presets/hex/RGB stay aligned
	// with the hue slider and saturation area.
	resolved := nrgbaToHSV(value)
	hue := state.color.hsv().h
	if resolved.s > 0 && resolved.v > 0 {
		hue = resolved.h
	}
	state.color.accept(value, hue)
	if recordHistory && picker.showHistory {
		state.pushHistory(value, picker.historySize)
	}
	state.requestValue(picker, value)
}

func (picker ColorPickerWidget) Style(value flowstyle.Style) ColorPickerWidget {
	picker.customStyle = value
	return picker
}

func (picker ColorPickerWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	key, pickerState := colorPickerStateFor(ctx, picker.key)

	// Bind disclosure and get current value
	pickerState.bind(picker)
	currentValue := pickerState.currentValue(picker)
	picker.value = currentValue

	pickerState.color.sync(currentValue)

	enabled := gtx.Enabled() && !picker.disabled
	wasOpen := pickerState.open
	if !enabled {
		pickerState.open = false
	}
	// Commit the last color to history when the popover closes (not during drag).
	if wasOpen && !pickerState.open && picker.showHistory {
		pickerState.pushHistory(pickerState.color.syncedColor, picker.historySize)
	}

	triggerGtx := gtx
	if !enabled {
		triggerGtx = triggerGtx.Disabled()
	}
	dimensions := layoutui.LayoutStyled(ctx, triggerGtx, key, flowstyle.StyleState{
		Hovered:  pickerState.trigger.Hovered(),
		Pressed:  pickerState.trigger.Pressed(),
		Focused:  triggerGtx.Focused(&pickerState.trigger),
		Disabled: !enabled,
		Open:     pickerState.open,
	}, picker.customStyle, frame.WidgetFunc(func(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
		return picker.layoutTrigger(ctx, gtx, pickerState, enabled)
	}))
	progress := pickerState.popoverProgress(gtx, pickerState.open && enabled, frame.ActiveTheme(ctx).Motion)
	if progress == 0 && !pickerState.open {
		return dimensions
	}

	picker.layoutPopover(ctx, gtx, pickerState, key, dimensions, progress, frame.OverlayNaturallyDisabled(gtx))
	return dimensions
}
