//go:build windows

package windows

import (
	"reflect"
	"syscall"
)

// Windows types
type (
	HWND    uintptr
	HMENU   uintptr
	HICON   uintptr
	HBITMAP uintptr
	WPARAM  uintptr
	LPARAM  uintptr
	LRESULT uintptr
	HANDLE  uintptr
	HMODULE uintptr
	UINT    uint32
	DWORD   uint32
	WORD    uint16
	BOOL    int32
)

// Window messages
const (
	WM_USER          = 0x0400
	WM_COMMAND       = 0x0111
	WM_DESTROY       = 0x0002
	WM_LBUTTONUP     = 0x0202
	WM_RBUTTONUP     = 0x0205
	WM_LBUTTONDBLCLK = 0x0203
	WM_RBUTTONDBLCLK = 0x0206
	WM_MOUSEMOVE     = 0x0200
)

// Notify icon messages
const (
	NIM_ADD        = 0x00000000
	NIM_MODIFY     = 0x00000001
	NIM_DELETE     = 0x00000002
	NIM_SETFOCUS   = 0x00000003
	NIM_SETVERSION = 0x00000004
)

// Notify icon flags
const (
	NIF_MESSAGE  = 0x00000001
	NIF_ICON     = 0x00000002
	NIF_TIP      = 0x00000004
	NIF_STATE    = 0x00000008
	NIF_INFO     = 0x00000010
	NIF_GUID     = 0x00000020
	NIF_REALTIME = 0x00000040
	NIF_SHOWTIP  = 0x00000080
)

// Menu flags
const (
	MF_STRING     = 0x00000000
	MF_SEPARATOR  = 0x00000800
	MF_CHECKED    = 0x00000008
	MF_UNCHECKED  = 0x00000000
	MF_DISABLED   = 0x00000002
	MF_ENABLED    = 0x00000000
	MF_GRAYED     = 0x00000001
	MF_POPUP      = 0x00000010
	MF_BYCOMMAND  = 0x00000000
	MF_BYPOSITION = 0x00000400
)

// Menu item info flags
const (
	MIIM_STATE      = 0x00000001
	MIIM_ID         = 0x00000002
	MIIM_SUBMENU    = 0x00000004
	MIIM_CHECKMARKS = 0x00000008
	MIIM_TYPE       = 0x00000010
	MIIM_DATA       = 0x00000020
	MIIM_STRING     = 0x00000040
	MIIM_BITMAP     = 0x00000080
	MIIM_FTYPE      = 0x00000100
)

// Menu item types
const (
	MFT_STRING       = 0x00000000
	MFT_BITMAP       = 0x00000004
	MFT_MENUBARBREAK = 0x00000020
	MFT_MENUBREAK    = 0x00000040
	MFT_OWNERDRAW    = 0x00000100
	MFT_RADIOCHECK   = 0x00000200
	MFT_SEPARATOR    = 0x00000800
	MFT_RIGHTORDER   = 0x00002000
	MFT_RIGHTJUSTIFY = 0x00004000
)

// Menu item states
const (
	MFS_CHECKED   = 0x00000008
	MFS_DEFAULT   = 0x00001000
	MFS_DISABLED  = 0x00000003
	MFS_ENABLED   = 0x00000000
	MFS_GRAYED    = 0x00000003
	MFS_HILITE    = 0x00000080
	MFS_UNCHECKED = 0x00000000
	MFS_UNHILITE  = 0x00000000
)

// Window styles
const (
	WS_OVERLAPPED       = 0x00000000
	WS_POPUP            = 0x80000000
	WS_CHILD            = 0x40000000
	WS_MINIMIZE         = 0x20000000
	WS_VISIBLE          = 0x10000000
	WS_DISABLED         = 0x08000000
	WS_CLIPSIBLINGS     = 0x04000000
	WS_CLIPCHILDREN     = 0x02000000
	WS_MAXIMIZE         = 0x01000000
	WS_CAPTION          = 0x00C00000
	WS_BORDER           = 0x00800000
	WS_DLGFRAME         = 0x00400000
	WS_VSCROLL          = 0x00200000
	WS_HSCROLL          = 0x00100000
	WS_SYSMENU          = 0x00080000
	WS_THICKFRAME       = 0x00040000
	WS_GROUP            = 0x00020000
	WS_TABSTOP          = 0x00010000
	WS_MINIMIZEBOX      = 0x00020000
	WS_MAXIMIZEBOX      = 0x00010000
	WS_OVERLAPPEDWINDOW = WS_OVERLAPPED | WS_CAPTION | WS_SYSMENU | WS_THICKFRAME | WS_MINIMIZEBOX | WS_MAXIMIZEBOX
)

// Track popup menu flags
const (
	TPM_LEFTBUTTON   = 0x0000
	TPM_RIGHTBUTTON  = 0x0002
	TPM_LEFTALIGN    = 0x0000
	TPM_CENTERALIGN  = 0x0004
	TPM_RIGHTALIGN   = 0x0008
	TPM_TOPALIGN     = 0x0000
	TPM_VCENTERALIGN = 0x0010
	TPM_BOTTOMALIGN  = 0x0020
	TPM_RETURNCMD    = 0x0100
	TPM_NONOTIFY     = 0x0080
)

// Image types
const (
	IMAGE_BITMAP = 0
	IMAGE_ICON   = 1
	IMAGE_CURSOR = 2
)

// LoadImage flags
const (
	LR_DEFAULTCOLOR     = 0x00000000
	LR_MONOCHROME       = 0x00000001
	LR_COLOR            = 0x00000002
	LR_COPYRETURNORG    = 0x00000004
	LR_COPYDELETEORG    = 0x00000008
	LR_LOADFROMFILE     = 0x00000010
	LR_LOADTRANSPARENT  = 0x00000020
	LR_DEFAULTSIZE      = 0x00000040
	LR_VGACOLOR         = 0x00000080
	LR_LOADMAP3DCOLORS  = 0x00001000
	LR_CREATEDIBSECTION = 0x00002000
	LR_COPYFROMRESOURCE = 0x00004000
	LR_SHARED           = 0x00008000
)

// NOTIFYICONDATA structure
type NOTIFYICONDATA struct {
	CbSize           DWORD
	HWnd             HWND
	UID              UINT
	UFlags           UINT
	UCallbackMessage UINT
	HIcon            HICON
	SzTip            [128]uint16
	DwState          DWORD
	DwStateMask      DWORD
	SzInfo           [256]uint16
	UTimeout         UINT
	SzInfoTitle      [64]uint16
	DwInfoFlags      DWORD
	GuidItem         syscall.GUID
	HBalloonIcon     HICON
}

// MENUITEMINFO structure
type MENUITEMINFO struct {
	CbSize        UINT
	FMask         UINT
	FType         UINT
	FState        UINT
	WID           UINT
	HSubMenu      HMENU
	HbmpChecked   HBITMAP
	HbmpUnchecked HBITMAP
	DwItemData    uintptr
	DwTypeData    *uint16
	Cch           UINT
	HbmpItem      HBITMAP
}

// POINT structure
type POINT struct {
	X int32
	Y int32
}

// RECT structure
type RECT struct {
	Left   int32
	Top    int32
	Right  int32
	Bottom int32
}

// WNDCLASSEX structure
type WNDCLASSEX struct {
	CbSize        UINT
	Style         UINT
	LpfnWndProc   uintptr
	CbClsExtra    int32
	CbWndExtra    int32
	HInstance     HMODULE
	HIcon         HICON
	HCursor       HANDLE
	HbrBackground HANDLE
	LpszMenuName  *uint16
	LpszClassName *uint16
	HIconSm       HICON
}

// MSG structure
type MSG struct {
	HWnd    HWND
	Message UINT
	WParam  WPARAM
	LParam  LPARAM
	Time    DWORD
	Pt      POINT
}

// Helper to convert size of struct to DWORD
func SizeOf(v any) DWORD {
	if v == nil {
		return 0
	}
	return DWORD(reflect.TypeOf(v).Size())
}

// UTF16PtrFromString converts a Go string to a UTF-16 pointer
func UTF16PtrFromString(s string) *uint16 {
	ptr, _ := syscall.UTF16PtrFromString(s)
	return ptr
}

// UTF16ToString converts a UTF-16 array to a Go string
func UTF16ToString(s []uint16) string {
	return syscall.UTF16ToString(s)
}
