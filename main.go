package main

import (
	"flag"

	"de.telekom-mms.corp-net-indicator/internal/logger"
	"de.telekom-mms.corp-net-indicator/internal/tray"
	"de.telekom-mms.corp-net-indicator/internal/ui"
)

var runAsTray bool
var quickConnect bool
var verbose bool

func init() {
	flag.BoolVar(&runAsTray, "tray", false, "start as tray icon")
	flag.BoolVar(&quickConnect, "quick", false, "quick connect to vpn")
	flag.BoolVar(&verbose, "v", false, "verbose logging")
}

// entry point
func main() {
	flag.Parse()

	if runAsTray {
		logger.Setup("TRAY", verbose)
		tray.New().Run()
	} else {
		logger.Setup("WINDOW", verbose)
		ui.NewStatus().Run(quickConnect)
	}
}
