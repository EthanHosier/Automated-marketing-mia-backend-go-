package types

type WebsiteData struct {
	Title           string              `json:"title"`
	MetaDescription string              `json:"meta_description"`
	Headings        map[string][]string `json:"headings"`
	Keywords        string              `json:"keywords"`
	Links           []string            `json:"links"`
	Summary         string              `json:"summary"`
	Categories      []string            `json:"categories"`
}
