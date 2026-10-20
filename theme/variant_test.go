package theme

import "testing"

// TestVariantVocabularyIsComplete pins the variant vocabulary: every value
// below VariantCount has a distinct name, and an out-of-range variant reports
// itself as unknown rather than borrowing a real name.
func TestVariantVocabularyIsComplete(t *testing.T) {
	seen := make(map[string]Variant, VariantCount)
	for value := range VariantCount {
		variant := Variant(value)
		name := variant.String()
		if name == "" || name == "variant(?)" {
			t.Errorf("Variant(%d).String() = %q, want a real name", value, name)
			continue
		}
		if previous, duplicate := seen[name]; duplicate {
			t.Errorf("Variant(%d) and Variant(%d) share the name %q", previous, value, name)
			continue
		}
		seen[name] = variant
	}
	if got := Variant(VariantCount).String(); got != "variant(?)" {
		t.Errorf("out-of-range variant String() = %q, want %q", got, "variant(?)")
	}
}

// TestVariantSetAllCoversTheWholeVocabulary checks the constant against the
// vocabulary size rather than against itself: adding a variant without widening
// VariantSetAll shows up here as an uncovered bit.
func TestVariantSetAllCoversTheWholeVocabulary(t *testing.T) {
	for value := range VariantCount {
		variant := Variant(value)
		if !VariantSetAll.Has(variant) {
			t.Errorf("VariantSetAll.Has(%s) = false, want true for every defined variant", variant)
		}
	}
	if bits := uint(VariantSetAll); bits != 1<<VariantCount-1 {
		t.Errorf("VariantSetAll bits = %#b, want %d low bits set", bits, VariantCount)
	}
}

// TestZeroVariantSetMeansEverything is the contract that keeps existing
// components working: a component that has not declared a subset renders every
// variant, so the split cannot silently remove behaviour from one that predates
// it.
func TestZeroVariantSetMeansEverything(t *testing.T) {
	var set VariantSet
	for value := range VariantCount {
		variant := Variant(value)
		if !set.Has(variant) {
			t.Errorf("zero VariantSet refuses %s; the zero value must mean all", variant)
		}
		if got := set.Resolve(variant); got != variant {
			t.Errorf("zero VariantSet.Resolve(%s) = %s, want %s", variant, got, variant)
		}
	}
}

// TestVariantSetResolveIsTotalAndSupported is the postcondition the fallback
// machinery has to satisfy for every possible declaration: Resolve always
// returns a variant the set supports, so a caller never has to re-check.
//
// The iteration covers every subset of a five-variant manifest rather than a
// hand-picked few, because the interesting failures are the chains that skip
// more than one step (link -> text -> ghost -> solid).
func TestVariantSetResolveIsTotalAndSupported(t *testing.T) {
	manifest := []Variant{VariantSolid, VariantSoft, VariantOutline, VariantGhost, VariantText}
	for mask := range 1 << len(manifest) {
		var set VariantSet
		for bit, variant := range manifest {
			if mask&(1<<bit) != 0 {
				set |= 1 << variant
			}
		}
		for value := range VariantCount {
			variant := Variant(value)
			resolved := set.Resolve(variant)
			if int(resolved) >= VariantCount {
				t.Fatalf("set %#b: Resolve(%s) = %s, outside the vocabulary", uint(set), variant, resolved)
			}
			if set.Has(variant) && resolved != variant {
				t.Fatalf("set %#b: Resolve(%s) = %s, but the set already supports it", uint(set), variant, resolved)
			}
			// A declared subset must never be answered with something outside
			// it — that is the property the lowest-set-bit fallback exists for.
			if set != 0 && !set.Has(resolved) {
				t.Fatalf("set %#b: Resolve(%s) = %s, which the set does not support", uint(set), variant, resolved)
			}
		}
	}
}

// TestVariantFallbacksStayInsideTheVocabulary keeps the fallback table honest:
// every candidate must be a real variant, and no variant may fall back to
// itself (which would make Resolve fail to converge).
func TestVariantFallbacksStayInsideTheVocabulary(t *testing.T) {
	for variant, candidates := range variantFallbacks {
		if int(variant) >= VariantCount {
			t.Errorf("fallback table has an entry for out-of-range %s", variant)
		}
		for _, candidate := range candidates {
			if int(candidate) >= VariantCount {
				t.Errorf("%s falls back to out-of-range %s", variant, candidate)
			}
			if candidate == variant {
				t.Errorf("%s falls back to itself", variant)
			}
		}
	}
}

// TestVariantSetResolveFallsBackAlongTheChain pins the actual fallback
// behaviour against a deliberately partial manifest, so the chains are asserted
// rather than merely exercised.
func TestVariantSetResolveFallsBackAlongTheChain(t *testing.T) {
	// Soft, outline and text, but no link, dashed, ghost or underlined.
	set := VariantSet(1<<VariantSoft | 1<<VariantOutline | 1<<VariantText)
	cases := []struct {
		request Variant
		want    Variant
		why     string
	}{
		{VariantSolid, VariantSoft, "no solid declared, so the nearest fill is soft"},
		{VariantSoft, VariantSoft, "declared variants are returned unchanged"},
		{VariantLink, VariantText, "a link degrades to plain text before a fill"},
		{VariantDashed, VariantOutline, "a dashed outline degrades to a plain outline, not a fill"},
		{VariantUnderlined, VariantOutline, "an underlined field degrades to an outlined one"},
		{VariantGhost, VariantText, "an unfilled surface degrades to text"},
		{VariantFilled, VariantSoft, "a filled face degrades to the soft tint"},
	}
	for _, testCase := range cases {
		if got := set.Resolve(testCase.request); got != testCase.want {
			t.Errorf("Resolve(%s) = %s, want %s (%s)", testCase.request, got, testCase.want, testCase.why)
		}
	}
}

// TestVariantSetAlwaysContainsARenderableVariant pins the invariant Resolve's
// last step relies on: every non-zero VariantSet has at least one valid bit,
// because the invalid bits are all above VariantCount and there are only
// 16-VariantCount of them.
//
// If someone widens the vocabulary past 16 variants this fails, which is the
// point — the alternative is a Resolve that silently starts returning garbage
// for sets built out of the high bits.
func TestVariantSetAlwaysContainsARenderableVariant(t *testing.T) {
	validMask := uint16(VariantSetAll)
	for mask := uint16(1); mask != 0; mask++ {
		if mask&validMask == 0 {
			continue
		}
		set := VariantSet(mask)
		found := false
		for value := range VariantCount {
			if set.Has(Variant(value)) {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("set %#016b has no valid variant bit, but Resolve assumes one exists", mask)
		}
	}
	if validMask == 0 {
		t.Fatal("VariantSetAll is empty")
	}
}

// TestVariantSetResolveFallsBackToSolidWhenNothingMatches covers the terminal
// case: a set built only from bits outside the vocabulary declares no renderable
// variant, and Resolve still has to terminate on something.
func TestVariantSetResolveFallsBackToSolidWhenNothingMatches(t *testing.T) {
	set := VariantSet(1 << 15)
	for _, variant := range []Variant{VariantSoft, VariantOutline, VariantGhost, VariantLink, VariantDashed} {
		if got := set.Resolve(variant); got != VariantSolid {
			t.Errorf("Resolve(%s) with no match = %s, want solid", variant, got)
		}
	}
}

// TestVariantSetResolveNeverEscapesADeclaredSubset is a regression guard for
// the fallback that used to be an unconditional VariantSolid.
//
// A component whose vocabulary excludes Solid — a text-only link style, say —
// must not be handed Solid back for an unsupported request: that is a variant
// it never declared. Resolve has to land inside the declared set instead.
func TestVariantSetResolveNeverEscapesADeclaredSubset(t *testing.T) {
	textOnly := VariantSet(1 << VariantText)
	for value := range VariantCount {
		variant := Variant(value)
		if variant == VariantText {
			continue
		}
		if got := textOnly.Resolve(variant); got != VariantText {
			t.Errorf("text-only set: Resolve(%s) = %s, want text", variant, got)
		}
	}
	if got := textOnly.Resolve(VariantText); got != VariantText {
		t.Errorf("text-only set: Resolve(text) = %s, want text", got)
	}
}

// TestVariantSetResolvePrefersTheDeclaredChainOverTheLowestBit is the ordering
// rule between the two fallback steps: a variant's own chain wins, and the
// lowest declared variant is only the last resort.
func TestVariantSetResolvePrefersTheDeclaredChainOverTheLowestBit(t *testing.T) {
	// Solid is the lowest bit, so an unconditional lowest-bit fallback would
	// answer every request with it; the chain must beat it.
	set := VariantSet(1<<VariantSolid | 1<<VariantOutline | 1<<VariantText)
	if got := set.Resolve(VariantLink); got != VariantText {
		t.Errorf("Resolve(link) = %s, want text from the chain, not the lowest declared variant", got)
	}
	if got := set.Resolve(VariantDashed); got != VariantOutline {
		t.Errorf("Resolve(dashed) = %s, want outline from the chain", got)
	}
}

// TestToneIsHueSplitsTheVocabularyInTwo pins the contract that ToneHover is a
// neutral interaction surface rather than a hue: the two families are derived
// differently (a hue tone comes off a ramp, a neutral one is constructed by the
// style package), so the predicate every consumer branches on has to be exact.
func TestToneIsHueSplitsTheVocabularyInTwo(t *testing.T) {
	hue := []Tone{TonePrimary, ToneSecondary, ToneSuccess, ToneWarning, ToneDanger}
	neutral := []Tone{ToneDefault, ToneHover}
	for _, tone := range hue {
		if !ToneIsHue(tone) {
			t.Errorf("ToneIsHue(%s) = false, want true: it draws from a hue ramp", tone)
		}
	}
	for _, tone := range neutral {
		if ToneIsHue(tone) {
			t.Errorf("ToneIsHue(%s) = true, want false: it is a neutral surface role", tone)
		}
		if int(tone) >= ToneCount {
			t.Errorf("%s is outside the vocabulary", tone)
		}
	}
	if len(hue)+len(neutral) != ToneCount {
		t.Errorf("classified %d tones, but ToneCount is %d — a tone has no family",
			len(hue)+len(neutral), ToneCount)
	}
	// Out-of-range reports false so the predicate agrees with Semantic.Tone's
	// fallback to ToneDefault.
	if ToneIsHue(Tone(200)) {
		t.Error("ToneIsHue(200) = true, want false for an out-of-range tone")
	}
}

// TestSizeVocabularyIsComplete mirrors the other two vocabularies, and pins the
// zero value to SizeSmall so an unset size cannot silently mean "large".
func TestSizeVocabularyIsComplete(t *testing.T) {
	if SizeSmall != 0 {
		t.Errorf("SizeSmall = %d, want 0: an unset size must read as the smallest control", SizeSmall)
	}
	seen := make(map[string]Size, 3)
	for _, size := range []Size{SizeSmall, SizeMedium, SizeLarge} {
		name := size.String()
		if name == "" || name == "size(?)" {
			t.Errorf("Size(%d).String() = %q, want a real name", size, name)
			continue
		}
		if _, duplicate := seen[name]; duplicate {
			t.Errorf("duplicate size name %q", name)
			continue
		}
		seen[name] = size
	}
	if got := Size(200).String(); got != "size(?)" {
		t.Errorf("out-of-range size String() = %q, want %q", got, "size(?)")
	}
}
