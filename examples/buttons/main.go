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
	"github.com/qianniancn/flowui/widget/text"
	"github.com/qianniancn/flowui/window"
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

	return layout.Center(
		layout.Box(
			layout.Scroll(layout.Column(
				text.New("FlowUI Buttons").Size(24),
				text.New(status).Size(16),
				layout.Divider(),
				section("Variants",
					buttonRow(
						demoButton("Primary", buttonpkg.Primary, send),
						demoButton("Secondary", buttonpkg.Secondary, send),
						demoButton("Tertiary", buttonpkg.Tertiary, send),
						demoButton("Ghost", buttonpkg.Ghost, send),
						demoButton("Outline", buttonpkg.Outline, send),
						demoButton("Danger", buttonpkg.Danger, send),
						demoButton("Danger soft", buttonpkg.DangerSoft, send),
					),
				),
				section("Sizes",
					buttonRow(
						buttonpkg.New("Small").
							Size(buttonpkg.Small).
							OnClick(func() {
								send(Pressed{Label: "Small"})
							}),
						buttonpkg.New("Medium").
							OnClick(func() {
								send(Pressed{Label: "Medium"})
							}),
						buttonpkg.New("Large").
							Size(buttonpkg.Large).
							OnClick(func() {
								send(Pressed{Label: "Large"})
							}),
					),
				),
				section("States",
					buttonRow(
						buttonpkg.New("Normal").
							OnClick(func() {
								send(Pressed{Label: "Normal"})
							}),
						buttonpkg.New("Loading").
							Loading(m.Loading).
							OnClick(func() {
								send(StartLoading{})
							}),
						buttonpkg.New("Disabled").
							Disabled(true),
					),
				),
				section("Instance style",
					buttonRow(
						buttonpkg.New("Default").
							Size(buttonpkg.Large).
							OnClick(func() {
								send(Pressed{Label: "Default style"})
							}),
						customStyleButton(send),
						(buttonpkg.New("Spinner").
							Variant(buttonpkg.Secondary).
							Size(buttonpkg.Large).
							Loading(m.Loading).
							OnClick(func() {
								send(StartLoading{})
							})).Style(style.New().Background(style.RGB(0xe8f3ec)).
							TextColor(style.RGB(0x125e39)).
							On(style.Hovered, style.New().Background(style.RGB(0xd7e9de)))),
					),
				),
				section("Full width",
					buttonpkg.New("Full width").
						FullWidth().
						Variant(buttonpkg.Outline).
						OnClick(func() {
							send(Pressed{Label: "Full width"})
						}),
				),
				section("Icon only",
					buttonRow(
						buttonpkg.New("+").
							IconOnly().
							OnClick(func() {
								send(Pressed{Label: "Add"})
							}),
						buttonpkg.New("...").
							IconOnly().
							Loading(m.Loading).
							Variant(buttonpkg.Secondary).
							OnClick(func() {
								send(StartLoading{})
							}),
						buttonpkg.New("x").
							IconOnly().
							Variant(buttonpkg.DangerSoft).
							OnClick(func() {
								send(Pressed{Label: "Delete"})
							}),
					),
				),
			).Gap(18),
			).Vertical(),
		).Style(style.New().FillWidth().MaxWidth(820).Padding(24)),
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
	return layout.Wrap(items...).Gap(8).AlignStart()
}

func demoButton(label string, variant buttonpkg.Variant, send ui.Send[Msg]) ui.Widget {
	return buttonpkg.New(label).Key(label).
		Variant(variant).
		OnClick(func() {
			send(Pressed{Label: label})
		})
}

func main() {
	ui.Run(ui.NewProgram(Model{},
		Update, View), window.Title("FlowUI Buttons"),
		window.Size(900, 640),
		window.Theme(base.Light()),
	)
}
