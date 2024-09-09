package canva

type Tokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

type Job struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

type UpdateTemplateResponse struct {
	Job Job `json:"job"`
}

type Design struct {
	CreatedAt int64  `json:"created_at"` // Use int64 for Unix timestamp
	ID        string `json:"id"`
	Title     string `json:"title"`
	UpdatedAt int64  `json:"updated_at"` // Use int64 for Unix timestamp
	Thumbnail struct {
		URL string `json:"url"`
	} `json:"thumbnail"`
	URL  string `json:"url"`
	URLs struct {
		EditURL string `json:"edit_url"`
		ViewURL string `json:"view_url"`
	} `json:"urls"`
}

// Define the Result struct
type UpdateTemplateResult struct {
	Type   string `json:"type"`
	Design Design `json:"design"`
}

// Define the Job struct
type UpdateTemplateJob struct {
	ID     string               `json:"id"`
	Result UpdateTemplateResult `json:"result"`
	Status string               `json:"status"`
}

// Define the UpdateTemplateJobStatus struct that references the above types
type UpdateTemplateJobStatus struct {
	Job UpdateTemplateJob `json:"job"`
}

type UploadAssetResponse struct {
	Job UploadAssetJob `json:"job"`
}

type UploadAssetJob struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Asset  Asset  `json:"asset"`
}

type Asset struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Tags      []string  `json:"tags"`
	CreatedAt int64     `json:"created_at"`
	UpdatedAt int64     `json:"updated_at"`
	Thumbnail Thumbnail `json:"thumbnail"`
}

type Thumbnail struct {
	Width  int    `json:"width"`
	Height int    `json:"height"`
	URL    string `json:"url"`
}

type ImageField struct {
	Name    string
	AssetId string
}

type TextField struct {
	Name string
	Text string
}

type ColorField struct {
	Name         string
	ColorAssetId string
}
