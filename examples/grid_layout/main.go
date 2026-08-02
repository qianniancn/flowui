package main

import "github.com/qianniancn/flowui/ui"

type Model struct{}
type Msg struct{}

func Update(*Model, Msg) ui.Cmd[Msg] { return nil }

func View(_ *ui.Context, _ Model, _ ui.Send[Msg]) ui.Widget {
	return ui.Scroll(
		"grid-layout",
		ui.Box(
			ui.Column(
				ui.Text("FlowUI Grid Layout").Size(24),
				ui.Text("Use Grid for a fixed column count and AutoGrid for responsive content."),
				ui.Divider(),
				section(
					"Fixed columns",
					"Grid keeps three columns at every window width. Each row uses the height of its tallest cell.",
					ui.Grid(3,
						gridCard("Inbox", "18 messages", "Review the latest conversations."),
						gridCard("In progress", "7 tasks", "Work currently being handled."),
						gridCard("Completed", "42 tasks", "Finished work from this week."),
						gridCard("Needs review", "3 items", "These items need a second look before they are closed.", "The extra line makes the row height difference visible."),
						gridCard("Scheduled", "12 items", "Upcoming work is ready to start."),
						gridCard("Archived", "128 items", "Older records remain available for reference."),
					).Gap(16),
				),
				section(
					"Responsive columns",
					"AutoGrid uses the minimum column width to choose how many columns fit. Resize the window to see it reflow.",
					ui.AutoGrid(190,
						gridCard("Design", "12 files", "Shared assets and design notes."),
						gridCard("Engineering", "24 files", "Implementation work and technical decisions."),
						gridCard("Research", "8 files", "References collected for the next release."),
						gridCard("Planning", "5 files", "Milestones, scope, and delivery dates."),
						gridCard("Feedback", "16 files", "Notes from users and internal review."),
						gridCard("Release", "9 files", "Checks and artifacts for the next version."),
					).ColumnGap(16).RowGap(12),
				),
				section(
					"Independent spacing",
					"ColumnGap and RowGap can be set independently when the horizontal and vertical rhythm needs differ.",
					ui.Grid(4,
						spacingCard("A"),
						spacingCard("B"),
						spacingCard("C"),
						spacingCard("D"),
						spacingCard("E"),
						spacingCard("F"),
						spacingCard("G"),
						spacingCard("H"),
					).ColumnGap(24).RowGap(8),
				),
			).Gap(20),
		).Style(ui.FillWidth().MaxWidth(1040).Padding(24)),
	).Vertical()
}

func section(title, description string, content ui.Widget) ui.Widget {
	return ui.Column(
		ui.Text(title).Size(18),
		ui.Text(description).Size(14),
		content,
	).Gap(10)
}

func gridCard(title, value, description string, extra ...string) ui.Widget {
	children := []ui.Widget{
		ui.Text(title).Size(14),
		ui.Text(value).Size(24),
		ui.Text(description).Size(13),
	}
	for _, line := range extra {
		children = append(children, ui.Text(line).Size(13))
	}
	return ui.Card(ui.Column(children...).Gap(6)).Variant(ui.CardDefault)
}

func spacingCard(label string) ui.Widget {
	return ui.Card(ui.Center(ui.Text(label).Size(20))).Variant(ui.CardSecondary)
}

func main() {
	ui.Run(ui.NewProgram(Model{}, Update, View),
		ui.Title("FlowUI Grid Layout"),
		ui.Size(1080, 760),
	)
}
