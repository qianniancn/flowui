package nodegraph

import (
	"image"
	"image/color"
	"math"
	"strings"

	"gioui.org/f32"
	"gioui.org/font"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"github.com/qianniancn/flowui/internal/frame"
	"github.com/qianniancn/flowui/internal/render"
	"github.com/qianniancn/flowui/internal/theme"
)

func drawNodeGraph(ctx *frame.Context, gtx layout.Context, graph resolvedGraph, viewport Viewport, selected, selectedEdges map[string]bool, hoveredNode, focusedNode, graphKey string, showGrid bool, gridPattern GridPattern, gridSize float32, nodesResizable bool, resizeHandles ResizeHandle, minimap, controls, culling bool, panels []Panel, viewportOverlays []ViewportOverlay, nodeToolbars []nodeToolbar, edgeToolbars []edgeToolbar, tokens theme.NodeGraphTheme) {
	activeTheme := frame.ActiveTheme(ctx)
	palette := activeTheme.Palette
	size := gtx.Constraints.Max
	canvas := image.Rectangle{Max: size}
	radius := min(max(gtx.Dp(tokens.CanvasRadius), 0), min(size.X, size.Y)/2)
	paint.FillShape(gtx.Ops, tokens.CanvasBackground, clip.UniformRRect(canvas, radius).Op(gtx.Ops))
	if width := max(gtx.Dp(tokens.CanvasBorderWidth), 0); width > 0 && tokens.CanvasBorder.A != 0 {
		drawGraphRoundedStroke(gtx, canvas, radius, width, tokens.CanvasBorder)
	}
	root := clip.UniformRRect(canvas, radius).Push(gtx.Ops)
	semantic.DescriptionOp("Node graph").Add(gtx.Ops)
	if showGrid {
		drawGrid(gtx, viewport, gridSize, tokens, gridPattern)
	}
	edgeIndices := graphDrawEdgeIndices(graph, viewport, gtx, culling)
	for index := len(edgeIndices) - 1; index >= 0; index-- {
		edge := graph.edges[edgeIndices[index]]
		if !culling || edgeVisible(gtx, viewport, edge) {
			drawEdge(ctx, gtx, viewport, edge, selectedEdges[edge.edge.ID], graphKey, tokens)
		}
		if edge.edge.isAnimated() && ctx != nil {
			ctx.Invalidate()
		}
	}
	nodeIndices := graphDrawNodeIndices(graph, viewport, gtx, culling)
	for index := len(nodeIndices) - 1; index >= 0; index-- {
		node := graph.nodes[nodeIndices[index]]
		drawNode(ctx, gtx, viewport, node, selected[node.node.ID], node.node.ID == hoveredNode, node.node.ID == focusedNode, node.node.isResizable(nodesResizable), node.node.resizeHandleSet(resizeHandles), graphKey, palette, tokens, culling)
	}
	drawGraphViewportOverlays(ctx, gtx, viewport, graphKey, viewportOverlays)
	drawGraphToolbars(ctx, gtx, graph, viewport, graphKey, nodeToolbars, edgeToolbars)
	drawGraphPanels(ctx, gtx, graphKey, panels)
	drawNodeMinimap(gtx, graph, viewport, selected, tokens, minimap)
	drawNodeControls(ctx, gtx, tokens, controls)
	root.Pop()
}

func drawGraphToolbars(ctx *frame.Context, gtx layout.Context, graph resolvedGraph, viewport Viewport, graphKey string, nodeToolbars []nodeToolbar, edgeToolbars []edgeToolbar) {
	overlays := make([]ViewportOverlay, 0, len(nodeToolbars)+len(edgeToolbars))
	for _, toolbar := range nodeToolbars {
		if node, ok := graph.byID[toolbar.nodeID]; ok {
			overlays = append(overlays, ViewportOverlay{Key: "node-toolbar:" + toolbar.key, Position: Point{X: node.position.X + node.size.Width/2, Y: node.position.Y - 12}, Content: toolbar.content})
		}
	}
	for _, toolbar := range edgeToolbars {
		if edge, ok := graph.edgeByID[toolbar.edgeID]; ok {
			mid := graphEdgeMidpoint(edge.edge.edgeTypeValue(), f32.Pt(edge.source.X, edge.source.Y), f32.Pt(edge.target.X, edge.target.Y), 32)
			overlays = append(overlays, ViewportOverlay{Key: "edge-toolbar:" + toolbar.key, Position: Point{X: mid.X, Y: mid.Y}, Content: toolbar.content})
		}
	}
	drawGraphViewportOverlays(ctx, gtx, viewport, graphKey, overlays)
}

func graphDrawNodeIndices(graph resolvedGraph, viewport Viewport, gtx layout.Context, culling bool) []int {
	if !culling {
		indices := make([]int, len(graph.nodes))
		for index := range indices {
			indices[index] = index
		}
		return indices
	}
	return graph.spatial.nodeCandidates(graphViewportBounds(viewport, gtx, 32))
}

func graphDrawEdgeIndices(graph resolvedGraph, viewport Viewport, gtx layout.Context, culling bool) []int {
	if !culling {
		indices := make([]int, len(graph.edges))
		for index := range indices {
			indices[index] = index
		}
		return indices
	}
	return graph.spatial.edgeCandidates(graphViewportBounds(viewport, gtx, 96))
}

func graphViewportBounds(viewport Viewport, gtx layout.Context, padding float32) graphRect {
	pixelsPerDP := graphPixelsPerDP(gtx)
	first := screenToWorld(viewport, f32.Point{}, pixelsPerDP)
	last := screenToWorld(viewport, f32.Pt(float32(gtx.Constraints.Max.X), float32(gtx.Constraints.Max.Y)), pixelsPerDP)
	return graphRect{min: Point{X: first.X - padding, Y: first.Y - padding}, max: Point{X: last.X + padding, Y: last.Y + padding}}
}

func drawGraphViewportOverlays(ctx *frame.Context, gtx layout.Context, viewport Viewport, graphKey string, overlays []ViewportOverlay) {
	if ctx == nil {
		return
	}
	pixelsPerDP := graphPixelsPerDP(gtx)
	for _, overlay := range overlays {
		screen := worldToScreen(viewport, overlay.Position, pixelsPerDP).Round()
		if screen.X < -gtx.Constraints.Max.X || screen.Y < -gtx.Constraints.Max.Y || screen.X > gtx.Constraints.Max.X*2 || screen.Y > gtx.Constraints.Max.Y*2 {
			continue
		}
		overlayGtx := gtx
		overlayGtx.Constraints = layout.Constraints{Max: gtx.Constraints.Max}
		overlayGtx.Metric.PxPerDp = pixelsPerDP * viewport.Zoom
		overlayGtx.Metric.PxPerSp = max(gtx.Metric.PxPerSp, 1) * viewport.Zoom
		macro := op.Record(gtx.Ops)
		restoreGraphKey := frame.PushKey(ctx, graphKey)
		restoreOverlayKey := frame.PushKey(ctx, "viewport-overlay:"+overlay.Key)
		_, placement := frame.TrackOverlayPlacement(ctx, func() layout.Dimensions {
			return overlay.Content.Layout(ctx, overlayGtx)
		})
		placement.PlaceOffset(screen)
		restoreOverlayKey()
		restoreGraphKey()
		call := macro.Stop()
		offset := op.Offset(screen).Push(gtx.Ops)
		call.Add(gtx.Ops)
		offset.Pop()
	}
}

func drawGraphPanels(ctx *frame.Context, gtx layout.Context, graphKey string, panels []Panel) {
	if ctx == nil {
		return
	}
	padding := gtx.Dp(unit.Dp(12))
	for _, panel := range panels {
		panelGtx := gtx
		panelGtx.Constraints = layout.Constraints{Max: gtx.Constraints.Max}
		macro := op.Record(gtx.Ops)
		restoreGraphKey := frame.PushKey(ctx, graphKey)
		restorePanelKey := frame.PushKey(ctx, "panel:"+panel.Key)
		dimensions, placement := frame.TrackOverlayPlacement(ctx, func() layout.Dimensions {
			return panel.Content.Layout(ctx, panelGtx)
		})
		restorePanelKey()
		restoreGraphKey()
		call := macro.Stop()
		position := graphPanelPosition(panel.Position, dimensions.Size, gtx.Constraints.Max, padding)
		placement.PlaceOffset(position)
		offset := op.Offset(position).Push(gtx.Ops)
		call.Add(gtx.Ops)
		offset.Pop()
	}
}

func graphPanelPosition(position PanelPosition, size, canvas image.Point, padding int) image.Point {
	x, y := padding, padding
	switch position {
	case PanelTopCenter, PanelBottomCenter:
		x = max((canvas.X-size.X)/2, padding)
	case PanelTopRight, PanelBottomRight:
		x = max(canvas.X-size.X-padding, padding)
	}
	switch position {
	case PanelBottomLeft, PanelBottomCenter, PanelBottomRight:
		y = max(canvas.Y-size.Y-padding, padding)
	}
	return image.Pt(x, y)
}

func edgeVisible(gtx layout.Context, viewport Viewport, edge resolvedEdge) bool {
	pixelsPerDP := graphPixelsPerDP(gtx)
	from := worldToScreen(viewport, edge.source, pixelsPerDP)
	to := worldToScreen(viewport, edge.target, pixelsPerDP)
	padding := max(float32(40), abs32(to.X-from.X)*.1)
	minX, maxX := min(from.X, to.X)-padding, max(from.X, to.X)+padding
	minY, maxY := min(from.Y, to.Y)-padding, max(from.Y, to.Y)+padding
	canvas := gtx.Constraints.Max
	return maxX >= 0 && maxY >= 0 && minX <= float32(canvas.X) && minY <= float32(canvas.Y)
}

func drawGrid(gtx layout.Context, viewport Viewport, gridSize float32, tokens theme.NodeGraphTheme, pattern GridPattern) {
	if gridSize <= 0 || !finite(gridSize) {
		return
	}
	scale := worldScale(viewport, graphPixelsPerDP(gtx))
	step := graphGridStep(gridSize, scale)
	if step <= 0 {
		return
	}
	colorValue := tokens.GridColor
	colorValue.A = byte(float32(colorValue.A)*min(max(tokens.GridOpacity, 0), 1) + .5)
	if colorValue.A == 0 {
		return
	}
	startX := -float32(math.Mod(float64(viewport.Origin.X*scale), float64(step)))
	startY := -float32(math.Mod(float64(viewport.Origin.Y*scale), float64(step)))
	if startX > 0 {
		startX -= step
	}
	if startY > 0 {
		startY -= step
	}
	if pattern == GridDots {
		drawDotGrid(gtx, startX, startY, step, colorValue)
		return
	}
	path := clip.Path{}
	path.Begin(gtx.Ops)
	for x := startX; x < float32(gtx.Constraints.Max.X); x += step {
		path.MoveTo(f32.Pt(x, 0))
		path.LineTo(f32.Pt(x, float32(gtx.Constraints.Max.Y)))
	}
	for y := startY; y < float32(gtx.Constraints.Max.Y); y += step {
		path.MoveTo(f32.Pt(0, y))
		path.LineTo(f32.Pt(float32(gtx.Constraints.Max.X), y))
	}
	stroke := clip.Stroke{Path: path.End(), Width: float32(max(gtx.Dp(unit.Dp(1)), 1))}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, colorValue)
	stroke.Pop()
}

// graphGridStep skips overly dense grid intervals at low zoom while keeping
// the visible grid aligned to an integer multiple of the configured spacing.
func graphGridStep(gridSize, scale float32) float32 {
	step := gridSize * scale
	if step <= 0 || !finite(step) {
		return 0
	}
	const minimumPixelStep = float32(8)
	if step < minimumPixelStep {
		step *= float32(math.Ceil(float64(minimumPixelStep / step)))
	}
	return step
}

func drawDotGrid(gtx layout.Context, startX, startY, step float32, colorValue color.NRGBA) {
	for y := startY; y < float32(gtx.Constraints.Max.Y); y += step {
		for x := startX; x < float32(gtx.Constraints.Max.X); x += step {
			center := image.Pt(int(math.Round(float64(x))), int(math.Round(float64(y))))
			paint.FillShape(gtx.Ops, colorValue, clip.Ellipse(image.Rect(center.X-1, center.Y-1, center.X+1, center.Y+1)).Op(gtx.Ops))
		}
	}
}

func drawEdge(ctx *frame.Context, gtx layout.Context, viewport Viewport, edge resolvedEdge, selected bool, graphKey string, tokens theme.NodeGraphTheme) {
	pixelsPerDP := graphPixelsPerDP(gtx)
	from := worldToScreen(viewport, edge.source, pixelsPerDP)
	to := worldToScreen(viewport, edge.target, pixelsPerDP)
	colorValue := edge.edge.Color
	if colorValue.A == 0 {
		colorValue = tokens.EdgeColor
	}
	if selected {
		colorValue = tokens.SelectedEdgeColor
	}
	width := edge.edge.width
	if width <= 0 {
		width = 1
	}
	width = max(float32(gtx.Dp(unit.Dp(width)))*viewport.Zoom, 1)
	minimumDistance := 32 * worldScale(viewport, pixelsPerDP)
	animated := edge.edge.isAnimated()
	phase := float32(0)
	if animated {
		phase = float32(gtx.Now.UnixNano()%1_000_000_000) / 1_000_000_000 * max(width*8, 16)
	}
	drawGraphEdge(gtx, edge.edge.edgeTypeValue(), from, to, width, colorValue, minimumDistance, edge.edge.isDashed() || animated, phase)
	drawEdgeMarker(gtx, edge.edge.sourceMark, edge.edge.edgeTypeValue(), from, to, true, colorValue)
	drawEdgeMarker(gtx, edge.edge.targetMark, edge.edge.edgeTypeValue(), from, to, false, colorValue)
	if edge.edge.label != "" || edge.edge.labelContent != nil {
		midpoint := graphEdgeMidpoint(edge.edge.edgeTypeValue(), from, to, minimumDistance)
		labelRect := image.Rect(int(math.Round(float64(midpoint.X)))-60, int(math.Round(float64(midpoint.Y)))-14, int(math.Round(float64(midpoint.X)))+60, int(math.Round(float64(midpoint.Y)))+14)
		if edge.edge.labelContent == nil {
			drawGraphCenteredText(ctx, gtx, edge.edge.label, unit.Sp(min(max(10*viewport.Zoom, 8), 14)), font.Normal, tokens.NodeForeground, labelRect)
			return
		}
		visible := labelRect.Intersect(image.Rectangle{Max: gtx.Constraints.Max})
		if visible.Empty() {
			return
		}
		restoreGraphKey := frame.PushKey(ctx, graphKey)
		restoreEdgeKey := frame.PushKey(ctx, "edge:"+edge.edge.ID)
		contentGtx := gtx
		contentGtx.Constraints = layout.Exact(labelRect.Size())
		offset := op.Offset(labelRect.Min).Push(gtx.Ops)
		edge.edge.labelContent.Layout(ctx, contentGtx)
		offset.Pop()
		restoreEdgeKey()
		restoreGraphKey()
	}
}

func drawConnectionPreview(gtx layout.Context, viewport Viewport, state *graphState, tokens theme.NodeGraphTheme) {
	from := worldToScreen(viewport, state.connectionSource.point, graphPixelsPerDP(gtx))
	to := state.connectionCurrent
	colorValue := tokens.EdgeColor
	if state.connectionTarget.endpoint.NodeID != "" {
		to = worldToScreen(viewport, state.connectionTarget.point, graphPixelsPerDP(gtx))
	}
	if state.connectionMode == connectionReconnectSource {
		from, to = to, from
	}
	drawGraphBezier(gtx, from, to, max(float32(gtx.Dp(unit.Dp(1)))*viewport.Zoom, 1), colorValue, 32)
}

func drawGraphBezier(gtx layout.Context, from, to f32.Point, width float32, colorValue color.NRGBA, minimumDistance float32) {
	distance := max(abs32(to.X-from.X)*.5, minimumDistance)
	path := clip.Path{}
	path.Begin(gtx.Ops)
	path.MoveTo(from)
	path.CubeTo(f32.Pt(from.X+distance, from.Y), f32.Pt(to.X-distance, to.Y), to)
	stroke := clip.Stroke{Path: path.End(), Width: width}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, colorValue)
	stroke.Pop()
}

func drawGraphEdge(gtx layout.Context, edgeType EdgeType, from, to f32.Point, width float32, colorValue color.NRGBA, minimumDistance float32, dashed bool, phase float32) {
	if dashed {
		points := graphEdgePolyline(edgeType, from, to, minimumDistance)
		drawGraphDashedPolyline(gtx, points, width, colorValue, phase)
		return
	}
	if edgeType == EdgeBezier {
		drawGraphBezier(gtx, from, to, width, colorValue, minimumDistance)
		return
	}
	if edgeType == EdgeSmoothStep {
		drawGraphSmoothStep(gtx, from, to, width, colorValue)
		return
	}
	drawGraphPolyline(gtx, graphEdgePolyline(edgeType, from, to, minimumDistance), width, colorValue)
}

func drawGraphPolyline(gtx layout.Context, points []f32.Point, width float32, colorValue color.NRGBA) {
	if len(points) < 2 {
		return
	}
	path := clip.Path{}
	path.Begin(gtx.Ops)
	path.MoveTo(points[0])
	for _, point := range points[1:] {
		path.LineTo(point)
	}
	stroke := clip.Stroke{Path: path.End(), Width: width}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, colorValue)
	stroke.Pop()
}

func drawGraphDashedPolyline(gtx layout.Context, points []f32.Point, width float32, colorValue color.NRGBA, phase float32) {
	dash, gap := max(width*4, 8), max(width*2, 5)
	cycle := dash + gap
	offset := float32(math.Mod(float64(phase), float64(cycle)))
	for index := 1; index < len(points); index++ {
		from, to := points[index-1], points[index]
		delta := to.Sub(from)
		length := float32(math.Hypot(float64(delta.X), float64(delta.Y)))
		if length <= 0 {
			continue
		}
		progress := -offset
		for progress < length {
			start := max(progress, 0)
			end := min(progress+dash, length)
			if end > start {
				a := f32.Pt(from.X+delta.X*start/length, from.Y+delta.Y*start/length)
				b := f32.Pt(from.X+delta.X*end/length, from.Y+delta.Y*end/length)
				drawGraphPolyline(gtx, []f32.Point{a, b}, width, colorValue)
			}
			progress += cycle
		}
		offset = float32(math.Mod(float64(offset+length), float64(cycle)))
	}
}

func drawGraphSmoothStep(gtx layout.Context, from, to f32.Point, width float32, colorValue color.NRGBA) {
	deltaX, deltaY := to.X-from.X, to.Y-from.Y
	if abs32(deltaY) < 1 {
		drawGraphPolyline(gtx, []f32.Point{from, to}, width, colorValue)
		return
	}
	midX := (from.X + to.X) / 2
	directionX, directionY := float32(1), float32(1)
	if deltaX < 0 {
		directionX = -1
	}
	if deltaY < 0 {
		directionY = -1
	}
	radius := min(float32(8), min(abs32(midX-from.X)/2, abs32(deltaY)/2))
	path := clip.Path{}
	path.Begin(gtx.Ops)
	path.MoveTo(from)
	path.LineTo(f32.Pt(midX-directionX*radius, from.Y))
	path.QuadTo(f32.Pt(midX, from.Y), f32.Pt(midX, from.Y+directionY*radius))
	path.LineTo(f32.Pt(midX, to.Y-directionY*radius))
	path.QuadTo(f32.Pt(midX, to.Y), f32.Pt(midX+directionX*radius, to.Y))
	path.LineTo(to)
	stroke := clip.Stroke{Path: path.End(), Width: width}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, colorValue)
	stroke.Pop()
}

func graphEdgeMidpoint(edgeType EdgeType, from, to f32.Point, minimumDistance float32) f32.Point {
	if edgeType == EdgeBezier {
		return graphBezierPoint(from, to, minimumDistance, .5)
	}
	return f32.Pt((from.X+to.X)/2, (from.Y+to.Y)/2)
}

func drawEdgeMarker(gtx layout.Context, marker EdgeMarker, edgeType EdgeType, from, to f32.Point, source bool, colorValue color.NRGBA) {
	if marker != MarkerArrow {
		return
	}
	points := graphEdgePolyline(edgeType, from, to, 32)
	if len(points) < 2 {
		return
	}
	endpoint, neighbor := to, points[len(points)-2]
	if source {
		endpoint, neighbor = from, points[1]
	}
	direction := endpoint.Sub(neighbor)
	length := float32(math.Hypot(float64(direction.X), float64(direction.Y)))
	if length <= 0 {
		return
	}
	direction = direction.Mul(1 / length)
	perpendicular := f32.Pt(-direction.Y, direction.X)
	size := max(float32(6), 5)
	base := endpoint.Sub(direction.Mul(size))
	path := clip.Path{}
	path.Begin(gtx.Ops)
	path.MoveTo(endpoint)
	path.LineTo(base.Add(perpendicular.Mul(size * .55)))
	path.LineTo(base.Sub(perpendicular.Mul(size * .55)))
	path.Close()
	paint.FillShape(gtx.Ops, colorValue, clip.Outline{Path: path.End()}.Op())
}

func drawSelectionBox(gtx layout.Context, first, second f32.Point, tokens theme.NodeGraphTheme) {
	rect := image.Rect(
		int(math.Floor(float64(min(first.X, second.X)))),
		int(math.Floor(float64(min(first.Y, second.Y)))),
		int(math.Ceil(float64(max(first.X, second.X)))),
		int(math.Ceil(float64(max(first.Y, second.Y)))),
	)
	if rect.Empty() {
		return
	}
	paint.FillShape(gtx.Ops, tokens.SelectionFill, clip.Rect(rect).Op())
	drawDottedRect(gtx, rect, tokens.SelectionBorder)
}

func drawNode(ctx *frame.Context, gtx layout.Context, viewport Viewport, node resolvedNode, selected, hovered, focused, resizable bool, resizeHandles ResizeHandle, graphKey string, palette theme.Palette, tokens theme.NodeGraphTheme, culling bool) {
	pixelsPerDP := graphPixelsPerDP(gtx)
	position := worldToScreen(viewport, node.position, pixelsPerDP).Round()
	scale := worldScale(viewport, pixelsPerDP)
	size := image.Pt(
		max(int(math.Round(float64(node.size.Width*scale))), 1),
		max(int(math.Round(float64(node.size.Height*scale))), 1),
	)
	rect := image.Rectangle{Min: position, Max: position.Add(size)}
	if culling && rect.Intersect(image.Rectangle{Max: gtx.Constraints.Max}).Empty() {
		return
	}
	radius := min(max(int(math.Round(float64(3*scale))), 1), min(size.X, size.Y)/2)
	if hovered && !selected {
		shadow := render.BoxShadow{
			OffsetY: 1,
			Blur:    4,
			Spread:  1,
			Color:   graphShadowColor(palette),
		}
		render.DrawShadow(gtx, rect, render.RoundedShadowCorners(3, 3, 3, 3), shadow)
	}
	paint.FillShape(gtx.Ops, tokens.NodeBackground, clip.UniformRRect(rect, radius).Op(gtx.Ops))
	headerHeight := min(max(int(math.Round(float64(nodeHeaderHeight*scale))), 1), size.Y)
	border := tokens.NodeBorder
	borderWidth := max(int(math.Round(float64(scale))), 1)
	if selected {
		border = tokens.SelectedNodeBorder
	}
	// Keep the node stroke inside its bounds. This avoids overlapping the
	// canvas stroke when a node is resized flush with the viewport edge.
	strokeInset := max((borderWidth+1)/2, 1)
	strokeRect := rect
	strokeRadius := radius
	if rect.Dx() > strokeInset*2 && rect.Dy() > strokeInset*2 {
		strokeRect = rect.Inset(strokeInset)
		strokeRadius = max(radius-strokeInset, 0)
	}
	drawGraphRoundedStroke(gtx, strokeRect, strokeRadius, borderWidth, border)

	clipped := clip.Rect(rect).Push(gtx.Ops)
	semantic.Button.Add(gtx.Ops)
	semantic.LabelOp(node.node.Title).Add(gtx.Ops)
	semantic.SelectedOp(selected).Add(gtx.Ops)
	semantic.EnabledOp(gtx.Enabled()).Add(gtx.Ops)
	if focused {
		semantic.DescriptionOp("Focused node").Add(gtx.Ops)
	}
	padding := max(int(math.Round(float64(10*scale))), 4)
	if node.node.content == nil {
		title := node.node.Title
		if title == "" {
			title = node.node.ID
		}
		drawGraphCenteredText(ctx, gtx, title, unit.Sp(min(max(12*viewport.Zoom, 9), 16)), font.Normal, tokens.NodeForeground, image.Rect(rect.Min.X+padding, rect.Min.Y, rect.Max.X-padding, rect.Min.Y+headerHeight))
	} else {
		inner := rect.Inset(padding)
		if !inner.Empty() {
			restoreGraphKey := frame.PushKey(ctx, graphKey)
			restoreNodeKey := frame.PushKey(ctx, "node:"+node.node.ID)
			contentGtx := gtx
			contentGtx.Constraints = layout.Exact(inner.Size())
			contentScale := max(viewport.Zoom, 0.01)
			contentGtx.Metric.PxPerDp = graphPixelsPerDP(gtx) * contentScale
			contentGtx.Metric.PxPerSp = max(gtx.Metric.PxPerSp, 1) * contentScale
			offset := op.Offset(inner.Min).Push(gtx.Ops)
			node.node.content.Layout(ctx, contentGtx)
			offset.Pop()
			restoreNodeKey()
			restoreGraphKey()
		}
	}
	clipped.Pop()

	// Ports and resize handles intentionally cross the node boundary. The node
	// content above remains clipped, while these interaction targets do not.
	for _, resolvedPort := range node.inputs {
		drawNodePort(ctx, gtx, viewport, resolvedPort, rect, size, scale, padding, tokens)
	}
	for _, resolvedPort := range node.outputs {
		drawNodePort(ctx, gtx, viewport, resolvedPort, rect, size, scale, padding, tokens)
	}
	if focused {
		drawDottedRect(gtx, rect.Inset(max(int(math.Round(float64(2*scale))), 1)), tokens.SelectionBorder)
	}
	if selected && resizable && resizeHandles != 0 {
		drawNodeResizeHandles(gtx, rect, scale, resizeHandles, tokens.SelectedNodeBorder)
	}
}

func drawNodeResizeHandles(gtx layout.Context, rect image.Rectangle, scale float32, handles ResizeHandle, colorValue color.NRGBA) {
	radius := max(int(math.Round(float64(4*scale))), 3)
	for _, handle := range resizeHandleOrder {
		if handles&handle == 0 {
			continue
		}
		center := resizeHandleScreenPoint(rect, handle)
		handleRect := image.Rect(center.X-radius, center.Y-radius, center.X+radius, center.Y+radius)
		paint.FillShape(gtx.Ops, color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}, clip.Rect(handleRect).Op())
		drawGraphRoundedStroke(gtx, handleRect, 1, max(int(math.Round(float64(scale))), 1), colorValue)
	}
}

func drawNodePort(ctx *frame.Context, gtx layout.Context, viewport Viewport, resolvedPort resolvedPort, rect image.Rectangle, size image.Point, scale float32, padding int, tokens theme.NodeGraphTheme) {
	port := resolvedPort.port
	anchor := worldToScreen(viewport, resolvedPort.point, graphPixelsPerDP(gtx))
	drawGraphPort(gtx, anchor, scale, port, tokens.PortColor, tokens.PortBorder)
	label := graphPortLabel(port)
	if strings.TrimSpace(label) == "" {
		return
	}
	textSize := unit.Sp(min(max(10*viewport.Zoom, 8), 14))
	textHeight := max(int(math.Round(float64(7*scale))), 3)
	maxWidth := max(size.X/2-padding, 1)
	position := image.Pt(rect.Min.X+padding, int(math.Round(float64(anchor.Y)))-textHeight)
	switch resolvedPort.position {
	case HandleRight:
		position = image.Pt(rect.Min.X+size.X/2, int(math.Round(float64(anchor.Y)))-textHeight)
	case HandleTop:
		position = image.Pt(int(math.Round(float64(anchor.X)))-maxWidth/2, rect.Min.Y+padding)
	case HandleBottom:
		position = image.Pt(int(math.Round(float64(anchor.X)))-maxWidth/2, rect.Max.Y-padding-textHeight)
	}
	drawGraphText(ctx, gtx, label, textSize, font.Normal, tokens.NodeMutedForeground, position, maxWidth)
}

func graphPortLabel(port Port) string {
	if port.Label != "" {
		return port.Label
	}
	return port.ID
}

func drawGraphPort(gtx layout.Context, center f32.Point, scale float32, port Port, fallback, border color.NRGBA) {
	radius := max(int(math.Round(float64(3*scale))), 2)
	position := center.Round()
	rect := image.Rect(position.X-radius, position.Y-radius, position.X+radius, position.Y+radius)
	colorValue := port.Color
	if colorValue.A == 0 {
		colorValue = fallback
	}
	paint.FillShape(gtx.Ops, border, clip.Ellipse(rect).Op(gtx.Ops))
	inset := max(int(math.Round(float64(scale))), 1)
	rect = rect.Inset(inset)
	if rect.Empty() {
		return
	}
	paint.FillShape(gtx.Ops, colorValue, clip.Ellipse(rect).Op(gtx.Ops))
}

func graphShadowColor(palette theme.Palette) color.NRGBA {
	shadow := palette.OverlayShadow
	shadow.A = 20
	return shadow
}

func drawGraphText(ctx *frame.Context, gtx layout.Context, value string, size unit.Sp, weight font.Weight, colorValue color.NRGBA, position image.Point, maxWidth int) {
	if value == "" || maxWidth <= 0 {
		return
	}
	textGtx := gtx
	textGtx.Constraints = layout.Constraints{Max: image.Pt(maxWidth, max(gtx.Constraints.Max.Y-position.Y, 1))}
	offset := op.Offset(position).Push(gtx.Ops)
	label := material.Label(frame.ActiveMaterial(ctx), size, value)
	label.Color = colorValue
	label.Font.Weight = weight
	label.MaxLines = 1
	label.Layout(textGtx)
	offset.Pop()
}

func drawGraphCenteredText(ctx *frame.Context, gtx layout.Context, value string, size unit.Sp, weight font.Weight, colorValue color.NRGBA, rect image.Rectangle) {
	if value == "" || rect.Empty() {
		return
	}
	textGtx := gtx
	textGtx.Constraints = layout.Constraints{Max: rect.Size()}
	macro := op.Record(gtx.Ops)
	label := material.Label(frame.ActiveMaterial(ctx), size, value)
	label.Color = colorValue
	label.Font.Weight = weight
	label.MaxLines = 1
	dimensions := label.Layout(textGtx)
	call := macro.Stop()
	position := image.Pt(rect.Min.X+max((rect.Dx()-dimensions.Size.X)/2, 0), rect.Min.Y+max((rect.Dy()-dimensions.Size.Y)/2, 0))
	offset := op.Offset(position).Push(gtx.Ops)
	call.Add(gtx.Ops)
	offset.Pop()
}

func drawDottedRect(gtx layout.Context, rect image.Rectangle, colorValue color.NRGBA) {
	if rect.Empty() || colorValue.A == 0 {
		return
	}
	thickness := max(gtx.Dp(unit.Dp(1)), 1)
	segment := max(gtx.Dp(unit.Dp(2)), 1)
	gap := max(gtx.Dp(unit.Dp(2)), 1)
	for x := rect.Min.X; x < rect.Max.X; x += segment + gap {
		end := min(x+segment, rect.Max.X)
		paint.FillShape(gtx.Ops, colorValue, clip.Rect(image.Rect(x, rect.Min.Y, end, min(rect.Min.Y+thickness, rect.Max.Y))).Op())
		paint.FillShape(gtx.Ops, colorValue, clip.Rect(image.Rect(x, max(rect.Max.Y-thickness, rect.Min.Y), end, rect.Max.Y)).Op())
	}
	for y := rect.Min.Y + segment; y < rect.Max.Y-segment; y += segment + gap {
		end := min(y+segment, rect.Max.Y)
		paint.FillShape(gtx.Ops, colorValue, clip.Rect(image.Rect(rect.Min.X, y, min(rect.Min.X+thickness, rect.Max.X), end)).Op())
		paint.FillShape(gtx.Ops, colorValue, clip.Rect(image.Rect(max(rect.Max.X-thickness, rect.Min.X), y, rect.Max.X, end)).Op())
	}
}

func drawGraphRoundedStroke(gtx layout.Context, rect image.Rectangle, radius, width int, colorValue color.NRGBA) {
	if rect.Empty() || width <= 0 || colorValue.A == 0 {
		return
	}
	stroke := clip.Stroke{Path: clip.UniformRRect(rect, radius).Path(gtx.Ops), Width: float32(width)}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, colorValue)
	stroke.Pop()
}
