package runtime

import (
	"testing"

	"github.com/qianniancn/FlowUI/internal/style"
)

func TestTokenCache(t *testing.T) {
	// Clear cache before test
	ClearTokenCache()

	// Create two identical cache keys
	key1 := tokenCacheKey{token: style.TokenControlHeight, themeHash: 12345}
	key2 := tokenCacheKey{token: style.TokenControlHeight, themeHash: 12345}
	keyDiff := tokenCacheKey{token: style.TokenControlRadius, themeHash: 12345}

	// Test cache miss
	if _, found := globalTokenCache.get(key1); found {
		t.Error("Expected cache miss on first access")
	}

	// Put value in cache
	value := style.Style{}.Height(36)
	globalTokenCache.put(key1, value)

	// Test cache hit with same key
	if cached, found := globalTokenCache.get(key2); !found {
		t.Error("Expected cache hit with identical key")
	} else {
		// Verify the cached value (basic check - Style equality would need more work)
		_ = cached
	}

	// Test cache miss with different key
	if _, found := globalTokenCache.get(keyDiff); found {
		t.Error("Expected cache miss with different token")
	}

	// Check stats
	stats := globalTokenCache.stats()
	if stats.Hits == 0 {
		t.Error("Expected at least one cache hit")
	}
	if stats.Misses == 0 {
		t.Error("Expected at least one cache miss")
	}
	if stats.Size == 0 {
		t.Error("Expected non-zero cache size")
	}
}

func TestTokenCacheLRU(t *testing.T) {
	// Create a small cache for testing
	cache := newTokenCache(3)

	// Fill cache
	cache.put(tokenCacheKey{token: style.TokenControlHeight, themeHash: 1}, style.Style{})
	cache.put(tokenCacheKey{token: style.TokenControlRadius, themeHash: 1}, style.Style{})
	cache.put(tokenCacheKey{token: style.TokenControlPaddingX, themeHash: 1}, style.Style{})

	if cache.size != 3 {
		t.Errorf("Expected cache size 3, got %d", cache.size)
	}

	// Access first entry to make it most recent
	cache.get(tokenCacheKey{token: style.TokenControlHeight, themeHash: 1})

	// Add fourth entry, should evict second entry (LRU)
	cache.put(tokenCacheKey{token: style.TokenBodyFontSize, themeHash: 1}, style.Style{})

	if cache.size != 3 {
		t.Errorf("Expected cache size to remain 3 after eviction, got %d", cache.size)
	}

	stats := cache.stats()
	if stats.Evictions == 0 {
		t.Error("Expected at least one eviction")
	}

	// First entry should still be in cache (was accessed)
	if _, found := cache.get(tokenCacheKey{token: style.TokenControlHeight, themeHash: 1}); !found {
		t.Error("Expected first entry to remain after being accessed")
	}

	// Second entry should be evicted
	if _, found := cache.get(tokenCacheKey{token: style.TokenControlRadius, themeHash: 1}); found {
		t.Error("Expected second entry to be evicted")
	}
}

func TestTokenCacheClear(t *testing.T) {
	ClearTokenCache()

	// Add some entries
	globalTokenCache.put(tokenCacheKey{token: style.TokenControlHeight, themeHash: 1}, style.Style{})
	globalTokenCache.put(tokenCacheKey{token: style.TokenControlRadius, themeHash: 1}, style.Style{})

	if globalTokenCache.size == 0 {
		t.Error("Expected non-zero cache size after putting entries")
	}

	// Clear and verify
	ClearTokenCache()

	if globalTokenCache.size != 0 {
		t.Errorf("Expected cache size 0 after clear, got %d", globalTokenCache.size)
	}

	stats := TokenCacheStats()
	if stats.Size != 0 {
		t.Errorf("Expected stats size 0 after clear, got %d", stats.Size)
	}
}

func TestTokenCacheThreadSafety(t *testing.T) {
	ClearTokenCache()

	// Test concurrent access
	done := make(chan bool)
	for i := range 10 {
		go func(id int) {
			key := tokenCacheKey{token: style.StyleToken(id % 5), themeHash: uint64(id % 3)}
			value := style.Style{}.Height(36)

			// Interleave puts and gets
			for range 100 {
				globalTokenCache.put(key, value)
				globalTokenCache.get(key)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for range 10 {
		<-done
	}

	// Just verify no panic occurred and cache is in valid state
	stats := TokenCacheStats()
	if stats.Size < 0 {
		t.Error("Cache in invalid state after concurrent access")
	}
}
