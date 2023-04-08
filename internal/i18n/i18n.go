package i18n

//go:generate gotext -srclang=en -dir=locales update -out=catalog.go -lang=en,de de.telekom-mms.corp-net-indicator

import (
	"os"

	"golang.org/x/text/message"
)

// TODO use context here...
var l *message.Printer

// returns printer to translate messages
func Localizer() *message.Printer {
	if l == nil {
		locale := os.Getenv("LANG")
		l = message.NewPrinter(message.MatchLanguage(locale[:2]))
	}
	return l
}
