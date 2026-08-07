package runtime

import (
	"sync"

	"github.com/qianniancn/flowui/internal/style"
)

// cascadeCache provides LRU caching for resolved style cascade results.
// It reduces redundant cascade operations by memoizing ResolvedStyle outputs.
type cascadeCache struct {
	mu      sync.RWMutex
	entries map[cascadeCacheKey]*cascadeEntry
	head    *cascadeEntry
	tail    *cascadeEntry
	size    int
	maxSize int

	// Statistics
	hits      uint64
	misses    uint64
	evictions uint64
}

type cascadeCacheKey struct {
	contextPtr   uintptr // Pointer to frame.Context for per-context caching
	componentKey string
	stateHash    uint16
	themeHash    uint64
	customHash   uint64
}

type cascadeEntry struct {
	key   cascadeCacheKey
	value style.ResolvedStyle
	prev  *cascadeEntry
	next  *cascadeEntry
}

// Cache size chosen to handle typical UI scenarios:
// - 50-100 unique component × state combinations
// - 4-6 interaction states per component
// - Room for multiple themes
const defaultCascadeCacheSize = 512

var globalCascadeCache = newCascadeCache(defaultCascadeCacheSize)

func newCascadeCache(maxSize int) *cascadeCache {
	return &cascadeCache{
		entries: make(map[cascadeCacheKey]*cascadeEntry, maxSize),
		maxSize: maxSize,
	}
}

// get retrieves a cached resolved style.
// Returns (value, true) on hit, (zero, false) on miss.
// IMPORTANT: Returns a deep copy to prevent mutation of cached values by animate().
func (c *cascadeCache) get(key cascadeCacheKey) (style.ResolvedStyle, bool) {
	c.mu.RLock()
	entry, found := c.entries[key]
	c.mu.RUnlock()

	if !found {
		c.mu.Lock()
		c.misses++
		c.mu.Unlock()
		return style.ResolvedStyle{}, false
	}

	// Move to front (most recently used)
	c.mu.Lock()
	c.hits++
	c.moveToFront(entry)
	value := cloneResolvedStyle(entry.value)
	c.mu.Unlock()

	return value, true
}

// put stores a resolved style in the cache.
func (c *cascadeCache) put(key cascadeCacheKey, value style.ResolvedStyle) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if already exists
	if entry, found := c.entries[key]; found {
		entry.value = cloneResolvedStyle(value)
		c.moveToFront(entry)
		return
	}

	// Evict if at capacity
	if c.size >= c.maxSize {
		c.evictLRU()
	}

	// Create new entry
	entry := &cascadeEntry{
		key:   key,
		value: cloneResolvedStyle(value),
	}

	c.entries[key] = entry
	c.addToFront(entry)
	c.size++
}

// clear removes all entries from the cache.
func (c *cascadeCache) clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[cascadeCacheKey]*cascadeEntry, c.maxSize)
	c.head = nil
	c.tail = nil
	c.size = 0
}

// stats returns cache performance statistics.
func (c *cascadeCache) stats() CascadeCacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	total := c.hits + c.misses
	var hitRate float64
	if total > 0 {
		hitRate = float64(c.hits) / float64(total)
	}

	return CascadeCacheStats{
		Hits:      c.hits,
		Misses:    c.misses,
		Evictions: c.evictions,
		HitRate:   hitRate,
		Size:      c.size,
	}
}

// CascadeCacheStats provides visibility into cascade cache performance.
type CascadeCacheStats struct {
	Hits      uint64
	Misses    uint64
	Evictions uint64
	HitRate   float64
	Size      int
}

// moveToFront moves an entry to the front of the LRU list.
// Caller must hold c.mu.
func (c *cascadeCache) moveToFront(entry *cascadeEntry) {
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
func (c *cascadeCache) addToFront(entry *cascadeEntry) {
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
func (c *cascadeCache) evictLRU() {
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

// computeStyleStateHash converts StyleState to a compact hash.
func computeStyleStateHash(state style.StyleState) uint16 {
	var hash uint16
	if state.Hovered {
		hash |= 1 << 0
	}
	if state.Pressed {
		hash |= 1 << 1
	}
	if state.Focused {
		hash |= 1 << 2
	}
	if state.FocusVisible {
		hash |= 1 << 3
	}
	if state.Disabled {
		hash |= 1 << 4
	}
	if state.Selected {
		hash |= 1 << 5
	}
	if state.Checked {
		hash |= 1 << 6
	}
	if state.Indeterminate {
		hash |= 1 << 7
	}
	if state.ReadOnly {
		hash |= 1 << 8
	}
	if state.Invalid {
		hash |= 1 << 9
	}
	if state.Loading {
		hash |= 1 << 10
	}
	if state.Open {
		hash |= 1 << 11
	}
	if state.Expanded {
		hash |= 1 << 12
	}
	if state.Dragging {
		hash |= 1 << 13
	}
	if state.DropTarget {
		hash |= 1 << 14
	}
	return hash
}

// computeCustomStyleHash generates a hash of a custom style.
// This detects when custom styles change so cache can invalidate.
func (c *cascadeCache) computeCustomStyleHash(custom style.Style) uint64 {
	return custom.Hash64()
}

// ClearCascadeCache invalidates all cached cascade results.
// Call this when the theme changes.
func ClearCascadeCache() {
	globalCascadeCache.clear()
	// Also clear token cache for consistency
	ClearTokenCache()
}

// GetCascadeCacheStats returns performance metrics for the global cascade cache.
func GetCascadeCacheStats() CascadeCacheStats {
	return globalCascadeCache.stats()
}

// DisableCascadeCache is a feature flag to disable the cache for debugging.
// Set to true to bypass caching entirely.
var DisableCascadeCache = false

// cloneResolvedStyle creates a deep copy of a ResolvedStyle.
// This is critical for caching: animate() modifies ResolvedStyle in-place,
// so we must return copies to prevent mutation of cached values.
func cloneResolvedStyle(src style.ResolvedStyle) style.ResolvedStyle {
	result := style.ResolvedStyle{}

	// Clone Box
	if src.Box != nil {
		boxCopy := *src.Box
		if boxCopy.Padding != nil {
			paddingCopy := *boxCopy.Padding
			boxCopy.Padding = &paddingCopy
		}
		if boxCopy.Margin != nil {
			marginCopy := *boxCopy.Margin
			boxCopy.Margin = &marginCopy
		}
		if boxCopy.Width != nil {
			widthCopy := *boxCopy.Width
			boxCopy.Width = &widthCopy
		}
		if boxCopy.Height != nil {
			heightCopy := *boxCopy.Height
			boxCopy.Height = &heightCopy
		}
		if boxCopy.MinWidth != nil {
			minWidthCopy := *boxCopy.MinWidth
			boxCopy.MinWidth = &minWidthCopy
		}
		if boxCopy.MinHeight != nil {
			minHeightCopy := *boxCopy.MinHeight
			boxCopy.MinHeight = &minHeightCopy
		}
		if boxCopy.MaxWidth != nil {
			maxWidthCopy := *boxCopy.MaxWidth
			boxCopy.MaxWidth = &maxWidthCopy
		}
		if boxCopy.MaxHeight != nil {
			maxHeightCopy := *boxCopy.MaxHeight
			boxCopy.MaxHeight = &maxHeightCopy
		}
		if boxCopy.AspectRatio != nil {
			aspectRatioCopy := *boxCopy.AspectRatio
			boxCopy.AspectRatio = &aspectRatioCopy
		}
		if boxCopy.FillWidth != nil {
			fillWidthCopy := *boxCopy.FillWidth
			boxCopy.FillWidth = &fillWidthCopy
		}
		if boxCopy.FillHeight != nil {
			fillHeightCopy := *boxCopy.FillHeight
			boxCopy.FillHeight = &fillHeightCopy
		}
		if boxCopy.Overflow != nil {
			overflowCopy := *boxCopy.Overflow
			boxCopy.Overflow = &overflowCopy
		}
		if boxCopy.Cursor != nil {
			cursorCopy := *boxCopy.Cursor
			boxCopy.Cursor = &cursorCopy
		}
		result.Box = &boxCopy
	}

	// Clone Paint
	if src.Paint != nil {
		paintCopy := *src.Paint
		// Clone interface fields that animate() modifies
		paintCopy.Background = clonePaintSource(paintCopy.Background)
		if paintCopy.Border != nil {
			borderCopy := *paintCopy.Border
			borderCopy.Color = cloneColorSource(borderCopy.Color)
			paintCopy.Border = &borderCopy
		}
		if paintCopy.BorderBottom != nil {
			borderCopy := *paintCopy.BorderBottom
			borderCopy.Color = cloneColorSource(borderCopy.Color)
			if borderCopy.Width != nil {
				widthCopy := *borderCopy.Width
				borderCopy.Width = &widthCopy
			}
			paintCopy.BorderBottom = &borderCopy
		}
		if paintCopy.Outline != nil {
			outlineCopy := *paintCopy.Outline
			outlineCopy.Color = cloneColorSource(outlineCopy.Color)
			paintCopy.Outline = &outlineCopy
		}
		if paintCopy.VisualOutset != nil {
			outsetCopy := *paintCopy.VisualOutset
			paintCopy.VisualOutset = &outsetCopy
		}
		if paintCopy.Opacity != nil {
			opacityCopy := *paintCopy.Opacity
			paintCopy.Opacity = &opacityCopy
		}
		if paintCopy.Radius != nil {
			radiusCopy := *paintCopy.Radius
			paintCopy.Radius = &radiusCopy
		}
		if paintCopy.Radii != nil {
			radiiCopy := *paintCopy.Radii
			paintCopy.Radii = &radiiCopy
		}
		if len(paintCopy.Shadows) > 0 {
			paintCopy.Shadows = append([]style.Shadow(nil), paintCopy.Shadows...)
		}
		result.Paint = &paintCopy
	}

	// Clone Text
	if src.Text != nil {
		textCopy := *src.Text
		// Clone Color interface that animate() modifies
		textCopy.Color = cloneColorSource(textCopy.Color)
		if textCopy.FontSize != nil {
			fontSizeCopy := *textCopy.FontSize
			textCopy.FontSize = &fontSizeCopy
		}
		if textCopy.FontWeight != nil {
			fontWeightCopy := *textCopy.FontWeight
			textCopy.FontWeight = &fontWeightCopy
		}
		if textCopy.LineHeight != nil {
			lineHeightCopy := *textCopy.LineHeight
			textCopy.LineHeight = &lineHeightCopy
		}
		if textCopy.LineHeightScale != nil {
			lineHeightScaleCopy := *textCopy.LineHeightScale
			textCopy.LineHeightScale = &lineHeightScaleCopy
		}
		if textCopy.MaxLines != nil {
			maxLinesCopy := *textCopy.MaxLines
			textCopy.MaxLines = &maxLinesCopy
		}
		if textCopy.Align != nil {
			alignCopy := *textCopy.Align
			textCopy.Align = &alignCopy
		}
		if textCopy.Wrap != nil {
			wrapCopy := *textCopy.Wrap
			textCopy.Wrap = &wrapCopy
		}
		if textCopy.Truncator != nil {
			truncatorCopy := *textCopy.Truncator
			textCopy.Truncator = &truncatorCopy
		}
		if textCopy.Typeface != nil {
			typefaceCopy := *textCopy.Typeface
			textCopy.Typeface = &typefaceCopy
		}
		if textCopy.FontStyle != nil {
			fontStyleCopy := *textCopy.FontStyle
			textCopy.FontStyle = &fontStyleCopy
		}
		result.Text = &textCopy
	}

	// Clone Transform
	if src.Trans != nil {
		transCopy := *src.Trans
		if transCopy.TranslateX != nil {
			translateXCopy := *transCopy.TranslateX
			transCopy.TranslateX = &translateXCopy
		}
		if transCopy.TranslateY != nil {
			translateYCopy := *transCopy.TranslateY
			transCopy.TranslateY = &translateYCopy
		}
		if transCopy.ScaleX != nil {
			scaleXCopy := *transCopy.ScaleX
			transCopy.ScaleX = &scaleXCopy
		}
		if transCopy.ScaleY != nil {
			scaleYCopy := *transCopy.ScaleY
			transCopy.ScaleY = &scaleYCopy
		}
		if transCopy.Rotate != nil {
			rotateCopy := *transCopy.Rotate
			transCopy.Rotate = &rotateCopy
		}
		result.Trans = &transCopy
	}

	// Clone Transitions (shallow copy is sufficient - they're immutable)
	if len(src.Transitions) > 0 {
		result.Transitions = append([]style.Transition(nil), src.Transitions...)
	}

	return result
}

// clonePaintSource creates a copy of a PaintSource interface value.
func clonePaintSource(src style.PaintSource) style.PaintSource {
	if src == nil {
		return nil
	}
	switch v := src.(type) {
	case style.SolidColor:
		return v // value type, already a copy
	case *style.SolidColor:
		if v == nil {
			return nil
		}
		copy := *v
		return &copy
	case style.StyleGradient:
		// Shallow copy is sufficient - gradients are treated as immutable after resolution
		return v
	default:
		// For unknown types, return as-is (may need extension for new types)
		return src
	}
}

// cloneColorSource creates a copy of a ColorSource interface value.
func cloneColorSource(src style.ColorSource) style.ColorSource {
	if src == nil {
		return nil
	}
	switch v := src.(type) {
	case style.SolidColor:
		return v // value type, already a copy
	case *style.SolidColor:
		if v == nil {
			return nil
		}
		copy := *v
		return &copy
	default:
		// For unknown types, return as-is
		return src
	}
}
