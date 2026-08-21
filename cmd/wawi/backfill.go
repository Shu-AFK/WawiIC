package wawi

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/Shu-AFK/WawiIC/cmd/wawi/wawi_structs"
)

// ParentAnnotation marks the items created by this tool.
const ParentAnnotation = "Mit API erstellt"

type BackfillResult struct {
	ParentSKU string
	ParentID  int
	// Children is how many child items the prices were compared across.
	Children int
	Prices   int
	// Details lists the winning price per sales channel, customer group and
	// quantity tier, together with the child it came from.
	Details []BackfillPrice
	Skipped string
	Err     error
}

// BackfillPrice is one price that would be written to the parent item.
type BackfillPrice struct {
	Price     wawi_structs.ItemSalesChannelPrice
	SourceSKU string
}

// salesChannelPriceKey identifies one price slot. A sales channel can hold a
// different price per customer group and per quantity tier, and each of those
// has to be compared separately.
type salesChannelPriceKey struct {
	SalesChannelID  string
	CustomerGroupID int
	FromQuantity    int
}

type BackfillOptions struct {
	// ItemIDs restricts the run to these parent items. When set, nothing is
	// searched at all and CategoryID is ignored.
	ItemIDs []int
	// CategoryID limits the search to one category. Zero walks every item, which
	// on a large catalogue means paging through the entire stock.
	CategoryID int
	// Apply writes the prices. When false the run only reports what it would do.
	Apply bool
}

// BackfillSalesChannelPrices copies the sales channel prices onto parent items that
// were created before the prices were carried over. Pages are processed as they
// arrive so the catalogue is never held in memory.
func BackfillSalesChannelPrices(opts BackfillOptions, progress func(string)) ([]BackfillResult, error) {
	if progress == nil {
		progress = func(string) {}
	}

	if len(opts.ItemIDs) > 0 {
		progress(fmt.Sprintf("%d vorgegebene Artikel-IDs werden bearbeitet.", len(opts.ItemIDs)))
		return backfillByID(opts.ItemIDs, opts.Apply, progress), nil
	}

	query := wawi_structs.QueryItemStruct{PageSize: 100}
	if opts.CategoryID != 0 {
		applied, inCategory, catalogue, err := categoryFilterApplied(opts.CategoryID)
		if err != nil {
			return nil, fmt.Errorf("failed to check the category filter: %w", err)
		}

		if applied {
			query.ItemCategory = strconv.Itoa(opts.CategoryID)
			progress(fmt.Sprintf("Kategoriefilter wirkt: %d von %d Artikeln in Kategorie %d.", inCategory, catalogue, opts.CategoryID))
		} else {
			progress(fmt.Sprintf(
				"Achtung: der Kategoriefilter wird vom Server ignoriert (Kategorie %d liefert %d Artikel, genau wie der gesamte Artikelstamm). Es wird stattdessen alles durchsucht, das dauert deutlich länger.",
				opts.CategoryID, inCategory,
			))
		}
	} else {
		progress("Suche Vaterartikel im gesamten Artikelstamm...")
	}

	results := make([]BackfillResult, 0)
	scanned := 0
	reportedTotal := false

	err := QueryItemsPaged(query, func(page []wawi_structs.GetItem, total int) error {
		// Surfacing the match count up front makes an ignored category filter
		// obvious before the run spends hours walking the whole catalogue.
		if !reportedTotal {
			reportedTotal = true
			progress(fmt.Sprintf("%d Artikel werden durchsucht.", total))
		}

		for _, item := range page {
			scanned++
			if scanned%500 == 0 {
				progress(fmt.Sprintf("%d von %d Artikeln geprüft, %d Vaterartikel gefunden.", scanned, total, len(results)))
			}

			if notAParent(item) != "" || item.Annotation != ParentAnnotation {
				continue
			}

			res := backfillParent(item, opts.Apply)
			results = append(results, res)
			progress(formatBackfillProgress(res, len(results)))
		}

		return nil
	})
	if err != nil {
		return results, fmt.Errorf("failed to query items: %w", err)
	}

	progress(fmt.Sprintf("%d Artikel geprüft, %d Vaterartikel bearbeitet.", scanned, len(results)))

	return results, nil
}

func backfillParent(parent wawi_structs.GetItem, apply bool) BackfillResult {
	res := BackfillResult{ParentSKU: parent.SKU, ParentID: parent.ID}

	children, err := fetchItemsByID(parent.ChildItems)
	if err != nil {
		res.Err = err
		return res
	}
	if len(children) == 0 {
		res.Skipped = "keine Kindartikel abrufbar"
		return res
	}
	res.Children = len(children)

	details, err := lowestPricePerChannel(children)
	if err != nil {
		res.Err = err
		return res
	}
	if len(details) == 0 {
		res.Skipped = "kein Kindartikel hat Verkaufskanalpreise"
		return res
	}
	res.Details = details
	res.Prices = len(details)

	if apply {
		if err := writeSalesChannelPrices(parent.ID, details); err != nil {
			res.Err = err
		}
	}

	return res
}

// writeSalesChannelPrices puts one price per sales channel slot onto an item.
func writeSalesChannelPrices(itemID int, details []BackfillPrice) error {
	target := strconv.Itoa(itemID)

	for _, detail := range details {
		if err := UpdateItemSalesChannelPrice(target, detail.Price); err != nil {
			return err
		}
	}

	return nil
}

// ApplyLowestPricePerChannel gives an item the cheapest price each sales channel
// has across the supplied children. Used both when a parent item is created and
// when the prices are backfilled onto an existing one, so both paths agree.
func ApplyLowestPricePerChannel(children []wawi_structs.GetItem, targetItemID int) error {
	details, err := lowestPricePerChannel(children)
	if err != nil {
		return err
	}

	return writeSalesChannelPrices(targetItemID, details)
}

// LowestPrice returns the smallest non nil value the getter finds across items.
func LowestPrice(items []wawi_structs.GetItem, get func(wawi_structs.ItemPriceData) *float64) *float64 {
	var best *float64

	for _, item := range items {
		value := get(item.ItemPriceData)
		if value == nil {
			continue
		}
		if best == nil || *value < *best {
			best = value
		}
	}

	return best
}

// lowestPricePerChannel collects the sales channel prices of every child and
// keeps the cheapest one per channel, customer group and quantity tier. Taking
// the minimum per slot rather than one child's whole price list means the parent
// never advertises more than the cheapest variant actually costs on that shop.
func lowestPricePerChannel(children []wawi_structs.GetItem) ([]BackfillPrice, error) {
	best := make(map[salesChannelPriceKey]BackfillPrice)
	order := make([]salesChannelPriceKey, 0)

	// Only online shops and JTL-POS accept item prices. Sending one to any other
	// channel is rejected with a 400, even though that channel can appear in the
	// prices read off a child item.
	acceptsPrices, capErr := channelsAcceptingPrices()

	for _, child := range children {
		prices, err := QueryItemSalesChannelPrices(strconv.Itoa(child.ID))
		if err != nil {
			return nil, err
		}

		for _, price := range prices {
			if price.NetPrice == nil && price.ReduceStandardPriceByPercent == nil {
				continue
			}
			if capErr == nil && !acceptsPrices[price.SalesChannelId] {
				continue
			}

			key := salesChannelPriceKey{
				SalesChannelID:  price.SalesChannelId,
				CustomerGroupID: price.CustomerGroupId,
				FromQuantity:    price.FromQuantity,
			}

			candidate := BackfillPrice{Price: price, SourceSKU: child.SKU}
			current, seen := best[key]
			if !seen {
				best[key] = candidate
				order = append(order, key)
				continue
			}
			if isCheaper(candidate.Price, current.Price) {
				best[key] = candidate
			}
		}
	}

	details := make([]BackfillPrice, 0, len(order))
	for _, key := range order {
		details = append(details, best[key])
	}

	return details, nil
}

// isCheaper reports whether a undercuts b. A concrete net price always wins over
// a percentage reduction, because the two cannot be compared without knowing the
// standard price they each apply to; among percentages the larger discount wins.
func isCheaper(a, b wawi_structs.ItemSalesChannelPrice) bool {
	switch {
	case a.NetPrice != nil && b.NetPrice != nil:
		return *a.NetPrice < *b.NetPrice
	case a.NetPrice != nil:
		return true
	case b.NetPrice != nil:
		return false
	case a.ReduceStandardPriceByPercent != nil && b.ReduceStandardPriceByPercent != nil:
		return *a.ReduceStandardPriceByPercent > *b.ReduceStandardPriceByPercent
	default:
		return false
	}
}

func formatBackfillProgress(res BackfillResult, index int) string {
	switch {
	case res.Err != nil:
		return fmt.Sprintf("[%d] %s: FEHLER %v", index, res.ParentSKU, res.Err)
	case res.Skipped != "":
		return fmt.Sprintf("[%d] %s: übersprungen", index, res.ParentSKU)
	default:
		return fmt.Sprintf("[%d] %s: %d Preise aus %d Kindartikeln", index, res.ParentSKU, res.Prices, res.Children)
	}
}

func fetchItemsByID(ids []int) ([]wawi_structs.GetItem, error) {
	items := make([]wawi_structs.GetItem, 0, len(ids))

	for _, id := range ids {
		item, found, err := GetItemByID(id)
		if err != nil {
			return nil, err
		}
		if !found {
			continue
		}

		items = append(items, *item)
	}

	return items, nil
}

// categoryFilterApplied reports whether the server actually narrows results by
// category. Older API versions accept the parameter and silently ignore it, which
// would turn a targeted backfill into a full catalogue scan with no visible error.
// Comparing the match counts costs two requests and settles it for the instance
// in front of us rather than relying on the documented behaviour.
func categoryFilterApplied(categoryID int) (applied bool, inCategory int, catalogue int, err error) {
	catalogue, err = countMatches(wawi_structs.QueryItemStruct{PageSize: 1})
	if err != nil {
		return false, 0, 0, err
	}

	inCategory, err = countMatches(wawi_structs.QueryItemStruct{
		PageSize:     1,
		ItemCategory: strconv.Itoa(categoryID),
	})
	if err != nil {
		return false, 0, 0, err
	}

	return inCategory < catalogue, inCategory, catalogue, nil
}

// countMatches reads only the first page to learn how many items a query matches.
func countMatches(query wawi_structs.QueryItemStruct) (int, error) {
	count := 0

	err := QueryItemsPaged(query, func(_ []wawi_structs.GetItem, total int) error {
		count = total
		return ErrStopPaging
	})
	if err != nil {
		return 0, err
	}

	return count, nil
}

// backfillByID works through an explicit list of parent items. Every item is
// fetched directly by its internal id, so nothing is searched at all - neither
// the catalogue size nor the category filter matters here.
func backfillByID(ids []int, apply bool, progress func(string)) []BackfillResult {
	results := make([]BackfillResult, 0, len(ids))

	for i, id := range ids {
		prefix := fmt.Sprintf("[%d/%d]", i+1, len(ids))

		parent, found, err := GetItemByID(id)
		if err != nil {
			results = append(results, BackfillResult{ParentID: id, Err: err})
			progress(fmt.Sprintf("%s ID %d: Fehler", prefix, id))
			continue
		}
		if !found {
			results = append(results, BackfillResult{ParentID: id, Skipped: "Artikel nicht gefunden"})
			progress(fmt.Sprintf("%s ID %d: nicht gefunden", prefix, id))
			continue
		}

		if reason := notAParent(*parent); reason != "" {
			results = append(results, BackfillResult{
				ParentID:  id,
				ParentSKU: parent.SKU,
				Skipped:   reason,
			})
			progress(fmt.Sprintf("%s %s: %s", prefix, parent.SKU, reason))
			continue
		}

		res := backfillParent(*parent, apply)
		results = append(results, res)
		progress(prefix + " " + strings.TrimPrefix(formatBackfillProgress(res, 0), "[0] "))

		// An explicit list overrides the annotation check, but a parent this tool
		// did not create is worth pointing out before prices are written to it.
		if parent.Annotation != ParentAnnotation {
			progress(fmt.Sprintf("  Hinweis: %s wurde nicht mit diesem Tool erstellt.", parent.SKU))
		}
	}

	return results
}

// ParseItemIDs reads one item ID per line. Blank lines and lines starting with #
// are ignored, and only the first column is read so a CSV export can be passed
// straight through.
func ParseItemIDs(data string) ([]int, error) {
	ids := make([]int, 0)
	seen := make(map[int]struct{})

	for i, line := range strings.Split(data, "\n") {
		field := strings.TrimSpace(line)
		field = strings.TrimPrefix(field, "\ufeff")
		if field == "" || strings.HasPrefix(field, "#") {
			continue
		}

		if cut := strings.IndexAny(field, ",;\t"); cut >= 0 {
			field = strings.TrimSpace(field[:cut])
		}
		field = strings.Trim(field, "\"'")

		id, err := strconv.Atoi(field)
		if err != nil {
			return nil, fmt.Errorf("line %d: %q is not an item ID", i+1, field)
		}
		if id <= 0 {
			return nil, fmt.Errorf("line %d: %d is not a valid item ID", i+1, id)
		}

		if _, dup := seen[id]; dup {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}

	if len(ids) == 0 {
		return nil, fmt.Errorf("no item IDs found")
	}

	return ids, nil
}

// notAParent explains why an item cannot receive combined prices, or returns an
// empty string when it is a proper parent article. Prices are only meaningful on
// an item that actually has variants to compare.
func notAParent(item wawi_structs.GetItem) string {
	if item.ParentItemID != 0 {
		return "kein Vaterartikel (ist selbst ein Kindartikel)"
	}
	if len(item.ChildItems) == 0 {
		return "kein Vaterartikel (keine Kindartikel)"
	}

	return ""
}

var (
	channelPriceSupportOnce sync.Once
	channelPriceSupport     map[string]bool
	channelPriceSupportErr  error
)

// channelsAcceptingPrices reports, per sales channel id, whether item prices can
// be written to it at all. Queried once and reused, since it is the same answer
// for every item in a run.
func channelsAcceptingPrices() (map[string]bool, error) {
	channelPriceSupportOnce.Do(func() {
		channels, err := QuerySalesChannels()
		if err != nil {
			channelPriceSupportErr = err
			return
		}

		support := make(map[string]bool, len(channels))
		for _, channel := range channels {
			support[channel.Id] = channel.ItemCapabilities.Prices
		}
		channelPriceSupport = support
	})

	return channelPriceSupport, channelPriceSupportErr
}
