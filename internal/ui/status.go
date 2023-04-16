package ui

import (
	"context"

	"de.telekom-mms.corp-net-indicator/internal/model"
	"de.telekom-mms.corp-net-indicator/internal/service"
	"de.telekom-mms.corp-net-indicator/internal/ui/gtkui"
)

// minimal interface to interact with an ui implementation
type StatusWindow interface {
	Open(ctx context.Context, iStatus *model.IdentityStatus, vStatus *model.VPNStatus, quickConnect bool)
	Close()
	ApplyIdentityStatus(ctx context.Context, status *model.IdentityStatus)
	ApplyVPNStatus(ctx context.Context, status *model.VPNStatus)
	NotifyError(err error)
}

// holds data channels for updates and a window handle
// is used to free memory after closing window
type Status struct {
	ctx context.Context

	connectDisconnectClicked chan *model.Credentials
	reLoginClicked           chan bool

	window StatusWindow
}

func NewStatus(ctx context.Context) *Status {
	s := &Status{
		ctx:                      ctx,
		connectDisconnectClicked: make(chan *model.Credentials),
		reLoginClicked:           make(chan bool),
	}
	s.window = gtkui.NewStatusWindow(s.connectDisconnectClicked, s.reLoginClicked)
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
				s.ctx = context.WithValue(s.ctx, model.VPNInProgress, true)
				if c != nil {
					if err := vSer.Connect(c.Password, c.Server); err != nil {
						go s.window.NotifyError(err)
					}
				} else {
					vSer.Disconnect()
					// TODO handle error
				}
			case <-s.reLoginClicked:
				s.ctx = context.WithValue(s.ctx, model.IdentityInProgress, true)
				iSer.ReLogin()
			// TODO handle error
			case status := <-iChan:
				s.ctx = context.WithValue(s.ctx, model.IdentityInProgress, status.InProgress)
				go s.window.ApplyIdentityStatus(s.ctx, status)
			case status := <-vChan:
				s.ctx = context.WithValue(s.ctx, model.VPNInProgress, status.InProgress)
				s.ctx = context.WithValue(s.ctx, model.Connected, status.Connected)
				go s.window.ApplyVPNStatus(s.ctx, status)
			case <-c:
				vSer.Close()
				iSer.Close()
				return
			}
		}
	}()

	// get actual status
	vStatus := vSer.GetStatus()
	s.ctx = context.WithValue(s.ctx, model.VPNInProgress, vStatus.InProgress)
	iStatus := iSer.GetStatus()
	s.ctx = context.WithValue(s.ctx, model.IdentityInProgress, iStatus.InProgress)
	// open window
	s.window.Open(s.ctx, iStatus, vStatus, quickConnect)
	close(c)
}
