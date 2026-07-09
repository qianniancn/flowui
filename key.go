package flowui

import "gioui.org/layout"

type KeyWidget struct {
	key   string
	child Widget
}

func Key(key string, child Widget) KeyWidget {
	return KeyWidget{key: key, child: child}
}

func (k KeyWidget) Layout(ctx *Context, gtx layout.Context) layout.Dimensions {
	pop := ctx.pushKey(k.key)
	defer pop()
	return k.child.Layout(ctx, gtx)
}
