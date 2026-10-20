package main

import (
	"image/color"

	lucide "github.com/qianniancn/flowui-icons-lucide"
	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme/base"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/description"
	"github.com/qianniancn/flowui/widget/icon"
	"github.com/qianniancn/flowui/widget/text"
	"github.com/qianniancn/flowui/widget/timeline"
	"github.com/qianniancn/flowui/window"
)

type Model struct{}
type Msg struct{}

func Update(_ *Model, _ Msg) ui.Cmd[Msg] { return nil }

func View(_ *ui.Context, _ Model, _ ui.Send[Msg]) ui.Widget {
	return layout.Scroll(layout.Box(layout.Column(
		text.New("FlowUI TimeLine").Size(24),
		text.New("Event timelines").Size(14),
		section("Basic", timeline.New([]timeline.Item{
			{Title: text.New("Created"), Content: text.New("The project was created"), Color: timeline.ColorBlue},
			{Title: text.New("Reviewed"), Content: text.New("The design review is complete"), Color: timeline.ColorGreen},
			{Title: text.New("Published"), Content: text.New("The release is now available"), Color: timeline.ColorGray},
		}).Gap(14)),
		section("Alternate and custom colors", timeline.New([]timeline.Item{
			{Title: text.New("Plan"), Content: description.New("Define the scope and milestones."), Color: timeline.ColorBlue},
			{Title: text.New("Build"), Content: description.New("Implement the selected work."), Color: timeline.ColorGreen, Placement: timeline.PlacementEnd},
			timeline.Item{Title: text.New("Review"), Content: description.New("Collect feedback from the team."), CustomColor: color.NRGBA{R: 0x7c, G: 0x3a, B: 0xed, A: 0xff}},
		}).Mode(timeline.ModeAlternate).TitleSpan(9).Variant(timeline.VariantFilled)),
		section("Pending", timeline.New([]timeline.Item{
			{Title: text.New("Queued"), Content: text.New("Waiting for the next worker"), Color: timeline.ColorBlue},
			{Title: text.New("Running"), Content: text.New("The task is in progress"), Color: timeline.ColorGreen},
		}).Pending(text.New("Deploying release...")).PendingIcon(icon.New(lucide.LoaderCircle).Size(16))),
		section("Icons and horizontal orientation", timeline.New([]timeline.Item{
			{Title: text.New("Plan"), Content: text.New("Scope approved"), Icon: icon.New(lucide.ListChecks).Size(16), Color: timeline.ColorBlue},
			{Title: text.New("Build"), Content: text.New("Artifacts ready"), Icon: icon.New(lucide.PackageCheck).Size(16), Color: timeline.ColorGreen},
			{Title: text.New("Ship"), Content: text.New("Release completed"), Icon: icon.New(lucide.Rocket).Size(16), Color: timeline.ColorRed},
		}).Orientation(timeline.OrientationHorizontal).Variant(timeline.VariantFilled)),
		section("Reverse order", timeline.New([]timeline.Item{
			{Title: text.New("Latest"), Content: text.New("Most recent event"), Color: timeline.ColorBlue},
			{Title: text.New("Earlier"), Content: text.New("Previous event"), Color: timeline.ColorGray},
		}).Reverse(true)),
	).Gap(28)).Style(style.New().FillWidth().MaxWidth(960).Padding(24)))
}

func section(title string, content ui.Widget) ui.Widget {
	return layout.Column(text.New(title).Size(18), content).Gap(10)
}

func main() {
	ui.Run(ui.NewProgram(Model{}, Update, View), window.Title("FlowUI TimeLine"), window.Size(980, 860), window.Theme(base.Light()))
}
