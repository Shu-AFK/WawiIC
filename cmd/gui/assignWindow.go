package gui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/Shu-AFK/WawiIC/cmd/wawi/wawi_structs"
)

func createAssignmentWindow(w fyne.Window, selected []wawi_structs.WItem, variations map[string][]string, labels map[string]string) {
	var selectedIDTree string
	var selectedIDList string

	w.Resize(fyne.NewSize(800, 500))

	itemList := widget.NewList(
		func() int {
			return len(selected)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("template")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			text := fmt.Sprintf("%s (SKU: %s)", selected[id].GuiItem.Name, selected[id].GuiItem.SKU)
			obj.(*widget.Label).SetText(text)
		},
	)
	itemScroll := container.NewVScroll(itemList)

	leftPanel := container.NewBorder(
		widget.NewLabel("Verfügbare Artikel"),
		nil,
		nil,
		nil,
		itemScroll,
	)

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
	treeScroll := container.NewVScroll(tree)

	assignBtn := widget.NewButton("Zusammenfügen", func() {
		if selectedIDList == "" || selectedIDTree == "" {
			dialog.ShowInformation("Achtung!", "Bitte wähle zuerst eine Variation und einen Artikel aus.", w)
			return
		}

		// TODO: Finish implementation
	})

	unassignBtn := widget.NewButton("Rückgängig machen", func() {
		// TODO: Finish implementation
	})

	rightPanel := container.NewBorder(
		widget.NewLabel("Variationen"),
		nil,
		nil,
		nil,
		treeScroll,
	)

	up := container.NewHSplit(leftPanel, rightPanel)
	up.SetOffset(0.45)

	content := container.NewVSplit(up, container.NewHBox(assignBtn, unassignBtn))
	content.SetOffset(0.95)

	w.SetContent(content)
	w.Show()
}
