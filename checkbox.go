package flowui

import (
	"time"

	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/unit"
)

type CheckboxWidget struct {
	key      string
	checked  bool
	label    string
	onChange func(bool)
	disabled bool
	invalid  bool
}

const (
	checkboxSelectDuration = 200 * time.Millisecond
	checkboxFocusDuration  = 100 * time.Millisecond
	checkboxSize           = unit.Dp(16)
	checkboxFocusSpace     = unit.Dp(2)
	checkboxBorderWidth    = unit.Dp(1)
	checkboxCheckStroke    = unit.Dp(1.5)
)

func Checkbox(key string, checked bool, label string) CheckboxWidget {
	return CheckboxWidget{
		key:     key,
		checked: checked,
		label:   label,
	}
}

func (c CheckboxWidget) OnChange(fn func(bool)) CheckboxWidget {
	c.onChange = fn
	return c
}

func (c CheckboxWidget) Disabled(disabled bool) CheckboxWidget {
	c.disabled = disabled
	return c
}

func (c CheckboxWidget) Invalid(invalid bool) CheckboxWidget {
	c.invalid = invalid
	return c
}

func (c CheckboxWidget) Layout(ctx *Context, gtx layout.Context) layout.Dimensions {
	key, state := ctx.boolStateWithKey(c.key)
	anim := ctx.checkboxState(key)
	state.Value = c.checked
	animGtx := gtx
	if c.disabled {
		gtx = gtx.Disabled()
	}

	presses := activePresses(state.History())
	dims := state.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		semantic.CheckBox.Add(gtx.Ops)
		if description := ctx.fieldDescription(key); description != "" {
			semantic.DescriptionOp(description).Add(gtx.Ops)
		}
		focusVisible := anim.focusVisible(gtx.Focused(state), state.History())
		style := checkboxStyleFor(ctx.Theme, state.Hovered(), c.disabled, c.invalid)
		style.selected = anim.selection(animGtx, state.Value)
		style.focus = anim.focusOpacity(animGtx, focusVisible && !c.disabled)
		return c.layoutContent(ctx, gtx, style)
	})
	if !c.disabled {
		ctx.focusOnPress(state, state.History(), presses)
	}
	if !c.disabled && state.Value != c.checked && c.onChange != nil {
		c.onChange(state.Value)
	}
	return dims
}
