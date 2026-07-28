package alert

import (
	"bytes"
	"image"
	"image/color"
	"testing"
	"time"

	"gioui.org/f32"
	gioinput "gioui.org/io/input"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/flowui-icons-lucide"
	"github.com/qianniancn/flowui/internal/components/button"
	"github.com/qianniancn/flowui/internal/frame"
	"github.com/qianniancn/flowui/internal/locale"
	flowstyle "github.com/qianniancn/flowui/internal/style"
	"github.com/qianniancn/flowui/internal/theme"
)

type alertProbe struct {
	size       image.Point
	foreground color.NRGBA
	background color.NRGBA
	layouts    int
}

func (p *alertProbe) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	p.layouts++
	p.foreground = ctx.ForegroundColor()
	p.background = ctx.BackgroundColor()
	return layout.Dimensions{Size: gtx.Constraints.Constrain(p.size)}
}

func newAlertContext(activeTheme *theme.Theme) *frame.Context {
	return frame.New(nil, activeTheme, locale.LanguageEnglish)
}

func alertTestContext() layout.Context {
	var ops op.Ops
	return layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(480, 200)},
		Ops:         &ops,
	}
}

func TestAlertStylePartsSeparateRootLabelAndIndicator(t *testing.T) {
	rootColor := color.NRGBA{R: 1, A: 0xff}
	labelColor := color.NRGBA{G: 2, A: 0xff}
	indicatorColor := color.NRGBA{B: 3, A: 0xff}
	descriptionColor := color.NRGBA{R: 4, A: 0xff}
	resolved := New("Title", "Description").Style(
		flowstyle.Style{}.
			TextColor(flowstyle.SolidColor{Color: rootColor}).
			Part(flowstyle.PartLabel, flowstyle.Style{}.TextColor(flowstyle.SolidColor{Color: labelColor})).
			Part(flowstyle.PartDescription, flowstyle.Style{}.TextColor(flowstyle.SolidColor{Color: descriptionColor})).
			Part(flowstyle.PartIndicator, flowstyle.Style{}.TextColor(flowstyle.SolidColor{Color: indicatorColor})),
	).resolveStyle(newAlertContext(nil), alertTestContext())

	rootText := resolved.root.Text.Color.(flowstyle.SolidColor).Color
	titleText := resolved.title.Text.Color.(flowstyle.SolidColor).Color
	descriptionText := resolved.description.Text.Color.(flowstyle.SolidColor).Color
	indicatorText := resolved.indicator.Text.Color.(flowstyle.SolidColor).Color
	if rootText != rootColor || titleText != labelColor || descriptionText != descriptionColor || indicatorText != indicatorColor {
		t.Fatalf("alert parts = root %#v title %#v description %#v indicator %#v", rootText, titleText, descriptionText, indicatorText)
	}
}

func TestAlertConditionalTransitionsAnimateRootAndPart(t *testing.T) {
	ctx := newAlertContext(nil)
	start := time.Unix(1, 0)
	rootFrom := flowstyle.RGB(0x102030).Color
	rootTo := flowstyle.RGB(0x8090a0).Color
	labelFrom := flowstyle.RGB(0x203040).Color
	labelTo := flowstyle.RGB(0x90a0b0).Color
	resolveAt := func(now time.Time, active bool) (color.NRGBA, color.NRGBA) {
		frame.BeginFrame(ctx)
		resolved := New("Status", "").Style(
			flowstyle.Style{}.
				Background(flowstyle.SolidColor{Color: rootFrom}).
				Transition(flowstyle.PropBackgroundColor, 100*time.Millisecond).
				WhenIf(active, flowstyle.Style{}.Background(flowstyle.SolidColor{Color: rootTo})).
				Part(flowstyle.PartLabel, flowstyle.Style{}.
					TextColor(flowstyle.SolidColor{Color: labelFrom}).
					Transition(flowstyle.PropTextColor, 100*time.Millisecond).
					WhenIf(active, flowstyle.Style{}.TextColor(flowstyle.SolidColor{Color: labelTo}))),
		).resolveStyle(ctx, layout.Context{Ops: new(op.Ops), Now: now})
		frame.EndFrame(ctx)
		root := resolved.root.Paint.Background.(flowstyle.SolidColor).Color
		label := resolved.title.Text.Color.(flowstyle.SolidColor).Color
		return root, label
	}

	if root, label := resolveAt(start, false); root != rootFrom || label != labelFrom {
		t.Fatalf("initial colors = %#v/%#v", root, label)
	}
	if root, label := resolveAt(start, true); root != rootFrom || label != labelFrom {
		t.Fatalf("transition start = %#v/%#v", root, label)
	}
	root, label := resolveAt(start.Add(50*time.Millisecond), true)
	if root == rootFrom || root == rootTo || label == labelFrom || label == labelTo {
		t.Fatalf("transition midpoint = %#v/%#v", root, label)
	}
	if root, label := resolveAt(start.Add(100*time.Millisecond), true); root != rootTo || label != labelTo {
		t.Fatalf("transition end = %#v/%#v", root, label)
	}
}

func TestAlertHeroUIDefaultTheme(t *testing.T) {
	tokens := theme.DefaultTheme().Components.Alert
	if tokens.PaddingX != 16 || tokens.PaddingY != 12 || tokens.Gap != 16 || tokens.Radius != 24 {
		t.Fatalf("alert geometry = %#v", tokens)
	}
	if tokens.IndicatorPadding != 4 || tokens.IconSize != 16 {
		t.Fatalf("alert indicator geometry = %#v", tokens)
	}
	if tokens.TitleSize != 14 || tokens.TitleLineHeight != 24 || tokens.DescriptionSize != 14 || tokens.DescriptionLineHeight != 20 {
		t.Fatalf("alert typography = %#v", tokens)
	}
}

func TestAlertFillsWidthAndMatchesHeroUIHeight(t *testing.T) {
	dims := New("New features available", "Check out the latest updates.").Layout(newAlertContext(nil), alertTestContext())
	if dims.Size.X != 480 {
		t.Fatalf("alert width = %d, want 480", dims.Size.X)
	}
	if dims.Size.Y != 68 {
		t.Fatalf("two-line alert height = %d, want 68", dims.Size.Y)
	}

	dims = New("Profile updated successfully", "").Layout(newAlertContext(nil), alertTestContext())
	if dims.Size.Y != 48 {
		t.Fatalf("title-only alert height = %d, want 48", dims.Size.Y)
	}
}

func TestAlertStatusColorsMatchHeroUI(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	tests := []struct {
		name   string
		status Status
		want   color.NRGBA
	}{
		{name: "default", status: StatusDefault, want: activeTheme.Palette.SurfaceForeground},
		{name: "accent", status: StatusAccent, want: activeTheme.Palette.AccentSoftForeground},
		{name: "success", status: StatusSuccess, want: activeTheme.Palette.SuccessSoftForeground},
		{name: "warning", status: StatusWarning, want: activeTheme.Palette.WarningSoftForeground},
		{name: "danger", status: StatusDanger, want: activeTheme.Palette.DangerSoftForeground},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			style := alertStyleFor(&activeTheme, test.status)
			if style.title != test.want || style.indicator != test.want {
				t.Fatalf("status colors = title %#v indicator %#v, want %#v", style.title, style.indicator, test.want)
			}
			if style.background != activeTheme.Palette.Surface || style.description != activeTheme.Palette.MutedForeground {
				t.Fatalf("shared alert colors = %#v", style)
			}
		})
	}
}

func TestAlertStatusColorsHonorTransparentPaletteValues(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	activeTheme.Palette.SurfaceForeground = color.NRGBA{}
	activeTheme.Palette.AccentSoftForeground = color.NRGBA{}
	activeTheme.Palette.SuccessSoftForeground = color.NRGBA{}
	activeTheme.Palette.WarningSoftForeground = color.NRGBA{}
	activeTheme.Palette.DangerSoftForeground = color.NRGBA{}
	if got := alertStyleFor(&activeTheme, StatusDefault).title; got.A != 0 {
		t.Fatalf("default color = %#v, want transparent", got)
	}
	if got := alertStyleFor(&activeTheme, StatusAccent).title; got.A != 0 {
		t.Fatalf("accent color = %#v, want transparent", got)
	}
	if got := alertStyleFor(&activeTheme, StatusSuccess).title; got.A != 0 {
		t.Fatalf("success color = %#v, want transparent", got)
	}
	if got := alertStyleFor(&activeTheme, StatusWarning).title; got.A != 0 {
		t.Fatalf("warning color = %#v, want transparent", got)
	}
	if got := alertStyleFor(&activeTheme, StatusDanger).title; got.A != 0 {
		t.Fatalf("danger color = %#v, want transparent", got)
	}
}

func TestAlertUsesLucideStatusIcons(t *testing.T) {
	tests := []struct {
		status Status
		want   []byte
	}{
		{status: StatusDefault, want: lucide.Info},
		{status: StatusAccent, want: lucide.Info},
		{status: StatusSuccess, want: lucide.CircleCheck},
		{status: StatusWarning, want: lucide.TriangleAlert},
		{status: StatusDanger, want: lucide.CircleAlert},
	}
	for _, test := range tests {
		if got := alertIcon(test.status); !bytes.Equal(got, test.want) {
			t.Fatalf("status %v uses the wrong Lucide icon", test.status)
		}
	}
}

func TestAlertCustomSlotsReceiveScopedColors(t *testing.T) {
	indicator := &alertProbe{size: image.Pt(16, 16)}
	content := &alertProbe{size: image.Pt(120, 20)}
	action := &alertProbe{size: image.Pt(64, 32)}
	activeTheme := theme.DefaultTheme()
	New("Processing", "ignored").
		Status(StatusAccent).
		Indicator(indicator).
		Content(content).
		Action(action).
		Layout(newAlertContext(&activeTheme), alertTestContext())

	if indicator.layouts != 1 || content.layouts != 1 || action.layouts != 1 {
		t.Fatalf("slot layouts = indicator %d content %d action %d", indicator.layouts, content.layouts, action.layouts)
	}
	if indicator.foreground != activeTheme.Palette.AccentSoftForeground {
		t.Fatalf("indicator foreground = %#v", indicator.foreground)
	}
	if content.foreground != activeTheme.Palette.MutedForeground {
		t.Fatalf("content foreground = %#v", content.foreground)
	}
	if action.foreground != activeTheme.Palette.SurfaceForeground {
		t.Fatalf("action foreground = %#v", action.foreground)
	}
	for name, got := range map[string]color.NRGBA{
		"indicator": indicator.background,
		"content":   content.background,
		"action":    action.background,
	} {
		if got != activeTheme.Palette.Surface {
			t.Fatalf("%s background = %#v", name, got)
		}
	}
}

func TestAlertSemanticsExposeTitleAndDescription(t *testing.T) {
	ctx := newAlertContext(nil)
	router := new(gioinput.Router)
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(480, 200)},
		Source:      router.Source(),
		Ops:         &ops,
	}
	New("Storage almost full", "You are using 90% of your quota.").Layout(ctx, gtx)
	router.Frame(&ops)
	if !alertSemanticTreeContains(router.AppendSemantics(nil), "Storage almost full", "You are using 90% of your quota.") {
		t.Fatal("alert semantics did not expose its title and description")
	}
}

func TestAlertActionButtonRemainsInteractive(t *testing.T) {
	ctx := newAlertContext(nil)
	router := new(gioinput.Router)
	clicked := false
	action := button.Button("refresh", &alertProbe{size: image.Pt(64, 16)}).
		Size(button.ButtonSmall).
		OnClick(func() { clicked = true })
	alert := New("Update available", "Refresh to get the latest version.").Action(action)
	layoutAlertTestFrame(ctx, router, alert, time.Unix(1, 0))
	router.Queue(
		pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: 1, Buttons: pointer.ButtonPrimary, Position: f32.Pt(432, 28)},
		pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, PointerID: 1, Position: f32.Pt(432, 28)},
	)
	layoutAlertTestFrame(ctx, router, alert, time.Unix(1, int64(time.Millisecond)))
	if !clicked {
		t.Fatal("alert action button did not receive the pointer click")
	}
}

func layoutAlertTestFrame(ctx *frame.Context, router *gioinput.Router, alert Widget, now time.Time) {
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(480, 200)},
		Source:      router.Source(),
		Ops:         &ops,
		Now:         now,
	}
	frame.BeginFrame(ctx)
	alert.Layout(ctx, gtx)
	frame.ApplyFrameCommands(ctx, gtx)
	frame.EndFrame(ctx)
	router.Frame(&ops)
}

func alertSemanticTreeContains(nodes []gioinput.SemanticNode, label, description string) bool {
	for _, node := range nodes {
		if node.Desc.Label == label && node.Desc.Description == description {
			return true
		}
		if alertSemanticTreeContains(node.Children, label, description) {
			return true
		}
	}
	return false
}
