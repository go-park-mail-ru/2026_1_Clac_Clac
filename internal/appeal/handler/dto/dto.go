package dto

import (
	"time"

	"github.com/go-park-mail-ru/2026_1_Clac_Clac/internal/appeal/common"
	"github.com/google/uuid"
)

// EntityAppealRequest представляет входные данные для создания обращения
// @Description Данные запроса на создание нового обращения (тикета)
type EntityAppealRequest struct {
	Mail        string          `json:"mail" example:"user@example.com"`
	Category    common.Category `json:"category" example:"bug"`
	Description string          `json:"description" example:"При нажатии на кнопку ничего не происходит."`
	DisplayName string          `json:"display_name" example:"Не работает кнопка оплаты"`
}

// Appeal представляет сущность одного обращения
// @Description Полная информация о созданном обращении
type Appeal struct {
	AppealID      int             `json:"appeal_id" example:"42"`
	AppealLink    uuid.UUID       `json:"appeal_link" example:"123e4567-e89b-12d3-a456-426614174000"`
	Email         string          `json:"email" example:"user@example.com"`
	DisplayName   string          `json:"display_name" example:"Не работает кнопка оплаты"`
	Status        common.Status   `json:"status" example:"new"`
	Category      common.Category `json:"category" example:"bug"`
	Description   string          `json:"description" example:"При нажатии на кнопку ничего не происходит."`
	AttachmentKey string          `json:"attachment_key" example:"attachments/screenshot_1.png"`
	CreatedAt     time.Time       `json:"created_at" example:"2023-10-12T07:20:50.52Z"`
}

// Appeals представляет список обращений
// @Description Ответ, содержащий массив обращений
type Appeals struct {
	Role    common.Role `json:"role" example:"user"`
	Appeals []Appeal    `json:"appeals"`
}

// ChangeAppealStatus представляет данные для обновления статуса
// @Description Запрос на изменение статуса обращения
type ChangeAppealStatus struct {
	Status common.Status `json:"status" example:"in_progress"`
}

// AppealStats представляет статистику по обращениям
// @Description Статистика количества обращений по их текущему статусу
type AppealStats struct {
	Open   int `json:"open" example:"15"`
	InWork int `json:"in_work" example:"4"`
	Close  int `json:"close" example:"42"`
}

// UploadAttachmentResponse содержит URL загруженного вложения
// @Description Ответ после загрузки вложения
type UploadAttachmentResponse struct {
	AttachmentURL string `json:"attachment_url" example:"https://bucket.endpoint/attachments/uuid.png"`
}
