package ui

import (
	"image"
	"testing"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/flowui/internal/frame"
)

func TestMeasureTextMatchesTextLayout(t *testing.T) {
	ctx := frame.New(nil, nil, LanguageEnglish)
	value := Text("FlowUI custom component").
		Size(16).
		Weight(font.Medium).
		LineHeight(22).
		MaxLines(1)
	constraints := layout.Constraints{Max: image.Pt(400, 100)}
	measured := MeasureText(ctx, layout.Context{Ops: new(op.Ops), Constraints: constraints}, value)
	laidOut := value.Layout(ctx, layout.Context{Ops: new(op.Ops), Constraints: constraints})
	if measured != laidOut {
		t.Fatalf("measured = %v, laid out = %v", measured, laidOut)
	}
	if measured.Size.X <= 0 || measured.Size.Y <= 0 {
		t.Fatalf("measured size = %v, want non-zero", measured.Size)
	}
}
