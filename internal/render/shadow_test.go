package render

import (
	"image"
	"image/color"
	"math"
	"sync"
	"testing"

	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/FlowUI/internal/theme"
)

func TestThemeShadowResolvesConfiguredLayers(t *testing.T) {
	style := theme.ShadowTheme{Layers: [theme.ShadowLayerCount]theme.ShadowLayerTheme{
		{OffsetX: 1, OffsetY: 2, Blur: 3, Spread: 4, Opacity: .8},
	}}
	layers := ThemeShadow(style, color.NRGBA{R: 10, A: 200}, .5).EffectiveLayers()
	if len(layers) != 1 {
		t.Fatalf("resolved layers = %d, want 1", len(layers))
	}
	if got := layers[0]; got.OffsetX != 1 || got.OffsetY != 2 || got.Blur != 3 || got.Spread != 4 || got.Color != (color.NRGBA{R: 10, A: 80}) {
		t.Fatalf("resolved layer = %#v", got)
	}
}

func TestThemeShadowRejectsInvalidOpacity(t *testing.T) {
	style := theme.ShadowTheme{Layers: [theme.ShadowLayerCount]theme.ShadowLayerTheme{{Blur: 4, Opacity: float32(math.NaN())}}}
	if layers := ThemeShadow(style, color.NRGBA{A: 200}, 1).EffectiveLayers(); len(layers) != 0 {
		t.Fatalf("invalid shadow layers = %#v", layers)
	}
}

func TestBoxShadowEffectiveLayersKeepsHardShadow(t *testing.T) {
	box := BoxShadow{
		OffsetY: 2,
		Blur:    0,
		Color:   color.NRGBA{A: 120},
	}
	layers := box.EffectiveLayers()
	if len(layers) != 1 {
		t.Fatalf("hard shadow layers = %d, want 1", len(layers))
	}
	if layers[0].Blur != 0 {
		t.Fatalf("hard shadow blur = %v, want 0", layers[0].Blur)
	}
}

func TestSoftShadowCacheReusesEntries(t *testing.T) {
	resetSoftShadowCacheForTest()
	col := color.NRGBA{R: 30, G: 80, B: 220, A: 120}
	shape := shadowShape{Kind: ShadowRoundedRect, Radii: cornerRadiiPx{NW: 4, NE: 8, SE: 6, SW: 2}}
	_ = softShadowEntry(image.Pt(24, 16), shape, 6, 1, 0, 2, col)
	_ = softShadowEntry(image.Pt(24, 16), shape, 6, 1, 0, 2, col)

	softShadowCache.Lock()
	defer softShadowCache.Unlock()
	if got := len(softShadowCache.entries); got != 1 {
		t.Fatalf("cache entries = %d, want 1", got)
	}
}

func TestSoftShadowCacheReusesAlphaVariants(t *testing.T) {
	resetSoftShadowCacheForTest()
	var ops op.Ops
	gtx := layout.Context{Ops: &ops}
	shape := shadowShape{Kind: ShadowRoundedRect, Radii: cornerRadiiPx{NW: 4, NE: 8, SE: 6, SW: 2}}
	for _, alpha := range []uint8{32, 64, 128, 255} {
		drawShadow(gtx, image.Rect(0, 0, 48, 32), shape, BoxShadow{
			Blur:  8,
			Color: color.NRGBA{R: 30, G: 80, B: 220, A: alpha},
		})
	}

	softShadowCache.Lock()
	defer softShadowCache.Unlock()
	if got := len(softShadowCache.entries); got != 1 {
		t.Fatalf("cache entries = %d, want 1", got)
	}
	for key, entry := range softShadowCache.entries {
		if key.color.A != 255 {
			t.Fatalf("cached alpha = %d, want 255", key.color.A)
		}
		if softShadowCache.bytes != entry.bytes {
			t.Fatalf("cache bytes = %d, want %d", softShadowCache.bytes, entry.bytes)
		}
	}
}

func TestSoftShadowCacheConcurrentMissesTrackBytesOnce(t *testing.T) {
	resetSoftShadowCacheForTest()
	col := color.NRGBA{R: 40, G: 90, B: 220, A: 120}
	shape := shadowShape{Kind: ShadowRoundedRect, Radii: cornerRadiiPx{NW: 20, NE: 10, SE: 24, SW: 14}}
	start := make(chan struct{})
	var group sync.WaitGroup
	for range 16 {
		group.Add(1)
		go func() {
			defer group.Done()
			<-start
			_ = softShadowEntry(image.Pt(320, 180), shape, 22, 4, 0, 8, col)
		}()
	}
	close(start)
	group.Wait()

	softShadowCache.Lock()
	defer softShadowCache.Unlock()
	if len(softShadowCache.entries) != 1 {
		t.Fatalf("cache entries = %d, want 1", len(softShadowCache.entries))
	}
	for _, entry := range softShadowCache.entries {
		if softShadowCache.bytes != entry.bytes {
			t.Fatalf("cache bytes = %d, want %d", softShadowCache.bytes, entry.bytes)
		}
	}
}

func TestSoftShadowCacheEvictsOldEntries(t *testing.T) {
	resetSoftShadowCacheForTest()
	shape := shadowShape{Kind: ShadowEllipse}
	for i := range softShadowCacheLimit + 20 {
		col := color.NRGBA{R: uint8(i), G: uint8(i >> 1), B: 180, A: 90}
		_ = softShadowEntry(image.Pt(8, 8), shape, 0, 0, i%3, i%5, col)
	}

	softShadowCache.Lock()
	defer softShadowCache.Unlock()
	if got := len(softShadowCache.entries); got > softShadowCacheLimit {
		t.Fatalf("cache entries = %d, want <= %d", got, softShadowCacheLimit)
	}
}

func TestCutRoundedRectMaskRemovesSurfaceInterior(t *testing.T) {
	w, h := 32, 24
	alpha := make([]uint8, w*h)
	rect := image.Rect(6, 5, 26, 19)
	shape := shadowShape{Kind: ShadowRoundedRect, Radii: cornerRadiiPx{NW: 4, NE: 8, SE: 6, SW: 2}}
	drawShapeMask(alpha, w, h, rect, shape, 200)
	cutShapeMask(alpha, w, h, rect, shape)

	if got := alpha[12*w+16]; got != 0 {
		t.Fatalf("center alpha = %d, want 0", got)
	}
}

func TestNonUniformCornerCoverage(t *testing.T) {
	rect := image.Rect(0, 0, 40, 30)
	shape := shadowShape{Kind: ShadowRoundedRect, Radii: cornerRadiiPx{NW: 14, NE: 2, SE: 10, SW: 0}}

	if got := shapeCoverage(0, 0, rect, shape); got != 0 {
		t.Fatalf("large rounded NW corner coverage = %v, want 0", got)
	}
	if got := shapeCoverage(39, 0, rect, shape); got == 0 {
		t.Fatalf("small rounded NE corner coverage = %v, want > 0", got)
	}
	if got := shapeCoverage(0, 29, rect, shape); got != 1 {
		t.Fatalf("square SW corner coverage = %v, want 1", got)
	}
}

func TestEllipseCoverage(t *testing.T) {
	rect := image.Rect(0, 0, 30, 30)
	shape := shadowShape{Kind: ShadowEllipse}

	if got := shapeCoverage(15, 15, rect, shape); got != 1 {
		t.Fatalf("ellipse center coverage = %v, want 1", got)
	}
	if got := shapeCoverage(0, 0, rect, shape); got != 0 {
		t.Fatalf("ellipse corner coverage = %v, want 0", got)
	}
}

func TestAntialiasCoverageReturnsFractionalEdges(t *testing.T) {
	ellipse := shadowShape{Kind: ShadowEllipse}
	got := shapeCoverage(19, 6, image.Rect(0, 0, 20, 20), ellipse)
	if got <= 0 || got >= 1 {
		t.Fatalf("ellipse edge coverage = %v, want fractional", got)
	}

	rounded := shadowShape{Kind: ShadowRoundedRect, Radii: cornerRadiiPx{NW: 8}}
	got = shapeCoverage(2, 2, image.Rect(0, 0, 20, 20), rounded)
	if got <= 0 || got >= 1 {
		t.Fatalf("rounded edge coverage = %v, want fractional", got)
	}
}

func TestSoftShadowEntryUsesDownsampledRasterForLargeBlur(t *testing.T) {
	resetSoftShadowCacheForTest()
	col := color.NRGBA{R: 40, G: 90, B: 220, A: 120}
	shape := shadowShape{Kind: ShadowRoundedRect, Radii: cornerRadiiPx{NW: 20, NE: 10, SE: 24, SW: 14}}
	entry := softShadowEntry(image.Pt(320, 180), shape, 22, 4, 0, 8, col)

	if entry.scale != 2 {
		t.Fatalf("entry scale = %d, want 2", entry.scale)
	}
	if entry.size.X <= entry.op.Size().X || entry.size.Y <= entry.op.Size().Y {
		t.Fatalf("full size = %v, op size = %v; want downsampled op", entry.size, entry.op.Size())
	}
}

func TestShadowRasterScaleLimitsLargeRasters(t *testing.T) {
	scale := shadowRasterScale(96, 5000, 3000)
	pixels := ceilDiv(5000, scale) * ceilDiv(3000, scale)
	if pixels > maxShadowRasterPixels {
		t.Fatalf("downsampled pixels = %d, want <= %d", pixels, maxShadowRasterPixels)
	}
}

func TestHugeShadowEntryIsRejected(t *testing.T) {
	col := color.NRGBA{R: 20, G: 80, B: 180, A: 120}
	shape := shadowShape{Kind: ShadowRoundedRect, Radii: cornerRadiiPx{NW: 24, NE: 24, SE: 24, SW: 24}}
	entry := buildSoftShadowEntry(image.Pt(20000, 20000), shape, 96, 4, 0, 8, col)
	if entry.op.Size() != (image.Point{}) {
		t.Fatalf("huge entry op size = %v, want empty", entry.op.Size())
	}
}

func TestOverflowShadowEntryIsRejected(t *testing.T) {
	col := color.NRGBA{R: 20, G: 80, B: 180, A: 120}
	shape := shadowShape{Kind: ShadowRoundedRect, Radii: cornerRadiiPx{NW: 24, NE: 24, SE: 24, SW: 24}}
	entry := buildSoftShadowEntry(image.Pt(maxIntValue(), maxIntValue()), shape, 4, 0, 0, 0, col)
	if entry.op.Size() != (image.Point{}) {
		t.Fatalf("overflow entry op size = %v, want empty", entry.op.Size())
	}
}

func TestRasterAlphaSkipsDitherWhenDownsampled(t *testing.T) {
	const alpha = uint8(100)
	if got := rasterAlpha(alpha, 2, 0, 0); got != alpha {
		t.Fatalf("downsampled alpha = %d, want unchanged %d", got, alpha)
	}
	if got := rasterAlpha(alpha, 1, 0, 0); got == alpha {
		t.Fatalf("full-res alpha = %d, want dithered value", got)
	}
}

func TestDitherAlphaHasZeroMeanThreshold(t *testing.T) {
	sum := 0
	for y := range 4 {
		for x := range 4 {
			sum += int(ditherAlpha(128, x, y)) - 128
		}
	}
	if sum != 0 {
		t.Fatalf("dither alpha threshold sum = %d, want 0", sum)
	}
}

func TestSoftShadowCacheTracksBytes(t *testing.T) {
	resetSoftShadowCacheForTest()
	col := color.NRGBA{R: 20, G: 80, B: 180, A: 120}
	shape := shadowShape{Kind: ShadowRoundedRect, Radii: cornerRadiiPx{NW: 6, NE: 4, SE: 8, SW: 2}}
	entry := softShadowEntry(image.Pt(48, 32), shape, 8, 1, 0, 3, col)

	softShadowCache.Lock()
	defer softShadowCache.Unlock()
	if softShadowCache.bytes != entry.bytes {
		t.Fatalf("cache bytes = %d, want %d", softShadowCache.bytes, entry.bytes)
	}
}

func TestPopupSurfaceUsesSoftShadowCache(t *testing.T) {
	resetSoftShadowCacheForTest()
	var ops op.Ops
	gtx := layout.Context{Ops: &ops}

	activeTheme := theme.DefaultTheme()
	DrawSurface(gtx, image.Rect(0, 0, 180, 96), 18, activeTheme.Palette.Overlay, ThemeShadow(activeTheme.Shadows.Overlay, activeTheme.Palette.OverlayShadow, 1))

	softShadowCache.Lock()
	defer softShadowCache.Unlock()
	if got := len(softShadowCache.entries); got == 0 {
		t.Fatal("popup surface did not cache a soft shadow")
	}
}

func BenchmarkBuildSoftShadowEntryRounded(b *testing.B) {
	col := color.NRGBA{R: 40, G: 90, B: 220, A: 120}
	shape := shadowShape{Kind: ShadowRoundedRect, Radii: cornerRadiiPx{NW: 20, NE: 10, SE: 24, SW: 14}}
	size := image.Pt(320, 180)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = buildSoftShadowEntry(size, shape, 22, 4, 0, 8, col)
	}
}

func BenchmarkSoftShadowCacheHit(b *testing.B) {
	resetSoftShadowCacheForTest()
	col := color.NRGBA{R: 40, G: 90, B: 220, A: 120}
	shape := shadowShape{Kind: ShadowRoundedRect, Radii: cornerRadiiPx{NW: 20, NE: 10, SE: 24, SW: 14}}
	size := image.Pt(320, 180)
	_ = softShadowEntry(size, shape, 22, 4, 0, 8, col)

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = softShadowEntry(size, shape, 22, 4, 0, 8, col)
	}
}

func BenchmarkDrawShadowAlphaVariants(b *testing.B) {
	resetSoftShadowCacheForTest()
	var ops op.Ops
	gtx := layout.Context{Ops: &ops}
	shape := shadowShape{Kind: ShadowRoundedRect, Radii: cornerRadiiPx{NW: 20, NE: 10, SE: 24, SW: 14}}
	rect := image.Rect(0, 0, 320, 180)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ops.Reset()
		alpha := uint8(i%255 + 1)
		drawShadow(gtx, rect, shape, BoxShadow{
			Blur:    22,
			Spread:  4,
			OffsetY: 8,
			Color:   color.NRGBA{R: 40, G: 90, B: 220, A: alpha},
		})
	}
}

func resetSoftShadowCacheForTest() {
	softShadowCache.Lock()
	softShadowCache.tick = 0
	softShadowCache.bytes = 0
	softShadowCache.entries = make(map[softShadowCacheKey]softShadowCacheEntry)
	softShadowCache.builds = make(map[softShadowCacheKey]*softShadowBuild)
	softShadowCache.Unlock()
}
