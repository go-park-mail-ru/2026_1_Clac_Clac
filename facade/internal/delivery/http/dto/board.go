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
