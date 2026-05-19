package dto

import "github.com/google/uuid"

// BoardInfo содержит основную информацию о доске.
//
//	@Description	Информация о доске
type BoardInfo struct {
	Link        uuid.UUID `json:"link"        example:"123e4567-e89b-12d3-a456-426614174000"`
	Name        string    `json:"name"        example:"My Project"`
	Description string    `json:"description" example:"Доска для трекинга задач"`
	Background  string    `json:"background"  example:"https://s3.example.com/backgrounds/bg.jpg"`
}

// GetBoardRequest содержит UUID доски для получения информации.
//
//	@Description	Запрос на получение доски по UUID
type GetBoardRequest struct {
	BoardLink uuid.UUID `json:"board_link" example:"123e4567-e89b-12d3-a456-426614174000"`
}

// CreateBoardRequest содержит данные для создания доски.
//
//	@Description	Данные для создания новой доски
type CreateBoardRequest struct {
	Name        string `json:"name"        example:"My Project"`
	Description string `json:"description" example:"Доска для трекинга задач"`
	Background  string `json:"background"  example:"blue"`
}

// UpdateBoardRequest содержит данные для обновления доски.
//
//	@Description	Данные для обновления информации о доске
type UpdateBoardRequest struct {
	BoardLink   uuid.UUID `json:"board_link"  example:"123e4567-e89b-12d3-a456-426614174000"`
	Name        string    `json:"name"        example:"My Project"`
	Description string    `json:"description" example:"Обновлённое описание"`
	Background  string    `json:"background"  example:"green"`
}

// UploadBackgroundRequest содержит данные для загрузки фона доски.
//
//	@Description	Данные для загрузки фонового изображения доски
type UploadBackgroundRequest struct {
	BoardLink uuid.UUID `json:"board_link" example:"123e4567-e89b-12d3-a456-426614174000"`
	Filename  string    `json:"filename"   example:"background.jpg"`
}

// UploadBackgroundResponse возвращает ключ загруженного фона.
//
//	@Description	Ответ после загрузки фона доски
type UploadBackgroundResponse struct {
	BackgroundKey string `json:"background_key" example:"https://s3.example.com/backgrounds/bg.jpg"`
}

// GetMembersRequest содержит UUID доски для получения списка участников.
//
//	@Description	Запрос на получение участников доски
type GetMembersRequest struct {
	BoardLink uuid.UUID `json:"board_link" example:"123e4567-e89b-12d3-a456-426614174000"`
}

// GetMembersResponse содержит список UUID участников доски.
//
//	@Description	Список участников доски
type GetMembersResponse struct {
	UserLinks []uuid.UUID `json:"user_links"`
}

// CreateInviteRequest содержит данные для создания приглашения на доску.
//
//	@Description	Данные для создания приглашения
type CreateInviteRequest struct {
	UserLink       string  `json:"user_link,omitempty" example:"123e4567-e89b-12d3-a456-426614174000"`
	DefaultRole    string  `json:"default_role" example:"editor"`
	ExpireSeconds  int64   `json:"expire_seconds,omitempty" example:"86400"`
}

// CreateInviteResponse содержит информацию о созданном приглашении.
//
//	@Description	Информация о приглашении
type CreateInviteResponse struct {
	InviteLink     string  `json:"invite_link" example:"123e4567-e89b-12d3-a456-426614174000"`
	BoardLink      string  `json:"board_link" example:"123e4567-e89b-12d3-a456-426614174000"`
	TargetUserLink *string `json:"target_user_link,omitempty"`
	DefaultRole    string  `json:"default_role" example:"editor"`
	Status         string  `json:"status" example:"active"`
	ExpireAt       *int64  `json:"expire_at,omitempty" example:"1712928000"`
	CreatedAt      int64   `json:"created_at" example:"1712841600"`
}

// AcceptInviteRequest содержит данные для принятия приглашения.
//
//	@Description	Данные для принятия приглашения
type AcceptInviteRequest struct {
	InviteLink string `json:"-"`
}

// AcceptInviteResponse содержит данные о результате принятия приглашения.
//
//	@Description	Результат принятия приглашения
type AcceptInviteResponse struct {
	BoardLink string `json:"board_link" example:"123e4567-e89b-12d3-a456-426614174000"`
	Role      string `json:"role" example:"editor"`
}

// UpdateMemberRoleRequest содержит данные для изменения роли пользователя на доске.
//
//	@Description	Данные для изменения роли
type UpdateMemberRoleRequest struct {
	NewRole string `json:"new_role" example:"editor"`
}

// InviteInfo содержит информацию о приглашении на доску.
//
//	@Description	Информация о приглашении
type InviteInfo struct {
	InviteLink     string  `json:"invite_link" example:"123e4567-e89b-12d3-a456-426614174000"`
	BoardLink      string  `json:"board_link" example:"123e4567-e89b-12d3-a456-426614174000"`
	TargetUserLink *string `json:"target_user_link,omitempty"`
	DefaultRole    string  `json:"default_role" example:"editor"`
	Status         string  `json:"status" example:"active"`
	ExpireAt       *int64  `json:"expire_at,omitempty" example:"1712928000"`
	CreatedAt      int64   `json:"created_at" example:"1712841600"`
}
