package main

import (
	"strconv"
	"strings"

	"gioui.org/font"
	"github.com/qianniancn/flowui-icons-lucide"
	"github.com/qianniancn/flowui/ui"
)

type model struct {
	page            string
	viewport        ui.NodeGraphViewport
	nodes           []ui.NodeGraphNode
	groupNodes      []ui.NodeGraphNode
	edges           []ui.NodeGraphEdge
	databaseNodes   []ui.NodeGraphNode
	databaseEdges   []ui.NodeGraphEdge
	creatorTitle    string
	creatorNodes    []ui.NodeGraphNode
	creatorEdges    []ui.NodeGraphEdge
	creatorSequence int
}

type message struct {
	page            string
	pageChanged     bool
	viewport        ui.NodeGraphViewport
	viewportChanged bool
	nodeChanges     []ui.NodeGraphNodeChange
	nodeGraph       string
	edgeChanges     []ui.NodeGraphEdgeChange
	edgeGraph       string
	connection      *ui.NodeGraphConnection
	connectionGraph string
	reconnection    *edgeReconnection
	creatorTitle    string
	creatorTitleSet bool
	addCreatorNode  bool
	creatorPosition *ui.NodeGraphPoint
}

type edgeReconnection struct {
	oldEdge    ui.NodeGraphEdge
	connection ui.NodeGraphConnection
	graph      string
}

func update(value *model, msg message) ui.Cmd[message] {
	if msg.pageChanged {
		value.page = msg.page
	}
	if msg.creatorTitleSet {
		value.creatorTitle = msg.creatorTitle
	}
	if msg.addCreatorNode {
		value.creatorSequence++
		title := strings.TrimSpace(value.creatorTitle)
		if title == "" {
			title = "Node " + strconv.Itoa(value.creatorSequence)
		}
		position := creatorNodePosition(value.creatorSequence)
		if msg.creatorPosition != nil {
			position = *msg.creatorPosition
		}
		value.creatorNodes = append(value.creatorNodes, newCreatorNode(value.creatorSequence, title, position))
		value.creatorTitle = ""
	}
	if msg.viewportChanged {
		value.viewport = msg.viewport
	}
	if len(msg.nodeChanges) > 0 {
		switch msg.nodeGraph {
		case "groups":
			value.groupNodes = ui.ApplyNodeGraphChanges(value.groupNodes, msg.nodeChanges)
		case "creator":
			value.creatorNodes = ui.ApplyNodeGraphChanges(value.creatorNodes, msg.nodeChanges)
		case "database":
			value.databaseNodes = ui.ApplyNodeGraphChanges(value.databaseNodes, msg.nodeChanges)
		default:
			value.nodes = ui.ApplyNodeGraphChanges(value.nodes, msg.nodeChanges)
		}
	}
	if len(msg.edgeChanges) > 0 {
		switch msg.edgeGraph {
		case "creator":
			value.creatorEdges = ui.ApplyNodeGraphEdgeChanges(value.creatorEdges, msg.edgeChanges)
		case "database":
			value.databaseEdges = ui.ApplyNodeGraphEdgeChanges(value.databaseEdges, msg.edgeChanges)
		default:
			value.edges = ui.ApplyNodeGraphEdgeChanges(value.edges, msg.edgeChanges)
		}
	}
	if msg.connection != nil {
		switch msg.connectionGraph {
		case "creator":
			value.creatorEdges = appendConnection(value.creatorEdges, *msg.connection)
		case "database":
			value.databaseEdges = appendConnection(value.databaseEdges, *msg.connection)
		default:
			value.edges = appendConnection(value.edges, *msg.connection)
		}
	}
	if msg.reconnection != nil {
		switch msg.reconnection.graph {
		case "database":
			value.databaseEdges = ui.ReconnectNodeGraphEdge(msg.reconnection.oldEdge, msg.reconnection.connection, value.databaseEdges)
		default:
			value.edges = ui.ReconnectNodeGraphEdge(msg.reconnection.oldEdge, msg.reconnection.connection, value.edges)
		}
	}
	return nil
}

func view(_ *ui.Context, value model, send ui.Send[message]) ui.Widget {
	return ui.Column(
		ui.Text("Node Graph").Size(20),
		ui.Tabs("node-graph-pages", value.page, []ui.TabItem{
			{Key: "editing", Label: "Editing", Panel: editingPanel(value, send)},
			{Key: "creator", Label: "Create nodes", Panel: creatorPanel(value, send)},
			{Key: "database", Label: "Relational data", Panel: databasePanel(value, send)},
			{Key: "groups", Label: "Groups", Panel: groupsPanel(value, send)},
			{Key: "overlays", Label: "Overlays", Panel: overlaysPanel()},
			{Key: "large", Label: "Large graph", Panel: largeGraphPanel()},
		}).OnChange(func(page string) { send(message{page: page, pageChanged: true}) }),
	).Gap(10)
}

func databasePanel(value model, send ui.Send[message]) ui.Widget {
	graph := ui.NodeGraphData{Nodes: value.databaseNodes, Edges: value.databaseEdges}
	return ui.NodeGraph("relational-data", graph).
		Height(520).
		FitView(true).
		FitViewPadding(.16).
		GridPattern(ui.NodeGraphGridDots).
		SelectionMode(ui.NodeGraphSelectionMultiple).
		SelectionOnDrag(true).
		EdgesSelectable(true).
		NodesDeletable(true).
		EdgesDeletable(true).
		SnapToGrid(true).
		SnapGrid(16, 16).
		OnNodesChange(func(changes []ui.NodeGraphNodeChange) { send(message{nodeChanges: changes, nodeGraph: "database"}) }).
		OnEdgesChange(func(changes []ui.NodeGraphEdgeChange) { send(message{edgeChanges: changes, edgeGraph: "database"}) }).
		OnConnect(func(connection ui.NodeGraphConnection) {
			send(message{connection: &connection, connectionGraph: "database"})
		}).
		OnReconnect(func(oldEdge ui.NodeGraphEdge, connection ui.NodeGraphConnection) {
			send(message{reconnection: &edgeReconnection{oldEdge: oldEdge, connection: connection, graph: "database"}})
		})
}

func creatorPanel(value model, send ui.Send[message]) ui.Widget {
	addNode := func() { send(message{addCreatorNode: true}) }
	controls := ui.Row(
		ui.Input("creator-title", value.creatorTitle).
			Placeholder("Node title").
			OnChange(func(title string) { send(message{creatorTitle: title, creatorTitleSet: true}) }).
			OnSubmit(func(title string) { send(message{creatorTitle: title, creatorTitleSet: true, addCreatorNode: true}) }).
			Style(ui.Width(180)),
		ui.Button("creator-add", ui.Row(
			ui.Icon(lucide.Plus).Size(16),
			ui.Text("Add").Size(13),
		).Gap(6).AlignMiddle()).
			Label("Add node").
			OnClick(addNode),
	).Gap(6).AlignMiddle()

	return ui.NodeGraph("creator", ui.NodeGraphData{Nodes: value.creatorNodes, Edges: value.creatorEdges}).
		Height(520).
		DefaultViewport(ui.NodeGraphViewport{Origin: ui.NodeGraphPoint{X: -32, Y: -32}, Zoom: 1}).
		GridPattern(ui.NodeGraphGridDots).
		SelectionMode(ui.NodeGraphSelectionMultiple).
		SelectionOnDrag(true).
		EdgesSelectable(true).
		NodesDeletable(true).
		EdgesDeletable(true).
		SnapToGrid(true).
		SnapGrid(16, 16).
		Panel("creator-controls", ui.NodeGraphPanelTopLeft, controls).
		OnCanvasDoubleClick(func(event ui.NodeGraphCanvasEvent) {
			position := event.Position
			send(message{addCreatorNode: true, creatorPosition: &position})
		}).
		OnNodesChange(func(changes []ui.NodeGraphNodeChange) { send(message{nodeChanges: changes, nodeGraph: "creator"}) }).
		OnEdgesChange(func(changes []ui.NodeGraphEdgeChange) { send(message{edgeChanges: changes, edgeGraph: "creator"}) }).
		OnConnect(func(connection ui.NodeGraphConnection) {
			send(message{connection: &connection, connectionGraph: "creator"})
		})
}

func editingPanel(value model, send ui.Send[message]) ui.Widget {
	graph := ui.NodeGraphData{
		Nodes: value.nodes,
		Edges: value.edges,
	}
	return ui.NodeGraph("workflow", graph).
		Height(520).
		Viewport(value.viewport).
		FitView(true).
		FitViewPadding(.1).
		SelectionMode(ui.NodeGraphSelectionMultiple).
		SelectionOnDrag(true).
		EdgesSelectable(true).
		SnapToGrid(true).
		SnapGrid(16, 16).
		OnViewportChange(func(next ui.NodeGraphViewport) { send(message{viewport: next, viewportChanged: true}) }).
		OnNodesChange(func(changes []ui.NodeGraphNodeChange) { send(message{nodeChanges: changes, nodeGraph: "editing"}) }).
		OnEdgesChange(func(changes []ui.NodeGraphEdgeChange) { send(message{edgeChanges: changes}) }).
		OnReconnect(func(oldEdge ui.NodeGraphEdge, connection ui.NodeGraphConnection) {
			send(message{reconnection: &edgeReconnection{oldEdge: oldEdge, connection: connection}})
		}).
		OnConnect(func(connection ui.NodeGraphConnection) { send(message{connection: &connection}) })
}

func groupsPanel(value model, send ui.Send[message]) ui.Widget {
	graph := ui.NodeGraphData{Nodes: value.groupNodes, Edges: groupEdges()}
	return ui.NodeGraph("groups", graph).Height(520).FitView(true).NodesResizable(true).ResizeHandles(ui.NodeGraphResizeHandleAll).GridPattern(ui.NodeGraphGridDots).
		OnNodesChange(func(changes []ui.NodeGraphNodeChange) { send(message{nodeChanges: changes, nodeGraph: "groups"}) })
}

func overlaysPanel() ui.Widget {
	graph := ui.NodeGraphData{Nodes: []ui.NodeGraphNode{
		ui.NewNodeGraphNode("request", "Request", ui.NodeGraphPoint{X: 100, Y: 160}).Outputs(ui.NewNodeGraphPort("out", "Out")),
		ui.NewNodeGraphNode("response", "Response", ui.NodeGraphPoint{X: 480, Y: 160}).Inputs(ui.NewNodeGraphPort("in", "In")),
	}, Edges: []ui.NodeGraphEdge{ui.NewNodeGraphEdge("request-response", ui.NewNodeGraphEndpoint("request", "out"), ui.NewNodeGraphEndpoint("response", "in")).Label("200 OK")}}
	return ui.NodeGraph("overlays", graph).
		Height(520).
		FitView(true).
		GridPattern(ui.NodeGraphGridDots).
		Panel("canvas-actions", ui.NodeGraphPanelTopRight, ui.Button("fit", ui.Text("Fit view"))).
		ViewportOverlay("note", ui.NodeGraphPoint{X: 270, Y: 96}, ui.Text("Viewport overlay").Size(12)).
		NodeToolbar("request-tools", "request", ui.Button("inspect-request", ui.Text("Inspect").Size(11))).
		EdgeToolbar("edge-tools", "request-response", ui.Button("inspect-edge", ui.Text("Trace").Size(11)))
}

func largeGraphPanel() ui.Widget {
	return ui.NodeGraph("large-graph", ui.NodeGraphData{Nodes: largeGraphNodes(), Edges: largeGraphEdges()}).
		Height(520).
		FitView(true).
		Culling(true).
		GridPattern(ui.NodeGraphGridDots)
}

func databaseNodes() []ui.NodeGraphNode {
	return []ui.NodeGraphNode{
		databaseNode("customers", "customers", ui.NodeGraphPoint{X: 60, Y: 80}, []string{
			"PK  id · uuid",
			"    email · text",
			"    name · text",
			"    created_at · timestamp",
		}),
		databaseNode("orders", "orders", ui.NodeGraphPoint{X: 390, Y: 80}, []string{
			"PK  id · uuid",
			"FK  customer_id · uuid",
			"    status · text",
			"    total · decimal",
			"    created_at · timestamp",
		}),
		databaseNode("products", "products", ui.NodeGraphPoint{X: 60, Y: 330}, []string{
			"PK  id · uuid",
			"    sku · text",
			"    name · text",
			"    unit_price · decimal",
		}),
		databaseNode("order-items", "order_items", ui.NodeGraphPoint{X: 390, Y: 330}, []string{
			"PK  id · uuid",
			"FK  order_id · uuid",
			"FK  product_id · uuid",
			"    quantity · integer",
			"    unit_price · decimal",
		}),
	}
}

func databaseNode(id, title string, position ui.NodeGraphPoint, fields []string) ui.NodeGraphNode {
	inputs := make([]ui.NodeGraphPort, 0, len(fields))
	outputs := make([]ui.NodeGraphPort, 0, len(fields))
	for index := range fields {
		portID := "field-" + strconv.Itoa(index)
		inputs = append(inputs, ui.NewNodeGraphPort(portID, " ").Type("uuid").Position(ui.NodeGraphHandleRight))
		outputs = append(outputs, ui.NewNodeGraphPort(portID, " ").Type("uuid").Position(ui.NodeGraphHandleLeft))
	}
	return ui.NewNodeGraphNode(id, title, position).
		WithSize(232, 154).
		Inputs(inputs...).
		Outputs(outputs...).
		Content(databaseTable(title, fields))
}

func databaseTable(title string, fields []string) ui.Widget {
	rows := make([]ui.Widget, 0, len(fields)+2)
	rows = append(rows,
		ui.Text(title).Size(12).Weight(font.SemiBold),
		ui.Spacer(0, 3),
	)
	for _, field := range fields {
		rows = append(rows, ui.Text(field).Size(11))
	}
	return ui.Column(rows...).Gap(5)
}

func databaseEdges() []ui.NodeGraphEdge {
	return []ui.NodeGraphEdge{
		ui.NewNodeGraphEdge("customers-orders", ui.NewNodeGraphEndpoint("customers", "field-0"), ui.NewNodeGraphEndpoint("orders", "field-1")).Label("customer_id"),
		ui.NewNodeGraphEdge("orders-order-items", ui.NewNodeGraphEndpoint("orders", "field-0"), ui.NewNodeGraphEndpoint("order-items", "field-1")).Label("order_id"),
		ui.NewNodeGraphEdge("products-order-items", ui.NewNodeGraphEndpoint("products", "field-0"), ui.NewNodeGraphEndpoint("order-items", "field-2")).Label("product_id"),
	}
}

func main() {
	ui.Run(ui.NewProgram(model{page: "editing", nodes: exampleNodes(), groupNodes: groupNodes(), edges: exampleEdges(), databaseNodes: databaseNodes(), databaseEdges: databaseEdges()}, update, view), ui.Title("FlowUI Node Graph"), ui.Size(1060, 680))
}

func creatorNodePosition(sequence int) ui.NodeGraphPoint {
	const columns = 4
	index := sequence - 1
	return ui.NodeGraphPoint{X: float32(48 + (index%columns)*208), Y: float32(96 + (index/columns)*136)}
}

func newCreatorNode(sequence int, title string, position ui.NodeGraphPoint) ui.NodeGraphNode {
	return ui.NewNodeGraphNode(
		"custom-"+strconv.Itoa(sequence),
		title,
		position,
	).Inputs(ui.NewNodeGraphPort("in", "In")).Outputs(ui.NewNodeGraphPort("out", "Out"))
}

func largeGraphNodes() []ui.NodeGraphNode {
	nodes := make([]ui.NodeGraphNode, 0, 144)
	for row := 0; row < 12; row++ {
		for column := 0; column < 12; column++ {
			id := largeNodeID(row, column)
			nodes = append(nodes, ui.NewNodeGraphNode(id, "Worker", ui.NodeGraphPoint{X: float32(column * 220), Y: float32(row * 140)}).Inputs(ui.NewNodeGraphPort("in", "In")).Outputs(ui.NewNodeGraphPort("out", "Out")))
		}
	}
	return nodes
}

func largeGraphEdges() []ui.NodeGraphEdge {
	edges := make([]ui.NodeGraphEdge, 0, 264)
	for row := 0; row < 12; row++ {
		for column := 0; column < 12; column++ {
			if column < 11 {
				edges = append(edges, ui.NewNodeGraphEdge("right-"+largeNodeID(row, column), ui.NewNodeGraphEndpoint(largeNodeID(row, column), "out"), ui.NewNodeGraphEndpoint(largeNodeID(row, column+1), "in")))
			}
			if row < 11 {
				edges = append(edges, ui.NewNodeGraphEdge("down-"+largeNodeID(row, column), ui.NewNodeGraphEndpoint(largeNodeID(row, column), "out"), ui.NewNodeGraphEndpoint(largeNodeID(row+1, column), "in")).Type(ui.NodeGraphEdgeSmoothStep))
			}
		}
	}
	return edges
}

func largeNodeID(row, column int) string {
	return "node-" + strconv.Itoa(row) + "-" + strconv.Itoa(column)
}

func groupNodes() []ui.NodeGraphNode {
	return []ui.NodeGraphNode{
		ui.NewNodeGraphNode("pipeline", "Pipeline", ui.NodeGraphPoint{X: 80, Y: 72}).WithSize(480, 300).Resizable(true).ResizeHandles(ui.NodeGraphResizeHandleAll),
		ui.NewNodeGraphNode("decode", "Decode", ui.NodeGraphPoint{X: 32, Y: 72}).Parent("pipeline").ConstrainToParent(true).WithSize(170, 92).Outputs(ui.NewNodeGraphPort("json", "JSON").Type("json")),
		ui.NewNodeGraphNode("validate", "Validate", ui.NodeGraphPoint{X: 260, Y: 72}).Parent("pipeline").ConstrainToParent(true).WithSize(170, 92).Inputs(ui.NewNodeGraphPort("json", "JSON").Type("json").MaxConnections(1)),
	}
}

func groupEdges() []ui.NodeGraphEdge {
	return []ui.NodeGraphEdge{ui.NewNodeGraphEdge("decode-validate", ui.NewNodeGraphEndpoint("decode", "json"), ui.NewNodeGraphEndpoint("validate", "json"))}
}

func exampleNodes() []ui.NodeGraphNode {
	return []ui.NodeGraphNode{
		ui.NewNodeGraphNode("source-a", "Source A", ui.NodeGraphPoint{X: 80, Y: 64}).
			Outputs(ui.NewNodeGraphPort("out", "Out")),
		ui.NewNodeGraphNode("source-b", "Source B", ui.NodeGraphPoint{X: 80, Y: 220}).
			Outputs(ui.NewNodeGraphPort("out", "Out")),
		ui.NewNodeGraphNode("source-c", "Source C", ui.NodeGraphPoint{X: 80, Y: 376}).
			Outputs(ui.NewNodeGraphPort("out", "Out")),
		ui.NewNodeGraphNode("target-a", "Target A", ui.NodeGraphPoint{X: 650, Y: 64}).
			Inputs(ui.NewNodeGraphPort("in", "In")),
		ui.NewNodeGraphNode("target-b", "Target B", ui.NodeGraphPoint{X: 650, Y: 220}).
			Inputs(ui.NewNodeGraphPort("in", "In")),
		ui.NewNodeGraphNode("target-c", "Target C", ui.NodeGraphPoint{X: 650, Y: 376}).
			Inputs(ui.NewNodeGraphPort("in", "In")),
	}
}

func exampleEdges() []ui.NodeGraphEdge {
	return []ui.NodeGraphEdge{
		ui.NewNodeGraphEdge("source-only", ui.NewNodeGraphEndpoint("source-a", "out"), ui.NewNodeGraphEndpoint("target-a", "in")).Reconnectable(ui.NodeGraphReconnectSource),
		ui.NewNodeGraphEdge("target-only", ui.NewNodeGraphEndpoint("source-b", "out"), ui.NewNodeGraphEndpoint("target-b", "in")).Reconnectable(ui.NodeGraphReconnectTarget),
		ui.NewNodeGraphEdge("both", ui.NewNodeGraphEndpoint("source-c", "out"), ui.NewNodeGraphEndpoint("target-c", "in")).Reconnectable(ui.NodeGraphReconnectBoth),
	}
}

func appendConnection(edges []ui.NodeGraphEdge, connection ui.NodeGraphConnection) []ui.NodeGraphEdge {
	for _, edge := range edges {
		if edge.Source == connection.Source && edge.Target == connection.Target {
			return edges
		}
	}
	id := connection.Source.NodeID + "-" + connection.Source.PortID + "-" + connection.Target.NodeID + "-" + connection.Target.PortID
	return append(edges, ui.NewNodeGraphEdge(id, connection.Source, connection.Target))
}
