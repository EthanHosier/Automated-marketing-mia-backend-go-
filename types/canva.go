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
