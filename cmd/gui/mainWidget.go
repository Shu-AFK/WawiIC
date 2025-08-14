package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type Item struct {
	ID       string
	Name     string
	IsFather bool
	IsChild  bool
}

func createMainWidget(canvas fyne.Canvas) fyne.CanvasObject {
	searchbar := widget.NewEntry()
	searchbar.SetPlaceHolder("Artikel...")

	rows := container.NewVBox()
	scroll := container.NewVScroll(rows)

	searchbar.OnSubmitted = func(query string) {
		onSearch(query, rows, canvas)
	}

	content := container.NewBorder(
		searchbar,
		nil,
		nil,
		nil,
		scroll,
	)

	return content
}

func onSearch(query string, rows *fyne.Container, canvas fyne.Canvas) {
	// TODO: Query items from API instead of test items
	items := []Item{
		{ID: "1", Name: "Binoculars", IsFather: true, IsChild: false},
		{ID: "2", Name: "Camera", IsFather: false, IsChild: true},
		{ID: "3", Name: "Tripod", IsFather: false, IsChild: false},
	}

	// Clear previous rows
	rows.Objects = nil

	for _, item := range items {
		row := container.NewHBox(
			truncatedLabelWithTooltip(item.ID, MaxIdLength, canvas),
			truncatedLabelWithTooltip(item.Name, MaxNameLength, canvas),
			createDisabledCheck("Vaterartikel", item.IsFather),
			createDisabledCheck("Kindartikel", item.IsChild),
		)
		rows.Add(row)
	}

	rows.Refresh()
}

func truncatedLabelWithTooltip(text string, maxLen int, canvas fyne.Canvas) fyne.CanvasObject {
	display := text
	if len(text) > maxLen {
		display = text[:maxLen] + "…"
	}

	btn := widget.NewButton(display, nil)
	btn.Importance = widget.LowImportance // makes it look more like a label

	if display != text {
		btn.OnTapped = func() {
			popup := widget.NewPopUp(widget.NewLabel(text), canvas)
			popup.Show()
		}
	}

	return btn
}

// createDisabledCheck creates a disabled checkbox with the given state
func createDisabledCheck(text string, checked bool) *widget.Check {
	cb := widget.NewCheck(text, nil)
	cb.SetChecked(checked)
	cb.Disable()
	return cb
}
