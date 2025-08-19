package wawi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
)

type ConfigEntry struct {
	Category    string `json:"category"`
	ShopWebsite string `json:"shop website"`
}

var config []ConfigEntry

func LoadConfig(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}

	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("parse JSON: %w", err)
	}

	if len(config) == 0 {
		return errors.New("config must contain at least one entry")
	}

	for i, e := range config {
		e.Category = strings.TrimSpace(e.Category)
		e.ShopWebsite = strings.TrimSpace(e.ShopWebsite)

		if e.Category == "" {
			return fmt.Errorf("entry %d: category must not be empty", i)
		}
		u, err := url.Parse(e.ShopWebsite)
		if err != nil || u.Scheme == "" || u.Host == "" {
			return fmt.Errorf("entry %d: shop website must be a valid URL", i)
		}

		if u.Scheme != "https" {
			fmt.Fprintf(os.Stderr, "warning: entry %d uses non-HTTPS URL: %s\n", i, e.ShopWebsite)
		}

		config[i] = e
	}

	return nil
}
