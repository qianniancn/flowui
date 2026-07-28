package ui

import "github.com/qianniancn/flowui/internal/components/toast"

type ToastItem = toast.ToastItem
type ToastProviderWidget = toast.ToastProviderWidget
type ToastVariant = toast.ToastVariant
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

func Toast(key, title string) ToastItem {
	return toast.Toast(key, title)
}

func ToastProvider(key string, items []ToastItem) ToastProviderWidget {
	return toast.ToastProvider(key, items)
}
