package storage

import "github.com/ethanhosier/mia-backend-go/canva"

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
	ImageUrl         string    `json:"image_url"`
}

type Similarity[T any] struct {
	Item       T
	Similarity float64
}

type Post struct {
	Platform string       `json:"platform"`
	Caption  string       `json:"caption"`
	Design   canva.Design `json:"design"`
}

type CampaignData struct {
	ResearchReport string `json:"research_report"`
	Posts          []Post `json:"posts"`
	Theme          string `json:"theme"`
	PrimaryKeyword string `json:"primary_keyword"`
}

type Campaign struct {
	ID   string       `json:"id"`
	Data CampaignData `json:"data"`
}
