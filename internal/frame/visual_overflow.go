package frame

// VisualOutset is the pixel space a visual may paint outside its layout box.
// It is used by clipping containers to reserve room inside their viewport.
type VisualOutset struct {
	Top    int
	Right  int
	Bottom int
	Left   int
}

// Empty reports whether no visual space is required.
func (o VisualOutset) Empty() bool {
	return o.Top <= 0 && o.Right <= 0 && o.Bottom <= 0 && o.Left <= 0
}

// Max returns the per-edge maximum of two visual outsets.
func (o VisualOutset) Max(other VisualOutset) VisualOutset {
	return VisualOutset{
		Top:    max(o.Top, other.Top),
		Right:  max(o.Right, other.Right),
		Bottom: max(o.Bottom, other.Bottom),
		Left:   max(o.Left, other.Left),
	}
}

// VisualOverflowCollector gathers the largest visual overflow reported by a
// subtree. A clipping container owns one while it lays out its content.
type VisualOverflowCollector struct {
	outset VisualOutset
}

// Outset returns the largest overflow reported to the collector.
func (c *VisualOverflowCollector) Outset() VisualOutset {
	if c == nil {
		return VisualOutset{}
	}
	return c.outset
}

// Add records a visual overflow requirement.
func (c *VisualOverflowCollector) Add(outset VisualOutset) {
	if c == nil || outset.Empty() {
		return
	}
	c.outset = c.outset.Max(outset)
}

// PushVisualOverflowCollector makes collector receive visual overflow from
// descendants. Nested clipping containers use their own collector so each
// viewport reserves only for the content it directly clips.
func PushVisualOverflowCollector(ctx *Context, collector *VisualOverflowCollector) func() {
	if ctx == nil {
		return func() {}
	}
	previous := ctx.visualOverflowCollector
	ctx.visualOverflowCollector = collector
	return func() {
		ctx.visualOverflowCollector = previous
	}
}

// CollectingVisualOverflow reports whether a clipping container is currently
// collecting visual overflow from this subtree.
func CollectingVisualOverflow(ctx *Context) bool {
	return ctx != nil && ctx.hiddenLayoutDepth == 0 && ctx.visualOverflowCollector != nil
}

// ReportVisualOverflow reports a descendant's visual overflow to the nearest
// active clipping container.
func ReportVisualOverflow(ctx *Context, outset VisualOutset) {
	if !CollectingVisualOverflow(ctx) {
		return
	}
	ctx.visualOverflowCollector.Add(outset)
}
