package dock

import (
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"math"
)

var ErrInvalidSnapshot = errors.New("flowui: invalid dock snapshot")

type collapseSnapshotJSON struct {
	First  bool `json:"first,omitempty"`
	Second bool `json:"second,omitempty"`
}

type snapshotJSON struct {
	Version      uint16                          `json:"version,omitempty"`
	RootKey      string                          `json:"rootKey,omitempty"`
	Ratios       map[string]float32              `json:"ratios,omitempty"`
	Collapsed    map[string]collapseSnapshotJSON `json:"collapsed,omitempty"`
	MaximizedKey string                          `json:"maximizedKey,omitempty"`
}

// Validate checks the structural and numeric invariants of a current dock
// snapshot.
func (s Snapshot) Validate() error {
	if s.Version != SnapshotVersion {
		return fmt.Errorf("%w: unsupported version %d", ErrInvalidSnapshot, s.Version)
	}
	for key, ratio := range s.Ratios {
		if key == "" {
			return fmt.Errorf("%w: empty ratio key", ErrInvalidSnapshot)
		}
		if math.IsNaN(float64(ratio)) || math.IsInf(float64(ratio), 0) || ratio < 0 || ratio > 1 {
			return fmt.Errorf("%w: ratio %q is %v", ErrInvalidSnapshot, key, ratio)
		}
	}
	for key := range s.Collapsed {
		if key == "" {
			return fmt.Errorf("%w: empty collapsed key", ErrInvalidSnapshot)
		}
	}
	return nil
}

// MarshalJSON writes the stable, versioned representation used for config or
// workspace files.
func (s Snapshot) MarshalJSON() ([]byte, error) {
	if err := s.Validate(); err != nil {
		return nil, err
	}
	value := snapshotJSON{
		Version:      s.Version,
		RootKey:      s.RootKey,
		Ratios:       cloneRatios(s.Ratios),
		Collapsed:    make(map[string]collapseSnapshotJSON, len(s.Collapsed)),
		MaximizedKey: s.MaximizedKey,
	}
	for key, collapsed := range s.Collapsed {
		value.Collapsed[key] = collapseSnapshotJSON(collapsed)
	}
	return json.Marshal(value)
}

// UnmarshalJSON reads the current dock snapshot format.
func (s *Snapshot) UnmarshalJSON(data []byte) error {
	if s == nil {
		return errors.New("flowui: nil dock snapshot")
	}
	var value snapshotJSON
	if err := json.Unmarshal(data, &value); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidSnapshot, err)
	}
	decoded := Snapshot{
		Version:      value.Version,
		RootKey:      value.RootKey,
		Ratios:       cloneRatios(value.Ratios),
		Collapsed:    make(map[string]CollapseState, len(value.Collapsed)),
		MaximizedKey: value.MaximizedKey,
	}
	for key, collapsed := range value.Collapsed {
		decoded.Collapsed[key] = CollapseState(collapsed)
	}
	if err := decoded.Validate(); err != nil {
		return err
	}
	*s = decoded
	return nil
}

func cloneRatios(values map[string]float32) map[string]float32 {
	if values == nil {
		return nil
	}
	clone := make(map[string]float32, len(values))
	maps.Copy(clone, values)
	return clone
}
