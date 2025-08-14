package wawi

import (
	"WawiIC/cmd/defines"
	"fmt"
	"io"
	"net/http"
	"os"
)

func wawiCreateRequest(method string, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	apiKey := os.Getenv(defines.APIKeyVarName)
	if apiKey == "" {
		return nil, fmt.Errorf("API key environment variable not set")
	}

	req.Header.Set("Authorization", apiKey)
	req.Header.Set("x-appid", defines.AppID)
	req.Header.Set("x-appversion", defines.Version)

	if method == "POST" {
		req.Header.Set("Content-Type", "application/json")
	}

	return req, nil
}
