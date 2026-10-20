package pagination

import (
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/theme"
)

type paginationVisualKey struct {
	size theme.PaginationSize
}

type paginationVisualCache struct {
	cache computed.ProviderCache[paginationVisualKey, theme.PaginationVisual]
}

type PaginationVisualCacheStats = computed.ProviderCacheStats

func (c *paginationVisualCache) stats() PaginationVisualCacheStats {
	if c == nil {
		return PaginationVisualCacheStats{}
	}
	return c.cache.Stats()
}

func (p Widget) visualWithCache(ctx *core.Context, override *theme.PaginationOverride) theme.PaginationVisual {
	active := core.ActiveTheme(ctx)
	core.RequireVisualProvider(ctx)
	request := theme.PaginationRequest{Size: theme.PaginationSize(p.size), Override: override}
	if override != nil || p.visualCache == nil {
		return resolvePaginationVisual(active, request)
	}
	key := paginationVisualKey{size: request.Size}
	return p.visualCache.cache.Resolve(active, key, func() theme.PaginationVisual {
		return resolvePaginationVisual(active, request)
	})
}

func resolvePaginationVisual(active *theme.Theme, request theme.PaginationRequest) theme.PaginationVisual {
	return theme.ResolveVisual(active, request, theme.ComponentVisualProvider.Pagination, theme.ValidatePaginationVisual)
}
