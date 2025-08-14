package wawi_registration

import (
	"WawiIC/cmd/defines"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// https://developer.jtl-software.com/products/erpapi/openapi/appregistration/authenticationheader_registerappasync
func postAppRegistration() (*defines.RegistrationResponse, error) {
	requestData := defines.ConstructAppData()

	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %v", err)
	}

	req, err := http.NewRequest("POST", defines.APIBaseURL+"authentication", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-ChallengeCode", defines.XChallangeCode) // "x-" war zu kurz
	req.Header.Set("X-AppID", defines.AppID)
	req.Header.Set("X-AppVersion", defines.Version)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("registration failed (status %d): %s", resp.StatusCode, string(body))
	}

	var regResp defines.RegistrationResponse
	if err := json.Unmarshal(body, &regResp); err != nil {
		return nil, fmt.Errorf("failed to parse registration response: %v", err)
	}

	return &regResp, nil
}

func waitForRegistrationAcc(registrationID string) (*defines.FetchRegistrationResponse, error) {
	req, err := http.NewRequest("GET", defines.APIBaseURL+"authentication/"+registrationID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("x-challengecode", defines.XChallangeCode)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %v", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("registration failed (status %d): %s", res.StatusCode, string(body))
	}

	var regResp defines.FetchRegistrationResponse
	if err := json.Unmarshal(body, &regResp); err != nil {
		return nil, fmt.Errorf("failed to parse registration response: %v", err)
	}

	return &regResp, nil
}

func Register() (string, error) {
	regResp, err := postAppRegistration()
	if err != nil {
		return "", err
	}

	if regResp.Status == 1 {
		return "", fmt.Errorf("registration failed (status 1): Rejected")
	}

	var waitRet *defines.FetchRegistrationResponse
	regID := regResp.RegistrationRequestId
	for true {
		waitRet, err = waitForRegistrationAcc(regID)

		if err != nil {
			return "", err
		}

		if waitRet.RequestStatusInfo.Status == 1 {
			return "", fmt.Errorf("registration failed (status %d): %s", waitRet.RequestStatusInfo.Status, string(regID))
		}

		if waitRet.RequestStatusInfo.Status == 2 {
			break
		}

		regID = waitRet.RequestStatusInfo.RegistrationRequestId
		time.Sleep(5 * time.Second)
	}

	return waitRet.Token.ApiKey, nil
}
