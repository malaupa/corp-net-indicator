package service

//go:generate dbus-codegen-go -client-only -prefix de.telekomMMS -package vpn -camelize -output ../generated/vpn/client/client.go ../schema/vpn.xml

import (
	"context"

	vpn "de.telekom-mms.corp-net-indicator/internal/generated/vpn/server"
	"de.telekom-mms.corp-net-indicator/internal/logger"
	"de.telekom-mms.corp-net-indicator/internal/model"
	"github.com/godbus/dbus/v5"
)

const V_DBUS_SERVICE_NAME = "de.telekomMMS.vpn"
const V_DBUS_OBJECT_PATH = "/de/telekomMMS/vpn"

type VPNService struct {
	conn       *dbus.Conn
	statusChan chan *model.VPNStatus
}

func NewVPNService() *VPNService {
	conn, err := dbus.ConnectSystemBus()
	if err != nil {
		panic(err)
	}
	return &VPNService{conn: conn, statusChan: make(chan *model.VPNStatus, 1)}
}

func (v *VPNService) ListenToVPN() <-chan *model.VPNStatus {
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
				logger.Logf("DBUS err: %v\n", err)
				panic(err)
			}

			if typed, ok := s.(*vpn.VpnStatusChangeSignal); ok {
				select {
				case v.statusChan <- MapDbusDictToStruct(typed.Body.Status, &model.VPNStatus{}):
				default:
				}
			}
		}
	}()
	return v.statusChan
}

func (v *VPNService) Connect(password string, server string) error {
	obj := vpn.NewVpn(v.conn.Object(V_DBUS_SERVICE_NAME, V_DBUS_OBJECT_PATH))
	return obj.Connect(context.Background(), password, server)
}

func (v *VPNService) Disconnect() error {
	obj := vpn.NewVpn(v.conn.Object(V_DBUS_SERVICE_NAME, V_DBUS_OBJECT_PATH))
	return obj.Disconnect(context.Background())
}

func (v *VPNService) GetStatus() (*model.VPNStatus, error) {
	obj := vpn.NewVpn(v.conn.Object(V_DBUS_SERVICE_NAME, V_DBUS_OBJECT_PATH))
	status, err := obj.GetStatus(context.Background())
	if err != nil {
		return nil, err
	}
	return MapDbusDictToStruct(status, &model.VPNStatus{}), nil
}

func (v *VPNService) GetServerList() ([]string, error) {
	obj := vpn.NewVpn(v.conn.Object(V_DBUS_SERVICE_NAME, V_DBUS_OBJECT_PATH))
	servers, err := obj.ListServers(context.Background())
	if err != nil {
		return []string{}, err
	}

	return servers, nil
}

func (v *VPNService) Close() {
	v.conn.Close()
	close(v.statusChan)
}
