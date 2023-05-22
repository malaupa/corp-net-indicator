package service

import (
	"com.telekom-mms.corp-net-indicator/internal/logger"
	"com.telekom-mms.corp-net-indicator/internal/model"
	oc "github.com/T-Systems-MMS/oc-daemon/pkg/client"
	"github.com/godbus/dbus/v5"
)

const (
	VPN_IFACE = "com.telekom_mms.oc_daemon.Daemon"
	VPN_PATH  = "/com/telekom_mms/oc_daemon/Daemon"
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
		dbusService: dbusService{conn: conn, iface: VPN_IFACE, path: VPN_PATH},
		statusChan:  make(chan *model.VPNStatus, 10),
	}
}

// attaches to the vpn DBUS status signal and delivers them by returned channel
func (v *VPNService) ListenToVPN() <-chan *model.VPNStatus {
	logger.Verbose("Listening to vpn status")
	v.listen(func(result map[string]dbus.Variant) {
		v.statusChan <- MapDbusDictToStruct(result, &model.VPNStatus{})
	})
	return v.statusChan
}

// triggers VPN connect
func (v *VPNService) Connect(password string, server string) error {
	config := oc.LoadUserSystemConfig()
	client := oc.NewClient(config)
	client.Config.Password = password
	client.Config.VPNServer = server

	err := client.Authenticate()
	if err != nil {
		return err
	}

	return v.callMethod("Connect",
		client.Login.Cookie,
		client.Login.Host,
		client.Login.ConnectURL,
		client.Login.Fingerprint,
		client.Login.Resolve).Store()
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
