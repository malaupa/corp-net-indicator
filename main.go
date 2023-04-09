package main

import (
	"context"

	"de.telekom-mms.corp-net-indicator/internal/tray"
)

// entry point
func main() {
	tray.New(context.Background())
}
