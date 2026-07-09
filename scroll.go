package flowui

import "gioui.org/layout"

type ScrollWidget struct {
	key           string
	child         Widget
	axis          layout.Axis
	align         layout.Alignment
	disabled      bool
	stickToEnd    bool
	scrollAnyAxis bool
}

func Scroll(key string, child Widget) ScrollWidget {
	return ScrollWidget{
		key:   key,
		child: child,
		axis:  layout.Vertical,
	}
}

func (s ScrollWidget) Vertical() ScrollWidget {
	s.axis = layout.Vertical
	return s
}

func (s ScrollWidget) Horizontal() ScrollWidget {
	s.axis = layout.Horizontal
	return s
}

func (s ScrollWidget) Disabled(disabled bool) ScrollWidget {
	s.disabled = disabled
	return s
}

func (s ScrollWidget) AlignStart() ScrollWidget {
	s.align = layout.Start
	return s
}

func (s ScrollWidget) AlignMiddle() ScrollWidget {
	s.align = layout.Middle
	return s
}

func (s ScrollWidget) AlignEnd() ScrollWidget {
	s.align = layout.End
	return s
}

func (s ScrollWidget) StickToEnd() ScrollWidget {
	s.stickToEnd = true
	return s
}

func (s ScrollWidget) ScrollAnyAxis() ScrollWidget {
	s.scrollAnyAxis = true
	return s
}

func (s ScrollWidget) Layout(ctx *Context, gtx layout.Context) layout.Dimensions {
	state := ctx.scrollState(s.key)
	state.Axis = s.axis
	state.Gap = 0
	state.Alignment = s.align
	state.ScrollToEnd = s.stickToEnd
	state.ScrollAnyAxis = s.scrollAnyAxis
	if s.disabled {
		gtx = gtx.Disabled()
	}
	return state.Layout(gtx, 1, func(gtx layout.Context, _ int) layout.Dimensions {
		return s.child.Layout(ctx, gtx)
	})
}
