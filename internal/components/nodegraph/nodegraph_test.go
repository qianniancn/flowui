package nodegraph

import (
	"image"
	"image/color"
	"math"
	"testing"

	"gioui.org/f32"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/flowui/internal/frame"
	"github.com/qianniancn/flowui/internal/locale"
)

func TestViewportRoundTrip(t *testing.T) {
	viewport := Viewport{Origin: Point{X: 32, Y: -16}, Zoom: 1.25}
	point := Point{X: 180, Y: 96}
	screen := worldToScreen(viewport, point, 1.5)
	got := screenToWorld(viewport, screen, 1.5)
	if !pointsClose(got, point) {
		t.Fatalf("round trip = %#v, want %#v", got, point)
	}
}

func TestNewUsesFlowUISurfaceDefaults(t *testing.T) {
	widget := New("graph", Graph{})
	if !widget.showGrid {
		t.Fatal("node graph grid is disabled by default")
	}
	if widget.gridPattern != GridLines {
		t.Fatalf("node graph grid pattern = %v, want lines", widget.gridPattern)
	}
}

func TestGraphGridStepRemainsVisibleAtLowZoom(t *testing.T) {
	if got := graphGridStep(16, .25); got < 8 {
		t.Fatalf("low-zoom grid step = %v, want at least 8 pixels", got)
	}
	if got := graphGridStep(16, 1); got != 16 {
		t.Fatalf("normal grid step = %v, want 16 pixels", got)
	}
}

func TestGridAppearanceOverrides(t *testing.T) {
	colorValue := color.NRGBA{R: 0x32, G: 0x78, B: 0xc8, A: 0xff}
	widget := New("graph", Graph{}).GridColor(colorValue).GridOpacity(.45)
	if !widget.hasGridColor || widget.gridColor != colorValue {
		t.Fatalf("grid color override = %#v, enabled=%v", widget.gridColor, widget.hasGridColor)
	}
	if !widget.hasGridOpacity || widget.gridOpacity != .45 {
		t.Fatalf("grid opacity override = %v, enabled=%v", widget.gridOpacity, widget.hasGridOpacity)
	}
}

func TestZoomAtKeepsPointerWorldPosition(t *testing.T) {
	viewport := Viewport{Origin: Point{X: 20, Y: 10}, Zoom: 1}
	pointer := f32.Pt(240, 120)
	before := screenToWorld(viewport, pointer, 1)
	after := zoomAt(viewport, 1.75, pointer, 1)
	got := screenToWorld(after, pointer, 1)
	if !pointsClose(got, before) {
		t.Fatalf("world point after zoom = %#v, want %#v", got, before)
	}
}

func TestFitGraphViewportContainsAndCentersNodes(t *testing.T) {
	graph := resolveGraph(Graph{Nodes: []Node{
		NewNode("first", "First", Point{X: 100, Y: 100}),
		NewNode("second", "Second", Point{X: 500, Y: 200}),
	}})
	viewport, ok := fitGraphViewport(graph, image.Pt(800, 400), 1, .1, .25, 2)
	if !ok {
		t.Fatal("fit viewport was not calculated")
	}
	first := worldToScreen(viewport, graph.nodes[0].node.Position, 1)
	second := worldToScreen(viewport, Point{X: graph.nodes[1].node.Position.X + graph.nodes[1].size.Width, Y: graph.nodes[1].node.Position.Y + graph.nodes[1].size.Height}, 1)
	if first.X < 80 || first.Y < 40 || second.X > 720 || second.Y > 360 {
		t.Fatalf("fitted node bounds = %v to %v", first, second)
	}
	if !pointsClose(screenToWorld(viewport, f32.Pt(400, 200), 1), Point{X: 375, Y: 182}) {
		t.Fatalf("viewport is not centered: %#v", viewport)
	}
}

func TestResolveGraphUsesPortAnchors(t *testing.T) {
	graph := Graph{
		Nodes: []Node{
			NewNode("source", "Source", Point{X: 20, Y: 30}).Outputs(NewPort("value", "Value")),
			NewNode("target", "Target", Point{X: 320, Y: 40}).Inputs(NewPort("input", "Input")),
		},
		Edges: []Edge{NewEdge("edge", NewEndpoint("source", "value"), NewEndpoint("target", "input"))},
	}
	resolved := resolveGraph(graph)
	if len(resolved.edges) != 1 {
		t.Fatalf("edges = %d, want 1", len(resolved.edges))
	}
	edge := resolved.edges[0]
	if edge.source.X <= graph.Nodes[0].Position.X || edge.target.X != graph.Nodes[1].Position.X {
		t.Fatalf("anchors = %#v -> %#v", edge.source, edge.target)
	}
	if edge.source.Y != graph.Nodes[0].Position.Y+nodeHeaderHeight+nodeVerticalPad+nodePortRowHeight/2 || edge.target.Y != graph.Nodes[1].Position.Y+nodeHeaderHeight+nodeVerticalPad+nodePortRowHeight/2 {
		t.Fatalf("first-row anchors = %#v -> %#v", edge.source, edge.target)
	}
}

func TestResolveGraphUsesConfiguredHandlePositions(t *testing.T) {
	node := NewNode("node", "Node", Point{X: 40, Y: 30}).WithSize(200, 100).
		Inputs(
			NewPort("left", "Left"),
			NewPort("top", "Top").Position(HandleTop),
		).
		Outputs(
			NewPort("right", "Right"),
			NewPort("bottom", "Bottom").Position(HandleBottom),
		)
	resolved := resolveGraph(Graph{Nodes: []Node{node}}).nodes[0]
	left, _ := resolved.port(false, "left")
	top, _ := resolved.port(false, "top")
	right, _ := resolved.port(true, "right")
	bottom, _ := resolved.port(true, "bottom")
	if left.point.X != node.Position.X || right.point.X != node.Position.X+resolved.size.Width {
		t.Fatalf("left/right anchors = %#v, %#v", left.point, right.point)
	}
	if top.point.Y != node.Position.Y || bottom.point.Y != node.Position.Y+resolved.size.Height {
		t.Fatalf("top/bottom anchors = %#v, %#v", top.point, bottom.point)
	}
	if top.point.X <= node.Position.X || bottom.point.X <= node.Position.X {
		t.Fatalf("horizontal handles were not distributed: %#v, %#v", top.point, bottom.point)
	}
}

func TestCustomNodeContentContributesToAutomaticSize(t *testing.T) {
	ctx := frame.New(nil, nil, locale.LanguageEnglish)
	frame.BeginFrameWithViewport(ctx, image.Pt(640, 360))
	var ops op.Ops
	gtx := layout.Context{Constraints: layout.Constraints{Max: image.Pt(640, 360)}, Ops: &ops}
	content := frame.WidgetFunc(func(_ *frame.Context, _ layout.Context) layout.Dimensions {
		return layout.Dimensions{Size: image.Pt(240, 80)}
	})
	measured := resolvedNodeSizeMeasured(ctx, gtx, NewNode("content", "Content", Point{}).Content(content))
	if measured.Width < 256 || measured.Height < 96 {
		t.Fatalf("content size = %#v, want at least 256x96", measured)
	}
	overridden := resolvedNodeSizeMeasured(ctx, gtx, NewNode("fixed", "Fixed", Point{}).WithSize(120, 120).Content(content))
	if overridden.Width != 120 || overridden.Height != 120 {
		t.Fatalf("explicit node size = %#v, want 120x120", overridden)
	}
}

func TestResolvedGraphFindsTopmostNode(t *testing.T) {
	graph := Graph{Nodes: []Node{
		NewNode("back", "Back", Point{X: 0, Y: 0}),
		NewNode("front", "Front", Point{X: 40, Y: 40}),
	}}
	if got := resolveGraph(graph).nodeAt(Point{X: 60, Y: 60}); got != "front" {
		t.Fatalf("node at overlap = %q, want front", got)
	}
}

func TestResolvedNodeSizeHonorsExplicitWidth(t *testing.T) {
	node := NewNode("node", "Node", Point{}).WithSize(120, 60).Inputs(NewPort("a", "A"), NewPort("b", "B"))
	size := resolvedNodeSize(node)
	if size.Width != 120 {
		t.Fatalf("width = %v, want 120", size.Width)
	}
	minimumHeight := nodeHeaderHeight + nodeVerticalPad*2 + 2*nodePortRowHeight
	if size.Height != minimumHeight {
		t.Fatalf("height = %v, want %v", size.Height, minimumHeight)
	}
}

func TestNormalizeViewportRepairsInvalidValues(t *testing.T) {
	got := normalizeViewport(Viewport{Origin: Point{X: float32(math.NaN()), Y: float32(math.Inf(1))}, Zoom: -1}, .5, 2)
	if got.Origin != (Point{}) || got.Zoom != 1 {
		t.Fatalf("normalized viewport = %#v", got)
	}
}

func TestApplyNodeChanges(t *testing.T) {
	nodes := []Node{
		NewNode("first", "First", Point{X: 20, Y: 40}),
		NewNode("second", "Second", Point{X: 120, Y: 80}),
	}
	updated := ApplyNodeChanges(nodes, []NodeChange{
		{ID: "first", Kind: NodeChangeSelection, Selected: true},
		{ID: "second", Kind: NodeChangePosition, Position: Point{X: 160, Y: 96}},
		{ID: "first", Kind: NodeChangeSize, Size: Size{Width: 180, Height: 100}},
		{ID: "missing", Kind: NodeChangeSelection, Selected: true},
	})
	if nodes[0].Selected || nodes[1].Position != (Point{X: 120, Y: 80}) {
		t.Fatal("ApplyNodeChanges mutated the input slice")
	}
	if !updated[0].Selected || updated[0].Size != (Size{Width: 180, Height: 100}) || updated[1].Position != (Point{X: 160, Y: 96}) {
		t.Fatalf("updated nodes = %#v", updated)
	}
}

func TestApplyEdgeChanges(t *testing.T) {
	edges := []Edge{
		NewEdge("first", NewEndpoint("source", "out"), NewEndpoint("target", "in")),
		NewEdge("second", NewEndpoint("source", "other"), NewEndpoint("target", "next")),
	}
	updated := ApplyEdgeChanges(edges, []EdgeChange{
		{ID: "first", Kind: EdgeChangeSelection, Selected: true},
		{ID: "second", Kind: EdgeChangeRemove},
	})
	if edges[0].Selected || len(edges) != 2 {
		t.Fatal("ApplyEdgeChanges mutated the input slice")
	}
	if len(updated) != 1 || updated[0].ID != "first" || !updated[0].Selected {
		t.Fatalf("updated edges = %#v", updated)
	}
}

func TestReconnectEdgePreservesIdentityAndInput(t *testing.T) {
	edges := []Edge{
		NewEdge("first", NewEndpoint("source", "out"), NewEndpoint("target", "in")),
		NewEdge("second", NewEndpoint("other", "out"), NewEndpoint("next", "in")),
	}
	updated := ReconnectEdge(edges[0], Connection{
		Source: NewEndpoint("replacement-source", "out"),
		Target: NewEndpoint("replacement-target", "in"),
	}, edges)
	if edges[0].Source != (Endpoint{NodeID: "source", PortID: "out"}) || edges[0].Target != (Endpoint{NodeID: "target", PortID: "in"}) {
		t.Fatal("ReconnectEdge mutated the input slice")
	}
	if updated[0].ID != "first" || updated[0].Source.NodeID != "replacement-source" || updated[0].Target.NodeID != "replacement-target" {
		t.Fatalf("reconnected edge = %#v", updated[0])
	}
	if updated[1] != edges[1] {
		t.Fatalf("unrelated edge changed: %#v", updated[1])
	}
}

func TestResolvedGraphFindsBezierEdge(t *testing.T) {
	graph := resolveGraph(Graph{
		Nodes: []Node{
			NewNode("source", "Source", Point{X: 40, Y: 80}).Outputs(NewPort("out", "Out")),
			NewNode("target", "Target", Point{X: 360, Y: 80}).Inputs(NewPort("in", "In")),
		},
		Edges: []Edge{NewEdge("edge", NewEndpoint("source", "out"), NewEndpoint("target", "in"))},
	})
	widget := New("graph", Graph{})
	if got := widget.edgeAt(graph, f32.Pt(275, 124), Viewport{Zoom: 1}, 1); got != "edge" {
		t.Fatalf("edge at center = %q, want edge", got)
	}
	if got := widget.edgeAt(graph, f32.Pt(275, 160), Viewport{Zoom: 1}, 1); got != "" {
		t.Fatalf("edge at offset = %q, want none", got)
	}
	if got := widget.EdgesSelectable(false).edgeAt(graph, f32.Pt(275, 124), Viewport{Zoom: 1}, 1); got != "edge" {
		t.Fatalf("non-selectable edge hit = %q, want edge", got)
	}
}

func TestEdgeGeometryTypesUseMatchingHitPaths(t *testing.T) {
	graph := resolveGraph(Graph{
		Nodes: []Node{
			NewNode("source", "Source", Point{X: 40, Y: 40}).Outputs(NewPort("out", "Out")),
			NewNode("target", "Target", Point{X: 300, Y: 180}).Inputs(NewPort("in", "In")),
		},
		Edges: []Edge{NewEdge("step", NewEndpoint("source", "out"), NewEndpoint("target", "in")).Type(EdgeStep)},
	})
	widget := New("graph", Graph{})
	if got := widget.edgeAt(graph, f32.Pt(245, 150), Viewport{Zoom: 1}, 1); got != "step" {
		t.Fatalf("step edge hit = %q, want step", got)
	}
	straight := graph
	straight.edges[0].edge = straight.edges[0].edge.Type(EdgeStraight)
	if got := widget.edgeAt(straight, f32.Pt(220, 160), Viewport{Zoom: 1}, 1); got != "" {
		t.Fatalf("straight edge incorrectly hit at step corner = %q", got)
	}
}

func TestEdgeStyleOptionsAreControlled(t *testing.T) {
	edge := NewEdge("edge", NewEndpoint("source", "out"), NewEndpoint("target", "in")).
		Type(EdgeSmoothStep).
		Width(2.5).
		Dashed(true).
		Animated(true).
		Markers(MarkerArrow, MarkerArrow).
		Label("value")
	if edge.edgeTypeValue() != EdgeSmoothStep || edge.width != 2.5 || !edge.isDashed() || !edge.isAnimated() || edge.sourceMark != MarkerArrow || edge.targetMark != MarkerArrow || edge.label != "value" {
		t.Fatalf("edge options were not retained: %#v", edge)
	}
}

func TestNextFocusedNodeMovesInRequestedDirection(t *testing.T) {
	graph := resolveGraph(Graph{Nodes: []Node{
		NewNode("left", "Left", Point{X: 0, Y: 0}),
		NewNode("right", "Right", Point{X: 220, Y: 0}),
		NewNode("below", "Below", Point{X: 0, Y: 160}),
	}})
	if got := nextFocusedNode(graph, "left", 1, 0); got != "right" {
		t.Fatalf("right navigation = %q, want right", got)
	}
	if got := nextFocusedNode(graph, "left", 0, 1); got != "below" {
		t.Fatalf("down navigation = %q, want below", got)
	}
	if got := nextFocusedNode(graph, "left", -1, 0); got != "left" {
		t.Fatalf("edge navigation lost focus = %q", got)
	}
}

func TestGraphControlsAdjustViewport(t *testing.T) {
	widget := New("graph", Graph{Nodes: []Node{NewNode("node", "Node", Point{X: 100, Y: 80})}}).Controls(true)
	viewport := Viewport{Zoom: 1}
	next, changed := widget.applyControl(controlZoomIn, resolveGraph(widget.graph), viewport, image.Pt(400, 240), 1)
	if !changed || next.Zoom <= viewport.Zoom {
		t.Fatalf("zoom control result = %#v, changed=%v", next, changed)
	}
	fit, changed := widget.applyControl(controlFit, resolveGraph(widget.graph), viewport, image.Pt(400, 240), 1)
	if !changed || fit.Zoom <= 0 {
		t.Fatalf("fit control result = %#v, changed=%v", fit, changed)
	}
}

func TestEnsureFocusedNodeVisiblePansOnlyWhenNeeded(t *testing.T) {
	graph := resolveGraph(Graph{Nodes: []Node{
		NewNode("left", "Left", Point{X: 0, Y: 0}),
		NewNode("right", "Right", Point{X: 600, Y: 0}),
	}})
	viewport := Viewport{Origin: Point{X: -24, Y: -24}, Zoom: 1}
	if next, changed := ensureFocusedNodeVisible(graph, "left", viewport, image.Pt(400, 220), 1); changed || next != viewport {
		t.Fatalf("visible focus changed viewport: %#v", next)
	}
	next, changed := ensureFocusedNodeVisible(graph, "right", viewport, image.Pt(400, 220), 1)
	if !changed || next.Origin.X <= 0 {
		t.Fatalf("offscreen focus did not pan: %#v", next)
	}
}

func TestMinimapMapsScreenPointToGraphCoordinates(t *testing.T) {
	graph := resolveGraph(Graph{Nodes: []Node{
		NewNode("first", "First", Point{X: 0, Y: 0}),
		NewNode("last", "Last", Point{X: 600, Y: 400}),
	}})
	geometry, ok := resolveMinimapGeometry(graph, image.Pt(480, 320))
	if !ok {
		t.Fatal("minimap geometry was not resolved")
	}
	world, ok := minimapWorldAt(graph, f32.Pt(float32(geometry.inner.Max.X-1), float32(geometry.inner.Max.Y-1)), image.Pt(480, 320))
	if !ok || world.X < 500 || world.Y < 300 {
		t.Fatalf("minimap world point = %#v, ok=%v", world, ok)
	}
}

func TestSnappedDragDeltaPreservesGroupOffset(t *testing.T) {
	delta := snappedDragDelta(Point{X: 10, Y: 20}, Point{X: 13, Y: 15}, Point{X: 16, Y: 16})
	if delta != (Point{X: 6, Y: 12}) {
		t.Fatalf("snapped delta = %#v, want {6 12}", delta)
	}
}

func TestSingleSelectionCollapsesExternalMultiSelection(t *testing.T) {
	got := updateSelection(map[string]bool{"first": true, "second": true}, "second", SelectionSingle, false)
	if len(got) != 1 || !got["second"] {
		t.Fatalf("selection = %#v, want only second", got)
	}
}

func TestSelectionBoxModes(t *testing.T) {
	node := resolvedNode{node: NewNode("node", "Node", Point{X: 20, Y: 20}).WithSize(100, 80), position: Point{X: 20, Y: 20}, size: Size{Width: 100, Height: 80}}
	partial := graphBox(Point{X: 90, Y: 70}, Point{X: 140, Y: 120})
	if !selectionBoxMatches(partial, node, SelectionBoxPartial) {
		t.Fatal("partial selection did not match intersecting node")
	}
	if selectionBoxMatches(partial, node, SelectionBoxFull) {
		t.Fatal("full selection matched a partially covered node")
	}
	full := graphBox(Point{X: 10, Y: 10}, Point{X: 140, Y: 120})
	if !selectionBoxMatches(full, node, SelectionBoxFull) {
		t.Fatal("full selection did not match contained node")
	}
}

func TestParentNodeResolvesChildWorldPositionAndPorts(t *testing.T) {
	parent := NewNode("group", "Group", Point{X: 100, Y: 80}).WithSize(400, 280)
	child := NewNode("child", "Child", Point{X: 40, Y: 30}).Parent("group").
		Outputs(NewPort("out", "Out"))
	target := NewNode("target", "Target", Point{X: 620, Y: 120}).Inputs(NewPort("in", "In"))
	resolved := resolveGraph(Graph{
		Nodes: []Node{child, target, parent},
		Edges: []Edge{NewEdge("edge", NewEndpoint("child", "out"), NewEndpoint("target", "in"))},
	})
	childResolved := resolved.byID["child"]
	if childResolved.position != (Point{X: 140, Y: 110}) {
		t.Fatalf("child world position = %#v", childResolved.position)
	}
	if childResolved.node.Position != child.Position {
		t.Fatalf("child local position mutated = %#v", childResolved.node.Position)
	}
	if len(resolved.nodes) < 2 || resolved.nodes[0].node.ID != "group" || resolved.nodes[1].node.ID != "child" {
		t.Fatalf("parent was not resolved before child: %#v", resolved.nodes)
	}
	wantAnchor := Point{X: 140 + childResolved.size.Width, Y: 110 + nodeHeaderHeight + nodeVerticalPad + nodePortRowHeight/2}
	if got := resolved.edges[0].source; got != wantAnchor {
		t.Fatalf("child edge anchor = %#v, want %#v", got, wantAnchor)
	}
}

func TestParentNodeRejectsUnknownParentAndCycles(t *testing.T) {
	t.Run("unknown parent", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Fatal("unknown parent did not panic")
			}
		}()
		resolveGraph(Graph{Nodes: []Node{NewNode("child", "Child", Point{}).Parent("missing")}})
	})
	t.Run("cycle", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Fatal("parent cycle did not panic")
			}
		}()
		resolveGraph(Graph{Nodes: []Node{
			NewNode("first", "First", Point{}).Parent("second"),
			NewNode("second", "Second", Point{}).Parent("first"),
		}})
	})
}

func TestParentDragDoesNotDuplicateSelectedChildMovement(t *testing.T) {
	graph := resolveGraph(Graph{Nodes: []Node{
		NewNode("parent", "Parent", Point{X: 100, Y: 100}).WithSize(400, 300),
		NewNode("child", "Child", Point{X: 20, Y: 20}).Parent("parent"),
	}})
	selected := map[string]bool{"parent": true, "child": true}
	if graph.hasSelectedAncestor(graph.byID["parent"].node, selected) {
		t.Fatal("root node reported a selected ancestor")
	}
	if !graph.hasSelectedAncestor(graph.byID["child"].node, selected) {
		t.Fatal("child did not report selected parent")
	}
	parentDrag := graph.dragNode(graph.byID["parent"])
	if parentDrag.position != (Point{X: 100, Y: 100}) {
		t.Fatalf("parent drag stores wrong model position: %#v", parentDrag)
	}
}

func TestChildDragCanBeConstrainedToParentBounds(t *testing.T) {
	graph := resolveGraph(Graph{Nodes: []Node{
		NewNode("parent", "Parent", Point{}).WithSize(300, 200),
		NewNode("child", "Child", Point{X: 20, Y: 20}).WithSize(120, 80).Parent("parent").ConstrainToParent(true),
	}})
	drag := graph.dragNode(graph.byID["child"])
	if drag.minimum != (Point{}) || drag.maximum != (Point{X: 180, Y: 120}) {
		t.Fatalf("child drag bounds = %#v", drag)
	}
}

func TestCopySelectionIncludesDescendantsAndInternalEdges(t *testing.T) {
	graph := Graph{
		Nodes: []Node{
			NewNode("group", "Group", Point{X: 100, Y: 80}).WithSize(400, 300),
			NewNode("child", "Child", Point{X: 30, Y: 40}).Parent("group").Inputs(NewPort("in", "In")).Outputs(NewPort("out", "Out")),
			NewNode("outside", "Outside", Point{X: 600, Y: 80}).Inputs(NewPort("in", "In")),
		},
		Edges: []Edge{
			NewEdge("internal", NewEndpoint("child", "out"), NewEndpoint("child", "in")),
			NewEdge("external", NewEndpoint("child", "out"), NewEndpoint("outside", "in")),
		},
	}
	fragment := CopySelection(graph, map[string]bool{"group": true})
	if len(fragment.Nodes) != 2 || fragment.Nodes[0].ID != "group" || fragment.Nodes[1].ID != "child" {
		t.Fatalf("copied nodes = %#v", fragment.Nodes)
	}
	if len(fragment.Edges) != 1 || fragment.Edges[0].ID != "internal" {
		t.Fatalf("copied edges = %#v", fragment.Edges)
	}
	childOnly := CopySelection(graph, map[string]bool{"child": true})
	if len(childOnly.Nodes) != 1 || childOnly.Nodes[0].parentID != "" || childOnly.Nodes[0].Position != (Point{X: 130, Y: 120}) {
		t.Fatalf("detached child copy = %#v", childOnly.Nodes)
	}
}

func TestPasteFragmentRemapsIDsAndOffsetsOnlyRoots(t *testing.T) {
	fragment := Fragment{
		Nodes: []Node{
			NewNode("group", "Group", Point{X: 20, Y: 30}),
			NewNode("child", "Child", Point{X: 10, Y: 12}).Parent("group").Outputs(NewPort("out", "Out")),
			NewNode("target", "Target", Point{X: 200, Y: 30}).Inputs(NewPort("in", "In")),
		},
		Edges: []Edge{NewEdge("edge", NewEndpoint("child", "out"), NewEndpoint("target", "in"))},
	}
	pasted := PasteFragment(Graph{}, fragment,
		func(node Node) string { return "copy-" + node.ID },
		func(edge Edge) string { return "copy-" + edge.ID },
		Point{X: 40, Y: 50},
	)
	if got := pasted.Nodes[0]; got.ID != "copy-group" || got.Position != (Point{X: 60, Y: 80}) {
		t.Fatalf("pasted root = %#v", got)
	}
	if got := pasted.Nodes[1]; got.ID != "copy-child" || got.parentID != "copy-group" || got.Position != (Point{X: 10, Y: 12}) {
		t.Fatalf("pasted child = %#v", got)
	}
	if got := pasted.Edges[0]; got.ID != "copy-edge" || got.Source.NodeID != "copy-child" || got.Target.NodeID != "copy-target" {
		t.Fatalf("pasted edge = %#v", got)
	}
}

func TestHistoryUndoesRedoesAndDropsRedoAfterCommit(t *testing.T) {
	initial := Graph{Nodes: []Node{NewNode("one", "One", Point{})}}
	history := NewHistory(initial).Limit(3)
	second := Graph{Nodes: []Node{NewNode("two", "Two", Point{})}}
	third := Graph{Nodes: []Node{NewNode("three", "Three", Point{})}}
	history.Commit(second)
	history.Commit(third)
	if got, ok := history.Undo(); !ok || got.Nodes[0].ID != "two" {
		t.Fatalf("undo = %#v, ok=%v", got, ok)
	}
	if got, ok := history.Redo(); !ok || got.Nodes[0].ID != "three" {
		t.Fatalf("redo = %#v, ok=%v", got, ok)
	}
	history.Undo()
	history.Commit(Graph{Nodes: []Node{NewNode("replacement", "Replacement", Point{})}})
	if history.CanRedo() {
		t.Fatal("commit retained redo history")
	}
}

func TestConnectionConstraintsEnforcePortTypeAndCapacity(t *testing.T) {
	source := NewNode("source", "Source", Point{}).Outputs(NewPort("out", "Out").Type("json").MaxConnections(1))
	target := NewNode("target", "Target", Point{X: 300, Y: 0}).Inputs(NewPort("in", "In").Type("json").MaxConnections(1))
	wrongType := NewNode("wrong", "Wrong", Point{X: 300, Y: 160}).Inputs(NewPort("in", "In").Type("text"))
	connection := Connection{Source: NewEndpoint("source", "out"), Target: NewEndpoint("target", "in")}
	graph := resolveGraph(Graph{Nodes: []Node{source, target, wrongType}})
	if !graph.connectionAllowed(connection, "") {
		t.Fatal("compatible connection was rejected")
	}
	if graph.connectionAllowed(Connection{Source: NewEndpoint("source", "out"), Target: NewEndpoint("wrong", "in")}, "") {
		t.Fatal("incompatible port types were accepted")
	}
	withEdge := resolveGraph(Graph{Nodes: []Node{source, target}, Edges: []Edge{NewEdge("edge", connection.Source, connection.Target)}})
	if withEdge.connectionAllowed(connection, "") {
		t.Fatal("connection capacity was not enforced")
	}
	if !withEdge.connectionAllowed(connection, "edge") {
		t.Fatal("reconnection did not exclude its existing edge")
	}
}

func TestGraphQueriesResolveGroupsBoundsConnectionsAndCoordinates(t *testing.T) {
	graph := Graph{
		Nodes: []Node{
			NewNode("group", "Group", Point{X: 100, Y: 80}).WithSize(320, 220),
			NewNode("child", "Child", Point{X: 20, Y: 30}).Parent("group").Outputs(NewPort("out", "Out")),
			NewNode("target", "Target", Point{X: 520, Y: 100}).Inputs(NewPort("in", "In")),
		},
		Edges: []Edge{NewEdge("edge", NewEndpoint("child", "out"), NewEndpoint("target", "in"))},
	}
	if position, ok := NodeWorldPosition(graph, "child"); !ok || position != (Point{X: 120, Y: 110}) {
		t.Fatalf("child world position = %#v, ok=%v", position, ok)
	}
	bounds, ok := NodesBounds(graph, "child")
	if !ok || bounds.Min != (Point{X: 120, Y: 110}) || bounds.Width() <= 0 || bounds.Height() <= 0 {
		t.Fatalf("child bounds = %#v, ok=%v", bounds, ok)
	}
	if nodes := IntersectingNodes(graph, bounds, false); len(nodes) != 1 || nodes[0].ID != "child" || nodes[0].Position != (Point{X: 20, Y: 30}) {
		t.Fatalf("contained nodes = %#v", nodes)
	}
	if connections := NodeConnections(graph, "child"); len(connections) != 1 || connections[0].ID != "edge" {
		t.Fatalf("node connections = %#v", connections)
	}
	if connections := PortConnections(graph, NewEndpoint("child", "out"), true); len(connections) != 1 || connections[0].ID != "edge" {
		t.Fatalf("port connections = %#v", connections)
	}
	viewport := Viewport{Origin: Point{X: 40, Y: 30}, Zoom: 1.5}
	world := Point{X: 220, Y: 130}
	if got := ScreenToWorld(viewport, WorldToScreen(viewport, world, 2), 2); !pointsClose(got, world) {
		t.Fatalf("coordinate round trip = %#v, want %#v", got, world)
	}
	if fitted, ok := FitViewport(graph, []string{"child"}, image.Pt(600, 300), 1, .1, .25, 2); !ok || fitted.Zoom <= 0 {
		t.Fatalf("fitted child viewport = %#v, ok=%v", fitted, ok)
	}
}

func TestGraphPanelPosition(t *testing.T) {
	canvas := image.Pt(400, 300)
	size := image.Pt(80, 40)
	if got := graphPanelPosition(PanelTopCenter, size, canvas, 12); got != image.Pt(160, 12) {
		t.Fatalf("top-center panel position = %v", got)
	}
	if got := graphPanelPosition(PanelBottomRight, size, canvas, 12); got != image.Pt(308, 248) {
		t.Fatalf("bottom-right panel position = %v", got)
	}
}

func TestViewportInputOptionsPreserveDefaultsAndAllowOverrides(t *testing.T) {
	widget := New("graph", Graph{})
	if !widget.zoomOnScroll || !widget.panOnPrimary || !widget.panOnMiddle || widget.panOnScroll {
		t.Fatalf("viewport defaults = %#v", widget)
	}
	widget = widget.ZoomOnScroll(false).PanOnScroll(true).PanOnPrimaryButton(false).PanOnMiddleButton(false)
	if widget.zoomOnScroll || !widget.panOnScroll || widget.panOnPrimary || widget.panOnMiddle {
		t.Fatalf("viewport overrides = %#v", widget)
	}
}

func TestNodeGraphCursorMatchesInteractionState(t *testing.T) {
	widget := New("graph", Graph{})
	graph := resolveGraph(Graph{Nodes: []Node{NewNode("node", "Node", Point{})}})
	tests := []struct {
		name  string
		state graphState
		want  pointer.Cursor
	}{
		{name: "resize corner", state: graphState{hoveredResize: ResizeHandleTopLeft}, want: pointer.CursorNorthWestSouthEastResize},
		{name: "resize edge", state: graphState{hoveredResize: ResizeHandleLeft}, want: pointer.CursorEastWestResize},
		{name: "resizing", state: graphState{gesture: gestureNodeResize, resizeNode: resizeNode{handle: ResizeHandleBottom}}, want: pointer.CursorNorthSouthResize},
		{name: "connection", state: graphState{gesture: gestureConnect}, want: pointer.CursorCrosshair},
		{name: "selection", state: graphState{gesture: gestureSelectionBox}, want: pointer.CursorCrosshair},
		{name: "dragging", state: graphState{gesture: gestureNodeDrag}, want: pointer.CursorGrabbing},
		{name: "node", state: graphState{hoveredNode: "node"}, want: pointer.CursorGrab},
		{name: "edge", state: graphState{hoveredEdge: "edge"}, want: pointer.CursorPointer},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := nodeGraphCursor(widget, &test.state, graph); got != test.want {
				t.Fatalf("cursor = %v, want %v", got, test.want)
			}
		})
	}
}

func TestNodeGraphPointerCallbacksDispatchTargets(t *testing.T) {
	graph := resolveGraph(Graph{
		Nodes: []Node{
			NewNode("source", "Source", Point{X: 40, Y: 40}).Outputs(NewPort("out", "Out")),
			NewNode("target", "Target", Point{X: 300, Y: 40}).Inputs(NewPort("in", "In")),
		},
		Edges: []Edge{NewEdge("edge", NewEndpoint("source", "out"), NewEndpoint("target", "in"))},
	})
	var canvasClicks, nodeClicks, nodeDoubles, edgeClicks, nodeEnters, nodeLeaves, nodeMenus int
	widget := New("graph", Graph{}).
		OnCanvasClick(func(CanvasEvent) { canvasClicks++ }).
		OnNodeClick(func(NodeEvent) { nodeClicks++ }).
		OnNodeDoubleClick(func(NodeEvent) { nodeDoubles++ }).
		OnNodeContextMenu(func(NodeEvent) { nodeMenus++ }).
		OnNodeHover(func(NodeEvent) { nodeEnters++ }).
		OnNodeLeave(func(NodeEvent) { nodeLeaves++ }).
		OnEdgeClick(func(EdgeEvent) { edgeClicks++ })
	viewport := Viewport{Zoom: 1}
	nodePosition := f32.Pt(80, 80)
	state := graphState{press: graphPress{target: graphTargetNode, id: "source", position: nodePosition}}
	widget.finishPress(&state, graph, viewport, 1, pointer.Event{Position: nodePosition})
	state.press = graphPress{target: graphTargetNode, id: "source", position: nodePosition}
	widget.finishPress(&state, graph, viewport, 1, pointer.Event{Position: nodePosition})
	if nodeClicks != 2 || nodeDoubles != 1 {
		t.Fatalf("node click callbacks = %d/%d, want 2/1", nodeClicks, nodeDoubles)
	}
	widget.emitContextMenu(graph, viewport, 1, pointer.Event{Buttons: pointer.ButtonSecondary, Position: nodePosition})
	if nodeMenus != 1 {
		t.Fatalf("node context menus = %d, want 1", nodeMenus)
	}
	widget.updateHover(&state, graph, nil, viewport, 1, nodePosition, 0)
	widget.updateHover(&state, graph, nil, viewport, 1, f32.Pt(520, 180), 0)
	if nodeEnters != 1 || nodeLeaves != 1 {
		t.Fatalf("node hover callbacks = %d/%d, want 1/1", nodeEnters, nodeLeaves)
	}

	edge := graph.edges[0]
	edgePosition := graphEdgeMidpoint(edge.edge.edgeTypeValue(), f32.Pt(edge.source.X, edge.source.Y), f32.Pt(edge.target.X, edge.target.Y), 32)
	target, id := widget.targetAt(graph, edgePosition, viewport, 1)
	if target != graphTargetEdge || id != "edge" {
		t.Fatalf("edge target = %v/%q, want edge", target, id)
	}
	state.press = graphPress{target: graphTargetEdge, id: id, position: edgePosition}
	widget.finishPress(&state, graph, viewport, 1, pointer.Event{Position: edgePosition})
	if edgeClicks != 1 {
		t.Fatalf("edge clicks = %d, want 1", edgeClicks)
	}

	canvasPosition := f32.Pt(520, 180)
	state.press = graphPress{target: graphTargetCanvas, position: canvasPosition}
	widget.finishPress(&state, graph, viewport, 1, pointer.Event{Position: canvasPosition})
	if canvasClicks != 1 {
		t.Fatalf("canvas clicks = %d, want 1", canvasClicks)
	}
}

func TestNodeGraphDropTypesAreValidatedAndUnique(t *testing.T) {
	widget := New("graph", Graph{}).DropTypes(" application/x-flowui-node ", "text/plain", "text/plain")
	if len(widget.dropTypes) != 2 || widget.dropTypes[0] != "application/x-flowui-node" || widget.dropTypes[1] != "text/plain" {
		t.Fatalf("drop types = %#v", widget.dropTypes)
	}
}

func TestResizeHandlesAdjustAnchoredNodeBounds(t *testing.T) {
	resize := resizeNode{
		position: Point{X: 100, Y: 80},
		size:     Size{Width: 200, Height: 120},
		minimum:  Size{Width: 120, Height: 72},
		maximum:  Size{Width: 320, Height: 240},
		handle:   ResizeHandleTopLeft,
	}
	position, size := resizedNodeBounds(resize, Point{X: 40, Y: 20})
	if position != (Point{X: 140, Y: 100}) || size != (Size{Width: 160, Height: 100}) {
		t.Fatalf("top-left resize = %#v %#v", position, size)
	}
	resize.handle = ResizeHandleLeft
	position, size = resizedNodeBounds(resize, Point{X: 160, Y: 0})
	if position.X != 180 || size.Width != 120 {
		t.Fatalf("minimum left resize = %#v %#v", position, size)
	}
	resize.handle = ResizeHandleBottom
	position, size = resizedNodeBounds(resize, Point{Y: 300})
	if position != resize.position || size.Height != 240 {
		t.Fatalf("maximum bottom resize = %#v %#v", position, size)
	}
}

func TestConnectionValidation(t *testing.T) {
	widget := New("graph", Graph{}).IsValidConnection(func(connection Connection) bool {
		return connection.Source.NodeID != connection.Target.NodeID
	})
	if widget.validConnection(Connection{Source: Endpoint{NodeID: "node"}, Target: Endpoint{NodeID: "node"}}) {
		t.Fatal("invalid connection was accepted")
	}
	if !widget.validConnection(Connection{Source: Endpoint{NodeID: "source"}, Target: Endpoint{NodeID: "target"}}) {
		t.Fatal("valid connection was rejected")
	}
}

func pointsClose(first, second Point) bool {
	return math.Abs(float64(first.X-second.X)) < 1e-4 && math.Abs(float64(first.Y-second.Y)) < 1e-4
}
