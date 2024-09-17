package storage

type TemplateFields struct {
	Name          string `json:"name"`
	Type          string `json:"type"`
	Label         string `json:"label"`
	Comment       string `json:"comment"`
	Page          string `json:"page"`
	MaxCharacters int    `json:"maxCharacters"`
}

type ColorField struct {
	Name    string `json:"name"`
	Label   string `json:"label"`
	Comment string `json:"comment"`
}

type Template struct {
	ID          string           `json:"id"`
	Title       string           `json:"title"`
	Platforms   []string         `json:"platforms"`
	ExportType  string           `json:"export_type"`
	Description string           `json:"description"`
	Fields      []TemplateFields `json:"fields"`
	ColorFields []ColorField     `json:"colors"`
}

type ImageFeature struct {
	ID               string    `json:"id"`
	Feature          string    `json:"feature"`
	FeatureEmbedding []float32 `json:"feature_embedding"`
	UserId           string    `json:"user_id"`
}
