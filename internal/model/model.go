package model

import "context"

type ctxKeys int

const (
	Trusted ctxKeys = iota
	Connected
	LoggedIn
	InProgress
)

func IncrementProgress(ctx context.Context) context.Context {
	return context.WithValue(ctx, InProgress, ctx.Value(InProgress).(int)+1)
}

func DecrementProgress(ctx context.Context) context.Context {
	val := ctx.Value(InProgress).(int)
	if val == 0 {
		return ctx
	}
	return context.WithValue(ctx, InProgress, val-1)
}

type IdentityStatus struct {
	TrustedNetwork  bool
	LoggedIn        bool
	LastKeepAliveAt int64
	KrbIssuedAt     int64
}

type VPNStatus struct {
	TrustedNetwork bool
	Connected      bool
	IP             string
	Device         string
	ConnectedAt    int64
	CertExpiresAt  int64
	ServerList     []string
}

type Credentials struct {
	Password string
	Server   string
}
