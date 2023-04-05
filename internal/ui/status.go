package ui

import (
	"os"

	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

func OpenStatusWindow() {
	app := gtk.NewApplication("de.telekom-mms.corp-net-indicator", gio.ApplicationFlagsNone)
	app.ConnectActivate(func() { connectStatusWindow(app) })

	if code := app.Run(os.Args); code > 0 {
		os.Exit(code)
	}
}

func connectStatusWindow(app *gtk.Application) {

}
