package layoutui

import (
	"github.com/qianniancn/FlowUI/internal/components/description"
	"github.com/qianniancn/FlowUI/internal/components/label"
	"github.com/qianniancn/FlowUI/internal/frame"
)

// PrepareFieldAssociations pre-registers semantic field relationships for a
// composite component before its children are laid out.
func PrepareFieldAssociations(ctx *frame.Context, widgets ...frame.Widget) {
	prepareFieldAssociations(ctx, widgets...)
}

func prepareFieldAssociations(ctx *frame.Context, widgets ...frame.Widget) {
	for _, widget := range widgets {
		prepareFieldAssociation(ctx, widget)
	}
}

func prepareFieldAssociation(ctx *frame.Context, widget frame.Widget) {
	if widget == nil || label.PrepareFieldAssociation(ctx, widget) || description.PrepareFieldAssociation(ctx, widget) {
		return
	}
	switch widget := widget.(type) {
	case AdaptiveWidget, *AdaptiveWidget, ListWidget, *ListWidget:
		// Their children depend on the current layout pass and are prepared there.
	case AspectRatioWidget:
		prepareFieldAssociation(ctx, widget.child)
	case *AspectRatioWidget:
		if widget != nil {
			prepareFieldAssociation(ctx, widget.child)
		}
	case BoxWidget:
		prepareFieldAssociation(ctx, widget.child)
	case *BoxWidget:
		if widget != nil {
			prepareFieldAssociation(ctx, widget.child)
		}
	case CenterWidget:
		prepareFieldAssociation(ctx, widget.child)
	case *CenterWidget:
		if widget != nil {
			prepareFieldAssociation(ctx, widget.child)
		}
	case ColumnWidget:
		prepareFieldAssociations(ctx, widget.children...)
	case *ColumnWidget:
		if widget != nil {
			prepareFieldAssociations(ctx, widget.children...)
		}
	case FlexWidget:
		prepareFieldAssociation(ctx, widget.child)
	case *FlexWidget:
		if widget != nil {
			prepareFieldAssociation(ctx, widget.child)
		}
	case GridWidget:
		prepareFieldAssociations(ctx, widget.children...)
	case *GridWidget:
		if widget != nil {
			prepareFieldAssociations(ctx, widget.children...)
		}
	case KeyWidget:
		prepareScopedFieldAssociations(ctx, widget.key, widget.child)
	case *KeyWidget:
		if widget != nil {
			prepareScopedFieldAssociations(ctx, widget.key, widget.child)
		}
	case RowWidget:
		prepareFieldAssociations(ctx, widget.children...)
	case *RowWidget:
		if widget != nil {
			prepareFieldAssociations(ctx, widget.children...)
		}
	case ScrollWidget:
		prepareFieldAssociation(ctx, widget.child)
	case *ScrollWidget:
		if widget != nil {
			prepareFieldAssociation(ctx, widget.child)
		}
	case ScrollbarWidget:
		prepareFieldAssociation(ctx, widget.child)
	case *ScrollbarWidget:
		if widget != nil {
			prepareFieldAssociation(ctx, widget.child)
		}
	case SplitPaneWidget:
		prepareFieldAssociation(ctx, widget.first)
		prepareFieldAssociation(ctx, widget.second)
	case *SplitPaneWidget:
		if widget != nil {
			prepareFieldAssociation(ctx, widget.first)
			prepareFieldAssociation(ctx, widget.second)
		}
	case StackWidget:
		prepareStackFieldAssociations(ctx, widget.layers)
	case *StackWidget:
		if widget != nil {
			prepareStackFieldAssociations(ctx, widget.layers)
		}
	case WrapWidget:
		prepareFieldAssociations(ctx, widget.children...)
	case *WrapWidget:
		if widget != nil {
			prepareFieldAssociations(ctx, widget.children...)
		}
	}
}

func prepareScopedFieldAssociations(ctx *frame.Context, key string, child frame.Widget) {
	pop := frame.PushKey(ctx, key)
	defer pop()
	prepareFieldAssociation(ctx, child)
}

func prepareStackFieldAssociations(ctx *frame.Context, layers []StackLayer) {
	for _, layer := range layers {
		prepareFieldAssociation(ctx, layer.child)
	}
}
