package colorpicker

import (
	"image/color"
	"time"

	"gioui.org/layout"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/theme"
)

type ColorPickerWidget struct {
	theme     func(*theme.Theme)
	key       string
	value     color.NRGBA
	label     string
	onChange  func(color.NRGBA)
	disabled  bool
	alpha     bool
	showField bool
	presets   []color.NRGBA
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

func (picker ColorPickerWidget) Theme(fn func(*theme.Theme)) ColorPickerWidget {
	picker.theme = fn
	return picker
}

func (picker ColorPickerWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	if restore := frame.PushInstanceTheme(ctx, picker.theme); restore != nil {
		defer restore()
	}
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
	dimensions := picker.layoutTrigger(ctx, triggerGtx, pickerState, enabled)
	progress := pickerState.popoverProgress(gtx, pickerState.open && enabled, frame.ActiveTheme(ctx).Motion)
	if progress == 0 && !pickerState.open {
		return dimensions
	}

	picker.layoutPopover(ctx, gtx, pickerState, key, dimensions, progress, frame.OverlayNaturallyDisabled(gtx))
	return dimensions
}
