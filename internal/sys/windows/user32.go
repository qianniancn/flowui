//go:build windows

package windows

import (
	"syscall"
	"unsafe"
)

var (
	user32                       = syscall.NewLazyDLL("user32.dll")
	procRegisterClassExW         = user32.NewProc("RegisterClassExW")
	procRegisterWindowMessageW   = user32.NewProc("RegisterWindowMessageW")
	procUnregisterClassW         = user32.NewProc("UnregisterClassW")
	procCreateWindowExW          = user32.NewProc("CreateWindowExW")
	procDefWindowProcW           = user32.NewProc("DefWindowProcW")
	procDestroyWindow            = user32.NewProc("DestroyWindow")
	procPostQuitMessage          = user32.NewProc("PostQuitMessage")
	procGetMessageW              = user32.NewProc("GetMessageW")
	procTranslateMessage         = user32.NewProc("TranslateMessage")
	procDispatchMessageW         = user32.NewProc("DispatchMessageW")
	procPostMessageW             = user32.NewProc("PostMessageW")
	procLoadIconW                = user32.NewProc("LoadIconW")
	procLoadCursorW              = user32.NewProc("LoadCursorW")
	procCreateIconFromResourceEx = user32.NewProc("CreateIconFromResourceEx")
	procDestroyIcon              = user32.NewProc("DestroyIcon")
	procCreatePopupMenu          = user32.NewProc("CreatePopupMenu")
	procDestroyMenu              = user32.NewProc("DestroyMenu")
	procAppendMenuW              = user32.NewProc("AppendMenuW")
	procInsertMenuItemW          = user32.NewProc("InsertMenuItemW")
	procSetMenuItemInfoW         = user32.NewProc("SetMenuItemInfoW")
	procGetMenuItemInfoW         = user32.NewProc("GetMenuItemInfoW")
	procTrackPopupMenu           = user32.NewProc("TrackPopupMenu")
	procSetForegroundWindow      = user32.NewProc("SetForegroundWindow")
	procGetCursorPos             = user32.NewProc("GetCursorPos")
	procGetSystemMetrics         = user32.NewProc("GetSystemMetrics")
)

// RegisterWindowMessage registers a system-wide window message.
func RegisterWindowMessage(name string) (UINT, error) {
	namePtr, err := syscall.UTF16PtrFromString(name)
	if err != nil {
		return 0, err
	}
	ret, _, callErr := procRegisterWindowMessageW.Call(uintptr(unsafe.Pointer(namePtr)))
	if ret == 0 {
		return 0, callErr
	}
	return UINT(ret), nil
}

// RegisterClassEx registers a window class
func RegisterClassEx(wc *WNDCLASSEX) (uint16, error) {
	ret, _, err := procRegisterClassExW.Call(uintptr(unsafe.Pointer(wc)))
	if ret == 0 {
		return 0, err
	}
	return uint16(ret), nil
}

// UnregisterClass unregisters a window class.
func UnregisterClass(className *uint16, instance HMODULE) error {
	ret, _, err := procUnregisterClassW.Call(
		uintptr(unsafe.Pointer(className)),
		uintptr(instance),
	)
	if ret == 0 {
		return err
	}
	return nil
}

// CreateWindowEx creates a window
func CreateWindowEx(
	exStyle DWORD,
	className, windowName *uint16,
	style DWORD,
	x, y, width, height int32,
	parent HWND,
	menu HMENU,
	instance HMODULE,
	param unsafe.Pointer,
) (HWND, error) {
	ret, _, err := procCreateWindowExW.Call(
		uintptr(exStyle),
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(windowName)),
		uintptr(style),
		uintptr(x),
		uintptr(y),
		uintptr(width),
		uintptr(height),
		uintptr(parent),
		uintptr(menu),
		uintptr(instance),
		uintptr(param),
	)
	if ret == 0 {
		return 0, err
	}
	return HWND(ret), nil
}

// DefWindowProc calls the default window procedure
func DefWindowProc(hwnd HWND, msg UINT, wParam WPARAM, lParam LPARAM) LRESULT {
	ret, _, _ := procDefWindowProcW.Call(
		uintptr(hwnd),
		uintptr(msg),
		uintptr(wParam),
		uintptr(lParam),
	)
	return LRESULT(ret)
}

// DestroyWindow destroys a window
func DestroyWindow(hwnd HWND) error {
	ret, _, err := procDestroyWindow.Call(uintptr(hwnd))
	if ret == 0 {
		return err
	}
	return nil
}

// PostQuitMessage posts a quit message
func PostQuitMessage(exitCode int32) {
	procPostQuitMessage.Call(uintptr(exitCode))
}

// GetMessage retrieves a message from the message queue
func GetMessage(msg *MSG, hwnd HWND, msgFilterMin, msgFilterMax UINT) BOOL {
	ret, _, _ := procGetMessageW.Call(
		uintptr(unsafe.Pointer(msg)),
		uintptr(hwnd),
		uintptr(msgFilterMin),
		uintptr(msgFilterMax),
	)
	return BOOL(ret)
}

// TranslateMessage translates virtual-key messages
func TranslateMessage(msg *MSG) BOOL {
	ret, _, _ := procTranslateMessage.Call(uintptr(unsafe.Pointer(msg)))
	return BOOL(ret)
}

// DispatchMessage dispatches a message to a window procedure
func DispatchMessage(msg *MSG) LRESULT {
	ret, _, _ := procDispatchMessageW.Call(uintptr(unsafe.Pointer(msg)))
	return LRESULT(ret)
}

// PostMessage posts a message to a window
func PostMessage(hwnd HWND, msg UINT, wParam WPARAM, lParam LPARAM) error {
	ret, _, err := procPostMessageW.Call(
		uintptr(hwnd),
		uintptr(msg),
		uintptr(wParam),
		uintptr(lParam),
	)
	if ret == 0 {
		return err
	}
	return nil
}

// LoadIcon loads an icon
func LoadIcon(instance HMODULE, iconName *uint16) (HICON, error) {
	ret, _, err := procLoadIconW.Call(uintptr(instance), uintptr(unsafe.Pointer(iconName)))
	if ret == 0 {
		return 0, err
	}
	return HICON(ret), nil
}

// LoadCursor loads a cursor
func LoadCursor(instance HMODULE, cursorName *uint16) (HANDLE, error) {
	ret, _, err := procLoadCursorW.Call(uintptr(instance), uintptr(unsafe.Pointer(cursorName)))
	if ret == 0 {
		return 0, err
	}
	return HANDLE(ret), nil
}

// CreateIconFromResourceEx creates an icon from resource data
func CreateIconFromResourceEx(presbits []byte, dwResSize DWORD, fIcon BOOL, dwVer DWORD, cxDesired, cyDesired int32, flags UINT) (HICON, error) {
	ret, _, err := procCreateIconFromResourceEx.Call(
		uintptr(unsafe.Pointer(&presbits[0])),
		uintptr(dwResSize),
		uintptr(fIcon),
		uintptr(dwVer),
		uintptr(cxDesired),
		uintptr(cyDesired),
		uintptr(flags),
	)
	if ret == 0 {
		return 0, err
	}
	return HICON(ret), nil
}

// DestroyIcon destroys an icon
func DestroyIcon(icon HICON) error {
	ret, _, err := procDestroyIcon.Call(uintptr(icon))
	if ret == 0 {
		return err
	}
	return nil
}

// CreatePopupMenu creates a popup menu
func CreatePopupMenu() (HMENU, error) {
	ret, _, err := procCreatePopupMenu.Call()
	if ret == 0 {
		return 0, err
	}
	return HMENU(ret), nil
}

// DestroyMenu destroys a menu
func DestroyMenu(menu HMENU) error {
	ret, _, err := procDestroyMenu.Call(uintptr(menu))
	if ret == 0 {
		return err
	}
	return nil
}

// AppendMenu appends a menu item
func AppendMenu(menu HMENU, flags UINT, idNewItem uintptr, lpNewItem *uint16) error {
	ret, _, err := procAppendMenuW.Call(
		uintptr(menu),
		uintptr(flags),
		idNewItem,
		uintptr(unsafe.Pointer(lpNewItem)),
	)
	if ret == 0 {
		return err
	}
	return nil
}

// InsertMenuItem inserts a menu item
func InsertMenuItem(menu HMENU, item UINT, fByPosition BOOL, lpmii *MENUITEMINFO) error {
	ret, _, err := procInsertMenuItemW.Call(
		uintptr(menu),
		uintptr(item),
		uintptr(fByPosition),
		uintptr(unsafe.Pointer(lpmii)),
	)
	if ret == 0 {
		return err
	}
	return nil
}

// SetMenuItemInfo sets menu item information
func SetMenuItemInfo(menu HMENU, item UINT, fByPosition BOOL, lpmii *MENUITEMINFO) error {
	ret, _, err := procSetMenuItemInfoW.Call(
		uintptr(menu),
		uintptr(item),
		uintptr(fByPosition),
		uintptr(unsafe.Pointer(lpmii)),
	)
	if ret == 0 {
		return err
	}
	return nil
}

// GetMenuItemInfo gets menu item information
func GetMenuItemInfo(menu HMENU, item UINT, fByPosition BOOL, lpmii *MENUITEMINFO) error {
	ret, _, err := procGetMenuItemInfoW.Call(
		uintptr(menu),
		uintptr(item),
		uintptr(fByPosition),
		uintptr(unsafe.Pointer(lpmii)),
	)
	if ret == 0 {
		return err
	}
	return nil
}

// TrackPopupMenu displays a shortcut menu
func TrackPopupMenu(menu HMENU, flags UINT, x, y int32, reserved int32, hwnd HWND, prcRect *RECT) BOOL {
	ret, _, _ := procTrackPopupMenu.Call(
		uintptr(menu),
		uintptr(flags),
		uintptr(x),
		uintptr(y),
		uintptr(reserved),
		uintptr(hwnd),
		uintptr(unsafe.Pointer(prcRect)),
	)
	return BOOL(ret)
}

// SetForegroundWindow brings a window to the foreground
func SetForegroundWindow(hwnd HWND) BOOL {
	ret, _, _ := procSetForegroundWindow.Call(uintptr(hwnd))
	return BOOL(ret)
}

// GetCursorPos retrieves the cursor position
func GetCursorPos(pt *POINT) error {
	ret, _, err := procGetCursorPos.Call(uintptr(unsafe.Pointer(pt)))
	if ret == 0 {
		return err
	}
	return nil
}

// GetSystemMetrics retrieves system metrics
func GetSystemMetrics(index int32) int32 {
	ret, _, _ := procGetSystemMetrics.Call(uintptr(index))
	return int32(ret)
}

// GetModuleHandle gets the module handle
func GetModuleHandle() (HMODULE, error) {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	procGetModuleHandleW := kernel32.NewProc("GetModuleHandleW")
	ret, _, err := procGetModuleHandleW.Call(0)
	if ret == 0 {
		return 0, err
	}
	return HMODULE(ret), nil
}
