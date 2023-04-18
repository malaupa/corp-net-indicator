package model

import (
	"sync"
)

type ContextValues struct {
	TrustedNetwork     bool
	Connected          bool
	LoggedIn           bool
	IdentityInProgress bool
	VPNInProgress      bool
}

type Context struct {
	lock sync.RWMutex

	values ContextValues
}

func NewContext() *Context {
	return &Context{lock: sync.RWMutex{}, values: ContextValues{}}
}

func (c *Context) Write(writer func(ctx *ContextValues)) ContextValues {
	c.lock.Lock()
	defer c.lock.Unlock()
	writer(&c.values)
	return c.values
}

func (c *Context) Read() ContextValues {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.values
}

type IdentityStatus struct {
	TrustedNetwork  bool
	LoggedIn        bool
	LastKeepAliveAt int64
	KrbIssuedAt     int64
	InProgress      bool
}

type VPNStatus struct {
	TrustedNetwork bool
	Connected      bool
	IP             string
	Device         string
	ConnectedAt    int64
	CertExpiresAt  int64
	ServerList     []string
	InProgress     bool
}

type CanBeInProgress interface {
	VPNStatus | IdentityStatus
}

type Credentials struct {
	Password string
	Server   string
}
