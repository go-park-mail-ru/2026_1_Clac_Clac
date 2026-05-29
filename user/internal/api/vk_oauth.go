package api

type VkIDTokenResponse struct {
	AccessToken string `json:"access_token"`
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
