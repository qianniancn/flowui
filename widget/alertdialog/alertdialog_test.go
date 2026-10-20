package alertdialog

import (
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
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/testtheme"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/modal"
)

type probeWidget struct {
	size       image.Point
	layouts    int
	foreground color.NRGBA
	background color.NRGBA
}

func (p *probeWidget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	p.layouts++
	p.foreground = ctx.ForegroundColor()
	p.background = ctx.BackgroundColor()
	return layout.Dimensions{Size: gtx.Constraints.Constrain(p.size)}
}

func TestAlertDialogDefaults(t *testing.T) {
	dialog := New("Delete project?", "This action cannot be undone.").Key("delete").Open(true)
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
}

func TestAlertDialogMapsContainerVariantsToModal(t *testing.T) {
	sizes := map[Size]modal.Size{
		SizeMedium: modal.Medium,
		SizeXSmall: modal.XSmall,
		SizeSmall:  modal.Small,
		SizeLarge:  modal.Large,
		SizeCover:  modal.Cover,
	}
	for value, want := range sizes {
		if got := New("Title", "Body").Key("test").Open(true).Size(value).modalSize(); got != want {
			t.Errorf("size %v maps to %v, want %v", value, got, want)
		}
	}
	placements := map[Placement]modal.Placement{
		PlacementAuto:   modal.Auto,
		PlacementTop:    modal.Top,
		PlacementCenter: modal.Center,
		PlacementBottom: modal.Bottom,
	}
	for value, want := range placements {
		if got := New("Title", "Body").Key("test").Open(true).Placement(value).modalPlacement(); got != want {
			t.Errorf("placement %v maps to %v, want %v", value, got, want)
		}
	}
	backdrops := map[BackdropVariant]modal.BackdropVariant{
		BackdropOpaque:      modal.BackdropOpaque,
		BackdropBlur:        modal.BackdropBlur,
		BackdropTransparent: modal.BackdropTransparent,
	}
	for value, want := range backdrops {
		if got := New("Title", "Body").Key("test").Open(true).Backdrop(value).modalBackdrop(); got != want {
			t.Errorf("backdrop %v maps to %v, want %v", value, got, want)
		}
	}
}

func TestAlertDialogCustomIconReceivesStatusColors(t *testing.T) {
	activeTheme := testtheme.Behavior()
	provider, err := theme.ComponentProvider(activeTheme)
	if err != nil {
		t.Fatal(err)
	}
	visual := provider.AlertDialog(activeTheme, theme.AlertDialogRequest{Status: theme.AlertSuccess})
	ctx := core.New(nil, activeTheme, locale.LanguageEnglish)
	customIcon := &probeWidget{size: image.Pt(20, 20)}
	gtx := alertDialogTestLayoutContext(nil, time.Time{})
	gtx.Constraints = layout.Constraints{Max: image.Pt(200, 100)}
	header := dialogHeader{title: "Complete task?", status: StatusSuccess, icon: customIcon}
	header.Layout(ctx, gtx)

	if customIcon.layouts != 1 {
		t.Fatalf("custom icon layouts = %d, want 1", customIcon.layouts)
	}
	if customIcon.foreground != visual.IconForeground || customIcon.background != visual.IconBackground {
		t.Fatalf("custom icon colors = %#v/%#v, want provider colors %#v/%#v", customIcon.foreground, customIcon.background, visual.IconForeground, visual.IconBackground)
	}
}

func TestAlertDialogCustomBodyDoesNotExposeReplacedDescription(t *testing.T) {
	customBody := &probeWidget{size: image.Pt(100, 24)}
	dialog := New("Custom body", "Default description").Key("custom").Open(true).Body(customBody)
	header := dialog.defaultHeader(nil)
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
	ctx := core.New(nil, testtheme.Behavior(), locale.LanguageEnglish)
	New("Delete project?", "This action cannot be undone.").Key("delete").Open(true).defaultHeader(nil).Layout(ctx, gtx)
	router.Frame(&ops)
	if !semanticTreeContains(router.AppendSemantics(nil), "Delete project?", "This action cannot be undone.") {
		t.Fatal("alert dialog header semantics did not expose its title and description")
	}
}

func TestAlertDialogDefaultsBlockBackdropAndEscape(t *testing.T) {
	ctx := core.New(nil, testtheme.Behavior(), locale.LanguageEnglish)
	router := new(input.Router)
	closeRequests := 0
	dialog := New("Delete?", "This action cannot be undone.").Key("locked").Open(true).
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
			ctx := core.New(nil, testtheme.Behavior(), locale.LanguageEnglish)
			router := new(input.Router)
			closeRequests := 0
			dialog := configure(New("Confirm?", "Choose an action.").Key(name).Open(true)).
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

func layoutAlertDialogFrame(ctx *core.Context, router *input.Router, dialog Widget, now time.Time) {
	gtx := alertDialogTestLayoutContext(router, now)
	core.BeginFrame(ctx)
	dialog.Layout(ctx, gtx)
	core.LayoutOverlays(ctx, gtx)
	core.ApplyFrameCommands(ctx, gtx)
	core.EndFrame(ctx)
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

func TestUncontrolledPreservesClosedInitialState(t *testing.T) {
	dialog := New("", "").Key("dialog").Uncontrolled()
	if dialog.hasOpen || !dialog.hasDefaultOpen || dialog.defaultOpen {
		t.Fatalf("Uncontrolled() = hasOpen=%v hasDefault=%v defaultOpen=%v", dialog.hasOpen, dialog.hasDefaultOpen, dialog.defaultOpen)
	}
}
