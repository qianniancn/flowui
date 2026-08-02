package main

import (
	_ "embed"
	"fmt"
	"runtime"

	"gioui.org/font"
	"github.com/qianniancn/flowui/ui"
)

//go:embed font/subset/SourceHanSansSC-Regular-Subset.otf
var regularFont []byte

// Medium and Bold remain available for comparison, but the example currently
// embeds only Regular to keep the default font path small.
/*
//go:embed font/subset/SourceHanSansSC-Medium-Subset.otf
var mediumFont []byte

//go:embed font/subset/SourceHanSansSC-Bold-Subset.otf
var boldFont []byte
*/

type Model struct {
	UseCustomFont bool
}

type Msg any

type ToggleFont struct{}

func update(application *ui.Application, systemTheme, customTheme ui.Theme) ui.Update[Model, Msg] {
	return func(model *Model, msg Msg) ui.Cmd[Msg] {
		switch msg.(type) {
		case ToggleFont:
			model.UseCustomFont = !model.UseCustomFont
			activeTheme := systemTheme
			if model.UseCustomFont {
				activeTheme = customTheme
			}
			application.SetTheme("fonts", activeTheme)
		}
		return nil
	}
}

func View(_ *ui.Context, m Model, send ui.Send[Msg]) ui.Widget {
	fontStatus := "system fallback"
	if m.UseCustomFont {
		fontStatus = "bundled Source Han Sans SC"
	}

	return ui.Scroll(
		"font-example",
		ui.Box(
			ui.Column(
				ui.Row(
					ui.Expanded(ui.Column(
						ui.Text("FlowUI Font Example").Size(26).Weight(font.Bold),
						ui.Text(fmt.Sprintf("Current font: %s", fontStatus)).Size(14),
					).Gap(4)),
					ui.Button("toggle", ui.Text("Toggle font source")).
						OnClick(func() {
							send(ToggleFont{})
						}).
						Variant(ui.ButtonPrimary),
				).AlignMiddle(),
				ui.Divider(),
				ui.Text("Weight samples").Size(18).Weight(font.Bold),
				ui.Grid(3,
					weightCard("Regular", font.Normal, "中文文本示例：常规字重。"),
					weightCard("Medium", font.Medium, "中文文本示例：中等字重。"),
					weightCard("Bold", font.Bold, "中文文本示例：加粗字重。"),
				).Gap(12),
				ui.Text("Font size samples").Size(18).Weight(font.Bold),
				ui.Grid(3,
					sizeCard("12 sp", 12),
					sizeCard("14 sp", 14),
					sizeCard("16 sp", 16),
					sizeCard("20 sp", 20),
					sizeCard("24 sp", 24),
					sizeCard("32 sp", 32),
				).Gap(12),
				ui.Text("Latin and symbols: FlowUI 123 !? … → ✓ ©").
					Size(14).
					Align(ui.TextAlignCenter),
				ui.Text("Chinese text is included only to verify CJK glyph coverage.").
					Size(14).
					Align(ui.TextAlignCenter),
			).Gap(16),
		).Style(ui.FillWidth().MaxWidth(1040).Padding(24)),
	).Vertical()
}

func weightCard(label string, weight font.Weight, chinese string) ui.Widget {
	return ui.Card(
		ui.Column(
			ui.Text(label+" weight").Size(17).Weight(weight),
			ui.Text("The quick brown fox jumps over the lazy dog.").Size(15).Weight(weight),
			ui.Text(chinese).Size(15).Weight(weight),
		).Gap(8),
	).Variant(ui.CardSecondary)
}

func sizeCard(label string, size float32) ui.Widget {
	return ui.Card(
		ui.Column(
			ui.Text(label).Size(13),
			ui.Text("FlowUI typography").Size(size),
		).Gap(8),
	).Variant(ui.CardSecondary)
}

func systemTheme() ui.Theme {
	theme := ui.DefaultTheme()
	if runtime.GOOS == "windows" {
		// Keep Windows UI text consistent with the system Chinese interface.
		theme.Typography.Typeface = "Microsoft YaHei, Segoe UI, sans-serif"
	}
	ui.SyncMaterialTheme(&theme)
	return theme
}

func main() {
	// Parse the bundled Regular face once and share it with the custom theme.
	regularFaces, err := ui.ParseFontCollection(regularFont)
	if err != nil {
		panic(fmt.Sprintf("failed to parse regular font: %v", err))
	}

	/*
		mediumFaces, err := ui.ParseFontCollection(mediumFont)
		if err != nil {
			panic(fmt.Sprintf("failed to parse medium font: %v", err))
		}

		boldFaces, err := ui.ParseFontCollection(boldFont)
		if err != nil {
			panic(fmt.Sprintf("failed to parse bold font: %v", err))
		}
	*/

	systemFontTheme := systemTheme()
	customTheme := ui.DefaultTheme()

	// Configure the subset font collection and keep the sample deterministic.
	customTheme.Fonts.Collection = regularFaces
	customTheme.Typography.Typeface = "Source Han Sans SC"
	// The subset includes the sample's Chinese, Latin, and symbol characters;
	// keep platform fallback for everything outside those lists.
	customTheme.Fonts.SystemFonts = true

	// Use Application so the button can replace the active window theme.
	application := ui.NewApplication()
	program := ui.Program[Model, Msg]{
		Init:   func() (Model, ui.Cmd[Msg]) { return Model{UseCustomFont: true}, nil },
		Update: update(application, systemFontTheme, customTheme),
		View:   View,
	}
	application.Run(ui.NewWindow(
		"fonts",
		program,
		ui.Title("FlowUI Font Example"),
		ui.Size(900, 700),
		ui.WithTheme(customTheme),
		ui.CenterOnStart(),
	))
}
