package gui

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/Shu-AFK/WawiIC/cmd/gui/gui_structs"
	"github.com/Shu-AFK/WawiIC/cmd/wawi/wawi_structs"
)

func createAssignmentWindow(w fyne.Window, selected []wawi_structs.WItem, variations map[string][]string, labels map[string]string) {
	selectedIndex := -1
	selectedCombinationIndex := -1

	var combinedItems []gui_structs.Combination
	availableItems := append([]wawi_structs.WItem(nil), selected...)

	w.Resize(fyne.NewSize(800, 500))

	itemList := widget.NewList(
		func() int {
			return len(availableItems)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("template")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			text := fmt.Sprintf("%s (SKU: %s)", availableItems[id].GuiItem.Name, availableItems[id].GuiItem.SKU)
			obj.(*widget.Label).SetText(text)
		},
	)
	itemList.OnSelected = func(id widget.ListItemID) {
		selectedIndex = id
	}
	itemScroll := container.NewVScroll(itemList)

	const visualRoot = ""
	const actualRoot = "root"

	if _, ok := variations[visualRoot]; !ok {
		variations[visualRoot] = []string{actualRoot}
	}
	if _, ok := variations[actualRoot]; !ok {
		variations[actualRoot] = []string{}
	}
	if _, ok := labels[actualRoot]; !ok {
		labels[actualRoot] = "Variationen"
	}

	branchIDs := variations[actualRoot]

	getValues := func(branchID string) []string {
		return variations[branchID]
	}

	type valueCombo struct {
		ids       []string
		label     string
		origIndex int
	}

	var allCombos []valueCombo
	var buildCombos func(idx int, picked []string)

	buildCombos = func(idx int, picked []string) {
		if idx == len(branchIDs) {
			parts := make([]string, len(picked))
			for i, id := range picked {
				if name, ok := labels[id]; ok {
					parts[i] = name
				} else {
					parts[i] = id
				}
			}
			allCombos = append(allCombos, valueCombo{
				ids:       append([]string(nil), picked...),
				label:     strings.Join(parts, " "),
				origIndex: len(allCombos),
			})
			return
		}
		values := getValues(branchIDs[idx])
		if len(values) == 0 {
			buildCombos(idx+1, picked)
			return
		}
		for _, v := range values {
			buildCombos(idx+1, append(picked, v))
		}
	}

	allCombos = allCombos[:0]
	buildCombos(0, nil)

	availableCombos := append([]valueCombo(nil), allCombos...)

	comboByID := make(map[string]valueCombo, len(allCombos))
	for _, c := range allCombos {
		cid := strings.Join(c.ids, "|")
		comboByID[cid] = c
	}

	selectedComboIndex := -1
	variationList := widget.NewList(
		func() int { return len(availableCombos) },
		func() fyne.CanvasObject { return widget.NewLabel("template") },
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(availableCombos[i].label)
		},
	)
	variationList.OnSelected = func(i widget.ListItemID) {
		selectedComboIndex = i
	}
	variationScroll := container.NewVScroll(variationList)

	assignmentList := widget.NewList(
		func() int {
			return len(combinedItems)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("template")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			c := combinedItems[id]
			name := labels[c.VariationID]
			if name == "" {
				name = c.VariationID
			}
			text := fmt.Sprintf("%s (SKU: %s): %s", c.Item.GuiItem.Name, c.Item.GuiItem.SKU, name)
			obj.(*widget.Label).SetText(text)
		},
	)
	assignmentList.OnSelected = func(id widget.ListItemID) {
		selectedCombinationIndex = id
	}
	assignmentScroll := container.NewVScroll(assignmentList)

	containsComboID := func(combinedID string) bool {
		for _, c := range availableCombos {
			if strings.Join(c.ids, "|") == combinedID {
				return true
			}
		}
		return false
	}

	insertComboByOrder := func(vc valueCombo) {
		if containsComboID(strings.Join(vc.ids, "|")) {
			return
		}
		pos := len(availableCombos)
		for i, c := range availableCombos {
			if vc.origIndex < c.origIndex {
				pos = i
				break
			}
		}
		availableCombos = append(availableCombos, valueCombo{})
		copy(availableCombos[pos+1:], availableCombos[pos:])
		availableCombos[pos] = vc
	}

	assignBtn := widget.NewButton("Zusammenfügen", func() {
		if selectedIndex < 0 || selectedComboIndex < 0 {
			dialog.ShowInformation("Achtung!", "Bitte wähle zuerst eine Variation und einen Artikel aus.", w)
			return
		}

		combo := availableCombos[selectedComboIndex]
		combinedID := strings.Join(combo.ids, "|")
		if _, ok := labels[combinedID]; !ok {
			labels[combinedID] = combo.label
		}

		availableCombos = append(availableCombos[:selectedComboIndex], availableCombos[selectedComboIndex+1:]...)
		variationList.UnselectAll()
		selectedComboIndex = -1
		variationList.Refresh()

		item := availableItems[selectedIndex]
		availableItems = append(availableItems[:selectedIndex], availableItems[selectedIndex+1:]...)
		itemList.UnselectAll()
		selectedIndex = -1
		itemList.Refresh()

		combinedItems = append(combinedItems, gui_structs.Combination{
			Item:        item,
			VariationID: combinedID,
			ParentID:    "",
			ParentIndex: -1,
		})

		assignmentList.Refresh()
	})

	unassignBtn := widget.NewButton("Rückgängig machen", func() {
		if selectedCombinationIndex < 0 || selectedCombinationIndex >= len(combinedItems) {
			dialog.ShowInformation("Achtung!", "Bitte wähle zuerst eine Kombination aus, die rückgängig gemacht werden soll.", w)
			return
		}

		c := combinedItems[selectedCombinationIndex]
		combinedItems = append(combinedItems[:selectedCombinationIndex], combinedItems[selectedCombinationIndex+1:]...)
		assignmentList.UnselectAll()
		selectedCombinationIndex = -1
		assignmentList.Refresh()

		availableItems = append(availableItems, c.Item)
		itemList.Refresh()

		if vc, ok := comboByID[c.VariationID]; ok {
			insertComboByOrder(vc)
			variationList.Refresh()
		}
	})

	cancelBtn := widget.NewButton("Abbrechen", func() {
		// Restore all items
		for _, c := range combinedItems {
			availableItems = append(availableItems, c.Item)
			// Restore all removed combos
			if vc, ok := comboByID[c.VariationID]; ok {
				insertComboByOrder(vc)
			}
		}

		combinedItems = nil
		itemList.Refresh()
		assignmentList.UnselectAll()
		assignmentList.Refresh()
		variationList.Refresh()

		w.Close()
	})

	doneBtn := widget.NewButton("Fertig", func() {
		if len(availableItems) != 0 {
			dialog.ShowInformation("Achtung!", "Bitte kombiniere zuerst alle Artikel..", w)
			return
		}
		if selectedCombinationIndex < 0 {
			dialog.ShowInformation("Achtung!", "Bitte wähle einen Artikel aus, von welchem das Bild an erster Stelle sein soll.", w)
			return
		}

		dialog.ShowConfirm("Achtung!", "Ist die ausgewählte Artikel/Variations-Kombination das Bild, welches als erstes beim Vaterartikel erscheinen soll?", func(b bool) {
			if !b {
				return
			}
			w.Close()
		}, w)
	})

	leftPanel := container.NewBorder(
		widget.NewLabel("Verfügbare Artikel"),
		nil,
		nil,
		nil,
		itemScroll,
	)

	rightPanel := container.NewBorder(
		widget.NewLabel("Kombinationen"),
		nil,
		nil,
		nil,
		assignmentScroll,
	)

	centerPanel := container.NewBorder(
		widget.NewLabel("Variationen"),
		nil,
		nil,
		nil,
		variationScroll,
	)

	mid := container.NewHSplit(centerPanel, rightPanel)
	mid.SetOffset(0.5)

	up := container.NewHSplit(leftPanel, mid)
	up.SetOffset(0.33)

	content := container.NewVSplit(up, container.NewHBox(assignBtn, unassignBtn, cancelBtn, layout.NewSpacer(), doneBtn))
	content.SetOffset(0.95)

	w.SetContent(content)
	w.Show()
}
