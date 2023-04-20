package testserver

import (
	"fmt"
	"log"
	"sync/atomic"
	"time"

	vpn "de.telekom-mms.corp-net-indicator/internal/generated/vpn/server"
	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/introspect"
	"github.com/godbus/dbus/v5/prop"
)

//go:generate dbus-codegen-go -prefix de.telekomMMS -package vpn -camelize -output ../generated/vpn/server/server.go ../schema/vpn.xml

const V_DBUS_SERVICE_NAME = "de.telekomMMS.vpn"
const V_DBUS_OBJECT_PATH = "/de/telekomMMS/vpn"

type VPNServer struct {
	state
}

var counter atomic.Int32
var vS *VPNServer

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
