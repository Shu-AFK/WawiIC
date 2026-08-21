package wawi

import (
	"testing"

	"github.com/Shu-AFK/WawiIC/cmd/wawi/wawi_structs"
)

func price(v float64) *float64 { return &v }

func channelPrice(channel string, group, qty int, net *float64, percent *float64) wawi_structs.ItemSalesChannelPrice {
	return wawi_structs.ItemSalesChannelPrice{
		SalesChannelId:               channel,
		CustomerGroupId:              group,
		FromQuantity:                 qty,
		NetPrice:                     net,
		ReduceStandardPriceByPercent: percent,
	}
}

func TestIsCheaper(t *testing.T) {
	tests := []struct {
		name string
		a, b wawi_structs.ItemSalesChannelPrice
		want bool
	}{
		{
			name: "lower net price wins",
			a:    channelPrice("1-1-1", 1, 0, price(10), nil),
			b:    channelPrice("1-1-1", 1, 0, price(12), nil),
			want: true,
		},
		{
			name: "higher net price loses",
			a:    channelPrice("1-1-1", 1, 0, price(15), nil),
			b:    channelPrice("1-1-1", 1, 0, price(12), nil),
			want: false,
		},
		{
			name: "equal net price does not replace",
			a:    channelPrice("1-1-1", 1, 0, price(12), nil),
			b:    channelPrice("1-1-1", 1, 0, price(12), nil),
			want: false,
		},
		{
			name: "a net price beats a percentage",
			a:    channelPrice("1-1-1", 1, 0, price(99), nil),
			b:    channelPrice("1-1-1", 1, 0, nil, price(50)),
			want: true,
		},
		{
			name: "a percentage never beats a net price",
			a:    channelPrice("1-1-1", 1, 0, nil, price(50)),
			b:    channelPrice("1-1-1", 1, 0, price(99), nil),
			want: false,
		},
		{
			name: "the bigger discount wins among percentages",
			a:    channelPrice("1-1-1", 1, 0, nil, price(20)),
			b:    channelPrice("1-1-1", 1, 0, nil, price(10)),
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isCheaper(tt.a, tt.b); got != tt.want {
				t.Errorf("isCheaper() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNotAParent(t *testing.T) {
	tests := []struct {
		name   string
		item   wawi_structs.GetItem
		wantOK bool
	}{
		{
			name:   "parent with children",
			item:   wawi_structs.GetItem{Item: wawi_structs.Item{ChildItems: []int{2, 3}}},
			wantOK: true,
		},
		{
			name:   "plain item without children",
			item:   wawi_structs.GetItem{Item: wawi_structs.Item{}},
			wantOK: false,
		},
		{
			name: "child item is never a parent",
			item: wawi_structs.GetItem{Item: wawi_structs.Item{
				ParentItemID: 7,
				ChildItems:   []int{2},
			}},
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if ok := notAParent(tt.item) == ""; ok != tt.wantOK {
				t.Errorf("notAParent() ok = %v, want %v (%q)", ok, tt.wantOK, notAParent(tt.item))
			}
		})
	}
}

func TestFindCheapestItemWithoutPrices(t *testing.T) {
	items := []wawi_structs.GetItem{
		{ID: 1, Item: wawi_structs.Item{SKU: "A"}},
		{ID: 2, Item: wawi_structs.Item{SKU: "B"}},
	}
	if got := findCheapestItem(items); got != 0 {
		t.Errorf("findCheapestItem() = %d, want 0", got)
	}
}

func TestLowestPrice(t *testing.T) {
	items := []wawi_structs.GetItem{
		{ID: 1, Item: wawi_structs.Item{SKU: "A", ItemPriceData: wawi_structs.ItemPriceData{
			SalesPriceNet: price(20), EbayPrice: price(25), AmazonPrice: nil,
		}}},
		{ID: 2, Item: wawi_structs.Item{SKU: "B", ItemPriceData: wawi_structs.ItemPriceData{
			SalesPriceNet: price(10), EbayPrice: price(30), AmazonPrice: price(40),
		}}},
		{ID: 3, Item: wawi_structs.Item{SKU: "C", ItemPriceData: wawi_structs.ItemPriceData{
			SalesPriceNet: price(15), EbayPrice: nil, AmazonPrice: price(35),
		}}},
	}

	// The cheapest item overall is B, but A is cheaper on eBay - each shop has to
	// be minimised on its own.
	if got := LowestPrice(items, func(p wawi_structs.ItemPriceData) *float64 { return p.SalesPriceNet }); *got != 10 {
		t.Errorf("SalesPriceNet = %v, want 10", *got)
	}
	if got := LowestPrice(items, func(p wawi_structs.ItemPriceData) *float64 { return p.EbayPrice }); *got != 25 {
		t.Errorf("EbayPrice = %v, want 25", *got)
	}
	if got := LowestPrice(items, func(p wawi_structs.ItemPriceData) *float64 { return p.AmazonPrice }); *got != 35 {
		t.Errorf("AmazonPrice = %v, want 35", *got)
	}
}

func TestLowestPriceWithoutAnyValue(t *testing.T) {
	items := []wawi_structs.GetItem{{ID: 1, Item: wawi_structs.Item{SKU: "A"}}}
	if got := LowestPrice(items, func(p wawi_structs.ItemPriceData) *float64 { return p.EbayPrice }); got != nil {
		t.Errorf("EbayPrice = %v, want nil", got)
	}
}

func TestSortPrimaryChannelFirst(t *testing.T) {
	details := []BackfillPrice{
		{Price: channelPrice("1-1-3", 1, 0, price(10), nil)},
		{Price: channelPrice("2-2-1", 1, 0, price(11), nil)},
		{Price: channelPrice(PrimarySalesChannel, 1, 0, price(12), nil)},
		{Price: channelPrice("2-2-5", 1, 0, price(13), nil)},
		{Price: channelPrice(PrimarySalesChannel, 2, 0, price(14), nil)},
	}

	sortPrimaryChannelFirst(details)

	// Both rows of the primary channel come first, in their original order.
	if details[0].Price.SalesChannelId != PrimarySalesChannel || details[0].Price.CustomerGroupId != 1 {
		t.Errorf("first = %s/%d", details[0].Price.SalesChannelId, details[0].Price.CustomerGroupId)
	}
	if details[1].Price.SalesChannelId != PrimarySalesChannel || details[1].Price.CustomerGroupId != 2 {
		t.Errorf("second = %s/%d", details[1].Price.SalesChannelId, details[1].Price.CustomerGroupId)
	}

	// The rest keeps the order it had.
	rest := []string{details[2].Price.SalesChannelId, details[3].Price.SalesChannelId, details[4].Price.SalesChannelId}
	want := []string{"1-1-3", "2-2-1", "2-2-5"}
	for i := range want {
		if rest[i] != want[i] {
			t.Errorf("rest[%d] = %s, want %s", i, rest[i], want[i])
		}
	}
}

func TestSortPrimaryChannelFirstWithoutPrimary(t *testing.T) {
	details := []BackfillPrice{
		{Price: channelPrice("1-1-3", 1, 0, price(10), nil)},
		{Price: channelPrice("2-2-1", 1, 0, price(11), nil)},
	}

	sortPrimaryChannelFirst(details)

	if details[0].Price.SalesChannelId != "1-1-3" || details[1].Price.SalesChannelId != "2-2-1" {
		t.Error("order changed although the primary channel is not present")
	}
}

// The standard price block has to be minimised per shop field, not taken from
// whichever child happens to be cheapest overall.
func TestLowestItemPriceData(t *testing.T) {
	children := []wawi_structs.GetItem{
		{ID: 1, Item: wawi_structs.Item{SKU: "A", ItemPriceData: wawi_structs.ItemPriceData{
			SalesPriceNet: price(12.75), EbayPrice: price(13), AmazonPrice: price(14),
		}}},
		{ID: 2, Item: wawi_structs.Item{SKU: "B", ItemPriceData: wawi_structs.ItemPriceData{
			SalesPriceNet: price(11.94688), EbayPrice: price(20), AmazonPrice: nil,
		}}},
	}

	got := lowestItemPriceData(children)
	if *got.SalesPriceNet != 11.94688 {
		t.Errorf("SalesPriceNet = %v, want 11.94688", *got.SalesPriceNet)
	}
	if *got.EbayPrice != 13 {
		t.Errorf("EbayPrice = %v, want 13", *got.EbayPrice)
	}
	if *got.AmazonPrice != 14 {
		t.Errorf("AmazonPrice = %v, want 14", *got.AmazonPrice)
	}
	// Not a shop price, so it must not be written at all.
	if got.SuggestedRetailPrice != nil || got.PurchasePriceNet != nil {
		t.Error("UVP or purchase price were set although they are not shop prices")
	}
}

// Zero is Wawi's "no price set", not a price of nothing. Treating it as the
// minimum would overwrite a real price with 0.00.
func TestLowestPriceIgnoresZero(t *testing.T) {
	items := []wawi_structs.GetItem{
		{ID: 1, Item: wawi_structs.Item{SKU: "A", ItemPriceData: wawi_structs.ItemPriceData{
			EbayPrice: price(0), AmazonPrice: price(0),
		}}},
		{ID: 2, Item: wawi_structs.Item{SKU: "B", ItemPriceData: wawi_structs.ItemPriceData{
			EbayPrice: price(25), AmazonPrice: price(0),
		}}},
	}

	if got := LowestPrice(items, func(p wawi_structs.ItemPriceData) *float64 { return p.EbayPrice }); got == nil || *got != 25 {
		t.Errorf("EbayPrice = %v, want 25", got)
	}
	// Every value is zero, so there is nothing to write at all.
	if got := LowestPrice(items, func(p wawi_structs.ItemPriceData) *float64 { return p.AmazonPrice }); got != nil {
		t.Errorf("AmazonPrice = %v, want nil", *got)
	}
}
