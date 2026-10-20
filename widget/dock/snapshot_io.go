package dock

import "encoding/json"

// MarshalSnapshot encodes dock geometry for persistence in a config or
// workspace file. It applies Snapshot's validation before writing.
func MarshalSnapshot(value Snapshot) ([]byte, error) {
	return json.Marshal(value)
}

// UnmarshalSnapshot decodes persisted dock geometry. It rejects snapshots that
// fail validation.
func UnmarshalSnapshot(data []byte) (Snapshot, error) {
	var value Snapshot
	if err := json.Unmarshal(data, &value); err != nil {
		return Snapshot{}, err
	}
	return value, nil
}
