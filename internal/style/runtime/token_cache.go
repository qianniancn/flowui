package runtime

import (
	"sync"

	"github.com/qianniancn/flowui/internal/style"
)

// tokenCache provides LRU caching for token expansion results.
// It reduces redundant token expansion during style resolution.
type tokenCache struct {
	mu      sync.RWMutex
	entries map[tokenCacheKey]*cacheEntry
	head    *cacheEntry
	tail    *cacheEntry
	size    int
	maxSize int

	// Statistics
	hits      uint64
	misses    uint64
	evictions uint64
}

type tokenCacheKey struct {
	token     style.StyleToken
	themeHash uint64
}

type cacheEntry struct {
	key   tokenCacheKey
	value style.Style
	prev  *cacheEntry
	next  *cacheEntry
}

// defaultCacheSize is chosen to handle typical UI scenarios:
// ~16 tokens × ~4 themes × 4 for headroom = 256 entries
const defaultCacheSize = 256

var globalTokenCache = newTokenCache(defaultCacheSize)

func newTokenCache(maxSize int) *tokenCache {
	return &tokenCache{
		entries: make(map[tokenCacheKey]*cacheEntry, maxSize),
		maxSize: maxSize,
	}
}

// get retrieves a cached token expansion result.
// Returns (value, true) on hit, (zero, false) on miss.
func (c *tokenCache) get(key tokenCacheKey) (style.Style, bool) {
	c.mu.RLock()
	entry, found := c.entries[key]
	c.mu.RUnlock()

	if !found {
		c.mu.Lock()
		c.misses++
		c.mu.Unlock()
		return style.Style{}, false
	}

	// Move to front (most recently used)
	c.mu.Lock()
	c.hits++
	c.moveToFront(entry)
	value := entry.value
	c.mu.Unlock()

	return value, true
}

// put stores a token expansion result in the cache.
func (c *tokenCache) put(key tokenCacheKey, value style.Style) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if already exists
	if entry, found := c.entries[key]; found {
		entry.value = value
		c.moveToFront(entry)
		return
	}

	// Evict if at capacity
	if c.size >= c.maxSize {
		c.evictLRU()
	}

	// Create new entry
	entry := &cacheEntry{
		key:   key,
		value: value,
	}

	c.entries[key] = entry
	c.addToFront(entry)
	c.size++
}

// clear removes all entries from the cache.
func (c *tokenCache) clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[tokenCacheKey]*cacheEntry, c.maxSize)
	c.head = nil
	c.tail = nil
	c.size = 0
}

// stats returns cache performance statistics.
func (c *tokenCache) stats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	total := c.hits + c.misses
	var hitRate float64
	if total > 0 {
		hitRate = float64(c.hits) / float64(total)
	}

	return CacheStats{
		Hits:      c.hits,
		Misses:    c.misses,
		Evictions: c.evictions,
		HitRate:   hitRate,
		Size:      c.size,
	}
}

// CacheStats provides visibility into token cache performance.
type CacheStats struct {
	Hits      uint64
	Misses    uint64
	Evictions uint64
	HitRate   float64
	Size      int
}

// moveToFront moves an entry to the front of the LRU list.
// Caller must hold c.mu.
func (c *tokenCache) moveToFront(entry *cacheEntry) {
	if entry == c.head {
		return
	}

	// Remove from current position
	if entry.prev != nil {
		entry.prev.next = entry.next
	}
	if entry.next != nil {
		entry.next.prev = entry.prev
	}
	if entry == c.tail {
		c.tail = entry.prev
	}

	// Add to front
	entry.prev = nil
	entry.next = c.head
	if c.head != nil {
		c.head.prev = entry
	}
	c.head = entry
	if c.tail == nil {
		c.tail = entry
	}
}

// addToFront adds a new entry to the front of the LRU list.
// Caller must hold c.mu.
func (c *tokenCache) addToFront(entry *cacheEntry) {
	entry.prev = nil
	entry.next = c.head
	if c.head != nil {
		c.head.prev = entry
	}
	c.head = entry
	if c.tail == nil {
		c.tail = entry
	}
}

// evictLRU removes the least recently used entry.
// Caller must hold c.mu.
func (c *tokenCache) evictLRU() {
	if c.tail == nil {
		return
	}

	// Remove tail
	delete(c.entries, c.tail.key)
	c.size--
	c.evictions++

	if c.tail.prev != nil {
		c.tail.prev.next = nil
		c.tail = c.tail.prev
	} else {
		c.head = nil
		c.tail = nil
	}
}

// ClearTokenCache invalidates all cached token expansions.
// Call this when the theme changes.
func ClearTokenCache() {
	globalTokenCache.clear()
}

// TokenCacheStats returns performance metrics for the global token cache.
func TokenCacheStats() CacheStats {
	return globalTokenCache.stats()
}
