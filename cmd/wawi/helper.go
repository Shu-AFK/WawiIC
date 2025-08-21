package wawi

import (
	"sort"
	"strings"

	"github.com/Shu-AFK/WawiIC/cmd/wawi/wawi_structs"
)

func getItemDangerous(items []wawi_structs.GetItem) *wawi_structs.DangerousGoods {
	for _, i := range items {
		if i.DangerousGoods.HazardNo != "" {
			return &i.DangerousGoods
		}
	}

	return nil
}

func findCheapestItem(items []wawi_structs.GetItem) int {
	cheapestItem := 0
	for i, item := range items {
		if *item.ItemPriceData.SalesPriceNet < *items[cheapestItem].ItemPriceData.SalesPriceNet {
			cheapestItem = i
		}
	}

	return cheapestItem
}

func removeUpToFirstDash(s string) string {
	_, after, found := strings.Cut(s, "-")
	if found {
		return after
	}
	return s
}

func getSearchTerms(items []wawi_structs.GetItem) string {
	searchTerms := make(map[string]bool)

	for _, item := range items {
		searchKeywords := strings.Split(item.SearchTerms, "")
		for _, keyword := range searchKeywords {
			kw := strings.TrimSpace(keyword)
			if kw != "" {
				searchTerms[keyword] = true
			}
		}
	}

	uniqueTerms := make([]string, 0, len(searchTerms))
	for term := range searchTerms {
		uniqueTerms = append(uniqueTerms, term)
	}

	return strings.Join(uniqueTerms, ", ")
}

func BuildVariationLabelIndex(variations map[string][]string, labels map[string]string) map[string][]string {
	out := make(map[string][]string)

	for parentID, childIDs := range variations {
		parentLabel, ok := labels[parentID]
		if !ok || parentLabel == "" {
			continue
		}

		seen := make(map[string]struct{})
		for _, cid := range childIDs {
			childLabel, ok := labels[cid]
			if !ok || childLabel == "" {
				continue
			}
			if _, dup := seen[childLabel]; dup {
				continue
			}
			seen[childLabel] = struct{}{}
			out[parentLabel] = append(out[parentLabel], childLabel)
		}

		if len(out[parentLabel]) > 1 {
			sort.Strings(out[parentLabel])
		}
	}
	delete(out, "Variationen")
	return out
}

func childNameFromVariationID(variationID string, labels map[string]string) (string, bool) {
	parts := strings.Split(variationID, "|")
	if len(parts) == 0 {
		return "", false
	}
	childID := strings.TrimSpace(parts[len(parts)-1])
	name, ok := labels[childID]
	return name, ok
}
