// SPDX-License-Identifier: Unlicense OR MIT

package render

import (
	"image"
	"image/color"
	"math"
	"sync"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/theme"
)

// ShadowShapeKind identifies the geometry used for a shadow.
type ShadowShapeKind uint8

const (
	// ShadowRoundedRect draws a rounded rectangle shadow.
	ShadowRoundedRect ShadowShapeKind = iota
	// ShadowEllipse draws an ellipse shadow.
	ShadowEllipse
)

// ShadowCornerRadii stores independent rounded-rectangle corner radii.
type ShadowCornerRadii struct {
	NW unit.Dp
	NE unit.Dp
	SE unit.Dp
	SW unit.Dp
}

// ShadowShape describes the surface casting a shadow.
type ShadowShape struct {
	Kind  ShadowShapeKind
	Radii ShadowCornerRadii
}

// RoundedShadowCorners returns a non-uniform rounded rectangle shadow shape.
func RoundedShadowCorners(nw, ne, se, sw unit.Dp) ShadowShape {
	return ShadowShape{
		Kind: ShadowRoundedRect,
		Radii: ShadowCornerRadii{
			NW: nw,
			NE: ne,
			SE: se,
			SW: sw,
		},
	}
}

// EllipseShadow returns an ellipse shadow shape.
func EllipseShadow() ShadowShape {
	return ShadowShape{Kind: ShadowEllipse}
}

func (s ShadowShape) pixels(gtx layout.Context, size image.Point) shadowShape {
	shape := shadowShape{Kind: s.Kind}
	if shape.Kind != ShadowEllipse {
		shape.Kind = ShadowRoundedRect
		shape.Radii = normalizeCornerRadii(image.Rectangle{Max: size}, cornerRadiiPx{
			NW: gtx.Dp(s.Radii.NW),
			NE: gtx.Dp(s.Radii.NE),
			SE: gtx.Dp(s.Radii.SE),
			SW: gtx.Dp(s.Radii.SW),
		})
	}
	return shape
}

// ClipOp returns a clip operation matching the shadow shape.
func (s ShadowShape) ClipOp(gtx layout.Context, rect image.Rectangle) clip.Op {
	shape := s.pixels(gtx, rect.Size())
	if shape.Kind == ShadowEllipse {
		return clip.Ellipse(rect).Op(gtx.Ops)
	}
	r := shape.Radii
	return clip.RRect{
		Rect: rect,
		NW:   r.NW,
		NE:   r.NE,
		SE:   r.SE,
		SW:   r.SW,
	}.Op(gtx.Ops)
}

// ShadowLayer is one soft shadow pass.
type ShadowLayer struct {
	OffsetX float32
	OffsetY float32
	Blur    float32
	Spread  float32
	Color   color.NRGBA
}

// BoxShadow describes a CSS-like shadow. Layers, when present, take precedence over
// the single shadow fields.
type BoxShadow struct {
	OffsetX float32
	OffsetY float32
	Blur    float32
	Spread  float32
	Color   color.NRGBA
	Layers  []ShadowLayer
}

// EffectiveLayers returns the drawable layers for a box shadow.
func (b BoxShadow) EffectiveLayers() []ShadowLayer {
	if len(b.Layers) > 0 {
		layers := make([]ShadowLayer, 0, len(b.Layers))
		for _, layer := range b.Layers {
			if layer.Blur >= 0 && layer.Color.A > 0 {
				layers = append(layers, layer)
			}
		}
		return layers
	}
	if b.Blur < 0 || b.Color.A == 0 {
		return nil
	}
	return []ShadowLayer{{
		OffsetX: b.OffsetX,
		OffsetY: b.OffsetY,
		Blur:    b.Blur,
		Spread:  b.Spread,
		Color:   b.Color,
	}}
}

func (b BoxShadow) forEachLayer(fn func(ShadowLayer)) {
	if len(b.Layers) > 0 {
		for _, layer := range b.Layers {
			if layer.Blur >= 0 && layer.Color.A > 0 {
				fn(layer)
			}
		}
		return
	}
	if b.Blur < 0 || b.Color.A == 0 {
		return
	}
	fn(ShadowLayer{
		OffsetX: b.OffsetX,
		OffsetY: b.OffsetY,
		Blur:    b.Blur,
		Spread:  b.Spread,
		Color:   b.Color,
	})
}

// ThemeShadow resolves theme-controlled layers against a base color.
func ThemeShadow(style theme.ShadowTheme, col color.NRGBA, opacity float32) BoxShadow {
	opacity = shadowOpacity(opacity)
	layers := make([]ShadowLayer, 0, len(style.Layers))
	for _, token := range style.Layers {
		layerOpacity := shadowOpacity(token.Opacity) * opacity
		layerColor := col
		layerColor.A = scaleAlpha(col.A, layerOpacity)
		if token.Blur < 0 || layerColor.A == 0 {
			continue
		}
		layers = append(layers, ShadowLayer{
			OffsetX: float32(token.OffsetX),
			OffsetY: float32(token.OffsetY),
			Blur:    float32(token.Blur),
			Spread:  float32(token.Spread),
			Color:   layerColor,
		})
	}
	return BoxShadow{Blur: -1, Layers: layers}
}

func shadowOpacity(value float32) float32 {
	if value <= 0 || math.IsNaN(float64(value)) || math.IsInf(float64(value), 0) {
		return 0
	}
	return min(value, 1)
}

// DrawShadow paints a cached soft shadow for bounds. The bounds rectangle may be
// offset; the shadow is drawn around that rectangle in the current operation
// coordinate space.
func DrawShadow(gtx layout.Context, bounds image.Rectangle, shape ShadowShape, box BoxShadow) {
	size := bounds.Size()
	if size.X <= 0 || size.Y <= 0 {
		return
	}
	drawShadow(gtx, bounds, shape.pixels(gtx, size), box)
}

func drawShadow(gtx layout.Context, bounds image.Rectangle, pxShape shadowShape, box BoxShadow) {
	size := bounds.Size()
	if size.X <= 0 || size.Y <= 0 {
		return
	}
	if bounds.Min != (image.Point{}) {
		stack := op.Offset(bounds.Min).Push(gtx.Ops)
		defer stack.Pop()
	}

	box.forEachLayer(func(layer ShadowLayer) {
		blur := gtx.Dp(unit.Dp(layer.Blur))
		if blur < 0 || layer.Color.A == 0 {
			return
		}
		opacity := float32(layer.Color.A) / 255
		cacheColor := layer.Color
		cacheColor.A = 255
		entry := softShadowEntry(
			size,
			pxShape,
			blur,
			gtx.Dp(unit.Dp(layer.Spread)),
			gtx.Dp(unit.Dp(layer.OffsetX)),
			gtx.Dp(unit.Dp(layer.OffsetY)),
			cacheColor,
		)
		if entry.op.Size() == (image.Point{}) {
			return
		}
		stack := op.Offset(image.Pt(-entry.padX, -entry.padY)).Push(gtx.Ops)
		clipStack := clip.Rect(image.Rectangle{Max: entry.size}).Push(gtx.Ops)
		var scaleStack op.TransformStack
		if entry.scale > 1 {
			scale := float32(entry.scale)
			scaleStack = op.Affine(f32.AffineId().Scale(f32.Point{}, f32.Pt(scale, scale))).Push(gtx.Ops)
		}
		var opacityStack paint.OpacityStack
		if opacity < 1 {
			opacityStack = paint.PushOpacity(gtx.Ops, opacity)
		}
		entry.op.Add(gtx.Ops)
		paint.PaintOp{}.Add(gtx.Ops)
		if opacity < 1 {
			opacityStack.Pop()
		}
		if entry.scale > 1 {
			scaleStack.Pop()
		}
		clipStack.Pop()
		stack.Pop()
	})
}

// DrawSurface paints a rounded surface with the supplied shadow. It is the
// themeable form used by higher-level widgets.
func DrawSurface(gtx layout.Context, rect image.Rectangle, radius int, surface color.NRGBA, shadow BoxShadow) {
	radius = max(radius, 0)
	shape := shadowShape{
		Kind: ShadowRoundedRect,
		Radii: cornerRadiiPx{
			NW: radius,
			NE: radius,
			SE: radius,
			SW: radius,
		},
	}
	drawShadow(gtx, rect, shape, shadow)
	paint.FillShape(gtx.Ops, surface, clip.UniformRRect(rect, radius).Op(gtx.Ops))
}

func scaleAlpha(a uint8, scale float32) uint8 {
	return clampByte(float32(a) * scale)
}

type softShadowCacheKey struct {
	version    int
	w, h       int
	shape      shadowShape
	blur       int
	scale      int
	spread     int
	offX, offY int
	color      color.NRGBA
}

type softShadowCacheEntry struct {
	op         paint.ImageOp
	padX, padY int
	size       image.Point
	scale      int
	bytes      int
	lastUsed   uint64
}

type softShadowBuild struct {
	done  chan struct{}
	entry softShadowCacheEntry
}

const (
	softShadowRendererVersion = 2
	softShadowCacheLimit      = 256
	softShadowCacheMaxBytes   = 32 << 20
	maxShadowRasterPixels     = 256 * 1024
	maxShadowRasterScale      = 8
)

var byteSlicePool = sync.Pool{
	New: func() any {
		buf := make([]uint8, 0, 128*1024)
		return &buf
	},
}

var softShadowCache = struct {
	sync.Mutex
	tick    uint64
	bytes   int
	entries map[softShadowCacheKey]softShadowCacheEntry
	builds  map[softShadowCacheKey]*softShadowBuild
}{
	entries: make(map[softShadowCacheKey]softShadowCacheEntry),
	builds:  make(map[softShadowCacheKey]*softShadowBuild),
}

type cornerRadiiPx struct {
	NW int
	NE int
	SE int
	SW int
}

type shadowShape struct {
	Kind  ShadowShapeKind
	Radii cornerRadiiPx
}

func softShadowEntry(size image.Point, shape shadowShape, blur, spread, offX, offY int, col color.NRGBA) softShadowCacheEntry {
	if size.X <= 0 || size.Y <= 0 || blur < 0 || col.A == 0 {
		return softShadowCacheEntry{}
	}
	blur = min(blur, 96)
	shape = normalizeShadowShape(image.Rectangle{Max: size}, shape)
	padX := blur*2 + absInt(spread) + absInt(offX) + 2
	padY := blur*2 + absInt(spread) + absInt(offY) + 2
	rasterSize, ok := shadowRasterSize(size, padX, padY)
	if !ok {
		return softShadowCacheEntry{}
	}
	scale := shadowRasterScale(blur, rasterSize.X, rasterSize.Y)
	key := softShadowCacheKey{
		version: softShadowRendererVersion,
		w:       size.X,
		h:       size.Y,
		shape:   shape,
		blur:    blur,
		scale:   scale,
		spread:  spread,
		offX:    offX, offY: offY, color: col,
	}

	softShadowCache.Lock()
	softShadowCache.tick++
	if entry, ok := softShadowCache.entries[key]; ok {
		entry.lastUsed = softShadowCache.tick
		softShadowCache.entries[key] = entry
		softShadowCache.Unlock()
		return entry
	}
	if build, ok := softShadowCache.builds[key]; ok {
		softShadowCache.Unlock()
		<-build.done
		return build.entry
	}
	build := &softShadowBuild{done: make(chan struct{})}
	softShadowCache.builds[key] = build
	softShadowCache.Unlock()

	entry := buildSoftShadowEntry(size, shape, blur, spread, offX, offY, col)
	finishBuild := func(entry softShadowCacheEntry) softShadowCacheEntry {
		softShadowCache.Lock()
		delete(softShadowCache.builds, key)
		build.entry = entry
		close(build.done)
		softShadowCache.Unlock()
		return entry
	}
	if entry.op.Size() == (image.Point{}) {
		return finishBuild(entry)
	}
	if entry.bytes > softShadowCacheMaxBytes {
		return finishBuild(entry)
	}

	softShadowCache.Lock()
	entry.lastUsed = softShadowCache.tick
	for len(softShadowCache.entries) >= softShadowCacheLimit || softShadowCache.bytes+entry.bytes > softShadowCacheMaxBytes {
		evictOldestShadowEntryLocked()
	}
	softShadowCache.entries[key] = entry
	softShadowCache.bytes += entry.bytes
	delete(softShadowCache.builds, key)
	build.entry = entry
	close(build.done)
	softShadowCache.Unlock()
	return entry
}

func evictOldestShadowEntryLocked() {
	var oldestKey softShadowCacheKey
	var oldestTick uint64
	first := true
	for key, entry := range softShadowCache.entries {
		if first || entry.lastUsed < oldestTick {
			oldestKey = key
			oldestTick = entry.lastUsed
			first = false
		}
	}
	if !first {
		softShadowCache.bytes -= softShadowCache.entries[oldestKey].bytes
		if softShadowCache.bytes < 0 {
			softShadowCache.bytes = 0
		}
		delete(softShadowCache.entries, oldestKey)
	}
}

func buildSoftShadowEntry(size image.Point, shape shadowShape, blur, spread, offX, offY int, col color.NRGBA) softShadowCacheEntry {
	padX := blur*2 + absInt(spread) + absInt(offX) + 2
	padY := blur*2 + absInt(spread) + absInt(offY) + 2
	rasterSize, ok := shadowRasterSize(size, padX, padY)
	if !ok {
		return softShadowCacheEntry{}
	}
	w := rasterSize.X
	h := rasterSize.Y
	if w <= 0 || h <= 0 {
		return softShadowCacheEntry{}
	}

	scale := shadowRasterScale(blur, w, h)
	lw := ceilDiv(w, scale)
	lh := ceilDiv(h, scale)
	if lw <= 0 || lh <= 0 {
		return softShadowCacheEntry{}
	}
	if lw > maxShadowRasterPixels/lh {
		return softShadowCacheEntry{}
	}

	alpha := borrowByteSlice(lw * lh)
	defer releaseByteSlice(alpha)

	shapeRect := image.Rect(
		padX+offX-spread,
		padY+offY-spread,
		padX+offX+size.X+spread,
		padY+offY+size.Y+spread,
	)
	lowShapeRect := scaleRectDown(shapeRect, scale)
	lowCutRect := scaleRectDown(image.Rect(padX, padY, padX+size.X, padY+size.Y), scale)
	lowShape := scaleShapeDown(shapeWithSpread(shape, spread), scale)
	lowCutShape := scaleShapeDown(shape, scale)
	drawShapeMask(alpha, lw, lh, lowShapeRect, lowShape, col.A)
	boxBlurAlpha(alpha, lw, lh, ceilDiv(blur, scale))
	cutShapeMask(alpha, lw, lh, lowCutRect, lowCutShape)

	img := image.NewRGBA(image.Rect(0, 0, lw, lh))
	for y := range lh {
		row := y * img.Stride
		for x := range lw {
			a := rasterAlpha(alpha[y*lw+x], scale, x, y)
			if a == 0 {
				continue
			}
			i := row + x*4
			img.Pix[i+0] = premul(col.R, a)
			img.Pix[i+1] = premul(col.G, a)
			img.Pix[i+2] = premul(col.B, a)
			img.Pix[i+3] = a
		}
	}
	op := paint.NewImageOp(img)
	op.Filter = paint.FilterLinear
	return softShadowCacheEntry{op: op, padX: padX, padY: padY, size: image.Pt(w, h), scale: scale, bytes: rgbaBytes(lw, lh)}
}

func shadowRasterScale(blur, w, h int) int {
	scale := 1
	switch {
	case blur >= 48 && pixelsAtLeast(w, h, 256*256):
		scale = 4
	case blur >= 16 && pixelsAtLeast(w, h, 128*128):
		scale = 2
	}
	for pixelsExceed(ceilDiv(w, scale), ceilDiv(h, scale), maxShadowRasterPixels) && scale < maxShadowRasterScale {
		scale *= 2
	}
	return scale
}

func shadowRasterSize(size image.Point, padX, padY int) (image.Point, bool) {
	if size.X <= 0 || size.Y <= 0 || padX < 0 || padY < 0 {
		return image.Point{}, false
	}
	maxInt := maxIntValue()
	if padX > (maxInt-size.X)/2 || padY > (maxInt-size.Y)/2 {
		return image.Point{}, false
	}
	return image.Pt(size.X+padX*2, size.Y+padY*2), true
}

func pixelsAtLeast(w, h, threshold int) bool {
	if threshold <= 0 {
		return true
	}
	if w <= 0 || h <= 0 {
		return false
	}
	return w > (threshold-1)/h
}

func pixelsExceed(w, h, limit int) bool {
	if limit < 0 {
		return true
	}
	if w <= 0 || h <= 0 {
		return false
	}
	return w > limit/h
}

func rgbaBytes(w, h int) int {
	if w <= 0 || h <= 0 {
		return 0
	}
	maxInt := maxIntValue()
	if w > maxInt/h/4 {
		return maxInt
	}
	return w * h * 4
}

func scaleRectDown(rect image.Rectangle, scale int) image.Rectangle {
	if scale <= 1 {
		return rect
	}
	return image.Rect(
		floorDiv(rect.Min.X, scale),
		floorDiv(rect.Min.Y, scale),
		ceilDiv(rect.Max.X, scale),
		ceilDiv(rect.Max.Y, scale),
	)
}

func scaleShapeDown(shape shadowShape, scale int) shadowShape {
	if scale <= 1 || shape.Kind == ShadowEllipse {
		return shape
	}
	shape.Radii.NW = ceilDiv(shape.Radii.NW, scale)
	shape.Radii.NE = ceilDiv(shape.Radii.NE, scale)
	shape.Radii.SE = ceilDiv(shape.Radii.SE, scale)
	shape.Radii.SW = ceilDiv(shape.Radii.SW, scale)
	return shape
}

func drawShapeMask(alpha []uint8, w, h int, rect image.Rectangle, shape shadowShape, a uint8) {
	if rect.Dx() <= 0 || rect.Dy() <= 0 || a == 0 {
		return
	}
	shape = normalizeShadowShape(rect, shape)
	if shape.Kind != ShadowEllipse {
		drawRoundedShapeMask(alpha, w, h, rect, shape.Radii, a)
		return
	}
	for y := max(rect.Min.Y, 0); y < min(rect.Max.Y, h); y++ {
		for x := max(rect.Min.X, 0); x < min(rect.Max.X, w); x++ {
			coverage := shapeCoverageNormalized(x, y, rect, shape)
			alpha[y*w+x] = uint8(float32(a)*coverage + .5)
		}
	}
}

func cutShapeMask(alpha []uint8, w, h int, rect image.Rectangle, shape shadowShape) {
	if rect.Dx() <= 0 || rect.Dy() <= 0 {
		return
	}
	shape = normalizeShadowShape(rect, shape)
	if shape.Kind != ShadowEllipse {
		cutRoundedShapeMask(alpha, w, h, rect, shape.Radii)
		return
	}
	for y := max(rect.Min.Y-1, 0); y < min(rect.Max.Y+1, h); y++ {
		for x := max(rect.Min.X-1, 0); x < min(rect.Max.X+1, w); x++ {
			coverage := shapeCoverageNormalized(x, y, rect, shape)
			if coverage <= 0 {
				continue
			}
			i := y*w + x
			alpha[i] = uint8(float32(alpha[i])*(1-coverage) + .5)
		}
	}
}

func drawRoundedShapeMask(alpha []uint8, w, h int, rect image.Rectangle, radii cornerRadiiPx, a uint8) {
	fillRoundedShapeRects(alpha, w, h, rect, radii, a)
	drawCornerMask(alpha, w, h, image.Rect(rect.Min.X, rect.Min.Y, rect.Min.X+radii.NW, rect.Min.Y+radii.NW), rect, radii, a)
	drawCornerMask(alpha, w, h, image.Rect(rect.Max.X-radii.NE, rect.Min.Y, rect.Max.X, rect.Min.Y+radii.NE), rect, radii, a)
	drawCornerMask(alpha, w, h, image.Rect(rect.Max.X-radii.SE, rect.Max.Y-radii.SE, rect.Max.X, rect.Max.Y), rect, radii, a)
	drawCornerMask(alpha, w, h, image.Rect(rect.Min.X, rect.Max.Y-radii.SW, rect.Min.X+radii.SW, rect.Max.Y), rect, radii, a)
}

func cutRoundedShapeMask(alpha []uint8, w, h int, rect image.Rectangle, radii cornerRadiiPx) {
	fillRoundedShapeRects(alpha, w, h, rect, radii, 0)
	cutCornerMask(alpha, w, h, image.Rect(rect.Min.X, rect.Min.Y, rect.Min.X+radii.NW, rect.Min.Y+radii.NW), rect, radii)
	cutCornerMask(alpha, w, h, image.Rect(rect.Max.X-radii.NE, rect.Min.Y, rect.Max.X, rect.Min.Y+radii.NE), rect, radii)
	cutCornerMask(alpha, w, h, image.Rect(rect.Max.X-radii.SE, rect.Max.Y-radii.SE, rect.Max.X, rect.Max.Y), rect, radii)
	cutCornerMask(alpha, w, h, image.Rect(rect.Min.X, rect.Max.Y-radii.SW, rect.Min.X+radii.SW, rect.Max.Y), rect, radii)
}

func fillRoundedShapeRects(alpha []uint8, w, h int, rect image.Rectangle, radii cornerRadiiPx, a uint8) {
	for y := max(rect.Min.Y, 0); y < min(rect.Max.Y, h); y++ {
		leftInset := 0
		if y < rect.Min.Y+radii.NW {
			leftInset = radii.NW
		} else if y >= rect.Max.Y-radii.SW {
			leftInset = radii.SW
		}
		rightInset := 0
		if y < rect.Min.Y+radii.NE {
			rightInset = radii.NE
		} else if y >= rect.Max.Y-radii.SE {
			rightInset = radii.SE
		}
		minX := max(rect.Min.X+leftInset, 0)
		maxX := min(rect.Max.X-rightInset, w)
		row := y * w
		for x := minX; x < maxX; x++ {
			alpha[row+x] = a
		}
	}
}

func drawCornerMask(alpha []uint8, w, h int, corner, rect image.Rectangle, radii cornerRadiiPx, a uint8) {
	for y := max(corner.Min.Y, 0); y < min(corner.Max.Y, h); y++ {
		for x := max(corner.Min.X, 0); x < min(corner.Max.X, w); x++ {
			coverage := roundedRectCoverage(x, y, rect, radii)
			alpha[y*w+x] = uint8(float32(a)*coverage + .5)
		}
	}
}

func cutCornerMask(alpha []uint8, w, h int, corner, rect image.Rectangle, radii cornerRadiiPx) {
	for y := max(corner.Min.Y, 0); y < min(corner.Max.Y, h); y++ {
		for x := max(corner.Min.X, 0); x < min(corner.Max.X, w); x++ {
			coverage := roundedRectCoverage(x, y, rect, radii)
			if coverage <= 0 {
				continue
			}
			i := y*w + x
			alpha[i] = uint8(float32(alpha[i])*(1-coverage) + .5)
		}
	}
}

func shapeCoverage(x, y int, rect image.Rectangle, shape shadowShape) float32 {
	shape = normalizeShadowShape(rect, shape)
	return shapeCoverageNormalized(x, y, rect, shape)
}

func shapeCoverageNormalized(x, y int, rect image.Rectangle, shape shadowShape) float32 {
	if shape.Kind == ShadowEllipse {
		return ellipseCoverage(x, y, rect)
	}
	return roundedRectCoverage(x, y, rect, shape.Radii)
}

func roundedRectCoverage(x, y int, rect image.Rectangle, radii cornerRadiiPx) float32 {
	if rect.Empty() || x < rect.Min.X || x >= rect.Max.X || y < rect.Min.Y || y >= rect.Max.Y {
		return 0
	}

	var cx, cy float64
	var radius int
	switch {
	case x < rect.Min.X+radii.NW && y < rect.Min.Y+radii.NW:
		radius = radii.NW
		cx = float64(rect.Min.X + radius)
		cy = float64(rect.Min.Y + radius)
	case x >= rect.Max.X-radii.NE && y < rect.Min.Y+radii.NE:
		radius = radii.NE
		cx = float64(rect.Max.X - radius)
		cy = float64(rect.Min.Y + radius)
	case x >= rect.Max.X-radii.SE && y >= rect.Max.Y-radii.SE:
		radius = radii.SE
		cx = float64(rect.Max.X - radius)
		cy = float64(rect.Max.Y - radius)
	case x < rect.Min.X+radii.SW && y >= rect.Max.Y-radii.SW:
		radius = radii.SW
		cx = float64(rect.Min.X + radius)
		cy = float64(rect.Max.Y - radius)
	default:
		return 1
	}
	if radius <= 0 {
		return 1
	}

	return circlePixelCoverage(x, y, cx, cy, float64(radius))
}

const aaGrid = 4

func circlePixelCoverage(x, y int, cx, cy, radius float64) float32 {
	px := float64(x) + 0.5
	py := float64(y) + 0.5
	dist := math.Hypot(px-cx, py-cy)
	edge := radius - dist
	switch {
	case edge >= 0.75:
		return 1
	case edge <= -0.75:
		return 0
	default:
		r2 := radius * radius
		inside := 0
		for sy := range aaGrid {
			fy := float64(y) + (float64(sy)+0.5)/aaGrid
			dy := fy - cy
			for sx := range aaGrid {
				fx := float64(x) + (float64(sx)+0.5)/aaGrid
				dx := fx - cx
				if dx*dx+dy*dy <= r2 {
					inside++
				}
			}
		}
		return float32(inside) / float32(aaGrid*aaGrid)
	}
}

func ellipseCoverage(x, y int, rect image.Rectangle) float32 {
	if rect.Empty() || x < rect.Min.X || x >= rect.Max.X || y < rect.Min.Y || y >= rect.Max.Y {
		return 0
	}
	rx := float64(rect.Dx()) / 2
	ry := float64(rect.Dy()) / 2
	if rx <= 0 || ry <= 0 {
		return 0
	}
	cx := float64(rect.Min.X+rect.Max.X) / 2
	cy := float64(rect.Min.Y+rect.Max.Y) / 2
	dx := (float64(x) + 0.5 - cx) / rx
	dy := (float64(y) + 0.5 - cy) / ry
	dist := math.Hypot(dx, dy)
	edge := (1 - dist) * math.Min(rx, ry)
	switch {
	case edge >= 0.75:
		return 1
	case edge <= -0.75:
		return 0
	default:
		inside := 0
		for sy := range aaGrid {
			fy := float64(y) + (float64(sy)+0.5)/aaGrid
			ny := (fy - cy) / ry
			for sx := range aaGrid {
				fx := float64(x) + (float64(sx)+0.5)/aaGrid
				nx := (fx - cx) / rx
				if nx*nx+ny*ny <= 1 {
					inside++
				}
			}
		}
		return float32(inside) / float32(aaGrid*aaGrid)
	}
}

func shapeWithSpread(shape shadowShape, spread int) shadowShape {
	if shape.Kind == ShadowEllipse {
		return shape
	}
	shape.Radii.NW = max(0, shape.Radii.NW+spread)
	shape.Radii.NE = max(0, shape.Radii.NE+spread)
	shape.Radii.SE = max(0, shape.Radii.SE+spread)
	shape.Radii.SW = max(0, shape.Radii.SW+spread)
	return shape
}

func normalizeShadowShape(rect image.Rectangle, shape shadowShape) shadowShape {
	if shape.Kind == ShadowEllipse {
		return shadowShape{Kind: ShadowEllipse}
	}
	shape.Kind = ShadowRoundedRect
	shape.Radii = normalizeCornerRadii(rect, shape.Radii)
	return shape
}

func normalizeCornerRadii(rect image.Rectangle, radii cornerRadiiPx) cornerRadiiPx {
	radii.NW = max(radii.NW, 0)
	radii.NE = max(radii.NE, 0)
	radii.SE = max(radii.SE, 0)
	radii.SW = max(radii.SW, 0)

	w := rect.Dx()
	h := rect.Dy()
	if w <= 0 || h <= 0 {
		return cornerRadiiPx{}
	}

	scale := 1.0
	scale = minFloat(scale, cornerScale(w, radii.NW+radii.NE))
	scale = minFloat(scale, cornerScale(w, radii.SW+radii.SE))
	scale = minFloat(scale, cornerScale(h, radii.NW+radii.SW))
	scale = minFloat(scale, cornerScale(h, radii.NE+radii.SE))
	if scale < 1 {
		radii.NW = int(math.Floor(float64(radii.NW)*scale + .5))
		radii.NE = int(math.Floor(float64(radii.NE)*scale + .5))
		radii.SE = int(math.Floor(float64(radii.SE)*scale + .5))
		radii.SW = int(math.Floor(float64(radii.SW)*scale + .5))
	}
	return radii
}

func cornerScale(size, sum int) float64 {
	if sum <= 0 || sum <= size {
		return 1
	}
	return float64(size) / float64(sum)
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func boxBlurAlpha(alpha []uint8, w, h, radius int) {
	if radius <= 0 || w <= 0 || h <= 0 {
		return
	}
	passes := 3
	boxRadius := max(int(math.Ceil(float64(radius)/2.0)), 1)
	tmp := borrowByteSlice(len(alpha))
	defer releaseByteSlice(tmp)
	for range passes {
		blurHorizontal(alpha, tmp, w, h, boxRadius)
		blurVertical(tmp, alpha, w, h, boxRadius)
	}
}

func borrowByteSlice(size int) []uint8 {
	bufp := byteSlicePool.Get().(*[]uint8)
	buf := *bufp
	if cap(buf) < size {
		buf = make([]uint8, size)
	} else {
		buf = buf[:size]
		clear(buf)
	}
	*bufp = nil
	return buf
}

func releaseByteSlice(buf []uint8) {
	if cap(buf) > 4*1024*1024 {
		return
	}
	buf = buf[:0]
	byteSlicePool.Put(&buf)
}

var bayer4 = [16]int{
	0, 8, 2, 10,
	12, 4, 14, 6,
	3, 11, 1, 9,
	15, 7, 13, 5,
}

func ditherAlpha(a uint8, x, y int) uint8 {
	if a == 0 || a == 255 {
		return a
	}
	rank := bayer4[(y&3)*4+(x&3)]
	threshold := rank - 8
	if rank >= 8 {
		threshold = rank - 7
	}
	if threshold < 0 && a <= uint8(-threshold) {
		return 0
	}
	if threshold > 0 && int(a)+threshold >= 255 {
		return 255
	}
	return uint8(int(a) + threshold)
}

func rasterAlpha(a uint8, scale, x, y int) uint8 {
	if scale > 1 {
		return a
	}
	return ditherAlpha(a, x, y)
}

func blurHorizontal(src, dst []uint8, w, h, r int) {
	window := r*2 + 1
	for y := range h {
		sum := 0
		row := y * w
		for x := -r; x <= r; x++ {
			sum += int(src[row+clampInt(x, 0, w-1)])
		}
		for x := range w {
			dst[row+x] = uint8(sum / window)
			remove := clampInt(x-r, 0, w-1)
			add := clampInt(x+r+1, 0, w-1)
			sum += int(src[row+add]) - int(src[row+remove])
		}
	}
}

func blurVertical(src, dst []uint8, w, h, r int) {
	window := r*2 + 1
	for x := range w {
		sum := 0
		for y := -r; y <= r; y++ {
			sum += int(src[clampInt(y, 0, h-1)*w+x])
		}
		for y := range h {
			dst[y*w+x] = uint8(sum / window)
			remove := clampInt(y-r, 0, h-1)
			add := clampInt(y+r+1, 0, h-1)
			sum += int(src[add*w+x]) - int(src[remove*w+x])
		}
	}
}

func premul(c, a uint8) uint8 {
	return uint8((uint16(c)*uint16(a) + 127) / 255)
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func maxIntValue() int {
	return int(^uint(0) >> 1)
}

func ceilDiv(v, d int) int {
	if d <= 0 {
		return v
	}
	if v >= 0 {
		return (v + d - 1) / d
	}
	return v / d
}

func floorDiv(v, d int) int {
	if d <= 0 {
		return v
	}
	if v >= 0 {
		return v / d
	}
	return -((-v + d - 1) / d)
}

func clampByte(v float32) uint8 {
	if v <= 0 {
		return 0
	}
	if v >= 255 {
		return 255
	}
	return uint8(math.Round(float64(v)))
}

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
