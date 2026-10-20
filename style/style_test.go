package style

import (
	"image/color"
	"math"
	"reflect"
	"testing"
	"time"

	"gioui.org/font"
	"gioui.org/io/pointer"
	giotext "gioui.org/text"
	"gioui.org/unit"
)

func TestStyleIsEmpty(t *testing.T) {
	tests := []struct {
		name  string
		style Style
		want  bool
	}{
		{name: "zero value", style: Style{}, want: true},
		{name: "box declaration", style: Style{}.MinHeight(1), want: false},
		{name: "empty part declaration", style: Style{}.Part(PartItem, Style{}), want: false},
		{name: "conditional rule", style: Style{}.On(Hovered, Style{}), want: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := test.style.IsEmpty(); got != test.want {
				t.Fatalf("Style.IsEmpty() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestBuilderCoversCommonStyleModel(t *testing.T) {
	ease := EaseFunc(func(value float32) float32 { return value })
	built := Style{}.
		Width(100).Height(40).
		MinWidth(80).MaxWidth(120).MinHeight(20).MaxHeight(60).
		FillWidth().FillHeight().AspectRatio(1.5).
		Padding(4).PaddingLeft(8).
		MarginX(2).MarginBottom(3).
		Overflow(OverflowHidden).Cursor(pointer.CursorPointer).
		FontWeight(600).Typeface("serif").FontStyle(font.Italic).
		LineHeight(20).LineHeightScale(1.2).MaxLines(2).TextAlign(TextAlignCenter).
		Wrap(giotext.WrapWords).Truncator("...").
		Translate(6, 7).Scale(.9).Rotate(.5).
		Transition(PropTransform, time.Second, TransitionDelay(20*time.Millisecond), TransitionEase(ease)).
		Resolve(State{})

	if built.Box == nil || *built.Box.MinWidth != 80 || !*built.Box.FillWidth || !*built.Box.FillHeight || *built.Box.AspectRatio != 1.5 || built.Box.Padding.Left != 8 || built.Box.Margin.Bottom != 3 || *built.Box.Overflow != OverflowHidden || *built.Box.Cursor != pointer.CursorPointer {
		t.Fatalf("box style = %#v", built.Box)
	}
	if built.Text == nil || *built.Text.FontWeight != 600 || *built.Text.Typeface != "serif" || *built.Text.FontStyle != font.Italic ||
		*built.Text.LineHeightScale != 1.2 || *built.Text.MaxLines != 2 || *built.Text.Align != TextAlignCenter ||
		*built.Text.Wrap != giotext.WrapWords || *built.Text.Truncator != "..." {
		t.Fatalf("text style = %#v", built.Text)
	}
	if built.Trans == nil || *built.Trans.TranslateX != 6 || *built.Trans.Rotate != .5 {
		t.Fatalf("transform style = %#v", built.Trans)
	}
	if len(built.Transitions) != 1 || built.Transitions[0].Delay != 20*time.Millisecond || built.Transitions[0].Ease == nil {
		t.Fatalf("transitions = %#v", built.Transitions)
	}
}

func TestTransitionIgnoresUnknownPropertyID(t *testing.T) {
	const unknown PropertyID = 255
	if unknown.Animatable() {
		t.Fatal("unknown property should not be animatable")
	}
	built := Style{}.
		Opacity(1).
		Transition(unknown, time.Second).
		Transition(PropOpacity, time.Millisecond).
		Resolve(State{})
	if len(built.Transitions) != 1 || built.Transitions[0].Property != PropOpacity {
		t.Fatalf("transitions = %#v, want only PropOpacity", built.Transitions)
	}
}

func TestTransformIgnoresNonFiniteValues(t *testing.T) {
	base := Style{}.Translate(2, 3).Scale(.5).Rotate(.7)
	for _, value := range []float32{float32(math.NaN()), float32(math.Inf(1)), float32(math.Inf(-1))} {
		resolved := base.
			Translate(unit.Dp(value), 4).
			Translate(4, unit.Dp(value)).
			Scale(value).
			Rotate(value).
			Resolve(State{})
		if *resolved.Trans.TranslateX != 2 || *resolved.Trans.TranslateY != 3 ||
			*resolved.Trans.ScaleX != .5 || *resolved.Trans.ScaleY != .5 || *resolved.Trans.Rotate != .7 {
			t.Fatalf("non-finite %v changed transform: %#v", value, resolved.Trans)
		}
	}
}

func TestLayoutValuesIgnoreNonFiniteValues(t *testing.T) {
	base := Style{}.
		Width(100).Height(40).
		MinWidth(80).MaxWidth(120).MinHeight(20).MaxHeight(60).
		Padding(4).Margin(5).
		BorderWidth(1).Radius(6).
		BoxShadow(1, 2, 3, 4, RGB(0x112233)).
		Outline(2, 3, RGB(0x445566)).
		Opacity(.75).
		FontSize(14).LineHeight(20).LineHeightScale(1.25)

	for _, value := range []float32{float32(math.NaN()), float32(math.Inf(1)), float32(math.Inf(-1))} {
		length := unit.Dp(value)
		textLength := unit.Sp(value)
		resolved := base.
			Width(length).Height(length).
			MinWidth(length).MaxWidth(length).MinHeight(length).MaxHeight(length).
			Padding(length).PaddingX(length).PaddingY(length).
			PaddingTop(length).PaddingRight(length).PaddingBottom(length).PaddingLeft(length).
			Margin(length).MarginX(length).MarginY(length).
			MarginTop(length).MarginRight(length).MarginBottom(length).MarginLeft(length).
			BorderWidth(length).
			Radius(length).
			RadiusTopLeft(length).RadiusTopRight(length).RadiusBottomRight(length).RadiusBottomLeft(length).
			BoxShadow(length, 2, 3, 4, RGB(0)).
			BoxShadow(1, length, 3, 4, RGB(0)).
			BoxShadow(1, 2, length, 4, RGB(0)).
			BoxShadow(1, 2, 3, length, RGB(0)).
			Outline(length, 2, RGB(0)).Outline(1, length, RGB(0)).
			Opacity(value).
			FontSize(textLength).LineHeight(textLength).LineHeightScale(value).
			Resolve(State{})

		want := base.Resolve(State{})
		if !reflect.DeepEqual(resolved, want) {
			t.Fatalf("non-finite %v changed layout style:\n got %#v\nwant %#v", value, resolved, want)
		}
	}
}

func TestPackedColors(t *testing.T) {
	if got := RGB(0x9333ea).Color; got != (color.NRGBA{R: 0x93, G: 0x33, B: 0xea, A: 0xff}) {
		t.Fatalf("RGB = %#v", got)
	}
	if got := RGBA(0x9333eacc).Color; got != (color.NRGBA{R: 0x93, G: 0x33, B: 0xea, A: 0xcc}) {
		t.Fatalf("RGBA = %#v", got)
	}
}

func TestCascadeReplacesShadowsAndOutline(t *testing.T) {
	shadowColor := RGBA(0x33445566)
	focusColor := RGB(0x00ccff)
	base := Style{}.
		BoxShadow(0, 2, 4, 0, shadowColor).
		Outline(1, 2, focusColor).
		VisualOutset(1, 2, 3, 4)

	override := Style{}.
		BoxShadow(0, 6, 12, 1, RGBA(0x11223344)).
		VisualOutset(5, 6, 7, 8)

	resolved := Cascade(State{}, base, override)
	if resolved.Paint == nil || len(resolved.Paint.Shadows) != 1 {
		t.Fatalf("shadows = %#v, want one overriding layer", resolved.Paint)
	}
	if got := resolved.Paint.Shadows[0]; got.OffsetY != 6 || got.Blur != 12 || got.Color != RGBA(0x11223344) {
		t.Fatalf("shadow = %#v", got)
	}
	if resolved.Paint.Outline == nil || resolved.Paint.Outline.Width != 1 || resolved.Paint.Outline.Color != focusColor {
		t.Fatalf("outline = %#v", resolved.Paint.Outline)
	}
	if resolved.Paint.VisualOutset == nil || *resolved.Paint.VisualOutset != (Insets{Top: 5, Right: 6, Bottom: 7, Left: 8}) {
		t.Fatalf("visual outset = %#v", resolved.Paint.VisualOutset)
	}
}

func TestResolveAppliesOnWithoutMutatingSource(t *testing.T) {
	normal := RGB(0x0066ff)
	hover := RGB(0x3388ff)
	source := Style{}.
		Background(normal).
		On(Hovered, Style{}.Background(hover))

	resolved := source.Resolve(State{Hovered: true})
	if resolved.Paint == nil || resolved.Paint.Background != hover {
		t.Fatalf("resolved background = %#v, want %#v", resolved.Paint.Background, hover)
	}
	normalResolved := source.Resolve(State{})
	if normalResolved.Paint == nil || normalResolved.Paint.Background != normal {
		t.Fatalf("source background mutated: %#v", normalResolved.Paint)
	}
}

func TestRuleRowsCompose(t *testing.T) {
	declaration := Style{}.Rules(
		Rule{When: Hovered, Not: Disabled, Set: Style{}.Opacity(.5)},
		// Rule.When is a bitwise AND, so an any-of pair is two rows sharing a
		// Set, not one row with an ORed mask.
		Rule{When: Pressed, Set: Style{}.Scale(.9)},
		Rule{When: FocusVisible, Set: Style{}.Scale(.9)},
	)

	hovered := declaration.Resolve(State{Hovered: true})
	if hovered.Paint == nil || hovered.Paint.Opacity == nil || *hovered.Paint.Opacity != .5 {
		t.Fatalf("all condition = %#v", hovered.Paint)
	}
	disabled := declaration.Resolve(State{Hovered: true, Disabled: true})
	if disabled.Paint != nil && disabled.Paint.Opacity != nil {
		t.Fatalf("not condition matched disabled state: %#v", disabled.Paint)
	}
	focused := declaration.Resolve(State{FocusVisible: true})
	if focused.Trans == nil || focused.Trans.ScaleX == nil || *focused.Trans.ScaleX != .9 {
		t.Fatalf("any condition = %#v", focused.Trans)
	}
}

func TestCascadeCanClearBackgroundAndShadows(t *testing.T) {
	base := Style{}.
		Background(RGB(0x112233)).
		BoxShadow(0, 2, 4, 0, RGBA(0x00000044))

	resolved := Cascade(State{}, base, Style{}.BackgroundNone().BoxShadowNone().VisualOutsetNone())

	if resolved.Paint == nil || resolved.Paint.Background != nil || len(resolved.Paint.Shadows) != 0 || resolved.Paint.VisualOutset != nil {
		t.Fatalf("cleared paint = %#v", resolved.Paint)
	}
}

func TestResolveAppliesDistinctComponentStates(t *testing.T) {
	source := Style{}.
		On(FocusVisible, Style{}.Opacity(.9)).
		On(Checked, Style{}.Scale(.8)).
		On(Indeterminate, Style{}.Radius(3)).
		On(ReadOnly, Style{}.BorderWidth(2)).
		On(Open, Style{}.Padding(4)).
		On(Expanded, Style{}.Margin(5)).
		On(Dragging, Style{}.Translate(6, 7)).
		On(DropTarget, Style{}.Background(RGB(0x010203)))

	resolved := source.Resolve(State{
		FocusVisible: true, Checked: true, Indeterminate: true, ReadOnly: true,
		Open: true, Expanded: true, Dragging: true, DropTarget: true,
	})
	if resolved.Paint == nil || *resolved.Paint.Opacity != .9 || *resolved.Paint.Radius != 3 || *resolved.Paint.Border.Width != 2 {
		t.Fatalf("paint state overrides = %#v", resolved.Paint)
	}
	if resolved.Trans == nil || *resolved.Trans.ScaleX != .8 || *resolved.Trans.TranslateX != 6 {
		t.Fatalf("transform state overrides = %#v", resolved.Trans)
	}
	if resolved.Box == nil || resolved.Box.Padding.Top != 4 || resolved.Box.Margin.Top != 5 {
		t.Fatalf("box state overrides = %#v", resolved.Box)
	}
}

func TestCascadePartInheritsTextWithoutRootPaint(t *testing.T) {
	base := Style{}.
		TextColor(RGB(0x111111)).
		Background(RGB(0xeeeeee)).
		Part(PartLabel, Style{}.FontWeight(600))

	custom := Style{}.
		TextColor(RGB(0x222222)).
		Part(PartLabel, Style{}.
			On(Hovered, Style{}.TextColor(RGB(0x333333))))

	resolved := CascadePart(State{Hovered: true}, PartLabel, base, custom)
	if resolved.Paint != nil {
		t.Fatalf("root paint leaked into label: %#v", resolved.Paint)
	}
	if resolved.Text == nil || resolved.Text.Color != RGB(0x333333) || *resolved.Text.FontWeight != 600 {
		t.Fatalf("label style = %#v", resolved.Text)
	}
}

func TestCustomPartNamesAndStyleSnapshot(t *testing.T) {
	const thumb Part = "thumb"
	part := Style{}.Radius(4)
	built := Style{}.Part(thumb, part)
	part = part.Radius(8)
	_ = part

	resolved := CascadePart(State{}, thumb, built)
	if resolved.Paint == nil || resolved.Paint.Radius == nil || *resolved.Paint.Radius != 4 {
		t.Fatalf("custom part = %#v", resolved.Paint)
	}
}

func TestBuiltInControlPartsHaveStableNames(t *testing.T) {
	if PartTrack != "track" || PartThumb != "thumb" || PartIndicator != "indicator" {
		t.Fatalf("control parts = %q/%q/%q", PartTrack, PartThumb, PartIndicator)
	}
}

func TestPartRootAppliesToRootDeclaration(t *testing.T) {
	resolved := Style{}.
		Part(PartRoot, Style{}.Padding(8).Background(RGB(0x123456))).
		Resolve(State{})
	if resolved.Box == nil || resolved.Box.Padding == nil || resolved.Box.Padding.Left != 8 {
		t.Fatalf("root box = %#v", resolved.Box)
	}
	if resolved.Paint == nil || resolved.Paint.Background != RGB(0x123456) {
		t.Fatalf("root paint = %#v", resolved.Paint)
	}
}

func TestCascadeUsesLaterSetPropertiesOnly(t *testing.T) {
	base := Style{}.
		Padding(unit.Dp(8)).
		Background(RGB(0x111111)).
		Opacity(1)

	override := Style{}.
		PaddingX(unit.Dp(12)).
		Opacity(0)

	resolved := Cascade(State{}, base, override)
	if resolved.Box == nil || resolved.Box.Padding == nil {
		t.Fatal("cascade lost padding")
	}
	if resolved.Box.Padding.Top != unit.Dp(8) || resolved.Box.Padding.Left != unit.Dp(12) {
		t.Fatalf("padding = %#v, want top 8 and left 12", *resolved.Box.Padding)
	}
	if resolved.Paint == nil || resolved.Paint.Opacity == nil || *resolved.Paint.Opacity != 0 {
		t.Fatalf("opacity = %#v, want explicit zero", resolved.Paint.Opacity)
	}
}

func TestCascadeMergesIndividualCornerRadii(t *testing.T) {
	base := Style{}.Radius(8)
	override := Style{}.RadiusTopLeft(1).RadiusBottomRight(3)
	resolved := Cascade(State{}, base, override)

	want := CornerRadii{TopLeft: 1, TopRight: 8, BottomRight: 3, BottomLeft: 8}
	if resolved.Paint == nil || resolved.Paint.Radii == nil || *resolved.Paint.Radii != want {
		t.Fatalf("radii = %#v, want %#v", resolved.Paint, want)
	}

	reset := Cascade(State{}, base, override, Style{}.Radius(4))
	if reset.Paint.Radius == nil || *reset.Paint.Radius != 4 || reset.Paint.Radii != nil {
		t.Fatalf("uniform radius did not replace corners: %#v", reset.Paint)
	}
}

func TestStyleDeepCopiesGradientAndNestedRule(t *testing.T) {
	gradient := Gradient{
		Stops: []GradientStop{{Offset: 0, Color: SolidColor{Color: color.NRGBA{A: 0xff}}}},
	}
	source := Style{}.
		Background(gradient).
		On(Pressed, Style{}.Scale(0.95))

	gradient.Stops[0].Offset = 1
	storedSource := source.Resolve(State{})
	if storedSource.Paint == nil {
		t.Fatal("source lost paint")
	}
	stored := storedSource.Paint.Background.(Gradient)
	if stored.Stops[0].Offset != 0 {
		t.Fatalf("gradient was aliased: %v", stored.Stops[0].Offset)
	}

	resolved := source.Resolve(State{Pressed: true})
	if resolved.Trans == nil || resolved.Trans.ScaleX == nil || *resolved.Trans.ScaleX != 0.95 {
		t.Fatalf("pressed transform = %#v, want 0.95", resolved.Trans)
	}
}

func TestColorSourceWorksAcrossPaintAndText(t *testing.T) {
	accent := RGB(0x0066ff)
	foreground := RGB(0x101418)
	base := Style{}.BorderColor(accent).TextColor(foreground)
	resolved := Cascade(State{}, base, Style{}.BorderWidth(2))

	if resolved.Text == nil || resolved.Text.Color != foreground {
		t.Fatalf("text color = %#v, want %#v", resolved.Text, foreground)
	}
	if resolved.Paint == nil || resolved.Paint.Border == nil || resolved.Paint.Border.Color != accent {
		t.Fatalf("border color = %#v, want %#v", resolved.Paint, accent)
	}
	if resolved.Paint.Border.Width == nil || *resolved.Paint.Border.Width != 2 {
		t.Fatalf("border width = %#v, want 2", resolved.Paint.Border.Width)
	}
}

func TestBottomBorderFluentStyle(t *testing.T) {
	accent := RGB(0x1677ff)
	declaration := Style{}.BorderBottom(1, accent)
	resolved := declaration.Resolve(State{})
	if resolved.Paint == nil || resolved.Paint.BorderBottom == nil {
		t.Fatalf("bottom border = %#v", resolved.Paint)
	}
	if resolved.Paint.BorderBottom.Color != accent || resolved.Paint.BorderBottom.Width == nil || *resolved.Paint.BorderBottom.Width != 1 {
		t.Fatalf("bottom border = %#v", resolved.Paint.BorderBottom)
	}

	changed := declaration.BorderBottom(2, RGB(0xff4d4f))
	base := declaration.Resolve(State{})
	if base.Paint.BorderBottom.Color != accent || *base.Paint.BorderBottom.Width != 1 {
		t.Fatalf("source bottom border changed: %#v", base.Paint.BorderBottom)
	}
	if changed.Hash64() == declaration.Hash64() {
		t.Fatal("bottom border changes were not included in style hash")
	}
}

func TestTransientShadowChangesHash(t *testing.T) {
	normal := New().BoxShadow(0, 0, 5, 0, RGB(0x3e63dd))
	transient := New().BoxShadowTransient(0, 0, 5, 0, RGB(0x3e63dd))
	if normal.Hash64() == transient.Hash64() {
		t.Fatal("transient shadow was not included in the style hash")
	}
}

func TestPartHashIsIndependentOfInsertionOrder(t *testing.T) {
	for _, count := range []int{1, 8, 9} {
		t.Run(string(rune('0'+count)), func(t *testing.T) {
			forward, reverse := Style{}, Style{}
			for i := range count {
				forward = forward.Part(Part(string(rune('a'+i))), Style{}.Width(unit.Dp(i+1)))
			}
			for i := count - 1; i >= 0; i-- {
				reverse = reverse.Part(Part(string(rune('a'+i))), Style{}.Width(unit.Dp(i+1)))
			}
			want := forward.Hash64()
			for range 20 {
				if got := reverse.Hash64(); got != want {
					t.Fatalf("%d parts hash = %#x, want %#x", count, got, want)
				}
			}
			changed := reverse.Part(Part("a"), Style{}.Width(100))
			if changed.Hash64() == want {
				t.Fatal("changing a Part did not change the hash")
			}
		})
	}
}

func TestResolveDeepCopiesPointerSources(t *testing.T) {
	background := &SolidColor{Color: color.NRGBA{R: 1, A: 0xff}}
	source := Style{
		paint: &PaintStyle{Background: background},
	}
	built := source.Resolve(State{})

	background.Color.R = 2
	if got := built.Paint.Background.(*SolidColor).Color.R; got != 1 {
		t.Fatalf("background red = %d, want 1", got)
	}
}

func TestMutatingResolvedStyleDoesNotChangeDeclaration(t *testing.T) {
	declaration := Style{}.
		Padding(8).
		Background(RGB(0x112233)).
		Opacity(.75).
		BoxShadow(0, 2, 4, 0, RGBA(0x00000040))

	resolved := declaration.Resolve(State{})
	resolved.Box.Padding.Left = 99
	*resolved.Paint.Opacity = 0
	resolved.Paint.Background = RGB(0xffffff)
	resolved.Paint.Shadows[0].Blur = 99

	again := declaration.Resolve(State{})
	if again.Box.Padding.Left != 8 || *again.Paint.Opacity != .75 || again.Paint.Background != RGB(0x112233) || again.Paint.Shadows[0].Blur != 4 {
		t.Fatalf("resolved mutation leaked into declaration: %#v", again)
	}
}

func TestFluentBranchesDoNotMutateTheirSource(t *testing.T) {
	base := Style{}.
		BorderColor(RGB(0x112233)).
		BorderWidth(1).
		BoxShadow(0, 2, 4, 0, RGBA(0x00000040)).
		On(Hovered, Style{}.Opacity(.8)).
		Part(PartLabel, Style{}.TextColor(RGB(0x445566)))
	changed := base.
		BorderWidth(3).
		BoxShadow(0, 4, 8, 0, RGBA(0x00000080)).
		On(Pressed, Style{}.Scale(.95)).
		Part(PartLabel, Style{}.FontWeight(600))

	baseResolved := base.Resolve(State{Hovered: true, Pressed: true})
	if baseResolved.Paint == nil || *baseResolved.Paint.Border.Width != 1 || len(baseResolved.Paint.Shadows) != 1 || baseResolved.Trans != nil {
		t.Fatalf("source branch changed: %#v", baseResolved)
	}
	baseLabel := CascadePart(State{}, PartLabel, base)
	if baseLabel.Text == nil || baseLabel.Text.FontWeight != nil || baseLabel.Text.Color != RGB(0x445566) {
		t.Fatalf("source part changed: %#v", baseLabel.Text)
	}

	changedResolved := changed.Resolve(State{Hovered: true, Pressed: true})
	if *changedResolved.Paint.Border.Width != 3 || len(changedResolved.Paint.Shadows) != 2 || changedResolved.Trans == nil || *changedResolved.Trans.ScaleX != .95 {
		t.Fatalf("changed branch = %#v", changedResolved)
	}
	changedLabel := CascadePart(State{}, PartLabel, changed)
	if changedLabel.Text == nil || changedLabel.Text.FontWeight == nil || *changedLabel.Text.FontWeight != 600 {
		t.Fatalf("changed part = %#v", changedLabel.Text)
	}
}
