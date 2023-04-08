package main

import (
	"context"
	"fmt"

	"de.telekom-mms.corp-net-indicator/internal/other"
)

func main() {
	fmt.Println(context.Background() == other.Test())
}
