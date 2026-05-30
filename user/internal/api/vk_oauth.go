package api

// VkIDTokenResponse — ответ от POST /oauth2/auth (токен-эндпоинт).
// Внимание: user_id здесь — число (797329160), в отличие от /user_info, где user_id — строка.
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

// VkIDUserInfoResponse — ответ от POST /oauth2/user_info.
type VkIDUserInfoResponse struct {
	User VkIDUser `json:"user"`
}

// VkIDUser — вложенная структура ответа /user_info.
// Внимание: user_id здесь — строка ("797329160"), в отличие от /oauth2/auth.
type VkIDUser struct {
	UserID    string `json:"user_id"`
	FirstName string `json:"first_name"`
	Email     string `json:"email"`
	Avatar    string `json:"avatar"`
}
