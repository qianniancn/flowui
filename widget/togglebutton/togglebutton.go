package togglebutton

import (
	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/text"
)

type Variant = theme.ToggleButtonVariant

const (
	Default = theme.ToggleButtonDefault
	Ghost   = theme.ToggleButtonGhost
)

type Size = theme.ToggleButtonSize

const (
	Medium = theme.ToggleButtonMedium
	Small  = theme.ToggleButtonSmall
	Large  = theme.ToggleButtonLarge
)

type Widget struct {
	override *theme.ToggleOverride
	key      string
	selected bool
	child    core.Widget
	onChange func(bool)
	variant  Variant
	size     Size
	disabled bool
	iconOnly bool
	label    string
}

// New creates a toggle with visible text and a matching accessible label. It
// starts unselected; Selected sets the controlled value.
func New(label string) Widget {
	return Widget{key: core.AutoKey(), child: text.New(label), label: label}
}

// Key sets an explicit identity for the toggle state.
func (b Widget) Key(key string) Widget {
	b.key = key
	return b
}

// Selected sets the controlled selection state.
func (b Widget) Selected(selected bool) Widget {
	b.selected = selected
	return b
}

// Content replaces the visible content, preserving the accessible label.
func (b Widget) Content(value core.Widget) Widget {
	b.child = value
	return b
}

func (b Widget) OnChange(onChange func(bool)) Widget {
	b.onChange = onChange
	return b
}

func (b Widget) Variant(variant Variant) Widget {
	b.variant = variant
	return b
}

func (b Widget) Size(size Size) Widget {
	b.size = size
	return b
}

func (b Widget) Disabled(disabled bool) Widget {
	b.disabled = disabled
	return b
}

func (b Widget) IconOnly() Widget {
	b.iconOnly = true
	return b
}

// Label overrides the accessible label without changing the visible content.
func (b Widget) Label(label string) Widget {
	b.label = label
	return b
}

func (b Widget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	return b.layoutWithOverride(ctx, gtx, b.override)
}

// Style merges instance declarations into the widget. The declaration
// reaches the provider at layout time, so Style can sit anywhere in the chain.
func (w Widget) Style(values ...style.Style) Widget {
	declaration := style.Join(values...)
	if w.override == nil {
		w.override = &theme.ToggleOverride{}
	}
	w.override.Declaration = style.Join(w.override.Declaration, declaration)
	return w
}
