package alertdialog

import (
	"bytes"
	"image"
	"image/color"
	"testing"
	"time"

	"gioui.org/f32"
	"gioui.org/io/input"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/flowui-icons-lucide"
	"github.com/qianniancn/flowui/internal/components/modal"
	"github.com/qianniancn/flowui/internal/frame"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/theme"
)

type probeWidget struct {
	size       image.Point
	layouts    int
	foreground color.NRGBA
	background color.NRGBA
}

func (p *probeWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	p.layouts++
	p.foreground = ctx.ForegroundColor()
	p.background = ctx.BackgroundColor()
	return layout.Dimensions{Size: gtx.Constraints.Constrain(p.size)}
}

func TestAlertDialogDefaultsMatchHeroUI(t *testing.T) {
	dialog := New("delete", true, "Delete project?", "This action cannot be undone.")
	if dialog.status != StatusDanger {
		t.Fatalf("default status = %v, want danger", dialog.status)
	}
	if dialog.dismissable {
		t.Fatal("alert dialog is dismissable by default")
	}
	if !dialog.keyboardDismissDisabled {
		t.Fatal("alert dialog allows Escape dismissal by default")
	}
	if !dialog.closeButton {
		t.Fatal("alert dialog default example should include the close trigger")
	}

	tokens := theme.DefaultTheme().Components.AlertDialog
	if tokens.IconSize != 40 || tokens.IconGlyphSize != 20 || tokens.HeaderGap != 12 || tokens.TitleSize != 16 {
		t.Fatalf("geometry = %+v, want icon 40/20, gap 12, title 16", tokens)
	}
}

func TestAlertDialogMapsContainerVariantsToModal(t *testing.T) {
	sizes := map[Size]modal.ModalSize{
		SizeMedium: modal.ModalMedium,
		SizeXSmall: modal.ModalXSmall,
		SizeSmall:  modal.ModalSmall,
		SizeLarge:  modal.ModalLarge,
		SizeCover:  modal.ModalCover,
	}
	for value, want := range sizes {
		if got := New("test", true, "Title", "Body").Size(value).modalSize(); got != want {
			t.Errorf("size %v maps to %v, want %v", value, got, want)
		}
	}
	placements := map[Placement]modal.ModalPlacement{
		PlacementAuto:   modal.ModalAuto,
		PlacementTop:    modal.ModalTop,
		PlacementCenter: modal.ModalCenter,
		PlacementBottom: modal.ModalBottom,
	}
	for value, want := range placements {
		if got := New("test", true, "Title", "Body").Placement(value).modalPlacement(); got != want {
			t.Errorf("placement %v maps to %v, want %v", value, got, want)
		}
	}
	backdrops := map[BackdropVariant]modal.ModalBackdropVariant{
		BackdropOpaque:      modal.ModalBackdropOpaque,
		BackdropBlur:        modal.ModalBackdropBlur,
		BackdropTransparent: modal.ModalBackdropTransparent,
	}
	for value, want := range backdrops {
		if got := New("test", true, "Title", "Body").Backdrop(value).modalBackdrop(); got != want {
			t.Errorf("backdrop %v maps to %v, want %v", value, got, want)
		}
	}
}

func TestAlertDialogStatusIconsUseLucide(t *testing.T) {
	tests := []struct {
		status Status
		want   []byte
	}{
		{StatusDefault, lucide.Info},
		{StatusAccent, lucide.Info},
		{StatusSuccess, lucide.CircleCheck},
		{StatusWarning, lucide.TriangleAlert},
		{StatusDanger, lucide.CircleAlert},
	}
	for _, test := range tests {
		if got := alertDialogIcon(test.status); !bytes.Equal(got, test.want) {
			t.Errorf("status %v uses the wrong Lucide icon", test.status)
		}
	}
}

func TestAlertDialogStatusStylesUseSoftSemanticColors(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	tests := []struct {
		status Status
		bg     color.NRGBA
		fg     color.NRGBA
	}{
		{StatusDefault, activeTheme.Palette.SurfaceSecondary, activeTheme.Palette.OverlayForeground},
		{StatusAccent, activeTheme.Palette.AccentSoft, activeTheme.Palette.AccentSoftForeground},
		{StatusSuccess, activeTheme.Palette.SuccessSoft, activeTheme.Palette.SuccessSoftForeground},
		{StatusWarning, activeTheme.Palette.WarningSoft, activeTheme.Palette.WarningSoftForeground},
		{StatusDanger, activeTheme.Palette.DangerSoft, activeTheme.Palette.DangerSoftForeground},
	}
	for _, test := range tests {
		style := alertDialogStyleFor(&activeTheme, test.status)
		if style.iconBackground != test.bg || style.iconForeground != test.fg {
			t.Errorf("status %v style = %#v/%#v, want %#v/%#v", test.status, style.iconBackground, style.iconForeground, test.bg, test.fg)
		}
	}
}

func TestAlertDialogCustomIconReceivesStatusColors(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	ctx := frame.New(nil, &activeTheme, locale.LanguageEnglish)
	customIcon := &probeWidget{size: image.Pt(20, 20)}
	gtx := alertDialogTestLayoutContext(nil, time.Time{})
	gtx.Constraints = layout.Constraints{Max: image.Pt(200, 100)}
	header := dialogHeader{title: "Complete task?", status: StatusSuccess, icon: customIcon}
	header.Layout(ctx, gtx)

	style := alertDialogStyleFor(&activeTheme, StatusSuccess)
	if customIcon.layouts != 1 {
		t.Fatalf("custom icon layouts = %d, want 1", customIcon.layouts)
	}
	if customIcon.foreground != style.iconForeground || customIcon.background != style.iconBackground {
		t.Fatalf("custom icon colors = %#v/%#v, want %#v/%#v", customIcon.foreground, customIcon.background, style.iconForeground, style.iconBackground)
	}
}

func TestAlertDialogCustomBodyDoesNotExposeReplacedDescription(t *testing.T) {
	customBody := &probeWidget{size: image.Pt(100, 24)}
	dialog := New("custom", true, "Custom body", "Default description").Body(customBody)
	header := dialog.defaultHeader()
	if header.description != "" {
		t.Fatalf("custom body retained replaced semantic description %q", header.description)
	}
}

func TestAlertDialogHeaderExposesTitleAndDescriptionSemantics(t *testing.T) {
	router := new(input.Router)
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 120)},
		Source:      router.Source(),
		Ops:         &ops,
	}
	ctx := frame.New(nil, nil, locale.LanguageEnglish)
	New("delete", true, "Delete project?", "This action cannot be undone.").defaultHeader().Layout(ctx, gtx)
	router.Frame(&ops)
	if !semanticTreeContains(router.AppendSemantics(nil), "Delete project?", "This action cannot be undone.") {
		t.Fatal("alert dialog header semantics did not expose its title and description")
	}
}

func TestAlertDialogDefaultsBlockBackdropAndEscape(t *testing.T) {
	ctx := frame.New(nil, nil, locale.LanguageEnglish)
	router := new(input.Router)
	closeRequests := 0
	dialog := New("locked", true, "Delete?", "This action cannot be undone.").
		OnOpenChange(func(open bool) {
			if !open {
				closeRequests++
			}
		})
	start := time.Unix(1, 0)
	layoutAlertDialogFrame(ctx, router, dialog, start)
	router.Queue(
		pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: 1, Buttons: pointer.ButtonPrimary, Position: f32.Pt(4, 4)},
		pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, PointerID: 1, Position: f32.Pt(4, 4)},
		key.Event{Name: key.NameEscape, State: key.Press},
	)
	layoutAlertDialogFrame(ctx, router, dialog, start.Add(10*time.Millisecond))
	if closeRequests != 0 {
		t.Fatalf("default alert dialog requested close %d times from backdrop or Escape", closeRequests)
	}
}

func TestAlertDialogCanEnableBackdropAndEscapeDismissal(t *testing.T) {
	for name, configure := range map[string]func(Widget) Widget{
		"backdrop": func(dialog Widget) Widget { return dialog.Dismissable(true) },
		"escape":   func(dialog Widget) Widget { return dialog.KeyboardDismissDisabled(false) },
	} {
		t.Run(name, func(t *testing.T) {
			ctx := frame.New(nil, nil, locale.LanguageEnglish)
			router := new(input.Router)
			closeRequests := 0
			dialog := configure(New(name, true, "Confirm?", "Choose an action.")).
				OnOpenChange(func(open bool) {
					if !open {
						closeRequests++
					}
				})
			start := time.Unix(1, 0)
			layoutAlertDialogFrame(ctx, router, dialog, start)
			if name == "backdrop" {
				router.Queue(
					pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: 1, Buttons: pointer.ButtonPrimary, Position: f32.Pt(4, 4)},
					pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, PointerID: 1, Position: f32.Pt(4, 4)},
				)
			} else {
				router.Queue(key.Event{Name: key.NameEscape, State: key.Press})
			}
			layoutAlertDialogFrame(ctx, router, dialog, start.Add(10*time.Millisecond))
			if closeRequests != 1 {
				t.Fatalf("close requests = %d, want 1", closeRequests)
			}
		})
	}
}

func alertDialogTestLayoutContext(router *input.Router, now time.Time) layout.Context {
	var source input.Source
	if router != nil {
		source = router.Source()
	}
	var ops op.Ops
	return layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 240)},
		Source:      source,
		Ops:         &ops,
		Now:         now,
	}
}

func layoutAlertDialogFrame(ctx *frame.Context, router *input.Router, dialog Widget, now time.Time) {
	gtx := alertDialogTestLayoutContext(router, now)
	frame.BeginFrame(ctx)
	dialog.Layout(ctx, gtx)
	frame.LayoutOverlays(ctx, gtx)
	frame.ApplyFrameCommands(ctx, gtx)
	frame.EndFrame(ctx)
	router.Frame(gtx.Ops)
}

func semanticTreeContains(nodes []input.SemanticNode, label, description string) bool {
	for _, node := range nodes {
		if node.Desc.Label == label && node.Desc.Description == description {
			return true
		}
		if semanticTreeContains(node.Children, label, description) {
			return true
		}
	}
	return false
}
