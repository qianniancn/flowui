package titlebar

import "github.com/qianniancn/flowui/internal/computed"

func (c *titleBarVisualCache) stats() computed.ProviderCacheStats {
	if c == nil {
		return computed.ProviderCacheStats{}
	}
	return c.cache.Stats()
}
