package button

import (
	"image"
	"image/color"
	"testing"

	"gioui.org/io/input"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/testtheme"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

func TestGroupInheritsOptionsAndPreservesButtonOverrides(t *testing.T) {
	group := Group(
		New("First"),
		New("Second").Variant(Danger).Size(Large).Disabled(false),
	).Variant(Secondary).Size(Small).Disabled(true)

	first := group.prepareButton(group.buttons[0], 0, 2)
	second := group.prepareButton(group.buttons[1], 1, 2)
	if first.variant != Secondary || first.size != Small || !first.disabled {
		t.Fatalf("inherited button = %#v", first)
	}
	if second.variant != Danger || second.size != Large || second.disabled {
		t.Fatalf("overridden button = %#v", second)
	}
}

func TestGroupLayout(t *testing.T) {
	buttons := []Widget{
		New("A").IconOnly(),
		New("B").IconOnly(),
	}
	horizontal := Group(buttons...).Size(Small)
	if dims := horizontal.Layout(newContext(nil), testLayoutContext()); dims.Size != image.Pt(64, 32) {
		t.Fatalf("horizontal ButtonGroup size = %v, want (64,32)", dims.Size)
	}
	vertical := horizontal.Orientation(GroupVertical)
	if dims := vertical.Layout(newContext(nil), testLayoutContext()); dims.Size != image.Pt(32, 64) {
		t.Fatalf("vertical ButtonGroup size = %v, want (32,64)", dims.Size)
	}
	constrained := testLayoutContext()
	constrained.Constraints = layout.Exact(image.Pt(300, 200))
	if dims := vertical.Layout(newContext(nil), constrained); dims.Size != image.Pt(32, 64) {
		t.Fatalf("constrained vertical ButtonGroup size = %v, want intrinsic (32,64)", dims.Size)
	}
	fullWidth := Group(buttons...).FullWidth().Separators(true)
	if dims := fullWidth.Layout(newContext(nil), testLayoutContext()); dims.Size != image.Pt(300, 40) {
		t.Fatalf("full-width ButtonGroup size = %v, want (300,40)", dims.Size)
	}
	verticalFullWidth := fullWidth.Orientation(GroupVertical)
	if dims := verticalFullWidth.Layout(newContext(nil), testLayoutContext()); dims.Size != image.Pt(300, 80) {
		t.Fatalf("vertical full-width ButtonGroup size = %v, want (300,80)", dims.Size)
	}
}

func TestGroupResolvesSeparatorVisualsOnlyWhenEnabled(t *testing.T) {
	for _, separators := range []bool{false, true} {
		t.Run(map[bool]string{false: "disabled", true: "enabled"}[separators], func(t *testing.T) {
			active := *testtheme.Behavior()
			provider := &separatorVisualProbe{ComponentVisualProvider: testtheme.BehaviorProvider()}
			active.VisualProvider = provider
			ctx := core.New(nil, &active, locale.LanguageAuto)
			group := Group(New("").Content(nil), New("").Content(nil)).Separators(separators)
			layoutFrame := func() {
				core.BeginFrame(ctx)
				group.Layout(ctx, testLayoutContext())
				core.EndFrame(ctx)
			}

			layoutFrame()
			provider.foregroundCalls = 0
			provider.separatorCalls = 0
			layoutFrame()

			wantForeground, wantSeparators := 0, 0
			if separators {
				wantForeground, wantSeparators = 2, 1
			}
			if provider.foregroundCalls != wantForeground || provider.separatorCalls != wantSeparators {
				t.Fatalf("separator resolver calls = %d foreground, %d renderer; want %d, %d", provider.foregroundCalls, provider.separatorCalls, wantForeground, wantSeparators)
			}
		})
	}
}

type separatorVisualProbe struct {
	theme.ComponentVisualProvider
	foregroundCalls int
	separatorCalls  int
}

func (*separatorVisualProbe) VisualsAreImmutable() {}

func (provider *separatorVisualProbe) Button(active *theme.Theme, request theme.ButtonRequest) theme.ButtonVisual {
	visual := provider.ComponentVisualProvider.Button(active, request)
	foreground := visual.ForegroundColor
	visual.ForegroundColor = func(resolved style.ResolvedStyle) (color.NRGBA, bool) {
		provider.foregroundCalls++
		return foreground(resolved)
	}
	separator := visual.Group.SeparatorRenderer
	visual.Group.SeparatorRenderer = func(gtx layout.Context, request theme.ButtonGroupSeparatorRequest) {
		provider.separatorCalls++
		separator(gtx, request)
	}
	return visual
}

func TestVerticalGroupEqualizesButtonWidths(t *testing.T) {
	narrow := &buttonGroupWidthProbe{size: image.Pt(16, 16)}
	wide := &buttonGroupWidthProbe{size: image.Pt(80, 16)}
	group := Group(
		New("").Content(narrow).Label("Narrow"),
		New("").Content(wide).Label("Wide"),
	).Orientation(GroupVertical)
	var router input.Router
	var operations op.Ops
	gtx := layout.Context{Constraints: layout.Constraints{Max: image.Pt(300, 200)}, Source: router.Source(), Ops: &operations}
	if dims := group.Layout(newContext(nil), gtx); dims.Size != image.Pt(112, 80) {
		t.Fatalf("vertical ButtonGroup size = %v, want (112,80)", dims.Size)
	}
	router.Frame(&operations)
	buttons := buttonGroupSemanticButtons(router.AppendSemantics(nil))
	widths := make(map[string]int, 2)
	for _, button := range buttons {
		widths[button.Desc.Label] = button.Desc.Bounds.Dx()
	}
	if widths["Narrow"] != 112 || widths["Wide"] != 112 {
		t.Fatalf("vertical button semantic bounds = %#v, want two 112px buttons", buttons)
	}
	if narrow.layouts != 1 || wide.layouts != 1 {
		t.Fatalf("vertical button content layouts = %d/%d, want 1/1", narrow.layouts, wide.layouts)
	}
}

func TestVerticalGroupPassesDisabledContext(t *testing.T) {
	for _, button := range []Widget{
		New("").Content(&enabledProbeWidget{}).Disabled(true),
		New("").Content(&enabledProbeWidget{}).Loading(true),
	} {
		probe := button.child.(*enabledProbeWidget)
		Group(button).Orientation(GroupVertical).Layout(newContext(nil), testLayoutContext())
		if probe.enabled {
			t.Fatalf("%s button content was laid out as enabled", button.key)
		}
	}
}

func TestVerticalGroupUsesStyleDuringMeasurement(t *testing.T) {
	group := (Group(
		New("Styled").Loading(true),
		New("Default"),
	).Orientation(GroupVertical)).Style(style.Style{}.PaddingX(48))
	defaultGroup := Group(
		New("Styled").Loading(true),
		New("Default"),
	).Orientation(GroupVertical)

	styledDims := group.Layout(newContext(nil), testLayoutContext())
	defaultDims := defaultGroup.Layout(newContext(nil), testLayoutContext())
	if styledDims.Size.X <= defaultDims.Size.X || styledDims.Size.Y != 80 {
		t.Fatalf("styled ButtonGroup size = %v, want wider than %v with height 80", styledDims.Size, defaultDims.Size)
	}
}

func TestGroupCorners(t *testing.T) {
	horizontalStart := buttonGroupCorners(buttonGroupItemStyle{grouped: true, position: buttonGroupStart})
	horizontalMiddle := buttonGroupCorners(buttonGroupItemStyle{grouped: true, position: buttonGroupMiddle})
	horizontalEnd := buttonGroupCorners(buttonGroupItemStyle{grouped: true, position: buttonGroupEnd})
	if !horizontalStart.NorthWest || !horizontalStart.SouthWest || horizontalStart.NorthEast || horizontalStart.SouthEast {
		t.Fatalf("horizontal start corners = %#v", horizontalStart)
	}
	if horizontalMiddle != (theme.ButtonGroupRequest{}) {
		t.Fatalf("horizontal middle corners = %#v", horizontalMiddle)
	}
	if !horizontalEnd.NorthEast || !horizontalEnd.SouthEast || horizontalEnd.NorthWest || horizontalEnd.SouthWest {
		t.Fatalf("horizontal end corners = %#v", horizontalEnd)
	}
	verticalEnd := buttonGroupCorners(buttonGroupItemStyle{grouped: true, orientation: GroupVertical, position: buttonGroupEnd})
	if !verticalEnd.SouthEast || !verticalEnd.SouthWest || verticalEnd.NorthWest || verticalEnd.NorthEast {
		t.Fatalf("vertical end corners = %#v", verticalEnd)
	}
}

type buttonGroupWidthProbe struct {
	size    image.Point
	layouts int
}

func (p *buttonGroupWidthProbe) Layout(_ *core.Context, gtx layout.Context) layout.Dimensions {
	p.layouts++
	return layout.Dimensions{Size: gtx.Constraints.Constrain(p.size)}
}

func buttonGroupSemanticButtons(nodes []input.SemanticNode) []input.SemanticNode {
	var buttons []input.SemanticNode
	for _, node := range nodes {
		if node.Desc.Class == semantic.Button {
			buttons = append(buttons, node)
		}
		buttons = append(buttons, buttonGroupSemanticButtons(node.Children)...)
	}
	return buttons
}
