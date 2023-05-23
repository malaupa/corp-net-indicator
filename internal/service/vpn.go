package service

import (
	"com.telekom-mms.corp-net-indicator/internal/logger"
	"com.telekom-mms.corp-net-indicator/internal/model"
	oc "github.com/T-Systems-MMS/oc-daemon/pkg/client"
	"github.com/T-Systems-MMS/oc-daemon/pkg/logininfo"
	"github.com/godbus/dbus/v5"
)

const (
	VPN_IFACE = "com.telekom_mms.oc_daemon.Daemon"
	VPN_PATH  = "/com/telekom_mms/oc_daemon/Daemon"
)

// allows to mock oc client api
type client interface {
	Authenticate() error
	SetConfig(password, server string)
	GetLoginInfo() *logininfo.LoginInfo
}

// wraps oc client
type ocClient struct {
	oc.Client
}

// sets necessary config attributes
func (c *ocClient) SetConfig(password, server string) {
	c.Config.Password = password
	c.Config.VPNServer = server
}

// returns login info object
func (c *ocClient) GetLoginInfo() *logininfo.LoginInfo {
	return c.Login
}

type VPNService struct {
	dbusService

	statusChan chan *model.VPNStatus
	ocClient   client
}

func NewVPNService() *VPNService {
	conn, err := dbus.ConnectSystemBus()
	if err != nil {
		panic(err)
	}
	return &VPNService{
		dbusService: dbusService{conn: conn, iface: VPN_IFACE, path: VPN_PATH},
		statusChan:  make(chan *model.VPNStatus, 10),
		ocClient:    &ocClient{Client: *oc.NewClient(oc.LoadUserSystemConfig())},
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
	v.ocClient.SetConfig(password, server)

	err := v.ocClient.Authenticate()
	if err != nil {
		return err
	}
	info := v.ocClient.GetLoginInfo()

	return v.callMethod("Connect", info.Cookie, info.Host, info.ConnectURL, info.Fingerprint, info.Resolve).Store()
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
