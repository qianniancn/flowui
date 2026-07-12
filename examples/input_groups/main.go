package main

import (
	"fmt"

	"github.com/qianniancn/FlowUI/ui"
	"github.com/qianniancn/flowui-icons-lucide"
)

type Model struct {
	Email   string
	Website string
	Price   string
	Search  string
	Token   string
	Last    string
}

type Msg struct {
	Field string
	Value string
	Copy  bool
}

func Update(m *Model, msg Msg) {
	if msg.Copy {
		m.Last = fmt.Sprintf("Copied %s", m.Token)
		return
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
	}
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
				).Gap(22)).Padding(3),
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

func labeledGroup(label, key string, group ui.InputGroupWidget) ui.Widget {
	return ui.Column(
		ui.Label(label).For(key),
		ui.Box(group).Width(320),
	).Gap(6)
}

func boundInput(key, value, placeholder string, send ui.Send[Msg]) ui.InputWidget {
	return ui.Input(key, value).
		Placeholder(placeholder).
		OnChange(func(value string) {
			send(Msg{Field: key, Value: value})
		})
}

func main() {
	ui.Run(
		Model{Website: "flowui", Price: "10", Token: "flow_live_7Q4M2"},
		Update,
		View,
		ui.Title("FlowUI InputGroup"),
		ui.Size(900, 760),
	)
}
