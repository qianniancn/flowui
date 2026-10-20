package resolve

import (
	"image"
	"image/color"
	"reflect"
	"testing"
	"time"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

func init() {
	// Match internal/computed's test setup: each test here builds a fresh,
	// short-lived context, so a per-context cascade cache keyed on the context
	// pointer can collide across tests when the allocator reuses an address.
	computed.DisableCascadeCache = true
}

func testGTX(now time.Time) layout.Context {
	return layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 200)},
		Ops:         new(op.Ops),
		Now:         now,
	}
}

func newTestContext() *Context {
	activeTheme := theme.NeutralTheme()
	return core.New(nil, &activeTheme, locale.LanguageAuto)
}

// scopePadding resolves base against an instance style inside a child widget,
// optionally wrapped in a Scope, and returns the resolved left padding. The child
// resolves through the same public entry point an out-of-module component would
// use, so the assertions below track the real cascade rather than a
// reimplementation of it.
func scopePadding(t *testing.T, scopeValue, base, instance style.Style, wrap bool) unit.Dp {
	t.Helper()
	ctx := newTestContext()
	var resolved style.ResolvedStyle
	child := WidgetFunc(func(c *Context, gtx layout.Context) layout.Dimensions {
		resolved = Resolve(c, gtx, "child", style.State{}, base, instance)
		return layout.Dimensions{}
	})

	if wrap {
		Scope(scopeValue, child).Layout(ctx, testGTX(time.Time{}))
	} else {
		child.Layout(ctx, testGTX(time.Time{}))
	}

	if resolved.Box == nil || resolved.Box.Padding == nil {
		t.Fatal("padding was not resolved")
	}
	return resolved.Box.Padding.Left
}

// TestScopeContract pins the rule that a scope reaches descendants that
// support the cascade, while an instance Style call still has the final say.
func TestScopeContract(t *testing.T) {
	base := style.Style{}.PaddingX(4)
	scopeValue := style.Style{}.PaddingX(12)
	instance := style.Style{}.PaddingX(20)

	unscoped := scopePadding(t, scopeValue, base, style.Style{}, false)
	if unscoped != 4 {
		t.Fatalf("unscoped padding = %v, want 4", unscoped)
	}

	scoped := scopePadding(t, scopeValue, base, style.Style{}, true)
	if scoped != 12 {
		t.Fatalf("scoped padding = %v, want 12", scoped)
	}

	overridden := scopePadding(t, scopeValue, base, instance, true)
	if overridden != 20 {
		t.Fatalf("scoped+instance padding = %v, want 20", overridden)
	}

	// Self-check: the three probes only order the layers if the middle one
	// actually differs from both neighbours. If Scope silently became a no-op,
	// the first and second reads would agree and this test would still pass.
	if scoped == unscoped || scoped == overridden {
		t.Fatalf("scope layer does not sit strictly between base and instance: %v/%v/%v",
			unscoped, scoped, overridden)
	}
}

// TestScopeRestoresAfterLayout covers the defer restore in Scope. A sibling
// resolved after the Scope widget returns must not inherit it.
func TestScopeRestoresAfterLayout(t *testing.T) {
	ctx := newTestContext()
	base := style.Style{}.PaddingX(4)

	var inside style.ResolvedStyle
	child := WidgetFunc(func(c *Context, gtx layout.Context) layout.Dimensions {
		inside = Resolve(c, gtx, "inside", style.State{}, base, style.Style{})
		return layout.Dimensions{}
	})
	Scope(style.Style{}.PaddingX(12), child).Layout(ctx, testGTX(time.Time{}))

	if inside.Box == nil || inside.Box.Padding == nil || inside.Box.Padding.Left != 12 {
		t.Fatalf("child did not see the scope: %#v", inside.Box)
	}
	if got := len(core.ActiveStyles(ctx)); got != 0 {
		t.Fatalf("scope leaked %d styles past the Scope widget", got)
	}

	after := Resolve(ctx, testGTX(time.Time{}), "after", style.State{}, base, style.Style{})
	if after.Box == nil || after.Box.Padding == nil || after.Box.Padding.Left != 4 {
		t.Fatalf("padding after the Scope = %#v, want 4", after.Box)
	}
}

// nestedScopePadding lays out a leaf under two nested scopes — outerValue wraps
// innerValue — and returns the leaf's resolved left padding plus the number of
// scope styles still active afterwards.
func nestedScopePadding(t *testing.T, outerValue, innerValue unit.Dp) (unit.Dp, int) {
	t.Helper()
	ctx := newTestContext()
	var leaf style.ResolvedStyle
	child := WidgetFunc(func(c *Context, gtx layout.Context) layout.Dimensions {
		leaf = Resolve(c, gtx, "leaf", style.State{}, style.Style{}.PaddingX(4), style.Style{})
		return layout.Dimensions{}
	})

	Scope(
		style.Style{}.PaddingX(outerValue),
		Scope(style.Style{}.PaddingX(innerValue), child),
	).Layout(ctx, testGTX(time.Time{}))

	leaked := len(core.ActiveStyles(ctx))
	if leaf.Box == nil || leaf.Box.Padding == nil {
		t.Fatal("padding was not resolved")
	}
	return leaf.Box.Padding.Left, leaked
}

// TestScopeNestedInnerWins locks the order of nested scopes. The scope closest
// to the leaf is pushed last, so it lands later in the cascade and outranks the
// one wrapping it.
func TestScopeNestedInnerWins(t *testing.T) {
	innerWins, leaked := nestedScopePadding(t, 8, 16)
	if innerWins != 16 {
		t.Fatalf("nested padding = %v, want 16 from the scope nearest the leaf", innerWins)
	}
	if leaked != 0 {
		t.Fatalf("nested scopes leaked %d styles", leaked)
	}

	// Self-check: the same two values nested the other way round must give the
	// other value, otherwise the assertion above cannot tell the two scopes apart.
	otherWay, _ := nestedScopePadding(t, 16, 8)
	if otherWay != 8 {
		t.Fatalf("swapped nesting = %v, want 8; the result does not follow the nesting order", otherWay)
	}
}

// TestScopeNilChildLeavesNoStyleBehind: a nil child returns empty dimensions and
// must not leave a scope active. It pins the observable contract — nothing is
// inherited afterwards and the value never reaches a later resolve — which is
// also what catches a guard that pushes the scope before checking for nil and
// then fails to restore it.
func TestScopeNilChildLeavesNoStyleBehind(t *testing.T) {
	ctx := newTestContext()

	if got := Scope(style.Style{}.PaddingX(12), nil).Layout(ctx, testGTX(time.Time{})); got != (layout.Dimensions{}) {
		t.Fatalf("nil child dimensions = %#v, want empty", got)
	}
	if got := len(core.ActiveStyles(ctx)); got != 0 {
		t.Fatalf("nil child pushed %d scope styles, want 0", got)
	}

	after := Resolve(ctx, testGTX(time.Time{}), "after", style.State{}, style.Style{}.PaddingX(4), style.Style{})
	if after.Box == nil || after.Box.Padding == nil || after.Box.Padding.Left != 4 {
		t.Fatalf("padding after a nil-child Scope = %#v, want 4", after.Box)
	}
}

// TestScopedKeyEmptyStaysEmpty pins the guard directly. core.FullKey is what the
// resolver used to call unconditionally; the self-check below proves it really
// does manufacture a non-empty key for input "", which is what makes the guard
// load-bearing rather than decorative.
func TestScopedKeyEmptyStaysEmpty(t *testing.T) {
	ctx := newTestContext()

	if got := scopedKey(ctx, ""); got != "" {
		t.Fatalf("unscoped empty key = %q, want \"\"", got)
	}
	if got := scopedKey(ctx, "button"); got != "button" {
		t.Fatalf("unscoped key = %q, want \"button\"", got)
	}

	restore := core.PushKey(ctx, "panel")
	defer restore()

	if got := core.FullKey(ctx, ""); got == "" {
		t.Fatalf("FullKey(ctx, %q) = %q in a key scope; the guard it motivates is untested", "", got)
	}
	if got := scopedKey(ctx, ""); got != "" {
		t.Fatalf("scoped empty key = %q, want \"\"", got)
	}
	if got, want := scopedKey(ctx, "button"), core.FullKey(ctx, "button"); got != want {
		t.Fatalf("scoped key = %q, want the qualified %q", got, want)
	}
}

// TestApplyTransitionsEmptyKeyInsideKeyScopeClaimsNoState is the behavioural half
// of the guard. An empty key documents "no identity", so inside a key scope it
// must still soft-fail. Without the guard, FullKey turned "" into "<scope>/",
// which reads as real identity: the call claimed a transition slot and shared it
// with every other keyless widget in the same scope.
func TestApplyTransitionsEmptyKeyInsideKeyScopeClaimsNoState(t *testing.T) {
	ctx := newTestContext()
	declaration := style.Style{}.
		Opacity(0.5).
		Transition(style.PropOpacity, time.Second)

	resolved := ResolveStatic(ctx, style.State{}, declaration, style.Style{})
	if !HasTransitions(resolved) {
		t.Fatal("declaration carries no transitions; the probes below cannot discriminate")
	}

	restore := core.PushKey(ctx, "panel")
	defer restore()

	out := ApplyTransitions(ctx, testGTX(time.Time{}), "", resolved)
	if got := core.StateLen(ctx); got != 0 {
		t.Fatalf("empty key inside a key scope retained %d states, want 0", got)
	}
	if out.Paint == nil || out.Paint.Opacity == nil || *out.Paint.Opacity != 0.5 {
		t.Fatalf("empty key should snap to the resolved target, got %#v", out.Paint)
	}

	// Positive control: the same resolve with a real key does claim a slot, so a
	// zero above means the guard fired rather than transitions being ignored.
	ApplyTransitions(ctx, testGTX(time.Time{}), "widget", resolved)
	if got := core.StateLen(ctx); got == 0 {
		t.Fatal("a keyed call claimed no state; the empty-key assertion cannot discriminate")
	}
}

// TestStaticPairMatchesOneCallResolve checks the equivalence the package doc
// promises: for the same key, ResolveStatic plus ApplyTransitions must animate
// exactly like Resolve. A component that resolves during measurement uses the
// static pair, so a divergence would make it animate differently from one that
// does not.
func TestStaticPairMatchesOneCallResolve(t *testing.T) {
	declaration := style.Style{}.
		Opacity(1).
		Transition(style.PropOpacity, 200*time.Millisecond).
		On(style.Hovered, style.Style{}.Opacity(0.25))

	start := time.Unix(1, 0)
	resting := style.State{}
	hovered := style.State{Hovered: true}
	midpoint := start.Add(100 * time.Millisecond)

	opacity := func(value style.ResolvedStyle) float32 {
		if value.Paint == nil || value.Paint.Opacity == nil {
			t.Fatal("opacity was not resolved")
		}
		return *value.Paint.Opacity
	}

	oneCall := newTestContext()
	Resolve(oneCall, testGTX(start), "widget", resting, declaration, style.Style{})
	Resolve(oneCall, testGTX(start), "widget", hovered, declaration, style.Style{})
	want := opacity(Resolve(oneCall, testGTX(midpoint), "widget", hovered, declaration, style.Style{}))

	staticPair := newTestContext()
	ApplyTransitions(staticPair, testGTX(start), "widget", ResolveStatic(staticPair, resting, declaration, style.Style{}))
	ApplyTransitions(staticPair, testGTX(start), "widget", ResolveStatic(staticPair, hovered, declaration, style.Style{}))
	got := opacity(ApplyTransitions(staticPair, testGTX(midpoint), "widget", ResolveStatic(staticPair, hovered, declaration, style.Style{})))

	if got != want {
		t.Fatalf("static pair opacity = %v, one call = %v", got, want)
	}
	// Self-check: at the midpoint the animation must sit strictly between the
	// two endpoints, otherwise the comparison above would also pass for a
	// declaration that ignored its transition entirely.
	if want <= 0.25 || want >= 1 {
		t.Fatalf("midpoint opacity = %v, want strictly between 0.25 and 1", want)
	}
}

// TestColorResolvesConcreteSources pins the value path: every ColorSource is
// concrete since the theme token types were retired, so Color resolves without
// a context.
func TestColorResolvesConcreteSources(t *testing.T) {
	literal := style.RGB(0x112233)
	if got, ok := Color(literal); !ok || got != literal.Color {
		t.Fatalf("Color(literal) = %v/%v, want %v/true", got, ok, literal.Color)
	}
}

// TestOpacityHelpersForward covers the nil branches callers rely on to skip work,
// and checks both helpers forward to the engine rather than re-deriving.
func TestOpacityHelpersForward(t *testing.T) {
	base := style.RGB(0x808080)
	half := float32(0.5)

	if got := ApplyOpacity(base.Color, nil); got != base.Color {
		t.Fatalf("ApplyOpacity(nil) = %v, want the colour unchanged", got)
	}
	if got, want := ApplyOpacity(base.Color, &half), computed.ApplyOpacity(base.Color, &half); got != want {
		t.Fatalf("ApplyOpacity = %v, want the engine's %v", got, want)
	}
	// Self-check: the comparison above only discriminates if scaling moved alpha.
	if got := ApplyOpacity(base.Color, &half); got.A == base.Color.A {
		t.Fatalf("ApplyOpacity left alpha at %d; the comparison cannot discriminate", got.A)
	}

	// An outline-less style has nothing to scale and must come back untouched,
	// so callers can chain the call unconditionally.
	plain := style.ResolvedStyle{Paint: &style.PaintStyle{Opacity: &half}}
	if got := ApplyOutlineOpacity(plain, 0.25); !reflect.DeepEqual(got, plain) {
		t.Fatalf("ApplyOutlineOpacity changed an outline-less style: %#v", got.Paint)
	}

	outlined := style.ResolvedStyle{Paint: &style.PaintStyle{
		Outline: &style.OutlineStyle{Color: style.SolidColor{Color: color.NRGBA{A: 200}}},
	}}
	got := ApplyOutlineOpacity(outlined, 0.5)
	if want := computed.ApplyOutlineOpacity(outlined, 0.5); !reflect.DeepEqual(got, want) {
		t.Fatalf("ApplyOutlineOpacity = %#v, want the engine's %#v", got.Paint.Outline, want.Paint.Outline)
	}
	// Self-check: the scaling must actually happen and the input must stay
	// intact, otherwise "equal to the engine" is satisfied by a shared no-op.
	if col, ok := Color(got.Paint.Outline.Color); !ok || col.A != 100 {
		t.Fatalf("scaled outline alpha = %v/%v, want 100", col, ok)
	}
	if col, _ := Color(outlined.Paint.Outline.Color); col.A != 200 {
		t.Fatalf("input outline was mutated in place: alpha %d, want 200", col.A)
	}
}
