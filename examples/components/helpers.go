package main

import (
	"fmt"
	"image/color"

	"github.com/qianniancn/flowui/ui"
)

type demoSection struct {
	Title   string
	Content ui.Widget
}

func componentPage(ctx *ui.Context, page string, model Model, send ui.Send[Msg]) ui.Widget {
	switch page {
	case "typography":
		return typographyPage(ctx, model, send)
	case "surfaces":
		return surfacesPage(ctx, model, send)
	case "motion-style":
		return motionStylePage(ctx, model, send)
	case "buttons":
		return buttonsPage(ctx, model, send)
	case "toolbars":
		return toolbarsPage(ctx, model, send)
	case "text-fields":
		return textFieldsPage(ctx, model, send)
	case "selection":
		return selectionPage(ctx, model, send)
	case "choice-fields":
		return choiceFieldsPage(ctx, model, send)
	case "dates":
		return datesPage(ctx, model, send)
	case "colors":
		return colorsPage(ctx, model, send)
	case "sliders":
		return slidersPage(ctx, model, send)
	case "tabs-pagination":
		return tabsPaginationPage(ctx, model, send)
	case "sidebar-tree":
		return sidebarTreePage(ctx, model, send)
	case "menus":
		return menusPage(ctx, model, send)
	case "status":
		return statusPage(ctx, model, send)
	case "disclosure":
		return disclosurePage(ctx, model, send)
	case "overlays":
		return overlaysPage(ctx, model, send)
	case "tables":
		return tablesPage(ctx, model, send)
	case "charts":
		return chartsPage(ctx, model, send)
	case "layout":
		return layoutPage(ctx, model, send)
	case "scrolling":
		return scrollingPage(ctx, model, send)
	case "split-pane":
		return splitPanePage(ctx, model, send)
	case "app-shell":
		return appShellPage(ctx, model, send)
	default:
		return typographyPage(ctx, model, send)
	}
}

func pageTitle(page string) string {
	return map[string]string{
		"typography":      "Typography & media",
		"surfaces":        "Surfaces & display",
		"motion-style":    "Motion & style",
		"buttons":         "Buttons",
		"toolbars":        "Toolbars",
		"text-fields":     "Text fields",
		"selection":       "Selection controls",
		"choice-fields":   "Choice fields",
		"dates":           "Date controls",
		"colors":          "Color controls",
		"sliders":         "Sliders",
		"tabs-pagination": "Tabs & pagination",
		"sidebar-tree":    "Sidebar & tree",
		"menus":           "Menus",
		"status":          "Status & progress",
		"disclosure":      "Disclosure",
		"overlays":        "Overlays",
		"tables":          "Tables",
		"charts":          "Charts",
		"layout":          "Layout primitives",
		"scrolling":       "Scrolling",
		"split-pane":      "Split pane",
		"app-shell":       "Application shell",
	}[page]
}

func demoPage(title string, sections ...demoSection) ui.Widget {
	children := []ui.Widget{ui.Text(title).Size(26), ui.Divider()}
	for index, section := range sections {
		children = append(children,
			ui.Column(
				ui.Text(section.Title).Size(17),
				section.Content,
			).Gap(12),
		)
		if index < len(sections)-1 {
			children = append(children, ui.Divider())
		}
	}
	return ui.Column(children...).Gap(20)
}

func demoRow(children ...ui.Widget) ui.Widget {
	return ui.Wrap(children...).Gap(10).LineGap(10).AlignMiddle()
}

func demoPanel(child ui.Widget) ui.Widget {
	return ui.Box(child).Style(ui.Padding(18))
}

func swatchLabel(value color.NRGBA) string {
	return fmt.Sprintf("#%02X%02X%02X", value.R, value.G, value.B)
}
