package cmp

import (
	"testing"

	"com.telekom-mms.corp-net-indicator/internal/model"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

func TestInitVPNDetail(t *testing.T) {
	gtk.Init()

	NewVPNDetail(
		model.NewContext(),
		make(chan *model.Credentials),
		&gtk.Window{},
		func() ([]string, error) {
			return []string{}, nil
		},
		nil,
	)
}

// TODO add assertions and tests
