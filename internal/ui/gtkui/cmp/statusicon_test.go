package cmp_test

import (
	"testing"

	"de.telekom-mms.corp-net-indicator/internal/ui/gtkui/cmp"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/stretchr/testify/assert"
)

func TestStatusIcon(t *testing.T) {
	gtk.Init()

	s := cmp.NewStatusIcon(true)
	assert.Equal(t, "emblem-default", s.IconName())

	s.SetStatus(false)
	assert.Equal(t, "emblem-important", s.IconName())
	s.SetStatus(true)
	assert.Equal(t, "emblem-default", s.IconName())
}
