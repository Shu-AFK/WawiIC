package wawi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/Shu-AFK/WawiIC/cmd/defines"
	"github.com/Shu-AFK/WawiIC/cmd/openai/openai_structs"
	"github.com/Shu-AFK/WawiIC/cmd/wawi/wawi_structs"
	gtf "github.com/bas24/googletranslatefree"
)

func QuerySalesChannels() ([]wawi_structs.SalesChannel, error) {
	var channels []wawi_structs.SalesChannel

	url := fmt.Sprintf("%s/salesChannels", defines.APIBaseURL)
	resp, err := wawiCreateRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET /salesChannels failed: %s", body)
	}

	if err := json.Unmarshal(body, &channels); err != nil {
		return nil, err
	}

	return channels, nil
}

func QueryItem(itemStruct wawi_structs.QueryItemStruct) ([]wawi_structs.GetItem, error) {
	var items []wawi_structs.GetItem
	pageNumber := 1

	for {
		resp, err := queryItemReq(itemStruct, pageNumber)
		if err != nil {
			return nil, err
		}
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()

		if err != nil {
			return nil, err
		}

		var respJSON wawi_structs.ResponseItemReq
		if err = json.Unmarshal(body, &respJSON); err != nil {
			return nil, err
		}

		items = append(items, respJSON.Items...)

		if !respJSON.HasNextPage {
			break
		}
		pageNumber = respJSON.NextPageNumber
	}

	return items, nil
}

func CreateParentItem(item wawi_structs.ItemCreate) (*wawi_structs.GetItem, error) {
	reqUrl := defines.APIBaseURL + "items?disableAutomaticWorkflows=true"
	body, err := json.Marshal(item)
	if err != nil {
		return nil, err
	}

	resp, err := wawiCreateRequest("POST", reqUrl, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		errorBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		resp.Body.Close()

		return nil, fmt.Errorf("failed to create parent item: %v (%v)", resp.StatusCode, string(errorBody))
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var respJSON wawi_structs.GetItem
	err = json.Unmarshal(respBody, &respJSON)
	if err != nil {
		return nil, err
	}

	return &respJSON, nil
}

func SetItemActiveSalesChannels(itemID string, _ []string) error {
	reqUrl := defines.APIBaseURL + "items/" + itemID

	payload := map[string]any{
		"ActiveSalesChannels": []string{"9-7-1-2"},
	}

	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal request failed: %w", err)
	}

	resp, err := wawiCreateRequest("PATCH", reqUrl, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("http error: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated, http.StatusNoContent:
		return nil
	default:
		return fmt.Errorf("activate sales channel failed: %s: %s", resp.Status, string(respBody))
	}
}

func AssignChildToParent(itemIDParent string, itemIDChild string, variationIDs []string) error {
	reqURL := defines.APIBaseURL + "items/" + itemIDParent + "/children/" + itemIDChild
	jsonBody, err := json.Marshal(variationIDs)
	if err != nil {
		return err
	}

	resp, err := wawiCreateRequest("POST", reqURL, bytes.NewReader(jsonBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		errorBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		resp.Body.Close()

		return fmt.Errorf("failed to assign children to parent: %v (%v)", resp.StatusCode, string(errorBody))
	}

	return nil
}

func QueryCategories(pageSize int) ([]wawi_structs.CategoryItem, error) {
	var categories []wawi_structs.CategoryItem
	pageNumber := 1

	for {
		resp, err := queryCategoriesReq(pageSize, pageNumber)
		if err != nil {
			return nil, err
		}
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()

		if err != nil {
			return nil, err
		}

		var respJSON wawi_structs.CategoryResponse
		if err = json.Unmarshal(body, &respJSON); err != nil {
			return nil, err
		}

		categories = append(categories, respJSON.Items...)

		if !respJSON.HasNextPage {
			break
		}
		pageNumber = respJSON.NextPageNumber
	}

	return categories, nil
}

func CreateItemImage(imageStruct wawi_structs.CreateImageStruct, itemID string) error {
	reqURL := defines.APIBaseURL + "items/" + itemID + "/images"
	reqBody, err := json.Marshal(imageStruct)
	if err != nil {
		return err
	}

	resp, err := wawiCreateRequest("POST", reqURL, bytes.NewReader(reqBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		errorBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		resp.Body.Close()

		return fmt.Errorf("failed to create item image: %v (%v)", resp.StatusCode, string(errorBody))
	}

	return nil
}

func CreateVariations(itemID string, variationName string) (*wawi_structs.ReturnVariationCreateStruct, error) {
	reqURL := defines.APIBaseURL + "items/" + itemID + "/variations"
	reqStruct := wawi_structs.CreateVariationStruct{
		Name:         variationName,
		Type:         0,
		Translations: []wawi_structs.Translation{},
	}
	reqBody, err := json.Marshal(reqStruct)
	if err != nil {
		return nil, err
	}

	resp, err := wawiCreateRequest("POST", reqURL, bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		errorBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		resp.Body.Close()

		return nil, fmt.Errorf("failed to create variation: %v (%v)", resp.StatusCode, string(errorBody))
	}

	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, err
	}

	var bodyJson wawi_structs.ReturnVariationCreateStruct
	err = json.Unmarshal(body, &bodyJson)
	if err != nil {
		return nil, err
	}

	return &bodyJson, nil
}

func CreateVariationValue(itemID string, variationID string, name string) (*wawi_structs.ReturnVariationValueCreateStruct, error) {
	nameEn, err := gtf.Translate(name, "de", "en")
	if err != nil {
		return nil, err
	}

	reqUrl := defines.APIBaseURL + "items/" + itemID + "/variations/" + variationID + "/values"
	reqBody, err := json.Marshal(wawi_structs.CreateVariationValueStruct{
		Name: name,
		Translations: []wawi_structs.Translation{
			{
				LanguageIso: "EN",
				Name:        nameEn,
			},
		},
	})

	if err != nil {
		return nil, err
	}

	resp, err := wawiCreateRequest("POST", reqUrl, bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		errorBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("failed to create variation value: %v (%v)", resp.StatusCode, string(errorBody))
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var respJSON wawi_structs.ReturnVariationValueCreateStruct
	err = json.Unmarshal(respBody, &respJSON)
	if err != nil {
		return nil, err
	}

	return &respJSON, nil
}

func UpdateDescription(itemID string, SEO openai_structs.ProductSEO) error {
	// Wawi Description
	salesChannelId := "1-1-1"
	reqUrl := defines.APIBaseURL + "items/" + itemID + "/descriptions/" + salesChannelId + "/de"
	updateBody := wawi_structs.UpdateMetaDesc{
		SeoMetaDescription: SEO.SEODescription,
		SeoTitleTag:        SEO.CombinedArticleName,
		SeoMetaKeywords:    strings.Join(SEO.SEOKeywords, ", "),
	}

	reqBody, err := json.Marshal(updateBody)
	if err != nil {
		return err
	}
	resp, err := wawiCreateRequest("PATCH", reqUrl, bytes.NewReader(reqBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		errorBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("failed to update description: %v (%v)", resp.StatusCode, string(errorBody))
	}

	return nil
}

func QueryItemProperties(itemId string) (*wawi_structs.QueryItemPropertiesStruct, error) {
	reqUrl := defines.APIBaseURL + "items/" + itemId + "/properties"
	resp, err := wawiCreateRequest("GET", reqUrl, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errorBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("failed to query item properties: %v (%v)", resp.StatusCode, string(errorBody))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var respJSON wawi_structs.QueryItemPropertiesStruct
	err = json.Unmarshal(body, &respJSON)
	if err != nil {
		return nil, err
	}

	return &respJSON, nil
}

func CreateItemProperty(itemId string, propertyValueId string) (*wawi_structs.Property, error) {
	reqUrl := defines.APIBaseURL + "items/" + itemId + "/properties"
	reqBody := "{\"PropertyValueId\":\"" + propertyValueId + "\"}"
	resp, err := wawiCreateRequest("POST", reqUrl, bytes.NewReader([]byte(reqBody)))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		errorBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("failed to create item property: %v (%v)", resp.StatusCode, string(errorBody))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var respJSON wawi_structs.Property
	err = json.Unmarshal(body, &respJSON)
	if err != nil {
		return nil, err
	}

	return &respJSON, nil
}

func QuerySuppliers() ([]wawi_structs.Suppliers, error) {
	reqUrl := defines.APIBaseURL + "suppliers"
	resp, err := wawiCreateRequest("GET", reqUrl, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errorBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("failed to query suppliers: %v (%v)", resp.StatusCode, string(errorBody))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var suppliers []wawi_structs.Suppliers
	err = json.Unmarshal(body, &suppliers)
	if err != nil {
		return nil, err
	}

	return suppliers, nil
}

func queryCategoriesReq(pageSize int, pageNumber int) (*http.Response, error) {
	if pageSize == 0 {
		return nil, fmt.Errorf("pageSize must be greater than zero")
	}

	reqURL := defines.APIBaseURL + "categories?pageNumber=" + strconv.Itoa(pageNumber) + "&pageSize=" + strconv.Itoa(pageSize)
	resp, err := wawiCreateRequest("GET", reqURL, nil)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		errorBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		resp.Body.Close()

		return nil, fmt.Errorf("failed to query categories: %v (%v)", resp.StatusCode, string(errorBody))
	}

	return resp, nil
}

func queryItemReq(itemStruct wawi_structs.QueryItemStruct, pageNumber int) (*http.Response, error) {
	if itemStruct.PageSize == 0 {
		return nil, fmt.Errorf("no page size provided")
	}

	baseURL, err := url.Parse(defines.APIBaseURL + "items")
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}

	params := url.Values{}
	params.Set("pageNumber", strconv.Itoa(pageNumber))
	params.Set("pageSize", strconv.Itoa(itemStruct.PageSize))

	if itemStruct.SearchKeyword != "" {
		params.Set("searchKeyWord", itemStruct.SearchKeyword)
	}
	if itemStruct.ItemSupplier != "" {
		params.Set("kHersteller", itemStruct.ItemSupplier)
	}
	if itemStruct.ItemCategory != "" {
		params.Set("kKategorie", itemStruct.ItemCategory)
	}
	if itemStruct.ItemID != "" {
		params.Set("id", itemStruct.ItemID)
	}

	baseURL.RawQuery = params.Encode()

	resp, err := wawiCreateRequest("GET", baseURL.String(), nil)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		errorBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		resp.Body.Close()

		return nil, fmt.Errorf("failed to query item: %v (%v)", resp.StatusCode, string(errorBody))
	}

	return resp, nil
}

func wawiCreateRequest(method string, url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	apiKey := os.Getenv(defines.APIKeyVarName)
	if apiKey == "" {
		return nil, fmt.Errorf("API key environment variable not set")
	}

	req.Header.Set("Authorization", fmt.Sprintf("Wawi %v", apiKey))
	req.Header.Set("x-appid", defines.AppID)
	req.Header.Set("x-runas", defines.AppID)
	req.Header.Set("api-version", defines.APIVersion)
	req.Header.Set("x-appversion", defines.Version)

	if method == "POST" || method == "PATCH" {
		req.Header.Set("Content-Type", "application/json")
	}

	fmt.Printf("Made request to: %s\n", req.URL)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
