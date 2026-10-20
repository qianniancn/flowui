package progress

import (
	"image/color"
	"reflect"
	"testing"

	"github.com/qianniancn/flowui/i18n"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/style"
)

func TestMeterDelegatesToBar(t *testing.T) {
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
	if meter.bar.color != BarSuccess || meter.bar.size != BarLarge || !meter.bar.disabled {
		t.Fatal("meter visual options were not delegated")
	}
	styled := meter.Style(style.Style{}.Radius(4))
	if overridden := styled.progressBar().override; overridden == nil ||
		overridden.Declaration.Resolve(style.State{}).Paint == nil {
		t.Fatal("meter override was not retained on the progress bar")
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

func TestMeterFillUsesConcreteColor(t *testing.T) {
	want := color.NRGBA{R: 0x12, G: 0x34, B: 0x56, A: 0xff}
	bar := Meter("storage", 42).Fill(want).progressBar()
	if bar.override == nil {
		t.Fatal("fill override was not passed to the progress bar")
	}
	fill := style.CascadePart(style.State{}, style.PartFill, bar.override.Declaration)
	solid, ok := fill.Paint.Background.(style.SolidColor)
	if !ok || solid.Color != want {
		t.Fatalf("fill background = %#v, want %v", fill.Paint.Background, want)
	}
}

func TestMeterHasNoIndeterminateMode(t *testing.T) {
	if _, ok := reflect.TypeFor[MeterWidget]().MethodByName("Indeterminate"); ok {
		t.Fatal("meter must represent a measured value, not indeterminate progress")
	}
}

func TestMeterKeepsDefaultSemanticLabel(t *testing.T) {
	bar := Meter("storage", 42).progressBar()
	if bar.semanticLabel != "" || bar.semanticKey != i18n.MeterDefault {
		t.Fatalf("default semantic label = %q key = %q", bar.semanticLabel, bar.semanticKey)
	}
	if got := bar.semanticDescription(); got != "Meter 42%" {
		t.Fatalf("default semantic description = %q", got)
	}
	chinese := core.NewWithCatalog(nil, nil, locale.LanguageChinese, i18n.Catalog{})
	if got := bar.semanticDescriptionFor(chinese); got != "仪表 42%" {
		t.Fatalf("Chinese semantic description = %q", got)
	}
	if got := Meter("storage", 42).Alt("Storage usage").progressBar().semanticDescription(); got != "Storage usage 42%" {
		t.Fatalf("semantic description = %q", got)
	}
}
