package wawi

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Shu-AFK/WawiIC/cmd/gui/gui_structs"
	"github.com/Shu-AFK/WawiIC/cmd/openai"
	"github.com/Shu-AFK/WawiIC/cmd/openai/openai_structs"
	"github.com/Shu-AFK/WawiIC/cmd/wawi/wawi_structs"
)

var NoCategory = errors.New("no category selected")

func GetItems(query string, selectedCategoryID string) ([]wawi_structs.WItem, error) {
	if selectedCategoryID == "" || selectedCategoryID == "Kategorien" {
		return nil, NoCategory
	}

	itemQuery := wawi_structs.QueryItemStruct{
		SearchKeyword: query,
		ItemCategory:  selectedCategoryID,
		PageSize:      1000,
	}

	items, err := QueryItem(itemQuery)
	if err != nil {
		return nil, err
	}

	var itemRet []wawi_structs.WItem

	for _, item := range items {
		isFater := false
		isChild := false

		if len(item.ChildItems) > 0 {
			isFater = true
		}
		if item.ParentItemID != 0 {
			isChild = true
		}

		nItem := wawi_structs.WItem{
			GuiItem: wawi_structs.GuiItem{
				SKU:      item.SKU,
				Name:     item.Name,
				IsFather: isFater,
				IsChild:  isChild,
				Combine:  false,
			},
			GetItem: item,
		}

		itemRet = append(itemRet, nItem)
	}

	return itemRet, nil
}

func GetCategories(pageSize int) (map[string][]string, map[string]string, error) {
	categories, err := QueryCategories(pageSize)
	if err != nil {
		return nil, nil, err
	}

	tree := make(map[string][]string)
	labels := make(map[string]string)

	const rootID = "root"
	labels[rootID] = "Kategorien"

	for _, c := range categories {
		parentKey := fmt.Sprintf("%d", c.ParentCategoryID)
		childKey := fmt.Sprintf("%d", c.ID)

		tree[parentKey] = append(tree[parentKey], childKey)
		labels[childKey] = c.Name

		if c.ParentCategoryID == 0 {
			tree[rootID] = append(tree[rootID], childKey)
		}
	}

	return tree, labels, nil
}

func HandleAssignDone(combinations []gui_structs.Combination, selectedCombinationIndex int, variations map[string][]string, labels map[string]string) (string, error) {
	productNames := make([]string, 0, len(combinations))
	variationLabels := "["
	oldSKUs := make([]string, 0, len(combinations))

	for _, c := range combinations {
		productNames = append(productNames, c.Item.GuiItem.Name)
		variationLabels += fmt.Sprintf("[%s], ", c.Label)
		oldSKUs = append(oldSKUs, c.Item.GuiItem.SKU)
	}
	variationLabels = variationLabels[:len(variationLabels)-2]
	variationLabels += "]"

	userPrompt := openai.GetUserPrompt(
		productNames,
		combinations[selectedCombinationIndex].Item.GetItem.Description,
		variationLabels,
		oldSKUs,
	)

	ctx := context.Background()
	productSEO, err := openai.MakeRequest(ctx, userPrompt)
	if err != nil {
		return "", err
	}

	items := make([]wawi_structs.GetItem, 0, len(combinations))
	for _, c := range combinations {
		items = append(items, c.Item.GetItem)
	}
	parentItem := createParentStruct(productSEO, items)
	item, err := CreateParentItem(parentItem)
	if err != nil {
		return "", err
	}
	if item.IsActive == false {
		return "", errors.New("item is not active")
	}

	/*
			err = SetItemActiveSalesChannels(strconv.Itoa(item.ID), combinations[selectedCombinationIndex].Item.GetItem.ActiveSalesChannels)
			if err != nil {
				return "", err
			}


			var images []wawi_structs.CreateImageStruct
			imageBuffer, err := GetImagesFromItem(combinations[selectedCombinationIndex].Item.GetItem)
			if err != nil {
				return "", err
			}
			images = append(images, imageBuffer...)
			for _, i := range combinations {
				if i.Item.GetItem.SKU == combinations[selectedCombinationIndex].Item.GuiItem.SKU {
					continue
				}
				imageBuffer, err = GetImagesFromItem(i.Item.GetItem)
				if err != nil {
					return "", err
				}
				images = append(images, imageBuffer...)
			}

			for _, image := range images {
				err = CreateItemImage(image, string(rune(item.ID)))
				if err != nil {
					return "", err
				}
			}

		err = UpdateDescription(strconv.Itoa(item.ID), *productSEO)
		if err != nil {
			return "", err
		}

		var propertyValueIds []string
		for _, c := range combinations {
			properties, err := QueryItemProperties(strconv.Itoa(c.Item.GetItem.ID))
			if err != nil {
				return "", err
			}

			for _, property := range properties.Properties {
				propertyValueIds = append(propertyValueIds, strconv.Itoa(property.PropertyValueId))
			}
		}
		uniquePropertyValueIds := uniqueStrings(propertyValueIds)
		for _, id := range uniquePropertyValueIds {
			_, err := CreateItemProperty(strconv.Itoa(item.ID), id)
			if err != nil {
				return "", err
			}
		}
	*/

	variationTree := BuildVariationLabelIndex(variations, labels)
	var variationOrder []string
	valueIDByLabel := map[string]map[string]string{}

	for parentName, children := range variationTree {
		parentVariation, err := CreateVariations(strconv.Itoa(item.ID), parentName)
		if err != nil {
			return "", err
		}

		variationOrder = append(variationOrder, parentName)
		if _, ok := valueIDByLabel[parentName]; !ok {
			valueIDByLabel[parentName] = map[string]string{}
		}

		for _, childName := range children {
			v, err := CreateVariationValue(strconv.Itoa(item.ID), strconv.Itoa(parentVariation.Id), childName)
			if err != nil {
				return "", err
			}
			valueIDByLabel[parentName][childName] = strconv.Itoa(v.Id)
		}
	}

	// Helper: parse a label like "Größe: 150ml, Farbe: Blau" or "[Farbe] Blau"
	parseNamesByLabel := func(raw string) map[string]string {
		out := map[string]string{}
		s := strings.TrimSpace(raw)
		s = strings.Trim(s, "[]")
		parts := strings.Split(s, ",")
		for _, p := range parts {
			seg := strings.TrimSpace(p)
			if seg == "" {
				continue
			}
			var key, val string
			if i := strings.IndexAny(seg, ":="); i >= 0 {
				key = strings.TrimSpace(seg[:i])
				val = strings.TrimSpace(seg[i+1:])
			} else {
				// Try bracketed "[Farbe] Blau" or "Farbe Blau"
				fields := strings.Fields(seg)
				if len(fields) >= 2 {
					key = strings.Trim(fields[0], "[]")
					val = strings.TrimSpace(seg[len(fields[0]):])
				}
			}
			key = strings.Trim(key, "[]- ")
			val = strings.Trim(val, "[]- ")
			if key != "" && val != "" {
				out[key] = val
			}
		}
		return out
	}

	// Helper: case-insensitive substring check
	containsCI := func(haystack, needle string) bool {
		return strings.Contains(strings.ToLower(haystack), strings.ToLower(needle))
	}

	for _, combination := range combinations {
		// First, parse what we can from the combination label
		namesByLabel := parseNamesByLabel(combination.Label)

		var comboValueIDs []string
		for _, vLabel := range variationOrder {
			name := namesByLabel[vLabel]

			// Fallback 1: try to detect which child value appears in the label text
			if name == "" {
				if children, ok := variationTree[vLabel]; ok {
					candidates := []string{}
					for _, childVal := range children {
						if containsCI(combination.Label, childVal) {
							candidates = append(candidates, childVal)
						}
					}
					if len(candidates) == 1 {
						name = candidates[0]
					}
				}
			}

			// Fallback 2: try to detect value from the combination item name
			if name == "" {
				if children, ok := variationTree[vLabel]; ok {
					candidates := []string{}
					for _, childVal := range children {
						if containsCI(combination.Item.GuiItem.Name, childVal) {
							candidates = append(candidates, childVal)
						}
					}
					if len(candidates) == 1 {
						name = candidates[0]
					}
				}
			}

			// Fallback 3: if the variation has only one value, use it
			if name == "" {
				if children, ok := variationTree[vLabel]; ok && len(children) == 1 {
					name = children[0]
				}
			}

			id, ok := valueIDByLabel[vLabel][name]
			if !ok {
				return "", fmt.Errorf("missing value ID for %s=%s", vLabel, name)
			}
			comboValueIDs = append(comboValueIDs, id)
		}

		if err := AssignChildToParent(
			strconv.Itoa(item.ID),
			strconv.Itoa(combination.Item.GetItem.ID),
			comboValueIDs,
		); err != nil {
			return "", err
		}
	}
	return parentItem.SKU, nil
}

func PtrIfSet[T comparable](v T) *T {
	var zero T
	if v == zero {
		return nil
	}
	return &v
}

func createParentStruct(seo *openai_structs.ProductSEO, items []wawi_structs.GetItem) wawi_structs.ItemCreate {
	cheapestItemIndex := findCheapestItem(items)
	dangerousStruct := getItemDangerous(items)
	searchTerms := getSearchTerms(items)

	_, seoSKUSplit, _ := strings.Cut(seo.NewSKU, "-")
	pNum, _, _ := strings.Cut(items[cheapestItemIndex].SKU, "-")

	newSKU := fmt.Sprintf("%s-%s", pNum, seoSKUSplit)

	ts := time.Now().UTC().Format(time.RFC3339)

	// TODO: Attribute übernehmen
	parentItem := wawi_structs.ItemCreate{
		SKU:                 newSKU,
		ManufacturerID:      PtrIfSet(items[cheapestItemIndex].ManufacturerID),
		ResponsiblePersonID: PtrIfSet(items[cheapestItemIndex].ResponsiblePersonID),
		IsActive:            true,
		Categories:          addCategoryToParent(items[cheapestItemIndex].Categories),
		Name:                seo.CombinedArticleName,
		Description:         seo.Description,
		ShortDescription:    seo.ShortDescription,
		Identifiers: &wawi_structs.Identifiers{
			ManufacturerNumber: PtrIfSet(removeUpToFirstDash(seo.NewSKU)),
		},
		ItemPriceData: &wawi_structs.ItemPriceData{
			SalesPriceNet:        items[cheapestItemIndex].ItemPriceData.SalesPriceNet,
			SuggestedRetailPrice: items[cheapestItemIndex].ItemPriceData.SuggestedRetailPrice,
			EbayPrice:            items[cheapestItemIndex].ItemPriceData.EbayPrice,
			AmazonPrice:          items[cheapestItemIndex].ItemPriceData.AmazonPrice,
		},
		Annotation:      "Mit API erstellt",
		Added:           ts,
		Changed:         ts,
		ReleasedOnDate:  ts,
		CountryOfOrigin: items[cheapestItemIndex].CountryOfOrigin,
		Weights: &wawi_structs.Weights{
			ItemWeight:     items[cheapestItemIndex].Weights.ItemWeight,
			ShippingWeight: items[cheapestItemIndex].Weights.ShippingWeight,
		},
		DangerousGoods:  dangerousStruct,
		Taric:           items[cheapestItemIndex].Taric,
		SearchTerms:     searchTerms,
		PriceListActive: false,
	}

	return parentItem
}
