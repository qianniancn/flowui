package chart

import (
	"reflect"

	"gioui.org/unit"
	"github.com/qianniancn/flowui/theme"
)

// DataCache caches resolved chart data by widget version, theme, and metric.
// A zero generation identifies an uncached resolution.
type DataCache[T any] struct {
	ready      bool
	version    uint64
	generation uint64
	theme      *theme.Theme
	metric     unit.Metric
	data       T
}

func (c *DataCache[T]) Resolve(hasVersion bool, version uint64, activeTheme *theme.Theme, metric unit.Metric, resolve func() T) (T, uint64) {
	if !hasVersion {
		*c = DataCache[T]{}
		return resolve(), 0
	}
	if c.ready && c.version == version && c.theme == activeTheme && c.metric == metric {
		return c.data, c.generation
	}
	c.generation++
	if c.generation == 0 {
		c.generation = 1
	}
	c.data = resolve()
	c.ready = true
	c.version = version
	c.theme = activeTheme
	c.metric = metric
	return c.data, c.generation
}

// GeometryCache retains the latest resolved geometry for one widget instance.
// Geometry is intentionally single-valued: viewport changes replace the old
// result instead of retaining an unbounded history of layouts.
type GeometryCache[T any, K comparable] struct {
	ready bool
	key   K
	value T
}

// Resolve returns the cached geometry when cacheable and key is unchanged.
// Uncacheable resolutions clear the previous value so transient data is never
// kept alive by the widget state.
func (c *GeometryCache[T, K]) Resolve(cacheable bool, key K, resolve func() T) T {
	if !cacheable {
		*c = GeometryCache[T, K]{}
		return resolve()
	}
	if c.ready && c.key == key {
		return c.value
	}
	c.value = resolve()
	c.key = key
	c.ready = true
	return c.value
}

// FuncID returns a stable code identity for a formatter callback. It is used
// only as a cache key; callers still need to advance DataVersion when a
// closure's captured values change.
func FuncID(fn any) uintptr {
	if fn == nil {
		return 0
	}
	value := reflect.ValueOf(fn)
	if value.Kind() != reflect.Func || value.IsNil() {
		return 0
	}
	return value.Pointer()
}
