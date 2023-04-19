package i18n

//go:generate gotext -srclang=en -dir=locales update -out=catalog.go -lang=en,de de.telekom-mms.corp-net-indicator

import (
	"os"

	"golang.org/x/text/message"
)

var L *message.Printer

// returns printer to translate messages
func init() {
	if L == nil {
		locale := os.Getenv("LANG")
		L = message.NewPrinter(message.MatchLanguage(locale[:2]))
	}
}
