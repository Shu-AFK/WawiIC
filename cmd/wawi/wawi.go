package wawi

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"image"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Shu-AFK/WawiIC/cmd/gui/gui_structs"
	"github.com/Shu-AFK/WawiIC/cmd/imagecomb"
	"github.com/Shu-AFK/WawiIC/cmd/openai"
	"github.com/Shu-AFK/WawiIC/cmd/openai/openai_structs"
	"github.com/Shu-AFK/WawiIC/cmd/wawi/wawi_structs"
)

var (
	NoCategory = errors.New("no category selected")
	NoSupplier = errors.New("no supplier selected")
)

func GetItems(query string, selectedCategoryID string, selectedSupplierID int) ([]wawi_structs.WItem, error) {
	itemQuery := wawi_structs.QueryItemStruct{
		SearchKeyword: query,
		PageSize:      20,
	}

	if SearchMode == "category" {
		if selectedCategoryID == "" || selectedCategoryID == "Kategorien" {
			return nil, NoCategory
		}

		itemQuery.ItemCategory = selectedCategoryID
	} else if SearchMode == "supplier" {
		if selectedSupplierID == 0 {
			return nil, NoSupplier
		}

		itemQuery.ItemSupplier = strconv.Itoa(selectedSupplierID)
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

func CheckIfSKUExists(sku string) (bool, error) {
	itemQuery := wawi_structs.QueryItemStruct{
		SearchKeyword: sku,
		PageSize:      5,
	}

	items, err := QueryItem(itemQuery)
	if err != nil {
		return false, err
	}

	return len(items) > 0, nil
}

func HandleAssignDone(combinations []gui_structs.Combination, variations map[string][]string, labels map[string]string, sku string, mergeImages bool, errorOnNoImages bool) (string, error) {
	productNames, variationLabels, oldSKUs := buildPromptInputs(combinations)

	if errorOnNoImages {
		missing := detectMissingPrimaryImages(combinations)
		if len(missing) > 0 {
			lines := make([]string, 0, len(missing))
			for _, item := range missing {
				lines = append(lines, fmt.Sprintf("- %s (SKU: %s)", item.Name, item.SKU))
			}
			return "", fmt.Errorf("fehlende Bilder für folgende Artikel:\n%s", strings.Join(lines, "\n"))
		}
	}

	images, err := getImagesAndBase64(combinations)
	if err != nil {
		return "", err
	}

	allImages, err := getAllItemImages(combinations)
	if err != nil {
		return "", err
	}

	productSEO, err := generateSEO(
		productNames,
		combinations[0].Item.GetItem.Description,
		variationLabels,
		oldSKUs,
	)
	if err != nil {
		return "", err
	}

	items := collectItemsFromCombinations(combinations)
	parentItem := createParentStruct(productSEO, items, sku)

	item, err := CreateParentItem(parentItem)
	if err != nil {
		return "", err
	}
	if !item.IsActive {
		return "", errors.New("item is not active")
	}

	activateSalesChannel := make(chan error, 1)
	descriptionChannel := make(chan error, 1)
	propertyChannel := make(chan error, 1)
	imageChannel := make(chan error, 1)
	type variantsResponse struct {
		order          []string
		valueIdByLabel map[string]map[string]string
		variationTree  map[string][]string
		err            error
	}
	variationsChannel := make(chan variantsResponse, 1)

	if ActivateSalesChannel {
		go func() {
			activateSalesChannel <- setActiveSalesChannels(item.ID, combinations[0].Item.GetItem.ActiveSalesChannels)
		}()
	} else {
		activateSalesChannel <- nil
	}

	// Meta description
	go func() {
		descriptionChannel <- UpdateDescription(strconv.Itoa(item.ID), *productSEO)
	}()

	// image handling logic
	go func() {
		if len(images) == 0 {
			if !mergeImages && errorOnNoImages {
				imageChannel <- fmt.Errorf("no images available and merge disabled")
				return
			}
			imageChannel <- nil
			return
		}

		if mergeImages {
			img, err := imagecomb.CombineImages(images)
			if err != nil {
				imageChannel <- fmt.Errorf("failed to combine images: %w", err)
				return
			}

			imageChannel <- uploadCombinedImage(img, *item)
		} else {
			imageChannel <- nil
		}
	}()

	// Error in older api version
	/*
		go func() {
			propIDs, err := collectUniquePropertyValueIDs(combinations)
			if err != nil {
				propertyChannel <- err
				return
			}

			for _, id := range propIDs {
				if _, err := CreateItemProperty(strconv.Itoa(item.ID), id); err != nil {
					propertyChannel <- err
					return
				}
			}

			propertyChannel <- nil
		}() */
	propertyChannel <- nil

	go func() {
		variationTree := BuildVariationLabelIndex(variations, labels)
		variationOrder, valueIdByLabel, err := createVariationsAndValues(item.ID, variationTree)

		variationsChannel <- variantsResponse{
			order:          variationOrder,
			valueIdByLabel: valueIdByLabel,
			variationTree:  variationTree,
			err:            err,
		}
	}()

	if err := <-activateSalesChannel; err != nil {
		return "", err
	}
	if err := <-descriptionChannel; err != nil {
		return "", err
	}
	if err := <-imageChannel; err != nil {
		return "", err
	}
	if err := <-propertyChannel; err != nil {
		return "", err
	}
	variationsResp := <-variationsChannel
	if variationsResp.err != nil {
		return "", variationsResp.err
	}

	assignChildChannel := make(chan error, 1)
	uploadItemImagesChannel := make(chan error, 1)

	if ActivateSalesChannel {
		go func() {
			assignChildChannel <- assignChildrenToParent(
				item.ID,
				combinations,
				variationsResp.variationTree,
				variationsResp.order,
				variationsResp.valueIdByLabel,
			)
		}()
	} else {
		assignChildChannel <- nil
	}
	go func() {
		uploadItemImagesChannel <- uploadAllItemImages(strconv.Itoa(item.ID), allImages)
	}()

	if err := <-uploadItemImagesChannel; err != nil {
		return "", err
	}
	if err := <-assignChildChannel; err != nil {
		return "", err
	}

	return parentItem.SKU, nil
}

func buildPromptInputs(combinations []gui_structs.Combination) ([]string, string, []string) {
	productNames := make([]string, 0, len(combinations))
	variationLabels := "["
	oldSKUs := make([]string, 0, len(combinations))

	for _, c := range combinations {
		productNames = append(productNames, c.Item.GuiItem.Name)
		variationLabels += fmt.Sprintf("[%s], ", c.Label)
		oldSKUs = append(oldSKUs, c.Item.GuiItem.SKU)
	}
	if len(variationLabels) >= 2 {
		variationLabels = variationLabels[:len(variationLabels)-2]
	}
	variationLabels += "]"

	return productNames, variationLabels, oldSKUs
}

func generateSEO(productNames []string, selectedDescription string, variationLabels string, oldSKUs []string) (*openai_structs.ProductSEO, error) {
	userPrompt := openai.GetUserPromptText(productNames, selectedDescription, variationLabels, oldSKUs)
	ctx := context.Background()
	return openai.MakeRequestText(ctx, userPrompt)
}

func collectItemsFromCombinations(combinations []gui_structs.Combination) []wawi_structs.GetItem {
	items := make([]wawi_structs.GetItem, 0, len(combinations))
	for _, c := range combinations {
		items = append(items, c.Item.GetItem)
	}
	return items
}

func setActiveSalesChannels(itemID int, channels []string) error {
	return SetItemActiveSalesChannels(strconv.Itoa(itemID), channels)
}

func collectUniquePropertyValueIDs(combinations []gui_structs.Combination) ([]string, error) {
	var propertyValueIds []string
	for _, c := range combinations {
		properties, err := QueryItemProperties(strconv.Itoa(c.Item.GetItem.ID))
		if err != nil {
			return nil, err
		}
		for _, property := range properties.Properties {
			propertyValueIds = append(propertyValueIds, strconv.Itoa(property.PropertyValueId))
		}
	}
	return uniqueStrings(propertyValueIds), nil
}

func createVariationsAndValues(itemID int, variationTree map[string][]string) ([]string, map[string]map[string]string, error) {
	var variationOrder []string
	valueIDByLabel := map[string]map[string]string{}

	for parentName, children := range variationTree {
		parentVariation, err := CreateVariations(strconv.Itoa(itemID), parentName)
		if err != nil {
			return nil, nil, err
		}
		variationOrder = append(variationOrder, parentName)
		if _, ok := valueIDByLabel[parentName]; !ok {
			valueIDByLabel[parentName] = map[string]string{}
		}
		for _, childName := range children {
			v, err := CreateVariationValue(strconv.Itoa(itemID), strconv.Itoa(parentVariation.Id), childName)
			if err != nil {
				return nil, nil, err
			}
			valueIDByLabel[parentName][childName] = strconv.Itoa(v.Id)
		}
	}
	return variationOrder, valueIDByLabel, nil
}

func assignChildrenToParent(
	parentItemID int,
	combinations []gui_structs.Combination,
	variationTree map[string][]string,
	variationOrder []string,
	valueIDByLabel map[string]map[string]string,
) error {
	for _, combination := range combinations {
		namesByLabel := parseNamesByLabel(combination.Label)

		var comboValueIDs []string
		for _, vLabel := range variationOrder {
			name := namesByLabel[vLabel]

			// Fallback 1: detect value in label text
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

			// Fallback 2: detect value in item name
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

			// Fallback 3: only one possible value
			if name == "" {
				if children, ok := variationTree[vLabel]; ok && len(children) == 1 {
					name = children[0]
				}
			}

			id, ok := valueIDByLabel[vLabel][name]
			if !ok {
				return fmt.Errorf("missing value ID for %s=%s", vLabel, name)
			}
			comboValueIDs = append(comboValueIDs, id)
		}

		if err := AssignChildToParent(
			strconv.Itoa(parentItemID),
			strconv.Itoa(combination.Item.GetItem.ID),
			comboValueIDs,
		); err != nil {
			return err
		}
	}
	return nil
}

func parseNamesByLabel(raw string) map[string]string {
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

func containsCI(haystack, needle string) bool {
	return strings.Contains(strings.ToLower(haystack), strings.ToLower(needle))
}

func PtrIfSet[T comparable](v T) *T {
	var zero T
	if v == zero {
		return nil
	}
	return &v
}

// Try reading an image as .jpg first, then .png. Returns bytes and picked extension ("jpg" or "png").
func readImageBytesWithFallback(basePath string) ([]byte, string, error) {
	jpgPath := basePath + ".jpg"
	data, err := os.ReadFile(jpgPath)
	if err == nil {
		return data, "jpg", nil
	}
	pngPath := basePath + ".png"
	dataPNG, errPNG := os.ReadFile(pngPath)
	if errPNG == nil {
		return dataPNG, "png", nil
	}
	return nil, "", fmt.Errorf("no image found: %s (%v) or %s (%v)", jpgPath, err, pngPath, errPNG)
}

func createParentStruct(seo *openai_structs.ProductSEO, items []wawi_structs.GetItem, newSKU string) wawi_structs.ItemCreate {
	cheapestItemIndex := findCheapestItem(items)
	dangerousStruct := getItemDangerous(items)
	searchTerms := getSearchTerms(items)

	ts := time.Now().UTC().Format(time.RFC3339)

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
			ManufacturerNumber: PtrIfSet(removeUpToFirstDash(newSKU)),
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

func getImagesAndBase64(combinations []gui_structs.Combination) ([]string, error) {
	images := make([]string, 0, len(combinations))

	for _, c := range combinations {
		base := fmt.Sprintf("%s%s-1", PathToFolder, c.Item.GuiItem.SKU)
		data, _, err := readImageBytesWithFallback(base)
		if err != nil {
			fmt.Fprintf(os.Stderr, "no image %s.(jpg|png): %v\n", base, err)
			continue
		}

		if _, _, err := image.Decode(bytes.NewReader(data)); err != nil {
			return nil, fmt.Errorf("decode %s.(jpg|png): %w", base, err)
		}

		b64 := base64.StdEncoding.EncodeToString(data)
		images = append(images, b64)
	}

	return images, nil
}

func uploadCombinedImage(imgB64 string, item wawi_structs.GetItem) error {
	cleaned, err := normalizeBase64(imgB64)
	if err != nil {
		return err
	}

	imgStruct := wawi_structs.CreateImageStruct{
		ImageData:      cleaned,
		Filename:       item.SKU + ".jpg",
		SalesChannelId: "1-1-1",
	}

	if err := CreateItemImage(imgStruct, strconv.Itoa(item.ID)); err != nil {
		return err
	}

	return nil
}

func getAllItemImages(combinations []gui_structs.Combination) ([]wawi_structs.CreateImageStruct, error) {
	images := make([]wawi_structs.CreateImageStruct, 0, len(combinations))

	for _, c := range combinations {
		for i := range 10 {
			base := fmt.Sprintf("%s%s-%v", PathToFolder, c.Item.GetItem.SKU, i+1)
			data, ext, err := readImageBytesWithFallback(base)

			if err != nil && i == 0 {
				fmt.Fprintf(os.Stderr, "no image %s.(jpg|png): %v\n", base, err)
				break
			} else if err != nil {
				break
			}

			if _, _, err := image.Decode(bytes.NewReader(data)); err != nil {
				fmt.Fprintf(os.Stderr, "decode %s.%s failed: %v\n", base, ext, err)
				continue
			}

			b64 := base64.StdEncoding.EncodeToString(data)
			img := wawi_structs.CreateImageStruct{
				ImageData:      b64,
				Filename:       fmt.Sprintf("%s-%v.%s", c.Item.GetItem.SKU, i+1, ext),
				SalesChannelId: "1-1-1",
			}

			images = append(images, img)
		}
	}

	return images, nil
}

func uploadAllItemImages(itemId string, images []wawi_structs.CreateImageStruct) error {
	for _, img := range images {
		if err := CreateItemImage(img, itemId); err != nil {
			return err
		}
	}

	return nil
}

type missingImageInfo struct {
	Name string
	SKU  string
}

func detectMissingPrimaryImages(combinations []gui_structs.Combination) []missingImageInfo {
	missing := make([]missingImageInfo, 0)
	for _, combo := range combinations {
		base := fmt.Sprintf("%s%s-1", PathToFolder, combo.Item.GuiItem.SKU)
		if !imageExists(base+".jpg") && !imageExists(base+".png") {
			missing = append(missing, missingImageInfo{
				Name: combo.Item.GuiItem.Name,
				SKU:  combo.Item.GuiItem.SKU,
			})
		}
	}
	return missing
}

func imageExists(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	}
	return false
}
