package fieldstyle

import (
	"gioui.org/io/pointer"
	"github.com/qianniancn/flowui/appearance"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/theme/internal/textrender"
)

// This file is the single visual implementation for Input, TextArea and
// InputGroup. State precedence is explicit in the rule data: matching rows are
// ordered by rank, matched-bit count and declaration order. The disabled row
// keeps rank 2 because it historically followed the two-bit validation rows.
// Provider and widget tests verify the rule shape, part effects and cascade
// behavior directly; no second visual implementation is retained as an oracle.

// BuildInput returns the complete provider-owned input declaration.
func BuildInput(activeTheme *theme.Theme, request theme.InputRequest) theme.InputVisual {
	if activeTheme == nil {
		panic(theme.ErrMissingVisualProvider)
	}
	ctx := styleContext{
		Theme: activeTheme, Variant: request.Variant, Mode: request.Mode,
		FullWidth: request.FullWidth, Height: request.Height, Size: request.Size,
	}
	var defaults style.Style
	switch request.Mode {
	case theme.InputModeTextArea:
		defaults = textAreaDefaultRules(ctx)
	case theme.InputModeGroup:
		defaults = groupDefaultRules(ctx)
	default:
		defaults = defaultRules(ctx)
	}
	return theme.InputVisual{
		Defaults: defaults,
		Override: theme.InputOverrideStyle(request),
		EditorStyle: func(content, placeholder, selection style.ResolvedStyle) theme.InputEditorStyle {
			return textrender.InputEditorStyle(activeTheme, content, placeholder, selection)
		},
	}
}

// defaultRules returns the single-line field's box declaration as a state rule
// table, in the shape the cascade's defaults slot wants.
func defaultRules(ctx styleContext) style.Style {
	tokens := ctx.Theme.Components.Input
	height := tokens.Height
	switch ctx.Size {
	case theme.InputSmall:
		if tokens.SmallHeight > 0 {
			height = tokens.SmallHeight
		}
	case theme.InputLarge:
		if tokens.LargeHeight > 0 {
			height = tokens.LargeHeight
		}
	}
	if ctx.Height > 0 {
		height = ctx.Height
	}
	radius, paddingX, textSize, lineHeight := inputTierGeometry(tokens, ctx.Size)
	return fieldRules(ctx.Theme, ctx.Variant, fieldDeclarationOptions{
		Height: height, Radius: radius, PaddingX: paddingX,
		TextSize: textSize, LineHeight: lineHeight,
		FocusRingWidth: tokens.FocusRingWidth, FocusGlowBlur: tokens.FocusGlowBlur, FocusHaloBlur: tokens.FocusHaloBlur, FocusHaloOpacity: tokens.FocusHaloOpacity, FocusShadowSpread: tokens.FocusShadowSpread, InvalidOutlineWidth: tokens.InvalidOutlineWidth,
		BorderWidth: tokens.BorderWidth,
		ShadowColor: tokens.ShadowColor, ShadowOpacity: tokens.ShadowOpacity,
		ShadowStrength: tokens.ShadowStrength, FillWidth: ctx.FullWidth,
	})
}

// textAreaDefaultRules is defaultRules for the multi-line field's tokens.
//
// Its height is the caller's, not the design's: a text area is sized by how many
// rows the widget currently holds, so the options carry the resolved height and
// the declaration takes the larger of it and the design's minimum. That value is
// geometry, so it stays in the base rather than becoming a row.
func textAreaDefaultRules(ctx styleContext) style.Style {
	tokens := ctx.Theme.Components.TextArea
	return fieldRules(ctx.Theme, ctx.Variant, fieldDeclarationOptions{
		Height: max(tokens.MinHeight, ctx.Height), Radius: tokens.Radius,
		PaddingX: tokens.PaddingX, PaddingY: tokens.PaddingY,
		TextSize: tokens.TextSize, LineHeight: tokens.LineHeight,
		FocusRingWidth: tokens.FocusRingWidth, FocusGlowBlur: tokens.FocusGlowBlur, FocusHaloBlur: tokens.FocusHaloBlur, FocusHaloOpacity: tokens.FocusHaloOpacity, FocusShadowSpread: tokens.FocusShadowSpread, InvalidOutlineWidth: tokens.InvalidOutlineWidth,
		BorderWidth: tokens.BorderWidth,
		ShadowColor: tokens.ShadowColor, ShadowOpacity: tokens.ShadowOpacity,
		ShadowStrength: tokens.ShadowStrength, FillWidth: ctx.FullWidth,
	})
}

// groupDefaultRules returns the compound group's declaration as a state rule
// table.
//
// It is the field table plus one row, and that row is the reason the group has
// six rather than five. The fluent form nests the predicate *inside* the
// PartContent declaration:
//
//	Part(PartContent, Style{}.Cursor(CursorText).
//	    When(Disabled, Style{}.Cursor(CursorNotAllowed)))
//
// The row below says the same thing from the root by naming the part in Set, which
// is the part-aware idiom §11.12 used for the progress bar. Both resolve to the
// same cursor; the root-level row is preferred here because RuleTable() reports a
// declaration's own rows and cannot see one that lives inside a part, so a
// part-level table would put this state beyond the reach of the lint that asks
// every component to declare its states.
func groupDefaultRules(ctx styleContext) style.Style {
	tokens := ctx.Theme.Components.InputGroup
	root := fieldRules(ctx.Theme, ctx.Variant, fieldDeclarationOptions{
		MinHeight: max(tokens.MinHeight, ctx.Height), Radius: tokens.Radius,
		TextSize: tokens.TextSize, LineHeight: tokens.LineHeight,
		FocusRingWidth: tokens.FocusRingWidth, FocusGlowBlur: tokens.FocusGlowBlur, FocusHaloBlur: tokens.FocusHaloBlur, FocusHaloOpacity: tokens.FocusHaloOpacity, FocusShadowSpread: tokens.FocusShadowSpread, InvalidOutlineWidth: tokens.InvalidOutlineWidth,
		BorderWidth: tokens.BorderWidth,
		ShadowColor: tokens.ShadowColor, ShadowOpacity: tokens.ShadowOpacity,
		ShadowStrength: tokens.ShadowStrength, FillWidth: ctx.FullWidth,
	})
	sem := ctx.Theme.Semantic()
	content := style.Style{}.
		PaddingX(tokens.PaddingX).
		PaddingY(tokens.TextAreaPaddingY).
		MinHeight(tokens.TextAreaMinHeight).
		Cursor(pointer.CursorText)
	return root.
		Rules(style.Rule{
			Rank: 2,
			When: style.Disabled,
			Set: style.Style{}.Part(style.PartContent, style.Style{}.
				Cursor(pointer.CursorNotAllowed)),
		}).
		Cursor(pointer.CursorDefault).
		Part(style.PartContent, content).
		Part(style.PartPrefix, style.Style{}.TextColor(style.Solid(sem.MutedForeground))).
		Part(style.PartSuffix, style.Style{}.TextColor(style.Solid(sem.MutedForeground))).
		Part(style.PartIndicator, style.Style{}.
			Width(tokens.DividerWidth).
			Background(style.Solid(sem.Border)))
}

// fieldRules builds the field box's base plus the five rows that used to be a
// When chain, and is shared by the three entry points above because they differ
// only in which tokens they pass.
//
// The rows are the whole state machine of the box: hover, focus, invalid,
// invalid-and-focused, disabled. Two of them restate a value the base already
// set -- the focused and invalid rows both repaint the background with the
// focused token -- because that is what the chain did, and the chain's later row
// winning is a fact the equivalence test checks rather than one this function
// re-derives.
func fieldRules(activeTheme *theme.Theme, variant theme.InputVariant, options fieldDeclarationOptions) style.Style {
	background, hover, _, shadowOpacity := surfaceColors(activeTheme, variant, options.ShadowOpacity)
	underlinedWidth := options.BorderWidth
	if underlinedWidth <= 0 {
		underlinedWidth = 1
	}
	spread := options.FocusShadowSpread
	sem := activeTheme.Semantic()
	// One read per tone through the shared vocabulary; the fields were already
	// being taken in pairs, so the pair is what the shared type is.
	accentTone := appearance.ToneColorsOf(sem, theme.TonePrimary)
	dangerTone := appearance.ToneColorsOf(sem, theme.ToneDanger)
	warningTone := appearance.ToneColorsOf(sem, theme.ToneWarning)
	accent, accentHover := accentTone.Fill, accentTone.Hover
	danger, dangerHover := dangerTone.Fill, dangerTone.Hover
	warning, warningHover := warningTone.Fill, warningTone.Hover

	// The focused field draws its ring as an inset border: the ring stays
	// inside the box, its corners follow the field radius exactly, and it never
	// inflates the visual extent that a clipping viewport reserves for the field.
	// Providers that express feedback as an
	// outer outline or a spread shadow keep that path through
	// FocusShadowSpread and FocusGlowBlur.
	ringSet := func(col style.ColorSource) style.Style {
		if spread > 0 || options.FocusGlowBlur > 0 {
			return style.Style{}.Outline(options.FocusRingWidth, 0, col)
		}
		return style.Style{}.
			BorderWidth(max(options.FocusRingWidth, options.BorderWidth)).
			BorderColor(col)
	}
	// Validation feedback uses the same inset stroke, one tier narrower than the
	// focus ring. InvalidOutlineWidth keeps an outline-based provider's model.
	validationSet := func(col style.ColorSource) style.Style {
		if options.InvalidOutlineWidth > 0 {
			return style.Style{}.Outline(options.InvalidOutlineWidth, 0, col)
		}
		return style.Style{}.
			BorderWidth(max(options.BorderWidth, 1)).
			BorderColor(col)
	}

	focusRow := style.Rule{When: style.Focused}
	switch variant {
	case theme.InputPrimary, theme.InputSecondary:
		focusRow.Set = ringSet(style.Solid(sem.Ring))
	case theme.InputFilled:
		focusRow.Set = ringSet(style.Solid(accent))
	case theme.InputBorderless:
		focusRow.Set = style.Style{}
	case theme.InputUnderlined:
		focusRow.Set = style.Style{}.BorderBottom(underlinedWidth, style.Solid(sem.Ring))
	}
	if spread > 0 {
		focusRow.Set = focusRow.Set.BoxShadow(0, 0, 0, spread, style.WithAlpha(accent, .1))
	}
	if options.FocusGlowBlur > 0 && variant != theme.InputBorderless {
		focusRow.Set = focusRow.Set.BoxShadowTransient(0, 0, options.FocusGlowBlur, 0, style.Solid(sem.Ring))
	}
	if options.FocusHaloBlur > 0 && variant != theme.InputBorderless && variant != theme.InputUnderlined {
		haloOpacity := options.FocusHaloOpacity
		if haloOpacity <= 0 {
			haloOpacity = 1
		}
		focusRow.Set = focusRow.Set.BoxShadowTransient(0, 0, options.FocusHaloBlur, 0, style.WithAlpha(accent, haloOpacity))
	}
	// Bordered fields darken their stroke on hover; the tinted variants keep
	// their own hover face, and the underlined rule is a hover affordance of its own.
	hoverSet := style.Style{}.Background(hover)
	switch variant {
	case theme.InputPrimary, theme.InputSecondary:
		hoverSet = hoverSet.BorderColor(style.Solid(theme.Mix(sem.Border, sem.Foreground, 18)))
	case theme.InputUnderlined:
		hoverSet = style.Style{}.BorderBottom(underlinedWidth, style.Solid(accentHover))
	}
	invalidSet := validationSet(style.Solid(danger))
	invalidHoverSet := invalidSet
	invalidFocusedSet := ringSet(style.Solid(danger))
	if spread > 0 {
		invalidSet = invalidSet.BoxShadow(0, 0, 0, spread, style.WithAlpha(danger, .06))
		invalidFocusedSet = invalidFocusedSet.BoxShadow(0, 0, 0, spread, style.WithAlpha(danger, .06))
	}

	switch variant {
	case theme.InputFilled:
		invalidSet, invalidHoverSet = validationTint(sem, theme.ToneDanger)
		invalidFocusedSet = ringSet(style.Solid(danger))
		if spread > 0 {
			invalidFocusedSet = invalidFocusedSet.BoxShadow(0, 0, 0, spread, style.WithAlpha(danger, .06))
		}
	case theme.InputBorderless:
		invalidSet = style.Style{}.TextColor(style.Solid(danger))
		invalidHoverSet = invalidSet
		invalidFocusedSet = invalidSet
	case theme.InputUnderlined:
		invalidSet = style.Style{}.BorderBottom(underlinedWidth, style.Solid(danger))
		invalidHoverSet = style.Style{}.BorderBottom(underlinedWidth, style.Solid(dangerHover))
		invalidFocusedSet = style.Style{}.BorderBottom(underlinedWidth, style.Solid(danger))
	}
	warningSet := validationSet(style.Solid(warning))
	warningHoverSet := warningSet
	warningFocusedSet := ringSet(style.Solid(warning))
	if spread > 0 {
		warningSet = warningSet.BoxShadow(0, 0, 0, spread, style.WithAlpha(warning, .1))
		warningFocusedSet = warningFocusedSet.BoxShadow(0, 0, 0, spread, style.WithAlpha(warning, .1))
	}
	switch variant {
	case theme.InputFilled:
		warningSet, warningHoverSet = validationTint(sem, theme.ToneWarning)
		warningFocusedSet = ringSet(style.Solid(warning))
		if spread > 0 {
			warningFocusedSet = warningFocusedSet.BoxShadow(0, 0, 0, spread, style.WithAlpha(warning, .1))
		}
	case theme.InputBorderless:
		warningSet = style.Style{}.TextColor(style.Solid(warning))
		warningHoverSet = warningSet
		warningFocusedSet = warningSet
	case theme.InputUnderlined:
		warningSet = style.Style{}.BorderBottom(underlinedWidth, style.Solid(warning))
		warningHoverSet = style.Style{}.BorderBottom(underlinedWidth, style.Solid(warningHover))
		warningFocusedSet = style.Style{}.BorderBottom(underlinedWidth, style.Solid(warning))
	}
	content := style.Style{}.
		Background(background).
		TextColor(style.Solid(sem.FieldForeground)).
		Radius(options.Radius).
		Overflow(style.OverflowHidden).
		Cursor(pointer.CursorText).
		Opacity(1).
		Transition(style.PropBackgroundColor, fieldDuration(activeTheme)).
		Transition(style.PropTextColor, fieldDuration(activeTheme)).
		Transition(style.PropBorderColor, fieldDuration(activeTheme)).
		Transition(style.PropOutlineColor, fieldDuration(activeTheme)).
		Transition(style.PropOpacity, fieldDuration(activeTheme))
	switch variant {
	case theme.InputUnderlined:
		content = content.Radius(0).BorderBottom(underlinedWidth, style.Solid(sem.Border))
	case theme.InputFilled, theme.InputBorderless:
		// The filled and borderless variants rest without a box border; focus draws one.
	default:
		if options.BorderWidth > 0 {
			content = content.BorderColor(style.Solid(sem.Border)).BorderWidth(options.BorderWidth)
		}
	}
	if options.Height > 0 {
		content = content.Height(options.Height)
	}
	if options.MinHeight > 0 {
		content = content.MinHeight(options.MinHeight)
	}
	if options.PaddingX != 0 {
		content = content.PaddingX(options.PaddingX)
	}
	if options.PaddingY != 0 {
		content = content.PaddingY(options.PaddingY)
	}
	if options.TextSize > 0 {
		content = content.FontSize(options.TextSize)
	}
	if options.LineHeight > 0 {
		content = content.LineHeight(options.LineHeight)
	}
	if options.FillWidth {
		content = content.FillWidth()
	}

	// The disabled row is the only one that needs a rank: the chain put it last,
	// and matched-bit count would place it before the two-bit invalid-and-focused
	// row. See the package comment.
	content = addShadows(content, activeTheme, options.ShadowColor, shadowOpacity, options.ShadowStrength).
		Rules(
			style.Rule{
				When: style.Hovered,
				Set:  hoverSet,
			},
			focusRow,
			style.Rule{
				When: style.Invalid,
				Set:  invalidSet,
			},
			style.Rule{
				When: style.Invalid | style.Hovered,
				Set:  invalidHoverSet,
			},
			style.Rule{
				When: style.Invalid | style.Focused,
				Set:  invalidFocusedSet,
			},
			style.Rule{
				When: style.Warning,
				Set:  warningSet,
			},
			style.Rule{
				When: style.Warning | style.Hovered,
				Set:  warningHoverSet,
			},
			style.Rule{
				When: style.Warning | style.Focused,
				Set:  warningFocusedSet,
			},
			// A read-only field shows the disabled face but keeps the text cursor
			// and its focus ring: it is still selectable and focusable.
			style.Rule{
				When: style.ReadOnly,
				Not:  style.Focused,
				Set:  readOnlySet(activeTheme, variant),
			},
			style.Rule{
				Rank: 2,
				When: style.Disabled,
				Set:  disabledSet(activeTheme, variant),
			},
		)
	return style.Style{}.
		Part(options.TargetPart, content).
		Part(style.PartPlaceholder, style.Style{}.TextColor(style.Solid(sem.FieldPlaceholder))).
		Part(style.PartSelection, style.Style{}.Background(style.Solid(sem.Selection)))
}

// GroupActionRules returns GroupActionWidget's chrome as a state rule table.
//
// This is the second state-dependent declaration in the package and the one the
// §11.8 census could not have found: it lives in input_group_action.go, next to
// the widget that uses it, rather than in recipe.go. Both of the shapes the
// earlier sections established for reading a predicate tree apply at once here --
// a grep for "When(" outside recipe.go is what finds it, and the census's "read
// recipe.go" step is what missed it.
//
// All four rows match one bit each, so orderedRules reproduces the chain's order
// from declaration order alone and no Rank is needed. The focus row is the one
// that needs two bitmask fields rather than a predicate tree: the chain wrote
// All(FocusVisible, Not(Disabled)), and losing the Not half would light the focus
// ring on a disabled action with no value comparison against the chain to notice,
// since the chain and the table would be wrong together.
//
// It takes the theme because its colors are concrete palette values:
// declarations quote values, so a declaration function needs the value
// source as a parameter.
// validationTint is the tinted field face a validation state shows on the filled
// variant: the tone's low-emphasis surface with its own label, and the same pair
// one interaction step further for hover.
//
// It is read through appearance.ToneColorsOf rather than off Semantic directly so the
// two deliberately separate implementations in this package — the fluent oracle
// in recipe.go and the rule table here — cannot drift on it, and so "the danger
// colour" is one decision shared with every other component.
func validationTint(sem theme.Semantic, tone theme.Tone) (resting, hovered style.Style) {
	pair := appearance.ToneColorsOf(sem, tone)
	return style.Style{}.
			Background(style.Solid(pair.Soft)).
			TextColor(style.Solid(pair.OnSoft)),
		style.Style{}.
			Background(style.Solid(pair.SoftHover)).
			TextColor(style.Solid(pair.OnSoft))
}
