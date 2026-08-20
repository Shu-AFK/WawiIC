package wawi

import (
	"encoding/json"
	"testing"

	"github.com/Shu-AFK/WawiIC/cmd/wawi/wawi_structs"
)

// The API reports component quantities as decimals, which used to abort a whole
// page of results.
func TestDecodeItemsAcceptsDecimalComponentQuantity(t *testing.T) {
	raw := []json.RawMessage{
		json.RawMessage(`{"Id":1,"SKU":"A","Components":[{"ItemId":7,"Quantity":20.0,"SortNumber":1}]}`),
		json.RawMessage(`{"Id":2,"SKU":"B","Components":[{"ItemId":8,"Quantity":0.5,"SortNumber":1}]}`),
	}

	items := decodeItems(raw)
	if len(items) != 2 {
		t.Fatalf("decodeItems() returned %d items, want 2", len(items))
	}
	if got := items[0].Components[0].Quantity; got != 20 {
		t.Errorf("Quantity = %v, want 20", got)
	}
	if got := items[1].Components[0].Quantity; got != 0.5 {
		t.Errorf("Quantity = %v, want 0.5", got)
	}
}

// A single item the model cannot represent must not cost the rest of the page.
func TestDecodeItemsSkipsUndecodableItem(t *testing.T) {
	raw := []json.RawMessage{
		json.RawMessage(`{"Id":1,"SKU":"GOOD"}`),
		json.RawMessage(`{"Id":2,"SKU":"BAD","ManufacturerId":"not-a-number"}`),
		json.RawMessage(`{"Id":3,"SKU":"ALSO-GOOD"}`),
	}

	items := decodeItems(raw)
	if len(items) != 2 {
		t.Fatalf("decodeItems() returned %d items, want 2", len(items))
	}
	if items[0].SKU != "GOOD" || items[1].SKU != "ALSO-GOOD" {
		t.Errorf("decodeItems() = %s, %s", items[0].SKU, items[1].SKU)
	}
}

func TestItemPriceDataDecodesDecimals(t *testing.T) {
	var item wawi_structs.GetItem
	if err := json.Unmarshal([]byte(`{"Id":1,"ItemPriceData":{"SalesPriceNet":12.34},"Quantities":{"MinimumOrderQuantity":1.5},"Weights":{"ItemWeigth":2.25}}`), &item); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if *item.ItemPriceData.SalesPriceNet != 12.34 {
		t.Errorf("SalesPriceNet = %v", *item.ItemPriceData.SalesPriceNet)
	}
	if item.Weights.ItemWeight != 2.25 {
		t.Errorf("ItemWeight = %v", item.Weights.ItemWeight)
	}
}
