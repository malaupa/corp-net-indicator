package gtkui

import (
	"log"

	"de.telekom-mms.corp-net-indicator/internal/i18n"
	"de.telekom-mms.corp-net-indicator/internal/model"
	"de.telekom-mms.corp-net-indicator/internal/ui/gtkui/cmp"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type statusWindow struct {
	ctx              *model.Context
	quickConnect     bool
	vpnActionClicked chan *model.Credentials
	reLoginClicked   chan bool

	window *gtk.ApplicationWindow

	identityDetail *cmp.IdentityDetails
	vpnDetail      *cmp.VPNDetail
}

func NewStatusWindow(ctx *model.Context, vpnActionClicked chan *model.Credentials, reLoginClicked chan bool) *statusWindow {
	return &statusWindow{vpnActionClicked: vpnActionClicked, reLoginClicked: reLoginClicked, ctx: ctx}
}

func (sw *statusWindow) Open(iStatus *model.IdentityStatus, vStatus *model.VPNStatus, servers []string, quickConnect bool) {
	sw.quickConnect = quickConnect
	app := gtk.NewApplication("de.telekom-mms.corp-net-indicator", gio.ApplicationFlagsNone)
	app.ConnectActivate(func() {
		l := i18n.Localizer()
		sw.window = gtk.NewApplicationWindow(app)
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

		sw.identityDetail = cmp.NewIdentityDetails(sw.ctx, sw.reLoginClicked, iStatus)
		sw.vpnDetail = cmp.NewVPNDetail(sw.ctx, sw.vpnActionClicked, &sw.window.Window, vStatus, servers, sw.identityDetail)

		details.Append(sw.identityDetail)
		details.Append(sw.vpnDetail)

		sw.window.SetChild(details)
		sw.window.Show()

		if sw.quickConnect {
			sw.vpnDetail.OnActionClicked()
		}
	})

	if code := app.Run([]string{}); code > 0 {
		// TODO enhance logging
		log.Println("Failed to open window")
	}
}

func (sw *statusWindow) ApplyIdentityStatus(status *model.IdentityStatus) {
	if sw.window == nil {
		return
	}
	sw.identityDetail.Apply(status)
}

func (sw *statusWindow) ApplyVPNStatus(status *model.VPNStatus) {
	if sw.window == nil {
		return
	}
	sw.vpnDetail.Apply(status, func() {
		if sw.quickConnect {
			sw.Close()
		}
	})
}

func (sw *statusWindow) Close() {
	if sw.window == nil {
		return
	}
	sw.vpnDetail.Close()
	sw.window.Close()
	sw.window.Destroy()
}

func (sw *statusWindow) NotifyError(err error) {
	if sw.window == nil {
		return
	}
	// sw.actionSpinner.Stop()
	// TODO handle error
}
