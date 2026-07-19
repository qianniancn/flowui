package alert

import (
	"gioui.org/layout"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/theme"
)

// Status identifies the semantic tone of an Alert.
type Status uint8

const (
	StatusDefault Status = iota
	StatusAccent
	StatusSuccess
	StatusWarning
	StatusDanger
)

// Widget presents persistent, in-page feedback without interrupting the user.
type Widget struct {
	theme       func(*theme.Theme)
	title       string
	description string
	status      Status
	indicator   frame.Widget
	content     frame.Widget
	action      frame.Widget
}

func New(title, description string) Widget {
	return Widget{title: title, description: description}
}

func (a Widget) Status(status Status) Widget {
	a.status = status
	return a
}

func (a Widget) Indicator(indicator frame.Widget) Widget {
	a.indicator = indicator
	return a
}

// Content replaces the standard description with custom content.
func (a Widget) Content(content frame.Widget) Widget {
	a.content = content
	return a
}

func (a Widget) Action(action frame.Widget) Widget {
	a.action = action
	return a
}

func (a Widget) Theme(fn func(*theme.Theme)) Widget {
	a.theme = fn
	return a
}

func (a Widget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	if restore := frame.PushInstanceTheme(ctx, a.theme); restore != nil {
		defer restore()
	}
	return a.layout(ctx, gtx, alertStyleFor(frame.ActiveTheme(ctx), a.status))
}
