package colorpicker

import "github.com/qianniancn/flowui/internal/computed"

func (cache *colorPickerVisualCache) stats() computed.ProviderCacheStats {
	if cache == nil {
		return computed.ProviderCacheStats{}
	}
	return cache.cache.Stats()
}
