package field

import (
	"image/color"
	"testing"

	"gioui.org/unit"
	"github.com/qianniancn/FlowUI/internal/theme"
)

func TestResolveStylePrimary(t *testing.T) {
	value := ResolveStyle(defaultTheme(), Primary, false, false, false, false)
	if value.Background != (color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}) {
		t.Fatalf("primary background = %v, want white", value.Background)
	}
	if value.Border.A != 0 || value.BorderWidth != 0 {
		t.Fatalf("primary border = %v at %v, want none", value.Border, value.BorderWidth)
	}
	if value.ShadowOpacity != 1 {
		t.Fatalf("primary shadow opacity = %v, want 1", value.ShadowOpacity)
	}
}

func TestResolveStyleSecondary(t *testing.T) {
	value := ResolveStyle(defaultTheme(), Secondary, false, false, false, false)
	if value.Background != (color.NRGBA{R: 0xeb, G: 0xeb, B: 0xec, A: 0xff}) {
		t.Fatalf("secondary background = %v, want default", value.Background)
	}
	if value.ShadowOpacity != 0 {
		t.Fatalf("secondary shadow opacity = %v, want 0", value.ShadowOpacity)
	}
}

func TestResolveStyleHover(t *testing.T) {
	value := ResolveStyle(defaultTheme(), Primary, true, false, false, false)
	if value.Background != (color.NRGBA{R: 0xf8, G: 0xf8, B: 0xf9, A: 0xff}) {
		t.Fatalf("hover background = %v, want field hover", value.Background)
	}
	if value.Border.A != 0 || value.BorderWidth != 0 {
		t.Fatalf("hover border = %v at %v, want none", value.Border, value.BorderWidth)
	}
}

func TestResolveStyleUsesDarkFieldTokens(t *testing.T) {
	activeTheme := theme.DarkTheme()
	primary := ResolveStyle(&activeTheme, Primary, false, false, false, false)
	hovered := ResolveStyle(&activeTheme, Primary, true, false, false, false)
	secondary := ResolveStyle(&activeTheme, Secondary, false, false, false, false)

	if primary.Background != activeTheme.Palette.FieldBackground || primary.Foreground != activeTheme.Palette.FieldForeground || primary.Placeholder != activeTheme.Palette.FieldPlaceholder {
		t.Fatalf("dark primary = %#v", primary)
	}
	if hovered.Background != activeTheme.Palette.FieldHover {
		t.Fatalf("dark hover = %#v, want %#v", hovered.Background, activeTheme.Palette.FieldHover)
	}
	if secondary.Background != activeTheme.Palette.Default {
		t.Fatalf("dark secondary = %#v, want %#v", secondary.Background, activeTheme.Palette.Default)
	}
}

func TestResolveStyleFocusAndInvalid(t *testing.T) {
	focused := ResolveStyle(defaultTheme(), Primary, true, true, false, false)
	if focused.Border != (color.NRGBA{R: 0x00, G: 0x6f, B: 0xee, A: 0xff}) || focused.BorderWidth != unit.Dp(2) {
		t.Fatalf("focus border = %v at %v, want accent at 2dp", focused.Border, focused.BorderWidth)
	}

	invalid := ResolveStyle(defaultTheme(), Primary, true, true, false, true)
	if invalid.Border != (color.NRGBA{R: 0xf3, G: 0x12, B: 0x60, A: 0xff}) || invalid.BorderWidth != unit.Dp(2) {
		t.Fatalf("focused invalid border = %v at %v, want danger at 2dp", invalid.Border, invalid.BorderWidth)
	}

	unfocused := ResolveStyle(defaultTheme(), Primary, false, false, false, true)
	if unfocused.Border != invalid.Border || unfocused.BorderWidth != unit.Dp(1) {
		t.Fatalf("unfocused invalid border = %v at %v, want danger at 1dp", unfocused.Border, unfocused.BorderWidth)
	}
}

func TestResolveStyleDisabled(t *testing.T) {
	value := ResolveStyle(defaultTheme(), Primary, true, true, true, true)
	if value.Border.A != 0x7f || value.Foreground.A != 0x7f {
		t.Fatalf("disabled alpha = border %d, foreground %d; want 127", value.Border.A, value.Foreground.A)
	}
	if value.ShadowOpacity != 0.5 {
		t.Fatalf("disabled shadow opacity = %v, want 0.5", value.ShadowOpacity)
	}
}

func defaultTheme() *theme.Theme {
	value := theme.DefaultTheme()
	return &value
}
