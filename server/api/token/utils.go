package token

// Response - TODO comment
type Response struct {
	TokenType    string  `json:"token_type"`
	AccessToken  string  `json:"access_token"`
	ExpiresIn    float64 `json:"expires_in"`
	ExpiresOn    int64   `json:"expires_on"`
	RefreshToken string  `json:"refresh_token"`
}
