package ui

import (
	"github.com/qianniancn/flowui/internal/components/workbench"
)

var (
	ErrWorkbenchEmptyKey           = workbench.ErrEmptyKey
	ErrWorkbenchDuplicateKey       = workbench.ErrDuplicateKey
	ErrWorkbenchUnknownGroup       = workbench.ErrUnknownGroup
	ErrWorkbenchUnknownTab         = workbench.ErrUnknownTab
	ErrWorkbenchInvalidPosition    = workbench.ErrInvalidPosition
	ErrWorkbenchUnsupportedVersion = workbench.ErrUnsupportedVersion
	ErrWorkbenchInvalidSnapshot    = workbench.ErrInvalidSnapshot
)

// WorkbenchTab is the model-side identity of a tab. It contains no editor
// implementation; render content remains a Tabs TabItem in the view layer.
type WorkbenchTab = workbench.Tab

// WorkbenchGroup is an ordered group of tabs hosted by a dock surface.
type WorkbenchGroup = workbench.Group

// WorkbenchChromeState stores visibility of the shell regions around content.
type WorkbenchChromeState = workbench.ChromeState

// WorkbenchState is the application-owned Workbench interaction model.
type WorkbenchState = workbench.State

// WorkbenchSnapshot is the versioned persistence contract for Workbench state.
type WorkbenchSnapshot = workbench.Snapshot

// WorkbenchGroupSnapshot stores one group's stable order and selection.
type WorkbenchGroupSnapshot = workbench.GroupSnapshot

// WorkbenchMigration remaps renamed groups/tabs while restoring a snapshot.
type WorkbenchMigration = workbench.Migration

// WorkbenchEventKind identifies a Workbench state transition.
type WorkbenchEventKind = workbench.EventKind

// WorkbenchEvent describes a state transition emitted by WorkbenchController.
type WorkbenchEvent = workbench.Event

// WorkbenchCommandID identifies a standard Workbench action.
type WorkbenchCommandID = workbench.CommandID

// WorkbenchController links the model to Tabs and Dock widgets.
type WorkbenchController = workbench.Controller

const (
	WorkbenchSnapshotVersion = workbench.SnapshotVersion
	WorkbenchGroupActivated  = workbench.EventGroupActivated
	WorkbenchTabActivated    = workbench.EventTabActivated
	WorkbenchTabAdded        = workbench.EventTabAdded
	WorkbenchTabClosed       = workbench.EventTabClosed
	WorkbenchTabReordered    = workbench.EventTabReordered
	WorkbenchTabMoved        = workbench.EventTabMoved
	WorkbenchGroupAdded      = workbench.EventGroupAdded
	WorkbenchGroupRemoved    = workbench.EventGroupRemoved
	WorkbenchGroupCollapsed  = workbench.EventGroupCollapsed
	WorkbenchDockChanged     = workbench.EventDockChanged
	WorkbenchChromeChanged   = workbench.EventChromeChanged
	WorkbenchLayoutRestored  = workbench.EventLayoutRestored

	WorkbenchCommandNextTab       = workbench.CommandNextTab
	WorkbenchCommandPreviousTab   = workbench.CommandPreviousTab
	WorkbenchCommandCloseTab      = workbench.CommandCloseTab
	WorkbenchCommandToggleSidebar = workbench.CommandToggleSidebar
	WorkbenchCommandTogglePanel   = workbench.CommandTogglePanel
	WorkbenchCommandToggleStatus  = workbench.CommandToggleStatus
)

// NewWorkbenchState creates a normalized Workbench model.
func NewWorkbenchState(groups []WorkbenchGroup) WorkbenchState {
	return workbench.NewState(groups)
}

// NewWorkbenchController creates an event-emitting Workbench adapter.
func NewWorkbenchController(state WorkbenchState) *WorkbenchController {
	return workbench.NewController(state)
}

// MarshalWorkbenchSnapshot encodes Workbench layout state for persistence.
func MarshalWorkbenchSnapshot(value WorkbenchSnapshot) ([]byte, error) {
	return workbench.MarshalSnapshot(value)
}

// UnmarshalWorkbenchSnapshot decodes a persisted Workbench layout document.
func UnmarshalWorkbenchSnapshot(data []byte) (WorkbenchSnapshot, error) {
	return workbench.UnmarshalSnapshot(data)
}

// ValidateWorkbenchSnapshot checks a snapshot before it is persisted or
// restored against an application model.
func ValidateWorkbenchSnapshot(value WorkbenchSnapshot) error {
	return workbench.ValidateSnapshot(value)
}
