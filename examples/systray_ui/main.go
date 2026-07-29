package main

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"sync"

	"github.com/qianniancn/flowui/systray"
	"github.com/qianniancn/flowui/ui"
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
		controller.app.Perform("main", ui.WindowActionRaise)
	}
	if send != nil {
		send(ControllerChanged{State: state})
	}
}

func (controller *Controller) closeDecision() ui.WindowCloseDecision {
	controller.mu.Lock()
	defer controller.mu.Unlock()
	if controller.tray != nil {
		controller.lastAction = "Window closed to tray"
		return ui.WindowCloseKeepAlive
	}
	controller.lastAction = "Window close requested"
	return ui.WindowCloseProceed
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

	return ui.Center(
		ui.Box(
			ui.Column(
				ui.Text("FlowUI System Tray").Size(26),
				ui.Row(
					ui.Text(fmt.Sprintf("Retained model count: %d", model.count)),
					ui.Button("increment", ui.Text("Increment")).
						OnClick(func() { send(Increment{}) }),
				).Gap(12),
				ui.Surface(
					ui.Box(
						ui.Column(
							ui.Text("Tray status").Size(14),
							ui.Text(trayState).Size(22),
							ui.Text("Last action: "+model.state.LastAction),
						).Gap(10),
					).Style(ui.Padding(20)),
				).Variant(ui.SurfaceSecondary),
				ui.Row(
					ui.Button("show-window", ui.Text("Show window")).
						OnClick(func() { send(TrayEvent{Action: WindowButton}) }),
					ui.Button("close-window", ui.Text("Close window")).
						Variant(ui.ButtonSecondary).
						OnClick(func() { send(TrayEvent{Action: WindowCloseButton}) }),
					ui.Button("toggle-tray", ui.Text(trayButton)).
						Variant(ui.ButtonSecondary).
						OnClick(func() { send(TrayEvent{Action: TrayButton}) }),
				).Gap(10),
			).Gap(18),
		).Style(ui.FillWidth().MaxWidth(620).Padding(28)),
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
	controller.window = ui.NewProgramWindow("main", program,
		ui.Title("FlowUI System Tray"),
		ui.Size(760, 500),
		ui.OnWindowCloseRequest(controller.closeDecision),
		ui.RetainModelOnClose(),
	)
	app.Run(controller.window)
}
