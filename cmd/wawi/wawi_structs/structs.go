package wawi_structs

type QueryItemStruct struct {
	SearchKeyword string
	ItemSupplier  string
	ItemCategory  string
	ItemID        string
	PageSize      int
}

type ResponseItemReq struct {
	TotalItems         int       `json:"TotalItems"`
	PageNumber         int       `json:"PageNumber"`
	PageSize           int       `json:"PageSize"`
	Items              []GetItem `json:"Items"`
	TotalPages         int       `json:"TotalPages"`
	HasPreviousPage    bool      `json:"HasPreviousPage"`
	HasNextPage        bool      `json:"HasNextPage"`
	NextPageNumber     int       `json:"NextPageNumber"`
	PreviousPageNumber int       `json:"PreviousPageNumber"`
}

type Item struct {
	SKU                 string         `json:"SKU"`
	ManufacturerID      int            `json:"ManufacturerId"`
	ResponsiblePersonID int            `json:"ResponsiblePersonId"`
	IsActive            bool           `json:"IsActive"`
	Categories          []Category     `json:"Categories"`
	Name                string         `json:"Name"`
	Description         string         `json:"Description"`
	ShortDescription    string         `json:"ShortDescription"`
	Identifiers         Identifiers    `json:"Identifiers"`
	Components          []Component    `json:"Components"`
	ChildItems          []int          `json:"ChildItems"`
	ParentItemID        int            `json:"ParentItemId"`
	ItemPriceData       ItemPriceData  `json:"ItemPriceData"`
	ActiveSalesChannels []string       `json:"ActiveSalesChannels"`
	SortNumber          int            `json:"SortNumber"`
	Annotation          string         `json:"Annotation"`
	Added               string         `json:"Added"`
	Changed             string         `json:"Changed"`
	ReleasedOnDate      string         `json:"ReleasedOnDate"`
	StorageOptions      StorageOptions `json:"StorageOptions"`
	CountryOfOrigin     string         `json:"CountryOfOrigin"`
	ConditionID         int            `json:"ConditionId"`
	ShippingClassID     int            `json:"ShippingClassId"`
	ProductGroupID      int            `json:"ProductGroupId"`
	TaxClassID          int            `json:"TaxClassId"`
	Dimensions          Dimensions     `json:"Dimensions"`
	Weights             Weights        `json:"Weights"`
	AllowNegativeStock  bool           `json:"AllowNegativeStock"`
	Quantities          Quantities     `json:"Quantities"`
	DangerousGoods      DangerousGoods `json:"DangerousGoods"`
	Taric               string         `json:"Taric"`
	SearchTerms         string         `json:"SearchTerms"`
	PriceListActive     bool           `json:"PriceListActive"`
	IgnoreDiscounts     bool           `json:"IgnoreDiscounts"`
	AvailabilityID      int            `json:"AvailabilityId"`
}

type ItemCreate struct {
	SKU                 string          `json:"SKU"`
	ManufacturerID      *int            `json:"ManufacturerId,omitempty"`
	ResponsiblePersonID *int            `json:"ResponsiblePersonId,omitempty"`
	IsActive            bool            `json:"IsActive"`
	Categories          []Category      `json:"Categories,omitempty"`
	Name                string          `json:"Name"`
	Description         string          `json:"Description,omitempty"`
	ShortDescription    string          `json:"ShortDescription,omitempty"`
	Identifiers         *Identifiers    `json:"Identifiers,omitempty"`
	Components          []Component     `json:"Components,omitempty"`
	ChildItems          []string        `json:"ChildItems,omitempty"`
	ParentItemID        *int            `json:"ParentItemId,omitempty"`
	ItemPriceData       *ItemPriceData  `json:"ItemPriceData,omitempty"`
	ActiveSalesChannels []string        `json:"ActiveSalesChannels,omitempty"`
	SortNumber          *int            `json:"SortNumber,omitempty"`
	Annotation          string          `json:"Annotation,omitempty"`
	Added               string          `json:"Added,omitempty"`
	Changed             string          `json:"Changed,omitempty"`
	ReleasedOnDate      string          `json:"ReleasedOnDate,omitempty"`
	StorageOptions      *StorageOptions `json:"StorageOptions,omitempty"`
	CountryOfOrigin     string          `json:"CountryOfOrigin,omitempty"`
	ConditionID         *int            `json:"ConditionId,omitempty"`
	ShippingClassID     *int            `json:"ShippingClassId,omitempty"`
	ProductGroupID      *int            `json:"ProductGroupId,omitempty"`
	TaxClassID          *int            `json:"TaxClassId,omitempty"`
	Dimensions          *Dimensions     `json:"Dimensions,omitempty"`
	Weights             *Weights        `json:"Weights,omitempty"`
	AllowNegativeStock  *bool           `json:"AllowNegativeStock,omitempty"`
	Quantities          *Quantities     `json:"Quantities,omitempty"`
	DangerousGoods      *DangerousGoods `json:"DangerousGoods,omitempty"`
	Taric               string          `json:"Taric,omitempty"`
	SearchTerms         string          `json:"SearchTerms,omitempty"`
	PriceListActive     bool            `json:"PriceListActive"`
	IgnoreDiscounts     *bool           `json:"IgnoreDiscounts,omitempty"`
	AvailabilityID      *int            `json:"AvailabilityId,omitempty"`
}

type GetItem struct {
	ID int `json:"Id"`
	Item
}

type Category struct {
	CategoryID int    `json:"CategoryId"`
	Name       string `json:"Name"`
}

type Identifiers struct {
	Gtin               *string   `json:"Gtin,omitempty"`
	ManufacturerNumber *string   `json:"ManufacturerNumber,omitempty"`
	ISBN               *string   `json:"ISBN,omitempty"`
	UPC                *string   `json:"UPC,omitempty"`
	AmazonFnsku        *string   `json:"AmazonFnsku,omitempty"`
	Asins              *[]string `json:"Asins,omitempty"`
	OwnIdentifier      *string   `json:"OwnIdentifier,omitempty"`
}

type Component struct {
	ItemID     int `json:"ItemId"`
	Quantity   int `json:"Quantity"`
	SortNumber int `json:"SortNumber"`
}

type ItemPriceData struct {
	SalesPriceNet        *float64 `json:"SalesPriceNet,omitempty"`
	SuggestedRetailPrice *float64 `json:"SuggestedRetailPrice,omitempty"`
	PurchasePriceNet     *float64 `json:"PurchasePriceNet,omitempty"`
	EbayPrice            *float64 `json:"EbayPrice,omitempty"`
	AmazonPrice          *float64 `json:"AmazonPrice,omitempty"`
}

type StorageOptions struct {
	InventoryManagementActive             bool    `json:"InventoryManagementActive"`
	SplitQuantity                         bool    `json:"SplitQuantity"`
	GlobalMinimumStockLevel               float32 `json:"GlobalMinimumStockLevel"`
	Buffer                                int     `json:"Buffer"`
	SerialNumberItem                      bool    `json:"SerialNumberItem"`
	SerialNumberTracking                  bool    `json:"SerialNumberTracking"`
	SubjectToShelfLifeExpirationDate      bool    `json:"SubjectToShelfLifeExpirationDate"`
	SubjectToBatchItem                    bool    `json:"SubjectToBatchItem"`
	ProcurementTime                       int     `json:"ProcurementTime"`
	DetermineProcurementTimeAutomatically bool    `json:"DetermineProcurementTimeAutomatically"`
	AdditionalHandlingTime                int     `json:"AdditionalHandlingTime"`
}

type Dimensions struct {
	Length float64 `json:"Length"`
	Width  float64 `json:"Width"`
	Height float64 `json:"Height"`
}

type Weights struct {
	ItemWeight     float64 `json:"ItemWeigth"` // Note: typo in JSON ("Weigth")
	ShippingWeight float64 `json:"ShippingWeight"`
}

type Quantities struct {
	MinimumOrderQuantity                    float32                                `json:"MinimumOrderQuantity"`
	MinimumPurchaseQuantityForCustomerGroup []MinimumPurchaseQuantityCustomerGroup `json:"MinimumPurchaseQuantityForCustomerGroup"`
	PermissibleOrderQuantity                float32                                `json:"PermissibleOrderQuantity"`
}

type MinimumPurchaseQuantityCustomerGroup struct {
	CustomerGroupID          int     `json:"CustomerGroupId"`
	PermissibleOrderQuantity float64 `json:"PermissibleOrderQuantity"`
	MinimumPurchaseQuantity  float32 `json:"MinimumPurchaseQuantity"`
	IsActive                 bool    `json:"IsActive"`
}

type DangerousGoods struct {
	UnNumber string `json:"UnNumber"`
	HazardNo string `json:"HazardNo"`
}

type CategoryItem struct {
	ID                  int      `json:"Id"`
	Name                string   `json:"Name"`
	Description         string   `json:"Description"`
	ParentCategoryID    int      `json:"ParentCategoryId"`
	Level               int      `json:"Level"`
	SortNumber          int      `json:"SortNumber"`
	ActiveSalesChannels []string `json:"ActiveSalesChannels"`
}

type CategoryResponse struct {
	TotalItems         int            `json:"TotalItems"`
	PageNumber         int            `json:"PageNumber"`
	PageSize           int            `json:"PageSize"`
	Items              []CategoryItem `json:"Items"`
	TotalPages         int            `json:"TotalPages"`
	HasPreviousPage    bool           `json:"HasPreviousPage"`
	HasNextPage        bool           `json:"HasNextPage"`
	NextPageNumber     int            `json:"NextPageNumber"`
	PreviousPageNumber int            `json:"PreviousPageNumber"`
}

type ItemImageReq struct {
	ItemId         int    `json:"ItemId"`
	ImageId        int    `json:"ImageId"`
	Filename       string `json:"Filename"`
	ImageDataType  string `json:"ImageDataType"`
	SalesChannelId string `json:"SalesChannelId"`
	EbayUserName   string `json:"EbayUserName"`
	SortNumber     int    `json:"SortNumber"`
	Size           int    `json:"Size"`
	Width          int    `json:"Width"`
	Height         int    `json:"Height"`
}

type CreateImageStruct struct {
	ImageData      string `json:"ImageData"`
	Filename       string `json:"Filename"`
	SalesChannelId string `json:"SalesChannelId"`
}

type GuiItem struct {
	SKU      string
	Name     string
	IsFather bool
	IsChild  bool
	Combine  bool
}

type WItem struct {
	GuiItem GuiItem
	GetItem GetItem
}

type CreateVariationStruct struct {
	Name         string        `json:"Name"`
	Type         int           `json:"Type"`
	Translations []Translation `json:"Translations"`
}

type ReturnVariationCreateStruct struct {
	Id           int           `json:"Id"`
	ItemId       int           `json:"ItemId"`
	Name         string        `json:"Name"`
	Type         int           `json:"Type"`
	Translations []Translation `json:"Translations"`
}

type Translation struct {
	LanguageIso string `json:"LanguageIso"`
	Name        string `json:"Name"`
}

type CreateVariationValueStruct struct {
	Name         string        `json:"Name"`
	Translations []Translation `json:"Translations"`
}

type ReturnVariationValueCreateStruct struct {
	Id           int           `json:"Id"`
	Name         string        `json:"Name"`
	Translations []Translation `json:"Translations,omitempty"`
}

type UpdateMetaDesc struct {
	Name               string `json:"Name,omitempty"`
	Description        string `json:"Description,omitempty"`
	ShortDescription   string `json:"ShortDescription,omitempty"`
	SeoPath            string `json:"SeoPath,omitempty"`
	SeoMetaDescription string `json:"SeoMetaDescription,omitempty"`
	SeoTitleTag        string `json:"SeoTitleTag,omitempty"`
	SeoMetaKeywords    string `json:"SeoMetaKeywords,omitempty"`
}

type SalesChannel struct {
	Id               string `json:"Id"`
	Type             int    `json:"Type"`
	Name             string `json:"Name"`
	DocumentationUrl string `json:"DocumentationUrl"`
	ItemCapabilities struct {
		Descriptions         bool `json:"Descriptions"`
		OnlineShopActivation bool `json:"OnlineShopActivation"`
		Prices               bool `json:"Prices"`
		SpecialPrices        bool `json:"SpecialPrices"`
		Images               bool `json:"Images"`
	} `json:"ItemCapabilities"`
	CategoryCapabilities struct {
		Descriptions         bool `json:"Descriptions"`
		OnlineShopActivation bool `json:"OnlineShopActivation"`
	} `json:"CategoryCapabilities"`
}
