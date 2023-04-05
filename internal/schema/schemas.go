package schema

import (
	"bytes"
	_ "embed"
	"io"
	"text/template"

	"github.com/godbus/dbus/v5/introspect"
)

//go:embed identity.xml
var identityIntro string

//go:embed vpn.xml
var vpnIntro string

type intros struct {
	IntrospectDataString string
}

type schema int

const (
	Identity schema = iota
	VPN
)

func Schema(s schema) string {
	t := template.New("tmpl")
	var err error
	switch s {
	case Identity:
		t, err = t.Parse(identityIntro)
	case VPN:
		t, err = t.Parse(vpnIntro)
	}
	if err != nil {
		return ""
	}
	var buff bytes.Buffer
	err = t.Execute(io.Writer(&buff), intros{introspect.IntrospectDataString})
	if err != nil {
		return ""
	}
	return buff.String()
}
