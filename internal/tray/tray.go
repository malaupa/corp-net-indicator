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
	t.handleVPNStatus(vSer.GetStatus())
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
			sw.Open(t.ctx, iSer.GetStatus(), vSer.GetStatus(), false)
		case <-t.trigger.ClickedCh:
			t.trigger.Disable()
			if t.ctx.Value(model.Connected).(bool) {
				sw.Close()
				t.ctx = model.IncrementProgress(t.ctx)
				vSer.Disconnect()
			} else {
				sw.Open(t.ctx, iSer.GetStatus(), vSer.GetStatus(), true)
			}
		// handle window clicks
		case c := <-sw.ConnectDisconnectClicked:
			t.ctx = model.IncrementProgress(t.ctx)
			if c != nil {
				if err := vSer.Connect(c.Password, c.Server); err != nil {
					sw.NotifyError(err)
				}
			} else {
				vSer.Disconnect()
			}
		case <-sw.ReLoginClicked:
			t.ctx = model.IncrementProgress(t.ctx)
			iSer.ReLogin()
		case <-sw.WindowClosed:
			if t.ctx.Value(model.InProgress).(int) == 0 {
				t.trigger.Enable()
			}
		// handle status updates
		case i := <-iChan:
			t.ctx = model.DecrementProgress(t.ctx)
			log.Println(i)
			t.UpdateIdentity(i)
			sw.IdentityUpdate(t.ctx, i)
		case v := <-vChan:
			t.ctx = model.DecrementProgress(t.ctx)
			t.handleVPNStatus(v)
			log.Printf("tray connected: %v", t.ctx.Value(model.Connected))
			sw.VPNUpdate(t.ctx, v)
		}
	}
}

// destroy tray
func (t *tray) onExit() {

}

func (t *tray) UpdateIdentity(identity *model.IdentityStatus) {
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

func (t *tray) handleVPNStatus(vpn *model.VPNStatus) {
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
