package label

import "github.com/qianniancn/flowui/internal/computed"

func (c *labelVisualCache) stats() computed.ProviderCacheStats {
	if c == nil {
		return computed.ProviderCacheStats{}
	}
	return c.cache.Stats()
}
