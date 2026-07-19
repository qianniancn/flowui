package toast

import (
	"image"
	"image/color"
	"testing"
	"time"

	"gioui.org/f32"
	"gioui.org/io/input"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/FlowUI/internal/components/button"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/locale"
	"github.com/qianniancn/FlowUI/internal/theme"
)

func TestToastOptionsUseValueSemantics(t *testing.T) {
	indicator := new(toastProbe)
	base := Toast("saved", "Saved")
	configured := base.
		Description("Changes saved").
		Variant(ToastSuccess).
		Indicator(indicator).
		Loading(true).
		Action("Undo").
		ActionVariant(button.ButtonDangerSoft).
		Timeout(3 * time.Second)

	if base.key != "saved" || base.title != "Saved" || base.description != "" || base.variant != ToastDefault || base.hasIndicator || base.loading || base.actionLabel != "" || base.actionVariant != button.ButtonPrimary || base.hasTimeout {
		t.Fatalf("base toast was mutated: %#v", base)
	}
	if configured.Key() != "saved" || configured.description != "Changes saved" || configured.variant != ToastSuccess || configured.indicator != indicator || !configured.loading {
		t.Fatalf("configured toast = %#v", configured)
	}
	if configured.actionLabel != "Undo" || configured.actionVariant != button.ButtonDangerSoft || configured.timeout != 3*time.Second || !configured.hasTimeout {
		t.Fatalf("configured toast action/timeout = %#v", configured)
	}
	if Toast("persistent", "Persistent").Timeout(-time.Second).timeout != 0 {
		t.Fatal("negative timeout was not clamped")
	}
}

func TestToastIndicatorNilHidesIndicator(t *testing.T) {
	if !Toast("default", "Default").showIndicator() {
		t.Fatal("default indicator is hidden")
	}
	if Toast("hidden", "Hidden").Indicator(nil).showIndicator() {
		t.Fatal("nil custom indicator did not hide the indicator")
	}
}

func TestToastProviderOptions(t *testing.T) {
	base := ToastProvider("toasts", nil)
	configured := base.
		Placement(ToastTopEnd).
		Offset(24).
		Gap(8).
		MaxVisible(5).
		ScaleFactor(0.1).
		Width(360).
		Paused(true).
		OnAction(func(string) {}).
		OnClose(func(string) {})

	if base.placement != ToastBottom || base.hasOffset || base.hasGap || base.hasMaxVisible || base.hasScale || base.hasWidth || base.paused {
		t.Fatalf("base provider was mutated: %#v", base)
	}
	if configured.placement != ToastTopEnd || configured.offset != 24 || configured.gap != 8 || configured.maxVisible != 5 || configured.scaleFactor != 0.1 || configured.width != 360 || !configured.paused {
		t.Fatalf("configured provider = %#v", configured)
	}
	if configured.onAction == nil || configured.onClose == nil {
		t.Fatal("provider callbacks are missing")
	}
}

func TestToastHeroUIDefaultTheme(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	tokens := activeTheme.Components.Toast
	if tokens.Width != 460 || tokens.Inset != 16 || tokens.Gap != 12 || tokens.Radius != 24 {
		t.Fatalf("toast geometry = %#v", tokens)
	}
	if tokens.MaxVisible != 3 || tokens.ScaleFactor != 0.05 || tokens.DefaultTimeout != 4*time.Second || tokens.AnimationDuration != 350*time.Millisecond {
		t.Fatalf("toast behavior tokens = %#v", tokens)
	}
	if activeTheme.Palette.SuccessSoftForeground != (color.NRGBA{R: 0x2b, G: 0x77, B: 0x45, A: 0xff}) || activeTheme.Palette.WarningSoftForeground != (color.NRGBA{R: 0x85, G: 0x5f, B: 0x2e, A: 0xff}) {
		t.Fatalf("light status foregrounds = success %#v warning %#v", activeTheme.Palette.SuccessSoftForeground, activeTheme.Palette.WarningSoftForeground)
	}
	darkTheme := theme.DarkTheme()
	if darkTheme.Palette.SuccessSoftForeground != (color.NRGBA{R: 0x74, G: 0xd8, B: 0x8f, A: 0xff}) || darkTheme.Palette.WarningSoftForeground != (color.NRGBA{R: 0xf9, G: 0xcb, B: 0x86, A: 0xff}) {
		t.Fatalf("dark status foregrounds = success %#v warning %#v", darkTheme.Palette.SuccessSoftForeground, darkTheme.Palette.WarningSoftForeground)
	}
}

func TestToastStateSyncUsesDefaultAndExplicitTimeouts(t *testing.T) {
	start := time.Unix(1, 0)
	var state toastProviderState
	state.sync(toastTestContextAt(start), []ToastItem{
		Toast("default", "Default"),
		Toast("persistent", "Persistent").Timeout(0),
	}, 4*time.Second)

	if got := state.entry("default"); got.remaining != 4*time.Second || got.deadline != start.Add(4*time.Second) || !got.timerRunning {
		t.Fatalf("default timer = %#v", got)
	}
	if got := state.entry("persistent"); got.remaining != 0 || got.timerRunning {
		t.Fatalf("persistent timer = %#v", got)
	}
}

func TestToastStateSameKeyLoadingUpdateRestartsTimerOnce(t *testing.T) {
	start := time.Unix(1, 0)
	var state toastProviderState
	state.sync(toastTestContextAt(start), []ToastItem{
		Toast("upload", "Uploading").Loading(true).Timeout(0),
	}, 4*time.Second)
	entry := state.entry("upload")
	if entry.remaining != 0 || entry.timerRunning {
		t.Fatalf("loading timer = remaining %v running %v", entry.remaining, entry.timerRunning)
	}

	updatedAt := start.Add(2 * time.Second)
	done := Toast("upload", "Uploaded").Variant(ToastSuccess)
	state.sync(toastTestContextAt(updatedAt), []ToastItem{done}, 4*time.Second)
	wantDeadline := updatedAt.Add(4 * time.Second)
	if entry.item.title != "Uploaded" || entry.configuredTimeout != 4*time.Second || entry.remaining != 4*time.Second || entry.deadline != wantDeadline || !entry.timerRunning {
		t.Fatalf("updated timer = %#v", entry)
	}

	state.sync(toastTestContextAt(updatedAt.Add(time.Second)), []ToastItem{done}, 4*time.Second)
	if entry.deadline != wantDeadline {
		t.Fatalf("unchanged toast reset deadline = %v, want %v", entry.deadline, wantDeadline)
	}
}

func TestToastStateRejectsDuplicateAndEmptyKeys(t *testing.T) {
	for _, items := range [][]ToastItem{
		{Toast("", "Empty")},
		{Toast("same", "One"), Toast("same", "Two")},
	} {
		func() {
			defer func() {
				if recover() == nil {
					t.Fatal("invalid toast keys did not panic")
				}
			}()
			var state toastProviderState
			state.sync(toastTestContextAt(time.Unix(1, 0)), items, 4*time.Second)
		}()
	}
}

func TestToastProviderCleansEntryRemovedBeforeAnimation(t *testing.T) {
	ctx := frame.New(nil, nil, locale.LanguageEnglish)
	router := new(input.Router)
	start := time.Unix(1, 0)
	viewport := image.Pt(900, 500)
	layoutToastFrame(ctx, router, ToastProvider("toasts", []ToastItem{
		Toast("saved", "Saved").Timeout(0),
	}), start, viewport)
	layoutToastFrame(ctx, router, ToastProvider("toasts", nil), start, viewport)

	stateValue := frame.UseState[toastProviderState](ctx, "toasts", stateSlotToast)
	if len(stateValue.entries) != 0 || len(stateValue.order) != 0 {
		t.Fatalf("removed invisible toast was retained: entries=%d order=%v", len(stateValue.entries), stateValue.order)
	}
}

func TestToastTimerPausesAndResumes(t *testing.T) {
	start := time.Unix(1, 0)
	var state toastProviderState
	state.sync(toastTestContextAt(start), []ToastItem{Toast("saved", "Saved")}, 4*time.Second)
	closed := false

	state.updateTimers(toastTestContextAt(start.Add(time.Second)), true, func(*toastEntryState) { closed = true })
	entry := state.entry("saved")
	if entry.remaining != 3*time.Second || entry.timerRunning || closed {
		t.Fatalf("paused timer = remaining %v running %v closed %v", entry.remaining, entry.timerRunning, closed)
	}
	state.updateTimers(toastTestContextAt(start.Add(2*time.Second)), false, func(*toastEntryState) { closed = true })
	if entry.deadline != start.Add(5*time.Second) || !entry.timerRunning || closed {
		t.Fatalf("resumed timer = deadline %v running %v closed %v", entry.deadline, entry.timerRunning, closed)
	}
	state.updateTimers(toastTestContextAt(start.Add(5*time.Second)), false, func(*toastEntryState) { closed = true })
	if !closed {
		t.Fatal("timer did not close at resumed deadline")
	}
}

func TestToastTimerPausesWhileActionIsFocused(t *testing.T) {
	ctx := frame.New(nil, nil, locale.LanguageEnglish)
	router := new(input.Router)
	start := time.Unix(1, 0)
	provider := ToastProvider("toasts", []ToastItem{
		Toast("saved", "Saved").Action("Undo"),
	})
	viewport := image.Pt(900, 500)
	layoutToastFrame(ctx, router, provider, start, viewport)
	layoutToastFrame(ctx, router, provider, start.Add(350*time.Millisecond), viewport)

	stateValue := frame.UseState[toastProviderState](ctx, "toasts", stateSlotToast)
	entry := stateValue.entry("saved")
	if entry == nil {
		t.Fatal("toast state is missing")
	}
	router.Source().Execute(key.FocusCmd{Tag: &entry.action})
	layoutToastFrame(ctx, router, provider, start.Add(time.Second), viewport)
	if entry.remaining != 3*time.Second || entry.timerRunning {
		t.Fatalf("focused action timer = remaining %v running %v", entry.remaining, entry.timerRunning)
	}
}

func TestToastEntryProgressEntersAndExits(t *testing.T) {
	start := time.Unix(1, 0)
	entry := toastEntryState{present: true}
	gtx := toastTestContextAt(start)
	if got := entry.progress(gtx, 350*time.Millisecond); got != 0 {
		t.Fatalf("initial progress = %v, want 0", got)
	}
	gtx.Now = start.Add(350 * time.Millisecond)
	if got := entry.progress(gtx, 350*time.Millisecond); got != 1 {
		t.Fatalf("entered progress = %v, want 1", got)
	}
	entry.present = false
	entry.progress(gtx, 350*time.Millisecond)
	gtx.Now = start.Add(700 * time.Millisecond)
	if got := entry.progress(gtx, 350*time.Millisecond); got != 0 {
		t.Fatalf("exited progress = %v, want 0", got)
	}
}

func TestToastStackPositionAnimatesBetweenIndexes(t *testing.T) {
	start := time.Unix(1, 0)
	entry := toastEntryState{}
	gtx := toastTestContextAt(start)
	if got := entry.stackPosition(gtx, 1, 350*time.Millisecond); got != 1 {
		t.Fatalf("initial stack position = %v, want 1", got)
	}
	entry.stackPosition(gtx, 0, 350*time.Millisecond)
	gtx.Now = start.Add(175 * time.Millisecond)
	if got := entry.stackPosition(gtx, 0, 350*time.Millisecond); got <= 0 || got >= 1 {
		t.Fatalf("mid-animation stack position = %v, want between 0 and 1", got)
	}
	gtx.Now = start.Add(350 * time.Millisecond)
	if got := entry.stackPosition(gtx, 0, 350*time.Millisecond); got != 0 {
		t.Fatalf("final stack position = %v, want 0", got)
	}
}

func TestToastRegionXMatchesPlacements(t *testing.T) {
	tests := []struct {
		placement ToastPlacement
		want      int
	}{
		{ToastBottomStart, 16},
		{ToastTopStart, 16},
		{ToastBottom, 70},
		{ToastTop, 70},
		{ToastBottomEnd, 124},
		{ToastTopEnd, 124},
	}
	for _, test := range tests {
		if got := toastRegionX(600, 460, 16, test.placement); got != test.want {
			t.Fatalf("placement %d x = %d, want %d", test.placement, got, test.want)
		}
	}
}

func TestToastStylesUseVariantColors(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	if got := toastStyleFor(&activeTheme, ToastAccent).title; got != activeTheme.Palette.AccentSoftForeground {
		t.Fatalf("accent title = %#v", got)
	}
	if got := toastStyleFor(&activeTheme, ToastDanger); got.title != activeTheme.Palette.DangerSoftForeground || got.indicator != got.title {
		t.Fatalf("danger style = %#v", got)
	}
	if got := toastStyleFor(&activeTheme, ToastSuccess); got.title != activeTheme.Palette.SuccessSoftForeground || got.indicator != got.title {
		t.Fatalf("success style = %#v", got)
	}
	if got := toastStyleFor(&activeTheme, ToastWarning); got.title != activeTheme.Palette.WarningSoftForeground || got.indicator != got.title {
		t.Fatalf("warning style = %#v", got)
	}
	if got := toastStyleFor(&activeTheme, ToastDefault); got.surface != activeTheme.Palette.Surface || got.description != activeTheme.Palette.MutedForeground {
		t.Fatalf("default style = %#v", got)
	}
	activeTheme.Palette.SuccessSoftForeground = color.NRGBA{}
	activeTheme.Palette.WarningSoftForeground = color.NRGBA{}
	if got := toastStyleFor(&activeTheme, ToastSuccess).title; got != activeTheme.Palette.Success {
		t.Fatalf("success fallback = %#v, want %#v", got, activeTheme.Palette.Success)
	}
	if got := toastStyleFor(&activeTheme, ToastWarning).title; got != activeTheme.Palette.Warning {
		t.Fatalf("warning fallback = %#v, want %#v", got, activeTheme.Palette.Warning)
	}
}

func TestToastProviderLayoutsAtHeroUIWidthAndPlacement(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	ctx := frame.New(nil, &activeTheme, locale.LanguageEnglish)
	router := new(input.Router)
	start := time.Unix(1, 0)
	provider := ToastProvider("toasts", []ToastItem{Toast("saved", "Saved").Timeout(0)})
	layoutToastFrame(ctx, router, provider, start, image.Pt(600, 400))
	layoutToastFrame(ctx, router, provider, start.Add(activeTheme.Components.Toast.AnimationDuration), image.Pt(600, 400))

	node, ok := semanticNodeWithLabel(router.AppendSemantics(nil), "Saved")
	if !ok {
		t.Fatal("toast semantic node is missing")
	}
	if node.Desc.Bounds.Dx() != 460 {
		t.Fatalf("toast width = %d, want 460", node.Desc.Bounds.Dx())
	}
	if node.Desc.Bounds.Min.X != 70 {
		t.Fatalf("toast x = %d, want 70", node.Desc.Bounds.Min.X)
	}
}

func TestToastProviderUsesThemeWidth(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	activeTheme.Components.Toast.Width = 320
	ctx := frame.New(nil, &activeTheme, locale.LanguageEnglish)
	router := new(input.Router)
	start := time.Unix(1, 0)
	provider := ToastProvider("toasts", []ToastItem{Toast("saved", "Saved").Timeout(0)})
	layoutToastFrame(ctx, router, provider, start, image.Pt(600, 400))
	layoutToastFrame(ctx, router, provider, start.Add(activeTheme.Components.Toast.AnimationDuration), image.Pt(600, 400))
	node, ok := semanticNodeWithLabel(router.AppendSemantics(nil), "Saved")
	if !ok || node.Desc.Bounds.Dx() != 320 {
		t.Fatalf("theme-controlled toast bounds = %v, found=%v; want width 320", node.Desc.Bounds, ok)
	}
}

func TestToastProviderUsesPlacementOffset(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	ctx := frame.New(nil, &activeTheme, locale.LanguageEnglish)
	router := new(input.Router)
	start := time.Unix(1, 0)
	provider := ToastProvider("toasts", []ToastItem{Toast("saved", "Saved").Timeout(0)}).
		Placement(ToastBottomEnd).
		Offset(44)
	viewport := image.Pt(600, 400)
	layoutToastFrame(ctx, router, provider, start, viewport)
	layoutToastFrame(ctx, router, provider, start.Add(activeTheme.Components.Toast.AnimationDuration), viewport)
	node, ok := semanticNodeWithLabel(router.AppendSemantics(nil), "Saved")
	if !ok || node.Desc.Bounds.Max.Y != viewport.Y-44 {
		t.Fatalf("offset toast bounds = %v, found=%v", node.Desc.Bounds, ok)
	}
}

func TestToastStackExpandsOnHover(t *testing.T) {
	for _, test := range []struct {
		name      string
		placement ToastPlacement
		expandsUp bool
	}{
		{name: "bottom", placement: ToastBottomEnd, expandsUp: true},
		{name: "top", placement: ToastTopEnd},
	} {
		t.Run(test.name, func(t *testing.T) {
			activeTheme := theme.DefaultTheme()
			ctx := frame.New(nil, &activeTheme, locale.LanguageEnglish)
			router := new(input.Router)
			frontIndicator := new(toastProbe)
			backIndicator := new(toastProbe)
			provider := ToastProvider("toasts", []ToastItem{
				Toast("front", "Front").Indicator(frontIndicator).Timeout(0),
				Toast("back", "Back").Description("A taller toast behind the front notification.").Indicator(backIndicator).Timeout(0),
			}).Placement(test.placement)
			viewport := image.Pt(900, 600)
			start := time.Unix(1, 0)
			duration := activeTheme.Components.Toast.AnimationDuration
			layoutToastFrame(ctx, router, provider, start, viewport)
			layoutToastFrame(ctx, router, provider, start.Add(duration), viewport)

			front, ok := semanticNodeWithLabel(router.AppendSemantics(nil), "Front")
			if !ok {
				t.Fatal("front toast semantic node is missing")
			}
			center := f32.Pt(
				float32(front.Desc.Bounds.Min.X+front.Desc.Bounds.Max.X)/2,
				float32(front.Desc.Bounds.Min.Y+front.Desc.Bounds.Max.Y)/2,
			)
			router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, PointerID: 1, Position: center})
			layoutToastFrame(ctx, router, provider, start.Add(duration+time.Millisecond), viewport)
			layoutToastFrame(ctx, router, provider, start.Add(2*duration+time.Millisecond), viewport)

			stateValue := frame.UseState[toastProviderState](ctx, "toasts", stateSlotToast)
			if !stateValue.regionHovered || stateValue.expansionValue != 1 {
				t.Fatalf("toast region expansion = hovered %v progress %v", stateValue.regionHovered, stateValue.expansionValue)
			}
			if !frontIndicator.enabled || !backIndicator.enabled {
				t.Fatalf("expanded indicator enabled states = front %v back %v", frontIndicator.enabled, backIndicator.enabled)
			}
			front, frontOK := semanticNodeWithLabel(router.AppendSemantics(nil), "Front")
			back, backOK := semanticNodeWithLabel(router.AppendSemantics(nil), "Back")
			separated := front.Desc.Bounds.Max.Y <= back.Desc.Bounds.Min.Y
			if test.expandsUp {
				separated = back.Desc.Bounds.Max.Y <= front.Desc.Bounds.Min.Y
			}
			if !frontOK || !backOK || !separated {
				t.Fatalf("expanded toast bounds = front %v back %v", front.Desc.Bounds, back.Desc.Bounds)
			}

			router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, PointerID: 1, Position: f32.Pt(1, float32(viewport.Y)/2)})
			layoutToastFrame(ctx, router, provider, start.Add(2*duration+2*time.Millisecond), viewport)
			layoutToastFrame(ctx, router, provider, start.Add(3*duration+2*time.Millisecond), viewport)
			if stateValue.regionHovered || stateValue.expansionValue != 0 || backIndicator.enabled {
				t.Fatalf("toast region remained expanded = hovered %v progress %v back enabled %v", stateValue.regionHovered, stateValue.expansionValue, backIndicator.enabled)
			}
		})
	}
}

func TestToastTouchModeShowsCloseButtonOnWideViewport(t *testing.T) {
	ctx := frame.New(nil, nil, locale.LanguageEnglish)
	router := new(input.Router)
	provider := ToastProvider("toasts", []ToastItem{Toast("saved", "Saved").Timeout(0)})
	viewport := image.Pt(900, 500)
	start := time.Unix(1, 0)
	layoutToastFrame(ctx, router, provider, start, viewport)
	layoutToastFrame(ctx, router, provider, start.Add(350*time.Millisecond), viewport)
	if _, ok := semanticButtonWithExactLabel(router.AppendSemantics(nil), "Close"); ok {
		t.Fatal("desktop close button was visible before interaction")
	}
	root, ok := semanticNodeWithLabel(router.AppendSemantics(nil), "Saved")
	if !ok {
		t.Fatal("toast semantic node is missing")
	}
	center := f32.Pt(
		float32(root.Desc.Bounds.Min.X+root.Desc.Bounds.Max.X)/2,
		float32(root.Desc.Bounds.Min.Y+root.Desc.Bounds.Max.Y)/2,
	)
	router.Queue(pointer.Event{Kind: pointer.Press, Source: pointer.Touch, PointerID: 1, Position: center})
	layoutToastFrame(ctx, router, provider, start.Add(360*time.Millisecond), viewport)

	stateValue := frame.UseState[toastProviderState](ctx, "toasts", stateSlotToast)
	if !stateValue.touchMode {
		t.Fatal("touch input did not enable toast touch mode")
	}
	if _, ok := semanticButtonWithExactLabel(router.AppendSemantics(nil), "Close"); !ok {
		t.Fatal("touch mode close button is missing")
	}
}

func TestToastActionReportsItemKey(t *testing.T) {
	ctx := frame.New(nil, nil, locale.LanguageEnglish)
	router := new(input.Router)
	start := time.Unix(1, 0)
	actionKey := ""
	provider := ToastProvider("toasts", []ToastItem{
		Toast("saved", "Saved").Action("Undo").Timeout(0),
	}).OnAction(func(key string) { actionKey = key })
	layoutToastFrame(ctx, router, provider, start, image.Pt(900, 500))
	layoutToastFrame(ctx, router, provider, start.Add(350*time.Millisecond), image.Pt(900, 500))
	node, ok := semanticNodeWithClass(router.AppendSemantics(nil), semantic.Button)
	if !ok {
		t.Fatal("toast action semantic node is missing")
	}
	center := image.Pt(
		(node.Desc.Bounds.Min.X+node.Desc.Bounds.Max.X)/2,
		(node.Desc.Bounds.Min.Y+node.Desc.Bounds.Max.Y)/2,
	)
	router.Queue(
		pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: 1, Buttons: pointer.ButtonPrimary, Position: f32.Pt(float32(center.X), float32(center.Y))},
		pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, PointerID: 1, Position: f32.Pt(float32(center.X), float32(center.Y))},
	)
	layoutToastFrame(ctx, router, provider, start.Add(360*time.Millisecond), image.Pt(900, 500))
	if actionKey != "saved" {
		t.Fatalf("action key = %q, want saved", actionKey)
	}
}

func TestToastCloseButtonAppearsOnHoverAndReportsKey(t *testing.T) {
	ctx := frame.New(nil, nil, locale.LanguageEnglish)
	router := new(input.Router)
	start := time.Unix(1, 0)
	closedKey := ""
	provider := ToastProvider("toasts", []ToastItem{
		Toast("saved", "Saved").Timeout(0),
	}).OnClose(func(key string) { closedKey = key })
	viewport := image.Pt(900, 500)
	layoutToastFrame(ctx, router, provider, start, viewport)
	layoutToastFrame(ctx, router, provider, start.Add(350*time.Millisecond), viewport)
	root, ok := semanticNodeWithLabel(router.AppendSemantics(nil), "Saved")
	if !ok {
		t.Fatal("toast semantic node is missing")
	}
	center := image.Pt((root.Desc.Bounds.Min.X+root.Desc.Bounds.Max.X)/2, (root.Desc.Bounds.Min.Y+root.Desc.Bounds.Max.Y)/2)
	router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, PointerID: 1, Position: f32.Pt(float32(center.X), float32(center.Y))})
	layoutToastFrame(ctx, router, provider, start.Add(360*time.Millisecond), viewport)
	closeNode, ok := semanticButtonWithExactLabel(router.AppendSemantics(nil), "Close")
	if !ok {
		t.Fatal("hovered toast close button is missing")
	}
	closeCenter := image.Pt(
		(closeNode.Desc.Bounds.Min.X+closeNode.Desc.Bounds.Max.X)/2,
		(closeNode.Desc.Bounds.Min.Y+closeNode.Desc.Bounds.Max.Y)/2,
	)
	router.Queue(
		pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: 1, Buttons: pointer.ButtonPrimary, Position: f32.Pt(float32(closeCenter.X), float32(closeCenter.Y))},
		pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, PointerID: 1, Position: f32.Pt(float32(closeCenter.X), float32(closeCenter.Y))},
	)
	layoutToastFrame(ctx, router, provider, start.Add(370*time.Millisecond), viewport)
	if closedKey != "saved" {
		t.Fatalf("closed key = %q, want saved", closedKey)
	}
}

func TestToastTimerStaysPausedOverCloseButtonOutsideRoot(t *testing.T) {
	ctx := frame.New(nil, nil, locale.LanguageEnglish)
	router := new(input.Router)
	start := time.Unix(1, 0)
	provider := ToastProvider("toasts", []ToastItem{Toast("saved", "Saved")})
	viewport := image.Pt(900, 500)
	layoutToastFrame(ctx, router, provider, start, viewport)
	layoutToastFrame(ctx, router, provider, start.Add(350*time.Millisecond), viewport)
	root, ok := semanticNodeWithLabel(router.AppendSemantics(nil), "Saved")
	if !ok {
		t.Fatal("toast semantic node is missing")
	}
	center := image.Pt((root.Desc.Bounds.Min.X+root.Desc.Bounds.Max.X)/2, (root.Desc.Bounds.Min.Y+root.Desc.Bounds.Max.Y)/2)
	router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, PointerID: 1, Position: f32.Pt(float32(center.X), float32(center.Y))})
	layoutToastFrame(ctx, router, provider, start.Add(360*time.Millisecond), viewport)
	closeNode, ok := semanticButtonWithExactLabel(router.AppendSemantics(nil), "Close")
	if !ok {
		t.Fatal("hovered toast close button is missing")
	}
	outside := image.Pt((closeNode.Desc.Bounds.Min.X+closeNode.Desc.Bounds.Max.X)/2, root.Desc.Bounds.Min.Y-1)
	router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, PointerID: 1, Position: f32.Pt(float32(outside.X), float32(outside.Y))})
	layoutToastFrame(ctx, router, provider, start.Add(370*time.Millisecond), viewport)

	stateValue := frame.UseState[toastProviderState](ctx, "toasts", stateSlotToast)
	entry := stateValue.entry("saved")
	if entry == nil || entry.hovered || !entry.close.Hovered() || entry.timerRunning {
		t.Fatalf("close hover pause = entry %#v", entry)
	}
}

func TestToastRootAcceptsKeyboardFocus(t *testing.T) {
	ctx := frame.New(nil, nil, locale.LanguageEnglish)
	router := new(input.Router)
	start := time.Unix(1, 0)
	provider := ToastProvider("toasts", []ToastItem{
		Toast("saved", "Saved").Timeout(0),
	})
	viewport := image.Pt(900, 500)
	layoutToastFrame(ctx, router, provider, start, viewport)
	layoutToastFrame(ctx, router, provider, start.Add(350*time.Millisecond), viewport)

	stateValue := frame.UseState[toastProviderState](ctx, "toasts", stateSlotToast)
	entry := stateValue.entry("saved")
	if entry == nil {
		t.Fatal("toast state is missing")
	}
	router.Source().Execute(key.FocusCmd{Tag: &entry.root})
	if !router.Source().Focused(&entry.root) {
		t.Fatal("toast root did not accept keyboard focus")
	}
}

func TestToastCustomIndicatorUsesHeroUISize(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	ctx := frame.New(nil, &activeTheme, locale.LanguageEnglish)
	probe := new(toastProbe)
	provider := ToastProvider("toasts", nil)
	provider.layoutToastIndicator(ctx, toastTestContextAt(time.Unix(1, 0)), Toast("custom", "Custom").Indicator(probe), toastStyleFor(&activeTheme, ToastDefault))
	if probe.constraints != layout.Exact(image.Pt(16, 16)) {
		t.Fatalf("indicator constraints = %#v, want exact 16x16", probe.constraints)
	}
}

func TestToastBackIndicatorReceivesDisabledContext(t *testing.T) {
	ctx := frame.New(nil, nil, locale.LanguageEnglish)
	router := new(input.Router)
	front := new(toastProbe)
	back := new(toastProbe)
	provider := ToastProvider("toasts", []ToastItem{
		Toast("front", "Front").Indicator(front).Timeout(0),
		Toast("back", "Back").Indicator(back).Timeout(0),
	})
	layoutToastFrame(ctx, router, provider, time.Unix(1, 0), image.Pt(900, 500))
	if !front.enabled || back.enabled {
		t.Fatalf("indicator enabled states = front %v back %v", front.enabled, back.enabled)
	}
}

type toastProbe struct {
	constraints layout.Constraints
	enabled     bool
}

func (p *toastProbe) Layout(_ *frame.Context, gtx layout.Context) layout.Dimensions {
	p.constraints = gtx.Constraints
	p.enabled = gtx.Enabled()
	return layout.Dimensions{Size: gtx.Constraints.Min}
}

func toastTestContextAt(now time.Time) layout.Context {
	var router input.Router
	var ops op.Ops
	return layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(600, 400)},
		Source:      router.Source(),
		Ops:         &ops,
		Now:         now,
	}
}

func layoutToastFrame(ctx *frame.Context, router *input.Router, provider ToastProviderWidget, now time.Time, viewport image.Point) {
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Exact(viewport),
		Source:      router.Source(),
		Ops:         &ops,
		Now:         now,
	}
	frame.BeginFrameWithViewport(ctx, viewport)
	provider.Layout(ctx, gtx)
	frame.LayoutOverlays(ctx, gtx)
	frame.ApplyFrameCommands(ctx, gtx)
	frame.EndFrame(ctx)
	router.Frame(&ops)
}

func semanticNodeWithLabel(nodes []input.SemanticNode, label string) (input.SemanticNode, bool) {
	for _, node := range nodes {
		if node.Desc.Label == label {
			return node, true
		}
		if child, ok := semanticNodeWithLabel(node.Children, label); ok {
			return child, true
		}
	}
	return input.SemanticNode{}, false
}

func semanticNodeWithClass(nodes []input.SemanticNode, class semantic.ClassOp) (input.SemanticNode, bool) {
	for _, node := range nodes {
		if node.Desc.Class == class {
			return node, true
		}
		if child, ok := semanticNodeWithClass(node.Children, class); ok {
			return child, true
		}
	}
	return input.SemanticNode{}, false
}

func semanticButtonWithExactLabel(nodes []input.SemanticNode, label string) (input.SemanticNode, bool) {
	for _, node := range nodes {
		if node.Desc.Class == semantic.Button && node.Desc.Label == label {
			return node, true
		}
		if child, ok := semanticButtonWithExactLabel(node.Children, label); ok {
			return child, true
		}
	}
	return input.SemanticNode{}, false
}
