package openai_structs

type ProductSEO struct {
	SEOKeywords         []string `json:"seo_keywords"`
	SEODescription      string   `json:"seo_description"`
	CombinedArticleName string   `json:"combined_article_name"`
	ShortDescription    string   `json:"short_description"`
	Description         string   `json:"description"`
	NewSKU              string   `json:"new_sku"`
	Image               string   `json:"combined_image_base64"`
}
