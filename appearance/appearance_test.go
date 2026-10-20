package appearance_test

import (
	"image/color"
	"testing"

	"gioui.org/unit"

	"github.com/qianniancn/flowui/appearance"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/theme/base"
)

func coreThemes() map[string]theme.Theme {
	return map[string]theme.Theme{
		"Default":     base.Light(),
		"DefaultDark": base.Dark(),
		"Neutral":     theme.NeutralTheme(),
		"NeutralDark": theme.NeutralDarkTheme(),
		"New(seed)":   theme.New(theme.Seed{}),
	}
}

// TestResolveAlwaysProducesAPaintableControl checks that every variant in the
// shared vocabulary resolves to something a component can paint. It deliberately
// does not claim the package *implements* them — five of the nine reach the
// fallback arm, and TestResolveModelsExactlyTheDeclaredVariants is what pins
// where that boundary sits.
//
// Only the resting surface is asserted non-transparent; hover colours may
// coincide when two derived ramp steps reach the same saturation floor.
func TestResolveAlwaysProducesAPaintableControl(t *testing.T) {
	for name, value := range coreThemes() {
		active := value
		for v := range theme.VariantCount {
			variant := theme.Variant(v)
			for _, tone := range []theme.Tone{theme.ToneDefault, theme.TonePrimary, theme.ToneDanger} {
				colors := appearance.Resolve(&active, appearance.Design{Tone: tone, Variant: variant, Accent: theme.TonePrimary})
				if variant == theme.VariantGhost {
					continue // transparent at rest by definition
				}
				if colors.Bg.A == 0 {
					t.Errorf("%s: %s × %s resolves to a transparent resting surface", name, tone, variant)
				}
				if colors.Fg.A == 0 && !colors.HasBorder {
					t.Errorf("%s: %s × %s paints neither a label nor a border, so it is invisible",
						name, tone, variant)
				}
			}
		}
	}
}

// modelledVariants is the set Resolve implements. It is written out rather than
// derived from the switch, so that implementing one more variant is a deliberate
// edit here and in the enum's doc.
var modelledVariants = []theme.Variant{
	theme.VariantSolid, theme.VariantSoft, theme.VariantOutline, theme.VariantGhost,
}

// unmodelledVariants is the vocabulary Resolve answers with the solid treatment.
// Two of them are field faces and belong to ResolveField; VariantDashed cannot be
// drawn at all today, because style.BorderStyle carries no dash. These are
// statements of scope, not of a queue.
var unmodelledVariants = []theme.Variant{
	theme.VariantText, theme.VariantLink, theme.VariantDashed,
	theme.VariantFilled, theme.VariantUnderlined,
}

// TestResolveModelsExactlyTheDeclaredVariants pins the boundary between the
// treatments this package implements and the vocabulary it merely accepts.
//
// Both lists and the code are checked against each other in both directions: a
// variant declared unmodelled that stops falling back, and one declared modelled
// that resolves exactly like the fallback, both fail. Without that, "the
// vocabulary is the union of three families' words" quietly becomes "all of it
// works", which is how a reader ends up believing VariantDashed draws a dash.
func TestResolveModelsExactlyTheDeclaredVariants(t *testing.T) {
	active := theme.NeutralTheme()
	unmodelled := map[theme.Variant]bool{}
	for _, variant := range unmodelledVariants {
		unmodelled[variant] = true
	}
	declared := map[theme.Variant]bool{}
	for _, variant := range modelledVariants {
		if declared[variant] {
			t.Errorf("%s is on the modelled list twice", variant)
		}
		declared[variant] = true
	}
	for _, variant := range unmodelledVariants {
		if declared[variant] {
			t.Errorf("%s is on both lists", variant)
		}
		declared[variant] = true
	}
	// A member on neither list would slip past the boundary check below.
	for v := range theme.VariantCount {
		if !declared[theme.Variant(v)] {
			t.Errorf("%s is on neither list", theme.Variant(v))
		}
	}

	solid := resolveVariant(&active, theme.VariantSolid)
	for v := range theme.VariantCount {
		variant := theme.Variant(v)
		// VariantSolid *is* the fallback, so it cannot be told apart from it.
		if variant == theme.VariantSolid {
			continue
		}
		isFallback := resolveVariant(&active, variant) == solid
		switch {
		case unmodelled[variant] && !isFallback:
			t.Errorf("%s is declared unmodelled but no longer falls back to solid — it is implemented now; move it to modelledVariants and update the enum doc", variant)
		case !unmodelled[variant] && isFallback:
			t.Errorf("%s is declared modelled but resolves exactly like solid — the treatment is not implemented", variant)
		}
	}
}

// resolveVariant reads one variant under a tone that exercises every arm.
func resolveVariant(t *theme.Theme, variant theme.Variant) appearance.Colors {
	return appearance.Resolve(t, appearance.Design{Tone: theme.ToneDanger, Variant: variant, Accent: theme.TonePrimary})
}

// TestOutlinedFamiliesReadDifferentSources is the invariant that a single
// "neutral surface" switch used to violate: the neutral outlined control and a
// toned one take their stroke and label from different places.
//
// The interaction step is not compared across families. A seed-derived theme with
// no brand hue resolves danger and primary to the same ramp, so their outlined
// buttons legitimately highlight identically; asserting inequality there would
// demand a difference the style has no reason to invent. What is asserted is that
// each family steps the tone it is supposed to.
func TestOutlinedFamiliesReadDifferentSources(t *testing.T) {
	for name, value := range coreThemes() {
		active := value
		semantic := active.Semantic()
		neutral := appearance.Resolve(&active, appearance.Design{
			Tone: theme.ToneDefault, Variant: theme.VariantOutline,
			Accent: theme.TonePrimary, NeutralSurface: true, NeutralBorder: true,
		})
		toned := appearance.Resolve(&active, appearance.Design{
			Tone: theme.ToneDanger, Variant: theme.VariantOutline, Accent: theme.ToneDanger,
		})

		if neutral.Border == toned.Border {
			t.Errorf("%s: the neutral and danger outlines share a stroke colour %v", name, neutral.Border)
		}
		if neutral.Fg == toned.Fg {
			t.Errorf("%s: both outlined families label with %v", name, neutral.Fg)
		}
		if want := semantic.Tone(theme.ToneDanger).Hover; toned.HoverFg != want {
			t.Errorf("%s: danger outline highlights with %v, want danger.Hover %v", name, toned.HoverFg, want)
		}
		if want := semantic.Tone(theme.TonePrimary).Hover; neutral.HoverFg != want {
			t.Errorf("%s: neutral outline highlights with %v, want primary.Hover %v", name, neutral.HoverFg, want)
		}
		// The surface is the one thing they do share: an outlined control never
		// moves its surface, and both are neutral boxes.
		if neutral.Bg != toned.Bg {
			t.Errorf("%s: outlined surfaces differ (%v vs %v); both rest on the neutral control face",
				name, neutral.Bg, toned.Bg)
		}
	}
}

// TestOutlineKeepsItsSurfaceUnderInteraction pins the rule that an outlined
// control recolours its label and stroke rather than its surface — the rule that
// distinguishes it from a soft control.
func TestOutlineKeepsItsSurfaceUnderInteraction(t *testing.T) {
	for name, value := range coreThemes() {
		active := value
		colors := appearance.Resolve(&active, appearance.Design{
			Tone: theme.ToneDefault, Variant: theme.VariantOutline,
			Accent: theme.TonePrimary, NeutralSurface: true, NeutralBorder: true,
		})
		if colors.Bg != colors.Hover || colors.Bg != colors.Pressed {
			t.Errorf("%s: outlined surface moves on interaction (%v → %v → %v)",
				name, colors.Bg, colors.Hover, colors.Pressed)
		}
		if colors.HoverFg == colors.Fg && colors.HoverBorder == colors.Border {
			t.Errorf("%s: outlined control changes nothing on hover", name)
		}
		if !colors.HasBorder {
			t.Errorf("%s: outlined control reports no border", name)
		}
	}
}

// TestGhostIsTransparentAtRest checks the one variant whose resting surface is
// deliberately nothing.
func TestGhostIsTransparentAtRest(t *testing.T) {
	for name, value := range coreThemes() {
		active := value
		colors := appearance.Resolve(&active, appearance.Design{
			Tone: theme.ToneDefault, Variant: theme.VariantGhost, Accent: theme.TonePrimary,
		})
		if colors.Bg.A != 0 {
			t.Errorf("%s: ghost rests on %v, want transparent", name, colors.Bg)
		}
		if colors.Hover.A == 0 {
			t.Errorf("%s: ghost has no hover fill", name)
		}
		if colors.HasBorder {
			t.Errorf("%s: ghost reports a border", name)
		}
	}
}

// TestShadowIsOptIn pins that a design without an UnderShadow value produces no
// shadow, which is what keeps the token meaningful rather than always-on.
func TestShadowIsOptIn(t *testing.T) {
	active := theme.NeutralTheme()
	without := appearance.Resolve(&active, appearance.Design{Tone: theme.TonePrimary, Variant: theme.VariantSolid})
	if without.Shadow.A != 0 {
		t.Errorf("a design with no UnderShadow produced shadow %v", without.Shadow)
	}
	with := appearance.Resolve(&active, appearance.Design{Tone: theme.TonePrimary, Variant: theme.VariantSolid, UnderShadow: .10})
	if with.Shadow.A == 0 {
		t.Error("UnderShadow produced no shadow")
	}
	black := appearance.Resolve(&active, appearance.Design{Tone: theme.TonePrimary, Variant: theme.VariantSolid, UnderShadow: .10, ShadowBlack: true})
	if black.Shadow.R != 0 || black.Shadow.G != 0 || black.Shadow.B != 0 {
		t.Errorf("ShadowBlack produced %v, want a black base", black.Shadow)
	}
}

// TestDeclarationCarriesTheVariantAsStateRules checks the lowering: the resting
// appearance lands on the base declaration and the variant's interaction lands on
// state rules, so a component resolving the declaration sees the variant's colours
// without the variant entering the runtime rule ordering.
func TestDeclarationCarriesTheVariantAsStateRules(t *testing.T) {
	active := theme.NeutralTheme()
	design := appearance.Design{Tone: theme.ToneDanger, Variant: theme.VariantSolid, Accent: theme.TonePrimary}
	colors := appearance.Resolve(&active, design)
	declaration := appearance.Declaration(&active, design, 1, 0)

	if colors.Bg == colors.Hover {
		t.Fatal("fixture is degenerate: fill and hover are equal, so the assertion below proves nothing")
	}

	resting := declaration.Resolve(style.State{})
	if got := solidOf(resting.Paint.Background); got != colors.Bg {
		t.Errorf("resting background = %v, want the resolved fill %v", got, colors.Bg)
	}

	hovered := declaration.Resolve(style.State{Hovered: true})
	if got := solidOf(hovered.Paint.Background); got != colors.Hover {
		t.Errorf("hover background = %v, want the resolved hover %v", got, colors.Hover)
	}

	pressed := declaration.Resolve(style.State{Pressed: true})
	if got := solidOf(pressed.Paint.Background); got != colors.Pressed {
		t.Errorf("pressed background = %v, want the resolved pressed %v", got, colors.Pressed)
	}
}

// solidOf reads the colour out of a resolved background, which is a SolidColor
// for every design this package produces.
func solidOf(source style.PaintSource) color.NRGBA {
	if solid, ok := source.(style.SolidColor); ok {
		return solid.Color
	}
	return color.NRGBA{}
}

// TestDeclarationIsHashedAndCacheSafe checks that lowering a design produces a
// declaration the cascade can hash, which is what lets the variant split avoid
// touching the cache key machinery.
func TestDeclarationIsHashedAndCacheSafe(t *testing.T) {
	active := theme.NeutralTheme()
	width, shadow := unit.Dp(1), unit.Dp(0)
	design := appearance.Design{Tone: theme.TonePrimary, Variant: theme.VariantSolid, Accent: theme.TonePrimary}
	declaration := appearance.Declaration(&active, design, width, shadow)
	if declaration.Hash64() == 0 {
		t.Error("declaration does not hash")
	}
	if declaration.Hash64() != appearance.Declaration(&active, design, width, shadow).Hash64() {
		t.Error("declaration hash is not reproducible")
	}
	other := appearance.Declaration(&active, appearance.Design{Tone: theme.ToneDanger, Variant: theme.VariantSolid, Accent: theme.TonePrimary}, width, shadow)
	if declaration.Hash64() == other.Hash64() {
		t.Error("two different designs hash identically")
	}
}

// TestDeclarationUsesTheTraitsItIsGiven pins why the two geometry tokens became
// parameters: a component has to get its own border width and shadow scale, not
// whichever component the shared package happened to name. Reading them off
// Components.Button meant every other caller silently inherited a button's ring.
func TestDeclarationUsesTheTraitsItIsGiven(t *testing.T) {
	active := theme.NeutralTheme()
	design := appearance.Design{Tone: theme.TonePrimary, Variant: theme.VariantOutline, Accent: theme.TonePrimary, UnderShadow: .5}
	base := appearance.Declaration(&active, design, 1, 4)
	if base.Hash64() == appearance.Declaration(&active, design, 3, 4).Hash64() {
		t.Error("the border width never reached the declaration")
	}
	if base.Hash64() == appearance.Declaration(&active, design, 1, 12).Hash64() {
		t.Error("the shadow scale never reached the declaration")
	}
}

// TestDeclarationTreatsAZeroShadowScaleAsOptOut pins the difference between "this
// design resolves a shadow colour" and "this component wants a resting shadow".
//
// They are separate decisions: Resolve always computes a colour for a design that
// carries an UnderShadow, while the scale token is the component's. A zero scale
// has to suppress the shadow on its own — reading only the colour would draw one
// for a component that turned it off, which is how this was found (button's rule
// table and its fluent oracle disagreed only for the themes whose UnderShadow is
// zero, and nowhere else).
func TestDeclarationTreatsAZeroShadowScaleAsOptOut(t *testing.T) {
	active := theme.NeutralTheme()
	design := appearance.Design{Tone: theme.TonePrimary, Variant: theme.VariantOutline, Accent: theme.TonePrimary, UnderShadow: .5}
	if got := appearance.Resolve(&active, design).Shadow.A; got == 0 {
		t.Fatal("fixture is degenerate: the design resolves no shadow colour")
	}

	drawn := appearance.Declaration(&active, design, 1, 4).Resolve(style.State{})
	if drawn.Paint.Shadows == nil {
		t.Error("a positive shadow scale drew no shadow")
	}
	suppressed := appearance.Declaration(&active, design, 1, 0).Resolve(style.State{})
	if suppressed.Paint.Shadows != nil {
		t.Errorf("a zero shadow scale still drew %v", suppressed.Paint.Shadows)
	}
}

// TestFieldFacesMapToTheirSurfaces pins what actually distinguishes the faces.
//
// The faces are *not* five distinct surfaces, and asserting that they are was
// wrong: the family splits by surface into three groups, and what separates the
// members of a group is the **border**, not the fill.
//
//   - Primary and Underlined both rest on the field surface — an underlined
//     field is the same box with the border collapsed to a bottom rule.
//   - Secondary and Filled both rest on the neutral control face — the tinted
//     faces, one boxed and one not.
//   - Borderless rests on nothing.
//
// The grouping is asserted because it is the real design; a face drifting into
// another group is the regression worth catching.
func TestFieldFacesMapToTheirSurfaces(t *testing.T) {
	for name, value := range coreThemes() {
		active := value
		semantic := active.Semantic()
		surfaceOf := func(face appearance.FieldVariant) color.NRGBA {
			return solidColorOf(appearance.ResolveField(&active, face, .12).Background)
		}

		if got, want := surfaceOf(appearance.FieldPrimary), semantic.FieldBackground; got != want {
			t.Errorf("%s: primary field rests on %v, want the field surface %v", name, got, want)
		}
		if got, want := surfaceOf(appearance.FieldUnderlined), semantic.FieldBackground; got != want {
			t.Errorf("%s: underlined field rests on %v, want the field surface %v", name, got, want)
		}
		// The tinted group is the neutral control face, which is by definition
		// the neutral tone's own fill.
		neutralFill := semantic.Tone(theme.ToneDefault).Fill
		if got := surfaceOf(appearance.FieldSecondary); got != neutralFill {
			t.Errorf("%s: secondary field rests on %v, want the neutral control face %v", name, got, neutralFill)
		}
		if got := surfaceOf(appearance.FieldFilled); got != neutralFill {
			t.Errorf("%s: filled field rests on %v, want the neutral control face %v", name, got, neutralFill)
		}
		if got := surfaceOf(appearance.FieldBorderless); got.A != 0 {
			t.Errorf("%s: borderless field rests on %v, want transparent", name, got)
		}
	}
}

// TestFieldShadowIsFlatUnlessTheFaceIsBoxed checks the one value the faces do not
// all share: the resting control shadow belongs to the boxed face, and a face
// without a box must not cast one.
func TestFieldShadowIsFlatUnlessTheFaceIsBoxed(t *testing.T) {
	active := theme.NeutralTheme()
	const configured = float32(.12)
	if got := appearance.ResolveField(&active, appearance.FieldPrimary, configured).ShadowOpacity; got != configured {
		t.Errorf("boxed field shadow = %v, want the configured %v", got, configured)
	}
	for _, face := range []appearance.FieldVariant{
		appearance.FieldSecondary, appearance.FieldFilled, appearance.FieldBorderless, appearance.FieldUnderlined,
	} {
		if got := appearance.ResolveField(&active, face, configured).ShadowOpacity; got != 0 {
			t.Errorf("%s casts shadow %v, want none: it has no box to cast from", face, got)
		}
	}
}

// TestToneRolesArePopulatedAndBounded guards the shared tone read.
//
// It deliberately does not compare the roles back to the TonePair they came
// from — that is what the function does, so the assertion would be vacuous. It
// checks the two properties a projection can actually get wrong: every role a
// caller may paint with is populated, and an out-of-range tone lands on the
// neutral one instead of a transparent hole.
func TestToneRolesArePopulatedAndBounded(t *testing.T) {
	for name, value := range coreThemes() {
		active := value
		for tone := range theme.ToneCount {
			colors := appearance.Tone(&active, theme.Tone(tone))
			roles := map[string]color.NRGBA{
				"Fill": colors.Fill, "OnFill": colors.OnFill,
				"Hover": colors.Hover, "Pressed": colors.Pressed,
				"Soft": colors.Soft, "OnSoft": colors.OnSoft, "SoftHover": colors.SoftHover,
			}
			for role, roleValue := range roles {
				if roleValue.A == 0 {
					t.Errorf("%s: %s.%s is transparent", name, theme.Tone(tone), role)
				}
			}
			if colors.Fill == colors.OnFill {
				t.Errorf("%s: %s paints its label in its own fill colour %v", name, theme.Tone(tone), colors.Fill)
			}
		}
		// semantic.Tone falls back to the neutral entry, so an out-of-range tone
		// must resolve to the same roles rather than to zero values.
		neutral := appearance.Tone(&active, theme.ToneDefault)
		if got := appearance.Tone(&active, theme.Tone(200)); got != neutral {
			t.Errorf("%s: an out-of-range tone resolved to %+v, want the neutral entry %+v", name, got, neutral)
		}
	}
}

// solidColorOf reads a resolved paint source as a colour; every field surface is
// a solid or a transparent fill.
func solidColorOf(source style.ColorSource) color.NRGBA {
	if source == nil {
		return color.NRGBA{}
	}
	switch value := source.(type) {
	case style.SolidColor:
		return value.Color
	default:
		return color.NRGBA{}
	}
}

// TestHoverAndPressedDiffer pins that pressing a control changes more than
// hovering it. The two states shared a colour before, which made a press look
// like a hover and left the interaction without a ladder.
func TestHoverAndPressedDiffer(t *testing.T) {
	for name, value := range coreThemes() {
		active := value
		for _, variant := range []theme.Variant{theme.VariantSolid, theme.VariantSoft, theme.VariantGhost} {
			colors := appearance.Resolve(&active, appearance.Design{
				Tone: theme.TonePrimary, Variant: variant, Accent: theme.TonePrimary,
			})
			if colors.Hover == colors.Pressed {
				t.Errorf("%s: %v hover and pressed share %v", name, variant, colors.Hover)
			}
		}
		danger := appearance.Resolve(&active, appearance.Design{
			Tone: theme.ToneDanger, Variant: theme.VariantSoft, Accent: theme.TonePrimary,
		})
		if danger.Hover == danger.Pressed {
			t.Errorf("%s: danger soft hover and pressed share %v", name, danger.Hover)
		}
	}
}
