package input

import (
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/core"
)

func (i Widget) withSemantics(ctx *core.Context, key string, enabled bool, child layout.Widget) layout.Widget {
	return withEditorSemantics(ctx, key, i.label, enabled, child)
}

func (t TextAreaWidget) withSemantics(ctx *core.Context, key string, enabled bool, child layout.Widget) layout.Widget {
	return withEditorSemantics(ctx, key, t.label, enabled, child)
}

func withEditorSemantics(ctx *core.Context, key, label string, enabled bool, child layout.Widget) layout.Widget {
	if label == "" {
		label = core.FieldLabel(ctx, key)
	}
	description := core.FieldDescription(ctx, key)
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
