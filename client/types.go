package client

//APIError - TODO comment
type APIError struct {
	Message string `json:"message"`
}

//Credentials - TODO comment
type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// RefreshToken - TODO comment
type RefreshToken struct {
	Token string `json:"refresh_token"`
}

// NewGroup - TODO comment
type NewGroup struct {
	Name        string `json:"username"`
	Description string `json:"password"`
	Members     string `json:"members"`
}
