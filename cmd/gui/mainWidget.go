package gui

// TODO: Filter for child categories not working

import (
	"errors"

	"github.com/Shu-AFK/WawiIC/cmd/helper"
	"github.com/Shu-AFK/WawiIC/cmd/wawi"
	"github.com/Shu-AFK/WawiIC/cmd/wawi/wawi_structs"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

var Selected []wawi_structs.WItem

func createMainWidget(canvas fyne.Canvas, app fyne.App, w fyne.Window) fyne.CanvasObject {
	var prevSearch string

	searchbar := widget.NewEntry()
	searchbar.SetPlaceHolder("Artikel...")

	rows := container.NewVBox()
	scroll := container.NewVScroll(rows)

	mergeButton := widget.NewButton("Zusammenfügen", func() {
		if len(Selected) <= 1 {
			dialog.ShowInformation("Achtung!", "Bitte wähle als erstes 2 oder mehr Artikel aus, welche du zusammenfügen willst.", w)
			return
		}

		combineW := app.NewWindow("Zusammenfügen")
		combineW.SetOnClosed(func() {
			Selected = Selected[:0]
			onSearch(prevSearch, rows, canvas, w)
		})

		CombineWindow(combineW, app, Selected)
	})

	mergeButton.Importance = widget.LowImportance
	mergeButton.Resize(fyne.NewSize(80, 40))

	buttonContainer := container.New(layout.NewHBoxLayout(), layout.NewSpacer(), mergeButton)

	searchbar.OnSubmitted = func(query string) {
		prevSearch = query
		onSearch(query, rows, canvas, w)
	}

	content := container.NewBorder(
		searchbar,
		buttonContainer,
		nil,
		nil,
		scroll,
	)

	return content
}

func onSearch(query string, rows *fyne.Container, canvas fyne.Canvas, w fyne.Window) {
	// Clear previous rows
	rows.Objects = nil
	rows.Refresh()

	items, err := wawi.GetItems(query, SelectedCategoryID)
	if errors.Is(err, wawi.NoCategory) {
		label := widget.NewLabel("Bitte eine Kategorie auswählen um nach Artikeln zu suchen")

		content := container.NewVBox(label)
		dialog.ShowCustom("Hinweis", "Schließen", content, w)

		return
	} else if err != nil {
		dialog.ShowError(err, w)
		return
	}

	for _, item := range items {
		combineCheck := widget.NewCheck("Zusammenfügen", func(checked bool) {
			if item.GuiItem.IsChild || item.GuiItem.IsFather {
				item.GuiItem.Combine = checked
			}

			if checked {
				Selected = append(Selected, item)
			} else {
				Selected = helper.FindAndRemoveItem(item, Selected)
			}
		})

		if item.GuiItem.IsFather || item.GuiItem.IsChild {
			combineCheck.Disable()
		}

		row := container.NewHBox(
			truncatedLabelWithTooltip(item.GuiItem.SKU, MaxIdLength, canvas),
			truncatedLabelWithTooltip(item.GuiItem.Name, MaxNameLength, canvas),
			layout.NewSpacer(),
			createDisabledCheck("Vaterartikel", item.GuiItem.IsFather),
			createDisabledCheck("Kindartikel", item.GuiItem.IsChild),
			combineCheck,
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

	cb.OnChanged = func(bool) {
		cb.SetChecked(checked)
	}

	return cb
}
