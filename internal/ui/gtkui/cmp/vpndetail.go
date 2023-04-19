package cmp

import (
	"de.telekom-mms.corp-net-indicator/internal/i18n"
	"de.telekom-mms.corp-net-indicator/internal/model"
	"de.telekom-mms.corp-net-indicator/internal/util"
	"github.com/diamondburned/gotk4/pkg/core/glib"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type VPNDetail struct {
	detail

	ctx *model.Context

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

func NewVPNDetail(
	context *model.Context,
	vpnActionClicked chan *model.Credentials,
	parent *gtk.Window,
	status *model.VPNStatus,
	servers []string,
	identityDetail *IdentityDetails) *VPNDetail {
	vd := &VPNDetail{detail: *newDetail(), ctx: context, actionClicked: vpnActionClicked, identityDetail: identityDetail}

	vd.loginDialog = newLoginDialog(parent, servers)

	vd.actionBtn = gtk.NewButtonWithLabel(i18n.L.Sprintf("Connect VPN"))
	vd.actionBtn.SetHAlign(gtk.AlignEnd)
	if status.Connected {
		vd.actionBtn.SetLabel(i18n.L.Sprintf("Disconnect VPN"))
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
	vd.ipLabel = gtk.NewLabel(util.FormatValue(status.IP))
	vd.deviceLabel = gtk.NewLabel(util.FormatValue(status.Device))
	vd.certExpiresLabel = gtk.NewLabel(util.FormatDate(status.CertExpiresAt))

	vd.
		buildBase(i18n.L.Sprintf("VPN Details")).
		addRow(i18n.L.Sprintf("Trusted Network"), vd.trustedNetworkImg).
		addRow(i18n.L.Sprintf("Connected"), vd.actionSpinner, vd.actionBtn, vd.connectedImg).
		addRow(i18n.L.Sprintf("Connected at"), vd.connectedAtLabel).
		addRow(i18n.L.Sprintf("IP"), vd.ipLabel).
		addRow(i18n.L.Sprintf("Device"), vd.deviceLabel).
		addRow(i18n.L.Sprintf("Certificate expires"), vd.certExpiresLabel)

		// set correct identity status
	vd.identityDetail.setButtonAndLoginState()
	// progress
	ctx := vd.ctx.Read()
	if ctx.VPNInProgress {
		vd.identityDetail.setReLoginBtn(false)
		vd.actionBtn.SetSensitive(false)
		vd.actionSpinner.Start()
	}
	return vd
}

func (vd *VPNDetail) Apply(status *model.VPNStatus, afterApply func()) {
	glib.IdleAdd(func() {
		ctx := vd.ctx.Read()
		if ctx.IdentityInProgress || ctx.VPNInProgress {
			if ctx.VPNInProgress {
				vd.actionSpinner.Start()
			}
			vd.actionBtn.SetSensitive(false)
			vd.identityDetail.setReLoginBtn(false)
			return
		}
		vd.trustedNetworkImg.setStatus(status.TrustedNetwork)
		vd.connectedImg.setStatus(status.Connected)
		vd.connectedAtLabel.SetText(util.FormatDate(status.ConnectedAt))
		vd.deviceLabel.SetText(util.FormatValue(status.Device))
		vd.ipLabel.SetText(util.FormatValue(status.IP))
		vd.SetButtonsAfterProgress()
		afterApply()
	})
}

func (vd *VPNDetail) SetButtonsAfterProgress() {
	ctx := vd.ctx.Read()
	vd.actionSpinner.Stop()
	if ctx.Connected {
		vd.actionBtn.SetLabel(i18n.L.Sprintf("Disconnect VPN"))
	} else {
		vd.actionBtn.SetLabel(i18n.L.Sprintf("Connect VPN"))
	}
	if ctx.TrustedNetwork {
		vd.actionBtn.SetSensitive(false)
	} else {
		vd.actionBtn.SetSensitive(true)
	}
	vd.identityDetail.setButtonAndLoginState()
}

func (vd *VPNDetail) OnActionClicked() {
	if vd.ctx.Read().Connected {
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
