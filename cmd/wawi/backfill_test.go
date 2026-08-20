package wawi

import (
	"testing"

	"github.com/Shu-AFK/WawiIC/cmd/wawi/wawi_structs"
)

func item(id int, sku string, price *float64) wawi_structs.GetItem {
	return wawi_structs.GetItem{
		ID: id,
		Item: wawi_structs.Item{
			SKU:           sku,
			ItemPriceData: wawi_structs.ItemPriceData{SalesPriceNet: price},
		},
	}
}

func price(v float64) *float64 { return &v }

func TestPickPriceSource(t *testing.T) {
	tests := []struct {
		name     string
		parent   wawi_structs.GetItem
		children []wawi_structs.GetItem
		wantSKU  string
	}{
		{
			name:   "matches the child the parent price came from",
			parent: item(1, "P", price(10)),
			children: []wawi_structs.GetItem{
				item(2, "A", price(20)),
				item(3, "B", price(10)),
				item(4, "C", price(15)),
			},
			wantSKU: "B",
		},
		{
			name:   "prefers the exact match over a cheaper child",
			parent: item(1, "P", price(10)),
			children: []wawi_structs.GetItem{
				item(2, "A", price(5)),
				item(3, "B", price(10)),
			},
			wantSKU: "B",
		},
		{
			name:   "falls back to the cheapest when no price matches",
			parent: item(1, "P", price(99)),
			children: []wawi_structs.GetItem{
				item(2, "A", price(20)),
				item(3, "B", price(8)),
			},
			wantSKU: "B",
		},
		{
			name:   "falls back when the parent has no price",
			parent: item(1, "P", nil),
			children: []wawi_structs.GetItem{
				item(2, "A", price(20)),
				item(3, "B", price(8)),
			},
			wantSKU: "B",
		},
		{
			name:   "ignores children without a price",
			parent: item(1, "P", nil),
			children: []wawi_structs.GetItem{
				item(2, "A", nil),
				item(3, "B", price(8)),
			},
			wantSKU: "B",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.children[pickPriceSource(tt.parent, tt.children)]
			if got.SKU != tt.wantSKU {
				t.Errorf("pickPriceSource() = %s, want %s", got.SKU, tt.wantSKU)
			}
		})
	}
}

func TestFindCheapestItemWithoutPrices(t *testing.T) {
	items := []wawi_structs.GetItem{item(1, "A", nil), item(2, "B", nil)}
	if got := findCheapestItem(items); got != 0 {
		t.Errorf("findCheapestItem() = %d, want 0", got)
	}
}
