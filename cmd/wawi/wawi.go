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
	"strconv"
)

func QueryItem(itemStruct wawi_structs.QueryItemStruct) ([]wawi_structs.GetItem, error) {
	var items []wawi_structs.GetItem
	pageNumber := 0

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

	req, err := wawiCreateRequest("POST", reqUrl, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("create item failed (status %d): %s", resp.StatusCode, string(body))
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

	req, err := wawiCreateRequest("POST", reqURL, bytes.NewReader(jsonBody))
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("failed to assign child(%s) to parent item(%s): %v", itemIDParent, itemIDChild, resp.StatusCode)
	}

	return nil
}

func QueryCategories(pageSize int) ([]wawi_structs.CategoryItem, error) {
	var categories []wawi_structs.CategoryItem
	pageNumber := 0

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

func queryCategoriesReq(pageSize int, pageNumber int) (*http.Response, error) {
	if pageSize == 0 {
		return nil, fmt.Errorf("pageSize must be greater than zero")
	}

	reqURL := defines.APIBaseURL + "categories?pageNumber=" + strconv.Itoa(pageNumber) + "&pageSize=" + strconv.Itoa(pageSize)
	req, err := wawiCreateRequest("GET", reqURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("failed to query categories: %v", resp.StatusCode)
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

	req, err := wawiCreateRequest("GET", baseURL.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("status code %d", resp.StatusCode)
	}

	return resp, nil
}
