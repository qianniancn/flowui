package state

import (
	"testing"
)

func TestMemo_CacheHit(t *testing.T) {
	store := &Store{}
	store.BeginFrame()

	// First call - cache miss
	memo1 := UseMemo[int](store, "key1", "slot1", 1, 100)
	result := memo1.Compute(func() int {
		return 42
	})
	if result != 42 {
		t.Fatalf("first call = %d, want 42", result)
	}

	// Second call with same inputs - cache hit
	memo2 := UseMemo[int](store, "key1", "slot1", 1, 100)
	callCount := 0
	result = memo2.Compute(func() int {
		callCount++
		return 99 // Should not be called
	})
	if result != 42 {
		t.Fatalf("second call = %d, want 42 (cached)", result)
	}
	if callCount != 0 {
		t.Fatalf("compute function called %d times, want 0 (cache hit)", callCount)
	}
}

func TestMemo_InvalidateOnThemeChange(t *testing.T) {
	store := &Store{}
	store.BeginFrame()

	// Cache with theme gen 1
	memo1 := UseMemo[string](store, "key1", "slot1", 1, 200)
	result := memo1.Compute(func() string {
		return "theme1"
	})
	if result != "theme1" {
		t.Fatalf("first call = %q, want %q", result, "theme1")
	}

	// Theme changes to gen 2 - cache miss
	memo2 := UseMemo[string](store, "key1", "slot1", 2, 200)
	result = memo2.Compute(func() string {
		return "theme2"
	})
	if result != "theme2" {
		t.Fatalf("after theme change = %q, want %q", result, "theme2")
	}

	// Verify new value is cached
	memo3 := UseMemo[string](store, "key1", "slot1", 2, 200)
	callCount := 0
	result = memo3.Compute(func() string {
		callCount++
		return "theme3"
	})
	if result != "theme2" {
		t.Fatalf("third call = %q, want %q (cached)", result, "theme2")
	}
	if callCount != 0 {
		t.Fatalf("compute called %d times, want 0", callCount)
	}
}

func TestMemo_InvalidateOnInputChange(t *testing.T) {
	store := &Store{}
	store.BeginFrame()

	// Cache with input hash 300
	memo1 := UseMemo[int](store, "key1", "slot1", 1, 300)
	result := memo1.Compute(func() int {
		return 100
	})
	if result != 100 {
		t.Fatalf("first call = %d, want 100", result)
	}

	// Input changes to hash 400 - cache miss
	memo2 := UseMemo[int](store, "key1", "slot1", 1, 400)
	result = memo2.Compute(func() int {
		return 200
	})
	if result != 200 {
		t.Fatalf("after input change = %d, want 200", result)
	}
}

func TestMemo_DifferentKeys(t *testing.T) {
	store := &Store{}
	store.BeginFrame()

	// Two different keys should have independent caches
	memo1 := UseMemo[int](store, "key1", "slot1", 1, 500)
	result1 := memo1.Compute(func() int {
		return 111
	})

	memo2 := UseMemo[int](store, "key2", "slot1", 1, 500)
	result2 := memo2.Compute(func() int {
		return 222
	})

	if result1 != 111 {
		t.Fatalf("key1 result = %d, want 111", result1)
	}
	if result2 != 222 {
		t.Fatalf("key2 result = %d, want 222", result2)
	}

	// Verify independence
	memo1Again := UseMemo[int](store, "key1", "slot1", 1, 500)
	if r, ok := memo1Again.Get(); !ok || r != 111 {
		t.Fatalf("key1 cache = (%d, %v), want (111, true)", r, ok)
	}
}

func TestMemo_DifferentSlots(t *testing.T) {
	store := &Store{}
	store.BeginFrame()

	// Same key, different slots should have independent caches
	memo1 := UseMemo[int](store, "key1", "slot-a", 1, 600)
	result1 := memo1.Compute(func() int {
		return 333
	})

	memo2 := UseMemo[int](store, "key1", "slot-b", 1, 600)
	result2 := memo2.Compute(func() int {
		return 444
	})

	if result1 != 333 {
		t.Fatalf("slot-a result = %d, want 333", result1)
	}
	if result2 != 444 {
		t.Fatalf("slot-b result = %d, want 444", result2)
	}
}

func TestMemo_GetSet(t *testing.T) {
	store := &Store{}
	store.BeginFrame()

	memo := UseMemo[string](store, "key1", "slot1", 1, 700)

	// Initially empty
	if _, ok := memo.Get(); ok {
		t.Fatal("Get on empty cache returned true, want false")
	}

	// Set a value
	memo.Set("hello")

	// Get should now return it
	result, ok := memo.Get()
	if !ok {
		t.Fatal("Get after Set returned false, want true")
	}
	if result != "hello" {
		t.Fatalf("Get after Set = %q, want %q", result, "hello")
	}
}

func TestMemo_FrameCleanup(t *testing.T) {
	store := &Store{}

	// Frame 1: create and cache
	store.BeginFrame()
	memo1 := UseMemo[int](store, "key1", "slot1", 1, 800)
	memo1.Set(555)
	store.EndFrame()

	// Frame 2: key not used, should be cleaned up
	store.BeginFrame()
	store.EndFrame()

	// Frame 3: try to access - should be gone
	store.BeginFrame()
	memo3 := UseMemo[int](store, "key1", "slot1", 1, 800)
	if _, ok := memo3.Get(); ok {
		t.Fatal("Get after cleanup returned true, want false")
	}
}

func TestMemo_NilStore(t *testing.T) {
	// Should not panic with nil store
	memo := UseMemo[int](nil, "key1", "slot1", 1, 900)

	if _, ok := memo.Get(); ok {
		t.Fatal("Get with nil store returned true, want false")
	}

	memo.Set(123) // Should not panic

	result := memo.Compute(func() int {
		return 456
	})
	if result != 456 {
		t.Fatalf("Compute with nil store = %d, want 456", result)
	}
}

func TestMemo_StructType(t *testing.T) {
	type Result struct {
		Value int
		Label string
	}

	store := &Store{}
	store.BeginFrame()

	memo := UseMemo[Result](store, "key1", "slot1", 1, 1000)
	result := memo.Compute(func() Result {
		return Result{Value: 42, Label: "answer"}
	})

	if result.Value != 42 || result.Label != "answer" {
		t.Fatalf("struct result = %+v, want {42, answer}", result)
	}

	// Verify cached
	memo2 := UseMemo[Result](store, "key1", "slot1", 1, 1000)
	result2, ok := memo2.Get()
	if !ok {
		t.Fatal("Get struct from cache returned false, want true")
	}
	if result2.Value != 42 || result2.Label != "answer" {
		t.Fatalf("cached struct = %+v, want {42, answer}", result2)
	}
}

func BenchmarkMemo_Hit(b *testing.B) {
	store := &Store{}
	store.BeginFrame()

	memo := UseMemo[int](store, "key1", "slot1", 1, 1100)
	memo.Set(42)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		memo := UseMemo[int](store, "key1", "slot1", 1, 1100)
		_ = memo.Compute(func() int {
			return 99 // Should not be called
		})
	}
}

func BenchmarkMemo_Miss(b *testing.B) {
	store := &Store{}
	store.BeginFrame()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		memo := UseMemo[int](store, "key1", "slot1", 1, uint64(i))
		_ = memo.Compute(func() int {
			return i
		})
	}
}
