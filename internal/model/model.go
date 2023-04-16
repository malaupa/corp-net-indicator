package model

type ctxKeys int

const (
	Trusted ctxKeys = iota
	Connected
	LoggedIn
	IdentityInProgress
	VPNInProgress
)

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
