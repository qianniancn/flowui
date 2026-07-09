package main

import (
	"fmt"
	"strings"

	ui "github.com/qianniancn/FlowUI"
)

type Model struct {
	Active  string
	Last    string
	Name    string
	Email   string
	Message string
}

type Msg any

type Open struct {
	Key string
}

type Close struct{}

type ChangeField struct {
	Field string
	Value string
}

func Update(m *Model, msg Msg) {
	switch msg := msg.(type) {
	case Open:
		m.Active = msg.Key
		m.Last = fmt.Sprintf("Opened %s modal", msg.Key)
	case Close:
		if m.Active != "" {
			m.Last = fmt.Sprintf("Closed %s modal", m.Active)
		}
		m.Active = ""
	case ChangeField:
		switch msg.Field {
		case "name":
			m.Name = msg.Value
		case "email":
			m.Email = msg.Value
		case "message":
			m.Message = msg.Value
		}
	}
}

func View(_ *ui.Context, m Model, send ui.Send[Msg]) ui.Widget {
	status := "No modal open"
	if m.Last != "" {
		status = m.Last
	}

	return ui.Stack(
		ui.Stacked(
			ui.Center(
				ui.Box(
					ui.Scroll("modals",
						ui.Column(
							ui.Text("FlowUI Modal").Size(24),
							ui.Text(status).Size(16),
							ui.Divider(),
							section("Default",
								buttonGrid(
									openButton("default", "Open modal", send),
									openButton("form", "Contact form", send),
								),
							),
							section("Placements",
								buttonGrid(
									openButton("top", "Top", send),
									openButton("center", "Center", send),
									openButton("bottom", "Bottom", send),
								),
							),
							section("Backdrop",
								buttonGrid(
									openButton("opaque", "Opaque", send),
									openButton("blur", "Blur", send),
									openButton("transparent", "Transparent", send),
								),
							),
							section("Sizes",
								buttonGrid(
									openButton("xs", "XSmall", send),
									openButton("sm", "Small", send),
									openButton("md", "Medium", send),
									openButton("lg", "Large", send),
									openButton("cover", "Cover", send),
									openButton("full", "Full", send),
								),
							),
							section("Scroll behavior",
								buttonGrid(
									openButton("scroll-inside", "Inside scroll", send),
									openButton("scroll-outside", "Outside scroll", send),
								),
							),
							section("Animations",
								buttonGrid(
									openButton("anim-fade", "Fade", send),
									openButton("anim-slide-down", "Slide down", send),
									openButton("anim-slide-up", "Slide up", send),
									openButton("anim-bounce", "Bounce scale", send),
									openButton("anim-zoom-out", "Zoom out", send),
									openButton("anim-pop", "Pop", send),
								),
							),
							section("Dismiss behavior",
								buttonGrid(
									openButton("locked-backdrop", "No backdrop close", send),
									openButton("locked-keyboard", "No Escape close", send),
								),
							),
						).Gap(18),
					).Vertical(),
				).FillWidth().MaxWidth(760).Padding(24),
			),
		),
		ui.Overlay(modalLayer(m, send)).Expanded(),
	)
}

func section(title string, child ui.Widget) ui.Widget {
	return ui.Column(
		ui.Text(title).Size(18),
		child,
	).Gap(10)
}

func buttonGrid(children ...ui.Widget) ui.Widget {
	items := make([]ui.Widget, 0, len(children))
	for _, child := range children {
		items = append(items, ui.Box(child))
	}
	return ui.Wrap(items...).Gap(8).AlignMiddle()
}

func openButton(key, label string, send ui.Send[Msg]) ui.Widget {
	return ui.Button("open-"+key, ui.Text(label)).
		Variant(ui.ButtonSecondary).
		OnClick(func() {
			send(Open{Key: key})
		})
}

func modalLayer(m Model, send ui.Send[Msg]) ui.Widget {
	return ui.Column(
		baseModal("default", m, send,
			"Welcome to FlowUI",
			"Modal is controlled by your Model and closes by calling OnOpenChange(false).",
		).Icon(ui.Text("!")),
		baseModal("form", m, send,
			"Contact us",
			"Compose inputs and buttons inside the modal body just like any other FlowUI view.",
		).Body(contactForm(m, send)),
		baseModal("top", m, send, "Top placement", placementBody("top")).Placement(ui.ModalTop),
		baseModal("center", m, send, "Center placement", placementBody("center")).Placement(ui.ModalCenter),
		baseModal("bottom", m, send, "Bottom placement", placementBody("bottom")).Placement(ui.ModalBottom),
		baseModal("opaque", m, send, "Opaque backdrop", backdropBody("opaque")).Backdrop(ui.ModalBackdropOpaque),
		baseModal("blur", m, send, "Blur-style backdrop", backdropBody("blur")).Backdrop(ui.ModalBackdropBlur),
		baseModal("transparent", m, send, "Transparent backdrop", backdropBody("transparent")).Backdrop(ui.ModalBackdropTransparent),
		baseModal("xs", m, send, "XSmall modal", sizeBody("xs")).Size(ui.ModalXSmall),
		baseModal("sm", m, send, "Small modal", sizeBody("sm")).Size(ui.ModalSmall),
		baseModal("md", m, send, "Medium modal", sizeBody("md")),
		baseModal("lg", m, send, "Large modal", sizeBody("lg")).Size(ui.ModalLarge),
		baseModal("cover", m, send, "Cover modal", sizeBody("cover")).Size(ui.ModalCover),
		baseModal("full", m, send, "Full modal", sizeBody("full")).Size(ui.ModalFull),
		baseModal("scroll-inside", m, send, "Inside scroll", "").
			Body(longBody("Inside scroll keeps the header and footer fixed while the body moves.")).
			Scroll(ui.ModalScrollInside),
		baseModal("scroll-outside", m, send, "Outside scroll", "").
			Body(longBody("Outside scroll moves the dialog content as a single surface.")).
			Scroll(ui.ModalScrollOutside),
		baseModal("anim-fade", m, send, "Fade animation", animationBody("fade")).
			Animation(ui.ModalAnimationFade),
		baseModal("anim-slide-down", m, send, "Slide down animation", animationBody("slide down")).
			Animation(ui.ModalAnimationSlideDown),
		baseModal("anim-slide-up", m, send, "Slide up animation", animationBody("slide up")).
			Animation(ui.ModalAnimationSlideUp),
		baseModal("anim-bounce", m, send, "Bounce scale animation", animationBody("bounce scale")).
			Animation(ui.ModalAnimationBounceScale),
		baseModal("anim-zoom-out", m, send, "Zoom out animation", animationBody("zoom out")).
			Animation(ui.ModalAnimationZoomOut),
		baseModal("anim-pop", m, send, "Pop animation", animationBody("pop")).
			Animation(ui.ModalAnimationPop),
		baseModal("locked-backdrop", m, send, "Explicit close required", "Clicking the backdrop will not close this modal.").
			Dismissable(false),
		baseModal("locked-keyboard", m, send, "Escape disabled", "Pressing Escape will not close this modal.").
			KeyboardDismissDisabled(true),
	)
}

func baseModal(key string, m Model, send ui.Send[Msg], title, body string) ui.ModalWidget {
	return ui.Modal(key, m.Active == key, title, ui.Text(body)).
		Footer(modalFooter(key, send)).
		OnOpenChange(func(open bool) {
			if !open {
				send(Close{})
			}
		})
}

func modalFooter(key string, send ui.Send[Msg]) ui.Widget {
	return ui.Row(
		ui.Button(key+"-cancel", ui.Text("Cancel")).
			Variant(ui.ButtonSecondary).
			OnClick(func() {
				send(Close{})
			}),
		ui.Button(key+"-confirm", ui.Text("Confirm")).
			OnClick(func() {
				send(Close{})
			}),
	).Gap(8).AlignMiddle()
}

func contactForm(m Model, send ui.Send[Msg]) ui.Widget {
	return ui.Column(
		ui.Input("contact-name", m.Name).
			Hint("Name").
			FullWidth().
			OnChange(func(value string) {
				send(ChangeField{Field: "name", Value: value})
			}),
		ui.Input("contact-email", m.Email).
			Hint("Email").
			FullWidth().
			OnChange(func(value string) {
				send(ChangeField{Field: "email", Value: value})
			}),
		ui.Input("contact-message", m.Message).
			Hint("Message").
			FullWidth().
			OnChange(func(value string) {
				send(ChangeField{Field: "message", Value: value})
			}),
	).Gap(12)
}

func longBody(prefix string) ui.Widget {
	rows := make([]ui.Widget, 0, 14)
	for i := range 14 {
		rows = append(rows, ui.Text(fmt.Sprintf("%s Paragraph %02d shows how longer modal content behaves when it exceeds the viewport height.", prefix, i+1)))
	}
	return ui.Column(rows...).Gap(8)
}

func placementBody(placement string) string {
	return fmt.Sprintf("This modal uses the %s placement option.", placement)
}

func backdropBody(backdrop string) string {
	return fmt.Sprintf("This modal uses the %s backdrop variant.", backdrop)
}

func sizeBody(size string) string {
	return fmt.Sprintf("The %s size changes the dialog width and viewport behavior.", strings.ToUpper(size))
}

func animationBody(animation string) string {
	return fmt.Sprintf("This modal uses the %s animation preset.", animation)
}

func main() {
	ui.Run(Model{}, Update, View,
		ui.Title("FlowUI Modal"),
		ui.Size(900, 720),
	)
}
