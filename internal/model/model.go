package model

import (
	"sync"
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
	TrustedNetwork  bool
	LoggedIn        bool
	LastKeepAliveAt int64
	KrbIssuedAt     int64
	InProgress      bool
}

// vpn status transmitted by DBUS
type VPNStatus struct {
	TrustedNetwork bool
	Connected      bool
	IP             string
	Device         string
	ConnectedAt    int64
	CertExpiresAt  int64
	InProgress     bool
}

// credentials to read on login
type Credentials struct {
	Password string
	Server   string
}
