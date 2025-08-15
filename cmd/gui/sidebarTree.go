package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

var SelectedCategoryID string

// categories: map[parentID][]childID
// labels: map[ID]Name
func createSidebarTree(categories map[string][]string, labels map[string]string) *container.Scroll {
	getChildIDs := func(uid widget.TreeNodeID) []widget.TreeNodeID {
		children := categories[uid]
		out := make([]widget.TreeNodeID, len(children))
		for i, child := range children {
			out[i] = widget.TreeNodeID(child)
		}
		return out
	}

	isBranch := func(uid widget.TreeNodeID) bool {
		_, exists := categories[uid]
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

	tree.OnSelected = func(id widget.TreeNodeID) {
		if id != "Kategorien" {
			SelectedCategoryID = id
		}
	}

	tree.Root = "root"

	return container.NewVScroll(tree)
}
