package core

import (
	"gioui.org/font"
	"gioui.org/io/system"
	"gioui.org/layout"
	giotext "gioui.org/text"
	"gioui.org/unit"
)

const textMeasureCacheCapacity = 256

// TextMeasureKey identifies every input that can change a non-interactive
// label's measured dimensions.
type TextMeasureKey struct {
	Text            string
	Shaper          *giotext.Shaper
	Font            font.Font
	TextSize        unit.Sp
	Alignment       giotext.Alignment
	MaxLines        int
	Truncator       string
	WrapPolicy      giotext.WrapPolicy
	LineHeight      unit.Sp
	LineHeightScale float32
	Constraints     layout.Constraints
	Metric          unit.Metric
	Locale          system.Locale
	ThemeGeneration uint64
}

type textMeasureEntry struct {
	dimensions layout.Dimensions
	lastUsed   uint64
}

type textMeasureCache struct {
	entries map[TextMeasureKey]textMeasureEntry
	clock   uint64
}

func (c *textMeasureCache) get(key TextMeasureKey) (layout.Dimensions, bool) {
	entry, ok := c.entries[key]
	if !ok {
		return layout.Dimensions{}, false
	}
	c.clock++
	entry.lastUsed = c.clock
	c.entries[key] = entry
	return entry.dimensions, true
}

func (c *textMeasureCache) put(key TextMeasureKey, dimensions layout.Dimensions) {
	if c.entries == nil {
		c.entries = make(map[TextMeasureKey]textMeasureEntry, textMeasureCacheCapacity)
	}
	if _, ok := c.entries[key]; !ok && len(c.entries) >= textMeasureCacheCapacity {
		var oldestKey TextMeasureKey
		var oldestUse uint64
		first := true
		for candidate, entry := range c.entries {
			if first || entry.lastUsed < oldestUse {
				oldestKey = candidate
				oldestUse = entry.lastUsed
				first = false
			}
		}
		delete(c.entries, oldestKey)
	}
	c.clock++
	c.entries[key] = textMeasureEntry{dimensions: dimensions, lastUsed: c.clock}
}

func (c *textMeasureCache) reset() {
	c.entries = nil
	c.clock = 0
}

// LookupTextMeasure returns a previously computed measurement for this
// context. It intentionally does not mark a state slot as used: measurements
// are derived render data, not component lifecycle state.
func LookupTextMeasure(ctx *Context, key TextMeasureKey) (layout.Dimensions, bool) {
	if ctx == nil {
		return layout.Dimensions{}, false
	}
	return ctx.textMeasureCache.get(key)
}

// StoreTextMeasure records a measurement for reuse by later frames.
func StoreTextMeasure(ctx *Context, key TextMeasureKey, dimensions layout.Dimensions) {
	if ctx == nil {
		return
	}
	ctx.textMeasureCache.put(key, dimensions)
}
