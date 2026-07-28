//go:build linux && !android

package systray

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"reflect"
	"testing"

	"github.com/godbus/dbus/v5"
)

func TestLinuxTrayLoadsMenuConfiguredBeforeRun(t *testing.T) {
	tray := New()
	menu := NewMenu()
	menu.Add("Show window")
	tray.SetLabel("FlowUI")
	tray.SetTooltip("FlowUI tray")
	tray.SetMenu(menu)

	impl := newSystemTrayImpl(tray).(*linuxSystemTray)
	if err := impl.loadInitialState(); err != nil {
		t.Fatalf("load initial state: %v", err)
	}
	if impl.label != "FlowUI" || impl.tooltip != "FlowUI tray" || impl.menu != menu {
		t.Fatalf("initial state = label %q tooltip %q menu %p", impl.label, impl.tooltip, impl.menu)
	}
	if impl.menuVersion.Load() != 1 {
		t.Fatalf("initial menu version = %d, want 1", impl.menuVersion.Load())
	}
	tray.Destroy()
}

func TestStatusNotifierPixmapsFromPNGUsesARGBPixels(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 2, 1))
	img.SetNRGBA(0, 0, color.NRGBA{R: 0x11, G: 0x22, B: 0x33, A: 0x44})
	img.SetNRGBA(1, 0, color.NRGBA{R: 0xaa, G: 0xbb, B: 0xcc, A: 0x00})
	var encoded bytes.Buffer
	if err := png.Encode(&encoded, img); err != nil {
		t.Fatalf("encode PNG: %v", err)
	}

	pixmaps, err := statusNotifierPixmapsFromPNG(encoded.Bytes())
	if err != nil {
		t.Fatalf("convert PNG: %v", err)
	}
	if len(pixmaps) != 1 {
		t.Fatalf("pixmap count = %d, want 1", len(pixmaps))
	}
	if pixmaps[0].Width != 2 || pixmaps[0].Height != 1 {
		t.Fatalf("pixmap dimensions = %dx%d, want 2x1", pixmaps[0].Width, pixmaps[0].Height)
	}
	want := []byte{0x44, 0x11, 0x22, 0x33, 0x00, 0xaa, 0xbb, 0xcc}
	if !bytes.Equal(pixmaps[0].Data, want) {
		t.Fatalf("ARGB pixels = %x, want %x", pixmaps[0].Data, want)
	}
}

func TestStatusNotifierPixmapsFromPNGRejectsInvalidData(t *testing.T) {
	if _, err := statusNotifierPixmapsFromPNG([]byte("not a PNG")); err == nil {
		t.Fatal("invalid PNG was accepted")
	}
}

func TestLinuxInitialIconPixmapContainsDecodedPixels(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 1, 1))
	img.SetNRGBA(0, 0, color.NRGBA{R: 0x10, G: 0x20, B: 0x30, A: 0x40})
	var encoded bytes.Buffer
	if err := png.Encode(&encoded, img); err != nil {
		t.Fatalf("encode PNG: %v", err)
	}

	tray := New().SetIcon(encoded.Bytes())
	impl := newSystemTrayImpl(tray).(*linuxSystemTray)
	if err := impl.loadInitialState(); err != nil {
		t.Fatalf("load initial state: %v", err)
	}
	properties := impl.createPropSpec()
	value := properties["org.kde.StatusNotifierItem"]["IconPixmap"].Value
	pixmaps, ok := value.([]statusNotifierPixmap)
	if !ok {
		t.Fatalf("IconPixmap has type %T", value)
	}
	if len(pixmaps) != 1 || !bytes.Equal(pixmaps[0].Data, []byte{0x40, 0x10, 0x20, 0x30}) {
		t.Fatalf("initial IconPixmap = %#v", pixmaps)
	}
	tray.Destroy()
}

func TestLinuxFirstMenuLayoutContainsItemProperties(t *testing.T) {
	menu := NewMenu()
	menu.Add("Show window")
	impl := &linuxSystemTray{
		menu:    menu,
		itemMap: make(map[int32]*dbusMenuItem),
	}

	layout := impl.buildMenuLayout(0, 1)
	if len(layout.V2) != 1 {
		t.Fatalf("root children = %d, want 1", len(layout.V2))
	}
	child := reflect.ValueOf(layout.V2[0].Value())
	properties, ok := child.Field(1).Interface().(map[string]dbus.Variant)
	if !ok {
		t.Fatalf("child properties have type %T", child.Field(1).Interface())
	}
	if got := properties["label"].Value(); got != "Show window" {
		t.Fatalf("first layout label = %v, want Show window", got)
	}
}

func TestLinuxClearingMenuAdvancesLayoutVersion(t *testing.T) {
	impl := &linuxSystemTray{
		menu:    NewMenu(),
		itemMap: make(map[int32]*dbusMenuItem),
	}
	impl.menuVersion.Store(4)
	impl.setMenu(nil)

	if got := impl.menuSnapshot(); got != nil {
		t.Fatalf("menu = %p, want nil", got)
	}
	if got := impl.menuVersion.Load(); got != 5 {
		t.Fatalf("menu version = %d, want 5", got)
	}
}
