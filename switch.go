package flowui

import (
	"time"

	"gioui.org/io/semantic"
	"gioui.org/layout"
)

type SwitchWidget struct {
	key         string
	checked     bool
	label       string
	description string
	onChange    func(bool)
	disabled    bool
	invalid     bool
	size        SwitchSize
	labelBefore bool
	thumb       func(bool) Widget
}

type SwitchSize int

const (
	SwitchMedium SwitchSize = iota
	SwitchSmall
	SwitchLarge
)

const (
	switchSelectDuration = 300 * time.Millisecond
	switchFocusDuration  = 100 * time.Millisecond
)

func Switch(key string, checked bool, label string) SwitchWidget {
	return SwitchWidget{
		key:     key,
		checked: checked,
		label:   label,
	}
}

func (s SwitchWidget) OnChange(fn func(bool)) SwitchWidget {
	s.onChange = fn
	return s
}

func (s SwitchWidget) Description(description string) SwitchWidget {
	s.description = description
	return s
}

func (s SwitchWidget) Disabled(disabled bool) SwitchWidget {
	s.disabled = disabled
	return s
}

func (s SwitchWidget) Invalid(invalid bool) SwitchWidget {
	s.invalid = invalid
	return s
}

func (s SwitchWidget) Size(size SwitchSize) SwitchWidget {
	s.size = size
	return s
}

func (s SwitchWidget) LabelBefore() SwitchWidget {
	s.labelBefore = true
	return s
}

func (s SwitchWidget) Thumb(content func(checked bool) Widget) SwitchWidget {
	s.thumb = content
	return s
}

func (s SwitchWidget) Layout(ctx *Context, gtx layout.Context) layout.Dimensions {
	state := ctx.switchState(s.key)
	key := ctx.fullKey(s.key)
	state.value.Value = s.checked
	animGtx := gtx
	disabled := s.disabled || !gtx.Enabled()
	presses := activePresses(state.value.History())
	if disabled {
		gtx = gtx.Disabled()
	}

	dims := state.value.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		semantic.Switch.Add(gtx.Ops)
		if s.label != "" {
			semantic.LabelOp(s.label).Add(gtx.Ops)
		}
		description := s.description
		if description == "" {
			description = ctx.fieldDescription(key)
		}
		if description != "" {
			semantic.DescriptionOp(description).Add(gtx.Ops)
		}

		focusVisible := state.focusVisible(gtx.Focused(&state.value), state.value.History())
		style := switchStyleFor(ctx.Theme, state.value.Hovered(), state.value.Pressed(), disabled, s.invalid)
		style.selected = state.selection(animGtx, state.value.Value)
		style.focus = state.focusOpacity(animGtx, focusVisible && !disabled)
		return s.layoutContent(ctx, gtx, style, switchSizeStyleFor(ctx.Theme, s.size), state.value.Value)
	})
	if !disabled {
		ctx.focusOnPress(&state.value, state.value.History(), presses)
	}
	if !disabled && state.value.Value != s.checked && s.onChange != nil {
		s.onChange(state.value.Value)
	}
	return dims
}

type SwitchGroupWidget struct {
	children   []Widget
	horizontal bool
}

func SwitchGroup(children ...Widget) SwitchGroupWidget {
	return SwitchGroupWidget{children: children}
}

func (g SwitchGroupWidget) Horizontal() SwitchGroupWidget {
	g.horizontal = true
	return g
}

func (g SwitchGroupWidget) Layout(ctx *Context, gtx layout.Context) layout.Dimensions {
	theme := ctx.Theme.Components.SwitchGroup
	children := make([]layout.Widget, 0, len(g.children))
	for _, child := range g.children {
		children = append(children, func(gtx layout.Context) layout.Dimensions {
			return child.Layout(ctx, gtx)
		})
	}
	return layoutRadioItems(gtx, g.horizontal, gtx.Dp(theme.HorizontalGap), gtx.Dp(theme.VerticalGap), children)
}
