package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func createSidebarTree(categories map[string][]string) *container.Scroll {
	getChildIDs := func(uid widget.TreeNodeID) []widget.TreeNodeID {
		children := categories[string(uid)]
		out := make([]widget.TreeNodeID, len(children))

		for i, child := range children {
			out[i] = widget.TreeNodeID(child)
		}

		return out
	}

	isBrance := func(uid widget.TreeNodeID) bool {
		_, yes := categories[uid]
		return yes
	}

	create := func(branch bool) fyne.CanvasObject {
		return widget.NewLabel("")
	}

	update := func(uid widget.TreeNodeID, branch bool, obj fyne.CanvasObject) {
		obj.(*widget.Label).SetText(uid)
	}

	tree := widget.NewTree(getChildIDs, isBrance, create, update)
	return container.NewVScroll(tree)
}
