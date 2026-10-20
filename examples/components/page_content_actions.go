package main

import (
	"bytes"
	"image"
	"image/jpeg"

	"gioui.org/font"
	"gioui.org/op/paint"
	lucide "github.com/qianniancn/flowui-icons-lucide"
	"github.com/qianniancn/flowui/assets/images"
	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/avatar"
	"github.com/qianniancn/flowui/widget/badge"
	"github.com/qianniancn/flowui/widget/button"
	"github.com/qianniancn/flowui/widget/card"
	"github.com/qianniancn/flowui/widget/chip"
	"github.com/qianniancn/flowui/widget/closebutton"
	"github.com/qianniancn/flowui/widget/description"
	iconpkg "github.com/qianniancn/flowui/widget/icon"
	imagepkg "github.com/qianniancn/flowui/widget/image"
	"github.com/qianniancn/flowui/widget/input"
	labelpkg "github.com/qianniancn/flowui/widget/label"
	"github.com/qianniancn/flowui/widget/surface"
	"github.com/qianniancn/flowui/widget/text"
	"github.com/qianniancn/flowui/widget/togglebutton"
	"github.com/qianniancn/flowui/widget/toolbar"
	"github.com/qianniancn/flowui/widget/tooltip"
)

var catalogLandscape = paint.NewImageOp(decodeCatalogImage(images.BGDesertJPG))

func typographyPage(_ *ui.Context, _ Model, _ ui.Send[Msg]) ui.Widget {
	return demoPage("Typography & media",
		demoSection{Title: "Text", Content: demoPanel(layout.Column(
			text.New("Display text").Size(26).Weight(font.SemiBold),
			text.New("Body text keeps dense interfaces readable."),
			text.New("Muted supporting text").Size(13),
			text.New("Monospace output: build completed").Typeface("monospace").Size(13),
		).Gap(8))},
		demoSection{Title: "SelectableText", Content: demoPanel(
			text.Selectable("catalog-selectable-text", "package main\n\nfunc main() {\n    println(\"FlowUI\")\n}").
				Typeface("monospace").Size(14).LineHeight(22),
		)},
		demoSection{Title: "Label & Description", Content: demoPanel(layout.Column(
			labelpkg.New("Project name").For("catalog-labeled-input").Required(true),
			input.New("FlowUI").FullWidth(),
			description.New("Used for generated artifacts and window titles.").For("catalog-labeled-input"),
		).Gap(7))},
		demoSection{Title: "Icon", Content: demoPanel(demoRow(
			iconpkg.New(lucide.Heart).Size(20),
			iconpkg.New(lucide.Star).Size(24),
			iconpkg.New(lucide.Bell).Size(28),
			iconpkg.New(lucide.Settings).Size(32),
		))},
		demoSection{Title: "Image", Content: demoPanel(demoRow(
			catalogImagePreview("Cover", imagepkg.New(catalogLandscape).Fit(imagepkg.FitCover).Width(220).Height(132).Radius(8).Alt("Desert landscape cropped to cover")),
			catalogImagePreview("Contain", imagepkg.New(catalogLandscape).Fit(imagepkg.FitContain).Width(220).Height(132).Radius(8).Alt("Desert landscape contained")),
		))},
	)
}

func catalogImagePreview(label string, content ui.Widget) ui.Widget {
	return layout.Column(
		(surface.New(content).Variant(surface.Secondary)).Style(style.New().Radius(8)),
		description.New(label),
	).Gap(6)
}

func surfacesPage(_ *ui.Context, _ Model, _ ui.Send[Msg]) ui.Widget {
	return demoPage("Surfaces & display",
		demoSection{Title: "Surface", Content: demoRow(
			surfaceSample("Default", surface.Default),
			surfaceSample("Secondary", surface.Secondary),
			surfaceSample("Tertiary", surface.Tertiary),
			(surface.New(layout.Box(text.New("Bordered")).Style(style.New().Width(180).Padding(18)))).Style(style.New().Radius(12).BorderWidth(2)),
		)},
		demoSection{Title: "Card", Content: demoRow(
			card.New(text.New("Default card").Size(16), description.New("Primary surface")),
			card.New(text.New("Secondary card").Size(16), description.New("Grouped content")).Variant(card.Secondary),
			card.New(text.New("Transparent card").Size(16), description.New("No surface fill")).Variant(card.Transparent),
		)},
		demoSection{Title: "Avatar", Content: demoPanel(demoRow(
			avatar.New("SM").Size(avatar.SizeSmall),
			avatar.New("MD").Color(avatar.ColorAccent).Variant(avatar.VariantSoft),
			avatar.New("LG").Size(avatar.SizeLarge).Color(avatar.ColorSuccess),
			avatar.New("").Fallback(iconpkg.New(lucide.UserRound).Size(20)),
		))},
		demoSection{Title: "Badge", Content: demoPanel(demoRow(
			badge.New(avatar.New("UI"), "5").Color(badge.ColorDanger).Size(badge.SizeSmall),
			badge.New(avatar.New("ON"), "").Color(badge.ColorSuccess).Placement(badge.PlacementBottomRight).Alt("Online"),
			badge.New(avatar.New("NEW"), "New").Color(badge.ColorAccent).Variant(badge.VariantSoft),
		))},
		demoSection{Title: "Chip", Content: demoPanel(demoRow(
			chip.New("Default"),
			chip.New("Accent").Color(chip.ColorAccent),
			chip.New("Success").Color(chip.ColorSuccess).Variant(chip.VariantSoft),
			chip.New("Warning").Color(chip.ColorWarning).Variant(chip.VariantSecondary),
			chip.New("Danger").Color(chip.ColorDanger).Size(chip.SizeSmall),
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
		demoSection{Title: "Button variants", Content: demoPanel(layout.Column(
			demoRow(
				button.New("Primary").OnClick(press("Primary")),
				button.New("Secondary").Variant(button.Secondary).OnClick(press("Secondary")),
				button.New("Tertiary").Variant(button.Tertiary).OnClick(press("Tertiary")),
				button.New("Ghost").Variant(button.Ghost).OnClick(press("Ghost")),
				button.New("Outline").Variant(button.Outline).OnClick(press("Outline")),
				button.New("Danger").Variant(button.Danger).OnClick(press("Danger")),
			),
			text.New(status).Size(12),
		).Gap(12))},
		demoSection{Title: "Sizes & states", Content: demoPanel(demoRow(
			button.New("Small").Size(button.Small),
			button.New("Medium"),
			button.New("Large").Size(button.Large),
			button.New("Loading").Loading(true),
			button.New("Disabled").Disabled(true),
			button.New("").Content(iconpkg.New(lucide.Plus).Size(18)).IconOnly().Label("Add item"),
		))},
		demoSection{Title: "ButtonGroup", Content: demoPanel(demoRow(
			button.Group(
				button.New("Left"),
				button.New("Center"),
				button.New("Right"),
			).Variant(button.Secondary).Separators(true),
			button.Group(
				button.New("").Content(iconpkg.New(lucide.Bold).Size(16)).IconOnly().Label("Bold"),
				button.New("").Content(iconpkg.New(lucide.Italic).Size(16)).IconOnly().Label("Italic"),
			).Variant(button.Tertiary),
		))},
		demoSection{Title: "ToggleButton & CloseButton", Content: demoPanel(demoRow(
			togglebutton.New("").Key("catalog-toggle").Selected(model.Checked).Content(layout.Row(iconpkg.New(lucide.Heart).Size(16), text.New("Favorite")).Gap(7)).
				OnChange(func(value bool) { send(func(model *Model) { model.Checked = value }) }),
			togglebutton.New("").Key("catalog-toggle-icon").Selected(model.SwitchOn).Content(iconpkg.New(lucide.Bold).Size(17)).IconOnly().Label("Bold").
				OnChange(func(value bool) { send(func(model *Model) { model.SwitchOn = value }) }),
			closebutton.New("catalog-close").Label("Close preview").OnClick(press("Close")),
		))},
	)
}

func toolbarsPage(_ *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
	tool := func(key, label string, icon []byte) ui.Widget {
		return tooltip.New("catalog-"+key+"-tooltip", button.New("").Key("catalog-"+key).Content(iconpkg.New(icon).Size(17)).Variant(button.Secondary).IconOnly().Label(label).
			OnClick(func() { send(func(model *Model) { model.LastAction = label }) }), label).Delay(0)
	}
	return demoPage("Toolbars",
		demoSection{Title: "Toolbar & ToolbarSeparator", Content: demoPanel(layout.Column(
			toolbar.New(
				tool("undo", "Undo", lucide.Undo2),
				tool("redo", "Redo", lucide.Redo2),
				toolbar.Separator(),
				tool("copy", "Copy", lucide.Copy),
				tool("cut", "Cut", lucide.Scissors),
			).Attached(true).Alt("Editing tools"),
			text.New(model.LastAction).Size(12),
		).Gap(10))},
		demoSection{Title: "Vertical toolbar", Content: demoPanel(
			toolbar.New(
				tool("align-left", "Align left", lucide.TextAlignStart),
				tool("align-center", "Align center", lucide.TextAlignCenter),
				tool("align-right", "Align right", lucide.TextAlignEnd),
			).Orientation(toolbar.Vertical).Attached(true).Alt("Alignment tools"),
		)},
	)
}

func surfaceSample(label string, variant surface.Variant) ui.Widget {
	declaration := style.New().Radius(12)
	if variant == surface.Default {
		declaration = declaration.Shadow(style.ShadowSurface)
	}
	return (surface.New(layout.Box(layout.Column(text.New(label).Size(15), description.New("Surface variant")).Gap(4)).Style(style.New().Width(180).Padding(18))).
		Variant(variant)).Style(declaration)
}

func decodeCatalogImage(data []byte) image.Image {
	value, err := jpeg.Decode(bytes.NewReader(data))
	if err != nil {
		panic(err)
	}
	return value
}
