package input

import (
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"github.com/qianniancn/FlowUI/internal/frame"
	flowstyle "github.com/qianniancn/FlowUI/internal/style"
	styleruntime "github.com/qianniancn/FlowUI/internal/style/runtime"
)

const stateSlotInputGroup = "input-group"

// InputGroupWidget combines an Input or TextArea with optional prefix and suffix content
// inside one HeroUI-style field shell.
type InputGroupWidget struct {
	input              InputWidget
	textArea           TextAreaWidget
	multiline          bool
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
	customStyle        flowstyle.Style
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

func InputGroupTextArea(textArea TextAreaWidget) InputGroupWidget {
	return InputGroupWidget{
		textArea:  textArea,
		multiline: true,
		variant:   textArea.variant,
		disabled:  textArea.disabled,
		invalid:   textArea.invalid,
		fullWidth: textArea.fullWidth,
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

func (g InputGroupWidget) Style(value flowstyle.Style) InputGroupWidget {
	g.customStyle = value
	return g
}

func (g InputGroupWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	var key string
	var editor *widget.Editor
	var enabled bool
	if g.multiline {
		gtx, key, editor, enabled = g.textArea.prepareEditor(ctx, gtx, g.disabled)
	} else {
		gtx, key, editor, enabled = g.input.prepareEditor(ctx, gtx, g.disabled)
	}
	disabled := !enabled
	state := inputGroupStateFor(ctx, key)
	state.State.Update(ctx, gtx, disabled, editor)

	focused := gtx.Focused(editor)
	focusVisible := frame.FocusVisible(ctx, editor, focused)
	styleState := flowstyle.StyleState{
		Hovered: state.Hovered, Focused: focused, FocusVisible: focusVisible,
		Disabled: disabled, ReadOnly: editor.ReadOnly, Invalid: g.invalid,
	}
	activeTheme := frame.ActiveTheme(ctx)
	tokens := activeTheme.Components.InputGroup
	minHeight := tokens.MinHeight
	if g.multiline {
		contentHeight := gtx.Sp(tokens.LineHeight)*g.textArea.resolvedRows() + gtx.Dp(tokens.TextAreaPaddingY)*2
		minHeight = max(tokens.TextAreaMinHeight, gtx.Metric.PxToDp(contentHeight))
	}
	defaults := inputGroupDefaultDeclaration(activeTheme, g.variant, g.fullWidth, minHeight)
	resolved := inputGroupResolvedStyle{
		root:        styleruntime.Resolve(ctx, gtx, key, styleState, defaults, flowstyle.Style{}, flowstyle.Style{}, g.customStyle),
		prefix:      styleruntime.ResolvePart(ctx, gtx, key, flowstyle.PartPrefix, styleState, defaults, flowstyle.Style{}, flowstyle.Style{}, g.customStyle),
		suffix:      styleruntime.ResolvePart(ctx, gtx, key, flowstyle.PartSuffix, styleState, defaults, flowstyle.Style{}, flowstyle.Style{}, g.customStyle),
		divider:     styleruntime.ResolvePart(ctx, gtx, key, flowstyle.PartIndicator, styleState, defaults, flowstyle.Style{}, flowstyle.Style{}, g.customStyle),
		placeholder: styleruntime.ResolvePart(ctx, gtx, key, flowstyle.PartPlaceholder, styleState, defaults, flowstyle.Style{}, flowstyle.Style{}, g.customStyle),
		selection:   styleruntime.ResolvePart(ctx, gtx, key, flowstyle.PartSelection, styleState, defaults, flowstyle.Style{}, flowstyle.Style{}, g.customStyle),
	}
	inputStyle := resolvedInputStyle(resolved.root, resolved.placeholder, resolved.selection, activeTheme)
	textSize, lineHeight := resolvedTypography(resolved.root, tokens.TextSize, tokens.LineHeight)
	var editorLayout layout.Widget
	if g.multiline {
		editorLayout = g.textArea.editorLayoutWithTypography(ctx, key, enabled, editor, inputStyle, textSize, lineHeight)
	} else {
		editorLayout = g.input.editorLayoutWithTypography(ctx, key, enabled, editor, inputStyle, textSize, lineHeight)
	}
	return g.layoutFrame(ctx, gtx, state, resolved, enabled, editorLayout)
}
