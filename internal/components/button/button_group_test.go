package button

import (
	"image"
	"testing"

	"gioui.org/io/input"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/FlowUI/internal/components/text"
	"github.com/qianniancn/FlowUI/internal/frame"
	flowstyle "github.com/qianniancn/FlowUI/internal/style"
	"github.com/qianniancn/FlowUI/internal/theme"
)

func TestButtonGroupOptionsUseValueSemantics(t *testing.T) {
	base := ButtonGroup(Button("first", text.New("First")), Button("second", text.New("Second")))
	configured := base.
		Orientation(ButtonGroupVertical).
		Variant(ButtonSecondary).
		Size(ButtonSmall).
		Disabled(true).
		FullWidth().
		Separators(true)

	if base.orientation != ButtonGroupHorizontal || base.variant != ButtonPrimary || base.size != ButtonMedium || base.disabled || base.fullWidth || base.separators {
		t.Fatalf("configuring ButtonGroup mutated base: %#v", base)
	}
	if configured.orientation != ButtonGroupVertical || configured.variant != ButtonSecondary || configured.size != ButtonSmall || !configured.disabled || !configured.fullWidth || !configured.separators {
		t.Fatalf("configured ButtonGroup = %#v", configured)
	}
}

func TestButtonGroupInheritsOptionsAndPreservesButtonOverrides(t *testing.T) {
	group := ButtonGroup(
		Button("first", text.New("First")),
		Button("second", text.New("Second")).Variant(ButtonDanger).Size(ButtonLarge).Disabled(false),
	).Variant(ButtonSecondary).Size(ButtonSmall).Disabled(true)

	first := group.prepareButton(group.buttons[0], 0, 2)
	second := group.prepareButton(group.buttons[1], 1, 2)
	if first.variant != ButtonSecondary || first.size != ButtonSmall || !first.disabled {
		t.Fatalf("inherited button = %#v", first)
	}
	if second.variant != ButtonDanger || second.size != ButtonLarge || second.disabled {
		t.Fatalf("overridden button = %#v", second)
	}
}

func TestButtonGroupLayout(t *testing.T) {
	buttons := []ButtonWidget{
		Button("first", text.New("A")).IconOnly(),
		Button("second", text.New("B")).IconOnly(),
	}
	horizontal := ButtonGroup(buttons...).Size(ButtonSmall)
	if dims := horizontal.Layout(newContext(nil), testLayoutContext()); dims.Size != image.Pt(72, 36) {
		t.Fatalf("horizontal ButtonGroup size = %v, want (72,36)", dims.Size)
	}
	vertical := horizontal.Orientation(ButtonGroupVertical)
	if dims := vertical.Layout(newContext(nil), testLayoutContext()); dims.Size != image.Pt(36, 72) {
		t.Fatalf("vertical ButtonGroup size = %v, want (36,72)", dims.Size)
	}
	constrained := testLayoutContext()
	constrained.Constraints = layout.Exact(image.Pt(300, 200))
	if dims := vertical.Layout(newContext(nil), constrained); dims.Size != image.Pt(36, 72) {
		t.Fatalf("constrained vertical ButtonGroup size = %v, want intrinsic (36,72)", dims.Size)
	}
	fullWidth := ButtonGroup(buttons...).FullWidth().Separators(true)
	if dims := fullWidth.Layout(newContext(nil), testLayoutContext()); dims.Size != image.Pt(300, 40) {
		t.Fatalf("full-width ButtonGroup size = %v, want (300,40)", dims.Size)
	}
	verticalFullWidth := fullWidth.Orientation(ButtonGroupVertical)
	if dims := verticalFullWidth.Layout(newContext(nil), testLayoutContext()); dims.Size != image.Pt(300, 80) {
		t.Fatalf("vertical full-width ButtonGroup size = %v, want (300,80)", dims.Size)
	}
}

func TestVerticalButtonGroupEqualizesButtonWidths(t *testing.T) {
	narrow := &buttonGroupWidthProbe{size: image.Pt(16, 16)}
	wide := &buttonGroupWidthProbe{size: image.Pt(80, 16)}
	group := ButtonGroup(
		Button("narrow", narrow).Label("Narrow"),
		Button("wide", wide).Label("Wide"),
	).Orientation(ButtonGroupVertical)
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

func TestVerticalButtonGroupPassesDisabledContext(t *testing.T) {
	for _, button := range []ButtonWidget{
		Button("disabled", &enabledProbeWidget{}).Disabled(true),
		Button("loading", &enabledProbeWidget{}).Loading(true),
	} {
		probe := button.child.(*enabledProbeWidget)
		ButtonGroup(button).Orientation(ButtonGroupVertical).Layout(newContext(nil), testLayoutContext())
		if probe.enabled {
			t.Fatalf("%s button content was laid out as enabled", button.key)
		}
	}
}

func TestVerticalButtonGroupUsesStyleDuringMeasurement(t *testing.T) {
	group := ButtonGroup(
		Button("styled", text.New("Styled")).Loading(true),
		Button("default", text.New("Default")),
	).
		Orientation(ButtonGroupVertical).
		Style(flowstyle.Style{}.PaddingX(48))
	defaultGroup := ButtonGroup(
		Button("styled", text.New("Styled")).Loading(true),
		Button("default", text.New("Default")),
	).Orientation(ButtonGroupVertical)

	styledDims := group.Layout(newContext(nil), testLayoutContext())
	defaultDims := defaultGroup.Layout(newContext(nil), testLayoutContext())
	if styledDims.Size.X <= defaultDims.Size.X || styledDims.Size.Y != 80 {
		t.Fatalf("styled ButtonGroup size = %v, want wider than %v with height 80", styledDims.Size, defaultDims.Size)
	}
}

func TestButtonGroupCorners(t *testing.T) {
	horizontalStart := buttonGroupCorners(buttonGroupItemStyle{grouped: true, position: buttonGroupStart})
	horizontalMiddle := buttonGroupCorners(buttonGroupItemStyle{grouped: true, position: buttonGroupMiddle})
	horizontalEnd := buttonGroupCorners(buttonGroupItemStyle{grouped: true, position: buttonGroupEnd})
	if !horizontalStart.nw || !horizontalStart.sw || horizontalStart.ne || horizontalStart.se {
		t.Fatalf("horizontal start corners = %#v", horizontalStart)
	}
	if horizontalMiddle != (buttonCorners{}) {
		t.Fatalf("horizontal middle corners = %#v", horizontalMiddle)
	}
	if !horizontalEnd.ne || !horizontalEnd.se || horizontalEnd.nw || horizontalEnd.sw {
		t.Fatalf("horizontal end corners = %#v", horizontalEnd)
	}
	verticalEnd := buttonGroupCorners(buttonGroupItemStyle{grouped: true, orientation: ButtonGroupVertical, position: buttonGroupEnd})
	if !verticalEnd.se || !verticalEnd.sw || verticalEnd.nw || verticalEnd.ne {
		t.Fatalf("vertical end corners = %#v", verticalEnd)
	}
}

func TestButtonGroupThemeMatchesHeroUI(t *testing.T) {
	tokens := theme.DefaultTheme().Components.ButtonGroup
	if tokens.SeparatorWidth != 1 || tokens.SeparatorLength != 0.5 || tokens.SeparatorOpacity != 0.15 {
		t.Fatalf("ButtonGroup theme = %#v", tokens)
	}
}

type buttonGroupWidthProbe struct {
	size    image.Point
	layouts int
}

func (p *buttonGroupWidthProbe) Layout(_ *frame.Context, gtx layout.Context) layout.Dimensions {
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
