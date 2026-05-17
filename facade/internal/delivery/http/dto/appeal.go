package dto

import (
	"time"

	"github.com/google/uuid"
)

// AppealInfo содержит полную информацию об обращении.
//
//	@Description	Информация об обращении пользователя
type AppealInfo struct {
	AppealID      int64     `json:"appeal_id"      example:"1"`
	AppealLink    uuid.UUID `json:"appeal_link"    example:"123e4567-e89b-12d3-a456-426614174000"`
	Email         string    `json:"email"          example:"user@example.com"`
	Category      string    `json:"category"       example:"technical"`
	Status        string    `json:"status"         example:"open"`
	DisplayName   string    `json:"display_name"   example:"Ivan Ivanov"`
	Description   string    `json:"description"    example:"Не могу войти в аккаунт"`
	AttachmentURL string    `json:"attachment_url" example:"https://s3.example.com/attachments/file.png"`
	CreatedAt     time.Time `json:"created_at"     example:"2026-04-12T14:35:00Z"`
}

// CreateAppealRequest содержит данные для создания обращения.
//
//	@Description	Данные для создания нового обращения
type CreateAppealRequest struct {
	Email       string `json:"email"        example:"user@example.com"`
	Category    string `json:"category"     example:"technical"`
	Description string `json:"description"  example:"Не могу войти в аккаунт"`
	DisplayName string `json:"display_name" example:"Ivan Ivanov"`
}

// CreateAppealResponse возвращает UUID созданного обращения.
//
//	@Description	Ответ при успешном создании обращения
type CreateAppealResponse struct {
	AppealLink uuid.UUID `json:"appeal_link" example:"123e4567-e89b-12d3-a456-426614174000"`
}

// GetAppealsResponse содержит список обращений и роль пользователя.
//
//	@Description	Список обращений текущего пользователя
type GetAppealsResponse struct {
	Role    string       `json:"role"    example:"user"`
	Appeals []AppealInfo `json:"appeals"`
}

// UploadAttachmentInfo содержит данные для привязки вложения к обращению.
//
//	@Description	Данные вложения обращения
type UploadAttachmentInfo struct {
	AppealLink uuid.UUID `json:"appeal_link" example:"123e4567-e89b-12d3-a456-426614174000"`
	Filename   string    `json:"filename"    example:"screenshot.png"`
}

// UploadAttachmentResponse возвращает URL загруженного вложения.
//
//	@Description	Ответ после загрузки вложения
type UploadAttachmentResponse struct {
	AttachmentURL string `json:"attachment_url" example:"https://s3.example.com/attachments/screenshot.png"`
}

// DeleteInfo содержит UUID обращения для удаления.
//
//	@Description	Данные для удаления обращения
type DeleteInfo struct {
	AppealLink uuid.UUID `json:"appeal_link" example:"123e4567-e89b-12d3-a456-426614174000"`
}

// AppealsStats содержит статистику обращений по статусам.
//
//	@Description	Статистика обращений
type AppealsStats struct {
	OpenAppeals   int64 `json:"open_appeals"    example:"5"`
	InWorkAppeals int64 `json:"in_work_appeals" example:"3"`
	CloseAppeals  int64 `json:"close_appeals"   example:"12"`
}

// ChangeAppealStatusInfo содержит данные для изменения статуса обращения.
//
//	@Description	Данные для изменения статуса обращения
type ChangeAppealStatusInfo struct {
	AppealLink uuid.UUID `json:"appeal_link" example:"123e4567-e89b-12d3-a456-426614174000"`
	NewStatus  string    `json:"new_status"  example:"in_work"`
}
