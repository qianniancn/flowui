package toolbar

import (
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/theme"
)

type toolbarVisualKey struct {
	orientation  theme.ToolbarOrientation
	attached     bool
	overrideHash uint64
	overrideSet  bool
}

type toolbarVisualCache struct {
	cache computed.ProviderCache[toolbarVisualKey, theme.ToolbarVisual]
}

type ToolbarVisualCacheStats = computed.ProviderCacheStats

func (c *toolbarVisualCache) stats() ToolbarVisualCacheStats {
	if c == nil {
		return ToolbarVisualCacheStats{}
	}
	return c.cache.Stats()
}

func toolbarOverrideKey(override *theme.ToolbarOverride) (uint64, bool) {
	if override == nil {
		return 0, false
	}
	return override.Declaration.Hash64(), true
}

func (w Widget) visualWithCache(ctx *core.Context, override *theme.ToolbarOverride) theme.ToolbarVisual {
	active := core.ActiveTheme(ctx)
	core.RequireVisualProvider(ctx)
	hash, set := toolbarOverrideKey(override)
	key := toolbarVisualKey{
		orientation:  toolbarOrientation(w.orientation),
		attached:     w.attached,
		overrideHash: hash,
		overrideSet:  set,
	}
	load := func() theme.ToolbarVisual {
		return resolveToolbarVisual(active, theme.ToolbarRequest{
			Orientation: key.orientation,
			Attached:    w.attached,
			Override:    override,
		})
	}
	if w.visualCache == nil {
		return load()
	}
	return w.visualCache.cache.Resolve(active, key, load)
}

func resolveToolbarVisual(active *theme.Theme, request theme.ToolbarRequest) theme.ToolbarVisual {
	return theme.ResolveVisual(active, request, theme.ComponentVisualProvider.Toolbar, theme.ValidateToolbarVisual)
}

func toolbarOrientation(value Orientation) theme.ToolbarOrientation {
	if value == Vertical {
		return theme.ToolbarVertical
	}
	return theme.ToolbarHorizontal
}
