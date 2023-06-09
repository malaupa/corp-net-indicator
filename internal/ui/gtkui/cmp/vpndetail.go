package cmp

import (
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"os"
	"os/exec"

	"github.com/diamondburned/gotk4/pkg/core/glib"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/telekom-mms/corp-net-indicator/internal/i18n"
	"github.com/telekom-mms/corp-net-indicator/internal/logger"
	"github.com/telekom-mms/corp-net-indicator/internal/model"
	"github.com/telekom-mms/corp-net-indicator/internal/util"
	"github.com/telekom-mms/oc-daemon/pkg/vpnstatus"
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
	// servers to list in login window
	getServers func() ([]string, error),
	// identity details to set button and icon state
	identityDetail *IdentityDetails) *VPNDetail {

	vd := &VPNDetail{detail: newDetail(), ctx: context, actionClicked: vpnActionClicked, identityDetail: identityDetail}

	// create login dialog
	vd.loginDialog = newLoginDialog(parent, getServers)

	// create action button with spinner, icons and labels
	vd.actionBtn = gtk.NewButtonWithLabel(i18n.L.Sprintf("Connect VPN"))
	vd.actionBtn.SetHAlign(gtk.AlignEnd)
	vd.actionBtn.ConnectClicked(vd.OnActionClicked)
	vd.actionSpinner = gtk.NewSpinner()
	vd.actionSpinner.SetHAlign(gtk.AlignEnd)
	vd.trustedNetworkLabel = gtk.NewLabel(i18n.L.Sprintf("not trusted"))
	vd.connectedImg = NewStatusIcon(false)
	vd.connectedAtLabel = gtk.NewLabel(util.DefaultValue)
	vd.ipLabel = gtk.NewLabel(util.DefaultValue)
	vd.deviceLabel = gtk.NewLabel(util.DefaultValue)
	vd.certExpiresLabel = gtk.NewLabel(util.DefaultValue)
	vd.applyTrustedNetwork(false)

	// set icons, labels and button with spinner in details box
	vd.
		buildBase(i18n.L.Sprintf("VPN Details")).
		addRow(i18n.L.Sprintf("Physical network"), vd.trustedNetworkLabel).
		addRow(i18n.L.Sprintf("Connected"), vd.actionSpinner, vd.actionBtn, vd.connectedImg).
		addRow(i18n.L.Sprintf("Connected at"), vd.connectedAtLabel).
		addRow(i18n.L.Sprintf("IP"), vd.ipLabel).
		addRow(i18n.L.Sprintf("Device"), vd.deviceLabel).
		addRow(i18n.L.Sprintf("Certificate expires"), vd.certExpiresLabel)

	go func() {
		configFile, err := os.ReadFile("/var/lib/oc-daemon/oc-client.json")
		if err != nil {
			logger.Logf("Warning: Can't read daemon config. %v", err)
			return
		}
		type c struct{ ClientCertificate string }
		var config c
		err = json.Unmarshal(configFile, &config)
		if err != nil {
			logger.Logf("Warning: Can't unmarshal daemon config: %v", err)
			return
		}

		// read cert
		out, err := exec.Command("p11tool", "--export", config.ClientCertificate).Output()
		if err != nil {
			logger.Logf("Warning: Can't read client certificate: %v", err)
			return
		}
		// decode
		der, _ := pem.Decode(out)
		// parse
		cert, err := x509.ParseCertificate(der.Bytes)
		if err != nil {
			logger.Logf("Warning: Can't parse client certificate: %v", err)
			return
		}
		// set value
		notAfter := cert.NotAfter.Unix()
		glib.IdleAdd(func() {
			vd.certExpiresLabel.SetText(util.FormatDate(notAfter))
		})
	}()

	vd.actionBtn.SetSensitive(false)
	vd.actionSpinner.Start()

	return vd
}

// applies new vpn status and calls afterApply after them
func (vd *VPNDetail) Apply(status *vpnstatus.Status, afterApply func(vpnConnected bool)) {
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
		connected := status.ConnectionState.Connected()
		vd.connectedImg.SetStatus(connected)
		vd.applyTrustedNetwork(status.TrustedNetwork.Trusted())
		vd.connectedAtLabel.SetText(util.FormatDate(status.ConnectedAt))
		vd.deviceLabel.SetText(util.FormatValue(status.Device))
		vd.ipLabel.SetText(util.FormatValue(status.IP))
		vd.certExpiresLabel.SetText(util.DefaultValue) // TODO
		vd.SetButtonsAfterProgress()
		afterApply(connected)
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
		err := vd.loginDialog.open(func(result *model.Credentials) {
			if result != nil {
				vd.triggerAction(result)
			}
		})
		if err != nil {
			// is handled in parent
			return
		}
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
		vd.connectedAtLabel.SetText(util.DefaultValue)
		vd.ipLabel.SetText(util.DefaultValue)
	} else {
		vd.trustedNetworkLabel.SetText(i18n.L.Sprintf("not trusted"))
		vd.actionBtn.SetSensitive(true)
		vd.connectedImg.SetOpacity(1)
		vd.connectedAtLabel.SetOpacity(1)
		vd.ipLabel.SetOpacity(1)
	}
}
