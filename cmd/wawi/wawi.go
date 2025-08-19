package wawi

import (
	"context"
	"errors"
	"fmt"

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
		PageSize:      20,
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

func HandleAssignDone(combinations []gui_structs.Combination, selectedCombinationIndex int) error {
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
		return err
	}

	parentItem := createParentStruct(productSEO, combinations[selectedCombinationIndex].Item.GetItem)
	item, err := CreateParentItem(parentItem)
	if err != nil {
		return err
	}
	if item.IsActive == false {
		return errors.New("item is not active")
	}

	// TODO: No possibility to query image data to add to item
	/*
		allImages := make([]wawi_structs.ItemImageReq, 0, len(combinations))
		images, err := QueryItemImages(string(rune(combinations[selectedCombinationIndex].Item.GetItem.ID)))
		if err != nil {
			return err
		}
		allImages = append(allImages, *images...)

		for _, c := range combinations {
			if c.Item.GetItem.ID == combinations[selectedCombinationIndex].Item.GetItem.ID {
				continue
			}

			images, err := QueryItemImages(string(rune(c.Item.GetItem.ID)))
			if err != nil {
				return err
			}
			allImages = append(allImages, *images...)
		}

		for _, image := range allImages {
			err := CreateItemImage(wawi_structs.CreateImageStruct{
				ImageData: ,
				Filename: ,
				SalesChannelId: ,
			}, string(rune(item.ID)))

			if err != nil {
				return err
			}
		}
	*/

	return nil
}

func createParentStruct(seo *openai_structs.ProductSEO, mainItem wawi_structs.GetItem) wawi_structs.Item {
	// TODO: Check what fields are actually necessary
	parentItem := wawi_structs.Item{
		SKU:                 seo.NewSKU,
		ManufacturerID:      mainItem.ManufacturerID,
		ResponsiblePersonID: mainItem.ResponsiblePersonID,
		Categories:          mainItem.Categories,
		Name:                seo.CombinedArticleName,
		Description:         seo.Description,
		ShortDescription:    seo.ShortDescription,
		Identifiers: wawi_structs.Identifiers{
			ManufacturerNumber: mainItem.Identifiers.ManufacturerNumber,
		},
		ActiveSalesChannels: mainItem.ActiveSalesChannels,
		SortNumber:          mainItem.SortNumber,
		Annotation:          mainItem.Annotation,
		CountryOfOrigin:     mainItem.CountryOfOrigin,
		AllowNegativeStock:  false,
		DangerousGoods:      mainItem.DangerousGoods,
		// TODO: PriceListActive?
	}

	return parentItem
}
