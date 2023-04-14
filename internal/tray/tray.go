package tray

import (
	"context"
	"os"
	"os/signal"

	"de.telekom-mms.corp-net-indicator/internal/assets"
	"de.telekom-mms.corp-net-indicator/internal/i18n"
	"de.telekom-mms.corp-net-indicator/internal/model"
	"de.telekom-mms.corp-net-indicator/internal/service"
	"de.telekom-mms.corp-net-indicator/internal/ui"
	"github.com/slytomcat/systray"
)

type tray struct {
	status *ui.Status

	statusItem   *systray.MenuItem
	actionItem   *systray.MenuItem
	startSystray func()
	quitSystray  func()

	ctx context.Context
}

// starts tray
func New(ctx context.Context) *tray {
	t := &tray{ctx: context.WithValue(ctx, model.InProgress, 0), status: ui.NewStatus()}
	// create tray
	t.startSystray, t.quitSystray = systray.RunWithExternalLoop(t.onReady, func() {})
	return t
}

// init tray
func (t *tray) onReady() {
	// translations
	l := i18n.Localizer()
	// set up menu
	systray.SetIcon(assets.GetIcon(assets.ShieldOff))
	t.statusItem = systray.AddMenuItem(l.Sprintf("Status"), l.Sprintf("Show Status"))
	t.statusItem.SetIcon(assets.GetIcon(assets.Status))
	t.actionItem = systray.AddMenuItem(l.Sprintf("Connect VPN"), l.Sprintf("Connect to VPN"))
	t.actionItem.SetIcon(assets.GetIcon(assets.Connect))
	t.actionItem.Hide()
}

func (t *tray) Run() {
	// start tray
	t.startSystray()
	// create services
	vSer := service.NewVPNService()
	iSer := service.NewIdentityService()
	// update tray
	t.applyVPNStatus(vSer.GetStatus())
	t.applyIdentityStatus(iSer.GetStatus())

	// listen to status changes
	vChan := vSer.ListenToVPN()
	iChan := iSer.ListenToIdentity()

	// catch interrupt and clean up
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	// main loop
	for {
		select {
		// handle tray menu clicks
		case <-t.statusItem.ClickedCh:
			t.status.OpenWindow(t.ctx, iSer.GetStatus(), vSer.GetStatus(), false)
		case <-t.actionItem.ClickedCh:
			if t.ctx.Value(model.Connected).(bool) {
				t.status.CloseWindow()
				t.actionItem.Disable()
				t.ctx = model.IncrementProgress(t.ctx)
				vSer.Disconnect()
			} else {
				t.status.OpenWindow(t.ctx, iSer.GetStatus(), vSer.GetStatus(), true)
			}
		// handle window clicks
		case c := <-t.status.ConnectDisconnectClicked:
			t.actionItem.Disable()
			t.ctx = model.IncrementProgress(t.ctx)
			if c != nil {
				if err := vSer.Connect(c.Password, c.Server); err != nil {
					t.status.NotifyError(err)
				}
			} else {
				vSer.Disconnect()
				// TODO handle error
			}
		case <-t.status.ReLoginClicked:
			t.ctx = model.IncrementProgress(t.ctx)
			iSer.ReLogin()
			// TODO handle error
		// handle status updates
		case status := <-iChan:
			t.ctx = model.DecrementProgress(t.ctx)
			t.applyIdentityStatus(status)
		case status := <-vChan:
			t.ctx = model.DecrementProgress(t.ctx)
			t.actionItem.Enable()
			t.applyVPNStatus(status)
		case <-c:
			t.status.CloseWindow()
			vSer.Close()
			iSer.Close()
			t.quitSystray()
			return
		}
	}
}

func (t *tray) applyIdentityStatus(status *model.IdentityStatus) {
	t.ctx = context.WithValue(t.ctx, model.LoggedIn, status.LoggedIn)
	t.status.ApplyIdentityStatus(t.ctx, status)
	trusted := t.ctx.Value(model.Trusted).(bool)
	connected := t.ctx.Value(model.Connected).(bool)
	if status.LoggedIn && (connected || trusted) {
		systray.SetIcon(assets.GetIcon(assets.Umbrella))
	} else {
		if connected || trusted {
			systray.SetIcon(assets.GetIcon(assets.ShieldOn))
		} else {
			systray.SetIcon(assets.GetIcon(assets.ShieldOff))
		}
	}
}

func (t *tray) applyVPNStatus(status *model.VPNStatus) {
	l := i18n.Localizer()
	t.ctx = context.WithValue(t.ctx, model.Trusted, status.TrustedNetwork)
	t.ctx = context.WithValue(t.ctx, model.Connected, status.Connected)
	t.status.ApplyVPNStatus(t.ctx, status)
	t.actionItem.Enable()
	if status.Connected {
		systray.SetIcon(assets.GetIcon(assets.ShieldOn))
		t.actionItem.SetTitle(l.Sprintf("Disconnect VPN"))
		t.actionItem.SetIcon(assets.GetIcon(assets.Disconnect))
		t.actionItem.Show()
	} else {
		t.actionItem.SetTitle(l.Sprintf("Connect VPN"))
		t.actionItem.SetIcon(assets.GetIcon(assets.Connect))
		if status.TrustedNetwork {
			systray.SetIcon(assets.GetIcon(assets.ShieldOn))
			t.actionItem.Hide()
		} else {
			systray.SetIcon(assets.GetIcon(assets.ShieldOff))
			t.actionItem.Show()
		}
	}
}
