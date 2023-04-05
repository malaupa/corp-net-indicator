package ui

import (
	"de.telekom-mms.corp-net-indicator/internal/i18n"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

func buildDetails() {
	l := i18n.Localizer()

	box := gtk.NewBox(gtk.OrientationVertical, 0)
	box.SetMarginTop(30)
	box.SetMarginBottom(30)
	box.SetMarginStart(60)
	box.SetMarginEnd(60)

	headBox := gtk.NewBox(gtk.OrientationHorizontal, 10)
	headBox.SetMarginBottom(10)
	headline := gtk.NewLabel(l.Sprintf("Corporate Network Status"))
	headline.AddCSSClass("title-2")
	headline.SetHAlign(gtk.AlignEnd)
	icon := gtk.NewImageFromIconName("preferences-system-network")
	icon.SetHAlign(gtk.AlignStart)
	icon.SetIconSize(gtk.IconSizeLarge)
	headBox.Append(icon)
	headBox.Append(headline)

	detailBox := gtk.NewListBox()
	detailBox.SetSelectionMode(gtk.SelectionNone)
	detailBox.SetShowSeparators(true)
	detailBox.AddCSSClass("rich-list")

	box.Append(headBox)
}
