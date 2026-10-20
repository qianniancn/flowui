package main

import (
	"context"
	"fmt"
	"time"

	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme/base"
	"github.com/qianniancn/flowui/ui"
	buttonpkg "github.com/qianniancn/flowui/widget/button"
	progresspkg "github.com/qianniancn/flowui/widget/progress"
	"github.com/qianniancn/flowui/widget/text"
	"github.com/qianniancn/flowui/window"
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

	return layout.Center(
		layout.Box(
			layout.Scroll(layout.Column(
				text.New("FlowUI ProgressBar").Size(24),
				text.New(status).Size(16),
				layout.Divider(),
				section("Task progress",
					layout.Column(
						progresspkg.Bar("upload", m.Progress).
							Label("Upload").
							ShowValue().
							Color(taskColor(m.Progress, m.Running)),
						buttonRow(
							buttonpkg.New(startLabel).Key("start").
								Loading(m.Running).
								OnClick(func() {
									send(Start{})
								}),
							buttonpkg.New("Reset").
								Variant(buttonpkg.Secondary).
								Disabled(m.Running && m.Progress < 100).
								OnClick(func() {
									send(Reset{})
								}),
						),
					).Gap(12),
				),
				section("Indeterminate",
					progresspkg.Bar("syncing", 0).
						Label("Syncing").
						Indeterminate().
						Color(progresspkg.BarWarning),
				),
				section("Sizes",
					layout.Column(
						progresspkg.Bar("small", 35).
							Label("Small").
							ShowValue().
							Size(progresspkg.BarSmall),
						progresspkg.Bar("medium", 60).
							Label("Medium").
							ShowValue(),
						progresspkg.Bar("large", 85).
							Label("Large").
							ShowValue().
							Size(progresspkg.BarLarge),
					).Gap(14),
				),
				section("Colors",
					layout.Column(
						progresspkg.Bar("default", 45).
							Label("Default").
							ShowValue().
							Color(progresspkg.BarDefault),
						progresspkg.Bar("accent", 55).
							Label("Accent").
							ShowValue(),
						progresspkg.Bar("success", 70).
							Label("Success").
							ShowValue().
							Color(progresspkg.BarSuccess),
						progresspkg.Bar("warning", 80).
							Label("Warning").
							ShowValue().
							Color(progresspkg.BarWarning),
						progresspkg.Bar("danger", 25).
							Label("Danger").
							ShowValue().
							Color(progresspkg.BarDanger),
					).Gap(14),
				),
			).Gap(18),
			).Vertical(),
		).Style(style.New().FillWidth().MaxWidth(760).Padding(24)),
	)
}

func section(title string, child ui.Widget) ui.Widget {
	return layout.Column(
		text.New(title).Size(18),
		child,
	).Gap(10)
}

func buttonRow(buttons ...ui.Widget) ui.Widget {
	items := make([]ui.Widget, 0, len(buttons))
	for _, button := range buttons {
		items = append(items, layout.Box(button))
	}
	return layout.Wrap(items...).Gap(8).AlignMiddle()
}

func taskColor(progress float64, running bool) progresspkg.BarColor {
	if running {
		return progresspkg.BarAccent
	}
	if progress >= 100 {
		return progresspkg.BarSuccess
	}
	return progresspkg.BarDefault
}

func main() {
	ui.Run(ui.Program[Model, Msg]{
		Init:          func() (Model, ui.Cmd[Msg]) { return Model{Status: "Ready"}, nil },
		Update:        Update,
		Subscriptions: Subscriptions,
		View:          View,
	}, window.Title("FlowUI ProgressBar"),
		window.Size(900, 720),
		window.Theme(base.Light()),
	)
}
