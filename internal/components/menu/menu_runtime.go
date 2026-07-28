package menu

import (
	"gioui.org/layout"
	"github.com/qianniancn/FlowUI/internal/frame"
)

// Runtime binds a Menu widget to one owner-specific state instance.
type Runtime struct {
	widget Widget
	state  *menuState
}

func (m Widget) Runtime(ctx *frame.Context, owner, role string, onRequestClose func(bool)) Runtime {
	m = m.withDerivedIdentity(owner, role).withClose(onRequestClose)
	return Runtime{widget: m, state: m.stateFor(ctx)}
}

func (r Runtime) Layout(ctx *frame.Context, gtx layout.Context, interactive bool) layout.Dimensions {
	return r.widget.layoutRoot(ctx, gtx, r.state, interactive)
}

// RootNavigation hands horizontal navigation at the root menu boundary to its
// owner. Nested submenus retain Left Arrow for returning to their parent.
func (r Runtime) RootNavigation(previous, next func()) Runtime {
	r.widget.onRootPrevious = previous
	r.widget.onRootNext = next
	return r
}

func (r Runtime) HasActiveSubmenu() bool {
	return r.state != nil && (r.state.openSubmenu != "" || r.state.submenuActive)
}

func (r Runtime) CloseSubmenus() {
	if r.state != nil {
		r.state.openSubmenu = ""
	}
}

func (r Runtime) FocusFirst(ctx *frame.Context, visible bool) bool {
	return r.state != nil && r.state.focusFirstEntry(ctx, r.widget, visible)
}

func (r Runtime) FocusLast(ctx *frame.Context, visible bool) bool {
	return r.state != nil && r.state.focusLastEntry(ctx, r.widget, visible)
}
