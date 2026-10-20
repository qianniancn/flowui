package tabs

import (
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/theme"
)

type tabsVisualKey struct {
	variant      uint8
	orientation  uint8
	size         uint8
	color        uint8
	overrideHash uint64
	overrideSet  bool
}

type tabsVisualCache struct {
	cache computed.ProviderCache[tabsVisualKey, theme.TabsVisual]
}

type TabsVisualCacheStats = computed.ProviderCacheStats

func tabsOverrideKey(override *theme.TabsOverride) (uint64, bool) {
	if override == nil {
		return 0, false
	}
	return override.Declaration.Hash64(), true
}

func (t Widget) visualWithCache(ctx *core.Context, override *theme.TabsOverride, cache *tabsVisualCache) theme.TabsVisual {
	active := core.ActiveTheme(ctx)
	core.RequireVisualProvider(ctx)
	hash, set := tabsOverrideKey(override)
	key := tabsVisualKey{
		variant: uint8(t.variant), orientation: uint8(t.orientation), size: uint8(t.size), color: uint8(t.color),
		overrideHash: hash, overrideSet: set,
	}
	load := func() theme.TabsVisual {
		return theme.ResolveVisual(active, theme.TabsRequest{
			Variant: uint8(t.variant), Orientation: uint8(t.orientation), Size: uint8(t.size), Color: uint8(t.color),
			Override: override,
		}, theme.ComponentVisualProvider.Tabs, theme.ValidateTabsVisual)
	}
	if cache == nil {
		return load()
	}
	return cache.cache.Resolve(active, key, load)
}

func (t Widget) visual(ctx *core.Context) theme.TabsVisual {
	return t.visualWithCache(ctx, nil, t.visualCache)
}
