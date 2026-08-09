package timeline

import (
	"image/color"
	"testing"

	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/theme"
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
	active := theme.DefaultTheme()
	custom := color.NRGBA{R: 1, G: 2, B: 3, A: 0xff}
	checks := []struct {
		item Item
		want color.NRGBA
	}{
		{Item{Color: ColorBlue}, active.Components.TimeLine.PrimaryColor},
		{Item{Color: ColorRed}, active.Components.TimeLine.ErrorColor},
		{Item{Color: ColorGreen}, active.Components.TimeLine.SuccessColor},
		{Item{Color: ColorGray}, active.Components.TimeLine.MutedColor},
		{Item{Color: ColorCustom, CustomColor: custom}, custom},
		{Item{Disabled: true}, active.Components.TimeLine.DisabledColor},
	}
	for _, check := range checks {
		if got := markerColor(&active, check.item); got != check.want {
			t.Fatalf("marker color = %#v, want %#v", got, check.want)
		}
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
	tests := []float32{0, -1, 25}
	for _, value := range tests {
		func() {
			defer func() {
				if recover() == nil {
					t.Fatalf("title span %v did not panic", value)
				}
			}()
			New(nil).TitleSpan(value)
		}()
	}
}

func TestResolvePlacementPrefersItemPlacement(t *testing.T) {
	widget := New(nil).Mode(ModeAlternate)
	if got := widget.resolvePlacement(Item{Placement: PlacementEnd}, 0); got != PlacementEnd {
		t.Fatalf("resolved placement = %v, want %v", got, PlacementEnd)
	}
	if got := widget.resolvePlacement(Item{}, 1); got != PlacementEnd {
		t.Fatalf("alternate placement = %v, want %v", got, PlacementEnd)
	}
}
