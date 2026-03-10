package api

// LogInRequest содержит данные для аутентификации
//
// @Description Данные для входа в систему
type LogInRequest struct {
	Email    string `json:"email"    example:"user@example.com"`
	Password string `json:"password" example:"p@ssword123"`
}

// RegisterRequest содержит данные для создания нового аккаунта
//
// @Description Модель регистрации нового пользователя
type RegisterRequest struct {
	DisplayName      string `json:"display_name"      example:"Ivan Ivanov"`
	Email            string `json:"email"             example:"ivan@mail.com"`
	Password         string `json:"password"          example:"securePassword"`
	RepeatedPassword string `json:"repeated_password" example:"securePassword"`
}

// PasswordRecoveryRequest используется для инициации сброса пароля
//
// @Description Запрос восстановления пароля через Email
type PasswordRecoveryRequest struct {
	Email string `json:"email" example:"user@example.com"`
}

// RecoveryCodeRequest используется для проверки кода подтверждения
//
// @Description Код подтверждения из письма
type RecoveryCodeRequest struct {
	Code string `json:"code" example:"123456"`
}

// NewPasswordRequest используется для установки нового пароля
//
// @Description Установка нового пароля после проверки токена
type NewPasswordRequest struct {
	TokenID          string `json:"token_id"          example:"uuid-token-string"`
	Password         string `json:"password"          example:"new_password_123"`
	RepeatedPassword string `json:"repeated_password" example:"new_password_123"`
}
