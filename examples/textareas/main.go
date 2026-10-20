package main

import (
	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme/base"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/input"
	labelpkg "github.com/qianniancn/flowui/widget/label"
	"github.com/qianniancn/flowui/widget/surface"
	"github.com/qianniancn/flowui/widget/text"
	"github.com/qianniancn/flowui/window"
)

type Model struct {
	Feedback string
	Notes    string
	Details  string
}

type Msg struct {
	Field string
	Value string
}

func Update(model *Model, msg Msg) ui.Cmd[Msg] {
	switch msg.Field {
	case "feedback":
		model.Feedback = msg.Value
	case "notes":
		model.Notes = msg.Value
	case "details":
		model.Details = msg.Value
	}
	return nil
}

func View(_ *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
	return layout.Center(
		layout.Box(
			layout.Scroll(layout.Box(layout.Column(
				text.New("FlowUI TextArea").Size(24),
				layout.Divider(),
				layout.Wrap(
					layout.Column(
						section("Variants",
							layout.Column(
								labeledTextArea("Primary", "feedback",
									boundTextArea("feedback", model.Feedback, "Describe your product", send)),
								labeledTextArea("Secondary", "notes",
									boundTextArea("notes", model.Notes, "Write meeting notes", send).
										Variant(input.Secondary)),
							).Gap(14),
						),
						section("Rows",
							labeledTextArea("Detailed notes", "details",
								boundTextArea("details", model.Details, "Write out the full meeting notes", send).
									Rows(6).
									MaxLength(800)),
						),
					).Gap(28),
					layout.Column(
						section("States",
							layout.Column(
								labeledTextArea("Read only", "read-only",
									input.NewTextArea("This value cannot be edited.").ReadOnly(true)),
								labeledTextArea("Invalid", "invalid",
									input.NewTextArea("The supplied details are incomplete.").Invalid(true)),
								labeledTextArea("Disabled", "disabled",
									input.NewTextArea("This field is unavailable.").Disabled(true)),
							).Gap(14),
						),
						section("On surface",
							layout.Box(
								(surface.New(
									layout.Box(
										input.NewTextArea("").
											Placeholder("Secondary textarea on a surface").
											Variant(input.Secondary).
											FullWidth(),
									).Style(style.New().Padding(16)),
								).Variant(surface.Secondary)).Style(style.New().Radius(20)),
							).Style(style.New().Width(340)),
						),
					).Gap(28),
				).Gap(28).LineGap(24).AlignStart(),
			).Gap(20)).Style(style.New().Padding(3)),
			).Vertical(),
		).Style(style.New().FillWidth().MaxWidth(800).Padding(24)),
	)
}

func section(title string, child ui.Widget) ui.Widget {
	return layout.Column(text.New(title).Size(18), child).Gap(10)
}

func labeledTextArea(label, key string, area input.TextAreaWidget) ui.Widget {
	return layout.Column(
		labelpkg.New(label).For(key),
		layout.Box(area).Style(style.New().Width(340)),
	).Gap(6)
}

func boundTextArea(key, value, placeholder string, send ui.Send[Msg]) input.TextAreaWidget {
	return input.NewTextArea(value).Key(key).
		Placeholder(placeholder).
		OnChange(func(value string) {
			send(Msg{Field: key, Value: value})
		})
}

func main() {
	ui.Run(ui.NewProgram(Model{Feedback: "FlowUI keeps desktop interfaces predictable."},
		Update, View), window.Title("FlowUI TextArea"),
		window.Size(920, 800),
		window.Theme(base.Light()),
	)
}
