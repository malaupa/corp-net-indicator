package main

//go:generate dbus-codegen-go -prefix de.telekomMMS -package identity -camelize -output ../internal/generated/identity/server/server.go ../internal/schema/identity.xml
//go:generate dbus-codegen-go -client-only -prefix de.telekomMMS -package identity -camelize -output ../internal/generated/identity/client/client.go ../internal/schema/identity.xml
//go:generate dbus-codegen-go -prefix de.telekomMMS -package vpn -camelize -output ../internal/generated/vpn/server/server.go ../internal/schema/vpn.xml
//go:generate dbus-codegen-go -client-only -prefix de.telekomMMS -package vpn -camelize -output ../internal/generated/vpn/client/client.go ../internal/schema/vpn.xml

import (
	_ "embed"
	"fmt"
	"log"
	"os"
	"time"

	identity "de.telekom-mms.corp-net-indicator/internal/generated/identity/server"
	vpn "de.telekom-mms.corp-net-indicator/internal/generated/vpn/server"
	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/introspect"
	"github.com/godbus/dbus/v5/prop"
)

const I_DBUS_SERVICE_NAME = "de.telekomMMS.identity"
const I_DBUS_OBJECT_PATH = "/de/telekomMMS/identity"
const V_DBUS_SERVICE_NAME = "de.telekomMMS.vpn"
const V_DBUS_OBJECT_PATH = "/de/telekomMMS/vpn"

var conn *dbus.Conn
var systemConn *dbus.Conn
var iProps *prop.Properties
var vProps *prop.Properties

type Identity struct {
	*identity.UnimplementedIdentity
}

func (i Identity) GetStatus() (map[string]dbus.Variant, *dbus.Error) {
	log.Println("Identity: GetStatus called!")
	variant, err := iProps.Get(I_DBUS_SERVICE_NAME, "Status")
	value := variant.Value()
	if v, ok := value.(map[string]dbus.Variant); ok {
		return v, err
	}
	return map[string]dbus.Variant{}, err
}

func (i Identity) ReLogin() *dbus.Error {
	log.Println("Identity: ReLogin called!")
	setAndEmitIdentitySignal(3)
	return nil
}

func setAndEmitIdentitySignal(sec time.Duration) {
	go func() {
		time.Sleep(time.Second * sec)
		status := buildIdentityStatus(true)
		err := iProps.Set(I_DBUS_SERVICE_NAME, "Status", dbus.MakeVariant(status))
		if err != nil {
			log.Println(err)
		}
		emitIdentitySignal(status)
	}()
}

func emitIdentitySignal(status map[string]dbus.Variant) {
	if err := identity.Emit(conn, &identity.IdentityStatusChangeSignal{
		Path: I_DBUS_OBJECT_PATH,
		Body: &identity.IdentityStatusChangeSignalBody{
			Status: status,
		},
	}); err != nil {
		log.Println(err)
	}
}

func buildIdentityStatus(loggedIn bool) map[string]dbus.Variant {
	return map[string]dbus.Variant{
		"TrustedNetwork":  dbus.MakeVariant(true),
		"LoggedIn":        dbus.MakeVariant(loggedIn),
		"LastKeepAliveAt": dbus.MakeVariant(time.Now().Unix()),
		"KrbIssuedAt":     dbus.MakeVariant(time.Now().Unix() - 60*60),
	}
}

type VPN struct {
	*vpn.UnimplementedVpn
}

func (v VPN) GetStatus() (map[string]dbus.Variant, *dbus.Error) {
	log.Println("VPN: GetStatus called!")
	variant, err := vProps.Get(V_DBUS_SERVICE_NAME, "Status")
	value := variant.Value()
	if v, ok := value.(map[string]dbus.Variant); ok {
		return v, err
	}
	return map[string]dbus.Variant{}, err
}

func (v VPN) Connect(password string, server string) *dbus.Error {
	log.Printf("VPN: Connect called! Password[%s] Server[%s]\n", password, server)
	setAndEmitVPNStatus(false, true)
	setAndEmitIdentitySignal(10)
	return nil
}

func (v VPN) Disconnect() *dbus.Error {
	log.Printf("VPN: Disconnect called!\n")
	setAndEmitVPNStatus(false, false)
	return nil
}

func (v VPN) ListServers() (servers []string, err *dbus.Error) {
	return []string{
		"server1.example.com",
		"server2.example.com",
		"server3.example.com",
	}, nil
}

func setAndEmitVPNStatus(trusted, connected bool) {
	go func() {
		time.Sleep(time.Second * 5)
		status := buildVPNStatus(trusted, connected)
		err := vProps.Set(V_DBUS_SERVICE_NAME, "Status", dbus.MakeVariant(status))
		if err != nil {
			log.Println(err)
		}
		emitVPNSignal(status)
	}()
}

func buildVPNStatus(trusted, connected bool) map[string]dbus.Variant {
	return map[string]dbus.Variant{
		"TrustedNetwork": dbus.MakeVariant(trusted),
		"Connected":      dbus.MakeVariant(connected),
		"IP":             dbus.MakeVariant("127.0.0.1"),
		"Device":         dbus.MakeVariant("vpn-tun0"),
		"ConnectedAt":    dbus.MakeVariant(time.Now().Unix()),
		"CertExpiresAt":  dbus.MakeVariant(time.Now().Unix() + 60*60*24*365),
		"ServerList":     dbus.MakeVariant([]string{"server1.example.org", "server2.example.org"}),
	}
}

func emitVPNSignal(status map[string]dbus.Variant) {
	if err := vpn.Emit(systemConn, &vpn.VpnStatusChangeSignal{
		Path: V_DBUS_OBJECT_PATH,
		Body: &vpn.VpnStatusChangeSignalBody{
			Status: status,
		},
	}); err != nil {
		log.Println(err)
	}
}

// func init() {
// 	runtime.LockOSThread()
// }

func main() {
	// dbus connection
	var err error
	conn, err = dbus.ConnectSessionBus()
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	systemConn, err = dbus.ConnectSystemBus()
	if err != nil {
		panic(err)
	}
	defer systemConn.Close()

	// identity introspection
	node := introspect.Node{
		Name: I_DBUS_OBJECT_PATH,
		Interfaces: []introspect.Interface{
			introspect.IntrospectData,
			prop.IntrospectData,
			identity.IntrospectDataIdentity,
		},
	}
	err = conn.Export(introspect.NewIntrospectable(&node), I_DBUS_OBJECT_PATH,
		"org.freedesktop.DBus.Introspectable")
	if err != nil {
		log.Printf("failed to export node introspection: %s\n", err)
		return
	}
	// identity methods
	i := Identity{}
	err = identity.ExportIdentity(conn, I_DBUS_OBJECT_PATH, i)
	if err != nil {
		panic(err)
	}
	// identity props
	iProps, err = prop.Export(conn, I_DBUS_OBJECT_PATH, map[string]map[string]*prop.Prop{
		I_DBUS_SERVICE_NAME: {
			"Version": {
				Value:    1,
				Writable: false,
				Emit:     prop.EmitTrue,
				Callback: nil,
			},
			"Status": {
				Value:    buildIdentityStatus(false),
				Writable: true,
				Emit:     prop.EmitTrue,
				Callback: nil,
			},
		},
	})
	if err != nil {
		panic(err)
	}

	reply, err := conn.RequestName(I_DBUS_SERVICE_NAME,
		dbus.NameFlagDoNotQueue)
	if err != nil {
		panic(err)
	}
	if reply != dbus.RequestNameReplyPrimaryOwner {
		fmt.Fprintln(os.Stderr, "name already taken")
		os.Exit(1)
	}
	fmt.Printf("Listening on interface - %v and path %v ...\n", I_DBUS_SERVICE_NAME, I_DBUS_OBJECT_PATH)

	// vpn introspection
	vNode := introspect.Node{
		Name: V_DBUS_OBJECT_PATH,
		Interfaces: []introspect.Interface{
			introspect.IntrospectData,
			prop.IntrospectData,
			vpn.IntrospectDataVpn,
		},
	}
	err = systemConn.Export(introspect.NewIntrospectable(&vNode), V_DBUS_OBJECT_PATH,
		"org.freedesktop.DBus.Introspectable")
	if err != nil {
		log.Printf("failed to export node introspection: %s\n", err)
		return
	}
	// vpn methods
	v := VPN{}
	err = vpn.ExportVpn(systemConn, V_DBUS_OBJECT_PATH, v)
	if err != nil {
		panic(err)
	}
	// identity props
	vProps, err = prop.Export(systemConn, V_DBUS_OBJECT_PATH, map[string]map[string]*prop.Prop{
		V_DBUS_SERVICE_NAME: {
			"Version": {
				Value:    1,
				Writable: false,
				Emit:     prop.EmitTrue,
				Callback: nil,
			},
			"Status": {
				Value:    buildVPNStatus(false, false),
				Writable: true,
				Emit:     prop.EmitTrue,
				Callback: nil,
			},
		},
	})
	if err != nil {
		panic(err)
	}

	reply, err = systemConn.RequestName(V_DBUS_SERVICE_NAME,
		dbus.NameFlagDoNotQueue)
	if err != nil {
		panic(err)
	}
	if reply != dbus.RequestNameReplyPrimaryOwner {
		fmt.Fprintln(os.Stderr, "name already taken")
		os.Exit(1)
	}
	fmt.Printf("Listening on interface - %v and path %v ...\n", V_DBUS_SERVICE_NAME, V_DBUS_OBJECT_PATH)

	if true {
		select {}
	}

	for {
		time.Sleep(time.Second * 10)
		vStatus := buildVPNStatus(false, true)
		vProps.SetMust(V_DBUS_SERVICE_NAME, "Status", dbus.MakeVariant(vStatus))
		emitVPNSignal(vStatus)
		log.Println("VPN connected")

		time.Sleep(time.Second * 2)
		iStatus := buildIdentityStatus(true)
		iProps.SetMust(I_DBUS_SERVICE_NAME, "Status", dbus.MakeVariant(iStatus))
		emitIdentitySignal(iStatus)
		log.Println("Identity loggedIn")

		time.Sleep(time.Second * 20)
		vStatus = buildVPNStatus(false, false)
		vProps.SetMust(V_DBUS_SERVICE_NAME, "Status", dbus.MakeVariant(vStatus))
		emitVPNSignal(vStatus)
		log.Println("VPN disconnected")

		time.Sleep(time.Second * 10)
		iStatus = buildIdentityStatus(false)
		iProps.SetMust(I_DBUS_SERVICE_NAME, "Status", dbus.MakeVariant(iStatus))
		emitIdentitySignal(iStatus)
		log.Println("Identity loggedOut")
	}
}
