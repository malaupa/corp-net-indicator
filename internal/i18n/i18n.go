package i18n

//go:generate gotext -srclang=en -dir=locales update -out=catalog.go -lang=en,de com.telekom-mms.corp-net-indicator

import (
	"os"

	"golang.org/x/text/message"
)

// printer to use
var L *message.Printer

// init printer for localized messages and labels
func init() {
	if L == nil {
		locale := os.Getenv("LANG")
		L = message.NewPrinter(message.MatchLanguage(locale[:2]))
	}
}
