package model

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
