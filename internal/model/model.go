package model

type IdentityStatus struct {
	TrustedNetwork bool
	LoggedIn       bool
}

type VPNStatus struct {
	TrustedNetwork bool
	Connected      bool
	IP             string
	Device         string
}

type Details struct {
	Identity *IdentityStatus
	VPN      *VPNStatus
}
