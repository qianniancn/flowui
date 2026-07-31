package ui

import "github.com/qianniancn/flowui/internal/components/alert"

// AlertStatus selects the visual tone of an alert.
type AlertStatus = alert.Status

type AlertWidget = alert.Widget

const (
	AlertDefault = alert.StatusDefault
	AlertAccent  = alert.StatusAccent
	AlertSuccess = alert.StatusSuccess
	AlertWarning = alert.StatusWarning
	AlertDanger  = alert.StatusDanger
)

// Alert creates a status alert with title and description text.
func Alert(title, description string) AlertWidget {
	return alert.New(title, description)
}
