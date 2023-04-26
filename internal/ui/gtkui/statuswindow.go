package gtkui

import (
	"de.telekom-mms.corp-net-indicator/internal/config"
	"de.telekom-mms.corp-net-indicator/internal/i18n"
	"de.telekom-mms.corp-net-indicator/internal/logger"
	"de.telekom-mms.corp-net-indicator/internal/model"
	"de.telekom-mms.corp-net-indicator/internal/ui/gtkui/cmp"
	"github.com/diamondburned/gotk4/pkg/core/glib"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

// holds all window parts
type statusWindow struct {
	ctx              *model.Context
	quickConnect     bool
	vpnActionClicked chan *model.Credentials
	reLoginClicked   chan bool

	window       *gtk.ApplicationWindow
	notification *cmp.Notification

	identityDetail *cmp.IdentityDetails
	vpnDetail      *cmp.VPNDetail
}

// creates new status window
func NewStatusWindow(ctx *model.Context, vpnActionClicked chan *model.Credentials, reLoginClicked chan bool) *statusWindow {
	return &statusWindow{vpnActionClicked: vpnActionClicked, reLoginClicked: reLoginClicked, ctx: ctx}
}

// opens a new status window
// initialization is done with given status data
func (sw *statusWindow) Open(iStatus *model.IdentityStatus, vStatus *model.VPNStatus, servers []string, quickConnect bool) {
	sw.quickConnect = quickConnect
	app := gtk.NewApplication("de.telekom-mms.corp-net-indicator", gio.ApplicationFlagsNone)
	app.ConnectActivate(func() {
		sw.window = gtk.NewApplicationWindow(app)
		sw.window.SetTitle("Corporate Network Status")
		sw.window.SetResizable(false)

		// header menu
		popover := gtk.NewPopover()
		aboutBtn := gtk.NewButtonWithLabel(i18n.L.Sprintf("About"))
		aboutBtn.ConnectClicked(func() {
			// about dialog
			aboutDialog := gtk.NewAboutDialog()
			aboutDialog.SetDestroyWithParent(true)
			aboutDialog.SetModal(true)
			aboutDialog.SetTransientFor(&sw.window.Window)
			aboutDialog.SetProgramName("Corporate Network Status")
			aboutDialog.SetComments("Program to show corporate network status.")
			aboutDialog.SetLogoIconName("applications-internet")
			commit := config.Commit
			if len(commit) > 11 {
				commit = config.Commit[0:11]
			}
			aboutDialog.SetVersion(config.Version + " (" + commit + ")")
			aboutDialog.SetCopyright("© 2023 The Linux Client Team")
			aboutDialog.SetAuthors([]string{"Stefan Schubert"})

			aboutDialog.Show()
			popover.Hide()
		})
		popover.SetChild(aboutBtn)
		menuBtn := gtk.NewMenuButton()
		menuBtn.SetIconName("open-menu-symbolic")
		menuBtn.SetPopover(popover)

		// important to get rounded bottom corners
		headerBar := gtk.NewHeaderBar()
		headerBar.SetShowTitleButtons(true)
		headerBar.PackEnd(menuBtn)
		sw.window.SetTitlebar(headerBar)

		// box for holding all detail boxes
		details := gtk.NewBox(gtk.OrientationVertical, 0)
		details.SetMarginTop(30)
		details.SetMarginBottom(30)
		details.SetMarginStart(60)
		details.SetMarginEnd(60)

		// create details
		sw.identityDetail = cmp.NewIdentityDetails(sw.ctx, sw.reLoginClicked, iStatus)
		sw.vpnDetail = cmp.NewVPNDetail(sw.ctx, sw.vpnActionClicked, &sw.window.Window, vStatus, servers, sw.identityDetail)

		// append all boxes
		details.Append(sw.identityDetail)
		details.Append(sw.vpnDetail)

		// create notification and overlay for them
		sw.notification = cmp.NewNotification()
		overlay := gtk.NewOverlay()
		// details are added as overlay child
		overlay.SetChild(details)
		overlay.AddOverlay(sw.notification.Revealer)

		// show window
		sw.window.SetChild(overlay)
		sw.window.Show()

		if sw.quickConnect && !vStatus.IsConnected(false) {
			sw.vpnDetail.OnActionClicked()
		}
	})

	// this call blocks until window is closed
	if code := app.Run([]string{}); code > 0 {
		logger.Log("Failed to open window")
	}
}

// applies identity status
func (sw *statusWindow) ApplyIdentityStatus(status *model.IdentityStatus) {
	if sw.window == nil {
		return
	}
	sw.identityDetail.Apply(status)
}

// applies vpn status
func (sw *statusWindow) ApplyVPNStatus(status *model.VPNStatus) {
	if sw.window == nil {
		return
	}
	sw.vpnDetail.Apply(status, func() {
		if sw.quickConnect {
			logger.Verbose("Closing window after quick connect")

			sw.Close()
		}
	})
}

// closes window
func (sw *statusWindow) Close() {
	if sw.window == nil {
		return
	}
	sw.vpnDetail.Close()
	sw.window.Close()
	sw.window.Destroy()
}

// triggers notification to show for given error
func (sw *statusWindow) NotifyError(err error) {
	if sw.window == nil {
		return
	}
	glib.IdleAdd(func() {
		sw.vpnDetail.SetButtonsAfterProgress()
		sw.notification.Show(i18n.L.Sprintf("Error: [%v]", err))
	})
}
