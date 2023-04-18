package main

import (
	"flag"

	"de.telekom-mms.corp-net-indicator/internal/tray"
	"de.telekom-mms.corp-net-indicator/internal/ui"
)

var runAsTray bool
var quickConnect bool

func init() {
	flag.BoolVar(&runAsTray, "tray", false, "start as tray icon")
	flag.BoolVar(&quickConnect, "quick", false, "quick connect to vpn")
}

// entry point
func main() {
	flag.Parse()

	if runAsTray {
		tray.New().Run()
	} else {
		ui.NewStatus().Run(quickConnect)
	}
}
