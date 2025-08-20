package gui

import (
	"fmt"

	"github.com/Shu-AFK/WawiIC/cmd/wawi/wawi_structs"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func CombineWindow(w fyne.Window, app fyne.App, selected []wawi_structs.WItem) {
	variations := map[string][]string{}
	labels := map[string]string{}
	id := 1
	selectedVariationID := ""

	const visualRoot = ""
	const actualRoot = "root"

	variations[visualRoot] = []string{actualRoot}
	variations[actualRoot] = []string{}
	labels[actualRoot] = "Variationen"

	getChildIDs := func(uid widget.TreeNodeID) []widget.TreeNodeID {
		children := variations[uid]
		out := make([]widget.TreeNodeID, len(children))
		for i, child := range children {
			out[i] = child
		}
		return out
	}

	isBranch := func(uid widget.TreeNodeID) bool {
		_, exists := variations[uid]
		return exists
	}

	create := func(branch bool) fyne.CanvasObject {
		return widget.NewLabel("")
	}

	update := func(uid widget.TreeNodeID, branch bool, obj fyne.CanvasObject) {
		if name, ok := labels[uid]; ok {
			obj.(*widget.Label).SetText(name)
		} else {
			obj.(*widget.Label).SetText(uid)
		}
	}

	tree := widget.NewTree(getChildIDs, isBranch, create, update)
	tree.OpenBranch(visualRoot)
	tree.OpenBranch(actualRoot)
	tree.OnSelected = func(id widget.TreeNodeID) {
		selectedVariationID = id
	}

	btnVariation := widget.NewButton("Variation anlegen", func() {
		entry := widget.NewEntry()
		entry.PlaceHolder = "Neue Variation"
		entry.Resize(fyne.NewSize(200, 100))
		popup := widget.NewPopUp(entry, w.Canvas())
		popup.Resize(fyne.NewSize(200, 100))
		entry.OnSubmitted = func(s string) {
			variationID := fmt.Sprintf("%d", id)
			id++

			variations[variationID] = []string{}
			labels[variationID] = s

			variations[actualRoot] = append(variations[actualRoot], variationID)

			tree.Refresh()
			tree.Select(variationID)
			popup.Hide()
		}
		popup.Show()
		w.Canvas().Focus(entry)
	})

	btnWert := widget.NewButton("Wert anlegen", func() {
		if selectedVariationID == "" {
			dialog.ShowInformation("Achtung!", "Bitte wähle zuerst eine Variation aus", w)
			return
		}

		if _, exists := variations[selectedVariationID]; !exists {
			dialog.ShowInformation("Achtung!",
				"Kann keinen Wert zu einem anderen Wert hinzufügen.\nBitte wähle eine Variation aus.",
				w)
			return
		}

		entry := widget.NewEntry()
		entry.PlaceHolder = "Neuer Wert"
		entry.Resize(fyne.NewSize(200, 100))
		popup := widget.NewPopUp(entry, w.Canvas())
		popup.Resize(fyne.NewSize(200, 100))
		entry.OnSubmitted = func(s string) {
			wertID := fmt.Sprintf("%d", id)
			id++

			labels[wertID] = s
			variations[selectedVariationID] = append(variations[selectedVariationID], wertID)

			tree.Refresh()
			tree.OpenBranch(selectedVariationID)
			popup.Hide()
		}
		popup.Show()
		w.Canvas().Focus(entry)
	})

	btnEdit := widget.NewButton("Name bearbeiten", func() {
		if selectedVariationID == "" || selectedVariationID == visualRoot || selectedVariationID == actualRoot {
			dialog.ShowInformation("Achtung!", "Bitte wähle eine Variation oder einen Wert zum Bearbeiten aus", w)
			return
		}

		var popup *widget.PopUp
		entry := widget.NewEntry()
		entry.SetText(labels[selectedVariationID])
		entry.Resize(fyne.NewSize(200, 40))

		content := container.NewVBox(
			entry,
			widget.NewButton("Speichern", func() {
				newName := entry.Text
				if newName == "" {
					dialog.ShowInformation("Achtung!", "Der Name darf nicht leer sein", w)
					return
				}
				labels[selectedVariationID] = newName
				tree.Refresh()
				popup.Hide()
			}),
			widget.NewButton("Abbrechen", func() {
				popup.Hide()
			}),
		)
		popup = widget.NewPopUp(content, w.Canvas())

		popup.Resize(fyne.NewSize(250, 120))
		entry.OnSubmitted = func(s string) {
			if s == "" {
				dialog.ShowInformation("Achtung!", "Der Name darf nicht leer sein", w)
				return
			}
			labels[selectedVariationID] = s
			tree.Refresh()
			popup.Hide()
		}
		popup.Show()
		w.Canvas().Focus(entry)
	})

	btnDel := widget.NewButton("Löschen", func() {
		if selectedVariationID == "" || selectedVariationID == visualRoot || selectedVariationID == actualRoot {
			dialog.ShowInformation("Achtung!", "Bitte wähle eine Variation oder einen Wert zum Löschen aus", w)
			return
		}

		dialog.ShowConfirm("Löschen bestätigen",
			fmt.Sprintf("Soll '%s' wirklich gelöscht werden?", labels[selectedVariationID]),
			func(confirmed bool) {
				if !confirmed {
					return
				}

				for parentID, children := range variations {
					for i, childID := range children {
						if childID == selectedVariationID {
							variations[parentID] = append(children[:i], children[i+1:]...)
							break
						}
					}
				}

				if children, exists := variations[selectedVariationID]; exists {
					for _, childID := range children {
						delete(labels, childID)
						delete(variations, childID)
					}
				}

				delete(labels, selectedVariationID)
				delete(variations, selectedVariationID)

				selectedVariationID = ""

				tree.Refresh()
				tree.UnselectAll()
			},
			w,
		)
	})

	btnCancel := widget.NewButton("Abbrechen", func() {
		w.Close()
	})

	btnFinished := widget.NewButton("Fertig", func() {
		if len(labels) <= len(selected) {
			dialog.ShowInformation("Achtung!", "Nicht genug Variationen für die ausgewählten Artikel!", w)
			return
		}

		newW := app.NewWindow("Artikel zuordnen")
		createAssignmentWindow(newW, selected, variations, labels)
		w.Close()
	})

	scroll := container.NewVScroll(tree)
	scroll.SetMinSize(fyne.NewSize(400, 500))

	content := container.NewVBox(
		scroll,
		container.NewHBox(btnVariation, btnWert),
		container.NewHBox(btnEdit, btnDel),
		container.NewHBox(btnCancel, btnFinished),
	)

	w.SetContent(content)
	w.Resize(fyne.NewSize(500, 600))
	w.Show()
}
