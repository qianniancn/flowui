package main

import (
	"fmt"
	"time"

	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme/base"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/datepicker"
	"github.com/qianniancn/flowui/widget/input"
	"github.com/qianniancn/flowui/widget/text"
	"github.com/qianniancn/flowui/window"
)

type Model struct {
	FieldDate   time.Time
	Trip        datepicker.DateRange
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
	Range datepicker.DateRange
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

	return layout.Center(
		layout.Box(
			layout.Scroll(layout.Box(layout.Column(
				text.New("FlowUI date components").Size(24),
				text.New(status).Size(16),
				layout.Divider(),
				section("DateField",
					layout.Column(
						layout.Box(datepicker.DateField("date-field", m.FieldDate).
							Label("Appointment date").
							Description("Use the arrow keys or type each segment").
							OnChange(func(value time.Time) {
								send(Msg{Field: fieldDate, Date: value})
							})).Style(style.New().Width(256)),
						layout.Box(datepicker.DateField("date-field-secondary", m.FieldDate).
							Label("Secondary date").
							Variant(input.Secondary).
							OnChange(func(value time.Time) {
								send(Msg{Field: fieldDate, Date: value})
							})).Style(style.New().Width(256)),
					).Gap(12),
				),
				section("DateRangePicker",
					layout.Box(datepicker.DateRangePicker("trip", m.Trip).
						Label("Trip dates").
						Description("Select your check-in and check-out dates").
						FullWidth().
						OnChange(func(value datepicker.DateRange) {
							send(Msg{Field: fieldTrip, Range: value})
						})).Style(style.New().Width(320)),
				),
				section("DatePicker variants",
					layout.Column(
						layout.Box(datePicker("start", fieldStart, m.Start, send).
							Label("Date").
							Description("Choose a date from the calendar or edit each segment").
							FullWidth()).
							Style(style.New().Width(280)),
						layout.Box(datePicker("review", fieldReview, m.Review, send).
							Label("Secondary date").
							Variant(input.Secondary).
							FullWidth()).
							Style(style.New().Width(280)),
						layout.Box(datePicker("picker-only", fieldPickerOnly, m.PickerOnly, send).
							Label("Picker-only date").
							Editable(false).
							FullWidth()).
							Style(style.New().Width(280)),
					).Gap(12),
				),
				section("States",
					layout.Column(
						layout.Box(datePicker("appointment", fieldAppointment, m.Appointment, send).
							Label("Appointment date").
							ErrorMessage("Select today or a future date").
							MinDate(today).
							Invalid(m.Appointment.IsZero()).
							FullWidth()).
							Style(style.New().Width(256)),
						layout.Box(datepicker.New("disabled", today).
							Label("Disabled date").
							Disabled(true).
							FullWidth()).
							Style(style.New().Width(256)),
					).Gap(12),
				),
				section("Full width",
					datePicker("full-width", fieldFull, m.Full, send).
						Label("Full-width date").
						Description("The date input fills the available width").
						Variant(input.Secondary).
						FullWidth(),
				),
			).Gap(18)).Style(style.New().Padding(3)),
			).Vertical(),
		).Style(style.New().FillWidth().MaxWidth(720).Padding(24)),
	)
}

func section(title string, child ui.Widget) ui.Widget {
	return layout.Column(
		text.New(title).Size(18),
		child,
	).Gap(10)
}

func datePicker(key string, field Field, value time.Time, send ui.Send[Msg]) datepicker.Widget {
	return datepicker.New(key, value).
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
		Trip: datepicker.DateRange{
			Start: today,
			End:   today.AddDate(0, 0, 4),
		},
	},
		Update, View), window.Title("FlowUI date components"),
		window.Size(900, 760),
		window.Theme(base.Light()),
	)
}
