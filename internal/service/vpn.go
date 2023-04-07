package service

import (
	"de.telekom-mms.corp-net-indicator/internal/model"
	"github.com/godbus/dbus/v5"
)

type VPN struct {
	conn       *dbus.Conn
	statusChan chan *model.VPNStatus
	stop       chan bool
}

func NewVPNService() *VPN {
	conn, err := dbus.ConnectSystemBus()
	if err != nil {
		panic(err)
	}
	return &VPN{conn: conn, statusChan: make(chan *model.VPNStatus), stop: make(chan bool)}
}

func (v *VPN) ListenToVPN() <-chan *model.VPNStatus {
	// TODO
	return v.statusChan
}

func (v *VPN) Connect(password string, server string) error {
	return nil
}

func (v *VPN) Disconnect() {

}

func (v *VPN) GetStatus() *model.VPNStatus {
	return &model.VPNStatus{}
}

func (v *VPN) GetStatusChan() <-chan *model.VPNStatus {
	return v.statusChan
}

func (v *VPN) Close() {
	v.stop <- true
	v.conn.Close()
}
