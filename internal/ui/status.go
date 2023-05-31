package ui

import (
	"os"
	"os/signal"

	"com.telekom-mms.corp-net-indicator/internal/logger"
	"com.telekom-mms.corp-net-indicator/internal/model"
	"com.telekom-mms.corp-net-indicator/internal/service"
	"com.telekom-mms.corp-net-indicator/internal/ui/gtkui"
)

// minimal interface to interact with an ui implementation
type StatusWindow interface {
	Open(quickConnect bool, getServers func() ([]string, error), onReady func())
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

	// listen to status changes
	vChan := vSer.ListenToVPN()
	iChan := iSer.ListenToIdentity()

	// catch interrupt and clean up
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	s.window.Open(quickConnect, vSer.GetServerList, func() {
		for {
			select {
			// handle window clicks
			case connect := <-s.connectDisconnectClicked:
				s.ctx.Write(func(ctx *model.ContextValues) {
					ctx.VPNInProgress = true
				})
				if connect != nil {
					logger.Verbose("Open dialog to connect to VPN")

					if err := vSer.Connect(connect.Password, connect.Server); err != nil {
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
					ctx.IdentityInProgress = status.InProgress(ctx.IdentityInProgress)
					ctx.LoggedIn = status.IsLoggedIn(ctx.LoggedIn)
				})
				go s.window.ApplyIdentityStatus(status)
			case status := <-vChan:
				logger.Verbosef("Apply vpn status: %+v\n", status)

				s.ctx.Write(func(ctx *model.ContextValues) {
					ctx.VPNInProgress = status.InProgress(ctx.VPNInProgress)
					ctx.Connected = status.IsConnected(ctx.Connected)
					ctx.TrustedNetwork = status.IsTrustedNetwork(ctx.TrustedNetwork)
				})
				go s.window.ApplyVPNStatus(status)
			case <-c:
				logger.Verbose("Received SIGINT -> closing")
				s.window.Close()
				return
			}
		}
	})

	logger.Verbose("Window closed")
	close(c)

	vSer.Close()
	iSer.Close()
}

func (s *Status) handleDBUSError(err error) {
	logger.Logf("DBUS Error: %v\n", err)

	s.ctx.Write(func(ctx *model.ContextValues) {
		ctx.VPNInProgress = false
		ctx.IdentityInProgress = false
	})
	go s.window.NotifyError(err)
}
