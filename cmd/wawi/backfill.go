package wawi

import (
	"fmt"
	"strconv"

	"github.com/Shu-AFK/WawiIC/cmd/wawi/wawi_structs"
)

// ParentAnnotation marks the items created by this tool.
const ParentAnnotation = "Mit API erstellt"

type BackfillResult struct {
	ParentSKU string
	ParentID  int
	SourceSKU string
	Prices    int
	Skipped   string
	Err       error
}

type BackfillOptions struct {
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

			if len(item.ChildItems) == 0 || item.Annotation != ParentAnnotation {
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

	source := children[pickPriceSource(parent, children)]
	res.SourceSKU = source.SKU

	prices, err := QueryItemSalesChannelPrices(strconv.Itoa(source.ID))
	if err != nil {
		res.Err = err
		return res
	}
	if len(prices) == 0 {
		res.Skipped = "Quellartikel hat keine Verkaufskanalpreise"
		return res
	}
	res.Prices = len(prices)

	if apply {
		if err := copySalesChannelPrices(source.ID, parent.ID); err != nil {
			res.Err = err
		}
	}

	return res
}

func formatBackfillProgress(res BackfillResult, index int) string {
	switch {
	case res.Err != nil:
		return fmt.Sprintf("[%d] %s: Fehler", index, res.ParentSKU)
	case res.Skipped != "":
		return fmt.Sprintf("[%d] %s: übersprungen", index, res.ParentSKU)
	default:
		return fmt.Sprintf("[%d] %s: %d Preise von %s", index, res.ParentSKU, res.Prices, res.SourceSKU)
	}
}

// pickPriceSource returns the index of the child the parent originally took its
// prices from. The parent kept the cheapest child's standard price, so an exact
// match on that price reproduces the original choice even if prices have moved
// since; otherwise the cheapest child today is the best guess.
func pickPriceSource(parent wawi_structs.GetItem, children []wawi_structs.GetItem) int {
	if parent.ItemPriceData.SalesPriceNet != nil {
		for i, child := range children {
			if child.ItemPriceData.SalesPriceNet == nil {
				continue
			}
			if *child.ItemPriceData.SalesPriceNet == *parent.ItemPriceData.SalesPriceNet {
				return i
			}
		}
	}

	return findCheapestItem(children)
}

func fetchItemsByID(ids []int) ([]wawi_structs.GetItem, error) {
	items := make([]wawi_structs.GetItem, 0, len(ids))

	for _, id := range ids {
		found, err := QueryItem(wawi_structs.QueryItemStruct{
			ItemID:   strconv.Itoa(id),
			PageSize: 5,
		})
		if err != nil {
			return nil, err
		}

		for _, item := range found {
			if item.ID == id {
				items = append(items, item)
				break
			}
		}
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
