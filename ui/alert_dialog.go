package ui

import "github.com/qianniancn/flowui/internal/components/alertdialog"

type AlertDialogWidget = alertdialog.Widget

// AlertDialogStatus selects the visual tone of an alert dialog.
type AlertDialogStatus = alertdialog.Status

// AlertDialogSize controls the dialog width.
type AlertDialogSize = alertdialog.Size

// AlertDialogPlacement controls where the dialog is placed in the window.
type AlertDialogPlacement = alertdialog.Placement

// AlertDialogBackdropVariant controls how the area behind the dialog is drawn.
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

// AlertDialog creates a modal alert dialog whose visibility is controlled by open.
func AlertDialog(key string, open bool, title, description string) AlertDialogWidget {
	return alertdialog.New(key, open, title, description)
}
