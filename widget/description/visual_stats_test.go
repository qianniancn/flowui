package description

import "github.com/qianniancn/flowui/internal/computed"

func (c *descriptionVisualCache) stats() computed.ProviderCacheStats {
	if c == nil {
		return computed.ProviderCacheStats{}
	}
	return c.cache.Stats()
}
