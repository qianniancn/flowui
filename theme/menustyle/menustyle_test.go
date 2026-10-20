package menustyle_test

import (
	"image/color"
	"testing"

	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/theme/base"
	"github.com/qianniancn/flowui/theme/menustyle"
)

// TestMenuHighlightText checks that a highlighted item keeps its label readable,
// and that the destructive item uses white text on the red fill.
func TestMenuHighlightText(t *testing.T) {
	light := base.Light()
	dark := base.Dark()
	white := color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
	for _, tc := range []struct {
		name   string
		active *theme.Theme
	}{
		{"light", &light},
		{"dark", &dark},
	} {
		t.Run(tc.name, func(t *testing.T) {
			visual := menustyle.BuildMenu(tc.active, theme.MenuRequest{})
			sem := tc.active.Semantic()
			for _, test := range []struct {
				name    string
				variant uint8
				fill    color.NRGBA
			}{
				{"default", 0, sem.Tone(theme.TonePrimary).Fill},
				{"danger", 1, sem.Tone(theme.ToneDanger).Fill},
			} {
				item := visual.Item(theme.MenuItemRequest{Variant: test.variant})
				part := style.CascadePart(style.State{Hovered: true}, style.PartItem, item.Defaults, item.Variant)
				background, ok := part.Paint.Background.(style.SolidColor)
				if !ok {
					t.Fatalf("%s fill = %#v", test.name, part.Paint.Background)
				}
				if background.Color != test.fill {
					t.Fatalf("%s fill = %v, want %v", test.name, background.Color, test.fill)
				}
				text := item.Colors(part).Foreground
				if text != white {
					t.Fatalf("%s highlight text = %v, want white", test.name, text)
				}
				if ratio := theme.ContrastRatio(test.fill, text); ratio < 3 {
					t.Fatalf("%s contrast = %.2f", test.name, ratio)
				}
			}
		})
	}
}
