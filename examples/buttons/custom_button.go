package main

import (
	"time"

	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/text"
)

var customButtonStyle = style.New().Height(46).
	PaddingX(22).
	Radius(3).
	Background(style.RGB(0x171717)).
	TextColor(style.RGB(0xf4ff80)).
	BorderColor(style.RGB(0xf4ff80)).
	BorderWidth(2).
	BoxShadow(3, 3, 0, 0, style.RGB(0x00a6a6)).
	Outline(2, 2, style.RGBA(0x00000000)).
	FontSize(15).
	FontWeight(700).
	Cursor(style.CursorPointer).
	Opacity(1).
	Scale(1).
	Transition(style.PropBackgroundColor, 120*time.Millisecond).
	Transition(style.PropTextColor, 120*time.Millisecond).
	Transition(style.PropOutlineColor, 120*time.Millisecond).
	Transition(style.PropTransform, 90*time.Millisecond).
	On(style.Hovered,
		style.New().Background(style.RGB(0xf4ff80)).
			TextColor(style.RGB(0x171717)),
	).
	On(style.Pressed,
		style.New().Background(style.RGB(0xff6b5f)).
			TextColor(style.RGB(0x171717)).
			Scale(0.96),
	).
	On(style.FocusVisible,
		style.New().Outline(2, 2, style.RGB(0x00a6a6)),
	)

func customStyleButton(send ui.Send[Msg]) ui.Widget {
	return layout.Box(layout.Center(text.New("Custom style"))).
		Key("style-custom").
		Label("Custom style").
		Style(customButtonStyle).
		OnClick(func() {
			send(Pressed{Label: "Custom style"})
		})
}
