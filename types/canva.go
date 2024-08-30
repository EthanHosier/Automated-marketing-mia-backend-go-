package types

type Token struct {
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

type JobStatus struct {
	Job struct {
		ID     string `json:"id"`
		Result struct {
			Type   string `json:"type"`
			Design struct {
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
			} `json:"design"`
		} `json:"result"`
		Status string `json:"status"`
	} `json:"job"`
}
