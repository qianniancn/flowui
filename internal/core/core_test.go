package core

import (
	"image"
	"testing"

	"gioui.org/io/input"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/style"
)

func TestWidgetFuncAndState(t *testing.T) {
	ctx := New(nil, nil, locale.LanguageAuto)
	value := UseStateWith(ctx, "counter", "test", func() *int {
		initial := 7
		return &initial
	})
	if *value != 7 {
		t.Fatalf("state value = %d, want 7", *value)
	}

	var ops op.Ops
	var router input.Router
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(40, 20)},
		Ops:         &ops,
		Source:      router.Source(),
	}
	widgetFunc := WidgetFunc(func(_ *Context, gtx layout.Context) layout.Dimensions {
		return layout.Dimensions{Size: gtx.Constraints.Max}
	})
	if got := widgetFunc.Layout(ctx, gtx).Size; got != image.Pt(40, 20) {
		t.Fatalf("widget dimensions = %v, want (40,20)", got)
	}
}

func TestScopePreservesState(t *testing.T) {
	ctx := New(nil, nil, locale.LanguageAuto)
	value := style.Style{}.Opacity(0.5)
	restore := PushStyle(ctx, value)
	if styles := ActiveStyles(ctx); len(styles) != 1 || styles[0].Hash64() != value.Hash64() {
		t.Fatalf("active styles = %#v, want one opacity style", styles)
	}
	restore()
	if styles := ActiveStyles(ctx); len(styles) != 0 {
		t.Fatalf("active styles after restore = %#v, want empty", styles)
	}
}

func TestClickableUsesScopedState(t *testing.T) {
	ctx := New(nil, nil, locale.LanguageAuto)
	first := ctx.Clickable("save")
	EndFrame(ctx)
	BeginFrame(ctx)
	second := ctx.Clickable("save")
	if first != second {
		t.Fatal("clickable did not retain state by key")
	}
	if _, ok := PeekState[widget.Clickable](ctx, "save", "clickable"); !ok {
		t.Fatal("clickable state was not retained")
	}
}
