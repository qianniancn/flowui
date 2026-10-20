package titlebar

import (
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/theme"
)

type titleBarVisualKey struct {
	clientDecorations bool
	overrideHash      uint64
	overrideSet       bool
}

type titleBarVisualCache struct {
	cache computed.ProviderCache[titleBarVisualKey, theme.TitleBarVisual]
}

func titleBarOverrideKey(override *theme.TitleBarOverride) (uint64, bool) {
	if override == nil {
		return 0, false
	}
	return override.Declaration.Hash64(), true
}

func (w Widget) visualWithCache(ctx *core.Context, override *theme.TitleBarOverride) theme.TitleBarVisual {
	active := core.ActiveTheme(ctx)
	core.RequireVisualProvider(ctx)
	hash, set := titleBarOverrideKey(override)
	key := titleBarVisualKey{
		clientDecorations: w.clientDecorations,
		overrideHash:      hash,
		overrideSet:       set,
	}
	load := func() theme.TitleBarVisual {
		return theme.ResolveVisual(active, theme.TitleBarRequest{
			ClientDecorations: w.clientDecorations,
			Override:          override,
		}, theme.ComponentVisualProvider.TitleBar, theme.ValidateTitleBarVisual)
	}
	if w.visualCache == nil || override != nil {
		return load()
	}
	return w.visualCache.cache.Resolve(active, key, load)
}
