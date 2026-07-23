package colorpicker

import (
	"image/color"
	"time"

	"gioui.org/layout"
	layoutui "github.com/qianniancn/FlowUI/internal/components/layout"
	"github.com/qianniancn/FlowUI/internal/frame"
	flowstyle "github.com/qianniancn/FlowUI/internal/style"
)

type ColorPickerWidget struct {
	key         string
	value       color.NRGBA
	label       string
	onChange    func(color.NRGBA)
	disabled    bool
	alpha       bool
	showField   bool
	presets     []color.NRGBA
	customStyle flowstyle.Style
}

const (
	colorPickerEnterDuration = 150 * time.Millisecond
	colorPickerExitDuration  = 100 * time.Millisecond
)

func ColorPicker(key string, value color.NRGBA) ColorPickerWidget {
	return ColorPickerWidget{key: key, value: value}
}

func (picker ColorPickerWidget) Label(label string) ColorPickerWidget {
	picker.label = label
	return picker
}

func (picker ColorPickerWidget) OnChange(fn func(color.NRGBA)) ColorPickerWidget {
	picker.onChange = fn
	return picker
}

func (picker ColorPickerWidget) Disabled(disabled bool) ColorPickerWidget {
	picker.disabled = disabled
	return picker
}

func (picker ColorPickerWidget) Alpha(enabled bool) ColorPickerWidget {
	picker.alpha = enabled
	return picker
}

func (picker ColorPickerWidget) ShowField() ColorPickerWidget {
	picker.showField = true
	return picker
}

func (picker ColorPickerWidget) Presets(values []color.NRGBA) ColorPickerWidget {
	picker.presets = append([]color.NRGBA(nil), values...)
	return picker
}

func (picker ColorPickerWidget) Style(value flowstyle.Style) ColorPickerWidget {
	picker.customStyle = value
	return picker
}

func (picker ColorPickerWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	key, pickerState := colorPickerStateFor(ctx, picker.key)
	pickerState.color.sync(picker.value)

	enabled := gtx.Enabled() && !picker.disabled
	if !enabled {
		pickerState.open = false
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
