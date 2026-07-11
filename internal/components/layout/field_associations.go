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
		if widget == nil || label.PrepareFieldAssociation(ctx, widget) || description.PrepareFieldAssociation(ctx, widget) {
			continue
		}
		switch widget := widget.(type) {
		case AdaptiveWidget, *AdaptiveWidget, ListWidget, *ListWidget:
			// Their children depend on the current layout pass and are prepared there.
		case AspectRatioWidget:
			prepareFieldAssociations(ctx, widget.child)
		case *AspectRatioWidget:
			if widget != nil {
				prepareFieldAssociations(ctx, widget.child)
			}
		case BoxWidget:
			prepareFieldAssociations(ctx, widget.child)
		case *BoxWidget:
			if widget != nil {
				prepareFieldAssociations(ctx, widget.child)
			}
		case CenterWidget:
			prepareFieldAssociations(ctx, widget.child)
		case *CenterWidget:
			if widget != nil {
				prepareFieldAssociations(ctx, widget.child)
			}
		case ColumnWidget:
			prepareFieldAssociations(ctx, widget.children...)
		case *ColumnWidget:
			if widget != nil {
				prepareFieldAssociations(ctx, widget.children...)
			}
		case FlexWidget:
			prepareFieldAssociations(ctx, widget.child)
		case *FlexWidget:
			if widget != nil {
				prepareFieldAssociations(ctx, widget.child)
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
			prepareFieldAssociations(ctx, widget.child)
		case *ScrollWidget:
			if widget != nil {
				prepareFieldAssociations(ctx, widget.child)
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
}

func prepareScopedFieldAssociations(ctx *frame.Context, key string, child frame.Widget) {
	pop := frame.PushKey(ctx, key)
	defer pop()
	prepareFieldAssociations(ctx, child)
}

func prepareStackFieldAssociations(ctx *frame.Context, layers []StackLayer) {
	for _, layer := range layers {
		prepareFieldAssociations(ctx, layer.child)
	}
}
