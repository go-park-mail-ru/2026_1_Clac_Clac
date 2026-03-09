package api

type LogInRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	DisplayName      string `json:"display_name"`
	Email            string `json:"email"`
	Password         string `json:"password"`
	RepeatedPassword string `json:"repeated_password"`
}

type PasswordRecoveryRequest struct {
	Email string `json:"email"`
}

type RecoveryCodeRequest struct {
	Code string `json:"code"`
}

type NewPasswordRequest struct {
	TokenID          string `jsin:"token_id"`
	Password         string `json:"password"`
	RepeatedPassword string `json:"repeated_password"`
}
