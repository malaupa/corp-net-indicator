package ui

import (
	"context"

	"de.telekom-mms.corp-net-indicator/internal/model"
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
	ConnectDisconnectClicked chan *model.Credentials
	ReLoginClicked           chan bool

	window    StatusWindow
	closeChan chan struct{}
}

func NewStatus() *Status {
	return &Status{
		ConnectDisconnectClicked: make(chan *model.Credentials),
		ReLoginClicked:           make(chan bool),
	}
}

func (s *Status) OpenWindow(ctx context.Context, iStatus *model.IdentityStatus, vStatus *model.VPNStatus, quickConnect bool) {
	if s.window != nil {
		s.CloseWindow()
	}
	s.window = gtkui.NewStatusWindow(s.ConnectDisconnectClicked, s.ReLoginClicked)
	go func() {
		s.closeChan = make(chan struct{})
		s.window.Open(ctx, iStatus, vStatus, quickConnect)
		s.window = nil
		close(s.closeChan)
	}()
}

func (s *Status) CloseWindow() {
	if s.window != nil {
		s.window.Close()
		<-s.closeChan
	}
}

func (s *Status) NotifyError(err error) {
	if s.window != nil {
		go s.window.NotifyError(err)
	}
}

func (s *Status) ApplyVPNStatus(ctx context.Context, status *model.VPNStatus) {
	if s.window != nil {
		go s.window.ApplyVPNStatus(ctx, status)
	}
}

func (s *Status) ApplyIdentityStatus(ctx context.Context, status *model.IdentityStatus) {
	if s.window != nil {
		go s.window.ApplyIdentityStatus(ctx, status)
	}
}
