package testserver

//go:generate dbus-codegen-go -prefix de.telekomMMS -package identity -camelize -output ../generated/identity/server/server.go ../schema/identity.xml

import (
	_ "embed"
	"log"
	"time"

	identity "de.telekom-mms.corp-net-indicator/internal/generated/identity/server"
	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/introspect"
	"github.com/godbus/dbus/v5/prop"
)

const I_DBUS_SERVICE_NAME = "de.telekomMMS.identity"
const I_DBUS_OBJECT_PATH = "/de/telekomMMS/identity"

type IdentityServer struct {
	state
}

var iS *IdentityServer

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
