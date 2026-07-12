package input

import (
	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/FlowUI/internal/frame"
)

const stateSlotInputGroup = "input-group"

// InputGroupWidget combines an Input with optional prefix and suffix content
// inside one HeroUI-style field shell.
type InputGroupWidget struct {
	input              InputWidget
	prefix             frame.Widget
	suffix             frame.Widget
	variant            InputVariant
	disabled           bool
	invalid            bool
	fullWidth          bool
	prefixLeftPadding  unit.Dp
	prefixRightPadding unit.Dp
	suffixLeftPadding  unit.Dp
	suffixRightPadding unit.Dp
	hasPrefixPadding   bool
	hasSuffixPadding   bool
}

func InputGroup(input InputWidget) InputGroupWidget {
	return InputGroupWidget{
		input:     input,
		variant:   input.variant,
		disabled:  input.disabled,
		invalid:   input.invalid,
		fullWidth: input.fullWidth,
	}
}

func (g InputGroupWidget) Prefix(prefix frame.Widget) InputGroupWidget {
	g.prefix = prefix
	return g
}

func (g InputGroupWidget) Suffix(suffix frame.Widget) InputGroupWidget {
	g.suffix = suffix
	return g
}

func (g InputGroupWidget) PrefixPadding(left, right int) InputGroupWidget {
	g.prefixLeftPadding = unit.Dp(max(left, 0))
	g.prefixRightPadding = unit.Dp(max(right, 0))
	g.hasPrefixPadding = true
	return g
}

func (g InputGroupWidget) SuffixPadding(left, right int) InputGroupWidget {
	g.suffixLeftPadding = unit.Dp(max(left, 0))
	g.suffixRightPadding = unit.Dp(max(right, 0))
	g.hasSuffixPadding = true
	return g
}

func (g InputGroupWidget) Variant(variant InputVariant) InputGroupWidget {
	g.variant = variant
	return g
}

func (g InputGroupWidget) Disabled(disabled bool) InputGroupWidget {
	g.disabled = disabled
	return g
}

func (g InputGroupWidget) Invalid(invalid bool) InputGroupWidget {
	g.invalid = invalid
	return g
}

func (g InputGroupWidget) FullWidth() InputGroupWidget {
	g.fullWidth = true
	return g
}

func (g InputGroupWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	gtx, key, editor, enabled := g.input.prepareEditor(ctx, gtx, g.disabled)
	disabled := !enabled
	state := inputGroupStateFor(ctx, key)
	state.State.Update(ctx, gtx, disabled, editor)

	focused := gtx.Focused(editor)
	style := inputGroupStyleFor(frame.ActiveTheme(ctx), g.variant, state.Hovered, focused, disabled, g.invalid)
	style.Background = state.Background(gtx, style.Background)
	style.Ring = state.BorderColor(gtx, style.Ring)

	inputStyle := inputStyle{
		Foreground:  style.Foreground,
		Placeholder: style.Placeholder,
		Selection:   style.Selection,
	}
	tokens := frame.ActiveTheme(ctx).Components.InputGroup
	return g.layoutFrame(ctx, gtx, state, style, enabled, g.input.editorLayoutWithTypography(
		ctx,
		key,
		enabled,
		editor,
		inputStyle,
		tokens.TextSize,
		tokens.LineHeight,
	))
}
