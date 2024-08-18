package types

type BedrockRequest struct {
	Prompt      string  `json:"prompt"`
	MaxTokens   int     `json:"max_gen_len,omitempty"`
	Temperature float64 `json:"temperature,omitempty"`
}

type BedrockResponse struct {
	PromptTokenCount     int    `json:"prompt_token_count"`
	GenerationTokenCount int    `json:"generation_token_count"`
	StopReason           string `json:"stop_reason"`
	Generation           string `json:"generation"`
}
