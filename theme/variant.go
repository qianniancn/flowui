package theme

// Tone is the semantic hue dimension shared by every component. It replaces
// the per-component colour enums of v1 (ButtonVariant's color axis, BadgeColor,
// AlertStatus, ...) with one vocabulary the whole library understands.
type Tone uint8

const (
	ToneDefault   Tone = iota // un-accented control: neutral surface
	TonePrimary               // the brand hue
	ToneSecondary             // muted accent, weaker than primary
	ToneHover                 // neutral interaction surface
	ToneSuccess
	ToneWarning
	ToneDanger
)

// ToneCount is the number of defined tones; Semantic indexes its pairs by it.
const ToneCount = int(ToneDanger) + 1

// ToneIsHue reports whether a tone draws its fill from a hue ramp, as opposed
// to a neutral surface role.
//
// The split matters because the two families are derived differently: a hue
// tone comes from Scales.Scale(t) through Scales.DeriveTone, while ToneDefault
// and ToneHover are neutral surfaces that no hue can produce and that style
// packages therefore construct directly. HoverSurface is the neutral
// interaction surface for a style whose brand colour is not its hover colour;
// v1 called the brand colour Accent, so v2 keeps the two apart by name.
//
// Out-of-range values report false, matching Semantic.Tone's fallback to
// ToneDefault.
func ToneIsHue(t Tone) bool {
	switch t {
	case TonePrimary, ToneSecondary, ToneSuccess, ToneWarning, ToneDanger:
		return true
	}
	return false
}

// String names the tone for diagnostics and test failures.
func (t Tone) String() string {
	switch t {
	case ToneDefault:
		return "default"
	case TonePrimary:
		return "primary"
	case ToneSecondary:
		return "secondary"
	case ToneHover:
		return "hover"
	case ToneSuccess:
		return "success"
	case ToneWarning:
		return "warning"
	case ToneDanger:
		return "danger"
	}
	return "tone(?)"
}

// Variant is the structural dimension of a control's visual treatment — the
// other half of the (Tone, Variant) pair that replaces the per-component
// one-dimensional enums of v1 (ButtonVariant, BadgeVariant, AlertStatus, ...).
//
// The vocabulary is the union of every component's structural variants, which is
// why a single enum can hold both the button's surface treatments and the input's
// field faces. Most of it is *vocabulary* rather than implementation: package
// styles resolves Solid, Soft, Outline and Ghost today and answers everything
// else with the solid treatment, so a member here is not a promise that anything
// draws it. Each component declares the subset it renders (VariantSet) and
// interprets the member itself: `filled` can mean a button tint or an input
// face, and those are different visuals.
type Variant uint8

const (
	// VariantSolid is a filled surface.
	VariantSolid Variant = iota
	// VariantSoft is a low-emphasis tinted fill. Flat and faded strengths
	// collapse into this one semantic.
	VariantSoft
	// VariantOutline is a border with no fill.
	VariantOutline
	// VariantGhost is an unfilled surface that still takes a hover fill.
	VariantGhost
	// VariantText is bare text with no fill and no hover fill.
	VariantText
	// VariantLink is text that reads as a link.
	VariantLink
	// VariantDashed is an outline drawn with a dashed stroke.
	VariantDashed
	// VariantFilled is a filled input face.
	VariantFilled
	// VariantUnderlined draws only the bottom edge.
	VariantUnderlined
)

// VariantCount is the number of defined variants; VariantSet indexes by it.
const VariantCount = int(VariantUnderlined) + 1

// String names the variant for diagnostics and test failures.
func (v Variant) String() string {
	switch v {
	case VariantSolid:
		return "solid"
	case VariantSoft:
		return "soft"
	case VariantOutline:
		return "outline"
	case VariantGhost:
		return "ghost"
	case VariantText:
		return "text"
	case VariantLink:
		return "link"
	case VariantDashed:
		return "dashed"
	case VariantFilled:
		return "filled"
	case VariantUnderlined:
		return "underlined"
	}
	return "variant(?)"
}

// VariantSet declares which variants a component renders.
//
// The zero value means "all of them". A component that has not stated its
// subset must keep working — the vocabulary is the union of every component's
// variants, and most of them predate the split — and the framework's own
// components state their subset explicitly, which is what turns it into
// something a test can check instead of something a reader reconstructs from a
// switch statement.
//
// ShapeSet is the same pattern for the outline dimension.
type VariantSet uint16

// VariantSetAll covers every variant in the vocabulary.
const VariantSetAll VariantSet = VariantSet(1<<VariantCount) - 1

// Has reports whether the component declares support for a variant.
func (s VariantSet) Has(v Variant) bool {
	if int(v) >= VariantCount {
		return false
	}
	return s == 0 || s&(1<<v) != 0
}

// Resolve returns a variant the component can render, falling back to the
// nearest one it does declare. It never reports failure: an unsupported variant
// is a no-op rather than an error, because the caller is a builder method with
// no error channel.
//
// Three steps, in order: the request itself, its entry in variantFallbacks,
// then the lowest variant the set declares. The last step is why a component
// that excludes VariantSolid still gets an answer it can render — falling back
// to solid unconditionally would hand back a variant the component never
// declared.
//
// The last step always finds a candidate: VariantSet is uint16, the valid bits
// are exactly the low VariantCount of them, and so a non-zero set has at least
// one valid bit. Only a set with no valid bits at all reaches VariantSolid,
// which is the best available answer for a malformed declaration.
func (s VariantSet) Resolve(v Variant) Variant {
	if s.Has(v) {
		return v
	}
	for _, candidate := range variantFallbacks[v] {
		if s.Has(candidate) {
			return candidate
		}
	}
	for value := range VariantCount {
		if s.Has(Variant(value)) {
			return Variant(value)
		}
	}
	return VariantSolid
}

// variantFallbacks orders the candidates for a variant a component does not
// render. The chains follow shape rather than emphasis: a dashed outline falls
// back to a plain outline before a filled surface, because an outline with the
// wrong stroke is closer to the author's intent than a solid block.
//
// A variant absent from the table skips this step, and Resolve then takes the
// lowest variant the set declares — not VariantSolid, which the component may
// not have declared either.
var variantFallbacks = map[Variant][]Variant{
	VariantSoft:       {VariantSolid},
	VariantOutline:    {VariantSoft, VariantSolid},
	VariantGhost:      {VariantText, VariantSoft, VariantSolid},
	VariantText:       {VariantGhost, VariantSoft, VariantSolid},
	VariantLink:       {VariantText, VariantGhost, VariantSolid},
	VariantDashed:     {VariantOutline, VariantSoft, VariantSolid},
	VariantFilled:     {VariantSoft, VariantSolid},
	VariantUnderlined: {VariantOutline, VariantSoft, VariantSolid},
}

// Size is the control size dimension shared by every component that has one.
//
// SizeSmall is the zero value so that an unset size reads as the smallest
// control rather than the largest: the existing per-component size enums put
// their medium member at zero, and a component migrating to this type must
// therefore map explicitly instead of aliasing.
type Size uint8

const (
	SizeSmall Size = iota
	SizeMedium
	SizeLarge
)

// String names the size for diagnostics and test failures.
func (s Size) String() string {
	switch s {
	case SizeSmall:
		return "small"
	case SizeMedium:
		return "medium"
	case SizeLarge:
		return "large"
	}
	return "size(?)"
}

// Scale returns the ramp a tone draws from. ToneHover is a neutral surface, so
// it reads the neutral ramp rather than a hue of its own.
func (s Scales) Scale(t Tone) ColorScale {
	switch t {
	case TonePrimary:
		return s.Primary
	case ToneSecondary:
		return s.Secondary
	case ToneSuccess:
		return s.Success
	case ToneWarning:
		return s.Warning
	case ToneDanger:
		return s.Danger
	default:
		return s.Neutral
	}
}
