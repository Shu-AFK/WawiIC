package defines

const (
	APIBaseURL      = "https://developer.jtl-software.com/_mock/products/erpapi/openapi/"
	AppID           = "WawiIC/v1"
	DisplayName     = "WawiIC"
	Description     = "Artikel zu Vaterartikeln zusammenführen"
	Version         = "1.0.0"
	ProviderName    = "Floyd Göttsch"
	ProviderWebsite = ""
	XChallangeCode  = "wh5x1kgdm2koqsc31rfly3s"
	APIKeyVarName   = "WAWIIC_APIKEY"
)

var MandatoryAPIScope = []string{
	"item.getitem",
	"all.read",
	"item.updateitem",
	"items.write",
	"item.assignchilditemtoparent",
}

type AppData struct {
	AppId              string   `json:"AppId"`
	DisplayName        string   `json:"DisplayName"`
	Description        string   `json:"Description"`
	Version            string   `json:"Version"`
	ProviderName       string   `json:"ProviderName"`
	ProviderWebsite    string   `json:"ProviderWebsite"`
	MandatoryApiScopes []string `json:"MandatoryApiScopes"`
	AppIcon            string   `json:"AppIcon"`
}

type RegistrationResponse struct {
	AppID                 string `json:"AppId"`
	RegistrationRequestId string `json:"RegistrationRequestId"`
	Status                int    `json:"Status"`
}

type FetchRegistrationResponse struct {
	RequestStatusInfo RequestStatus `json:"RequestStatusInfo"`
	Token             TokenInfo     `json:"Token"`
	GrantedScopes     string        `json:"GrantedScopes"`
}

type RequestStatus struct {
	AppId                 string `json:"AppId"`
	RegistrationRequestId string `json:"RegistrationRequestId"`
	Status                int    `json:"Status"`
}

type TokenInfo struct {
	ApiKey string `json:"ApiKey"`
}

func ConstructAppData() *AppData {
	return &AppData{
		AppId:              AppID,
		DisplayName:        DisplayName,
		Description:        Description,
		Version:            Version,
		ProviderName:       ProviderName,
		ProviderWebsite:    ProviderWebsite,
		MandatoryApiScopes: MandatoryAPIScope,
		AppIcon:            "",
	}
}
