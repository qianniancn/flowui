package main

import (
	"image/color"

	lucide "github.com/qianniancn/flowui-icons-lucide"
	"github.com/qianniancn/flowui/ui"
)

type Model struct{}
type Msg struct{}

func Update(_ *Model, _ Msg) ui.Cmd[Msg] { return nil }

func View(_ *ui.Context, _ Model, _ ui.Send[Msg]) ui.Widget {
	return ui.Scroll("timeline-page", ui.Box(ui.Column(
		ui.Text("FlowUI TimeLine").Size(24),
		ui.Text("Ant Design-inspired event timelines").Size(14),
		section("Basic", ui.TimeLine([]ui.TimeLineItem{
			{Title: ui.Text("Created"), Content: ui.Text("The project was created"), Color: ui.TimeLineBlue},
			{Title: ui.Text("Reviewed"), Content: ui.Text("The design review is complete"), Color: ui.TimeLineGreen},
			{Title: ui.Text("Published"), Content: ui.Text("The release is now available"), Color: ui.TimeLineGray},
		}).Gap(14)),
		section("Alternate and custom colors", ui.TimeLine([]ui.TimeLineItem{
			{Title: ui.Text("Plan"), Content: ui.Description("Define the scope and milestones."), Color: ui.TimeLineBlue},
			{Title: ui.Text("Build"), Content: ui.Description("Implement the selected work."), Color: ui.TimeLineGreen, Placement: ui.TimeLinePlacementEnd},
			ui.CustomTimeLineItem(ui.Text("Review"), ui.Description("Collect feedback from the team."), color.NRGBA{R: 0x7c, G: 0x3a, B: 0xed, A: 0xff}),
		}).Mode(ui.TimeLineAlternate).TitleSpan(9).Variant(ui.TimeLineFilled)),
		section("Pending", ui.TimeLine([]ui.TimeLineItem{
			{Title: ui.Text("Queued"), Content: ui.Text("Waiting for the next worker"), Color: ui.TimeLineBlue},
			{Title: ui.Text("Running"), Content: ui.Text("The task is in progress"), Color: ui.TimeLineGreen},
		}).Pending(ui.Text("Deploying release...")).PendingIcon(ui.Icon(lucide.LoaderCircle).Size(16))),
		section("Icons and horizontal orientation", ui.TimeLine([]ui.TimeLineItem{
			{Title: ui.Text("Plan"), Content: ui.Text("Scope approved"), Icon: ui.Icon(lucide.ListChecks).Size(16), Color: ui.TimeLineBlue},
			{Title: ui.Text("Build"), Content: ui.Text("Artifacts ready"), Icon: ui.Icon(lucide.PackageCheck).Size(16), Color: ui.TimeLineGreen},
			{Title: ui.Text("Ship"), Content: ui.Text("Release completed"), Icon: ui.Icon(lucide.Rocket).Size(16), Color: ui.TimeLineRed},
		}).Orientation(ui.TimeLineHorizontal).Variant(ui.TimeLineFilled)),
		section("Reverse order", ui.TimeLine([]ui.TimeLineItem{
			{Title: ui.Text("Latest"), Content: ui.Text("Most recent event"), Color: ui.TimeLineBlue},
			{Title: ui.Text("Earlier"), Content: ui.Text("Previous event"), Color: ui.TimeLineGray},
		}).Reverse(true)),
	).Gap(28)).Style(ui.FillWidth().MaxWidth(960).Padding(24)))
}

func section(title string, content ui.Widget) ui.Widget {
	return ui.Column(ui.Text(title).Size(18), content).Gap(10)
}

func main() {
	ui.Run(ui.NewProgram(Model{}, Update, View), ui.Title("FlowUI TimeLine"), ui.Size(980, 860))
}
