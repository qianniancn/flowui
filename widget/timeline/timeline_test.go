package timeline

import (
	"image/color"
	"math"
	"testing"

	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/theme"
)

func TestNewCopiesItems(t *testing.T) {
	items := []Item{{Key: "one"}}
	widget := New(items)
	items[0].Key = "changed"
	if widget.items[0].Key != "one" {
		t.Fatal("timeline copied items were mutated by the caller")
	}
}

func TestMarkerColors(t *testing.T) {
	custom := color.NRGBA{R: 1, G: 2, B: 3, A: 0xff}
	colors := theme.TimelineColors{
		Primary:  color.NRGBA{R: 4, A: 0xff},
		Error:    color.NRGBA{G: 5, A: 0xff},
		Success:  color.NRGBA{B: 6, A: 0xff},
		Muted:    color.NRGBA{R: 7, G: 7, A: 0xff},
		Disabled: color.NRGBA{R: 8, G: 8, B: 8, A: 0xff},
	}
	checks := []struct {
		item Item
		want color.NRGBA
	}{
		{item: Item{Color: ColorBlue}, want: colors.Primary},
		{item: Item{Color: ColorRed}, want: colors.Error},
		{item: Item{Color: ColorGreen}, want: colors.Success},
		{item: Item{Color: ColorGray}, want: colors.Muted},
		{item: Item{Color: ColorCustom, CustomColor: custom}, want: custom},
		{item: Item{Disabled: true}, want: colors.Disabled},
	}
	for _, check := range checks {
		if got := markerColor(colors, check.item); got != check.want {
			t.Fatalf("marker Color = %#v, want %#v", got, check.want)
		}
	}
}

func TestMarkerColorPrecedence(t *testing.T) {
	custom := color.NRGBA{R: 1, G: 2, B: 3, A: 0xff}
	colors := theme.TimelineColors{
		Primary:  color.NRGBA{R: 4, A: 0xff},
		Error:    color.NRGBA{G: 5, A: 0xff},
		Success:  color.NRGBA{B: 6, A: 0xff},
		Muted:    color.NRGBA{R: 7, G: 7, A: 0xff},
		Disabled: color.NRGBA{R: 8, G: 8, B: 8, A: 0xff},
	}
	tests := []struct {
		name string
		item Item
		want color.NRGBA
	}{
		{name: "disabled wins over loading and custom", item: Item{Disabled: true, Loading: true, Color: ColorCustom, CustomColor: custom}, want: colors.Disabled},
		{name: "loading wins over custom", item: Item{Loading: true, Color: ColorCustom, CustomColor: custom}, want: colors.Primary},
		{name: "custom color wins over semantic color", item: Item{Color: ColorCustom, CustomColor: custom}, want: custom},
		{name: "transparent custom falls back to primary", item: Item{Color: ColorCustom}, want: colors.Primary},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := markerColor(colors, test.item); got != test.want {
				t.Fatalf("marker color = %#v, want %#v", got, test.want)
			}
		})
	}
}

func TestTitleWidthRejectsNegativeValues(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("negative title width did not panic")
		}
	}()
	New(nil).TitleWidth(-1)
}

func TestTitleWidthUsesDeviceIndependentPixels(t *testing.T) {
	gtx := layout.Context{Metric: unit.Metric{PxPerDp: 2}}
	if got := New(nil).TitleWidth(80).resolveTitleSpanPx(gtx, 500); got != 160 {
		t.Fatalf("title width = %d px, want 160 px at 2x scale", got)
	}
}

func TestGapRejectsNegativeValues(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("negative gap did not panic")
		}
	}()
	New(nil).Gap(-1)
}

func TestTitleSpanRejectsInvalidValues(t *testing.T) {
	tests := []struct {
		name  string
		value float32
	}{
		{name: "zero", value: 0},
		{name: "negative", value: -1},
		{name: "above twenty four", value: 25},
		{name: "not a number", value: float32(math.NaN())},
		{name: "positive infinity", value: float32(math.Inf(1))},
		{name: "negative infinity", value: float32(math.Inf(-1))},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatalf("title span %v did not panic", test.value)
				}
			}()
			New(nil).TitleSpan(test.value)
		})
	}
}

func TestResolvePlacementPrefersItemPlacement(t *testing.T) {
	widget := New(nil).Mode(ModeAlternate)
	if got := widget.resolvePlacement(Item{Placement: PlacementEnd}, 0); got != PlacementEnd {
		t.Fatalf("resolved Placement = %v, want %v", got, PlacementEnd)
	}
	if got := widget.resolvePlacement(Item{}, 1); got != PlacementEnd {
		t.Fatalf("alternate Placement = %v, want %v", got, PlacementEnd)
	}
}

func TestResolveItemsAppliesModeReverseAndDisabled(t *testing.T) {
	items := []Item{
		{Key: "one"},
		{Key: "two"},
		{Key: "three", Placement: PlacementEnd},
	}
	tests := []struct {
		name        string
		widget      Widget
		wantKeys    []string
		wantPlaces  []Placement
		wantDisable bool
	}{
		{
			name:       "alternate preserves resolved placement when reversed",
			widget:     New(items).Mode(ModeAlternate).Reverse(true),
			wantKeys:   []string{"three", "two", "one"},
			wantPlaces: []Placement{PlacementEnd, PlacementEnd, PlacementStart},
		},
		{
			name:        "end mode disables every item",
			widget:      New(items).Mode(ModeEnd).Disabled(true),
			wantKeys:    []string{"one", "two", "three"},
			wantPlaces:  []Placement{PlacementEnd, PlacementEnd, PlacementEnd},
			wantDisable: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resolved := test.widget.resolveItems(test.widget.items)
			if len(resolved) != len(test.wantKeys) {
				t.Fatalf("resolved item count = %d, want %d", len(resolved), len(test.wantKeys))
			}
			for index, item := range resolved {
				if item.Key != test.wantKeys[index] {
					t.Errorf("item %d key = %q, want %q", index, item.Key, test.wantKeys[index])
				}
				if item.placement != test.wantPlaces[index] {
					t.Errorf("item %d placement = %v, want %v", index, item.placement, test.wantPlaces[index])
				}
				if item.Disabled != test.wantDisable {
					t.Errorf("item %d disabled = %v, want %v", index, item.Disabled, test.wantDisable)
				}
			}
		})
	}
}

func TestUsesStructuredVerticalDependsOnModeAndTitles(t *testing.T) {
	tests := []struct {
		name   string
		widget Widget
		items  []resolvedItem
		want   bool
	}{
		{name: "alternate mode is structured", widget: New(nil).Mode(ModeAlternate), want: true},
		{name: "title makes start mode structured", widget: New(nil).Mode(ModeStart), items: []resolvedItem{{Item: Item{Title: timelineTestWidget{}}}}, want: true},
		{name: "content only start mode is inline", widget: New(nil).Mode(ModeStart), items: []resolvedItem{{Item: Item{Content: timelineTestWidget{}}}}, want: false},
		{name: "content only end mode is inline", widget: New(nil).Mode(ModeEnd), items: []resolvedItem{{Item: Item{Content: timelineTestWidget{}}}}, want: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := test.widget.usesStructuredVertical(test.items); got != test.want {
				t.Fatalf("structured vertical = %v, want %v", got, test.want)
			}
		})
	}
}

type timelineTestWidget struct{}

func (timelineTestWidget) Layout(_ *core.Context, gtx layout.Context) layout.Dimensions {
	return layout.Dimensions{Size: gtx.Constraints.Min}
}
