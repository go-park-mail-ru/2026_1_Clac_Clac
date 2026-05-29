package api

type VkIDTokenResponse struct {
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	IDToken          string `json:"id_token"`
	TokenType        string `json:"token_type"`
	ExpiresIn        int    `json:"expires_in"`
	UserID           int64  `json:"user_id"`
	State            string `json:"state"`
	Scope            string `json:"scope"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

type VkIDUserInfoResponse struct {
	User VkIDUser `json:"user"`
}

type VkIDUser struct {
	UserID    int64  `json:"user_id"`
	FirstName string `json:"first_name"`
	Email     string `json:"email"`
	Avatar    string `json:"avatar"`
}
