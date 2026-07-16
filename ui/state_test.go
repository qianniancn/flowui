package ui

import (
	"testing"

	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/locale"
)

type customFrameState struct {
	frames int
}

func (s *customFrameState) BeginFrame() {
	s.frames++
}

func TestUseStateRetainsAndSweepsCustomState(t *testing.T) {
	ctx := frame.New(nil, nil, locale.LanguageEnglish)
	frame.BeginFrame(ctx)
	first := UseStateWith(ctx, "canvas", func() *customFrameState {
		return &customFrameState{frames: 1}
	})
	frame.EndFrame(ctx)

	frame.BeginFrame(ctx)
	second := UseState[customFrameState](ctx, "canvas")
	frame.EndFrame(ctx)
	if first != second || second.frames != 2 {
		t.Fatalf("custom state = same %v frames %d", first == second, second.frames)
	}

	frame.BeginFrame(ctx)
	frame.EndFrame(ctx)
	if got := frame.StateLen(ctx); got != 0 {
		t.Fatalf("custom state retained after removal: %d", got)
	}
}
