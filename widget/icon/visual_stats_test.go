package icon

import "github.com/qianniancn/flowui/internal/computed"

func (c *iconVisualCache) stats() computed.ProviderCacheStats {
	if c == nil {
		return computed.ProviderCacheStats{}
	}
	return c.cache.Stats()
}
