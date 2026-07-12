package ui

import "github.com/qianniancn/FlowUI/internal/components/alertdialog"

type AlertDialogWidget = alertdialog.Widget
type AlertDialogStatus = alertdialog.Status
type AlertDialogSize = alertdialog.Size
type AlertDialogPlacement = alertdialog.Placement
type AlertDialogBackdropVariant = alertdialog.BackdropVariant

const (
	AlertDialogDefault = alertdialog.StatusDefault
	AlertDialogAccent  = alertdialog.StatusAccent
	AlertDialogSuccess = alertdialog.StatusSuccess
	AlertDialogWarning = alertdialog.StatusWarning
	AlertDialogDanger  = alertdialog.StatusDanger

	AlertDialogMedium = alertdialog.SizeMedium
	AlertDialogXSmall = alertdialog.SizeXSmall
	AlertDialogSmall  = alertdialog.SizeSmall
	AlertDialogLarge  = alertdialog.SizeLarge
	AlertDialogCover  = alertdialog.SizeCover

	AlertDialogAuto   = alertdialog.PlacementAuto
	AlertDialogTop    = alertdialog.PlacementTop
	AlertDialogCenter = alertdialog.PlacementCenter
	AlertDialogBottom = alertdialog.PlacementBottom

	AlertDialogBackdropOpaque      = alertdialog.BackdropOpaque
	AlertDialogBackdropBlur        = alertdialog.BackdropBlur
	AlertDialogBackdropTransparent = alertdialog.BackdropTransparent
)

func AlertDialog(key string, open bool, title, description string) AlertDialogWidget {
	return alertdialog.New(key, open, title, description)
}
