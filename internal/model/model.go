package model

import (
	"fmt"
	"reflect"
	"sync"
)

// trusted state
const (
	TrustUnknown uint32 = iota
	NotTrusted
	Trusted
)

// identity agent login states
const (
	LoginUnknown uint32 = iota
	LoggedOut
	LoggingIn
	LoggedIn
	LoggingOut
)

// vpn connection state
const (
	ConnectUnknown uint32 = iota
	Disconnected
	Connecting
	Connected
	Disconnecting
)

// type to hold context values
type ContextValues struct {
	// is current network trusted
	TrustedNetwork bool
	// is vpn connected
	Connected bool
	// is identity agent logged in
	LoggedIn bool
	// is identity agent action in progress
	IdentityInProgress bool
	// is vpn action in progress
	VPNInProgress bool
}

// holds context values and handles write and read accesses
type Context struct {
	lock sync.RWMutex

	values ContextValues
}

// creates new context
func NewContext() *Context {
	return &Context{lock: sync.RWMutex{}, values: ContextValues{}}
}

// provides writer to write context values
func (c *Context) Write(writer func(ctx *ContextValues)) ContextValues {
	c.lock.Lock()
	defer c.lock.Unlock()
	writer(&c.values)
	return c.values
}

// returns copy of context values to read
func (c *Context) Read() ContextValues {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.values
}

// identity status transmitted by DBUS
type IdentityStatus struct {
	TrustedNetwork       *uint32
	LoginState           *uint32
	LastKeepAliveAt      *int64
	KerberosTGTStartTime *int64
	KerberosTGTEndTime   *int64
}

func (s *IdentityStatus) String() string {
	return ToString(s)
}

func (s *IdentityStatus) InProgress(ctxInProgress bool) bool {
	return (s.LoginState != nil && (*s.LoginState == LoggingIn || *s.LoginState == LoggingOut)) ||
		(s.LoginState == nil && ctxInProgress)
}

func (s *IdentityStatus) IsLoggedIn(ctxIsLoggedIn bool) bool {
	return (s.LoginState != nil && *s.LoginState == LoggedIn) ||
		(s.LoginState == nil && ctxIsLoggedIn)
}

// vpn status transmitted by DBUS
type VPNStatus struct {
	TrustedNetwork  *uint32
	ConnectionState *uint32
	IP              *string
	Device          *string
	ConnectedAt     *int64
	CertExpiresAt   *int64
}

func (s *VPNStatus) String() string {
	return ToString(s)
}

func (s *VPNStatus) IsTrustedNetwork(ctxTrusted bool) bool {
	return (s.TrustedNetwork != nil && *s.TrustedNetwork == Trusted) ||
		(s.TrustedNetwork == nil && ctxTrusted)
}

func (s *VPNStatus) IsConnected(ctxConnected bool) bool {
	return (s.ConnectionState != nil && *s.ConnectionState == Connected) ||
		(s.ConnectionState == nil && ctxConnected)
}

func (s *VPNStatus) InProgress(ctxInProgress bool) bool {
	return (s.ConnectionState != nil && (*s.ConnectionState == Connecting || *s.ConnectionState == Disconnecting)) ||
		(s.ConnectionState == nil && ctxInProgress)
}

// credentials to read on login
type Credentials struct {
	Password string
	Server   string
}

// expands pointers to values in structs
func ToString[T interface{}](structType T) string {
	result := "{"
	// unwrap result struct
	elem := reflect.ValueOf(structType).Elem()
	elemType := elem.Type()
	// iterate over struct fields
	for i := 0; i < elem.NumField(); i++ {
		// get and check field value
		fieldValue := elem.Field(i)
		if !fieldValue.IsValid() {
			continue
		}
		// add separator
		if i != 0 {
			result += ", "
		}
		// get field definition
		field := elemType.Field(i)
		var val any = fieldValue
		if field.Type.Kind() == reflect.Pointer {
			if fieldValue.IsNil() {
				val = "<nil>"
			} else {
				val = fieldValue.Elem()
			}
		}
		result += fmt.Sprintf("%s: %v", field.Name, val)
	}
	return result + "}"
}
