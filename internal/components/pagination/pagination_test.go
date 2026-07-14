package pagination

import (
	"image"
	"testing"
	"time"

	"gioui.org/io/input"
	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/locale"
	"github.com/qianniancn/FlowUI/internal/theme"
)

func TestPaginationOptionsUseValueSemantics(t *testing.T) {
	base := New("pages", 2, 12)
	configured := base.
		Size(SizeSmall).
		Siblings(2).
		Boundaries(2).
		ShowControls(false).
		Labels("Back", "Forward").
		Disabled(true).
		OnChange(func(int) {})
	if base.size != SizeMedium || base.siblings != 1 || base.boundaries != 1 || !base.showControls {
		t.Fatal("configuring Pagination mutated the base value")
	}
	if configured.size != SizeSmall || configured.siblings != 2 || configured.boundaries != 2 || configured.showControls || !configured.disabled || configured.onChange == nil {
		t.Fatalf("configured Pagination = %+v", configured)
	}
}

func TestPaginationPageItems(t *testing.T) {
	tests := []struct {
		name  string
		page  int
		total int
		want  []int
	}{
		{name: "short", page: 2, total: 4, want: []int{1, 2, 3, 4}},
		{name: "start", page: 1, total: 12, want: []int{1, 2, 0, 12}},
		{name: "middle", page: 6, total: 12, want: []int{1, 0, 5, 6, 7, 0, 12}},
		{name: "end", page: 12, total: 12, want: []int{1, 0, 11, 12}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := pageItems(test.page, test.total, 1, 1)
			if !samePages(got, test.want) {
				t.Fatalf("pageItems(%d, %d) = %v, want %v", test.page, test.total, got, test.want)
			}
		})
	}
}

func TestPaginationControlledPageClick(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	ctx := frame.New(nil, &activeTheme, locale.LanguageEnglish)
	router := new(input.Router)
	changed := 0
	widget := New("pages", 1, 5).Size(SizeSmall).OnChange(func(page int) { changed = page })
	gtx := paginationTestContext(router, image.Pt(480, 80), time.Unix(1, 0))

	frame.BeginFrame(ctx)
	dims := widget.Layout(ctx, gtx)
	frame.EndFrame(ctx)
	router.Frame(gtx.Ops)
	if dims.Size.Y != gtx.Dp(activeTheme.Components.Pagination.SmallSize) {
		t.Fatalf("Pagination height = %d", dims.Size.Y)
	}
	value, ok := frame.PeekState[paginationState](ctx, "pages", stateSlotPagination)
	if !ok || value.items["page:2"] == nil {
		t.Fatal("page item state was not retained")
	}
	value.items["page:2"].clickable.Click()

	gtx = paginationTestContext(router, image.Pt(480, 80), time.Unix(1, int64(time.Millisecond)))
	frame.BeginFrame(ctx)
	widget.Layout(ctx, gtx)
	frame.EndFrame(ctx)
	router.Frame(gtx.Ops)
	if changed != 2 {
		t.Fatalf("changed page = %d, want 2", changed)
	}
}

func TestPaginationThemeMatchesHeroUI(t *testing.T) {
	tokens := theme.DefaultTheme().Components.Pagination
	if tokens.SmallSize != 28 || tokens.MediumSize != 32 || tokens.LargeSize != 36 || tokens.ItemGap != 4 || tokens.FocusRingWidth != 2 {
		t.Fatalf("Pagination theme = %+v", tokens)
	}
}

func paginationTestContext(router *input.Router, max image.Point, now time.Time) layout.Context {
	var source input.Source
	if router != nil {
		source = router.Source()
	}
	return layout.Context{Constraints: layout.Constraints{Max: max}, Source: source, Ops: new(op.Ops), Now: now}
}

func samePages(left, right []int) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
