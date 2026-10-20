package testtheme

import (
	"image/color"

	"gioui.org/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

func (behaviorProvider) NodeGraph(active *theme.Theme, request theme.NodeGraphRequest) theme.NodeGraphVisual {
	tokens := active.Components.NodeGraph
	return theme.NodeGraphVisual{
		Defaults: style.Style{}.
			FillWidth().
			Background(style.Solid(background)).
			Opacity(1).
			Rules(style.Rule{When: style.Disabled, Set: style.Style{}.
				Opacity(active.DisabledOpacityValue())}),
		Override: theme.NodeGraphOverrideStyle(request),
		Metrics: theme.NodeGraphMetrics{
			CanvasRadius: tokens.CanvasRadius, CanvasBorderWidth: tokens.CanvasBorderWidth,
			GridWidth: 1, GridMinimumStep: 8, EdgeWidth: 1, EdgeMinimumDistance: 32,
			EdgeLabelWidth: 60, EdgeLabelHeight: 14, NodeHeaderHeight: 26, NodeVerticalPadding: 8, NodePortRowHeight: 20,
			NodeDefaultWidth: 150, NodeDefaultHeight: 64,
			NodePadding: 10, NodeRadius: 3, NodeBorderWidth: 1, NodeTitleSize: 12,
			NodePortTextSize: 10, NodeTextMinScale: .75, NodeTextMaxScale: 1.333333,
			PortRadius: 3, PortBorderWidth: 1, EdgeMarkerSize: 5, EdgeSmoothRadius: 8,
			SelectionWidth: 1, SelectionDashLength: 2, SelectionDashGap: 2,
			FocusInset: 2, FocusWidth: 1, FocusDashLength: 2, FocusDashGap: 2,
			ResizeHandleRadius: 4, ResizeHandleBorderWidth: 1, ControlWidth: 108,
			ControlHeight: 36, ControlMargin: 12, ControlRadius: 5, ControlDividerInset: 8,
			ControlTextSize: 16, ControlFitSpan: .18, ControlFitArm: .8, PanelPadding: 12, ToolbarOffset: 12,
			MinimapWidth:  180,
			MinimapHeight: 120, MinimapMargin: 12, MinimapPadding: 8, MinimapRadius: 5,
			MinimapBorderWidth: 1, MinimapEdgeWidth: .7, MinimapViewportWidth: 1,
			MinimapNodeMinSize: 2, ShadowOffsetY: 1, ShadowBlur: 4, ShadowSpread: 1,
			ShadowOpacity: 1,
		},
		Colors: theme.NodeGraphColors{
			CanvasBackground: background, CanvasBorder: muted, Grid: muted,
			NodeBackground: background, NodeBorder: muted, NodeForeground: foreground,
			NodeMutedForeground: muted, PortColor: accent, PortBorder: foreground,
			EdgeColor: accent, SelectedEdgeColor: focus, SelectedNodeBorder: focus,
			SelectionFill: accent, SelectionBorder: focus, ControlBackground: background,
			ControlForeground: foreground, ControlBorder: muted, Shadow: muted,
			ResizeHandleFill: color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff},
		},
		Canvas:          func(layout.Context, theme.NodeGraphCanvasRequest) {},
		Grid:            func(layout.Context, theme.NodeGraphGridRequest) {},
		Edge:            func(layout.Context, theme.NodeGraphEdgeRequest) {},
		Connection:      func(layout.Context, theme.NodeGraphConnectionRequest) {},
		Node:            func(layout.Context, theme.NodeGraphNodeRequest) {},
		Port:            func(layout.Context, theme.NodeGraphPortRequest) {},
		Selection:       func(layout.Context, theme.NodeGraphSelectionRequest) {},
		Resize:          func(layout.Context, theme.NodeGraphResizeRequest) {},
		Controls:        func(layout.Context, theme.NodeGraphControlsRequest) {},
		Minimap:         func(layout.Context, theme.NodeGraphMinimapRequest) {},
		DisabledOpacity: active.DisabledOpacityValue(),
	}
}
