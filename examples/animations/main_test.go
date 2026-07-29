package main

import (
	"image"
	"testing"
	"time"

	"github.com/qianniancn/flowui/uitest"
)

func TestUpdateAnimationState(t *testing.T) {
	model := Model{forward: true, timelinePlaying: true}

	Update(&model, ToggleDirection{})
	if model.forward {
		t.Fatal("ToggleDirection did not reverse the easing direction")
	}
	Update(&model, ToggleTimeline{})
	if model.timelinePlaying {
		t.Fatal("ToggleTimeline did not pause the timeline")
	}
	Update(&model, RestartTimeline{})
	if model.timelineRun != 1 {
		t.Fatalf("RestartTimeline revision = %d, want 1", model.timelineRun)
	}
	Update(&model, ToggleLayout{})
	if !model.layoutExpanded {
		t.Fatal("ToggleLayout did not expand the animated layout")
	}
	Update(&model, ToggleRect{})
	if !model.rectAlternate {
		t.Fatal("ToggleRect did not select the alternate rectangle")
	}
}

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

func TestSpringDemoSpecsAreCompleteAndUnique(t *testing.T) {
	if got := len(springDemoSpecs); got != 4 {
		t.Fatalf("spring preset count = %d, want 4", got)
	}

	keys := make(map[string]struct{}, len(springDemoSpecs))
	for _, spec := range springDemoSpecs {
		if spec.key == "" || spec.label == "" {
			t.Fatalf("incomplete spring demo spec: %+v", spec)
		}
		if _, exists := keys[spec.key]; exists {
			t.Fatalf("duplicate spring demo key %q", spec.key)
		}
		keys[spec.key] = struct{}{}
		if spec.config.Stiffness <= 0 || spec.config.Damping < 0 || spec.config.Mass <= 0 ||
			spec.config.RestDisplacement <= 0 || spec.config.RestVelocity <= 0 {
			t.Fatalf("invalid spring config for %q: %+v", spec.key, spec.config)
		}
		if spec.mix < 0 || spec.mix > 1 {
			t.Fatalf("spring color mix for %q = %v, want [0, 1]", spec.key, spec.mix)
		}
	}
}

func TestViewLayoutsAtDesktopAndNarrowSizes(t *testing.T) {
	for _, size := range []image.Point{image.Pt(1100, 760), image.Pt(720, 640)} {
		t.Run(size.String(), func(t *testing.T) {
			model := Model{forward: true, timelinePlaying: true}
			harness := uitest.New(size)
			frame := func() {
				dimensions := harness.Frame(View(harness.Context(), model, func(Msg) {}))
				if dimensions.Size != size {
					t.Fatalf("root size = %v, want %v", dimensions.Size, size)
				}
			}

			frame()
			model.forward = false
			model.layoutExpanded = true
			model.rectAlternate = true
			model.timelineRun++
			harness.Advance(100 * time.Millisecond)
			frame()
			model.timelinePlaying = false
			harness.Advance(500 * time.Millisecond)
			frame()
		})
	}
}
