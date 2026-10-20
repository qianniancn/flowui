// Package rulesetest holds the value-comparison helpers shared by the
// rule-table equivalence tests.
//
// A component that migrates its state styling from a fluent declaration to a
// rule table has to prove, for every interaction state, that the
// two produce the same resolved style. Doing that needs three things that are
// identical for every component: a complete state sweep, a way to compare two
// ResolvedStyle values while ignoring the one field a value comparison cannot
// read, and a readable rendering for failure messages. They live here so that
// each component's test file keeps only what is actually component-specific.
//
// This package exists for tests only; it is not part of the module's runtime
// surface.
package rulesetest

import (
	"fmt"
	"strings"

	"github.com/qianniancn/flowui/style"
)

// StateSpaceSize is the number of interaction states a StateMask can encode.
const StateSpaceSize = 1 << 16

// EveryState enumerates all interaction states by round-tripping bitmasks
// through StateMask.Split, so the sweep covers exactly the bits StateMask defines.
func EveryState() []style.State {
	states := make([]style.State, StateSpaceSize)
	for bits := range StateSpaceSize {
		states[bits] = style.StateMask(bits).Split()
	}
	return states
}

// Normalize drops the one field a value comparison cannot read. EaseFunc is a
// function value; neither side is expected to carry one, and reflect.DeepEqual
// reports any two non-nil functions as unequal regardless of behavior.
func Normalize(r style.ResolvedStyle) style.ResolvedStyle {
	for i := range r.Transitions {
		r.Transitions[i].Ease = nil
	}
	return r
}

// StateName renders a state's interaction bits, or "(none)" when it carries none.
func StateName(state style.State) string {
	v := state.Bits()
	if v == 0 {
		return "(none)"
	}
	return StateMaskBitsName(v)
}

// PartName renders a Part for failure messages.
//
// PartRoot's own value is the empty string, so printing a Part directly produces
// "part  differs for state disabled" -- a message that names the state but not
// the part, and sends the reader back to the source to find out which one
// disagreed. Components whose sweep covers PartRoot therefore have to render it
// as something; doing that here keeps each component's test file to the parts it
// actually compares.
func PartName(part style.Part) string {
	if part == style.PartRoot {
		return "root"
	}
	return string(part)
}

// Panics reports whether calling call panics.
//
// It exists for the one assertion a value comparison cannot make: a component
// declaration that dereferences its theme has no nil-theme behaviour to compare,
// and the contract worth pinning is that both implementations refuse the same
// way. The first component to need it (progress) kept a local copy; this is the
// second, which is where the shared home became worth having.
func Panics(call func()) (recovered bool) {
	defer func() {
		if recover() != nil {
			recovered = true
		}
	}()
	call()
	return false
}

// StateMaskBitsName renders a StateMask as a stable, readable list. It is hand
// written rather than derived from reflection so that the name a failure
// message prints is the same name a document would use.
func StateMaskBitsName(v style.StateMask) string {
	if v == 0 {
		return ""
	}
	bits := []struct {
		flag style.StateMask
		name string
	}{
		{style.Disabled, "disabled"},
		{style.Hovered, "hovered"},
		{style.Pressed, "pressed"},
		{style.Focused, "focused"},
		{style.FocusVisible, "focus-visible"},
		{style.Selected, "selected"},
		{style.Checked, "checked"},
		{style.Indeterminate, "indeterminate"},
		{style.ReadOnly, "read-only"},
		{style.Invalid, "invalid"},
		{style.Loading, "loading"},
		{style.Open, "open"},
		{style.Expanded, "expanded"},
		{style.Dragging, "dragging"},
		{style.DropTarget, "drop-target"},
		{style.Warning, "warning"},
	}
	parts := make([]string, 0, 4)
	for _, entry := range bits {
		if v&entry.flag != 0 {
			parts = append(parts, entry.name)
		}
	}
	return strings.Join(parts, "|")
}

// Describe renders a resolved style field by field, so a failure message names
// the property that differs instead of dumping two opaque structs.
func Describe(r style.ResolvedStyle) string {
	var b strings.Builder
	b.WriteString("box{")
	if r.Box != nil {
		fmt.Fprintf(&b, "w=%s h=%s minW=%s maxW=%s minH=%s maxH=%s fillW=%s fillH=%s aspect=%s pad=%s margin=%s overflow=%s cursor=%s",
			ptrString(r.Box.Width), ptrString(r.Box.Height), ptrString(r.Box.MinWidth), ptrString(r.Box.MaxWidth),
			ptrString(r.Box.MinHeight), ptrString(r.Box.MaxHeight), ptrString(r.Box.FillWidth), ptrString(r.Box.FillHeight),
			ptrString(r.Box.AspectRatio), ptrString(r.Box.Padding), ptrString(r.Box.Margin),
			ptrString(r.Box.Overflow), ptrString(r.Box.Cursor))
	}
	b.WriteString("} paint{")
	if r.Paint != nil {
		fmt.Fprintf(&b, "bg=%s border=%s bottom=%s radius=%s radii=%s shadows=%s outline=%s outset=%s opacity=%s",
			describePaintSource(r.Paint.Background), describeBorder(r.Paint.Border), describeBorder(r.Paint.BorderBottom),
			ptrString(r.Paint.Radius), ptrString(r.Paint.Radii), describeShadows(r.Paint.Shadows),
			describeOutline(r.Paint.Outline), ptrString(r.Paint.VisualOutset), ptrString(r.Paint.Opacity))
	}
	b.WriteString("} text{")
	if r.Text != nil {
		fmt.Fprintf(&b, "color=%s size=%s weight=%s typeface=%s style=%s lineHeight=%s lineHeightScale=%s maxLines=%s align=%s wrap=%s truncator=%s",
			describeColorSource(r.Text.Color), ptrString(r.Text.FontSize), ptrString(r.Text.FontWeight),
			ptrString(r.Text.Typeface), ptrString(r.Text.FontStyle), ptrString(r.Text.LineHeight),
			ptrString(r.Text.LineHeightScale), ptrString(r.Text.MaxLines), ptrString(r.Text.Align),
			ptrString(r.Text.Wrap), ptrString(r.Text.Truncator))
	}
	b.WriteString("} transform{")
	if r.Trans != nil {
		fmt.Fprintf(&b, "translateX=%s translateY=%s scaleX=%s scaleY=%s rotate=%s",
			ptrString(r.Trans.TranslateX), ptrString(r.Trans.TranslateY),
			ptrString(r.Trans.ScaleX), ptrString(r.Trans.ScaleY), ptrString(r.Trans.Rotate))
	}
	b.WriteString("} transitions[")
	for _, transition := range r.Transitions {
		fmt.Fprintf(&b, "%d/%v/%v ", transition.Property, transition.Duration, transition.Delay)
	}
	b.WriteString("]")
	return b.String()
}

func describeBorder(b *style.BorderStyle) string {
	if b == nil {
		return "-"
	}
	return fmt.Sprintf("%s/%s", describeColorSource(b.Color), ptrString(b.Width))
}

func describeOutline(o *style.OutlineStyle) string {
	if o == nil {
		return "-"
	}
	return fmt.Sprintf("w=%v offset=%v color=%s", o.Width, o.Offset, describeColorSource(o.Color))
}

func describeShadows(shadows []style.Shadow) string {
	if len(shadows) == 0 {
		return "-"
	}
	parts := make([]string, 0, len(shadows))
	for _, shadow := range shadows {
		parts = append(parts, fmt.Sprintf("(%v,%v,%v,%v,%s,%s)",
			shadow.OffsetX, shadow.OffsetY, shadow.Blur, shadow.Spread,
			describeColorSource(shadow.Color), ptrString(shadow.Profile)))
	}
	return strings.Join(parts, ";")
}

func describePaintSource(source style.PaintSource) string {
	switch value := source.(type) {
	case nil:
		return "-"
	case style.SolidColor:
		return fmt.Sprintf("solid%v", value.Color)
	case *style.SolidColor:
		if value == nil {
			return "-"
		}
		return fmt.Sprintf("solid%v*", value.Color)
	case style.Gradient:
		return fmt.Sprintf("gradient%v/%d", value.AngleDegrees, len(value.Stops))
	default:
		return fmt.Sprintf("%T", source)
	}
}

func describeColorSource(source style.ColorSource) string {
	if source == nil {
		return "-"
	}
	return describePaintSource(source)
}

func ptrString[T any](value *T) string {
	if value == nil {
		return "-"
	}
	return fmt.Sprint(*value)
}
