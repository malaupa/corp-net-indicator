package tray

import (
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
	iSer := service.NewIdentityService()
	vSer := service.NewVPNService()
	// update tray
	t.UpdateIdentity(iSer.GetStatus())
	t.UpdateVPN(vSer.GetStatus())

	// listen to status changes
	iChan := iSer.ListenToIdentity()
	vChan := vSer.ListenToVPN()

	// init window
	sw := ui.NewStatusWindow()

	// main loop
	for {
		select {
		// handle tray menu clicks
		case <-t.status.ClickedCh:
			sw.Open(iSer.GetStatus(), vSer.GetStatus(), false)
		case <-t.trigger.ClickedCh:
			if t.connected {
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
