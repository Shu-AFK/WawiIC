package gui

import (
	"github.com/Shu-AFK/WawiIC/assets"
	"github.com/Shu-AFK/WawiIC/cmd/wawi"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
)

func RunGUI() {
	WawiIC := app.New()
	iconRes := fyne.NewStaticResource("WawiIC.png", assets.Icon)
	WawiIC.SetIcon(iconRes)

	w := WawiIC.NewWindow("WawiIC")

	tree, labels, err := wawi.GetCategories(50)
	if err != nil {
		dialog.ShowError(err, w)
	}

	split := container.NewHSplit(createSidebarTree(tree, labels), createMainWidget(w.Canvas(), WawiIC, w))
	split.Offset = 0.15

	w.CenterOnScreen()
	w.SetMaster()
	w.SetPadded(true)
	w.SetContent(split)
	w.Resize(fyne.NewSize(800, 600))

	w.ShowAndRun()
}
