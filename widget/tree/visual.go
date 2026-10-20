package tree

import (
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/theme"
)

type treeVisualKey struct {
	Variant theme.TreeVariant
	Size    theme.TreeSize
}

type treeVisualCache struct {
	cache computed.ProviderCacheMap[treeVisualKey, theme.TreeVisual]
}

type treeVisualCacheStats = computed.ProviderCacheStats

func (c *treeVisualCache) Stats() treeVisualCacheStats {
	if c == nil {
		return treeVisualCacheStats{}
	}
	return c.cache.Stats()
}

func (c *treeVisualCache) resolve(active *theme.Theme, request theme.TreeRequest) theme.TreeVisual {
	if request.Override != nil || c == nil {
		return resolveTreeVisual(active, request)
	}
	key := treeVisualKey{Variant: request.Variant, Size: request.Size}
	return c.cache.Resolve(active, key, func() theme.TreeVisual {
		return resolveTreeVisual(active, theme.TreeRequest{Variant: request.Variant, Size: request.Size})
	})
}

func resolveTreeVisual(active *theme.Theme, request theme.TreeRequest) theme.TreeVisual {
	return theme.ResolveVisual(active, request, theme.ComponentVisualProvider.Tree, theme.ValidateTreeVisual)
}

func treeThemeVariant(value Variant) theme.TreeVariant {
	if value == VariantSurface {
		return theme.TreeVariantSurface
	}
	return theme.TreeVariantDefault
}

func treeThemeSize(value Size) theme.TreeSize {
	if value == SizeSmall {
		return theme.TreeSizeSmall
	}
	return theme.TreeSizeMedium
}
