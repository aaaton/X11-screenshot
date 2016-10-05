// Example simple-mousebinding shows how to grab buttons on the root window and
// respond to them via callback functions. It also shows how to remove such
// callbacks so that they no longer respond to the button events.
// Note that more documentation can be found in the mousebind package.
package main

import (
	"image/png"
	"log"
	"os"
	"os/user"
	"time"
	"xgbutil/keybind"
	"xgbutil/xcursor"
	"xgbutil/xwindow"

	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/mousebind"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/kbinani/screenshot"
)

var normal, crosshair xproto.Cursor
var isActive bool

func main() {
	// Connect to the X server using the DISPLAY environment variable.
	X, err := xgbutil.NewConn()
	if err != nil {
		log.Fatal(err)
	}
	mousebind.Initialize(X)
	keybind.Initialize(X)

	normal, err = xcursor.CreateCursor(X, xcursor.XCursor)
	if err != nil {
		log.Println(err)
	}
	crosshair, err = xcursor.CreateCursor(X, xcursor.Crosshair)
	if err != nil {
		log.Println(err)
	}

	//Setup the hotkey
	err = keybind.KeyPressFun(
		func(X *xgbutil.XUtil, ev xevent.KeyPressEvent) {
			log.Println("Ready")
			changeCursor(X)
			err = mousebind.ButtonPressFun(startSS).Connect(X, X.RootWin(), "1", false, true)
			if err != nil {
				log.Fatal(err)
			}
			err = mousebind.ButtonReleaseFun(endSS).Connect(X, X.RootWin(), "1", false, true)
			if err != nil {
				log.Fatal(err)
			}
		}).Connect(X, X.RootWin(), "Control-Shift-F4", true)
	if err != nil {
		log.Fatal(err)
	}

	// Finally, start the main event loop. This will route any appropriate
	// ButtonPressEvents to your callback function.
	log.Println("Press ctrl+shift+F4 to take screenshot")
	xevent.Main(X)
}

var x0, x1, y0, y1 int

func startSS(X *xgbutil.XUtil, e xevent.ButtonPressEvent) {
	x0 = int(e.EventX)
	y0 = int(e.EventY)
	log.Println("Starting screenshot")
}

func endSS(X *xgbutil.XUtil, e xevent.ButtonReleaseEvent) {
	x1 = int(e.EventX)
	y1 = int(e.EventY)
	width := x1 - x0
	height := y1 - y0
	if width < 10 || height < 10 {
		log.Println("Too small...")
	} else {
		img, err := screenshot.Capture(x0, y0, width, height)
		if err != nil {
			log.Println(err)
		}
		t := time.Now()
		filename := desktop() + "screenshot-" + t.Format("2006-01-02 15:04:05") + ".png"
		file, err := os.Create(filename)
		if err != nil {
			log.Println(err)
		}
		defer file.Close()
		png.Encode(file, img)
		log.Println("Saved as ", filename)
	}
	changeCursor(X)
	mousebind.Detach(X, X.RootWin())
	mousebind.Detach(X, X.RootWin())
}

func changeCursor(X *xgbutil.XUtil) {
	win, err := xwindow.Create(X, X.RootWin())
	if err != nil {
		log.Println(err)
	}
	win, err = win.Parent()
	if err != nil {
		log.Println(err)
	}
	var c xproto.Cursor
	if isActive {
		c = normal
	} else {
		c = crosshair
	}
	win.Change(xproto.CwBackPixel|xproto.CwCursor,
		0xffffffff, uint32(c))
	isActive = !isActive
}

func desktop() string {
	usr, err := user.Current()
	if err != nil {
		log.Println(err)
	}
	return usr.HomeDir + "/Desktop/"
}
