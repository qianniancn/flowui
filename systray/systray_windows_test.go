//go:build windows

package systray

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"image/png"
	"testing"
	"time"

	"github.com/qianniancn/FlowUI/internal/sys/windows"
)

func TestWindowsTaskbarCreatedRestoresIcon(t *testing.T) {
	tray := New().SetTooltip("FlowUI")
	impl := newSystemTrayImpl(tray).(*windowsSystemTray)
	impl.taskbarCreatedMessage = 0xc123
	impl.icon = 42
	impl.visible = false

	originalShellNotifyIcon := shellNotifyIcon
	var gotMessage windows.DWORD
	var gotData windows.NOTIFYICONDATA
	shellNotifyIcon = func(message windows.DWORD, data *windows.NOTIFYICONDATA) error {
		gotMessage = message
		gotData = *data
		return nil
	}
	t.Cleanup(func() {
		shellNotifyIcon = originalShellNotifyIcon
		tray.Destroy()
	})

	impl.wndProc(100, impl.taskbarCreatedMessage, 0, 0)
	if gotMessage != windows.NIM_ADD {
		t.Fatalf("Shell_NotifyIcon message = %d, want NIM_ADD", gotMessage)
	}
	if gotData.HWnd != 100 || gotData.HIcon != 42 {
		t.Fatalf("restored icon = hwnd %d icon %d", gotData.HWnd, gotData.HIcon)
	}
	wantFlags := windows.UINT(windows.NIF_MESSAGE | windows.NIF_ICON | windows.NIF_TIP | windows.NIF_STATE)
	if gotData.UFlags != wantFlags || gotData.DwState != 1 || gotData.DwStateMask != 1 {
		t.Fatalf("restored state = flags %#x state %d mask %d", gotData.UFlags, gotData.DwState, gotData.DwStateMask)
	}
	if !impl.iconAdded {
		t.Fatal("restored icon was not marked as added")
	}
}

func TestWindowsTaskbarCreatedReportsRestoreFailure(t *testing.T) {
	tray := New()
	impl := newSystemTrayImpl(tray).(*windowsSystemTray)
	impl.taskbarCreatedMessage = 0xc123
	want := errors.New("notify failed")

	originalShellNotifyIcon := shellNotifyIcon
	shellNotifyIcon = func(windows.DWORD, *windows.NOTIFYICONDATA) error { return want }
	t.Cleanup(func() {
		shellNotifyIcon = originalShellNotifyIcon
		tray.Destroy()
	})

	impl.wndProc(100, impl.taskbarCreatedMessage, 0, 0)
	select {
	case got := <-tray.Errors():
		if !errors.Is(got, want) {
			t.Fatalf("restore error = %v, want %v", got, want)
		}
	default:
		t.Fatal("restore failure was not reported")
	}
}

func TestCreateIconFromPNG(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 16, 16))
	for y := range 16 {
		for x := range 16 {
			img.SetNRGBA(x, y, color.NRGBA{R: 0x18, G: 0x65, B: 0xd8, A: 0xff})
		}
	}

	var data bytes.Buffer
	if err := png.Encode(&data, img); err != nil {
		t.Fatalf("encode PNG: %v", err)
	}

	icon, err := createIconFromPNG(data.Bytes())
	if err != nil {
		t.Fatalf("create icon: %v", err)
	}
	if icon == 0 {
		t.Fatal("create icon returned a null handle")
	}
	t.Cleanup(func() { _ = windows.DestroyIcon(icon) })
}

func TestWindowsTrayMessageLoopSurvivesRecreation(t *testing.T) {
	for attempt := range 2 {
		tray := New()
		clicked := make(chan struct{}, 1)
		tray.OnRightClick(func() { clicked <- struct{}{} })
		done := make(chan struct{})
		go func() {
			tray.Run()
			close(done)
		}()

		impl, hwnd := waitForWindowsTray(t, tray)
		taskRan := make(chan struct{}, 1)
		impl.post(func() { taskRan <- struct{}{} })
		select {
		case <-taskRan:
		case <-time.After(2 * time.Second):
			tray.Destroy()
			t.Fatal("native task did not run on the tray message loop")
		}

		if err := windows.PostMessage(hwnd, wmSystray, 0, windows.LPARAM(windows.WM_RBUTTONUP)); err != nil {
			tray.Destroy()
			t.Fatalf("post right-click message: %v", err)
		}
		select {
		case <-clicked:
		case <-time.After(2 * time.Second):
			tray.Destroy()
			t.Fatalf("right-click callback did not run on attempt %d", attempt+1)
		}

		tray.Destroy()
		select {
		case <-done:
		case <-time.After(2 * time.Second):
			t.Fatalf("tray message loop did not stop on attempt %d", attempt+1)
		}
	}
}

func TestWindowsRightClickShowsConfiguredMenu(t *testing.T) {
	shown := make(chan struct{}, 1)
	originalTrackPopupMenu := trackWindowsPopupMenu
	trackWindowsPopupMenu = func(
		windows.HMENU,
		windows.UINT,
		int32,
		int32,
		int32,
		windows.HWND,
		*windows.RECT,
	) windows.BOOL {
		shown <- struct{}{}
		return 1
	}
	t.Cleanup(func() { trackWindowsPopupMenu = originalTrackPopupMenu })

	tray := New()
	menu := NewMenu()
	menu.Add("Show window")
	tray.SetMenu(menu)
	done := make(chan struct{})
	go func() {
		tray.Run()
		close(done)
	}()
	_, hwnd := waitForWindowsTray(t, tray)
	if err := windows.PostMessage(hwnd, wmSystray, 0, windows.LPARAM(windows.WM_RBUTTONUP)); err != nil {
		tray.Destroy()
		t.Fatalf("post right-click message: %v", err)
	}
	select {
	case <-shown:
	case <-time.After(2 * time.Second):
		tray.Destroy()
		t.Fatal("configured menu was not shown after right-click")
	}
	tray.Destroy()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("tray message loop did not stop")
	}
}

func TestWindowsImmediateDestroyDoesNotPoisonNextMessageLoop(t *testing.T) {
	for attempt := range 10 {
		tray := New()
		done := make(chan struct{})
		go func() {
			tray.Run()
			close(done)
		}()
		tray.Destroy()
		select {
		case <-done:
		case <-time.After(2 * time.Second):
			t.Fatalf("immediately destroyed tray did not stop on attempt %d", attempt+1)
		}
	}

	tray := New()
	clicked := make(chan struct{}, 1)
	tray.OnRightClick(func() { clicked <- struct{}{} })
	done := make(chan struct{})
	go func() {
		tray.Run()
		close(done)
	}()
	_, hwnd := waitForWindowsTray(t, tray)
	if err := windows.PostMessage(hwnd, wmSystray, 0, windows.LPARAM(windows.WM_RBUTTONUP)); err != nil {
		tray.Destroy()
		t.Fatalf("post right-click message: %v", err)
	}
	select {
	case <-clicked:
	case <-time.After(2 * time.Second):
		tray.Destroy()
		t.Fatal("message loop after immediate destroys did not receive right-click")
	}
	tray.Destroy()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("final tray message loop did not stop")
	}
}

func waitForWindowsTray(t *testing.T, tray *SystemTray) (*windowsSystemTray, windows.HWND) {
	t.Helper()
	select {
	case <-tray.Ready():
	case err := <-tray.Errors():
		tray.Destroy()
		t.Fatalf("tray startup failed: %v", err)
	case <-time.After(2 * time.Second):
		tray.Destroy()
		t.Fatal("tray did not become ready")
	}
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		tray.mu.Lock()
		impl, _ := tray.impl.(*windowsSystemTray)
		tray.mu.Unlock()
		if impl != nil {
			if hwnd := impl.windowHandle(); hwnd != 0 {
				return impl, hwnd
			}
		}
		time.Sleep(5 * time.Millisecond)
	}
	tray.Destroy()
	t.Fatal("tray window was not created")
	return nil, 0
}

func TestWindowsMenuBuild(t *testing.T) {
	menu := NewMenu()
	item := menu.Add("Show FlowUI window")
	wm := newWindowsMenu(menu, nil)
	if wm == nil || wm.hmenu == 0 {
		t.Fatal("menu did not create a native popup handle")
	}
	t.Cleanup(wm.destroyOnThread)
	if item.impl == nil {
		t.Fatal("menu item did not receive a native implementation")
	}
}
