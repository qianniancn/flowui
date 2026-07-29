package main

import (
	"fmt"
	"strings"

	"github.com/qianniancn/flowui-icons-lucide"
	"github.com/qianniancn/flowui/ui"
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

	return ui.Center(
		ui.Box(
			ui.Scroll("input-groups",
				ui.Box(ui.Column(
					ui.Text("FlowUI InputGroup").Size(24),
					ui.Text(status).Size(14),
					ui.Divider(),
					ui.Wrap(
						ui.Column(
							section("Prefix and suffix",
								ui.Column(
									labeledGroup("Email address", "email",
										ui.InputGroup(boundInput("email", m.Email, "name@email.com", send)).
											Prefix(ui.Icon(lucide.Mail).Size(16)).
											FullWidth()),
									labeledGroup("Website", "website",
										ui.InputGroup(boundInput("website", m.Website, "heroui", send)).
											Prefix(ui.Icon(lucide.Globe).Size(16)).
											Suffix(ui.Text(".com")).
											FullWidth()),
									labeledGroup("Price", "price",
										ui.InputGroup(boundInput("price", m.Price, "10", send).Type(ui.InputNumber)).
											Prefix(ui.Text("$")).
											Suffix(ui.Text("USD")).
											FullWidth()),
								).Gap(12),
							),
							section("Variants",
								ui.Column(
									labeledGroup("Primary", "search-primary",
										ui.InputGroup(boundInput("search-primary", m.Search, "Search projects", send)).
											Prefix(ui.Icon(lucide.Search).Size(16)).
											FullWidth()),
									labeledGroup("Secondary", "search",
										ui.InputGroup(boundInput("search", m.Search, "Search projects", send)).
											Prefix(ui.Icon(lucide.Search).Size(16)).
											Variant(ui.InputSecondary).
											FullWidth()),
								).Gap(12),
							),
						).Gap(28),
						ui.Column(
							section("Text area",
								labeledGroup("Prompt", "prompt",
									ui.InputGroupTextArea(
										boundTextArea("prompt", m.Prompt, "Assign tasks or ask anything", send).Rows(5),
									).
										Prefix(ui.Icon(lucide.MessageSquare).Size(16)).
										Suffix(
											ui.Button("send-prompt", ui.Icon(lucide.SendHorizontal).Size(16)).
												Variant(ui.ButtonGhost).
												Size(ui.ButtonSmall).
												IconOnly().
												Disabled(strings.TrimSpace(m.Prompt) == "").
												OnClick(func() { send(Msg{Submit: true}) }),
										).
										SuffixPadding(8, 4).
										FullWidth(),
								),
							),
							section("Action suffix",
								labeledGroup("API token", "token",
									ui.InputGroup(boundInput("token", m.Token, "flow_live_...", send)).
										Suffix(
											ui.Button("copy-token", ui.Icon(lucide.Copy).Size(16)).
												Variant(ui.ButtonGhost).
												Size(ui.ButtonSmall).
												IconOnly().
												OnClick(func() { send(Msg{Copy: true}) }),
										).
										SuffixPadding(12, 0).
										FullWidth(),
								),
							),
							section("States",
								ui.Column(
									labeledGroup("Invalid", "invalid",
										ui.InputGroup(ui.Input("invalid", "invalid-address").Label("Invalid email")).
											Prefix(ui.Icon(lucide.Mail).Size(16)).
											Suffix(ui.Icon(lucide.CircleAlert).Size(16)).
											Invalid(true).
											FullWidth()),
									labeledGroup("Disabled", "disabled",
										ui.InputGroup(ui.Input("disabled", "name@email.com")).
											Prefix(ui.Icon(lucide.Mail).Size(16)).
											Disabled(true).
											FullWidth()),
								).Gap(12),
							),
						).Gap(28),
					).Gap(28).LineGap(22).AlignStart(),
				).Gap(22)).Style(ui.Padding(3)),
			).Vertical(),
		).Style(ui.FillWidth()).Style(ui.MaxWidth(760)).Style(ui.Padding(24)),
	)
}

func section(title string, child ui.Widget) ui.Widget {
	return ui.Column(
		ui.Text(title).Size(18),
		child,
	).Gap(10)
}

func labeledGroup(label, key string, group ui.InputGroupWidget) ui.Widget {
	return ui.Column(
		ui.Label(label).For(key),
		ui.Box(group).Style(ui.Width(320)),
	).Gap(6)
}

func boundInput(key, value, placeholder string, send ui.Send[Msg]) ui.InputWidget {
	return ui.Input(key, value).
		Placeholder(placeholder).
		OnChange(func(value string) {
			send(Msg{Field: key, Value: value})
		})
}

func boundTextArea(key, value, placeholder string, send ui.Send[Msg]) ui.TextAreaWidget {
	return ui.TextArea(key, value).
		Placeholder(placeholder).
		OnChange(func(value string) {
			send(Msg{Field: key, Value: value})
		})
}

func main() {
	ui.Run(ui.NewProgram(Model{Website: "flowui", Price: "10", Token: "flow_live_7Q4M2"},
		Update, View), ui.Title("FlowUI InputGroup"),
		ui.Size(900, 760),
	)
}
