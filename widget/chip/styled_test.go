package chip

import (
	"image"
	"testing"

	"gioui.org/io/input"
	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/testtheme"
	"github.com/qianniancn/flowui/style"
)

func TestStyledPreservesChipBehavior(t *testing.T) {
	ctx := core.New(nil, testtheme.Behavior(), locale.LanguageEnglish)

	baseStart := &chipProbe{size: image.Pt(12, 12)}
	baseEnd := &chipProbe{size: image.Pt(12, 12)}
	base := New("Styled label").StartContent(baseStart).EndContent(baseEnd)
	baseRouter := new(input.Router)
	baseDims := layoutStyledChipFrame(ctx, baseRouter, base, image.Pt(320, 80))

	styledStart := &chipProbe{size: image.Pt(12, 12)}
	styledEnd := &chipProbe{size: image.Pt(12, 12)}
	styled := (New("Styled label").StartContent(styledStart).EndContent(styledEnd)).Style(style.Style{}.Radius(3))
	styledRouter := new(input.Router)
	styledDims := layoutStyledChipFrame(ctx, styledRouter, styled, image.Pt(320, 80))

	if styledDims != baseDims {
		t.Fatalf("styled dimensions = %v, want base dimensions %v", styledDims, baseDims)
	}
	if baseStart.layouts != 1 || baseEnd.layouts != 1 || styledStart.layouts != 1 || styledEnd.layouts != 1 {
		t.Fatalf("content layouts = base %d/%d styled %d/%d, want one each", baseStart.layouts, baseEnd.layouts, styledStart.layouts, styledEnd.layouts)
	}
	if !chipSemanticTreeContains(styledRouter.AppendSemantics(nil), "Styled label") {
		t.Fatal("styled chip did not preserve label semantics")
	}
}

func layoutStyledChipFrame(ctx *core.Context, router *input.Router, value core.Widget, max image.Point) layout.Dimensions {
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: max},
		Source:      router.Source(),
		Ops:         &ops,
	}
	core.BeginFrame(ctx)
	dims := value.Layout(ctx, gtx)
	core.ApplyFrameCommands(ctx, gtx)
	core.EndFrame(ctx)
	router.Frame(&ops)
	return dims
}
