package main

import (
	"fmt"
	"strings"

	lucide "github.com/qianniancn/flowui-icons-lucide"
	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme/base"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/icon"
	"github.com/qianniancn/flowui/widget/input"
	labelpkg "github.com/qianniancn/flowui/widget/label"
	"github.com/qianniancn/flowui/widget/text"
	"github.com/qianniancn/flowui/window"
)

type Model struct {
	Email   string
	Website string
	Price   string
	Search  string
	Token   string
	Prompt  string
	Last    string
}

type Msg struct {
	Field  string
	Value  string
	Copy   bool
	Submit bool
}

func Update(m *Model, msg Msg) ui.Cmd[Msg] {
	if msg.Copy {
		m.Last = fmt.Sprintf("Copied %s", m.Token)
		return nil
	}
	if msg.Submit {
		if strings.TrimSpace(m.Prompt) == "" {
			return nil
		}
		m.Last = fmt.Sprintf("Submitted prompt: %s", m.Prompt)
		m.Prompt = ""
		return nil
	}
	switch msg.Field {
	case "email":
		m.Email = msg.Value
	case "website":
		m.Website = msg.Value
	case "price":
		m.Price = msg.Value
	case "search", "search-primary":
		m.Search = msg.Value
	case "token":
		m.Token = msg.Value
	case "prompt":
		m.Prompt = msg.Value
	}
	return nil
}

func View(_ *ui.Context, m Model, send ui.Send[Msg]) ui.Widget {
	status := "Edit a field to update its value"
	if m.Last != "" {
		status = m.Last
	}

	return layout.Center(
		layout.Box(
			layout.Scroll(layout.Box(layout.Column(
				text.New("FlowUI InputGroup").Size(24),
				text.New(status).Size(14),
				layout.Divider(),
				layout.Wrap(
					layout.Column(
						section("Prefix and suffix",
							layout.Column(
								labeledGroup("Email address", "email",
									input.Group(boundInput("email", m.Email, "name@email.com", send)).
										Prefix(icon.New(lucide.Mail).Size(16)).
										FullWidth()),
								labeledGroup("Website", "website",
									input.Group(boundInput("website", m.Website, "flowui.dev", send)).
										Prefix(icon.New(lucide.Globe).Size(16)).
										Suffix(text.New(".com")).
										FullWidth()),
								labeledGroup("Price", "price",
									input.Group(boundInput("price", m.Price, "10", send).Type(input.Number)).
										Prefix(text.New("$")).
										Suffix(text.New("USD")).
										FullWidth()),
							).Gap(12),
						),
						section("Variants",
							layout.Column(
								labeledGroup("Primary", "search-primary",
									input.Group(boundInput("search-primary", m.Search, "Search projects", send)).
										Prefix(icon.New(lucide.Search).Size(16)).
										FullWidth()),
								labeledGroup("Secondary", "search",
									input.Group(boundInput("search", m.Search, "Search projects", send)).
										Prefix(icon.New(lucide.Search).Size(16)).
										Variant(input.Secondary).
										FullWidth()),
							).Gap(12),
						),
					).Gap(28),
					layout.Column(
						section("Text area",
							labeledGroup("Prompt", "prompt",
								input.GroupTextArea(
									boundTextArea("prompt", m.Prompt, "Assign tasks or ask anything", send).Rows(5),
								).
									Prefix(icon.New(lucide.MessageSquare).Size(16)).
									SuffixAction(
										input.GroupAction("send-prompt", "Send prompt").Content(icon.New(lucide.SendHorizontal).Size(16)).
											Disabled(strings.TrimSpace(m.Prompt) == "").
											OnClick(func() { send(Msg{Submit: true}) }),
									).
									FullWidth(),
							),
						),
						section("Action suffix",
							labeledGroup("API token", "token",
								input.Group(boundInput("token", m.Token, "flow_live_...", send)).
									SuffixAction(
										input.GroupAction("copy-token", "Copy token").Content(icon.New(lucide.Copy).Size(16)).
											OnClick(func() { send(Msg{Copy: true}) }),
									).
									FullWidth(),
							),
						),
						section("States",
							layout.Column(
								labeledGroup("Invalid", "invalid",
									input.Group(input.New("invalid-address").Label("Invalid email")).
										Prefix(icon.New(lucide.Mail).Size(16)).
										Suffix(icon.New(lucide.CircleAlert).Size(16)).
										Invalid(true).
										FullWidth()),
								labeledGroup("Disabled", "disabled",
									input.Group(input.New("name@email.com")).
										Prefix(icon.New(lucide.Mail).Size(16)).
										Disabled(true).
										FullWidth()),
							).Gap(12),
						),
					).Gap(28),
				).Gap(28).LineGap(22).AlignStart(),
			).Gap(22)).Style(style.New().Padding(3)),
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

func labeledGroup(label, key string, group input.GroupWidget) ui.Widget {
	return layout.Column(
		labelpkg.New(label).For(key),
		layout.Box(group).Style(style.New().Width(320)),
	).Gap(6)
}

func boundInput(key, value, placeholder string, send ui.Send[Msg]) input.Widget {
	return input.New(value).Key(key).
		Placeholder(placeholder).
		OnChange(func(value string) {
			send(Msg{Field: key, Value: value})
		})
}

func boundTextArea(key, value, placeholder string, send ui.Send[Msg]) input.TextAreaWidget {
	return input.NewTextArea(value).Key(key).
		Placeholder(placeholder).
		OnChange(func(value string) {
			send(Msg{Field: key, Value: value})
		})
}

func main() {
	ui.Run(ui.NewProgram(Model{Website: "flowui", Price: "10", Token: "flow_live_7Q4M2"},
		Update, View), window.Title("FlowUI InputGroup"),
		window.Size(900, 760),
		window.Theme(base.Light()),
	)
}
