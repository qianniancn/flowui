package layoutui

import (
	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/frame"
)

type KeyWidget struct {
	key   string
	child frame.Widget
}

func Key(key string, child frame.Widget) KeyWidget {
	return KeyWidget{key: key, child: child}
}

func (k KeyWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	pop := frame.PushKey(ctx, k.key)
	defer pop()
	prepareFieldAssociations(ctx, k.child)
	return k.child.Layout(ctx, gtx)
}
