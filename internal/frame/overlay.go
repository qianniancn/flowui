package frame

import (
	"fmt"
	"image"
	"math"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
)

// OverlayLayer controls the root-level stacking group for an overlay.
type OverlayLayer uint8

const (
	OverlayLayerPopup OverlayLayer = iota
	OverlayLayerModal
)

// OverlayRequest describes root-level work registered while laying out a
// component. Layout receives the anchor in viewport coordinates and whether
// this overlay owned input in the preceding frame.
type OverlayRequest struct {
	Key       string
	Layer     OverlayLayer
	Anchor    image.Rectangle
	HasAnchor bool
	Disabled  bool
	Layout    func(layout.Context, image.Rectangle, bool) layout.Dimensions
	// Tail records a focus boundary for the nearest topmost overlay ancestor.
	// It runs after all overlay layouts and must not register another overlay.
	Tail func(layout.Context)
}

type overlayIdentity struct {
	key   string
	layer OverlayLayer
}

type overlayRequest struct {
	OverlayRequest
	identity  overlayIdentity
	host      *overlayHost
	transform int
	scope     []string
	group     OverlayLayer
	order     uint64
	rootOrder uint64
	parent    *overlayRequest
	rendered  bool
}

type overlayTransform struct {
	parent  int
	local   f32.Affine2D
	opacity float32
	placed  bool
	clip    image.Rectangle
	hasClip bool
}

type overlayHost struct {
	transforms      []overlayTransform
	current         int
	requests        []*overlayRequest
	orders          map[overlayIdentity]uint64
	seen            map[overlayIdentity]struct{}
	nextOrder       uint64
	previousTop     overlayIdentity
	hasPreviousTop  bool
	eventTop        overlayIdentity
	hasEventTop     bool
	becameTop       overlayIdentity
	hasBecameTop    bool
	previousTail    overlayIdentity
	hasPreviousTail bool
	eventTail       overlayIdentity
	hasEventTail    bool
	becameTail      overlayIdentity
	hasBecameTail   bool
	active          *overlayRequest
	inTail          bool
	afterLayout     []func()
}

const invalidOverlayTransform = -1

func (h *overlayHost) beginFrame() {
	h.eventTop = h.previousTop
	h.hasEventTop = h.hasPreviousTop
	h.eventTail = h.previousTail
	h.hasEventTail = h.hasPreviousTail
	clear(h.requests)
	h.requests = h.requests[:0]
	h.resetTransforms()
	clear(h.afterLayout)
	h.afterLayout = h.afterLayout[:0]
	h.active = nil
	h.inTail = false
	h.hasBecameTop = false
	h.hasBecameTail = false
	if h.seen == nil {
		h.seen = make(map[overlayIdentity]struct{})
	} else {
		clear(h.seen)
	}
}

// RegisterOverlay adds an overlay to the root host for the current frame.
func RegisterOverlay(ctx *Context, request OverlayRequest) {
	if request.Key == "" {
		panic("flowui: empty overlay key")
	}
	if request.Layout == nil {
		panic("flowui: nil overlay layout")
	}
	identity := overlayIdentity{key: request.Key, layer: request.Layer}
	host := &ctx.overlays
	if host.inTail {
		panic("flowui: overlay tail cannot register another overlay")
	}
	host.ensureRootTransform()
	if host.seen == nil {
		host.seen = make(map[overlayIdentity]struct{})
	}
	if _, exists := host.seen[identity]; exists {
		panic(fmt.Sprintf("flowui: duplicate overlay key %q", request.Key))
	}
	host.seen[identity] = struct{}{}
	if host.orders == nil {
		host.orders = make(map[overlayIdentity]uint64)
	}
	order, exists := host.orders[identity]
	if !exists {
		host.nextOrder++
		order = host.nextOrder
		host.orders[identity] = order
	}
	group := request.Layer
	rootOrder := order
	if host.active != nil && host.active.group > group {
		group = host.active.group
	}
	if host.active != nil {
		rootOrder = host.active.rootOrder
	}
	host.requests = append(host.requests, &overlayRequest{
		OverlayRequest: request,
		identity:       identity,
		host:           host,
		transform:      host.current,
		scope:          ctx.keys.Scope(),
		group:          group,
		order:          order,
		rootOrder:      rootOrder,
		parent:         host.active,
	})
}

// OverlayInteractive reports whether key owned overlay input in the preceding
// frame. With no preceding overlay, it returns true for compatibility with
// direct component layout tests and first-frame rendering.
func OverlayInteractive(ctx *Context, layer OverlayLayer, key string) bool {
	host := &ctx.overlays
	return !host.hasEventTop || host.eventTop == (overlayIdentity{key: key, layer: layer})
}

// OverlayTopmost reports whether key is the topmost overlay from the most
// recently completed overlay layout. It is intended for AfterOverlays work.
func OverlayTopmost(ctx *Context, layer OverlayLayer, key string) bool {
	host := &ctx.overlays
	return host.hasPreviousTop && host.previousTop == (overlayIdentity{key: key, layer: layer})
}

// HasTopOverlay reports whether the most recently completed overlay layout
// produced a visible overlay.
func HasTopOverlay(ctx *Context) bool {
	return ctx.overlays.hasPreviousTop
}

// OverlayBecameTopmost reports whether key became the topmost overlay in the
// overlay layout that just completed. It is intended for AfterOverlays work;
// input ownership does not transfer until the next frame.
func OverlayBecameTopmost(ctx *Context, layer OverlayLayer, key string) bool {
	host := &ctx.overlays
	return host.hasBecameTop && host.becameTop == (overlayIdentity{key: key, layer: layer})
}

// OverlayFocusScopeTopmost reports whether key owns the focus boundary for the
// current topmost overlay ancestry.
func OverlayFocusScopeTopmost(ctx *Context, layer OverlayLayer, key string) bool {
	host := &ctx.overlays
	return host.hasPreviousTail && host.previousTail == (overlayIdentity{key: key, layer: layer})
}

// OverlayFocusScopeBecameTopmost reports whether key became the active focus
// boundary during the overlay layout that just completed.
func OverlayFocusScopeBecameTopmost(ctx *Context, layer OverlayLayer, key string) bool {
	host := &ctx.overlays
	return host.hasBecameTail && host.becameTail == (overlayIdentity{key: key, layer: layer})
}

// OverlayNaturallyDisabled reports the disabled state inherited from the
// registration site.
func OverlayNaturallyDisabled(gtx layout.Context) bool {
	return !gtx.Enabled()
}

// AfterOverlays schedules state cleanup after all dynamically registered
// overlays have completed their layout.
func AfterOverlays(ctx *Context, fn func()) {
	if fn != nil {
		ctx.overlays.afterLayout = append(ctx.overlays.afterLayout, fn)
	}
}

// LayoutOverlays resolves anchors and records every overlay at the root. The
// resulting macro is deferred once so it is painted after deferred work from
// the main widget tree.
func LayoutOverlays(ctx *Context, gtx layout.Context) {
	host := &ctx.overlays
	viewport := OverlayViewport(ctx, gtx.Constraints.Max)
	if viewport.X <= 0 || viewport.Y <= 0 {
		host.finishFrame(overlayIdentity{}, false, nil)
		return
	}
	rootGtx := gtx
	rootGtx.Constraints = layout.Exact(viewport)
	macro := op.Record(gtx.Ops)

	var top overlayIdentity
	hasTop := false
	var topRequest *overlayRequest
	for {
		request := host.nextResolvable(viewport)
		if request == nil {
			break
		}
		anchor, _, _, inheritedOpacity := resolveOverlayAnchor(request)
		request.rendered = true
		interactive := !host.hasEventTop || host.eventTop == request.identity
		func() {
			host.active = request
			previousCurrent := host.current
			requestRoot := host.appendTransform(overlayTransform{
				parent:  invalidOverlayTransform,
				local:   f32.AffineId(),
				opacity: inheritedOpacity,
				placed:  true,
			})
			host.current = requestRoot
			restoreScope := ctx.keys.UseScope(request.scope)
			defer func() {
				restoreScope()
				host.current = previousCurrent
				host.active = nil
			}()
			requestGtx := rootGtx
			if request.Disabled {
				requestGtx = requestGtx.Disabled()
			}
			if inheritedOpacity < 1 {
				opacity := paint.PushOpacity(requestGtx.Ops, inheritedOpacity)
				request.Layout(requestGtx, anchor, interactive)
				opacity.Pop()
			} else {
				request.Layout(requestGtx, anchor, interactive)
			}
		}()
		top = request.identity
		topRequest = request
		hasTop = true
	}
	tail := topOverlayTail(topRequest)
	if tail != nil {
		host.layoutTail(ctx, rootGtx, tail)
	}

	op.Defer(gtx.Ops, macro.Stop())
	host.finishFrame(top, hasTop, tail)
}

func topOverlayTail(request *overlayRequest) *overlayRequest {
	for request != nil {
		if request.Tail != nil {
			return request
		}
		request = request.parent
	}
	return nil
}

func (h *overlayHost) layoutTail(ctx *Context, rootGtx layout.Context, request *overlayRequest) {
	if request.Tail == nil {
		return
	}
	_, _, _, inheritedOpacity := resolveOverlayAnchor(request)
	requestGtx := rootGtx
	if request.Disabled {
		requestGtx = requestGtx.Disabled()
	}
	restoreScope := ctx.keys.UseScope(request.scope)
	defer restoreScope()
	h.inTail = true
	defer func() { h.inTail = false }()
	if inheritedOpacity < 1 {
		opacity := paint.PushOpacity(requestGtx.Ops, inheritedOpacity)
		request.Tail(requestGtx)
		opacity.Pop()
		return
	}
	request.Tail(requestGtx)
}

func (h *overlayHost) nextResolvable(viewport image.Point) *overlayRequest {
	var candidate *overlayRequest
	for _, request := range h.requests {
		if request.rendered {
			continue
		}
		_, ok, visible, _ := resolveOverlayAnchor(request)
		if !ok {
			continue
		}
		if request.HasAnchor && visible.Intersect(image.Rectangle{Max: viewport}).Empty() {
			request.rendered = true
			continue
		}
		if candidate == nil || overlayRequestLess(request, candidate) {
			candidate = request
		}
	}
	return candidate
}

func overlayRequestLess(a, b *overlayRequest) bool {
	if a.group != b.group {
		return a.group < b.group
	}
	if a.rootOrder != b.rootOrder {
		return a.rootOrder < b.rootOrder
	}
	// An owner is rendered before it can register descendants. Layer priority
	// therefore applies only among pending work in the same root branch.
	if a.Layer != b.Layer {
		return a.Layer < b.Layer
	}
	return a.order < b.order
}

func (h *overlayHost) finishFrame(top overlayIdentity, hasTop bool, tail *overlayRequest) {
	h.hasBecameTop = hasTop && (!h.hasEventTop || h.eventTop != top)
	if h.hasBecameTop {
		h.becameTop = top
	} else {
		h.becameTop = overlayIdentity{}
	}
	h.previousTop = top
	h.hasPreviousTop = hasTop
	h.hasBecameTail = tail != nil && (!h.hasEventTail || h.eventTail != tail.identity)
	if h.hasBecameTail {
		h.becameTail = tail.identity
	} else {
		h.becameTail = overlayIdentity{}
	}
	if tail != nil {
		h.previousTail = tail.identity
		h.hasPreviousTail = true
	} else {
		h.previousTail = overlayIdentity{}
		h.hasPreviousTail = false
	}
	for identity := range h.orders {
		if _, ok := h.seen[identity]; !ok {
			delete(h.orders, identity)
		}
	}
	h.runAfterLayout()
}

func (h *overlayHost) runAfterLayout() {
	callbacks := h.afterLayout
	h.afterLayout = nil
	for _, callback := range callbacks {
		callback()
	}
}

func resolveOverlayAnchor(request *overlayRequest) (image.Rectangle, bool, image.Rectangle, float32) {
	if request == nil || request.host == nil {
		return image.Rectangle{}, false, image.Rectangle{}, 1
	}
	if !request.HasAnchor {
		_, _, opacity, ok := resolveOverlayTransform(request.host, request.transform)
		return image.Rectangle{}, ok, image.Rectangle{}, opacity
	}
	transform, clips, opacity, ok := resolveOverlayTransform(request.host, request.transform)
	if !ok {
		return image.Rectangle{}, false, image.Rectangle{}, opacity
	}
	anchor := transformRectangle(transform, request.Anchor)
	visible := anchor
	for _, clip := range clips {
		visible = visible.Intersect(clip)
		if visible.Empty() {
			return anchor, true, image.Rectangle{}, opacity
		}
	}
	return anchor, true, visible, opacity
}

func resolveOverlayTransform(host *overlayHost, index int) (f32.Affine2D, []image.Rectangle, float32, bool) {
	if index == invalidOverlayTransform {
		return f32.AffineId(), nil, 1, true
	}
	if host == nil || index < 0 || index >= len(host.transforms) {
		return f32.Affine2D{}, nil, 1, false
	}
	chain := make([]int, 0, 8)
	for current := index; current != invalidOverlayTransform; current = host.transforms[current].parent {
		if current < 0 || current >= len(host.transforms) {
			return f32.Affine2D{}, nil, 1, false
		}
		if !host.transforms[current].placed {
			return f32.Affine2D{}, nil, 1, false
		}
		chain = append(chain, current)
	}
	transform := f32.AffineId()
	opacity := float32(1)
	clips := make([]image.Rectangle, 0, 2)
	for i := len(chain) - 1; i >= 0; i-- {
		node := host.transforms[chain[i]]
		if node.hasClip {
			clips = append(clips, transformRectangle(transform, node.clip))
		}
		transform = transform.Mul(node.local)
		opacity *= node.opacity
	}
	return transform, clips, opacity, true
}

func transformRectangle(transform f32.Affine2D, rect image.Rectangle) image.Rectangle {
	points := [4]f32.Point{
		transform.Transform(f32.Pt(float32(rect.Min.X), float32(rect.Min.Y))),
		transform.Transform(f32.Pt(float32(rect.Max.X), float32(rect.Min.Y))),
		transform.Transform(f32.Pt(float32(rect.Max.X), float32(rect.Max.Y))),
		transform.Transform(f32.Pt(float32(rect.Min.X), float32(rect.Max.Y))),
	}
	minX, maxX := points[0].X, points[0].X
	minY, maxY := points[0].Y, points[0].Y
	for _, point := range points[1:] {
		minX = min(minX, point.X)
		maxX = max(maxX, point.X)
		minY = min(minY, point.Y)
		maxY = max(maxY, point.Y)
	}
	return image.Rect(
		int(math.Floor(float64(minX))),
		int(math.Floor(float64(minY))),
		int(math.Ceil(float64(maxX))),
		int(math.Ceil(float64(maxY))),
	)
}

// OverlayPlacement tracks the transform applied to a measured child after its
// final position becomes known. Placements are valid only for the current
// frame.
type OverlayPlacement struct {
	host  *overlayHost
	index int
}

// TrackOverlayPlacement lays out child under a mutable transform node.
func TrackOverlayPlacement(ctx *Context, child func() layout.Dimensions) (layout.Dimensions, OverlayPlacement) {
	host := &ctx.overlays
	host.ensureRootTransform()
	parent := host.current
	node := host.appendTransform(overlayTransform{
		parent:  parent,
		local:   f32.AffineId(),
		opacity: 1,
	})
	host.current = node
	dims := func() layout.Dimensions {
		defer func() { host.current = parent }()
		return child()
	}()
	return dims, OverlayPlacement{host: host, index: node}
}

// PlaceOffset sets the child's local-to-parent translation.
func (p OverlayPlacement) PlaceOffset(offset image.Point) {
	p.PlaceTransform(f32.AffineId().Offset(f32.Pt(float32(offset.X), float32(offset.Y))))
}

// PlaceTransform sets the child's local-to-parent affine transform.
func (p OverlayPlacement) PlaceTransform(transform f32.Affine2D) {
	node := p.node()
	if node == nil {
		return
	}
	node.local = transform
	node.placed = true
}

// SetOpacity sets the opacity inherited by overlays registered inside the
// tracked child. The child's own paint operations remain the caller's
// responsibility.
func (p OverlayPlacement) SetOpacity(opacity float32) {
	node := p.node()
	if node == nil {
		return
	}
	node.opacity = min(max(opacity, 0), 1)
}

// ClipTo limits overlay visibility to a rectangle in the placement's parent
// coordinate space. The full anchor is retained for positioning when any part
// of it remains visible.
func (p OverlayPlacement) ClipTo(rect image.Rectangle) {
	node := p.node()
	if node == nil {
		return
	}
	node.clip = rect
	node.hasClip = true
}

func (p OverlayPlacement) node() *overlayTransform {
	if p.host == nil || p.index < 0 || p.index >= len(p.host.transforms) {
		return nil
	}
	return &p.host.transforms[p.index]
}

func (h *overlayHost) resetTransforms() {
	root := overlayTransform{
		parent:  invalidOverlayTransform,
		local:   f32.AffineId(),
		opacity: 1,
		placed:  true,
	}
	if len(h.transforms) == 0 {
		h.transforms = append(h.transforms, root)
	} else {
		h.transforms = h.transforms[:1]
		h.transforms[0] = root
	}
	h.current = 0
}

func (h *overlayHost) ensureRootTransform() {
	if len(h.transforms) == 0 {
		h.resetTransforms()
	}
}

func (h *overlayHost) appendTransform(transform overlayTransform) int {
	index := len(h.transforms)
	h.transforms = append(h.transforms, transform)
	return index
}
