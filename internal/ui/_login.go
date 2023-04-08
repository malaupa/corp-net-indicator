package ui

import (
	"os"

	"de.telekom-mms.corp-net-indicator/internal/i18n"
	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

func OpenConnectWindow() {
	go func() {
		app := gtk.NewApplication("de.telekom-mms.corp-net-indicator", gio.ApplicationFlagsNone)
		app.ConnectActivate(func() { openConnectDialog(app) })

		if code := app.Run(os.Args); code > 0 {
			os.Exit(code)
		}
	}()
}

func openConnectDialog(app *gtk.Application) {
	// create localizer
	l := i18n.Localizer()
	// create window
	window := gtk.NewApplicationWindow(app)
	passwordInput := gtk.NewPasswordEntry()
	connectBtn := gtk.NewButtonWithLabel(l.Sprintf("Connect"))
	grid := gtk.NewGrid()
	label := gtk.NewLabel(l.Sprintf("Password"))

	// vpn connect handler
	handleConnect := func() {
		// TODO vpn trigger login
		window.Close()
	}

	window.SetTitle(l.Sprintf("Connect to VPN"))
	window.SetResizable(false)
	// label.SetHAlign(gtk.AlignStart)
	passwordInput.SetHExpand(false)
	passwordInput.SetVExpand(false)
	passwordInput.SetHAlign(gtk.AlignStart)
	passwordInput.SetVAlign(gtk.AlignCenter)
	connectBtn.SetHExpand(false)
	connectBtn.SetVExpand(false)
	connectBtn.SetHAlign(gtk.AlignStart)
	connectBtn.SetVAlign(gtk.AlignCenter)

	// popover := gtk.NewPopover()
	// popErr := gtk.NewLabel(l.Sprintf("Error!"))
	// popErr.AddCSSClass("error")
	// popover.SetChild(popErr)
	// popover.SetParent(grid)
	// popover.SetHasArrow(false)
	// popover.SetPosition(gtk.PosTop)
	toastO := adw.NewToastOverlay()
	toastO.SetChild(grid)
	grid.Attach(label, 0, 0, 1, 1)
	grid.Attach(passwordInput, 0, 1, 1, 1)
	grid.Attach(connectBtn, 1, 1, 1, 1)
	grid.SetColumnSpacing(10)
	grid.SetRowSpacing(10)
	grid.SetMarginTop(20)
	grid.SetMarginBottom(20)
	grid.SetMarginEnd(20)
	grid.SetMarginStart(20)
	window.SetChild(toastO)

	connectBtn.AddCSSClass("suggested-action")
	connectBtn.ConnectClicked(handleConnect)
	passwordInput.ConnectActivate(func() {
		t := adw.NewToast(l.Sprintf("Error!"))
		t.SetTimeout(5)
		toastO.AddToast(t)
		// handleConnect()
	})

	window.Show()
}
