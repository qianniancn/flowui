package input

import (
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"github.com/qianniancn/FlowUI/internal/frame"
)

func (i InputWidget) withSemantics(ctx *frame.Context, key string, enabled bool, child layout.Widget) layout.Widget {
	return withEditorSemantics(ctx, key, i.label, enabled, child)
}

func (t TextAreaWidget) withSemantics(ctx *frame.Context, key string, enabled bool, child layout.Widget) layout.Widget {
	return withEditorSemantics(ctx, key, t.label, enabled, child)
}

func withEditorSemantics(ctx *frame.Context, key, label string, enabled bool, child layout.Widget) layout.Widget {
	if label == "" {
		label = frame.FieldLabel(ctx, key)
	}
	description := frame.FieldDescription(ctx, key)
	return func(gtx layout.Context) layout.Dimensions {
		// Gio suppresses the editor class when its event source is disabled.
		if !enabled {
			semantic.Editor.Add(gtx.Ops)
		}
		semantic.EnabledOp(enabled).Add(gtx.Ops)
		if label != "" {
			semantic.LabelOp(label).Add(gtx.Ops)
		}
		if description != "" {
			semantic.DescriptionOp(description).Add(gtx.Ops)
		}
		return child(gtx)
	}
}
