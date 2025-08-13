package defines

const (
	APIBaseURL      = ""
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
	OptionalApiScopes  []string `json:"OptionalApiScopes"`
	AppIcon            string   `json:"AppIcon"`
	RegistrationType   int      `json:"RegistrationType"`
	Signature          string   `json:"Signature"`
	SignatureData      string   `json:"SignatureData"`
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
		RegistrationType:   0,
		Signature:          "",
		SignatureData:      "",
	}
}
