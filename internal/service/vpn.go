package service

import (
	"de.telekom-mms.corp-net-indicator/internal/model"
	"github.com/godbus/dbus/v5"
)

type VPNService struct {
	dbusService

	statusChan chan *model.VPNStatus
}

func NewVPNService() *VPNService {
	conn, err := dbus.ConnectSystemBus()
	if err != nil {
		panic(err)
	}
	return &VPNService{
		dbusService: dbusService{conn: conn, iface: "de.telekomMMS.vpn", path: "/de/telekomMMS/vpn"},
		statusChan:  make(chan *model.VPNStatus, 1),
	}
}

// attaches to the vpn DBUS status signal and delivers them by returned channel
func (v *VPNService) ListenToVPN() <-chan *model.VPNStatus {
	v.listen(func(result map[string]dbus.Variant) {
		select {
		case v.statusChan <- MapDbusDictToStruct(result, &model.VPNStatus{}):
		default:
		}
	})
	return v.statusChan
}

// triggers VPN connect
func (v *VPNService) Connect(password string, server string) error {
	return v.callMethod("Connect", "cookie", "host", "connectUrl", "fingerprint", "resolve").Store()
}

// triggers VPN disconnect
func (v *VPNService) Disconnect() error {
	return v.callMethod("Disconnect").Store()
}

// retrieves vpn status by DBUS
func (v *VPNService) GetStatus() (*model.VPNStatus, error) {
	status, err := v.getStatus()
	if err != nil {
		return nil, err
	}
	return MapDbusDictToStruct(status, &model.VPNStatus{}), nil
}

// retrieves server list by DBUS
func (v *VPNService) GetServerList() ([]string, error) {
	servers, err := v.getProp("Servers")
	if err != nil {
		return []string{}, err
	}
	val := servers.Value()
	if list, ok := val.([]string); ok {
		return list, nil
	}
	return []string{}, nil
}

// closes DBUS connection and signal channel
func (v *VPNService) Close() {
	v.conn.Close()
	close(v.statusChan)
}
