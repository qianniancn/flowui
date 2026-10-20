package toast

import (
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/theme"
)

type toastVisualKey struct {
	variant     Variant
	override    uint64
	overrideSet bool
}

type toastVisualCache struct {
	cache computed.ProviderCache[toastVisualKey, theme.ToastVisual]
}

type ToastVisualCacheStats = computed.ProviderCacheStats

func toastOverrideKey(override *theme.ToastOverride) (uint64, bool) {
	if override == nil {
		return 0, false
	}
	return override.Declaration.Hash64(), true
}

func (p Widget) baseVisual(ctx *core.Context, override *theme.ToastOverride) theme.ToastVisual {
	return p.visualWithCache(ctx, Default, override)
}

// toastVisualFor is kept as a small package-local probe for behavior tests;
// production layout uses the per-widget cache above.
func toastVisualFor(ctx *core.Context, variant Variant) theme.ToastVisual {
	active := core.ActiveTheme(ctx)
	core.RequireVisualProvider(ctx)
	return resolveToastProvider(active, variant, nil)
}

func (p Widget) visualWithCache(ctx *core.Context, variant Variant, override *theme.ToastOverride) theme.ToastVisual {
	active := core.ActiveTheme(ctx)
	core.RequireVisualProvider(ctx)
	hash, set := toastOverrideKey(override)
	key := toastVisualKey{variant: variant, override: hash, overrideSet: set}
	load := func() theme.ToastVisual {
		return resolveToastProvider(active, variant, override)
	}
	if p.visualCache == nil {
		return load()
	}
	return p.visualCache.cache.Resolve(active, key, load)
}

func resolveToastProvider(active *theme.Theme, variant Variant, override *theme.ToastOverride) theme.ToastVisual {
	return theme.ResolveVisual(active, theme.ToastRequest{Variant: variant, Override: override},
		theme.ComponentVisualProvider.Toast, theme.ValidateToastVisual)
}
