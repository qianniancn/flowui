package listbox

import "github.com/qianniancn/flowui/internal/computed"

func (c *listBoxVisualCache) stats() computed.ProviderCacheStats {
	if c == nil {
		return computed.ProviderCacheStats{}
	}
	return c.cache.Stats()
}
