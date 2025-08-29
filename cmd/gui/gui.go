package gui

import (
	"fmt"

	"github.com/Shu-AFK/WawiIC/assets"
	"github.com/Shu-AFK/WawiIC/cmd/wawi"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
)

func RunGUI() {
	fmt.Println("Starting GUI...")
	WawiIC := app.New()
	fmt.Println("Created app")
	iconRes := fyne.NewStaticResource("WawiIC.png", assets.Icon)
	WawiIC.SetIcon(iconRes)

	w := WawiIC.NewWindow("WawiIC")
	fmt.Println("Created window")

	var split *container.Split
	var isSearchmode bool = false
	if wawi.SearchMode == "category" {
		isSearchmode = true
		fmt.Println("Getting categories...")
		tree, labels, err := wawi.GetCategories(100)
		if err != nil {
			fmt.Println(err)
			dialog.ShowError(err, w)
			return
		}
		fmt.Println("Got categories")

		split = container.NewHSplit(createSidebarTree(tree, labels), createMainWidget(w.Canvas(), WawiIC, w))
	} else if wawi.SearchMode == "supplier" {
		isSearchmode = true
		fmt.Println("Getting suppliers...")
		list, err := createSupplierList()
		if err != nil {
			fmt.Println(err)
			dialog.ShowError(err, w)
			return
		}

		fmt.Println("Got suppliers")
		split = container.NewHSplit(list, createMainWidget(w.Canvas(), WawiIC, w))
	}

	w.CenterOnScreen()
	w.SetMaster()
	w.SetPadded(true)

	if isSearchmode {
		split.Offset = 0.2
		w.SetContent(split)
	} else {
		w.SetContent(createMainWidget(w.Canvas(), WawiIC, w))
	}

	w.Resize(fyne.NewSize(800, 600))

	fmt.Println("Created content")
	w.ShowAndRun()
}
