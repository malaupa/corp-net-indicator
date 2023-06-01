package service

import (
	"strings"
	"time"

	"context"

	"com.telekom-mms.corp-net-indicator/internal/logger"
	oc "github.com/T-Systems-MMS/oc-daemon/pkg/client"
	"github.com/T-Systems-MMS/oc-daemon/pkg/vpnstatus"
)

func VPNInProgress(state vpnstatus.ConnectionState) bool {
	return state == vpnstatus.ConnectionStateConnecting || state == vpnstatus.ConnectionStateDisconnecting
}

type VPNService struct {
	oc.Client
	statusChan chan *vpnstatus.Status
}

func NewVPNService() *VPNService {
	client, err := oc.NewClient(oc.LoadUserSystemConfig())
	if err != nil {
		panic(err)
	}
	return &VPNService{
		Client:     client,
		statusChan: make(chan *vpnstatus.Status, 10),
	}
}

// attaches to the vpn DBUS status signal and delivers them by returned channel
func (v *VPNService) SubscribeToVPN() (<-chan *vpnstatus.Status, func()) {
	logger.Verbose("Start listening to vpn status")
	ctx, close := context.WithCancel(context.Background())
	go func() {
		for {
			select {
			case <-ctx.Done():
				v.close()
				return
			default:
			}
			status, err := v.Query()
			if err != nil && !strings.Contains(err.Error(), ERR_SUFFIX) {
				logger.Logf("DBUS error: %v\n", err)
			}
			if status != nil {
				select {
				case v.statusChan <- status:
				case <-ctx.Done():
					v.close()
					return
				}
				break
			}
			logger.Verbosef("Wait %d seconds for service to come up...", DEBOUNCE)
			time.Sleep(time.Second * DEBOUNCE)
		}

		statusChan, err := v.Subscribe()
		if err != nil {
			panic(err)
		}

		for {
			select {
			case status := <-statusChan:
				v.statusChan <- status
			case <-ctx.Done():
				v.close()
				return
			}
		}
	}()
	return v.statusChan, close
}

// triggers VPN connect
func (v *VPNService) ConnectWithPasswordAndServer(password string, server string) error {
	config := v.GetConfig()
	config.Password = password
	config.VPNServer = server
	v.SetConfig(config)
	// v.SetLogin(&logininfo.LoginInfo{})

	err := v.Authenticate()
	if err != nil {
		return err
	}

	return v.Connect()
}

// Returns servers to connect to
func (v *VPNService) GetServers() ([]string, error) {
	result, err := v.Query()
	return result.Servers, err
}

// closes DBUS connection and signal channel
func (v *VPNService) close() {
	v.Close()
	close(v.statusChan)
}
