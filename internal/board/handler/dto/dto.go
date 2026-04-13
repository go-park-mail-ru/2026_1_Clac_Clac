package dto

import (
	"time"

	"github.com/google/uuid"
)

// CreateBoardRequest описывает модель для создания новой доски
type CreateBoardRequest struct {
	Name        string `json:"name" example:"Project Alpha"`
	Description string `json:"description" example:"Основной рабочий процесс"`
	Background  string `json:"background" example:"#FFFFFF"`
}

// DeleteBoardRequest описывает модель для удаления доски
type DeleteBoardRequest struct {
	Link uuid.UUID `json:"link" example:"123e4567-e89b-12d3-a456-426614174000"`
}

// UpdateBoardRequest описывает модель для обновления данных доски
type UpdateBoardRequest struct {
	Name        string `json:"name" example:"Project Beta"`
	Description string `json:"description" example:"Обновленное описание"`
	Background  string `json:"background" example:"https://example.com/bg.png"`
}

// BoardInfoResponse описывает модель ответа с информацией о доске
type BoardInfoResponse struct {
	Link        uuid.UUID `json:"link" example:"123e4567-e89b-12d3-a456-426614174000"`
	Name        string    `json:"name" example:"Project Alpha"`
	Description string    `json:"description" example:"Основной рабочий процесс"`
	Background  string    `json:"background" example:"#FFFFFF"`
	CreatedAt   time.Time `json:"created_at" example:"2026-04-11T12:10:06Z"`
}

// BackgroundUpdateResponse описывает ответ после успешной загрузки фона
type BackgroundUpdateResponse struct {
	BackgroundURL string `json:"background_url" example:"/static/backgrounds/123e4567.png"`
}
