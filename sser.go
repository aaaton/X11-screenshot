package main

import (
	"fmt"
	"image/png"
	"log"
	"os"
	"os/user"
	"time"
	"xgbutil/ewmh"
	"xgbutil/icccm"
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

	//Hotkey for fullscreen
	err = keybind.KeyPressFun(
		func(X *xgbutil.XUtil, ev xevent.KeyPressEvent) {
			if isActive {
				return
			}
			width := int(X.Screen().WidthInPixels)
			height := int(X.Screen().HeightInPixels)

			var filename string
			filename, err = saveImage(0, 0, width, height)
			check(err)
			log.Println("Saved", filename)
		}).Connect(X, X.RootWin(), "Control-Shift-F1", true)
	check(err)

	//Hotkey for window
	err = keybind.KeyPressFun(
		func(X *xgbutil.XUtil, ev xevent.KeyPressEvent) {
			if isActive {
				return
			}
			changeCursor(X, cameraOV)
			fmt.Println("Choose a window")
		}).Connect(X, X.RootWin(), "Control-Shift-F2", true)
	check(err)

	//Setup the hotkey for area
	err = keybind.KeyPressFun(
		func(X *xgbutil.XUtil, ev xevent.KeyPressEvent) {
			if isActive {
				return
			}
			changeCursor(X, crossOV)
			log.Println("Ready for region")
		}).Connect(X, X.RootWin(), "Control-Shift-F3", true)
	check(err)

	log.Println("Press ctrl+shift+F4 to take screenshot")
	xevent.Main(X)
}

var x0, y0 int

func begin(X *xgbutil.XUtil, rootX, rootY, eventX, eventY int) (bool, xproto.Cursor) {
	x0 = eventX
	y0 = eventY
	log.Println("Starting screenshot")
	return true, cross
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
	// err := ewmh.WmOpaqueRegionSet(X, crossOV.Id, []ewmh.WmOpaqueRegion{region})
	// check(err)
}

func end(X *xgbutil.XUtil, rootX, rootY, eventX, eventY int) {
	width := eventX - x0
	height := eventY - y0
	closeCursor()
	if width < 10 || height < 10 {
		log.Println("Too small...")
	} else {
		filename, err := saveImage(x0, y0, width, height)
		check(err)
		log.Println("Saved as ", filename)
	}

}

func saveImage(x0, y0, width, height int) (filename string, err error) {
	img, err := screenshot.Capture(x0, y0, width, height)
	if err != nil {
		return
	}
	t := time.Now()
	filename = desktop() + "Screenshot " + t.Format("2006-01-02 15:04:05") + ".png"
	file, err := os.Create(filename)
	if err != nil {
		return
	}
	defer file.Close()
	png.Encode(file, img)
	return
}

var crossOV, cameraOV *xwindow.Window
var cross, camera xproto.Cursor
var isActive bool

func initOverlay(X *xgbutil.XUtil) {
	var err error
	camera, err = xcursor.CreateCursor(X, xcursor.Circle)
	check(err)
	cross, err = xcursor.CreateCursor(X, xcursor.Crosshair)
	check(err)

	crossOV, err = xwindow.Generate(X)
	check(err)
	err = crossOV.CreateChecked(X.RootWin(), 0, 0, 100, 100,
		xproto.CwBackPixel|xproto.CwCursor,
		0xffffffff, uint32(cross))
	check(err)
	ewmh.WmNameSet(X, crossOV.Id, "Screenshot")
	ewmh.WmWindowOpacitySet(X, crossOV.Id, 0)
	//Drag mouse
	mousebind.Drag(X, crossOV.Id, crossOV.Id, "1", true, begin, step, end)
	//Abort
	cb := keybind.KeyPressFun(func(X *xgbutil.XUtil, ev xevent.KeyPressEvent) {
		closeCursor()
	})
	err = cb.Connect(X, crossOV.Id, "Escape", true)
	check(err)

	//Window shot
	cameraOV, err = xwindow.Generate(X)
	check(err)
	err = cameraOV.CreateChecked(X.RootWin(), 0, 0, 1, 1,
		xproto.CwCursor, uint32(camera))
	check(err)
	ewmh.WmWindowOpacitySet(X, cameraOV.Id, 0)

	//CLick binding
	err = mousebind.ButtonPressFun(
		func(X *xgbutil.XUtil, e xevent.ButtonPressEvent) {
			closeCursor()
			time.Sleep(time.Millisecond * 100)
			xproto.AllowEvents(X.Conn(), xproto.AllowReplayPointer, 0)
			time.Sleep(time.Millisecond * 500)

			win, _ := ewmh.ActiveWindowGet(X)
			fmt.Println(getName(X, win))
			g, err2 := xwindow.New(X, win).DecorGeometry()
			check(err2)
			fmt.Println(g)
		}).Connect(X, X.RootWin(), "1", true, true)
	cameraOV.Listen(xproto.EventMaskFocusChange)
	check(err)

	//Abort
	cb.Connect(X, cameraOV.Id, "Escape", true)

}

func changeCursor(X *xgbutil.XUtil, win *xwindow.Window) {
	if !isActive {
		win.Map()
		ewmh.WmStateReq(X, win.Id, ewmh.StateToggle, "_NET_WM_STATE_FULLSCREEN")
	}
	isActive = !isActive
}

func closeCursor() {
	if crossOV != nil {
		crossOV.Unmap()
	}
	if cameraOV != nil {
		cameraOV.Unmap()
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

func getName(X *xgbutil.XUtil, id xproto.Window) string {
	name, err := ewmh.WmNameGet(X, id)

	// If there was a problem getting _NET_WM_NAME or if its empty,
	// try the old-school version.
	if err != nil || len(name) == 0 {
		name, err = icccm.WmNameGet(X, id)

		// If we still can't find anything, give up.
		if err != nil || len(name) == 0 {
			name = "N/A"
		}
	}
	return name
}

func check(err error) {
	if err != nil {
		log.Println(err)
		closeCursor()
	}
}
