package ui_test

import (
	"image"
	"math"
	"testing"

	"gioui.org/f32"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/uitest"
)

func TestNodeGraphViewportGesturesAreControlled(t *testing.T) {
	harness := uitest.New(image.Pt(420, 280))
	viewport := ui.NodeGraphViewport{Zoom: 1}
	graph := ui.NodeGraphData{Nodes: []ui.NodeGraphNode{
		ui.NewNodeGraphNode("node", "Node", ui.NodeGraphPoint{X: 160, Y: 100}),
	}}
	widget := func() ui.NodeGraphWidget {
		return ui.NodeGraph("graph", graph).
			Viewport(viewport).
			OnViewportChange(func(next ui.NodeGraphViewport) { viewport = next })
	}

	harness.Frame(widget())
	harness.Router().Queue(
		pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: 1, Buttons: pointer.ButtonPrimary, Position: f32.Pt(20, 20)},
		pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, PointerID: 1, Buttons: pointer.ButtonPrimary, Position: f32.Pt(80, 60)},
		pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, PointerID: 1, Position: f32.Pt(80, 60)},
	)
	harness.Frame(widget())
	if viewport.Origin != (ui.NodeGraphPoint{X: -60, Y: -40}) {
		t.Fatalf("viewport after pan = %#v", viewport)
	}

	pointerPosition := f32.Pt(200, 120)
	worldBefore := ui.NodeGraphPoint{X: viewport.Origin.X + pointerPosition.X/viewport.Zoom, Y: viewport.Origin.Y + pointerPosition.Y/viewport.Zoom}
	harness.Router().Queue(pointer.Event{Kind: pointer.Scroll, Source: pointer.Mouse, Position: pointerPosition, Scroll: f32.Pt(0, -40)})
	harness.Frame(widget())
	if viewport.Zoom <= 1 {
		t.Fatalf("viewport zoom = %v, want > 1", viewport.Zoom)
	}
	worldAfter := ui.NodeGraphPoint{X: viewport.Origin.X + pointerPosition.X/viewport.Zoom, Y: viewport.Origin.Y + pointerPosition.Y/viewport.Zoom}
	if math.Abs(float64(worldAfter.X-worldBefore.X)) > 1e-4 || math.Abs(float64(worldAfter.Y-worldBefore.Y)) > 1e-4 {
		t.Fatalf("world point under cursor moved: before=%#v after=%#v", worldBefore, worldAfter)
	}
}

func TestNodeGraphCustomContentReceivesPointerInput(t *testing.T) {
	harness := uitest.New(image.Pt(360, 220))
	clicked := false
	graph := ui.NodeGraphData{Nodes: []ui.NodeGraphNode{
		ui.NewNodeGraphNode("node", "Node", ui.NodeGraphPoint{X: 40, Y: 40}).
			WithSize(180, 80).
			Content(ui.Button("inner", ui.Text("Run")).OnClick(func() { clicked = true })),
	}}
	widget := func() ui.NodeGraphWidget { return ui.NodeGraph("custom-content", graph) }
	harness.Frame(widget())
	harness.Click(f32.Pt(100, 80))
	harness.Frame(widget())
	if !clicked {
		t.Fatal("custom node content did not receive the pointer click")
	}
}

func TestNodeGraphPointerCallbacksUseGraphTargets(t *testing.T) {
	harness := uitest.New(image.Pt(520, 280))
	graph := ui.NodeGraphData{Nodes: []ui.NodeGraphNode{
		ui.NewNodeGraphNode("node", "Node", ui.NodeGraphPoint{X: 80, Y: 60}).WithSize(160, 80),
	}}
	var canvasClicks, canvasMenus, nodeClicks int
	widget := func() ui.NodeGraphWidget {
		return ui.NodeGraph("pointer-callbacks", graph).
			OnCanvasClick(func(ui.NodeGraphCanvasEvent) { canvasClicks++ }).
			OnCanvasContextMenu(func(ui.NodeGraphCanvasEvent) { canvasMenus++ }).
			OnNodeClick(func(event ui.NodeGraphNodeEvent) {
				if event.NodeID == "node" {
					nodeClicks++
				}
			})
	}
	harness.Frame(widget())
	harness.Click(f32.Pt(24, 220))
	harness.Frame(widget())
	harness.Click(f32.Pt(120, 90))
	harness.Frame(widget())
	harness.Router().Queue(pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: 2, Buttons: pointer.ButtonSecondary, Position: f32.Pt(24, 220)})
	harness.Frame(widget())
	if canvasClicks != 1 || canvasMenus != 1 || nodeClicks != 1 {
		t.Fatalf("pointer callbacks = canvas:%d menu:%d node:%d", canvasClicks, canvasMenus, nodeClicks)
	}
}

func TestNodeGraphKeyboardNavigationSelectsFocusedNode(t *testing.T) {
	harness := uitest.New(image.Pt(520, 260))
	graph := ui.NodeGraphData{Nodes: []ui.NodeGraphNode{
		ui.NewNodeGraphNode("left", "Left", ui.NodeGraphPoint{X: 40, Y: 60}),
		ui.NewNodeGraphNode("right", "Right", ui.NodeGraphPoint{X: 260, Y: 60}),
	}}
	widget := func() ui.NodeGraphWidget {
		return ui.NodeGraph("keyboard", graph).
			OnNodesChange(func(changes []ui.NodeGraphNodeChange) {
				graph.Nodes = ui.ApplyNodeGraphChanges(graph.Nodes, changes)
			})
	}
	harness.Frame(widget())
	harness.Click(f32.Pt(80, 90))
	harness.Frame(widget())
	harness.Key(key.NameRightArrow, 0)
	harness.Frame(widget())
	harness.Key(key.NameEnter, 0)
	harness.Frame(widget())
	if !graph.Nodes[1].Selected || graph.Nodes[0].Selected {
		t.Fatalf("keyboard selection = %#v", graph.Nodes)
	}
}

func TestNodeGraphControlsChangeControlledViewport(t *testing.T) {
	harness := uitest.New(image.Pt(360, 220))
	viewport := ui.NodeGraphViewport{Zoom: 1}
	graph := ui.NodeGraphData{Nodes: []ui.NodeGraphNode{
		ui.NewNodeGraphNode("node", "Node", ui.NodeGraphPoint{X: 80, Y: 60}),
	}}
	widget := func() ui.NodeGraphWidget {
		return ui.NodeGraph("controls", graph).
			Controls(true).
			Minimap(true).
			Viewport(viewport).
			OnViewportChange(func(next ui.NodeGraphViewport) { viewport = next })
	}
	harness.Frame(widget())
	harness.Click(f32.Pt(66, 190))
	harness.Frame(widget())
	if viewport.Zoom <= 1 {
		t.Fatalf("viewport after zoom control = %#v", viewport)
	}
}

func TestNodeGraphMinimapClickChangesControlledViewport(t *testing.T) {
	harness := uitest.New(image.Pt(360, 220))
	viewport := ui.NodeGraphViewport{Zoom: 1}
	graph := ui.NodeGraphData{Nodes: []ui.NodeGraphNode{
		ui.NewNodeGraphNode("first", "First", ui.NodeGraphPoint{X: 0, Y: 0}),
		ui.NewNodeGraphNode("last", "Last", ui.NodeGraphPoint{X: 600, Y: 400}),
	}}
	widget := func() ui.NodeGraphWidget {
		return ui.NodeGraph("minimap", graph).
			Minimap(true).
			Viewport(viewport).
			OnViewportChange(func(next ui.NodeGraphViewport) { viewport = next })
	}
	harness.Frame(widget())
	harness.Click(f32.Pt(310, 180))
	harness.Frame(widget())
	if viewport.Origin == (ui.NodeGraphPoint{}) {
		t.Fatalf("viewport after minimap click = %#v", viewport)
	}
}

func TestNodeGraphMinimapDragPansControlledViewport(t *testing.T) {
	harness := uitest.New(image.Pt(360, 220))
	viewport := ui.NodeGraphViewport{Zoom: 1}
	graph := ui.NodeGraphData{Nodes: []ui.NodeGraphNode{
		ui.NewNodeGraphNode("first", "First", ui.NodeGraphPoint{X: 0, Y: 0}),
		ui.NewNodeGraphNode("last", "Last", ui.NodeGraphPoint{X: 600, Y: 400}),
	}}
	widget := func() ui.NodeGraphWidget {
		return ui.NodeGraph("minimap-drag", graph).
			Minimap(true).
			Viewport(viewport).
			OnViewportChange(func(next ui.NodeGraphViewport) { viewport = next })
	}
	harness.Frame(widget())
	harness.Router().Queue(
		pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: 9, Buttons: pointer.ButtonPrimary, Position: f32.Pt(220, 120)},
		pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, PointerID: 9, Buttons: pointer.ButtonPrimary, Position: f32.Pt(250, 120)},
		pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, PointerID: 9, Position: f32.Pt(250, 120)},
	)
	harness.Frame(widget())
	if viewport.Origin.X <= 0 {
		t.Fatalf("viewport after minimap drag = %#v", viewport)
	}
}

func TestNodeGraphResizeIsControlledAndConstrained(t *testing.T) {
	harness := uitest.New(image.Pt(420, 280))
	graph := ui.NodeGraphData{Nodes: []ui.NodeGraphNode{
		ui.NewNodeGraphNode("node", "Node", ui.NodeGraphPoint{X: 40, Y: 60}).
			WithSize(150, 80).
			Select(true).
			Resizable(true).
			SizeRange(100, 70, 180, 100),
	}}
	widget := func() ui.NodeGraphWidget {
		return ui.NodeGraph("resize", graph).
			OnNodesChange(func(changes []ui.NodeGraphNodeChange) {
				graph.Nodes = ui.ApplyNodeGraphChanges(graph.Nodes, changes)
			})
	}
	harness.Frame(widget())
	harness.Router().Queue(
		pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: 12, Buttons: pointer.ButtonPrimary, Position: f32.Pt(190, 140)},
		pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, PointerID: 12, Buttons: pointer.ButtonPrimary, Position: f32.Pt(330, 280)},
		pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, PointerID: 12, Position: f32.Pt(330, 280)},
	)
	harness.Frame(widget())
	if graph.Nodes[0].Size != (ui.NodeGraphSize{Width: 180, Height: 100}) {
		t.Fatalf("resized node = %#v", graph.Nodes[0])
	}
}

func TestNodeGraphFitViewRequestsControlledViewportOnce(t *testing.T) {
	harness := uitest.New(image.Pt(800, 400))
	viewport := ui.NodeGraphViewport{Zoom: 1}
	fit := true
	requests := 0
	graph := ui.NodeGraphData{Nodes: []ui.NodeGraphNode{
		ui.NewNodeGraphNode("first", "First", ui.NodeGraphPoint{X: 100, Y: 100}),
		ui.NewNodeGraphNode("second", "Second", ui.NodeGraphPoint{X: 500, Y: 200}),
	}}
	widget := func() ui.NodeGraphWidget {
		return ui.NodeGraph("fit-view", graph).
			Viewport(viewport).
			FitView(fit).
			OnViewportChange(func(next ui.NodeGraphViewport) {
				requests++
				viewport = next
			})
	}

	harness.Frame(widget())
	if requests != 1 || viewport.Zoom == 1 || viewport.Origin == (ui.NodeGraphPoint{}) {
		t.Fatalf("fit request = %#v, count = %d", viewport, requests)
	}
	harness.Frame(widget())
	if requests != 1 {
		t.Fatalf("fit view repeated without a new request: %d", requests)
	}
	fit = false
	harness.Frame(widget())
	fit = true
	harness.Frame(widget())
	if requests != 2 {
		t.Fatalf("fit view did not trigger again: %d", requests)
	}
}

func TestNodeGraphNodeChangesAreControlled(t *testing.T) {
	harness := uitest.New(image.Pt(520, 320))
	graph := ui.NodeGraphData{Nodes: []ui.NodeGraphNode{
		ui.NewNodeGraphNode("first", "First", ui.NodeGraphPoint{X: 80, Y: 60}),
		ui.NewNodeGraphNode("second", "Second", ui.NodeGraphPoint{X: 320, Y: 60}),
	}}
	widget := func() ui.NodeGraphWidget {
		return ui.NodeGraph("graph", graph).
			SelectionMode(ui.NodeGraphSelectionMultiple).
			NodeDragThreshold(0).
			OnNodesChange(func(changes []ui.NodeGraphNodeChange) {
				graph.Nodes = ui.ApplyNodeGraphChanges(graph.Nodes, changes)
			})
	}

	harness.Frame(widget())
	harness.Router().Queue(
		pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: 1, Buttons: pointer.ButtonPrimary, Position: f32.Pt(100, 80)},
		pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, PointerID: 1, Buttons: pointer.ButtonPrimary, Position: f32.Pt(150, 120)},
		pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, PointerID: 1, Position: f32.Pt(150, 120)},
	)
	harness.Frame(widget())
	if !graph.Nodes[0].Selected || graph.Nodes[0].Position != (ui.NodeGraphPoint{X: 130, Y: 100}) {
		t.Fatalf("first node after drag = %#v", graph.Nodes[0])
	}

	harness.Router().Queue(
		pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: 2, Buttons: pointer.ButtonPrimary, Position: f32.Pt(340, 80), Modifiers: key.ModShift},
		pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, PointerID: 2, Position: f32.Pt(340, 80), Modifiers: key.ModShift},
	)
	harness.Frame(widget())
	if !graph.Nodes[0].Selected || !graph.Nodes[1].Selected {
		t.Fatalf("shift selection = %#v", graph.Nodes)
	}

	harness.Router().Queue(
		pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: 3, Buttons: pointer.ButtonPrimary, Position: f32.Pt(340, 80)},
		pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, PointerID: 3, Buttons: pointer.ButtonPrimary, Position: f32.Pt(360, 96)},
		pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, PointerID: 3, Position: f32.Pt(360, 96)},
	)
	harness.Frame(widget())
	if graph.Nodes[0].Position != (ui.NodeGraphPoint{X: 150, Y: 116}) || graph.Nodes[1].Position != (ui.NodeGraphPoint{X: 340, Y: 76}) {
		t.Fatalf("group drag positions = %#v, %#v", graph.Nodes[0].Position, graph.Nodes[1].Position)
	}
}

func TestNodeGraphSelectionBoxIsControlled(t *testing.T) {
	harness := uitest.New(image.Pt(600, 280))
	graph := ui.NodeGraphData{Nodes: []ui.NodeGraphNode{
		ui.NewNodeGraphNode("first", "First", ui.NodeGraphPoint{X: 80, Y: 60}),
		ui.NewNodeGraphNode("second", "Second", ui.NodeGraphPoint{X: 340, Y: 60}),
	}}
	widget := func() ui.NodeGraphWidget {
		return ui.NodeGraph("selection-box", graph).
			SelectionMode(ui.NodeGraphSelectionMultiple).
			SelectionOnDrag(true).
			NodeDragThreshold(0).
			OnNodesChange(func(changes []ui.NodeGraphNodeChange) {
				graph.Nodes = ui.ApplyNodeGraphChanges(graph.Nodes, changes)
			})
	}

	harness.Frame(widget())
	harness.Router().Queue(
		pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: 1, Buttons: pointer.ButtonPrimary, Position: f32.Pt(40, 40)},
		pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, PointerID: 1, Buttons: pointer.ButtonPrimary, Position: f32.Pt(580, 180)},
		pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, PointerID: 1, Position: f32.Pt(580, 180)},
	)
	harness.Frame(widget())
	if !graph.Nodes[0].Selected || !graph.Nodes[1].Selected {
		t.Fatalf("box selection = %#v", graph.Nodes)
	}

	harness.Router().Queue(
		pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: 2, Buttons: pointer.ButtonPrimary, Position: f32.Pt(20, 20)},
		pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, PointerID: 2, Position: f32.Pt(20, 20)},
	)
	harness.Frame(widget())
	if graph.Nodes[0].Selected || graph.Nodes[1].Selected {
		t.Fatalf("background click did not clear selection: %#v", graph.Nodes)
	}
}

func TestNodeGraphConnectionIsControlled(t *testing.T) {
	harness := uitest.New(image.Pt(620, 280))
	graph := ui.NodeGraphData{Nodes: []ui.NodeGraphNode{
		ui.NewNodeGraphNode("source", "Source", ui.NodeGraphPoint{X: 40, Y: 80}).Outputs(ui.NewNodeGraphPort("out", "Out")),
		ui.NewNodeGraphNode("target", "Target", ui.NodeGraphPoint{X: 360, Y: 80}).Inputs(ui.NewNodeGraphPort("in", "In")),
	}}
	var connected ui.NodeGraphConnection
	widget := func() ui.NodeGraphWidget {
		return ui.NodeGraph("connection", graph).
			IsValidConnection(func(connection ui.NodeGraphConnection) bool {
				return connection.Source.NodeID != connection.Target.NodeID
			}).
			OnConnect(func(connection ui.NodeGraphConnection) { connected = connection })
	}

	harness.Frame(widget())
	harness.Router().Queue(
		pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: 1, Buttons: pointer.ButtonPrimary, Position: f32.Pt(190, 124)},
		pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, PointerID: 1, Buttons: pointer.ButtonPrimary, Position: f32.Pt(360, 124)},
		pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, PointerID: 1, Position: f32.Pt(360, 124)},
	)
	harness.Frame(widget())
	if connected.Source != (ui.NodeGraphEndpoint{NodeID: "source", PortID: "out"}) || connected.Target != (ui.NodeGraphEndpoint{NodeID: "target", PortID: "in"}) {
		t.Fatalf("connection = %#v", connected)
	}
}

func TestNodeGraphEdgeSelectionAndDeletionAreControlled(t *testing.T) {
	harness := uitest.New(image.Pt(620, 280))
	graph := ui.NodeGraphData{
		Nodes: []ui.NodeGraphNode{
			ui.NewNodeGraphNode("source", "Source", ui.NodeGraphPoint{X: 40, Y: 80}).Outputs(ui.NewNodeGraphPort("out", "Out")),
			ui.NewNodeGraphNode("target", "Target", ui.NodeGraphPoint{X: 360, Y: 80}).Inputs(ui.NewNodeGraphPort("in", "In")),
		},
		Edges: []ui.NodeGraphEdge{
			ui.NewNodeGraphEdge("edge", ui.NewNodeGraphEndpoint("source", "out"), ui.NewNodeGraphEndpoint("target", "in")),
		},
	}
	widget := func() ui.NodeGraphWidget {
		return ui.NodeGraph("edge-selection", graph).
			SelectionMode(ui.NodeGraphSelectionMultiple).
			OnNodesChange(func(changes []ui.NodeGraphNodeChange) {
				graph.Nodes = ui.ApplyNodeGraphChanges(graph.Nodes, changes)
			}).
			OnEdgesChange(func(changes []ui.NodeGraphEdgeChange) {
				graph.Edges = ui.ApplyNodeGraphEdgeChanges(graph.Edges, changes)
			})
	}

	harness.Frame(widget())
	harness.Click(f32.Pt(275, 124))
	harness.Frame(widget())
	if !graph.Edges[0].Selected {
		t.Fatalf("edge selection = %#v", graph.Edges)
	}

	harness.Key(key.NameDeleteForward, 0)
	harness.Frame(widget())
	if len(graph.Edges) != 0 {
		t.Fatalf("edges after deletion = %#v", graph.Edges)
	}
}

func TestNodeGraphDeletingNodeRemovesConnectedEdges(t *testing.T) {
	harness := uitest.New(image.Pt(620, 280))
	graph := ui.NodeGraphData{
		Nodes: []ui.NodeGraphNode{
			ui.NewNodeGraphNode("source", "Source", ui.NodeGraphPoint{X: 40, Y: 80}).Outputs(ui.NewNodeGraphPort("out", "Out")),
			ui.NewNodeGraphNode("target", "Target", ui.NodeGraphPoint{X: 360, Y: 80}).Inputs(ui.NewNodeGraphPort("in", "In")),
		},
		Edges: []ui.NodeGraphEdge{
			ui.NewNodeGraphEdge("edge", ui.NewNodeGraphEndpoint("source", "out"), ui.NewNodeGraphEndpoint("target", "in")),
		},
	}
	widget := func() ui.NodeGraphWidget {
		return ui.NodeGraph("node-deletion", graph).
			OnNodesChange(func(changes []ui.NodeGraphNodeChange) {
				graph.Nodes = ui.ApplyNodeGraphChanges(graph.Nodes, changes)
			}).
			OnEdgesChange(func(changes []ui.NodeGraphEdgeChange) {
				graph.Edges = ui.ApplyNodeGraphEdgeChanges(graph.Edges, changes)
			})
	}

	harness.Frame(widget())
	harness.Click(f32.Pt(100, 100))
	harness.Frame(widget())
	harness.Key(key.NameDeleteBackward, 0)
	harness.Frame(widget())
	if len(graph.Nodes) != 1 || graph.Nodes[0].ID != "target" || len(graph.Edges) != 0 {
		t.Fatalf("graph after node deletion = %#v", graph)
	}
}

func TestNodeGraphReconnectsBothEndpointsUnderControlledState(t *testing.T) {
	graph := reconnectGraph(ui.NodeGraphReconnectBoth)
	harness := uitest.New(image.Pt(620, 360))
	widget := func() ui.NodeGraphWidget {
		return ui.NodeGraph("reconnect", graph).
			OnReconnect(func(oldEdge ui.NodeGraphEdge, connection ui.NodeGraphConnection) {
				graph.Edges = ui.ReconnectNodeGraphEdge(oldEdge, connection, graph.Edges)
			})
	}

	harness.Frame(widget())
	harness.Router().Queue(
		pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: 1, Buttons: pointer.ButtonPrimary, Position: f32.Pt(190, 124)},
		pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, PointerID: 1, Buttons: pointer.ButtonPrimary, Position: f32.Pt(190, 234)},
		pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, PointerID: 1, Position: f32.Pt(190, 234)},
	)
	harness.Frame(widget())
	if edge := graph.Edges[0]; edge.ID != "edge" || edge.Source.NodeID != "source-b" || edge.Target.NodeID != "target-a" {
		t.Fatalf("source reconnection = %#v", edge)
	}

	harness.Router().Queue(
		pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: 2, Buttons: pointer.ButtonPrimary, Position: f32.Pt(360, 124)},
		pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, PointerID: 2, Buttons: pointer.ButtonPrimary, Position: f32.Pt(360, 234)},
		pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, PointerID: 2, Position: f32.Pt(360, 234)},
	)
	harness.Frame(widget())
	if edge := graph.Edges[0]; edge.ID != "edge" || edge.Source.NodeID != "source-b" || edge.Target.NodeID != "target-b" {
		t.Fatalf("target reconnection = %#v", edge)
	}
}

func TestNodeGraphReconnectModeLimitsEndpoints(t *testing.T) {
	graph := reconnectGraph(ui.NodeGraphReconnectSource)
	harness := uitest.New(image.Pt(620, 360))
	called := false
	widget := func() ui.NodeGraphWidget {
		return ui.NodeGraph("reconnect-source-only", graph).
			OnReconnect(func(ui.NodeGraphEdge, ui.NodeGraphConnection) { called = true })
	}

	harness.Frame(widget())
	harness.Router().Queue(
		pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: 1, Buttons: pointer.ButtonPrimary, Position: f32.Pt(360, 124)},
		pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, PointerID: 1, Buttons: pointer.ButtonPrimary, Position: f32.Pt(360, 234)},
		pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, PointerID: 1, Position: f32.Pt(360, 234)},
	)
	harness.Frame(widget())
	if called || graph.Edges[0].Target.NodeID != "target-a" {
		t.Fatalf("source-only reconnection changed target: %#v", graph.Edges[0])
	}
}

func reconnectGraph(mode ui.NodeGraphReconnectMode) ui.NodeGraphData {
	return ui.NodeGraphData{
		Nodes: []ui.NodeGraphNode{
			ui.NewNodeGraphNode("source-a", "Source A", ui.NodeGraphPoint{X: 40, Y: 80}).Outputs(ui.NewNodeGraphPort("out", "Out")),
			ui.NewNodeGraphNode("source-b", "Source B", ui.NodeGraphPoint{X: 40, Y: 190}).Outputs(ui.NewNodeGraphPort("out", "Out")),
			ui.NewNodeGraphNode("target-a", "Target A", ui.NodeGraphPoint{X: 360, Y: 80}).Inputs(ui.NewNodeGraphPort("in", "In")),
			ui.NewNodeGraphNode("target-b", "Target B", ui.NodeGraphPoint{X: 360, Y: 190}).Inputs(ui.NewNodeGraphPort("in", "In")),
		},
		Edges: []ui.NodeGraphEdge{
			ui.NewNodeGraphEdge("edge", ui.NewNodeGraphEndpoint("source-a", "out"), ui.NewNodeGraphEndpoint("target-a", "in")).Reconnectable(mode),
		},
	}
}
