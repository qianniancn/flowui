package chart

import (
	"image"
	"image/color"
	"testing"

	"gioui.org/font"
	"gioui.org/io/input"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/frame"
	"github.com/qianniancn/flowui/internal/locale"
)

func TestRecordTextDefersOpsUntilPlaced(t *testing.T) {
	ctx := frame.New(nil, nil, locale.LanguageAuto)
	gtx, router := chartTextContext()
	RecordText(ctx, gtx, "Recorded", 14, font.Normal, color.NRGBA{A: 0xff}, 200)
	router.Frame(gtx.Ops)
	if chartSemanticLabel(router.AppendSemantics(nil), "Recorded") {
		t.Fatal("recorded text was emitted before placement")
	}

	gtx, router = chartTextContext()
	call, dims := RecordText(ctx, gtx, "Recorded", 14, font.Normal, color.NRGBA{A: 0xff}, 200)
	PlaceRecorded(gtx, call, dims, image.Pt(10, 10))
	router.Frame(gtx.Ops)
	if !chartSemanticLabel(router.AppendSemantics(nil), "Recorded") {
		t.Fatal("placed text did not emit its recorded operations")
	}
}

func chartTextContext() (layout.Context, *input.Router) {
	router := new(input.Router)
	return layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 100)},
		Metric:      unit.Metric{PxPerDp: 1, PxPerSp: 1},
		Source:      router.Source(),
		Ops:         new(op.Ops),
	}, router
}

func chartSemanticLabel(nodes []input.SemanticNode, label string) bool {
	for _, node := range nodes {
		if node.Desc.Label == label || chartSemanticLabel(node.Children, label) {
			return true
		}
	}
	return false
}
