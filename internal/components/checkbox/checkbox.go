package checkbox

import (
	"time"

	"gioui.org/io/semantic"
	"gioui.org/layout"
	layoutui "github.com/qianniancn/flowui/internal/components/layout"
	"github.com/qianniancn/flowui/internal/frame"
	"github.com/qianniancn/flowui/internal/state"
	flowstyle "github.com/qianniancn/flowui/internal/style"
)

type CheckboxVariant uint8

const (
	CheckboxPrimary CheckboxVariant = iota
	CheckboxSecondary
)

// IndicatorState describes the controlled state passed to a custom indicator.
type IndicatorState struct {
	Checked       bool
	Indeterminate bool
	Disabled      bool
	Invalid       bool
}

type CheckboxWidget struct {
	key            string
	checked        bool
	hasChecked     bool
	defaultChecked bool
	hasDefault     bool
	label          string
	description    string
	errorMessage   string
	onChange       func(bool)
	indicator      func(IndicatorState) frame.Widget
	variant        CheckboxVariant
	disabled       bool
	invalid        bool
	indeterminate  bool
	readOnly       bool
	required       bool
	customStyle    flowstyle.Style
}

const (
	checkboxSelectDuration = 200 * time.Millisecond
	checkboxFocusDuration  = 100 * time.Millisecond
)

func Checkbox(key string, checked bool, label string) CheckboxWidget {
	// Always treat constructor's checked parameter as controlled if called with explicit true/false
	// For uncontrolled mode, users should call Checkbox(key, false, label).DefaultChecked(value)
	return CheckboxWidget{
		key:        key,
		checked:    checked,
		hasChecked: true,
		label:      label,
	}
}

func (c CheckboxWidget) OnChange(fn func(bool)) CheckboxWidget {
	c.onChange = fn
	return c
}

func (c CheckboxWidget) Checked(checked bool) CheckboxWidget {
	c.checked = checked
	c.hasChecked = true
	return c
}

func (c CheckboxWidget) DefaultChecked(checked bool) CheckboxWidget {
	c.defaultChecked = checked
	c.hasDefault = true
	c.hasChecked = false // Switch to uncontrolled mode
	return c
}

func (c CheckboxWidget) Variant(variant CheckboxVariant) CheckboxWidget {
	c.variant = variant
	return c
}

func (c CheckboxWidget) Indeterminate(indeterminate bool) CheckboxWidget {
	c.indeterminate = indeterminate
	return c
}

func (c CheckboxWidget) ReadOnly(readOnly bool) CheckboxWidget {
	c.readOnly = readOnly
	return c
}

func (c CheckboxWidget) Required(required bool) CheckboxWidget {
	c.required = required
	return c
}

func (c CheckboxWidget) Description(description string) CheckboxWidget {
	c.description = description
	return c
}

func (c CheckboxWidget) ErrorMessage(message string) CheckboxWidget {
	c.errorMessage = message
	return c
}

func (c CheckboxWidget) Indicator(indicator func(IndicatorState) frame.Widget) CheckboxWidget {
	c.indicator = indicator
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

func (c CheckboxWidget) Style(value flowstyle.Style) CheckboxWidget {
	c.customStyle = value
	return c
}

func (c CheckboxWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	key, valueState := frame.BoolStateWithKey(ctx, c.key)
	anim := checkboxStateFor(ctx, key)

	// Bind disclosure state
	anim.bind(c)
	checked := anim.currentChecked(c)

	valueState.Value = checked
	animGtx := gtx
	presses := state.ActivePresses(valueState.History())
	disabled := c.disabled || !gtx.Enabled()
	message := c.supportingText()
	if message != "" {
		frame.PrepareFieldDescription(ctx, key, message)
	}
	if disabled {
		gtx = gtx.Disabled()
	} else if valueState.Update(gtx) {
		next := valueState.Value
		valueState.Value = checked
		if !c.readOnly {
			checked = anim.requestChecked(c, next)
		}
	}
	dims := valueState.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		semantic.CheckBox.Add(gtx.Ops)
		if c.label != "" {
			semantic.LabelOp(c.label).Add(gtx.Ops)
		}
		if description := c.semanticDescription(ctx, key); description != "" {
			semantic.DescriptionOp(description).Add(gtx.Ops)
		}
		semantic.SelectedOp(checked).Add(gtx.Ops)
		semantic.EnabledOp(!disabled).Add(gtx.Ops)
		if !disabled {
			frame.FocusOnPress(ctx, valueState, valueState.History(), presses)
		}
		focusVisible := frame.FocusVisible(ctx, valueState, gtx.Focused(valueState))
		styleState := flowstyle.StyleState{
			Hovered:       valueState.Hovered(),
			Pressed:       valueState.Pressed(),
			Focused:       gtx.Focused(valueState),
			FocusVisible:  focusVisible,
			Disabled:      disabled,
			Selected:      checked || c.indeterminate,
			Checked:       checked,
			Indeterminate: c.indeterminate,
			ReadOnly:      c.readOnly,
			Invalid:       c.invalid,
		}
		resolved := c.resolveStyle(ctx, animGtx, key, styleState)
		motion := frame.ActiveTheme(ctx).Motion
		selection := anim.selection(animGtx, checked || c.indeterminate, motion)
		focus := anim.focusOpacity(animGtx, focusVisible && !disabled, motion)
		indicatorState := IndicatorState{
			Checked: checked, Indeterminate: c.indeterminate,
			Disabled: disabled, Invalid: c.invalid,
		}
		options := ControlOptions{
			Variant: c.variant, Selection: selection, Indeterminate: c.indeterminate,
			Hovered: valueState.Hovered(), Pressed: valueState.Pressed(),
			Focused: focus, Disabled: disabled, Invalid: c.invalid,
			CustomIndicator: c.indicator != nil,
			Indicator:       c.indicatorWidget(indicatorState),
			resolvedStyle:   &resolved.indicator,
		}
		return layoutui.LayoutStyled(ctx, gtx, key, styleState, c.customStyle, frame.WidgetFunc(func(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
			return c.layoutContent(ctx, gtx, options, indicatorState, resolved)
		}))
	})
	return dims
}

func (c CheckboxWidget) supportingText() string {
	if c.invalid && c.errorMessage != "" {
		return c.errorMessage
	}
	return c.description
}

func (c CheckboxWidget) semanticDescription(ctx *frame.Context, key string) string {
	if message := c.supportingText(); message != "" {
		return message
	}
	return frame.FieldDescription(ctx, key)
}

func (c CheckboxWidget) indicatorWidget(state IndicatorState) frame.Widget {
	if c.indicator == nil {
		return nil
	}
	return c.indicator(state)
}
