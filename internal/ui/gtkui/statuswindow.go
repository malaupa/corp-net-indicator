package gtkui

import (
	"context"
	"log"
	"os"

	"de.telekom-mms.corp-net-indicator/internal/i18n"
	"de.telekom-mms.corp-net-indicator/internal/model"
	"de.telekom-mms.corp-net-indicator/internal/ui/gtkui/cmp"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type statusWindow struct {
	ctx              context.Context
	quickConnect     bool
	vpnActionClicked chan *model.Credentials
	reLoginClicked   chan bool

	window *gtk.ApplicationWindow

	identityDetail *cmp.IdentityDetails
	vpnDetail      *cmp.VPNDetail
}

func NewStatusWindow(vpnActionClicked chan *model.Credentials, reLoginClicked chan bool) *statusWindow {
	return &statusWindow{vpnActionClicked: vpnActionClicked, reLoginClicked: reLoginClicked}
}

func (sw *statusWindow) Open(ctx context.Context, iStatus *model.IdentityStatus, vStatus *model.VPNStatus, quickConnect bool) {
	sw.ctx = context.WithValue(ctx, model.Connected, vStatus.Connected)
	sw.quickConnect = quickConnect
	app := gtk.NewApplication("de.telekom-mms.corp-net-indicator", gio.ApplicationFlagsNone)
	app.ConnectActivate(func() {
		l := i18n.Localizer()
		sw.window = gtk.NewApplicationWindow(app)
		// sw.window.ConnectCloseRequest(func() (ok bool) {
		// 	sw.Close()
		// 	app.Quit()
		// 	log.Println("quit")
		// 	return true
		// })
		sw.window.SetTitle(l.Sprintf("Corporate Network Status"))
		sw.window.SetResizable(false)

		headerBar := gtk.NewHeaderBar()
		headerBar.SetShowTitleButtons(true)
		sw.window.SetTitlebar(headerBar)

		details := gtk.NewBox(gtk.OrientationVertical, 0)
		details.SetMarginTop(30)
		details.SetMarginBottom(30)
		details.SetMarginStart(60)
		details.SetMarginEnd(60)

		sw.identityDetail = cmp.NewIdentityDetails(ctx, sw.reLoginClicked, iStatus)
		sw.vpnDetail = cmp.NewVPNDetail(ctx, sw.vpnActionClicked, &sw.window.Window, vStatus, sw.identityDetail)

		details.Append(sw.identityDetail)
		details.Append(sw.vpnDetail)

		sw.window.SetChild(details)
		sw.window.Show()

		if sw.quickConnect {
			sw.vpnDetail.OnActionClicked()
		}
	})

	if code := app.Run(os.Args); code > 0 {
		// TODO enhance logging
		log.Println("Failed to open window")
	}
}

func (sw *statusWindow) ApplyIdentityStatus(ctx context.Context, status *model.IdentityStatus) {
	sw.identityDetail.Apply(ctx, status)
}

func (sw *statusWindow) ApplyVPNStatus(ctx context.Context, status *model.VPNStatus) {
	sw.vpnDetail.Apply(ctx, status, func() {
		if sw.quickConnect {
			sw.Close()
		}
	})
}

func (sw *statusWindow) Close() {
	sw.vpnDetail.Close()
	sw.window.Close()
	sw.window.Destroy()
}

func (sw *statusWindow) NotifyError(err error) {
	// sw.actionSpinner.Stop()
	// TODO handle error
}
