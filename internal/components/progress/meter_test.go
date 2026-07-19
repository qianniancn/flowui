package progress

import (
	"reflect"
	"testing"
)

func TestMeterDelegatesToProgressBar(t *testing.T) {
	meter := Meter("storage", 60).
		Label("Storage").
		Alt("Storage usage").
		ShowValue().
		Range(0, 120).
		Color(MeterSuccess).
		Size(MeterLarge).
		Disabled(true)

	if meter.bar.key != "storage" || meter.bar.value != 60 || meter.bar.label != "Storage" {
		t.Fatal("meter did not configure the shared progress bar")
	}
	if meter.alt != "Storage usage" || !meter.bar.showValue || meter.bar.maxValue != 120 {
		t.Fatal("meter semantics or range were not retained")
	}
	if meter.bar.color != ProgressBarSuccess || meter.bar.size != ProgressBarLarge || !meter.bar.disabled {
		t.Fatal("meter visual options were not delegated")
	}
}

func TestMeterKeepsValueTextOverFormatter(t *testing.T) {
	meter := Meter("storage", 42).
		ValueFormatter(func(float64) string { return "formatted" }).
		ValueText("42 GB")
	bar := meter.progressBar()
	if !bar.hasValueText || bar.valueText != "42 GB" {
		t.Fatal("explicit meter value text was not retained")
	}
	if got := Meter("storage", 42).ValueFormatter(func(float64) string { return "formatted" }).progressBar().valueText; got != "formatted" {
		t.Fatalf("formatted value = %q", got)
	}
}

func TestMeterHasNoIndeterminateMode(t *testing.T) {
	if _, ok := reflect.TypeFor[MeterWidget]().MethodByName("Indeterminate"); ok {
		t.Fatal("meter must represent a measured value, not indeterminate progress")
	}
}

func TestMeterKeepsDefaultSemanticLabel(t *testing.T) {
	if got := Meter("storage", 42).progressBar().semanticLabel; got != "Meter" {
		t.Fatalf("semantic label = %q", got)
	}
	if got := Meter("storage", 42).Alt("Storage usage").progressBar().semanticDescription(); got != "Storage usage 42%" {
		t.Fatalf("semantic description = %q", got)
	}
}
