package main

import "testing"

func TestEasingCurveSpecsAreCompleteAndUnique(t *testing.T) {
	if got := len(easingCurveSpecs); got != 31 {
		t.Fatalf("easing curve count = %d, want 31", got)
	}

	keys := make(map[string]struct{}, len(easingCurveSpecs))
	for _, spec := range easingCurveSpecs {
		if spec.label == "" || spec.key == "" || spec.easing == nil {
			t.Fatalf("incomplete easing curve spec: %+v", spec)
		}
		if _, exists := keys[spec.key]; exists {
			t.Fatalf("duplicate easing curve key %q", spec.key)
		}
		keys[spec.key] = struct{}{}
	}
}
