package ui

import (
	"log"
	"os"

	"de.telekom-mms.corp-net-indicator/internal/i18n"
	"de.telekom-mms.corp-net-indicator/internal/model"
	"github.com/diamondburned/gotk4/pkg/core/glib"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type Credentials struct {
	Password string
	Server   string
}

type StatusWindow struct {
	ConnectDisconnectClicked chan *Credentials
	ReLoginClicked           chan bool

	// identity data
	loggedInImg      *gtk.Image
	keepAliveAtLabel *gtk.Label
	krbIssuedAtLabel *gtk.Label
	reLoginBtn       *gtk.Button

	// vpn data
}

func NewStatusWindow() *StatusWindow {
	return &StatusWindow{ConnectDisconnectClicked: make(chan *Credentials), ReLoginClicked: make(chan bool)}
}

func (sw *StatusWindow) Open(i *model.IdentityStatus, v *model.VPNStatus, connect bool) {
	go func() {
		app := gtk.NewApplication("de.telekom-mms.corp-net-indicator", gio.ApplicationFlagsNone)
		app.ConnectActivate(func() {
			l := i18n.Localizer()
			// create window
			window := gtk.NewApplicationWindow(app)
			window.SetTitle(l.Sprintf("Corporate Network Status"))
			window.SetResizable(false)
			headerBar := gtk.NewHeaderBar()
			headerBar.SetShowTitleButtons(true)
			icon := gtk.NewImageFromIconName("preferences-system-network")
			icon.SetHAlign(gtk.AlignStart)
			icon.SetIconSize(gtk.IconSizeLarge)
			headerBar.PackStart(icon)

			details := sw.buildDetails(i, v)

			window.SetTitlebar(headerBar)
			window.SetChild(details)
			window.Show()
		})

		if code := app.Run(os.Args); code > 0 {
			// TODO enhance logging
			log.Println("Failed to open window")
		}
	}()
}

func (sw *StatusWindow) IdentityUpdate(u *model.IdentityStatus) {
	go func() {
		glib.IdleAdd(func() {
			if u.LoggedIn {
				sw.loggedInImg.SetFromIconName("emblem-default")
			} else {
				sw.loggedInImg.SetFromIconName("emblem-important")
			}
			sw.keepAliveAtLabel.SetText(formatDate(u.LastKeepAliveAt))
			sw.krbIssuedAtLabel.SetText(formatDate(u.KrbIssuedAt))
		})
	}()
}

func (sw *StatusWindow) VPNUpdate(u *model.VPNStatus) {
	// TODO implement
}

func (sw *StatusWindow) NotifyError(err error) {
	// TODO handle error
}

func (sw *StatusWindow) buildDetails(i *model.IdentityStatus, v *model.VPNStatus) *gtk.Box {
	l := i18n.Localizer()

	box := gtk.NewBox(gtk.OrientationVertical, 0)
	box.SetMarginTop(30)
	box.SetMarginBottom(30)
	box.SetMarginStart(60)
	box.SetMarginEnd(60)

	identityBox := sw.buildIdentity(i)
	box.Append(identityBox)

	// button box
	btnBox := gtk.NewBox(gtk.OrientationHorizontal, 10)
	btnBox.SetMarginBottom(10)
	btnBox.SetHAlign(gtk.AlignEnd)

	// action button
	var btnLabel string
	if v.Connected {
		btnLabel = l.Sprintf("Disconnect to VPN")
	} else {
		btnLabel = l.Sprintf("Connect to VPN")
	}
	actionBtn := gtk.NewButtonWithLabel(btnLabel)
	actionBtn.SetHAlign(gtk.AlignEnd)
	btnBox.Append(actionBtn)
	box.Append(btnBox)

	return box
}

func (sw *StatusWindow) buildIdentity(identity *model.IdentityStatus) gtk.Widgetter {
	sw.reLoginBtn = gtk.NewButtonWithLabel("ReLogin")
	box, list := buildListBoxBase("Identity Detail", nil)
	if identity.LoggedIn {
		sw.loggedInImg = gtk.NewImageFromIconName("emblem-default")
	} else {
		sw.loggedInImg = gtk.NewImageFromIconName("emblem-important")
	}
	sw.loggedInImg.SetIconSize(gtk.IconSizeNormal)
	list.Append(addRow(gtk.NewLabel("Logged in"), sw.loggedInImg, sw.reLoginBtn))
	sw.keepAliveAtLabel = gtk.NewLabel(formatDate(identity.LastKeepAliveAt))
	list.Append(addRow(gtk.NewLabel("Last Refresh"), sw.keepAliveAtLabel))
	sw.krbIssuedAtLabel = gtk.NewLabel(formatDate(identity.KrbIssuedAt))
	list.Append(addRow(gtk.NewLabel("Kerberos ticket issued"), sw.krbIssuedAtLabel))
	return box
}

func buildListBoxBase(title string, titleBtn *gtk.Button) (gtk.Widgetter, *gtk.ListBox) {
	l := i18n.Localizer()
	box := gtk.NewBox(gtk.OrientationVertical, 10)
	box.SetMarginBottom(20)
	list := gtk.NewListBox()
	list.SetSelectionMode(gtk.SelectionNone)
	list.SetShowSeparators(true)
	list.AddCSSClass("rich-list")
	titleBox := gtk.NewBox(gtk.OrientationHorizontal, 10)
	label := gtk.NewLabel(l.Sprint(title))
	label.SetMarginBottom(10)
	label.SetHAlign(gtk.AlignStart)
	label.AddCSSClass("title-4")
	titleBox.Append(label)
	if titleBtn != nil {
		titleBtn.SetHAlign(gtk.AlignEnd)
		titleBox.Append(titleBtn)
	}
	box.Append(titleBox)
	frame := gtk.NewFrame("")
	frame.SetChild(list)
	// frame.SetCanTarget(false)
	box.Append(frame)
	return box, list
}

func addRow(label *gtk.Label, value ...gtk.Widgetter) *gtk.ListBoxRow {
	box := gtk.NewBox(gtk.OrientationHorizontal, 10)
	label.SetHAlign(gtk.AlignStart)
	label.SetHExpand(true)
	box.Append(label)
	for _, w := range value {
		box.Append(w)
	}
	row := gtk.NewListBoxRow()
	row.SetChild(box)
	row.SetActivatable(false)
	return row
}
