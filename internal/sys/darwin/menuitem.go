//go:build darwin && !ios

package darwin

/*
#cgo CFLAGS: -mmacosx-version-min=10.13 -x objective-c
#cgo LDFLAGS: -framework Cocoa

#include "systray.h"
*/
import "C"
import "unsafe"

// CreateNSMenuItem creates a new menu item
func CreateNSMenuItem(label string, id int64) unsafe.Pointer {
	var cLabel *C.char
	if label != "" {
		cLabel = C.CString(label)
	}
	return C.createNSMenuItem(cLabel, C.long(id))
	// Note: createNSMenuItem frees cLabel
}

// SetMenuItemChecked sets the checked state
func SetMenuItemChecked(nsMenuItem unsafe.Pointer, checked bool) {
	C.setMenuItemChecked(nsMenuItem, C.bool(checked))
}

// SetMenuItemEnabled sets the enabled state
func SetMenuItemEnabled(nsMenuItem unsafe.Pointer, enabled bool) {
	C.setMenuItemEnabled(nsMenuItem, C.bool(enabled))
}

// SetMenuItemHidden sets the hidden state
func SetMenuItemHidden(nsMenuItem unsafe.Pointer, hidden bool) {
	C.setMenuItemHidden(nsMenuItem, C.bool(hidden))
}

// ReleaseNSMenuItem releases a menu item
func ReleaseNSMenuItem(nsMenuItem unsafe.Pointer) {
	C.releaseNSMenuItem(nsMenuItem)
}
