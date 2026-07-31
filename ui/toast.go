package ui

import "github.com/qianniancn/flowui/internal/components/toast"

// ToastItem describes one transient notification.
type ToastItem = toast.ToastItem

type ToastProviderWidget = toast.ToastProviderWidget

// ToastVariant selects a notification's visual tone.
type ToastVariant = toast.ToastVariant

// ToastPlacement selects where notifications appear in the window.
type ToastPlacement = toast.ToastPlacement

const (
	ToastDefault = toast.ToastDefault
	ToastAccent  = toast.ToastAccent
	ToastSuccess = toast.ToastSuccess
	ToastWarning = toast.ToastWarning
	ToastDanger  = toast.ToastDanger

	ToastBottom      = toast.ToastBottom
	ToastBottomStart = toast.ToastBottomStart
	ToastBottomEnd   = toast.ToastBottomEnd
	ToastTop         = toast.ToastTop
	ToastTopStart    = toast.ToastTopStart
	ToastTopEnd      = toast.ToastTopEnd
)

// Toast creates one transient notification item.
func Toast(key, title string) ToastItem {
	return toast.Toast(key, title)
}

// ToastProvider renders and manages a list of transient notifications.
func ToastProvider(key string, items []ToastItem) ToastProviderWidget {
	return toast.ToastProvider(key, items)
}
