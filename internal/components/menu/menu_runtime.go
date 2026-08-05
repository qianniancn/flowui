package menu

import (
	"image/color"

	"gioui.org/layout"
	"github.com/qianniancn/flowui/internal/frame"
	flowstyle "github.com/qianniancn/flowui/internal/style"
	styleruntime "github.com/qianniancn/flowui/internal/style/runtime"
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

// HoveredWithSubmenus includes the visible descendants of a root menu. It is
// used by hover-triggered owners that need to know whether the pointer remains
// inside a nested popup, rather than merely whether a submenu is open.
func (r Runtime) HoveredWithSubmenus(ctx *frame.Context) bool {
	if r.state == nil {
		return false
	}
	return menuHovered(ctx, r.widget, r.state)
}

func menuHovered(ctx *frame.Context, widget Widget, state *menuState) bool {
	entries := state.resolveEntries(widget)
	actionable := state.actionableEntries(entries)
	for _, item := range state.items {
		if item.clickable.Hovered() {
			return true
		}
	}
	if state.openSubmenu == "" {
		return false
	}
	for _, entry := range actionable {
		if entry.item.Key != state.openSubmenu || !itemHasSubmenu(entry.item) {
			continue
		}
		child := widget.submenu(state, entry.item)
		childState := child.peekState(ctx)
		if childState == nil {
			continue
		}
		if childState.dialog.Hovered() || menuHovered(ctx, child, childState) {
			return true
		}
	}
	return false
}

// peekState reads an existing menu state without claiming its key for the
// current layout pass. Hover inspection runs before the menu is laid out and
// must not make the later layout claim the same derived submenu key twice.
func (m Widget) peekState(ctx *frame.Context) *menuState {
	var key string
	if m.derivedOwner == "" {
		key = frame.FullKey(ctx, m.key)
	} else {
		key = frame.DerivedKey(ctx, m.derivedOwner, m.derivedRole)
	}
	state, _ := frame.PeekState[menuState](ctx, key, stateSlotMenu)
	return state
}

func (r Runtime) CloseSubmenus() {
	if r.state != nil {
		r.state.openSubmenu = ""
	}
}

// PanelColors returns the resolved popup surface and border colors, including
// the menu's instance style. Non-solid custom backgrounds fall back to the
// theme surface because an arrow cannot reproduce an arbitrary brush.
func (r Runtime) PanelColors(ctx *frame.Context) (background, border color.NRGBA) {
	if r.state == nil {
		return color.NRGBA{}, color.NRGBA{}
	}
	activeTheme := frame.ActiveTheme(ctx)
	panel := menuPanelStyle(activeTheme)
	background = panel.background
	border = panel.border
	resolved := styleruntime.ResolveStatic(ctx, flowstyle.StyleState{}, menuRootDeclaration(activeTheme, r.widget.themeTokens(ctx)), flowstyle.Style{}, flowstyle.Style{}, r.widget.customStyle)
	if resolved.Paint == nil {
		return background, border
	}
	if source, ok := resolved.Paint.Background.(flowstyle.ColorSource); ok {
		if value, ok := styleruntime.ResolveColor(ctx, source); ok {
			background = value
		}
	}
	if resolved.Paint.Border != nil {
		if value, ok := styleruntime.ResolveColor(ctx, resolved.Paint.Border.Color); ok {
			border = value
		}
	}
	return background, border
}

func (r Runtime) FocusFirst(ctx *frame.Context, visible bool) bool {
	return r.state != nil && r.state.focusFirstEntry(ctx, r.widget, visible)
}

func (r Runtime) FocusLast(ctx *frame.Context, visible bool) bool {
	return r.state != nil && r.state.focusLastEntry(ctx, r.widget, visible)
}
