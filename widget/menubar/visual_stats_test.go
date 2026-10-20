package menubar

import "github.com/qianniancn/flowui/internal/computed"

func (c *menubarVisualCache) stats() computed.ProviderCacheStats {
	if c == nil {
		return computed.ProviderCacheStats{}
	}
	return c.cache.Stats()
}
