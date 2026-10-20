package style_test

import (
	"image/color"
	"testing"
	"time"

	giolayout "gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/draw"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/style/resolve"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/ui"
)

// TestStyleResolversUsePublicContextAndWidgetContracts pins the whole public
// resolution door. style/resolve is the only path out-of-module code has: the
// engine (internal/computed) is unreachable, and style itself is a pure leaf
// that cannot import it. So the door must expose both call shapes —
//
//   - the one-call pair, Resolve / ResolvePart, which applies transitions; and
//   - the static pair plus HasTransitions / ApplyTransitions, which is what a
//     component that also resolves during measurement needs in order to apply
//     transition state exactly once per frame;
//
// — plus the value helpers the same code needs to read what it resolved.
func TestStyleResolversUsePublicContextAndWidgetContracts(t *testing.T) {
	var _ func(*ui.Context, giolayout.Context, string, style.State, style.Style, style.Style) style.ResolvedStyle = resolve.Resolve
	var _ func(*ui.Context, style.State, style.Style, style.Style) style.ResolvedStyle = resolve.ResolveStatic
	var _ func(*ui.Context, giolayout.Context, string, style.Part, style.State, style.Style, style.Style) style.ResolvedStyle = resolve.ResolvePart
	var _ func(*ui.Context, style.Part, style.State, style.Style, style.Style) style.ResolvedStyle = resolve.ResolvePartStatic
	var _ func(style.ResolvedStyle) style.ResolvedStyle = resolve.CloneResolvedStyle
	var _ func(...style.ResolvedStyle) bool = resolve.HasTransitions
	var _ func(*ui.Context, giolayout.Context, string, style.ResolvedStyle) style.ResolvedStyle = resolve.ApplyTransitions
	var _ func(*ui.Context, giolayout.Context, string, style.Part, style.ResolvedStyle) style.ResolvedStyle = resolve.ApplyPartTransitions
	var _ func(style.ColorSource) (color.NRGBA, bool) = resolve.Color
	var _ func(color.NRGBA, *float32) color.NRGBA = resolve.ApplyOpacity
	var _ func(style.ResolvedStyle, float32) style.ResolvedStyle = resolve.ApplyOutlineOpacity
	var _ func(style.PaintSource) (draw.Brush, bool) = resolve.Brush
	var _ func(style.Style, ui.Widget) ui.Widget = resolve.Scope
}

func TestStyleFluentContractResolvesPublicProperties(t *testing.T) {
	base := style.New()
	if resolved := style.New().Resolve(style.State{}); resolved.Box != nil || resolved.Paint != nil || resolved.Text != nil {
		t.Fatalf("new style is not empty: %#v", resolved)
	}
	configured := base.
		Width(unit.Dp(240)).
		Padding(unit.Dp(8)).
		Background(style.RGB(0x102030)).
		BorderWidth(unit.Dp(1)).
		TextColor(style.RGB(0xf0f0f0)).
		FontSize(unit.Sp(14)).
		Translate(unit.Dp(2), unit.Dp(3)).
		Transition(style.PropOpacity, time.Millisecond*120, style.TransitionEase(func(progress float32) float32 {
			return progress
		}))

	resolved := configured.Resolve(style.State{})
	if resolved.Box == nil || resolved.Box.Width == nil || *resolved.Box.Width != unit.Dp(240) {
		t.Fatalf("resolved width = %#v, want 240dp", resolved.Box)
	}
	if resolved.Box.Padding == nil || resolved.Box.Padding.Top != unit.Dp(8) {
		t.Fatalf("resolved padding = %#v, want 8dp", resolved.Box.Padding)
	}
	if resolved.Paint == nil || resolved.Paint.Border == nil || resolved.Paint.Border.Width == nil {
		t.Fatal("resolved border is missing")
	}
	if resolved.Text == nil || resolved.Text.FontSize == nil || *resolved.Text.FontSize != unit.Sp(14) {
		t.Fatal("resolved text size is missing")
	}
	if len(resolved.Transitions) != 1 || resolved.Transitions[0].Property != style.PropOpacity {
		t.Fatalf("resolved transitions = %#v, want one opacity transition", resolved.Transitions)
	}

	// Style values are immutable declarations: editing a copy does not mutate
	// the original zero value.
	if original := base.Resolve(style.State{}); original.Box != nil {
		t.Fatal("style fluent call mutated the original value")
	}
}

func TestStyleCompositionHelpersPreserveDeclarations(t *testing.T) {
	joined := style.Join(
		(style.Style{}).Width(unit.Dp(120)),
		(style.Style{}).Height(unit.Dp(32)),
	)
	resolved := joined.Resolve(style.State{})
	if resolved.Box == nil || resolved.Box.Width == nil || resolved.Box.Height == nil {
		t.Fatal("joined style did not preserve box declarations")
	}
	if *resolved.Box.Width != unit.Dp(120) || *resolved.Box.Height != unit.Dp(32) {
		t.Fatalf("joined dimensions = %v x %v", *resolved.Box.Width, *resolved.Box.Height)
	}

	fontSize := unit.Sp(13)
	text := style.TextDeclaration(&style.TextStyle{FontSize: &fontSize})
	textResolved := text.Resolve(style.State{})
	if textResolved.Text == nil || textResolved.Text.FontSize == nil || *textResolved.Text.FontSize != fontSize {
		t.Fatal("text declaration did not preserve text properties")
	}
}

func TestStyleRulesThemeAndColorSources(t *testing.T) {
	conditional := style.Style{}.
		On(style.Hovered, style.Style{}.Opacity(0.5)).
		Part(style.PartContent, style.Style{}.TextColor(style.RGB(0xff0000)))
	resolved := conditional.Resolve(style.State{Hovered: true})
	if resolved.Paint == nil || resolved.Paint.Opacity == nil || *resolved.Paint.Opacity != 0.5 {
		t.Fatal("hover rule did not resolve")
	}
	part := style.CascadePart(style.State{}, style.PartContent, conditional)
	if part.Text == nil || part.Text.Color == nil {
		t.Fatal("content part did not resolve")
	}

	light := theme.NeutralTheme()
	dark := theme.NeutralDarkTheme()
	if light.Palette.Background == dark.Palette.Background {
		t.Fatal("default and dark themes unexpectedly share a background")
	}

	resolvedColor, ok := resolve.Color(style.Color(color.NRGBA{R: 1, G: 2, B: 3, A: 255}))
	if !ok || resolvedColor != (color.NRGBA{R: 1, G: 2, B: 3, A: 255}) {
		t.Fatalf("resolved color = %#v, ok=%v", resolvedColor, ok)
	}

	// Theme values are plain fields a declaration reads directly; switching
	// themes means rebuilding declarations with the new values.
	if accent := style.Solid(light.Palette.Accent); accent.Color != light.Palette.Accent {
		t.Fatalf("Solid did not capture the palette value: %#v", accent)
	}
}
