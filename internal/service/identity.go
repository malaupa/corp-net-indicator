package service

//go:generate dbus-codegen-go -client-only -prefix de.telekomMMS -package identity -camelize -output ../generated/identity/client/client.go ../schema/identity.xml

import (
	"de.telekom-mms.corp-net-indicator/internal/logger"
	"de.telekom-mms.corp-net-indicator/internal/model"
	"github.com/godbus/dbus/v5"
)

type IdentityService struct {
	dbusService

	statusChan chan *model.IdentityStatus
}

func NewIdentityService() *IdentityService {
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		panic(err)
	}
	return &IdentityService{
		dbusService: dbusService{conn: conn, iface: "de.telekomMMS.identity", path: "/de/telekomMMS/identity"},
		statusChan:  make(chan *model.IdentityStatus, 1),
	}
}

// attaches to DBUS properties changed signal, maps to status and delivers them by returned channel
func (i *IdentityService) ListenToIdentity() <-chan *model.IdentityStatus {
	logger.Verbose("Listening to identity status")
	i.listen(func(sig map[string]dbus.Variant) {
		select {
		case i.statusChan <- MapDbusDictToStruct(sig, &model.IdentityStatus{}):
		default:
		}
	})
	return i.statusChan
}

// retrieves identity status by DBUS
func (i *IdentityService) GetStatus() (*model.IdentityStatus, error) {
	logger.Verbose("Call GetStatus")
	status, err := i.getStatus()
	if err != nil {
		return nil, err
	}
	return MapDbusDictToStruct(status, &model.IdentityStatus{}), nil
}

// triggers identity agent login
func (i *IdentityService) ReLogin() error {
	logger.Verbose("Call ReLogin")
	return i.callMethod("ReLogin").Store()
}

// closes DBUS connection and signal channel
func (i *IdentityService) Close() {
	i.conn.Close()
	close(i.statusChan)
	logger.Verbose("Service closed")
}
