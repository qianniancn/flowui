package alert

import (
	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

// Status identifies the semantic tone of an alert.
type Status = theme.AlertStatus

const (
	Default = theme.AlertDefault
	Accent  = theme.AlertAccent
	Success = theme.AlertSuccess
	Warning = theme.AlertWarning
	Danger  = theme.AlertDanger
)

// Widget presents persistent, in-page feedback without interrupting the user.
type Widget struct {
	override    *theme.AlertOverride
	key         string
	title       string
	description string
	status      Status
	indicator   core.Widget
	content     core.Widget
	action      core.Widget
	closable    bool
	onClose     func()
	visualCache *alertVisualCache
}

func New(title, description string) Widget {
	return Widget{title: title, description: description, visualCache: &alertVisualCache{}}
}

// Key sets the close button's identity. Alerts with the same text otherwise
// derive the same key, so set Key when two identical alerts can coexist.
func (a Widget) Key(key string) Widget {
	a.key = key
	return a
}

// Closable adds a close button at the trailing edge.
func (a Widget) Closable(closable bool) Widget {
	a.closable = closable
	return a
}

// OnClose runs when the close button is activated.
func (a Widget) OnClose(fn func()) Widget {
	a.onClose = fn
	return a
}

func (a Widget) Status(status Status) Widget {
	a.status = status
	return a
}

func (a Widget) Indicator(indicator core.Widget) Widget {
	a.indicator = indicator
	return a
}

// Content replaces the standard description with custom content.
func (a Widget) Content(content core.Widget) Widget {
	a.content = content
	return a
}

func (a Widget) Action(action core.Widget) Widget {
	a.action = action
	return a
}

func (a Widget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	return a.layoutWithOverride(ctx, gtx, a.override)
}

// Style merges instance declarations into the alert. The declaration reaches
// the provider at layout time, so Style can sit anywhere in the chain.
func (a Widget) Style(values ...style.Style) Widget {
	declaration := style.Join(values...)
	if a.override == nil {
		a.override = &theme.AlertOverride{}
	}
	a.override.Declaration = style.Join(a.override.Declaration, declaration)
	return a
}
