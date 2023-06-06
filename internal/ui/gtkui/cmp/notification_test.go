package cmp_test

import (
	"testing"

	"com.telekom-mms.corp-net-indicator/internal/ui/gtkui/cmp"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/stretchr/testify/assert"
)

func TestNotification(t *testing.T) {
	gtk.Init()

	n := cmp.NewNotification()
	assert.Equal(t, n.Revealer.RevealChild(), false)

	n.Show("Hu!")
	assert.Equal(t, n.Revealer.RevealChild(), true)
}

// TODO add assertions and tests
