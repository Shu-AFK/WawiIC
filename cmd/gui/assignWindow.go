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
	w.Resize(fyne.NewSize(800, 500))
	state := newAssignmentState(selected, variations, labels)
	ui := buildAssignmentUI(w, state)

	ui.assignBtn.OnTapped = func() { handleAssign(w, state, ui) }
	ui.unassignBtn.OnTapped = func() { handleUnassign(w, state, ui) }
	ui.cancelBtn.OnTapped = func() { handleCancel(w, state, ui) }
	ui.doneBtn.OnTapped = func() { handleDone(w, state, ui) }

	w.SetContent(ui.content)
	w.Show()
}

type assignmentState struct {
	selectedIndex            int
	selectedComboIndex       int
	selectedCombinationIndex int
	availableItems           []wawi_structs.WItem
	availableCombos          []valueCombo
	combinedItems            []gui_structs.Combination
	comboByID                map[string]valueCombo
	variations               map[string][]string
	labels                   map[string]string
	mergeImages              bool
	errorOnNoImages          bool
}

func newAssignmentState(selected []wawi_structs.WItem, variations map[string][]string, labels map[string]string) *assignmentState {
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
	comboByID := make(map[string]valueCombo, len(allCombos))
	for _, c := range allCombos {
		comboByID[strings.Join(c.ids, "|")] = c
	}

	return &assignmentState{
		selectedIndex:            -1,
		selectedComboIndex:       -1,
		selectedCombinationIndex: -1,
		availableItems:           append([]wawi_structs.WItem(nil), selected...),
		availableCombos:          append([]valueCombo(nil), allCombos...),
		comboByID:                comboByID,
		variations:               variations,
		labels:                   labels,
		mergeImages:              false,
		errorOnNoImages:          false,
	}
}

type assignmentUI struct {
	itemList, variationList, assignmentList    *widget.List
	assignBtn, unassignBtn, cancelBtn, doneBtn *widget.Button
	mergeCheck, errorCheck                     *widget.Check
	content                                    fyne.CanvasObject
}

func buildAssignmentUI(w fyne.Window, s *assignmentState) *assignmentUI {
	itemList := widget.NewList(
		itemListLengthFn(&s.availableItems),
		labelTemplate,
		itemListUpdateFn(&s.availableItems),
	)
	itemList.OnSelected = func(id widget.ListItemID) { s.selectedIndex = id }

	variationList := widget.NewList(
		variationListLengthFn(&s.availableCombos),
		labelTemplate,
		variationListUpdateFn(&s.availableCombos),
	)
	variationList.OnSelected = func(i widget.ListItemID) { s.selectedComboIndex = i }

	assignmentList := widget.NewList(
		assignmentListLengthFn(&s.combinedItems),
		labelTemplate,
		assignmentListUpdateFn(&s.combinedItems, s.labels),
	)
	assignmentList.OnSelected = func(id widget.ListItemID) { s.selectedCombinationIndex = id }

	assignBtn := widget.NewButton("Zusammenfügen", nil)
	unassignBtn := widget.NewButton("Rückgängig machen", nil)
	cancelBtn := widget.NewButton("Abbrechen", nil)
	doneBtn := widget.NewButton("Fertig", nil)

	mergeCheck := widget.NewCheck("Bilder zusammenfügen", func(val bool) {
		s.mergeImages = val
	})
	errorCheck := widget.NewCheck("Error wenn kein Bild vorhanden", func(val bool) {
		s.errorOnNoImages = val
	})

	left := container.NewBorder(widget.NewLabel("Verfügbare Artikel"), nil, nil, nil, container.NewVScroll(itemList))
	center := container.NewBorder(widget.NewLabel("Variationen"), nil, nil, nil, container.NewVScroll(variationList))
	right := container.NewBorder(widget.NewLabel("Kombinationen"), nil, nil, nil, container.NewVScroll(assignmentList))

	rightSplit := container.NewHSplit(center, right)
	rightSplit.SetOffset(0.5)
	mainSplit := container.NewHSplit(left, rightSplit)
	mainSplit.SetOffset(0.33)

	buttons := container.NewHBox(
		assignBtn,
		unassignBtn,
		cancelBtn,
		layout.NewSpacer(),
		mergeCheck,
		errorCheck,
		doneBtn,
	)

	content := container.NewVSplit(mainSplit, buttons)
	content.SetOffset(0.95)

	return &assignmentUI{
		itemList, variationList, assignmentList,
		assignBtn, unassignBtn, cancelBtn, doneBtn,
		mergeCheck, errorCheck,
		content,
	}
}

func handleAssign(w fyne.Window, s *assignmentState, ui *assignmentUI) {
	if s.selectedIndex < 0 || s.selectedComboIndex < 0 {
		dialog.ShowInformation("Achtung!", "Bitte wähle zuerst eine Variation und einen Artikel aus.", w)
		return
	}
	combo := s.availableCombos[s.selectedComboIndex]
	combinedID := strings.Join(combo.ids, "|")
	if _, ok := s.labels[combinedID]; !ok {
		s.labels[combinedID] = combo.label
	}
	s.availableCombos = append(s.availableCombos[:s.selectedComboIndex], s.availableCombos[s.selectedComboIndex+1:]...)
	ui.variationList.UnselectAll()
	s.selectedComboIndex = -1
	ui.variationList.Refresh()

	item := s.availableItems[s.selectedIndex]
	s.availableItems = append(s.availableItems[:s.selectedIndex], s.availableItems[s.selectedIndex+1:]...)
	ui.itemList.UnselectAll()
	s.selectedIndex = -1
	ui.itemList.Refresh()

	s.combinedItems = append(s.combinedItems, gui_structs.Combination{
		Item:        item,
		Label:       s.labels[combinedID],
		VariationID: combinedID,
		ParentID:    "",
		ParentIndex: -1,
	})
	ui.assignmentList.Refresh()
}

func handleUnassign(w fyne.Window, s *assignmentState, ui *assignmentUI) {
	if s.selectedCombinationIndex < 0 || s.selectedCombinationIndex >= len(s.combinedItems) {
		dialog.ShowInformation("Achtung!", "Bitte wähle zuerst eine Kombination aus, die rückgängig gemacht werden soll.", w)
		return
	}
	c := s.combinedItems[s.selectedCombinationIndex]
	s.combinedItems = append(s.combinedItems[:s.selectedCombinationIndex], s.combinedItems[s.selectedCombinationIndex+1:]...)
	ui.assignmentList.UnselectAll()
	s.selectedCombinationIndex = -1
	ui.assignmentList.Refresh()
	s.availableItems = append(s.availableItems, c.Item)
	ui.itemList.Refresh()
	if vc, ok := s.comboByID[c.VariationID]; ok {
		insertComboByOrder(&s.availableCombos, vc)
		ui.variationList.Refresh()
	}
}

func handleCancel(w fyne.Window, s *assignmentState, ui *assignmentUI) {
	for _, c := range s.combinedItems {
		s.availableItems = append(s.availableItems, c.Item)
		if vc, ok := s.comboByID[c.VariationID]; ok {
			insertComboByOrder(&s.availableCombos, vc)
		}
	}
	s.combinedItems = nil
	ui.itemList.Refresh()
	ui.assignmentList.UnselectAll()
	ui.assignmentList.Refresh()
	ui.variationList.Refresh()
	w.Close()
}

func handleDone(w fyne.Window, s *assignmentState, ui *assignmentUI) {
	if len(s.availableItems) != 0 {
		dialog.ShowInformation("Achtung!", "Bitte kombiniere zuerst alle Artikel..", w)
		return
	}

	entry := widget.NewEntry()
	d := dialog.NewForm(
		"SKU des Vaterartikels",
		"Speichern",
		"Abbrechen",
		[]*widget.FormItem{widget.NewFormItem("SKU", entry)},
		func(confirmed bool) {
			if !confirmed {
				return
			}
			sku := strings.TrimSpace(entry.Text)
			if sku == "" {
				dialog.ShowInformation("Ungültige Eingabe", "Bitte eine SKU eingeben.", w)
				return
			}
			proceedAssign := func() {
				spinner := widget.NewProgressBarInfinite()
				waitDlg := dialog.NewCustomWithoutButtons(
					"Bitte warten",
					container.NewVBox(widget.NewLabel("Vorgang läuft…"), spinner),
					w,
				)
				waitDlg.Show()
				go func() {
					SKU, err := wawi.HandleAssignDone(
						s.combinedItems,
						s.variations,
						s.labels,
						sku,
						s.mergeImages,
						s.errorOnNoImages,
					)
					fyne.Do(func() {
						waitDlg.Hide()
						if err != nil {
							dialog.ShowError(fmt.Errorf("etwas lief schief: %w", err), w)
							fmt.Println(fmt.Sprintf("etwas lief schief: %s", err))
							return
						}
						FatherSKU = SKU
						w.Close()
					})
				}()
			}
			dialog.ShowConfirm(
				"SKU prüfen?",
				"Möchtest du prüfen, ob die SKU bereits existiert?",
				func(shouldCheck bool) {
					if !shouldCheck {
						proceedAssign()
						return
					}
					checkSpinner := widget.NewProgressBarInfinite()
					checkDlg := dialog.NewCustomWithoutButtons(
						"Bitte warten",
						container.NewVBox(widget.NewLabel("Prüfe SKU…"), checkSpinner),
						w,
					)
					checkDlg.Show()
					go func() {
						existsAlready, err := wawi.CheckIfSKUExists(sku)
						fyne.Do(func() {
							checkDlg.Hide()
							if err != nil {
								dialog.ShowError(fmt.Errorf("etwas lief schief: %w", err), w)
								fmt.Println(fmt.Sprintf("etwas lief schief: %s", err))
								return
							}
							if existsAlready {
								dialog.ShowInformation("Achtung!", "Diese SKU existiert bereits.", w)
								fmt.Println(fmt.Sprintf("Diese SKU existiert bereits: %s", sku))
								return
							}
							proceedAssign()
						})
					}()
				},
				w,
			)
		},
		w,
	)
	d.Show()
}

// list helper functions unchanged
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
