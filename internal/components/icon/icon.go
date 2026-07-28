package icon

import (
	"container/list"
	"image"
	"image/color"
	"sync"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/internal/frame"
)

const defaultSize = unit.Dp(24)

const lucideStrokeScale = 12

const (
	// The byte limit tracks encoded IconVG input; widget.Icon does not expose decoded size.
	iconCacheMaxEntries = 256
	iconCacheMaxBytes   = 4 << 20
)

type Widget struct {
	data     []byte
	size     unit.Dp
	color    color.NRGBA
	hasColor bool
}

type cacheKey struct {
	first  *byte
	length int
}

type renderer struct {
	mu   sync.Mutex
	icon *widget.Icon
}

type cacheEntry struct {
	renderer *renderer
	bytes    int
	elem     *list.Element
}

var renderers = struct {
	sync.Mutex
	entries map[cacheKey]*cacheEntry
	lru     list.List
	bytes   int
}{
	entries: make(map[cacheKey]*cacheEntry),
}

func New(data []byte) Widget {
	return Widget{data: data}
}

func (w Widget) Size(dp float32) Widget {
	w.size = unit.Dp(max(dp, 0))
	return w
}

func (w Widget) Color(col color.NRGBA) Widget {
	w.color = col
	w.hasColor = true
	return w
}

// LucideSizeForStroke returns the icon size that preserves Lucide's standard
// 2-unit stroke in its 24-unit viewBox.
func LucideSizeForStroke(gtx layout.Context, stroke unit.Dp) int {
	return max(gtx.Dp(stroke*lucideStrokeScale), lucideStrokeScale)
}

func (w Widget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	if w.data == nil {
		return layout.Dimensions{}
	}
	size := w.size
	if size == 0 {
		size = defaultSize
	}
	target := gtx.Dp(size)
	outerSize := gtx.Constraints.Constrain(image.Pt(target, target))
	diameter := min(outerSize.X, outerSize.Y)

	col := ctx.ForegroundColor()
	if w.hasColor {
		col = w.color
	}
	iconGtx := gtx
	iconGtx.Constraints = layout.Exact(image.Pt(diameter, diameter))
	offset := image.Pt((outerSize.X-diameter)/2, (outerSize.Y-diameter)/2)
	stack := op.Offset(offset).Push(gtx.Ops)
	Layout(w.data, iconGtx, col)
	stack.Pop()
	return layout.Dimensions{Size: outerSize}
}

// Layout renders IconVG data using the supplied constraints and color.
func Layout(data []byte, gtx layout.Context, col color.NRGBA) layout.Dimensions {
	if len(data) == 0 {
		return layout.Dimensions{}
	}
	key := cacheKey{first: &data[0], length: len(data)}
	renderers.Lock()
	if cached, ok := renderers.entries[key]; ok {
		renderers.lru.MoveToFront(cached.elem)
		valueRenderer := cached.renderer
		renderers.Unlock()
		valueRenderer.mu.Lock()
		defer valueRenderer.mu.Unlock()
		return valueRenderer.icon.Layout(gtx, col)
	}
	renderers.Unlock()

	resolved, err := widget.NewIcon(data)
	if err != nil {
		panic(err)
	}
	valueRenderer := &renderer{icon: resolved}

	renderers.Lock()
	if cached, ok := renderers.entries[key]; ok {
		renderers.lru.MoveToFront(cached.elem)
		valueRenderer = cached.renderer
	} else if len(data) <= iconCacheMaxBytes {
		for len(renderers.entries) >= iconCacheMaxEntries || renderers.bytes+len(data) > iconCacheMaxBytes {
			evictOldestLocked()
		}
		entry := &cacheEntry{renderer: valueRenderer, bytes: len(data)}
		entry.elem = renderers.lru.PushFront(key)
		renderers.entries[key] = entry
		renderers.bytes += entry.bytes
	}
	renderers.Unlock()

	valueRenderer.mu.Lock()
	defer valueRenderer.mu.Unlock()
	return valueRenderer.icon.Layout(gtx, col)
}

func evictOldestLocked() {
	oldest := renderers.lru.Back()
	if oldest == nil {
		return
	}
	key := oldest.Value.(cacheKey)
	entry := renderers.entries[key]
	delete(renderers.entries, key)
	renderers.bytes -= entry.bytes
	if renderers.bytes < 0 {
		renderers.bytes = 0
	}
	renderers.lru.Remove(oldest)
}
