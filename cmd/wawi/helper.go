package wawi

import (
	"encoding/base64"
	"regexp"
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

func collectUniqueCategories(items []wawi_structs.GetItem) []wawi_structs.Category {
	seen := make(map[int]wawi_structs.Category)
	order := make([]int, 0)

	for _, item := range items {
		for _, category := range item.Categories {
			if _, alreadyAdded := seen[category.CategoryID]; alreadyAdded {
				continue
			}

			seen[category.CategoryID] = category
			order = append(order, category.CategoryID)
		}
	}

	categories := make([]wawi_structs.Category, 0, len(order)+1)
	for _, id := range order {
		categories = append(categories, seen[id])
	}

	if _, exists := seen[categoryID]; !exists {
		categories = append(categories, wawi_structs.Category{CategoryID: categoryID})
	}

	return categories
}

func uniqueStrings(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, s := range in {
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}

func normalizeBase64(b64 string) (string, error) {
	re := regexp.MustCompile(`\s+`)
	b64 = re.ReplaceAllString(b64, "")

	_, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return "", err
	}

	return b64, nil
}
