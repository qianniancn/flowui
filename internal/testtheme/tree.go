package testtheme

import (
	"image"
	"image/color"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

func (behaviorProvider) Tree(active *theme.Theme, request theme.TreeRequest) theme.TreeVisual {
	return treeBehaviorVisual(active, request)
}

// treeBehaviorVisual supplies geometry sufficient for interaction tests without
// executing the production Tree recipe or drawing implementation.
func treeBehaviorVisual(active *theme.Theme, request theme.TreeRequest) theme.TreeVisual {
	metrics := theme.TreeMetrics{
		Padding: 4, Gap: 2, MaxHeight: 320, RowHeight: 32, DescriptionRowHeight: 48,
		RowRadius: 4, RowPaddingX: 8, RowPaddingY: 4, Indent: 16,
		ChevronSlotSize: 20, ChevronIconSize: 12, ContentGap: 6,
		ItemTextSize: 14, ItemDescriptionSize: 12, ItemTextWeight: font.Medium,
		FocusRingWidth: 1, SurfaceRadius: 4, DragPreviewOffset: 8,
		DragPreviewMaxWidth: 240, DragPreviewPaddingX: 8, DragPreviewPaddingY: 6,
		DragPreviewRadius: 4,
	}
	if request.Size == theme.TreeSizeSmall {
		metrics.Padding, metrics.Gap = 2, 1
		metrics.RowHeight, metrics.DescriptionRowHeight = 24, 40
		metrics.RowPaddingX, metrics.RowPaddingY, metrics.Indent = 4, 2, 12
		metrics.ChevronSlotSize, metrics.ChevronIconSize, metrics.ContentGap = 16, 12, 5
		metrics.ItemTextSize, metrics.ItemDescriptionSize = 13, 11
	}
	colors := theme.TreeColors{
		RootForeground: foreground, Foreground: foreground, Description: muted,
		RowHover: background, RowSelected: accent, RowSelectedHover: accent,
		SelectedForeground: accentFg, SelectedDescription: accentFg,
		Chevron: muted, Focus: focus, Guides: muted, DropIndicator: focus,
		Retry: focus, ToggleBackground: background,
		DragPreviewBackground: background, DragPreviewForeground: foreground,
		EditorSelection: accent,
	}
	if request.Variant == theme.TreeVariantSurface {
		colors.RootBackground = background
	}
	item := func(value theme.TreeItemRequest) theme.TreeItemVisual {
		result := theme.TreeItemVisual{
			Foreground: foreground, Description: muted, Chevron: muted, Focus: focus, Opacity: 1,
		}
		if value.Hovered {
			result.Background = background
		}
		if value.Selected {
			result.Background = accent
			result.Foreground, result.Description, result.Chevron = accentFg, accentFg, accentFg
		}
		if value.Disabled {
			result.Opacity = 0.5
			result.Focus = color.NRGBA{}
		}
		return result
	}
	return theme.TreeVisual{
		Defaults: style.Style{}.FillWidth().Background(style.Solid(colors.RootBackground)).Opacity(1),
		Override: theme.TreeOverrideStyle(request),
		Metrics:  metrics, Colors: colors, DisabledOpacity: 0.5, Motion: active.Motion,
		Item: item,
		EditorStyle: func(value theme.TreeItemRequest) theme.TreeEditorStyle {
			return theme.TreeEditorStyle{TextSize: metrics.ItemTextSize, Foreground: item(value).Foreground, Selection: accent, FontWeight: font.Medium}
		},
		Root:          func(layout.Context, image.Point, unit.Dp, color.NRGBA, bool) {},
		Row:           func(layout.Context, image.Point, theme.TreeItemVisual, float32, unit.Dp) {},
		DropIndicator: func(layout.Context, image.Point, theme.TreeDropPosition, int, theme.TreeMetrics, color.NRGBA) {},
		Toggle:        func(layout.Context, image.Point, theme.TreeMetrics, float32, bool, color.NRGBA, color.NRGBA) {},
		Retry:         func(layout.Context, image.Point, theme.TreeMetrics, color.NRGBA) {},
		Guides:        func(layout.Context, theme.TreeGuidesRequest, unit.Dp) {},
		Text: func(gtx layout.Context, value theme.TreeTextRequest) layout.Dimensions {
			width := len([]rune(value.Text)) * gtx.Sp(value.Size) / 2
			return layout.Dimensions{Size: gtx.Constraints.Constrain(image.Pt(width, gtx.Sp(value.Size)+4))}
		},
		DragPreview: func(gtx layout.Context, value theme.TreeDragPreviewRequest, metrics theme.TreeMetrics) layout.Dimensions {
			width := min(len([]rune(value.Label))*gtx.Sp(metrics.ItemTextSize)/2+2*gtx.Dp(metrics.DragPreviewPaddingX), gtx.Dp(metrics.DragPreviewMaxWidth))
			height := gtx.Dp(metrics.RowHeight + 2*metrics.DragPreviewPaddingY)
			return layout.Dimensions{Size: image.Pt(width, height)}
		},
	}
}
