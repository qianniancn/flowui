package main

import "github.com/qianniancn/FlowUI/ui"

type Model struct {
	Feedback string
	Notes    string
	Details  string
}

type Msg struct {
	Field string
	Value string
}

func Update(model *Model, msg Msg) {
	switch msg.Field {
	case "feedback":
		model.Feedback = msg.Value
	case "notes":
		model.Notes = msg.Value
	case "details":
		model.Details = msg.Value
	}
}

func View(_ *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
	return ui.Center(
		ui.Box(
			ui.Scroll("textareas",
				ui.Box(ui.Column(
					ui.Text("FlowUI TextArea").Size(24),
					ui.Divider(),
					ui.Wrap(
						ui.Column(
							section("Variants",
								ui.Column(
									labeledTextArea("Primary", "feedback",
										boundTextArea("feedback", model.Feedback, "Describe your product", send)),
									labeledTextArea("Secondary", "notes",
										boundTextArea("notes", model.Notes, "Write meeting notes", send).
											Variant(ui.TextAreaSecondary)),
								).Gap(14),
							),
							section("Rows",
								labeledTextArea("Detailed notes", "details",
									boundTextArea("details", model.Details, "Write out the full meeting notes", send).
										Rows(6).
										MaxLength(800)),
							),
						).Gap(28),
						ui.Column(
							section("States",
								ui.Column(
									labeledTextArea("Read only", "read-only",
										ui.TextArea("read-only", "This value cannot be edited.").ReadOnly(true)),
									labeledTextArea("Invalid", "invalid",
										ui.TextArea("invalid", "The supplied details are incomplete.").Invalid(true)),
									labeledTextArea("Disabled", "disabled",
										ui.TextArea("disabled", "This field is unavailable.").Disabled(true)),
								).Gap(14),
							),
							section("On surface",
								ui.Box(
									ui.Surface(
										ui.Box(
											ui.TextArea("surface", "").
												Placeholder("Secondary textarea on a surface").
												Variant(ui.TextAreaSecondary).
												FullWidth(),
										).Style(ui.Padding(16)),
									).Variant(ui.SurfaceSecondary).Style(ui.Radius(20)),
								).Style(ui.Width(340)),
							),
						).Gap(28),
					).Gap(28).LineGap(24).AlignStart(),
				).Gap(20)).Style(ui.Padding(3)),
			).Vertical(),
		).Style(ui.FillWidth().MaxWidth(800).Padding(24)),
	)
}

func section(title string, child ui.Widget) ui.Widget {
	return ui.Column(ui.Text(title).Size(18), child).Gap(10)
}

func labeledTextArea(label, key string, area ui.TextAreaWidget) ui.Widget {
	return ui.Column(
		ui.Label(label).For(key),
		ui.Box(area).Style(ui.Width(340)),
	).Gap(6)
}

func boundTextArea(key, value, placeholder string, send ui.Send[Msg]) ui.TextAreaWidget {
	return ui.TextArea(key, value).
		Placeholder(placeholder).
		OnChange(func(value string) {
			send(Msg{Field: key, Value: value})
		})
}

func main() {
	ui.Run(
		Model{Feedback: "FlowUI keeps desktop interfaces predictable."},
		Update,
		View,
		ui.Title("FlowUI TextArea"),
		ui.Size(920, 800),
	)
}
