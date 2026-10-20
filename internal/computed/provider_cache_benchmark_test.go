package computed

import (
	"testing"

	"github.com/qianniancn/flowui/theme"
)

func BenchmarkProviderCacheHit(b *testing.B) {
	active := theme.NeutralTheme()
	cache := ProviderCache[int, string]{}
	load := func() string { return "value" }
	cache.Resolve(&active, 0, load)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		cache.Resolve(&active, 0, load)
	}
	b.StopTimer()
	reportProviderCacheStats(b, cache.Stats())
}

func BenchmarkProviderCacheMiss(b *testing.B) {
	active := theme.NeutralTheme()
	cache := ProviderCache[int, string]{}
	load := func() string { return "value" }

	b.ReportAllocs()
	b.ResetTimer()
	for key := 0; b.Loop(); key++ {
		cache.Resolve(&active, key, load)
	}
	b.StopTimer()
	reportProviderCacheStats(b, cache.Stats())
}

func BenchmarkProviderCacheThemeInvalidation(b *testing.B) {
	first := theme.NeutralTheme()
	second := theme.NeutralTheme()
	cache := ProviderCache[int, string]{}
	load := func() string { return "value" }
	cache.Resolve(&first, 0, load)

	b.ReportAllocs()
	b.ResetTimer()
	for iteration := 0; b.Loop(); iteration++ {
		active := &first
		if iteration%2 == 1 {
			active = &second
		}
		cache.Resolve(active, 0, load)
	}
	b.StopTimer()
	reportProviderCacheStats(b, cache.Stats())
}

func BenchmarkProviderCacheMapHit(b *testing.B) {
	active := theme.NeutralTheme()
	cache := ProviderCacheMap[int, string]{}
	load := func() string { return "value" }
	cache.Resolve(&active, 0, load)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		cache.Resolve(&active, 0, load)
	}
	b.StopTimer()
	reportProviderCacheMapStats(b, cache.Stats())
}

func BenchmarkProviderCacheMapMiss(b *testing.B) {
	active := theme.NeutralTheme()
	cache := ProviderCacheMap[int, string]{}
	load := func() string { return "value" }

	b.ReportAllocs()
	b.ResetTimer()
	for key := 0; b.Loop(); key++ {
		cache.Resolve(&active, key, load)
	}
	b.StopTimer()
	reportProviderCacheMapStats(b, cache.Stats())
}

func BenchmarkProviderCacheMapThemeInvalidation(b *testing.B) {
	first := theme.NeutralTheme()
	second := theme.NeutralTheme()
	cache := ProviderCacheMap[int, string]{}
	load := func() string { return "value" }
	cache.Resolve(&first, 0, load)

	b.ReportAllocs()
	b.ResetTimer()
	for iteration := 0; b.Loop(); iteration++ {
		active := &first
		if iteration%2 == 1 {
			active = &second
		}
		cache.Resolve(active, 0, load)
	}
	b.StopTimer()
	reportProviderCacheMapStats(b, cache.Stats())
}

func reportProviderCacheStats(b *testing.B, stats ProviderCacheStats) {
	b.Helper()
	denominator := float64(b.N)
	b.ReportMetric(float64(stats.Hits)/denominator, "provider-cache-hits/op")
	b.ReportMetric(float64(stats.Misses)/denominator, "provider-cache-misses/op")
	b.ReportMetric(float64(stats.ThemeInvalidations)/denominator, "provider-cache-theme-invalidations/op")
	b.ReportMetric(float64(stats.KeyInvalidations)/denominator, "provider-cache-key-invalidations/op")
}

func reportProviderCacheMapStats(b *testing.B, stats ProviderCacheStats) {
	b.Helper()
	denominator := float64(b.N)
	b.ReportMetric(float64(stats.Hits)/denominator, "provider-cache-map-hits/op")
	b.ReportMetric(float64(stats.Misses)/denominator, "provider-cache-map-misses/op")
	b.ReportMetric(float64(stats.ThemeInvalidations)/denominator, "provider-cache-map-theme-invalidations/op")
	b.ReportMetric(float64(stats.KeyInvalidations)/denominator, "provider-cache-map-key-invalidations/op")
}
