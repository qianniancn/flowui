package state

import "testing"

func TestFrameMapRetainsOnlySeenValues(t *testing.T) {
	var values map[string]*int
	var seen map[string]struct{}
	BeginFrameMap(&seen)
	*UseFrameMap(&values, &seen, "a") = 1
	*UseFrameMap(&values, &seen, "b") = 2
	SweepFrameMap(values, seen)
	BeginFrameMap(&seen)
	if got := UseFrameMap(&values, &seen, "b"); *got != 2 {
		t.Fatalf("retained value = %d, want 2", *got)
	}
	SweepFrameMap(values, seen)
	if values["a"] != nil || values["b"] == nil {
		t.Fatalf("values after sweep = %#v", values)
	}
}
