package cmp

import (
	"context"

	"de.telekom-mms.corp-net-indicator/internal/i18n"
	"de.telekom-mms.corp-net-indicator/internal/model"
	"de.telekom-mms.corp-net-indicator/internal/util"
	"github.com/diamondburned/gotk4/pkg/core/glib"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type VPNDetail struct {
	detail

	ctx           context.Context
	actionClicked chan *model.Credentials

	trustedNetworkImg *statusIcon
	connectedImg      *statusIcon
	actionSpinner     *gtk.Spinner
	actionBtn         *gtk.Button
	connectedAtLabel  *gtk.Label
	ipLabel           *gtk.Label
	deviceLabel       *gtk.Label
	certExpiresLabel  *gtk.Label
	loginDialog       *loginDialog

	identityDetail *IdentityDetails
}

func NewVPNDetail(ctx context.Context, vpnActionClicked chan *model.Credentials, parent *gtk.Window, status *model.VPNStatus, identityDetail *IdentityDetails) *VPNDetail {
	vd := &VPNDetail{detail: *newDetail(), ctx: ctx, actionClicked: vpnActionClicked, identityDetail: identityDetail}
	l := i18n.Localizer()

	vd.loginDialog = newLoginDialog(parent, status.ServerList)

	vd.actionBtn = gtk.NewButtonWithLabel(l.Sprintf("Connect VPN"))
	vd.actionBtn.SetHAlign(gtk.AlignEnd)
	if status.Connected {
		vd.actionBtn.SetLabel(l.Sprintf("Disconnect VPN"))
	}
	if status.TrustedNetwork {
		vd.actionBtn.SetSensitive(false)
	}
	vd.actionBtn.ConnectClicked(vd.OnActionClicked)
	vd.actionSpinner = gtk.NewSpinner()
	vd.actionSpinner.SetHAlign(gtk.AlignEnd)
	vd.trustedNetworkImg = newStatusIcon(status.TrustedNetwork)
	vd.connectedImg = newStatusIcon(status.Connected)
	vd.connectedAtLabel = gtk.NewLabel(util.FormatDate(status.ConnectedAt))
	vd.ipLabel = gtk.NewLabel(status.IP)
	vd.deviceLabel = gtk.NewLabel(status.Device)
	vd.certExpiresLabel = gtk.NewLabel(util.FormatDate(status.CertExpiresAt))

	vd.
		buildBase(l.Sprintf("VPN Details")).
		addRow(l.Sprintf("Trusted Network"), vd.trustedNetworkImg).
		addRow(l.Sprintf("Connected"), vd.actionSpinner, vd.actionBtn, vd.connectedImg).
		addRow(l.Sprintf("Connect at"), vd.connectedAtLabel).
		addRow(l.Sprintf("IP"), vd.ipLabel).
		addRow(l.Sprintf("Device"), vd.deviceLabel).
		addRow(l.Sprintf("Certificate expires"), vd.certExpiresLabel)

	// set correct identity status
	if !status.Connected && !status.TrustedNetwork {
		vd.identityDetail.setLoggedOut()
		vd.identityDetail.setReLoginBtn(false)
	}
	// progress
	if ctx.Value(model.InProgress).(int) > 0 {
		vd.actionBtn.SetSensitive(false)
		vd.identityDetail.setReLoginBtn(false)
	}
	return vd
}

func (vd *VPNDetail) Apply(ctx context.Context, status *model.VPNStatus, afterApply func()) {
	vd.ctx = ctx
	l := i18n.Localizer()
	glib.IdleAdd(func() {
		vd.actionSpinner.Stop()
		vd.trustedNetworkImg.setStatus(status.TrustedNetwork)
		vd.connectedImg.setStatus(status.Connected)
		vd.connectedAtLabel.SetText(util.FormatDate(status.ConnectedAt))
		if status.Connected {
			vd.actionBtn.SetLabel(l.Sprintf("Disconnect VPN"))
		} else {
			vd.actionBtn.SetLabel(l.Sprintf("Connect VPN"))
		}
		if status.TrustedNetwork {
			vd.actionBtn.SetSensitive(false)
		} else {
			vd.actionBtn.SetSensitive(true)
		}
		if !status.Connected && !status.TrustedNetwork {
			vd.identityDetail.setLoggedOut()
			vd.identityDetail.setReLoginBtn(false)
		} else {
			vd.identityDetail.setReLoginBtn(true)
		}
		afterApply()
	})
}

func (vd *VPNDetail) OnActionClicked() {
	if vd.ctx.Value(model.Connected).(bool) {
		go vd.triggerAction(nil)
	} else {
		resultChan := vd.loginDialog.open()
		go func() {
			result := <-resultChan
			if result != nil {
				vd.triggerAction(result)
			}
		}()
	}
}

func (vd *VPNDetail) triggerAction(cred *model.Credentials) {
	glib.IdleAdd(func() {
		vd.actionSpinner.Start()
		vd.actionBtn.SetSensitive(false)
		vd.identityDetail.setReLoginBtn(false)
	})
	vd.actionClicked <- cred
}

func (vd *VPNDetail) Close() {
	vd.loginDialog.close()
}
