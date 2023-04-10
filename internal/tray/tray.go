package tray

import (
	"context"
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

	ctx context.Context
}

// starts tray
func New(ctx context.Context) {
	t := &tray{ctx: context.WithValue(ctx, model.InProgress, 0)}
	systray.Run(t.onReady, t.onExit)
}

// init tray
func (t *tray) onReady() {
	// translations
	l := i18n.Localizer()
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
	t.applyVPNStatus(vSer.GetStatus())
	t.ApplyIdentityStatus(iSer.GetStatus())

	// listen to status changes
	vChan := vSer.ListenToVPN()
	iChan := iSer.ListenToIdentity()

	// init window
	s := ui.NewStatus()

	// main loop
	for {
		select {
		// handle tray menu clicks
		case <-t.status.ClickedCh:
			s.OpenWindow(t.ctx, iSer.GetStatus(), vSer.GetStatus(), false)
		case <-t.trigger.ClickedCh:
			if t.ctx.Value(model.Connected).(bool) {
				s.Close()
				t.trigger.Disable()
				t.ctx = model.IncrementProgress(t.ctx)
				vSer.Disconnect()
			} else {
				s.OpenWindow(t.ctx, iSer.GetStatus(), vSer.GetStatus(), true)
			}
		// handle window clicks
		case c := <-s.ConnectDisconnectClicked:
			t.trigger.Disable()
			t.ctx = model.IncrementProgress(t.ctx)
			if c != nil {
				if err := vSer.Connect(c.Password, c.Server); err != nil {
					s.NotifyError(err)
				}
			} else {
				vSer.Disconnect()
				// TODO handle error
			}
		case <-s.ReLoginClicked:
			t.ctx = model.IncrementProgress(t.ctx)
			iSer.ReLogin()
			// TODO handle error
		// handle status updates
		case status := <-iChan:
			t.ctx = model.DecrementProgress(t.ctx)
			log.Println(status)
			t.ApplyIdentityStatus(status)
			s.ApplyIdentityStatus(t.ctx, status)
		case status := <-vChan:
			t.ctx = model.DecrementProgress(t.ctx)
			t.trigger.Enable()
			t.applyVPNStatus(status)
			s.ApplyVPNStatus(t.ctx, status)
		}
	}
}

// destroy tray
func (t *tray) onExit() {

}

func (t *tray) ApplyIdentityStatus(identity *model.IdentityStatus) {
	t.ctx = context.WithValue(t.ctx, model.LoggedIn, identity.LoggedIn)
	trusted := t.ctx.Value(model.Trusted).(bool)
	connected := t.ctx.Value(model.Connected).(bool)
	if identity.LoggedIn && (connected || trusted) {
		systray.SetIcon(assets.GetIcon(assets.Umbrella))
	} else {
		if connected || trusted {
			systray.SetIcon(assets.GetIcon(assets.ShieldOn))
		} else {
			systray.SetIcon(assets.GetIcon(assets.ShieldOff))
		}
	}
}

func (t *tray) applyVPNStatus(vpn *model.VPNStatus) {
	l := i18n.Localizer()
	t.ctx = context.WithValue(t.ctx, model.Trusted, vpn.TrustedNetwork)
	t.ctx = context.WithValue(t.ctx, model.Connected, vpn.Connected)
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
