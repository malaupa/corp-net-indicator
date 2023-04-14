package cmp

import "github.com/diamondburned/gotk4/pkg/gtk/v4"

type statusIcon struct {
	gtk.Image
}

func newStatusIcon(status bool) *statusIcon {
	icon := &statusIcon{*gtk.NewImage()}
	icon.setStatus(status)
	return icon
}

func (i *statusIcon) setStatus(status bool) {
	if status {
		i.SetFromIconName("emblem-default")
	} else {
		i.SetFromIconName("emblem-important")
	}
}
