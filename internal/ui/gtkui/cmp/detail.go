package cmp

import (
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type detail struct {
	gtk.Box
	list *gtk.ListBox
}

func newDetail() *detail {
	return &detail{Box: *gtk.NewBox(gtk.OrientationVertical, 10)}
}

func (d *detail) buildBase(title string) *detail {
	d.SetMarginBottom(20)
	d.list = gtk.NewListBox()
	d.list.SetSelectionMode(gtk.SelectionNone)
	d.list.SetShowSeparators(true)
	d.list.AddCSSClass("rich-list")
	label := gtk.NewLabel(title)
	label.SetMarginBottom(10)
	label.SetHAlign(gtk.AlignStart)
	label.AddCSSClass("title-4")
	d.Append(label)
	frame := gtk.NewFrame("")
	frame.SetChild(d.list)
	d.Append(frame)
	return d
}

func (d *detail) addRow(labelText string, value ...gtk.Widgetter) *detail {
	box := gtk.NewBox(gtk.OrientationHorizontal, 10)
	label := gtk.NewLabel(labelText)
	label.SetHAlign(gtk.AlignStart)
	label.SetHExpand(true)
	box.Append(label)
	for _, w := range value {
		box.Append(w)
	}
	row := gtk.NewListBoxRow()
	row.SetChild(box)
	row.SetActivatable(false)
	row.SetSizeRequest(340, 0)
	d.list.Append(row)
	return d
}
