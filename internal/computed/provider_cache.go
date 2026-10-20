package computed

import (
	"reflect"

	"github.com/qianniancn/flowui/theme"
)

// ProviderCache memoizes one provider result for a theme and a comparable
// request key. Replacing the active theme's provider also invalidates the cache.
// The cache is intentionally instance-owned: widgets decide the
// request key and retain control of provider validation and request creation.
//
// Resolve invokes load only after a miss. A panic from load therefore leaves
// the previous entry intact and does not turn an unsuccessful request into a
// valid cache value.
type ProviderCache[K comparable, V any] struct {
	theme    *theme.Theme
	provider providerIdentity
	key      K
	value    V
	valid    bool

	hits               uint64
	misses             uint64
	themeInvalidations uint64
	keyInvalidations   uint64
}

// ProviderCacheMap memoizes provider results for multiple comparable request
// keys while sharing the same theme and invalidation lifecycle as
// ProviderCache. It is useful for compound widgets that reuse more than one
// variant in a frame or across frames.
type ProviderCacheMap[K comparable, V any] struct {
	theme    *theme.Theme
	provider providerIdentity
	values   map[K]V
	valid    bool

	// bounded is opt-in so existing map caches keep their zero-value, unbounded
	// behavior without paying for LRU bookkeeping.
	bounded  bool
	capacity int
	clock    uint64
	lastUsed map[K]uint64

	hits               uint64
	misses             uint64
	themeInvalidations uint64
	keyInvalidations   uint64
	evictions          uint64
}

// SetCapacity enables a bounded LRU policy for this cache. A positive
// capacity is required because zero-value ProviderCacheMap instances are
// intentionally unbounded for existing callers. Changing the capacity clears
// any value already held by the cache.
func (c *ProviderCacheMap[K, V]) SetCapacity(capacity int) {
	if c == nil {
		return
	}
	if capacity <= 0 {
		panic("computed: provider cache capacity must be positive")
	}
	if c.bounded && c.capacity == capacity {
		return
	}
	if c.valid {
		c.theme = nil
		c.provider = providerIdentity{}
		c.values = nil
		c.valid = false
	}
	c.bounded = true
	c.capacity = capacity
	c.clock = 0
	c.lastUsed = nil
}

// Resolve returns the cached value for active and key, or loads and stores a
// fresh value on a miss. Loading happens before a new theme map is committed,
// so a panic cannot turn an unsuccessful request into a valid cache value.
func (c *ProviderCacheMap[K, V]) Resolve(active *theme.Theme, key K, load func() V) V {
	value, _ := c.ResolveWithStatus(active, key, load)
	return value
}

// ResolveWithStatus is Resolve with a hit flag for callers that expose cache
// evidence. The loading and invalidation semantics are identical to Resolve.
func (c *ProviderCacheMap[K, V]) ResolveWithStatus(active *theme.Theme, key K, load func() V) (V, bool) {
	if c == nil {
		return load(), false
	}

	provider := visualProvider(active)
	providerMatches := c.valid && c.provider.matches(provider)
	themeChanged := c.valid && (c.theme != active || !providerMatches)
	if themeChanged {
		c.themeInvalidations++
	} else if c.valid {
		if value, ok := c.values[key]; ok {
			c.hits++
			c.touch(key)
			return value, true
		}
		if len(c.values) > 0 {
			c.keyInvalidations++
		}
	}

	c.misses++
	value := load()
	if themeChanged || !c.valid {
		capacity := 2
		if c.bounded {
			capacity = c.capacity
		}
		c.values = make(map[K]V, capacity)
		c.lastUsed = nil
		c.clock = 0
		c.theme = active
		if !providerMatches {
			c.provider = identifyProvider(provider)
		}
		c.valid = true
	}
	if c.bounded {
		if len(c.values) >= c.capacity {
			var oldestKey K
			var oldestUse uint64
			first := true
			for candidate, lastUsed := range c.lastUsed {
				if first || lastUsed < oldestUse {
					oldestKey = candidate
					oldestUse = lastUsed
					first = false
				}
			}
			if !first {
				delete(c.values, oldestKey)
				delete(c.lastUsed, oldestKey)
				c.evictions++
			}
		}
		if c.lastUsed == nil {
			c.lastUsed = make(map[K]uint64, c.capacity)
		}
	}
	c.values[key] = value
	c.touch(key)
	return value, false
}

func (c *ProviderCacheMap[K, V]) touch(key K) {
	if c == nil || !c.bounded {
		return
	}
	c.clock++
	c.lastUsed[key] = c.clock
}

// Stats returns a snapshot of cache counters without exposing cached values.
func (c *ProviderCacheMap[K, V]) Stats() ProviderCacheStats {
	if c == nil {
		return ProviderCacheStats{}
	}
	return ProviderCacheStats{
		Hits:               c.hits,
		Misses:             c.misses,
		ThemeInvalidations: c.themeInvalidations,
		KeyInvalidations:   c.keyInvalidations,
		Evictions:          c.evictions,
		Size:               len(c.values),
	}
}

// Resolve returns the cached value for active and key, or loads and stores a
// fresh value on a miss. A nil receiver deliberately behaves as an uncached
// resolver, which lets callers keep optional test or benchmark caches simple.
func (c *ProviderCache[K, V]) Resolve(active *theme.Theme, key K, load func() V) V {
	if c == nil {
		return load()
	}
	provider := visualProvider(active)
	providerMatches := c.valid && c.provider.matches(provider)
	themeChanged := c.valid && (c.theme != active || !providerMatches)
	if c.valid {
		if themeChanged {
			c.themeInvalidations++
		}
		if c.key != key {
			c.keyInvalidations++
		}
	}
	if c.valid && !themeChanged && c.key == key {
		c.hits++
		return c.value
	}

	value := load()
	c.theme = active
	if !providerMatches {
		c.provider = identifyProvider(provider)
	}
	c.key = key
	c.value = value
	c.valid = true
	c.misses++
	return value
}

func visualProvider(active *theme.Theme) theme.VisualProvider {
	if active == nil {
		return nil
	}
	return active.VisualProvider
}

// Check comparability only when storing a provider to avoid reflection and
// allocations on cache hits. A comparable value can safely be compared with any
// other interface value, even one containing a non-comparable provider.
type providerIdentity struct {
	value      theme.VisualProvider
	comparable bool
}

func identifyProvider(value theme.VisualProvider) providerIdentity {
	return providerIdentity{
		value:      value,
		comparable: value == nil || reflect.ValueOf(value).Comparable(),
	}
}

// Non-comparable providers always miss, including structs whose interface
// fields contain slices. Changes behind the same pointer still require replacing
// the theme; identity checks cannot detect arbitrary internal state changes.
func (id providerIdentity) matches(value theme.VisualProvider) bool {
	return id.comparable && id.value == value
}

// ProviderCacheStats is a snapshot of one ProviderCache instance.
type ProviderCacheStats struct {
	Hits               uint64
	Misses             uint64
	ThemeInvalidations uint64 // Theme pointer or visual provider changes.
	KeyInvalidations   uint64
	Evictions          uint64
	Size               int
}

// Stats returns cache counters without exposing the cached value or key.
func (c *ProviderCache[K, V]) Stats() ProviderCacheStats {
	if c == nil {
		return ProviderCacheStats{}
	}
	return ProviderCacheStats{
		Hits:               c.hits,
		Misses:             c.misses,
		ThemeInvalidations: c.themeInvalidations,
		KeyInvalidations:   c.keyInvalidations,
	}
}
