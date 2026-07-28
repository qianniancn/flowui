package field

import (
	"image"
	"testing"

	"gioui.org/io/input"
	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/flowui/internal/frame"
	"github.com/qianniancn/flowui/internal/locale"
)

func TestDisabledStateClearsHover(t *testing.T) {
	state := &State{Hovered: true}
	state.Update(frame.New(nil, nil, locale.LanguageAuto), fieldTestContext(), true, new(int))
	if state.Hovered {
		t.Fatal("disabled field kept hover state")
	}
}

func fieldTestContext() layout.Context {
	var router input.Router
	var ops op.Ops
	return layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 200)},
		Source:      router.Source(),
		Ops:         &ops,
	}
}
