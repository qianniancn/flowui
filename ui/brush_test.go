package ui

import (
	"image/color"
	"testing"

	"github.com/qianniancn/flowui/internal/frame"
)

func TestResolveColorUsesActiveTheme(t *testing.T) {
	activeTheme := DefaultTheme()
	activeTheme.Palette.Accent = color.NRGBA{R: 0x12, G: 0x34, B: 0x56, A: 0xff}
	ctx := frame.New(nil, &activeTheme, LanguageEnglish)

	resolved, ok := ResolveColor(ctx, WithAlpha(TokenAccent, .5))
	if !ok {
		t.Fatal("ResolveColor rejected a theme color")
	}
	if want := (color.NRGBA{R: 0x12, G: 0x34, B: 0x56, A: 0x80}); resolved != want {
		t.Fatalf("resolved color = %#v, want %#v", resolved, want)
	}
	if _, ok := ResolveColor(ctx, nil); ok {
		t.Fatal("ResolveColor accepted nil")
	}
	if got, ok := ResolveColor(nil, TokenAccent); !ok || got != DefaultTheme().Palette.Accent {
		t.Fatalf("nil-context color = %#v, %v", got, ok)
	}
}

func TestResolveBrushResolvesGradientStopsWithoutMutatingSource(t *testing.T) {
	activeTheme := DefaultTheme()
	activeTheme.Palette.Accent = color.NRGBA{R: 0x10, G: 0x20, B: 0x30, A: 0xff}
	activeTheme.Palette.Danger = color.NRGBA{R: 0xa0, G: 0xb0, B: 0xc0, A: 0xff}
	ctx := frame.New(nil, &activeTheme, LanguageEnglish)
	source := LinearGradient(
		ColorStop(0, TokenAccent),
		ColorStop(1, TokenDanger),
	)

	brush, ok := ResolveBrush(ctx, source)
	if !ok {
		t.Fatal("ResolveBrush rejected a gradient")
	}
	if got := brush.ColorAt(0); got != activeTheme.Palette.Accent {
		t.Fatalf("first color = %#v, want %#v", got, activeTheme.Palette.Accent)
	}
	if got := brush.ColorAt(1); got != activeTheme.Palette.Danger {
		t.Fatalf("last color = %#v, want %#v", got, activeTheme.Palette.Danger)
	}
	if source.Stops[0].Color != TokenAccent || source.Stops[1].Color != TokenDanger {
		t.Fatal("ResolveBrush mutated the source gradient")
	}
}
