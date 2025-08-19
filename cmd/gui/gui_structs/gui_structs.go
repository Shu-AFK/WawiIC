package gui_structs

import "github.com/Shu-AFK/WawiIC/cmd/wawi/wawi_structs"

type Combination struct {
	Item        wawi_structs.WItem
	Label       string
	VariationID string
	ParentID    string
	ParentIndex int
}
