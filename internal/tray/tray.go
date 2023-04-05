package tray

import (
	"de.telekom-mms.corp-net-indicator/internal/assets"
	"de.telekom-mms.corp-net-indicator/internal/i18n"
	"de.telekom-mms.corp-net-indicator/internal/ui"
	"github.com/slytomcat/systray"
)

// starts tray
func Run() {
	systray.Run(onReady, onExit)
}

// init tray
func onReady() {
	l := i18n.Localizer()

	systray.SetIcon(assets.GetIcon(assets.ShieldOff))
	sItem := systray.AddMenuItem(l.Sprintf("Status"), l.Sprintf("Show Status"))
	sItem.SetIcon(assets.GetIcon(assets.Status))
	go func() {
		for {
			<-sItem.ClickedCh
			ui.OpenStatusWindow()
		}
	}()

	// TODO detect current state -> dbus
	systray.AddSeparator()
	cItem := systray.AddMenuItem(l.Sprintf("Connect VPN"), l.Sprintf("Connect to VPN"))
	cItem.SetIcon(assets.GetIcon(assets.Connect))
	go func() {
		for {
			<-cItem.ClickedCh
			ui.OpenConnectWindow()
		}
	}()
}

// destroy tray
func onExit() {

}
