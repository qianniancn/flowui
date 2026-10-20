package pagination

import (
	"errors"
	"image"
	"testing"

	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/testtheme"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

type countingPaginationProvider struct {
	theme.ComponentVisualProvider
	calls int
}

func (p *countingPaginationProvider) Pagination(active *theme.Theme, request theme.PaginationRequest) theme.PaginationVisual {
	p.calls++
	return p.ComponentVisualProvider.Pagination(active, request)
}

type incompletePaginationProvider struct{ theme.ComponentVisualProvider }

func (incompletePaginationProvider) Pagination(*theme.Theme, theme.PaginationRequest) theme.PaginationVisual {
	return theme.PaginationVisual{}
}

func TestPaginationVisualCacheTracksSizeAndTheme(t *testing.T) {
	first := *testtheme.Behavior()
	provider := &countingPaginationProvider{ComponentVisualProvider: first.VisualProvider.(theme.ComponentVisualProvider)}
	first.VisualProvider = provider
	second := *testtheme.Behavior()
	second.VisualProvider = provider
	firstContext := core.New(nil, &first, locale.LanguageEnglish)
	secondContext := core.New(nil, &second, locale.LanguageEnglish)

	value := New("cache", 1, 5)
	value.visualWithCache(firstContext, nil)
	value.visualWithCache(firstContext, nil)
	value.Size(SizeSmall).visualWithCache(firstContext, nil)
	value.visualWithCache(secondContext, nil)

	if provider.calls != 3 {
		t.Fatalf("pagination provider calls = %d, want 3", provider.calls)
	}
	stats := value.visualCache.stats()
	if stats.Hits != 1 || stats.Misses != 3 || stats.KeyInvalidations != 2 || stats.ThemeInvalidations != 1 {
		t.Fatalf("pagination cache stats = %#v", stats)
	}
}

func TestPaginationIncompleteVisualDoesNotPolluteCache(t *testing.T) {
	active := testtheme.Behavior()
	active.VisualProvider = incompletePaginationProvider{ComponentVisualProvider: testtheme.BehaviorProvider()}
	ctx := core.New(nil, active, locale.LanguageEnglish)
	value := New("incomplete", 1, 3)
	defer func() {
		recovered := recover()
		if recoveredErr, ok := recovered.(error); !ok || !errors.Is(recoveredErr, theme.ErrIncompleteVisualProvider) {
			t.Fatalf("incomplete pagination visual panic = %v", recovered)
		}
		if stats := value.visualCache.stats(); stats.Hits != 0 || stats.Misses != 0 {
			t.Fatalf("incomplete pagination visual polluted cache: %#v", stats)
		}
	}()
	value.Layout(ctx, layout.Context{Constraints: layout.Constraints{Max: image.Pt(320, 80)}})
}

func TestPaginationStyledPassesOverrideToProvider(t *testing.T) {
	active := *testtheme.Behavior()
	base := active.VisualProvider.(theme.ComponentVisualProvider)
	provider := &recordingPaginationProvider{ComponentVisualProvider: base}
	active.VisualProvider = provider
	ctx := core.New(nil, &active, locale.LanguageEnglish)
	(New("styled", 1, 3)).Style(style.Style{}.Radius(3)).Layout(ctx, layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(320, 80)},
		Ops:         new(op.Ops),
	})
	if len(provider.requests) != 1 || provider.requests[0].Override == nil {
		t.Fatalf("pagination provider requests = %#v", provider.requests)
	}
}

type recordingPaginationProvider struct {
	theme.ComponentVisualProvider
	requests []theme.PaginationRequest
}

func (p *recordingPaginationProvider) Pagination(active *theme.Theme, request theme.PaginationRequest) theme.PaginationVisual {
	p.requests = append(p.requests, request)
	return p.ComponentVisualProvider.Pagination(active, request)
}
