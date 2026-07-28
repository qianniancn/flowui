package main

import (
	"time"

	"github.com/qianniancn/flowui/ui"
)

var customButtonStyle = ui.Height(46).
	PaddingX(22).
	Radius(3).
	Background(ui.RGB(0x171717)).
	TextColor(ui.RGB(0xf4ff80)).
	BorderColor(ui.RGB(0xf4ff80)).
	BorderWidth(2).
	BoxShadow(3, 3, 0, 0, ui.RGB(0x00a6a6)).
	Outline(2, 2, ui.RGBA(0x00000000)).
	FontSize(15).
	FontWeight(700).
	Cursor(ui.CursorPointer).
	Opacity(1).
	Scale(1, 1).
	Transition(ui.PropBackgroundColor, 120*time.Millisecond).
	Transition(ui.PropTextColor, 120*time.Millisecond).
	Transition(ui.PropOutlineColor, 120*time.Millisecond).
	Transition(ui.PropTransform, 90*time.Millisecond).
	When(ui.Hovered,
		ui.Background(ui.RGB(0xf4ff80)).
			TextColor(ui.RGB(0x171717)),
	).
	When(ui.Pressed,
		ui.Background(ui.RGB(0xff6b5f)).
			TextColor(ui.RGB(0x171717)).
			Scale(0.96, 0.96),
	).
	When(ui.FocusVisible,
		ui.Outline(2, 2, ui.RGB(0x00a6a6)),
	)

func customStyleButton(send ui.Send[Msg]) ui.Widget {
	return ui.Box(ui.Center(ui.Text("Custom style"))).
		Key("style-custom").
		Label("Custom style").
		Style(customButtonStyle).
		OnClick(func() {
			send(Pressed{Label: "Custom style"})
		})
}
