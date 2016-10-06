package main

import (
	"image/png"
	"log"
	"os"
	"os/user"
	"time"
	"xgbutil/ewmh"
	"xgbutil/keybind"
	"xgbutil/xcursor"
	"xgbutil/xwindow"

	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/mousebind"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/kbinani/screenshot"
)

func main() {
	// Connect to the X server using the DISPLAY environment variable.
	X, err := xgbutil.NewConn()
	check(err)
	mousebind.Initialize(X)
	keybind.Initialize(X)

	initOverlay(X)

	//Setup the hotkey
	err = keybind.KeyPressFun(
		func(X *xgbutil.XUtil, ev xevent.KeyPressEvent) {
			changeCursor(X)
			log.Println("Ready")
		}).Connect(X, X.RootWin(), "Control-Shift-F4", true)
	check(err)

	log.Println("Press ctrl+shift+F4 to take screenshot")
	xevent.Main(X)
}

var x0, y0 int

func begin(X *xgbutil.XUtil, rootX, rootY, eventX, eventY int) (bool, xproto.Cursor) {
	x0 = eventX
	y0 = eventY
	log.Println("Starting screenshot")
	return true, crosshair
}

func step(X *xgbutil.XUtil, rootX, rootY, eventX, eventY int) {
	//TODO: This is not working :/ Wanted to make a square where the mouse is marking
	// startX := x0
	// width := eventX - startX
	// if eventX < x0 {
	// 	width = x0 - eventX
	// 	startX = eventX
	// }
	// startY := y0
	// height := eventY - startY
	// if eventY < y0 {
	// 	height = y0 - eventY
	// 	startY = eventY
	// }
	// region := ewmh.WmOpaqueRegion{X: startX, Y: startY, Width: uint(width), Height: uint(height)}
	// err := ewmh.WmOpaqueRegionSet(X, ov.Id, []ewmh.WmOpaqueRegion{region})
	// check(err)
}

func end(X *xgbutil.XUtil, rootX, rootY, eventX, eventY int) {
	width := eventX - x0
	height := eventY - y0
	closeCursor()
	if width < 10 || height < 10 {
		log.Println("Too small...")
	} else {
		img, err := screenshot.Capture(x0, y0, width, height)
		check(err)

		t := time.Now()
		filename := desktop() + "Screenshot " + t.Format("2006-01-02 15:04:05") + ".png"
		file, err := os.Create(filename)
		check(err)
		defer file.Close()

		png.Encode(file, img)
		log.Println("Saved as ", filename)
	}

}

var ov *xwindow.Window
var crosshair xproto.Cursor
var isActive bool

func initOverlay(X *xgbutil.XUtil) {
	var err error
	ov, err = xwindow.Generate(X)
	check(err)
	crosshair, err = xcursor.CreateCursor(X, xcursor.Crosshair)
	check(err)

	err = ov.CreateChecked(X.RootWin(), 0, 0, 100, 100,
		xproto.CwBackPixel|xproto.CwCursor,
		0xffffffff, uint32(crosshair))
	check(err)
	ewmh.WmNameSet(X, ov.Id, "Screenshot")
	ewmh.WmWindowOpacitySet(X, ov.Id, 0)
	//Drag mouse
	mousebind.Drag(X, ov.Id, ov.Id, "1", true, begin, step, end)

	//Abort
	err = keybind.KeyPressFun(func(X *xgbutil.XUtil, ev xevent.KeyPressEvent) {
		closeCursor()
	}).Connect(X, ov.Id, "Escape", true)
	check(err)
}

func changeCursor(X *xgbutil.XUtil) {
	if !isActive {
		ov.Map()
		ewmh.WmStateReq(X, ov.Id, ewmh.StateToggle, "_NET_WM_STATE_FULLSCREEN")
	}
	isActive = !isActive
}

func closeCursor() {
	if ov != nil {
		ov.Unmap()
	}
	isActive = !isActive
}

func desktop() string {
	usr, err := user.Current()
	if err != nil {
		log.Println(err)
	}
	return usr.HomeDir + "/Desktop/"
}

func check(err error) {
	if err != nil {
		log.Println(err)
		closeCursor()
	}
}
