package cmp

import (
	"com.telekom-mms.corp-net-indicator/internal/i18n"
	"com.telekom-mms.corp-net-indicator/internal/model"
	"com.telekom-mms.corp-net-indicator/internal/util"
	"github.com/diamondburned/gotk4/pkg/core/glib"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type VPNDetail struct {
	detail

	ctx *model.Context

	actionClicked chan *model.Credentials

	trustedNetworkLabel *gtk.Label
	connectedImg        *statusIcon
	actionSpinner       *gtk.Spinner
	actionBtn           *gtk.Button
	connectedAtLabel    *gtk.Label
	ipLabel             *gtk.Label
	deviceLabel         *gtk.Label
	certExpiresLabel    *gtk.Label
	loginDialog         *loginDialog

	identityDetail *IdentityDetails
}

// creates new vpn details
func NewVPNDetail(
	// shared context
	context *model.Context,
	// channel to notify connect or disconnect clicks
	vpnActionClicked chan *model.Credentials,
	// parent window to attach login dialog
	parent *gtk.Window,
	// fresh vpn status to process
	status *model.VPNStatus,
	// servers to list in login window
	servers []string,
	// identity details to set button and icon state
	identityDetail *IdentityDetails) *VPNDetail {

	vd := &VPNDetail{detail: newDetail(), ctx: context, actionClicked: vpnActionClicked, identityDetail: identityDetail}

	// create login dialog
	vd.loginDialog = newLoginDialog(parent, servers)

	// create action button with spinner, icons and labels
	vd.actionBtn = gtk.NewButtonWithLabel(i18n.L.Sprintf("Connect VPN"))
	vd.actionBtn.SetHAlign(gtk.AlignEnd)
	if status.IsConnected(false) {
		vd.actionBtn.SetLabel(i18n.L.Sprintf("Disconnect VPN"))
	}
	vd.actionBtn.ConnectClicked(vd.OnActionClicked)
	vd.actionSpinner = gtk.NewSpinner()
	vd.actionSpinner.SetHAlign(gtk.AlignEnd)
	vd.trustedNetworkLabel = gtk.NewLabel(i18n.L.Sprintf("not trusted"))
	vd.connectedImg = NewStatusIcon(status.IsConnected(false))
	vd.connectedAtLabel = gtk.NewLabel(util.FormatDate(status.ConnectedAt))
	vd.ipLabel = gtk.NewLabel(util.FormatValue(status.IP))
	vd.deviceLabel = gtk.NewLabel(util.FormatValue(status.Device))
	vd.certExpiresLabel = gtk.NewLabel(util.FormatDate(status.CertExpiresAt))
	vd.applyTrustedNetwork(status.IsTrustedNetwork(false))

	// set icons, labels and button with spinner in details box
	vd.
		buildBase(i18n.L.Sprintf("VPN Details")).
		addRow(i18n.L.Sprintf("Physical network"), vd.trustedNetworkLabel).
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

// applies new vpn status and calls afterApply after them
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
		vd.connectedImg.SetStatus(status.IsConnected(ctx.Connected))
		vd.applyTrustedNetwork(status.IsTrustedNetwork(ctx.TrustedNetwork))
		if status.ConnectedAt != nil {
			vd.connectedAtLabel.SetText(util.FormatDate(status.ConnectedAt))
		}
		if status.Device != nil {
			vd.deviceLabel.SetText(util.FormatValue(status.Device))
		}
		if status.IP != nil {
			vd.ipLabel.SetText(util.FormatValue(status.IP))
		}
		vd.SetButtonsAfterProgress()
		afterApply()
	})
}

// set button state after progress -> can be after status update or if error occurs
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

// is triggered on action click, triggers action according state
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

// sets widget state and sends credentials over channel
func (vd *VPNDetail) triggerAction(cred *model.Credentials) {
	glib.IdleAdd(func() {
		vd.actionSpinner.Start()
		vd.actionBtn.SetSensitive(false)
		vd.identityDetail.setReLoginBtn(false)
	})
	vd.actionClicked <- cred
}

// triggers dialog closing
func (vd *VPNDetail) Close() {
	vd.loginDialog.close()
}

// apply values related to trusted network setting
func (vd *VPNDetail) applyTrustedNetwork(trustedNetwork bool) {
	if trustedNetwork {
		vd.trustedNetworkLabel.SetText(i18n.L.Sprintf("trusted"))
		vd.actionBtn.SetSensitive(false)
		vd.connectedImg.SetOpacity(0.5)
		vd.connectedAtLabel.SetOpacity(0.5)
		vd.ipLabel.SetOpacity(0.5)
		vd.connectedImg.SetIgnore()
		vd.connectedAtLabel.SetText(util.FormatDate(nil))
		vd.ipLabel.SetText(util.FormatValue(nil))
	} else {
		vd.trustedNetworkLabel.SetText(i18n.L.Sprintf("not trusted"))
		vd.actionBtn.SetSensitive(true)
		vd.connectedImg.SetOpacity(1)
		vd.connectedAtLabel.SetOpacity(1)
		vd.ipLabel.SetOpacity(1)
	}
}
