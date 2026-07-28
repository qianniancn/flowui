//go:build windows

package windows

import (
	"syscall"
	"unsafe"
)

var (
	shell32               = syscall.NewLazyDLL("shell32.dll")
	procShell_NotifyIconW = shell32.NewProc("Shell_NotifyIconW")
)

// Shell_NotifyIcon adds, modifies, or deletes an icon from the taskbar notification area
func Shell_NotifyIcon(message DWORD, data *NOTIFYICONDATA) error {
	ret, _, err := procShell_NotifyIconW.Call(
		uintptr(message),
		uintptr(unsafe.Pointer(data)),
	)
	if ret == 0 {
		return err
	}
	return nil
}
