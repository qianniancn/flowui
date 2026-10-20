//go:generate go run ./internal/cataloggen

package main

import (
	"fmt"
	"image/color"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"gioui.org/font"
	giolayout "gioui.org/layout"
	"gioui.org/unit"

	lucide "github.com/qianniancn/flowui-icons-lucide"
	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/theme/base"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/button"
	"github.com/qianniancn/flowui/widget/icon"
	"github.com/qianniancn/flowui/widget/input"
	"github.com/qianniancn/flowui/widget/surface"
	"github.com/qianniancn/flowui/widget/text"
	"github.com/qianniancn/flowui/window"
)

var paper = color.NRGBA{R: 0xf7, G: 0xf4, B: 0xed, A: 0xff}

const (
	tileHeight = 104
	tileGap    = 10
)

type Model struct {
	query    string
	letter   string
	selected string
}

type Msg interface{ msg() }

type SearchChanged struct{ Value string }
type SearchSubmitted struct{ Value string }
type LetterSelected struct{ Letter string }
type IconSelected struct{ Name string }
type ClearFilters struct{}

func (SearchChanged) msg()   {}
func (SearchSubmitted) msg() {}
func (LetterSelected) msg()  {}
func (IconSelected) msg()    {}
func (ClearFilters) msg()    {}

func Update(model *Model, message Msg) ui.Cmd[Msg] {
	switch value := message.(type) {
	case SearchChanged:
		model.query = value.Value
		if strings.TrimSpace(model.query) != "" {
			model.letter = ""
		}
	case SearchSubmitted:
		model.query = value.Value
		if strings.TrimSpace(model.query) != "" {
			model.letter = ""
		}
		if matches := filterIcons(*model); len(matches) > 0 {
			model.selected = matches[0].name
		}
	case LetterSelected:
		model.letter = value.Letter
		model.query = ""
	case IconSelected:
		model.selected = value.Name
	case ClearFilters:
		model.query = ""
		model.letter = ""
	}
	return nil
}

func View(ctx *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
	icons := filterIcons(model)
	return layout.Box(
		layout.Column(
			header(ctx),
			searchRow(ctx, model, send),
			selectedStrip(ctx, model.selected),
			letterIndex(ctx, model, send),
			resultLine(ctx, model, len(icons)),
			layout.Expanded(iconGrid(ctx, icons, model, send)),
		).Gap(14),
	).Style(style.New().
		FillWidth().
		FillHeight().
		Background(style.Color(paper)).
		Padding(28))
}

func header(ctx *ui.Context) ui.Widget {
	active := ctx.Theme()
	sem := active.Semantic()
	accent := sem.Tone(theme.TonePrimary)
	return layout.Column(
		layout.Row(
			layout.Expanded(layout.Row(
				icon.New(lucide.Grid3x3).Size(12).Color(accent.Fill),
				text.New("LUCIDE / ICON INDEX").
					Size(11).
					Weight(font.Medium).
					Color(sem.MutedForeground),
			).Gap(6).AlignMiddle()),
			text.New("v0.4").
				Typeface("monospace").
				Size(11).
				Color(sem.MutedForeground),
		).AlignMiddle(),
		text.New("The whole set.").Size(34).Weight(font.Medium).Color(sem.Foreground),
		text.New(fmt.Sprintf("Search %s icons by name, or browse the alphabet.", formatCount(len(catalog)))).
			Size(13).
			Color(sem.MutedForeground),
	).Gap(5)
}

func searchRow(ctx *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
	active := ctx.Theme()
	sem := active.Semantic()
	search := input.Group(
		input.New("lucide-search").
			Value(model.query).
			Placeholder("Search icons...").
			Size(input.Large).
			OnChange(func(value string) { send(SearchChanged{Value: value}) }).
			OnSubmit(func(value string) { send(SearchSubmitted{Value: value}) }).
			FullWidth(),
	).
		Prefix(icon.New(lucide.Search).Size(18).Color(sem.MutedForeground)).
		FullWidth()

	children := []layout.Widget{layout.Expanded(search)}
	if model.query != "" || model.letter != "" {
		children = append(children, button.New("clear-filters").
			Content(text.New("Clear")).
			Variant(button.Ghost).
			Size(button.Small).
			OnClick(func() { send(ClearFilters{}) }))
	}
	return layout.Row(children...).Gap(8).AlignMiddle()
}

func selectedStrip(ctx *ui.Context, selected string) ui.Widget {
	active := ctx.Theme()
	sem := active.Semantic()
	tone := sem.Tone(theme.TonePrimary)
	entry := findIcon(selected)
	if entry == nil {
		return surface.New(layout.Row(
			icon.New(lucide.MousePointerClick).Size(26).Color(sem.MutedForeground),
			layout.Column(
				text.New("Select an icon").Size(14).Weight(font.Medium),
				text.New("Its Go identifier appears here.").
					Size(11).
					Color(sem.MutedForeground),
			).Gap(2),
		).Gap(12).AlignMiddle()).
			Variant(surface.Secondary).
			Style(style.New().FillWidth().Padding(12).Radius(10))
	}

	return surface.New(layout.Row(
		layout.Box(layout.Center(
			icon.New(entry.data).Size(28).Color(tone.Fill),
		)).Style(style.New().
			Width(46).
			Height(46).
			Radius(9).
			Background(style.Color(tone.Soft))),
		layout.Expanded(layout.Column(
			text.New(entry.name).Size(15).Weight(font.Medium),
			text.New("lucide."+entry.name).
				Typeface("monospace").
				Size(12).
				Color(sem.MutedForeground),
		).Gap(1)),
		text.New("SELECTED").
			Typeface("monospace").
			Size(11).
			Color(sem.MutedForeground),
	).Gap(12).AlignMiddle()).
		Variant(surface.Secondary).
		Style(style.New().FillWidth().Padding(12).Radius(10))
}

func letterIndex(ctx *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
	letters := make([]layout.Widget, 0, 27)
	letters = append(letters, letterChip(ctx, "", "All", model.letter == "", send))
	for r := 'A'; r <= 'Z'; r++ {
		letter := string(r)
		letters = append(letters, letterChip(ctx, letter, letter, model.letter == letter, send))
	}
	return layout.Wrap(letters...).Gap(3).LineGap(3)
}

func letterChip(ctx *ui.Context, letter, label string, selected bool, send ui.Send[Msg]) ui.Widget {
	active := ctx.Theme()
	sem := active.Semantic()
	tone := sem.Tone(theme.TonePrimary)
	key := "letter:all"
	if letter != "" {
		key = "letter:" + letter
	}
	labelColor := sem.MutedForeground
	if selected {
		labelColor = tone.OnSoft
	}
	return layout.Box(
		text.New(label).
			Typeface("monospace").
			Size(14).
			Weight(font.Medium).
			Color(labelColor),
	).
		Key(key).
		Label("Icons starting with " + label).
		OnClick(func() { send(LetterSelected{Letter: letter}) }).
		Align(layout.AlignCenter).
		StyleFunc(func(_ *layout.Context, _ giolayout.Context, state style.State) style.Style {
			background := sem.Surface
			foreground := sem.MutedForeground
			border := sem.Border
			if selected {
				background = tone.Soft
				foreground = tone.OnSoft
				border = tone.Fill
			}
			if state.Hovered {
				if selected {
					background = tone.SoftHover
				} else {
					background = sem.SurfaceHover
				}
			}
			if state.Pressed {
				background = sem.SurfacePressed
			}
			return style.New().
				Height(30).
				PaddingX(8).
				Radius(6).
				Background(style.Color(background)).
				BorderWidth(1).
				BorderColor(style.Color(border)).
				TextColor(style.Color(foreground))
		})
}

func resultLine(ctx *ui.Context, model Model, total int) ui.Widget {
	active := ctx.Theme()
	sem := active.Semantic()
	label := formatCount(total) + " icons"
	if strings.TrimSpace(model.query) != "" {
		label = fmt.Sprintf("%s matches for “%s”", formatCount(total), model.query)
	} else if model.letter != "" {
		label = fmt.Sprintf("%s icons starting with %s", formatCount(total), model.letter)
	}
	children := []layout.Widget{text.New(label).Size(12).Color(sem.MutedForeground)}
	if total > 0 {
		children = append(children,
			layout.Expanded(layout.Spacer(0, 0)),
			text.New("Click a tile to inspect").
				Size(11).
				Color(sem.MutedForeground),
		)
	}
	return layout.Row(children...).AlignMiddle()
}

func iconGrid(ctx *ui.Context, icons []catalogIcon, model Model, send ui.Send[Msg]) ui.Widget {
	return ui.WidgetFunc(func(ctx *ui.Context, gtx giolayout.Context) giolayout.Dimensions {
		if len(icons) == 0 {
			return emptyIcons(ctx, send).Layout(ctx, gtx)
		}
		width := max(gtx.Constraints.Max.X, 1)
		columns := max(2, min(8, (width+tileGap)/(168+tileGap)))
		rows := (len(icons) + columns - 1) / columns
		key := fmt.Sprintf("icon-grid:%d:%s:%s", columns, model.query, model.letter)
		return layout.List(rows, func(row int) ui.Widget {
			children := make([]layout.Widget, 0, columns)
			for column := 0; column < columns; column++ {
				index := row*columns + column
				if index >= len(icons) {
					break
				}
				children = append(children, iconTile(ctx, icons[index], icons[index].name == model.selected, send))
			}
			return layout.Grid(columns, children...).Gap(tileGap)
		}).Key(key).Gap(tileGap).Layout(ctx, gtx)
	})
}

func iconTile(ctx *ui.Context, entry catalogIcon, selected bool, send ui.Send[Msg]) ui.Widget {
	active := ctx.Theme()
	sem := active.Semantic()
	tone := sem.Tone(theme.TonePrimary)
	foreground := sem.Foreground
	nameColor := sem.MutedForeground
	if selected {
		foreground = tone.OnSoft
		nameColor = tone.OnSoft
	}
	return layout.Box(
		layout.Column(
			layout.Center(icon.New(entry.data).Size(26).Color(foreground)),
			text.New(entry.name).
				Size(12).
				Weight(font.Medium).
				MaxLines(1).
				Truncator("…").
				Color(nameColor),
		).Gap(8).AlignMiddle(),
	).
		Key("icon:" + entry.name).
		Label(entry.name).
		OnClick(func() { send(IconSelected{Name: entry.name}) }).
		Align(layout.AlignCenter).
		StyleFunc(func(_ *layout.Context, _ giolayout.Context, state style.State) style.Style {
			background := sem.Surface
			border := sem.Border
			borderWidth := 1
			if selected {
				background = tone.Soft
				border = tone.Fill
				borderWidth = 2
			}
			if state.Hovered {
				if selected {
					background = tone.SoftHover
				} else {
					background = sem.SurfaceHover
				}
			}
			if state.Pressed {
				background = sem.SurfacePressed
			}
			if state.Focused {
				border = tone.Fill
				borderWidth = 2
			}
			return style.New().
				Height(tileHeight).
				Padding(8).
				Radius(8).
				Background(style.Color(background)).
				BorderWidth(unit.Dp(borderWidth)).
				BorderColor(style.Color(border))
		})
}

func emptyIcons(ctx *ui.Context, send ui.Send[Msg]) ui.Widget {
	active := ctx.Theme()
	sem := active.Semantic()
	return layout.Center(layout.Column(
		icon.New(lucide.SearchX).Size(36).Color(sem.MutedForeground),
		text.New("No icons found").Size(17).Weight(font.Medium),
		text.New("Try a shorter name, or clear the filters.").
			Size(13).
			Color(sem.MutedForeground),
		button.New("empty-clear").
			Content(text.New("Clear filters")).
			Variant(button.Secondary).
			OnClick(func() { send(ClearFilters{}) }),
	).Gap(8).AlignMiddle())
}

func filterIcons(model Model) []catalogIcon {
	query := strings.TrimSpace(model.query)
	tokens := searchTokens(query)
	compactQuery := strings.Join(tokens, "")
	letter := strings.ToUpper(strings.TrimSpace(model.letter))

	result := make([]catalogIcon, 0, len(catalog))
	for _, item := range catalog {
		if letter != "" && !strings.HasPrefix(item.name, letter) {
			continue
		}
		if len(tokens) > 0 {
			lowerName := strings.ToLower(item.name)
			matched := true
			for _, token := range tokens {
				if !strings.Contains(lowerName, token) {
					matched = false
					break
				}
			}
			if !matched {
				continue
			}
		}
		result = append(result, item)
	}
	if len(tokens) == 0 {
		return result
	}
	sort.SliceStable(result, func(i, j int) bool {
		leftRank := matchRank(result[i].name, compactQuery)
		rightRank := matchRank(result[j].name, compactQuery)
		if leftRank != rightRank {
			return leftRank < rightRank
		}
		return result[i].name < result[j].name
	})
	return result
}

func matchRank(name, compactQuery string) int {
	compactName := compactIconName(name)
	switch {
	case compactName == compactQuery:
		return 0
	case strings.HasPrefix(compactName, compactQuery):
		return 1
	}
	for _, word := range strings.FieldsFunc(name, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	}) {
		if strings.HasPrefix(strings.ToLower(word), compactQuery) {
			return 2
		}
	}
	return 3
}

func compactIconName(name string) string {
	var builder strings.Builder
	builder.Grow(len(name))
	for _, r := range strings.ToLower(name) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			builder.WriteRune(r)
		}
	}
	return builder.String()
}

func searchTokens(query string) []string {
	return strings.FieldsFunc(strings.ToLower(query), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
}

func findIcon(name string) *catalogIcon {
	for index := range catalog {
		if catalog[index].name == name {
			return &catalog[index]
		}
	}
	return nil
}

func formatCount(value int) string {
	plain := strconv.Itoa(value)
	if len(plain) <= 3 {
		return plain
	}
	var builder strings.Builder
	first := len(plain) % 3
	if first == 0 {
		first = 3
	}
	builder.WriteString(plain[:first])
	for index := first; index < len(plain); index += 3 {
		builder.WriteByte(',')
		builder.WriteString(plain[index : index+3])
	}
	return builder.String()
}

func main() {
	ui.Run(ui.NewProgram(Model{selected: "AArrowDown"}, Update, View),
		window.Title("FlowUI Lucide Icons"),
		window.Size(1080, 780),
		window.Theme(base.Light()),
	)
}
