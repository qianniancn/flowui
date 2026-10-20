package main

import (
	_ "embed"
	"fmt"
	"runtime"

	"gioui.org/font"
	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/theme/base"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/button"
	"github.com/qianniancn/flowui/widget/card"
	"github.com/qianniancn/flowui/widget/text"
	"github.com/qianniancn/flowui/window"
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

func update(application *ui.Application, systemTheme, customTheme theme.Theme) ui.Update[Model, Msg] {
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

	return layout.Scroll(layout.Box(
		layout.Column(
			layout.Row(
				layout.Expanded(layout.Column(
					text.New("FlowUI Font Example").Size(26).Weight(font.Bold),
					text.New(fmt.Sprintf("Current font: %s", fontStatus)).Size(14),
				).Gap(4)),
				button.New("Toggle font source").
					OnClick(func() {
						send(ToggleFont{})
					}).
					Variant(button.Primary),
			).AlignMiddle(),
			layout.Divider(),
			text.New("Weight samples").Size(18).Weight(font.Bold),
			layout.Grid(3,
				weightCard("Regular", font.Normal, "中文文本示例：常规字重。"),
				weightCard("Medium", font.Medium, "中文文本示例：中等字重。"),
				weightCard("Bold", font.Bold, "中文文本示例：加粗字重。"),
			).Gap(12),
			text.New("Font size samples").Size(18).Weight(font.Bold),
			layout.Grid(3,
				sizeCard("12 sp", 12),
				sizeCard("14 sp", 14),
				sizeCard("16 sp", 16),
				sizeCard("20 sp", 20),
				sizeCard("24 sp", 24),
				sizeCard("32 sp", 32),
			).Gap(12),
			text.New("Latin and symbols: FlowUI 123 !? … → ✓ ©").
				Size(14).
				Align(text.AlignCenter),
			text.New("Chinese text is included only to verify CJK glyph coverage.").
				Size(14).
				Align(text.AlignCenter),
		).Gap(16),
	).Style(style.New().FillWidth().MaxWidth(1040).Padding(24)),
	).Vertical()
}

func weightCard(label string, weight font.Weight, chinese string) ui.Widget {
	return card.New(
		layout.Column(
			text.New(label+" weight").Size(17).Weight(weight),
			text.New("The quick brown fox jumps over the lazy dog.").Size(15).Weight(weight),
			text.New(chinese).Size(15).Weight(weight),
		).Gap(8),
	).Variant(card.Secondary)
}

func sizeCard(label string, size float32) ui.Widget {
	return card.New(
		layout.Column(
			text.New(label).Size(13),
			text.New("FlowUI typography").Size(size),
		).Gap(8),
	).Variant(card.Secondary)
}

func systemTheme() theme.Theme {
	result := base.Light()
	if runtime.GOOS == "windows" {
		// Keep Windows UI text consistent with the system Chinese interface.
		result.Typography.Typeface = "Microsoft YaHei, Segoe UI, sans-serif"
	}
	theme.SyncMaterialTheme(&result)
	return result
}

func main() {
	// Parse the bundled Regular face once and share it with the custom theme.
	regularFaces, err := style.ParseFontCollection(regularFont)
	if err != nil {
		panic(fmt.Sprintf("failed to parse regular font: %v", err))
	}

	/*
		mediumFaces, err := style.ParseFontCollection(mediumFont)
		if err != nil {
			panic(fmt.Sprintf("failed to parse medium font: %v", err))
		}

		boldFaces, err := style.ParseFontCollection(boldFont)
		if err != nil {
			panic(fmt.Sprintf("failed to parse bold font: %v", err))
		}
	*/

	systemFontTheme := systemTheme()
	customTheme := base.Light()

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
		window.Title("FlowUI Font Example"),
		window.Size(900, 700),
		window.Theme(customTheme),
		window.CenterOnStart(),
	))
}
