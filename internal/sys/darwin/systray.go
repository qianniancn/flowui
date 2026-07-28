//go:build darwin && !ios

package darwin

/*
#cgo CFLAGS: -mmacosx-version-min=10.13 -x objective-c
#cgo LDFLAGS: -framework Cocoa

#include <stdlib.h>
#include "Cocoa/Cocoa.h"
#include "systray.h"
*/
import "C"
import "unsafe"

// StatusItemController wraps NSStatusItem with event handling
// The Objective-C implementation is in systray.m

// ImageFromBytes creates an NSImage from PNG/JPEG data
func ImageFromBytes(bytes []byte) unsafe.Pointer {
	if len(bytes) == 0 {
		return nil
	}
	return unsafe.Pointer(C.imageFromBytes((*C.uchar)(&bytes[0]), C.int(len(bytes))))
}

// SystemTrayNew creates a new status bar item
func SystemTrayNew(id int64) unsafe.Pointer {
	return C.systemTrayNew(C.long(id))
}

// SystemTraySetLabel sets the label text
func SystemTraySetLabel(nsStatusItem unsafe.Pointer, label string) {
	cLabel := C.CString(label)
	C.systemTraySetLabel(nsStatusItem, cLabel)
	// Note: systemTraySetLabel frees cLabel
}

// SystemTraySetIcon sets the icon image
func SystemTraySetIcon(nsStatusItem, nsImage unsafe.Pointer, position int, isTemplate bool) {
	C.systemTraySetIcon(nsStatusItem, nsImage, C.int(position), C.bool(isTemplate))
}

// SystemTrayDestroy removes the status item
func SystemTrayDestroy(nsStatusItem unsafe.Pointer) {
	C.systemTrayDestroy(nsStatusItem)
}

// SystemTrayShow shows the status item
func SystemTrayShow(nsStatusItem unsafe.Pointer) {
	C.systemTrayShow(nsStatusItem)
}

// SystemTrayHide hides the status item
func SystemTrayHide(nsStatusItem unsafe.Pointer) {
	C.systemTrayHide(nsStatusItem)
}

// ShowMenu displays the menu
func ShowMenu(nsStatusItem, nsMenu unsafe.Pointer) {
	C.showMenu(nsStatusItem, nsMenu)
}

// SystemTraySetCachedMenu caches the menu for event handling
func SystemTraySetCachedMenu(nsStatusItem, nsMenu unsafe.Pointer) {
	C.systemTraySetCachedMenu(nsStatusItem, nsMenu)
}

// CreateNSMenu creates a new NSMenu
func CreateNSMenu(label string) unsafe.Pointer {
	var cLabel *C.char
	if label != "" {
		cLabel = C.CString(label)
	}
	return C.createNSMenu(cLabel)
	// Note: createNSMenu frees cLabel
}

// ReleaseNSMenu releases a menu
func ReleaseNSMenu(nsMenu unsafe.Pointer) {
	C.releaseNSMenu(nsMenu)
}

// AddMenuItem adds a menu item to a menu
func AddMenuItem(nsMenu, nsMenuItem unsafe.Pointer) {
	C.addMenuItem(nsMenu, nsMenuItem)
}

// AddMenuSeparator adds a separator
func AddMenuSeparator(nsMenu unsafe.Pointer) {
	C.addMenuSeparator(nsMenu)
}

// SetMenuItemSubmenu sets the submenu of a menu item
func SetMenuItemSubmenu(nsMenuItem, nsMenu unsafe.Pointer) {
	C.setMenuItemSubmenu(nsMenuItem, nsMenu)
}

// ClearMenu removes all items from a menu
func ClearMenu(nsMenu unsafe.Pointer) {
	C.clearMenu(nsMenu)
}
