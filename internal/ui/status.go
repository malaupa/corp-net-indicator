package ui

import (
	"os"

	"de.telekom-mms.corp-net-indicator/internal/logger"
	"de.telekom-mms.corp-net-indicator/internal/model"
	"de.telekom-mms.corp-net-indicator/internal/service"
	"de.telekom-mms.corp-net-indicator/internal/ui/gtkui"
)

// minimal interface to interact with an ui implementation
type StatusWindow interface {
	Open(iStatus *model.IdentityStatus, vStatus *model.VPNStatus, servers []string, quickConnect bool)
	Close()
	ApplyIdentityStatus(status *model.IdentityStatus)
	ApplyVPNStatus(status *model.VPNStatus)
	NotifyError(err error)
}

// holds data channels for updates and a window handle
// is used to free memory after closing window
type Status struct {
	ctx *model.Context

	connectDisconnectClicked chan *model.Credentials
	reLoginClicked           chan bool

	window StatusWindow
}

func NewStatus() *Status {
	s := &Status{
		ctx:                      model.NewContext(),
		connectDisconnectClicked: make(chan *model.Credentials),
		reLoginClicked:           make(chan bool),
	}
	s.window = gtkui.NewStatusWindow(s.ctx, s.connectDisconnectClicked, s.reLoginClicked)
	return s
}

func (s *Status) Run(quickConnect bool) {
	// create services
	vSer := service.NewVPNService()
	iSer := service.NewIdentityService()

	// get actual status
	vStatus, err := vSer.GetStatus()
	if err != nil {
		logger.Logf("DBUS error: %v\n", err)
		os.Exit(1)
	}
	iStatus, err := iSer.GetStatus()
	if err != nil {
		logger.Logf("DBUS error: %v\n", err)
		os.Exit(1)
	}
	s.ctx.Write(func(ctx *model.ContextValues) {
		ctx.VPNInProgress = vStatus.InProgress
		ctx.Connected = vStatus.Connected
		ctx.TrustedNetwork = vStatus.TrustedNetwork
		ctx.IdentityInProgress = iStatus.InProgress
		ctx.LoggedIn = iStatus.LoggedIn
	})
	servers, err := vSer.GetServerList()
	if err != nil {
		logger.Logf("DBUS error: %v\n", err)
		os.Exit(1)
	}

	// listen to status changes
	vChan := vSer.ListenToVPN()
	iChan := iSer.ListenToIdentity()

	c := make(chan struct{})
	go func() {
		for {
			select {
			// handle window clicks
			case c := <-s.connectDisconnectClicked:
				s.ctx.Write(func(ctx *model.ContextValues) {
					ctx.VPNInProgress = true
				})
				if c != nil {
					logger.Verbose("Open dialog to connect to VPN")

					if err := vSer.Connect(c.Password, c.Server); err != nil {
						s.handleDBUSError(err)
					}
				} else {
					logger.Verbose("Tray to disconnect")

					if err := vSer.Disconnect(); err != nil {
						s.handleDBUSError(err)
					}
				}
			case <-s.reLoginClicked:
				logger.Verbose("Try to login to identity service")

				s.ctx.Write(func(ctx *model.ContextValues) {
					ctx.IdentityInProgress = true
				})
				if err := iSer.ReLogin(); err != nil {
					s.handleDBUSError(err)
				}
			case status := <-iChan:
				logger.Verbosef("Apply identity status: %+v\n", status)

				s.ctx.Write(func(ctx *model.ContextValues) {
					ctx.IdentityInProgress = status.InProgress
					ctx.LoggedIn = status.LoggedIn
				})
				go s.window.ApplyIdentityStatus(status)
			case status := <-vChan:
				logger.Verbosef("Apply vpn status: %+v\n", status)

				s.ctx.Write(func(ctx *model.ContextValues) {
					ctx.VPNInProgress = status.InProgress
					ctx.Connected = status.Connected
					ctx.TrustedNetwork = status.TrustedNetwork
				})
				go s.window.ApplyVPNStatus(status)
			case <-c:
				logger.Verbose("Received SIGINT -> closing")

				vSer.Close()
				iSer.Close()
				return
			}
		}
	}()

	// open window
	s.window.Open(iStatus, vStatus, servers, quickConnect)
	logger.Verbose("Window closed")

	close(c)
}

func (s *Status) handleDBUSError(err error) {
	logger.Logf("DBUS Error: %v\n", err)

	s.ctx.Write(func(ctx *model.ContextValues) {
		ctx.VPNInProgress = false
		ctx.IdentityInProgress = false
	})
	go s.window.NotifyError(err)
}
