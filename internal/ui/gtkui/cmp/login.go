package cmp

import (
	"de.telekom-mms.corp-net-indicator/internal/i18n"
	"de.telekom-mms.corp-net-indicator/internal/model"
	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type loginDialog struct {
	parent     *gtk.Window
	dialog     *gtk.Dialog
	serverList []string
}

func newLoginDialog(parent *gtk.Window, serverList []string) *loginDialog {
	return &loginDialog{parent: parent, serverList: serverList}
}

func (d *loginDialog) open() <-chan *model.Credentials {
	const dialogFlags = 0 |
		gtk.DialogDestroyWithParent |
		gtk.DialogModal |
		gtk.DialogUseHeaderBar

	d.dialog = gtk.NewDialogWithFlags("", d.parent, dialogFlags)
	d.dialog.SetResizable(false)

	passwordLabel := gtk.NewLabel(i18n.L.Sprintf("Password"))
	passwordLabel.SetHAlign(gtk.AlignStart)
	passwordEntry := gtk.NewPasswordEntry()
	passwordEntry.SetHExpand(false)
	passwordEntry.SetVExpand(false)
	passwordEntry.SetHAlign(gtk.AlignStart)
	serverLabel := gtk.NewLabel(i18n.L.Sprintf("Server"))
	serverLabel.SetHAlign(gtk.AlignStart)
	// currently dropdown is buggy, selection is not set in every case
	// serverListEntry := gtk.NewDropDownFromStrings(d.serverList)
	serverListEntry := gtk.NewComboBoxText()
	serverListEntry.SetPopupFixedWidth(true)
	for _, server := range d.serverList {
		serverListEntry.AppendText(server)
	}
	serverListEntry.SetActive(0)

	grid := gtk.NewGrid()
	grid.SetColumnSpacing(10)
	grid.SetRowSpacing(10)
	grid.SetMarginTop(20)
	grid.SetMarginBottom(20)
	grid.SetMarginEnd(20)
	grid.SetMarginStart(20)
	grid.Attach(passwordLabel, 0, 0, 1, 1)
	grid.Attach(passwordEntry, 1, 0, 1, 1)
	grid.Attach(serverLabel, 0, 1, 1, 1)
	grid.Attach(serverListEntry, 1, 1, 1, 1)

	d.dialog.SetChild(grid)

	result := make(chan *model.Credentials)

	okBtn := d.dialog.AddButton(i18n.L.Sprintf("Connect"), int(gtk.ResponseOK)).(*gtk.Button)
	okBtn.SetSensitive(false)
	okBtn.AddCSSClass("suggested-action")
	okBtn.ConnectClicked(func() {
		result <- &model.Credentials{Password: passwordEntry.Text(), Server: serverListEntry.ActiveText()}
		d.close()
	})

	// connect enter in password entry
	passwordEntry.ConnectActivate(func() {
		if okBtn.Sensitive() {
			okBtn.Activate()
		}
	})
	passwordEntry.ConnectChanged(func() {
		if passwordEntry.Text() == "" {
			okBtn.SetSensitive(false)
		} else {
			okBtn.SetSensitive(true)
		}
	})

	ccBtn := d.dialog.AddButton(i18n.L.Sprintf("Cancel"), int(gtk.ResponseCancel)).(*gtk.Button)
	ccBtn.ConnectClicked(d.close)

	// bind esc
	esc := gtk.NewEventControllerKey()
	esc.SetName("dialog-escape")
	esc.ConnectKeyPressed(func(val, code uint, state gdk.ModifierType) bool {
		switch val {
		case gdk.KEY_Escape:
			if ccBtn.Sensitive() {
				ccBtn.Activate()
				return true
			}
		}

		return false
	})
	d.dialog.AddController(esc)

	d.dialog.Show()

	return result
}

func (d *loginDialog) close() {
	if d.dialog != nil {
		d.dialog.Close()
		d.dialog.Destroy()
		d.dialog = nil
	}
}
