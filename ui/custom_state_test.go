package ui

import (
	"testing"

	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/theme"
)

type customFrameState struct {
	frames int
}

func (s *customFrameState) BeginFrame() {
	s.frames++
}

func TestUseStateRetainsAndSweepsCustomState(t *testing.T) {
	active := theme.NeutralTheme()
	ctx := core.New(nil, &active, locale.LanguageEnglish)
	core.BeginFrame(ctx)
	first := UseStateWith(ctx, "canvas", func() *customFrameState {
		return &customFrameState{frames: 1}
	})
	core.EndFrame(ctx)

	core.BeginFrame(ctx)
	second := UseState[customFrameState](ctx, "canvas")
	core.EndFrame(ctx)
	if first != second || second.frames != 2 {
		t.Fatalf("custom state = same %v frames %d", first == second, second.frames)
	}

	core.BeginFrame(ctx)
	core.EndFrame(ctx)
	if got := core.StateLen(ctx); got != 0 {
		t.Fatalf("custom state retained after removal: %d", got)
	}
}
