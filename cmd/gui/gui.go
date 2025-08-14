package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
)

func RunGUI() {
	WawiIC := app.New()
	w := WawiIC.NewWindow("WawiIC")

	// TODO: Make dynamic using JTL WAWI api
	data := map[string][]string{
		"":      {"Shop1", "Shop2", "Shop3"}, // root nodes
		"Shop1": {"Electronics", "Clothing", "Books"},
		"Shop2": {"Groceries", "Toys"},
		"Shop3": {"Home", "Garden", "Tools", "Sports", "Stationery"},
	}

	split := container.NewHSplit(createSidebarTree(data), createMainWidget(w.Canvas()))
	split.Offset = 0.15

	w.CenterOnScreen()
	w.SetMaster()
	w.SetPadded(true)
	w.SetContent(split)
	w.Resize(fyne.NewSize(800, 600))

	w.ShowAndRun()
}
