package cmp

import (
	"testing"

	"com.telekom-mms.corp-net-indicator/internal/model"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

func TestIdentityDetailInit(t *testing.T) {
	gtk.Init()

	NewIdentityDetails(model.NewContext(), make(chan bool))
}

// TODO add assertions and tests
