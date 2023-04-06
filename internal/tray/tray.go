package tray

import (
	"log"

	"de.telekom-mms.corp-net-indicator/internal/assets"
	"de.telekom-mms.corp-net-indicator/internal/i18n"
	"de.telekom-mms.corp-net-indicator/internal/model"
	"de.telekom-mms.corp-net-indicator/internal/service"
	"de.telekom-mms.corp-net-indicator/internal/ui"
	"github.com/slytomcat/systray"
)

type tray struct {
	status  *systray.MenuItem
	trigger *systray.MenuItem

	connected bool
}

// starts tray
func Run() {
	systray.Run(onReady, onExit)
}

// init tray
func onReady() {
	// translations
	l := i18n.Localizer()
	// ref holder
	t := &tray{}

	// set up menu
	systray.SetIcon(assets.GetIcon(assets.ShieldOff))
	t.status = systray.AddMenuItem(l.Sprintf("Status"), l.Sprintf("Show Status"))
	t.status.SetIcon(assets.GetIcon(assets.Status))
	t.trigger = systray.AddMenuItem(l.Sprintf("Connect VPN"), l.Sprintf("Connect to VPN"))
	t.trigger.SetIcon(assets.GetIcon(assets.Connect))
	t.trigger.Hide()

	// create services
	iService := service.NewIdentityService()
	vService := service.NewVPNService()
	// update tray
	t.UpdateIdentity(iService.GetStatus())
	t.UpdateVPN(vService.GetStatus())

	// listen to status changes
	iListener := iService.ListenToIdentity()
	vListener := vService.ListenToVPN()

	// init window
	statusWindow := ui.NewStatusWindow()

	// main loop
	for {
		select {
		case <-t.status.ClickedCh:
			statusWindow.Open(&model.Details{Identity: iService.GetStatus(), VPN: vService.GetStatus()}, false)
		case <-t.trigger.ClickedCh:
			if t.connected {
				vService.Disconnect()
			} else {
				statusWindow.Open(&model.Details{Identity: iService.GetStatus(), VPN: vService.GetStatus()}, true)
			}
		case i := <-iListener:
			log.Println("loop")
			log.Println(i)
			select {
			case statusWindow.IdentityChan <- i:
			default:
			}
			t.UpdateIdentity(i)
		case v := <-vListener:
			t.UpdateVPN(v)
		}
	}
}

// destroy tray
func onExit() {

}

func (t *tray) UpdateIdentity(identity *model.IdentityStatus) {
	if identity.LoggedIn {
		systray.SetIcon(assets.GetIcon(assets.Umbrella))
	} else {
		systray.SetIcon(assets.GetIcon(assets.ShieldOn))
	}
}

func (t *tray) UpdateVPN(vpn *model.VPNStatus) {
	l := i18n.Localizer()
	t.connected = vpn.Connected
	if vpn.Connected {
		systray.SetIcon(assets.GetIcon(assets.ShieldOn))
		t.trigger.SetTitle(l.Sprintf("Disconnect VPN"))
	} else {
		if !vpn.TrustedNetwork {
			systray.SetIcon(assets.GetIcon(assets.ShieldOff))
		}
		t.trigger.SetTitle(l.Sprintf("Connect VPN"))
	}
	if vpn.TrustedNetwork {
		systray.SetIcon(assets.GetIcon(assets.ShieldOn))
		t.trigger.Hide()
	} else {
		systray.SetIcon(assets.GetIcon(assets.ShieldOff))
		t.trigger.Show()
	}
}
