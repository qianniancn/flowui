package toast

import "github.com/qianniancn/flowui/internal/computed"

func (c *toastVisualCache) stats() computed.ProviderCacheStats {
	if c == nil {
		return computed.ProviderCacheStats{}
	}
	return c.cache.Stats()
}
