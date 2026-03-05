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
)

type User struct {
	ID uuid.UUID `json:"id"`
	//   uuid Link
	DisplayName  string  `json:"display_name"`
	PasswordHash string  `json:"-"`
	Email        string  `json:"email"`
	Avatar       *string `json:"background,omitempty"`
	// timestamp    CreatedAt
	// timestamp    UpdateAt
}

type MemberBoard struct {
	BoardID uuid.UUID `json:"board_id"`
	UserID  uuid.UUID `json:"user_id"`

	Level     LevelUser `json:"level"`
	IsLike    bool      `json:"is_like"`
	IsArchive bool      `json:"is_archive"`
	// timestamp CreatedAt
	// timestamp UpdateAt
}

type Board struct {
	ID uuid.UUID `json:"id"`
	// uuid Link
	// CreatedAt time.Time `json:"created_at"`
	// UpdateAt
}

type BoardVersion struct {
	ID      uuid.UUID `json:"id"`
	BoardID uuid.UUID `json:"board_id"`

	BoardName   string `json:"board_name"`
	Description string `json:"description"`
	Background  string `json:"background"`

	ValidFrom time.Time  `json:"valid_from"`
	ValidTo   *time.Time `json:"valid_to,omitempty"`
}

type BoardTemplate struct {
	ID       uuid.UUID  `json:"id"`
	AuthorID *uuid.UUID `json:"author_id,omitempty"`

	TemplateName string `json:"template_name"`
	//  timestamp CreatedAt
	//     timestamp UpdateAt
}

type SectionTemplate struct {
	ID         uuid.UUID `json:"id"`
	TemplateID uuid.UUID `json:"template_id"`

	SectionName string `json:"section_name"`
	Position    int    `json:"position"`
	IsMandatory bool   `json:"is_mandatory"`
	MaxTasks    *int   `json:"max_tasks,omitempty"`
	// timestamp   CreatedAt
	// *timestamp   UpdateAt
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
	// timestamp   CreatedAt
	// *timestamp UpdateAt
}

type CommentTask struct {
	ID       uuid.UUID  `json:"id"`
	TaskID   uuid.UUID  `json:"task_id"`
	ParentID *uuid.UUID `json:"parent_id,omitempty"`

	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
	// *timestamp UpdateAt
}

type TaskDependency struct {
	BlockingTaskID uuid.UUID `json:"blocking_task_id"`
	BlockedTaskID  uuid.UUID `json:"blocked_task_id"`
	// timestamp      CreatedAt
	// *timestamp UpdateAt
}

type WorkerTask struct {
	AssigneeID uuid.UUID `json:"assignee_id"`
	TaskID     uuid.UUID `json:"task_id"`

	// timestamp CreatedAt
	// timestamp UpdateAt
}

type ListenerTask struct {
	ListenerID uuid.UUID `json:"listener_id"`
	TaskID     uuid.UUID `json:"task_id"`
	// timestamp  CreatedAt
	// // *timestamp UpdateAt
}
