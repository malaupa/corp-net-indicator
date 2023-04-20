package testserver

//go:generate dbus-codegen-go -prefix de.telekomMMS -package identity -camelize -output ../generated/identity/server/server.go ../schema/identity.xml
//go:generate dbus-codegen-go -prefix de.telekomMMS -package vpn -camelize -output ../generated/vpn/server/server.go ../schema/vpn.xml

import (
	_ "embed"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
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

var counter atomic.Int32

type state struct {
	sync.Mutex
	state    map[string]dbus.Variant
	conn     *dbus.Conn
	simulate bool
}

func (s *state) getState() map[string]dbus.Variant {
	s.Lock()
	defer s.Unlock()
	return s.state
}

func (s *state) setInProgress() map[string]dbus.Variant {
	s.Lock()
	defer s.Unlock()
	s.state["InProgress"] = dbus.MakeVariant(true)
	return s.state
}

type IdentityServer struct {
	state
}

type VPNServer struct {
	state
}

var iS *IdentityServer
var vS *VPNServer

type Identity struct {
	identity.UnimplementedIdentity
}

func (i *Identity) GetStatus() (map[string]dbus.Variant, *dbus.Error) {
	log.Println("Identity: GetStatus called!")
	return iS.getState(), nil
}

func (i *Identity) ReLogin() *dbus.Error {
	log.Println("Identity: ReLogin called!")
	iS.emitIdentitySignal(iS.setInProgress())
	go func() {
		if iS.simulate {
			time.Sleep(time.Second * 3)
		}
		iS.emitIdentitySignal(iS.buildIdentityStatus(true))
	}()
	return nil
}

func NewIdentityServer(simulate bool) *IdentityServer {
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		panic(err)
	}
	iS = &IdentityServer{state: state{conn: conn, state: make(map[string]dbus.Variant), simulate: simulate}}
	iS.buildIdentityStatus(false)

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
		log.Panicf("failed to export node introspection: %s\n", err)
	}
	// identity methods
	i := &Identity{}
	err = identity.ExportIdentity(conn, I_DBUS_OBJECT_PATH, i)
	if err != nil {
		panic(err)
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

	return iS
}

func (s *IdentityServer) emitIdentitySignal(state map[string]dbus.Variant) {
	if err := identity.Emit(s.conn, &identity.IdentityStatusChangeSignal{
		Path: I_DBUS_OBJECT_PATH,
		Body: &identity.IdentityStatusChangeSignalBody{
			Status: state,
		},
	}); err != nil {
		log.Println(err)
	}
}

func (s *IdentityServer) buildIdentityStatus(loggedIn bool) map[string]dbus.Variant {
	s.Lock()
	defer s.Unlock()
	var now int64 = 60 * 60
	if s.simulate {
		now = time.Now().Unix()
	}
	s.state.state = map[string]dbus.Variant{
		"InProgress":      dbus.MakeVariant(false),
		"TrustedNetwork":  dbus.MakeVariant(true),
		"LoggedIn":        dbus.MakeVariant(loggedIn),
		"LastKeepAliveAt": dbus.MakeVariant(now),
		"KrbIssuedAt":     dbus.MakeVariant(now - 60*60),
	}
	return s.state.state
}

func (s *IdentityServer) Close() {
	s.conn.Close()
}

type VPN struct {
	vpn.UnimplementedVpn
}

func (v *VPN) GetStatus() (map[string]dbus.Variant, *dbus.Error) {
	log.Printf("VPN: GetStatus called!\n")
	return vS.getState(), nil
}

func (v *VPN) Connect(password string, server string) *dbus.Error {
	log.Printf("VPN: Connect called! Password[%s] Server[%s]\n", password, server)
	vS.emitVPNSignal(vS.setInProgress())
	go func() {
		if vS.simulate {
			time.Sleep(time.Second * 5)
		}
		vS.emitVPNSignal(vS.buildVPNStatus(false, true))
		if vS.simulate {
			time.Sleep(time.Second * 5)
		}
		if iS != nil {
			iS.emitIdentitySignal(iS.buildIdentityStatus(true))
		}
	}()
	return nil
}

func (v *VPN) Disconnect() *dbus.Error {
	if vS.simulate {
		counter.Add(1)
		if counter.Load()%3 == 0 {
			counter.Store(0)
			return dbus.MakeFailedError(fmt.Errorf("Disconnect failed"))
		}
	}
	log.Printf("VPN: Disconnect called!\n")
	vS.emitVPNSignal(vS.setInProgress())
	go func() {
		if vS.simulate {
			time.Sleep(time.Second * 5)
		}
		vS.emitVPNSignal(vS.buildVPNStatus(false, false))
		if vS.simulate {
			time.Sleep(time.Second * 5)
		}
		if iS != nil {
			iS.emitIdentitySignal(iS.buildIdentityStatus(false))
		}
	}()
	return nil
}

func (v *VPN) ListServers() (servers []string, err *dbus.Error) {
	log.Printf("VPN: ListServers called!\n")
	return []string{
		"server1.example.com",
		"server2.example.com",
		"server3.example.com",
	}, nil
}

func NewVPNServer(simulate bool) *VPNServer {
	systemConn, err := dbus.ConnectSystemBus()
	if err != nil {
		panic(err)
	}
	vS = &VPNServer{state: state{conn: systemConn, state: make(map[string]dbus.Variant), simulate: simulate}}
	vS.buildVPNStatus(false, false)

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
		log.Panicf("failed to export node introspection: %s\n", err)
	}
	// vpn methods
	v := &VPN{}
	err = vpn.ExportVpn(systemConn, V_DBUS_OBJECT_PATH, v)
	if err != nil {
		panic(err)
	}

	reply, err := systemConn.RequestName(V_DBUS_SERVICE_NAME,
		dbus.NameFlagDoNotQueue)
	if err != nil {
		panic(err)
	}
	if reply != dbus.RequestNameReplyPrimaryOwner {
		log.Panic("name already taken")
	}
	log.Printf("Listening on interface - %v and path %v ...\n", V_DBUS_SERVICE_NAME, V_DBUS_OBJECT_PATH)

	return vS
}

func (s *VPNServer) buildVPNStatus(trusted, connected bool) map[string]dbus.Variant {
	s.Lock()
	defer s.Unlock()
	var now int64 = 0
	if s.simulate {
		now = time.Now().Unix()
	}
	s.state.state = map[string]dbus.Variant{
		"InProgress":     dbus.MakeVariant(false),
		"TrustedNetwork": dbus.MakeVariant(trusted),
		"Connected":      dbus.MakeVariant(connected),
		"IP":             dbus.MakeVariant("127.0.0.1"),
		"Device":         dbus.MakeVariant("vpn-tun0"),
		"ConnectedAt":    dbus.MakeVariant(now),
		"CertExpiresAt":  dbus.MakeVariant(now + 60*60*24*365),
	}
	return s.state.state
}

func (s *VPNServer) emitVPNSignal(status map[string]dbus.Variant) {
	if err := vpn.Emit(s.conn, &vpn.VpnStatusChangeSignal{
		Path: V_DBUS_OBJECT_PATH,
		Body: &vpn.VpnStatusChangeSignalBody{
			Status: status,
		},
	}); err != nil {
		log.Println(err)
	}
}

func (s *VPNServer) Close() {
	s.conn.Close()
}
