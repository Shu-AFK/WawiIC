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

// Each shop position gets its own channel, so shops that genuinely have
// different prices stay apart instead of all receiving the cheapest one.
func TestRetargetChannels(t *testing.T) {
	details := []BackfillPrice{
		{Price: channelPrice("1-1-3", 3, 0, price(11.94688), nil), Occurrence: 0},
		{Price: channelPrice("1-1-3", 3, 0, price(16.42696), nil), Occurrence: 1},
		{Price: channelPrice("1-1-3", 2, 0, price(18.84), nil), Occurrence: 0},
	}

	out := retargetChannels(details, []string{"2-2-2", "2-2-6"})

	if len(out) != 3 {
		t.Fatalf("got %d prices, want 3", len(out))
	}
	if out[0].Price.SalesChannelId != "2-2-2" || *out[0].Price.NetPrice != 11.94688 {
		t.Errorf("first = %s %v", out[0].Price.SalesChannelId, *out[0].Price.NetPrice)
	}
	if out[1].Price.SalesChannelId != "2-2-6" || *out[1].Price.NetPrice != 16.42696 {
		t.Errorf("second = %s %v", out[1].Price.SalesChannelId, *out[1].Price.NetPrice)
	}
	if out[2].Price.SalesChannelId != "2-2-2" || *out[2].Price.NetPrice != 18.84 {
		t.Errorf("third = %s %v", out[2].Price.SalesChannelId, *out[2].Price.NetPrice)
	}
}

// More shops on the children than channels named: the surplus is dropped rather
// than written to the wrong shop.
func TestRetargetChannelsDropsSurplus(t *testing.T) {
	details := []BackfillPrice{
		{Price: channelPrice("1-1-3", 3, 0, price(10), nil), Occurrence: 0},
		{Price: channelPrice("1-1-3", 3, 0, price(20), nil), Occurrence: 1},
	}

	out := retargetChannels(details, []string{"2-2-2"})

	if len(out) != 1 {
		t.Fatalf("got %d prices, want 1", len(out))
	}
	if *out[0].Price.NetPrice != 10 {
		t.Errorf("kept %v, want 10", *out[0].Price.NetPrice)
	}
}

func TestFormatStoredValues(t *testing.T) {
	if got := formatStoredValues(nil); got != "nichts" {
		t.Errorf("formatStoredValues(nil) = %q, want \"nichts\"", got)
	}
	if got := formatStoredValues([]float64{11.94688, 18.84}); got != "11.94688, 18.84" {
		t.Errorf("formatStoredValues() = %q", got)
	}
}

// A child that is not active in a shop has no price there, so it must not count
// towards that shop's minimum - the cheapest of the remaining children wins.
func TestSourcesCountsOnlyChildrenWithAPrice(t *testing.T) {
	details := []BackfillPrice{
		{Price: channelPrice("1-1-3", 3, 0, price(11.94688), nil), SourceSKU: "A", Sources: 2},
		{Price: channelPrice("1-1-3", 3, 0, price(16.42696), nil), Occurrence: 1, SourceSKU: "A", Sources: 1},
	}

	// Shop 1 was priced by both children, shop 2 only by one.
	if details[0].Sources != 2 {
		t.Errorf("shop 1 sources = %d, want 2", details[0].Sources)
	}
	if details[1].Sources != 1 {
		t.Errorf("shop 2 sources = %d, want 1", details[1].Sources)
	}

	out := retargetChannels(details, []string{"2-2-2", "2-2-1"})
	if len(out) != 2 {
		t.Fatalf("got %d prices, want 2", len(out))
	}
	// The shop only one child sells in keeps that child's price rather than the
	// cheaper price of a shop it is not in.
	if *out[1].Price.NetPrice != 16.42696 {
		t.Errorf("shop 2 price = %v, want 16.42696", *out[1].Price.NetPrice)
	}
}
