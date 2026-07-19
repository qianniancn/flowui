package togglebutton

import (
	"time"

	"gioui.org/layout"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/theme"
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
	theme    func(*theme.Theme)
	key      string
	selected bool
	child    frame.Widget
	onChange func(bool)
	variant  ToggleButtonVariant
	size     ToggleButtonSize
	disabled bool
	iconOnly bool
	label    string
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

func (b ToggleButtonWidget) Theme(fn func(*theme.Theme)) ToggleButtonWidget {
	b.theme = fn
	return b
}

func (b ToggleButtonWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	if restore := frame.PushInstanceTheme(ctx, b.theme); restore != nil {
		defer restore()
	}
	return b.layout(ctx, gtx)
}
