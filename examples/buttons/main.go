package main

import (
	"context"
	"fmt"
	"time"

	"github.com/qianniancn/flowui/ui"
)

type Model struct {
	Clicks  int
	Last    string
	Loading bool
}

type Msg any

type Pressed struct {
	Label string
}

type StartLoading struct{}

type FinishLoading struct{}

func Update(m *Model, msg Msg) ui.Cmd[Msg] {
	switch msg := msg.(type) {
	case Pressed:
		m.Clicks++
		m.Last = msg.Label
	case StartLoading:
		if m.Loading {
			return nil
		}
		m.Clicks++
		m.Last = "Loading"
		m.Loading = true
		return ui.DoContext(func(ctx context.Context, send ui.Send[Msg]) error {
			timer := time.NewTimer(1200 * time.Millisecond)
			defer timer.Stop()
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-timer.C:
				send(FinishLoading{})
				return nil
			}
		})
	case FinishLoading:
		m.Loading = false
		m.Last = "Loaded"
	}
	return nil
}

func View(_ *ui.Context, m Model, send ui.Send[Msg]) ui.Widget {
	status := "No button pressed yet."
	if m.Last != "" {
		status = fmt.Sprintf("%s pressed, total %d", m.Last, m.Clicks)
	}

	return ui.Center(
		ui.Box(
			ui.Scroll("buttons",
				ui.Column(
					ui.Text("FlowUI Buttons").Size(24),
					ui.Text(status).Size(16),
					ui.Divider(),
					section("Variants",
						buttonRow(
							demoButton("Primary", ui.ButtonPrimary, send),
							demoButton("Secondary", ui.ButtonSecondary, send),
							demoButton("Tertiary", ui.ButtonTertiary, send),
							demoButton("Ghost", ui.ButtonGhost, send),
							demoButton("Outline", ui.ButtonOutline, send),
							demoButton("Danger", ui.ButtonDanger, send),
							demoButton("Danger soft", ui.ButtonDangerSoft, send),
						),
					),
					section("Sizes",
						buttonRow(
							ui.Button("small", ui.Text("Small")).
								Size(ui.ButtonSmall).
								OnClick(func() {
									send(Pressed{Label: "Small"})
								}),
							ui.Button("medium", ui.Text("Medium")).
								OnClick(func() {
									send(Pressed{Label: "Medium"})
								}),
							ui.Button("large", ui.Text("Large")).
								Size(ui.ButtonLarge).
								OnClick(func() {
									send(Pressed{Label: "Large"})
								}),
						),
					),
					section("States",
						buttonRow(
							ui.Button("normal", ui.Text("Normal")).
								OnClick(func() {
									send(Pressed{Label: "Normal"})
								}),
							ui.Button("loading", ui.Text("Loading")).
								Loading(m.Loading).
								OnClick(func() {
									send(StartLoading{})
								}),
							ui.Button("disabled", ui.Text("Disabled")).
								Disabled(true),
						),
					),
					section("Instance style",
						buttonRow(
							ui.Button("style-default", ui.Text("Default")).
								Size(ui.ButtonLarge).
								OnClick(func() {
									send(Pressed{Label: "Default style"})
								}),
							customStyleButton(send),
							ui.Button("style-spinner", ui.Text("Spinner")).
								Variant(ui.ButtonSecondary).
								Size(ui.ButtonLarge).
								Loading(m.Loading).
								Style(ui.Background(ui.RGB(0xe8f3ec)).
									TextColor(ui.RGB(0x125e39)).
									When(ui.Hovered, ui.Background(ui.RGB(0xd7e9de))),
								).
								OnClick(func() {
									send(StartLoading{})
								}),
						),
					),
					section("Full width",
						ui.Button("full-width", ui.Text("Full width")).
							FullWidth().
							Variant(ui.ButtonOutline).
							OnClick(func() {
								send(Pressed{Label: "Full width"})
							}),
					),
					section("Icon only",
						buttonRow(
							ui.Button("icon-add", ui.Text("+")).
								IconOnly().
								OnClick(func() {
									send(Pressed{Label: "Add"})
								}),
							ui.Button("icon-loading", ui.Text("...")).
								IconOnly().
								Loading(m.Loading).
								Variant(ui.ButtonSecondary).
								OnClick(func() {
									send(StartLoading{})
								}),
							ui.Button("icon-delete", ui.Text("x")).
								IconOnly().
								Variant(ui.ButtonDangerSoft).
								OnClick(func() {
									send(Pressed{Label: "Delete"})
								}),
						),
					),
				).Gap(18),
			).Vertical(),
		).Style(ui.FillWidth()).Style(ui.MaxWidth(820)).Style(ui.Padding(24)),
	)
}

func section(title string, child ui.Widget) ui.Widget {
	return ui.Column(
		ui.Text(title).Size(18),
		child,
	).Gap(10)
}

func buttonRow(buttons ...ui.Widget) ui.Widget {
	items := make([]ui.Widget, 0, len(buttons))
	for _, button := range buttons {
		items = append(items, ui.Box(button))
	}
	return ui.Wrap(items...).Gap(8).AlignMiddle()
}

func demoButton(label string, variant ui.ButtonVariant, send ui.Send[Msg]) ui.Widget {
	return ui.Button(label, ui.Text(label)).
		Variant(variant).
		OnClick(func() {
			send(Pressed{Label: label})
		})
}

func main() {
	ui.RunCmd(Model{}, Update, View,
		ui.Title("FlowUI Buttons"),
		ui.Size(900, 640),
	)
}
