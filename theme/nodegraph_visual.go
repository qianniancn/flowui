package theme

import (
	"image"
	"image/color"

	"gioui.org/f32"
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/style"
)

// NodeGraphRequest contains the instance-level visual declaration for a node
// graph. Graph data, viewport state, selection, and pointer behavior remain
// widget-owned.
type NodeGraphRequest struct {
	Override *NodeGraphOverride
}

// NodeGraphOverride carries caller customization across the neutral provider
// boundary. Grid overrides remain declarations owned by the provider; the
// widget only transports the values and never interprets their visual effect.
type NodeGraphOverride struct {
	Declaration style.Style
	GridColor   *color.NRGBA
	GridOpacity *float32
}

// NodeGraphOverrideStyle extracts the provider-owned instance layer.
func NodeGraphOverrideStyle(request NodeGraphRequest) style.Style {
	if request.Override == nil {
		return style.Style{}
	}
	return request.Override.Declaration
}

// NodeGraphGridPattern identifies the provider-neutral grid primitive.
type NodeGraphGridPattern uint8

const (
	NodeGraphGridLines NodeGraphGridPattern = iota
	NodeGraphGridDots
)

// NodeGraphEdgeType identifies the route geometry selected by graph behavior.
type NodeGraphEdgeType uint8

const (
	NodeGraphEdgeBezier NodeGraphEdgeType = iota
	NodeGraphEdgeStraight
	NodeGraphEdgeStep
	NodeGraphEdgeSmoothStep
)

// NodeGraphEdgeMarker identifies an optional endpoint marker.
type NodeGraphEdgeMarker uint8

const (
	NodeGraphMarkerNone NodeGraphEdgeMarker = iota
	NodeGraphMarkerArrow
)

// NodeGraphMetrics contains geometry and typography consumed by graph layout
// and provider renderers. It contains no semantic or Gio state.
type NodeGraphMetrics struct {
	CanvasRadius            unit.Dp
	CanvasBorderWidth       unit.Dp
	GridWidth               unit.Dp
	GridMinimumStep         unit.Dp
	EdgeWidth               unit.Dp
	EdgeMinimumDistance     unit.Dp
	EdgeLabelWidth          unit.Dp
	EdgeLabelHeight         unit.Dp
	NodeHeaderHeight        unit.Dp
	NodeVerticalPadding     unit.Dp
	NodePortRowHeight       unit.Dp
	NodeDefaultWidth        unit.Dp
	NodeDefaultHeight       unit.Dp
	NodePadding             unit.Dp
	NodeRadius              unit.Dp
	NodeBorderWidth         unit.Dp
	NodeTitleSize           unit.Sp
	NodePortTextSize        unit.Sp
	NodeTextMinScale        float32
	NodeTextMaxScale        float32
	PortRadius              unit.Dp
	PortBorderWidth         unit.Dp
	EdgeMarkerSize          unit.Dp
	EdgeSmoothRadius        unit.Dp
	SelectionWidth          unit.Dp
	SelectionDashLength     unit.Dp
	SelectionDashGap        unit.Dp
	FocusInset              unit.Dp
	FocusWidth              unit.Dp
	FocusDashLength         unit.Dp
	FocusDashGap            unit.Dp
	ResizeHandleRadius      unit.Dp
	ResizeHandleBorderWidth unit.Dp
	ControlWidth            unit.Dp
	ControlHeight           unit.Dp
	ControlMargin           unit.Dp
	ControlRadius           unit.Dp
	ControlDividerInset     unit.Dp
	ControlTextSize         unit.Sp
	ControlFitSpan          float32
	ControlFitArm           float32
	PanelPadding            unit.Dp
	ToolbarOffset           unit.Dp
	MinimapWidth            unit.Dp
	MinimapHeight           unit.Dp
	MinimapMargin           unit.Dp
	MinimapPadding          unit.Dp
	MinimapRadius           unit.Dp
	MinimapBorderWidth      unit.Dp
	MinimapEdgeWidth        unit.Dp
	MinimapViewportWidth    unit.Dp
	MinimapNodeMinSize      unit.Dp
	ShadowOffsetY           unit.Dp
	ShadowBlur              unit.Dp
	ShadowSpread            unit.Dp
	ShadowOpacity           float32
}

// NodeGraphColors contains all palette decisions used by graph behavior and
// renderers. Explicit port and edge colors remain widget inputs.
type NodeGraphColors struct {
	CanvasBackground    color.NRGBA
	CanvasBorder        color.NRGBA
	Grid                color.NRGBA
	NodeBackground      color.NRGBA
	NodeBorder          color.NRGBA
	NodeForeground      color.NRGBA
	NodeMutedForeground color.NRGBA
	PortColor           color.NRGBA
	PortBorder          color.NRGBA
	EdgeColor           color.NRGBA
	SelectedEdgeColor   color.NRGBA
	SelectedNodeBorder  color.NRGBA
	SelectionFill       color.NRGBA
	SelectionBorder     color.NRGBA
	ControlBackground   color.NRGBA
	ControlForeground   color.NRGBA
	ControlBorder       color.NRGBA
	Shadow              color.NRGBA
	ResizeHandleFill    color.NRGBA
}

// NodeGraphCanvasRequest describes the graph canvas primitive.
type NodeGraphCanvasRequest struct {
	Size        image.Point
	Radius      int
	BorderWidth int
	Background  color.NRGBA
	Border      color.NRGBA
}

// NodeGraphGridRequest describes one visible graph grid.
type NodeGraphGridRequest struct {
	Size    image.Point
	Pattern NodeGraphGridPattern
	StartX  float32
	StartY  float32
	Step    float32
	Width   float32
	Color   color.NRGBA
}

// NodeGraphEdgeRequest describes one edge route and optional label/markers.
type NodeGraphEdgeRequest struct {
	EdgeType        NodeGraphEdgeType
	From            f32.Point
	To              f32.Point
	Points          []f32.Point
	Width           float32
	MinimumDistance float32
	Dashed          bool
	Phase           float32
	Color           color.NRGBA
	SourceMarker    NodeGraphEdgeMarker
	TargetMarker    NodeGraphEdgeMarker
	MarkerSize      float32
	Label           string
	LabelRect       image.Rectangle
	LabelSize       unit.Sp
	LabelColor      color.NRGBA
	LabelWeight     font.Weight
}

// NodeGraphConnectionRequest describes the temporary edge during a connect
// gesture.
type NodeGraphConnectionRequest struct {
	From            f32.Point
	To              f32.Point
	Width           float32
	MinimumDistance float32
	Color           color.NRGBA
}

// NodeGraphNodeRequest describes one node shell and its default title.
type NodeGraphNodeRequest struct {
	Rect         image.Rectangle
	Radius       int
	BorderWidth  int
	Padding      int
	HeaderHeight int
	Background   color.NRGBA
	Border       color.NRGBA
	Shadow       color.NRGBA
	DrawShadow   bool
	Title        string
	TitleRect    image.Rectangle
	TitleSize    unit.Sp
	TitleColor   color.NRGBA
	TitleWeight  font.Weight
}

// NodeGraphPortRequest describes one port dot and its optional label.
type NodeGraphPortRequest struct {
	Center      f32.Point
	Scale       float32
	Radius      int
	BorderWidth int
	Color       color.NRGBA
	Border      color.NRGBA
	Label       string
	LabelRect   image.Rectangle
	LabelSize   unit.Sp
	LabelColor  color.NRGBA
	LabelWeight font.Weight
}

// NodeGraphSelectionRequest describes a selection or focus outline.
type NodeGraphSelectionRequest struct {
	Rect       image.Rectangle
	Fill       color.NRGBA
	Border     color.NRGBA
	Width      int
	DashLength int
	DashGap    int
	Filled     bool
}

// NodeGraphResizeRequest describes the visible resize handles.
type NodeGraphResizeRequest struct {
	Rect        image.Rectangle
	Scale       float32
	Handles     uint16
	Radius      int
	BorderWidth int
	Border      color.NRGBA
	Fill        color.NRGBA
}

// NodeGraphControlsRequest describes the fixed zoom/fit control strip.
type NodeGraphControlsRequest struct {
	Rect         image.Rectangle
	ButtonWidth  int
	Radius       int
	DividerInset int
	TextSize     unit.Sp
	FitSpan      float32
	FitArm       float32
	Background   color.NRGBA
	Foreground   color.NRGBA
	Border       color.NRGBA
}

// NodeGraphMinimapNode is one node rectangle in minimap coordinates.
type NodeGraphMinimapNode struct {
	Rect     image.Rectangle
	Selected bool
}

// NodeGraphMinimapEdge is one edge segment in minimap coordinates.
type NodeGraphMinimapEdge struct {
	Points []f32.Point
	Color  color.NRGBA
}

// NodeGraphMinimapRequest describes the complete minimap primitive.
type NodeGraphMinimapRequest struct {
	Rect          image.Rectangle
	Inner         image.Rectangle
	Nodes         []NodeGraphMinimapNode
	Edges         []NodeGraphMinimapEdge
	Viewport      image.Rectangle
	Radius        int
	BorderWidth   int
	EdgeWidth     float32
	ViewportWidth float32
	Background    color.NRGBA
	Border        color.NRGBA
	Selected      color.NRGBA
}

// NodeGraphCanvasRenderer draws the graph canvas.
type NodeGraphCanvasRenderer func(layout.Context, NodeGraphCanvasRequest)

// NodeGraphGridRenderer draws the graph grid.
type NodeGraphGridRenderer func(layout.Context, NodeGraphGridRequest)

// NodeGraphEdgeRenderer draws one edge, including its markers and default
// text label.
type NodeGraphEdgeRenderer func(layout.Context, NodeGraphEdgeRequest)

// NodeGraphConnectionRenderer draws a temporary connection preview.
type NodeGraphConnectionRenderer func(layout.Context, NodeGraphConnectionRequest)

// NodeGraphNodeRenderer draws a node shell and its default title.
type NodeGraphNodeRenderer func(layout.Context, NodeGraphNodeRequest)

// NodeGraphPortRenderer draws a port dot and its optional label.
type NodeGraphPortRenderer func(layout.Context, NodeGraphPortRequest)

// NodeGraphSelectionRenderer draws a selection or focus outline.
type NodeGraphSelectionRenderer func(layout.Context, NodeGraphSelectionRequest)

// NodeGraphResizeRenderer draws resize handles.
type NodeGraphResizeRenderer func(layout.Context, NodeGraphResizeRequest)

// NodeGraphControlsRenderer draws the fixed graph controls.
type NodeGraphControlsRenderer func(layout.Context, NodeGraphControlsRequest)

// NodeGraphMinimapRenderer draws the complete minimap.
type NodeGraphMinimapRenderer func(layout.Context, NodeGraphMinimapRequest)

// NodeGraphVisual is the complete visual plan consumed by NodeGraph behavior.
// The widget computes graph state and geometry; metrics, colors, and all
// visual primitives are owned by the provider.
type NodeGraphVisual struct {
	Defaults        style.Style
	Override        style.Style
	Metrics         NodeGraphMetrics
	Colors          NodeGraphColors
	Canvas          NodeGraphCanvasRenderer
	Grid            NodeGraphGridRenderer
	Edge            NodeGraphEdgeRenderer
	Connection      NodeGraphConnectionRenderer
	Node            NodeGraphNodeRenderer
	Port            NodeGraphPortRenderer
	Selection       NodeGraphSelectionRenderer
	Resize          NodeGraphResizeRenderer
	Controls        NodeGraphControlsRenderer
	Minimap         NodeGraphMinimapRenderer
	DisabledOpacity float32
}

// NodeGraphProvider supplies graph metrics, palette, and renderer strategies
// without taking ownership of graph data, viewport, selection, or pointer
// behavior.
type NodeGraphProvider interface {
	NodeGraph(active *Theme, request NodeGraphRequest) NodeGraphVisual
}

// ValidateNodeGraphVisual checks the complete graph geometry, palette, and
// renderer contract before a visual enters the widget cache. NodeGraph is a
// domain surface, so the widget may supply geometry facts but cannot recover
// any missing visual primitive locally.
func ValidateNodeGraphVisual(visual NodeGraphVisual) {
	if visual.Canvas == nil {
		invalidVisual("NodeGraph", "Canvas", "must be provided")
	}
	if visual.Grid == nil {
		invalidVisual("NodeGraph", "Grid", "must be provided")
	}
	if visual.Edge == nil {
		invalidVisual("NodeGraph", "Edge", "must be provided")
	}
	if visual.Connection == nil {
		invalidVisual("NodeGraph", "Connection", "must be provided")
	}
	if visual.Node == nil {
		invalidVisual("NodeGraph", "Node", "must be provided")
	}
	if visual.Port == nil {
		invalidVisual("NodeGraph", "Port", "must be provided")
	}
	if visual.Selection == nil {
		invalidVisual("NodeGraph", "Selection", "must be provided")
	}
	if visual.Resize == nil {
		invalidVisual("NodeGraph", "Resize", "must be provided")
	}
	if visual.Controls == nil {
		invalidVisual("NodeGraph", "Controls", "must be provided")
	}
	if visual.Minimap == nil {
		invalidVisual("NodeGraph", "Minimap", "must be provided")
	}

	m := visual.Metrics
	if !validPositiveDp(m.CanvasRadius) {
		invalidVisual("NodeGraph", "Metrics.CanvasRadius", "must be finite and greater than 0")
	}
	if !validPositiveDp(m.CanvasBorderWidth) {
		invalidVisual("NodeGraph", "Metrics.CanvasBorderWidth", "must be finite and greater than 0")
	}
	if !validPositiveDp(m.GridWidth) {
		invalidVisual("NodeGraph", "Metrics.GridWidth", "must be finite and greater than 0")
	}
	if !validPositiveDp(m.GridMinimumStep) {
		invalidVisual("NodeGraph", "Metrics.GridMinimumStep", "must be finite and greater than 0")
	}
	if !validPositiveDp(m.EdgeWidth) {
		invalidVisual("NodeGraph", "Metrics.EdgeWidth", "must be finite and greater than 0")
	}
	if !validPositiveDp(m.EdgeMinimumDistance) {
		invalidVisual("NodeGraph", "Metrics.EdgeMinimumDistance", "must be finite and greater than 0")
	}
	if !validPositiveDp(m.EdgeLabelWidth) {
		invalidVisual("NodeGraph", "Metrics.EdgeLabelWidth", "must be finite and greater than 0")
	}
	if !validPositiveDp(m.EdgeLabelHeight) {
		invalidVisual("NodeGraph", "Metrics.EdgeLabelHeight", "must be finite and greater than 0")
	}
	if !validPositiveDp(m.NodeHeaderHeight) {
		invalidVisual("NodeGraph", "Metrics.NodeHeaderHeight", "must be finite and greater than 0")
	}
	if !validPositiveDp(m.NodeVerticalPadding) {
		invalidVisual("NodeGraph", "Metrics.NodeVerticalPadding", "must be finite and greater than 0")
	}
	if !validPositiveDp(m.NodePortRowHeight) {
		invalidVisual("NodeGraph", "Metrics.NodePortRowHeight", "must be finite and greater than 0")
	}
	if !validPositiveDp(m.NodeDefaultWidth) {
		invalidVisual("NodeGraph", "Metrics.NodeDefaultWidth", "must be finite and greater than 0")
	}
	if !validPositiveDp(m.NodeDefaultHeight) {
		invalidVisual("NodeGraph", "Metrics.NodeDefaultHeight", "must be finite and greater than 0")
	}
	if !validPositiveDp(m.NodePadding) {
		invalidVisual("NodeGraph", "Metrics.NodePadding", "must be finite and greater than 0")
	}
	if !validPositiveDp(m.NodeRadius) {
		invalidVisual("NodeGraph", "Metrics.NodeRadius", "must be finite and greater than 0")
	}
	if !validPositiveDp(m.NodeBorderWidth) {
		invalidVisual("NodeGraph", "Metrics.NodeBorderWidth", "must be finite and greater than 0")
	}
	if !validPositiveSp(m.NodeTitleSize) {
		invalidVisual("NodeGraph", "Metrics.NodeTitleSize", "must be finite and greater than 0")
	}
	if !validPositiveSp(m.NodePortTextSize) {
		invalidVisual("NodeGraph", "Metrics.NodePortTextSize", "must be finite and greater than 0")
	}
	if !validFiniteFloat32(m.NodeTextMinScale) {
		invalidVisual("NodeGraph", "Metrics.NodeTextMinScale", "must be finite")
	}
	if !validFiniteFloat32(m.NodeTextMaxScale) {
		invalidVisual("NodeGraph", "Metrics.NodeTextMaxScale", "must be finite")
	}
	if m.NodeTextMinScale <= 0 {
		invalidVisual("NodeGraph", "Metrics.NodeTextMinScale", "must be greater than 0")
	}
	if m.NodeTextMaxScale < m.NodeTextMinScale {
		invalidVisual("NodeGraph", "Metrics.NodeTextMaxScale", "must be at least m.NodeTextMinScale")
	}
	if !validPositiveDp(m.PortRadius) {
		invalidVisual("NodeGraph", "Metrics.PortRadius", "must be finite and greater than 0")
	}
	if !validPositiveDp(m.PortBorderWidth) {
		invalidVisual("NodeGraph", "Metrics.PortBorderWidth", "must be finite and greater than 0")
	}
	if !validPositiveDp(m.EdgeMarkerSize) {
		invalidVisual("NodeGraph", "Metrics.EdgeMarkerSize", "must be finite and greater than 0")
	}
	if !validPositiveDp(m.EdgeSmoothRadius) {
		invalidVisual("NodeGraph", "Metrics.EdgeSmoothRadius", "must be finite and greater than 0")
	}
	if !validPositiveDp(m.SelectionWidth) {
		invalidVisual("NodeGraph", "Metrics.SelectionWidth", "must be finite and greater than 0")
	}
	if !validPositiveDp(m.SelectionDashLength) {
		invalidVisual("NodeGraph", "Metrics.SelectionDashLength", "must be finite and greater than 0")
	}
	if !validPositiveDp(m.SelectionDashGap) {
		invalidVisual("NodeGraph", "Metrics.SelectionDashGap", "must be finite and greater than 0")
	}
	if !validPositiveDp(m.FocusInset) {
		invalidVisual("NodeGraph", "Metrics.FocusInset", "must be finite and greater than 0")
	}
	if !validPositiveDp(m.FocusWidth) {
		invalidVisual("NodeGraph", "Metrics.FocusWidth", "must be finite and greater than 0")
	}
	if !validPositiveDp(m.FocusDashLength) {
		invalidVisual("NodeGraph", "Metrics.FocusDashLength", "must be finite and greater than 0")
	}
	if !validPositiveDp(m.FocusDashGap) {
		invalidVisual("NodeGraph", "Metrics.FocusDashGap", "must be finite and greater than 0")
	}
	if !validPositiveDp(m.ResizeHandleRadius) {
		invalidVisual("NodeGraph", "Metrics.ResizeHandleRadius", "must be finite and greater than 0")
	}
	if !validPositiveDp(m.ResizeHandleBorderWidth) {
		invalidVisual("NodeGraph", "Metrics.ResizeHandleBorderWidth", "must be finite and greater than 0")
	}
	if !validPositiveDp(m.ControlWidth) {
		invalidVisual("NodeGraph", "Metrics.ControlWidth", "must be finite and greater than 0")
	}
	if !validPositiveDp(m.ControlHeight) {
		invalidVisual("NodeGraph", "Metrics.ControlHeight", "must be finite and greater than 0")
	}
	if !validPositiveDp(m.ControlMargin) {
		invalidVisual("NodeGraph", "Metrics.ControlMargin", "must be finite and greater than 0")
	}
	if !validPositiveDp(m.ControlRadius) {
		invalidVisual("NodeGraph", "Metrics.ControlRadius", "must be finite and greater than 0")
	}
	if !validPositiveDp(m.ControlDividerInset) {
		invalidVisual("NodeGraph", "Metrics.ControlDividerInset", "must be finite and greater than 0")
	}
	if !validPositiveSp(m.ControlTextSize) {
		invalidVisual("NodeGraph", "Metrics.ControlTextSize", "must be finite and greater than 0")
	}
	if !validFiniteFloat32(m.ControlFitSpan) {
		invalidVisual("NodeGraph", "Metrics.ControlFitSpan", "must be finite")
	}
	if !validFiniteFloat32(m.ControlFitArm) {
		invalidVisual("NodeGraph", "Metrics.ControlFitArm", "must be finite")
	}
	if m.ControlFitSpan <= 0 {
		invalidVisual("NodeGraph", "Metrics.ControlFitSpan", "must be greater than 0")
	}
	if m.ControlFitArm <= 0 {
		invalidVisual("NodeGraph", "Metrics.ControlFitArm", "must be greater than 0")
	}
	if !validNonNegativeDp(m.PanelPadding) {
		invalidVisual("NodeGraph", "Metrics.PanelPadding", "must be finite and non-negative")
	}
	if !validNonNegativeDp(m.ToolbarOffset) {
		invalidVisual("NodeGraph", "Metrics.ToolbarOffset", "must be finite and non-negative")
	}
	if !validPositiveDp(m.MinimapWidth) {
		invalidVisual("NodeGraph", "Metrics.MinimapWidth", "must be finite and greater than 0")
	}
	if !validPositiveDp(m.MinimapHeight) {
		invalidVisual("NodeGraph", "Metrics.MinimapHeight", "must be finite and greater than 0")
	}
	if !validPositiveDp(m.MinimapMargin) {
		invalidVisual("NodeGraph", "Metrics.MinimapMargin", "must be finite and greater than 0")
	}
	if !validPositiveDp(m.MinimapPadding) {
		invalidVisual("NodeGraph", "Metrics.MinimapPadding", "must be finite and greater than 0")
	}
	if !validPositiveDp(m.MinimapRadius) {
		invalidVisual("NodeGraph", "Metrics.MinimapRadius", "must be finite and greater than 0")
	}
	if !validPositiveDp(m.MinimapBorderWidth) {
		invalidVisual("NodeGraph", "Metrics.MinimapBorderWidth", "must be finite and greater than 0")
	}
	if !validPositiveDp(m.MinimapEdgeWidth) {
		invalidVisual("NodeGraph", "Metrics.MinimapEdgeWidth", "must be finite and greater than 0")
	}
	if !validPositiveDp(m.MinimapViewportWidth) {
		invalidVisual("NodeGraph", "Metrics.MinimapViewportWidth", "must be finite and greater than 0")
	}
	if !validPositiveDp(m.MinimapNodeMinSize) {
		invalidVisual("NodeGraph", "Metrics.MinimapNodeMinSize", "must be finite and greater than 0")
	}
	if !validNonNegativeDp(m.ShadowOffsetY) {
		invalidVisual("NodeGraph", "Metrics.ShadowOffsetY", "must be finite and non-negative")
	}
	if !validNonNegativeDp(m.ShadowBlur) {
		invalidVisual("NodeGraph", "Metrics.ShadowBlur", "must be finite and non-negative")
	}
	if !validNonNegativeDp(m.ShadowSpread) {
		invalidVisual("NodeGraph", "Metrics.ShadowSpread", "must be finite and non-negative")
	}
	if !hasValidOpacityValue(m.ShadowOpacity) {
		invalidVisual("NodeGraph", "Metrics.ShadowOpacity", "must be finite and between 0 and 1")
	}
	if !hasValidOpacityValue(visual.DisabledOpacity) {
		invalidVisual("NodeGraph", "DisabledOpacity", "must be finite and between 0 and 1")
	}

	if visual.Colors.CanvasBackground.A == 0 {
		invalidVisual("NodeGraph", "Colors.CanvasBackground", "must have nonzero alpha")
	}
	if visual.Colors.CanvasBorder.A == 0 {
		invalidVisual("NodeGraph", "Colors.CanvasBorder", "must have nonzero alpha")
	}
	if visual.Colors.Grid.A == 0 {
		invalidVisual("NodeGraph", "Colors.Grid", "must have nonzero alpha")
	}
	if visual.Colors.NodeBackground.A == 0 {
		invalidVisual("NodeGraph", "Colors.NodeBackground", "must have nonzero alpha")
	}
	if visual.Colors.NodeBorder.A == 0 {
		invalidVisual("NodeGraph", "Colors.NodeBorder", "must have nonzero alpha")
	}
	if visual.Colors.NodeForeground.A == 0 {
		invalidVisual("NodeGraph", "Colors.NodeForeground", "must have nonzero alpha")
	}
	if visual.Colors.NodeMutedForeground.A == 0 {
		invalidVisual("NodeGraph", "Colors.NodeMutedForeground", "must have nonzero alpha")
	}
	if visual.Colors.PortColor.A == 0 {
		invalidVisual("NodeGraph", "Colors.PortColor", "must have nonzero alpha")
	}
	if visual.Colors.PortBorder.A == 0 {
		invalidVisual("NodeGraph", "Colors.PortBorder", "must have nonzero alpha")
	}
	if visual.Colors.EdgeColor.A == 0 {
		invalidVisual("NodeGraph", "Colors.EdgeColor", "must have nonzero alpha")
	}
	if visual.Colors.SelectedEdgeColor.A == 0 {
		invalidVisual("NodeGraph", "Colors.SelectedEdgeColor", "must have nonzero alpha")
	}
	if visual.Colors.SelectedNodeBorder.A == 0 {
		invalidVisual("NodeGraph", "Colors.SelectedNodeBorder", "must have nonzero alpha")
	}
	if visual.Colors.SelectionFill.A == 0 {
		invalidVisual("NodeGraph", "Colors.SelectionFill", "must have nonzero alpha")
	}
	if visual.Colors.SelectionBorder.A == 0 {
		invalidVisual("NodeGraph", "Colors.SelectionBorder", "must have nonzero alpha")
	}
	if visual.Colors.ControlBackground.A == 0 {
		invalidVisual("NodeGraph", "Colors.ControlBackground", "must have nonzero alpha")
	}
	if visual.Colors.ControlForeground.A == 0 {
		invalidVisual("NodeGraph", "Colors.ControlForeground", "must have nonzero alpha")
	}
	if visual.Colors.ControlBorder.A == 0 {
		invalidVisual("NodeGraph", "Colors.ControlBorder", "must have nonzero alpha")
	}
	if visual.Colors.Shadow.A == 0 {
		invalidVisual("NodeGraph", "Colors.Shadow", "must have nonzero alpha")
	}
	if visual.Colors.ResizeHandleFill.A == 0 {
		invalidVisual("NodeGraph", "Colors.ResizeHandleFill", "must have nonzero alpha")
	}
	root := style.Cascade(style.State{}, visual.Defaults, visual.Override)
	if root.Box == nil {
		invalidVisual("NodeGraph", "Root.Box", "must be provided")
	}
	if root.Paint == nil {
		invalidVisual("NodeGraph", "Root.Paint", "must be provided")
	}
	if !hasValidOpacity(root.Paint.Opacity) {
		invalidVisual("NodeGraph", "Root.Paint.Opacity", "must be finite and between 0 and 1")
	}
}
