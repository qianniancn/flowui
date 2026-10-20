// Package appearance owns the shared (Tone, Variant) → visual resolution that every
// component's design table feeds into.
//
// It maps tones and variants to common roles, such as a neutral outlined
// surface carrying danger text. Providers supply values and design tables,
// while this package keeps their colour interpretation in one place.
//
// The package deliberately depends on theme and style only. It knows nothing
// about widgets, which is what lets a component's design table live beside the
// component while the rules live here.
package appearance

import (
	"image/color"

	"gioui.org/unit"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

// Design is one component variant's reading of a tone.
//
// Tone is the semantic hue; Accent is the tone whose interaction steps the
// control highlights with. They differ for the neutral outlined button: it
// strokes and labels neutrally but highlights in the brand colour, because
// ToneDefault has no hue to step.
type Design struct {
	Tone    theme.Tone
	Variant theme.Variant
	Accent  theme.Tone

	// NeutralSurface routes a soft or outlined surface to the neutral control
	// face instead of the tone's own low-emphasis colour.
	//
	// It is not a convenience switch over one shape: a soft button and a
	// variant-soft *tint* are different things (Secondary is the neutral control
	// face; DangerSoft is a danger tint), and an outlined button and a toned
	// outline differ in where all four of surface, stroke, label and interaction
	// come from. Resolve reads the flag per family rather than driving both from
	// it.
	NeutralSurface bool

	// NeutralBorder strokes and labels with the neutral outline roles rather
	// than with the tone's fill.
	NeutralBorder bool

	// UnderShadow is the resting drop shadow's alpha; zero draws none. It is a
	// value rather than a constant because the three shapes differ (a filled
	// action casts more than an outlined one) and a disabled control casts none.
	UnderShadow float32
	// ShadowBlack takes the shadow's colour from black rather than the tone's
	// fill.
	ShadowBlack bool
}

// Colors is the resolved palette for one design. Every field is a static value.
//
// Fg is the resting label. HoverFg, HoverBorder, PressedFg and PressedBorder are
// zero when the design does not recolor that property on that state, which is
// the same "zero means inherit" convention the palette has always used.
type Colors struct {
	Bg, Hover, Pressed       color.NRGBA
	Fg, Border               color.NRGBA
	HasBorder                bool
	HoverFg, HoverBorder     color.NRGBA
	PressedFg, PressedBorder color.NRGBA
	Shadow                   color.NRGBA
}

// Resolve returns the static colours for a design.
//
// The rules below are what the three families agree on, verified value by value
// against their shipped themes:
//
//   - A filled control steps its whole surface through the tone's fill ladder.
//   - A soft control keeps the neutral tone's low-emphasis pair; only a toned
//     tint (DangerSoft) reads its own tone's soft colour.
//   - An outlined control never moves its surface. Its stroke and label are
//     neutral roles for the neutral family and the tone's fill for a toned one,
//     and its interaction recolours label and stroke — never the surface.
//   - A ghost control is transparent and takes a fill only on interaction.
func Resolve(t *theme.Theme, design Design) Colors {
	semantic := t.Semantic()
	tone := semantic.Tone(design.Tone)
	accent := semantic.Tone(design.Accent)
	neutral := semantic.Tone(theme.ToneDefault)
	transparent := color.NRGBA{}

	// The neutral control face is the neutral tone's own fill, so both names
	// resolve to one value; the role is kept because it says what it is for.
	surface, surfaceHover := semantic.ControlSurface, semantic.ControlSurfaceHover
	if !design.NeutralSurface {
		surface, surfaceHover = tone.Soft, tone.SoftHover
	}

	var result Colors
	switch design.Variant {
	case theme.VariantSolid:
		result = Colors{Bg: tone.Fill, Hover: tone.Hover, Pressed: tone.Pressed, Fg: tone.OnFill}

	case theme.VariantSoft:
		// Pressing deepens the hovered tint by one step; the hover state alone
		// must not already read as pressed.
		pressed := theme.Mix(surfaceHover, tone.OnSoft, 8)
		result = Colors{Bg: surface, Hover: surfaceHover, Pressed: pressed, Fg: tone.OnSoft}

	case theme.VariantGhost:
		result = Colors{Bg: transparent, Hover: tone.Fill, Pressed: tone.Hover, Fg: tone.OnFill}

	case theme.VariantOutline:
		// The resting surface is the neutral control face for both families; the
		// hand-filled palettes route it through the neutral tone's fill, which is
		// the same value.
		result = Colors{Bg: neutral.Fill, Hover: neutral.Fill, Pressed: neutral.Fill, HasBorder: true}
		if design.NeutralBorder {
			result.Fg = semantic.ControlOutlineForeground
			result.Border = semantic.ControlOutline
		} else {
			result.Fg = tone.Fill
			result.Border = tone.Fill
		}
		result.HoverFg, result.HoverBorder = accent.Hover, accent.Hover
		result.PressedFg, result.PressedBorder = accent.Pressed, accent.Pressed

	default:
		// A variant this table does not model resolves to a filled control rather
		// than to nothing: a component that asks for a treatment it does not
		// implement gets its tone's primary reading, not a transparent box.
		//
		// That covers VariantText, VariantLink, VariantDashed, VariantFilled and
		// VariantUnderlined — the last two are field faces, which belong to
		// ResolveField, and VariantDashed cannot be drawn at all because
		// style.BorderStyle carries no dash. Reaching this arm is a *fallback*,
		// not an implementation, so TestResolveModelsExactlyTheDeclaredVariants
		// fails the moment one of them starts resolving differently: a treatment
		// has to be declared here, not arrive by accident.
		result = Colors{Bg: tone.Fill, Hover: tone.Hover, Pressed: tone.Pressed, Fg: tone.OnFill}
	}

	if design.UnderShadow > 0 {
		base := tone.Fill
		if design.ShadowBlack {
			base = color.NRGBA{A: 0xff}
		}
		result.Shadow = style.WithAlpha(base, design.UnderShadow).Color
	}
	return result
}

// ToneColors is the colour role set a component reads off one tone, with the
// naming each family uses for those roles.
//
// It exists because two different component shapes want the same five values
// under different names. An action button asks for the fill it rests on and the
// label above it; a status badge asks for its solid, its tinted variant and the
// neutral entry that has no hue. Both are reading TonePair — this type just says
// so once, so a component does not have to decide independently what "the danger
// colour" means.
type ToneColors struct {
	// Fill is the solid surface and OnFill the label above it.
	Fill, OnFill color.NRGBA
	// Hover and Pressed are the interaction steps of the fill.
	Hover, Pressed color.NRGBA
	// Soft is the low-emphasis surface and OnSoft the label above it.
	Soft, OnSoft color.NRGBA
	// SoftHover is the low-emphasis surface's interaction step.
	SoftHover color.NRGBA
}

// Tone reads one tone's colour roles.
//
// Callers pass the tone they mean; semantic.Tone already falls back to
// ToneDefault for an out-of-range value, so this needs no branch of its own.
//
// Use it for the *tone arm* of a component's colour axis only. Two patterns
// recur and neither belongs here:
//
//   - A value on the axis that is not a tone at all. A spinner's "current"
//     colour means "read as text" and takes the canvas foreground; a progress
//     bar's "default" colour takes it too, while a progress circle's takes the
//     neutral control's label. Resolve that case before calling Tone.
//   - A component whose neutral entry has its own surface rather than the
//     neutral tone's — a chip's tertiary surface, an alert's panel surface, an
//     avatar's halved fill. Keep that surface in the component, because
//     ToneDefault would change it.
//
// The rule of thumb: Tone answers "what colours does this tone have", not "what
// colour does this component use when it has no tone".
func Tone(t *theme.Theme, tone theme.Tone) ToneColors {
	return ToneColorsOf(t.Semantic(), tone)
}

// ToneColorsOf reads one tone's colour roles off an already-resolved semantic
// layer, for callers that hold one (a rule table resolving several tones at once
// should not re-enter the theme for each).
//
// It is the same read as Tone; the two exist so a caller does not have to choose
// between re-resolving the theme and reaching past this package into the pair's
// fields.
func ToneColorsOf(semantic theme.Semantic, tone theme.Tone) ToneColors {
	pair := semantic.Tone(tone)
	return ToneColors{
		Fill: pair.Fill, OnFill: pair.OnFill,
		Hover: pair.Hover, Pressed: pair.Pressed,
		Soft: pair.Soft, OnSoft: pair.OnSoft, SoftHover: pair.SoftHover,
	}
}

// FieldVariant is the *other* structural axis: the face an editable field
// presents, rather than the treatment an action button presents.
//
// It is a separate vocabulary because the two families answer different
// questions. A button variant says how prominent an action is (solid, soft,
// outline); a field face says what shape the editable area has (a bordered box,
// a bordered box on a tinted surface, a bottom rule, nothing). No shipped style
// treats them as one axis, and forcing them together would make every style
// declare two unrelated subsets in one enum.
type FieldVariant uint8

const (
	// FieldPrimary is the boxed field: the surface role plus a full border.
	FieldPrimary FieldVariant = iota
	// FieldSecondary is the same box on the neutral control face, without the
	// control's resting shadow.
	FieldSecondary
	// FieldFilled is a tinted surface with no border.
	FieldFilled
	// FieldBorderless is a transparent surface.
	FieldBorderless
	// FieldUnderlined keeps the surface and collapses the border to a bottom
	// rule.
	FieldUnderlined
)

// String names the face for diagnostics and test failures.
func (v FieldVariant) String() string {
	switch v {
	case FieldPrimary:
		return "primary"
	case FieldSecondary:
		return "secondary"
	case FieldFilled:
		return "filled"
	case FieldBorderless:
		return "borderless"
	case FieldUnderlined:
		return "underlined"
	}
	return "field(?)"
}

// FieldSurface is the resolved surface trio for one field face, plus the resting
// shadow's opacity.
//
// Focused is separate from Hover rather than derived: a field's focused surface
// is its own role, and a style may give it the canvas back instead of
// stepping the hover colour.
type FieldSurface struct {
	Background style.ColorSource
	Hover      style.ColorSource
	Focused    style.ColorSource
	// ShadowOpacity is the resting control shadow the face carries; zero when
	// the face is flat.
	ShadowOpacity float32
}

// ResolveField returns the surface trio for a field face.
//
// shadowOpacity is the field's configured resting shadow, threaded through
// rather than decided here because it is a component token (InputTheme and
// TextAreaTheme each carry one) and only two of the five faces keep it.
func ResolveField(t *theme.Theme, variant FieldVariant, shadowOpacity float32) FieldSurface {
	semantic := t.Semantic()
	neutral := semantic.Tone(theme.ToneDefault)
	transparent := style.RGBA(0)

	surface := FieldSurface{
		Background:    style.Solid(semantic.FieldBackground),
		Hover:         style.Solid(semantic.FieldHover),
		Focused:       style.Solid(semantic.FieldFocus),
		ShadowOpacity: shadowOpacity,
	}

	switch variant {
	case FieldSecondary, FieldFilled:
		// The tinted faces sit on the neutral control face and are flat: the
		// shadow is the boxed field's depth cue, and these have no box.
		surface.Background = style.Solid(neutral.Fill)
		surface.Hover = style.Solid(neutral.Hover)
		surface.Focused = style.Solid(neutral.Fill)
		surface.ShadowOpacity = 0
	case FieldBorderless:
		surface.Background, surface.Hover, surface.Focused = transparent, transparent, transparent
		surface.ShadowOpacity = 0
	case FieldUnderlined:
		// The bottom rule replaces the box, so the surface keeps the field
		// background and does not change on hover.
		surface.Hover = surface.Background
		surface.ShadowOpacity = 0
	}
	return surface
}

// Declaration lowers a design into the declaration a component hands to the
// cascade: the resting appearance plus a rule per interaction state.
//
// borderWidth and shadowScale are the calling component's own geometry tokens
// (Components.X.BorderWidth and .UnderShadow), threaded through the way
// ResolveField takes a field's configured shadow: they belong to a component and
// this package owns none, so reading them off a named component here would make
// every other caller inherit that component's geometry.
//
// The variant is chosen at build time and does not enter the rule table, so a
// component's Rank/When ordering and its style hash are unaffected by how many
// variants the vocabulary grows to.
func Declaration(t *theme.Theme, design Design, borderWidth unit.Dp, shadowScale unit.Dp) style.Style {
	colors := Resolve(t, design)
	declaration := style.Style{}.Background(style.Solid(colors.Bg)).TextColor(style.Solid(colors.Fg))
	if colors.HasBorder {
		declaration = declaration.
			BorderColor(style.Solid(colors.Border)).
			BorderWidth(borderWidth)
	}
	// A zero scale means the calling component opts out of the resting shadow, so
	// the colour being non-transparent is not on its own a reason to draw one.
	if shadowScale > 0 && colors.Shadow.A != 0 {
		declaration = declaration.BoxShadow(0, shadowScale, 0, 0, style.Solid(colors.Shadow))
	}

	hover := style.Style{}.Background(style.Solid(colors.Hover))
	if colors.HoverFg.A != 0 {
		hover = hover.TextColor(style.Solid(colors.HoverFg))
	}
	if colors.HasBorder && colors.HoverBorder.A != 0 {
		hover = hover.BorderColor(style.Solid(colors.HoverBorder))
	}

	pressed := style.Style{}.Background(style.Solid(colors.Pressed))
	if colors.PressedFg.A != 0 {
		pressed = pressed.TextColor(style.Solid(colors.PressedFg))
	}
	if colors.HasBorder && colors.PressedBorder.A != 0 {
		pressed = pressed.BorderColor(style.Solid(colors.PressedBorder))
	}

	return declaration.
		On(style.Hovered, hover).
		On(style.Pressed, pressed)
}
