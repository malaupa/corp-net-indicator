package i18n

//go:generate gotext -srclang=en -dir=locales update -out=catalog.go -lang=en,de de.telekom-mms.corp-net-indicator

import (
	"os"

	"golang.org/x/text/message"
)

// returns printer to translate messages
func Localizer() *message.Printer {
	locale := os.Getenv("LANG")
	return message.NewPrinter(message.MatchLanguage(locale[:2]))
}
