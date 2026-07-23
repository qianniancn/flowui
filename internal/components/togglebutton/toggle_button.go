package togglebutton

import (
	"time"

	"gioui.org/layout"
	"github.com/qianniancn/FlowUI/internal/frame"
	flowstyle "github.com/qianniancn/FlowUI/internal/style"
)

type ToggleButtonVariant uint8

const (
	ToggleButtonDefault ToggleButtonVariant = iota
	ToggleButtonGhost
)

type ToggleButtonSize uint8

const (
	ToggleButtonMedium ToggleButtonSize = iota
	ToggleButtonSmall
	ToggleButtonLarge
)

const (
	toggleButtonColorDuration = 100 * time.Millisecond
	toggleButtonScaleDuration = 250 * time.Millisecond
)

type ToggleButtonWidget struct {
	key         string
	selected    bool
	child       frame.Widget
	onChange    func(bool)
	variant     ToggleButtonVariant
	size        ToggleButtonSize
	disabled    bool
	iconOnly    bool
	label       string
	customStyle flowstyle.Style
}

func ToggleButton(key string, selected bool, child frame.Widget) ToggleButtonWidget {
	return ToggleButtonWidget{key: key, selected: selected, child: child}
}

func (b ToggleButtonWidget) OnChange(onChange func(bool)) ToggleButtonWidget {
	b.onChange = onChange
	return b
}

func (b ToggleButtonWidget) Variant(variant ToggleButtonVariant) ToggleButtonWidget {
	b.variant = variant
	return b
}

func (b ToggleButtonWidget) Size(size ToggleButtonSize) ToggleButtonWidget {
	b.size = size
	return b
}

func (b ToggleButtonWidget) Disabled(disabled bool) ToggleButtonWidget {
	b.disabled = disabled
	return b
}

func (b ToggleButtonWidget) IconOnly() ToggleButtonWidget {
	b.iconOnly = true
	return b
}

func (b ToggleButtonWidget) Label(label string) ToggleButtonWidget {
	b.label = label
	return b
}

func (b ToggleButtonWidget) Style(value flowstyle.Style) ToggleButtonWidget {
	b.customStyle = value
	return b
}

func (b ToggleButtonWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	return b.layout(ctx, gtx)
}
