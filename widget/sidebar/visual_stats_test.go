package sidebar

import "github.com/qianniancn/flowui/internal/computed"

func (c *sidebarVisualCache) stats() computed.ProviderCacheStats {
	if c == nil {
		return computed.ProviderCacheStats{}
	}
	return c.cache.Stats()
}
