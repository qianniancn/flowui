package layout

import (
	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/core"
)

type KeyWidget struct {
	key   string
	child Widget
}

func Key(key string, child Widget) KeyWidget {
	return KeyWidget{key: key, child: child}
}

func (k KeyWidget) Layout(ctx *Context, gtx layout.Context) layout.Dimensions {
	pop := core.PushKey(ctx, k.key)
	defer pop()
	prepareFieldAssociations(ctx, k.child)
	return k.child.Layout(ctx, gtx)
}
