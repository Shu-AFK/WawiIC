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
	"github.com/Shu-AFK/WawiIC/cmd/wawi"
	"github.com/Shu-AFK/WawiIC/cmd/wawi/wawi_structs"
)

type valueCombo struct {
	ids       []string
	label     string
	origIndex int
}

func createAssignmentWindow(w fyne.Window, selected []wawi_structs.WItem, variations map[string][]string, labels map[string]string) {
	selectedIndex := -1
	selectedCombinationIndex := -1

	var combinedItems []gui_structs.Combination
	availableItems := append([]wawi_structs.WItem(nil), selected...)

	w.Resize(fyne.NewSize(800, 500))

	itemList := widget.NewList(
		itemListLengthFn(&availableItems),
		labelTemplate,
		itemListUpdateFn(&availableItems),
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

	allCombos := buildAllCombos(variations, labels, branchIDs)

	availableCombos := append([]valueCombo(nil), allCombos...)

	comboByID := make(map[string]valueCombo, len(allCombos))
	for _, c := range allCombos {
		cid := strings.Join(c.ids, "|")
		comboByID[cid] = c
	}

	selectedComboIndex := -1
	variationList := widget.NewList(
		variationListLengthFn(&availableCombos),
		labelTemplate,
		variationListUpdateFn(&availableCombos),
	)
	variationList.OnSelected = func(i widget.ListItemID) {
		selectedComboIndex = i
	}
	variationScroll := container.NewVScroll(variationList)

	assignmentList := widget.NewList(
		assignmentListLengthFn(&combinedItems),
		labelTemplate,
		assignmentListUpdateFn(&combinedItems, labels),
	)
	assignmentList.OnSelected = func(id widget.ListItemID) {
		selectedCombinationIndex = id
	}
	assignmentScroll := container.NewVScroll(assignmentList)

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
			Label:       labels[combinedID],
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
			insertComboByOrder(&availableCombos, vc)
			variationList.Refresh()
		}
	})

	cancelBtn := widget.NewButton("Abbrechen", func() {
		for _, c := range combinedItems {
			availableItems = append(availableItems, c.Item)
			if vc, ok := comboByID[c.VariationID]; ok {
				insertComboByOrder(&availableCombos, vc)
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

			err := wawi.HandleAssignDone(combinedItems, selectedCombinationIndex, variations, labels)
			if err != nil {
				dialog.ShowError(fmt.Errorf("etwas lief schief: %w", err), w)
				fmt.Println(fmt.Sprintf("etwas lief schief: %s", err))
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

	right := container.NewHSplit(centerPanel, rightPanel)
	right.SetOffset(0.5)

	up := container.NewHSplit(leftPanel, right)
	up.SetOffset(0.33)

	content := container.NewVSplit(up, container.NewHBox(assignBtn, unassignBtn, cancelBtn, layout.NewSpacer(), doneBtn))
	content.SetOffset(0.95)

	w.SetContent(content)
	w.Show()
}

func itemListLengthFn(items *[]wawi_structs.WItem) func() int {
	return func() int { return len(*items) }
}

func labelTemplate() fyne.CanvasObject { return widget.NewLabel("template") }

func itemListUpdateFn(items *[]wawi_structs.WItem) func(id widget.ListItemID, obj fyne.CanvasObject) {
	return func(id widget.ListItemID, obj fyne.CanvasObject) {
		text := fmt.Sprintf("%s (SKU: %s)", (*items)[id].GuiItem.Name, (*items)[id].GuiItem.SKU)
		obj.(*widget.Label).SetText(text)
	}
}

func variationListLengthFn(combos *[]valueCombo) func() int {
	return func() int { return len(*combos) }
}
func variationListUpdateFn(combos *[]valueCombo) func(id widget.ListItemID, obj fyne.CanvasObject) {
	return func(i widget.ListItemID, o fyne.CanvasObject) {
		o.(*widget.Label).SetText((*combos)[i].label)
	}
}

func assignmentListLengthFn(items *[]gui_structs.Combination) func() int {
	return func() int { return len(*items) }
}
func assignmentListUpdateFn(items *[]gui_structs.Combination, labels map[string]string) func(id widget.ListItemID, obj fyne.CanvasObject) {
	return func(id widget.ListItemID, obj fyne.CanvasObject) {
		c := (*items)[id]
		name := labels[c.VariationID]
		if name == "" {
			name = c.VariationID
		}
		text := fmt.Sprintf("%s (SKU: %s): %s", c.Item.GuiItem.Name, c.Item.GuiItem.SKU, name)
		obj.(*widget.Label).SetText(text)
	}
}

func buildAllCombos(variations map[string][]string, labels map[string]string, branchIDs []string) []valueCombo {
	var all []valueCombo
	var rec func(idx int, picked []string)
	rec = func(idx int, picked []string) {
		if idx == len(branchIDs) {
			parts := make([]string, len(picked))
			for i, id := range picked {
				if name, ok := labels[id]; ok {
					parts[i] = name
				} else {
					parts[i] = id
				}
			}
			all = append(all, valueCombo{
				ids:       append([]string(nil), picked...),
				label:     strings.Join(parts, " "),
				origIndex: len(all),
			})
			return
		}
		values := variations[branchIDs[idx]]
		if len(values) == 0 {
			rec(idx+1, picked)
			return
		}
		for _, v := range values {
			rec(idx+1, append(picked, v))
		}
	}
	rec(0, nil)
	return all
}

func containsComboID(availableCombos []valueCombo, combinedID string) bool {
	for _, c := range availableCombos {
		if strings.Join(c.ids, "|") == combinedID {
			return true
		}
	}
	return false
}

func insertComboByOrder(availableCombos *[]valueCombo, vc valueCombo) {
	if containsComboID(*availableCombos, strings.Join(vc.ids, "|")) {
		return
	}
	pos := len(*availableCombos)
	for i, c := range *availableCombos {
		if vc.origIndex < c.origIndex {
			pos = i
			break
		}
	}
	*availableCombos = append(*availableCombos, valueCombo{})
	copy((*availableCombos)[pos+1:], (*availableCombos)[pos:])
	(*availableCombos)[pos] = vc
}
