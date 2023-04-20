package cmp

import (
	"de.telekom-mms.corp-net-indicator/internal/i18n"
	"de.telekom-mms.corp-net-indicator/internal/model"
	"de.telekom-mms.corp-net-indicator/internal/util"
	"github.com/diamondburned/gotk4/pkg/core/glib"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type IdentityDetails struct {
	detail
	ctx *model.Context

	reLoginClicked chan bool

	loggedInImg      *statusIcon
	keepAliveAtLabel *gtk.Label
	krbIssuedAtLabel *gtk.Label
	reLoginBtn       *gtk.Button
	reLoginSpinner   *gtk.Spinner
}

// creates new identity details box
func NewIdentityDetails(
	// shared context
	ctx *model.Context,
	// channel to notify reLogin clicks
	reLoginClicked chan bool,
	// fresh identity status to process
	status *model.IdentityStatus) *IdentityDetails {

	id := &IdentityDetails{detail: newDetail(), ctx: ctx, reLoginClicked: reLoginClicked}

	// keep all elements in struct to instrument their behavior
	id.reLoginBtn = gtk.NewButtonWithLabel(i18n.L.Sprintf("ReLogin"))
	id.reLoginBtn.SetHAlign(gtk.AlignEnd)
	// set click handler
	id.reLoginBtn.ConnectClicked(id.onReLoginClicked)
	id.loggedInImg = NewStatusIcon(status.LoggedIn)
	id.reLoginSpinner = gtk.NewSpinner()
	id.reLoginSpinner.SetHAlign(gtk.AlignEnd)
	id.keepAliveAtLabel = gtk.NewLabel(util.FormatDate(status.LastKeepAliveAt))
	id.krbIssuedAtLabel = gtk.NewLabel(util.FormatDate(status.KrbIssuedAt))

	// build details box and attach them to the values and actions
	id.
		buildBase(i18n.L.Sprintf("Identity Details")).
		addRow(i18n.L.Sprintf("Logged in"), id.reLoginSpinner, id.reLoginBtn, id.loggedInImg).
		addRow(i18n.L.Sprintf("Last Refresh"), id.keepAliveAtLabel).
		addRow(i18n.L.Sprintf("Kerberos ticket issued"), id.krbIssuedAtLabel)

	// set progress if needed
	if ctx.Read().IdentityInProgress {
		id.reLoginSpinner.Start()
		id.reLoginBtn.SetSensitive(false)
	}

	return id
}

// applies new status to identity details
func (id *IdentityDetails) Apply(status *model.IdentityStatus) {
	glib.IdleAdd(func() {
		ctx := id.ctx.Read()
		// quick path for in progress updates
		if ctx.IdentityInProgress || ctx.VPNInProgress {
			if ctx.IdentityInProgress {
				id.reLoginSpinner.Start()
			}
			id.reLoginBtn.SetSensitive(false)
			return
		}
		// set new status values
		id.loggedInImg.SetStatus(status.LoggedIn)
		id.keepAliveAtLabel.SetText(util.FormatDate(status.LastKeepAliveAt))
		id.krbIssuedAtLabel.SetText(util.FormatDate(status.KrbIssuedAt))
		id.setReLoginBtn(status.LoggedIn)
		// set button state
		id.setButtonAndLoginState()
	})
}

// action handler to trigger login to identity service
func (id *IdentityDetails) onReLoginClicked() {
	go func() {
		glib.IdleAdd(func() {
			id.reLoginSpinner.Start()
			id.reLoginBtn.SetSensitive(false)
		})
		id.reLoginClicked <- true
	}()
}

// sets button state -> true = activated, false deactivated
func (id *IdentityDetails) setReLoginBtn(status bool) {
	id.reLoginBtn.SetSensitive(status)
	id.reLoginSpinner.Stop()
}

// set logged in icon state and button state
func (id *IdentityDetails) setButtonAndLoginState() {
	ctx := id.ctx.Read()
	if !ctx.Connected && !ctx.TrustedNetwork {
		id.loggedInImg.SetStatus(false)
		id.setReLoginBtn(false)
	} else if !ctx.IdentityInProgress {
		id.setReLoginBtn(true)
	}
}
