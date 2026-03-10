package models

import (
	"time"

	"github.com/google/uuid"
)

type LevelUser int

const (
	Viewer LevelUser = iota + 1
	Editor
	Admin
	Creater
)

// User описывает сущность пользователя в системе
//
// @Description Полная информация о пользователе
type User struct {
	ID          uuid.UUID `json:"id"                   example:"123e4567-e89b-12d3-a456-426614174000"`
	DisplayName string    `json:"display_name"         example:"Ivan Ivanov"`
	// PasswordHash не отправляется клиенту
	PasswordHash string `json:"-"                    swaggerignore:"true"`
	Email        string `json:"email"                example:"ivan@mail.com"`
	// Поле имеет тег 'avatar', так оно и будет отображаться в JSON
	Avatar *string `json:"avatar,omitempty" example:"https://example.com/avatar.jpg"`
	Boards []Board `json:"boards"`
}
type MemberBoard struct {
	BoardID uuid.UUID `json:"board_id"`
	UserID  uuid.UUID `json:"user_id"`

	Level     LevelUser `json:"level"`
	IsLike    bool      `json:"is_like"`
	IsArchive bool      `json:"is_archive"`
}

// Board представляет рабочую доску пользователя
//
// @Description Краткая информация о доске
type Board struct {
	ID uuid.UUID `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
}

type BoardTemplate struct {
	ID       uuid.UUID  `json:"id"`
	AuthorID *uuid.UUID `json:"author_id,omitempty"`

	BoardName   string `json:"board_name"`
	Description string `json:"description"`
	Background  string `json:"background"`

	TemplateName string `json:"template_name"`
}

type SectionTemplate struct {
	ID         uuid.UUID `json:"id"`
	TemplateID uuid.UUID `json:"template_id"`

	SectionName string `json:"section_name"`
	Position    int    `json:"position"`
	IsMandatory bool   `json:"is_mandatory"`
	MaxTasks    *int   `json:"max_tasks,omitempty"`
}

type Section struct {
	ID      uuid.UUID `json:"id"`
	BoardID uuid.UUID `json:"board_id"`
	Link    uuid.UUID `json:"link_id"`
}

type SectionVersion struct {
	ID        uuid.UUID `json:"id"`
	SectionID uuid.UUID `json:"section_id"`

	SectionName string `json:"section_name"`
	Position    int    `json:"position"`
	IsMandatory bool   `json:"is_mandatory"`
	MaxTasks    *int   `json:"max_tasks,omitempty"`

	ValidFrom time.Time  `json:"valid_from"`
	ValidTo   *time.Time `json:"valid_to,omitempty"`
}

type Task struct {
	ID        uuid.UUID `json:"id"`
	AuthorID  uuid.UUID `json:"author_id"`
	SectionID uuid.UUID `json:"section_id"`
}

type TaskVersion struct {
	ID     uuid.UUID `json:"id"`
	TaskID uuid.UUID `json:"task_id"`

	SectionID   uuid.UUID `json:"section_id"`
	Title       string    `json:"title"`
	Description *string   `json:"description,omitempty"`
	Position    int       `json:"position"`

	TaskStartAt *time.Time `json:"task_start_at,omitempty"`
	Duedate     *time.Time `json:"due_date,omitempty"`

	ValidFrom time.Time  `json:"valid_from"`
	ValidTo   *time.Time `json:"valid_to,omitempty"`
}

type Subtask struct {
	ID          uuid.UUID `json:"id"`
	TaskID      uuid.UUID `json:"task_id"`
	Description string    `json:"description"`
	IsDone      bool      `json:"is_done"`
	Position    int       `json:"position"`
}

type CommentTask struct {
	ID       uuid.UUID  `json:"id"`
	TaskID   uuid.UUID  `json:"task_id"`
	ParentID *uuid.UUID `json:"parent_id,omitempty"`

	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
}

type TaskDependency struct {
	BlockingTaskID uuid.UUID `json:"blocking_task_id"`
	BlockedTaskID  uuid.UUID `json:"blocked_task_id"`
}

type WorkerTask struct {
	AssigneeID uuid.UUID `json:"assignee_id"`
	TaskID     uuid.UUID `json:"task_id"`
}

type ListenerTask struct {
	ListenerID uuid.UUID `json:"listener_id"`
	TaskID     uuid.UUID `json:"task_id"`
}
