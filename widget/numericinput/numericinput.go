// Package numericinput provides a controlled number field with step controls.
package numericinput

import (
	"math"
	"strconv"
	"strings"

	giolayout "gioui.org/layout"
	"github.com/qianniancn/flowui/internal/core"
	flowlayout "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/widget/button"
	"github.com/qianniancn/flowui/widget/input"
)

// Widget is a controlled number field. Edits commit on Enter or focus loss;
// step buttons commit immediately. Incomplete edits stay local to the field.
type Widget struct {
	key       string
	value     float64
	min       float64
	max       float64
	step      float64
	precision int
	label     string
	disabled  bool
	onChange  func(float64)
}

// New creates a number field with step 1 and no range restriction.
func New(key string, value float64) Widget {
	requireFinite(value)
	return Widget{key: key, value: value, min: -math.MaxFloat64, max: math.MaxFloat64, step: 1, precision: -1}
}

func (n Widget) Range(minimum, maximum float64) Widget {
	requireFinite(minimum)
	requireFinite(maximum)
	if maximum < minimum {
		panic("flowui: numeric input maximum must be greater than or equal to minimum")
	}
	n.min, n.max = minimum, maximum
	n.value = clamp(n.value, minimum, maximum)
	return n
}

func (n Widget) Step(step float64) Widget {
	requireFinite(step)
	if step <= 0 {
		panic("flowui: numeric input step must be positive")
	}
	n.step = step
	return n
}

// Precision rounds committed values to 0..15 fractional digits.
// Use -1 (the default) for the shortest decimal representation.
func (n Widget) Precision(precision int) Widget {
	if precision < -1 || precision > 15 {
		panic("flowui: numeric input precision must be -1 or between 0 and 15")
	}
	n.precision = precision
	return n
}

func (n Widget) Label(label string) Widget {
	n.label = label
	return n
}

func (n Widget) Disabled(disabled bool) Widget {
	n.disabled = disabled
	return n
}

func (n Widget) OnChange(fn func(float64)) Widget {
	n.onChange = fn
	return n
}

func (n Widget) Value(value float64) Widget {
	requireFinite(value)
	n.value = clamp(value, n.min, n.max)
	return n
}

func (n Widget) Layout(ctx *core.Context, gtx giolayout.Context) giolayout.Dimensions {
	owner := core.ClaimKey(ctx, core.KindCustom, n.key)
	state := core.UseState[editState](ctx, owner, "numeric-input")
	fieldKey := core.DerivedKey(ctx, owner, "field")
	tag := core.FieldFocusTag(ctx, core.FullKey(ctx, fieldKey))
	focused := tag != nil && gtx.Focused(tag)
	if !state.ready || state.value != n.value || !state.dirty {
		state.draft = n.format(n.value)
		state.value, state.ready, state.dirty = n.value, true, false
	}
	commit := func(value float64) {
		value = n.normalize(value)
		state.draft, state.dirty = n.format(value), false
		if !n.disabled && gtx.Enabled() && value != n.value && n.onChange != nil {
			n.onChange(value)
		}
	}
	commitDraft := func() {
		value, err := strconv.ParseFloat(strings.TrimSpace(state.draft), 64)
		if err != nil || math.IsNaN(value) || math.IsInf(value, 0) {
			state.draft, state.dirty = n.format(n.value), false
			return
		}
		commit(value)
	}
	if state.focused && !focused && state.dirty {
		commitDraft()
	}
	state.focused = focused
	field := input.New(state.draft).Key(fieldKey).
		Type(input.Number).
		Label(n.label).
		FullWidth().
		Disabled(n.disabled).
		OnChange(func(value string) {
			state.draft, state.dirty = value, true
		}).OnSubmit(func(value string) {
		state.draft = value
		commitDraft()
	})
	step := func(delta float64) {
		value := n.value
		if parsed, err := strconv.ParseFloat(strings.TrimSpace(state.draft), 64); err == nil && !math.IsNaN(parsed) && !math.IsInf(parsed, 0) {
			value = parsed
		}
		commit(value + delta)
	}
	minus := button.New("−").Key(core.DerivedKey(ctx, owner, "decrement")).
		Variant(button.Secondary).Label(n.label + " −").
		Disabled(n.disabled || n.value <= n.min).
		OnClick(func() { step(-n.step) })
	plus := button.New("+").Key(core.DerivedKey(ctx, owner, "increment")).
		Variant(button.Secondary).Label(n.label + " +").
		Disabled(n.disabled || n.value >= n.max).
		OnClick(func() { step(n.step) })
	return flowlayout.Row(minus, flowlayout.Expanded(field), plus).Gap(4).AlignMiddle().Layout(ctx, gtx)
}

type editState struct {
	draft                 string
	value                 float64
	ready, dirty, focused bool
}

func (n Widget) format(value float64) string {
	return strconv.FormatFloat(value, 'f', n.precision, 64)
}

func (n Widget) normalize(value float64) float64 {
	value = clamp(value, n.min, n.max)
	if n.precision >= 0 {
		value, _ = strconv.ParseFloat(n.format(value), 64)
	}
	return clamp(value, n.min, n.max)
}

func requireFinite(value float64) {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		panic("flowui: numeric input requires finite values")
	}
}

func clamp(value, minimum, maximum float64) float64 {
	if value < minimum {
		return minimum
	}
	if value > maximum {
		return maximum
	}
	return value
}

var _ core.Widget = Widget{}
