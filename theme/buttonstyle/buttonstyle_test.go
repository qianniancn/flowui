package buttonstyle_test

import (
	"testing"

	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/theme/buttonstyle"
)

func TestDefaultConfigCustomization(t *testing.T) {
	active := theme.New(theme.DefaultSeed())
	config := buttonstyle.DefaultConfig()
	config.Radius = 0
	config.BorderWidth = 0
	config.ContentGap = 0
	config.Height = 44
	visual := buttonstyle.BuildButton(&active, theme.ButtonRequest{}, config)
	theme.ValidateButtonVisual(visual)
	root := style.Cascade(style.State{}, visual.Defaults, visual.Variant, visual.Size)
	if root.Paint.Radius == nil || *root.Paint.Radius != 0 {
		t.Fatalf("radius = %v, want explicit zero", root.Paint.Radius)
	}
	if gap := visual.ContentGap(theme.ButtonContentGapRequest{}); gap != 0 {
		t.Fatalf("content gap = %v, want explicit zero", gap)
	}
	if root.Box.Height == nil || *root.Box.Height != 44 {
		t.Fatalf("height = %v, want custom height 44", root.Box.Height)
	}
	if buttonstyle.DefaultConfig().Radius == 0 {
		t.Fatal("customizing a config modified subsequent defaults")
	}
}
