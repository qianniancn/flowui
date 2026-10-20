package tabs

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

type incompleteTabsProvider struct{ theme.ComponentVisualProvider }

func (incompleteTabsProvider) Tabs(*theme.Theme, theme.TabsRequest) theme.TabsVisual {
	return theme.TabsVisual{}
}

type recordingTabsProvider struct {
	theme.ComponentVisualProvider
	requests []theme.TabsRequest
}

func (p *recordingTabsProvider) Tabs(active *theme.Theme, request theme.TabsRequest) theme.TabsVisual {
	p.requests = append(p.requests, request)
	return p.ComponentVisualProvider.Tabs(active, request)
}

func TestTabsVisualCacheTracksThemeAndKey(t *testing.T) {
	first := testtheme.Behavior()
	second := testtheme.Behavior()
	firstContext := core.New(nil, first, locale.LanguageEnglish)
	secondContext := core.New(nil, second, locale.LanguageEnglish)
	value := New("cache", "one", []TabItem{{Key: "one", Label: "One"}})
	cache := tabsVisualCache{}
	value.visualWithCache(firstContext, nil, &cache)
	value.visualWithCache(firstContext, nil, &cache)
	value.Size(Small).visualWithCache(firstContext, nil, &cache)
	value.visualWithCache(secondContext, nil, &cache)
	value.visualWithCache(firstContext, &theme.TabsOverride{Declaration: style.Style{}.Padding(2)}, &cache)
	stats := cache.cache.Stats()
	if stats.Hits != 1 || stats.Misses != 4 || stats.KeyInvalidations != 3 || stats.ThemeInvalidations != 2 {
		t.Fatalf("tabs cache stats = %#v", stats)
	}
}

func TestTabsIncompleteVisualDoesNotPolluteCache(t *testing.T) {
	active := testtheme.Behavior()
	active.VisualProvider = incompleteTabsProvider{ComponentVisualProvider: testtheme.BehaviorProvider()}
	ctx := core.New(nil, active, locale.LanguageEnglish)
	value := New("incomplete", "one", []TabItem{{Key: "one", Label: "One"}})
	defer func() {
		recovered := recover()
		if recoveredErr, ok := recovered.(error); !ok || !errors.Is(recoveredErr, theme.ErrIncompleteVisualProvider) {
			t.Fatalf("incomplete tabs visual panic = %v", recovered)
		}
		if stats := value.visualCache.cache.Stats(); stats.Hits != 0 || stats.Misses != 0 {
			t.Fatalf("incomplete tabs visual polluted cache: %#v", stats)
		}
	}()
	value.Layout(ctx, layout.Context{Constraints: layout.Constraints{Max: image.Pt(320, 120)}})
}

func TestTabsStyledPassesOverrideToProvider(t *testing.T) {
	active := testtheme.Behavior()
	provider := &recordingTabsProvider{ComponentVisualProvider: testtheme.BehaviorProvider()}
	active.VisualProvider = provider
	ctx := core.New(nil, active, locale.LanguageEnglish)
	(New("styled", "one", []TabItem{{Key: "one", Label: "One"}})).Style(style.Style{}.Radius(3)).Layout(ctx, layout.Context{Constraints: layout.Constraints{Max: image.Pt(320, 120)}, Ops: new(op.Ops)})
	if len(provider.requests) != 1 || provider.requests[0].Override == nil {
		t.Fatalf("tabs provider requests = %#v", provider.requests)
	}
}
