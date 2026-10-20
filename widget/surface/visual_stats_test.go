package surface

import "github.com/qianniancn/flowui/internal/computed"

func (c *surfaceVisualCache) stats() computed.ProviderCacheStats {
	if c == nil {
		return computed.ProviderCacheStats{}
	}
	return c.cache.Stats()
}
