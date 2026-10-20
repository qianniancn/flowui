package main

import (
	"strconv"
	"strings"

	"gioui.org/font"
	lucide "github.com/qianniancn/flowui-icons-lucide"
	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme/base"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/button"
	"github.com/qianniancn/flowui/widget/icon"
	"github.com/qianniancn/flowui/widget/input"
	"github.com/qianniancn/flowui/widget/nodegraph"
	"github.com/qianniancn/flowui/widget/tabs"
	"github.com/qianniancn/flowui/widget/text"
	"github.com/qianniancn/flowui/window"
)

type model struct {
	page            string
	viewport        nodegraph.Viewport
	nodes           []nodegraph.Node
	groupNodes      []nodegraph.Node
	edges           []nodegraph.Edge
	databaseNodes   []nodegraph.Node
	databaseEdges   []nodegraph.Edge
	creatorTitle    string
	creatorNodes    []nodegraph.Node
	creatorEdges    []nodegraph.Edge
	creatorSequence int
}

type message struct {
	page            string
	pageChanged     bool
	viewport        nodegraph.Viewport
	viewportChanged bool
	nodeChanges     []nodegraph.NodeChange
	nodeGraph       string
	edgeChanges     []nodegraph.EdgeChange
	edgeGraph       string
	connection      *nodegraph.Connection
	connectionGraph string
	reconnection    *edgeReconnection
	creatorTitle    string
	creatorTitleSet bool
	addCreatorNode  bool
	creatorPosition *nodegraph.Point
}

type edgeReconnection struct {
	oldEdge    nodegraph.Edge
	connection nodegraph.Connection
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
			value.groupNodes = nodegraph.ApplyNodeChanges(value.groupNodes, msg.nodeChanges)
		case "creator":
			value.creatorNodes = nodegraph.ApplyNodeChanges(value.creatorNodes, msg.nodeChanges)
		case "database":
			value.databaseNodes = nodegraph.ApplyNodeChanges(value.databaseNodes, msg.nodeChanges)
		default:
			value.nodes = nodegraph.ApplyNodeChanges(value.nodes, msg.nodeChanges)
		}
	}
	if len(msg.edgeChanges) > 0 {
		switch msg.edgeGraph {
		case "creator":
			value.creatorEdges = nodegraph.ApplyEdgeChanges(value.creatorEdges, msg.edgeChanges)
		case "database":
			value.databaseEdges = nodegraph.ApplyEdgeChanges(value.databaseEdges, msg.edgeChanges)
		default:
			value.edges = nodegraph.ApplyEdgeChanges(value.edges, msg.edgeChanges)
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
			value.databaseEdges = nodegraph.ReconnectEdge(msg.reconnection.oldEdge, msg.reconnection.connection, value.databaseEdges)
		default:
			value.edges = nodegraph.ReconnectEdge(msg.reconnection.oldEdge, msg.reconnection.connection, value.edges)
		}
	}
	return nil
}

func view(_ *ui.Context, value model, send ui.Send[message]) ui.Widget {
	return layout.Column(
		text.New("Node Graph").Size(20),
		tabs.New("node-graph-pages", value.page, []tabs.TabItem{
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
	graph := nodegraph.Graph{Nodes: value.databaseNodes, Edges: value.databaseEdges}
	return nodegraph.New("relational-data", graph).
		Height(520).
		FitView(true).
		FitViewPadding(.16).
		GridPattern(nodegraph.GridDots).
		SelectionMode(nodegraph.SelectionMultiple).
		SelectionOnDrag(true).
		EdgesSelectable(true).
		NodesDeletable(true).
		EdgesDeletable(true).
		SnapToGrid(true).
		SnapGrid(16, 16).
		OnNodesChange(func(changes []nodegraph.NodeChange) { send(message{nodeChanges: changes, nodeGraph: "database"}) }).
		OnEdgesChange(func(changes []nodegraph.EdgeChange) { send(message{edgeChanges: changes, edgeGraph: "database"}) }).
		OnConnect(func(connection nodegraph.Connection) {
			send(message{connection: &connection, connectionGraph: "database"})
		}).
		OnReconnect(func(oldEdge nodegraph.Edge, connection nodegraph.Connection) {
			send(message{reconnection: &edgeReconnection{oldEdge: oldEdge, connection: connection, graph: "database"}})
		})
}

func creatorPanel(value model, send ui.Send[message]) ui.Widget {
	addNode := func() { send(message{addCreatorNode: true}) }
	controls := layout.Row(
		(input.New(value.creatorTitle).Key("creator-title").
			Placeholder("Node title").
			OnChange(func(title string) { send(message{creatorTitle: title, creatorTitleSet: true}) }).
			OnSubmit(func(title string) { send(message{creatorTitle: title, creatorTitleSet: true, addCreatorNode: true}) })).Style(style.New().Width(180)),
		button.New("").Content(layout.Row(
			icon.New(lucide.Plus).Size(16),
			text.New("Add").Size(13),
		).Gap(6).AlignMiddle()).
			Label("Add node").
			OnClick(addNode),
	).Gap(6).AlignMiddle()

	return nodegraph.New("creator", nodegraph.Graph{Nodes: value.creatorNodes, Edges: value.creatorEdges}).
		Height(520).
		DefaultViewport(nodegraph.Viewport{Origin: nodegraph.Point{X: -32, Y: -32}, Zoom: 1}).
		GridPattern(nodegraph.GridDots).
		SelectionMode(nodegraph.SelectionMultiple).
		SelectionOnDrag(true).
		EdgesSelectable(true).
		NodesDeletable(true).
		EdgesDeletable(true).
		SnapToGrid(true).
		SnapGrid(16, 16).
		Panel("creator-controls", nodegraph.PanelTopLeft, controls).
		OnCanvasDoubleClick(func(event nodegraph.CanvasEvent) {
			position := event.Position
			send(message{addCreatorNode: true, creatorPosition: &position})
		}).
		OnNodesChange(func(changes []nodegraph.NodeChange) { send(message{nodeChanges: changes, nodeGraph: "creator"}) }).
		OnEdgesChange(func(changes []nodegraph.EdgeChange) { send(message{edgeChanges: changes, edgeGraph: "creator"}) }).
		OnConnect(func(connection nodegraph.Connection) {
			send(message{connection: &connection, connectionGraph: "creator"})
		})
}

func editingPanel(value model, send ui.Send[message]) ui.Widget {
	graph := nodegraph.Graph{
		Nodes: value.nodes,
		Edges: value.edges,
	}
	return nodegraph.New("workflow", graph).
		Height(520).
		Viewport(value.viewport).
		FitView(true).
		FitViewPadding(.1).
		SelectionMode(nodegraph.SelectionMultiple).
		SelectionOnDrag(true).
		EdgesSelectable(true).
		SnapToGrid(true).
		SnapGrid(16, 16).
		OnViewportChange(func(next nodegraph.Viewport) { send(message{viewport: next, viewportChanged: true}) }).
		OnNodesChange(func(changes []nodegraph.NodeChange) { send(message{nodeChanges: changes, nodeGraph: "editing"}) }).
		OnEdgesChange(func(changes []nodegraph.EdgeChange) { send(message{edgeChanges: changes}) }).
		OnReconnect(func(oldEdge nodegraph.Edge, connection nodegraph.Connection) {
			send(message{reconnection: &edgeReconnection{oldEdge: oldEdge, connection: connection}})
		}).
		OnConnect(func(connection nodegraph.Connection) { send(message{connection: &connection}) })
}

func groupsPanel(value model, send ui.Send[message]) ui.Widget {
	graph := nodegraph.Graph{Nodes: value.groupNodes, Edges: groupEdges()}
	return nodegraph.New("groups", graph).Height(520).FitView(true).NodesResizable(true).ResizeHandles(nodegraph.ResizeHandleAll).GridPattern(nodegraph.GridDots).
		OnNodesChange(func(changes []nodegraph.NodeChange) { send(message{nodeChanges: changes, nodeGraph: "groups"}) })
}

func overlaysPanel() ui.Widget {
	graph := nodegraph.Graph{Nodes: []nodegraph.Node{
		nodegraph.NewNode("request", "Request", nodegraph.Point{X: 100, Y: 160}).Outputs(nodegraph.NewPort("out", "Out")),
		nodegraph.NewNode("response", "Response", nodegraph.Point{X: 480, Y: 160}).Inputs(nodegraph.NewPort("in", "In")),
	}, Edges: []nodegraph.Edge{nodegraph.NewEdge("request-response", nodegraph.NewEndpoint("request", "out"), nodegraph.NewEndpoint("response", "in")).Label("200 OK")}}
	return nodegraph.New("overlays", graph).
		Height(520).
		FitView(true).
		GridPattern(nodegraph.GridDots).
		Panel("canvas-actions", nodegraph.PanelTopRight, button.New("Fit view")).
		ViewportOverlay("note", nodegraph.Point{X: 270, Y: 96}, text.New("Viewport overlay").Size(12)).
		NodeToolbar("request-tools", "request", button.New("Inspect").Size(11)).
		EdgeToolbar("edge-tools", "request-response", button.New("Trace").Size(11))
}

func largeGraphPanel() ui.Widget {
	return nodegraph.New("large-graph", nodegraph.Graph{Nodes: largeGraphNodes(), Edges: largeGraphEdges()}).
		Height(520).
		FitView(true).
		Culling(true).
		GridPattern(nodegraph.GridDots)
}

func databaseNodes() []nodegraph.Node {
	return []nodegraph.Node{
		databaseNode("customers", "customers", nodegraph.Point{X: 60, Y: 80}, []string{
			"PK  id · uuid",
			"    email · text",
			"    name · text",
			"    created_at · timestamp",
		}),
		databaseNode("orders", "orders", nodegraph.Point{X: 390, Y: 80}, []string{
			"PK  id · uuid",
			"FK  customer_id · uuid",
			"    status · text",
			"    total · decimal",
			"    created_at · timestamp",
		}),
		databaseNode("products", "products", nodegraph.Point{X: 60, Y: 330}, []string{
			"PK  id · uuid",
			"    sku · text",
			"    name · text",
			"    unit_price · decimal",
		}),
		databaseNode("order-items", "order_items", nodegraph.Point{X: 390, Y: 330}, []string{
			"PK  id · uuid",
			"FK  order_id · uuid",
			"FK  product_id · uuid",
			"    quantity · integer",
			"    unit_price · decimal",
		}),
	}
}

func databaseNode(id, title string, position nodegraph.Point, fields []string) nodegraph.Node {
	inputs := make([]nodegraph.Port, 0, len(fields))
	outputs := make([]nodegraph.Port, 0, len(fields))
	for index := range fields {
		portID := "field-" + strconv.Itoa(index)
		inputs = append(inputs, nodegraph.NewPort(portID, " ").Type("uuid").Position(nodegraph.HandleRight))
		outputs = append(outputs, nodegraph.NewPort(portID, " ").Type("uuid").Position(nodegraph.HandleLeft))
	}
	return nodegraph.NewNode(id, title, position).
		WithSize(232, 154).
		Inputs(inputs...).
		Outputs(outputs...).
		Content(databaseTable(title, fields))
}

func databaseTable(title string, fields []string) ui.Widget {
	rows := make([]ui.Widget, 0, len(fields)+2)
	rows = append(rows,
		text.New(title).Size(12).Weight(font.SemiBold),
		layout.Spacer(0, 3),
	)
	for _, field := range fields {
		rows = append(rows, text.New(field).Size(11))
	}
	return layout.Column(rows...).Gap(5)
}

func databaseEdges() []nodegraph.Edge {
	return []nodegraph.Edge{
		nodegraph.NewEdge("customers-orders", nodegraph.NewEndpoint("customers", "field-0"), nodegraph.NewEndpoint("orders", "field-1")).Label("customer_id"),
		nodegraph.NewEdge("orders-order-items", nodegraph.NewEndpoint("orders", "field-0"), nodegraph.NewEndpoint("order-items", "field-1")).Label("order_id"),
		nodegraph.NewEdge("products-order-items", nodegraph.NewEndpoint("products", "field-0"), nodegraph.NewEndpoint("order-items", "field-2")).Label("product_id"),
	}
}

func main() {
	ui.Run(ui.NewProgram(model{page: "editing", nodes: exampleNodes(), groupNodes: groupNodes(), edges: exampleEdges(), databaseNodes: databaseNodes(), databaseEdges: databaseEdges()}, update, view), window.Title("FlowUI Node Graph"), window.Size(1060, 680), window.Theme(base.Light()))
}

func creatorNodePosition(sequence int) nodegraph.Point {
	const columns = 4
	index := sequence - 1
	return nodegraph.Point{X: float32(48 + (index%columns)*208), Y: float32(96 + (index/columns)*136)}
}

func newCreatorNode(sequence int, title string, position nodegraph.Point) nodegraph.Node {
	return nodegraph.NewNode(
		"custom-"+strconv.Itoa(sequence),
		title,
		position,
	).Inputs(nodegraph.NewPort("in", "In")).Outputs(nodegraph.NewPort("out", "Out"))
}

func largeGraphNodes() []nodegraph.Node {
	nodes := make([]nodegraph.Node, 0, 144)
	for row := range 12 {
		for column := range 12 {
			id := largeNodeID(row, column)
			nodes = append(nodes, nodegraph.NewNode(id, "Worker", nodegraph.Point{X: float32(column * 220), Y: float32(row * 140)}).Inputs(nodegraph.NewPort("in", "In")).Outputs(nodegraph.NewPort("out", "Out")))
		}
	}
	return nodes
}

func largeGraphEdges() []nodegraph.Edge {
	edges := make([]nodegraph.Edge, 0, 264)
	for row := range 12 {
		for column := range 12 {
			if column < 11 {
				edges = append(edges, nodegraph.NewEdge("right-"+largeNodeID(row, column), nodegraph.NewEndpoint(largeNodeID(row, column), "out"), nodegraph.NewEndpoint(largeNodeID(row, column+1), "in")))
			}
			if row < 11 {
				edges = append(edges, nodegraph.NewEdge("down-"+largeNodeID(row, column), nodegraph.NewEndpoint(largeNodeID(row, column), "out"), nodegraph.NewEndpoint(largeNodeID(row+1, column), "in")).Type(nodegraph.EdgeSmoothStep))
			}
		}
	}
	return edges
}

func largeNodeID(row, column int) string {
	return "node-" + strconv.Itoa(row) + "-" + strconv.Itoa(column)
}

func groupNodes() []nodegraph.Node {
	return []nodegraph.Node{
		nodegraph.NewNode("pipeline", "Pipeline", nodegraph.Point{X: 80, Y: 72}).WithSize(480, 300).Resizable(true).ResizeHandles(nodegraph.ResizeHandleAll),
		nodegraph.NewNode("decode", "Decode", nodegraph.Point{X: 32, Y: 72}).Parent("pipeline").ConstrainToParent(true).WithSize(170, 92).Outputs(nodegraph.NewPort("json", "JSON").Type("json")),
		nodegraph.NewNode("validate", "Validate", nodegraph.Point{X: 260, Y: 72}).Parent("pipeline").ConstrainToParent(true).WithSize(170, 92).Inputs(nodegraph.NewPort("json", "JSON").Type("json").MaxConnections(1)),
	}
}

func groupEdges() []nodegraph.Edge {
	return []nodegraph.Edge{nodegraph.NewEdge("decode-validate", nodegraph.NewEndpoint("decode", "json"), nodegraph.NewEndpoint("validate", "json"))}
}

func exampleNodes() []nodegraph.Node {
	return []nodegraph.Node{
		nodegraph.NewNode("source-a", "Source A", nodegraph.Point{X: 80, Y: 64}).
			Outputs(nodegraph.NewPort("out", "Out")),
		nodegraph.NewNode("source-b", "Source B", nodegraph.Point{X: 80, Y: 220}).
			Outputs(nodegraph.NewPort("out", "Out")),
		nodegraph.NewNode("source-c", "Source C", nodegraph.Point{X: 80, Y: 376}).
			Outputs(nodegraph.NewPort("out", "Out")),
		nodegraph.NewNode("target-a", "Target A", nodegraph.Point{X: 650, Y: 64}).
			Inputs(nodegraph.NewPort("in", "In")),
		nodegraph.NewNode("target-b", "Target B", nodegraph.Point{X: 650, Y: 220}).
			Inputs(nodegraph.NewPort("in", "In")),
		nodegraph.NewNode("target-c", "Target C", nodegraph.Point{X: 650, Y: 376}).
			Inputs(nodegraph.NewPort("in", "In")),
	}
}

func exampleEdges() []nodegraph.Edge {
	return []nodegraph.Edge{
		nodegraph.NewEdge("source-only", nodegraph.NewEndpoint("source-a", "out"), nodegraph.NewEndpoint("target-a", "in")).Reconnectable(nodegraph.ReconnectSource),
		nodegraph.NewEdge("target-only", nodegraph.NewEndpoint("source-b", "out"), nodegraph.NewEndpoint("target-b", "in")).Reconnectable(nodegraph.ReconnectTarget),
		nodegraph.NewEdge("both", nodegraph.NewEndpoint("source-c", "out"), nodegraph.NewEndpoint("target-c", "in")).Reconnectable(nodegraph.ReconnectBoth),
	}
}

func appendConnection(edges []nodegraph.Edge, connection nodegraph.Connection) []nodegraph.Edge {
	for _, edge := range edges {
		if edge.Source == connection.Source && edge.Target == connection.Target {
			return edges
		}
	}
	id := connection.Source.NodeID + "-" + connection.Source.PortID + "-" + connection.Target.NodeID + "-" + connection.Target.PortID
	return append(edges, nodegraph.NewEdge(id, connection.Source, connection.Target))
}
