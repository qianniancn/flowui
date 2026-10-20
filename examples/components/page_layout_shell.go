package main

import (
	"fmt"
	"image"

	"gioui.org/layout"
	"gioui.org/op"
	lucide "github.com/qianniancn/flowui-icons-lucide"
	flowlayout "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/button"
	"github.com/qianniancn/flowui/widget/chip"
	commandpkg "github.com/qianniancn/flowui/widget/command"
	"github.com/qianniancn/flowui/widget/description"
	"github.com/qianniancn/flowui/widget/dock"
	iconpkg "github.com/qianniancn/flowui/widget/icon"
	menupkg "github.com/qianniancn/flowui/widget/menu"
	"github.com/qianniancn/flowui/widget/menubar"
	panelpkg "github.com/qianniancn/flowui/widget/panel"
	portalpkg "github.com/qianniancn/flowui/widget/portal"
	"github.com/qianniancn/flowui/widget/statusbar"
	"github.com/qianniancn/flowui/widget/surface"
	"github.com/qianniancn/flowui/widget/text"
	"github.com/qianniancn/flowui/widget/titlebar"
	"github.com/qianniancn/flowui/widget/toolbar"
)

func layoutPage(_ *ui.Context, _ Model, _ ui.Send[Msg]) ui.Widget {
	return demoPage("Layout primitives",
		demoSection{Title: "AspectRatio, Box & Center", Content: demoPanel(
			flowlayout.Box(flowlayout.AspectRatio(16.0/9.0,
				(surface.New(flowlayout.Center(text.New("16:9 preview"))).Variant(surface.Tertiary)).Style(style.New().Radius(8)),
			)).Style(style.New().MaxWidth(520).Overflow(style.OverflowHidden)),
		)},
		demoSection{Title: "Column, Row, Expanded & Flexible", Content: demoPanel(flowlayout.Column(
			flowlayout.Row(flowlayout.Box(layoutTile("Fixed")).Style(style.New().Width(150)), flowlayout.Expanded(layoutTile("Expanded"))).Gap(10),
			flowlayout.Row(flowlayout.Flexible(1, layoutTile("1")), flowlayout.Flexible(2, layoutTile("2"))).Gap(10),
		).Gap(10))},
		demoSection{Title: "Grid & AutoGrid", Content: demoPanel(flowlayout.Column(
			flowlayout.Grid(3, layoutTile("One"), layoutTile("Two"), layoutTile("Three")).Gap(10),
			flowlayout.AutoGrid(150, layoutTile("Auto A"), layoutTile("Auto B"), layoutTile("Auto C"), layoutTile("Auto D")).Gap(10),
		).Gap(10))},
		demoSection{Title: "Wrap", Content: demoPanel(flowlayout.Wrap(
			chip.New("Design"), chip.New("Engineering"), chip.New("Product"), chip.New("Research"), chip.New("Operations"),
		).Gap(8).LineGap(8))},
		demoSection{Title: "Stack, Divider, Separator & Spacer", Content: demoPanel(flowlayout.Column(
			flowlayout.Key("catalog-layout-stack", flowlayout.Box(flowlayout.Stack(
				flowlayout.Stacked((surface.New(flowlayout.Box(text.New("Stacked layer")).Style(style.New().FillWidth().Height(110).Padding(16))).Variant(surface.Tertiary)).Style(style.New().Radius(8))),
				flowlayout.Overlay(chip.New("Overlay").Color(chip.ColorAccent)).Align(flowlayout.AlignTopEnd),
			)).Style(style.New().MaxWidth(520))),
			flowlayout.Divider(),
			flowlayout.Row(text.New("Left"), flowlayout.Box(flowlayout.Divider()).Style(style.New().Height(22)), flowlayout.Spacer(14, 1), text.New("Right")).AlignMiddle().Gap(10),
		).Gap(12))},
	)
}

func scrollingPage(_ *ui.Context, _ Model, _ ui.Send[Msg]) ui.Widget {
	rows := func(prefix string, count int) ui.Widget {
		items := make([]ui.Widget, count)
		for index := range items {
			items[index] = catalogListRow(prefix, index)
		}
		return flowlayout.Column(items...).Gap(8)
	}
	horizontal := make([]ui.Widget, 10)
	for index := range horizontal {
		horizontal[index] = flowlayout.Box(layoutTile(fmt.Sprintf("Item %02d", index+1))).Style(style.New().Width(132))
	}
	return demoPage("Scrolling",
		demoSection{Title: "Scroll", Content: demoPanel(
			flowlayout.Box(flowlayout.Scroll(rows("Scroll", 24)).Vertical()).Style(style.New().FillWidth().Height(220)),
		)},
		demoSection{Title: "Scrollbar", Content: demoPanel(
			flowlayout.Box(flowlayout.Scrollbar(rows("Scrollbar", 24)).Overlay()).Style(style.New().FillWidth().Height(220)),
		)},
		demoSection{Title: "Horizontal Scrollbar", Content: demoPanel(
			flowlayout.Box(flowlayout.Scrollbar(flowlayout.Row(horizontal...).Gap(10)).Horizontal().Overlay()).Style(style.New().FillWidth().Height(82)),
		)},
		demoSection{Title: "List", Content: demoPanel(
			flowlayout.Box(flowlayout.List(1000, func(index int) ui.Widget {
				return catalogListRow("Virtual item", index)
			}).Gap(8)).Style(style.New().FillWidth().Height(240)),
		)},
	)
}

func splitPanePage(_ *ui.Context, _ Model, _ ui.Send[Msg]) ui.Widget {
	workspace := flowlayout.SplitPane(catalogPane("Explorer", iconpkg.New(lucide.Files).Size(18)),
		flowlayout.SplitPane(catalogPane("Editor", iconpkg.New(lucide.FileCode).Size(18)),
			catalogPane("Output", iconpkg.New(lucide.Terminal).Size(18)),
		).Vertical().DefaultRatio(.68).MinFirst(180).MinSecond(100).Label("Resize editor and output"),
	).DefaultRatio(.26).MinFirst(180).MinSecond(360).Label("Resize explorer and editor")
	dockWorkspace := dock.New("catalog-dock-workspace", dock.Split("catalog-dock-root", dock.Horizontal,
		dock.Panel("catalog-dock-explorer", catalogPane("Explorer", iconpkg.New(lucide.Files).Size(18))),
		dock.Split("catalog-dock-editor-bottom", dock.Vertical,
			dock.Panel("catalog-dock-editor", catalogPane("Editor", iconpkg.New(lucide.FileCode).Size(18))),
			dock.Panel("catalog-dock-output", catalogPane("Output", iconpkg.New(lucide.Terminal).Size(18))),
		).Ratio(.72),
	).Ratio(.24))
	panelHost := panelpkg.New("catalog-panel-host", "editor", []panelpkg.Item{
		{Key: "editor", Content: flowlayout.Center(text.New("Editor view"))},
		{Key: "output", Content: flowlayout.Center(text.New("Output view"))},
	})
	viewStack := panelpkg.New("catalog-view-stack", "overview", []panelpkg.Item{
		{Key: "overview", Content: flowlayout.Center(text.New("Overview view"))},
		{Key: "settings", Content: flowlayout.Center(text.New("Settings view"))},
	})
	return demoPage("Split pane",
		demoSection{Title: "Horizontal & vertical SplitPane", Content: demoPanel(
			flowlayout.Box(workspace).Style(style.New().FillWidth().Height(430)),
		)},
		demoSection{Title: "DockLayout, PanelHost & ViewStack", Content: demoPanel(flowlayout.Column(
			flowlayout.Box(dockWorkspace).Style(style.New().FillWidth().Height(300)),
			flowlayout.Row(
				flowlayout.Box(panelHost).Style(style.New().FillWidth().Height(100)),
				flowlayout.Box(viewStack).Style(style.New().FillWidth().Height(100)),
			).Gap(12),
		).Gap(12))},
	)
}

func appShellPage(_ *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
	command := func(key, label string, icon []byte) commandpkg.Command {
		return commandpkg.New(key, label).
			Icon(iconpkg.New(icon).Size(16)).
			OnExecute(func() { send(func(model *Model) { model.LastAction = label }) })
	}
	newCommand := command("catalog-command-new", "New", lucide.FilePlus)
	saveCommand := command("catalog-command-save", "Save", lucide.Save).Shortcut(commandpkg.KeyShortcut("S", commandpkg.ShortcutPrimary))
	deleteCommand := command("catalog-command-delete", "Delete", lucide.Trash).Danger(true)
	runCommand := command("catalog-command-run", "Run", lucide.Play).Shortcut(commandpkg.KeyShortcut("R", commandpkg.ShortcutPrimary))
	refreshCommand := command("catalog-command-refresh", "Refresh", lucide.RefreshCw)
	commands := []commandpkg.Command{newCommand, saveCommand, deleteCommand, runCommand, refreshCommand}

	menu := menubar.New("catalog-shell-menu", []menubar.Item{
		menubar.NewMenu("catalog-shell-file", "File", []menupkg.Item{
			commandpkg.MenuItem(newCommand),
			commandpkg.MenuItem(saveCommand),
			menupkg.Separator(),
			commandpkg.MenuItem(deleteCommand),
		}),
	}).Compact(true).Alt("Application menu")
	portal := portalpkg.New("catalog-portal", model.PortalOpen,
		button.New("").Content(iconpkg.New(lucide.PanelTopOpen).Size(16)).
			Variant(button.Secondary).IconOnly().Label("Open portal").
			OnClick(func() { send(func(model *Model) { model.PortalOpen = !model.PortalOpen }) }),
		func(anchor image.Rectangle, interactive bool) ui.Widget {
			return catalogPortalPanel(anchor, interactive, func() {
				send(func(model *Model) { model.PortalOpen = false })
			})
		},
	)
	status := model.LastAction
	if status == "" {
		status = "Ready"
	}
	shell := (surface.New(flowlayout.Column(
		titlebar.New("catalog-title-bar", "FlowUI Components", menu),
		flowlayout.Expanded(flowlayout.Row(
			surface.New(flowlayout.Box(flowlayout.Column(
				text.New("EXPLORER").Size(11),
				flowlayout.Divider(),
				text.New("components").Size(13),
				text.New("  button.go").Size(13),
				text.New("  surface.go").Size(13),
			).Gap(9)).Style(style.New().Width(190).FillHeight().Padding(12))).Variant(surface.Secondary),
			flowlayout.Expanded(flowlayout.Box(flowlayout.Column(
				toolbar.New(
					commandpkg.Button("catalog-command-new-button", newCommand),
					commandpkg.Button("catalog-command-save-button", saveCommand),
					toolbar.Separator(),
					portal,
				).Attached(true).Alt("Application commands"),
				flowlayout.Divider(),
				text.New("Component workspace").Size(18),
				description.New("Commands, title bar, status bar, and portal share one application shell."),
			).Gap(12)).Style(style.New().FillWidth().FillHeight().Padding(16))),
		)),
		statusbar.New(
			flowlayout.Row(iconpkg.New(lucide.GitBranch).Size(13), text.New("main").Size(12)).AlignMiddle().Gap(5),
			text.New(status).Size(12),
		),
	))).Style(style.New().BorderWidth(1))
	commandDemo := commandpkg.Scope(commands,
		flowlayout.Column(
			toolbar.New(
				commandpkg.Button("catalog-command-run-button", runCommand),
				commandpkg.Button("catalog-command-refresh-button", refreshCommand),
			).Attached(true).Alt("Command buttons"),
			flowlayout.Box(menupkg.New("catalog-command-menu", []menupkg.Item{
				commandpkg.MenuItem(runCommand),
				commandpkg.MenuItem(refreshCommand),
			})).Style(style.New().Width(260)),
		).Gap(12),
	)
	return demoPage("Application shell",
		demoSection{Title: "WindowTitleBar, StatusBar, Commands & Portal", Content: flowlayout.Box(shell).Style(style.New().FillWidth().Height(440).Overflow(style.OverflowHidden))},
		demoSection{Title: "CommandScope, CommandButton & CommandMenuItem", Content: demoPanel(commandDemo)},
	)
}

func layoutTile(label string) ui.Widget {
	return (surface.New(flowlayout.Box(flowlayout.Center(text.New(label).Size(13))).Style(style.New().FillWidth().Height(54))).Variant(surface.Tertiary)).Style(style.New().Radius(7))
}

func catalogListRow(prefix string, index int) ui.Widget {
	return (surface.New(flowlayout.Box(flowlayout.Row(
		text.New(fmt.Sprintf("%s %03d", prefix, index+1)).Size(13),
		flowlayout.Expanded(flowlayout.Spacer(0, 0)),
		text.New("Ready").Size(12),
	).AlignMiddle()).Style(style.New().FillWidth().Padding(10))).Variant(surface.Tertiary)).Style(style.New().Radius(6))
}

func catalogPane(title string, icon ui.Widget) ui.Widget {
	return surface.New(flowlayout.Box(flowlayout.Column(
		flowlayout.Row(icon, text.New(title).Size(13)).AlignMiddle().Gap(8),
		flowlayout.Divider(),
		flowlayout.Expanded(flowlayout.Center(text.New(title+" content").Size(13))),
	).Gap(10)).Style(style.New().FillWidth().FillHeight().Padding(12))).Variant(surface.Secondary)
}

func catalogPortalPanel(anchor image.Rectangle, interactive bool, close func()) ui.Widget {
	return flowlayout.WidgetFunc(func(ctx *ui.Context, gtx layout.Context) layout.Dimensions {
		if !interactive {
			gtx = gtx.Disabled()
		}
		panelGtx := gtx
		panelGtx.Constraints.Min = image.Point{}
		panelGtx.Constraints.Max = image.Pt(min(gtx.Dp(300), gtx.Constraints.Max.X), min(gtx.Dp(180), gtx.Constraints.Max.Y))
		panel := flowlayout.Key("catalog-portal-panel", (surface.New(
			flowlayout.Box(flowlayout.Column(
				text.New("Portal content").Size(15),
				description.New("Rendered at the root overlay host."),
				button.New("Close").Size(button.Small).OnClick(close),
			).Gap(10)).Style(style.New().Width(280).Padding(14)),
		)).Style(style.New().Radius(8).Shadow(style.ShadowSurface)))

		macro := op.Record(gtx.Ops)
		_, placement := flowlayout.TrackOverlayPlacement(ctx, func() layout.Dimensions {
			return panel.Layout(ctx, panelGtx)
		})
		call := macro.Stop()
		position := image.Pt(anchor.Min.X, anchor.Max.Y+gtx.Dp(8))
		placement.PlaceOffset(position)
		offset := op.Offset(position).Push(gtx.Ops)
		call.Add(gtx.Ops)
		offset.Pop()
		return layout.Dimensions{Size: gtx.Constraints.Max}
	})
}
