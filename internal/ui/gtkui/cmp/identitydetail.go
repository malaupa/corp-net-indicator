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

func NewIdentityDetails(ctx *model.Context, reLoginClicked chan bool, status *model.IdentityStatus) *IdentityDetails {
	id := &IdentityDetails{detail: *newDetail(), ctx: ctx, reLoginClicked: reLoginClicked}

	l := i18n.Localizer()
	id.reLoginBtn = gtk.NewButtonWithLabel(l.Sprintf("ReLogin"))
	id.reLoginBtn.SetHAlign(gtk.AlignEnd)
	id.reLoginBtn.ConnectClicked(id.onReLoginClicked)
	id.loggedInImg = newStatusIcon(status.LoggedIn)
	id.reLoginSpinner = gtk.NewSpinner()
	id.reLoginSpinner.SetHAlign(gtk.AlignEnd)
	id.keepAliveAtLabel = gtk.NewLabel(util.FormatDate(status.LastKeepAliveAt))
	id.krbIssuedAtLabel = gtk.NewLabel(util.FormatDate(status.KrbIssuedAt))

	id.
		buildBase(l.Sprintf("Identity Details")).
		addRow(l.Sprintf("Logged in"), id.reLoginSpinner, id.reLoginBtn, id.loggedInImg).
		addRow(l.Sprintf("Last Refresh"), id.keepAliveAtLabel).
		addRow(l.Sprintf("Kerberos ticket issued"), id.krbIssuedAtLabel)

	return id
}

func (id *IdentityDetails) Apply(status *model.IdentityStatus) {
	glib.IdleAdd(func() {
		ctx := id.ctx.Read()
		if ctx.IdentityInProgress || ctx.VPNInProgress {
			if ctx.IdentityInProgress {
				id.reLoginSpinner.Start()
			}
			id.reLoginBtn.SetSensitive(false)
			return
		}
		id.loggedInImg.setStatus(status.LoggedIn)
		id.setReLoginBtn(status.LoggedIn)
		id.setButtonAndLoginState()
		id.keepAliveAtLabel.SetText(util.FormatDate(status.LastKeepAliveAt))
		id.krbIssuedAtLabel.SetText(util.FormatDate(status.KrbIssuedAt))
	})
}

func (id *IdentityDetails) onReLoginClicked() {
	go func() {
		glib.IdleAdd(func() {
			id.reLoginSpinner.Start()
			id.reLoginBtn.SetSensitive(false)
		})
		id.reLoginClicked <- true
	}()
}

func (id *IdentityDetails) setReLoginBtn(status bool) {
	id.reLoginBtn.SetSensitive(status)
	id.reLoginSpinner.Stop()
}

func (id *IdentityDetails) setButtonAndLoginState() {
	ctx := id.ctx.Read()
	if !ctx.Connected && !ctx.TrustedNetwork {
		id.loggedInImg.setStatus(false)
		id.setReLoginBtn(false)
	} else if !ctx.IdentityInProgress {
		id.setReLoginBtn(true)
	}
}
