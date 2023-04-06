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

type StatusWindow struct {
	IdentityChan chan *model.IdentityStatus
	VPNChan      chan *model.VPNStatus
}

func NewStatusWindow() *StatusWindow {
	return &StatusWindow{IdentityChan: make(chan *model.IdentityStatus), VPNChan: make(chan *model.VPNStatus)}
}

func (sw *StatusWindow) Open(details *model.Details, connect bool) {
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

			details := sw.buildDetails(details)

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

func (sw *StatusWindow) buildDetails(details *model.Details) *gtk.Box {
	l := i18n.Localizer()

	box := gtk.NewBox(gtk.OrientationVertical, 0)
	box.SetMarginTop(30)
	box.SetMarginBottom(30)
	box.SetMarginStart(60)
	box.SetMarginEnd(60)

	headBox := gtk.NewBox(gtk.OrientationHorizontal, 10)
	headBox.SetMarginBottom(20)
	icon := gtk.NewImageFromIconName("preferences-system-network")
	icon.SetHAlign(gtk.AlignStart)
	icon.SetIconSize(gtk.IconSizeLarge)
	headline := gtk.NewLabel(l.Sprintf("Corporate Network Status"))
	headline.AddCSSClass("title-2")
	headBox.Append(icon)
	headBox.Append(headline)
	btnBox := gtk.NewBox(gtk.OrientationHorizontal, 10)
	btnBox.SetMarginBottom(10)
	btnBox.SetHAlign(gtk.AlignEnd)
	var btnLabel string
	if details.VPN.Connected {
		btnLabel = l.Sprintf("Disconnect to VPN")
	} else {
		btnLabel = l.Sprintf("Connect to VPN")
	}
	connectBtn := gtk.NewButtonWithLabel(btnLabel)
	connectBtn.SetHAlign(gtk.AlignEnd)
	btnBox.Append(connectBtn)

	identityBox := sw.buildIdentity(details.Identity)

	// box.Append(headBox)
	box.Append(identityBox)
	box.Append(btnBox)

	return box
}

func (sw *StatusWindow) buildIdentity(identity *model.IdentityStatus) gtk.Widgetter {
	box, list := buildListBoxBase("Identity Detail")
	var loggedInImg *gtk.Image
	if identity.LoggedIn {
		loggedInImg = gtk.NewImageFromIconName("emblem-default")
	} else {
		loggedInImg = gtk.NewImageFromIconName("emblem-important")
	}
	loggedInImg.SetIconSize(gtk.IconSizeNormal)
	list.Append(addRow(gtk.NewLabel("Logged in"), loggedInImg))
	// TODO enhance status
	list.Append(addRow(gtk.NewLabel("Last Refresh"), gtk.NewLabel("12:12")))
	go func() {
		for {
			if u, ok := <-sw.IdentityChan; ok {
				log.Println(u)
				glib.IdleAdd(func() {
					if u.LoggedIn {
						loggedInImg.SetFromIconName("emblem-default")
					} else {
						loggedInImg.SetFromIconName("emblem-important")
					}
				})
			} else {
				log.Println("break")
				break
			}
		}
	}()
	return box
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
	frame.SetCanTarget(false)
	box.Append(frame)
	return box, list
}

func addRow(label *gtk.Label, value gtk.Widgetter) *gtk.ListBoxRow {
	box := gtk.NewBox(gtk.OrientationHorizontal, 10)
	label.SetHAlign(gtk.AlignStart)
	label.SetHExpand(true)
	box.Append(label)
	box.Append(value)
	row := gtk.NewListBoxRow()
	row.SetChild(box)
	return row
}
