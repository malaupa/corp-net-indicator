package testserver

import (
	_ "embed"
	"log"
	"time"

	"de.telekom-mms.corp-net-indicator/internal/model"
	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/introspect"
	"github.com/godbus/dbus/v5/prop"
)

const I_DBUS_SERVICE_NAME = "de.telekomMMS.identity"
const I_DBUS_OBJECT_PATH = "/de/telekomMMS/identity"

type identityAgent struct {
	agent
}

func (i identityAgent) ReLogin() *dbus.Error {
	log.Println("Identity: ReLogin called!")
	i.props.SetMust(I_DBUS_SERVICE_NAME, "LoginState", model.LoggingIn)
	go func() {
		if i.simulate {
			time.Sleep(time.Second * 3)
		}
		var now int64 = 60 * 60
		if i.simulate {
			now = time.Now().Unix()
		}
		i.props.SetMustMany(I_DBUS_SERVICE_NAME, map[string]interface{}{
			"LoginState":      model.LoggedIn,
			"LastKeepAliveAt": now,
		})
	}()
	return nil
}

var iA *identityAgent

func NewIdentityServer(simulate bool) *dbus.Conn {
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		panic(err)
	}
	a := identityAgent{agent: agent{simulate: simulate}}
	var now int64 = 60 * 60
	if a.simulate {
		now = time.Now().Unix()
	}
	iA = &a

	// identity properties
	a.props, err = prop.Export(conn, I_DBUS_OBJECT_PATH, prop.Map{
		I_DBUS_SERVICE_NAME: {
			"TrustedNetwork":   {Value: model.LoginUnknown, Writable: false, Emit: prop.EmitTrue, Callback: nil},
			"LoginState":       {Value: model.LoginUnknown, Writable: false, Emit: prop.EmitTrue, Callback: nil},
			"LastKeepAliveAt":  {Value: now, Writable: false, Emit: prop.EmitTrue, Callback: nil},
			"KerberosIssuedAt": {Value: now - 60*60, Writable: false, Emit: prop.EmitTrue, Callback: nil},
		},
	})
	if err != nil {
		panic(err)
	}
	// identity methods
	err = conn.Export(a, I_DBUS_OBJECT_PATH, I_DBUS_SERVICE_NAME)
	if err != nil {
		panic(err)
	}
	// identity introspection
	n := &introspect.Node{
		Name: I_DBUS_OBJECT_PATH,
		Interfaces: []introspect.Interface{
			introspect.IntrospectData,
			prop.IntrospectData,
			{
				Name:       I_DBUS_SERVICE_NAME,
				Methods:    introspect.Methods(a),
				Properties: a.props.Introspection(I_DBUS_SERVICE_NAME),
			},
		},
	}
	err = conn.Export(introspect.NewIntrospectable(n), I_DBUS_OBJECT_PATH, "org.freedesktop.DBus.Introspectable")
	if err != nil {
		log.Panicf("failed to export node introspection: %s\n", err)
	}

	reply, err := conn.RequestName(I_DBUS_SERVICE_NAME,
		dbus.NameFlagDoNotQueue)
	if err != nil {
		panic(err)
	}
	if reply != dbus.RequestNameReplyPrimaryOwner {
		log.Panic("name already taken")
	}
	log.Printf("Listening on interface - %v and path %v ...\n", I_DBUS_SERVICE_NAME, I_DBUS_OBJECT_PATH)

	return conn
}
