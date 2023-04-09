package ui

import (
	"context"
	"log"
	"os"

	"de.telekom-mms.corp-net-indicator/internal/i18n"
	"de.telekom-mms.corp-net-indicator/internal/model"
	"github.com/diamondburned/gotk4/pkg/core/glib"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type StatusWindow struct {
	window       *gtk.ApplicationWindow
	quickConnect bool

	ConnectDisconnectClicked chan *model.Credentials
	ReLoginClicked           chan bool
	WindowClosed             chan bool

	// identity data
	loggedInImg      *gtk.Image
	keepAliveAtLabel *gtk.Label
	krbIssuedAtLabel *gtk.Label
	reLoginBtn       *gtk.Button
	reLoginSpinner   *gtk.Spinner

	// vpn data
	trustedNetworkImg *gtk.Image
	connectedImg      *gtk.Image
	actionSpinner     *gtk.Spinner
	actionBtn         *gtk.Button
	connectedAtLabel  *gtk.Label
	ipLabel           *gtk.Label
	deviceLabel       *gtk.Label
	certExpiresLabel  *gtk.Label
	loginDialog       *LoginDialog

	connected bool
	closeChan chan struct{}
}

// TODO separate handle to open window and really instantiated window
func NewStatusWindow() *StatusWindow {
	return &StatusWindow{
		ConnectDisconnectClicked: make(chan *model.Credentials),
		ReLoginClicked:           make(chan bool),
		WindowClosed:             make(chan bool),
	}
}

func (sw *StatusWindow) Open(ctx context.Context, i *model.IdentityStatus, v *model.VPNStatus, quickConnect bool) {
	sw.quickConnect = quickConnect
	if sw.window != nil {
		sw.Close()
		<-sw.closeChan
	}
	go func() {
		sw.closeChan = make(chan struct{})
		app := gtk.NewApplication("de.telekom-mms.corp-net-indicator", gio.ApplicationFlagsNone)
		app.ConnectActivate(func() {
			l := i18n.Localizer()
			// create window
			sw.window = gtk.NewApplicationWindow(app)
			sw.window.SetTitle(l.Sprintf("Corporate Network Status"))
			sw.window.SetResizable(false)
			headerBar := gtk.NewHeaderBar()
			headerBar.SetShowTitleButtons(true)

			sw.loginDialog = NewLoginDialog(&sw.window.Window, v.ServerList)

			details := sw.buildDetails(i, v)

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
			return
		}
		close(sw.closeChan)
		sw.WindowClosed <- true
	}()
}

func (sw *StatusWindow) Close() {
	sw.loginDialog.Close()
	sw.window.Close()
	sw.window.Destroy()
	sw.window = nil
}

func (sw *StatusWindow) IdentityUpdate(ctx context.Context, u *model.IdentityStatus) {
	if sw.loggedInImg == nil {
		return
	}
	glib.IdleAdd(func() {
		sw.reLoginSpinner.Stop()
		setStatusIcon(sw.loggedInImg, u.LoggedIn)
		sw.keepAliveAtLabel.SetText(formatDate(u.LastKeepAliveAt))
		sw.krbIssuedAtLabel.SetText(formatDate(u.KrbIssuedAt))
	})
}

func (sw *StatusWindow) VPNUpdate(ctx context.Context, u *model.VPNStatus) {
	sw.connected = u.Connected
	l := i18n.Localizer()
	if sw.trustedNetworkImg == nil {
		return
	}
	glib.IdleAdd(func() {
		sw.actionSpinner.Stop()
		setStatusIcon(sw.trustedNetworkImg, u.TrustedNetwork)
		setStatusIcon(sw.connectedImg, u.Connected)
		sw.connectedAtLabel.SetText(formatDate(u.ConnectedAt))
		if u.Connected {
			sw.actionBtn.SetLabel(l.Sprintf("Disconnect VPN"))
		} else {
			sw.actionBtn.SetLabel(l.Sprintf("Connect VPN"))
		}
		if u.TrustedNetwork {
			sw.actionBtn.SetSensitive(false)
		} else {
			sw.actionBtn.SetSensitive(true)
		}
		if !u.Connected && !u.TrustedNetwork {
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

func (sw *StatusWindow) NotifyError(err error) {
	sw.actionSpinner.Stop()
	// TODO handle error
}

func (sw *StatusWindow) buildDetails(i *model.IdentityStatus, v *model.VPNStatus) *gtk.Box {
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

func (sw *StatusWindow) buildIdentity(identity *model.IdentityStatus) gtk.Widgetter {
	l := i18n.Localizer()
	box, list := buildListBoxBase("Identity Details")
	sw.reLoginBtn = gtk.NewButtonWithLabel(l.Sprintf("ReLogin"))
	sw.reLoginBtn.SetHAlign(gtk.AlignEnd)
	sw.reLoginBtn.ConnectClicked(sw.handleReLogin)
	sw.loggedInImg = buildStatusIcon(identity.LoggedIn)
	sw.reLoginSpinner = gtk.NewSpinner()
	sw.reLoginSpinner.SetHAlign(gtk.AlignEnd)
	list.Append(addRow(l.Sprintf("Logged in"), sw.reLoginSpinner, sw.reLoginBtn, sw.loggedInImg))
	sw.keepAliveAtLabel = gtk.NewLabel(formatDate(identity.LastKeepAliveAt))
	list.Append(addRow(l.Sprintf("Last Refresh"), sw.keepAliveAtLabel))
	sw.krbIssuedAtLabel = gtk.NewLabel(formatDate(identity.KrbIssuedAt))
	list.Append(addRow(l.Sprintf("Kerberos ticket issued"), sw.krbIssuedAtLabel))
	return box
}

func (sw *StatusWindow) buildVPN(vpn *model.VPNStatus) gtk.Widgetter {
	l := i18n.Localizer()
	box, list := buildListBoxBase("VPN Details")
	sw.connected = vpn.Connected
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
	sw.connectedAtLabel = gtk.NewLabel(formatDate(vpn.ConnectedAt))
	list.Append(addRow(l.Sprintf("Connect at"), sw.connectedAtLabel))
	sw.ipLabel = gtk.NewLabel(vpn.IP)
	list.Append(addRow(l.Sprintf("IP"), sw.ipLabel))
	sw.deviceLabel = gtk.NewLabel(vpn.Device)
	list.Append(addRow(l.Sprintf("Device"), sw.deviceLabel))
	sw.certExpiresLabel = gtk.NewLabel(formatDate(vpn.CertExpiresAt))
	list.Append(addRow(l.Sprintf("Certificate expires"), sw.certExpiresLabel))
	// set correct identity status
	if !vpn.Connected && !vpn.TrustedNetwork {
		setStatusIcon(sw.loggedInImg, false)
		sw.reLoginBtn.SetSensitive(false)
	}
	return box
}

func (sw *StatusWindow) handleAction() {
	if sw.connected {
		sw.actionSpinner.Start()
		sw.actionBtn.SetSensitive(false)
		sw.reLoginBtn.SetSensitive(false)
		go func() {
			log.Println("start 1")
			sw.ConnectDisconnectClicked <- nil
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
				sw.ConnectDisconnectClicked <- result
			}
		}()
	}
}

func (sw *StatusWindow) handleReLogin() {
	go func() {
		glib.IdleAdd(func() {
			sw.reLoginSpinner.Start()
		})
		sw.ReLoginClicked <- true
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
