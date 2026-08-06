package ui

import (
	"encoding/json"

	"github.com/qianniancn/flowui/internal/components/dock"
)

var ErrDockInvalidSnapshot = dock.ErrInvalidSnapshot

// DockOrientation controls the axis of a DockSplit node.
type DockOrientation = dock.Orientation

// DockCollapseState stores controlled collapsed state for one split node.
type DockCollapseState = dock.CollapseState

// DockLayoutSnapshot is the serializable geometry state of a DockLayout.
type DockLayoutSnapshot = dock.Snapshot

// DockSnapshotVersion is the current serializable DockLayout schema version.
const DockSnapshotVersion = dock.SnapshotVersion

// DockNode describes one panel or split in a DockLayout tree.
type DockNode = dock.Node

// DockLayoutWidget lays out a recursively composable workbench tree.
type DockLayoutWidget = dock.DockLayoutWidget

const (
	DockHorizontal DockOrientation = dock.Horizontal
	DockVertical   DockOrientation = dock.Vertical
)

// DockPanel creates a leaf node containing content.
func DockPanel(key string, content Widget) DockNode {
	return dock.Panel(key, content)
}

// DockSplit creates a recursively composable split node.
func DockSplit(key string, orientation DockOrientation, first, second DockNode) DockNode {
	return dock.Split(key, orientation, first, second)
}

// DockLayout creates a workbench layout from a declarative node tree.
func DockLayout(key string, root DockNode) DockLayoutWidget {
	return dock.New(key, root)
}

// MarshalDockLayoutSnapshot encodes dock geometry for persistence.
func MarshalDockLayoutSnapshot(value DockLayoutSnapshot) ([]byte, error) {
	return json.Marshal(value)
}

// UnmarshalDockLayoutSnapshot decodes persisted dock geometry.
func UnmarshalDockLayoutSnapshot(data []byte) (DockLayoutSnapshot, error) {
	var value DockLayoutSnapshot
	if err := json.Unmarshal(data, &value); err != nil {
		return DockLayoutSnapshot{}, err
	}
	return value, nil
}
