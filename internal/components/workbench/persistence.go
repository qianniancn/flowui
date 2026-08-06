package workbench

import (
	"encoding/json"
	"errors"
	"fmt"
)

var ErrInvalidSnapshot = errors.New("flowui: invalid workbench snapshot")

type groupSnapshotJSON struct {
	Key       string   `json:"key"`
	TabOrder  []string `json:"tabOrder,omitempty"`
	ActiveKey string   `json:"activeKey,omitempty"`
	Collapsed bool     `json:"collapsed,omitempty"`
}

type chromeSnapshotJSON struct {
	SidebarVisible     bool `json:"sidebarVisible"`
	BottomPanelVisible bool `json:"bottomPanelVisible"`
	StatusBarVisible   bool `json:"statusBarVisible"`
}

type snapshotJSON struct {
	Version      uint16              `json:"version,omitempty"`
	ActiveGroup  string              `json:"activeGroup,omitempty"`
	FocusedGroup string              `json:"focusedGroup,omitempty"`
	Groups       []groupSnapshotJSON `json:"groups,omitempty"`
	Dock         any                 `json:"dock,omitempty"`
	Chrome       *chromeSnapshotJSON `json:"chrome,omitempty"`
}

// Validate checks stable-key and version invariants before a snapshot is
// written or applied. Unknown keys are intentionally allowed here because
// Restore may resolve them through Migration against a newer model.
func (s Snapshot) Validate() error {
	if s.Version > SnapshotVersion {
		return fmt.Errorf("%w: unsupported version %d", ErrInvalidSnapshot, s.Version)
	}
	if err := s.Dock.Validate(); err != nil {
		return fmt.Errorf("%w: dock: %v", ErrInvalidSnapshot, err)
	}
	groups := make(map[string]struct{}, len(s.Groups))
	for _, group := range s.Groups {
		if group.Key == "" {
			return fmt.Errorf("%w: empty group key", ErrInvalidSnapshot)
		}
		if _, exists := groups[group.Key]; exists {
			return fmt.Errorf("%w: duplicate group key %q", ErrInvalidSnapshot, group.Key)
		}
		groups[group.Key] = struct{}{}
		keys := make(map[string]struct{}, len(group.TabOrder))
		for _, key := range group.TabOrder {
			if key == "" {
				return fmt.Errorf("%w: empty tab key in group %q", ErrInvalidSnapshot, group.Key)
			}
			if _, exists := keys[key]; exists {
				return fmt.Errorf("%w: duplicate tab key %q in group %q", ErrInvalidSnapshot, key, group.Key)
			}
			keys[key] = struct{}{}
		}
	}
	return nil
}

// MarshalJSON writes a stable, versioned Workbench layout document. It stores
// only group/tab identities and dock geometry; editor buffers are never part
// of this payload.
func (s Snapshot) MarshalJSON() ([]byte, error) {
	if err := s.Validate(); err != nil {
		return nil, err
	}
	if s.Version == 0 {
		s.Version = SnapshotVersion
	}
	value := snapshotJSON{
		Version:      s.Version,
		ActiveGroup:  s.ActiveGroup,
		FocusedGroup: s.FocusedGroup,
		Groups:       make([]groupSnapshotJSON, len(s.Groups)),
		Dock:         s.Dock,
		Chrome: &chromeSnapshotJSON{
			SidebarVisible:     s.Chrome.SidebarVisible,
			BottomPanelVisible: s.Chrome.BottomPanelVisible,
			StatusBarVisible:   s.Chrome.StatusBarVisible,
		},
	}
	for index, group := range s.Groups {
		value.Groups[index] = groupSnapshotJSON{
			Key:       group.Key,
			TabOrder:  append([]string(nil), group.TabOrder...),
			ActiveKey: group.ActiveKey,
			Collapsed: group.Collapsed,
		}
	}
	return json.Marshal(value)
}

// UnmarshalJSON reads a current or legacy snapshot. It does not apply model
// migrations; call State.Restore with Migration after loading the document.
func (s *Snapshot) UnmarshalJSON(data []byte) error {
	if s == nil {
		return errors.New("flowui: nil workbench snapshot")
	}
	var value struct {
		Version      uint16              `json:"version"`
		ActiveGroup  string              `json:"activeGroup"`
		FocusedGroup string              `json:"focusedGroup"`
		Groups       []groupSnapshotJSON `json:"groups"`
		Dock         json.RawMessage     `json:"dock"`
		Chrome       *chromeSnapshotJSON `json:"chrome"`
	}
	if err := json.Unmarshal(data, &value); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidSnapshot, err)
	}
	decoded := Snapshot{
		Version:      value.Version,
		ActiveGroup:  value.ActiveGroup,
		FocusedGroup: value.FocusedGroup,
		Groups:       make([]GroupSnapshot, len(value.Groups)),
	}
	if value.Chrome != nil {
		decoded.Chrome = ChromeState{
			SidebarVisible:     value.Chrome.SidebarVisible,
			BottomPanelVisible: value.Chrome.BottomPanelVisible,
			StatusBarVisible:   value.Chrome.StatusBarVisible,
		}
	}
	for index, group := range value.Groups {
		decoded.Groups[index] = GroupSnapshot{
			Key:       group.Key,
			TabOrder:  append([]string(nil), group.TabOrder...),
			ActiveKey: group.ActiveKey,
			Collapsed: group.Collapsed,
		}
	}
	if len(value.Dock) != 0 && string(value.Dock) != "null" {
		if err := json.Unmarshal(value.Dock, &decoded.Dock); err != nil {
			return fmt.Errorf("%w: dock: %v", ErrInvalidSnapshot, err)
		}
	}
	if err := decoded.Validate(); err != nil {
		return err
	}
	*s = decoded
	return nil
}

// MarshalSnapshot is the explicit codec entry point for callers that prefer a
// function over json.Marshal on the aliased public type.
func MarshalSnapshot(value Snapshot) ([]byte, error) {
	return json.Marshal(value)
}

// UnmarshalSnapshot is the explicit codec entry point for callers that prefer
// a function over json.Unmarshal.
func UnmarshalSnapshot(data []byte) (Snapshot, error) {
	var value Snapshot
	if err := json.Unmarshal(data, &value); err != nil {
		return Snapshot{}, err
	}
	return value, nil
}

// ValidateSnapshot is a convenience wrapper for persistence boundaries.
func ValidateSnapshot(value Snapshot) error {
	return value.Validate()
}
