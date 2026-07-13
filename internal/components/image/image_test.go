package imageview

import (
	"image"
	"testing"

	"gioui.org/io/input"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"gioui.org/widget"
	"github.com/qianniancn/FlowUI/internal/frame"
	flowlayout "github.com/qianniancn/FlowUI/internal/layout"
	"github.com/qianniancn/FlowUI/internal/locale"
)

func TestImageOptionsUseValueSemantics(t *testing.T) {
	source := imageSource(120, 80)
	base := New(source)
	configured := base.
		Fit(FitCover).
		Position(flowlayout.AlignBottomEnd).
		Width(200).
		Height(120).
		Radius(24).
		Opacity(0.5).
		Alt("Landscape")
	if base.fit != FitScaleDown || base.position != flowlayout.AlignCenter || base.hasWidth || base.hasHeight || base.radius != 0 || base.hasOpacity || base.alt != "" {
		t.Fatalf("configuring Image mutated base: %#v", base)
	}
	if configured.fit != FitCover || configured.position != flowlayout.AlignBottomEnd || configured.width != 200 || configured.height != 120 || configured.radius != 24 || configured.opacity != 0.5 || configured.alt != "Landscape" {
		t.Fatalf("configured Image = %#v", configured)
	}
}

func TestImageFitModesMapToGio(t *testing.T) {
	tests := []struct {
		fit  Fit
		want widget.Fit
	}{{FitScaleDown, widget.ScaleDown}, {FitContain, widget.Contain}, {FitCover, widget.Cover}, {FitFill, widget.Fill}, {FitUnscaled, widget.Unscaled}}
	for _, test := range tests {
		if got := New(paint.ImageOp{}).Fit(test.fit).gioFit(); got != test.want {
			t.Fatalf("Image fit %d = %d, want %d", test.fit, got, test.want)
		}
	}
}

func TestImageDefaultsToIntrinsicSizeAndScalesDown(t *testing.T) {
	ctx := imageTestFrameContext()
	if dims := New(imageSource(120, 80)).Layout(ctx, imageTestLayoutContext(image.Pt(300, 300))); dims.Size != image.Pt(120, 80) {
		t.Fatalf("intrinsic Image size = %v", dims.Size)
	}
	if dims := New(imageSource(120, 80)).Layout(ctx, imageTestLayoutContext(image.Pt(60, 60))); dims.Size != image.Pt(60, 40) {
		t.Fatalf("scaled Image size = %v", dims.Size)
	}
}

func TestImageFixedBoundsSupportContainAndCover(t *testing.T) {
	ctx := imageTestFrameContext()
	for _, fit := range []Fit{FitContain, FitCover, FitFill} {
		dims := New(imageSource(120, 80)).Fit(fit).Width(60).Height(60).Layout(ctx, imageTestLayoutContext(image.Pt(300, 300)))
		if dims.Size != image.Pt(60, 60) {
			t.Fatalf("Image fit %d size = %v", fit, dims.Size)
		}
	}
	if dims := New(imageSource(120, 80)).Width(60).Layout(ctx, imageTestLayoutContext(image.Pt(300, 300))); dims.Size != image.Pt(60, 40) {
		t.Fatalf("width-only Image size = %v", dims.Size)
	}
}

func TestImageFixedBoundsRespectParentMinimums(t *testing.T) {
	ctx := imageTestFrameContext()
	gtx := imageTestLayoutContext(image.Pt(200, 200))
	gtx.Constraints.Min.X = 100
	if dims := New(imageSource(120, 80)).Width(50).Layout(ctx, gtx); dims.Size != image.Pt(100, 66) {
		t.Fatalf("minimum-width Image size = %v", dims.Size)
	}

	gtx = imageTestLayoutContext(image.Pt(200, 200))
	gtx.Constraints.Min.Y = 100
	if dims := New(imageSource(80, 120)).Height(50).Layout(ctx, gtx); dims.Size != image.Pt(66, 100) {
		t.Fatalf("minimum-height Image size = %v", dims.Size)
	}
}

func TestEmptyImageKeepsExplicitLayoutBounds(t *testing.T) {
	dims := New(paint.ImageOp{}).Width(80).Height(48).Layout(imageTestFrameContext(), imageTestLayoutContext(image.Pt(300, 300)))
	if dims.Size != image.Pt(80, 48) {
		t.Fatalf("empty Image size = %v", dims.Size)
	}
}

func TestEmptyImageDoesNotExposeAltText(t *testing.T) {
	router := new(input.Router)
	var ops op.Ops
	gtx := layout.Context{Ops: &ops, Source: router.Source(), Constraints: layout.Constraints{Max: image.Pt(200, 200)}}
	New(paint.ImageOp{}).Width(80).Height(48).Alt("Missing image").Layout(imageTestFrameContext(), gtx)
	router.Frame(&ops)
	if imageSemanticTreeContains(router.AppendSemantics(nil), "Missing image") {
		t.Fatal("empty Image exposed alt text")
	}
}

func TestImageOpacityIsClamped(t *testing.T) {
	if got := New(paint.ImageOp{}).Opacity(-1); !got.hasOpacity || got.opacity != 0 {
		t.Fatalf("negative opacity = %v", got.opacity)
	}
	if got := New(paint.ImageOp{}).Opacity(2); !got.hasOpacity || got.opacity != 1 {
		t.Fatalf("large opacity = %v", got.opacity)
	}
}

func TestImageExposesAccessibleLabel(t *testing.T) {
	router := new(input.Router)
	var ops op.Ops
	gtx := layout.Context{Ops: &ops, Source: router.Source(), Constraints: layout.Constraints{Max: image.Pt(200, 200)}}
	New(imageSource(80, 60)).Alt("Mountain landscape").Layout(imageTestFrameContext(), gtx)
	router.Frame(&ops)
	if !imageSemanticTreeContains(router.AppendSemantics(nil), "Mountain landscape") {
		t.Fatal("Image semantics did not expose alt text")
	}
}

func imageSource(width, height int) paint.ImageOp {
	return paint.NewImageOp(image.NewRGBA(image.Rect(0, 0, width, height)))
}

func imageTestFrameContext() *frame.Context {
	return frame.New(nil, nil, locale.LanguageEnglish)
}

func imageTestLayoutContext(maximum image.Point) layout.Context {
	return layout.Context{Ops: new(op.Ops), Constraints: layout.Constraints{Max: maximum}}
}

func imageSemanticTreeContains(nodes []input.SemanticNode, label string) bool {
	for _, node := range nodes {
		if node.Desc.Label == label || imageSemanticTreeContains(node.Children, label) {
			return true
		}
	}
	return false
}
