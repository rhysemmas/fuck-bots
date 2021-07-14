package spotify

// Token contains a response from spotify's token endpoint
type Token struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
}

// Refresh contains a refresh token and time until token expiry
type Refresh struct {
	RefreshToken string
	ExpiresIn    int
}

// Playlist contains a response from spotify's playlist endpoint
type Playlist struct {
	Name string `json:"name"`
}
