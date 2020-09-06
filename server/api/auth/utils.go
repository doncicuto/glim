package auth

// Tokens - TODO comment
type Tokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// Response - TODO comment
type Response struct {
	TokenType string  `json:"token_type"`
	ExpiresIn float64 `json:"expires_in"`
	ExpiresOn int64   `json:"expires_on"`
	Tokens
}
