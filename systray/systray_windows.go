//go:build windows

package systray

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/png"
	"runtime"
	"sync"
	"syscall"

	"github.com/qianniancn/flowui/internal/sys/windows"
)

const (
	wmSystray        = windows.WM_USER + 1
	wmSystrayTask    = windows.WM_USER + 2
	wmSystrayDestroy = windows.WM_USER + 3
)

type windowsSystemTray struct {
	parent *SystemTray
	uid    uint32

	stateMu sync.Mutex
	hwnd    windows.HWND
	closing bool
	tasks   []func()

	// Native resources below are only accessed from the window thread.
	menu      *windowsMenu
	icon      windows.HICON
	iconAdded bool
	visible   bool

	taskbarCreatedMessage windows.UINT
}

var shellNotifyIcon = windows.Shell_NotifyIcon

// newSystemTrayImpl creates a Windows system tray implementation.
func newSystemTrayImpl(parent *SystemTray) systemTrayImpl {
	return &windowsSystemTray{
		parent:  parent,
		uid:     uint32(parent.id),
		visible: true,
	}
}

func (w *windowsSystemTray) setLabel(string) {
	// Windows doesn't support labels in the notification area.
}

func (w *windowsSystemTray) setTooltip(tooltip string) {
	w.post(func() { w.setTooltipOnThread(tooltip) })
}

func (w *windowsSystemTray) setTooltipOnThread(tooltip string) {
	hwnd := w.windowHandle()
	if hwnd == 0 || !w.iconAdded {
		return
	}
	var nid windows.NOTIFYICONDATA
	nid.CbSize = windows.SizeOf(nid)
	nid.HWnd = hwnd
	nid.UID = windows.UINT(w.uid)
	nid.UFlags = windows.NIF_TIP
	utf16Tooltip, _ := syscall.UTF16FromString(tooltip)
	copy(nid.SzTip[:], utf16Tooltip)
	if err := shellNotifyIcon(windows.NIM_MODIFY, &nid); err != nil {
		w.parent.reportError(fmt.Errorf("systray: update Windows tooltip: %w", err))
	}
}

func (w *windowsSystemTray) setIcon(iconData []byte) {
	data := append([]byte(nil), iconData...)
	w.post(func() { w.setIconOnThread(data) })
}

func (w *windowsSystemTray) setIconOnThread(iconData []byte) {
	icon, err := createIconFromPNG(iconData)
	if err != nil {
		w.parent.reportError(fmt.Errorf("systray: decode Windows icon: %w", err))
		return
	}
	if w.icon != 0 {
		_ = windows.DestroyIcon(w.icon)
	}
	w.icon = icon

	hwnd := w.windowHandle()
	if hwnd == 0 || !w.iconAdded {
		return
	}
	var nid windows.NOTIFYICONDATA
	nid.CbSize = windows.SizeOf(nid)
	nid.HWnd = hwnd
	nid.UID = windows.UINT(w.uid)
	nid.UFlags = windows.NIF_ICON
	nid.HIcon = icon
	if err := shellNotifyIcon(windows.NIM_MODIFY, &nid); err != nil {
		w.parent.reportError(fmt.Errorf("systray: update Windows icon: %w", err))
	}
}

func (w *windowsSystemTray) setMenu(menu *Menu) {
	w.post(func() { w.setMenuOnThread(menu) })
}

func (w *windowsSystemTray) setMenuOnThread(menu *Menu) {
	if w.menu != nil {
		w.menu.detach()
		w.menu.destroyOnThread()
		w.menu = nil
	}
	if menu == nil {
		return
	}
	menu.processRadioGroups()
	w.menu = newWindowsMenu(menu, w)
	menu.mu.Lock()
	menu.impl = w.menu
	menu.mu.Unlock()
}

func (w *windowsSystemTray) run() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	hInstance, err := windows.GetModuleHandle()
	if err != nil {
		w.parent.fail(fmt.Errorf("systray: get Windows module handle: %w", err))
		return
	}
	taskbarCreatedMessage, err := windows.RegisterWindowMessage("TaskbarCreated")
	if err != nil {
		w.parent.fail(fmt.Errorf("systray: register TaskbarCreated message: %w", err))
		return
	}
	w.taskbarCreatedMessage = taskbarCreatedMessage
	className := windows.UTF16PtrFromString(fmt.Sprintf("FlowUISystrayClass_%d", w.uid))
	wc := windows.WNDCLASSEX{
		CbSize:        windows.UINT(windows.SizeOf(windows.WNDCLASSEX{})),
		LpfnWndProc:   syscall.NewCallback(w.wndProc),
		HInstance:     hInstance,
		LpszClassName: className,
	}
	if _, err := windows.RegisterClassEx(&wc); err != nil {
		w.parent.fail(fmt.Errorf("systray: register Windows tray class: %w", err))
		return
	}
	defer func() { _ = windows.UnregisterClass(className, hInstance) }()

	hwnd, err := windows.CreateWindowEx(
		0,
		className,
		windows.UTF16PtrFromString("FlowUISystray"),
		0,
		0, 0, 0, 0,
		0, 0, hInstance, nil,
	)
	if err != nil {
		w.parent.fail(fmt.Errorf("systray: create Windows tray window: %w", err))
		return
	}
	w.stateMu.Lock()
	w.hwnd = hwnd
	closing := w.closing
	w.stateMu.Unlock()
	if closing {
		w.cleanupOnThread(hwnd)
		w.runMessageLoop()
		return
	}

	w.parent.mu.Lock()
	iconData := append([]byte(nil), w.parent.icon...)
	menu := w.parent.menu
	w.parent.mu.Unlock()

	if len(iconData) > 0 {
		icon, iconErr := createIconFromPNG(iconData)
		if iconErr != nil {
			w.parent.fail(fmt.Errorf("systray: decode initial Windows icon: %w", iconErr))
			w.cleanupOnThread(hwnd)
			w.runMessageLoop()
			return
		}
		w.icon = icon
	}
	if menu != nil {
		w.setMenuOnThread(menu)
	}
	if err := w.addIconOnThread(hwnd); err != nil {
		w.parent.fail(fmt.Errorf("systray: add Windows tray icon: %w", err))
		w.cleanupOnThread(hwnd)
		w.runMessageLoop()
		return
	}
	w.parent.markReady()
	w.drainTasksOnThread()

	w.runMessageLoop()
	w.cleanupOnThread(hwnd)
	w.parent.finish()
}

func (w *windowsSystemTray) addIconOnThread(hwnd windows.HWND) error {
	var nid windows.NOTIFYICONDATA
	nid.CbSize = windows.SizeOf(nid)
	nid.HWnd = hwnd
	nid.UID = windows.UINT(w.uid)
	nid.UFlags = windows.NIF_MESSAGE | windows.NIF_TIP
	nid.UCallbackMessage = windows.UINT(wmSystray)
	if w.icon != 0 {
		nid.UFlags |= windows.NIF_ICON
		nid.HIcon = w.icon
	}
	w.parent.mu.Lock()
	tooltip := w.parent.tooltip
	w.parent.mu.Unlock()
	if tooltip != "" {
		utf16Tooltip, _ := syscall.UTF16FromString(tooltip)
		copy(nid.SzTip[:], utf16Tooltip)
	}
	if !w.visible {
		nid.UFlags |= windows.NIF_STATE
		nid.DwStateMask = 1 // NIS_HIDDEN
		nid.DwState = 1
	}
	if err := shellNotifyIcon(windows.NIM_ADD, &nid); err != nil {
		w.iconAdded = false
		return err
	}
	w.iconAdded = true
	return nil
}

func (w *windowsSystemTray) runMessageLoop() {
	var msg windows.MSG
	for windows.GetMessage(&msg, 0, 0, 0) > 0 {
		windows.TranslateMessage(&msg)
		windows.DispatchMessage(&msg)
	}
}

func (w *windowsSystemTray) post(task func()) {
	if task == nil {
		return
	}
	w.stateMu.Lock()
	if w.closing {
		w.stateMu.Unlock()
		return
	}
	w.tasks = append(w.tasks, task)
	hwnd := w.hwnd
	w.stateMu.Unlock()
	if hwnd != 0 {
		_ = windows.PostMessage(hwnd, wmSystrayTask, 0, 0)
	}
}

func (w *windowsSystemTray) drainTasksOnThread() {
	w.stateMu.Lock()
	tasks := w.tasks
	w.tasks = nil
	w.stateMu.Unlock()
	for _, task := range tasks {
		task()
	}
}

func (w *windowsSystemTray) windowHandle() windows.HWND {
	w.stateMu.Lock()
	defer w.stateMu.Unlock()
	return w.hwnd
}

func (w *windowsSystemTray) destroy() {
	w.stateMu.Lock()
	if w.closing {
		w.stateMu.Unlock()
		return
	}
	w.closing = true
	w.tasks = nil
	hwnd := w.hwnd
	w.stateMu.Unlock()
	if hwnd != 0 {
		_ = windows.PostMessage(hwnd, wmSystrayDestroy, 0, 0)
	}
}

func (w *windowsSystemTray) cleanupOnThread(hwnd windows.HWND) {
	w.stateMu.Lock()
	w.closing = true
	w.tasks = nil
	ownsWindow := w.hwnd == hwnd
	if ownsWindow {
		w.hwnd = 0
	}
	w.stateMu.Unlock()

	if w.iconAdded {
		var nid windows.NOTIFYICONDATA
		nid.CbSize = windows.SizeOf(nid)
		nid.HWnd = hwnd
		nid.UID = windows.UINT(w.uid)
		_ = shellNotifyIcon(windows.NIM_DELETE, &nid)
		w.iconAdded = false
	}
	if w.icon != 0 {
		_ = windows.DestroyIcon(w.icon)
		w.icon = 0
	}
	if w.menu != nil {
		w.menu.detach()
		w.menu.destroyOnThread()
		w.menu = nil
	}

	if ownsWindow && hwnd != 0 {
		_ = windows.DestroyWindow(hwnd)
	}
}

func (w *windowsSystemTray) show() {
	w.post(func() { w.setVisibleOnThread(true) })
}

func (w *windowsSystemTray) hide() {
	w.post(func() { w.setVisibleOnThread(false) })
}

func (w *windowsSystemTray) setVisibleOnThread(visible bool) {
	w.visible = visible
	hwnd := w.windowHandle()
	if hwnd == 0 || !w.iconAdded {
		return
	}
	var nid windows.NOTIFYICONDATA
	nid.CbSize = windows.SizeOf(nid)
	nid.HWnd = hwnd
	nid.UID = windows.UINT(w.uid)
	nid.UFlags = windows.NIF_STATE
	nid.DwStateMask = 1 // NIS_HIDDEN
	if !visible {
		nid.DwState = 1
	}
	if err := shellNotifyIcon(windows.NIM_MODIFY, &nid); err != nil {
		w.parent.reportError(fmt.Errorf("systray: update Windows tray visibility: %w", err))
	}
}

func (w *windowsSystemTray) wndProc(hwnd windows.HWND, msg windows.UINT, wParam windows.WPARAM, lParam windows.LPARAM) uintptr {
	if msg == w.taskbarCreatedMessage && w.taskbarCreatedMessage != 0 {
		if err := w.addIconOnThread(hwnd); err != nil {
			w.parent.reportError(fmt.Errorf("systray: restore Windows tray icon after Explorer restart: %w", err))
		}
		return 0
	}
	switch msg {
	case windows.WM_DESTROY:
		windows.PostQuitMessage(0)
		return 0
	case wmSystrayTask:
		w.drainTasksOnThread()
		return 0
	case wmSystrayDestroy:
		w.cleanupOnThread(hwnd)
		return 0
	case wmSystray:
		switch lParam {
		case windows.WM_LBUTTONUP:
			w.parent.mu.Lock()
			handler := w.parent.clickHandler
			w.parent.mu.Unlock()
			if handler != nil {
				go handler()
			}
		case windows.WM_RBUTTONUP:
			w.parent.mu.Lock()
			handler := w.parent.rightClickHandler
			w.parent.mu.Unlock()
			if handler != nil {
				go handler()
			} else if w.menu != nil {
				w.menu.showOnThread(hwnd)
			}
		case windows.WM_LBUTTONDBLCLK:
			w.parent.mu.Lock()
			handler := w.parent.doubleClickHandler
			w.parent.mu.Unlock()
			if handler != nil {
				go handler()
			}
		case windows.WM_MOUSEMOVE:
			w.parent.mu.Lock()
			handler := w.parent.mouseEnterHandler
			w.parent.mu.Unlock()
			if handler != nil {
				go handler()
			}
		}
		return 0
	case windows.WM_COMMAND:
		itemID := uint(loword(uint32(wParam)))
		if item := GetMenuItemByID(itemID); item != nil {
			go item.Click()
		}
		return 0
	}
	return uintptr(windows.DefWindowProc(hwnd, msg, wParam, lParam))
}

// createIconFromPNG creates a Windows icon from PNG data.
func createIconFromPNG(data []byte) (windows.HICON, error) {
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return 0, err
	}
	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			rgba.Set(x, y, img.At(x, y))
		}
	}
	width := bounds.Dx()
	height := bounds.Dy()
	buf := new(bytes.Buffer)
	_ = binary.Write(buf, binary.LittleEndian, uint16(0))
	_ = binary.Write(buf, binary.LittleEndian, uint16(1))
	_ = binary.Write(buf, binary.LittleEndian, uint16(1))
	_ = binary.Write(buf, binary.LittleEndian, uint8(width))
	_ = binary.Write(buf, binary.LittleEndian, uint8(height))
	_ = binary.Write(buf, binary.LittleEndian, uint8(0))
	_ = binary.Write(buf, binary.LittleEndian, uint8(0))
	_ = binary.Write(buf, binary.LittleEndian, uint16(1))
	_ = binary.Write(buf, binary.LittleEndian, uint16(32))
	_ = binary.Write(buf, binary.LittleEndian, uint32(len(data)))
	_ = binary.Write(buf, binary.LittleEndian, uint32(22))
	buf.Write(data)
	iconData := buf.Bytes()
	return windows.CreateIconFromResourceEx(
		iconData[22:],
		windows.DWORD(len(data)),
		1,
		0x00030000,
		int32(width),
		int32(height),
		windows.LR_DEFAULTCOLOR,
	)
}

func loword(l uint32) uint16 {
	return uint16(l & 0xFFFF)
}
