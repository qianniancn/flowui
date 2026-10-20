package chart

import "testing"

func TestGeometryCacheReusesMatchingKey(t *testing.T) {
	var cache GeometryCache[string, int]
	resolutions := 0
	resolve := func() string {
		resolutions++
		return "geometry"
	}

	first := cache.Resolve(true, 1, resolve)
	second := cache.Resolve(true, 1, resolve)
	if first != second || resolutions != 1 {
		t.Fatalf("matching geometry key resolved %d times, want once", resolutions)
	}
}

func TestGeometryCacheReplacesChangedKey(t *testing.T) {
	var cache GeometryCache[int, int]
	resolutions := 0
	resolve := func() int {
		resolutions++
		return resolutions
	}

	if got := cache.Resolve(true, 1, resolve); got != 1 {
		t.Fatalf("first geometry = %d, want 1", got)
	}
	if got := cache.Resolve(true, 2, resolve); got != 2 {
		t.Fatalf("changed-key geometry = %d, want 2", got)
	}
	if resolutions != 2 {
		t.Fatalf("changed geometry key resolved %d times, want twice", resolutions)
	}
}

func TestGeometryCacheDoesNotRetainUnversionedResolution(t *testing.T) {
	var cache GeometryCache[int, int]
	resolutions := 0
	resolve := func() int {
		resolutions++
		return resolutions
	}

	if got := cache.Resolve(false, 1, resolve); got != 1 {
		t.Fatalf("first unversioned geometry = %d, want 1", got)
	}
	if got := cache.Resolve(false, 1, resolve); got != 2 {
		t.Fatalf("second unversioned geometry = %d, want 2", got)
	}
	if resolutions != 2 {
		t.Fatalf("unversioned geometry resolved %d times, want twice", resolutions)
	}
	if got := cache.Resolve(true, 1, resolve); got != 3 {
		t.Fatalf("versioned geometry after unversioned result = %d, want 3", got)
	}
}
