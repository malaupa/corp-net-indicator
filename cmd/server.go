package main

//go:generate dbus-codegen-go -prefix de.telekomMMS -package identity -camelize -output ../internal/generated/identity/identity.go ../internal/schema/identity.xml

import (
	_ "embed"
	"fmt"
	"log"
	"os"
	"time"

	"de.telekom-mms.corp-net-indicator/internal/generated/identity"
	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/introspect"
	"github.com/godbus/dbus/v5/prop"
)

const DBUS_SERVICE_NAME = "de.telekomMMS.identity"
const DBUS_OBJECT_PATH = "/de/telekomMMS/identity"

var props *prop.Properties

type Identity struct {
	*identity.UnimplementedIdentity
}

func (i Identity) GetStatus() (map[string]dbus.Variant, *dbus.Error) {
	log.Println("GetStatus called!")
	variant, err := props.Get(DBUS_SERVICE_NAME, "Status")
	value := variant.Value()
	if v, ok := value.(map[string]dbus.Variant); ok {
		return v, err
	}
	return map[string]dbus.Variant{}, err
}

func (i Identity) ReLogin() *dbus.Error {
	log.Println("ReLogin called!")
	return props.Set(DBUS_SERVICE_NAME, "Status", dbus.MakeVariant(map[string]dbus.Variant{"TrustedNetwork": dbus.MakeVariant(true), "LoggedIn": dbus.MakeVariant(true)}))
}

func main() {
	// debus connection
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// introspection
	node := introspect.Node{
		Name: DBUS_OBJECT_PATH,
		Interfaces: []introspect.Interface{
			introspect.IntrospectData,
			prop.IntrospectData,
			identity.IntrospectDataIdentity,
		},
	}
	err = conn.Export(introspect.NewIntrospectable(&node), DBUS_OBJECT_PATH,
		"org.freedesktop.DBus.Introspectable")
	if err != nil {
		log.Printf("failed to export node introspection: %s\n", err)
		return
	}
	// methods
	i := Identity{}
	err = identity.ExportIdentity(conn, DBUS_OBJECT_PATH, i)
	if err != nil {
		panic(err)
	}
	// props
	props, err = prop.Export(conn, DBUS_OBJECT_PATH, map[string]map[string]*prop.Prop{
		DBUS_SERVICE_NAME: {
			"Version": {
				Value:    1,
				Writable: false,
				Emit:     prop.EmitTrue,
				Callback: nil,
			},
			"Status": {
				Value:    map[string]dbus.Variant{"TrustedNetwork": dbus.MakeVariant(false), "LoggedIn": dbus.MakeVariant(false)},
				Writable: true,
				Emit:     prop.EmitTrue,
				Callback: nil,
			},
		},
	})
	if err != nil {
		panic(err)
	}

	reply, err := conn.RequestName(DBUS_SERVICE_NAME,
		dbus.NameFlagDoNotQueue)
	if err != nil {
		panic(err)
	}
	if reply != dbus.RequestNameReplyPrimaryOwner {
		fmt.Fprintln(os.Stderr, "name already taken")
		os.Exit(1)
	}
	fmt.Printf("Listening on interface - %v and path %v ...\n", DBUS_SERVICE_NAME, DBUS_OBJECT_PATH)
	// select {}
	t, l := true, false
	for {
		time.Sleep(time.Second * 10)
		status := map[string]dbus.Variant{"TrustedNetwork": dbus.MakeVariant(t), "LoggedIn": dbus.MakeVariant(l)}
		props.SetMust(DBUS_SERVICE_NAME, "Status", status)
		if err := identity.Emit(conn, &identity.IdentityStatusChangeSignal{
			Path: DBUS_OBJECT_PATH,
			Body: &identity.IdentityStatusChangeSignalBody{
				Status: status,
			},
		}); err != nil {
			log.Println(err)
		}

		l = !l
	}
}
