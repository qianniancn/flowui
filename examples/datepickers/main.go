package main

import (
	"fmt"
	"time"

	"github.com/qianniancn/FlowUI/ui"
)

type Model struct {
	Start       time.Time
	Review      time.Time
	Appointment time.Time
	Full        time.Time
	Last        string
}

type Field string

const (
	fieldStart       Field = "start"
	fieldReview      Field = "review"
	fieldAppointment Field = "appointment"
	fieldFull        Field = "full"
)

type Msg struct {
	Field Field
	Date  time.Time
}

func Update(m *Model, msg Msg) {
	switch msg.Field {
	case fieldStart:
		m.Start = msg.Date
	case fieldReview:
		m.Review = msg.Date
	case fieldAppointment:
		m.Appointment = msg.Date
	case fieldFull:
		m.Full = msg.Date
	}
	m.Last = fmt.Sprintf("%s selected %s", msg.Field, formatDate(msg.Date))
}

func View(_ *ui.Context, m Model, send ui.Send[Msg]) ui.Widget {
	status := "No date selected"
	if m.Last != "" {
		status = m.Last
	}
	today := dateOnly(time.Now())

	return ui.Center(
		ui.Box(
			ui.Scroll("datepickers",
				ui.Column(
					ui.Text("FlowUI DatePicker").Size(24),
					ui.Text(status).Size(16),
					ui.Divider(),
					section("Variants",
						ui.Column(
							ui.Box(datePicker("start", fieldStart, m.Start, send)).
								Width(320),
							ui.Box(datePicker("review", fieldReview, m.Review, send).
								Variant(ui.InputSecondary)).
								Width(320),
						).Gap(12),
					),
					section("States",
						ui.Column(
							ui.Box(datePicker("appointment", fieldAppointment, m.Appointment, send).
								MinDate(today).
								Invalid(m.Appointment.IsZero())).
								Width(320),
							ui.Box(ui.DatePicker("disabled", today).
								Disabled(true)).
								Width(320),
						).Gap(12),
					),
					section("Full width",
						datePicker("full-width", fieldFull, m.Full, send).
							Variant(ui.InputSecondary).
							FullWidth(),
					),
				).Gap(18),
			).Vertical(),
		).FillWidth().MaxWidth(720).Padding(24),
	)
}

func section(title string, child ui.Widget) ui.Widget {
	return ui.Column(
		ui.Text(title).Size(18),
		child,
	).Gap(10)
}

func datePicker(key string, field Field, value time.Time, send ui.Send[Msg]) ui.DatePickerWidget {
	return ui.DatePicker(key, value).
		OnChange(func(date time.Time) {
			send(Msg{
				Field: field,
				Date:  date,
			})
		})
}

func formatDate(date time.Time) string {
	if date.IsZero() {
		return "(empty)"
	}
	return date.Format("2006-01-02")
}

func dateOnly(date time.Time) time.Time {
	year, month, day := date.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, date.Location())
}

func main() {
	ui.Run(Model{}, Update, View,
		ui.Title("FlowUI DatePicker"),
		ui.Size(900, 680),
	)
}
