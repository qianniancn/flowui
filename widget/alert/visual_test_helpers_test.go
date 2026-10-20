package alert

import "github.com/qianniancn/flowui/internal/computed"

func (c *alertVisualCache) stats() computed.ProviderCacheStats {
	if c == nil {
		return computed.ProviderCacheStats{}
	}
	return c.cache.Stats()
}
