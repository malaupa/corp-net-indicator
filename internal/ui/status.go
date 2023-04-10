package ui

import (
	"context"

	"de.telekom-mms.corp-net-indicator/internal/model"
)

type Status struct {
	ConnectDisconnectClicked chan *model.Credentials
	ReLoginClicked           chan bool

	window    *statusWindow
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
	s.window = newStatusWindow(s.ConnectDisconnectClicked, s.ReLoginClicked)
	go func() {
		s.closeChan = make(chan struct{})
		s.window.open(ctx, iStatus, vStatus, quickConnect)
		s.window = nil
		close(s.closeChan)
	}()
}

func (s *Status) Close() {
	if s.window != nil {
		s.window.close()
		<-s.closeChan
	}
}

func (s *Status) NotifyError(err error) {
	if s.window != nil {
		go s.window.notifyError(err)
	}
}

func (s *Status) ApplyVPNStatus(ctx context.Context, status *model.VPNStatus) {
	if s.window != nil {
		go s.window.applyVPNStatus(ctx, status)
	}
}

func (s *Status) ApplyIdentityStatus(ctx context.Context, status *model.IdentityStatus) {
	if s.window != nil {
		go s.window.applyIdentityStatus(ctx, status)
	}
}
