package gui

// TODO: Filter for child categories not working

import (
	"errors"
	"fmt"

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

var FatherSKU string

func createMainWidget(canvas fyne.Canvas, app fyne.App, w fyne.Window) fyne.CanvasObject {
	var prevSearch string

	searchbar := widget.NewEntry()
	searchbar.SetPlaceHolder("Artikel...")

	autoSearchCheck := widget.NewCheck("nach Zusammenfügen erneut suchen", nil)
	autoSearchCheck.SetChecked(true)

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

			if FatherSKU != "" {
				dialog.ShowInformation("Erfolg", fmt.Sprintf("Der Vather Artikel wurde erfolgreich erstellt. SKU %s\nBitte überprüfe alle informationen nochmal in JTL Wawi.", FatherSKU), w)
			}
			FatherSKU = ""

			if autoSearchCheck.Checked {
				searchbar.SetText(prevSearch)
				onSearch(prevSearch, rows, canvas, w)
			}
		})

		CombineWindow(combineW, app, Selected)
	})

	mergeButton.Importance = widget.LowImportance
	mergeButton.Resize(fyne.NewSize(80, 40))

	clearButton := widget.NewButton("Auswahl leeren", func() {
		Selected = Selected[:0]
		FatherSKU = ""

		for _, obj := range rows.Objects {
			if rowContainer, ok := obj.(*fyne.Container); ok {
				for _, child := range rowContainer.Objects {
					chk, ok := child.(*widget.Check)
					if !ok {
						continue
					}

					if dis, ok := child.(fyne.Disableable); ok && dis.Disabled() {
						continue
					}

					chk.SetChecked(false)
				}
			}
		}

		rows.Refresh()
	})

	buttonContainer := container.New(layout.NewHBoxLayout(), layout.NewSpacer(), clearButton, mergeButton)

	searchbar.OnSubmitted = func(query string) {
		prevSearch = query
		onSearch(query, rows, canvas, w)
	}

	content := container.NewBorder(
		container.NewVBox(
			searchbar,
			container.NewHBox(autoSearchCheck, layout.NewSpacer()),
		),
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

	spinner := widget.NewProgressBarInfinite()
	waitDlg := dialog.NewCustomWithoutButtons(
		"Bitte warten",
		container.NewVBox(
			widget.NewLabel("Artikel werden gesucht…"),
			spinner,
		),
		w,
	)
	waitDlg.Show()

	go func() {
		var items []wawi_structs.WItem
		var err error

		if wawi.SearchMode == "category" {
			items, err = wawi.GetItems(query, SelectedCategoryID, 0)
		} else if wawi.SearchMode == "supplier" {
			items, err = wawi.GetItems(query, "", SelectedSupplierID)
		} else if wawi.SearchMode == "none" {
			items, err = wawi.GetItems(query, "", 0)
		}

		fyne.Do(func() {
			defer waitDlg.Hide()

			if errors.Is(err, wawi.NoCategory) {
				label := widget.NewLabel("Bitte eine Kategorie auswählen um nach Artikeln zu suchen")
				content := container.NewVBox(label)
				dialog.ShowCustom("Hinweis", "Schließen", content, w)
				return
			} else if errors.Is(err, wawi.NoSupplier) {
				label := widget.NewLabel("Bitte einen Hersteller ausählen um nach Artikeln zu suchen")
				content := container.NewVBox(label)
				dialog.ShowCustom("Hinweis", "Schließen", content, w)
				return
			} else if err != nil {
				dialog.ShowError(err, w)
				return
			}

			for _, it := range items {
				item := it

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
		})
	}()
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
