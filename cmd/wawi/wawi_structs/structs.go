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
	ChildItems          string         `json:"ChildItems"`
	ParentItemID        int            `json:"ParentItemId"`
	ItemPriceData       ItemPriceData  `json:"ItemPriceData"`
	ActiveSalesChannels string         `json:"ActiveSalesChannels"`
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

type GetItem struct {
	ID int `json:"Id"`
	Item
}

type Category struct {
	CategoryID int    `json:"CategoryId"`
	Name       string `json:"Name"`
}

type Identifiers struct {
	Gtin               string `json:"Gtin"`
	ManufacturerNumber string `json:"ManufacturerNumber"`
	ISBN               string `json:"ISBN"`
	UPC                string `json:"UPC"`
	AmazonFnsku        string `json:"AmazonFnsku"`
	Asins              string `json:"Asins"`
	OwnIdentifier      string `json:"OwnIdentifier"`
}

type Component struct {
	ItemID     int `json:"ItemId"`
	Quantity   int `json:"Quantity"`
	SortNumber int `json:"SortNumber"`
}

type ItemPriceData struct {
	SalesPriceNet        float64 `json:"SalesPriceNet"`
	SuggestedRetailPrice float64 `json:"SuggestedRetailPrice"`
	PurchasePriceNet     float64 `json:"PurchasePriceNet"`
	EbayPrice            float64 `json:"EbayPrice"`
	AmazonPrice          float64 `json:"AmazonPrice"`
}

type StorageOptions struct {
	InventoryManagementActive             bool `json:"InventoryManagementActive"`
	SplitQuantity                         bool `json:"SplitQuantity"`
	GlobalMinimumStockLevel               int  `json:"GlobalMinimumStockLevel"`
	Buffer                                int  `json:"Buffer"`
	SerialNumberItem                      bool `json:"SerialNumberItem"`
	SerialNumberTracking                  bool `json:"SerialNumberTracking"`
	SubjectToShelfLifeExpirationDate      bool `json:"SubjectToShelfLifeExpirationDate"`
	SubjectToBatchItem                    bool `json:"SubjectToBatchItem"`
	ProcurementTime                       int  `json:"ProcurementTime"`
	DetermineProcurementTimeAutomatically bool `json:"DetermineProcurementTimeAutomatically"`
	AdditionalHandlingTime                int  `json:"AdditionalHandlingTime"`
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
	MinimumOrderQuantity                    int                                    `json:"MinimumOrderQuantity"`
	MinimumPurchaseQuantityForCustomerGroup []MinimumPurchaseQuantityCustomerGroup `json:"MinimumPurchaseQuantityForCustomerGroup"`
	PermissibleOrderQuantity                int                                    `json:"PermissibleOrderQuantity"`
}

type MinimumPurchaseQuantityCustomerGroup struct {
	CustomerGroupID          int     `json:"CustomerGroupId"`
	PermissibleOrderQuantity float64 `json:"PermissibleOrderQuantity"`
	MinimumPurchaseQuantity  int     `json:"MinimumPurchaseQuantity"`
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
