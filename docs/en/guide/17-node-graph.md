# 17 - Node graph

`NodeGraph` is the canvas foundation for node-based workflows, data flows, and
visual editors. The application owns the graph and viewport. The component
renders nodes, ports, and edges, then reports viewport changes to the model.

## Graph data

```go
graph := nodegraph.Graph{
	Nodes: []nodegraph.Node{
		nodegraph.NewNode("request", "HTTP Request", nodegraph.Point{X: 80, Y: 140}).
			Outputs(nodegraph.NewPort("body", "Body")),
		nodegraph.NewNode("parse", "Parse JSON", nodegraph.Point{X: 380, Y: 92}).
			Inputs(nodegraph.NewPort("input", "Input")),
	},
	Edges: []nodegraph.Edge{
		nodegraph.NewEdge("request-body",
			nodegraph.NewEndpoint("request", "body"),
			nodegraph.NewEndpoint("parse", "input")),
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
	nodegraph.NewNode("node-"+strconv.Itoa(nextID), title,
		nodegraph.Point{X: 80, Y: 96}).
		Inputs(nodegraph.NewPort("in", "In")).
		Outputs(nodegraph.NewPort("out", "Out")),
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
nodegraph.New("workflow", graph).
	Grid(true).
	GridPattern(nodegraph.GridDots)
```

To override the grid color and opacity for one graph without changing the
application theme, pass a provider-owned override:

```go
gridColor := color.NRGBA{R: 0x32, G: 0x78, B: 0xc8, A: 0xff}
gridOpacity := float32(.35)
nodegraph.StyledWithOverride(
	nodegraph.New("workflow", graph).GridPattern(nodegraph.GridDots),
	theme.NodeGraphOverride{GridColor: &gridColor, GridOpacity: &gridOpacity},
)
```

Grid color and opacity are provider-owned overrides; visual configuration does
not belong on the behavior widget.
draggable nodes use a grab cursor and connectable ports use a crosshair.

The canvas, grid, nodes, ports, edges, and selection state use the dedicated
tokens in `Theme.Components.NodeGraph`. Applications can override individual
values:

```go
window.CustomizeTheme(func(t *theme.Theme) {
	t.Components.NodeGraph.CanvasBackground = color.NRGBA{R: 0xf8, G: 0xfa, B: 0xfc, A: 0xff}
	t.Components.NodeGraph.CanvasBorder = color.NRGBA{R: 0xd9, G: 0xe1, B: 0xeb, A: 0xff}
	t.Components.NodeGraph.GridColor = color.NRGBA{R: 0xd9, G: 0xe1, B: 0xeb, A: 0xff}
	t.Components.NodeGraph.GridOpacity = .65
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
nodegraph.New("workflow", graph).
	OnCanvasDoubleClick(func(event nodegraph.CanvasEvent) {
		addNode(event.Position)
	}).
	OnNodeContextMenu(func(event nodegraph.NodeEvent) {
		openNodeMenu(event.NodeID, event.Position)
	}).
	OnEdgeHover(func(event nodegraph.EdgeEvent) {
		setInspectedEdge(event.EdgeID)
	})
```

`OnCanvasClick`, `OnNodeClick`, and `OnEdgeClick` report primary clicks;
each has a `DoubleClick` variant. `OnCanvasContextMenu`,
`OnNodeContextMenu`, and `OnEdgeContextMenu` report secondary-button presses.

For data from an external Gio drag source, declare each accepted MIME type and
handle the copied payload in `OnDrop`:

```go
nodegraph.New("workflow", graph).
	DropTypes("application/x-flowui-node").
	OnDrop(func(event nodegraph.DropEvent) {
		addNodeFromPayload(event.Position, event.Data)
	})
```

The drag source remains application-owned. It must offer one of the declared
MIME types through Gio's transfer API; NodeGraph validates the type, limits a
payload to 1 MiB, and reports the graph-coordinate drop position.

## Controlled viewport

Keep the viewport in the model and apply change requests through messages:

```go
nodegraph.New("workflow", graph).
	Height(520).
	Viewport(model.Viewport).
	OnViewportChange(func(next nodegraph.Viewport) {
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
nodegraph.New("workflow", graph).
	Viewport(model.Viewport).
	FitView(model.RequestFit).
	FitViewPadding(.1).
	OnViewportChange(func(next nodegraph.Viewport) {
		send(ViewportChanged{Viewport: next})
	})
```

The Minimap and zoom controls are optional. Both use the controlled viewport;
zoom, zoom-out, and Fit actions report through `OnViewportChange`. Clicking the
Minimap centers the viewport there, and dragging its viewport frame pans
continuously:

```go
nodegraph.New("workflow", graph).
	Viewport(model.Viewport).
	Minimap(true).
	Controls(true).
	OnViewportChange(func(next nodegraph.Viewport) {
		send(ViewportChanged{Viewport: next})
	})
```

## Canvas Overlays and Queries

`Panel` is a fixed FlowUI Widget above the canvas, equivalent to XYFlow's
Panel, and does not move or scale with the graph. `ViewportOverlay` is placed
in graph coordinates, equivalent to ViewportPortal, and follows pan and zoom.
Both scope state by the NodeGraph key and their own key:

```go
nodegraph.New("workflow", graph).
	Panel("actions", nodegraph.PanelTopRight,
		button.New("Fit")).
	ViewportOverlay("annotation", nodegraph.Point{X: 240, Y: 128},
		text.New("Inspect this node"))
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
nodegraph.New("workflow", graph).
	SelectionMode(nodegraph.SelectionMultiple).
	SnapToGrid(true).
	SnapGrid(16, 16).
	OnNodesChange(func(changes []nodegraph.NodeChange) {
		send(NodesChanged{Nodes: nodegraph.ApplyNodeChanges(model.Nodes, changes)})
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
node := nodegraph.NewNode("transform", "Transform", nodegraph.Point{X: 80, Y: 80}).
	WithSize(180, 100).
	Resizable(true).
	SizeRange(120, 72, 480, 360)
```

The default resize handle is the lower-right corner. Use
`ResizeHandles(nodegraph.ResizeHandleAll)` for four edges and four corners;
an individual node can override that set with `ResizeHandles`. Resizing from
the left or top reports both a position and a size request, which
`ApplyNodeGraphChanges` applies directly:

```go
nodegraph.New("workflow", graph).
	NodesResizable(true).
	ResizeHandles(nodegraph.ResizeHandleAll)
```

## Groups and Parent Nodes

`Parent` places a node inside another node. A child's `Position` always stays
in the direct parent's local coordinates, while NodeGraph resolves world
coordinates for layout, edge anchors, box selection, fit view, and the
minimap. Moving a parent therefore needs only the parent position update; the
child model data is not changed a second time:

```go
group := nodegraph.NewNode("extract", "Extract", nodegraph.Point{X: 120, Y: 80}).
	WithSize(440, 300)
child := nodegraph.NewNode("parse", "Parse", nodegraph.Point{X: 32, Y: 64}).
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
history := nodegraph.NewHistory(model.Graph).Limit(100)

nodegraph.New("workflow", model.Graph).
	OnPaste(func(fragment nodegraph.Fragment, offset nodegraph.Point) {
		next := nodegraph.PasteFragment(model.Graph, fragment,
			func(node nodegraph.Node) string { return newNodeID(node) },
			func(edge nodegraph.Edge) string { return newEdgeID(edge) },
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
nodegraph.New("workflow", graph).
	OnEdgesChange(func(changes []nodegraph.EdgeChange) {
		send(EdgesChanged{Edges: nodegraph.ApplyEdgeChanges(model.Edges, changes)})
	})
```

Once the graph has focus, Backspace or Delete requests removal of selected
elements. `NodesDeletable` and `EdgesDeletable` configure the global defaults;
`Node.Deletable(false)` and `Edge.Deletable(false)` override either setting.
Removing a node also requests removal of every connected edge.

## Box selection and connections

With `SelectionOnDrag(true)`, primary-button dragging on empty space selects
nodes while middle-button dragging continues to pan. Use
`SelectionBoxMode(nodegraph.SelectionBoxPartial)` for intersecting nodes, or
`NodeGraphSelectionBoxFull` to require full containment.

Dragging from an output port to an input port shows a connection preview.
`IsValidConnection` rejects candidates according to application rules; release
on a valid input delivers `OnConnect`:

```go
nodegraph.New("workflow", graph).
	IsValidConnection(func(connection nodegraph.Connection) bool {
		return connection.Source.NodeID != connection.Target.NodeID
	}).
	OnConnect(func(connection nodegraph.Connection) {
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
source := nodegraph.NewPort("payload", "Payload").Type("json")
target := nodegraph.NewPort("input", "Input").Type("json").MaxConnections(1)
```

## Edge reconnection

Drag an existing edge endpoint to request that endpoint's replacement.
Reconnection uses `IsValidConnection` just like a new connection, but the
component never changes application data itself. `OnReconnect` receives the
old edge and complete replacement connection. `ReconnectNodeGraphEdge` is a
pure helper that writes the change back while retaining the edge ID:

```go
nodegraph.New("workflow", graph).
	OnReconnect(func(oldEdge nodegraph.Edge, connection nodegraph.Connection) {
		send(EdgeReconnected{
			Edges: nodegraph.ReconnectEdge(oldEdge, connection, model.Edges),
		})
	})
```

Both endpoints are reconnectable by default. `EdgesReconnectable(false)`
disables that default, while an individual edge can limit its range:

```go
sourceOnly := nodegraph.NewEdge("source-only", source, target).
	Reconnectable(nodegraph.ReconnectSource)
targetOnly := nodegraph.NewEdge("target-only", source, target).
	Reconnectable(nodegraph.ReconnectTarget)
```

`NodeGraphReconnectBoth` permits both endpoints and
`NodeGraphReconnectNone` disables reconnection for that edge. A source
endpoint may only land on an output port and a target endpoint may only land
on an input port, preserving the output-to-input direction in every callback.

## Edge styles and labels

Edges use Bezier curves by default, and can switch to straight, stepped, or
rounded stepped geometry. Drawing and hit testing share the same geometry:

```go
edge := nodegraph.NewEdge("request-body", source, target).
	Type(nodegraph.EdgeSmoothStep).
	Width(1.5).
	Dashed(true).
	Markers(nodegraph.MarkerNone, nodegraph.MarkerArrow).
	Label("JSON")
```

`Animated(true)` moves the dash pattern along the edge and enables dashed
rendering. `Width` is expressed in dp; zero uses the theme default. `Color`,
`Selected`, and `Selectable` remain independently controlled.

`LabelContent` replaces a static label with any FlowUI Widget. Its state is
scoped by graph and edge ID, making it appropriate for interactive buttons,
toggles, and other inline controls:

```go
edge := nodegraph.NewEdge("request-body", source, target).
	LabelContent(button.New("Inspect"))
```

## Custom node content and handles

Nodes can provide an inner widget with `Content`. Content receives a stable
state scope derived from the graph and node IDs. When `WithSize` does not set a
dimension, the widget's measured size contributes to the automatic node size;
explicit dimensions still take precedence:

```go
content := ui.WidgetFunc(func(ctx *ui.Context, gtx layout.Context) layout.Dimensions {
	return text.New("Running").Layout(ctx, gtx)
})
node := nodegraph.NewNode("worker", "Worker", nodegraph.Point{X: 120, Y: 80}).
	Content(content).
	Inputs(nodegraph.NewPort("request", "Request").Position(nodegraph.HandleTop)).
	Outputs(nodegraph.NewPort("result", "Result").Position(nodegraph.HandleBottom))
```

Inputs default to the left side and outputs to the right. `Port.Position` can
place either role on `NodeGraphHandleLeft`, `NodeGraphHandleRight`,
`NodeGraphHandleTop`, or `NodeGraphHandleBottom`. Edge anchors, hit testing, and
reconnection all use the same resolved positions. Custom content replaces the
built-in title/body rendering; handles remain rendered and managed by
NodeGraph.

Graph mutations remain controlled callbacks owned by the application model.

See the component tests under `widget/nodegraph` for interaction coverage.
The code in this chapter is intended to be combined with an application.
