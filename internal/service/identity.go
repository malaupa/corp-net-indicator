package service

import (
	"context"
	"log"

	identity "de.telekom-mms.corp-net-indicator/internal/generated/identity/client"
	"de.telekom-mms.corp-net-indicator/internal/model"
	"github.com/godbus/dbus/v5"
)

const I_DBUS_SERVICE_NAME = "de.telekomMMS.identity"
const I_DBUS_OBJECT_PATH = "/de/telekomMMS/identity"

type Identity struct {
	conn       *dbus.Conn
	statusChan chan *model.IdentityStatus
	stop       chan bool
}

func NewIdentityService() *Identity {
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		panic(err)
	}
	return &Identity{conn: conn, statusChan: make(chan *model.IdentityStatus), stop: make(chan bool)}
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
				case <-i.stop:
					close(i.statusChan)
					return
				case i.statusChan <- MapDbusDictToStruct(typed.Body.Status, &model.IdentityStatus{}):
				default:
				}
			}
		}
	}()
	return i.statusChan
}

func (i *Identity) GetStatus() *model.IdentityStatus {
	obj := identity.NewIdentity(i.conn.Object(I_DBUS_SERVICE_NAME, I_DBUS_OBJECT_PATH))
	status, err := obj.GetStatus(context.Background())
	if err != nil {
		// TODO enhance logging
		log.Println(err)
	}

	return MapDbusDictToStruct(status, &model.IdentityStatus{})
}

func (i *Identity) ReLogin() {
	obj := identity.NewIdentity(i.conn.Object(I_DBUS_SERVICE_NAME, I_DBUS_OBJECT_PATH))
	err := obj.ReLogin(context.Background())
	if err != nil {
		// TODO enhance logging
		log.Println(err)
	}
}

func (i *Identity) Close() {
	i.stop <- true
	i.conn.Close()
}
