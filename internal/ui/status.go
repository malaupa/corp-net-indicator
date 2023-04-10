package ui

import (
	"context"

	"de.telekom-mms.corp-net-indicator/internal/model"
	"de.telekom-mms.corp-net-indicator/internal/ui/gtkui"
)

type StatusWindow interface {
	Open(ctx context.Context, iStatus *model.IdentityStatus, vStatus *model.VPNStatus, quickConnect bool)
	Close()
	ApplyIdentityStatus(ctx context.Context, status *model.IdentityStatus)
	ApplyVPNStatus(ctx context.Context, status *model.VPNStatus)
	NotifyError(err error)
}

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
		s.Close()
	}
	s.window = gtkui.NewStatusWindow(s.ConnectDisconnectClicked, s.ReLoginClicked)
	go func() {
		s.closeChan = make(chan struct{})
		s.window.Open(ctx, iStatus, vStatus, quickConnect)
		s.window = nil
		close(s.closeChan)
	}()
}

func (s *Status) Close() {
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
