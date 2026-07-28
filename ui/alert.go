package ui

import "github.com/qianniancn/flowui/internal/components/alert"

type AlertStatus = alert.Status
type AlertWidget = alert.Widget

const (
	AlertDefault = alert.StatusDefault
	AlertAccent  = alert.StatusAccent
	AlertSuccess = alert.StatusSuccess
	AlertWarning = alert.StatusWarning
	AlertDanger  = alert.StatusDanger
)

func Alert(title, description string) AlertWidget {
	return alert.New(title, description)
}
