//go:build windows

package systray

import (
	"syscall"

	"github.com/qianniancn/flowui/internal/sys/windows"
)

type windowsMenu struct {
	parent *Menu
	owner  *windowsSystemTray
	hmenu  windows.HMENU
}

var trackWindowsPopupMenu = windows.TrackPopupMenu

func newWindowsMenu(parent *Menu, owner *windowsSystemTray) *windowsMenu {
	hmenu, err := windows.CreatePopupMenu()
	if err != nil {
		return nil
	}
	wm := &windowsMenu{parent: parent, owner: owner, hmenu: hmenu}
	wm.buildMenu()
	return wm
}

func (wm *windowsMenu) buildMenu() {
	if wm == nil || wm.hmenu == 0 {
		return
	}
	for _, item := range wm.parent.Items() {
		wm.addMenuItem(item)
	}
}

func (wm *windowsMenu) addMenuItem(item *MenuItem) {
	item.mu.Lock()
	itemType := item.itemType
	label := item.label
	disabled := item.disabled
	checked := item.checked
	hidden := item.hidden
	submenu := item.submenu
	itemID := item.id
	item.mu.Unlock()
	if hidden {
		return
	}

	switch itemType {
	case MenuItemSeparator:
		_ = windows.AppendMenu(wm.hmenu, windows.MF_SEPARATOR, 0, nil)
	case MenuItemSubmenu:
		if submenu == nil {
			return
		}
		submenuImpl := newWindowsMenu(submenu, wm.owner)
		if submenuImpl == nil {
			return
		}
		submenu.mu.Lock()
		submenu.impl = submenuImpl
		submenu.mu.Unlock()
		utf16Label, _ := syscall.UTF16PtrFromString(label)
		flags := windows.MF_STRING | windows.MF_POPUP
		if disabled {
			flags |= windows.MF_DISABLED | windows.MF_GRAYED
		}
		_ = windows.AppendMenu(wm.hmenu, windows.UINT(flags), uintptr(submenuImpl.hmenu), utf16Label)
	default:
		utf16Label, _ := syscall.UTF16PtrFromString(label)
		flags := windows.MF_STRING
		if disabled {
			flags |= windows.MF_DISABLED | windows.MF_GRAYED
		}
		if checked {
			flags |= windows.MF_CHECKED
		}
		_ = windows.AppendMenu(wm.hmenu, windows.UINT(flags), uintptr(itemID), utf16Label)
		item.mu.Lock()
		item.impl = &windowsMenuItem{parent: item, menu: wm}
		item.mu.Unlock()
	}
}

func (wm *windowsMenu) dispatch(fn func()) {
	if wm.owner == nil {
		fn()
		return
	}
	wm.owner.post(fn)
}

func (wm *windowsMenu) update() {
	wm.dispatch(wm.updateOnThread)
}

func (wm *windowsMenu) updateOnThread() {
	if wm.hmenu != 0 {
		_ = windows.DestroyMenu(wm.hmenu)
	}
	hmenu, err := windows.CreatePopupMenu()
	if err != nil {
		wm.hmenu = 0
		return
	}
	wm.hmenu = hmenu
	wm.buildMenu()
}

func (wm *windowsMenu) show() {
	wm.dispatch(func() {
		if wm.owner != nil {
			wm.showOnThread(wm.owner.windowHandle())
		}
	})
}

func (wm *windowsMenu) showOnThread(hwnd windows.HWND) {
	if wm == nil || wm.hmenu == 0 || hwnd == 0 {
		return
	}
	var pt windows.POINT
	if err := windows.GetCursorPos(&pt); err != nil {
		return
	}
	windows.SetForegroundWindow(hwnd)
	trackWindowsPopupMenu(
		wm.hmenu,
		windows.TPM_BOTTOMALIGN|windows.TPM_LEFTALIGN|windows.TPM_RIGHTBUTTON,
		pt.X,
		pt.Y,
		0,
		hwnd,
		nil,
	)
	_ = windows.PostMessage(hwnd, windows.WM_USER, 0, 0)
}

func (wm *windowsMenu) destroy() {
	wm.detach()
	wm.dispatch(wm.destroyOnThread)
}

func (wm *windowsMenu) destroyOnThread() {
	if wm != nil && wm.hmenu != 0 {
		_ = windows.DestroyMenu(wm.hmenu)
		wm.hmenu = 0
	}
}

func (wm *windowsMenu) detach() {
	if wm == nil || wm.parent == nil {
		return
	}
	wm.parent.mu.Lock()
	if wm.parent.impl == wm {
		wm.parent.impl = nil
	}
	wm.parent.mu.Unlock()
}

type windowsMenuItem struct {
	parent *MenuItem
	menu   *windowsMenu
}

func (wmi *windowsMenuItem) setLabel(string) {
	wmi.menu.update()
}

func (*windowsMenuItem) setTooltip(string) {
	// Windows doesn't support tooltips in menus.
}

func (wmi *windowsMenuItem) setDisabled(disabled bool) {
	wmi.menu.dispatch(func() {
		if wmi.menu.hmenu == 0 {
			return
		}
		var mii windows.MENUITEMINFO
		mii.CbSize = windows.UINT(windows.SizeOf(mii))
		mii.FMask = windows.MIIM_STATE
		if disabled {
			mii.FState = windows.MFS_DISABLED | windows.MFS_GRAYED
		} else {
			mii.FState = windows.MFS_ENABLED
		}
		_ = windows.SetMenuItemInfo(wmi.menu.hmenu, windows.UINT(wmi.parent.id), 1, &mii)
	})
}

func (wmi *windowsMenuItem) setChecked(checked bool) {
	wmi.menu.dispatch(func() {
		if wmi.menu.hmenu == 0 {
			return
		}
		var mii windows.MENUITEMINFO
		mii.CbSize = windows.UINT(windows.SizeOf(mii))
		mii.FMask = windows.MIIM_STATE
		if checked {
			mii.FState = windows.MFS_CHECKED
		} else {
			mii.FState = windows.MFS_UNCHECKED
		}
		_ = windows.SetMenuItemInfo(wmi.menu.hmenu, windows.UINT(wmi.parent.id), 1, &mii)
	})
}

func (wmi *windowsMenuItem) setHidden(bool) {
	wmi.menu.update()
}

func (*windowsMenuItem) destroy() {
	// The containing menu owns native item resources.
}
