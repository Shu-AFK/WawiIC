package helper

import "WawiIC/cmd/wawi/wawi_structs"

func FindAndRemoveItem(item wawi_structs.WItem, selected []wawi_structs.WItem) []wawi_structs.WItem {
	for i, val := range selected {
		if val.GuiItem.SKU == item.GuiItem.SKU {
			return append(selected[:i], selected[i+1:]...)
		}
	}

	return selected
}
