// Example simple-mousebinding shows how to grab buttons on the root window and
// respond to them via callback functions. It also shows how to remove such
// callbacks so that they no longer respond to the button events.
// Note that more documentation can be found in the mousebind package.
package main

import (
	"image/png"
	"log"
	"os"

	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/mousebind"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/kbinani/screenshot"
)

func main() {
	// Connect to the X server using the DISPLAY environment variable.
	X, err := xgbutil.NewConn()
	if err != nil {
		log.Fatal(err)
	}
	mousebind.Initialize(X)

	err = mousebind.ButtonPressFun(activate).Connect(X, X.RootWin(), "Control-1", false, true)
	if err != nil {
		log.Fatal(err)
	}

	cb2 := mousebind.ButtonReleaseFun(release)
	err = cb2.Connect(X, X.RootWin(), "Control-Shift-1", false, true)
	if err != nil {
		log.Fatal(err)
	}

	// Finally, start the main event loop. This will route any appropriate
	// ButtonPressEvents to your callback function.
	log.Println("Program initialized. Start pressing mouse buttons!")
	xevent.Main(X)
}

func activate(X *xgbutil.XUtil, e xevent.ButtonPressEvent) {
	log.Println("Activated mouse tracking")
	cb1 := mousebind.ButtonPressFun(
		func(X *xgbutil.XUtil, e xevent.ButtonPressEvent) {
			x0 := int(e.EventX - 100)
			y0 := int(e.EventY - 100)

			img, err := screenshot.Capture(x0, y0, 200, 200)
			if err != nil {
				log.Fatal(err)
			}
			file, _ := os.Create("smallss.png")
			defer file.Close()
			png.Encode(file, img)

			xproto.AllowEvents(X.Conn(), xproto.AllowReplayPointer, 0)
		})
	err := cb1.Connect(X, X.RootWin(), "1", true, true)
	if err != nil {
		log.Fatal(err)
	}
}

func release(X *xgbutil.XUtil, e xevent.ButtonReleaseEvent) {
	// from all ButtonPress *and* ButtonRelease handlers.
	mousebind.Detach(X, X.RootWin())

	log.Printf("Detached all Button{Press,Release}Events from the "+
		"root window (%d).", X.RootWin())
}
