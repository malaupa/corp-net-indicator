package ui

import (
	"de.telekom-mms.corp-net-indicator/internal/model"
	"de.telekom-mms.corp-net-indicator/internal/service"
	"de.telekom-mms.corp-net-indicator/internal/ui/gtkui"
)

// minimal interface to interact with an ui implementation
type StatusWindow interface {
	Open(iStatus *model.IdentityStatus, vStatus *model.VPNStatus, quickConnect bool)
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
					if err := vSer.Connect(c.Password, c.Server); err != nil {
						go s.window.NotifyError(err)
					}
				} else {
					vSer.Disconnect()
					// TODO handle error
				}
			case <-s.reLoginClicked:
				s.ctx.Write(func(ctx *model.ContextValues) {
					ctx.IdentityInProgress = true
				})
				iSer.ReLogin()
			// TODO handle error
			case status := <-iChan:
				s.ctx.Write(func(ctx *model.ContextValues) {
					ctx.IdentityInProgress = status.InProgress
					ctx.LoggedIn = status.LoggedIn
				})
				go s.window.ApplyIdentityStatus(status)
			case status := <-vChan:
				s.ctx.Write(func(ctx *model.ContextValues) {
					ctx.VPNInProgress = status.InProgress
					ctx.Connected = status.Connected
					ctx.TrustedNetwork = status.TrustedNetwork
				})
				go s.window.ApplyVPNStatus(status)
			case <-c:
				vSer.Close()
				iSer.Close()
				return
			}
		}
	}()

	// get actual status
	vStatus := vSer.GetStatus()
	s.ctx.Write(func(ctx *model.ContextValues) {
		ctx.VPNInProgress = vStatus.InProgress
		ctx.Connected = vStatus.Connected
		ctx.TrustedNetwork = vStatus.TrustedNetwork
	})
	iStatus := iSer.GetStatus()
	s.ctx.Write(func(ctx *model.ContextValues) {
		ctx.IdentityInProgress = iStatus.InProgress
		ctx.LoggedIn = iStatus.LoggedIn
	})
	// open window
	s.window.Open(iStatus, vStatus, quickConnect)
	close(c)
}
