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

	trusted   bool
	connected bool
	loggedIn  bool
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
	// TODO use context to hold state
	// TODO hold action in progress in state -> load UI in loading state

	// set up menu
	systray.SetIcon(assets.GetIcon(assets.ShieldOff))
	t.status = systray.AddMenuItem(l.Sprintf("Status"), l.Sprintf("Show Status"))
	t.status.SetIcon(assets.GetIcon(assets.Status))
	t.trigger = systray.AddMenuItem(l.Sprintf("Connect VPN"), l.Sprintf("Connect to VPN"))
	t.trigger.SetIcon(assets.GetIcon(assets.Connect))
	t.trigger.Hide()

	// create services
	vSer := service.NewVPNService()
	iSer := service.NewIdentityService()
	// update tray
	t.UpdateVPN(vSer.GetStatus())
	t.UpdateIdentity(iSer.GetStatus())

	// listen to status changes
	vChan := vSer.ListenToVPN()
	iChan := iSer.ListenToIdentity()

	// init window
	sw := ui.NewStatusWindow()

	// main loop
	for {
		select {
		// handle tray menu clicks
		case <-t.status.ClickedCh:
			sw.Open(iSer.GetStatus(), vSer.GetStatus(), false)
		case <-t.trigger.ClickedCh:
			t.trigger.Disable()
			if t.connected {
				sw.Close()
				vSer.Disconnect()
			} else {
				sw.Open(iSer.GetStatus(), vSer.GetStatus(), true)
			}
		// handle window clicks
		case c := <-sw.ConnectDisconnectClicked:
			if c != nil {
				if err := vSer.Connect(c.Password, c.Server); err != nil {
					sw.NotifyError(err)
				}
			} else {
				vSer.Disconnect()
			}
		case <-sw.ReLoginClicked:
			iSer.ReLogin()
		// handle status updates
		case i := <-iChan:
			log.Println(i)
			sw.IdentityUpdate(i)
			t.UpdateIdentity(i)
		case v := <-vChan:
			sw.VPNUpdate(v)
			t.UpdateVPN(v)
		}
	}
}

// destroy tray
func onExit() {

}

func (t *tray) UpdateIdentity(identity *model.IdentityStatus) {
	t.loggedIn = identity.LoggedIn
	if identity.LoggedIn {
		systray.SetIcon(assets.GetIcon(assets.Umbrella))
	} else {
		if t.connected || t.trusted {
			systray.SetIcon(assets.GetIcon(assets.ShieldOn))
		} else {
			systray.SetIcon(assets.GetIcon(assets.ShieldOff))
		}
	}
}

func (t *tray) UpdateVPN(vpn *model.VPNStatus) {
	l := i18n.Localizer()
	t.trusted = vpn.TrustedNetwork
	t.connected = vpn.Connected
	t.trigger.Enable()
	if vpn.Connected {
		systray.SetIcon(assets.GetIcon(assets.ShieldOn))
		t.trigger.SetTitle(l.Sprintf("Disconnect VPN"))
		t.trigger.SetIcon(assets.GetIcon(assets.Disconnect))
		t.trigger.Show()
	} else {
		t.trigger.SetTitle(l.Sprintf("Connect VPN"))
		t.trigger.SetIcon(assets.GetIcon(assets.Connect))
		if vpn.TrustedNetwork {
			systray.SetIcon(assets.GetIcon(assets.ShieldOn))
			t.trigger.Hide()
		} else {
			systray.SetIcon(assets.GetIcon(assets.ShieldOff))
			t.trigger.Show()
		}
	}
}
