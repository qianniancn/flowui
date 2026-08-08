# 17 - Node graph

`NodeGraph` is the canvas foundation for node-based workflows, data flows, and
visual editors. The application owns the graph and viewport. The component
renders nodes, ports, and edges, then reports viewport changes to the model.

## Graph data

```go
graph := ui.NodeGraphData{
	Nodes: []ui.NodeGraphNode{
		ui.NewNodeGraphNode("request", "HTTP Request", ui.NodeGraphPoint{X: 80, Y: 140}).
			Outputs(ui.NewNodeGraphPort("body", "Body")),
		ui.NewNodeGraphNode("parse", "Parse JSON", ui.NodeGraphPoint{X: 380, Y: 92}).
			Inputs(ui.NewNodeGraphPort("input", "Input")),
	},
	Edges: []ui.NodeGraphEdge{
		ui.NewNodeGraphEdge("request-body",
			ui.NewNodeGraphEndpoint("request", "body"),
			ui.NewNodeGraphEndpoint("parse", "input")),
	},
}
```

Node and port IDs are stable, unique identities. An edge starts at an output
port and ends at an input port. Positions and sizes use density-independent
graph coordinates: one graph unit equals one dp at zoom `1`.

Applications add nodes by updating their own model, then passing the new slice
back to `NodeGraph`. NodeGraph does not mutate graph data internally:

```go
nextID := len(model.Nodes) + 1
model.Nodes = append(model.Nodes,
	ui.NewNodeGraphNode("node-"+strconv.Itoa(nextID), title,
		ui.NodeGraphPoint{X: 80, Y: 96}).
		Inputs(ui.NewNodeGraphPort("in", "In")).
		Outputs(ui.NewNodeGraphPort("out", "Out")),
)
```

NodeGraph uses FlowUI's tool-surface hierarchy: the canvas is a
`SurfaceSecondary` area with a thin border and 8dp radius, while nodes use the
`FieldBackground` (pure white in the light theme). A visible, low-emphasis
line grid is shown by default and may be disabled with `Grid(false)`. Hovered
nodes gain a subtle shadow;

`Grid(true)` controls visibility, and `GridPattern` switches between lines and
dots:

```go
ui.NodeGraph("workflow", graph).
	Grid(true).
	GridPattern(ui.NodeGraphGridDots)
```

`GridColor` and `GridOpacity` override those values for one graph without
changing the application theme:

```go
ui.NodeGraph("workflow", graph).
	GridPattern(ui.NodeGraphGridDots).
	GridColor(color.NRGBA{R: 0x32, G: 0x78, B: 0xc8, A: 0xff}).
	GridOpacity(.35)
```
draggable nodes use a grab cursor and connectable ports use a crosshair.

The canvas, grid, nodes, ports, edges, and selection state use the dedicated
tokens in `Theme.Components.NodeGraph`. Applications can override individual
values:

```go
ui.CustomizeTheme(func(theme *ui.Theme) {
	theme.Components.NodeGraph.CanvasBackground = color.NRGBA{R: 0xf8, G: 0xfa, B: 0xfc, A: 0xff}
	theme.Components.NodeGraph.CanvasBorder = color.NRGBA{R: 0xd9, G: 0xe1, B: 0xeb, A: 0xff}
	theme.Components.NodeGraph.GridColor = color.NRGBA{R: 0xd9, G: 0xe1, B: 0xeb, A: 0xff}
	theme.Components.NodeGraph.GridOpacity = .65
})
```

`Culling(true)` is enabled by default. It skips drawing nodes and edges outside
the canvas viewport for large graphs. Hit testing, the minimap, and FitView
still use the complete graph data; disable it only when diagnosing rendering.

## Canvas Events and Drops

Canvas callbacks return graph coordinates, so applications can create a node
exactly where a user interacts. Node and edge callbacks carry the target ID;
hover callbacks fire when the pointer enters or leaves a target:

```go
ui.NodeGraph("workflow", graph).
	OnCanvasDoubleClick(func(event ui.NodeGraphCanvasEvent) {
		addNode(event.Position)
	}).
	OnNodeContextMenu(func(event ui.NodeGraphNodeEvent) {
		openNodeMenu(event.NodeID, event.Position)
	}).
	OnEdgeHover(func(event ui.NodeGraphEdgeEvent) {
		setInspectedEdge(event.EdgeID)
	})
```

`OnCanvasClick`, `OnNodeClick`, and `OnEdgeClick` report primary clicks;
each has a `DoubleClick` variant. `OnCanvasContextMenu`,
`OnNodeContextMenu`, and `OnEdgeContextMenu` report secondary-button presses.

For data from an external Gio drag source, declare each accepted MIME type and
handle the copied payload in `OnDrop`:

```go
ui.NodeGraph("workflow", graph).
	DropTypes("application/x-flowui-node").
	OnDrop(func(event ui.NodeGraphDropEvent) {
		addNodeFromPayload(event.Position, event.Data)
	})
```

The drag source remains application-owned. It must offer one of the declared
MIME types through Gio's transfer API; NodeGraph validates the type, limits a
payload to 1 MiB, and reports the graph-coordinate drop position.

## Controlled viewport

Keep the viewport in the model and apply change requests through messages:

```go
ui.NodeGraph("workflow", graph).
	Height(520).
	Viewport(model.Viewport).
	OnViewportChange(func(next ui.NodeGraphViewport) {
		send(ViewportChanged{Viewport: next})
	})
```

Drag empty space with the primary button, or drag anywhere with the middle
button, to pan. The mouse wheel zooms around its pointer position. `ZoomRange`
sets zoom bounds; `Grid` and `GridSize` control the background. Without a
controlled `Viewport`, use `DefaultViewport` for initial local state.

`ZoomOnScroll`, `PanOnScroll`, `PanOnPrimaryButton`, and
`PanOnMiddleButton` control those input modes independently. The defaults
preserve wheel zooming and primary/middle panning. After disabling
`ZoomOnScroll`, enable `PanOnScroll(true)` to pan with the wheel.

`FitView(true)` calculates a viewport containing every node once nodes and a
canvas size are available, taking precedence over `DefaultViewport`. A
controlled viewport receives that calculated value through
`OnViewportChange`, and becomes stable when the application writes it back.
`FitViewPadding` reserves a fraction at every canvas edge; its default is
`0.1`. To request another fit, toggle `FitView` from `false` to `true`:

```go
ui.NodeGraph("workflow", graph).
	Viewport(model.Viewport).
	FitView(model.RequestFit).
	FitViewPadding(.1).
	OnViewportChange(func(next ui.NodeGraphViewport) {
		send(ViewportChanged{Viewport: next})
	})
```

The Minimap and zoom controls are optional. Both use the controlled viewport;
zoom, zoom-out, and Fit actions report through `OnViewportChange`. Clicking the
Minimap centers the viewport there, and dragging its viewport frame pans
continuously:

```go
ui.NodeGraph("workflow", graph).
	Viewport(model.Viewport).
	Minimap(true).
	Controls(true).
	OnViewportChange(func(next ui.NodeGraphViewport) {
		send(ViewportChanged{Viewport: next})
	})
```

## Canvas Overlays and Queries

`Panel` is a fixed FlowUI Widget above the canvas, equivalent to XYFlow's
Panel, and does not move or scale with the graph. `ViewportOverlay` is placed
in graph coordinates, equivalent to ViewportPortal, and follows pan and zoom.
Both scope state by the NodeGraph key and their own key:

```go
ui.NodeGraph("workflow", graph).
	Panel("actions", ui.NodeGraphPanelTopRight,
		ui.Button("fit", "Fit")).
	ViewportOverlay("annotation", ui.NodeGraphPoint{X: 240, Y: 128},
		ui.Text("Inspect this node"))
```

`NodeGraphNodesBounds`, `NodeGraphIntersectingNodes`,
`NodeGraphNodeConnections`, `NodeGraphPortConnections`,
`NodeGraphNodeWorldPosition`, and the coordinate conversion helpers provide
queries that do not depend on widget state. They return application-data
copies for inspectors, auto-layout, and business commands.
`NodeGraphFitViewport` calculates a controlled viewport for all or specified
nodes.

## Selection and dragging

Selection and position requests are returned through `OnNodesChange`.
`ApplyNodeGraphChanges` is a pure helper for the common Model update:

```go
ui.NodeGraph("workflow", graph).
	SelectionMode(ui.NodeGraphSelectionMultiple).
	SnapToGrid(true).
	SnapGrid(16, 16).
	OnNodesChange(func(changes []ui.NodeGraphNodeChange) {
		send(NodesChanged{Nodes: ui.ApplyNodeGraphChanges(model.Nodes, changes)})
	})
```

Primary-click selects one node; Shift, Ctrl, or Command-click toggles a node
in a multi-selection. Dragging a selected node moves the draggable selection as
a group. `NodeDragThreshold` controls the movement required to start a drag.
`NodesDraggable` and `NodesSelectable` provide the global defaults, while
`Draggable(false)` and `Selectable(false)` override either behavior per node.
`SelectedKeys` supplies a separate controlled selection; otherwise Node's
`Selected` field is the selection source.

`NodesResizable(true)` shows a lower-right resize handle for selected nodes.
An individual node can override it with `Resizable(true)` or
`Resizable(false)`. Resizing reports `NodeChangeSize`, while `SizeRange`
constrains the interactive dimensions:

```go
node := ui.NewNodeGraphNode("transform", "Transform", ui.NodeGraphPoint{X: 80, Y: 80}).
	WithSize(180, 100).
	Resizable(true).
	SizeRange(120, 72, 480, 360)
```

The default remains the lower-right handle for compatibility. Use
`ResizeHandles(ui.NodeGraphResizeHandleAll)` for four edges and four corners;
an individual node can override that set with `ResizeHandles`. Resizing from
the left or top reports both a position and a size request, which
`ApplyNodeGraphChanges` applies directly:

```go
ui.NodeGraph("workflow", graph).
	NodesResizable(true).
	ResizeHandles(ui.NodeGraphResizeHandleAll)
```

## Groups and Parent Nodes

`Parent` places a node inside another node. A child's `Position` always stays
in the direct parent's local coordinates, while NodeGraph resolves world
coordinates for layout, edge anchors, box selection, fit view, and the
minimap. Moving a parent therefore needs only the parent position update; the
child model data is not changed a second time:

```go
group := ui.NewNodeGraphNode("extract", "Extract", ui.NodeGraphPoint{X: 120, Y: 80}).
	WithSize(440, 300)
child := ui.NewNodeGraphNode("parse", "Parse", ui.NodeGraphPoint{X: 32, Y: 64}).
	Parent("extract").
	ConstrainToParent(true)
```

Parents may appear anywhere in `Nodes`; NodeGraph resolves and paints each
parent before its descendants. Parent references must exist and cannot form a
cycle. Dragging a selected parent with selected descendants reports only the
parent movement. `ConstrainToParent(true)` keeps interactive child dragging
inside the direct parent bounds. It is off by default so application rules can
allow deliberate overflow.

## Editing Commands and History

With canvas focus, `Ctrl/Cmd+A` selects all, `Ctrl/Cmd+C`, `Ctrl/Cmd+X`, and
`Ctrl/Cmd+V` request copy, cut, and paste, and `Ctrl/Cmd+Z`, `Ctrl/Cmd+Y`, and
`Ctrl/Cmd+Shift+Z` request undo and redo. NodeGraph never changes application
graph data itself: `OnCopy`, `OnCut`, `OnPaste`, `OnUndo`, and `OnRedo` are
controlled command callbacks.

`Fragment` preserves selected descendants and internal edges. Copying a child
without its parent turns it into a root at its current world position.
`PasteFragment` requires application-owned node and edge ID allocators, so the
component never guesses business identities. `History` is an optional snapshot
history; call `Commit` only for completed operations such as a drag or resize
release:

```go
history := ui.NewNodeGraphHistory(model.Graph).Limit(100)

ui.NodeGraph("workflow", model.Graph).
	OnPaste(func(fragment ui.NodeGraphFragment, offset ui.NodeGraphPoint) {
		next := ui.PasteNodeGraphFragment(model.Graph, fragment,
			func(node ui.NodeGraphNode) string { return newNodeID(node) },
			func(edge ui.NodeGraphEdge) string { return newEdgeID(edge) },
			offset,
		)
		send(GraphReplaced{Graph: next})
	}).
	OnUndo(func() {
		if previous, ok := history.Undo(); ok {
			send(GraphReplaced{Graph: previous})
		}
	})
```

After the canvas receives keyboard focus, arrow keys move focus to the nearest
node in that direction. Home and End jump to the first and last node, while
Enter or Space selects the focused node. When the focused node is outside the
visible area, the viewport pans by the minimum amount needed to reveal it.
Nodes expose Gio Button semantics with their label and selected state; controls
inside custom content retain their own focus and semantics.

## Edge selection and deletion

Edges use the same controlled selection model as nodes. The component keeps a
larger hit area around the visible stroke; primary-click selects an edge, and
Shift, Ctrl, or Command enables multi-selection. `SelectedEdgeKeys` controls
edge selection independently; otherwise `Edge.Selected` is the source.
`EdgesSelectable` and `Edge.Selectable(false)` configure the default.

`OnEdgesChange` reports selection and removal requests, and
`ApplyNodeGraphEdgeChanges` can write them back to the model:

```go
ui.NodeGraph("workflow", graph).
	OnEdgesChange(func(changes []ui.NodeGraphEdgeChange) {
		send(EdgesChanged{Edges: ui.ApplyNodeGraphEdgeChanges(model.Edges, changes)})
	})
```

Once the graph has focus, Backspace or Delete requests removal of selected
elements. `NodesDeletable` and `EdgesDeletable` configure the global defaults;
`Node.Deletable(false)` and `Edge.Deletable(false)` override either setting.
Removing a node also requests removal of every connected edge.

## Box selection and connections

With `SelectionOnDrag(true)`, primary-button dragging on empty space selects
nodes while middle-button dragging continues to pan. Use
`SelectionBoxMode(ui.NodeGraphSelectionBoxPartial)` for intersecting nodes, or
`NodeGraphSelectionBoxFull` to require full containment.

Dragging from an output port to an input port shows a connection preview.
`IsValidConnection` rejects candidates according to application rules; release
on a valid input delivers `OnConnect`:

```go
ui.NodeGraph("workflow", graph).
	IsValidConnection(func(connection ui.NodeGraphConnection) bool {
		return connection.Source.NodeID != connection.Target.NodeID
	}).
	OnConnect(func(connection ui.NodeGraphConnection) {
		// Create an Edge in Update, then render it through graph.Edges.
	})
```

`NodesConnectable` is the global default. `Node.Connectable(false)` and
`Port.Connectable(false)` disable connections locally.

Ports can also express common constraints without a business callback. When
both endpoints have a non-empty `Type`, they must match. `MaxConnections(1)`
limits a port to one connection. Existing edges count against the limit, while
the edge being reconnected is excluded:

```go
source := ui.NewNodeGraphPort("payload", "Payload").Type("json")
target := ui.NewNodeGraphPort("input", "Input").Type("json").MaxConnections(1)
```

## Edge reconnection

Drag an existing edge endpoint to request that endpoint's replacement.
Reconnection uses `IsValidConnection` just like a new connection, but the
component never changes application data itself. `OnReconnect` receives the
old edge and complete replacement connection. `ReconnectNodeGraphEdge` is a
pure helper that writes the change back while retaining the edge ID:

```go
ui.NodeGraph("workflow", graph).
	OnReconnect(func(oldEdge ui.NodeGraphEdge, connection ui.NodeGraphConnection) {
		send(EdgeReconnected{
			Edges: ui.ReconnectNodeGraphEdge(oldEdge, connection, model.Edges),
		})
	})
```

Both endpoints are reconnectable by default. `EdgesReconnectable(false)`
disables that default, while an individual edge can limit its range:

```go
sourceOnly := ui.NewNodeGraphEdge("source-only", source, target).
	Reconnectable(ui.NodeGraphReconnectSource)
targetOnly := ui.NewNodeGraphEdge("target-only", source, target).
	Reconnectable(ui.NodeGraphReconnectTarget)
```

`NodeGraphReconnectBoth` permits both endpoints and
`NodeGraphReconnectNone` disables reconnection for that edge. A source
endpoint may only land on an output port and a target endpoint may only land
on an input port, preserving the output-to-input direction in every callback.

## Edge styles and labels

Edges use Bezier curves by default, and can switch to straight, stepped, or
rounded stepped geometry. Drawing and hit testing share the same geometry:

```go
edge := ui.NewNodeGraphEdge("request-body", source, target).
	Type(ui.NodeGraphEdgeSmoothStep).
	Width(1.5).
	Dashed(true).
	Markers(ui.NodeGraphMarkerNone, ui.NodeGraphMarkerArrow).
	Label("JSON")
```

`Animated(true)` moves the dash pattern along the edge and enables dashed
rendering. `Width` is expressed in dp; zero uses the theme default. `Color`,
`Selected`, and `Selectable` remain independently controlled.

`LabelContent` replaces a static label with any FlowUI Widget. Its state is
scoped by graph and edge ID, making it appropriate for interactive buttons,
toggles, and other inline controls:

```go
edge := ui.NewNodeGraphEdge("request-body", source, target).
	LabelContent(ui.Button("edge-action", "Inspect"))
```

## Custom node content and handles

Nodes can provide an inner widget with `Content`. Content receives a stable
state scope derived from the graph and node IDs. When `WithSize` does not set a
dimension, the widget's measured size contributes to the automatic node size;
explicit dimensions still take precedence:

```go
content := ui.WidgetFunc(func(ctx *ui.Context, gtx layout.Context) layout.Dimensions {
	return ui.Text("Running").Layout(ctx, gtx)
})
node := ui.NewNodeGraphNode("worker", "Worker", ui.NodeGraphPoint{X: 120, Y: 80}).
	Content(content).
	Inputs(ui.NewNodeGraphPort("request", "Request").Position(ui.NodeGraphHandleTop)).
	Outputs(ui.NewNodeGraphPort("result", "Result").Position(ui.NodeGraphHandleBottom))
```

Inputs default to the left side and outputs to the right. `Port.Position` can
place either role on `NodeGraphHandleLeft`, `NodeGraphHandleRight`,
`NodeGraphHandleTop`, or `NodeGraphHandleBottom`. Edge anchors, hit testing, and
reconnection all use the same resolved positions. Custom content replaces the
built-in title/body rendering; handles remain rendered and managed by
NodeGraph.

Graph mutations remain controlled callbacks owned by the application model.

See `examples/node_graph` for a runnable program.
