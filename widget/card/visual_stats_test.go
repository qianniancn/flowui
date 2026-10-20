package card

import "github.com/qianniancn/flowui/internal/computed"

func (c *cardVisualCache) stats() computed.ProviderCacheStats {
	if c == nil {
		return computed.ProviderCacheStats{}
	}
	return c.cache.Stats()
}
