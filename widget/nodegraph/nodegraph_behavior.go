package nodegraph

import "time"

// nodeGraphBehaviorMetrics is the widget-owned interaction and visibility
// contract. Values that are expressed in pixels are converted at the call
// site using the current world scale; values expressed in graph units are
// deliberately independent of the active visual provider.
//
// This table is intentionally separate from theme.NodeGraphMetrics. A theme
// may change how a graph looks, but it must not change what is clickable,
// draggable, focusable, or eligible for culling.
type nodeGraphBehaviorMetrics struct {
	// Culling padding is in graph units and keeps near-viewport work warm.
	nodeCullingPadding float32
	edgeCullingPadding float32

	// The spatial index and final screen visibility check use separate values:
	// the former protects curve bounds in graph space, the latter protects the
	// rendered curve and marker overshoot in screen space.
	curveCullingMinPadding     float32
	curveCullingOvershootRatio float32
	edgeCullingMinPaddingPx    float32
	edgeCullingOvershootRatio  float32

	// Pointer tolerances are screen-pixel contracts, so they remain usable at
	// every graph zoom after conversion through worldScale.
	portHitRadiusPx       float32
	resizeHitSlopPx       float32
	edgeHitQueryPaddingPx float32
	edgeHitRadiusPx       float32

	// A Bezier edge's hit-test route uses this graph-space minimum control
	// distance. It is independent of the provider's visual route metrics.
	edgeHitMinimumDistance float32

	focusRevealMarginPx float32
	clickSlopPx         float32
	doubleClickSlopPx   float32
	doubleClickWait     time.Duration

	// Resize constraints below are neutral interaction guards. Normal node
	// measurement still comes from theme.NodeGraphMetrics.
	resizeMinimumWidth  float32
	resizeMinimumHeight float32
}

var defaultNodeGraphBehavior = nodeGraphBehaviorMetrics{
	nodeCullingPadding:         32,
	edgeCullingPadding:         96,
	curveCullingMinPadding:     64,
	curveCullingOvershootRatio: .1,
	edgeCullingMinPaddingPx:    40,
	edgeCullingOvershootRatio:  .1,
	portHitRadiusPx:            10,
	resizeHitSlopPx:            10,
	edgeHitQueryPaddingPx:      14,
	edgeHitRadiusPx:            10,
	edgeHitMinimumDistance:     32,
	focusRevealMarginPx:        24,
	clickSlopPx:                3,
	doubleClickSlopPx:          6,
	doubleClickWait:            500 * time.Millisecond,
	resizeMinimumWidth:         80,
	resizeMinimumHeight:        48,
}

func (b nodeGraphBehaviorMetrics) worldRadius(pixels, scale float32) float32 {
	if scale <= 0 || !finite(scale) {
		return 0
	}
	return pixels / scale
}

func (b nodeGraphBehaviorMetrics) edgeHitRadiusSquared() float32 {
	return b.edgeHitRadiusPx * b.edgeHitRadiusPx
}
