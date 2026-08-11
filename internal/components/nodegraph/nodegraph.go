// Package nodegraph provides the geometry, viewport, and rendering foundation
// for node-based flow editors. Graph data remains owned by the application.
package nodegraph

import (
	"image/color"
	"slices"
	"strings"

	"gioui.org/io/key"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/frame"
)

const (
	stateSlotNodeGraph    = "node-graph"
	defaultNodeWidth      = float32(150)
	defaultNodeHeight     = float32(64)
	nodeHeaderHeight      = float32(26)
	nodePortRowHeight     = float32(20)
	nodeVerticalPad       = float32(8)
	defaultGridSize       = float32(16)
	defaultMinZoom        = float32(0.25)
	defaultMaxZoom        = float32(2)
	defaultFitViewPadding = float32(.1)
	defaultDragThreshold  = float32(3)
)

// Point is a location in graph coordinates. One graph unit maps to one dp at
// zoom 1, making stored graphs independent of display density.
type Point struct {
	X float32
	Y float32
}

// HandlePosition identifies the side of a node where a port is rendered.
type HandlePosition uint8

const (
	HandleLeft HandlePosition = iota
	HandleRight
	HandleTop
	HandleBottom
)

// ResizeHandle is a bit set selecting one or more node resize handles.
type ResizeHandle uint16

const (
	ResizeHandleTop ResizeHandle = 1 << iota
	ResizeHandleRight
	ResizeHandleBottom
	ResizeHandleLeft
	ResizeHandleTopLeft
	ResizeHandleTopRight
	ResizeHandleBottomRight
	ResizeHandleBottomLeft
	ResizeHandleAll = ResizeHandleTop | ResizeHandleRight | ResizeHandleBottom | ResizeHandleLeft | ResizeHandleTopLeft | ResizeHandleTopRight | ResizeHandleBottomRight | ResizeHandleBottomLeft
)

// Size is a width and height in graph coordinates.
type Size struct {
	Width  float32
	Height float32
}

// Viewport identifies the graph point at the canvas origin and its zoom.
type Viewport struct {
	Origin Point
	Zoom   float32
}

// CanvasEvent describes a pointer interaction on empty graph canvas space.
type CanvasEvent struct {
	Position  Point
	Modifiers key.Modifiers
}

// NodeEvent describes a pointer interaction with a node.
type NodeEvent struct {
	NodeID    string
	Position  Point
	Modifiers key.Modifiers
}

// EdgeEvent describes a pointer interaction with an edge.
type EdgeEvent struct {
	EdgeID    string
	Position  Point
	Modifiers key.Modifiers
}

// DropEvent describes externally transferred data dropped on the graph.
// Data is a copy owned by the receiver and is limited to 1 MiB.
type DropEvent struct {
	Position Point
	MIME     string
	Data     []byte
}

// PanelPosition positions content above the canvas without applying viewport
// pan or zoom.
type PanelPosition uint8

const (
	PanelTopLeft PanelPosition = iota
	PanelTopCenter
	PanelTopRight
	PanelBottomLeft
	PanelBottomCenter
	PanelBottomRight
)

// Panel is fixed canvas content, comparable to XYFlow's Panel. Key scopes any
// state owned by Content to this NodeGraph instance.
type Panel struct {
	Key      string
	Position PanelPosition
	Content  frame.Widget
}

// ViewportOverlay is content placed in graph coordinates. It moves and scales
// with the graph viewport, comparable to XYFlow's ViewportPortal.
type ViewportOverlay struct {
	Key      string
	Position Point
	Content  frame.Widget
}

type nodeToolbar struct {
	key, nodeID string
	content     frame.Widget
}
type edgeToolbar struct {
	key, edgeID string
	content     frame.Widget
}

// Port is one named input or output connection point on a Node.
type Port struct {
	ID          string
	Label       string
	Color       color.NRGBA
	connectable *bool
	position    *HandlePosition
	dataType    string
	maxLinks    int
}

// NewPort creates a named node port. The label defaults to its ID.
func NewPort(id, label string) Port {
	if id == "" {
		panic("flowui: empty node graph port ID")
	}
	if label == "" {
		label = id
	}
	return Port{ID: id, Label: label}
}

// Connectable overrides whether this port participates in connection gestures.
func (p Port) Connectable(enabled bool) Port {
	p.connectable = &enabled
	return p
}

// Position places this port on a specific side of its node. Without an
// explicit position, inputs use the left side and outputs use the right side.
func (p Port) Position(position HandlePosition) Port {
	if position > HandleBottom {
		panic("flowui: invalid node graph handle position")
	}
	p.position = &position
	return p
}

// Type assigns an optional connection type. When both endpoints define a
// type, NodeGraph only accepts connections with matching values.
func (p Port) Type(value string) Port {
	p.dataType = value
	return p
}

// MaxConnections limits connections on this port. Zero keeps the default of
// no component-level limit; use one for a single-value input or output.
func (p Port) MaxConnections(maximum int) Port {
	if maximum < 0 {
		panic("flowui: node graph port maximum connections must not be negative")
	}
	p.maxLinks = maximum
	return p
}

func (p Port) handlePosition(output bool) HandlePosition {
	if p.position != nil {
		return *p.position
	}
	if output {
		return HandleRight
	}
	return HandleLeft
}

func (p Port) isConnectable(defaultValue bool) bool {
	if p.connectable != nil {
		return *p.connectable
	}
	return defaultValue
}

// Node is one positioned graph node. Width and Height may be zero to use a
// compact size derived from the number of ports.
type Node struct {
	ID          string
	Title       string
	Position    Point
	Size        Size
	InputPorts  []Port
	OutputPorts []Port
	// content replaces the built-in title/body rendering while handles remain
	// managed by NodeGraph. A nil value keeps the default title and port labels.
	content          frame.Widget
	Selected         bool
	draggable        *bool
	resizable        *bool
	selectable       *bool
	connectable      *bool
	deletable        *bool
	minSize          Size
	maxSize          Size
	parentID         string
	parentBound      *bool
	resizeHandles    ResizeHandle
	hasResizeHandles bool
}

// NewNode creates a graph node at position.
func NewNode(id, title string, position Point) Node {
	if id == "" {
		panic("flowui: empty node graph node ID")
	}
	if title == "" {
		title = id
	}
	return Node{ID: id, Title: title, Position: position}
}

// WithSize sets the explicit node size in graph coordinates.
func (n Node) WithSize(width, height float32) Node {
	if width < 0 || height < 0 || !finite(width) || !finite(height) {
		panic("flowui: invalid node graph node size")
	}
	n.Size = Size{Width: width, Height: height}
	return n
}

// Parent makes this node a child of parentID. A child position is stored in
// the parent's local graph coordinates; its resolved visual position includes
// every ancestor's position. Passing an empty ID removes the parent relation.
func (n Node) Parent(parentID string) Node {
	if parentID == n.ID && parentID != "" {
		panic("flowui: node graph node cannot be its own parent")
	}
	n.parentID = parentID
	return n
}

// ConstrainToParent keeps interactive child dragging inside its direct
// parent's bounds. It does not alter application-owned positions directly.
func (n Node) ConstrainToParent(enabled bool) Node {
	n.parentBound = &enabled
	return n
}

func (n Node) isConstrainedToParent() bool {
	return n.parentBound != nil && *n.parentBound
}

// Inputs assigns input ports to the node.
func (n Node) Inputs(ports ...Port) Node {
	n.InputPorts = append([]Port(nil), ports...)
	return n
}

// Outputs assigns output ports to the node.
func (n Node) Outputs(ports ...Port) Node {
	n.OutputPorts = append([]Port(nil), ports...)
	return n
}

// Content sets a custom widget rendered inside the node body. If a width or
// height is not supplied with WithSize, the measured content contributes to
// the automatic node size.
func (n Node) Content(content frame.Widget) Node {
	n.content = content
	return n
}

// Select sets the node's controlled selected state.
func (n Node) Select(selected bool) Node {
	n.Selected = selected
	return n
}

// Draggable overrides whether this node may be dragged.
func (n Node) Draggable(enabled bool) Node {
	n.draggable = &enabled
	return n
}

// Resizable overrides whether this node exposes a resize handle when selected.
func (n Node) Resizable(enabled bool) Node {
	n.resizable = &enabled
	return n
}

// ResizeHandles overrides the handles shown for this node while it is
// resizable. The default is the lower-right corner for compatibility.
func (n Node) ResizeHandles(handles ResizeHandle) Node {
	if handles&^ResizeHandleAll != 0 {
		panic("flowui: invalid node graph resize handles")
	}
	n.resizeHandles = handles
	n.hasResizeHandles = true
	return n
}

func (n Node) resizeHandleSet(defaultValue ResizeHandle) ResizeHandle {
	if n.hasResizeHandles {
		return n.resizeHandles
	}
	return defaultValue
}

// SizeRange constrains interactive resizing in graph coordinates. A maximum
// of zero leaves that axis unbounded.
func (n Node) SizeRange(minWidth, minHeight, maxWidth, maxHeight float32) Node {
	if minWidth < 0 || minHeight < 0 || maxWidth < 0 || maxHeight < 0 || !finite(minWidth) || !finite(minHeight) || !finite(maxWidth) || !finite(maxHeight) || (maxWidth > 0 && maxWidth < minWidth) || (maxHeight > 0 && maxHeight < minHeight) {
		panic("flowui: invalid node graph resize range")
	}
	n.minSize = Size{Width: minWidth, Height: minHeight}
	n.maxSize = Size{Width: maxWidth, Height: maxHeight}
	return n
}

// Selectable overrides whether this node may be selected.
func (n Node) Selectable(enabled bool) Node {
	n.selectable = &enabled
	return n
}

// Connectable overrides whether this node's ports accept connection gestures.
func (n Node) Connectable(enabled bool) Node {
	n.connectable = &enabled
	return n
}

// Deletable overrides whether this node participates in delete requests.
func (n Node) Deletable(enabled bool) Node {
	n.deletable = &enabled
	return n
}

func (n Node) isDraggable(defaultValue bool) bool {
	if n.draggable != nil {
		return *n.draggable
	}
	return defaultValue
}

func (n Node) isResizable(defaultValue bool) bool {
	if n.resizable != nil {
		return *n.resizable
	}
	return defaultValue
}

func (n Node) isSelectable(defaultValue bool) bool {
	if n.selectable != nil {
		return *n.selectable
	}
	return defaultValue
}

func (n Node) isConnectable(defaultValue bool) bool {
	if n.connectable != nil {
		return *n.connectable
	}
	return defaultValue
}

func (n Node) isDeletable(defaultValue bool) bool {
	if n.deletable != nil {
		return *n.deletable
	}
	return defaultValue
}

// Endpoint identifies one port of one node.
type Endpoint struct {
	NodeID string
	PortID string
}

// EdgeType controls the geometry used to render and hit-test an edge.
type EdgeType uint8

const (
	EdgeBezier EdgeType = iota
	EdgeStraight
	EdgeStep
	EdgeSmoothStep
)

// EdgeMarker identifies an optional arrow marker at an edge endpoint.
type EdgeMarker uint8

const (
	MarkerNone EdgeMarker = iota
	MarkerArrow
)

// NewEndpoint creates a graph edge endpoint.
func NewEndpoint(nodeID, portID string) Endpoint {
	if nodeID == "" || portID == "" {
		panic("flowui: node graph endpoint requires node and port IDs")
	}
	return Endpoint{NodeID: nodeID, PortID: portID}
}

// Edge connects an output endpoint to an input endpoint.
type Edge struct {
	ID           string
	Source       Endpoint
	Target       Endpoint
	Color        color.NRGBA
	Selected     bool
	label        string
	selectable   *bool
	deletable    *bool
	reconnect    *ReconnectMode
	edgeType     *EdgeType
	width        float32
	dashed       *bool
	animated     *bool
	sourceMark   EdgeMarker
	targetMark   EdgeMarker
	labelContent frame.Widget
}

// NewEdge creates a directed edge from source to target.
func NewEdge(id string, source, target Endpoint) Edge {
	if id == "" {
		panic("flowui: empty node graph edge ID")
	}
	return Edge{ID: id, Source: source, Target: target}
}

// Select sets the edge's controlled selected state.
func (e Edge) Select(selected bool) Edge {
	e.Selected = selected
	return e
}

// Type selects the edge geometry. The default is Bezier.
func (e Edge) Type(edgeType EdgeType) Edge {
	if edgeType > EdgeSmoothStep {
		panic("flowui: invalid node graph edge type")
	}
	e.edgeType = &edgeType
	return e
}

func (e Edge) edgeTypeValue() EdgeType {
	if e.edgeType != nil {
		return *e.edgeType
	}
	return EdgeBezier
}

// Width sets the edge stroke width in dp. Zero uses the theme default.
func (e Edge) Width(width float32) Edge {
	if width < 0 || !finite(width) {
		panic("flowui: invalid node graph edge width")
	}
	e.width = width
	return e
}

// Dashed toggles a dashed stroke.
func (e Edge) Dashed(enabled bool) Edge {
	e.dashed = &enabled
	return e
}

func (e Edge) isDashed() bool {
	return e.dashed != nil && *e.dashed
}

// Animated toggles a moving dashed stroke. It implies Dashed when enabled.
func (e Edge) Animated(enabled bool) Edge {
	e.animated = &enabled
	return e
}

func (e Edge) isAnimated() bool {
	return e.animated != nil && *e.animated
}

// Markers configures optional arrow markers at the source and target ends.
func (e Edge) Markers(source, target EdgeMarker) Edge {
	if source > MarkerArrow || target > MarkerArrow {
		panic("flowui: invalid node graph edge marker")
	}
	e.sourceMark, e.targetMark = source, target
	return e
}

// Label sets the optional label rendered at the edge midpoint.
func (e Edge) Label(label string) Edge {
	e.label = label
	return e
}

// LabelContent replaces the plain edge label with an interactive widget at the
// edge midpoint. It receives a state scope unique to this graph and edge ID.
func (e Edge) LabelContent(content frame.Widget) Edge {
	e.labelContent = content
	return e
}

// Selectable overrides whether this edge may be selected.
func (e Edge) Selectable(enabled bool) Edge {
	e.selectable = &enabled
	return e
}

// Deletable overrides whether this edge participates in delete requests.
func (e Edge) Deletable(enabled bool) Edge {
	e.deletable = &enabled
	return e
}

// Reconnectable overrides which endpoints of this edge may be reconnected.
func (e Edge) Reconnectable(mode ReconnectMode) Edge {
	if mode > ReconnectNone {
		panic("flowui: invalid node graph reconnect mode")
	}
	e.reconnect = &mode
	return e
}

func (e Edge) isSelectable(defaultValue bool) bool {
	if e.selectable != nil {
		return *e.selectable
	}
	return defaultValue
}

func (e Edge) isDeletable(defaultValue bool) bool {
	if e.deletable != nil {
		return *e.deletable
	}
	return defaultValue
}

func (e Edge) isReconnectable(endpoint reconnectEndpoint, defaultValue bool) bool {
	mode := ReconnectBoth
	if e.reconnect != nil {
		mode = *e.reconnect
	} else if !defaultValue {
		mode = ReconnectNone
	}
	return mode == ReconnectBoth || (mode == ReconnectSource && endpoint == reconnectSource) || (mode == ReconnectTarget && endpoint == reconnectTarget)
}

// Graph is the application-owned, declarative node graph model.
type Graph struct {
	Nodes []Node
	Edges []Edge
}

// Connection is a requested edge between one output and one input port.
type Connection struct {
	Source Endpoint
	Target Endpoint
}

// ReconnectMode controls which endpoints of an edge may be dragged to a new
// port. The zero value permits both endpoints.
type ReconnectMode uint8

const (
	ReconnectBoth ReconnectMode = iota
	ReconnectSource
	ReconnectTarget
	ReconnectNone
)

// SelectionMode controls how many nodes can be selected by pointer gestures.
type SelectionMode uint8

const (
	SelectionSingle SelectionMode = iota
	SelectionMultiple
	SelectionNone
)

// SelectionBoxMode controls how a dragged selection box matches nodes.
type SelectionBoxMode uint8

const (
	SelectionBoxPartial SelectionBoxMode = iota
	SelectionBoxFull
)

// GridPattern controls the visual form used when the graph grid is enabled.
type GridPattern uint8

const (
	GridLines GridPattern = iota
	GridDots
)

// NodeChangeKind identifies a requested node data change.
type NodeChangeKind uint8

const (
	NodeChangePosition NodeChangeKind = iota
	NodeChangeSelection
	NodeChangeRemove
	NodeChangeSize
)

// NodeChange requests an application-owned node update. Position changes are
// emitted throughout a drag and once more with Dragging set to false on release.
type NodeChange struct {
	ID       string
	Kind     NodeChangeKind
	Position Point
	Size     Size
	Selected bool
	Dragging bool
	Resizing bool
}

// EdgeChangeKind identifies a requested edge update.
type EdgeChangeKind uint8

const (
	EdgeChangeSelection EdgeChangeKind = iota
	EdgeChangeRemove
)

// EdgeChange requests an application-owned edge selection or removal.
type EdgeChange struct {
	ID       string
	Kind     EdgeChangeKind
	Selected bool
}

// ApplyNodeChanges applies position and selection requests to a copy of nodes.
// It is optional; applications may instead apply changes in their own model code.
func ApplyNodeChanges(nodes []Node, changes []NodeChange) []Node {
	updated := append([]Node(nil), nodes...)
	indexByID := make(map[string]int, len(updated))
	for index, node := range updated {
		indexByID[node.ID] = index
	}
	for _, change := range changes {
		index, ok := indexByID[change.ID]
		if !ok {
			continue
		}
		switch change.Kind {
		case NodeChangePosition:
			updated[index].Position = change.Position
		case NodeChangeSelection:
			updated[index].Selected = change.Selected
		case NodeChangeRemove:
			updated[index].ID = ""
		case NodeChangeSize:
			updated[index].Size = change.Size
		}
	}
	result := updated[:0]
	for _, node := range updated {
		if node.ID != "" {
			result = append(result, node)
		}
	}
	return result
}

// ApplyEdgeChanges applies selection and removal requests to a copy of edges.
// It is optional; applications may instead apply changes in their own model code.
func ApplyEdgeChanges(edges []Edge, changes []EdgeChange) []Edge {
	updated := append([]Edge(nil), edges...)
	indexByID := make(map[string]int, len(updated))
	for index, edge := range updated {
		indexByID[edge.ID] = index
	}
	for _, change := range changes {
		index, ok := indexByID[change.ID]
		if !ok {
			continue
		}
		switch change.Kind {
		case EdgeChangeSelection:
			updated[index].Selected = change.Selected
		case EdgeChangeRemove:
			updated[index].ID = ""
		}
	}
	result := updated[:0]
	for _, edge := range updated {
		if edge.ID != "" {
			result = append(result, edge)
		}
	}
	return result
}

// ReconnectEdge applies connection endpoints to a copy of the matching edge.
// It preserves the edge ID because FlowUI graph identities are application-owned.
func ReconnectEdge(old Edge, connection Connection, edges []Edge) []Edge {
	updated := append([]Edge(nil), edges...)
	if connection.Source.NodeID == "" || connection.Source.PortID == "" || connection.Target.NodeID == "" || connection.Target.PortID == "" {
		return updated
	}
	for index, edge := range updated {
		if edge.ID == old.ID {
			updated[index].Source = connection.Source
			updated[index].Target = connection.Target
			break
		}
	}
	return updated
}

// Widget lays out an interactive node graph. Nodes and edges are controlled by
// the caller; viewport, selection, and position changes are reported as callbacks.
type Widget struct {
	key                 string
	graph               Graph
	viewport            Viewport
	hasViewport         bool
	defaultViewport     Viewport
	hasDefault          bool
	fitView             bool
	fitViewPadding      float32
	onViewportChange    func(Viewport)
	height              unit.Dp
	showGrid            bool
	gridPattern         GridPattern
	gridSize            float32
	gridColor           color.NRGBA
	hasGridColor        bool
	gridOpacity         float32
	hasGridOpacity      bool
	minimapEnabled      bool
	controlsEnabled     bool
	cullingEnabled      bool
	zoomOnScroll        bool
	panOnScroll         bool
	panOnPrimary        bool
	panOnMiddle         bool
	minZoom             float32
	maxZoom             float32
	selectedKeys        []string
	hasSelectedKeys     bool
	selectedEdgeKeys    []string
	hasSelectedEdges    bool
	selectionMode       SelectionMode
	selectionOnDrag     bool
	selectionBoxMode    SelectionBoxMode
	nodesDraggable      bool
	nodesResizable      bool
	resizeHandles       ResizeHandle
	nodesSelectable     bool
	nodesConnectable    bool
	nodesDeletable      bool
	edgesSelectable     bool
	edgesDeletable      bool
	edgesReconnectable  bool
	dragThreshold       float32
	snapToGrid          bool
	snapGrid            Point
	onNodesChange       func([]NodeChange)
	onEdgesChange       func([]EdgeChange)
	onReconnect         func(Edge, Connection)
	isValidConnection   func(Connection) bool
	onConnect           func(Connection)
	onCopy              func(Fragment)
	onCut               func(Fragment)
	onPaste             func(Fragment, Point)
	onUndo              func()
	onRedo              func()
	onCanvasClick       func(CanvasEvent)
	onCanvasDoubleClick func(CanvasEvent)
	onCanvasContextMenu func(CanvasEvent)
	onNodeClick         func(NodeEvent)
	onNodeDoubleClick   func(NodeEvent)
	onNodeContextMenu   func(NodeEvent)
	onNodeHover         func(NodeEvent)
	onNodeLeave         func(NodeEvent)
	onEdgeClick         func(EdgeEvent)
	onEdgeDoubleClick   func(EdgeEvent)
	onEdgeContextMenu   func(EdgeEvent)
	onEdgeHover         func(EdgeEvent)
	onEdgeLeave         func(EdgeEvent)
	dropTypes           []string
	onDrop              func(DropEvent)
	panels              []Panel
	viewportOverlays    []ViewportOverlay
	nodeToolbars        []nodeToolbar
	edgeToolbars        []edgeToolbar
	disabled            bool
}

// New creates a node graph widget for graph.
func New(key string, graph Graph) Widget {
	if key == "" {
		panic("flowui: empty node graph key")
	}
	return Widget{
		key:                key,
		graph:              graph,
		showGrid:           true,
		cullingEnabled:     true,
		zoomOnScroll:       true,
		panOnPrimary:       true,
		panOnMiddle:        true,
		gridSize:           defaultGridSize,
		minZoom:            defaultMinZoom,
		maxZoom:            defaultMaxZoom,
		fitViewPadding:     defaultFitViewPadding,
		selectionMode:      SelectionSingle,
		nodesDraggable:     true,
		resizeHandles:      ResizeHandleBottomRight,
		nodesSelectable:    true,
		nodesConnectable:   true,
		nodesDeletable:     true,
		edgesSelectable:    true,
		edgesDeletable:     true,
		edgesReconnectable: true,
		dragThreshold:      defaultDragThreshold,
		snapGrid:           Point{X: defaultGridSize, Y: defaultGridSize},
	}
}

// Viewport makes the viewport controlled. Its Origin is the graph coordinate
// displayed at the canvas top-left.
func (w Widget) Viewport(value Viewport) Widget {
	w.viewport = value
	w.hasViewport = true
	return w
}

// DefaultViewport sets the initial internal viewport when Viewport is not
// supplied. It is read once per widget state lifetime.
func (w Widget) DefaultViewport(value Viewport) Widget {
	w.defaultViewport = value
	w.hasDefault = true
	return w
}

// FitView requests an initial viewport that contains every node. It takes
// precedence over DefaultViewport and reports the fitted viewport through
// OnViewportChange when the viewport is controlled. Toggle it from false to
// true to request another fit.
func (w Widget) FitView(enabled bool) Widget {
	w.fitView = enabled
	return w
}

// FitViewPadding sets the fraction of each canvas edge reserved around the
// fitted nodes. The default is 0.1, or ten percent on every edge.
func (w Widget) FitViewPadding(padding float32) Widget {
	if padding < 0 || padding >= .5 || !finite(padding) {
		panic("flowui: node graph fit view padding must be between 0 and 0.5")
	}
	w.fitViewPadding = padding
	return w
}

// OnViewportChange receives panning and zooming requests. Supplying this
// callback normally goes with a controlled Viewport in the application model.
func (w Widget) OnViewportChange(fn func(Viewport)) Widget {
	w.onViewportChange = fn
	return w
}

// OnCanvasClick receives an unmodified primary click on empty canvas space.
func (w Widget) OnCanvasClick(fn func(CanvasEvent)) Widget {
	w.onCanvasClick = fn
	return w
}

// OnCanvasDoubleClick receives a double click on empty canvas space.
func (w Widget) OnCanvasDoubleClick(fn func(CanvasEvent)) Widget {
	w.onCanvasDoubleClick = fn
	return w
}

// OnCanvasContextMenu receives a secondary-button press on empty canvas space.
func (w Widget) OnCanvasContextMenu(fn func(CanvasEvent)) Widget {
	w.onCanvasContextMenu = fn
	return w
}

// OnNodeClick receives an unmodified primary click on a node.
func (w Widget) OnNodeClick(fn func(NodeEvent)) Widget {
	w.onNodeClick = fn
	return w
}

// OnNodeDoubleClick receives a double click on a node.
func (w Widget) OnNodeDoubleClick(fn func(NodeEvent)) Widget {
	w.onNodeDoubleClick = fn
	return w
}

// OnNodeContextMenu receives a secondary-button press on a node.
func (w Widget) OnNodeContextMenu(fn func(NodeEvent)) Widget {
	w.onNodeContextMenu = fn
	return w
}

// OnNodeHover receives a node when the pointer enters it.
func (w Widget) OnNodeHover(fn func(NodeEvent)) Widget {
	w.onNodeHover = fn
	return w
}

// OnNodeLeave receives a node when the pointer leaves it.
func (w Widget) OnNodeLeave(fn func(NodeEvent)) Widget {
	w.onNodeLeave = fn
	return w
}

// OnEdgeClick receives an unmodified primary click on an edge.
func (w Widget) OnEdgeClick(fn func(EdgeEvent)) Widget {
	w.onEdgeClick = fn
	return w
}

// OnEdgeDoubleClick receives a double click on an edge.
func (w Widget) OnEdgeDoubleClick(fn func(EdgeEvent)) Widget {
	w.onEdgeDoubleClick = fn
	return w
}

// OnEdgeContextMenu receives a secondary-button press on an edge.
func (w Widget) OnEdgeContextMenu(fn func(EdgeEvent)) Widget {
	w.onEdgeContextMenu = fn
	return w
}

// OnEdgeHover receives an edge when the pointer enters it.
func (w Widget) OnEdgeHover(fn func(EdgeEvent)) Widget {
	w.onEdgeHover = fn
	return w
}

// OnEdgeLeave receives an edge when the pointer leaves it.
func (w Widget) OnEdgeLeave(fn func(EdgeEvent)) Widget {
	w.onEdgeLeave = fn
	return w
}

// DropTypes declares MIME types accepted from external drag sources.
func (w Widget) DropTypes(types ...string) Widget {
	if len(types) == 0 {
		panic("flowui: node graph drop types must not be empty")
	}
	unique := make([]string, 0, len(types))
	for _, value := range types {
		value = strings.TrimSpace(value)
		if value == "" {
			panic("flowui: node graph drop type must not be empty")
		}
		alreadyAdded := slices.Contains(unique, value)
		if !alreadyAdded {
			unique = append(unique, value)
		}
	}
	w.dropTypes = unique
	return w
}

// OnDrop receives data dropped on the graph for a declared MIME type.
func (w Widget) OnDrop(fn func(DropEvent)) Widget {
	w.onDrop = fn
	return w
}

// Height sets the graph height in dp. Without it, the graph consumes its
// layout constraints' maximum height.
func (w Widget) Height(dp int) Widget {
	if dp <= 0 {
		panic("flowui: node graph height must be positive")
	}
	w.height = unit.Dp(dp)
	return w
}

// Grid controls the background graph grid.
func (w Widget) Grid(enabled bool) Widget {
	w.showGrid = enabled
	return w
}

// Minimap toggles the compact graph overview in the canvas corner.
func (w Widget) Minimap(enabled bool) Widget {
	w.minimapEnabled = enabled
	return w
}

// Controls toggles zoom and fit-view controls in the canvas corner.
func (w Widget) Controls(enabled bool) Widget {
	w.controlsEnabled = enabled
	return w
}

// Panel adds fixed content above the graph canvas. A Panel is clipped to the
// canvas and is painted after nodes and edges, so it may contain controls or
// tool surfaces. Keys must be non-empty and unique within this graph.
func (w Widget) Panel(key string, position PanelPosition, content frame.Widget) Widget {
	if key == "" || content == nil || position > PanelBottomRight {
		panic("flowui: invalid node graph panel")
	}
	for _, panel := range w.panels {
		if panel.Key == key {
			panic("flowui: duplicate node graph panel key " + key)
		}
	}
	w.panels = append(w.panels, Panel{Key: key, Position: position, Content: content})
	return w
}

// ViewportOverlay adds content at a graph coordinate. It uses the current
// zoom for layout density, allowing annotations and controls to stay attached
// to the graph rather than the window.
func (w Widget) ViewportOverlay(key string, position Point, content frame.Widget) Widget {
	if key == "" || content == nil || !finite(position.X) || !finite(position.Y) {
		panic("flowui: invalid node graph viewport overlay")
	}
	for _, overlay := range w.viewportOverlays {
		if overlay.Key == key {
			panic("flowui: duplicate node graph viewport overlay key " + key)
		}
	}
	w.viewportOverlays = append(w.viewportOverlays, ViewportOverlay{Key: key, Position: position, Content: content})
	return w
}

// NodeToolbar adds content above nodeID in graph coordinates. The application
// controls visibility by including or omitting the toolbar from the widget.
func (w Widget) NodeToolbar(key, nodeID string, content frame.Widget) Widget {
	if key == "" || nodeID == "" || content == nil {
		panic("flowui: invalid node graph node toolbar")
	}
	w.nodeToolbars = append(w.nodeToolbars, nodeToolbar{key: key, nodeID: nodeID, content: content})
	return w
}

// EdgeToolbar adds content at the midpoint of edgeID in graph coordinates.
func (w Widget) EdgeToolbar(key, edgeID string, content frame.Widget) Widget {
	if key == "" || edgeID == "" || content == nil {
		panic("flowui: invalid node graph edge toolbar")
	}
	w.edgeToolbars = append(w.edgeToolbars, edgeToolbar{key: key, edgeID: edgeID, content: content})
	return w
}

// Culling skips node and edge drawing work outside the canvas viewport. It is
// enabled by default and should normally remain enabled for large graphs.
func (w Widget) Culling(enabled bool) Widget {
	w.cullingEnabled = enabled
	return w
}

// ZoomOnScroll controls wheel zooming around the pointer. It defaults to true.
func (w Widget) ZoomOnScroll(enabled bool) Widget {
	w.zoomOnScroll = enabled
	return w
}

// PanOnScroll enables wheel panning when zoom-on-scroll is disabled. It
// defaults to false so existing wheel behavior remains zooming.
func (w Widget) PanOnScroll(enabled bool) Widget {
	w.panOnScroll = enabled
	return w
}

// PanOnPrimaryButton controls panning from empty canvas space with the primary
// pointer button when selection-box dragging is not active. It defaults to true.
func (w Widget) PanOnPrimaryButton(enabled bool) Widget {
	w.panOnPrimary = enabled
	return w
}

// PanOnMiddleButton controls middle-button panning. It defaults to true.
func (w Widget) PanOnMiddleButton(enabled bool) Widget {
	w.panOnMiddle = enabled
	return w
}

// GridPattern selects line or dot rendering for an enabled graph grid.
func (w Widget) GridPattern(pattern GridPattern) Widget {
	if pattern > GridDots {
		panic("flowui: invalid node graph grid pattern")
	}
	w.gridPattern = pattern
	return w
}

// GridColor overrides the theme color used for lines or dots.
func (w Widget) GridColor(value color.NRGBA) Widget {
	w.gridColor = value
	w.hasGridColor = true
	return w
}

// GridOpacity overrides the theme opacity used for lines or dots.
func (w Widget) GridOpacity(value float32) Widget {
	if value < 0 || value > 1 || !finite(value) {
		panic("flowui: node graph grid opacity must be between 0 and 1")
	}
	w.gridOpacity = value
	w.hasGridOpacity = true
	return w
}

// GridSize sets the grid interval in graph coordinates.
func (w Widget) GridSize(value float32) Widget {
	if value <= 0 || !finite(value) {
		panic("flowui: node graph grid size must be positive")
	}
	w.gridSize = value
	return w
}

// ZoomRange bounds scroll zooming. Minimum must be positive and no greater
// than maximum.
func (w Widget) ZoomRange(minimum, maximum float32) Widget {
	if minimum <= 0 || maximum < minimum || !finite(minimum) || !finite(maximum) {
		panic("flowui: invalid node graph zoom range")
	}
	w.minZoom, w.maxZoom = minimum, maximum
	return w
}

// SelectedKeys supplies a controlled selection independent of Node.Selected.
// When omitted, each node's Selected field controls the rendered selection.
func (w Widget) SelectedKeys(keys []string) Widget {
	w.selectedKeys = append([]string(nil), keys...)
	w.hasSelectedKeys = true
	return w
}

// SelectedEdgeKeys supplies a controlled edge selection independent of
// Edge.Selected. When omitted, each edge's Selected field controls selection.
func (w Widget) SelectedEdgeKeys(keys []string) Widget {
	w.selectedEdgeKeys = append([]string(nil), keys...)
	w.hasSelectedEdges = true
	return w
}

// SelectionMode controls pointer selection behavior.
func (w Widget) SelectionMode(mode SelectionMode) Widget {
	if mode > SelectionNone {
		panic("flowui: invalid node graph selection mode")
	}
	w.selectionMode = mode
	return w
}

// SelectionOnDrag makes a primary-button drag on empty space draw a node
// selection box. Middle-button dragging remains available for viewport panning.
func (w Widget) SelectionOnDrag(enabled bool) Widget {
	w.selectionOnDrag = enabled
	return w
}

// SelectionBoxMode chooses whether a node must be fully contained by the
// selection box or may merely intersect it.
func (w Widget) SelectionBoxMode(mode SelectionBoxMode) Widget {
	if mode > SelectionBoxFull {
		panic("flowui: invalid node graph selection box mode")
	}
	w.selectionBoxMode = mode
	return w
}

// NodesDraggable enables or disables dragging for all nodes that do not have
// an explicit Node.Draggable override.
func (w Widget) NodesDraggable(enabled bool) Widget {
	w.nodesDraggable = enabled
	return w
}

// NodesResizable enables selected-node resize handles unless a node has an
// explicit Resizable override. It defaults to false.
func (w Widget) NodesResizable(enabled bool) Widget {
	w.nodesResizable = enabled
	return w
}

// ResizeHandles sets the default handles for resizable nodes. Use
// ResizeHandleAll for four edges and four corners; a node may override it with
// Node.ResizeHandles.
func (w Widget) ResizeHandles(handles ResizeHandle) Widget {
	if handles&^ResizeHandleAll != 0 {
		panic("flowui: invalid node graph resize handles")
	}
	w.resizeHandles = handles
	return w
}

// NodesSelectable enables or disables selection for all nodes that do not have
// an explicit Node.Selectable override.
func (w Widget) NodesSelectable(enabled bool) Widget {
	w.nodesSelectable = enabled
	return w
}

// NodesConnectable enables or disables connection gestures for nodes and ports
// without an explicit Connectable override.
func (w Widget) NodesConnectable(enabled bool) Widget {
	w.nodesConnectable = enabled
	return w
}

// NodesDeletable enables or disables deletion for nodes without a Deletable
// override. Deleting a node also removes its connected edges.
func (w Widget) NodesDeletable(enabled bool) Widget {
	w.nodesDeletable = enabled
	return w
}

// EdgesSelectable enables or disables selection for edges without an explicit
// Edge.Selectable override.
func (w Widget) EdgesSelectable(enabled bool) Widget {
	w.edgesSelectable = enabled
	return w
}

// EdgesDeletable enables or disables deletion for edges without a Deletable
// override.
func (w Widget) EdgesDeletable(enabled bool) Widget {
	w.edgesDeletable = enabled
	return w
}

// EdgesReconnectable enables or disables reconnection for edges without an
// explicit Edge.Reconnectable override.
func (w Widget) EdgesReconnectable(enabled bool) Widget {
	w.edgesReconnectable = enabled
	return w
}

// NodeDragThreshold sets the pointer movement in pixels required to start a
// node drag. Zero starts dragging immediately.
func (w Widget) NodeDragThreshold(pixels float32) Widget {
	if pixels < 0 || !finite(pixels) {
		panic("flowui: node graph drag threshold must not be negative")
	}
	w.dragThreshold = pixels
	return w
}

// SnapToGrid enables snapping node drag deltas to SnapGrid.
func (w Widget) SnapToGrid(enabled bool) Widget {
	w.snapToGrid = enabled
	return w
}

// SnapGrid sets the horizontal and vertical grid intervals used for node drag
// snapping. It does not change the drawn background grid.
func (w Widget) SnapGrid(x, y float32) Widget {
	if x <= 0 || y <= 0 || !finite(x) || !finite(y) {
		panic("flowui: node graph snap grid must be positive")
	}
	w.snapGrid = Point{X: x, Y: y}
	return w
}

// OnNodesChange receives controlled node selection, position, and size requests.
func (w Widget) OnNodesChange(fn func([]NodeChange)) Widget {
	w.onNodesChange = fn
	return w
}

// OnEdgesChange receives controlled edge selection and removal requests.
func (w Widget) OnEdgesChange(fn func([]EdgeChange)) Widget {
	w.onEdgesChange = fn
	return w
}

// OnReconnect receives a validated request to replace one edge's endpoints.
// Use ReconnectEdge to apply the request to application-owned edges.
func (w Widget) OnReconnect(fn func(Edge, Connection)) Widget {
	w.onReconnect = fn
	return w
}

// IsValidConnection validates a candidate connection before it is highlighted
// or submitted. It must not mutate application state.
func (w Widget) IsValidConnection(fn func(Connection) bool) Widget {
	w.isValidConnection = fn
	return w
}

// OnConnect receives a validated output-to-input connection request.
func (w Widget) OnConnect(fn func(Connection)) Widget {
	w.onConnect = fn
	return w
}

// OnCopy receives the current node selection when Ctrl/Cmd+C is pressed while
// the graph canvas owns keyboard focus.
func (w Widget) OnCopy(fn func(Fragment)) Widget {
	w.onCopy = fn
	return w
}

// OnCut receives the copied selection before NodeGraph requests its deletion
// when Ctrl/Cmd+X is pressed.
func (w Widget) OnCut(fn func(Fragment)) Widget {
	w.onCut = fn
	return w
}

// OnPaste receives the most recently copied in-process Fragment and a
// suggested incremental offset when Ctrl/Cmd+V is pressed. Use PasteFragment
// with application-owned ID allocators to create the new model data.
func (w Widget) OnPaste(fn func(Fragment, Point)) Widget {
	w.onPaste = fn
	return w
}

// OnUndo receives Ctrl/Cmd+Z command requests. Graph data remains owned by
// the application; History is available for a snapshot-based implementation.
func (w Widget) OnUndo(fn func()) Widget {
	w.onUndo = fn
	return w
}

// OnRedo receives Ctrl/Cmd+Y and Ctrl/Cmd+Shift+Z command requests.
func (w Widget) OnRedo(fn func()) Widget {
	w.onRedo = fn
	return w
}

// Disabled prevents viewport gestures while preserving rendering.
func (w Widget) Disabled(disabled bool) Widget {
	w.disabled = disabled
	return w
}
