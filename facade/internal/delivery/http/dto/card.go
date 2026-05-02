package dto

import (
	"time"

	"github.com/google/uuid"
)

// CreateCardRequest содержит данные для создания карточки.
//
//	@Description	Данные для создания новой карточки в секции
type CreateCardRequest struct {
	SectionLink  string     `json:"section_link"`
	Title        string     `json:"title"`
	Description  string     `json:"description"`
	ExecutorLink *string    `json:"executor_link,omitempty"`
	Deadline     *time.Time `json:"deadline,omitempty"`
}

// UpdateCardRequest содержит данные для обновления карточки.
//
//	@Description	Данные для обновления карточки
type UpdateCardRequest struct {
	Title        string     `json:"title"`
	Description  string     `json:"description"`
	ExecutorLink *string    `json:"executor_link,omitempty"`
	Deadline     *time.Time `json:"deadline,omitempty"`
}

// ReorderCardsRequest содержит данные для перемещения карточки.
//
//	@Description	Данные для изменения позиции карточки в секции
type ReorderCardsRequest struct {
	SectionLink string `json:"section_link"`
	Position    int    `json:"position"`
}

// SubtaskResponse описывает подзадачу карточки.
//
//	@Description	Подзадача карточки
type SubtaskResponse struct {
	SubtaskLink uuid.UUID `json:"subtask_link"`
	Description string    `json:"description"`
	IsDone      bool      `json:"is_done"`
	Position    int       `json:"position"`
}

// CardResponse описывает полную информацию о карточке.
//
//	@Description	Полная информация о карточке
type CardResponse struct {
	CardLink     uuid.UUID         `json:"card_link"`
	ExecutorLink *string           `json:"executor_link,omitempty"`
	Title        string            `json:"title"`
	Description  string            `json:"description"`
	Deadline     *time.Time        `json:"deadline,omitempty"`
	Subtasks     []SubtaskResponse `json:"subtasks"`
}

// CreateCardResponse содержит ответ при создании карточки.
//
//	@Description	Ответ при успешном создании карточки
type CreateCardResponse struct {
	CardLink    uuid.UUID `json:"card_link"`
	SectionLink uuid.UUID `json:"section_link"`
	Position    int       `json:"position"`
}

// CreateCommentRequest содержит данные для создания комментария.
//
//	@Description	Данные для создания комментария к карточке
type CreateCommentRequest struct {
	Text       string  `json:"text"`
	ParentLink *string `json:"parent_link,omitempty"`
}

// UpdateCommentRequest содержит данные для обновления комментария.
//
//	@Description	Данные для редактирования текста комментария
type UpdateCommentRequest struct {
	Text string `json:"text"`
}

// CommentResponse описывает комментарий к карточке.
//
//	@Description	Комментарий к карточке
type CommentResponse struct {
	CommentLink uuid.UUID  `json:"comment_link"`
	ParentLink  *uuid.UUID `json:"parent_link,omitempty"`
	AuthorLink  uuid.UUID  `json:"author_link"`
	Text        string     `json:"text"`
}

// CommentsResponse содержит список комментариев к карточке.
//
//	@Description	Список комментариев к карточке
type CommentsResponse struct {
	Comments []CommentResponse `json:"comments"`
}

// CreateCommentResponse содержит ответ при создании комментария.
//
//	@Description	Ответ при успешном создании комментария
type CreateCommentResponse struct {
	CommentLink uuid.UUID `json:"comment_link"`
}

// CreateSubtaskRequest содержит данные для создания подзадачи.
//
//	@Description	Данные для создания подзадачи карточки
type CreateSubtaskRequest struct {
	Description string `json:"description"`
}

// UpdateSubtaskRequest содержит данные для обновления подзадачи.
//
//	@Description	Данные для обновления подзадачи карточки
type UpdateSubtaskRequest struct {
	IsDone      bool   `json:"is_done"`
	Description string `json:"description"`
}
