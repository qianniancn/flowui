package checkbox

import (
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

type Variant = theme.CheckboxVariant

const (
	Primary   = theme.CheckboxPrimary
	Secondary = theme.CheckboxSecondary
)

// IndicatorState describes the controlled state passed to a custom indicator.
type IndicatorState struct {
	Checked       bool
	Indeterminate bool
	Disabled      bool
	Invalid       bool
}

type Widget struct {
	override       *theme.CheckboxOverride
	key            string
	checked        bool
	hasChecked     bool
	defaultChecked bool
	hasDefault     bool
	label          string
	description    string
	errorMessage   string
	onChange       func(bool)
	indicator      func(IndicatorState) core.Widget
	variant        Variant
	disabled       bool
	invalid        bool
	indeterminate  bool
	readOnly       bool
	required       bool
}

// New creates an uncontrolled checkbox that starts unchecked. Use Checked for
// controlled state and DefaultChecked for an uncontrolled starting value.
func New(label string) Widget {
	return Widget{
		key:   core.AutoKey(),
		label: label,
	}
}

// Key sets an explicit identity for the checkbox state.
func (c Widget) Key(key string) Widget {
	c.key = key
	return c
}

func (c Widget) OnChange(fn func(bool)) Widget {
	c.onChange = fn
	return c
}

func (c Widget) Checked(checked bool) Widget {
	c.checked = checked
	c.hasChecked = true
	return c
}

func (c Widget) DefaultChecked(checked bool) Widget {
	c.defaultChecked = checked
	c.hasDefault = true
	c.hasChecked = false // Switch to uncontrolled mode
	return c
}

func (c Widget) Variant(variant Variant) Widget {
	c.variant = variant
	return c
}

func (c Widget) Indeterminate(indeterminate bool) Widget {
	c.indeterminate = indeterminate
	return c
}

func (c Widget) ReadOnly(readOnly bool) Widget {
	c.readOnly = readOnly
	return c
}

func (c Widget) Required(required bool) Widget {
	c.required = required
	return c
}

func (c Widget) Description(description string) Widget {
	c.description = description
	return c
}

func (c Widget) ErrorMessage(message string) Widget {
	c.errorMessage = message
	return c
}

func (c Widget) Indicator(indicator func(IndicatorState) core.Widget) Widget {
	c.indicator = indicator
	return c
}

func (c Widget) Disabled(disabled bool) Widget {
	c.disabled = disabled
	return c
}

func (c Widget) Invalid(invalid bool) Widget {
	c.invalid = invalid
	return c
}

func (c Widget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	return c.layoutWithOverride(ctx, gtx, c.override)
}

func (c Widget) layoutWithOverride(ctx *core.Context, gtx layout.Context, override *theme.CheckboxOverride) layout.Dimensions {
	core.RequireVisualProvider(ctx)
	key, valueState := core.BoolStateWithKey(ctx, c.key)
	anim := checkboxStateFor(ctx, key)

	// Bind disclosure state
	anim.bind(c)
	checked := anim.currentChecked(c)

	valueState.Value = checked
	animGtx := gtx
	presses := core.ActivePresses(valueState.History())
	disabled := c.disabled || !gtx.Enabled()
	message := c.supportingText()
	if message != "" {
		core.PrepareFieldDescription(ctx, key, message)
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
		if !disabled {
			core.FocusOnPress(ctx, valueState, valueState.History(), presses)
		}
		focusVisible := core.FocusVisible(ctx, valueState, gtx.Focused(valueState))
		styleState := style.State{
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
		resolved := c.resolveStyle(ctx, animGtx, key, styleState, override, &anim.styleCache)
		selection := anim.selection(animGtx, checked || c.indeterminate, resolved.selectionDuration, resolved.motion)
		focus := anim.focusOpacity(animGtx, focusVisible && !disabled, resolved.focusDuration, resolved.motion)
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
			resolvedVisual:  &resolved.visual,
		}
		dims := layoutui.Styled(ctx, gtx, key, styleState, resolved.override, core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
			return c.layoutContent(ctx, gtx, options, indicatorState, resolved)
		}))
		semanticClip := clip.Rect{Max: dims.Size}.Push(gtx.Ops)
		semantic.CheckBox.Add(gtx.Ops)
		if c.label != "" {
			semantic.LabelOp(c.label).Add(gtx.Ops)
		}
		if description := c.semanticDescription(ctx, key); description != "" {
			semantic.DescriptionOp(description).Add(gtx.Ops)
		}
		semantic.SelectedOp(checked).Add(gtx.Ops)
		semantic.EnabledOp(!disabled).Add(gtx.Ops)
		semanticClip.Pop()
		return dims
	})
	return dims
}

func (c Widget) supportingText() string {
	if c.invalid && c.errorMessage != "" {
		return c.errorMessage
	}
	return c.description
}

func (c Widget) semanticDescription(ctx *core.Context, key string) string {
	if message := c.supportingText(); message != "" {
		return message
	}
	return core.FieldDescription(ctx, key)
}

func (c Widget) indicatorWidget(state IndicatorState) core.Widget {
	if c.indicator == nil {
		return nil
	}
	return c.indicator(state)
}

// Style merges instance declarations into the widget. The declaration
// reaches the provider at layout time, so Style can sit anywhere in the chain.
func (w Widget) Style(values ...style.Style) Widget {
	declaration := style.Join(values...)
	if w.override == nil {
		w.override = &theme.CheckboxOverride{}
	}
	w.override.Declaration = style.Join(w.override.Declaration, declaration)
	return w
}
