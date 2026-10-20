package menubar

import (
	"testing"

	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/testtheme"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

type countingMenubarProvider struct {
	theme.ComponentVisualProvider
	calls int
}

func (p *countingMenubarProvider) Menubar(active *theme.Theme, request theme.MenubarRequest) theme.MenubarVisual {
	p.calls++
	return p.ComponentVisualProvider.Menubar(active, request)
}

func newMenubarVisualCacheFixture() (*core.Context, *countingMenubarProvider) {
	active := testtheme.Behavior()
	provider := &countingMenubarProvider{ComponentVisualProvider: testtheme.BehaviorProvider()}
	active.VisualProvider = provider
	return core.New(nil, active, locale.LanguageEnglish), provider
}

func TestMenubarVisualCacheKeys(t *testing.T) {
	tests := []struct {
		name  string
		check func(t *testing.T, ctx *core.Context, provider *countingMenubarProvider, cache *menubarVisualCache)
	}{
		{
			name: "same request hits",
			check: func(t *testing.T, ctx *core.Context, provider *countingMenubarProvider, cache *menubarVisualCache) {
				widget := Widget{}
				widget.visualWithCache(ctx, nil, cache)
				widget.visualWithCache(ctx, nil, cache)
				stats := cache.stats()
				if provider.calls != 1 || stats.Hits != 1 || stats.Misses != 1 {
					t.Fatalf("cache stats = calls %d, stats %#v; want 1, one hit and one miss", provider.calls, stats)
				}
			},
		},
		{
			name: "override hash invalidates",
			check: func(t *testing.T, ctx *core.Context, provider *countingMenubarProvider, cache *menubarVisualCache) {
				widget := Widget{}
				first := &theme.MenubarOverride{Declaration: style.Style{}.TextColor(style.RGB(0x101010))}
				second := &theme.MenubarOverride{Declaration: style.Style{}.TextColor(style.RGB(0x202020))}
				widget.visualWithCache(ctx, first, cache)
				widget.visualWithCache(ctx, second, cache)
				stats := cache.stats()
				if provider.calls != 2 || stats.KeyInvalidations != 1 {
					t.Fatalf("override change stats = calls %d, stats %#v; want 2 calls and one key invalidation", provider.calls, stats)
				}
			},
		},
		{
			name: "compact invalidates",
			check: func(t *testing.T, ctx *core.Context, provider *countingMenubarProvider, cache *menubarVisualCache) {
				widget := Widget{}
				widget.visualWithCache(ctx, nil, cache)
				widget.compact = true
				widget.visualWithCache(ctx, nil, cache)
				stats := cache.stats()
				if provider.calls != 2 || stats.KeyInvalidations != 1 {
					t.Fatalf("compact change stats = calls %d, stats %#v; want 2 calls and one key invalidation", provider.calls, stats)
				}
			},
		},
		{
			name: "theme pointer invalidates",
			check: func(t *testing.T, ctx *core.Context, provider *countingMenubarProvider, cache *menubarVisualCache) {
				widget := Widget{}
				widget.visualWithCache(ctx, nil, cache)
				replacement := *core.ActiveTheme(ctx)
				core.ReplaceTheme(ctx, replacement)
				widget.visualWithCache(ctx, nil, cache)
				stats := cache.stats()
				if provider.calls != 2 || stats.ThemeInvalidations != 1 {
					t.Fatalf("theme change stats = calls %d, stats %#v; want 2 calls and one theme invalidation", provider.calls, stats)
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx, provider := newMenubarVisualCacheFixture()
			test.check(t, ctx, provider, &menubarVisualCache{})
		})
	}
}

func TestMenubarVisualCacheValidatesProviderOnHit(t *testing.T) {
	ctx, _ := newMenubarVisualCacheFixture()
	widget := Widget{}
	cache := &menubarVisualCache{}
	widget.visualWithCache(ctx, nil, cache)
	core.ActiveTheme(ctx).VisualProvider = nil

	deferred := func() {
		widget.visualWithCache(ctx, nil, cache)
	}
	deferredPanic := func() (recovered any) {
		defer func() { recovered = recover() }()
		deferred()
		return nil
	}()
	if deferredPanic == nil {
		t.Fatal("cache hit accepted a theme without a visual provider")
	}
}

func TestMenubarMissingProviderFailsExplicitly(t *testing.T) {
	active := testtheme.Behavior()
	active.VisualProvider = nil
	ctx := newMenubarContext(t, active)

	deferredPanic := func() (recovered any) {
		defer func() { recovered = recover() }()
		(Widget{}).visualWithCache(ctx, nil, nil)
		return nil
	}()
	if deferredPanic != theme.ErrMissingVisualProvider {
		t.Fatalf("missing menubar provider panic = %v, want %v", deferredPanic, theme.ErrMissingVisualProvider)
	}
}
