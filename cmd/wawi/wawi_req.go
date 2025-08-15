package wawi

import (
	"WawiIC/cmd/defines"
	"WawiIC/cmd/wawi/wawi_structs"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

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

func CreateParentItem(item wawi_structs.Item) (*wawi_structs.GetItem, error) {
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

		return nil, fmt.Errorf("failed to query categories: %v (%v)", resp.StatusCode, string(errorBody))
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

		return fmt.Errorf("failed to query categories: %v (%v)", resp.StatusCode, string(errorBody))
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

func QueryItemImages(itemID string) (*[]wawi_structs.ItemImageReq, error) {
	reqURL := defines.APIBaseURL + "items/" + itemID + "/images"
	resp, err := wawiCreateRequest("GET", reqURL, nil)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get item images failed (status %d): %s", resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, err
	}

	var bodyJSON []wawi_structs.ItemImageReq
	err = json.Unmarshal(body, &bodyJSON)
	if err != nil {
		return nil, err
	}

	return &bodyJSON, nil
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

		return fmt.Errorf("failed to query categories: %v (%v)", resp.StatusCode, string(errorBody))
	}

	return nil
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
	if itemStruct.ItemCategory == "" {
		return nil, fmt.Errorf("no item category provided")
	}

	baseURL, err := url.Parse(defines.APIBaseURL + "items")
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}

	params := url.Values{}
	params.Set("kKategorie", itemStruct.ItemCategory)
	params.Set("pageNumber", strconv.Itoa(pageNumber))
	params.Set("pageSize", strconv.Itoa(itemStruct.PageSize))

	if itemStruct.SearchKeyword != "" {
		params.Set("searchKeyWord", itemStruct.SearchKeyword)
	}
	if itemStruct.ItemSupplier != "" {
		params.Set("kHersteller", itemStruct.ItemSupplier)
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

		return nil, fmt.Errorf("failed to query categories: %v (%v)", resp.StatusCode, string(errorBody))
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
	req.Header.Set("x-appversion", defines.Version)
	req.Header.Set("x-runas", defines.AppID)

	if method == "POST" || method == "PATCH" {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
