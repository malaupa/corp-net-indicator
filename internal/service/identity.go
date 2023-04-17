package service

import (
	"context"

	identity "de.telekom-mms.corp-net-indicator/internal/generated/identity/client"
	"de.telekom-mms.corp-net-indicator/internal/model"
	"github.com/godbus/dbus/v5"
)

const I_DBUS_SERVICE_NAME = "de.telekomMMS.identity"
const I_DBUS_OBJECT_PATH = "/de/telekomMMS/identity"

type Identity struct {
	conn       *dbus.Conn
	statusChan chan *model.IdentityStatus
}

func NewIdentityService() *Identity {
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		panic(err)
	}
	return &Identity{conn: conn, statusChan: make(chan *model.IdentityStatus, 1)}
}

func (i *Identity) ListenToIdentity() <-chan *model.IdentityStatus {
	go func() {
		var sigI *identity.IdentityStatusChangeSignal = nil
		identity.AddMatchSignal(i.conn, sigI)
		defer identity.RemoveMatchSignal(i.conn, sigI)

		c := make(chan *dbus.Signal, 1)
		i.conn.Signal(c)

		for sig := range c {
			s, err := identity.LookupSignal(sig)
			if err != nil {
				if err == identity.ErrUnknownSignal {
					continue
				}
				panic(err)
			}

			if typed, ok := s.(*identity.IdentityStatusChangeSignal); ok {
				select {
				case i.statusChan <- MapDbusDictToStruct(typed.Body.Status, &model.IdentityStatus{}):
				default:
				}
			}
		}
	}()
	return i.statusChan
}

func (i *Identity) GetStatus() (*model.IdentityStatus, error) {
	obj := identity.NewIdentity(i.conn.Object(I_DBUS_SERVICE_NAME, I_DBUS_OBJECT_PATH))
	status, err := obj.GetStatus(context.Background())
	if err != nil {
		return nil, err
	}

	return MapDbusDictToStruct(status, &model.IdentityStatus{}), nil
}

func (i *Identity) ReLogin() error {
	obj := identity.NewIdentity(i.conn.Object(I_DBUS_SERVICE_NAME, I_DBUS_OBJECT_PATH))
	return obj.ReLogin(context.Background())
}

func (i *Identity) Close() {
	i.conn.Close()
}
