package cmp

import gtk "github.com/diamondburned/gotk4/pkg/gtk/v4"

type statusIcon struct {
	gtk.Image
}

// creates new status icon
func NewStatusIcon(status bool) *statusIcon {
	icon := &statusIcon{*gtk.NewImage()}
	icon.SetStatus(status)
	return icon
}

// changes icon -> true = green check, false = red cross
func (i *statusIcon) SetStatus(status bool) {
	if status {
		i.SetFromIconName("emblem-default")
	} else {
		i.SetFromIconName("emblem-important")
	}
}
