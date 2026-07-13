package animation

import (
	"math"
	"testing"
)

func TestCoreEasingsClampInputAndPreserveEndpoints(t *testing.T) {
	for name, easing := range allEasings() {
		t.Run(name, func(t *testing.T) {
			if got := easing(-1); got != 0 {
				t.Fatalf("easing(-1) = %v, want 0", got)
			}
			if got := easing(2); got != 1 {
				t.Fatalf("easing(2) = %v, want 1", got)
			}
			if got := easing(float32(math.NaN())); got != 0 {
				t.Fatalf("easing(NaN) = %v, want 0", got)
			}
			for step := 0; step <= 100; step++ {
				progress := float32(step) / 100
				got := easing(progress)
				if math.IsNaN(float64(got)) || math.IsInf(float64(got), 0) {
					t.Fatalf("easing(%v) returned non-finite value %v", progress, got)
				}
			}
		})
	}
}

func TestEasingFamilySymmetry(t *testing.T) {
	families := map[string]struct {
		in    Easing
		out   Easing
		inOut Easing
	}{
		"quadratic":  {EaseQuadraticIn, EaseQuadraticOut, EaseQuadraticInOut},
		"cubic":      {EaseCubicIn, EaseCubicOut, EaseCubicInOut},
		"quartic":    {EaseQuarticIn, EaseQuarticOut, EaseQuarticInOut},
		"quintic":    {EaseQuinticIn, EaseQuinticOut, EaseQuinticInOut},
		"sinusoidal": {EaseSinusoidalIn, EaseSinusoidalOut, EaseSinusoidalInOut},
		"exponential": {EaseExponentialIn, EaseExponentialOut,
			EaseExponentialInOut},
		"circular": {EaseCircularIn, EaseCircularOut, EaseCircularInOut},
		"elastic":  {EaseElasticIn, EaseElasticOut, EaseElasticInOut},
		"back":     {EaseBackIn, EaseBackOut, EaseBackInOut},
		"bounce":   {EaseBounceIn, EaseBounceOut, EaseBounceInOut},
	}
	for name, family := range families {
		t.Run(name, func(t *testing.T) {
			for step := 0; step <= 100; step++ {
				progress := float32(step) / 100
				want := 1 - family.out(1-progress)
				if got := family.in(progress); !closeEnough(got, want, 2e-5) {
					t.Fatalf("in(%v) = %v, want mirrored out %v", progress, got, want)
				}
			}
			if got := family.inOut(0.5); !closeEnough(got, 0.5, 1e-6) {
				t.Fatalf("inOut(0.5) = %v, want 0.5", got)
			}
		})
	}
}

func TestEasingsMatchZRenderReferenceValues(t *testing.T) {
	if got := EaseCubicOut(0.5); math.Abs(float64(got-0.875)) > 1e-6 {
		t.Fatalf("EaseCubicOut(0.5) = %v, want 0.875", got)
	}
	if got := EaseElasticOut(0.25); !closeEnough(got, 1.125, 1e-6) {
		t.Fatalf("EaseElasticOut(0.25) = %v, want 1.125", got)
	}
	if got := EaseBackIn(0.5); !closeEnough(got, -0.0876975, 1e-6) {
		t.Fatalf("EaseBackIn(0.5) = %v, want -0.0876975", got)
	}
	if got := EaseBounceOut(0.5); !closeEnough(got, 0.765625, 1e-6) {
		t.Fatalf("EaseBounceOut(0.5) = %v, want 0.765625", got)
	}
	if got := EaseBackOut(0.7); got <= 1 {
		t.Fatalf("EaseBackOut(0.7) = %v, want overshoot", got)
	}
}

func allEasings() map[string]Easing {
	return map[string]Easing{
		"linear":            EaseLinear,
		"quadratic in":      EaseQuadraticIn,
		"quadratic out":     EaseQuadraticOut,
		"quadratic inout":   EaseQuadraticInOut,
		"cubic in":          EaseCubicIn,
		"cubic out":         EaseCubicOut,
		"cubic inout":       EaseCubicInOut,
		"quartic in":        EaseQuarticIn,
		"quartic out":       EaseQuarticOut,
		"quartic inout":     EaseQuarticInOut,
		"quintic in":        EaseQuinticIn,
		"quintic out":       EaseQuinticOut,
		"quintic inout":     EaseQuinticInOut,
		"sinusoidal in":     EaseSinusoidalIn,
		"sinusoidal out":    EaseSinusoidalOut,
		"sinusoidal inout":  EaseSinusoidalInOut,
		"exponential in":    EaseExponentialIn,
		"exponential out":   EaseExponentialOut,
		"exponential inout": EaseExponentialInOut,
		"circular in":       EaseCircularIn,
		"circular out":      EaseCircularOut,
		"circular inout":    EaseCircularInOut,
		"elastic in":        EaseElasticIn,
		"elastic out":       EaseElasticOut,
		"elastic inout":     EaseElasticInOut,
		"back in":           EaseBackIn,
		"back out":          EaseBackOut,
		"back inout":        EaseBackInOut,
		"bounce in":         EaseBounceIn,
		"bounce out":        EaseBounceOut,
		"bounce inout":      EaseBounceInOut,
		"smoothstep":        EaseSmoothstep,
	}
}

func closeEnough(got, want, tolerance float32) bool {
	return math.Abs(float64(got-want)) <= float64(tolerance)
}

func TestApplyEasingRejectsNonFiniteOutput(t *testing.T) {
	for _, result := range []float32{float32(math.NaN()), float32(math.Inf(1))} {
		func() {
			defer func() {
				if recover() == nil {
					t.Fatal("non-finite easing result did not panic")
				}
			}()
			applyEasing(func(float32) float32 { return result }, 0.5)
		}()
	}
}
