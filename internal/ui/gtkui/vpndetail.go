package gtkui

import (
	"context"
	"log"

	"de.telekom-mms.corp-net-indicator/internal/i18n"
	"de.telekom-mms.corp-net-indicator/internal/model"
	"de.telekom-mms.corp-net-indicator/internal/util"
	"github.com/diamondburned/gotk4/pkg/core/glib"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type vpnDetail struct {
	ctx              context.Context
	vpnActionClicked chan *model.Credentials

	root gtk.Widgetter

	trustedNetworkImg *gtk.Image
	connectedImg      *gtk.Image
	actionSpinner     *gtk.Spinner
	actionBtn         *gtk.Button
	connectedAtLabel  *gtk.Label
	ipLabel           *gtk.Label
	deviceLabel       *gtk.Label
	certExpiresLabel  *gtk.Label
	loginDialog       *LoginDialog
}

func new(ctx context.Context, vpnActionClicked chan *model.Credentials, vpn *model.VPNStatus) *vpnDetail {
	vd := &vpnDetail{ctx: ctx, vpnActionClicked: vpnActionClicked}
	l := i18n.Localizer()
	box, list := buildListBoxBase("VPN Details")
	vd.root = box
	vd.actionBtn = gtk.NewButtonWithLabel(l.Sprintf("Connect VPN"))
	vd.actionBtn.SetHAlign(gtk.AlignEnd)
	if vpn.Connected {
		vd.actionBtn.SetLabel(l.Sprintf("Disconnect VPN"))
	}
	if vpn.TrustedNetwork {
		vd.actionBtn.SetSensitive(false)
	}
	vd.actionBtn.ConnectClicked(vd.actionClicked)
	vd.actionSpinner = gtk.NewSpinner()
	vd.actionSpinner.SetHAlign(gtk.AlignEnd)
	vd.trustedNetworkImg = buildStatusIcon(vpn.TrustedNetwork)
	list.Append(addRow(l.Sprintf("Trusted Network"), vd.trustedNetworkImg))
	vd.connectedImg = buildStatusIcon(vpn.Connected)
	list.Append(addRow(l.Sprintf("Connected"), vd.actionSpinner, vd.actionBtn, vd.connectedImg))
	vd.connectedAtLabel = gtk.NewLabel(util.FormatDate(vpn.ConnectedAt))
	list.Append(addRow(l.Sprintf("Connect at"), vd.connectedAtLabel))
	vd.ipLabel = gtk.NewLabel(vpn.IP)
	list.Append(addRow(l.Sprintf("IP"), vd.ipLabel))
	vd.deviceLabel = gtk.NewLabel(vpn.Device)
	list.Append(addRow(l.Sprintf("Device"), vd.deviceLabel))
	vd.certExpiresLabel = gtk.NewLabel(util.FormatDate(vpn.CertExpiresAt))
	list.Append(addRow(l.Sprintf("Certificate expires"), vd.certExpiresLabel))
	// set correct identity status
	// if !vpn.Connected && !vpn.TrustedNetwork {
	// 	setStatusIcon(vd.loggedInImg, false)
	// 	vd.reLoginBtn.SetSensitive(false)
	// }
	return vd
}

func (vd *vpnDetail) applyVPNStatus(ctx context.Context, status *model.VPNStatus, afterApply func()) {
	vd.ctx = ctx
	l := i18n.Localizer()
	glib.IdleAdd(func() {
		vd.actionSpinner.Stop()
		setStatusIcon(vd.trustedNetworkImg, status.TrustedNetwork)
		setStatusIcon(vd.connectedImg, status.Connected)
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
		// if !status.Connected && !status.TrustedNetwork {
		// 	setStatusIcon(vd.loggedInImg, false)
		// 	vd.reLoginBtn.SetSensitive(false)
		// } else {
		// 	vd.reLoginBtn.SetSensitive(true)
		// }
		afterApply()
		// if vd.quickConnect {
		// 	vd.Close()
		// }
	})
}

func (sw *vpnDetail) actionClicked() {
	if sw.ctx.Value(model.Connected).(bool) {
		sw.actionSpinner.Start()
		sw.actionBtn.SetSensitive(false)
		// sw.reLoginBtn.SetSensitive(false)
		go func() {
			log.Println("start 1")
			sw.vpnActionClicked <- nil
			log.Println("end 1")
		}()
	} else {
		if sw.loginDialog.IsOpen() {
			return
		}
		resultChan := sw.loginDialog.Open()
		go func() {
			result := <-resultChan
			if result != nil {
				log.Println(result)
				glib.IdleAdd(func() {
					sw.actionSpinner.Start()
					sw.actionBtn.SetSensitive(false)
					// sw.reLoginBtn.SetSensitive(false)
				})
				sw.vpnActionClicked <- result
			}
		}()
	}
}
