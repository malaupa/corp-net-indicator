package cmp

import (
	"com.telekom-mms.corp-net-indicator/internal/i18n"
	"com.telekom-mms.corp-net-indicator/internal/model"
	"github.com/diamondburned/gotk4/pkg/core/glib"
	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type loginDialog struct {
	parent     *gtk.Window
	dialog     *gtk.Dialog
	getServers func() ([]string, error)
}

// creates new login window handle
func newLoginDialog(parent *gtk.Window, getServers func() ([]string, error)) *loginDialog {
	return &loginDialog{parent: parent, getServers: getServers}
}

// triggers dialog opening
func (d *loginDialog) open(onResult func(*model.Credentials)) error {
	servers, err := d.getServers()
	if err != nil {
		return err
	}

	glib.IdleAdd(func() {

		const dialogFlags = 0 |
			gtk.DialogDestroyWithParent |
			gtk.DialogModal |
			gtk.DialogUseHeaderBar

		d.dialog = gtk.NewDialogWithFlags("", d.parent, dialogFlags)
		d.dialog.SetResizable(false)

		// create inputs/entries to collected user data
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
		for _, server := range servers {
			serverListEntry.AppendText(server)
		}
		serverListEntry.SetActive(0)

		// build grid to place entries
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

		// create ok action with click handler
		okBtn := d.dialog.AddButton(i18n.L.Sprintf("Connect"), int(gtk.ResponseOK)).(*gtk.Button)
		okBtn.SetSensitive(false)
		okBtn.AddCSSClass("suggested-action")
		okBtn.ConnectClicked(func() {
			onResult(&model.Credentials{Password: passwordEntry.Text(), Server: serverListEntry.ActiveText()})
			d.close()
		})

		// connect enter in password entry to trigger ok action
		passwordEntry.ConnectActivate(func() {
			if okBtn.Sensitive() {
				okBtn.Activate()
			}
		})
		// activate input validation
		passwordEntry.ConnectChanged(func() {
			if passwordEntry.Text() == "" {
				okBtn.SetSensitive(false)
			} else {
				okBtn.SetSensitive(true)
			}
		})

		// create cancel button with handler to close dialog
		ccBtn := d.dialog.AddButton(i18n.L.Sprintf("Cancel"), int(gtk.ResponseCancel)).(*gtk.Button)
		ccBtn.ConnectClicked(d.close)

		// bind esc to close dialog analogous to cancel
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

		// show dialog
		d.dialog.Show()
	})

	// return result channel -> credentials are given for success
	return nil
}

// closes dialog
func (d *loginDialog) close() {
	if d.dialog != nil {
		d.dialog.Close()
		d.dialog.Destroy()
		d.dialog = nil
	}
}
