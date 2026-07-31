package main

import (
	"bytes"
	"image"
	"image/jpeg"

	"gioui.org/font"
	"gioui.org/op/paint"
	lucide "github.com/qianniancn/flowui-icons-lucide"
	"github.com/qianniancn/flowui/assets/images"
	"github.com/qianniancn/flowui/ui"
)

var catalogLandscape = paint.NewImageOp(decodeCatalogImage(images.BGDesertJPG))

func typographyPage(_ *ui.Context, _ Model, _ ui.Send[Msg]) ui.Widget {
	return demoPage("Typography & media",
		demoSection{Title: "Text", Content: demoPanel(ui.Column(
			ui.Text("Display text").Size(26).Weight(font.SemiBold),
			ui.Text("Body text keeps dense interfaces readable."),
			ui.Text("Muted supporting text").Size(13),
			ui.Text("Monospace output: build completed").Typeface("monospace").Size(13),
		).Gap(8))},
		demoSection{Title: "SelectableText", Content: demoPanel(
			ui.SelectableText("catalog-selectable-text", "package main\n\nfunc main() {\n    println(\"FlowUI\")\n}").
				Typeface("monospace").Size(14).LineHeight(22),
		)},
		demoSection{Title: "Label & Description", Content: demoPanel(ui.Column(
			ui.Label("Project name").For("catalog-labeled-input").Required(true),
			ui.Input("catalog-labeled-input", "FlowUI").FullWidth(),
			ui.Description("Used for generated artifacts and window titles.").For("catalog-labeled-input"),
		).Gap(7))},
		demoSection{Title: "Icon", Content: demoPanel(demoRow(
			ui.Icon(lucide.Heart).Size(20),
			ui.Icon(lucide.Star).Size(24),
			ui.Icon(lucide.Bell).Size(28),
			ui.Icon(lucide.Settings).Size(32),
		))},
		demoSection{Title: "Image", Content: demoPanel(demoRow(
			catalogImagePreview("Cover", ui.Image(catalogLandscape).Fit(ui.ImageCover).Width(220).Height(132).Radius(8).Alt("Desert landscape cropped to cover")),
			catalogImagePreview("Contain", ui.Image(catalogLandscape).Fit(ui.ImageContain).Width(220).Height(132).Radius(8).Alt("Desert landscape contained")),
		))},
	)
}

func catalogImagePreview(label string, content ui.Widget) ui.Widget {
	return ui.Column(
		ui.Surface(content).Variant(ui.SurfaceSecondary).Style(ui.Radius(8)),
		ui.Description(label),
	).Gap(6)
}

func surfacesPage(_ *ui.Context, _ Model, _ ui.Send[Msg]) ui.Widget {
	return demoPage("Surfaces & display",
		demoSection{Title: "Surface", Content: demoRow(
			surfaceSample("Default", ui.SurfaceDefault),
			surfaceSample("Secondary", ui.SurfaceSecondary),
			surfaceSample("Tertiary", ui.SurfaceTertiary),
			ui.Surface(ui.Box(ui.Text("Bordered")).Style(ui.Width(180).Padding(18))).Style(ui.Radius(12).BorderWidth(2)),
		)},
		demoSection{Title: "Card", Content: demoRow(
			ui.Card(ui.Text("Default card").Size(16), ui.Description("Primary surface")),
			ui.Card(ui.Text("Secondary card").Size(16), ui.Description("Grouped content")).Variant(ui.CardSecondary),
			ui.Card(ui.Text("Transparent card").Size(16), ui.Description("No surface fill")).Variant(ui.CardTransparent),
		)},
		demoSection{Title: "Avatar", Content: demoPanel(demoRow(
			ui.Avatar("SM").Size(ui.AvatarSmall),
			ui.Avatar("MD").Color(ui.AvatarAccent).Variant(ui.AvatarSoft),
			ui.Avatar("LG").Size(ui.AvatarLarge).Color(ui.AvatarSuccess),
			ui.Avatar("").Fallback(ui.Icon(lucide.UserRound).Size(20)),
		))},
		demoSection{Title: "Badge", Content: demoPanel(demoRow(
			ui.Badge(ui.Avatar("UI"), "5").Color(ui.BadgeDanger).Size(ui.BadgeSmall),
			ui.Badge(ui.Avatar("ON"), "").Color(ui.BadgeSuccess).Placement(ui.BadgeBottomRight).Alt("Online"),
			ui.Badge(ui.Avatar("NEW"), "New").Color(ui.BadgeAccent).Variant(ui.BadgeSoft),
		))},
		demoSection{Title: "Chip", Content: demoPanel(demoRow(
			ui.Chip("Default"),
			ui.Chip("Accent").Color(ui.ChipAccent),
			ui.Chip("Success").Color(ui.ChipSuccess).Variant(ui.ChipSoft),
			ui.Chip("Warning").Color(ui.ChipWarning).Variant(ui.ChipSecondary),
			ui.Chip("Danger").Color(ui.ChipDanger).Size(ui.ChipSmall),
		))},
	)
}

func buttonsPage(_ *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
	status := model.LastAction
	if status == "" {
		status = "Ready"
	}
	press := func(label string) func() {
		return func() { send(func(model *Model) { model.LastAction = label }) }
	}
	return demoPage("Buttons",
		demoSection{Title: "Button variants", Content: demoPanel(ui.Column(
			demoRow(
				ui.Button("catalog-button-primary", ui.Text("Primary")).OnClick(press("Primary")),
				ui.Button("catalog-button-secondary", ui.Text("Secondary")).Variant(ui.ButtonSecondary).OnClick(press("Secondary")),
				ui.Button("catalog-button-tertiary", ui.Text("Tertiary")).Variant(ui.ButtonTertiary).OnClick(press("Tertiary")),
				ui.Button("catalog-button-ghost", ui.Text("Ghost")).Variant(ui.ButtonGhost).OnClick(press("Ghost")),
				ui.Button("catalog-button-outline", ui.Text("Outline")).Variant(ui.ButtonOutline).OnClick(press("Outline")),
				ui.Button("catalog-button-danger", ui.Text("Danger")).Variant(ui.ButtonDanger).OnClick(press("Danger")),
			),
			ui.Text(status).Size(12),
		).Gap(12))},
		demoSection{Title: "Sizes & states", Content: demoPanel(demoRow(
			ui.Button("catalog-button-small", ui.Text("Small")).Size(ui.ButtonSmall),
			ui.Button("catalog-button-medium", ui.Text("Medium")),
			ui.Button("catalog-button-large", ui.Text("Large")).Size(ui.ButtonLarge),
			ui.Button("catalog-button-loading", ui.Text("Loading")).Loading(true),
			ui.Button("catalog-button-disabled", ui.Text("Disabled")).Disabled(true),
			ui.Button("catalog-button-icon", ui.Icon(lucide.Plus).Size(18)).IconOnly().Label("Add item"),
		))},
		demoSection{Title: "ButtonGroup", Content: demoPanel(demoRow(
			ui.ButtonGroup(
				ui.Button("catalog-group-left", ui.Text("Left")),
				ui.Button("catalog-group-center", ui.Text("Center")),
				ui.Button("catalog-group-right", ui.Text("Right")),
			).Variant(ui.ButtonSecondary).Separators(true),
			ui.ButtonGroup(
				ui.Button("catalog-group-bold", ui.Icon(lucide.Bold).Size(16)).IconOnly().Label("Bold"),
				ui.Button("catalog-group-italic", ui.Icon(lucide.Italic).Size(16)).IconOnly().Label("Italic"),
			).Variant(ui.ButtonTertiary),
		))},
		demoSection{Title: "ToggleButton & CloseButton", Content: demoPanel(demoRow(
			ui.ToggleButton("catalog-toggle", model.Checked, ui.Row(ui.Icon(lucide.Heart).Size(16), ui.Text("Favorite")).Gap(7)).
				OnChange(func(value bool) { send(func(model *Model) { model.Checked = value }) }),
			ui.ToggleButton("catalog-toggle-icon", model.SwitchOn, ui.Icon(lucide.Bold).Size(17)).IconOnly().Label("Bold").
				OnChange(func(value bool) { send(func(model *Model) { model.SwitchOn = value }) }),
			ui.CloseButton("catalog-close").Label("Close preview").OnClick(press("Close")),
		))},
	)
}

func toolbarsPage(_ *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
	tool := func(key, label string, icon []byte) ui.Widget {
		return ui.Tooltip("catalog-"+key+"-tooltip",
			ui.Button("catalog-"+key, ui.Icon(icon).Size(17)).Variant(ui.ButtonSecondary).IconOnly().Label(label).
				OnClick(func() { send(func(model *Model) { model.LastAction = label }) }),
			ui.Text(label),
		).Delay(0)
	}
	return demoPage("Toolbars",
		demoSection{Title: "Toolbar & ToolbarSeparator", Content: demoPanel(ui.Column(
			ui.Toolbar(
				tool("undo", "Undo", lucide.Undo2),
				tool("redo", "Redo", lucide.Redo2),
				ui.ToolbarSeparator(),
				tool("copy", "Copy", lucide.Copy),
				tool("cut", "Cut", lucide.Scissors),
			).Attached(true).Alt("Editing tools"),
			ui.Text(model.LastAction).Size(12),
		).Gap(10))},
		demoSection{Title: "Vertical toolbar", Content: demoPanel(
			ui.Toolbar(
				tool("align-left", "Align left", lucide.TextAlignStart),
				tool("align-center", "Align center", lucide.TextAlignCenter),
				tool("align-right", "Align right", lucide.TextAlignEnd),
			).Orientation(ui.ToolbarVertical).Attached(true).Alt("Alignment tools"),
		)},
	)
}

func surfaceSample(label string, variant ui.SurfaceVariant) ui.Widget {
	declaration := ui.Radius(12)
	if variant == ui.SurfaceDefault {
		declaration = declaration.Shadow(ui.ShadowSurface)
	}
	return ui.Surface(ui.Box(ui.Column(ui.Text(label).Size(15), ui.Description("Surface variant")).Gap(4)).Style(ui.Width(180).Padding(18))).
		Variant(variant).Style(declaration)
}

func decodeCatalogImage(data []byte) image.Image {
	value, err := jpeg.Decode(bytes.NewReader(data))
	if err != nil {
		panic(err)
	}
	return value
}
