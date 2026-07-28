package chart

import (
	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/theme"
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
