package style

import (
	"testing"
	"time"
)

const (
	testHoverOpacity   = float32(0.90)
	testPressOpacity   = float32(0.80)
	testDisableOpacity = float32(0.38)
)

func resolvedOpacity(t *testing.T, s Style, state State) string {
	t.Helper()
	resolved := s.Resolve(state)
	if resolved.Paint == nil || resolved.Paint.Opacity == nil {
		return "nil"
	}
	return formatFloat(*resolved.Paint.Opacity)
}

func formatFloat(v float32) string {
	switch {
	case v == testDisableOpacity:
		return "disabled"
	case v == testPressOpacity:
		return "pressed"
	case v == testHoverOpacity:
		return "hovered"
	default:
		return "other"
	}
}

// TestRuleDeclarationOrderBreaksTies covers the property the old When chain had
// implicitly: with equal Rank and equal specificity, the later row wins. This
// is what makes a mechanical When -> Rule migration safe.
func TestRuleDeclarationOrderBreaksTies(t *testing.T) {
	decl := Style{}.Rules(
		Rule{When: Hovered, Set: Style{}.Opacity(testHoverOpacity)},
		Rule{When: Hovered, Set: Style{}.Opacity(testPressOpacity)},
	)
	if got := resolvedOpacity(t, decl, State{Hovered: true}); got != "pressed" {
		t.Fatalf("later row must win the tie, got %s", got)
	}
}

// TestRuleMatchedBitsBeatDeclarationOrder covers the rule the When chain could
// not express at all: a more specific row wins even when it was declared first.
func TestRuleMatchedBitsBeatDeclarationOrder(t *testing.T) {
	decl := Style{}.Rules(
		Rule{When: Checked | Disabled, Set: Style{}.Opacity(testDisableOpacity)},
		Rule{When: Checked, Set: Style{}.Opacity(testHoverOpacity)},
	)
	both := State{Checked: true, Disabled: true}
	if got := resolvedOpacity(t, decl, both); got != "disabled" {
		t.Fatalf("2-bit row must beat the 1-bit row declared after it, got %s", got)
	}
	only := State{Checked: true}
	if got := resolvedOpacity(t, decl, only); got != "hovered" {
		t.Fatalf("1-bit row must still apply alone, got %s", got)
	}
}

// TestRuleRankOrdersOrthogonalFamilies covers the case bit count cannot order.
//
// "Disabled always beats Invalid and Hovered" is a fixed precedence, but the
// two families are orthogonal: with bit-count priority alone, an
// Invalid|Hovered row (2 bits) would outrank a Disabled row (1 bit), which is
// the wrong answer. Expressing the intent through specificity would need one
// row per combination of the losing family's bits.
func TestRuleRankOrdersOrthogonalFamilies(t *testing.T) {
	// Rank at zero: the higher-specificity row wins, which is wrong here.
	unranked := Style{}.Rules(
		Rule{When: Disabled, Set: Style{}.Opacity(testDisableOpacity)},
		Rule{When: Invalid | Hovered, Set: Style{}.Opacity(testHoverOpacity)},
	)
	state := State{Disabled: true, Invalid: true, Hovered: true}
	if got := resolvedOpacity(t, unranked, state); got != "hovered" {
		t.Fatalf("without Rank the high-specificity row should win; got %s", got)
	}

	// Rank on the Disabled family makes the intent explicit and wins.
	ranked := Style{}.Rules(
		Rule{Rank: 1, When: Disabled, Set: Style{}.Opacity(testDisableOpacity)},
		Rule{When: Invalid | Hovered, Set: Style{}.Opacity(testHoverOpacity)},
	)
	if got := resolvedOpacity(t, ranked, state); got != "disabled" {
		t.Fatalf("Rank must beat matched-bit count, got %s", got)
	}
}

// TestRuleNotIsIsolated covers the Not field: a row is skipped when any of its
// Not bits is set, which is how "focus ring, but not while disabled" is said
// without a nested All/Not predicate.
func TestRuleNotIsIsolated(t *testing.T) {
	decl := Style{}.Rules(
		Rule{When: FocusVisible, Not: Disabled, Set: Style{}.Opacity(testHoverOpacity)},
	)
	if got := resolvedOpacity(t, decl, State{FocusVisible: true}); got != "hovered" {
		t.Fatalf("row should apply when no Not bit is set, got %s", got)
	}
	if got := resolvedOpacity(t, decl, State{FocusVisible: true, Disabled: true}); got != "nil" {
		t.Fatalf("row must be skipped when a Not bit is set, got %s", got)
	}
}

// TestStateMaskMappingIsOneToOne pins the design decision that State -> StateMask
// does not suppress any family. Suppression belongs in the component that
// produces State, where it can be asserted; hiding it here would make a
// disabled control silently stop reporting hover, and would put a rule into the
// conversion that no document states.
func TestStateMaskMappingIsOneToOne(t *testing.T) {
	all := State{
		Disabled: true, Hovered: true, Pressed: true, Focused: true, FocusVisible: true,
		Selected: true, Checked: true, Indeterminate: true, ReadOnly: true, Invalid: true,
		Loading: true, Open: true, Expanded: true, Dragging: true, DropTarget: true,
		Warning: true,
	}
	if got := all.Bits(); got != StateMaskAll {
		t.Fatalf("State.Bits() = %#x, want every bit %#x", got, StateMaskAll)
	}
	if round := all.Bits().Split(); round != all {
		t.Fatalf("StateMask.Split() did not round-trip:\n got %+v\nwant %+v", round, all)
	}
	// A disabled control keeps reporting hover/press, on purpose.
	disabledHover := State{Disabled: true, Hovered: true}.Bits()
	if disabledHover&Hovered == 0 {
		t.Fatal("Disabled must not clear Hovered inside StateMask")
	}
}

// TestRuleTableIsCacheSafeAndStaticallyHashable is the payoff claim of the
// design: because a rule holds no closures, a declaration built from rules can
// be hashed and cached without knowing the interaction state.
func TestRuleTableIsCacheSafeAndStaticallyHashable(t *testing.T) {
	build := func() Style {
		return Style{}.Radius(4).Rules(
			Rule{When: Hovered, Set: Style{}.Opacity(testHoverOpacity)},
			Rule{When: Checked | Indeterminate, Set: Style{}.Opacity(testPressOpacity)},
			Rule{Rank: 1, When: Disabled, Set: Style{}.Opacity(testDisableOpacity)},
		)
	}

	if build().CacheUnsafe() {
		t.Fatal("a declaration built only from Rules must be cache safe")
	}

	want := build().Hash64()
	for i := range 16 {
		if got := build().Hash64(); got != want {
			t.Fatalf("static hash drifted when the same table was rebuilt (iteration %d): %#x != %#x",
				i, got, want)
		}
	}

	// And the hash must actually distinguish tables that differ.
	other := Style{}.Radius(4).Rules(
		Rule{When: Hovered, Set: Style{}.Opacity(testPressOpacity)},
	)
	if other.Hash64() == want {
		t.Fatal("two different rule tables produced the same static hash")
	}
}

// TestRuleTablesHashApartAndStayCacheSafe pins the caching contract of rule
// rows: the key is the bits themselves, so two tables differing only in When
// never share a hash, and a declaration built purely from rules is
// cache-safe. (The predicate-condition slot this once also covered was
// removed with the retired When mechanism.)
func TestRuleTablesHashApartAndStayCacheSafe(t *testing.T) {
	checked := Style{}.Rules(Rule{When: Checked, Set: Style{}.Opacity(testPressOpacity)})
	hovered := Style{}.Rules(Rule{When: Hovered, Set: Style{}.Opacity(testPressOpacity)})
	if checked.Hash64() == hovered.Hash64() {
		t.Fatal("rule tables differing only in When must not share a hash")
	}
	if checked.CacheUnsafe() || hovered.CacheUnsafe() {
		t.Fatal("rule tables must stay cache safe")
	}
}

// The On family is the authoring form of rule rows. Its contract is the D8
// expansion table: every sugar construct expands to exactly one table shape,
// so each test below pairs a sugared declaration with the raw table a user
// could have written and asserts the two are the same declaration -- same
// static hash (the structural hash over every field, including the Set
// payloads), same rows, and same resolution across the relevant states.

// TestOnAppendsExactlyTheRawRule pins On(mask, set) == Rule{When: mask, Set: set}.
func TestOnAppendsExactlyTheRawRule(t *testing.T) {
	set := Style{}.Background(RGB(0x336699)).Scale(0.97)
	sugared := Style{}.On(Pressed, set)
	raw := Style{}.Rules(Rule{When: Pressed, Set: set})

	if sugared.Hash64() != raw.Hash64() {
		t.Fatalf("On form hashed %#x, raw form %#x", sugared.Hash64(), raw.Hash64())
	}
	rows := sugared.RuleTable()
	if len(rows) != 1 || rows[0].When != Pressed || rows[0].Not != 0 ||
		rows[0].Tween != nil || rows[0].Set.Hash64() != set.Hash64() {
		t.Fatalf("On must expand to exactly Rule{When: mask, Set: set}, got %+v", rows)
	}
}

// TestOnNotAppendsWhenAndNot pins OnNot(a, not, set) == Rule{When: a, Not: not,
// Set: set}, including a multi-bit not mask.
func TestOnNotAppendsWhenAndNot(t *testing.T) {
	set := Style{}.Opacity(0.5)
	sugared := Style{}.OnNot(FocusVisible, Selected|Disabled, set)
	raw := Style{}.Rules(Rule{When: FocusVisible, Not: Selected | Disabled, Set: set})

	if sugared.Hash64() != raw.Hash64() {
		t.Fatalf("OnNot form hashed %#x, raw form %#x", sugared.Hash64(), raw.Hash64())
	}
	if got := resolvedOpacity(t, sugared, State{FocusVisible: true}); got != "other" {
		t.Fatalf("OnNot row should apply when no Not bit is set, got %s", got)
	}
	if got := resolvedOpacity(t, sugared, State{FocusVisible: true, Disabled: true}); got != "nil" {
		t.Fatalf("OnNot row must be skipped when a Not bit is set, got %s", got)
	}
}

// TestOnAnyExpandsOneRowPerBitLSBFirst pins the disjunction expansion: one row
// per bit, ordered from the least significant bit up.
func TestOnAnyExpandsOneRowPerBitLSBFirst(t *testing.T) {
	set := Style{}.Opacity(testHoverOpacity)
	mask := Pressed | Hovered | FocusVisible // bits 2, 1, 4
	rows := Style{}.OnAny(mask, set).RuleTable()

	if len(rows) != 3 {
		t.Fatalf("OnAny must expand to one row per bit, got %d rows", len(rows))
	}
	want := []StateMask{Hovered, Pressed, FocusVisible}
	for i, row := range rows {
		if row.When != want[i] || row.Tween != nil || row.Set.Hash64() != set.Hash64() {
			t.Fatalf("row %d = {When: %#x}, want {When: %#x} with an untouched set", i, row.When, want[i])
		}
	}
}

// TestOnAnyAttachesTweenToFirstRowOnly settles the open question from the ADR:
// the tween lands on the first expanded row (the lowest bit's), not on every
// row.
func TestOnAnyAttachesTweenToFirstRowOnly(t *testing.T) {
	set := Style{}.Radius(2)
	rows := Style{}.OnAny(Hovered|Pressed, set, Quick(PropTransform, 90*time.Millisecond)).RuleTable()

	if len(rows) != 2 {
		t.Fatalf("expected two rows, got %d", len(rows))
	}
	if rows[0].Tween == nil || rows[0].Tween.Property != PropTransform || rows[0].Tween.Duration != 90*time.Millisecond {
		t.Fatalf("first row must carry the tween, got %+v", rows[0].Tween)
	}
	if rows[1].Tween != nil {
		t.Fatalf("second row must not carry the tween, got %+v", rows[1].Tween)
	}
}

// TestOnAnyEmptyMaskAppendsNothing keeps the empty mask a no-op rather than an
// accidental unconditional row; the unconditional row is On's job.
func TestOnAnyEmptyMaskAppendsNothing(t *testing.T) {
	decl := Style{}.OnAny(0, Style{}.Opacity(0.5))
	if decl.RuleCount() != 0 {
		t.Fatalf("OnAny(0) must append no rows, got %d", decl.RuleCount())
	}
}

// TestOnZeroMaskIsUnconditionalRow pins the escape hatch checkbox's resting
// row relies on: When 0 matches every state and sorts first.
func TestOnZeroMaskIsUnconditionalRow(t *testing.T) {
	decl := Style{}.On(0, Style{}.Opacity(0.5))
	rows := decl.RuleTable()
	if len(rows) != 1 || rows[0].When != 0 {
		t.Fatalf("On(0) must append one unconditional row, got %+v", rows)
	}
	for _, state := range []State{{}, {Hovered: true}, {Disabled: true, Checked: true}} {
		resolved := decl.Resolve(state)
		if resolved.Paint == nil || resolved.Paint.Opacity == nil || *resolved.Paint.Opacity != 0.5 {
			t.Fatalf("unconditional row must resolve in state %+v", state)
		}
	}
}

// TestQuickBuildsRowTween pins the tween constructor: explicit property,
// clamped non-negative duration.
func TestQuickBuildsRowTween(t *testing.T) {
	tween := Quick(PropTransform, 90*time.Millisecond)
	if tween == nil || tween.Property != PropTransform || tween.Duration != 90*time.Millisecond || tween.Delay != 0 {
		t.Fatalf("Quick built %+v", tween)
	}
	if got := Quick(PropOpacity, -time.Second).Duration; got != 0 {
		t.Fatalf("negative duration must clamp to zero, got %v", got)
	}
}

// TestOnFormsResolveLikeTheirRawTables walks every state of a five-bit mask
// and asserts the sugared chain resolves identically to the raw table. This is
// the equivalence the component migrations lean on.
func TestOnFormsResolveLikeTheirRawTables(t *testing.T) {
	set := Style{}.Background(RGB(0xff0000)).Scale(0.97)
	alt := Style{}.TextColor(RGB(0x00ff00))
	tween := Quick(PropTransform, 90*time.Millisecond)

	sugared := Style{}.Opacity(1).
		On(Hovered, set).
		OnAny(Pressed|FocusVisible, alt, tween).
		OnNot(Selected, Disabled, Style{}.Opacity(0.5))
	raw := Style{}.Opacity(1).Rules(
		Rule{When: Hovered, Set: set},
		Rule{When: Pressed, Set: alt, Tween: tween},
		Rule{When: FocusVisible, Set: alt},
		Rule{When: Selected, Not: Disabled, Set: Style{}.Opacity(0.5)},
	)

	if sugared.Hash64() != raw.Hash64() {
		t.Fatalf("sugared declaration hashed %#x, raw table %#x", sugared.Hash64(), raw.Hash64())
	}
	const states = 1 << 5 // hovered, pressed, focusVisible, selected, disabled
	for bits := range states {
		state := StateMask(bits).Split()
		a := sugared.resolve(state).hash64(nil)
		b := raw.resolve(state).hash64(nil)
		if a != b {
			t.Fatalf("state %#x: sugared resolved to %#x, raw table to %#x", bits, a, b)
		}
	}
}

// TestOnDeclarationsStayCacheSafe guards the payoff the rule table exists for:
// On-built declarations hold no closures and hash statically.
func TestOnDeclarationsStayCacheSafe(t *testing.T) {
	decl := Style{}.
		On(Hovered, Style{}.Opacity(0.5)).
		OnAny(Pressed|Checked, Style{}.Radius(2), Quick(PropTransform, time.Second)).
		OnNot(FocusVisible, Disabled, Style{}.Outline(1, 0, SolidColor{}))
	if decl.CacheUnsafe() {
		t.Fatal("a declaration built only from On rows must be cache safe")
	}
	want := decl.Hash64()
	for i := range 8 {
		if decl.Hash64() != want {
			t.Fatalf("static hash drifted on rebuild (iteration %d)", i)
		}
	}
}
