package main

import (
	"context"
	"fmt"
	"time"

	"github.com/qianniancn/FlowUI/ui"
)

type Model struct {
	Progress float64
	Running  bool
	RunID    int
	Status   string
}

type Msg any

type Start struct{}

type Reset struct{}

type Tick struct {
	RunID int
	Value float64
}

type Done struct {
	RunID int
}

func Update(m *Model, msg Msg) ui.Cmd[Msg] {
	switch msg := msg.(type) {
	case Start:
		if m.Running {
			return nil
		}
		m.RunID++
		m.Progress = 0
		m.Running = true
		m.Status = "Starting upload..."
	case Tick:
		if msg.RunID != m.RunID || !m.Running {
			return nil
		}
		m.Progress = msg.Value
		m.Status = fmt.Sprintf("Uploading %.0f%%", msg.Value)
	case Done:
		if msg.RunID != m.RunID {
			return nil
		}
		m.Progress = 100
		m.Running = false
		m.Status = "Upload complete"
	case Reset:
		m.RunID++
		m.Progress = 0
		m.Running = false
		m.Status = "Ready"
	}
	return nil
}

func Subscriptions(m Model) []ui.Subscription[Msg] {
	if !m.Running {
		return nil
	}
	runID := m.RunID
	return []ui.Subscription[Msg]{
		ui.Subscribe(fmt.Sprintf("upload:%d", runID), func(ctx context.Context, send ui.Send[Msg]) error {
			ticker := time.NewTicker(180 * time.Millisecond)
			defer ticker.Stop()
			for step := 1; step <= 10; step++ {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-ticker.C:
					send(Tick{
						RunID: runID,
						Value: float64(step * 10),
					})
				}
			}
			send(Done{RunID: runID})
			return nil
		}),
	}
}

func View(_ *ui.Context, m Model, send ui.Send[Msg]) ui.Widget {
	status := m.Status
	if status == "" {
		status = "Ready"
	}

	startLabel := "Start"
	if m.Running {
		startLabel = "Uploading"
	}

	return ui.Center(
		ui.Box(
			ui.Scroll("progress-bars",
				ui.Column(
					ui.Text("FlowUI ProgressBar").Size(24),
					ui.Text(status).Size(16),
					ui.Divider(),
					section("Task progress",
						ui.Column(
							ui.ProgressBar("upload", m.Progress).
								Label("Upload").
								ShowValue().
								Color(taskColor(m.Progress, m.Running)),
							buttonRow(
								ui.Button("start", ui.Text(startLabel)).
									Loading(m.Running).
									OnClick(func() {
										send(Start{})
									}),
								ui.Button("reset", ui.Text("Reset")).
									Variant(ui.ButtonSecondary).
									Disabled(m.Running && m.Progress < 100).
									OnClick(func() {
										send(Reset{})
									}),
							),
						).Gap(12),
					),
					section("Indeterminate",
						ui.ProgressBar("syncing", 0).
							Label("Syncing").
							Indeterminate().
							Color(ui.ProgressBarWarning),
					),
					section("Sizes",
						ui.Column(
							ui.ProgressBar("small", 35).
								Label("Small").
								ShowValue().
								Size(ui.ProgressBarSmall),
							ui.ProgressBar("medium", 60).
								Label("Medium").
								ShowValue(),
							ui.ProgressBar("large", 85).
								Label("Large").
								ShowValue().
								Size(ui.ProgressBarLarge),
						).Gap(14),
					),
					section("Colors",
						ui.Column(
							ui.ProgressBar("default", 45).
								Label("Default").
								ShowValue().
								Color(ui.ProgressBarDefault),
							ui.ProgressBar("accent", 55).
								Label("Accent").
								ShowValue(),
							ui.ProgressBar("success", 70).
								Label("Success").
								ShowValue().
								Color(ui.ProgressBarSuccess),
							ui.ProgressBar("warning", 80).
								Label("Warning").
								ShowValue().
								Color(ui.ProgressBarWarning),
							ui.ProgressBar("danger", 25).
								Label("Danger").
								ShowValue().
								Color(ui.ProgressBarDanger),
						).Gap(14),
					),
				).Gap(18),
			).Vertical(),
		).FillWidth().MaxWidth(760).Padding(24),
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

func taskColor(progress float64, running bool) ui.ProgressBarColor {
	if running {
		return ui.ProgressBarAccent
	}
	if progress >= 100 {
		return ui.ProgressBarSuccess
	}
	return ui.ProgressBarDefault
}

func main() {
	ui.RunWithSubscriptions(Model{Status: "Ready"}, Update, Subscriptions, View,
		ui.Title("FlowUI ProgressBar"),
		ui.Size(900, 720),
	)
}
