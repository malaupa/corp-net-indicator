package tray

import (
	"de.telekom-mms.corp-net-indicator/internal/assets"
	"github.com/slytomcat/systray"
)

// starts tray
func Run() {
	systray.Run(onReady, onExit)
}

// init tray
func onReady() {
	systray.SetIcon(assets.GetIcon(assets.ShieldOff))
}

// destroy tray
func onExit() {

}
