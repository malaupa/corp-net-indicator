package gtkui

import (
	"context"
	"log"
	"os"

	"de.telekom-mms.corp-net-indicator/internal/i18n"
	"de.telekom-mms.corp-net-indicator/internal/model"
	"de.telekom-mms.corp-net-indicator/internal/util"
	"github.com/diamondburned/gotk4/pkg/core/glib"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type statusWindow struct {
	ctx              context.Context
	quickConnect     bool
	vpnActionClicked chan *model.Credentials
	reLoginClicked   chan bool

	window *gtk.ApplicationWindow

	// identity data TODO -> separate struct/file/component
	loggedInImg      *gtk.Image
	keepAliveAtLabel *gtk.Label
	krbIssuedAtLabel *gtk.Label
	reLoginBtn       *gtk.Button
	reLoginSpinner   *gtk.Spinner

	// vpn data TODO -> separate struct/file/component
	trustedNetworkImg *gtk.Image
	connectedImg      *gtk.Image
	actionSpinner     *gtk.Spinner
	actionBtn         *gtk.Button
	connectedAtLabel  *gtk.Label
	ipLabel           *gtk.Label
	deviceLabel       *gtk.Label
	certExpiresLabel  *gtk.Label
	loginDialog       *LoginDialog
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
		// create window
		sw.window = gtk.NewApplicationWindow(app)
		sw.window.SetTitle(l.Sprintf("Corporate Network Status"))
		sw.window.SetResizable(false)
		headerBar := gtk.NewHeaderBar()
		headerBar.SetShowTitleButtons(true)

		sw.loginDialog = NewLoginDialog(&sw.window.Window, vStatus.ServerList)

		details := sw.buildDetails(iStatus, vStatus)

		// progress
		if ctx.Value(model.InProgress).(int) > 0 {
			sw.actionBtn.SetSensitive(false)
			sw.reLoginBtn.SetSensitive(false)
		}

		sw.window.SetTitlebar(headerBar)
		sw.window.SetChild(details)
		sw.window.Show()

		if sw.quickConnect {
			sw.handleAction()
		}
	})

	if code := app.Run(os.Args); code > 0 {
		// TODO enhance logging
		log.Println("Failed to open window")
	}
}

func (sw *statusWindow) Close() {
	sw.loginDialog.Close()
	sw.window.Close()
	sw.window.Destroy()
}

func (sw *statusWindow) ApplyIdentityStatus(ctx context.Context, status *model.IdentityStatus) {
	glib.IdleAdd(func() {
		sw.reLoginSpinner.Stop()
		setStatusIcon(sw.loggedInImg, status.LoggedIn)
		sw.keepAliveAtLabel.SetText(util.FormatDate(status.LastKeepAliveAt))
		sw.krbIssuedAtLabel.SetText(util.FormatDate(status.KrbIssuedAt))
	})
}

func (sw *statusWindow) ApplyVPNStatus(ctx context.Context, status *model.VPNStatus) {
	sw.ctx = ctx
	l := i18n.Localizer()
	glib.IdleAdd(func() {
		sw.actionSpinner.Stop()
		setStatusIcon(sw.trustedNetworkImg, status.TrustedNetwork)
		setStatusIcon(sw.connectedImg, status.Connected)
		sw.connectedAtLabel.SetText(util.FormatDate(status.ConnectedAt))
		if status.Connected {
			sw.actionBtn.SetLabel(l.Sprintf("Disconnect VPN"))
		} else {
			sw.actionBtn.SetLabel(l.Sprintf("Connect VPN"))
		}
		if status.TrustedNetwork {
			sw.actionBtn.SetSensitive(false)
		} else {
			sw.actionBtn.SetSensitive(true)
		}
		if !status.Connected && !status.TrustedNetwork {
			setStatusIcon(sw.loggedInImg, false)
			sw.reLoginBtn.SetSensitive(false)
		} else {
			sw.reLoginBtn.SetSensitive(true)
		}
		if sw.quickConnect {
			sw.Close()
		}
	})
}

func (sw *statusWindow) NotifyError(err error) {
	sw.actionSpinner.Stop()
	// TODO handle error
}

func (sw *statusWindow) buildDetails(i *model.IdentityStatus, v *model.VPNStatus) *gtk.Box {
	box := gtk.NewBox(gtk.OrientationVertical, 0)
	box.SetMarginTop(30)
	box.SetMarginBottom(30)
	box.SetMarginStart(60)
	box.SetMarginEnd(60)

	box.Append(sw.buildIdentity(i))
	box.Append(sw.buildVPN(v))

	// button box
	btnBox := gtk.NewBox(gtk.OrientationHorizontal, 10)
	btnBox.SetMarginBottom(10)
	btnBox.SetHAlign(gtk.AlignEnd)

	return box
}

func (sw *statusWindow) buildIdentity(identity *model.IdentityStatus) gtk.Widgetter {
	l := i18n.Localizer()
	box, list := buildListBoxBase("Identity Details")
	sw.reLoginBtn = gtk.NewButtonWithLabel(l.Sprintf("ReLogin"))
	sw.reLoginBtn.SetHAlign(gtk.AlignEnd)
	sw.reLoginBtn.ConnectClicked(sw.handleReLogin)
	sw.loggedInImg = buildStatusIcon(identity.LoggedIn)
	sw.reLoginSpinner = gtk.NewSpinner()
	sw.reLoginSpinner.SetHAlign(gtk.AlignEnd)
	list.Append(addRow(l.Sprintf("Logged in"), sw.reLoginSpinner, sw.reLoginBtn, sw.loggedInImg))
	sw.keepAliveAtLabel = gtk.NewLabel(util.FormatDate(identity.LastKeepAliveAt))
	list.Append(addRow(l.Sprintf("Last Refresh"), sw.keepAliveAtLabel))
	sw.krbIssuedAtLabel = gtk.NewLabel(util.FormatDate(identity.KrbIssuedAt))
	list.Append(addRow(l.Sprintf("Kerberos ticket issued"), sw.krbIssuedAtLabel))
	return box
}

func (sw *statusWindow) buildVPN(vpn *model.VPNStatus) gtk.Widgetter {
	l := i18n.Localizer()
	box, list := buildListBoxBase("VPN Details")
	sw.actionBtn = gtk.NewButtonWithLabel(l.Sprintf("Connect VPN"))
	sw.actionBtn.SetHAlign(gtk.AlignEnd)
	if vpn.Connected {
		sw.actionBtn.SetLabel(l.Sprintf("Disconnect VPN"))
	}
	if vpn.TrustedNetwork {
		sw.actionBtn.SetSensitive(false)
	}
	sw.actionBtn.ConnectClicked(sw.handleAction)
	sw.actionSpinner = gtk.NewSpinner()
	sw.actionSpinner.SetHAlign(gtk.AlignEnd)
	sw.trustedNetworkImg = buildStatusIcon(vpn.TrustedNetwork)
	list.Append(addRow(l.Sprintf("Trusted Network"), sw.trustedNetworkImg))
	sw.connectedImg = buildStatusIcon(vpn.Connected)
	list.Append(addRow(l.Sprintf("Connected"), sw.actionSpinner, sw.actionBtn, sw.connectedImg))
	sw.connectedAtLabel = gtk.NewLabel(util.FormatDate(vpn.ConnectedAt))
	list.Append(addRow(l.Sprintf("Connect at"), sw.connectedAtLabel))
	sw.ipLabel = gtk.NewLabel(vpn.IP)
	list.Append(addRow(l.Sprintf("IP"), sw.ipLabel))
	sw.deviceLabel = gtk.NewLabel(vpn.Device)
	list.Append(addRow(l.Sprintf("Device"), sw.deviceLabel))
	sw.certExpiresLabel = gtk.NewLabel(util.FormatDate(vpn.CertExpiresAt))
	list.Append(addRow(l.Sprintf("Certificate expires"), sw.certExpiresLabel))
	// set correct identity status
	if !vpn.Connected && !vpn.TrustedNetwork {
		setStatusIcon(sw.loggedInImg, false)
		sw.reLoginBtn.SetSensitive(false)
	}
	return box
}

func (sw *statusWindow) handleAction() {
	if sw.ctx.Value(model.Connected).(bool) {
		sw.actionSpinner.Start()
		sw.actionBtn.SetSensitive(false)
		sw.reLoginBtn.SetSensitive(false)
		go func() {
			log.Println("start 1")
			sw.vpnActionClicked <- nil
			log.Println("end 1")
		}()
	} else {
		if sw.loginDialog.IsOpen() {
			return
		}
		resultChan := sw.loginDialog.Open()
		go func() {
			result := <-resultChan
			if result != nil {
				log.Println(result)
				glib.IdleAdd(func() {
					sw.actionSpinner.Start()
					sw.actionBtn.SetSensitive(false)
					sw.reLoginBtn.SetSensitive(false)
				})
				sw.vpnActionClicked <- result
			}
		}()
	}
}

func (sw *statusWindow) handleReLogin() {
	go func() {
		glib.IdleAdd(func() {
			sw.reLoginSpinner.Start()
		})
		sw.reLoginClicked <- true
	}()
}

func buildStatusIcon(check bool) *gtk.Image {
	var icon *gtk.Image
	if check {
		icon = gtk.NewImageFromIconName("emblem-default")
	} else {
		icon = gtk.NewImageFromIconName("emblem-important")
	}
	icon.SetIconSize(gtk.IconSizeNormal)
	return icon
}

func setStatusIcon(icon *gtk.Image, check bool) {
	if check {
		icon.SetFromIconName("emblem-default")
	} else {
		icon.SetFromIconName("emblem-important")
	}
}

func buildListBoxBase(title string) (gtk.Widgetter, *gtk.ListBox) {
	l := i18n.Localizer()
	box := gtk.NewBox(gtk.OrientationVertical, 10)
	box.SetMarginBottom(20)
	list := gtk.NewListBox()
	list.SetSelectionMode(gtk.SelectionNone)
	list.SetShowSeparators(true)
	list.AddCSSClass("rich-list")
	label := gtk.NewLabel(l.Sprint(title))
	label.SetMarginBottom(10)
	label.SetHAlign(gtk.AlignStart)
	label.AddCSSClass("title-4")
	box.Append(label)
	frame := gtk.NewFrame("")
	frame.SetChild(list)
	box.Append(frame)
	return box, list
}

func addRow(labelText string, value ...gtk.Widgetter) *gtk.ListBoxRow {
	box := gtk.NewBox(gtk.OrientationHorizontal, 10)
	label := gtk.NewLabel(labelText)
	label.SetHAlign(gtk.AlignStart)
	label.SetHExpand(true)
	box.Append(label)
	for _, w := range value {
		box.Append(w)
	}
	row := gtk.NewListBoxRow()
	row.SetChild(box)
	row.SetActivatable(false)
	row.SetSizeRequest(340, 0)
	return row
}
