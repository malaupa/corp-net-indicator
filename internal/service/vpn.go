package service

import (
	"context"
	"log"

	vpn "de.telekom-mms.corp-net-indicator/internal/generated/vpn/server"
	"de.telekom-mms.corp-net-indicator/internal/model"
	"github.com/godbus/dbus/v5"
)

const V_DBUS_SERVICE_NAME = "de.telekomMMS.vpn"
const V_DBUS_OBJECT_PATH = "/de/telekomMMS/vpn"

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
	go func() {
		var sigI *vpn.VpnStatusChangeSignal = nil
		vpn.AddMatchSignal(v.conn, sigI)
		defer vpn.RemoveMatchSignal(v.conn, sigI)

		c := make(chan *dbus.Signal, 1)
		v.conn.Signal(c)

		for sig := range c {
			s, err := vpn.LookupSignal(sig)
			if err != nil {
				if err == vpn.ErrUnknownSignal {
					continue
				}
				panic(err)
			}

			if typed, ok := s.(*vpn.VpnStatusChangeSignal); ok {
				select {
				case <-v.stop:
					close(v.statusChan)
					return
				case v.statusChan <- MapDbusDictToStruct(typed.Body.Status, &model.VPNStatus{}):
				default:
				}
			}
		}
	}()
	return v.statusChan
}

func (v *VPN) Connect(password string, server string) error {
	obj := vpn.NewVpn(v.conn.Object(V_DBUS_SERVICE_NAME, V_DBUS_OBJECT_PATH))
	err := obj.Connect(context.Background(), password, server)
	if err != nil {
		// TODO enhance logging
		log.Println(err)
		return err
	}
	return nil
}

func (v *VPN) Disconnect() {
	obj := vpn.NewVpn(v.conn.Object(V_DBUS_SERVICE_NAME, V_DBUS_OBJECT_PATH))
	err := obj.Disconnect(context.Background())
	if err != nil {
		// TODO enhance logging
		log.Println(err)
	}
}

func (v *VPN) GetStatus() *model.VPNStatus {
	obj := vpn.NewVpn(v.conn.Object(V_DBUS_SERVICE_NAME, V_DBUS_OBJECT_PATH))
	status, err := obj.GetStatus(context.Background())
	if err != nil {
		// TODO enhance logging
		log.Println(err)
	}

	return MapDbusDictToStruct(status, &model.VPNStatus{})
}

func (v *VPN) Close() {
	v.stop <- true
	v.conn.Close()
}
