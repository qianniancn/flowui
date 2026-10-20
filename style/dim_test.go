package style

import (
	"image/color"
	"math"
	"reflect"
	"testing"
)

// Dim is the primitive that lets a rule table express "scale whatever the base
// said", so these tests are about the two things that make it usable at all:
// it must dim the merged declaration rather than only the row that asked for it,
// and it must leave shadow layers alone.

func disabledState() State {
	return Disabled.Split()
}

func solid(alpha uint8) SolidColor {
	return SolidColor{Color: color.NRGBA{R: 10, G: 20, B: 30, A: alpha}}
}

func alphaOf(t *testing.T, source ColorSource) uint8 {
	t.Helper()
	switch value := source.(type) {
	case SolidColor:
		return value.Color.A
	case *SolidColor:
		if value == nil {
			t.Fatal("nil solid color")
		}
		return value.Color.A
	default:
		t.Fatalf("expected a solid color, got %T", source)
		return 0
	}
}

func colorOf(t *testing.T, source PaintSource) uint8 {
	t.Helper()
	if source == nil {
		t.Fatal("nil paint source")
	}
	return alphaOf(t, source.(ColorSource))
}

// TestDimScalesEveryColorButNotShadows pins the field list, because "which
// fields does it touch" is the whole contract.
func TestDimScalesEveryColorButNotShadows(t *testing.T) {
	declaration := Style{}.
		Background(solid(200)).
		BorderWidth(2).
		BorderColor(solid(160)).
		Outline(1, 0, solid(40)).
		TextColor(solid(100)).
		BoxShadow(0, 1, 0, 1, solid(80)).
		Dim(0.5)

	resolved := declaration.Resolve(State{})

	if got := colorOf(t, resolved.Paint.Background); got != 100 {
		t.Fatalf("background alpha = %d, want 100", got)
	}
	if got := alphaOf(t, resolved.Paint.Border.Color); got != 80 {
		t.Fatalf("border alpha = %d, want 80", got)
	}
	if got := alphaOf(t, resolved.Paint.Outline.Color); got != 20 {
		t.Fatalf("outline alpha = %d, want 20", got)
	}
	if got := alphaOf(t, resolved.Text.Color); got != 50 {
		t.Fatalf("text alpha = %d, want 50", got)
	}
	// The one exclusion. A shadow whose alpha derives from a color has to be
	// restated by whoever wrote it, because the built-in components disagree
	// about whether it should dim.
	if len(resolved.Paint.Shadows) != 1 {
		t.Fatalf("expected 1 shadow, got %d", len(resolved.Paint.Shadows))
	}
	if got := alphaOf(t, resolved.Paint.Shadows[0].Color); got != 80 {
		t.Fatalf("shadow alpha = %d, want it untouched at 80", got)
	}
}

// TestDimDimsTheMergedDeclaration is the reason Dim is a resolution-boundary
// step rather than a row-local transform. The row only says "dim"; the colors it
// has to dim were declared by the base.
func TestDimDimsTheMergedDeclaration(t *testing.T) {
	declaration := Style{}.
		Background(solid(200)).
		TextColor(solid(100)).
		Rules(Rule{
			When: Disabled,
			Set:  Style{}.Dim(0.5),
		})

	enabled := declaration.Resolve(State{})
	if got := colorOf(t, enabled.Paint.Background); got != 200 {
		t.Fatalf("enabled background alpha = %d, want 200", got)
	}

	disabled := declaration.Resolve(disabledState())
	if got := colorOf(t, disabled.Paint.Background); got != 100 {
		t.Fatalf("disabled background alpha = %d, want 100: the row must dim the base, not only its own fields", got)
	}
	if got := alphaOf(t, disabled.Text.Color); got != 50 {
		t.Fatalf("disabled text alpha = %d, want 50", got)
	}
}

// A later layer's Dim replaces rather than multiplies: "dim by a half" should
// mean a half, not a half of whatever the layer below asked for.
func TestDimIsLastLayerWins(t *testing.T) {
	declaration := Join(
		Style{}.Background(solid(200)).Dim(0.9),
		Style{}.Dim(0.5),
	)
	if got := colorOf(t, declaration.Resolve(State{}).Paint.Background); got != 100 {
		t.Fatalf("background alpha = %d, want 100 (0.5, not 0.9*0.5)", got)
	}
}

func TestDimClampsAndIgnoresNonFiniteFactors(t *testing.T) {
	if got := colorOf(t, Style{}.Background(solid(200)).Dim(2).Resolve(State{}).Paint.Background); got != 200 {
		t.Fatalf("alpha = %d, want the factor clamped to 1", got)
	}
	if got := colorOf(t, Style{}.Background(solid(200)).Dim(-1).Resolve(State{}).Paint.Background); got != 0 {
		t.Fatalf("alpha = %d, want the factor clamped to 0", got)
	}
	// A non-finite factor is dropped rather than coerced to zero, matching how
	// every other numeric setter in this package treats NaN.
	source := color.NRGBA{R: 10, G: 20, B: 30, A: 200}
	resolved := Style{}.Background(SolidColor{Color: source}).Dim(float32(math.NaN())).Resolve(State{})
	if got := colorOf(t, resolved.Paint.Background); got != 200 {
		t.Fatalf("alpha = %d, want 200: a NaN factor must leave the declaration alone", got)
	}
}

// Dim is a scalar field, so it must reach the cache key. Two declarations that
// differ only in their factor must not share an entry.
func TestDimIsHashableAndCacheSafe(t *testing.T) {
	base := Style{}.Background(solid(200))
	half := base.Dim(0.5)
	quarter := base.Dim(0.25)

	if half.Hash64() == quarter.Hash64() {
		t.Fatal("dim factor must be part of the cache key")
	}
	if half.Hash64() != half.Dim(0.5).Hash64() {
		t.Fatal("static hash must be reproducible")
	}
	if half.CacheUnsafe() {
		t.Fatal("a scalar factor is not a captured closure; Dim must stay cache safe")
	}
	if base.Hash64() == half.Hash64() {
		t.Fatal("dimmed and undimmed declarations must hash differently")
	}
}

// The merge fast path drops a layer it considers empty. A layer carrying only a
// Dim is not empty, and treating it as such would silently un-dim it.
func TestDimSurvivesMergeFastPath(t *testing.T) {
	declaration := Join(Style{}.Background(solid(200)))
	withDim := merge(declaration, Style{}.Dim(0.5))
	if withDim.dim == nil {
		t.Fatal("a layer carrying only Dim was treated as empty")
	}
	if got := colorOf(t, withDim.Resolve(State{}).Paint.Background); got != 100 {
		t.Fatalf("alpha = %d, want 100", got)
	}
}

// Texture: a gradient's stops are colors too, and a partially applied Dim would
// be worse than none.
func TestDimScalesGradientStops(t *testing.T) {
	declaration := Style{}.
		Background(LinearGradient(ColorStop(0, solid(200)), ColorStop(1, solid(100)))).
		Dim(0.5)

	resolved := declaration.Resolve(State{})
	gradient, ok := resolved.Paint.Background.(Gradient)
	if !ok {
		t.Fatalf("background = %T, want a Gradient", resolved.Paint.Background)
	}
	want := []uint8{100, 50}
	for index, stop := range gradient.Stops {
		if got := alphaOf(t, stop.Color); got != want[index] {
			t.Fatalf("stop %d alpha = %d, want %d", index, got, want[index])
		}
	}
}

// The dimmed result must be a plain solid color, not a wrapper: the built-in
// disabled colors are plain solids, and Dim has to be indistinguishable
// from them for the equivalence tests to mean anything.
func TestDimOnSolidStaysSolid(t *testing.T) {
	resolved := Style{}.Background(solid(200)).Dim(0.5).Resolve(State{})
	want := SolidColor{Color: color.NRGBA{R: 10, G: 20, B: 30, A: 100}}
	if !reflect.DeepEqual(resolved.Paint.Background, want) {
		t.Fatalf("background = %#v, want %#v", resolved.Paint.Background, want)
	}
}
