package combobox

import (
	"fmt"
	"testing"
)

// Benchmark for #15 - visible items caching optimization

// BenchmarkVisibleItems_CacheHit benchmarks the cache hit path
func BenchmarkVisibleItems_CacheHit(b *testing.B) {
	s := &comboBoxState{}

	// Create a large dataset
	items := make([]ComboBoxItem, 1000)
	for i := range items {
		items[i] = ComboBoxItem{
			Key:   fmt.Sprintf("key%d", i),
			Label: fmt.Sprintf("Option %d", i),
		}
	}

	widget := ComboBox("bench", "key0", items)

	// Populate cache
	_ = s.visibleItems(widget, "", "Option 0")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Same query and selection should hit cache
		_ = s.visibleItems(widget, "", "Option 0")
	}
}

// BenchmarkVisibleItems_Filtering benchmarks the filtering logic
func BenchmarkVisibleItems_Filtering(b *testing.B) {
	s := &comboBoxState{}

	// Create a large dataset
	items := make([]ComboBoxItem, 1000)
	for i := range items {
		items[i] = ComboBoxItem{
			Key:   fmt.Sprintf("key%d", i),
			Label: fmt.Sprintf("Option %d", i),
		}
	}

	widget := ComboBox("bench", "key0", items).InputValue("Option 1")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Query that matches a subset (forces full scan and filtering)
		_ = s.visibleItems(widget, "Option 1", "Option 0")
	}
}

// BenchmarkVisibleItems_NoFilter benchmarks when all items are visible
func BenchmarkVisibleItems_NoFilter(b *testing.B) {
	s := &comboBoxState{}

	// Create a large dataset
	items := make([]ComboBoxItem, 1000)
	for i := range items {
		items[i] = ComboBoxItem{
			Key:   fmt.Sprintf("key%d", i),
			Label: fmt.Sprintf("Option %d", i),
		}
	}

	widget := ComboBox("bench", "key0", items)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Empty query shows all items
		_ = s.visibleItems(widget, "", "Option 0")
	}
}

// BenchmarkVisibleItems_CaseInsensitiveMatch benchmarks case-insensitive filtering
func BenchmarkVisibleItems_CaseInsensitiveMatch(b *testing.B) {
	s := &comboBoxState{}

	// Create mixed-case dataset
	items := make([]ComboBoxItem, 1000)
	for i := range items {
		items[i] = ComboBoxItem{
			Key:   fmt.Sprintf("key%d", i),
			Label: fmt.Sprintf("Option %d", i),
		}
	}

	widget := ComboBox("bench", "key0", items).InputValue("option")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Case-insensitive query (forces strings.ToLower on each comparison)
		_ = s.visibleItems(widget, "option", "Option 0")
	}
}
