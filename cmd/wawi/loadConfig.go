package wawi

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/Shu-AFK/WawiIC/cmd/defines"
)

type configRoot struct {
	ApiBaseURL           string `json:"api base url"`
	ApiVersion           string `json:"api version"`
	SearchMode           string `json:"search mode"`
	CategoryID           string `json:"category id"`
	PathToFolder         string `json:"path to image folder"`
	ActivateSalesChannel bool   `json:"activate sales channel"`
}

var SearchMode string
var PathToFolder string
var ActivateSalesChannel bool
var categoryID int

func LoadConfig(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}

	var root configRoot
	if err := json.Unmarshal(data, &root); err != nil {
		return fmt.Errorf("parse JSON: %w", err)
	}

	defines.APIBaseURL = strings.TrimSpace(root.ApiBaseURL)
	if defines.APIBaseURL == "" {
		return errors.New("api base url must not be empty")
	}
	categoryID, err = strconv.Atoi(strings.TrimSpace(root.CategoryID))

	if root.ApiVersion == "" {
		return errors.New("api version must not be empty")
	}
	defines.APIVersion = strings.TrimSpace(root.ApiVersion)

	SearchMode = strings.TrimSpace(root.SearchMode)
	if SearchMode != "category" && SearchMode != "supplier" && SearchMode != "none" {
		return fmt.Errorf("search mode must be either 'category', 'supplier' or 'none' but was '%s'", SearchMode)
	}

	PathToFolder = strings.TrimSpace(root.PathToFolder)
	if PathToFolder == "" {
		return errors.New("path to image folder must not be empty")
	}

	ActivateSalesChannel = false
	ActivateSalesChannel = root.ActivateSalesChannel

	return nil
}
