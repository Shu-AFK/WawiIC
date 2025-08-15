package defines

const (
	APIBaseURL      = "http://127.0.0.1:5883/api/eazybusiness/"
	AppID           = "WawiIC/v1"
	DisplayName     = "WawiIC"
	Description     = "Artikel zu Vaterartikeln zusammenführen"
	Version         = "1.0.0"
	ProviderName    = "Floyd Göttsch"
	ProviderWebsite = "https://www.alpa-industrievertretungen.de/"
	XChallangeCode  = "wh5x1kgdm2koqsc31rfly3s"
	APIKeyVarName   = "WAWIIC_APIKEY"
	APIVersion      = "1.1"
)

var MandatoryAPIScope = []string{
	"item.getitem",
	"category.querycategories",
	"all.read",
	"item.queryitemimages",
	"item.createitemimage",
	"item.updateitem",
	"item.createitem",
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
	GrantedScopes     []string      `json:"GrantedScopes"`
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
