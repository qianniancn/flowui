package main

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"sync"

	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/systray"
	"github.com/qianniancn/flowui/theme/base"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/button"
	"github.com/qianniancn/flowui/widget/surface"
	"github.com/qianniancn/flowui/widget/text"
	"github.com/qianniancn/flowui/window"
)

type Controller struct {
	app    *ui.Application
	window ui.WindowSpec

	lifecycleMu sync.Mutex
	mu          sync.Mutex

	tray       *systray.SystemTray
	trayReady  bool
	send       ui.Send[Msg]
	connection uint64
	lastAction string
}

type ControllerState struct {
	TrayStarting bool
	TrayEnabled  bool
	LastAction   string
}

type Model struct {
	controller *Controller
	state      ControllerState
	count      int
}

type Msg any

type ControllerChanged struct {
	State ControllerState
}

type TrayAction string

type TrayEvent struct {
	Action TrayAction
}

type Increment struct{}

const (
	TrayClicked       TrayAction = "Tray icon clicked"
	WindowShown       TrayAction = "Window shown"
	WindowButton      TrayAction = "Show window button clicked"
	WindowCloseButton TrayAction = "Close window button clicked"
	TrayButton        TrayAction = "Tray button clicked"
)

func NewController(app *ui.Application) *Controller {
	return &Controller{
		app:        app,
		lastAction: "Tray is disabled",
	}
}

func (controller *Controller) snapshotLocked() ControllerState {
	return ControllerState{
		TrayStarting: controller.tray != nil && !controller.trayReady,
		TrayEnabled:  controller.tray != nil && controller.trayReady,
		LastAction:   controller.lastAction,
	}
}

func (controller *Controller) snapshot() ControllerState {
	controller.mu.Lock()
	defer controller.mu.Unlock()
	return controller.snapshotLocked()
}

func (controller *Controller) subscribe(ctx context.Context, send ui.Send[Msg]) error {
	controller.mu.Lock()
	controller.connection++
	connection := controller.connection
	controller.send = send
	state := controller.snapshotLocked()
	controller.mu.Unlock()

	send(ControllerChanged{State: state})
	<-ctx.Done()

	controller.mu.Lock()
	if controller.connection == connection {
		controller.send = nil
	}
	controller.mu.Unlock()
	return nil
}

func (controller *Controller) enableTrayLocked() {
	controller.mu.Lock()
	if controller.tray != nil {
		controller.mu.Unlock()
		return
	}
	controller.mu.Unlock()

	controller.app.SetKeepAlive(true)

	tray := systray.New()
	tray.SetIcon(trayIcon()).SetTooltip("FlowUI Tray UI")

	menu := systray.NewMenu()
	menu.Add("Show FlowUI window").OnClick(func() {
		controller.showWindow(WindowShown)
	})
	menu.AddSeparator()
	menu.Add("Quit").OnClick(controller.quit)

	tray.SetMenu(menu)
	tray.OnClick(func() {
		controller.showWindow(TrayClicked)
	})

	controller.mu.Lock()
	if controller.tray != nil {
		controller.mu.Unlock()
		tray.Destroy()
		return
	}
	controller.tray = tray
	controller.trayReady = false
	controller.lastAction = "Starting tray"
	state := controller.snapshotLocked()
	send := controller.send
	controller.mu.Unlock()
	if send != nil {
		send(ControllerChanged{State: state})
	}

	go controller.observeTray(tray)
	go tray.Run()
}

func (controller *Controller) observeTray(tray *systray.SystemTray) {
	ready := tray.Ready()
	for {
		select {
		case <-ready:
			ready = nil
			controller.mu.Lock()
			if controller.tray != tray || controller.trayReady {
				controller.mu.Unlock()
				continue
			}
			controller.trayReady = true
			controller.lastAction = "Tray enabled"
			state := controller.snapshotLocked()
			send := controller.send
			controller.mu.Unlock()
			if send != nil {
				send(ControllerChanged{State: state})
			}
		case err := <-tray.Errors():
			controller.mu.Lock()
			if controller.tray != tray {
				controller.mu.Unlock()
				continue
			}
			controller.lastAction = "Tray error: " + err.Error()
			state := controller.snapshotLocked()
			send := controller.send
			controller.mu.Unlock()
			if send != nil {
				send(ControllerChanged{State: state})
			}
		case <-tray.Done():
			controller.handleTrayDone(tray)
			return
		}
	}
}

func (controller *Controller) handleTrayDone(tray *systray.SystemTray) {
	var terminalErr error
	select {
	case terminalErr = <-tray.Errors():
	default:
	}

	controller.lifecycleMu.Lock()
	defer controller.lifecycleMu.Unlock()

	controller.mu.Lock()
	if controller.tray != tray {
		controller.mu.Unlock()
		return
	}
	wasReady := controller.trayReady
	controller.tray = nil
	controller.trayReady = false
	if terminalErr != nil {
		controller.lastAction = "Tray error: " + terminalErr.Error()
	} else if wasReady {
		controller.lastAction = "Tray stopped"
	} else if controller.lastAction == "Starting tray" {
		controller.lastAction = "Tray failed to start"
	}
	state := controller.snapshotLocked()
	send := controller.send
	controller.mu.Unlock()

	tray.Destroy()
	controller.app.SetKeepAlive(false)
	if send != nil {
		send(ControllerChanged{State: state})
	}
}

func (controller *Controller) disableTrayLocked() {
	controller.mu.Lock()
	tray := controller.tray
	if tray == nil {
		controller.mu.Unlock()
		return
	}
	controller.tray = nil
	controller.trayReady = false
	controller.lastAction = "Tray disabled"
	state := controller.snapshotLocked()
	send := controller.send
	controller.mu.Unlock()

	tray.Destroy()
	controller.app.SetKeepAlive(false)
	if send != nil {
		send(ControllerChanged{State: state})
	}
}

func (controller *Controller) toggleTray() {
	controller.lifecycleMu.Lock()
	defer controller.lifecycleMu.Unlock()

	controller.mu.Lock()
	enabled := controller.tray != nil
	controller.mu.Unlock()
	if enabled {
		controller.disableTrayLocked()
	} else {
		controller.enableTrayLocked()
	}
}

func (controller *Controller) showWindow(action TrayAction) {
	controller.mu.Lock()
	controller.lastAction = string(action)
	state := controller.snapshotLocked()
	send := controller.send
	controller.mu.Unlock()

	if !controller.app.IsOpen("main") {
		controller.app.Open(controller.window)
	} else {
		controller.app.Perform("main", window.ActionRaise)
	}
	if send != nil {
		send(ControllerChanged{State: state})
	}
}

func (controller *Controller) closeDecision() window.CloseDecision {
	controller.mu.Lock()
	defer controller.mu.Unlock()
	if controller.tray != nil {
		controller.lastAction = "Window closed to tray"
		return window.CloseKeepAlive
	}
	controller.lastAction = "Window close requested"
	return window.CloseProceed
}

func (controller *Controller) quit() {
	controller.lifecycleMu.Lock()
	defer controller.lifecycleMu.Unlock()

	controller.mu.Lock()
	tray := controller.tray
	controller.tray = nil
	controller.trayReady = false
	controller.lastAction = "Quit requested from tray"
	controller.mu.Unlock()
	if tray != nil {
		tray.Destroy()
	}
	controller.app.Quit()
}

func Init(controller *Controller) (Model, ui.Cmd[Msg]) {
	return Model{
		controller: controller,
		state:      controller.snapshot(),
	}, nil
}

func Subscriptions(model Model) []ui.Subscription[Msg] {
	return []ui.Subscription[Msg]{
		ui.Subscribe("tray-controller", model.controller.subscribe),
	}
}

func Update(model *Model, msg Msg) ui.Cmd[Msg] {
	switch value := msg.(type) {
	case ControllerChanged:
		model.state = value.State
	case TrayEvent:
		controller := model.controller
		switch value.Action {
		case WindowButton:
			return ui.Do(func(ui.Send[Msg]) {
				controller.showWindow(WindowButton)
			})
		case WindowCloseButton:
			return ui.Do(func(ui.Send[Msg]) {
				controller.app.RequestClose("main")
			})
		case TrayButton:
			return ui.Do(func(ui.Send[Msg]) {
				controller.toggleTray()
			})
		}
	case Increment:
		model.count++
	}
	return nil
}

func View(_ *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
	trayState := "Disabled"
	trayButton := "Enable tray"
	if model.state.TrayStarting {
		trayState = "Starting"
		trayButton = "Disable tray"
	} else if model.state.TrayEnabled {
		trayState = "Enabled"
		trayButton = "Disable tray"
	}

	return layout.Center(
		layout.Box(
			layout.Column(
				text.New("FlowUI System Tray").Size(26),
				layout.Row(
					text.New(fmt.Sprintf("Retained model count: %d", model.count)),
					button.New("Increment").
						OnClick(func() { send(Increment{}) }),
				).Gap(12),
				surface.New(
					layout.Box(
						layout.Column(
							text.New("Tray status").Size(14),
							text.New(trayState).Size(22),
							text.New("Last action: "+model.state.LastAction),
						).Gap(10),
					).Style(style.New().Padding(20)),
				).Variant(surface.Secondary),
				layout.Row(
					button.New("Show window").
						OnClick(func() { send(TrayEvent{Action: WindowButton}) }),
					button.New("Close window").
						Variant(button.Secondary).
						OnClick(func() { send(TrayEvent{Action: WindowCloseButton}) }),
					button.New(trayButton).Key("toggle-tray").
						Variant(button.Secondary).
						OnClick(func() { send(TrayEvent{Action: TrayButton}) }),
				).Gap(10),
			).Gap(18),
		).Style(style.New().FillWidth().MaxWidth(620).Padding(28)),
	)
}

func trayIcon() []byte {
	const size = 16
	img := image.NewNRGBA(image.Rect(0, 0, size, size))
	for y := range size {
		for x := range size {
			dx, dy := x-7, y-7
			if dx*dx+dy*dy <= 42 {
				img.SetNRGBA(x, y, color.NRGBA{R: 0x18, G: 0x65, B: 0xd8, A: 0xff})
			}
		}
	}
	var encoded bytes.Buffer
	if err := png.Encode(&encoded, img); err != nil {
		return nil
	}
	return encoded.Bytes()
}

func main() {
	app := ui.NewApplication()
	controller := NewController(app)
	program := ui.Program[Model, Msg]{
		Init: func() (Model, ui.Cmd[Msg]) {
			return Init(controller)
		},
		Update:        Update,
		Subscriptions: Subscriptions,
		View:          View,
	}
	controller.window = ui.NewWindow("main", program,
		window.Title("FlowUI System Tray"),
		window.Size(760, 500),
		window.OnWindowCloseRequest(controller.closeDecision),
		window.RetainModelOnClose(),
		window.Theme(base.Light()),
	)
	app.Run(controller.window)
}
