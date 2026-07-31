package main

import (
	"fmt"
	"image"

	"gioui.org/layout"
	"gioui.org/op"
	lucide "github.com/qianniancn/flowui-icons-lucide"
	"github.com/qianniancn/flowui/ui"
)

func layoutPage(_ *ui.Context, _ Model, _ ui.Send[Msg]) ui.Widget {
	return demoPage("Layout primitives",
		demoSection{Title: "AspectRatio, Box & Center", Content: demoPanel(
			ui.Box(ui.AspectRatio(16.0/9.0,
				ui.Surface(ui.Center(ui.Text("16:9 preview"))).Variant(ui.SurfaceTertiary).Style(ui.Radius(8)),
			)).Style(ui.MaxWidth(520).Overflow(ui.StyleOverflowHidden)),
		)},
		demoSection{Title: "Column, Row, Expanded & Flexible", Content: demoPanel(ui.Column(
			ui.Row(ui.Box(layoutTile("Fixed")).Style(ui.Width(150)), ui.Expanded(layoutTile("Expanded"))).Gap(10),
			ui.Row(ui.Flexible(1, layoutTile("1")), ui.Flexible(2, layoutTile("2"))).Gap(10),
		).Gap(10))},
		demoSection{Title: "Grid & AutoGrid", Content: demoPanel(ui.Column(
			ui.Grid(3, layoutTile("One"), layoutTile("Two"), layoutTile("Three")).Gap(10),
			ui.AutoGrid(150, layoutTile("Auto A"), layoutTile("Auto B"), layoutTile("Auto C"), layoutTile("Auto D")).Gap(10),
		).Gap(10))},
		demoSection{Title: "Wrap", Content: demoPanel(ui.Wrap(
			ui.Chip("Design"), ui.Chip("Engineering"), ui.Chip("Product"), ui.Chip("Research"), ui.Chip("Operations"),
		).Gap(8).LineGap(8))},
		demoSection{Title: "Stack, Divider, Separator & Spacer", Content: demoPanel(ui.Column(
			ui.Key("catalog-layout-stack", ui.Box(ui.Stack(
				ui.Stacked(ui.Surface(ui.Box(ui.Text("Stacked layer")).Style(ui.FillWidth().Height(110).Padding(16))).Variant(ui.SurfaceTertiary).Style(ui.Radius(8))),
				ui.Overlay(ui.Chip("Overlay").Color(ui.ChipAccent)).Align(ui.AlignTopEnd),
			)).Style(ui.MaxWidth(520))),
			ui.Divider(),
			ui.Row(ui.Text("Left"), ui.Box(ui.Separator()).Style(ui.Height(22)), ui.Spacer(14, 1), ui.Text("Right")).AlignMiddle().Gap(10),
		).Gap(12))},
	)
}

func scrollingPage(_ *ui.Context, _ Model, _ ui.Send[Msg]) ui.Widget {
	rows := func(prefix string, count int) ui.Widget {
		items := make([]ui.Widget, count)
		for index := range items {
			items[index] = catalogListRow(prefix, index)
		}
		return ui.Column(items...).Gap(8)
	}
	horizontal := make([]ui.Widget, 10)
	for index := range horizontal {
		horizontal[index] = ui.Box(layoutTile(fmt.Sprintf("Item %02d", index+1))).Style(ui.Width(132))
	}
	return demoPage("Scrolling",
		demoSection{Title: "Scroll", Content: demoPanel(
			ui.Box(ui.Scroll("catalog-scroll", rows("Scroll", 24)).Vertical()).Style(ui.FillWidth().Height(220)),
		)},
		demoSection{Title: "Scrollbar", Content: demoPanel(
			ui.Box(ui.Scrollbar("catalog-scrollbar", rows("Scrollbar", 24)).Overlay()).Style(ui.FillWidth().Height(220)),
		)},
		demoSection{Title: "Horizontal Scrollbar", Content: demoPanel(
			ui.Box(ui.Scrollbar("catalog-horizontal-scrollbar", ui.Row(horizontal...).Gap(10)).Horizontal().Overlay()).Style(ui.FillWidth().Height(82)),
		)},
		demoSection{Title: "List", Content: demoPanel(
			ui.Box(ui.List("catalog-list", 1000, func(index int) ui.Widget {
				return catalogListRow("Virtual item", index)
			}).Gap(8)).Style(ui.FillWidth().Height(240)),
		)},
	)
}

func splitPanePage(_ *ui.Context, _ Model, _ ui.Send[Msg]) ui.Widget {
	workspace := ui.SplitPane(
		"catalog-split-horizontal",
		catalogPane("Explorer", ui.Icon(lucide.Files).Size(18)),
		ui.SplitPane(
			"catalog-split-vertical",
			catalogPane("Editor", ui.Icon(lucide.FileCode).Size(18)),
			catalogPane("Output", ui.Icon(lucide.Terminal).Size(18)),
		).Vertical().DefaultRatio(.68).MinFirst(180).MinSecond(100).Label("Resize editor and output"),
	).DefaultRatio(.26).MinFirst(180).MinSecond(360).Label("Resize explorer and editor")
	return demoPage("Split pane",
		demoSection{Title: "Horizontal & vertical SplitPane", Content: demoPanel(
			ui.Box(workspace).Style(ui.FillWidth().Height(430)),
		)},
	)
}

func appShellPage(_ *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
	action := func(key, label string, icon []byte) ui.Action {
		return ui.NewAction(key, label).
			Icon(ui.Icon(icon).Size(16)).
			OnExecute(func() { send(func(model *Model) { model.LastAction = label }) })
	}
	newAction := action("catalog-action-new", "New", lucide.FilePlus)
	saveAction := action("catalog-action-save", "Save", lucide.Save).Shortcut(ui.KeyShortcut("S", ui.ShortcutPrimary))
	deleteAction := action("catalog-action-delete", "Delete", lucide.Trash2).Danger(true)
	actions := []ui.Action{newAction, saveAction, deleteAction}

	command := func(key, label string, icon []byte) ui.Command {
		return ui.NewCommand(key, label).
			Icon(ui.Icon(icon).Size(16)).
			OnExecute(func() { send(func(model *Model) { model.LastAction = label }) })
	}
	runCommand := command("catalog-command-run", "Run", lucide.Play).Shortcut(ui.KeyShortcut("R", ui.ShortcutPrimary))
	refreshCommand := command("catalog-command-refresh", "Refresh", lucide.RefreshCw)
	commands := []ui.Command{runCommand, refreshCommand}

	menu := ui.Menubar("catalog-shell-menu", []ui.MenubarItem{
		ui.MenubarMenu("catalog-shell-file", "File", []ui.MenuItem{
			ui.ActionMenuItem(newAction),
			ui.ActionMenuItem(saveAction),
			ui.MenuSeparator(),
			ui.ActionMenuItem(deleteAction),
		}),
	}).Compact(true).Alt("Application menu")
	portal := ui.Portal("catalog-portal", model.PortalOpen,
		ui.Button("catalog-portal-trigger", ui.Icon(lucide.PanelTopOpen).Size(16)).
			Variant(ui.ButtonSecondary).IconOnly().Label("Open portal").
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
	shell := ui.Surface(ui.Column(
		ui.WindowTitleBar("catalog-title-bar", "FlowUI Components", menu),
		ui.Expanded(ui.Row(
			ui.Surface(ui.Box(ui.Column(
				ui.Text("EXPLORER").Size(11),
				ui.Divider(),
				ui.Text("components").Size(13),
				ui.Text("  button.go").Size(13),
				ui.Text("  surface.go").Size(13),
			).Gap(9)).Style(ui.Width(190).FillHeight().Padding(12))).Variant(ui.SurfaceSecondary),
			ui.Expanded(ui.Box(ui.Column(
				ui.Toolbar(
					ui.ActionButton("catalog-action-new-button", newAction),
					ui.ActionButton("catalog-action-save-button", saveAction),
					ui.ToolbarSeparator(),
					portal,
				).Attached(true).Alt("Application commands"),
				ui.Divider(),
				ui.Text("Component workspace").Size(18),
				ui.Description("Commands, title bar, status bar, and portal share one application shell."),
			).Gap(12)).Style(ui.FillWidth().FillHeight().Padding(16))),
		)),
		ui.StatusBar(
			ui.Row(ui.Icon(lucide.GitBranch).Size(13), ui.Text("main").Size(12)).AlignMiddle().Gap(5),
			ui.Text(status).Size(12),
		),
	)).Style(ui.BorderWidth(1))
	commandDemo := ui.CommandScope(commands,
		ui.Column(
			ui.Toolbar(
				ui.CommandButton("catalog-command-run-button", runCommand),
				ui.CommandButton("catalog-command-refresh-button", refreshCommand),
			).Attached(true).Alt("Command buttons"),
			ui.Box(ui.Menu("catalog-command-menu", []ui.MenuItem{
				ui.CommandMenuItem(runCommand),
				ui.CommandMenuItem(refreshCommand),
			})).Style(ui.Width(260)),
		).Gap(12),
	)

	return demoPage("Application shell",
		demoSection{Title: "WindowTitleBar, StatusBar, Actions & Portal", Content: ui.ActionScope(actions, ui.Box(shell).Style(ui.FillWidth().Height(440).Overflow(ui.StyleOverflowHidden)))},
		demoSection{Title: "CommandScope, CommandButton & CommandMenuItem", Content: demoPanel(commandDemo)},
	)
}

func layoutTile(label string) ui.Widget {
	return ui.Surface(ui.Box(ui.Center(ui.Text(label).Size(13))).Style(ui.FillWidth().Height(54))).Variant(ui.SurfaceTertiary).Style(ui.Radius(7))
}

func catalogListRow(prefix string, index int) ui.Widget {
	return ui.Surface(ui.Box(ui.Row(
		ui.Text(fmt.Sprintf("%s %03d", prefix, index+1)).Size(13),
		ui.Expanded(ui.Spacer(0, 0)),
		ui.Text("Ready").Size(12),
	).AlignMiddle()).Style(ui.FillWidth().Padding(10))).Variant(ui.SurfaceTertiary).Style(ui.Radius(6))
}

func catalogPane(title string, icon ui.Widget) ui.Widget {
	return ui.Surface(ui.Box(ui.Column(
		ui.Row(icon, ui.Text(title).Size(13)).AlignMiddle().Gap(8),
		ui.Divider(),
		ui.Expanded(ui.Center(ui.Text(title+" content").Size(13))),
	).Gap(10)).Style(ui.FillWidth().FillHeight().Padding(12))).Variant(ui.SurfaceSecondary)
}

func catalogPortalPanel(anchor image.Rectangle, interactive bool, close func()) ui.Widget {
	return ui.WidgetFunc(func(ctx *ui.Context, gtx layout.Context) layout.Dimensions {
		if !interactive {
			gtx = gtx.Disabled()
		}
		panelGtx := gtx
		panelGtx.Constraints.Min = image.Point{}
		panelGtx.Constraints.Max = image.Pt(min(gtx.Dp(300), gtx.Constraints.Max.X), min(gtx.Dp(180), gtx.Constraints.Max.Y))
		panel := ui.Key("catalog-portal-panel", ui.Surface(
			ui.Box(ui.Column(
				ui.Text("Portal content").Size(15),
				ui.Description("Rendered at the root overlay host."),
				ui.Button("catalog-portal-close", ui.Text("Close")).Size(ui.ButtonSmall).OnClick(close),
			).Gap(10)).Style(ui.Width(280).Padding(14)),
		).Style(ui.Radius(8).Shadow(ui.ShadowSurface)))

		macro := op.Record(gtx.Ops)
		_, placement := ui.TrackOverlayPlacement(ctx, func() layout.Dimensions {
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
