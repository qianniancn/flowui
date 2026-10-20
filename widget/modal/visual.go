package modal

import (
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/theme"
)

// ModalVisualCacheStats is a snapshot of one Modal instance's visual cache.
type ModalVisualCacheStats struct {
	Hits               uint64
	Misses             uint64
	ThemeInvalidations uint64
	KeyInvalidations   uint64
}

type modalVisualKey struct {
	overrideHash uint64
	overrideSet  bool
	size         Size
	backdrop     BackdropVariant
}

type modalVisualCache struct {
	cache computed.ProviderCache[modalVisualKey, theme.ModalVisual]
}

func (c *modalVisualCache) stats() ModalVisualCacheStats {
	if c == nil {
		return ModalVisualCacheStats{}
	}
	cacheStats := c.cache.Stats()
	return ModalVisualCacheStats{
		Hits:               cacheStats.Hits,
		Misses:             cacheStats.Misses,
		ThemeInvalidations: cacheStats.ThemeInvalidations,
		KeyInvalidations:   cacheStats.KeyInvalidations,
	}
}

func modalOverrideKey(override *theme.ModalOverride) (uint64, bool) {
	if override == nil {
		return 0, false
	}
	return override.Declaration.Hash64(), true
}

// visualWithCache is the only bridge from modal behavior to theme visuals.
// Layout resolves it once and passes the immutable bundle through every
// geometry, style, motion, and drawing path in the overlay.
func (m Widget) visualWithCache(ctx *core.Context, override *theme.ModalOverride, cache *modalVisualCache) theme.ModalVisual {
	active := core.ActiveTheme(ctx)
	core.RequireVisualProvider(ctx)
	hash, set := modalOverrideKey(override)
	key := modalVisualKey{overrideHash: hash, overrideSet: set, size: m.size, backdrop: m.backdrop}
	return cache.resolve(active, key, func() theme.ModalVisual {
		return theme.ResolveVisual(active, theme.ModalRequest{
			Size: modalSize(m.size), Backdrop: modalBackdrop(m.backdrop), Override: override,
		}, theme.ComponentVisualProvider.Modal, theme.ValidateModalVisual)
	})
}

func (c *modalVisualCache) resolve(active *theme.Theme, key modalVisualKey, load func() theme.ModalVisual) theme.ModalVisual {
	if c == nil {
		return load()
	}
	return c.cache.Resolve(active, key, load)
}

func modalSize(value Size) theme.ModalSize {
	return theme.ModalSize(value)
}

func modalBackdrop(value BackdropVariant) theme.ModalBackdrop {
	return theme.ModalBackdrop(value)
}
