package menubar

import (
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/theme"
)

type menubarVisualKey struct {
	overrideHash uint64
	overrideSet  bool
	compact      bool
}

type menubarVisualCache struct {
	cache computed.ProviderCache[menubarVisualKey, theme.MenubarVisual]
}

func menubarOverrideKey(override *theme.MenubarOverride) (uint64, bool) {
	if override == nil {
		return 0, false
	}
	return override.Declaration.Hash64(), true
}

func (m Widget) visualWithCache(ctx *core.Context, override *theme.MenubarOverride, cache *menubarVisualCache) theme.MenubarVisual {
	active := core.ActiveTheme(ctx)
	core.RequireVisualProvider(ctx)
	hash, set := menubarOverrideKey(override)
	key := menubarVisualKey{overrideHash: hash, overrideSet: set, compact: m.compact}
	load := func() theme.MenubarVisual {
		return theme.ResolveVisual(active, theme.MenubarRequest{
			Compact:  m.compact,
			Override: override,
		}, theme.ComponentVisualProvider.Menubar, theme.ValidateMenubarVisual)
	}
	if cache == nil {
		return load()
	}
	return cache.cache.Resolve(active, key, load)
}
