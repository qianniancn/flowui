package main

import (
	"fmt"
	"time"

	"github.com/qianniancn/flowui/ui"
)

type Model struct {
	FieldDate   time.Time
	Trip        ui.DateRange
	Start       time.Time
	Review      time.Time
	PickerOnly  time.Time
	Appointment time.Time
	Full        time.Time
	Last        string
}

type Field string

const (
	fieldStart       Field = "start"
	fieldReview      Field = "review"
	fieldPickerOnly  Field = "picker-only"
	fieldAppointment Field = "appointment"
	fieldFull        Field = "full"
	fieldDate        Field = "date-field"
	fieldTrip        Field = "trip"
)

type Msg struct {
	Field Field
	Date  time.Time
	Range ui.DateRange
}

func Update(m *Model, msg Msg) ui.Cmd[Msg] {
	switch msg.Field {
	case fieldDate:
		m.FieldDate = msg.Date
	case fieldTrip:
		m.Trip = msg.Range
	case fieldStart:
		m.Start = msg.Date
	case fieldReview:
		m.Review = msg.Date
	case fieldPickerOnly:
		m.PickerOnly = msg.Date
	case fieldAppointment:
		m.Appointment = msg.Date
	case fieldFull:
		m.Full = msg.Date
	}
	if msg.Field == fieldTrip {
		m.Last = fmt.Sprintf("%s: %s – %s", msg.Field, formatDate(msg.Range.Start), formatDate(msg.Range.End))
	} else {
		m.Last = fmt.Sprintf("%s selected %s", msg.Field, formatDate(msg.Date))
	}
	return nil
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
				ui.Box(ui.Column(
					ui.Text("FlowUI date components").Size(24),
					ui.Text(status).Size(16),
					ui.Divider(),
					section("DateField",
						ui.Column(
							ui.Box(ui.DateField("date-field", m.FieldDate).
								Label("Appointment date").
								Description("Use the arrow keys or type each segment").
								OnChange(func(value time.Time) {
									send(Msg{Field: fieldDate, Date: value})
								})).Style(ui.Width(256)),
							ui.Box(ui.DateField("date-field-secondary", m.FieldDate).
								Label("Secondary date").
								Variant(ui.InputSecondary).
								OnChange(func(value time.Time) {
									send(Msg{Field: fieldDate, Date: value})
								})).Style(ui.Width(256)),
						).Gap(12),
					),
					section("DateRangePicker",
						ui.Box(ui.DateRangePicker("trip", m.Trip).
							Label("Trip dates").
							Description("Select your check-in and check-out dates").
							FullWidth().
							OnChange(func(value ui.DateRange) {
								send(Msg{Field: fieldTrip, Range: value})
							})).Style(ui.Width(320)),
					),
					section("DatePicker variants",
						ui.Column(
							ui.Box(datePicker("start", fieldStart, m.Start, send).
								Label("Date").
								Description("Choose a date from the calendar or edit each segment").
								FullWidth()).
								Style(ui.Width(280)),
							ui.Box(datePicker("review", fieldReview, m.Review, send).
								Label("Secondary date").
								Variant(ui.InputSecondary).
								FullWidth()).
								Style(ui.Width(280)),
							ui.Box(datePicker("picker-only", fieldPickerOnly, m.PickerOnly, send).
								Label("Picker-only date").
								Editable(false).
								FullWidth()).
								Style(ui.Width(280)),
						).Gap(12),
					),
					section("States",
						ui.Column(
							ui.Box(datePicker("appointment", fieldAppointment, m.Appointment, send).
								Label("Appointment date").
								ErrorMessage("Select today or a future date").
								MinDate(today).
								Invalid(m.Appointment.IsZero()).
								FullWidth()).
								Style(ui.Width(256)),
							ui.Box(ui.DatePicker("disabled", today).
								Label("Disabled date").
								Disabled(true).
								FullWidth()).
								Style(ui.Width(256)),
						).Gap(12),
					),
					section("Full width",
						datePicker("full-width", fieldFull, m.Full, send).
							Label("Full-width date").
							Description("The date input fills the available width").
							Variant(ui.InputSecondary).
							FullWidth(),
					),
				).Gap(18)).Style(ui.Padding(3)),
			).Vertical(),
		).Style(ui.FillWidth().MaxWidth(720).Padding(24)),
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
	today := dateOnly(time.Now())
	ui.Run(ui.NewProgram(Model{
		FieldDate: today,
		Trip: ui.DateRange{
			Start: today,
			End:   today.AddDate(0, 0, 4),
		},
	},
		Update, View), ui.Title("FlowUI date components"),
		ui.Size(900, 760),
	)
}
